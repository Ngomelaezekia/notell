package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RelationshipHandler struct {
	DB *gorm.DB
}

func NewRelationshipHandler(db *gorm.DB) *RelationshipHandler {
	return &RelationshipHandler{
		DB: db,
	}
}

// ===============================
// Follow User
// ===============================
func (h *RelationshipHandler) FollowUser(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	followerID := userID.(uint)

	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	followingID := uint(targetID)

	// Prevent self follow
	if followerID == followingID {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "you cannot follow yourself",
		})
		return
	}

	// Check target user
	var targetUser models.User

	err = h.DB.
		Select("id", "allow_followers").
		First(&targetUser, followingID).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if !targetUser.AllowFollowers {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "this user does not allow followers",
		})
		return
	}

	// Check existing relationship
	var existing models.Relationship

	err = h.DB.
		Where(
			"follower_id = ? AND following_id = ?",
			followerID,
			followingID,
		).
		First(&existing).
		Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"message": "already following this user",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	relationship := models.Relationship{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      "accepted",
	}

	if err := h.DB.Create(&relationship).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to follow user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully followed user",
		"data":    relationship,
	})
}

// ===============================
// Unfollow User
// ===============================
func (h *RelationshipHandler) UnfollowUser(c *gin.Context) {

	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	followerID := userID.(uint)

	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	result := h.DB.
		Where(
			"follower_id = ? AND following_id = ?",
			followerID,
			uint(targetID),
		).
		Delete(&models.Relationship{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "not following this user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully unfollowed user",
	})
}

// ===============================
// Followers
// ===============================
func (h *RelationshipHandler) GetFollowers(c *gin.Context) {

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	var followers []models.User

	err = h.DB.
		Table("users").
		Select(`
			users.id,
			users.username,
			users.profile_picture,
			users.bio,
			users.country,
			users.city,
			users.status
		`).
		Joins(`
			JOIN user_relationships
			ON user_relationships.follower_id = users.id
		`).
		Where(`
			user_relationships.following_id = ?
			AND user_relationships.status = ?
		`,
			uint(userID),
			"accepted",
		).
		Scan(&followers).
		Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": followers,
	})
}

// ===============================
// Following
// ===============================
func (h *RelationshipHandler) GetFollowing(c *gin.Context) {

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	var following []models.User

	err = h.DB.
		Table("users").
		Select(`
			users.id,
			users.username,
			users.profile_picture,
			users.bio,
			users.country,
			users.city,
			users.status
		`).
		Joins(`
			JOIN user_relationships
			ON user_relationships.following_id = users.id
		`).
		Where(`
			user_relationships.follower_id = ?
			AND user_relationships.status = ?
		`,
			uint(userID),
			"accepted",
		).
		Scan(&following).
		Error

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": following,
	})
}

// ===============================
// Remove Follower
// ===============================
func (h *RelationshipHandler) RemoveFollower(c *gin.Context) {

	userID, exists := c.Get("userId")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	myID := userID.(uint)

	followerID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	result := h.DB.
		Where(
			"follower_id = ? AND following_id = ?",
			uint(followerID),
			myID,
		).
		Delete(&models.Relationship{})

	if result.Error != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": result.Error.Error(),
		})

		return
	}

	if result.RowsAffected == 0 {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user is not following you",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "follower removed successfully",
	})
}
