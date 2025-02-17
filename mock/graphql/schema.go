package graphql

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql/language/ast"
	"go-slack-ics/mock"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go-slack-ics/system"
)

// ShopifyGraphQl ist der Haupt-Typ, der unser Schema und weitere Abhängigkeiten hält.
type ShopifyGraphQl struct {
	redis  *system.Redis
	schema graphql.Schema
	// store wird pro Request gesetzt (um Datenrennen zu vermeiden)
	store  string
	domain string
}

// URLScalar repräsentiert einen URL-Scalar-Typ, der intern als String behandelt wird.
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

func NewShopifyGraphQl() *ShopifyGraphQl {
	s := &ShopifyGraphQl{
		redis: system.NewRedis(),
	}

	// ─── Query Type ──────────────────────────────────────────────

	appInstallationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AppInstallation",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.ID},
			"activeSubscriptions": &graphql.Field{
				Type: graphql.NewList(AppSubscriptionType),
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

					var subscription mock.RecurringApplicationCharge
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
			"node": &graphql.Field{
				Type: AppSubscriptionType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("id muss ein String sein")
					}
					parts := strings.Split(id, "/")
					numStr := parts[len(parts)-1]
					numId, _ := strconv.Atoi(numStr)

					key := fmt.Sprintf("recurring_application_charge:%d", numId)
					var charge mock.RecurringApplicationCharge
					if err := s.redis.Get(key, "no-alias", &charge); err == nil {
						graphqlCharge := charge.GetSubscription()
						return graphqlCharge, nil
					}

					return nil, fmt.Errorf("Subscription not found")
				},
			},
		},
	})

	// ─── Mutation Type ──────────────────────────────────────────────

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"appSubscriptionCreate": &graphql.Field{
				Type: AppSubscriptionCreatePayloadType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
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
						Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(AppSubscriptionLineItemInputType))),
					},
					"replacementBehavior": &graphql.ArgumentConfig{
						Type: ReplacementBehaviorEnum,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name := p.Args["name"].(string)
					returnUrl := p.Args["returnUrl"].(string)
					trialDays, _ := p.Args["trialDays"].(int)
					replacementBehavior, _ := p.Args["replacementBehavior"].(string)

					var testPtr *bool
					if testVal, ok := p.Args["test"].(bool); ok {
						testPtr = new(bool)
						*testPtr = testVal
					}

					// Extrahiere das Array der lineItems – hier verwenden wir das erste Element
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

					unixTime := time.Now().Unix()
					confirmationUrl := fmt.Sprintf("%s/confirm/%d", s.domain, int(unixTime))
					trialEndsOn := time.Now().Add(time.Duration(trialDays) * 24 * time.Hour)

					subscription := mock.RecurringApplicationCharge{
						ActivatedOn:         nil,
						BillingOn:           nil,
						CancelledOn:         nil,
						CappedAmount:        "100",
						ConfirmationURL:     confirmationUrl,
						CreatedAt:           time.Now(),
						ID:                  int(unixTime),
						Name:                name,
						Price:               amount,
						ReturnURL:           returnUrl,
						Status:              "PENDING",
						Terms:               "Standard Terms",
						Test:                testPtr,
						TrialDays:           trialDays,
						TrialEndsOn:         &trialEndsOn,
						UpdatedAt:           time.Now(),
						Currency:            currency,
						ApiClientId:         "",
						DecoratedReturnUrl:  "",
						ReplacementBehavior: replacementBehavior,
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

// GraphQLHandler verarbeitet die HTTP-Anfrage.
func (s *ShopifyGraphQl) GraphQLHandler(c *gin.Context) {
	// Lese den "store"-Parameter (z. B. aus dem Header "X-Shopify-Access-Token")
	s.store = c.GetHeader("X-Shopify-Access-Token")
	// Bestimme die Domain und erzeuge die ConfirmationURL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	s.domain = fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	h := handler.New(&handler.Config{
		Schema:   &s.schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Erstelle unseren ResponseCapture, um die Ausgabe abzufangen.
	capture := &ResponseCapture{
		ResponseWriter: c.Writer,
		statusCode:     http.StatusOK,
	}

	// Lasse den Handler die Anfrage bearbeiten und schreibe in unseren Capture.
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
