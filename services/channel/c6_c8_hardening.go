package main

import (
    "strings"
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func registerC6C8Hardened(db *gorm.DB, p *gin.RouterGroup) {
    p.GET("/channels/:id/live", listLive(db)); p.POST("/channels/:id/live", hardenedCreateLive(db)); p.PATCH("/channels/:id/live/:sessionId", hardenedUpdateLive(db)); p.POST("/channels/:id/live/:sessionId/start", hardenedStartLive(db)); p.POST("/channels/:id/live/:sessionId/end", hardenedEndLive(db)); p.DELETE("/channels/:id/live/:sessionId", hardenedCancelLive(db))
    p.GET("/channels/:id/media", listMediaRefs(db)); p.POST("/channels/:id/media", addMediaRef(db)); p.DELETE("/channels/:id/media/:mediaId", removeMediaRef(db)); p.GET("/channels/:id/playback/:mediaId", hardenedPlaybackAuthorization(db))
    p.POST("/channels/:id/reports", hardenedReportCase(db)); p.GET("/channels/:id/safety/cases", hardenedListCases(db)); p.PATCH("/channels/:id/safety/cases/:caseId", hardenedUpdateCase(db)); p.POST("/channels/:id/rights", setRights(db)); p.GET("/channels/:id/rights", listRights(db))
    p.GET("/channels/:id/monetization", getEligibility(db)); p.PUT("/channels/:id/monetization", setEligibility(db)); p.GET("/channels/:id/revenue", listRevenue(db)); p.POST("/channels/:id/ads", createAd(db)); p.GET("/channels/:id/ads", listAds(db)); p.GET("/channels/:id/analytics", analytics(db))
    p.POST("/channels/:id/revenue", func(c *gin.Context){c.JSON(403,gin.H{"message":"revenue events must be ingested by a trusted service"})})
    p.POST("/internal/channels/:id/revenue", internalRevenueIngest(db))
    p.GET("/internal/channels/:id/entitlements/:userId", internalEntitlementCheck(db))
}

func internalServiceAuthorized(c *gin.Context) bool { expected:=strings.TrimSpace(getenv("CHANNEL_INTERNAL_SERVICE_KEY","")); supplied:=strings.TrimSpace(c.GetHeader("X-Channel-Service-Key")); return expected!="" && supplied!="" && supplied==expected }

func internalEntitlementCheck(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    if !internalServiceAuthorized(c){c.JSON(401,gin.H{"message":"trusted service authentication required"});return}
    channel:=channelID(c); var ch Channel
    if db.First(&ch,channel).Error!=nil || ch.Status!=ChannelActive {c.JSON(404,gin.H{"message":"channel not found"});return}
    var m ChannelMember
    err:=db.Where("channel_id=? AND user_id=? AND membership_type=? AND status=?",channel,c.Param("userId"),MembershipSubscriber,"ACTIVE").First(&m).Error
    allowed:=err==nil
    if err!=nil && err!=gorm.ErrRecordNotFound {c.JSON(500,gin.H{"message":"failed to check entitlement"});return}
    c.JSON(200,gin.H{"channelId":channel,"userId":c.Param("userId"),"entitled":allowed,"source":"channel_membership","membershipType":MembershipSubscriber})
} }

func hardenedPlaybackAuthorization(db *gorm.DB) gin.HandlerFunc { return func(c *gin.Context) {
    var x MediaReference
    if db.Where("channel_id=? AND media_id=? AND status=?",c.Param("id"),c.Param("mediaId"),"ACTIVE").First(&x).Error!=nil {c.JSON(404,gin.H{"message":"media not found"});return}
    if x.Access=="PRIVATE" {
        var m ChannelMember
        err:=db.Where("channel_id=? AND user_id=? AND membership_type=? AND status=?",x.ChannelID,userID(c),MembershipSubscriber,"ACTIVE").First(&m).Error
        if err==nil {c.JSON(200,gin.H{"allowed":true,"mediaId":x.MediaID,"access":"PRIVATE","entitlement":"subscriber"});return}
        if err!=gorm.ErrRecordNotFound {c.JSON(500,gin.H{"message":"failed to verify playback entitlement"});return}
        c.JSON(200,gin.H{"allowed":false,"reason":"subscriber_entitlement_required","mediaId":x.MediaID});return
    }
    c.JSON(200,gin.H{"allowed":true,"mediaId":x.MediaID,"access":"PUBLIC"})
} }

