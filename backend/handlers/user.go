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

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

type updateProfileInput struct {
	Username       *string `json:"username" binding:"omitempty,gt=0"`
	Email          *string `json:"email" binding:"omitempty,email"`
	Country        *string `json:"country"`
	City           *string `json:"city"`
	Bio            *string `json:"bio"`
	ProfilePicture *string `json:"profilePicture" binding:"omitempty,url"`
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
		trimmedUsername := strings.TrimSpace(*input.Username)
		if trimmedUsername == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "username cannot be empty"})
			return
		}
		var existingUser models.User
		err := h.DB.Where("username = ? AND id != ?", trimmedUsername, userID).First(&existingUser).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "username is already taken"})
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
		updates["username"] = trimmedUsername
	}

	if input.Email != nil {
		trimmedEmail := strings.TrimSpace(strings.ToLower(*input.Email))
		if trimmedEmail == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "email cannot be empty"})
			return
		}
		var existingUser models.User
		err := h.DB.Where("email = ? AND id != ?", trimmedEmail, userID).First(&existingUser).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "email is already in use"})
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
			return
		}
		updates["email"] = trimmedEmail
	}

	if input.Country != nil { updates["country"] = strings.TrimSpace(*input.Country) }
	if input.City != nil { updates["city"] = strings.TrimSpace(*input.City) }
	if input.Bio != nil { updates["bio"] = strings.TrimSpace(*input.Bio) }
	if input.ProfilePicture != nil { updates["profile_picture"] = strings.TrimSpace(*input.ProfilePicture) }
	if input.AllowFollowers != nil { updates["allow_followers"] = *input.AllowFollowers }

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no valid fields provided for update"})
		return
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update profile"})
		return
	}

	var updatedUser models.User
	if err := h.DB.First(&updatedUser, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch updated profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully", "data": gin.H{"user": updatedUser}})
}

// SearchUsers searches public user profiles by username, bio, city, or country.
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query must be at least 2 characters"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	if limit > 50 { limit = 50 }

	pattern := "%" + query + "%"
	var total int64
	base := h.DB.Model(&models.User{}).Where(
		"username ILIKE ? OR bio ILIKE ? OR city ILIKE ? OR country ILIKE ?",
		pattern, pattern, pattern, pattern,
	)
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	var users []models.User
	err := base.Select("id, username, profile_picture, bio, country, city, status, allow_followers, created_at").
		Order("username ASC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"users": users,
			"pagination": gin.H{
				"page": page, "limit": limit, "total": total,
				"hasMore": int64(page*limit) < total,
			},
		},
	})
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	targetIDParam := c.Param("id")
	targetIDUint, err := strconv.ParseUint(targetIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	targetID := uint(targetIDUint)

	var user models.User
	err = h.DB.Select("id, username, profile_picture, bio, country, city, allow_followers, created_at").First(&user, targetID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": user}})
}
