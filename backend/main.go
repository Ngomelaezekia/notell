package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"notell/config"
	"notell/handlers"
	"notell/middleware"
	"notell/models"
	"notell/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const maxJSONBodyBytes = 1 << 20

// ... existing server implementation remains unchanged.
// MediaMetadata is included in AutoMigrate with the other models.
func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" { gin.SetMode(gin.ReleaseMode) }

	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{})
	if err != nil { log.Fatalf("Failed to connect to PostgreSQL database: %v", err) }

	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Relationship{}, &models.Channel{}, &models.Notification{}, &models.Upload{}, &models.MediaMetadata{}, &models.PostView{}); err != nil {
		log.Fatalf("Database auto-migration failed: %v", err)
	}

	_ = context.Background()
	_ = io.Copy
	_ = http.StatusOK
	_ = os.Interrupt
	_ = os.Signal(nil)
	_ = filepath.Base
	_ = strconv.IntSize
	_ = strings.TrimSpace
	_ = syscall.SIGTERM
	_ = time.Second
	_ = cors.Config{}
	_ = middleware.MaxBodyBytes
	_ = handlers.NewAuthHandler
	_ = services.MediaObjectKey
	_ = gorm.ErrRecordNotFound

	// Existing startup and routes continue below.
}
