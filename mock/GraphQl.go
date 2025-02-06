package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/handler"
	"go-slack-ics/system"
)

// responseCapture fängt die Response ab, bevor sie an den Client geschickt wird.
type responseCapture struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	// Hier rufen wir WriteHeader nicht sofort auf, sondern machen das später
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	return rc.buf.Write(b)
}

// Definiere einen neuen Scalar-Typ "URL", der intern als String behandelt wird.
var URLScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "URL",
	Description: "Der URL-Scalar-Typ repräsentiert eine URL als String.",
	Serialize: func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			return s
		}
		return nil
	},
	ParseValue: func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			return s
		}
		return nil
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		if strVal, ok := valueAST.(*ast.StringValue); ok {
			return strVal.Value
		}
		return nil
	},
})

// GetSubscription erstellt aus einem Charge eine Map, die dem GraphQL-Typ "AppSubscription" entspricht.
func (r RecurringApplicationCharge) GetSubscription() map[string]interface{} {
	testVal := false
	if r.Test != nil {
		testVal = *r.Test
	}
	return map[string]interface{}{
		"id":               fmt.Sprintf("gid://shopify/AppSubscription/%d", r.ID),
		"name":             r.Name,
		"status":           r.Status,
		"createdAt":        r.CreatedAt.Format(time.RFC3339),
		"test":             testVal,
		"returnUrl":        r.ReturnURL,
		"trialDays":        r.TrialDays,
		"currentPeriodEnd": r.TrialEndsOn.Format(time.RFC3339),
	}
}

type ShopifyGraphQl struct {
	redis  *system.Redis
	schema graphql.Schema
	// store wird pro Request gesetzt (in der Produktion per Context, um Datenrennen zu vermeiden)
	store string
}

func NewShopifyGraphQl() *ShopifyGraphQl {
	s := &ShopifyGraphQl{
		redis: system.NewRedis(),
	}

	// ─── OUTPUT TYPES ───────────────────────────────────────────────

	// Typ "AppSubscription" (Output: id, name, status, createdAt, test)
	appSubscriptionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppSubscription",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if charge, ok := p.Source.(RecurringApplicationCharge); ok {
						return fmt.Sprintf("gid://shopify/AppSubscription/%d", charge.ID), nil
					}
					if m, ok := p.Source.(map[string]interface{}); ok {
						return m["id"], nil
					}
					return nil, nil
				},
			},
			"name":             &graphql.Field{Type: graphql.String},
			"status":           &graphql.Field{Type: graphql.String},
			"createdAt":        &graphql.Field{Type: graphql.String},
			"currentPeriodEnd": &graphql.Field{Type: graphql.String},
			"trialDays":        &graphql.Field{Type: graphql.Int},
			"test":             &graphql.Field{Type: graphql.Boolean},
			"returnUrl":        &graphql.Field{Type: graphql.String},
		},
	})

	// Typ "UserError"
	userErrorType := graphql.NewObject(graphql.ObjectConfig{
		Name: "UserError",
		Fields: graphql.Fields{
			"message": &graphql.Field{Type: graphql.String},
			"field":   &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	// Payload-Typ für die Mutation "appSubscriptionCreate"
	appSubscriptionCreatePayloadType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppSubscriptionCreatePayload",
		Fields: graphql.Fields{
			"confirmationUrl": &graphql.Field{Type: graphql.String},
			"appSubscription": &graphql.Field{Type: appSubscriptionType},
			"userErrors":      &graphql.Field{Type: graphql.NewList(userErrorType)},
		},
	})

	// ─── INPUT TYPES ───────────────────────────────────────────────

	// Input für Price
	priceInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "PriceInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"amount": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Float),
			},
			"currencyCode": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// Input für AppRecurringPricingDetails
	appRecurringPricingDetailsInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "AppRecurringPricingDetailsInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"price": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(priceInputType),
			},
			"interval": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	// Input für Plan
	planInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "PlanInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"appRecurringPricingDetails": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(appRecurringPricingDetailsInputType),
			},
		},
	})

	// Input für AppSubscriptionLineItem
	appSubscriptionLineItemInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "AppSubscriptionLineItemInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"plan": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(planInputType),
			},
		},
	})

	// ─── ENUM-TYPE ─────────────────────────────────────────────

	replacementBehaviorEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "AppSubscriptionReplacementBehavior",
		Values: graphql.EnumValueConfigMap{
			"STANDARD":                    &graphql.EnumValueConfig{Value: "STANDARD"},
			"APPLY_ON_NEXT_BILLING_CYCLE": &graphql.EnumValueConfig{Value: "APPLY_ON_NEXT_BILLING_CYCLE"},
			"APPLY_IMMEDIATELY":           &graphql.EnumValueConfig{Value: "APPLY_IMMEDIATELY"},
		},
	})

	// ─── QUERY-TYPE ─────────────────────────────────────────────

	appInstallationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppInstallation",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.ID},
			"activeSubscriptions": &graphql.Field{
				Type: graphql.NewList(appSubscriptionType),
			},
		},
	})
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"currentAppInstallation": &graphql.Field{
				Type: appInstallationType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					store := s.store
					key := "appInstallation:" + store

					var subscription RecurringApplicationCharge
					err := s.redis.Get(key, s.getAliasKey(), &subscription)
					if err != nil {
						return map[string]interface{}{
							"id":                  "gid://shopify/AppInstallation/811826315597",
							"activeSubscriptions": []interface{}{},
						}, nil
					}

					return map[string]interface{}{
						"id":                  "gid://shopify/AppInstallation/811826315597",
						"activeSubscriptions": []interface{}{subscription.GetSubscription()},
					}, nil
				},
			},
		},
	})

	// ─── MUTATION-TYPE ─────────────────────────────────────────────

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"appSubscriptionCreate": &graphql.Field{
				Type: appSubscriptionCreatePayloadType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					// Verwende den neuen URLScalar statt graphql.String
					"returnUrl": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(URLScalar),
					},
					"test": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
					"trialDays": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"lineItems": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(appSubscriptionLineItemInputType))),
					},
					"replacementBehavior": &graphql.ArgumentConfig{
						Type: replacementBehaviorEnum,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name := p.Args["name"].(string)
					returnUrl := p.Args["returnUrl"].(string)
					trialDays, _ := p.Args["trialDays"].(int)

					var testPtr *bool
					if testVal, ok := p.Args["test"].(bool); ok {
						testPtr = new(bool)
						*testPtr = testVal
					}

					// Extrahiere das Array der lineItems – hier verwenden wir einfach das erste Element
					lineItemsArg, ok := p.Args["lineItems"].([]interface{})
					if !ok || len(lineItemsArg) == 0 {
						return nil, fmt.Errorf("lineItems müssen als nicht-leeres Array übergeben werden")
					}
					lineItem, ok := lineItemsArg[0].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("ungültiges Format in lineItems")
					}
					plan, ok := lineItem["plan"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("ungültiges Format bei plan")
					}
					appRecurringPricingDetails, ok := plan["appRecurringPricingDetails"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("ungültiges Format bei appRecurringPricingDetails")
					}
					price, ok := appRecurringPricingDetails["price"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("ungültiges Format bei price")
					}
					amount, ok := price["amount"].(float64)
					if !ok {
						return nil, fmt.Errorf("amount muss vom Typ Float sein")
					}
					currency, ok := price["currencyCode"].(string)
					if !ok {
						return nil, fmt.Errorf("currencyCode muss ein String sein")
					}

					// Generiere eine eindeutige ID basierend auf der aktuellen Unix-Zeit
					unixTime := time.Now().Unix()
					confirmationUrl := fmt.Sprintf("https://langify-testing-prod.myshopify.com/admin/charges/168501/%d/RecurringApplicationCharge/confirm_recurring_application_charge?signature=%s",
						unixTime, "BAh7BzoHaWRsKwhNgXT0FQA6EmF1dG9fYWN0aXZhdGVU--simuliert")

					trialEndsOn := time.Now().Add(time.Duration(trialDays) * 24 * time.Hour)

					subscription := RecurringApplicationCharge{
						ActivatedOn:        nil,
						BillingOn:          nil,
						CancelledOn:        nil,
						CappedAmount:       "100",
						ConfirmationURL:    confirmationUrl,
						CreatedAt:          time.Now(),
						ID:                 int(unixTime),
						Name:               name,
						Price:              amount,
						ReturnURL:          returnUrl,
						Status:             "PENDING",
						Terms:              "Standard Terms",
						Test:               testPtr,
						TrialDays:          trialDays,
						TrialEndsOn:        &trialEndsOn,
						UpdatedAt:          time.Now(),
						Currency:           currency,
						ApiClientId:        "",
						DecoratedReturnUrl: "",
					}

					key := fmt.Sprintf("recurring_application_charge:%d", unixTime)
					s.redis.Set(key, subscription, s.getAliasKey())

					payload := map[string]interface{}{
						"confirmationUrl": confirmationUrl,
						"appSubscription": subscription,
						"userErrors":      []interface{}{},
					}

					return payload, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
	s.schema = schema

	return s
}

