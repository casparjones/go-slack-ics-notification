package mock

import "time"

type RecurringApplicationCharge struct {
	ActivatedOn        *time.Time `json:"activated_on"`
	BillingOn          *time.Time `json:"billing_on"`
	CancelledOn        *time.Time `json:"cancelled_on"`
	CappedAmount       string     `json:"capped_amount"`
	ConfirmationURL    string     `json:"confirmation_url"`
	CreatedAt          time.Time  `json:"created_at"`
	ID                 int        `json:"id"`
	Name               string     `json:"name"`
	Price              float64    `json:"price"`
	ReturnURL          string     `json:"return_url"`
	Status             string     `json:"status"`
	Terms              string     `json:"terms"`
	Test               *bool      `json:"test"`
	TrialDays          int        `json:"trial_days"`
	TrialEndsOn        *time.Time `json:"trial_ends_on"`
	UpdatedAt          time.Time  `json:"updated_at"`
	Currency           string     `json:"currency"`
	ApiClientId        string     `json:"api_client_id"`
	DecoratedReturnUrl string     `json:"decorated_return_url"`
}

type Charges struct {
	Charges []RecurringApplicationCharge `json:"recurring_application_charges"`
}

type Charge struct {
	Charge RecurringApplicationCharge `json:"recurring_application_charge"`
}
