package mock

import (
	"fmt"
	"strings"
	"time"
)

type RecurringApplicationCharge struct {
	ActivatedOn         *time.Time `json:"activated_on"`
	BillingOn           *time.Time `json:"billing_on"`
	CancelledOn         *time.Time `json:"cancelled_on"`
	CappedAmount        string     `json:"capped_amount"`
	ConfirmationURL     string     `json:"confirmation_url"`
	CreatedAt           time.Time  `json:"created_at"`
	ID                  int        `json:"id"`
	Name                string     `json:"name"`
	Price               float64    `json:"price"`
	ReturnURL           string     `json:"return_url"`
	Status              string     `json:"status"`
	Terms               string     `json:"terms"`
	Test                *bool      `json:"test"`
	TrialDays           int        `json:"trial_days"`
	TrialEndsOn         *time.Time `json:"trial_ends_on"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Currency            string     `json:"currency"`
	ApiClientId         string     `json:"api_client_id"`
	DecoratedReturnUrl  string     `json:"decorated_return_url"`
	ReplacementBehavior string     `json:"replacementBehavior"`
}

func (charge RecurringApplicationCharge) GetReplacementBehavior() string {
	if charge.ReplacementBehavior == "" {
		return "default"
	} else if charge.ReplacementBehavior == "APPLY_ON_NEXT_BILLING_CYCLE" {
		return "next billing"
	} else {
		return "default"
	}
}

// GetSubscription erstellt aus einem Charge eine Map, die dem GraphQL-Typ "AppSubscription" entspricht.
func (charge RecurringApplicationCharge) GetSubscription() RecurringApplicationChargeGraphQl {
	testVal := false
	if charge.Test != nil {
		testVal = *charge.Test
	}

	return RecurringApplicationChargeGraphQl{
		Gid:              fmt.Sprintf("gid://shopify/AppSubscription/%d", charge.ID),
		Name:             charge.Name,
		Status:           strings.ToUpper(charge.Status),
		CreatedAt:        charge.CreatedAt,
		Test:             testVal,
		ReturnURL:        charge.ReturnURL,
		CurrentPeriodEnd: *charge.TrialEndsOn,
		TrialDays:        charge.TrialDays,
		LineItems: []LineItem{
			{
				Id: "gid://shopify/LineItem/1",
				Plan: Plan{
					PricingDetails: PriceDetails{
						Price: Price{
							Amount:   fmt.Sprintf("%.2f", charge.Price),
							Currency: charge.Currency,
						},
						Interval: "EVERY_30_DAYS",
					},
				},
			},
		},
	}
}
