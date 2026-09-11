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
	Visibility  string `json:"visibility" binding:"omitempty,oneof=public private"`
	Caption     string `json:"caption" binding:"max=2000"`
}

type createCommentInput struct {
	Content  string `json:"content" binding:"required,max=2000"`
	ParentID *uint  `json:"parentId"`
}

const postEngagementSelect = `posts.*, (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS like_count, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comment_count, EXISTS(SELECT 1 FROM likes WHERE likes.post_id = posts.id AND likes.user_id = ?) AS liked`

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
		if strings.EqualFold(candidate.Scheme, public.Scheme) && strings.EqualFold(candidate.Host, public.Host) && strings.HasPrefix(candidate.Path, "/uploads/") && candidate.RawQuery == "" && candidate.Fragment == "" {
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

func validateManagedMediaReference(filename, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
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
	userID := c.MustGet("userId").(uint)
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

	candidate, err := url.Parse(contentURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid uploaded media URL"})
		return
	}
	decodedPath, err := url.PathUnescape(candidate.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid uploaded media URL"})
		return
	}
	relative := strings.TrimPrefix(decodedPath, "/uploads/")
	filename := filepath.Base(filepath.FromSlash(relative))
	if relative == "" || filename != relative || filename == "." || filename == string(filepath.Separator) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid uploaded media URL"})
		return
	}

	var upload models.Upload
	if err := h.DB.Where("filename = ? AND user_id = ? AND post_id IS NULL", filename, userID).First(&upload).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "uploaded media is not owned by the current user or has already been used"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate uploaded media ownership"})
		return
	}

	var contentTypeMatches bool
	switch input.ContentType {
	case "image":
		contentTypeMatches = upload.MediaType == "image/jpeg" || upload.MediaType == "image/png" || upload.MediaType == "image/webp"
	case "video":
		contentTypeMatches = upload.MediaType == "video/mp4" || upload.MediaType == "video/quicktime"
	}
	if !contentTypeMatches {
		c.JSON(http.StatusBadRequest, gin.H{"message": "uploaded media type does not match content type"})
		return
	}
	if err := validateManagedMediaReference(filename, input.ContentType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	visibility := strings.TrimSpace(input.Visibility)
	if visibility == "" {
		visibility = "public"
	}

	var post models.Post
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var claimed models.Upload
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ? AND post_id IS NULL", upload.ID, userID).First(&claimed).Error; err != nil {
			return err
		}
		post = models.Post{
			UserID:      userID,
			ContentType: input.ContentType,
			ContentURL:  contentURL,
			Visibility:  visibility,
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
	post.CommentCount = 0
	post.Liked = false
	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully", "data": post})
}
