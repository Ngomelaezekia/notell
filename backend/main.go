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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"notell/config"
	"notell/handlers"
	"notell/middleware"
	"notell/models"
	"notell/services"
)

const maxJSONBodyBytes int64 = 2 << 20
const privatePlaybackURLExpiry = 5 * time.Minute

type playbackSigner interface {
	GeneratePlaybackURL(context.Context, string, time.Duration) (string, error)
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" { return fallback }
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 { return fallback }
	return n
}

func trustedProxies() []string {
	value := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if value == "" { return nil }
	parts := strings.Split(value, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts { if proxy := strings.TrimSpace(part); proxy != "" { proxies = append(proxies, proxy) } }
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

func serveMedia(storage services.MediaStorage, access func(context.Context, string, uint) (bool, bool, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawFilename := c.Param("filename")
		filename := filepath.Base(rawFilename)
		if filename == "." || filename == "" || filename != rawFilename {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid media filename"})
			return
		}
		userID := uint(0)
		if value, ok := c.Get(middleware.ContextUserIDKey); ok { if id, ok := value.(uint); ok { userID = id } }
		allowed, private, err := access(c.Request.Context(), filename, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed checking media"})
			return
		}
		if !allowed {
			c.JSON(http.StatusNotFound, gin.H{"message": "media not found"})
			return
		}

		key := services.MediaObjectKey(filename)
		exists, err := storage.Exists(c.Request.Context(), key)
		if err != nil {
			log.Printf("failed checking media availability %q: %v", key, err)
			c.JSON(http.StatusBadGateway, gin.H{"message": "media storage unavailable"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"message": "media not found"})
			return
		}

		if private {
			signer, ok := storage.(playbackSigner)
			if !ok {
				log.Printf("private media playback signer unavailable for %q", key)
				c.JSON(http.StatusServiceUnavailable, gin.H{"message": "private media playback is unavailable"})
				return
			}
			playbackURL, signErr := signer.GeneratePlaybackURL(c.Request.Context(), key, privatePlaybackURLExpiry)
			if signErr != nil {
				log.Printf("failed generating private media playback URL %q: %v", key, signErr)
				c.JSON(http.StatusBadGateway, gin.H{"message": "failed generating media playback URL"})
				return
			}
			c.Header("Cache-Control", "private, no-store")
			c.Redirect(http.StatusFound, playbackURL)
			return
		}

		byteRange := c.GetHeader("Range")
		body, contentType, contentLength, contentRange, err := storage.Open(c.Request.Context(), key, byteRange)
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
		if private { c.Header("Cache-Control", "private, no-store") } else { c.Header("Cache-Control", "public, max-age=31536000, immutable") }
		if contentType != "" { c.Header("Content-Type", contentType) }
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		if contentRange != "" { c.Header("Content-Range", contentRange); c.Status(http.StatusPartialContent) }
		if _, err := io.Copy(c.Writer, body); err != nil { log.Printf("failed streaming media %q: %v", key, err) }
	}
}

func main() {
	cfg := config.Load()
	if cfg.AppEnv == "production" { gin.SetMode(gin.ReleaseMode) }
	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{})
	if err != nil { log.Fatalf("Failed to connect to PostgreSQL database: %v", err) }
	log.Println("Database connection established successfully")
	sqlDB, err := db.DB()
	if err != nil { log.Fatalf("Failed to access PostgreSQL connection pool: %v", err) }
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 25))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(envInt("DB_CONN_MAX_IDLE_MINUTES", 5)) * time.Minute)
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil { log.Fatalf("Failed to ping PostgreSQL database: %v", err) }

	// Older deployments could retain view rows for posts that were already deleted.
	// Remove those orphans before GORM adds the post_views -> posts foreign key so
	// auto-migration remains safe and repeatable on existing production databases.
	if result := db.Exec(`DELETE FROM post_views WHERE NOT EXISTS (SELECT 1 FROM posts WHERE posts.id = post_views.post_id)`); result.Error != nil {
		log.Fatalf("Failed cleaning orphan post views: %v", result.Error)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Relationship{}, &models.Channel{}, &models.Notification{}, &models.Upload{}, &models.MediaMetadata{}, &models.MediaJob{}, &models.PostView{}); err != nil { log.Fatalf("Database auto-migration failed: %v", err) }
	mediaStorage, err := services.NewMediaStorage(cfg)
	if err != nil { log.Fatalf("Media storage initialization failed: %v", err) }
	storageCtx, storageCancel := context.WithCancel(context.Background())
	defer storageCancel()
	services.StartMediaReconciler(storageCtx, mediaStorage, db)
	services.StartMediaWorker(storageCtx, db)
	r := gin.Default()
	if err := r.SetTrustedProxies(trustedProxies()); err != nil { log.Fatalf("Invalid TRUSTED_PROXIES configuration: %v", err) }
	r.Use(securityHeaders(), middleware.MaxBodyBytes(maxJSONBodyBytes), cors.New(cors.Config{AllowOrigins: []string{cfg.FrontendURL}, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"}, ExposeHeaders: []string{"Content-Length", "Content-Range", "Accept-Ranges"}, AllowCredentials: true, MaxAge: 12 * time.Hour}))
	if cfg.AppEnv == "production" { r.Use(func(c *gin.Context) { c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains"); c.Next() }) }
	r.GET("/uploads/:filename", middleware.OptionalAuth(cfg.JWTSecret), serveMedia(mediaStorage, func(ctx context.Context, filename string, userID uint) (bool, bool, error) {
		var post models.Post
		err := db.Select("posts.user_id, posts.visibility").Joins("JOIN uploads ON uploads.post_id = posts.id").Where("uploads.filename = ? AND uploads.post_id IS NOT NULL", filename).First(&post).Error
		if err != nil { if err == gorm.ErrRecordNotFound { return false, false, nil }; return false, false, err }
		private := post.Visibility == "private"
		if private && (userID == 0 || userID != post.UserID) { return false, true, nil }
		return true, private, nil
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
			if err := sqlDB.PingContext(ctx); err != nil { c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"}); return }
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		})
		api.POST("/auth/register", middleware.RateLimit(10, time.Minute), auth.Register)
		api.POST("/auth/login", middleware.RateLimit(10, time.Minute), auth.Login)
		api.GET("/auth/google", middleware.RateLimit(20, time.Minute), auth.GoogleLogin)
		api.GET("/auth/google/callback", middleware.RateLimit(20, time.Minute), auth.GoogleCallback)
		api.GET("/posts/:id", middleware.RateLimit(120, time.Minute), post.GetPostByID)
		api.GET("/posts/:id/comments", middleware.OptionalAuth(cfg.JWTSecret), middleware.RateLimit(120, time.Minute), post.GetComments)
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
	server := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 120 * time.Second, IdleTimeout: 120 * time.Second}
	go func() {
		log.Printf("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatalf("Server failed: %v", err) }
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil { log.Printf("Server shutdown failed: %v", err) }
}
