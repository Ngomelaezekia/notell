package main

import (
	"log"
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
	// 1. Load Configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Connect to Database
	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	log.Println("Database connection established successfully")

	// 3. Auto-Migrate Database Schemas
	if err := db.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		log.Fatalf("Database auto-migration failed: %v", err)
	}

	// 4. Initialize Router
	r := gin.Default()

	// 5. Setup CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 6. Serve Uploaded Files (Images & Videos)
	r.Static("/uploads", "./uploads")

	// 7. Initialize Handlers
	auth := handlers.NewAuthHandler(db, cfg)
	post := handlers.NewPostHandler(db)
	userHandler := handlers.NewUserHandler(db)
	relationshipHandler := handlers.NewRelationshipHandler(db)
	uploadHandler := handlers.NewUploadHandler()
	// 8. Register API Routes
	api := r.Group("/api")
	{
		// Unprotected Auth Routes
		api.POST("/auth/register", auth.Register)
		api.POST("/auth/login", auth.Login)
		api.GET("/auth/google", auth.GoogleLogin)
		api.GET("/auth/google/callback", auth.GoogleCallback)

		// Public Content & Profile Routes
		api.GET("/posts/:id", post.GetPostByID)
		api.GET("/posts/:id/comments", post.GetComments)
		api.GET("/users/:id", userHandler.GetUserProfile)

		// Protected Routes (JWT Required)
		protected := api.Group("/").Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			//uploads
			protected.POST("/upload", uploadHandler.UploadMedia) // New upload endpoint
			// Auth
			protected.GET("/auth/me", auth.Me)
			protected.POST("/auth/logout", auth.Logout)

			// User Updates
			protected.PUT("/users/profile", userHandler.UpdateProfile)

			// Posts Management
			protected.POST("/posts", post.CreatePost)
			protected.GET("/posts/feed", post.GetFeed)
			protected.DELETE("/posts/:id", post.DeletePost) // Fixed endpoint typo

			// Likes & Comments
			protected.POST("/posts/:id/like", post.ToggleLike)
			protected.POST("/posts/:id/comments", post.AddComment)

			// Relationship Management
			protected.POST("/users/:id/follow", relationshipHandler.FollowUser)
			protected.DELETE("/users/:id/unfollow", relationshipHandler.UnfollowUser)
			protected.DELETE("/users/followers/:id", relationshipHandler.RemoveFollower)

			protected.GET("/users/:id/followers", relationshipHandler.GetFollowers)
			protected.GET("/users/:id/following", relationshipHandler.GetFollowing)
		}
	}

	// 9. Start Server
	serverAddress := ":" + cfg.Port
	log.Printf("Server running in %s mode on port %s...", cfg.AppEnv, cfg.Port)
	if err := r.Run(serverAddress); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
