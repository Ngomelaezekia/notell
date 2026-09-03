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
	DB             *gorm.DB
	PublicURL      string
	MediaPublicURL string
}

// NewPostHandler preserves the existing two-argument call shape while allowing
// production to supply a separate durable media origin such as Cloudflare R2.
func NewPostHandler(db *gorm.DB, publicURL string, mediaPublicURL ...string) *PostHandler {
	mediaURL := publicURL
	if len(mediaPublicURL) > 0 && strings.TrimSpace(mediaPublicURL[0]) != "" {
		mediaURL = mediaPublicURL[0]
	}
	return &PostHandler{
		DB:             db,
		PublicURL:      strings.TrimRight(publicURL, "/"),
		MediaPublicURL: strings.TrimRight(mediaURL, "/"),
	}
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
	for _, publicURL := range []string{h.MediaPublicURL, h.PublicURL} {
		public, err := url.Parse(publicURL)
		if err != nil || public.Scheme == "" || public.Host == "" {
			continue
		}
		if strings.EqualFold(candidate.Scheme, public.Scheme) &&
			strings.EqualFold(candidate.Host, public.Host) &&
			strings.HasPrefix(candidate.Path, "/uploads/") &&
			candidate.RawQuery == "" && candidate.Fragment == "" {
			return true
		}
	}
	return false
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "contentUrl must be a managed uploaded media URL"})
		return
	}
	mediaPath, ok := h.managedMediaPath(contentURL)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid uploaded media URL"})
		return
	}
	if err := validateManagedMedia(mediaPath, input.ContentType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	filename := filepath.Base(mediaPath)
	var upload models.Upload
	if err := h.DB.Where("filename = ? AND user_id = ? AND post_id IS NULL", filename, userID).First(&upload).Error; err != nil {
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
