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

// GetCategorizedFeed serves the complete discoverable feed with explicit
// categories. Pagination is retained so the client can progressively fetch
// every page without loading the whole database into memory at once.
func (h *PostHandler) GetCategorizedFeed(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	category := strings.ToLower(strings.TrimSpace(c.DefaultQuery("category", "all")))
	if category == "friends" {
		category = "following"
	}
	if category != "all" && category != "popular" && category != "local" && category != "following" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid feed category"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	query := h.DB.Model(&models.Post{}).
		Select(postEngagementSelect, userID)

	switch category {
	case "following":
		query = query.
			Joins(`JOIN user_relationships ON user_relationships.following_id = posts.user_id AND user_relationships.follower_id = ? AND user_relationships.status = ?`, userID, "accepted").
			Where("posts.user_id <> ?", userID)
	case "local":
		var viewer models.User
		if err := h.DB.Select("city", "country").First(&viewer, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load user location"})
			return
		}

		city := strings.TrimSpace(stringValue(viewer.City))
		country := strings.TrimSpace(stringValue(viewer.Country))
		if city == "" && country == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": []models.Post{},
				"pagination": gin.H{"page": page, "limit": limit, "hasMore": false},
				"category": category,
			})
			return
		}

		query = query.Joins("JOIN users ON users.id = posts.user_id")
		if city != "" && country != "" {
			query = query.Where("LOWER(users.city) = LOWER(?) AND LOWER(users.country) = LOWER(?)", city, country)
		} else if city != "" {
			query = query.Where("LOWER(users.city) = LOWER(?)", city)
		} else {
			query = query.Where("LOWER(users.country) = LOWER(?)", country)
		}
	case "popular":
		// Popularity combines engagement with a gentle age decay so an old
		// viral post does not permanently dominate the feed.
		query = query.Order(gorm.Expr(`(
			(COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = posts.id), 0) * 3) +
			(COALESCE((SELECT COUNT(*) FROM comments cm WHERE cm.post_id = posts.id), 0) * 2)
		) / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, 0.5) DESC`))
	}

	if category != "popular" {
		query = query.Order("posts.created_at DESC")
	} else {
		query = query.Order("posts.created_at DESC")
	}
	query = query.Order("posts.id DESC")

	var posts []models.Post
	if err := query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "profile_picture")
		}).
		Offset((page - 1) * limit).
		Limit(limit + 1).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch feed"})
		return
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       posts,
		"category":   category,
		"pagination": gin.H{"page": page, "limit": limit, "hasMore": hasMore},
	})
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
