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

func (h *PostHandler) CreatePost(c *gin.Context) {
	authUserID, exists := c.Get("userId"); if !exists { c.JSON(http.StatusUnauthorized, gin.H{"message":"unauthorized"}); return }; userID := authUserID.(uint)
	var input createPostInput; if err:=c.ShouldBindJSON(&input); err!=nil { c.JSON(http.StatusBadRequest,gin.H{"message":err.Error()}); return }
	post:=models.Post{UserID:userID,ContentType:input.ContentType,ContentURL:input.ContentURL,Caption:strings.TrimSpace(input.Caption)}
	if err:=h.DB.Create(&post).Error;err!=nil { c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to create post"});return }
	h.DB.Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&post,post.ID)
	c.JSON(http.StatusCreated,gin.H{"message":"post created successfully","data":post})
}

func (h *PostHandler) GetFeed(c *gin.Context) {
	userID:=c.MustGet("userId").(uint); page,_:=strconv.Atoi(c.DefaultQuery("page","1")); limit,_:=strconv.Atoi(c.DefaultQuery("limit","10")); if page<1{page=1};if limit<1||limit>50{limit=10}
	var posts []models.Post; err:=h.DB.Model(&models.Post{}).Joins(`JOIN user_relationships ON user_relationships.following_id = posts.user_id`).Where(`user_relationships.follower_id = ? AND user_relationships.status = ?`,userID,"accepted").Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Order("posts.created_at DESC").Limit(limit).Offset((page-1)*limit).Find(&posts).Error
	if err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to fetch feed"});return};c.JSON(http.StatusOK,gin.H{"data":posts,"page":page,"limit":limit})
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	id,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return};var post models.Post
	err=h.DB.Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&post,uint(id)).Error
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
	if err==nil{if err:=h.DB.Delete(&like).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to remove like"});return};c.JSON(http.StatusOK,gin.H{"liked":false});return}
	if !errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to check like"});return};if err:=h.DB.Create(&models.Like{UserID:userID,PostID:uint(postID)}).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to like post"});return};c.JSON(http.StatusOK,gin.H{"liked":true})
}

func (h *PostHandler) AddComment(c *gin.Context) {
	userID:=c.MustGet("userId").(uint);postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return}
	var input createCommentInput;if err:=c.ShouldBindJSON(&input);err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":err.Error()});return};content:=strings.TrimSpace(input.Content);if content==""{c.JSON(http.StatusBadRequest,gin.H{"message":"comment cannot be empty"});return}
	var post models.Post;if err:=h.DB.First(&post,uint(postID)).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusNotFound,gin.H{"message":"post not found"})}else{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to find post"})};return}
	if input.ParentID!=nil{var parent models.Comment;if err:=h.DB.Where("id=? AND post_id=?",*input.ParentID,uint(postID)).First(&parent).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusBadRequest,gin.H{"message":"parent comment not found for this post"})}else{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to validate parent comment"})};return}}
	comment:=models.Comment{UserID:userID,PostID:uint(postID),Content:content,ParentID:input.ParentID};if err:=h.DB.Create(&comment).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to add comment"});return}
	if err:=h.DB.Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).First(&comment,comment.ID).Error;err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"comment created but failed to load author"});return};c.JSON(http.StatusCreated,gin.H{"message":"comment added","data":comment})
}

func (h *PostHandler) GetComments(c *gin.Context) {
	postID,err:=strconv.ParseUint(c.Param("id"),10,32);if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"message":"invalid post ID"});return};var post models.Post
	if err:=h.DB.First(&post,uint(postID)).Error;err!=nil{if errors.Is(err,gorm.ErrRecordNotFound){c.JSON(http.StatusNotFound,gin.H{"message":"post not found"})}else{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to find post"})};return}
	var comments []models.Comment;err=h.DB.Where("post_id=? AND parent_id IS NULL",uint(postID)).Preload("User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Preload("Replies.User",func(db *gorm.DB)*gorm.DB{return db.Select("id","username","profile_picture")}).Order("created_at ASC").Find(&comments).Error
	if err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"message":"failed to fetch comments"});return};c.JSON(http.StatusOK,gin.H{"data":comments})
}
