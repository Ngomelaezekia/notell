package main

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// ChannelContent.PostID is an external reference to a post owned by the main
// Notell service. The channel service has its own database, so it must not
// query the main service's posts table directly.
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
