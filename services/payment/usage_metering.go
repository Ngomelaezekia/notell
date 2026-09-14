package main

import (
 "errors"
 "math"
 "time"

 "gorm.io/gorm"
)

type UsageSummary struct {
 Metric string `json:"metric"`
 Used int64 `json:"used"`
 Included int64 `json:"included"`
 Overage int64 `json:"overage"`
 UnitPrice int64 `json:"overage_unit_price"`
 Amount int64 `json:"overage_amount"`
}

func calculateOverage(used, included, unitPrice int64) (int64, int64, bool) {
 if used < 0 || included < 0 || unitPrice < 0 { return 0, 0, false }
 if used <= included { return 0, 0, true }
 over := used - included
 if over > math.MaxInt64/unitPrice && unitPrice != 0 { return 0, 0, false }
 return over, over * unitPrice, true
}

func (s *Server) summarizeSubscriptionUsage(userID, subscriptionID string, start, end time.Time) ([]UsageSummary, error) {
 var sub Subscription
 if err := s.db.Table("subscriptions").Where("id = ? AND user_id = ?", subscriptionID, userID).First(&sub).Error; err != nil { return nil, err }
 var allowances []PackageAllowance
 if err := s.db.Table("package_allowances").Where("package_id = ?", sub.PackageID).Find(&allowances).Error; err != nil { return nil, err }
 out := make([]UsageSummary, 0, len(allowances))
 for _, a := range allowances {
  var total int64
  if err := s.db.Table("usage_events").Where("user_id = ? AND subscription_id = ? AND metric = ? AND occurred_at >= ? AND occurred_at < ?", userID, subscriptionID, a.Metric, start, end).Select("COALESCE(SUM(quantity),0)").Scan(&total).Error; err != nil { return nil, err }
  over, amount, ok := calculateOverage(total, a.IncludedQuantity, a.OverageUnitPrice)
  if !ok { return nil, errors.New("usage overage overflow") }
  out = append(out, UsageSummary{Metric:a.Metric,Used:total,Included:a.IncludedQuantity,Overage:over,UnitPrice:a.OverageUnitPrice,Amount:amount})
 }
 return out, nil
}

func (s *Server) calculateCurrentUsage(userID, subscriptionID string) ([]UsageSummary, error) {
 var sub Subscription
 if err := s.db.Table("subscriptions").Where("id = ? AND user_id = ?", subscriptionID, userID).First(&sub).Error; err != nil { return nil, err }
 return s.summarizeSubscriptionUsage(userID, subscriptionID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
}

func subscriptionExists(err error) bool { return !errors.Is(err, gorm.ErrRecordNotFound) }
