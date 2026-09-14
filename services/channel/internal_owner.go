package main

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func internalChannelOwnerCheck(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !internalServiceAuthorized(c) {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "trusted service authentication required"})
            return
        }
        channelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
        if err != nil || channelID == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid channel id"})
            return
        }
        userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
        if err != nil || userID == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
            return
        }
        var ch Channel
        if err := db.First(&ch, channelID).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"message": "channel not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load channel"})
            return
        }
        if ch.Status != ChannelActive {
            c.JSON(http.StatusNotFound, gin.H{"message": "channel not found"})
            return
        }
        allowed := ch.OwnerID == userID
        c.JSON(http.StatusOK, gin.H{
            "channelId": channelID,
            "userId": userID,
            "ownerId": ch.OwnerID,
            "allowed": allowed,
            "active": true,
            "source": strings.TrimSpace("channel_owner"),
        })
    }
}
