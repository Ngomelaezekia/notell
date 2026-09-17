package main

import (
    "errors"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type giftTokenResponse struct {
    GiftToken
    ShareURL string `json:"shareUrl"`
}

func giftAdminKeyAuthorized(c *gin.Context) bool {
    expected := strings.TrimSpace(os.Getenv("GIFT_ADMIN_KEY"))
    supplied := strings.TrimSpace(c.GetHeader("X-Gift-Admin-Key"))
    return expected != "" && supplied != "" && supplied == expected
}

func giftShareURL(code string) string {
    base := strings.TrimRight(strings.TrimSpace(getenv("PUBLIC_APP_URL", getenv("FRONTEND_URL", "http://localhost:5173"))), "/")
    return base + "/auth?mode=signup&gift=" + url.QueryEscape(code)
}

func createGiftToken(db *gorm.DB, tokenType, createdBy string, maxRedemptions int, expiresAt *time.Time) (GiftToken, error) {
    discount, err := giftDiscountPercent(tokenType)
    if err != nil {
        return GiftToken{}, err
    }
    tokenType = strings.ToUpper(strings.TrimSpace(tokenType))
    if maxRedemptions <= 0 {
        switch tokenType {
        case GiftAdminFull:
            maxRedemptions = 1
        case GiftSocialMedia:
            maxRedemptions = 100
        default:
            maxRedemptions = 25
        }
    }
    if maxRedemptions > 1000 {
        maxRedemptions = 1000
    }
    if tokenType == GiftAdminFull {
        var count int64
        if err := db.Model(&GiftToken{}).Where("type = ?", GiftAdminFull).Count(&count).Error; err != nil {
            return GiftToken{}, err
        }
        if count >= 2 {
            return GiftToken{}, errors.New("only two admin gift tokens are allowed")
        }
        maxRedemptions = 1
    }
    for attempt := 0; attempt < 8; attempt++ {
        code, err := generateGiftCode()
        if err != nil {
            return GiftToken{}, err
        }
        token := GiftToken{ID: newID("gft"), Code: code, Type: tokenType, DiscountPercent: discount, MaxRedemptions: maxRedemptions, CreatedByUserID: createdBy, ExpiresAt: expiresAt, Active: true}
        if err := db.Create(&token).Error; err != nil {
            if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
                continue
            }
            return GiftToken{}, err
        }
        return token, nil
    }
    return GiftToken{}, errors.New("failed to generate a unique gift token")
}

func registerGiftRoutes(r *gin.Engine, s *Server) {
    api := r.Group("/v1")
    api.Use(s.auth)

    api.POST("/gift-tokens", func(c *gin.Context) {
        var in struct {
            Type           string `json:"type" binding:"required"`
            MaxRedemptions int    `json:"maxRedemptions"`
            ExpiresAt      *time.Time `json:"expiresAt"`
        }
        if err := c.ShouldBindJSON(&in); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
            return
        }
        tokenType := strings.ToUpper(strings.TrimSpace(in.Type))
        if tokenType != GiftUserAffiliation && !giftAdminKeyAuthorized(c) {
            c.JSON(http.StatusForbidden, gin.H{"error": "gift token type requires admin authorization"})
            return
        }
        token, err := createGiftToken(s.db, tokenType, c.GetString("user_id"), in.MaxRedemptions, in.ExpiresAt)
        if err != nil {
            status := http.StatusBadRequest
            if strings.Contains(err.Error(), "only two") {
                status = http.StatusConflict
            }
            c.JSON(status, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusCreated, giftTokenResponse{GiftToken: token, ShareURL: giftShareURL(token.Code)})
    })

    api.POST("/gift-tokens/admin/bootstrap", func(c *gin.Context) {
        if !giftAdminKeyAuthorized(c) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "gift admin authorization required"})
            return
        }
        var existing int64
        if err := s.db.Model(&GiftToken{}).Where("type = ?", GiftAdminFull).Count(&existing).Error; err != nil {
            c.JSON(500, gin.H{"error": "failed to inspect admin gift tokens"})
            return
        }
        created := make([]giftTokenResponse, 0, 2-int(existing))
        for existing < 2 {
            token, err := createGiftToken(s.db, GiftAdminFull, "system", 1, nil)
            if err != nil {
                c.JSON(500, gin.H{"error": err.Error()})
                return
            }
            created = append(created, giftTokenResponse{GiftToken: token, ShareURL: giftShareURL(token.Code)})
            existing++
        }
        c.JSON(http.StatusCreated, gin.H{"tokens": created})
    })

    api.GET("/gift-tokens/mine", func(c *gin.Context) {
        uid := c.GetString("user_id")
        var rows []GiftToken
        if err := s.db.Where("created_by_user_id = ?", uid).Order("created_at DESC").Limit(100).Find(&rows).Error; err != nil {
            c.JSON(500, gin.H{"error": "failed to load gift tokens"})
            return
        }
        out := make([]giftTokenResponse, 0, len(rows))
        for _, token := range rows {
            out = append(out, giftTokenResponse{GiftToken: token, ShareURL: giftShareURL(token.Code)})
        }
        c.JSON(200, gin.H{"tokens": out})
    })

    api.POST("/gift-tokens/claim", func(c *gin.Context) {
        var in struct { Code string `json:"code" binding:"required"` }
        if err := c.ShouldBindJSON(&in); err != nil {
            c.JSON(400, gin.H{"error": "gift code is required"})
            return
        }
        token, err := claimGiftToken(s.db, c.GetString("user_id"), in.Code)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "inactive") || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "redeemed") {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }
            c.JSON(500, gin.H{"error": "gift code could not be claimed"})
            return
        }
        c.JSON(200, giftTokenResponse{GiftToken: token, ShareURL: giftShareURL(token.Code)})
    })

    r.GET("/v1/internal/gift-claims", func(c *gin.Context) {
        key := strings.TrimSpace(c.GetHeader("X-Internal-Service-Key"))
        expected := strings.TrimSpace(os.Getenv("INTERNAL_SERVICE_KEY"))
        if key == "" || expected == "" || key != expected {
            c.JSON(401, gin.H{"error": "unauthorized"})
            return
        }
        userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
        if userID == "" {
            c.JSON(400, gin.H{"error": "X-User-ID is required"})
            return
        }
        token, ok, err := claimedGiftForUser(s.db, userID)
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to load gift claim"})
            return
        }
        if !ok {
            c.JSON(200, gin.H{"claimed": false})
            return
        }
        c.JSON(200, gin.H{"claimed": true, "giftToken": token})
    })

    _ = strconv.Itoa
    _ = clause.Locking{}
}
