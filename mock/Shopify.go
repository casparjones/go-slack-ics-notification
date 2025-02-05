package mock

import (
	"fmt"
	"go-slack-ics/system"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Shopify struct {
	redis *system.Redis
}

func NewShopify() *Shopify {
	s := &Shopify{
		redis: system.NewRedis(),
	}
	go s.cleanupCharges()
	return s
}

// cleanupCharges löscht Charges, die älter als 24 Stunden sind.
func (s *Shopify) cleanupCharges() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		listKey := "recurring_application_charges_ids"
		keys, err := s.redis.LRange(listKey, 0, -1)
		if err != nil {
			continue
		}
		for _, key := range keys {
			var charge RecurringApplicationCharge
			if err := s.redis.Get(key, &charge); err != nil {
				continue
			}
			if time.Since(charge.CreatedAt) > 24*time.Hour {
				s.redis.Del(key)
				s.redis.LRem(listKey, 0, key)
			}
		}
	}
}

func (s *Shopify) createRecurringApplicationCharge(c *gin.Context) {
	var chargeWrapper Charge
	if err := c.BindJSON(&chargeWrapper); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	charge := chargeWrapper.Charge

	// Generiere eine eindeutige ID, z.B. basierend auf UnixNano
	charge.ID = int(time.Now().UnixNano() & 0x7fffffff)
	now := time.Now()
	charge.CreatedAt = now
	charge.UpdatedAt = now
	charge.ActivatedOn = &now
	charge.BillingOn = &now
	charge.CancelledOn = &now
	charge.CappedAmount = "100"

	// Bestimme die Domain und erzeuge die ConfirmationURL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	domain := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	charge.ConfirmationURL = fmt.Sprintf("%s/confirm/%d", domain, charge.ID)
	charge.Status = "pending"
	charge.Terms = "Standard Terms"
	charge.TrialEndsOn = &now
	charge.Currency = "USD"
	charge.ApiClientId = "123456"

	// Speichere den Charge in Redis
	key := fmt.Sprintf("recurring_application_charge:%d", charge.ID)
	if err := s.redis.Set(key, charge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Füge den Key der Liste aller Charges hinzu
	listKey := "recurring_application_charges_ids"
	if err := s.redis.LPush(listKey, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chargeWrapper.Charge = charge
	c.JSON(http.StatusOK, chargeWrapper)
}

func (s *Shopify) getRecurringApplicationCharges(c *gin.Context) {
	listKey := "recurring_application_charges_ids"
	keys, err := s.redis.LRange(listKey, 0, -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var charges []RecurringApplicationCharge
	for _, key := range keys {
		var charge RecurringApplicationCharge
		if err := s.redis.Get(key, &charge); err == nil {
			charges = append(charges, charge)
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Not Found"})
		return
	}
	key := fmt.Sprintf("recurring_application_charge:%d", id)
	var charge RecurringApplicationCharge
	if err := s.redis.Get(key, &charge); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": "Not Found"})
		return
	}
	c.JSON(http.StatusOK, Charge{charge})
}

func (s *Shopify) updateRecurringApplicationCharge(c *gin.Context) {
	idStr := c.Param("recurring_application_charge_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Not Found"})
		return
	}

	var update RecurringApplicationCharge
	if err := c.BindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := fmt.Sprintf("recurring_application_charge:%d", id)
	var charge RecurringApplicationCharge
	if err := s.redis.Get(key, &charge); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

	charge.CappedAmount = update.CappedAmount
	charge.UpdatedAt = time.Now()

	if err := s.redis.Set(key, charge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Charge{charge})
}

func (s *Shopify) deleteRecurringApplicationCharge(c *gin.Context) {
	idStr := s.Param(c, "recurring_application_charge_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Not Found"})
		return
	}
	key := fmt.Sprintf("recurring_application_charge:%d", id)

	if err := s.redis.Del(key); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

	// Entferne den Key auch aus der Liste
	listKey := "recurring_application_charges_ids"
	s.redis.LRem(listKey, 0, key)
	c.JSON(http.StatusOK, gin.H{"message": "Charge deleted"})
}

func (s *Shopify) confirmCharge(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "Not Found"})
		return
	}
	key := fmt.Sprintf("recurring_application_charge:%d", id)
	var charge RecurringApplicationCharge
	if err := s.redis.Get(key, &charge); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charge not found"})
		return
	}

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

	if err := s.redis.Set(key, charge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	}

	shopify = engine.Group("/:store/admin/api/:version/")
	{
		shopify.GET("/recurring_application_charges.json", s.getRecurringApplicationCharges)
		shopify.POST("/recurring_application_charges.json", s.createRecurringApplicationCharge)
		shopify.GET("/recurring_application_charges/:recurring_application_charge_id.json", s.getRecurringApplicationCharge)
		shopify.PUT("/recurring_application_charges/:recurring_application_charge_id/customize.json", s.updateRecurringApplicationCharge)
		shopify.DELETE("/recurring_application_charges/:recurring_application_charge_id.json", s.deleteRecurringApplicationCharge)
		engine.GET("/confirm/:id", s.confirmCharge)
	}
}
