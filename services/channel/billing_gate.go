package main

import (
    "net/http"
    "os"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
)

func channelPlanActive(c *gin.Context) bool {
    if strings.TrimSpace(os.Getenv("APP_ENV")) != "production" && strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_URL")) == "" {
        return true
    }
    base := strings.TrimRight(strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_URL")), "/")
    key := strings.TrimSpace(os.Getenv("INTERNAL_SERVICE_KEY"))
    if base == "" || key == "" {
        return false
    }
    uid := userID(c)
    req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, base+"/v1/internal/entitlements/channel/0", nil)
    if err != nil { return false }
    req.Header.Set("X-Internal-Service-Key", key)
    req.Header.Set("X-User-ID", strconv.FormatUint(uid, 10))
    resp, err := internalHTTPClient.Do(req)
    if err != nil { return false }
    defer resp.Body.Close()
    return resp.StatusCode == http.StatusOK
}

func requireChannelPlan(next gin.HandlerFunc) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !channelPlanActive(c) {
            c.JSON(http.StatusPaymentRequired, gin.H{"message":"active channel plan required","code":"CHANNEL_PLAN_REQUIRED"})
            return
        }
        next(c)
    }
}
