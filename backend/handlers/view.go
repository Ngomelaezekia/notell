package handlers

import (
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
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message": "invalid post ID"}); return }

	var post models.Post
	if err := h.DB.Select("id, user_id, visibility").First(&post, uint(postID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound { c.JSON(http.StatusNotFound, gin.H{"message": "post not found"}); return }
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to validate post"}); return
	}
	if post.Visibility == "private" && post.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	result := h.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "post_id"}, {Name: "user_id"}}, DoNothing: true}).Create(&models.PostView{PostID: uint(postID), UserID: userID})
	if result.Error != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to record post view"}); return }
	if result.RowsAffected == 1 {
		if err := h.DB.Model(&models.Post{}).Where("id = ?", uint(postID)).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update post view count"}); return
		}
	}
	c.JSON(http.StatusOK, gin.H{"viewed": true})
}
