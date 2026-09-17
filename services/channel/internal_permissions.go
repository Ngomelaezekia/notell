package main

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func internalLivePermission(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !internalServiceAuthorized(c) {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "trusted service authentication required"})
            return
        }
        channelID := channelID(c)
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
        var member ChannelTeamMember
        err := db.Where("channel_id=? AND user_id=? AND status=?", channelID, c.Param("userId"), "ACTIVE").First(&member).Error
        allowed := ch.OwnerID == userIDValue(c.Param("userId"))
        if !allowed && err == nil {
            allowed = rolePermissions[member.Role]["live"]
        }
        if err != nil && err != gorm.ErrRecordNotFound {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to check channel permission"})
            return
        }
        role := ""
        if ch.OwnerID == userIDValue(c.Param("userId")) {
            role = string(RoleOwner)
        } else if err == nil {
            role = string(member.Role)
        }
        c.JSON(http.StatusOK, gin.H{"channelId": channelID, "userId": c.Param("userId"), "allowed": allowed, "active": true, "role": role, "permission": "live"})
    }
}

func userIDValue(value string) uint64 {
    var out uint64
    for _, r := range strings.TrimSpace(value) {
        if r < '0' || r > '9' {
            return 0
        }
        out = out*10 + uint64(r-'0')
    }
    return out
}
