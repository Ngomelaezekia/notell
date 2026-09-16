package main

import (
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

func stripePriceForPackage(p Package) string {
    code := strings.ToUpper(strings.NewReplacer("-", "_", " ", "_").Replace(strings.TrimSpace(p.Code)))
    if code == "" {
        return ""
    }
    return strings.TrimSpace(os.Getenv("STRIPE_PRICE_PKG_" + code))
}

func registerPlanCheckoutRoutes(r *gin.Engine, s *Server) {
    api := r.Group("/v1")
    api.Use(s.auth)

    api.POST("/plans/:id/checkout", func(c *gin.Context) {
        uid := c.GetString("user_id")
        var p Package
        if err := s.db.Table("packages").Where("id = ? AND active = true", strings.TrimSpace(c.Param("id"))).First(&p).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
            return
        }
        priceID := stripePriceForPackage(p)
        if priceID == "" {
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("Stripe price is not configured for plan %s", p.ID)})
            return
        }

        var sub Subscription
        err := s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ? AND package_id = ?", uid, "platform", "channel", p.ID).Order("updated_at desc").First(&sub).Error
        if err != nil {
            now := time.Now().UTC()
            sub = Subscription{ID: newID("sub"), UserID: uid, PackageID: p.ID, ResourceType: "platform", ResourceID: "channel", PriceID: priceID, Provider: "stripe", Status: "incomplete", CurrentPeriodStart: now, CurrentPeriodEnd: now}
            if err := s.db.Create(&sub).Error; err != nil {
                c.JSON(http.StatusConflict, gin.H{"error": "subscription could not be created"})
                return
            }
        } else if subscriptionActive(sub) {
            c.JSON(http.StatusOK, gin.H{"subscription": sub, "active": true})
            return
        } else if sub.Status != "incomplete" {
            c.JSON(http.StatusConflict, gin.H{"error": "existing plan subscription is not available for checkout"})
            return
        }

        success := strings.TrimSpace(os.Getenv("SUBSCRIPTION_SUCCESS_URL"))
        cancel := strings.TrimSpace(os.Getenv("SUBSCRIPTION_CANCEL_URL"))
        if success == "" || cancel == "" {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription checkout URLs are not configured"})
            return
        }
        url, err := stripeFromEnv().CreateSubscriptionCheckout(sub, success, cancel)
        if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Stripe checkout could not be created"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"url": url, "subscription": sub, "active": false})
    })

    api.GET("/plans/status", func(c *gin.Context) {
        uid := c.GetString("user_id")
        var rows []Subscription
        if err := s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", uid, "platform", "channel").Order("updated_at desc").Find(&rows).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load plan status"})
            return
        }
        for _, row := range rows {
            if subscriptionActive(row) {
                c.JSON(http.StatusOK, gin.H{"active": true, "subscription": row})
                return
            }
        }
        c.JSON(http.StatusOK, gin.H{"active": false})
    })
}
