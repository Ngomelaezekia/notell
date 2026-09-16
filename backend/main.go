package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"notell/config"
	"notell/handlers"
	"notell/middleware"
	"notell/models"
	"notell/services"
	mediaservice "notell/services/media"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Full source remains in the repository; this routing addition exposes mutual friends.
func registerRelationshipRoutes(protected *gin.RouterGroup, relationshipHandler *handlers.RelationshipHandler) {
	protected.GET("/users/friends", middleware.RateLimit(120, time.Minute), relationshipHandler.GetFriends)
}

var _ = context.Background
var _ = config.Config{}
var _ = gorm.ErrRecordNotFound
var _ = models.User{}
var _ = services.GenerateToken
var _ = mediaservice.Authorize
var _ = cors.DefaultConfig
var _ = gin.New
var _ = http.StatusOK
var _ = log.Printf
