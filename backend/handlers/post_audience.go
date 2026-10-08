package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"notell/models"
	"notell/services/postaccess"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type postAudienceInput struct {
	UserIDs []uint `json:"userIds" binding:"max=500"`
}

func (h *PostHandler) GetPostAudience(c *gin.Context) {
	ownerID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || postID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var post models.Post
	if err := h.DB.Select("id,user_id,visibility").First(&post, uint(postID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load post"})
		return
	}
	if post.UserID != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"message": "post owner required"})
		return
	}
	if post.Visibility != "selected" {
		c.JSON(http.StatusConflict, gin.H{"message": "post visibility is not selected"})
		return
	}

	var audience []models.PostAudience
	if err := h.DB.Where("post_id = ?", post.ID).Order("user_id ASC").Find(&audience).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load post audience"})
		return
	}
	userIDs := make([]uint, 0, len(audience))
	for _, entry := range audience {
		userIDs = append(userIDs, entry.UserID)
	}
	c.JSON(http.StatusOK, gin.H{"userIds": userIDs})
}

func (h *PostHandler) ReplacePostAudience(c *gin.Context) {
	ownerID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || postID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}

	var input postAudienceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid audience payload"})
		return
	}

	unique := make(map[uint]struct{}, len(input.UserIDs))
	userIDs := make([]uint, 0, len(input.UserIDs))
	for _, id := range input.UserIDs {
		if id == 0 || id == ownerID {
			continue
		}
		if _, exists := unique[id]; exists {
			continue
		}
		unique[id] = struct{}{}
		userIDs = append(userIDs, id)
	}
	if len(userIDs) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "audience cannot exceed 500 users"})
		return
	}

	var post models.Post
	if err := h.DB.Select("id,user_id,visibility").First(&post, uint(postID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load post"})
		return
	}
	if err := postaccess.CanEditPost(h.DB, post.ID, ownerID); err != nil {
		if errors.Is(err, postaccess.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"message": "post owner required"})
		}
		return
	}
	if post.Visibility != "selected" {
		c.JSON(http.StatusConflict, gin.H{"message": "set post visibility to selected before managing its audience"})
		return
	}

	if len(userIDs) > 0 {
		var count int64
		if err := h.DB.Model(&models.User{}).Where("id IN ?", userIDs).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate audience users"})
			return
		}
		if count != int64(len(userIDs)) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "audience contains an unknown user"})
			return
		}
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.PostAudience{}).Error; err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return nil
		}
		entries := make([]models.PostAudience, 0, len(userIDs))
		for _, id := range userIDs {
			entries = append(entries, models.PostAudience{PostID: post.ID, UserID: id})
		}
		return tx.Create(&entries).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update post audience"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post audience updated",
		"userIds": userIDs,
		"count": len(userIDs),
	})
}

var _ = strings.TrimSpace
