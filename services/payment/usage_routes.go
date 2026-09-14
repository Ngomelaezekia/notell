package main

import (
 "net/http"
 "strings"

 "github.com/gin-gonic/gin"
)

func registerUsageRoutes(r *gin.Engine, s *Server) {
 api := r.Group("/v1")
 api.Use(s.auth)
 api.GET("/subscriptions/:id/usage", func(c *gin.Context) {
  uid := c.GetString("user_id")
  id := strings.TrimSpace(c.Param("id"))
  if id == "" { c.JSON(http.StatusBadRequest, gin.H{"error":"subscription id is required"}); return }
  summaries, err := s.calculateCurrentUsage(uid, id)
  if err != nil { c.JSON(http.StatusNotFound, gin.H{"error":"subscription not found"}); return }
  var total int64
  for _, v := range summaries { if v.Amount > 0 { if total > int64(^uint64(0)>>1)-v.Amount { c.JSON(500,gin.H{"error":"usage total overflow"}); return }; total += v.Amount } }
  c.JSON(http.StatusOK, gin.H{"subscription_id":id,"usage":summaries,"overage_total":total})
 })
}
