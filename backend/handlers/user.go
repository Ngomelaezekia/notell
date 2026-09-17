package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct{ DB *gorm.DB }

func NewUserHandler(db *gorm.DB) *UserHandler { return &UserHandler{DB: db} }

type updateProfileInput struct {
	Username       *string `json:"username" binding:"omitempty,gt=0,max=50"`
	Email          *string `json:"email" binding:"omitempty,email,max=254"`
	Country        *string `json:"country" binding:"max=100"`
	City           *string `json:"city" binding:"max=100"`
	Bio            *string `json:"bio" binding:"max=2000"`
	ProfilePicture *string `json:"profilePicture" binding:"omitempty,url,max=2048"`
	CoverPicture   *string `json:"coverPicture" binding:"omitempty,url,max=2048"`
	AllowFollowers *bool   `json:"allowFollowers"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	authUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	userID := authUserID.(uint)
	var input updateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if input.Username != nil {
		trimmed := strings.TrimSpace(*input.Username)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "username cannot be empty"})
			return
		}
		var existing models.User
		err := h.DB.Where("username = ? AND id != ?", trimmed, userID).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "username is already taken"})
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
		updates["username"] = trimmed
	}
	if input.Email != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*input.Email))
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "email cannot be empty"})
			return
		}
		var existing models.User
		err := h.DB.Where("email = ? AND id != ?", trimmed, userID).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "email is already in use"})
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
		updates["email"] = trimmed
	}
	if input.Country != nil {
		updates["country"] = strings.TrimSpace(*input.Country)
	}
	if input.City != nil {
		updates["city"] = strings.TrimSpace(*input.City)
	}
	if input.Bio != nil {
		updates["bio"] = strings.TrimSpace(*input.Bio)
	}
	if input.ProfilePicture != nil {
		updates["profile_picture"] = strings.TrimSpace(*input.ProfilePicture)
	}
	if input.CoverPicture != nil {
		updates["cover_picture"] = strings.TrimSpace(*input.CoverPicture)
	}
	if input.AllowFollowers != nil {
		updates["allow_followers"] = *input.AllowFollowers
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no valid fields provided for update"})
		return
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"message": "username or email is already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update profile"})
		return
	}
	var updatedUser models.User
	if err := h.DB.Select("id", "username", "email", "profile_picture", "cover_picture", "bio", "city", "country", "status", "allow_followers", "created_at", "updated_at").First(&updatedUser, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch updated profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully", "data": gin.H{"user": updatedUser}})
}

type searchUserResult struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
	CoverPicture   *string   `json:"coverPicture,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Country        *string   `json:"country,omitempty"`
	City           *string   `json:"city,omitempty"`
	Status         string    `json:"status"`
	AllowFollowers bool      `json:"allowFollowers"`
	CreatedAt      time.Time `json:"createdAt"`
	Following      bool      `json:"following"`
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query must be at least 2 characters"})
		return
	}
	if len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query is too long"})
		return
	}
	viewerIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	viewerID := viewerIDValue.(uint)
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
	usernameQuery := strings.TrimPrefix(query, "@")
	if usernameQuery == "" {
		usernameQuery = query
	}
	escaped := escapeLikePattern(usernameQuery)
	pattern := "%" + escaped + "%"
	prefixPattern := escaped + "%"
	var total int64
	base := h.DB.Model(&models.User{}).Where("(username ILIKE ? ESCAPE '\\' OR COALESCE(bio,'') ILIKE ? ESCAPE '\\' OR COALESCE(city,'') ILIKE ? ESCAPE '\\' OR COALESCE(country,'') ILIKE ? ESCAPE '\\')", pattern, pattern, pattern, pattern)
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}
	var users []searchUserResult
	orderRank := gorm.Expr("CASE WHEN username ILIKE ? THEN 0 WHEN username ILIKE ? ESCAPE '\\' THEN 1 WHEN username ILIKE ? ESCAPE '\\' THEN 2 ELSE 3 END", usernameQuery, prefixPattern, pattern)
	err := base.Select("id, username, profile_picture, cover_picture, bio, country, city, status, allow_followers, created_at").Order(orderRank).Order("username ASC").Order("id ASC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}
	if len(users) > 0 {
		ids := make([]uint, len(users))
		for i := range users {
			ids[i] = users[i].ID
		}
		var relationships []models.Relationship
		if err := h.DB.Where("follower_id = ? AND following_id IN ? AND status = ?", viewerID, ids, "accepted").Find(&relationships).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
		following := make(map[uint]struct{}, len(relationships))
		for _, r := range relationships {
			following[r.FollowingID] = struct{}{}
		}
		for i := range users {
			_, users[i].Following = following[users[i].ID]
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"users": users, "pagination": gin.H{"page": page, "limit": limit, "total": total, "hasMore": int64(page*limit) < total}}})
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	targetIDUint, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	targetID := uint(targetIDUint)
	viewerID := uint(0)
	if value, ok := c.Get("userId"); ok {
		if id, ok := value.(uint); ok {
			viewerID = id
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 24
	}
	if limit > 36 {
		limit = 36
	}
	var user models.User
	err = h.DB.Select("id, username, profile_picture, cover_picture, bio, city, country, allow_followers, created_at").Preload("Posts", func(db *gorm.DB) *gorm.DB {
		q := db.Select("id,user_id,upload_id,content_type,content_url,visibility,caption,created_at,updated_at")
		if viewerID != targetID {
			q = q.Where("COALESCE(visibility,'public') = ?", "public")
		}
		return q.Preload("Upload.MediaMetadata").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit)
	}).First(&user, targetID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}
	var postCount int64
	countQuery := h.DB.Model(&models.Post{}).Where("user_id = ?", targetID)
	if viewerID != targetID {
		countQuery = countQuery.Where("COALESCE(visibility,'public') = ?", "public")
	}
	if err := countQuery.Count(&postCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count profile posts"})
		return
	}
	user.PostCount = postCount
	var followerCount int64
	if err := h.DB.Model(&models.Relationship{}).Where("following_id = ? AND status = ?", targetID, "accepted").Count(&followerCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count followers"})
		return
	}
	var followingCount int64
	if err := h.DB.Model(&models.Relationship{}).Where("follower_id = ? AND status = ?", targetID, "accepted").Count(&followingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count following"})
		return
	}
	user.FollowerCount = followerCount
	user.FollowingCount = followingCount
	if viewerID != 0 && viewerID != targetID {
		var outgoing models.Relationship
		outgoingErr := h.DB.Where("follower_id = ? AND following_id = ?", viewerID, targetID).First(&outgoing).Error
		if outgoingErr == nil {
			user.RelationshipStatus = outgoing.Status
			user.IsFollowing = outgoing.Status == "accepted"
		} else if !errors.Is(outgoingErr, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load relationship"})
			return
		}
		var incoming models.Relationship
		incomingErr := h.DB.Where("follower_id = ? AND following_id = ?", targetID, viewerID).First(&incoming).Error
		if incomingErr == nil {
			user.IsFollowedBy = incoming.Status == "accepted"
		} else if !errors.Is(incomingErr, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load relationship"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": user, "pagination": gin.H{"page": page, "limit": limit, "total": postCount, "hasMore": int64(page*limit) < postCount}}})
}
