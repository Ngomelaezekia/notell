package main

// syncSubscriptionEntitlementState applies the subscription lifecycle state
// to the entitlement boundary used by channel/live/message services.
// The caller remains responsible for wrapping this in the same transaction as
// subscription state changes when atomic payment processing is required.
func syncSubscriptionEntitlementState(s *Server, sub Subscription) error {
	event := buildEntitlementEvent(
		sub.UserID,
		sub.ResourceType,
		sub.ResourceID,
		sub.Status,
	)
	return upsertEntitlement(s.db, event, sub.PackageID)
}
