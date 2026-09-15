package main

import "time"

// Subscription lifecycle to entitlement state mapping.
// Channel, live and message services consume entitlement state and never
// depend directly on payment provider details.

func entitlementStatusForSubscription(status string) string {
 switch status {
 case PaymentSucceeded, "active":
  return "active"
 case "past_due", PaymentFailed:
  return "suspended"
 case PaymentCanceled, "canceled", "expired":
  return "revoked"
 default:
  return "pending"
 }
}

type EntitlementEvent struct {
 UserID string
 ResourceType string
 ResourceID string
 Status string
 EffectiveAt time.Time
 Reason string
}

func buildEntitlementEvent(userID, resourceType, resourceID, subscriptionStatus string) EntitlementEvent {
 return EntitlementEvent{
  UserID:userID,
  ResourceType:resourceType,
  ResourceID:resourceID,
  Status:entitlementStatusForSubscription(subscriptionStatus),
  EffectiveAt:time.Now().UTC(),
  Reason:"subscription_state_changed",
 }
}
