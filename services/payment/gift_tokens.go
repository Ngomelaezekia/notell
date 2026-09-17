package main

import (
    "crypto/rand"
    "errors"
    "fmt"
    "strings"
    "time"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

const (
    GiftAdminFull       = "ADMIN_FULL"
    GiftSocialMedia     = "SOCIAL_MEDIA"
    GiftUserAffiliation = "USER_AFFILIATION"
)

type GiftToken struct {
    ID                string     `gorm:"primaryKey" json:"id"`
    Code              string     `gorm:"size:6;uniqueIndex;not null" json:"code"`
    Type              string     `gorm:"size:32;not null;index" json:"type"`
    DiscountPercent   int        `gorm:"not null" json:"discountPercent"`
    MaxRedemptions    int        `gorm:"not null;default:1" json:"maxRedemptions"`
    RedeemedCount     int        `gorm:"not null;default:0" json:"redeemedCount"`
    CreatedByUserID   string     `gorm:"index" json:"createdByUserId,omitempty"`
    ProviderCouponID  string     `gorm:"uniqueIndex" json:"-"`
    Active            bool       `gorm:"not null;default:true;index" json:"active"`
    ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
    CreatedAt         time.Time  `json:"createdAt"`
    UpdatedAt         time.Time  `json:"updatedAt"`
}

type GiftClaim struct {
    ID           string     `gorm:"primaryKey" json:"id"`
    GiftTokenID  string     `gorm:"not null;index" json:"giftTokenId"`
    UserID       string     `gorm:"not null;index" json:"userId"`
    ClaimedAt    time.Time  `json:"claimedAt"`
    UsedAt       *time.Time `json:"usedAt,omitempty"`
    SubscriptionID string   `json:"subscriptionId,omitempty"`
}

type GiftRedemption struct {
    ID           string    `gorm:"primaryKey" json:"id"`
    GiftTokenID  string    `gorm:"not null;index" json:"giftTokenId"`
    UserID       string    `gorm:"not null;index" json:"userId"`
    SubscriptionID string  `gorm:"not null;index" json:"subscriptionId"`
    RedeemedAt   time.Time `json:"redeemedAt"`
}

func giftDiscountPercent(tokenType string) (int, error) {
    switch strings.ToUpper(strings.TrimSpace(tokenType)) {
    case GiftAdminFull:
        return 100, nil
    case GiftSocialMedia:
        return 40, nil
    case GiftUserAffiliation:
        return 50, nil
    default:
        return 0, errors.New("unsupported gift token type")
    }
}

func normalizeGiftCode(value string) string {
    return strings.ToUpper(strings.TrimSpace(value))
}

func generateGiftCode() (string, error) {
    const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
    raw := make([]byte, 6)
    if _, err := rand.Read(raw); err != nil {
        return "", fmt.Errorf("generate gift code: %w", err)
    }
    code := make([]byte, len(raw))
    for i, b := range raw {
        code[i] = alphabet[int(b)%len(alphabet)]
    }
    return string(code), nil
}

func giftTokenAvailable(token GiftToken, now time.Time) bool {
    if !token.Active || token.RedeemedCount >= token.MaxRedemptions {
        return false
    }
    return token.ExpiresAt == nil || token.ExpiresAt.After(now)
}

func loadGiftTokenLocked(tx *gorm.DB, code string) (GiftToken, error) {
    var token GiftToken
    err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", normalizeGiftCode(code)).First(&token).Error
    return token, err
}

func claimGiftToken(db *gorm.DB, userID, code string) (GiftToken, error) {
    var token GiftToken
    err := db.Transaction(func(tx *gorm.DB) error {
        var err error
        token, err = loadGiftTokenLocked(tx, code)
        if err != nil {
            return err
        }
        if !giftTokenAvailable(token, time.Now().UTC()) {
            return errors.New("gift code is inactive, expired, or fully redeemed")
        }
        var redemption GiftRedemption
        if err := tx.Where("gift_token_id = ? AND user_id = ?", token.ID, userID).First(&redemption).Error; err == nil {
            return errors.New("gift code has already been redeemed by this user")
        } else if !errors.Is(err, gorm.ErrRecordNotFound) {
            return err
        }
        var claim GiftClaim
        err = tx.Where("gift_token_id = ? AND user_id = ?", token.ID, userID).First(&claim).Error
        if errors.Is(err, gorm.ErrRecordNotFound) {
            claim = GiftClaim{ID: newID("gcl"), GiftTokenID: token.ID, UserID: userID, ClaimedAt: time.Now().UTC()}
            if err := tx.Create(&claim).Error; err != nil {
                return err
            }
        } else if err != nil {
            return err
        }
        return nil
    })
    return token, err
}

func firstChannelSubscriptionExists(db *gorm.DB, userID string) (bool, error) {
    var count int64
    if err := db.Model(&Subscription{}).Where("user_id = ? AND resource_type = ? AND resource_id = ? AND status <> ?", userID, "platform", "channel", "incomplete").Count(&count).Error; err != nil {
        return false, err
    }
    return count > 0, nil
}

func claimedGiftForUser(db *gorm.DB, userID string) (GiftToken, bool, error) {
    var token GiftToken
    err := db.Table("gift_tokens").
        Joins("JOIN gift_claims ON gift_claims.gift_token_id = gift_tokens.id").
        Where("gift_claims.user_id = ? AND gift_claims.used_at IS NULL", userID).
        Order("gift_claims.claimed_at DESC").First(&token).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return GiftToken{}, false, nil
    }
    if err != nil {
        return GiftToken{}, false, err
    }
    if !giftTokenAvailable(token, time.Now().UTC()) {
        return GiftToken{}, false, nil
    }
    return token, true, nil
}

func redeemGiftToken(db *gorm.DB, tokenID, userID, subscriptionID string) error {
    return db.Transaction(func(tx *gorm.DB) error {
        var token GiftToken
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", tokenID).First(&token).Error; err != nil {
            return err
        }
        var existing GiftRedemption
        if err := tx.Where("gift_token_id = ? AND user_id = ?", tokenID, userID).First(&existing).Error; err == nil {
            return nil
        } else if !errors.Is(err, gorm.ErrRecordNotFound) {
            return err
        }
        if !giftTokenAvailable(token, time.Now().UTC()) {
            return errors.New("gift code is no longer available")
        }
        redemption := GiftRedemption{ID: newID("grd"), GiftTokenID: tokenID, UserID: userID, SubscriptionID: subscriptionID, RedeemedAt: time.Now().UTC()}
        if err := tx.Create(&redemption).Error; err != nil {
            return err
        }
        if err := tx.Model(&token).Update("redeemed_count", gorm.Expr("redeemed_count + 1")).Error; err != nil {
            return err
        }
        now := time.Now().UTC()
        if err := tx.Model(&GiftClaim{}).Where("gift_token_id = ? AND user_id = ? AND used_at IS NULL", tokenID, userID).Updates(map[string]any{"used_at": now, "subscription_id": subscriptionID}).Error; err != nil {
            return err
        }
        return nil
    })
}
