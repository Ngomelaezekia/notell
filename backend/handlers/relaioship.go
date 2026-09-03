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

type RelationshipHandler struct {
	DB *gorm.DB
}

func NewRelationshipHandler(db *gorm.DB) *RelationshipHandler {
	return &RelationshipHandler{DB: db}
}

const relationshipStatus = "accepted"

func parsePageLimit(c *gin.Context, defaultLimit int) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > 50 {
		limit = 50
	}
	return page, limit
}

func (h *RelationshipHandler) FollowUser(c *gin.Context) {
	followerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	followingID := uint(targetID)
	if followerID == followingID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "you cannot follow yourself"})
		return
	}

	var targetUser models.User
	if err := h.DB.Select("id", "allow_followers").First(&targetUser, followingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to find user"})
		return
	}
	if !targetUser.AllowFollowers {
		c.JSON(http.StatusForbidden, gin.H{"message": "this user does not allow followers"})
		return
	}

	relationship := models.Relationship{FollowerID: followerID, FollowingID: followingID, Status: relationshipStatus}
	result := h.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "follower_id"}, {Name: "following_id"}}, DoNothing: true}).Create(&relationship)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to follow user"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "already following this user"})
		return
	}

	_ = CreateNotification(h.DB, followingID, followerID, "follow", nil, nil)
	c.JSON(http.StatusOK, gin.H{"message": "successfully followed user", "data": gin.H{"following": true, "status": relationshipStatus}})
}

func (h *RelationshipHandler) UnfollowUser(c *gin.Context) {
	followerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	result := h.DB.Where("follower_id = ? AND following_id = ?", followerID, uint(targetID)).Delete(&models.Relationship{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to unfollow user"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "not following this user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully unfollowed user", "data": gin.H{"following": false}})
}

func (h *RelationshipHandler) GetRelationshipStatus(c *gin.Context) {
	viewerID := c.MustGet("userId").(uint)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	id := uint(targetID)

	var target models.User
	if err := h.DB.Select("id", "allow_followers").First(&target, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to find user"})
		return
	}

	var followingCount, followerCount int64
	if err := h.DB.Model(&models.Relationship{}).Where("following_id = ? AND status = ?", id, relationshipStatus).Count(&followerCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count followers"})
		return
	}
	if err := h.DB.Model(&models.Relationship{}).Where("follower_id = ? AND status = ?", id, relationshipStatus).Count(&followingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count following"})
		return
	}

	following := false
	follower := false
	if viewerID != id {
		var relation models.Relationship
		err := h.DB.Where("follower_id = ? AND following_id = ? AND status = ?", viewerID, id, relationshipStatus).First(&relation).Error
		if err == nil {
			following = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load relationship"})
			return
		}

		err = h.DB.Where("follower_id = ? AND following_id = ? AND status = ?", id, viewerID, relationshipStatus).First(&relation).Error
		if err == nil {
			follower = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load relationship"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"following": following,
		"follower": follower,
		"followerCount": followerCount,
		"followingCount": followingCount,
		"allowFollowers": target.AllowFollowers,
		"isSelf": viewerID == id,
	}})
}

func (h *RelationshipHandler) listUsers(c *gin.Context, following bool) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	page, limit := parsePageLimit(c, 20)
	offset := (page - 1) * limit

	base := h.DB.Table("users").
		Select("users.id, users.username, users.profile_picture, users.bio, users.country, users.city, users.status").
		Joins("JOIN user_relationships ON user_relationships." + map[bool]string{true: "following_id", false: "follower_id"}[following] + " = users.id").
		Where("user_relationships." + map[bool]string{true: "follower_id", false: "following_id"}[following] + " = ? AND user_relationships.status = ?", uint(userID), relationshipStatus)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count relationships"})
		return
	}

	var users []models.User
	if err := base.Order("users.username ASC").Order("users.id ASC").Offset(offset).Limit(limit).Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch relationships"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users, "pagination": gin.H{
		"page": page,
		"limit": limit,
		"total": total,
		"hasMore": int64(page*limit) < total,
	}})
}

func (h *RelationshipHandler) GetFollowers(c *gin.Context) {
	h.listUsers(c, true)
}

func (h *RelationshipHandler) GetFollowing(c *gin.Context) {
	h.listUsers(c, false)
}

func (h *RelationshipHandler) RemoveFollower(c *gin.Context) {
	myID := c.MustGet("userId").(uint)
	followerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
	result := h.DB.Where("follower_id = ? AND following_id = ?", uint(followerID), myID).Delete(&models.Relationship{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to remove follower"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "user is not following you"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "follower removed successfully"})
}
