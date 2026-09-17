package main

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func linkOwnedContent(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := channelID(c)
        if !hasPermission(db, c, id, "content") {
            return
        }
        var in struct {
            PostID uint64 `json:"postId"`
            Status string `json:"status"`
        }
        if err := c.ShouldBindJSON(&in); err != nil || in.PostID == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"message": "postId is required"})
            return
        }
        var ch Channel
        if err := db.Select("id,owner_id,status").First(&ch, id).Error; err != nil || ch.Status != ChannelActive {
            c.JSON(http.StatusNotFound, gin.H{"message": "channel not found"})
            return
        }
        var post struct {
            ID uint64
            UserID uint64
            Visibility string
        }
        if err := db.Table("posts").Select("id,user_id,visibility").Where("id = ?", in.PostID).First(&post).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
            return
        }
        if post.UserID != ch.OwnerID {
            c.JSON(http.StatusForbidden, gin.H{"message": "channel content must belong to the channel owner"})
            return
        }
        status := strings.ToUpper(strings.TrimSpace(in.Status))
        if status == "" {
            status = "PUBLISHED"
        }
        if status != "PUBLISHED" && status != "DRAFT" && status != "ARCHIVED" {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid content status"})
            return
        }
        row := ChannelContent{ChannelID: id, PostID: in.PostID, Status: status, CreatedBy: userID(c)}
        if err := db.Create(&row).Error; err != nil {
            c.JSON(http.StatusConflict, gin.H{"message": "post is already linked or invalid"})
            return
        }
        c.JSON(http.StatusCreated, row)
    }
}
