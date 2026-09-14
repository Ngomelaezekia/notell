package main

// Phase 2 subscription primitives stay behind the payment service boundary.
// Provider credentials and recurring billing activation remain configuration-gated.

type SubscriptionStatus string

const (
	SubscriptionIncomplete SubscriptionStatus = "incomplete"
	SubscriptionActive     SubscriptionStatus = "active"
	SubscriptionPastDue    SubscriptionStatus = "past_due"
	SubscriptionCanceled   SubscriptionStatus = "canceled"
	SubscriptionPaused     SubscriptionStatus = "paused"
)

type PlatformFee struct {
	Amount   int64
	Currency string
}

func calculatePlatformFee(amount, basisPoints int64) PlatformFee {
	fee, ok := bpsCeil(amount, basisPoints)
	if !ok {
		return PlatformFee{}
	}
	return PlatformFee{Amount: fee}
}
