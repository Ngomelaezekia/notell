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
	cfg := config.Load()
	if cfg.AppEnv == "production" { gin.SetMode(gin.ReleaseMode) }
	db, err := gorm.Open(postgres.Open(cfg.GetDBDSN()), &gorm.Config{})
	if err != nil { log.Fatalf("Failed to connect to PostgreSQL database: %v", err) }
	log.Println("Database connection established successfully")
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Relationship{}, &models.Channel{}); err != nil { log.Fatalf("Database auto-migration failed: %v", err) }

	r := gin.Default()
	r.Use(cors.New(cors.Config{AllowOrigins:[]string{cfg.FrontendURL}, AllowMethods:[]string{"GET","POST","PUT","PATCH","DELETE","OPTIONS"}, AllowHeaders:[]string{"Origin","Content-Type","Accept","Authorization"}, ExposeHeaders:[]string{"Content-Length"}, AllowCredentials:true, MaxAge:12*time.Hour}))
	r.Static("/uploads", "./uploads")
	auth := handlers.NewAuthHandler(db, cfg)
	post := handlers.NewPostHandler(db)
	userHandler := handlers.NewUserHandler(db)
	relationshipHandler := handlers.NewRelationshipHandler(db)
	uploadHandler := handlers.NewUploadHandler()

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
		}
	}
	serverAddress := ":" + cfg.Port
	log.Printf("Server running in %s mode on port %s...", cfg.AppEnv, cfg.Port)
	if err := r.Run(serverAddress); err != nil { log.Fatalf("Server failed to start: %v", err) }
}
