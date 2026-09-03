package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostHandler struct {
	DB        *gorm.DB
	PublicURL string
}

func NewPostHandler(db *gorm.DB, publicURL string) *PostHandler {
	return &PostHandler{DB: db, PublicURL: strings.TrimRight(publicURL, "/")}
}

type createPostInput struct {
	ContentType string `json:"contentType" binding:"required,oneof=image video"`
	ContentURL  string `json:"contentUrl" binding:"required,url"`
	Caption     string `json:"caption" binding:"max=2000"`
}

type createCommentInput struct {
	Content  string `json:"content" binding:"required,max=2000"`
	ParentID *uint  `json:"parentId"`
}

const postEngagementSelect = `posts.*, (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS like_count, EXISTS(SELECT 1 FROM likes WHERE likes.post_id = posts.id AND likes.user_id = ?) AS liked`

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return replacer.Replace(value)
}

func (h *PostHandler) isManagedMediaURL(value string) bool {
	candidate, err := url.Parse(strings.TrimSpace(value))
	if err != nil || candidate.Scheme == "" || candidate.Host == "" {
		return false
	}
	public, err := url.Parse(h.PublicURL)
	if err != nil || public.Scheme == "" || public.Host == "" {
		return false
	}
	if !strings.EqualFold(candidate.Scheme, public.Scheme) || !strings.EqualFold(candidate.Host, public.Host) {
		return false
	}
	return strings.HasPrefix(candidate.Path, "/uploads/") && candidate.RawQuery == "" && candidate.Fragment == ""
}

func (h *PostHandler) managedMediaPath(value string) (string, bool) {
	candidate, err := url.Parse(strings.TrimSpace(value))
	if err != nil || !h.isManagedMediaURL(value) {
		return "", false
	}
	decodedPath, err := url.PathUnescape(candidate.Path)
	if err != nil {
		return "", false
	}
	relative := strings.TrimPrefix(decodedPath, "/uploads/")
	if relative == "" {
		return "", false
	}
	filename := filepath.Base(filepath.FromSlash(relative))
	if filename != relative || filename == "." || filename == string(filepath.Separator) {
		return "", false
	}
	return filepath.Join("uploads", filename), true
}

func validateManagedMedia(path, contentType string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("uploaded media file not found")
		}
		return err
	}
	if info.IsDir() {
		return errors.New("uploaded media path is not a file")
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch contentType {
	case "image":
		if ext != ".jpg" && ext != ".png" && ext != ".webp" {
			return errors.New("media file type does not match image content type")
		}
	case "video":
		if ext != ".mp4" && ext != ".mov" {
			return errors.New("media file type does not match video content type")
		}
	default:
		return errors.New("unsupported content type")
	}
	return nil
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	authUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	userID := authUserID.(uint)

	var input createPostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	contentURL := strings.TrimSpace(input.ContentURL)
	if !h.isManagedMediaURL(contentURL) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "contentUrl must reference media uploaded to this server"})
		return
	}
	mediaPath, ok := h.managedMediaPath(contentURL)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid uploaded media path"})
		return
	}
	if err := validateManagedMedia(mediaPath, input.ContentType); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "uploaded media file not found"})
			return
		}
		if strings.Contains(err.Error(), "media") || strings.Contains(err.Error(), "content type") || strings.Contains(err.Error(), "not a file") {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate uploaded media"})
		return
	}

	var upload models.Upload
	if err := h.DB.Where("filename = ? AND user_id = ? AND post_id IS NULL", filepath.Base(mediaPath), userID).First(&upload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "uploaded media is not owned by the current user or has already been used"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate uploaded media ownership"})
		return
	}

	var post models.Post
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var claimed models.Upload
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ? AND post_id IS NULL", upload.ID, userID).First(&claimed).Error; err != nil {
			return err
		}
		post = models.Post{
			UserID:      userID,
			ContentType: input.ContentType,
			ContentURL:  contentURL,
			Caption:     strings.TrimSpace(input.Caption),
		}
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&claimed).Updates(map[string]interface{}{"post_id": post.ID, "claimed_at": now}).Error; err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "uploaded media is not owned by the current user or has already been used"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create post"})
		return
	}

	if err := h.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "profile_picture")
	}).First(&post, post.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load created post"})
		return
	}

	post.LikeCount = 0
	post.Liked = false
	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully", "data": post})
}

func (h *PostHandler) GetFeed(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var posts []models.Post
	err := h.DB.Model(&models.Post{}).
		Select(postEngagementSelect, userID).
		Joins(`LEFT JOIN user_relationships ON user_relationships.following_id = posts.user_id AND user_relationships.follower_id = ? AND user_relationships.status = ?`, userID, "accepted").
		Where(`posts.user_id = ? OR user_relationships.follower_id IS NOT NULL`, userID).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		Order("posts.created_at DESC").
		Order("posts.id DESC").
		Limit(limit + 1).
		Offset((page - 1) * limit).
		Find(&posts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch feed"})
		return
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}
	c.JSON(http.StatusOK, gin.H{
		"data": posts,
		"pagination": gin.H{
			"page":    page,
			"limit":   limit,
			"hasMore": hasMore,
		},
	})
}

