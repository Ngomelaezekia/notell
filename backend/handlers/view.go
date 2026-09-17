package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordPostView records one unique authenticated view per user/post pair.
// A repeated visibility event is intentionally a no-op so autoplay/scrolling
// cannot artificially inflate popularity.
func (h *PostHandler) RecordPostView(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"})
		return
	}
	postIDUint := uint(postID)

	viewed := false
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.Select("id, user_id, visibility").First(&post, postIDUint).Error; err != nil {
			return err
		}
		if post.Visibility == "private" && post.UserID != userID {
			return gorm.ErrRecordNotFound
		}

		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&models.PostView{PostID: postIDUint, UserID: userID})
		if result.Error != nil {
			return result.Error
		}
		viewed = true
		if result.RowsAffected == 1 {
			if err := tx.Model(&models.Post{}).Where("id = ?", postIDUint).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to record post view"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"viewed": viewed})
}
