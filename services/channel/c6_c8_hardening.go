package main

import (
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// registerC6C8Hardened replaces the vulnerable C6-C8 mutation handlers while
// the original handlers remain source-compatible. All sensitive mutations use
// the existing C4 permission model instead of owner-only checks.
func registerC6C8Hardened(db *gorm.DB, p *gin.RouterGroup) {
    p.POST("/channels/:id/live", hardenedCreateLive(db))
    p.PATCH("/channels/:id/live/:sessionId", hardenedUpdateLive(db))
    p.POST("/channels/:id/live/:sessionId/start", hardenedStartLive(db))
    p.POST("/channels/:id/live/:sessionId/end", hardenedEndLive(db))
    p.DELETE("/channels/:id/live/:sessionId", hardenedCancelLive(db))
    p.POST("/channels/:id/reports", hardenedReportCase(db))
    p.GET("/channels/:id/safety/cases", hardenedListCases(db))
    p.PATCH("/channels/:id/safety/cases/:caseId", hardenedUpdateCase(db))
    p.POST("/channels/:id/revenue", hardenedRecordRevenue(db))
}

func hardenedChannelPermission(db *gorm.DB, c *gin.Context, permission string) (Channel, bool) {
    id := channelID(c)
    var ch Channel
    if db.First(&ch, id).Error != nil || ch.Status == ChannelDisabled {
        c.JSON(404, gin.H{"message": "channel not found"})
        return ch, false
    }
    if !hasPermission(db, c, id, permission) {
        return ch, false
    }
    return ch, true
}

func hardenedCreateLive(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    ch, ok := hardenedChannelPermission(db, c, "live"); if !ok { return }
    var in struct { Title string `json:"title"`; PlaybackRef string `json:"playbackRef"` }
    if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Title) == "" { c.JSON(400, gin.H{"message":"title is required"}); return }
    x := LiveSession{ChannelID: ch.ID, OwnerID: ch.OwnerID, Title: strings.TrimSpace(in.Title), State: LiveScheduled, PlaybackRef: strings.TrimSpace(in.PlaybackRef)}
    if err := db.Create(&x).Error; err != nil { c.JSON(500, gin.H{"message":"failed to create live session"}); return }
    c.JSON(201, x)
} }

func hardenedUpdateLive(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    ch, ok := hardenedChannelPermission(db, c, "live"); if !ok { return }
    var x LiveSession
    if db.Where("id=? AND channel_id=?", c.Param("sessionId"), ch.ID).First(&x).Error != nil { c.JSON(404, gin.H{"message":"live session not found"}); return }
    if x.State != LiveScheduled { c.JSON(409, gin.H{"message":"live session is no longer editable"}); return }
    var in struct { Title string `json:"title"`; PlaybackRef string `json:"playbackRef"` }
    if c.ShouldBindJSON(&in) != nil { c.JSON(400, gin.H{"message":"invalid request"}); return }
    if strings.TrimSpace(in.Title) != "" { x.Title = strings.TrimSpace(in.Title) }
    if in.PlaybackRef != "" { x.PlaybackRef = strings.TrimSpace(in.PlaybackRef) }
    if err := db.Save(&x).Error; err != nil { c.JSON(500, gin.H{"message":"failed to update live session"}); return }
    c.JSON(200, x)
} }

func hardenedTransitionLive(db *gorm.DB, target LiveState) gin.HandlerFunc { return func(c *gin.Context) {
    ch, ok := hardenedChannelPermission(db, c, "live"); if !ok { return }
    var x LiveSession
    if db.Where("id=? AND channel_id=?", c.Param("sessionId"), ch.ID).First(&x).Error != nil { c.JSON(404, gin.H{"message":"live session not found"}); return }
    now := time.Now().UTC()
    switch target {
    case LiveLive:
        if x.State != LiveScheduled { c.JSON(409, gin.H{"message":"only scheduled sessions can start"}); return }
        x.State = LiveLive; x.StartedAt = &now
    case LiveEnded:
        if x.State != LiveLive { c.JSON(409, gin.H{"message":"only live sessions can end"}); return }
        x.State = LiveEnded; x.EndedAt = &now
    case LiveCancelled:
        if x.State != LiveScheduled { c.JSON(409, gin.H{"message":"only scheduled sessions can be cancelled"}); return }
        x.State = LiveCancelled
    default:
        c.JSON(400, gin.H{"message":"invalid live transition"}); return
    }
    if err := db.Save(&x).Error; err != nil { c.JSON(500, gin.H{"message":"failed to update live state"}); return }
    c.JSON(200, x)
} }
func hardenedStartLive(db *gorm.DB) gin.HandlerFunc { return hardenedTransitionLive(db, LiveLive) }
func hardenedEndLive(db *gorm.DB) gin.HandlerFunc { return hardenedTransitionLive(db, LiveEnded) }
func hardenedCancelLive(db *gorm.DB) gin.HandlerFunc { return hardenedTransitionLive(db, LiveCancelled) }

