package main

import "math"

type CostInput struct {
	StreamMinutes       int64
	ViewerMinutes       int64
	TranscodeMinutes    int64
	StorageGBMonths     int64
	ComputeMinutes      int64
	CDNMinutes          int64
	PaymentFeeMicros    int64
	OperationalMicros   int64
	RiskReserveBps      int64
	PlatformMarginBps   int64
}

type CostRates struct {
	StreamMinuteMicros    int64
	ViewerMinuteMicros    int64
	TranscodeMinuteMicros int64
	StorageGBMonthMicros  int64
	ComputeMinuteMicros   int64
	CDNMinuteMicros       int64
}

type CostBreakdown struct {
	StreamingMicros  int64
	DeliveryMicros   int64
	TranscodeMicros  int64
	StorageMicros    int64
	ComputeMicros    int64
	CDNMicros        int64
	PaymentMicros    int64
	OperationsMicros int64
	RiskMicros       int64
	InfrastructureMicros int64
	PlatformFeeMicros    int64
	CustomerPriceMicros  int64
}

func nonNegative(v int64) int64 { if v < 0 { return 0 }; return v }

func mulSafe(a, b int64) (int64, bool) {
	if a < 0 || b < 0 { return 0, false }
	if a == 0 || b == 0 { return 0, true }
	if a > math.MaxInt64/b { return 0, false }
	return a * b, true
}

func bpsCeil(amount, bps int64) (int64, bool) {
	if amount <= 0 || bps <= 0 { return 0, true }
	q, r := amount/10000, amount%10000
	a, ok := mulSafe(q, bps); if !ok { return 0, false }
	b, ok := mulSafe(r, bps); if !ok { return 0, false }
	b = (b + 9999) / 10000
	if a > math.MaxInt64-b { return 0, false }
	return a + b, true
}

func CalculateCost(in CostInput, r CostRates) (CostBreakdown, bool) {
	var out CostBreakdown
	var ok bool
	parts := [][3]int64{
		{nonNegative(in.StreamMinutes), nonNegative(r.StreamMinuteMicros), 0},
		{nonNegative(in.ViewerMinutes), nonNegative(r.ViewerMinuteMicros), 1},
		{nonNegative(in.TranscodeMinutes), nonNegative(r.TranscodeMinuteMicros), 2},
		{nonNegative(in.StorageGBMonths), nonNegative(r.StorageGBMonthMicros), 3},
		{nonNegative(in.ComputeMinutes), nonNegative(r.ComputeMinuteMicros), 4},
		{nonNegative(in.CDNMinutes), nonNegative(r.CDNMinuteMicros), 5},
	}
	for _, p := range parts {
		v, good := mulSafe(p[0], p[1]); if !good { return CostBreakdown{}, false }
		switch p[2] { case 0: out.StreamingMicros=v; case 1: out.DeliveryMicros=v; case 2: out.TranscodeMicros=v; case 3: out.StorageMicros=v; case 4: out.ComputeMicros=v; case 5: out.CDNMicros=v }
	}
	out.PaymentMicros = nonNegative(in.PaymentFeeMicros)
	out.OperationsMicros = nonNegative(in.OperationalMicros)
	for _, v := range []int64{out.StreamingMicros,out.DeliveryMicros,out.TranscodeMicros,out.StorageMicros,out.ComputeMicros,out.CDNMicros,out.PaymentMicros,out.OperationsMicros} {
		if math.MaxInt64-out.InfrastructureMicros < v { return CostBreakdown{}, false }
		out.InfrastructureMicros += v
	}
	out.RiskMicros, ok = bpsCeil(out.InfrastructureMicros, nonNegative(in.RiskReserveBps)); if !ok { return CostBreakdown{}, false }
	base, ok := addSafe(out.InfrastructureMicros, out.RiskMicros); if !ok { return CostBreakdown{}, false }
	out.PlatformFeeMicros, ok = bpsCeil(base, nonNegative(in.PlatformMarginBps)); if !ok { return CostBreakdown{}, false }
	out.CustomerPriceMicros, ok = addSafe(base, out.PlatformFeeMicros); if !ok { return CostBreakdown{}, false }
	return out, true
}

func addSafe(a,b int64) (int64,bool) { if b > 0 && a > math.MaxInt64-b { return 0,false }; return a+b,true }

func monthlyStreamMinutes(dailyMinutes, days int64) int64 { if dailyMinutes <= 0 || days <= 0 { return 0 }; v,ok:=mulSafe(dailyMinutes,days); if !ok{return 0}; return v }
func monthly24x7Hours(days int64) int64 { if days <= 0 { return 0 }; v,ok:=mulSafe(days,24); if !ok{return 0}; return v }
