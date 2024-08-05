package mock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RecurringApplicationCharge struct {
	ActivatedOn     *time.Time `json:"activated_on"`
	BillingOn       *time.Time `json:"billing_on"`
	CancelledOn     *time.Time `json:"cancelled_on"`
	CappedAmount    string     `json:"capped_amount"`
	ConfirmationURL string     `json:"confirmation_url"`
	CreatedAt       time.Time  `json:"created_at"`
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Price           float64    `json:"price"`
	ReturnURL       string     `json:"return_url"`
	Status          string     `json:"status"`
	Terms           string     `json:"terms"`
	Test            *bool      `json:"test"`
	TrialDays       int        `json:"trial_days"`
	TrialEndsOn     *time.Time `json:"trial_ends_on"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Currency        string     `json:"currency"`
	ApiClientId     string     `json:"api_client_id"`
}

type Charges struct {
	Charges []RecurringApplicationCharge `json:"recurring_application_charges"`
}

type Charge struct {
	Charge RecurringApplicationCharge `json:"recurring_application_charge"`
}

type Shopify struct {
	mu                          sync.Mutex
	recurringApplicationCharges map[int]RecurringApplicationCharge
}

func NewShopify() *Shopify {
	s := &Shopify{
		recurringApplicationCharges: make(map[int]RecurringApplicationCharge),
	}
	go s.cleanupCharges()
	return s
}

func (s *Shopify) cleanupCharges() {
	for {
		time.Sleep(24 * time.Hour)
		s.mu.Lock()
		for id, charge := range s.recurringApplicationCharges {
			if time.Since(charge.CreatedAt) > 24*time.Hour {
				delete(s.recurringApplicationCharges, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Shopify) createRecurringApplicationCharge(c *gin.Context) {
	var chargeWrapper Charge
	if err := c.BindJSON(&chargeWrapper); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	charge := chargeWrapper.Charge

	// Generiere eine ID für den neuen Charge
	s.mu.Lock()
	idBase := 10000000
	charge.ID = len(s.recurringApplicationCharges) + idBase
	now := time.Now()
	charge.CreatedAt = now
	charge.UpdatedAt = now

	// Fülle die Felder aus
	charge.ActivatedOn = &now
	charge.BillingOn = &now
	charge.CancelledOn = &now
	charge.CappedAmount = "100"
	// Bestimme die aktuelle Domain
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	domain := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	charge.ConfirmationURL = fmt.Sprintf("%s/confirm/%d", domain, charge.ID)
	charge.Status = "pending"
	charge.Terms = "Standard Terms"
	test := true
	charge.Test = &test
	charge.TrialEndsOn = &now
	charge.Currency = "USD"
	charge.ApiClientId = "123456"

	s.recurringApplicationCharges[charge.ID] = charge
	s.mu.Unlock()

	chargeWrapper.Charge = charge
	c.JSON(http.StatusOK, chargeWrapper)
}

func (s *Shopify) getRecurringApplicationCharges(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var charges []RecurringApplicationCharge
	for _, charge := range s.recurringApplicationCharges {
		charges = append(charges, charge)
	}

	c.JSON(http.StatusOK, Charges{charges})
}

func (s *Shopify) Param(c *gin.Context, key string) string {
	value := c.Param(key + ".json")
	if value != "" {
		value = value[:len(value)-5]
	}

	return value
}

func (s *Shopify) getRecurringApplicationCharge(c *gin.Context) {
	idStr := s.Param(c, "recurring_application_charge_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid charge ID"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	charge, exists := s.recurringApplicationCharges[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

	c.JSON(http.StatusOK, Charge{charge})
}

func (s *Shopify) updateRecurringApplicationCharge(c *gin.Context) {
	idStr := c.Param("recurring_application_charge_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid charge ID"})
		return
	}

	var update RecurringApplicationCharge
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	charge, exists := s.recurringApplicationCharges[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

	charge.CappedAmount = update.CappedAmount
	s.recurringApplicationCharges[id] = charge

	c.JSON(http.StatusOK, Charge{charge})
}

func (s *Shopify) deleteRecurringApplicationCharge(c *gin.Context) {
	idStr := s.Param(c, "recurring_application_charge_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid charge ID"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.recurringApplicationCharges[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

	delete(s.recurringApplicationCharges, id)
	c.JSON(http.StatusOK, gin.H{"message": "Charge deleted"})
}

func (s *Shopify) confirmCharge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid charge ID"})
		return
	}

	s.mu.Lock()
	charge, exists := s.recurringApplicationCharges[id]
	if !exists {
		s.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}
	s.mu.Unlock()

	action := c.Query("action")
	if action == "accept" {
		charge.Status = "active"
	} else if action == "decline" {
		charge.Status = "declined"
	} else {
		c.HTML(http.StatusOK, "confirm.html", gin.H{
			"charge": charge,
		})
		return
	}

	s.mu.Lock()
	s.recurringApplicationCharges[id] = charge
	s.mu.Unlock()

	// Leite zur ReturnURL weiter
	c.Redirect(http.StatusFound, charge.ReturnURL)
}

func (s *Shopify) Routes(engine *gin.Engine) {
	shopify := engine.Group("/admin/api/:version/")
	{
		shopify.GET("/recurring_application_charges.json", s.getRecurringApplicationCharges)
		shopify.POST("/recurring_application_charges.json", s.createRecurringApplicationCharge)
		shopify.GET("/recurring_application_charges/:recurring_application_charge_id.json", s.getRecurringApplicationCharge)
		shopify.PUT("/recurring_application_charges/:recurring_application_charge_id/customize.json", s.updateRecurringApplicationCharge)
		shopify.DELETE("/recurring_application_charges/:recurring_application_charge_id.json", s.deleteRecurringApplicationCharge)
		engine.GET("/confirm/:id", s.confirmCharge)
	}
}
