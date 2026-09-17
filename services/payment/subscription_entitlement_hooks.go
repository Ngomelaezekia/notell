package main

import (
	"strings"
	"sync"
)

var stripeReconciliationOnce sync.Once

func applyStripeSubscriptionEventAndSync(s *Server, eventType string, object map[string]any) error {
	stripeReconciliationOnce.Do(func() { startStripeSubscriptionReconciliation(s) })
	if err := applyStripeSubscriptionEvent(s, eventType, object); err != nil {
		return err
	}
	providerID := stripeSubscriptionIDForEvent(eventType, object)
	if providerID == "" {
		return nil
	}
	var sub Subscription
	if err := s.db.Where("provider = ? AND provider_subscription_id = ?", "stripe", providerID).First(&sub).Error; err != nil {
		return nil
	}
	if err := syncSubscriptionEntitlementState(s, sub); err != nil {
		return err
	}
	if sub.Status == "active" && sub.GiftTokenID != nil && strings.TrimSpace(*sub.GiftTokenID) != "" {
		return redeemGiftToken(s.db, *sub.GiftTokenID, sub.UserID, sub.ID)
	}
	return nil
}
