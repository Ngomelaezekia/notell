package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	defer sqlDB.Close()

	if err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Relationship{},
		&models.Channel{},
		&models.Notification{},
	); err != nil {
		log.Fatalf("Database auto-migration failed: %v", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Static("/uploads", "./uploads")

	auth := handlers.NewAuthHandler(db, cfg)
	post := handlers.NewPostHandler(db)
	userHandler := handlers.NewUserHandler(db)
	relationshipHandler := handlers.NewRelationshipHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db)
	uploadHandler := handlers.NewUploadHandler(cfg.PublicURL)

	api := r.Group("/api")
	{
		api.POST("/auth/register", auth.Register)
		api.POST("/auth/login", auth.Login)
		api.GET("/auth/google", auth.GoogleLogin)
		api.GET("/auth/google/callback", auth.GoogleCallback)
		api.GET("/posts/:id", post.GetPostByID)
		api.GET("/posts/:id/comments", post.GetComments)
		api.GET("/users/:id", userHandler.GetUserProfile)

		protected := api.Group("/").Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			protected.POST("/upload", uploadHandler.UploadMedia)
			protected.GET("/auth/me", auth.Me)
			protected.POST("/auth/logout", auth.Logout)
			protected.PUT("/users/profile", userHandler.UpdateProfile)
			protected.GET("/users/search", userHandler.SearchUsers)
			protected.GET("/posts/search", post.SearchPosts)
			protected.POST("/posts", post.CreatePost)
			protected.GET("/posts/feed", post.GetFeed)
			protected.DELETE("/posts/:id", post.DeletePost)
			protected.POST("/posts/:id/like", post.ToggleLike)
			protected.POST("/posts/:id/comments", post.AddComment)
			protected.POST("/users/:id/follow", relationshipHandler.FollowUser)
			protected.DELETE("/users/:id/unfollow", relationshipHandler.UnfollowUser)
			protected.GET("/users/:id/relationship", relationshipHandler.GetRelationshipStatus)
			protected.DELETE("/users/followers/:id", relationshipHandler.RemoveFollower)
			protected.GET("/users/:id/followers", relationshipHandler.GetFollowers)
			protected.GET("/users/:id/following", relationshipHandler.GetFollowing)
			protected.GET("/notifications", notificationHandler.List)
			protected.POST("/notifications/:id/read", notificationHandler.MarkRead)
			protected.POST("/notifications/read-all", notificationHandler.MarkAllRead)
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
