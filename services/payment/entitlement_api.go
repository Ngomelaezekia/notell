package main

import (
 "net/http"
 "strings"

 "github.com/gin-gonic/gin"
)

// Entitlement API exposes feature access to independent Notell services.
// Channel, live and message services consume this contract.

func registerEntitlementRoutes(r *gin.Engine, s *Server) {
 api := r.Group("/v1")
 api.Use(s.auth)

 api.GET("/entitlements", func(c *gin.Context) {
  userID := c.GetString("user_id")
  var entitlements []Entitlement
  if err := s.db.Table("entitlements").Where("user_id = ? AND active = true", userID).Find(&entitlements).Error; err != nil {
   c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to load entitlements"})
   return
  }
  c.JSON(http.StatusOK, gin.H{"entitlements": entitlements})
 })

 api.GET("/entitlements/check/:resource/:id", func(c *gin.Context) {
  userID := c.GetString("user_id")
  resource := strings.TrimSpace(c.Param("resource"))
  resourceID := strings.TrimSpace(c.Param("id"))

  var entitlement Entitlement
  err := s.db.Table("entitlements").Where(
   "user_id = ? AND resource_type = ? AND resource_id = ? AND active = true",
   userID, resource, resourceID,
  ).First(&entitlement).Error

  if err != nil {
   c.JSON(http.StatusOK, gin.H{"allowed": false})
   return
  }

  c.JSON(http.StatusOK, gin.H{
   "allowed": true,
   "entitlement": entitlement,
  })
 })
}
