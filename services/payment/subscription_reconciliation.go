package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func stripeReconciliationInterval() time.Duration {
	value := strings.TrimSpace(os.Getenv("STRIPE_RECONCILIATION_INTERVAL"))
	if value == "" {
		return 6 * time.Hour
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func reconcileStripeSubscriptions(s *Server) error {
	provider := stripeFromEnv()
	if !provider.enabled() {
		return nil
	}

	var rows []Subscription
	if err := s.db.Where("provider = ? AND provider_subscription_id <> ''", "stripe").Find(&rows).Error; err != nil {
		return err
	}

	var firstErr error
	for _, sub := range rows {
		object, err := provider.RetrieveSubscription(sub.ProviderSubscriptionID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := applyStripeSubscriptionEventAndSync(s, "customer.subscription.updated", object); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func startStripeSubscriptionReconciliation(s *Server) {
	interval := stripeReconciliationInterval()
	if interval <= 0 || !stripeFromEnv().enabled() {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := reconcileStripeSubscriptions(s); err != nil {
				log.Printf("stripe subscription reconciliation failed: %v", err)
			}
		}
	}()
}
