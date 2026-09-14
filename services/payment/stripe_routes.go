package main

import (
    "encoding/json"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

func stripeFromEnv() *StripeProvider { return NewStripeProvider(strings.TrimSpace(getenv("STRIPE_SECRET_KEY", ""))) }

func registerStripeRoutes(r *gin.Engine, s *Server) {
    // Webhooks are intentionally outside JWT auth. Stripe authenticates them with its signing secret.
    r.POST("/v1/webhooks/stripe", func(c *gin.Context) {
        secret := strings.TrimSpace(getenv("STRIPE_WEBHOOK_SECRET", ""))
        payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
        if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook body"}); return }
        if err := verifyStripeSignature(payload, c.GetHeader("Stripe-Signature"), secret, 5*time.Minute); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"}); return
        }
        var event struct { ID string `json:"id"`; Type string `json:"type"`; Data struct { Object map[string]any `json:"object"` } `json:"data"` }
        if err := json.Unmarshal(payload, &event); err != nil || event.ID == "" || event.Type == "" { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Stripe event"}); return }

        var existing WebhookEvent
        if err := s.db.Where("provider = ? AND provider_event_id = ?", "stripe", event.ID).First(&existing).Error; err == nil {
            c.JSON(http.StatusOK, gin.H{"received": true, "duplicate": true}); return
        }
        wh := WebhookEvent{ID: newID("wh"), Provider: "stripe", ProviderEventID: event.ID, EventType: event.Type, Payload: string(payload), Status: "received"}
        if err := s.db.Create(&wh).Error; err != nil { c.JSON(http.StatusConflict, gin.H{"error": "webhook already recorded"}); return }
        if err := applyStripeEvent(s, event.Type, event.Data.Object); err != nil {
            now := time.Now(); s.db.Model(&wh).Updates(map[string]any{"status": "failed", "error_message": err.Error(), "processed_at": now})
            c.JSON(http.StatusOK, gin.H{"received": true}); return
        }
        now := time.Now(); s.db.Model(&wh).Updates(map[string]any{"status": "processed", "processed_at": now})
        c.JSON(http.StatusOK, gin.H{"received": true})
    })

    api := r.Group("/v1")
    api.Use(s.auth)

    api.POST("/payments/:id/intent", func(c *gin.Context) {
        uid := c.GetString("user_id")
        var p Payment
        if err := s.db.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&p).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"}); return }
        if p.Provider != "stripe" { c.JSON(http.StatusBadRequest, gin.H{"error": "payment provider is not stripe"}); return }
        if p.Status != PaymentPending { c.JSON(http.StatusConflict, gin.H{"error": "payment is not pending"}); return }
        provider := stripeFromEnv(); id, clientSecret, err := provider.CreatePaymentIntent(p)
        if err != nil { c.JSON(http.StatusBadGateway, gin.H{"error": "stripe payment intent could not be created"}); return }
        if err := s.db.Model(&p).Updates(map[string]any{"provider_payment_id": id, "status": PaymentProcessing}).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "payment could not be updated"}); return }
        _ = s.db.Create(&AuditEvent{ID:newID("aud"), PaymentID:p.ID, UserID:uid, EventType:"stripe_payment_intent_created", FromStatus:PaymentPending, ToStatus:PaymentProcessing, Metadata:"{}"}).Error
        p.Status = PaymentProcessing; p.ProviderPaymentID = id
        c.JSON(http.StatusOK, gin.H{"payment": p, "client_secret": clientSecret})
    })

    api.POST("/payments/:id/refund", func(c *gin.Context) {
        uid := c.GetString("user_id")
        var in struct { Amount int64 `json:"amount"`; Reason string `json:"reason"` }
        if err := c.ShouldBindJSON(&in); err != nil && err != io.EOF { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund request"}); return }
        var p Payment
        if err := s.db.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&p).Error; err != nil { c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"}); return }
        if p.Status != PaymentSucceeded || p.ProviderPaymentID == "" { c.JSON(http.StatusConflict, gin.H{"error": "payment is not refundable"}); return }
        if in.Amount < 0 || (in.Amount > 0 && in.Amount > p.Amount) { c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund amount"}); return }
        provider := stripeFromEnv(); refundID, status, err := provider.RefundPayment(p, in.Amount, strings.TrimSpace(in.Reason))
        if err != nil { c.JSON(http.StatusBadGateway, gin.H{"error": "stripe refund failed"}); return }
        refund := Refund{ID:newID("ref"), PaymentID:p.ID, Amount:in.Amount, Currency:p.Currency, Status:status, ProviderRefundID:refundID, Reason:in.Reason}
        if err := s.db.Create(&refund).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "refund record could not be saved"}); return }
        _ = s.db.Create(&AuditEvent{ID:newID("aud"), PaymentID:p.ID, UserID:uid, EventType:"refund_created", Metadata:"{}"}).Error
        c.JSON(http.StatusCreated, gin.H{"refund": refund})
    })
}

func applyStripeEvent(s *Server, eventType string, object map[string]any) error {
    providerPaymentID, _ := object["payment_intent"].(string)
    if providerPaymentID == "" { providerPaymentID, _ = object["id"].(string) }
    if providerPaymentID == "" { return nil }
    var p Payment
    if err := s.db.Where("provider = ? AND provider_payment_id = ?", "stripe", providerPaymentID).First(&p).Error; err != nil { return nil }
    target := ""
    switch eventType {
    case "payment_intent.processing": target = PaymentProcessing
    case "payment_intent.succeeded", "charge.succeeded": target = PaymentSucceeded
    case "payment_intent.payment_failed", "charge.failed": target = PaymentFailed
    case "payment_intent.canceled": target = PaymentCanceled
    default: return nil
    }
    if p.Status == target { return nil }
    if !validTransition(p.Status, target) { return nil }
    from := p.Status
    p.Status = target
    if err := s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(&p).Error; err != nil { return err }
        if err := tx.Create(&AuditEvent{ID:newID("aud"), PaymentID:p.ID, UserID:p.UserID, EventType:"stripe_webhook_status", FromStatus:from, ToStatus:target, Metadata:"{}"}).Error; err != nil { return err }
        if target == PaymentSucceeded { return tx.Create(&LedgerEntry{ID:newID("led"), PaymentID:p.ID, UserID:p.UserID, EntryType:"payment_captured", Amount:p.Amount, Currency:p.Currency, Reference:p.ID}).Error }
        return nil
    }); err != nil { return err }
    return nil
}
