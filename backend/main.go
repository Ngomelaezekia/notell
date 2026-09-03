package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"notell/config"
	"notell/handlers"
	"notell/middleware"
	"notell/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const maxJSONBodyBytes = 1 << 20

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
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
	sqlDB.SetConnMaxIdleTime(time.Duration(envInt("DB_CONN_MAX_IDLE_MINUTES", 5)) * time.Minute)
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Relationship{},
		&models.Channel{},
		&models.Notification{},
		&models.Upload{},
	); err != nil {
		log.Fatalf("Database auto-migration failed: %v", err)
	}

	r := gin.Default()
	r.Use(
		middleware.MaxBodyBytes(maxJSONBodyBytes),
		cors.New(cors.Config{
			AllowOrigins:     []string{cfg.FrontendURL},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	)
	r.Static("/uploads", "./uploads")

	auth := handlers.NewAuthHandler(db, cfg)
	post := handlers.NewPostHandler(db, cfg.PublicURL)
	userHandler := handlers.NewUserHandler(db)
	relationshipHandler := handlers.NewRelationshipHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	uploadHandler := handlers.NewUploadHandler(db, cfg.PublicURL)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.GET("/ready", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
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

		protected := api.Group("/").Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			protected.POST("/upload", middleware.RateLimit(20, time.Minute), uploadHandler.UploadMedia)
			protected.GET("/auth/me", middleware.RateLimit(120, time.Minute), auth.Me)
			protected.POST("/auth/logout", middleware.RateLimit(30, time.Minute), auth.Logout)
			protected.PUT("/users/profile", middleware.RateLimit(30, time.Minute), userHandler.UpdateProfile)
			protected.GET("/users/search", middleware.RateLimit(60, time.Minute), userHandler.SearchUsers)
			protected.GET("/posts/search", middleware.RateLimit(60, time.Minute), post.SearchPosts)
			protected.POST("/posts", middleware.RateLimit(30, time.Minute), post.CreatePost)
			protected.GET("/posts/feed", middleware.RateLimit(120, time.Minute), post.GetFeed)
			protected.DELETE("/posts/:id", middleware.RateLimit(30, time.Minute), post.DeletePost)
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
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		log.Printf("Server running in %s mode on port %s...", cfg.AppEnv, cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
}
