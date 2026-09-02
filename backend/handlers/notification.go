package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const notificationActorSelect = "id, username, profile_picture"

// CreateNotification records an in-app notification. Notification failures
// should not break the action that triggered them, so callers can ignore the error.
func CreateNotification(db *gorm.DB, recipientID, actorID uint, notificationType string, postID, commentID *uint) error {
	if recipientID == 0 || actorID == 0 || recipientID == actorID {
		return nil
	}
	return db.Create(&models.Notification{
		UserID: recipientID,
		ActorID: actorID,
		Type: notificationType,
		PostID: postID,
		CommentID: commentID,
	}).Error
}

type NotificationHandler struct {
	DB *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{DB: db}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	var total int64
	if err := h.DB.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count notifications"})
		return
	}

	var unreadCount int64
	if err := h.DB.Model(&models.Notification{}).Where("user_id = ? AND read = ?", userID, false).Count(&unreadCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count unread notifications"})
		return
	}

	var notifications []models.Notification
	err := h.DB.Where("user_id = ?", userID).
		Preload("Actor", func(db *gorm.DB) *gorm.DB { return db.Select(notificationActorSelect) }).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"notifications": notifications,
			"pagination": gin.H{
				"page": page,
				"limit": limit,
				"total": total,
				"hasMore": int64(page*limit) < total,
			},
			"unreadCount": unreadCount,
		},
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid notification ID"})
		return
	}

	result := h.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", uint(notificationID), userID).
		Update("read", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to mark notification as read"})
		return
	}
	if result.RowsAffected == 0 {
		var notification models.Notification
		if err := h.DB.Where("id = ?", uint(notificationID)).First(&notification).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "notification not found"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("userId").(uint)
	if err := h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to mark notifications as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notifications marked as read"})
}
