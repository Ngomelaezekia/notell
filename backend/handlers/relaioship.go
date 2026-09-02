package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RelationshipHandler struct { DB *gorm.DB }
func NewRelationshipHandler(db *gorm.DB) *RelationshipHandler { return &RelationshipHandler{DB: db} }

func (h *RelationshipHandler) FollowUser(c *gin.Context) {
	followerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	followingID := uint(targetID)
	if followerID == followingID { c.JSON(http.StatusBadRequest, gin.H{"message":"you cannot follow yourself"}); return }

	var targetUser models.User
	if err := h.DB.Select("id", "allow_followers").First(&targetUser, followingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { c.JSON(http.StatusNotFound, gin.H{"message":"user not found"}); return }
		c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to find user"}); return
	}
	if !targetUser.AllowFollowers { c.JSON(http.StatusForbidden, gin.H{"message":"this user does not allow followers"}); return }

	var existing models.Relationship
	err = h.DB.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existing).Error
	if err == nil { c.JSON(http.StatusConflict, gin.H{"message":"already following this user"}); return }
	if !errors.Is(err, gorm.ErrRecordNotFound) { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to check relationship"}); return }

	relationship := models.Relationship{FollowerID:followerID, FollowingID:followingID, Status:"accepted"}
	if err := h.DB.Create(&relationship).Error; err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to follow user"}); return }
	_ = CreateNotification(h.DB, followingID, followerID, "follow", nil, nil)
	c.JSON(http.StatusOK, gin.H{"message":"successfully followed user", "data":gin.H{"following":true, "status":"accepted"}})
}

func (h *RelationshipHandler) UnfollowUser(c *gin.Context) {
	followerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	result := h.DB.Where("follower_id = ? AND following_id = ?", followerID, uint(targetID)).Delete(&models.Relationship{})
	if result.Error != nil { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to unfollow user"}); return }
	if result.RowsAffected == 0 { c.JSON(http.StatusNotFound, gin.H{"message":"not following this user"}); return }
	c.JSON(http.StatusOK, gin.H{"message":"successfully unfollowed user", "data":gin.H{"following":false}})
}

func (h *RelationshipHandler) GetRelationshipStatus(c *gin.Context) {
	viewerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	id := uint(targetID)

	var target models.User
	if err := h.DB.Select("id", "allow_followers").First(&target, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { c.JSON(http.StatusNotFound, gin.H{"message":"user not found"}); return }
		c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to find user"}); return
	}

	var followingCount, followerCount int64
	h.DB.Model(&models.Relationship{}).Where("following_id = ? AND status = ?", id, "accepted").Count(&followerCount)
	h.DB.Model(&models.Relationship{}).Where("follower_id = ? AND status = ?", id, "accepted").Count(&followingCount)

	following := false
	follower := false
	if viewerID != id {
		var relation models.Relationship
		if err := h.DB.Where("follower_id = ? AND following_id = ? AND status = ?", viewerID, id, "accepted").First(&relation).Error; err == nil { following = true }
		if err := h.DB.Where("follower_id = ? AND following_id = ? AND status = ?", id, viewerID, "accepted").First(&relation).Error; err == nil { follower = true }
	}

	c.JSON(http.StatusOK, gin.H{"data":gin.H{
		"following": following,
		"follower": follower,
		"followerCount": followerCount,
		"followingCount": followingCount,
		"allowFollowers": target.AllowFollowers,
		"isSelf": viewerID == id,
	}})
}

func (h *RelationshipHandler) GetFollowers(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	var followers []models.User
	err = h.DB.Table("users").Select("users.id, users.username, users.profile_picture, users.bio, users.country, users.city, users.status").Joins("JOIN user_relationships ON user_relationships.follower_id = users.id").Where("user_relationships.following_id = ? AND user_relationships.status = ?", uint(userID), "accepted").Scan(&followers).Error
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to fetch followers"}); return }
	c.JSON(http.StatusOK, gin.H{"data":followers})
}

func (h *RelationshipHandler) GetFollowing(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	var following []models.User
	err = h.DB.Table("users").Select("users.id, users.username, users.profile_picture, users.bio, users.country, users.city, users.status").Joins("JOIN user_relationships ON user_relationships.following_id = users.id").Where("user_relationships.follower_id = ? AND user_relationships.status = ?", uint(userID), "accepted").Scan(&following).Error
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to fetch following"}); return }
	c.JSON(http.StatusOK, gin.H{"data":following})
}

func (h *RelationshipHandler) RemoveFollower(c *gin.Context) {
	myID := c.MustGet("userId").(uint)
	followerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"message":"invalid user ID"}); return }
	result := h.DB.Where("follower_id = ? AND following_id = ?", uint(followerID), myID).Delete(&models.Relationship{})
	if result.Error != nil { c.JSON(http.StatusInternalServerError, gin.H{"message":"failed to remove follower"}); return }
	if result.RowsAffected == 0 { c.JSON(http.StatusNotFound, gin.H{"message":"user is not following you"}); return }
	c.JSON(http.StatusOK, gin.H{"message":"follower removed successfully"})
}
