package main

import "time"

func syncSubscriptionEntitlement(s *Server, sub Subscription) error {
	event := buildEntitlementEvent(
		sub.UserID,
		sub.ResourceType,
		sub.ResourceID,
		sub.Status,
		sub.CurrentPeriodEnd,
	)
	return upsertEntitlement(s.db, event, sub.PackageID)
}

func entitlementEffectiveEnd(sub Subscription) *time.Time {
	if sub.CurrentPeriodEnd.IsZero() {
		return nil
	}
	return &sub.CurrentPeriodEnd
}
