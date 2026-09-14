package main

import "testing"

func TestMonthlyStreamingScenarios(t *testing.T) {
	if got := monthlyStreamMinutes(30, 30); got != 900 { t.Fatalf("30 min/day should be 900 min/month, got %d", got) }
	if got := monthly24x7Hours(30); got != 720 { t.Fatalf("24/7 should be 720 hours/month, got %d", got) }
}

func TestCalculateCostIncludesViewerDeliveryAndMargin(t *testing.T) {
	in := CostInput{StreamMinutes:900, ViewerMinutes:150000, TranscodeMinutes:900, OperationalMicros:10_000_000, RiskReserveBps:500, PlatformMarginBps:2000}
	r := CostRates{StreamMinuteMicros:100, ViewerMinuteMicros:1000, TranscodeMinuteMicros:200, ComputeMinuteMicros:0, CDNMinuteMicros:0, StorageGBMonthMicros:0}
	got, ok := CalculateCost(in, r)
	if !ok { t.Fatal("cost calculation failed") }
	if got.DeliveryMicros != 150_000_000 { t.Fatalf("unexpected delivery cost: %d", got.DeliveryMicros) }
	if got.CustomerPriceMicros <= got.InfrastructureMicros { t.Fatalf("customer price must exceed infrastructure cost") }
}

func TestCalculateCostRejectsOverflow(t *testing.T) {
	_, ok := CalculateCost(CostInput{ViewerMinutes: 2}, CostRates{ViewerMinuteMicros: 1<<62})
	if ok { t.Fatal("expected overflow rejection") }
}
