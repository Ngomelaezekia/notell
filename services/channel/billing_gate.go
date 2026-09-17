package main

import (
 "encoding/json"
 "net/http"
 "os"
 "strconv"
 "strings"
 "time"
 "github.com/gin-gonic/gin"
)

type paymentEntitlementResponse struct {
 Active bool `json:"active"`
 Subscription struct { PackageID string `json:"packageId"` } `json:"subscription"`
}

func channelPlanCodeFromPackage(packageID string) string {
 switch strings.TrimSpace(packageID) {
 case "pkg_starter": return "STARTER"
 case "pkg_creator": return "CREATOR"
 case "pkg_pro": return "STUDIO"
 case "pkg_247": return "PRO_NETWORK"
 default: return ""
 }
}

func channelPlanActive(c *gin.Context) (bool, string) {
 if strings.TrimSpace(os.Getenv("APP_ENV")) != "production" && strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_URL")) == "" { return true, "CREATOR" }
 base := strings.TrimRight(strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_URL")), "/")
 key := strings.TrimSpace(os.Getenv("INTERNAL_SERVICE_KEY"))
 if base == "" || key == "" { return false, "" }
 uid := userID(c)
 req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, base+"/v1/internal/entitlements/platform/channel", nil)
 if err != nil { return false, "" }
 req.Header.Set("X-Internal-Service-Key", key)
 req.Header.Set("X-User-ID", strconv.FormatUint(uid,10))
 client := &http.Client{Timeout:5*time.Second}
 resp, err := client.Do(req)
 if err != nil { return false, "" }
 defer resp.Body.Close()
 if resp.StatusCode != http.StatusOK { return false, "" }
 var result paymentEntitlementResponse
 if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Active { return false, "" }
 planCode := channelPlanCodeFromPackage(result.Subscription.PackageID)
 if planCode == "" { return false, "" }
 return true, planCode
}

func requireChannelPlan(next gin.HandlerFunc) gin.HandlerFunc {
 return func(c *gin.Context) {
  active, planCode := channelPlanActive(c)
  if !active { c.JSON(http.StatusPaymentRequired, gin.H{"message":"active channel plan required","code":"CHANNEL_PLAN_REQUIRED"}); return }
  c.Set("channelPlanCode", planCode)
  next(c)
 }
}
