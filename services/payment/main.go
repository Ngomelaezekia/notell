package main

import (
 "errors"
 "log"
 "net/http"
 "os"
 "strings"
 "time"

 "github.com/gin-contrib/cors"
 "github.com/gin-gonic/gin"
 "github.com/golang-jwt/jwt/v5"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
)

type Config struct { Port string; DatabaseURL string; JWTSecret string; FrontendURL string; Env string }
func loadConfig() Config { return Config{Port:getenv("PORT","10000"), DatabaseURL:os.Getenv("DATABASE_URL"), JWTSecret:os.Getenv("JWT_SECRET"), FrontendURL:getenv("FRONTEND_URL","http://localhost:5173"), Env:getenv("APP_ENV","development")} }
func getenv(k,d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }

type Payment struct {
 ID string `gorm:"primaryKey"`; UserID string `gorm:"index;not null"`; CustomerID string; PriceID string
 Amount int64 `gorm:"not null"`; Currency string `gorm:"not null"`; Status string `gorm:"index;not null"`
 Provider string `gorm:"not null"`; ProviderPaymentID string `gorm:"index"`; IdempotencyKey string `gorm:"uniqueIndex;not null"`
 FailureCode string; FailureMessage string; CreatedAt time.Time; UpdatedAt time.Time
}

type PaymentCustomer struct { ID string `gorm:"primaryKey"`; UserID string `gorm:"uniqueIndex;not null"`; Provider string; ProviderCustomerID string; CreatedAt time.Time; UpdatedAt time.Time }
type PaymentProduct struct { ID string `gorm:"primaryKey"`; Name string; Description string; Active bool; Provider string; ProviderProductID string; CreatedAt time.Time; UpdatedAt time.Time }
type PaymentPrice struct { ID string `gorm:"primaryKey"`; ProductID string `gorm:"index;not null"`; Currency string; UnitAmount int64; Interval string; Active bool; ProviderPriceID string; CreatedAt time.Time; UpdatedAt time.Time }
type WebhookEvent struct { ID string `gorm:"primaryKey"`; Provider string; ProviderEventID string; EventType string; Payload string `gorm:"type:jsonb"`; Status string; ErrorMessage string; ProcessedAt *time.Time; CreatedAt time.Time }
type Refund struct { ID string `gorm:"primaryKey"`; PaymentID string `gorm:"index;not null"`; Amount int64; Currency string; Status string; ProviderRefundID string; Reason string; CreatedAt time.Time; UpdatedAt time.Time }
type LedgerEntry struct { ID string `gorm:"primaryKey"`; PaymentID string `gorm:"index"`; UserID string `gorm:"index;not null"`; EntryType string; Amount int64; Currency string; Reference string; CreatedAt time.Time }
type AuditEvent struct { ID string `gorm:"primaryKey"`; PaymentID string `gorm:"index"`; UserID string; EventType string; FromStatus string; ToStatus string; Metadata string `gorm:"type:jsonb"`; CreatedAt time.Time }

type Server struct { cfg Config; db *gorm.DB }
func (s *Server) auth(c *gin.Context) { token:=c.GetHeader("Authorization"); if strings.HasPrefix(token,"Bearer ") { token=strings.TrimSpace(strings.TrimPrefix(token,"Bearer ")) } else { token,_=c.Cookie("auth_token") }; if token=="" { c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }; t,err:=jwt.Parse(token,func(t *jwt.Token)(any,error){ if t.Method.Alg()!=jwt.SigningMethodHS256.Alg() { return nil,errors.New("invalid signing method") }; return []byte(s.cfg.JWTSecret),nil },jwt.WithIssuer("notell-api")); if err!=nil || !t.Valid { c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }; sub,err:=t.Claims.GetSubject(); if err!=nil || sub=="" { c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }; c.Set("user_id",sub); c.Next() }
func main() { cfg:=loadConfig(); if cfg.JWTSecret=="" { log.Fatal("JWT_SECRET is required") }; if cfg.DatabaseURL=="" { log.Fatal("DATABASE_URL is required") }; db,err:=gorm.Open(postgres.Open(cfg.DatabaseURL),&gorm.Config{}); if err!=nil { log.Fatal(err) }; s:=&Server{cfg:cfg,db:db}; r:=gin.New(); r.Use(gin.Recovery()); r.Use(cors.New(cors.Config{AllowOrigins:[]string{cfg.FrontendURL},AllowMethods:[]string{"GET","POST","OPTIONS"},AllowHeaders:[]string{"Origin","Content-Type","Authorization","Idempotency-Key"},AllowCredentials:true})); r.GET("/health",func(c *gin.Context){ c.JSON(200,gin.H{"status":"ok","service":"payment"}) }); r.GET("/ready",func(c *gin.Context){ sqlDB,e:=db.DB(); if e!=nil || sqlDB.Ping()!=nil { c.JSON(503,gin.H{"status":"not_ready"}); return }; c.JSON(200,gin.H{"status":"ready"}) }); api:=r.Group("/v1"); api.Use(s.auth); api.GET("/payments",func(c *gin.Context){ uid:=c.GetString("user_id"); var p []Payment; if err:=db.Where("user_id = ?",uid).Order("created_at desc").Limit(100).Find(&p).Error; err!=nil { c.JSON(500,gin.H{"error":"failed to load payments"}); return }; c.JSON(200,gin.H{"payments":p}) }); r.Run(":"+cfg.Port) }