func hardenedChannelPermission(db *gorm.DB,c *gin.Context,permission string)(Channel,bool){id:=channelID(c);var ch Channel;if db.First(&ch,id).Error!=nil||ch.Status==ChannelDisabled{c.JSON(404,gin.H{"message":"channel not found"});return ch,false};if !hasPermission(db,c,id,permission){return ch,false};return ch,true}
func hardenedCreateLive(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){ch,ok:=hardenedChannelPermission(db,c,"live");if !ok{return};var in struct{Title string `json:"title"`;PlaybackRef string `json:"playbackRef"`};if c.ShouldBindJSON(&in)!=nil||strings.TrimSpace(in.Title)==""{c.JSON(400,gin.H{"message":"title is required"});return};x:=LiveSession{ChannelID:ch.ID,OwnerID:ch.OwnerID,Title:strings.TrimSpace(in.Title),State:LiveScheduled,PlaybackRef:strings.TrimSpace(in.PlaybackRef)};if err:=db.Create(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to create live session"});return};c.JSON(201,x)}}
func hardenedUpdateLive(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){ch,ok:=hardenedChannelPermission(db,c,"live");if !ok{return};var x LiveSession;if db.Where("id=? AND channel_id=?",c.Param("sessionId"),ch.ID).First(&x).Error!=nil{c.JSON(404,gin.H{"message":"live session not found"});return};if x.State!=LiveScheduled{c.JSON(409,gin.H{"message":"live session is no longer editable"});return};var in struct{Title string `json:"title"`;PlaybackRef string `json:"playbackRef"`};if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"message":"invalid request"});return};if strings.TrimSpace(in.Title)!=""{x.Title=strings.TrimSpace(in.Title)};if in.PlaybackRef!=""{x.PlaybackRef=strings.TrimSpace(in.PlaybackRef)};if err:=db.Save(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to update live session"});return};c.JSON(200,x)}}
func hardenedTransitionLive(db *gorm.DB,target LiveState)gin.HandlerFunc{return func(c *gin.Context){ch,ok:=hardenedChannelPermission(db,c,"live");if !ok{return};var x LiveSession;if db.Where("id=? AND channel_id=?",c.Param("sessionId"),ch.ID).First(&x).Error!=nil{c.JSON(404,gin.H{"message":"live session not found"});return};now:=time.Now().UTC();switch target{case LiveLive:if x.State!=LiveScheduled{c.JSON(409,gin.H{"message":"only scheduled sessions can start"});return};x.State=LiveLive;x.StartedAt=&now;case LiveEnded:if x.State!=LiveLive{c.JSON(409,gin.H{"message":"only live sessions can end"});return};x.State=LiveEnded;x.EndedAt=&now;case LiveCancelled:if x.State!=LiveScheduled{c.JSON(409,gin.H{"message":"only scheduled sessions can be cancelled"});return};x.State=LiveCancelled;default:c.JSON(400,gin.H{"message":"invalid live transition"});return};if err:=db.Save(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to update live state"});return};c.JSON(200,x)}}
func hardenedStartLive(db *gorm.DB)gin.HandlerFunc{return hardenedTransitionLive(db,LiveLive)}
func hardenedEndLive(db *gorm.DB)gin.HandlerFunc{return hardenedTransitionLive(db,LiveEnded)}
func hardenedCancelLive(db *gorm.DB)gin.HandlerFunc{return hardenedTransitionLive(db,LiveCancelled)}
func hardenedReportCase(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){var in struct{TargetType string `json:"targetType"`;TargetID string `json:"targetId"`;Reason string `json:"reason"`;Notes string `json:"notes"`};if c.ShouldBindJSON(&in)!=nil||strings.TrimSpace(in.TargetType)==""||strings.TrimSpace(in.TargetID)==""||strings.TrimSpace(in.Reason)==""{c.JSON(400,gin.H{"message":"targetType, targetId and reason are required"});return};x:=SafetyCase{ChannelID:channelID(c),ReporterID:userID(c),TargetType:strings.TrimSpace(in.TargetType),TargetID:strings.TrimSpace(in.TargetID),Reason:strings.TrimSpace(in.Reason),Notes:strings.TrimSpace(in.Notes),Status:"OPEN"};if err:=db.Create(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to create report"});return};c.JSON(201,x)}}
func hardenedListCases(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){if _,ok:=hardenedChannelPermission(db,c,"moderate");!ok{return};var x []SafetyCase;if err:=db.Where("channel_id=?",channelID(c)).Order("created_at DESC").Limit(100).Find(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to list safety cases"});return};c.JSON(200,gin.H{"cases":x})}}
func hardenedUpdateCase(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){ch,ok:=hardenedChannelPermission(db,c,"moderate");if !ok{return};var x SafetyCase;if db.Where("id=? AND channel_id=?",c.Param("caseId"),ch.ID).First(&x).Error!=nil{c.JSON(404,gin.H{"message":"case not found"});return};var in struct{Status string `json:"status"`;Notes string `json:"notes"`};if c.ShouldBindJSON(&in)!=nil{c.JSON(400,gin.H{"message":"invalid request"});return};switch in.Status{case "OPEN","REVIEWING","RESOLVED","REJECTED":default:c.JSON(400,gin.H{"message":"invalid status"});return};x.Status=in.Status;x.Notes=strings.TrimSpace(in.Notes);if err:=db.Save(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to update safety case"});return};c.JSON(200,x)}}
func internalRevenueIngest(db *gorm.DB)gin.HandlerFunc{return func(c *gin.Context){if !internalServiceAuthorized(c){c.JSON(401,gin.H{"message":"trusted service authentication required"});return};channel:=channelID(c);var ch Channel;if db.First(&ch,channel).Error!=nil||ch.Status==ChannelDisabled{c.JSON(404,gin.H{"message":"channel not found"});return};var in struct{UserID uint64 `json:"userId"`;EventType string `json:"eventType"`;ReferenceID string `json:"referenceId"`;Currency string `json:"currency"`;GrossCents int64 `json:"grossCents"`};if c.ShouldBindJSON(&in)!=nil||strings.TrimSpace(in.EventType)==""||in.GrossCents<=0{c.JSON(400,gin.H{"message":"eventType and positive grossCents are required"});return};in.ReferenceID=strings.TrimSpace(in.ReferenceID);if in.ReferenceID==""{c.JSON(400,gin.H{"message":"referenceId is required"});return};var existing RevenueEvent;err:=db.Where("channel_id=? AND reference_id=?",channel,in.ReferenceID).First(&existing).Error;if err==nil{c.JSON(409,gin.H{"message":"revenue reference already exists"});return};if err!=gorm.ErrRecordNotFound{c.JSON(500,gin.H{"message":"failed to check revenue reference"});return};currency:=strings.ToUpper(strings.TrimSpace(in.Currency));if currency==""{currency="USD"};fee:=in.GrossCents/5;x:=RevenueEvent{ChannelID:channel,UserID:in.UserID,EventType:strings.TrimSpace(in.EventType),ReferenceID:in.ReferenceID,GrossCents:in.GrossCents,PlatformFeeCents:fee,CreatorCents:in.GrossCents-fee,Currency:currency,Status:"PENDING"};if err:=db.Create(&x).Error;err!=nil{c.JSON(500,gin.H{"message":"failed to record revenue event"});return};c.JSON(201,x)}}
