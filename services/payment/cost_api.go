package main

import (
 "net/http"
 "strings"

 "github.com/gin-gonic/gin"
)

type CostEstimateRequest struct {
 StreamMinutes int64 `json:"stream_minutes"`
 ViewerCount int64 `json:"viewer_count"`
 ViewerMinutes int64 `json:"viewer_minutes"`
 TranscodeMinutes int64 `json:"transcode_minutes"`
 StorageGBMonths int64 `json:"storage_gb_months"`
 ComputeMinutes int64 `json:"compute_minutes"`
 CDNMinutes int64 `json:"cdn_minutes"`
 PaymentFeeMicros int64 `json:"payment_fee_micros"`
 OperationalMicros int64 `json:"operational_micros"`
 RiskReserveBps int64 `json:"risk_reserve_bps"`
 PlatformMarginBps int64 `json:"platform_margin_bps"`
 Currency string `json:"currency"`
 AdRevenueMicros int64 `json:"ad_revenue_micros"`
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

type SubscriptionCostResponse struct {
 SubscriptionID string `json:"subscription_id"`
 PackageID string `json:"package_id"`
 Currency string `json:"currency"`
 PackagePriceMicros int64 `json:"package_price_micros"`
 Usage []UsageSummary `json:"usage"`
 OverageMicros int64 `json:"overage_micros"`
 ProjectedPackageChargeMicros int64 `json:"projected_package_charge_micros"`
 EstimatedInfrastructureMicros int64 `json:"estimated_infrastructure_micros"`
 EstimatedPlatformFeeMicros int64 `json:"estimated_platform_fee_micros"`
}

func validateCostRequest(in *CostEstimateRequest) (string, bool) {
 if in.StreamMinutes < 0 || in.ViewerCount < 0 || in.ViewerMinutes < 0 || in.TranscodeMinutes < 0 || in.StorageGBMonths < 0 || in.ComputeMinutes < 0 || in.CDNMinutes < 0 || in.PaymentFeeMicros < 0 || in.OperationalMicros < 0 || in.AdRevenueMicros < 0 { return "cost inputs cannot be negative", false }
 if in.ViewerMinutes == 0 && in.ViewerCount > 0 && in.StreamMinutes > 0 { v,ok:=mulSafe(in.ViewerCount,in.StreamMinutes); if !ok { return "viewer minute calculation overflow",false }; in.ViewerMinutes=v }
 if in.RiskReserveBps < 0 || in.RiskReserveBps > 10000 || in.PlatformMarginBps < 0 || in.PlatformMarginBps > 10000 || in.CreatorAdShareBps < 0 || in.CreatorAdShareBps > 10000 { return "basis-point values must be between 0 and 10000",false }
 currency:=strings.ToLower(strings.TrimSpace(in.Currency)); if currency=="" { currency="usd" }; if len(currency)!=3 { return "currency must be 3 characters",false }; in.Currency=currency
 return "",true
}

func costInputFromUsage(usage []UsageSummary) CostInput {
 var in CostInput
 for _,u:=range usage { switch u.Metric { case "stream_minutes": in.StreamMinutes=u.Used; case "viewer_minutes": in.ViewerMinutes=u.Used; case "transcode_minutes": in.TranscodeMinutes=u.Used; case "storage_gb_months": in.StorageGBMonths=u.Used; case "compute_minutes": in.ComputeMinutes=u.Used; case "cdn_minutes": in.CDNMinutes=u.Used } }
 return in
}

func registerCostRoutes(r *gin.Engine, s *Server) {
 api:=r.Group("/v1"); api.Use(s.auth)
 api.POST("/cost/estimate",func(c *gin.Context){
  var in CostEstimateRequest; if err:=c.ShouldBindJSON(&in); err!=nil { c.JSON(http.StatusBadRequest,gin.H{"error":"invalid cost estimate request"}); return }
  if msg,ok:=validateCostRequest(&in); !ok { c.JSON(http.StatusBadRequest,gin.H{"error":msg}); return }
  if in.RiskReserveBps==0 { in.RiskReserveBps=s.cfg.DefaultRiskReserveBps }; if in.PlatformMarginBps==0 { in.PlatformMarginBps=s.cfg.DefaultPlatformMarginBps }
  rates:=s.cfg.CostRates
  costInput:=CostInput{StreamMinutes:in.StreamMinutes,ViewerMinutes:in.ViewerMinutes,TranscodeMinutes:in.TranscodeMinutes,StorageGBMonths:in.StorageGBMonths,ComputeMinutes:in.ComputeMinutes,CDNMinutes:in.CDNMinutes,PaymentFeeMicros:in.PaymentFeeMicros,OperationalMicros:in.OperationalMicros,RiskReserveBps:in.RiskReserveBps,PlatformMarginBps:in.PlatformMarginBps}
  breakdown,ok:=CalculateCost(costInput,rates); if !ok { c.JSON(http.StatusBadRequest,gin.H{"error":"cost calculation overflow"}); return }
  creatorAd,ok:=bpsCeil(in.AdRevenueMicros,in.CreatorAdShareBps); if !ok { c.JSON(http.StatusBadRequest,gin.H{"error":"ad revenue calculation overflow"}); return }; if creatorAd>in.AdRevenueMicros { creatorAd=in.AdRevenueMicros }
  c.JSON(http.StatusOK,CostEstimateResponse{Input:in,ViewerMinutes:in.ViewerMinutes,Breakdown:breakdown,InfrastructureMicros:breakdown.InfrastructureMicros,PlatformFeeMicros:breakdown.PlatformFeeMicros,CustomerPriceMicros:breakdown.CustomerPriceMicros,AdRevenueMicros:in.AdRevenueMicros,CreatorAdRevenueMicros:creatorAd,NotellAdRevenueMicros:in.AdRevenueMicros-creatorAd,Currency:in.Currency})
 })
 api.GET("/subscriptions/:id/cost",func(c *gin.Context){
  uid:=c.GetString("user_id"); var sub Subscription
  if err:=s.db.Table("subscriptions").Where("id = ? AND user_id = ?",c.Param("id"),uid).First(&sub).Error; err!=nil { c.JSON(http.StatusNotFound,gin.H{"error":"subscription not found"}); return }
  var p Package; if err:=s.db.Table("packages").Where("id = ?",sub.PackageID).First(&p).Error; err!=nil { c.JSON(http.StatusNotFound,gin.H{"error":"package not found"}); return }
  usage,err:=s.calculateCurrentUsage(uid,sub.ID); if err!=nil { c.JSON(http.StatusInternalServerError,gin.H{"error":"failed to calculate subscription usage"}); return }
  var overage int64; for _,u:=range usage { if u.Amount>0 && overage>int64(^uint64(0)>>1)-u.Amount { c.JSON(http.StatusInternalServerError,gin.H{"error":"cost total overflow"}); return }; overage+=u.Amount }
  packageCharge,ok:=addSafe(p.MonthlyPrice*1000000,overage); if p.MonthlyPrice>0 && !ok { c.JSON(http.StatusInternalServerError,gin.H{"error":"package charge overflow"}); return }
  ci:=costInputFromUsage(usage); ci.RiskReserveBps=s.cfg.DefaultRiskReserveBps; ci.PlatformMarginBps=s.cfg.DefaultPlatformMarginBps
  breakdown,ok:=CalculateCost(ci,s.cfg.CostRates); if !ok { c.JSON(http.StatusInternalServerError,gin.H{"error":"cost calculation overflow"}); return }
  c.JSON(http.StatusOK,SubscriptionCostResponse{SubscriptionID:sub.ID,PackageID:p.ID,Currency:p.Currency,PackagePriceMicros:p.MonthlyPrice*1000000,Usage:usage,OverageMicros:overage,ProjectedPackageChargeMicros:packageCharge,EstimatedInfrastructureMicros:breakdown.InfrastructureMicros,EstimatedPlatformFeeMicros:breakdown.PlatformFeeMicros})
 })
}
