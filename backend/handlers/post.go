package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostHandler struct {
	DB *gorm.DB
}

func NewPostHandler(db *gorm.DB) *PostHandler {
	return &PostHandler{
		DB: db,
	}
}

// ===============================
// DTOs
// ===============================

type createPostInput struct {
	ContentType string `json:"contentType" binding:"required,oneof=image video"`
	ContentURL  string `json:"contentUrl" binding:"required,url"`
	Caption     string `json:"caption"`
}

type createCommentInput struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parentId"`
}

// ===============================
// Create Post
// ===============================

func (h *PostHandler) CreatePost(c *gin.Context) {

	authUserID, exists := c.Get("userId")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	userID := authUserID.(uint)

	var input createPostInput

	if err := c.ShouldBindJSON(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	post := models.Post{
		UserID:      userID,
		ContentType: input.ContentType,
		ContentURL:  input.ContentURL,
		Caption:     input.Caption,
	}

	if err := h.DB.Create(&post).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create post",
		})

		return
	}

	h.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"username",
				"profile_picture",
			)
		}).
		First(&post, post.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "post created successfully",
		"data":    post,
	})
}

// ===============================
// Feed
// ===============================

func (h *PostHandler) GetFeed(c *gin.Context) {

	authUserID, exists := c.Get("userId")

	if !exists {

		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})

		return
	}

	userID := authUserID.(uint)

	page, _ := strconv.Atoi(
		c.DefaultQuery("page", "1"),
	)

	limit, _ := strconv.Atoi(
		c.DefaultQuery("limit", "10"),
	)

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	var posts []models.Post

	err := h.DB.
		Model(&models.Post{}).

		// FIXED TABLE NAME
		Joins(`
			JOIN user_relationships
			ON user_relationships.following_id = posts.user_id
		`).
		Where(`
			user_relationships.follower_id = ?
			AND user_relationships.status = ?
		`,
			userID,
			"accepted",
		).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"username",
				"profile_picture",
			)
		}).
		Order(
			"posts.created_at DESC",
		).
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": err.Error(),
			},
		)

		return
	}

	c.JSON(http.StatusOK, gin.H{

		"data": posts,

		"page": page,

		"limit": limit,
	})
}

// ===============================
// Get Single Post
// ===============================

func (h *PostHandler) GetPostByID(c *gin.Context) {

	id, err := strconv.ParseUint(
		c.Param("id"),
		10,
		32,
	)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid post ID",
		})

		return
	}

	var post models.Post

	err = h.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"username",
				"profile_picture",
			)
		}).
		First(
			&post,
			uint(id),
		).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {

			c.JSON(http.StatusNotFound, gin.H{
				"message": "post not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// ===============================
// Delete Post
// ===============================

func (h *PostHandler) DeletePost(c *gin.Context) {

	userID := c.MustGet("userId").(uint)

	postID, err := strconv.ParseUint(
		c.Param("id"),
		10,
		32,
	)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid post ID",
		})

		return
	}

	result := h.DB.
		Where(
			"id=? AND user_id=?",
			uint(postID),
			userID,
		).
		Delete(&models.Post{})

	if result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": result.Error.Error(),
		})

		return
	}

	if result.RowsAffected == 0 {

		c.JSON(http.StatusNotFound, gin.H{
			"message": "post not found",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post deleted successfully",
	})
}

// ===============================
// Like / Unlike
// ===============================

func (h *PostHandler) ToggleLike(c *gin.Context) {

	userID := c.MustGet("userId").(uint)

	postID, err := strconv.ParseUint(
		c.Param("id"),
		10,
		32,
	)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid post ID",
		})

		return
	}

	var like models.Like

	err = h.DB.
		Where(
			"user_id=? AND post_id=?",
			userID,
			uint(postID),
		).
		First(&like).
		Error

	if err == nil {

		h.DB.Delete(&like)

		c.JSON(http.StatusOK, gin.H{
			"liked": false,
		})

		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {

		newLike := models.Like{
			UserID: userID,
			PostID: uint(postID),
		}

		if err := h.DB.Create(&newLike).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"liked": true,
		})

		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"message": err.Error(),
	})
}

// ===============================
// Add Comment
// ===============================

func (h *PostHandler) AddComment(c *gin.Context) {

	userID := c.MustGet("userId").(uint)

	postID, err := strconv.ParseUint(
		c.Param("id"),
		10,
		32,
	)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid post ID",
		})

		return
	}

	var input createCommentInput

	if err := c.ShouldBindJSON(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		return
	}

	comment := models.Comment{

		UserID: userID,

		PostID: uint(postID),

		Content: input.Content,

		ParentID: input.ParentID,
	}

	if err := h.DB.Create(&comment).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "comment added",
		"data":    comment,
	})
}

// ===============================
// Get Comments
// ===============================

func (h *PostHandler) GetComments(c *gin.Context) {

	postID, err := strconv.ParseUint(
		c.Param("id"),
		10,
		32,
	)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid post ID",
		})

		return
	}

	var comments []models.Comment

	err = h.DB.
		Where(
			"post_id=? AND parent_id IS NULL",
			uint(postID),
		).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"username",
				"profile_picture",
			)
		}).
		Preload("Replies.User", func(db *gorm.DB) *gorm.DB {
			return db.Select(
				"id",
				"username",
				"profile_picture",
			)
		}).
		Order(
			"created_at ASC",
		).
		Find(&comments).
		Error

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": comments,
	})
}
