package main

// applyStripeSubscriptionEventAndSync applies Stripe subscription changes and
// then updates the entitlement boundary consumed by channel/live/message
// services. Keeping this wrapper separate avoids changing webhook behavior
// until callers are migrated to the synchronized path.
func applyStripeSubscriptionEventAndSync(s *Server, eventType string, object map[string]any) error {
	if err := applyStripeSubscriptionEvent(s, eventType, object); err != nil {
		return err
	}

	providerID, _ := object["id"].(string)
	if providerID == "" {
		return nil
	}

	var sub Subscription
	if err := s.db.Where("provider = ? AND provider_subscription_id = ?", "stripe", providerID).First(&sub).Error; err != nil {
		return nil
	}

	return syncSubscriptionEntitlementState(s, sub)
}
