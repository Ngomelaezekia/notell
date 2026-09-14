package main

import (
    "encoding/json"
    "net/http"
    "strings"
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Subscription struct {
    ID string `gorm:"primaryKey" json:"id"`
    UserID string `gorm:"index;not null" json:"userId"`
    ResourceType string `gorm:"not null" json:"resourceType"`
    ResourceID string `gorm:"not null" json:"resourceId"`
    PriceID string `gorm:"not null" json:"priceId"`
    Provider string `json:"provider"`
    ProviderSubscriptionID string `json:"providerSubscriptionId,omitempty"`
    Status string `gorm:"index;not null" json:"status"`
    CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
    CurrentPeriodEnd *time.Time `json:"currentPeriodEnd,omitempty"`
    CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

func registerSubscriptionRoutes(r *gin.Engine, s *Server) {
    api:=r.Group("/v1"); api.Use(s.auth)
    api.GET("/subscriptions", func(c *gin.Context){ uid:=c.GetString("user_id");var rows []Subscription;if err:=s.db.Where("user_id = ?",uid).Order("created_at desc").Find(&rows).Error;err!=nil{c.JSON(500,gin.H{"error":"failed to load subscriptions"});return};c.JSON(200,gin.H{"subscriptions":rows}) })
    api.GET("/entitlements/:resourceType/:resourceId", func(c *gin.Context){ uid:=c.GetString("user_id");var row Subscription;err:=s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?",uid,c.Param("resourceType"),c.Param("resourceId")).Order("updated_at desc").First(&row).Error;if err!=nil{c.JSON(200,gin.H{"active":false});return};active:=row.Status=="active"||row.Status=="trialing";if row.CurrentPeriodEnd!=nil&&row.CurrentPeriodEnd.Before(time.Now().UTC()){active=false};c.JSON(200,gin.H{"active":active,"subscription":row}) })
    api.POST("/subscriptions", func(c *gin.Context){
        var in struct{ResourceType string `json:"resourceType" binding:"required"`;ResourceID string `json:"resourceId" binding:"required"`;PriceID string `json:"priceId" binding:"required"`};if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"resourceType, resourceId and priceId are required"});return};in.ResourceType=strings.TrimSpace(in.ResourceType);in.ResourceID=strings.TrimSpace(in.ResourceID);in.PriceID=strings.TrimSpace(in.PriceID);if len(in.ResourceType)>64||len(in.ResourceID)>128||len(in.PriceID)>128{c.JSON(400,gin.H{"error":"subscription identifiers are too long"});return};uid:=c.GetString("user_id");var existing Subscription;if err:=s.db.Where("user_id = ? AND resource_type = ? AND resource_id = ?",uid,in.ResourceType,in.ResourceID).First(&existing).Error;err==nil{c.JSON(200,gin.H{"subscription":existing,"idempotent":true});return};row:=Subscription{ID:newID("sub"),UserID:uid,ResourceType:in.ResourceType,ResourceID:in.ResourceID,PriceID:in.PriceID,Provider:"stripe",Status:"incomplete"};if err:=s.db.Create(&row).Error;err!=nil{c.JSON(409,gin.H{"error":"subscription could not be created"});return};c.JSON(201,gin.H{"subscription":row,"next":"checkout"})
    })
    _ = json.Valid
    _ = http.MethodGet
    _ = gorm.ErrRecordNotFound
}
