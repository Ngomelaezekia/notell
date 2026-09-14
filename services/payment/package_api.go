package main

import (
 "net/http"
 "strings"
 "time"
 "github.com/gin-gonic/gin"
)

type Package struct { ID string `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Currency string `json:"currency"`; MonthlyPrice int64 `json:"monthly_price"`; Active bool `json:"active"` }
type PackageAllowance struct { ID string `json:"id"`; PackageID string `json:"package_id"`; Metric string `json:"metric"`; IncludedQuantity int64 `json:"included_quantity"`; OverageUnitPrice int64 `json:"overage_unit_price"` }
type Subscription struct { ID string `json:"id"`; UserID string `json:"user_id"`; PackageID string `json:"package_id"`; Status SubscriptionStatus `json:"status"`; Provider string `json:"provider"`; CurrentPeriodStart time.Time `json:"current_period_start"`; CurrentPeriodEnd time.Time `json:"current_period_end"`; CancelAtPeriodEnd bool `json:"cancel_at_period_end"` }
type UsageEvent struct { ID string `json:"id"`; UserID string `json:"user_id"`; SubscriptionID string `json:"subscription_id"`; Metric string `json:"metric"`; Quantity int64 `json:"quantity"`; IdempotencyKey string `json:"idempotency_key"`; OccurredAt time.Time `json:"occurred_at"` }

func registerPackageRoutes(r *gin.Engine, s *Server) {
 api:=r.Group("/v1"); api.Use(s.auth)
 api.GET("/packages",func(c *gin.Context){ var p []Package; if err:=s.db.Table("packages").Where("active = true").Order("monthly_price asc").Find(&p).Error; err!=nil { c.JSON(500,gin.H{"error":"failed to load packages"}); return }; c.JSON(200,gin.H{"packages":p}) })
 api.GET("/packages/:id/allowances",func(c *gin.Context){ var a []PackageAllowance; if err:=s.db.Table("package_allowances").Where("package_id = ?",c.Param("id")).Find(&a).Error; err!=nil { c.JSON(500,gin.H{"error":"failed to load allowances"}); return }; c.JSON(200,gin.H{"allowances":a}) })
 api.GET("/subscriptions",func(c *gin.Context){ var sub []Subscription; if err:=s.db.Table("subscriptions").Where("user_id = ?",c.GetString("user_id")).Order("created_at desc").Find(&sub).Error; err!=nil { c.JSON(500,gin.H{"error":"failed to load subscriptions"}); return }; c.JSON(200,gin.H{"subscriptions":sub}) })
 api.POST("/subscriptions",func(c *gin.Context){ var in struct{ PackageID string `json:"package_id"` }; if err:=c.ShouldBindJSON(&in); err!=nil || strings.TrimSpace(in.PackageID)=="" { c.JSON(400,gin.H{"error":"package_id is required"}); return }; var p Package; if err:=s.db.Table("packages").Where("id = ? AND active = true",in.PackageID).First(&p).Error; err!=nil { c.JSON(404,gin.H{"error":"package not found"}); return }; uid:=c.GetString("user_id"); now:=time.Now().UTC(); sub:=Subscription{ID:newID("sub"),UserID:uid,PackageID:p.ID,Status:SubscriptionIncomplete,Provider:"stripe",CurrentPeriodStart:now,CurrentPeriodEnd:now.AddDate(0,1,0)}; if err:=s.db.Table("subscriptions").Create(&sub).Error; err!=nil { c.JSON(500,gin.H{"error":"subscription could not be created"}); return }; c.JSON(http.StatusCreated,gin.H{"subscription":sub,"payment_required":p.MonthlyPrice>0}) })
 api.POST("/usage",func(c *gin.Context){ key:=strings.TrimSpace(c.GetHeader("Idempotency-Key")); if key=="" || len(key)>128 { c.JSON(400,gin.H{"error":"Idempotency-Key is required"}); return }; var in struct{ SubscriptionID string `json:"subscription_id"`; Metric string `json:"metric"`; Quantity int64 `json:"quantity"`; OccurredAt *time.Time `json:"occurred_at"` }; if err:=c.ShouldBindJSON(&in); err!=nil || strings.TrimSpace(in.Metric)=="" || in.Quantity<0 { c.JSON(400,gin.H{"error":"invalid usage event"}); return }; uid:=c.GetString("user_id"); var existing UsageEvent; if err:=s.db.Table("usage_events").Where("idempotency_key = ?",key).First(&existing).Error; err==nil { if existing.UserID!=uid { c.JSON(409,gin.H{"error":"idempotency key belongs to another user"}); return }; c.JSON(200,gin.H{"usage":existing,"idempotent":true}); return }; at:=time.Now().UTC(); if in.OccurredAt!=nil { at=in.OccurredAt.UTC() }; ev:=UsageEvent{ID:newID("use"),UserID:uid,SubscriptionID:strings.TrimSpace(in.SubscriptionID),Metric:strings.TrimSpace(in.Metric),Quantity:in.Quantity,IdempotencyKey:key,OccurredAt:at}; if err:=s.db.Table("usage_events").Create(&ev).Error; err!=nil { c.JSON(409,gin.H{"error":"usage event could not be recorded"}); return }; c.JSON(http.StatusCreated,gin.H{"usage":ev}) })
}
