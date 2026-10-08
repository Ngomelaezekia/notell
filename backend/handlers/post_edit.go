package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"notell/models"
	"notell/services/postaccess"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type updatePostInput struct {
	ContentType *string `json:"contentType"`
	ContentURL  *string `json:"contentUrl"`
	Visibility  *string `json:"visibility"`
	Caption     *string `json:"caption"`
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || postID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var input updatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if input.ContentType == nil && input.ContentURL == nil && input.Visibility == nil && input.Caption == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no post changes supplied"})
		return
	}
	if input.ContentURL != nil && input.ContentType == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "contentType is required when replacing media"})
		return
	}
	if input.ContentType != nil && *input.ContentType != "image" && *input.ContentType != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "contentType must be image or video"})
		return
	}

	var updated models.Post
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&post, uint(postID)).Error; err != nil {
			return err
		}
		if post.UserID != userID {
			return postaccess.ErrAccessDenied
		}

		if input.Visibility != nil {
			visibility := strings.TrimSpace(*input.Visibility)
			switch visibility {
			case "public", "private", "followers", "subscribers", "selected":
				post.Visibility = visibility
			default:
				return errors.New("unsupported post visibility")
			}
		}
		if input.Caption != nil {
			post.Caption = strings.TrimSpace(*input.Caption)
		}

		if input.ContentURL != nil {
			contentURL := strings.TrimSpace(*input.ContentURL)
			if !h.isManagedMediaURL(contentURL) {
				return errors.New("contentUrl must be a managed uploaded media URL")
			}
			candidate, err := url.Parse(contentURL)
			if err != nil {
				return errors.New("invalid uploaded media URL")
			}
			decodedPath, err := url.PathUnescape(candidate.Path)
			if err != nil {
				return errors.New("invalid uploaded media URL")
			}
			relative := strings.TrimPrefix(decodedPath, "/uploads/")
			filename := filepath.Base(filepath.FromSlash(relative))
			if relative == "" || filename != relative || filename == "." || filename == string(filepath.Separator) {
				return errors.New("invalid uploaded media URL")
			}
			if err := validateManagedMediaReference(filename, *input.ContentType); err != nil {
				return err
			}

			var replacement models.Upload
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id <> COALESCE(?, 0) AND filename = ? AND user_id = ? AND post_id IS NULL", post.UploadID, filename, userID).
				First(&replacement).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("replacement media is not owned by the current user or has already been used")
				}
				return err
			}

			var metadata models.MediaMetadata
			if err := tx.Where("upload_id = ?", replacement.ID).First(&metadata).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("replacement media is not ready for publication")
				}
				return err
			}
			if strings.ToLower(strings.TrimSpace(metadata.Status)) != models.MediaStatusReady {
				return errors.New("replacement media is not ready for publication")
			}

			var contentTypeMatches bool
			switch *input.ContentType {
			case "image":
				contentTypeMatches = replacement.MediaType == "image/jpeg" || replacement.MediaType == "image/png" || replacement.MediaType == "image/webp"
			case "video":
				contentTypeMatches = replacement.MediaType == "video/mp4" || replacement.MediaType == "video/quicktime" || replacement.MediaType == "video/webm"
			}
			if !contentTypeMatches {
				return errors.New("replacement media type does not match content type")
			}

			now := time.Now()
			if post.UploadID != nil {
				if err := tx.Model(&models.Upload{}).Where("id = ? AND post_id = ?", *post.UploadID, post.ID).
					Updates(map[string]interface{}{"post_id": nil, "claimed_at": nil}).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&replacement).Updates(map[string]interface{}{"post_id": post.ID, "claimed_at": now}).Error; err != nil {
				return err
			}
			post.UploadID = &replacement.ID
			post.ContentURL = contentURL
			post.ContentType = *input.ContentType
		} else if input.ContentType != nil {
			return errors.New("contentUrl is required when changing contentType")
		}

		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		updated = post
		return nil
	})

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	case errors.Is(err, postaccess.ErrAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{"message": "only the post owner can edit this post"})
		return
	case err != nil:
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "replacement media") ||
			strings.Contains(err.Error(), "contentUrl") ||
			strings.Contains(err.Error(), "contentType") ||
			strings.Contains(err.Error(), "visibility") ||
			strings.Contains(err.Error(), "unsupported") {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	if err := h.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "profile_picture")
	}).First(&updated, updated.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "post updated but failed to load it"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post updated successfully", "data": updated})
}
