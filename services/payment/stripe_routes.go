package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func stripeFromEnv() *StripeProvider {
	return NewStripeProvider(strings.TrimSpace(getenv("STRIPE_SECRET_KEY", "")))
}

func ensureGiftCoupon(s *Server, sub Subscription) (string, error) {
	if sub.GiftTokenID == nil || strings.TrimSpace(*sub.GiftTokenID) == "" || sub.GiftDiscountPercent <= 0 {
		return "", nil
	}
	var token GiftToken
	if err := s.db.Where("id = ?", strings.TrimSpace(*sub.GiftTokenID)).First(&token).Error; err != nil {
		return "", err
	}
	if !giftTokenAvailable(token, time.Now().UTC()) {
		return "", gorm.ErrInvalidData
	}
	if token.ProviderCouponID != nil && strings.TrimSpace(*token.ProviderCouponID) != "" {
		return strings.TrimSpace(*token.ProviderCouponID), nil
	}
	couponID, err := stripeFromEnv().CreateCoupon(token.DiscountPercent, token.MaxRedemptions, "Notell "+token.Type+" "+token.Code, token.ID)
	if err != nil {
		return "", err
	}
	if err := s.db.Model(&GiftToken{}).Where("id = ? AND provider_coupon_id IS NULL", token.ID).Update("provider_coupon_id", couponID).Error; err != nil {
		return "", err
	}
	return couponID, nil
}

