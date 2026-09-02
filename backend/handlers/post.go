package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostHandler struct { DB *gorm.DB }
func NewPostHandler(db *gorm.DB) *PostHandler { return &PostHandler{DB: db} }

type createPostInput struct { ContentType string `json:"contentType" binding:"required,oneof=image video"`; ContentURL string `json:"contentUrl" binding:"required,url"`; Caption string `json:"caption"` }
type createCommentInput struct { Content string `json:"content" binding:"required,max=2000"`; ParentID *uint `json:"parentId"` }

const postEngagementSelect = `posts.*, (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS like_count, EXISTS(SELECT 1 FROM likes WHERE likes.post_id = posts.id AND likes.user_id = ?) AS liked`

func (h *PostHandler) CreatePost(c *gin.Context) {
	authUserID, exists := c.Get("userId"); if !exists { c.JSON(http.StatusUnauthorized, gin.H{"message":"unauthorized"}); return }; userID := authUserID.(uint)
	var input createPostInput; if err:=c.ShouldBindJSON(&input);err!=nil { c.JSON(http.StatusBadRequest,gin.H{"message":err.Error()}); return }
	post:=models.Post{UserID:userID,ContentType:input.ContentType,ContentURL:input.ContentURL,Caption:strings.TrimSpace(input.Caption)}
	if err:=h.DB.Create(&post).Error;err!=nil { c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to create post"});return }
	h.DB.Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&post,post.ID)
	post.LikeCount = 0; post.Liked = false
	c.JSON(http.StatusCreated,gin.H{"message":"post created successfully","data":post})
}

func (h *PostHandler) GetFeed(c *gin.Context) {
	userID:=c.MustGet("userId").(uint); page,_:=strconv.Atoi(c.DefaultQuery("page","1")); limit,_:=strconv.Atoi(c.DefaultQuery("limit","10")); if page<1{page=1};if limit<1||limit>50{limit=10}
	var posts []models.Post
	err:=h.DB.Model(&models.Post{}).Select(postEngagementSelect,userID).Joins(`LEFT JOIN user_relationships ON user_relationships.following_id = posts.user_id AND user_relationships.follower_id = ? AND user_relationships.status = ?`, userID, "accepted").Where(`posts.user_id = ? OR user_relationships.follower_id IS NOT NULL`, userID).Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Order("posts.created_at DESC").Order("posts.id DESC").Limit(limit+1).Offset((page-1)*limit).Find(&posts).Error
	if err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to fetch feed"});return}
	hasMore:=len(posts)>limit
	if hasMore { posts=posts[:limit] }
	c.JSON(http.StatusOK,gin.H{"data":posts,"pagination":gin.H{"page":page,"limit":limit,"hasMore":hasMore}})
}

// SearchPosts searches captions and author usernames. Results prioritize exact
// username matches, username prefixes, caption prefixes, then recency.
func (h *PostHandler) SearchPosts(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "search query must be at least 2 characters"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	if limit > 50 { limit = 50 }

	pattern := "%" + query + "%"
	prefix := query + "%"
	var total int64
	base := h.DB.Model(&models.Post{}).
		Joins("JOIN users ON users.id = posts.user_id").
		Where("posts.caption ILIKE ? OR users.username ILIKE ?", pattern, pattern)
	if err := base.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	var posts []models.Post
	authUserID := uint(0)
	if value, ok := c.Get("userId"); ok {
		if id, ok := value.(uint); ok { authUserID = id }
	}
	err := base.Select(postEngagementSelect, authUserID).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "profile_picture") }).
		Order(gorm.Expr(`CASE
			WHEN LOWER(users.username) = LOWER(?) THEN 0
			WHEN LOWER(users.username) LIKE LOWER(?) THEN 1
			WHEN LOWER(posts.caption) LIKE LOWER(?) THEN 2
			ELSE 3
		END`, query, prefix, prefix)).
		Order("posts.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&posts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"posts": posts,
			"pagination": gin.H{
				"page": page, "limit": limit, "total": total,
				"hasMore": int64(page*limit) < total,
			},
		},
	})
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	id,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return};var post models.Post
	userID:=uint(0);if value,ok:=c.Get("userId");ok{if id,ok:=value.(uint);ok{userID=id}}
	err=h.DB.Model(&models.Post{}).Select(postEngagementSelect,userID).Where("posts.id = ?",uint(id)).Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&post).Error
	if err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusNotFound,gin.H{"message":"post not found"});return};c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to fetch post"});return};c.JSON(http.StatusOK,gin.H{"data":post})
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	userID:=c.MustGet("userId").(uint);postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return}
	result:=h.DB.Where("id=? AND user_id=?",uint(postID),userID).Delete(&models.Post{});if result.Error!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to delete post"});return};if result.RowsAffected==0{c.JSON(http.StatusNotFound,gin.H{"message":"post not found"});return};c.JSON(http.StatusOK,gin.H{"message":"post deleted successfully"})
}

