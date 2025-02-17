package mock

import "time"

type Price struct {
	Amount   string `json:"amount"`
	Currency string `json:"currencyCode"`
}

type PriceDetails struct {
	Price    Price  `json:"price"`
	Interval string `json:"interval"`
}

type Plan struct {
	PricingDetails PriceDetails `json:"pricingDetails"`
}

type LineItem struct {
	Plan Plan `json:"plan"`
}

type RecurringApplicationChargeGraphQl struct {
	Gid              string     `json:"id"`
	CreatedAt        time.Time  `json:"createdAt"`
	CurrentPeriodEnd time.Time  `json:"currentPeriodEnd"`
	Name             string     `json:"name"`
	ReturnURL        string     `json:"returnUrl"`
	Status           string     `json:"status"`
	Test             bool       `json:"test"`
	LineItems        []LineItem `json:"lineItems"`
	TrialDays        int        `json:"trialDays"`
}

type Charges struct {
	Charges []RecurringApplicationCharge `json:"recurring_application_charges"`
}

type Charge struct {
	Charge RecurringApplicationCharge `json:"recurring_application_charge"`
}
