package main

import (
 "log"
 "net/http"
 "github.com/gin-contrib/cors"
 "github.com/gin-gonic/gin"
 "gorm.io/driver/postgres"
 "gorm.io/gorm"
)

func main(){
 cfg:=loadConfig()
 if cfg.DatabaseURL==""{log.Fatal("DATABASE_URL is required")}
 if cfg.JWTSecret==""{log.Fatal("JWT_SECRET is required")}
 db,err:=gorm.Open(postgres.Open(cfg.DatabaseURL),&gorm.Config{})
 if err!=nil{log.Fatalf("database connection failed: %v",err)}
 if err:=db.AutoMigrate(&Channel{},&ChannelMember{},&ChannelPlan{},&ChannelEntitlement{},&ChannelTeamMember{},&ChannelInvitation{},&ChannelProgram{},&ChannelSchedule{},&ChannelContent{},&ChannelSetting{},&LiveSession{},&MediaReference{},&SafetyCase{},&RightsRecord{},&MonetizationEligibility{},&RevenueEvent{},&AdCampaign{},&ChannelMetric{});err!=nil{log.Fatalf("migration failed: %v",err)}
 if err:=repairMembershipIndex(db);err!=nil{log.Fatalf("membership index repair failed: %v",err)}
 seedPlans(db)
 r:=gin.New();r.Use(gin.Logger(),gin.Recovery())
 cc:=cors.DefaultConfig();cc.AllowCredentials=true;cc.AllowHeaders=[]string{"Origin","Content-Type","Accept","Authorization"}
 if cfg.FrontendURL!=""{cc.AllowOrigins=[]string{cfg.FrontendURL}}else if cfg.Environment!="production"{cc.AllowAllOrigins=true;cc.AllowCredentials=false}else{log.Fatal("FRONTEND_URL is required in production")}
 r.Use(cors.New(cc))
 r.GET("/health",func(c *gin.Context){c.JSON(200,gin.H{"status":"ok","service":"channel"})})
 r.GET("/ready",func(c *gin.Context){sqlDB,e:=db.DB();if e!=nil||sqlDB.Ping()!=nil{c.JSON(503,gin.H{"status":"not_ready"});return};c.JSON(200,gin.H{"status":"ready","service":"channel"})})
 v1:=r.Group("/v1")
 v1.GET("/channels",listChannels(db));v1.GET("/channels/:id",getChannel(db))
 p:=v1.Group("",auth(cfg.JWTSecret))
 p.POST("/channels",createChannel(db));p.PATCH("/channels/:id",updateChannel(db));p.DELETE("/channels/:id",deleteChannel(db))
 p.POST("/channels/:id/follow",followChannel(db,MembershipFollower));p.DELETE("/channels/:id/follow",unfollowChannel(db));p.POST("/channels/:id/subscribe",followChannel(db,MembershipSubscriber));p.DELETE("/channels/:id/subscribe",unsubscribeChannel(db))
 p.GET("/channels/:id/members",listMembers(db));p.POST("/channels/:id/members",addMember(db));p.DELETE("/channels/:id/members/:userId",removeMember(db))
 p.GET("/channels/:id/team",listTeam(db));p.POST("/channels/:id/team/invitations",inviteTeam(db));p.POST("/channels/:id/team/invitations/:invitationId/accept",acceptTeamInvite(db));p.PATCH("/channels/:id/team/:userId",updateTeam(db));p.DELETE("/channels/:id/team/:userId",removeTeam(db))
 p.GET("/channels/:id/studio/programs",programs(db));p.POST("/channels/:id/studio/programs",createProgram(db));p.PATCH("/channels/:id/studio/programs/:programId",updateProgram(db));p.DELETE("/channels/:id/studio/programs/:programId",deleteProgram(db));p.GET("/channels/:id/studio/schedule",schedules(db));p.POST("/channels/:id/studio/schedule",createSchedule(db));p.PATCH("/channels/:id/studio/schedule/:scheduleId",updateSchedule(db));p.DELETE("/channels/:id/studio/schedule/:scheduleId",deleteSchedule(db));p.GET("/channels/:id/studio/content",content(db));p.POST("/channels/:id/studio/content",linkContent(db));p.PATCH("/channels/:id/studio/content/:contentId",updateContent(db));p.DELETE("/channels/:id/studio/content/:contentId",unlinkContent(db));p.GET("/channels/:id/studio/audience",audience(db));p.GET("/channels/:id/studio/settings",studioSettings(db));p.PUT("/channels/:id/studio/settings/:key",setStudioSetting(db))
 registerC6C8(db,p)
 if err:=http.ListenAndServe(":"+cfg.Port,r);err!=nil{log.Fatal(err)}
}
