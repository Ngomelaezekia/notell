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

func registerStripeRoutes(r *gin.Engine, s *Server) {

	r.POST("/v1/webhooks/stripe", func(c *gin.Context) {

		// signature verification
		// webhook idempotency
		// event parsing
		// webhook storage

		if err := applyStripeEvent(s, event.Type, event.Data.Object); err != nil {
			// failure handling
			return
		}

		// ✅ FIXED LINE
		if err := applyStripeSubscriptionEventAndSync(
			s,
			event.Type,
			event.Data.Object,
		); err != nil {
			now := time.Now()

			s.db.Model(&wh).Updates(map[string]any{
				"status":        "failed",
				"error_message": err.Error(),
				"processed_at":  now,
			})

			c.JSON(200, gin.H{
				"received": true,
			})
			return
		}

		now := time.Now()

		s.db.Model(&wh).Updates(map[string]any{
			"status":       "processed",
			"processed_at": now,
		})

		c.JSON(200, gin.H{
			"received": true,
		})
	})
}
