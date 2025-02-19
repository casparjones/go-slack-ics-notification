package graphql

import (
	"fmt"
	"go-slack-ics/mock"

	"github.com/graphql-go/graphql"
)

// ─── Output Types ──────────────────────────────────────────────

// PriceType definiert den Preis.
var PriceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Price",
	Fields: graphql.Fields{
		"amount":       &graphql.Field{Type: graphql.String},
		"currencyCode": &graphql.Field{Type: graphql.String},
	},
})

// PriceDetailsType umschließt den Price.
var PriceDetailsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PriceDetails",
	Fields: graphql.Fields{
		"price": &graphql.Field{Type: PriceType},
	},
})

// PlanType enthält PricingDetails und das Abrechnungsintervall.
var PlanType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Plan",
	Fields: graphql.Fields{
		"pricingDetails": &graphql.Field{
			Type: AppRecurringPricingType,
		},
	},
})

// LineItemType enthält den Plan.
var LineItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LineItem",
	Fields: graphql.Fields{
		"id":   &graphql.Field{Type: graphql.String},
		"plan": &graphql.Field{Type: PlanType},
	},
})

// AppSubscriptionType repräsentiert ein Abo (Hinweis: Die Typen
// RecurringApplicationCharge und RecurringApplicationChargeGraphQl
// müssen in deinem Projekt definiert sein).
var AppSubscriptionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AppSubscription",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// Prüfe die möglichen Typen
				if charge, ok := p.Source.(mock.RecurringApplicationCharge); ok {
					return fmt.Sprintf("gid://shopify/AppSubscription/%d", charge.ID), nil
				}
				if m, ok := p.Source.(map[string]interface{}); ok {
					return m["id"], nil
				}
				if charge, ok := p.Source.(mock.RecurringApplicationChargeGraphQl); ok {
					return fmt.Sprintf("%s", charge.Gid), nil
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
		"lineItems": &graphql.Field{
			Type: graphql.NewList(LineItemType),
		},
	},
})

// UserErrorType für Fehlermeldungen.
var UserErrorType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UserError",
	Fields: graphql.Fields{
		"message": &graphql.Field{Type: graphql.String},
		"field":   &graphql.Field{Type: graphql.NewList(graphql.String)},
	},
})

// AppSubscriptionCreatePayloadType als Rückgabe der Mutation.
var AppSubscriptionCreatePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AppSubscriptionCreatePayload",
	Fields: graphql.Fields{
		"confirmationUrl": &graphql.Field{Type: graphql.String},
		"appSubscription": &graphql.Field{Type: AppSubscriptionType},
		"userErrors":      &graphql.Field{Type: graphql.NewList(UserErrorType)},
	},
})

// ─── Input Types ──────────────────────────────────────────────

// PriceInputType definiert den Input für Price.
var PriceInputType = graphql.NewInputObject(graphql.InputObjectConfig{
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

// AppRecurringPricingDetailsInputType für die Preisdaten und das Intervall.
var AppRecurringPricingDetailsInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AppRecurringPricingDetailsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"price": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(PriceInputType),
		},
		"interval": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

// PlanInputType fasst die PricingDetails zusammen.
var PlanInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "PlanInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appRecurringPricingDetails": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(AppRecurringPricingDetailsInputType),
		},
	},
})

// AppSubscriptionLineItemInputType enthält den Plan.
var AppSubscriptionLineItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AppSubscriptionLineItemInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"plan": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(PlanInputType),
		},
	},
})

// ─── Enum Type ──────────────────────────────────────────────

// ReplacementBehaviorEnum definiert das Verhalten beim Austausch eines Abos.
var ReplacementBehaviorEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "AppSubscriptionReplacementBehavior",
	Values: graphql.EnumValueConfigMap{
		"STANDARD":                    &graphql.EnumValueConfig{Value: "STANDARD"},
		"APPLY_ON_NEXT_BILLING_CYCLE": &graphql.EnumValueConfig{Value: "APPLY_ON_NEXT_BILLING_CYCLE"},
		"APPLY_IMMEDIATELY":           &graphql.EnumValueConfig{Value: "APPLY_IMMEDIATELY"},
	},
})

var AppRecurringPricingType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AppRecurringPricing",
	Fields: graphql.Fields{
		"price": &graphql.Field{
			Type: PriceType, // sollte z. B. Felder amount und currencyCode enthalten
		},
		"interval": &graphql.Field{
			Type: graphql.String,
		},
	},
})