func hardenedReportCase(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    var in struct { TargetType string `json:"targetType"`; TargetID string `json:"targetId"`; Reason string `json:"reason"`; Notes string `json:"notes"` }
    if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.TargetType)=="" || strings.TrimSpace(in.TargetID)=="" || strings.TrimSpace(in.Reason)=="" { c.JSON(400, gin.H{"message":"targetType, targetId and reason are required"}); return }
    x := SafetyCase{ChannelID: channelID(c), ReporterID:userID(c), TargetType:strings.TrimSpace(in.TargetType), TargetID:strings.TrimSpace(in.TargetID), Reason:strings.TrimSpace(in.Reason), Notes:strings.TrimSpace(in.Notes), Status:"OPEN"}
    if err:=db.Create(&x).Error; err!=nil { c.JSON(500,gin.H{"message":"failed to create report"}); return }
    c.JSON(201,x)
} }
func hardenedListCases(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    if _,ok:=hardenedChannelPermission(db,c,"moderate"); !ok { return }
    var x []SafetyCase
    if err:=db.Where("channel_id=?",channelID(c)).Order("created_at DESC").Limit(100).Find(&x).Error;err!=nil { c.JSON(500,gin.H{"message":"failed to list safety cases"});return }
    c.JSON(200,gin.H{"cases":x})
} }
func hardenedUpdateCase(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    ch,ok:=hardenedChannelPermission(db,c,"moderate");if !ok{return}
    var x SafetyCase
    if db.Where("id=? AND channel_id=?",c.Param("caseId"),ch.ID).First(&x).Error!=nil { c.JSON(404,gin.H{"message":"case not found"});return }
    var in struct{Status string `json:"status"`;Notes string `json:"notes"`}
    if c.ShouldBindJSON(&in)!=nil {c.JSON(400,gin.H{"message":"invalid request"});return}
    switch in.Status {case "OPEN","REVIEWING","RESOLVED","REJECTED":default:c.JSON(400,gin.H{"message":"invalid status"});return}
    x.Status=in.Status;x.Notes=strings.TrimSpace(in.Notes)
    if err:=db.Save(&x).Error;err!=nil {c.JSON(500,gin.H{"message":"failed to update safety case"});return}
    c.JSON(200,x)
} }

func hardenedRecordRevenue(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    ch,ok:=hardenedChannelPermission(db,c,"channel");if !ok{return}
    var in struct{UserID uint64 `json:"userId"`;EventType string `json:"eventType"`;ReferenceID string `json:"referenceId"`;Currency string `json:"currency"`;GrossCents int64 `json:"grossCents"`}
    if c.ShouldBindJSON(&in)!=nil || strings.TrimSpace(in.EventType)=="" || in.GrossCents<=0 {c.JSON(400,gin.H{"message":"eventType and positive grossCents are required"});return}
    in.ReferenceID=strings.TrimSpace(in.ReferenceID)
    if in.ReferenceID=="" {c.JSON(400,gin.H{"message":"referenceId is required for idempotent revenue events"});return}
    var existing RevenueEvent
    if db.Where("channel_id=? AND reference_id=?",ch.ID,in.ReferenceID).First(&existing).Error==nil {c.JSON(409,gin.H{"message":"revenue reference already exists"});return}
    currency:=strings.ToUpper(strings.TrimSpace(in.Currency));if currency==""{currency="USD"}
    fee:=in.GrossCents/5
    x:=RevenueEvent{ChannelID:ch.ID,UserID:in.UserID,EventType:strings.TrimSpace(in.EventType),ReferenceID:in.ReferenceID,GrossCents:in.GrossCents,PlatformFeeCents:fee,CreatorCents:in.GrossCents-fee,Currency:currency,Status:"PENDING"}
    if err:=db.Create(&x).Error;err!=nil {c.JSON(500,gin.H{"message":"failed to record revenue event"});return}
    c.JSON(201,x)
} }
