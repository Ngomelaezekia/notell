package main

func syncSubscriptionEntitlementState(s *Server, sub Subscription) error {
	event := buildEntitlementEvent(
		sub.UserID,
		sub.ResourceType,
		sub.ResourceID,
		sub.Status,
		sub.CurrentPeriodEnd,
	)
	return upsertEntitlement(s.db, event, sub.PackageID)
}
