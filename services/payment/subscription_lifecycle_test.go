package main

import (
	"testing"
	"time"
)

func TestSubscriptionActiveGraceAndExpiry(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	for _, status := range []string{"active", "trialing", "past_due", "canceled"} {
		if !subscriptionActive(Subscription{Status: status, CurrentPeriodEnd: future}) {
			t.Fatalf("expected %s subscription to remain active before period end", status)
		}
	}
	for _, status := range []string{"active", "trialing", "past_due", "canceled"} {
		if subscriptionActive(Subscription{Status: status, CurrentPeriodEnd: past}) {
			t.Fatalf("expected %s subscription to be inactive after period end", status)
		}
	}
}

func TestStripeSubscriptionIDForEvent(t *testing.T) {
	if got := stripeSubscriptionIDForEvent("customer.subscription.updated", map[string]any{"id": "sub_123"}); got != "sub_123" {
		t.Fatalf("got %q, want sub_123", got)
	}
	if got := stripeSubscriptionIDForEvent("invoice.paid", map[string]any{"id": "in_123", "subscription": "sub_123"}); got != "sub_123" {
		t.Fatalf("got %q, want sub_123", got)
	}
}

func TestEntitlementStatusRespectsPeriodEnd(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)
	if got := entitlementStatusForSubscription("canceled", future); got != "active" {
		t.Fatalf("got %q, want active during paid period", got)
	}
	if got := entitlementStatusForSubscription("past_due", future); got != "active" {
		t.Fatalf("got %q, want active during grace window", got)
	}
	if got := entitlementStatusForSubscription("active", past); got != "revoked" {
		t.Fatalf("got %q, want revoked after period end", got)
	}
}
