package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Subscription struct {
	ID                     string    `gorm:"primaryKey" json:"id"`
	UserID                 string    `gorm:"index;not null" json:"userId"`
	PackageID              string    `gorm:"not null" json:"packageId"`
	ResourceType           string    `gorm:"not null;index:idx_subscription_resource" json:"resourceType"`
	ResourceID             string    `gorm:"not null;index:idx_subscription_resource" json:"resourceId"`
	PriceID                string    `gorm:"not null" json:"priceId"`
	Provider               string    `json:"provider"`
	ProviderSubscriptionID string    `gorm:"uniqueIndex" json:"providerSubscriptionId,omitempty"`
	Status                 string    `gorm:"index;not null" json:"status"`
	CurrentPeriodStart     time.Time `json:"currentPeriodStart"`
	CurrentPeriodEnd       time.Time `json:"currentPeriodEnd"`
	CancelAtPeriodEnd      bool      `json:"cancelAtPeriodEnd"`
	LastSyncedAt           time.Time `json:"lastSyncedAt"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

func subscriptionActive(s Subscription) bool {
	return (s.Status == "active" || s.Status == "trialing" || s.Status == "past_due" || s.Status == "canceled") && s.CurrentPeriodEnd.After(time.Now().UTC())
}

func registerSubscriptionRoutes(r *gin.Engine, s *Server) {
	api := r.Group("/v1")
	api.Use(s.auth)
	api.GET("/subscriptions", func(c *gin.Context) {
		uid := c.GetString("user_id")
		var rows []Subscription
		if err := s.db.Where("user_id = ?", uid).Order("created_at desc").Find(&rows).Error; err != nil {
			c.JSON(500, gin.H{"error": "failed to load subscriptions"})
			return
		}
		c.JSON(200, gin.H{"subscriptions": rows})
	})
	api.GET("/entitlements/:resourceType/:resourceId", func(c *gin.Context) {
		uid := c.GetString("user_id")
		var row Subscription
		if err := s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", uid, c.Param("resourceType"), c.Param("resourceId")).Order("updated_at desc").First(&row).Error; err != nil {
			c.JSON(200, gin.H{"active": false})
			return
		}
		c.JSON(200, gin.H{"active": subscriptionActive(row), "subscription": row})
	})
	r.GET("/v1/internal/entitlements/:resourceType/:resourceId", func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-Internal-Service-Key"))
		if key == "" || key != strings.TrimSpace(getenv("INTERNAL_SERVICE_KEY", "")) {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}
		ownerID := strings.TrimSpace(c.GetHeader("X-User-ID"))
		if ownerID == "" {
			c.JSON(400, gin.H{"error": "X-User-ID is required"})
			return
		}
		var rows []Subscription
		if err := s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", ownerID, c.Param("resourceType"), c.Param("resourceId")).Order("updated_at desc").Find(&rows).Error; err != nil {
			c.JSON(500, gin.H{"error": "failed to load entitlement"})
			return
		}
		for _, row := range rows {
			if subscriptionActive(row) {
				c.JSON(200, gin.H{"active": true, "subscription": row})
				return
			}
		}
		c.JSON(200, gin.H{"active": false})
	})
	api.POST("/subscriptions", func(c *gin.Context) {
		var in struct {
			ResourceType string `json:"resourceType" binding:"required"`
			ResourceID   string `json:"resourceId" binding:"required"`
			PriceID      string `json:"priceId" binding:"required"`
		}
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"error": "resourceType, resourceId and priceId are required"})
			return
		}
		in.ResourceType = strings.TrimSpace(in.ResourceType)
		in.ResourceID = strings.TrimSpace(in.ResourceID)
		in.PriceID = strings.TrimSpace(in.PriceID)
		uid := c.GetString("user_id")
		var existing Subscription
		if err := s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", uid, in.ResourceType, in.ResourceID).First(&existing).Error; err == nil {
			c.JSON(200, gin.H{"subscription": existing, "idempotent": true})
			return
		}
		now := time.Now().UTC()
		row := Subscription{ID: newID("sub"), UserID: uid, PackageID: in.PriceID, ResourceType: in.ResourceType, ResourceID: in.ResourceID, PriceID: in.PriceID, Provider: "stripe", Status: "incomplete", CurrentPeriodStart: now, CurrentPeriodEnd: now}
		if err := s.db.Create(&row).Error; err != nil {
			c.JSON(409, gin.H{"error": "subscription could not be created"})
			return
		}
		c.JSON(201, gin.H{"subscription": row, "next": "checkout"})
	})
}

func stripeObjectString(object map[string]any, key string) string {
	v, _ := object[key].(string)
	return strings.TrimSpace(v)
}

func stripeObjectUnix(object map[string]any, key string) (time.Time, bool) {
	v, ok := object[key].(float64)
	if !ok || v <= 0 {
		return time.Time{}, false
	}
	return time.Unix(int64(v), 0).UTC(), true
}

func stripeSubscriptionIDForEvent(eventType string, object map[string]any) string {
	if strings.HasPrefix(eventType, "customer.subscription.") {
		return stripeObjectString(object, "id")
	}
	return stripeObjectString(object, "subscription")
}

func applyStripeSubscriptionEvent(s *Server, eventType string, object map[string]any) error {
	providerID := stripeSubscriptionIDForEvent(eventType, object)
	if providerID == "" {
		return nil
	}
	var sub Subscription
	err := s.db.Where("provider = ? AND provider_subscription_id = ?", "stripe", providerID).First(&sub).Error
	if err != nil {
		meta, _ := object["metadata"].(map[string]any)
		uid, _ := meta["user_id"].(string)
		rt, _ := meta["resource_type"].(string)
		rid, _ := meta["resource_id"].(string)
		if uid == "" || rt == "" || rid == "" {
			return nil
		}
		if err = s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", uid, rt, rid).Order("updated_at desc").First(&sub).Error; err != nil {
			return nil
		}
		sub.ProviderSubscriptionID = providerID
	}

	status := stripeObjectString(object, "status")
	switch eventType {
	case "customer.subscription.created", "customer.subscription.updated":
	case "customer.subscription.deleted":
		status = "canceled"
	case "invoice.paid":
		status = "active"
	case "invoice.payment_failed":
		status = "past_due"
	case "customer.subscription.trial_will_end":
		return nil
	default:
		return nil
	}
	if status != "" {
		sub.Status = status
	}
	sub.CancelAtPeriodEnd = false
	if v, ok := object["cancel_at_period_end"].(bool); ok {
		sub.CancelAtPeriodEnd = v
	}
	if ts, ok := stripeObjectUnix(object, "current_period_start"); ok {
		sub.CurrentPeriodStart = ts
	}
	if ts, ok := stripeObjectUnix(object, "current_period_end"); ok {
		sub.CurrentPeriodEnd = ts
	}
	if sub.CurrentPeriodEnd.IsZero() {
		if ts, ok := stripeObjectUnix(object, "period_end"); ok {
			sub.CurrentPeriodEnd = ts
		}
	}
	sub.LastSyncedAt = time.Now().UTC()
	return s.db.Save(&sub).Error
}
