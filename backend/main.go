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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const maxJSONBodyBytes int64 = 2 << 20

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func trustedProxies() []string {
	value := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		if proxy := strings.TrimSpace(part); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func serveMedia(storage services.MediaStorage, claimed func(context.Context, string) (bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := filepath.Base(c.Param("filename"))
		if filename == "." || filename == "" || filename != c.Param("filename") {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid media filename"})
			return
		}
		isClaimed, err := claimed(c.Request.Context(), filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed checking media"})
			return
		}
		if !isClaimed {
			c.JSON(http.StatusNotFound, gin.H{"message": "media not found"})
			return
		}

		byteRange := c.GetHeader("Range")
		body, contentType, contentLength, contentRange, err := storage.Open(c.Request.Context(), services.MediaObjectKey(filename), byteRange)
		if err != nil {
			if byteRange != "" {
				c.Header("Content-Range", "bytes */*")
				c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"message": "invalid media range"})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"message": "media not found"})
			return
		}
		defer body.Close()

		c.Header("Accept-Ranges", "bytes")
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		if contentType != "" {
			c.Header("Content-Type", contentType)
		}
		if contentLength > 0 {
			c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		}
		if contentRange != "" {
			c.Header("Content-Range", contentRange)
			c.Status(http.StatusPartialContent)
		}
		if _, err := io.Copy(c.Writer, body); err != nil {
			log.Printf("failed streaming media %q: %v", services.MediaObjectKey(filename), err)
		}
	}
}

func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}
	log.Println("Database connection established successfully")
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to access PostgreSQL connection pool: %v", err)
	}
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 25))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(envInt("DB_MAX_IDLE_MINUTES", 5)) * time.Minute)
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Relationship{}, &models.Channel{}, &models.Notification{}, &models.Upload{}, &models.MediaMetadata{}, &models.MediaJob{}, &models.PostView{}); err != nil {
		log.Fatalf("Database auto-migration failed: %v", err)
	}

	mediaStorage, err := services.NewMediaStorage(cfg)
	if err != nil {
		log.Fatalf("Media storage initialization failed: %v", err)
	}
	storageCtx, storageCancel := context.WithCancel(context.Background())
	defer storageCancel()
	services.StartMediaReconciler(storageCtx, mediaStorage, db)
	services.StartMediaWorker(storageCtx, db)

	r := gin.Default()
	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		log.Fatalf("Invalid TRUSTED_PROXIES configuration: %v", err)
	}
	r.Use(securityHeaders(), middleware.MaxBodyBytes(maxJSONBodyBytes), cors.New(cors.Config{
		AllowOrigins: []string{cfg.FrontendURL}, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"}, ExposeHeaders: []string{"Content-Length"}, AllowCredentials: true, MaxAge: 12 * time.Hour,
	}))
	if cfg.AppEnv == "production" {
		r.Use(func(c *gin.Context) {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Next()
		})
	}

	r.GET("/uploads/:filename", serveMedia(mediaStorage, func(ctx context.Context, filename string) (bool, error) {
		var count int64
		err := db.Model(&models.Upload{}).
			Where("filename = ? AND post_id IS NOT NULL", filename).
			Count(&count).Error
		return count > 0, err
	}))
	auth := handlers.NewAuthHandler(db, cfg)
	post := handlers.NewPostHandler(db, cfg.PublicURL, cfg.MediaPublicURL)
	userHandler := handlers.NewUserHandler(db)
	relationshipHandler := handlers.NewRelationshipHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	uploadHandler := handlers.NewUploadHandler(db, cfg.PublicURL, mediaStorage)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
		api.GET("/ready", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		})
		api.POST("/auth/register", middleware.RateLimit(10, time.Minute), auth.Register)
		api.POST("/auth/login", middleware.RateLimit(10, time.Minute), auth.Login)
		api.GET("/auth/google", middleware.RateLimit(20, time.Minute), auth.GoogleLogin)
		api.GET("/auth/google/callback", middleware.RateLimit(20, time.Minute), auth.GoogleCallback)
		api.GET("/posts/:id", middleware.RateLimit(120, time.Minute), post.GetPostByID)
		api.GET("/posts/:id/comments", middleware.RateLimit(120, time.Minute), post.GetComments)
		api.GET("/users/:id", middleware.RateLimit(120, time.Minute), userHandler.GetUserProfile)

		protected := api.Group("/").Use(middleware.CSRFProtection(cfg.FrontendURL), middleware.AuthRequired(cfg.JWTSecret))
		{
			protected.POST("/upload", middleware.RateLimit(20, time.Minute), uploadHandler.UploadMedia)
			protected.GET("/auth/me", middleware.RateLimit(120, time.Minute), auth.Me)
			protected.POST("/auth/logout", middleware.RateLimit(30, time.Minute), auth.Logout)
			protected.PUT("/users/profile", middleware.RateLimit(30, time.Minute), userHandler.UpdateProfile)
			protected.GET("/users/search", middleware.RateLimit(60, time.Minute), userHandler.SearchUsers)
			protected.GET("/posts/search", middleware.RateLimit(60, time.Minute), post.SearchPosts)
			protected.POST("/posts", middleware.RateLimit(30, time.Minute), post.CreatePost)
			protected.GET("/posts/feed", middleware.RateLimit(120, time.Minute), post.GetCategorizedFeed)
			protected.DELETE("/posts/:id", middleware.RateLimit(30, time.Minute), post.DeletePost)
			protected.POST("/posts/:id/view", middleware.RateLimit(240, time.Minute), post.RecordPostView)
			protected.POST("/posts/:id/like", middleware.RateLimit(120, time.Minute), post.ToggleLike)
			protected.POST("/posts/:id/comments", middleware.RateLimit(60, time.Minute), post.AddComment)
			protected.POST("/users/:id/follow", middleware.RateLimit(60, time.Minute), relationshipHandler.FollowUser)
			protected.DELETE("/users/:id/unfollow", middleware.RateLimit(60, time.Minute), relationshipHandler.UnfollowUser)
			protected.GET("/users/:id/relationship", middleware.RateLimit(120, time.Minute), relationshipHandler.GetRelationshipStatus)
			protected.DELETE("/users/followers/:id", middleware.RateLimit(60, time.Minute), relationshipHandler.RemoveFollower)
			protected.GET("/users/:id/followers", middleware.RateLimit(120, time.Minute), relationshipHandler.GetFollowers)
			protected.GET("/users/:id/following", middleware.RateLimit(120, time.Minute), relationshipHandler.GetFollowing)
			protected.GET("/notifications", middleware.RateLimit(120, time.Minute), notificationHandler.List)
			protected.POST("/notifications/:id/read", middleware.RateLimit(120, time.Minute), notificationHandler.MarkRead)
			protected.POST("/notifications/read-all", middleware.RateLimit(60, time.Minute), notificationHandler.MarkAllRead)
		}
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