func (h *PostHandler) ToggleLike(c *gin.Context) {
	userID:=c.MustGet("userId").(uint);postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return}
	var post models.Post;if err:=h.DB.First(&post,uint(postID)).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusNotFound,gin.H{"message":"post not found"})}else{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to find post"})};return}
	var like models.Like;err=h.DB.Where("user_id=? AND post_id=?",userID,uint(postID)).First(&like).Error
	liked:=false
	if err==nil{
		if err:=h.DB.Delete(&like).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to remove like"});return}
		h.DB.Where("user_id = ? AND actor_id = ? AND post_id = ? AND type = ?", post.UserID, userID, uint(postID), "like").Delete(&models.Notification{})
	}
	if errors.Is(err,gorm.ErrRecordNotFound){
		if err:=h.DB.Create(&models.Like{UserID:userID,PostID:uint(postID)}).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to like post"});return}
		liked=true
		_ = CreateNotification(h.DB, post.UserID, userID, "like", func() *uint { id:=uint(postID); return &id }(), nil)
	}
	var likeCount int64; h.DB.Model(&models.Like{}).Where("post_id=?",uint(postID)).Count(&likeCount)
	c.JSON(http.StatusOK,gin.H{"liked":liked,"likeCount":likeCount})
}

func (h *PostHandler) AddComment(c *gin.Context) {
	userID:=c.MustGet("userId").(uint);postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return}
	var input createCommentInput;if err:=c.ShouldBindJSON(&input);err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":err.Error()});return}
	var post models.Post;if err:=h.DB.Select("id,user_id").First(&post,uint(postID)).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusNotFound,gin.H{"message":"post not found"});return};c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to find post"});return}
	content:=strings.TrimSpace(input.Content);if content==""{c.JSON(http.StatusBadRequest,gin.H{"message":"comment cannot be empty"});return}
	if input.ParentID!=nil{var parent models.Comment;if err:=h.DB.Select("id,post_id").First(&parent,*input.ParentID).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusBadRequest,gin.H{"message":"parent comment not found"});return};c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to validate parent comment"});return};if parent.PostID!=uint(postID){c.JSON(http.StatusBadRequest,gin.H{"message":"parent comment belongs to another post"});return}}
	comment:=models.Comment{PostID:uint(postID),UserID:userID,Content:content,ParentID:input.ParentID};if err:=h.DB.Create(&comment).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to add comment"});return}
	if err:=h.DB.Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&comment,comment.ID).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to load comment"});return}
	targetUserID:=post.UserID;if input.ParentID!=nil{var parent models.Comment;if h.DB.Select("user_id").First(&parent,*input.ParentID).Error==nil{targetUserID=parent.UserID}}
	_ = CreateNotification(h.DB,targetUserID,userID,"comment",func()*uint{id:=uint(postID);return &id}(),&comment.ID)
	c.JSON(http.StatusCreated,gin.H{"data":comment})
}

func (h *PostHandler) GetComments(c *gin.Context) {
	postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return}
	var comments []models.Comment;if err:=h.DB.Where("post_id=? AND parent_id IS NULL",uint(postID)).Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Preload("Replies",func(db *gorm.DB)*gorm.DB{return db.Order("created_at ASC")}).Preload("Replies.User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Order("created_at ASC").Find(&comments).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to fetch comments"});return};c.JSON(http.StatusOK,gin.H{"data":comments})
}