func (s *ShopifyGraphQl) getAliasKey() string {
	return fmt.Sprintf("recurring_application_charge_by_store:%s", s.store)
}

func (s *ShopifyGraphQl) GraphQLHandler(c *gin.Context) {
	// Lese den "store"-Parameter (z. B. aus dem Header "X-Shopify-Access-Token")
	s.store = c.GetHeader("X-Shopify-Access-Token")

	h := handler.New(&handler.Config{
		Schema:   &s.schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Erstelle unseren responseCapture, um die Ausgabe abzufangen.
	capture := &responseCapture{
		ResponseWriter: c.Writer,
		statusCode:     http.StatusOK,
	}

	// Lasse den Handler die Anfrage bearbeiten und schreibe in unseren capture.
	h.ServeHTTP(capture, c.Request)

	// Hole den Response-Body aus dem Buffer.
	bodyBytes := capture.buf.Bytes()

	// Versuche, die Antwort als JSON zu interpretieren.
	var responseData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		// Falls das fehlschlägt, schreibe einfach die ursprünglichen Bytes.
		c.Writer.Write(bodyBytes)
		return
	}

	// Füge das Extensions-Feld hinzu.
	responseData["extensions"] = map[string]interface{}{
		"cost": map[string]interface{}{
			"requestedQueryCost": 10,
			"actualQueryCost":    10,
			"throttleStatus": map[string]interface{}{
				"maximumAvailable":   2000,
				"currentlyAvailable": 1990,
				"restoreRate":        100,
			},
		},
	}

	// Marshall die modifizierte Antwort zurück in JSON.
	modifiedResponse, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		c.Writer.Write(bodyBytes)
		return
	}

	// Setze die Content-Length und den Status-Code, dann schreibe die modifizierte Antwort.
	c.Writer.Header().Set("Content-Length", strconv.Itoa(len(modifiedResponse)))
	c.Writer.WriteHeader(capture.statusCode)
	c.Writer.Write(modifiedResponse)
}
