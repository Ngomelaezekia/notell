package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CostEstimateRequest struct {
	StreamMinutes    int64 `json:"stream_minutes"`
	ViewerCount      int64 `json:"viewer_count"`
	ViewerMinutes    int64 `json:"viewer_minutes"`
	TranscodeMinutes int64 `json:"transcode_minutes"`
	StorageGBMonths  int64 `json:"storage_gb_months"`
	ComputeMinutes   int64 `json:"compute_minutes"`
	CDNMinutes       int64 `json:"cdn_minutes"`
	PaymentFeeMicros int64 `json:"payment_fee_micros"`
	OperationalMicros int64 `json:"operational_micros"`
	RiskReserveBps   int64 `json:"risk_reserve_bps"`
	PlatformMarginBps int64 `json:"platform_margin_bps"`
	Currency         string `json:"currency"`
	AdRevenueMicros  int64 `json:"ad_revenue_micros"`
	CreatorAdShareBps int64 `json:"creator_ad_share_bps"`
}

type CostEstimateResponse struct {
	Input CostEstimateRequest `json:"input"`
	ViewerMinutes int64 `json:"viewer_minutes"`
	Breakdown CostBreakdown `json:"breakdown"`
	InfrastructureMicros int64 `json:"infrastructure_micros"`
	PlatformFeeMicros int64 `json:"platform_fee_micros"`
	CustomerPriceMicros int64 `json:"customer_price_micros"`
	AdRevenueMicros int64 `json:"ad_revenue_micros"`
	CreatorAdRevenueMicros int64 `json:"creator_ad_revenue_micros"`
	NotellAdRevenueMicros int64 `json:"notell_ad_revenue_micros"`
	Currency string `json:"currency"`
}

func registerCostRoutes(r *gin.Engine, s *Server) {
	api := r.Group("/v1")
	api.Use(s.auth)
	api.POST("/cost/estimate", func(c *gin.Context) {
		var in CostEstimateRequest
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cost estimate request"})
			return
		}
		if in.StreamMinutes < 0 || in.ViewerCount < 0 || in.ViewerMinutes < 0 || in.TranscodeMinutes < 0 || in.StorageGBMonths < 0 || in.ComputeMinutes < 0 || in.CDNMinutes < 0 || in.PaymentFeeMicros < 0 || in.OperationalMicros < 0 || in.AdRevenueMicros < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cost inputs cannot be negative"})
			return
		}
		if in.ViewerMinutes == 0 && in.ViewerCount > 0 && in.StreamMinutes > 0 {
			v, ok := mulSafe(in.ViewerCount, in.StreamMinutes)
			if !ok { c.JSON(http.StatusBadRequest, gin.H{"error": "viewer minute calculation overflow"}); return }
			in.ViewerMinutes = v
		}
		if in.RiskReserveBps < 0 || in.RiskReserveBps > 10000 || in.PlatformMarginBps < 0 || in.PlatformMarginBps > 10000 || in.CreatorAdShareBps < 0 || in.CreatorAdShareBps > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "basis-point values must be between 0 and 10000"})
			return
		}
		currency := strings.ToLower(strings.TrimSpace(in.Currency))
		if currency == "" { currency = "usd" }
		if len(currency) != 3 { c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be 3 characters"}); return }

		// Provider-agnostic default rate profile. These are configuration-safe
		// estimates, not hard-coded customer prices. Provider rates can replace them later.
		rates := CostRates{
			StreamMinuteMicros: 0,
			ViewerMinuteMicros: 1000,
			TranscodeMinuteMicros: 0,
			StorageGBMonthMicros: 0,
			ComputeMinuteMicros: 0,
			CDNMinuteMicros: 0,
		}
		costInput := CostInput{
			StreamMinutes: in.StreamMinutes,
			ViewerMinutes: in.ViewerMinutes,
			TranscodeMinutes: in.TranscodeMinutes,
			StorageGBMonths: in.StorageGBMonths,
			ComputeMinutes: in.ComputeMinutes,
			CDNMinutes: in.CDNMinutes,
			PaymentFeeMicros: in.PaymentFeeMicros,
			OperationalMicros: in.OperationalMicros,
			RiskReserveBps: in.RiskReserveBps,
			PlatformMarginBps: in.PlatformMarginBps,
		}
		breakdown, ok := CalculateCost(costInput, rates)
		if !ok { c.JSON(http.StatusBadRequest, gin.H{"error": "cost calculation overflow"}); return }

		creatorAd, ok := bpsCeil(in.AdRevenueMicros, in.CreatorAdShareBps)
		if !ok { c.JSON(http.StatusBadRequest, gin.H{"error": "ad revenue calculation overflow"}); return }
		if creatorAd > in.AdRevenueMicros { creatorAd = in.AdRevenueMicros }
		resp := CostEstimateResponse{
			Input: in,
			ViewerMinutes: in.ViewerMinutes,
			Breakdown: breakdown,
			InfrastructureMicros: breakdown.InfrastructureMicros,
			PlatformFeeMicros: breakdown.PlatformFeeMicros,
			CustomerPriceMicros: breakdown.CustomerPriceMicros,
			AdRevenueMicros: in.AdRevenueMicros,
			CreatorAdRevenueMicros: creatorAd,
			NotellAdRevenueMicros: in.AdRevenueMicros - creatorAd,
			Currency: currency,
		}
		c.JSON(http.StatusOK, resp)
	})
}
