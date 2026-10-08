package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"
	"notell/services/postaccess"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *PostHandler) SharePost(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
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
	if err := postaccess.CanViewRecordWithDB(h.DB, post, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	var existing models.PostShare
	if err := h.DB.Where("post_id = ? AND user_id = ?", post.ID, userID).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to check share state"})
			return
		}
		if err := h.DB.Create(&models.PostShare{PostID: post.ID, UserID: userID}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to share post"})
			return
		}
	}

	var shareCount int64
	if err := h.DB.Model(&models.PostShare{}).Where("post_id = ?", post.ID).Count(&shareCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "share recorded but count failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post shared",
		"shared": true,
		"shareCount": shareCount,
	})
}