// SearchPosts searches captions and author usernames. Results prioritize exact
// username matches, username prefixes, caption prefixes, then recency.
func (h *PostHandler) SearchPosts(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query must be at least 2 characters"})
		return
	}
	if len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query is too long"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	escapedQuery := escapeLikePattern(query)
	pattern := "%" + escapedQuery + "%"
	prefix := escapedQuery + "%"
	var total int64
	base := h.DB.Model(&models.Post{}).
		Joins("JOIN users ON users.id = posts.user_id").
		Where("posts.caption ILIKE ? ESCAPE '\\' OR users.username ILIKE ? ESCAPE '\\'", pattern, pattern)
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	var posts []models.Post
	authUserID := uint(0)
	if value, ok := c.Get("userId"); ok {
		if id, ok := value.(uint); ok {
			authUserID = id
		}
	}

	err = base.Select(postEngagementSelect, authUserID).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		Order(gorm.Expr(`CASE
			WHEN LOWER(users.username) = LOWER(?) THEN 0
			WHEN LOWER(users.username) LIKE LOWER(?) ESCAPE '\\' THEN 1
			WHEN LOWER(posts.caption) LIKE LOWER(?) ESCAPE '\\' THEN 2
			ELSE 3
		END`, query, prefix, prefix)).
		Order("posts.created_at DESC").
		Order("posts.id DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&posts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"posts": posts,
			"pagination": gin.H{
				"page":    page,
				"limit":   limit,
				"total":   total,
				"hasMore": int64(page*limit) < total,
			},
		},
	})
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var post models.Post
	userID := uint(0)
	if value, ok := c.Get("userId"); ok {
		if id, ok := value.(uint); ok {
			userID = id
		}
	}

	err = h.DB.Model(&models.Post{}).
		Select(postEngagementSelect, userID).
		Where("posts.id = ?", uint(id)).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": post})
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}
	postIDUint := uint(postID)
	var contentURL string

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.Select("id, content_url").Where("id = ? AND user_id = ?", postIDUint, userID).First(&post).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrRecordNotFound
			}
			return err
		}
		contentURL = post.ContentURL

		if err := tx.Delete(&post).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", postIDUint).Delete(&models.Notification{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", postIDUint).Delete(&models.Upload{}).Error; err != nil {
			return err
		}
		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete post"})
		return
	}

	if path, ok := h.managedMediaPath(contentURL); ok {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("failed to remove deleted post media %q: %v", path, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

func (h *PostHandler) ToggleLike(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}
	postIDUint := uint(postID)

	var post models.Post
	liked := false

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id, user_id").First(&post, postIDUint).Error; err != nil {
			return err
		}

		var like models.Like
		likeErr := tx.Where("user_id = ? AND post_id = ?", userID, postIDUint).First(&like).Error
		switch {
		case likeErr == nil:
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ? AND actor_id = ? AND post_id = ? AND type = ?", post.UserID, userID, postIDUint, "like").Delete(&models.Notification{}).Error; err != nil {
				return err
			}
			liked = false
		case errors.Is(likeErr, gorm.ErrRecordNotFound):
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Like{UserID: userID, PostID: postIDUint}).Error; err != nil {
				return err
			}
			liked = true
			_ = CreateNotification(tx, post.UserID, userID, "like", func() *uint {
				id := postIDUint
				return &id
			}(), nil)
		default:
			return likeErr
		}
		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to toggle like"})
		return
	}

	var likeCount int64
	if err := h.DB.Model(&models.Like{}).Where("post_id = ?", postIDUint).Count(&likeCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count likes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": liked, "likeCount": likeCount})
}

func (h *PostHandler) AddComment(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var input createCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var post models.Post
	if err := h.DB.Select("id, user_id").First(&post, uint(postID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to find post"})
		return
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "comment cannot be empty"})
		return
	}

	var parent *models.Comment
	notificationType := "comment"
	targetUserID := post.UserID
	if input.ParentID != nil {
		parent = &models.Comment{}
		if err := h.DB.Select("id, post_id, parent_id, user_id").First(parent, *input.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"message": "parent comment not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate parent comment"})
			return
		}
		if parent.PostID != uint(postID) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "parent comment belongs to another post"})
			return
		}
		if parent.ParentID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "replies can only target top-level comments"})
			return
		}
		targetUserID = parent.UserID
		notificationType = "reply"
	}

	comment := models.Comment{
		PostID:   uint(postID),
		UserID:   userID,
		Content:  content,
		ParentID: input.ParentID,
	}
	if err := h.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to add comment"})
		return
	}
	if err := h.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "profile_picture")
	}).First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load comment"})
		return
	}

	_ = CreateNotification(h.DB, targetUserID, userID, notificationType, func() *uint {
		id := uint(postID)
		return &id
	}(), &comment.ID)
	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

func (h *PostHandler) GetComments(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var comments []models.Comment
	if err := h.DB.Where("post_id = ? AND parent_id IS NULL", uint(postID)).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Order("id ASC")
		}).
		Preload("Replies.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		Order("created_at ASC").
		Order("id ASC").
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": comments})
}
