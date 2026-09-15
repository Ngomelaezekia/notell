package main

import "time"

// syncSubscriptionEntitlement keeps billing state changes mapped to access state.
// Payment provider details remain isolated inside payment service.
func syncSubscriptionEntitlement(s *Server, sub Subscription) error {
	event := buildEntitlementEvent(
		sub.UserID,
		sub.ResourceType,
		sub.ResourceID,
		sub.Status,
	)

	return upsertEntitlement(s.db, event, sub.PackageID)
}

// entitlementEffectiveEnd is reserved for future subscription expiry handling.
func entitlementEffectiveEnd(sub Subscription) *time.Time {
	if sub.CurrentPeriodEnd.IsZero() {
		return nil
	}
	return &sub.CurrentPeriodEnd
}
