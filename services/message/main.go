package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	messageTypeText     = "TEXT"
	messageTypeCall     = "CALL_INVITE"
	messageTypeSystem   = "SYSTEM"
	conversationDirect  = "DIRECT"
	conversationGroup   = "GROUP"
)

type Config struct {
	Port       string
	DatabaseURL string
	JWTSecret  string
	FrontendURL string
	AppEnv     string
	LiveKitURL string
	LiveKitKey string
	LiveKitSecret string
}

func loadConfig() (Config, error) {
	c := Config{Port: env("PORT", "8080"), AppEnv: env("APP_ENV", "development"), FrontendURL: os.Getenv("FRONTEND_URL"), LiveKitURL: os.Getenv("LIVEKIT_URL"), LiveKitKey: os.Getenv("LIVEKIT_API_KEY"), LiveKitSecret: os.Getenv("LIVEKIT_API_SECRET")}
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	c.JWTSecret = os.Getenv("JWT_SECRET")
	if c.DatabaseURL == "" || c.JWTSecret == "" { return c, errors.New("DATABASE_URL and JWT_SECRET are required") }
	return c, nil
}
func env(k, d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }

type Conversation struct {
	ID string `gorm:"primaryKey;size:32" json:"id"`
	Type string `gorm:"size:16;not null" json:"type"`
	Title string `gorm:"size:160" json:"title"`
	CreatedBy string `gorm:"size:128;not null;index" json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Members []ConversationMember `json:"-"`
}
type ConversationMember struct {
	ID uint `gorm:"primaryKey"`
	ConversationID string `gorm:"size:32;not null;uniqueIndex:ux_conv_member"`
	UserID string `gorm:"size:128;not null;uniqueIndex:ux_conv_member;index"`
	JoinedAt time.Time `json:"joinedAt"`
}
type Message struct {
	ID string `gorm:"primaryKey;size:32" json:"id"`
	ConversationID string `gorm:"size:32;not null;index" json:"conversationId"`
	SenderID string `gorm:"size:128;not null;index" json:"senderId"`
	Type string `gorm:"size:24;not null" json:"type"`
	Body string `gorm:"type:text" json:"body"`
	ClientID string `gorm:"size:64;index" json:"clientId"`
	CreatedAt time.Time `json:"createdAt"`
}
type ReadReceipt struct { ID uint `gorm:"primaryKey"`; MessageID string `gorm:"size:32;not null;uniqueIndex:ux_read"`; UserID string `gorm:"size:128;not null;uniqueIndex:ux_read"`; ReadAt time.Time `json:"readAt"` }
type Call struct { ID string `gorm:"primaryKey;size:32" json:"id"`; ConversationID string `gorm:"size:32;not null;index"`; RoomName string `gorm:"size:160;not null;uniqueIndex" json:"roomName"`; CreatedBy string `gorm:"size:128;not null" json:"createdBy"`; Status string `gorm:"size:16;not null" json:"status"`; CreatedAt time.Time `json:"createdAt"`; EndedAt *time.Time `json:"endedAt,omitempty"` }

func newID() string { b:=make([]byte,16); _,_ = rand.Read(b); return hex.EncodeToString(b) }

func userIDFromJWT(c *gin.Context, secret string) (string, bool) {
	token := ""
	if h:=c.GetHeader("Authorization"); strings.HasPrefix(h,"Bearer ") { token=strings.TrimSpace(strings.TrimPrefix(h,"Bearer ")) }
	if token=="" { token,_=c.Cookie("auth_token") }
	if token=="" { return "",false }
	parsed,err:=jwt.Parse(token, func(t *jwt.Token)(any,error){ if t.Method != jwt.SigningMethodHS256 { return nil, errors.New("unexpected signing method") }; return []byte(secret),nil }, jwt.WithIssuer("notell-api"))
	if err!=nil || !parsed.Valid { return "",false }
	claims,ok:=parsed.Claims.(jwt.MapClaims); if !ok { return "",false }
	for _,k:=range []string{"userId","user_id","sub"} { if v,ok:=claims[k].(string); ok && v!="" { return v,true } }
	return "",false
}
func auth(secret string) gin.HandlerFunc { return func(c *gin.Context){ id,ok:=userIDFromJWT(c,secret); if !ok { c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }; c.Set("userID",id); c.Next() } }
func uid(c *gin.Context) string { v,_:=c.Get("userID"); return v.(string) }

var upgrader = websocket.Upgrader{ReadBufferSize:4096,WriteBufferSize:4096,CheckOrigin:func(r *http.Request)bool{return true}}
type hub struct { mu sync.RWMutex; clients map[string]map[*websocket.Conn]struct{} }
func newHub()*hub{return &hub{clients:map[string]map[*websocket.Conn]struct{}{}}}
func(h *hub) add(conv string,ws *websocket.Conn){h.mu.Lock();defer h.mu.Unlock();if h.clients[conv]==nil{h.clients[conv]=map[*websocket.Conn]struct{}{}};h.clients[conv][ws]=struct{}{}}
func(h *hub) remove(conv string,ws *websocket.Conn){h.mu.Lock();defer h.mu.Unlock();if m:=h.clients[conv];m!=nil{delete(m,ws);if len(m)==0{delete(h.clients,conv)}}}
func(h *hub) broadcast(conv string,v any){h.mu.RLock();defer h.mu.RUnlock();for ws:=range h.clients[conv]{_ = ws.WriteJSON(v)}}

func member(db *gorm.DB, conv,user string) bool { var m ConversationMember; return db.Where("conversation_id = ? AND user_id = ?",conv,user).First(&m).Error==nil }
func requireMember(db *gorm.DB,c *gin.Context,conv string)bool{if !member(db,conv,uid(c)){c.JSON(http.StatusForbidden,gin.H{"error":"conversation membership required"});return false};return true}

func main(){
	cfg,err:=loadConfig();if err!=nil{log.Fatal(err)}
	db,err:=gorm.Open(postgres.Open(cfg.DatabaseURL),&gorm.Config{});if err!=nil{log.Fatal(err)}
	if err=db.AutoMigrate(&Conversation{},&ConversationMember{},&Message{},&ReadReceipt{},&Call{});err!=nil{log.Fatal(err)}
	h:=newHub();r:=gin.New();r.Use(gin.Logger(),gin.Recovery())
	if cfg.FrontendURL!="" { co:=cors.DefaultConfig();co.AllowOrigins=[]string{cfg.FrontendURL};co.AllowCredentials=true;co.AllowHeaders=[]string{"Origin","Content-Type","Authorization"};r.Use(cors.New(co)) }
	r.GET("/health",func(c *gin.Context){c.JSON(200,gin.H{"status":"ok","service":"message"})})
	r.GET("/ready",func(c *gin.Context){sqlDB,e:=db.DB();if e!=nil||sqlDB.Ping()!=nil{c.JSON(503,gin.H{"status":"not_ready"});return};c.JSON(200,gin.H{"status":"ready"})})
	api:=r.Group("/v1",auth(cfg.JWTSecret))
	api.POST("/conversations",func(c *gin.Context){
		var in struct{Type string `json:"type"`;Title string `json:"title"`;MemberIDs []string `json:"memberIds"`};if c.BindJSON(&in)!=nil{c.JSON(400,gin.H{"error":"invalid request"});return}
		if in.Type!=""&&in.Type!=conversationDirect&&in.Type!=conversationGroup{c.JSON(400,gin.H{"error":"invalid conversation type"});return};if in.Type==""{in.Type=conversationDirect}
		members:=map[string]bool{uid(c):true};for _,m:=range in.MemberIDs{if m!=""{members[m]=true}};if in.Type==conversationDirect&&len(members)!=2{c.JSON(400,gin.H{"error":"direct conversation requires two users"});return}
		conv:=Conversation{ID:newID(),Type:in.Type,Title:in.Title,CreatedBy:uid(c)};tx:=db.Begin();if err:=tx.Create(&conv).Error;err!=nil{tx.Rollback();c.JSON(500,gin.H{"error":"create conversation failed"});return};for m:=range members{if err:=tx.Create(&ConversationMember{ConversationID:conv.ID,UserID:m,JoinedAt:time.Now()}).Error;err!=nil{tx.Rollback();c.JSON(500,gin.H{"error":"add member failed"});return}};if err:=tx.Commit().Error;err!=nil{c.JSON(500,gin.H{"error":"commit failed"});return};c.JSON(201,conv)
	})
	api.GET("/conversations",func(c *gin.Context){var ms []ConversationMember;if db.Where("user_id = ?",uid(c)).Find(&ms).Error!=nil{c.JSON(500,gin.H{"error":"list failed"});return};ids:=make([]string,0,len(ms));for _,m:=range ms{ids=append(ids,m.ConversationID)};var cs []Conversation;if len(ids)>0{db.Where("id IN ?",ids).Order("updated_at DESC").Find(&cs)};c.JSON(200,cs)})
	api.GET("/conversations/:id/messages",func(c *gin.Context){id:=c.Param("id");if !requireMember(db,c,id){return};limit,_:=strconv.Atoi(c.DefaultQuery("limit","50"));if limit<1||limit>100{limit=50};var ms []Message;db.Where("conversation_id = ?",id).Order("created_at DESC").Limit(limit).Find(&ms);for i,j:=0,len(ms)-1;i<j;i,j=i+1,j-1{ms[i],ms[j]=ms[j],ms[i]};c.JSON(200,ms)})
	api.POST("/conversations/:id/read",func(c *gin.Context){id:=c.Param("id");if !requireMember(db,c,id){return};var in struct{MessageID string `json:"messageId"`};if c.BindJSON(&in)!=nil||in.MessageID==""{c.JSON(400,gin.H{"error":"messageId required"});return};var m Message;if db.Where("id = ? AND conversation_id = ?",in.MessageID,id).First(&m).Error!=nil{c.JSON(404,gin.H{"error":"message not found"});return};db.Where("message_id=? AND user_id=?",in.MessageID,uid(c)).Assign(ReadReceipt{MessageID:in.MessageID,UserID:uid(c),ReadAt:time.Now()}).FirstOrCreate(&ReadReceipt{});h.broadcast(id,gin.H{"type":"read","messageId":in.MessageID,"userId":uid(c),"readAt":time.Now()});c.JSON(200,gin.H{"ok":true})})
	api.POST("/conversations/:id/calls",func(c *gin.Context){
		id:=c.Param("id");if !requireMember(db,c,id){return};if cfg.LiveKitURL==""||cfg.LiveKitKey==""||cfg.LiveKitSecret==""{c.JSON(503,gin.H{"error":"live calls are not configured"});return}
		call:=Call{ID:newID(),ConversationID:id,RoomName:"notell-call-"+newID(),CreatedBy:uid(c),Status:"ACTIVE",CreatedAt:time.Now()};if err:=db.Create(&call).Error;err!=nil{c.JSON(500,gin.H{"error":"create call failed"});return};token:=lksdk.NewAccessToken(cfg.LiveKitKey,cfg.LiveKitSecret).SetIdentity(uid(c)).SetName(uid(c)).SetValidFor(2*time.Hour);grant:=&lksdk.VideoGrant{RoomJoin:true,Room:call.RoomName};token.SetVideoGrant(grant);jwtToken,err:=token.ToJWT();if err!=nil{c.JSON(500,gin.H{"error":"create call token failed"});return};h.broadcast(id,gin.H{"type":"call_invite","callId":call.ID,"roomName":call.RoomName,"from":uid(c)});c.JSON(201,gin.H{"call":call,"wsUrl":cfg.LiveKitURL,"token":jwtToken})
	})
	api.POST("/calls/:id/token",func(c *gin.Context){var call Call;if db.First(&call,"id = ?",c.Param("id")).Error!=nil{c.JSON(404,gin.H{"error":"call not found"});return};if !member(db,call.ConversationID,uid(c)){c.JSON(403,gin.H{"error":"conversation membership required"});return};if cfg.LiveKitURL==""||cfg.LiveKitKey==""||cfg.LiveKitSecret==""{c.JSON(503,gin.H{"error":"live calls are not configured"});return};token:=lksdk.NewAccessToken(cfg.LiveKitKey,cfg.LiveKitSecret).SetIdentity(uid(c)).SetValidFor(2*time.Hour);token.SetVideoGrant(&lksdk.VideoGrant{RoomJoin:true,Room:call.RoomName});jwtToken,err:=token.ToJWT();if err!=nil{c.JSON(500,gin.H{"error":"token failed"});return};c.JSON(200,gin.H{"wsUrl":cfg.LiveKitURL,"token":jwtToken,"roomName":call.RoomName})})
	api.GET("/ws/:conversationId",func(c *gin.Context){conv:=c.Param("conversationId");user,ok:=userIDFromJWT(c,cfg.JWTSecret);if !ok||!member(db,conv,user){c.JSON(403,gin.H{"error":"conversation membership required"});return};ws,err:=upgrader.Upgrade(c.Writer,c.Request,nil);if err!=nil{return};h.add(conv,ws);defer func(){h.remove(conv,ws);ws.Close()}();_ = ws.WriteJSON(gin.H{"type":"connected","conversationId":conv,"userId":user})
		for {var in struct{Type string `json:"type"`;Body string `json:"body"`;ClientID string `json:"clientId"`;MessageID string `json:"messageId"`};err:=ws.ReadJSON(&in);if err!=nil{return};switch in.Type{
			case "message": if strings.TrimSpace(in.Body)==""{continue};m:=Message{ID:newID(),ConversationID:conv,SenderID:user,Type:messageTypeText,Body:in.Body,ClientID:in.ClientID,CreatedAt:time.Now()};if db.Create(&m).Error==nil{db.Model(&Conversation{}).Where("id=?",conv).Update("updated_at",m.CreatedAt);h.broadcast(conv,gin.H{"type":"message","message":m})}
			case "typing": h.broadcast(conv,gin.H{"type":"typing","userId":user,"typing":true})
			case "stop_typing": h.broadcast(conv,gin.H{"type":"typing","userId":user,"typing":false})
			case "read": if in.MessageID!=""{db.Where("message_id=? AND user_id=?",in.MessageID,user).Assign(ReadReceipt{MessageID:in.MessageID,UserID:user,ReadAt:time.Now()}).FirstOrCreate(&ReadReceipt{});h.broadcast(conv,gin.H{"type":"read","messageId":in.MessageID,"userId":user,"readAt":time.Now()})}
		}}
	})
	log.Printf("message service listening on :%s",cfg.Port);log.Fatal(r.Run(":"+cfg.Port))
}

var _ = fmt.Sprintf