func registerStripeRoutes(r *gin.Engine, s *Server) {
	r.POST("/v1/webhooks/stripe", func(c *gin.Context) {
		secret := strings.TrimSpace(getenv("STRIPE_WEBHOOK_SECRET", ""))
		payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid webhook body"})
			return
		}
		if err := verifyStripeSignature(payload, c.GetHeader("Stripe-Signature"), secret, 5*time.Minute); err != nil {
			c.JSON(400, gin.H{"error": "invalid webhook signature"})
			return
		}
		var event struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Data struct {
				Object map[string]any `json:"object"`
			} `json:"data"`
		}
		if err := json.Unmarshal(payload, &event); err != nil || event.ID == "" || event.Type == "" {
			c.JSON(400, gin.H{"error": "invalid Stripe event"})
			return
		}
		var existing WebhookEvent
		if err := s.db.Where("provider = ? AND provider_event_id = ?", "stripe", event.ID).First(&existing).Error; err == nil {
			c.JSON(200, gin.H{"received": true, "duplicate": true})
			return
		}
		wh := WebhookEvent{ID: newID("wh"), Provider: "stripe", ProviderEventID: event.ID, EventType: event.Type, Payload: string(payload), Status: "received"}
		if err := s.db.Create(&wh).Error; err != nil {
			c.JSON(409, gin.H{"error": "webhook already recorded"})
			return
		}
		if err := applyStripeEvent(s, event.Type, event.Data.Object); err != nil {
			now := time.Now()
			s.db.Model(&wh).Updates(map[string]any{"status": "failed", "error_message": err.Error(), "processed_at": now})
			c.JSON(200, gin.H{"received": true})
			return
		}
		if err := applyStripeSubscriptionEventAndSync(s, event.Type, event.Data.Object); err != nil {
			now := time.Now()
			s.db.Model(&wh).Updates(map[string]any{"status": "failed", "error_message": err.Error(), "processed_at": now})
			c.JSON(200, gin.H{"received": true})
			return
		}
		now := time.Now()
		s.db.Model(&wh).Updates(map[string]any{"status": "processed", "processed_at": now})
		c.JSON(200, gin.H{"received": true})
	})

	api := r.Group("/v1")
	api.Use(s.auth)
	api.POST("/subscriptions/:id/checkout", func(c *gin.Context) {
		uid := c.GetString("user_id")
		var sub Subscription
		if err := s.db.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&sub).Error; err != nil {
			c.JSON(404, gin.H{"error": "subscription not found"})
			return
		}
		if sub.Provider != "stripe" || sub.Status != "incomplete" {
			c.JSON(409, gin.H{"error": "subscription is not ready for checkout"})
			return
		}
		success := strings.TrimSpace(os.Getenv("SUBSCRIPTION_SUCCESS_URL"))
		cancel := strings.TrimSpace(os.Getenv("SUBSCRIPTION_CANCEL_URL"))
		if success == "" || cancel == "" {
			c.JSON(500, gin.H{"error": "subscription checkout URLs are not configured"})
			return
		}
		couponID, err := ensureGiftCoupon(s, sub)
		if err != nil {
			c.JSON(502, gin.H{"error": "gift discount could not be prepared"})
			return
		}
		url, err := stripeFromEnv().CreateSubscriptionCheckout(sub, success, cancel, couponID)
		if err != nil {
			c.JSON(502, gin.H{"error": "stripe checkout could not be created"})
			return
		}
		c.JSON(200, gin.H{"url": url, "subscription": sub, "giftDiscountPercent": sub.GiftDiscountPercent})
	})

	api.POST("/payments/:id/intent", func(c *gin.Context) {
		uid := c.GetString("user_id")
		var p Payment
		if err := s.db.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&p).Error; err != nil {
			c.JSON(404, gin.H{"error": "payment not found"})
			return
		}
		if p.Provider != "stripe" || p.Status != PaymentPending {
			c.JSON(409, gin.H{"error": "payment is not ready for Stripe intent"})
			return
		}
		id, clientSecret, err := stripeFromEnv().CreatePaymentIntent(p)
		if err != nil {
			c.JSON(502, gin.H{"error": "stripe payment intent could not be created"})
			return
		}
		if err := s.db.Model(&p).Updates(map[string]any{"provider_payment_id": id, "status": PaymentProcessing}).Error; err != nil {
			c.JSON(500, gin.H{"error": "payment could not be updated"})
			return
		}
		p.Status = PaymentProcessing
		p.ProviderPaymentID = id
		c.JSON(200, gin.H{"payment": p, "client_secret": clientSecret})
	})

	api.POST("/payments/:id/refund", func(c *gin.Context) {
		uid := c.GetString("user_id")
		var in struct {
			Amount int64  `json:"amount"`
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&in); err != nil && err != io.EOF {
			c.JSON(400, gin.H{"error": "invalid refund request"})
			return
		}
		var p Payment
		if err := s.db.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&p).Error; err != nil {
			c.JSON(404, gin.H{"error": "payment not found"})
			return
		}
		if p.Status != PaymentSucceeded || p.ProviderPaymentID == "" {
			c.JSON(409, gin.H{"error": "payment is not refundable"})
			return
		}
		var refunded int64
		if err := s.db.Model(&Refund{}).Where("payment_id = ?", p.ID).Select("COALESCE(SUM(amount),0)").Scan(&refunded).Error; err != nil {
			c.JSON(500, gin.H{"error": "refund balance could not be checked"})
			return
		}
		remaining := p.Amount - refunded
		if remaining <= 0 {
			c.JSON(409, gin.H{"error": "payment has no refundable balance"})
			return
		}
		if in.Amount <= 0 {
			in.Amount = remaining
		}
		if in.Amount > remaining {
			c.JSON(400, gin.H{"error": "refund amount exceeds refundable balance"})
			return
		}
		refundID, status, err := stripeFromEnv().RefundPayment(p, in.Amount, strings.TrimSpace(in.Reason))
		if err != nil {
			c.JSON(502, gin.H{"error": "stripe refund failed"})
			return
		}
		refund := Refund{ID: newID("ref"), PaymentID: p.ID, Amount: in.Amount, Currency: p.Currency, Status: status, ProviderRefundID: refundID, Reason: in.Reason}
		if err := s.db.Create(&refund).Error; err != nil {
			c.JSON(500, gin.H{"error": "refund record could not be saved"})
			return
		}
		c.JSON(201, gin.H{"refund": refund})
	})
}

func applyStripeEvent(s *Server, eventType string, object map[string]any) error {
	providerPaymentID, _ := object["payment_intent"].(string)
	if providerPaymentID == "" { providerPaymentID, _ = object["id"].(string) }
	if providerPaymentID == "" { return nil }
	var p Payment
	if err := s.db.Where("provider = ? AND provider_payment_id = ?", "stripe", providerPaymentID).First(&p).Error; err != nil { return nil }
	target := ""
	switch eventType { case "payment_intent.processing": target = PaymentProcessing; case "payment_intent.succeeded", "charge.succeeded": target = PaymentSucceeded; case "payment_intent.payment_failed", "charge.failed": target = PaymentFailed; case "payment_intent.canceled": target = PaymentCanceled; default: return nil }
	if p.Status == target || !validProviderTransition(p.Status, target) { return nil }
	p.Status = target
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&p).Error; err != nil { return err }
		if target == PaymentSucceeded { var n int64; if err := tx.Model(&LedgerEntry{}).Where("payment_id = ? AND entry_type = ?", p.ID, "payment_captured").Count(&n).Error; err != nil { return err }; if n == 0 { return tx.Create(&LedgerEntry{ID:newID("led"),PaymentID:p.ID,UserID:p.UserID,EntryType:"payment_captured",Amount:p.Amount,Currency:p.Currency,Reference:p.ID}).Error } }
		return nil
	})
}
