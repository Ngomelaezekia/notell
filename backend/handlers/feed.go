package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCategorizedFeed serves ranked home-feed categories. Ranking is deliberately
// deterministic and database-side so pagination stays cheap and the client does
// not have to load the whole feed before sorting it. The scoring inputs can later
// be replaced/extended by a learned ranking model without changing the API.
func (h *PostHandler) GetCategorizedFeed(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	category := strings.ToLower(strings.TrimSpace(c.DefaultQuery("category", "all")))
	if category == "friends" || category == "my-friends" {
		category = "following"
	}
	if category != "all" && category != "popular" && category != "local" && category != "following" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid feed category"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 50 { limit = 20 }

	query := h.DB.Model(&models.Post{}).Select(postEngagementSelect, userID)
	// Private posts are visible only to their owner until the future
	// subscription/channel authorization service is introduced.
	query = query.Where("(COALESCE(posts.visibility, 'public') = ? OR posts.user_id = ?)", "public", userID)

	switch category {
	case "following":
		query = query.
			Joins(`LEFT JOIN user_relationships ON user_relationships.following_id = posts.user_id AND user_relationships.follower_id = ? AND user_relationships.status = ?`, userID, "accepted").
			Where(`posts.user_id = ? OR user_relationships.follower_id IS NOT NULL`, userID).
			Order(gorm.Expr(`((COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = posts.id), 0) * 3) + (COALESCE((SELECT COUNT(*) FROM comments cm WHERE cm.post_id = posts.id), 0) * 2) + COALESCE(posts.view_count, 0)) / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, 0.35) DESC`))
	case "local":
		var viewer models.User
		if err := h.DB.Select("city", "country").First(&viewer, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found"}); return }
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load user location"}); return
		}
		city := strings.TrimSpace(stringValue(viewer.City))
		country := strings.TrimSpace(stringValue(viewer.Country))
		if city == "" && country == "" {
			c.JSON(http.StatusOK, gin.H{"data": []models.Post{}, "category": category, "pagination": gin.H{"page": page, "limit": limit, "total": 0, "hasMore": false}})
			return
		}
		query = query.Joins("JOIN users feed_users ON feed_users.id = posts.user_id")
		if city != "" { query = query.Where("LOWER(feed_users.city) = LOWER(?)", city) } else { query = query.Where("LOWER(feed_users.country) = LOWER(?)", country) }
		query = query.Order(gorm.Expr(`((COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = posts.id), 0) * 3) + (COALESCE((SELECT COUNT(*) FROM comments cm WHERE cm.post_id = posts.id), 0) * 2) + COALESCE(posts.view_count, 0)) / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, 0.45) DESC`))
	case "popular":
		query = query.Joins("JOIN users feed_users ON feed_users.id = posts.user_id").Order(gorm.Expr(`((COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = posts.id), 0) * 3) + (COALESCE((SELECT COUNT(*) FROM comments cm WHERE cm.post_id = posts.id), 0) * 2) + (COALESCE(posts.view_count, 0) * 4) + (COALESCE((SELECT COUNT(*) FROM user_relationships r WHERE r.following_id = posts.user_id AND r.status = 'accepted'), 0) * 0.5)) / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, 0.5) DESC`))
	case "all":
		query = query.Joins("JOIN users feed_users ON feed_users.id = posts.user_id").Order(gorm.Expr(`((COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = posts.id), 0) * 3) + (COALESCE((SELECT COUNT(*) FROM comments cm WHERE cm.post_id = posts.id), 0) * 2) + (COALESCE(posts.view_count, 0) * 4) + (COALESCE((SELECT COUNT(*) FROM user_relationships r WHERE r.following_id = posts.user_id AND r.status = 'accepted'), 0) * 0.5) + CASE WHEN posts.user_id = ? THEN 4 ELSE 0 END) / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, 0.45) DESC`, userID))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count feed posts"}); return }
	query = query.Order("posts.id DESC").Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "profile_picture", "city", "country") }).Preload("Upload.MediaMetadata").Offset((page-1)*limit).Limit(limit+1)

	var posts []models.Post
	if err := query.Find(&posts).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch feed"}); return }
	hasMore := len(posts) > limit
	if hasMore { posts = posts[:limit] }
	c.JSON(http.StatusOK, gin.H{"data": posts, "category": category, "pagination": gin.H{"page": page, "limit": limit, "total": total, "hasMore": hasMore}})
}

func stringValue(value *string) string { if value == nil { return "" }; return *value }
