package main

import "time"

func entitlementStatusForSubscription(status string, periodEnd time.Time) string {
	now := time.Now().UTC()
	if !periodEnd.IsZero() && !periodEnd.After(now) {
		return "revoked"
	}
	switch status {
	case PaymentSucceeded, "active", "trialing", "past_due":
		return "active"
	case "canceled", "expired":
		return "revoked"
	default:
		return "pending"
	}
}

type EntitlementEvent struct {
	UserID       string
	ResourceType string
	ResourceID   string
	Status       string
	EffectiveAt  time.Time
	Reason       string
	EndsAt       *time.Time
}

func buildEntitlementEvent(userID, resourceType, resourceID, subscriptionStatus string, periodEnd time.Time) EntitlementEvent {
	status := entitlementStatusForSubscription(subscriptionStatus, periodEnd)
	var endsAt *time.Time
	if !periodEnd.IsZero() {
		end := periodEnd.UTC()
		endsAt = &end
	}
	return EntitlementEvent{
		UserID: userID, ResourceType: resourceType, ResourceID: resourceID,
		Status: status, EffectiveAt: time.Now().UTC(), Reason: "subscription_state_changed", EndsAt: endsAt,
	}
}
