package main

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type premiumChannelCreateInput struct {
    Name          string      `json:"name"`
    Description   string      `json:"description"`
    Category      string      `json:"category"`
    Type          ChannelType `json:"type"`
    AvatarMediaID string      `json:"avatarMediaId"`
    BannerMediaID string      `json:"bannerMediaId"`
}

func createPremiumChannel(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        planCode, _ := c.Get("channelPlanCode")
        code, ok := planCode.(string)
        if !ok || strings.TrimSpace(code) == "" {
            c.JSON(http.StatusServiceUnavailable, gin.H{"message": "channel plan could not be resolved"})
            return
        }
        code = strings.TrimSpace(code)
        var plan ChannelPlan
        if err := db.Where("code = ? AND active = ?", code, true).First(&plan).Error; err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{"message": "channel plan is not available"})
            return
        }

        var in premiumChannelCreateInput
        if err := c.ShouldBindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
            return
        }
        in.Name = strings.TrimSpace(in.Name)
        if len(in.Name) < 2 || len(in.Name) > 100 {
            c.JSON(http.StatusBadRequest, gin.H{"message": "name must be between 2 and 100 characters"})
            return
        }
        if in.Type == "" {
            in.Type = ChannelPublic
        }
        if in.Type != ChannelPublic && in.Type != ChannelPrivate {
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid channel type"})
            return
        }
        if in.Type == ChannelPrivate && !plan.PrivateAllowed {
            c.JSON(http.StatusPaymentRequired, gin.H{"message": "private channels require a plan that allows private access"})
            return
        }
        slugBase := makeSlug(in.Name)
        if slugBase == "" {
            c.JSON(http.StatusBadRequest, gin.H{"message": "name cannot produce a valid slug"})
            return
        }

        ownerID := userID(c)
        for attempt := 0; attempt < 8; attempt++ {
            slug := slugBase
            if attempt > 0 {
                slug = fmt.Sprintf("%s-%d", slugBase, attempt+1)
            }
            ch := Channel{OwnerID: ownerID, Name: in.Name, Slug: slug, Description: strings.TrimSpace(in.Description), Category: strings.TrimSpace(in.Category), Type: in.Type, Status: ChannelActive, AvatarMediaID: strings.TrimSpace(in.AvatarMediaID), BannerMediaID: strings.TrimSpace(in.BannerMediaID)}
            err := db.Transaction(func(tx *gorm.DB) error {
                if err := tx.Create(&ch).Error; err != nil {
                    return err
                }
                if err := tx.Create(&ChannelMember{ChannelID: ch.ID, UserID: ownerID, MembershipType: MembershipTeam, Role: RoleOwner, Status: "ACTIVE"}).Error; err != nil {
                    return err
                }
                return tx.Create(&ChannelEntitlement{ChannelID: ch.ID, PlanCode: plan.Code, Status: "ACTIVE", PeriodEndsAt: nil}).Error
            })
            if err == nil {
                c.JSON(http.StatusCreated, ch)
                return
            }
            lower := strings.ToLower(err.Error())
            if strings.Contains(lower, "unique") || strings.Contains(lower, "duplicate") {
                continue
            }
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create channel"})
            return
        }
        c.JSON(http.StatusConflict, gin.H{"message": "could not allocate a unique channel slug"})
    }
}
