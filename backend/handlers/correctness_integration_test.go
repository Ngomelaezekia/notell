package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newCorrectnessIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("NOTELL_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTELL_TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Notification{}); err != nil {
		_ = closePostLifecycleDB(db)
		t.Fatalf("migrate correctness test database: %v", err)
	}

	t.Cleanup(func() {
		if err := closePostLifecycleDB(db); err != nil {
			t.Errorf("close correctness test database: %v", err)
		}
	})
	return db
}

func createCorrectnessUser(t *testing.T, db *gorm.DB, suffix string) models.User {
	t.Helper()
	stamp := time.Now().UnixNano()
	user := models.User{
		Username: fmt.Sprintf("correctness_%s_%d", suffix, stamp),
		Email:    fmt.Sprintf("correctness_%s_%d@example.test", suffix, stamp),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create correctness user: %v", err)
	}
	t.Cleanup(func() { _ = db.Delete(&user).Error })
	return user
}

func createCorrectnessPost(t *testing.T, db *gorm.DB, userID uint) models.Post {
	t.Helper()
	var post models.Post
	// Use SQL for the fixture so the comment tests do not need to manufacture
	// an Upload row and physical media file just to create a post.
	if err := db.Exec(
		"INSERT INTO posts (user_id, content_type, content_url, caption, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		userID, "image", "https://example.test/uploads/correctness.jpg", "correctness", time.Now(), time.Now(),
	).Scan(&post.ID).Error; err != nil {
		t.Fatalf("create correctness post: %v", err)
	}
	post.UserID = userID
	post.ContentType = "image"
	post.ContentURL = "https://example.test/uploads/correctness.jpg"
	return post
}

func invokeAddComment(t *testing.T, h *PostHandler, userID, postID uint, content string, parentID *uint) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/posts/:id/comments", func(c *gin.Context) {
		c.Set("userId", userID)
		h.AddComment(c)
	})

	payload := map[string]interface{}{"content": content}
	if parentID != nil {
		payload["parentId"] = *parentID
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal comment payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%d/comments", postID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAddCommentRejectsMissingPost(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "missing-post")
	h := NewPostHandler(db, "https://example.test")

	w := invokeAddComment(t, h, user.ID, 999999999, "hello", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestAddCommentRejectsParentFromAnotherPost(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "cross-post")
	postA := createCorrectnessPost(t, db, user.ID)
	postB := createCorrectnessPost(t, db, user.ID)

	parent := models.Comment{UserID: user.ID, PostID: postA.ID, Content: "parent"}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("create parent comment: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	w := invokeAddComment(t, h, user.ID, postB.ID, "invalid reply", &parent.ID)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	var count int64
	if err := db.Model(&models.Comment{}).Where("post_id = ?", postB.ID).Count(&count).Error; err != nil {
		t.Fatalf("count post B comments: %v", err)
	}
	if count != 0 {
		t.Fatalf("post B comment count = %d, want 0", count)
	}
}

func TestAddCommentRejectsNestedReply(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "nested")
	post := createCorrectnessPost(t, db, user.ID)

	parent := models.Comment{UserID: user.ID, PostID: post.ID, Content: "parent"}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("create parent comment: %v", err)
	}
	child := models.Comment{UserID: user.ID, PostID: post.ID, Content: "child", ParentID: &parent.ID}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("create child comment: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	w := invokeAddComment(t, h, user.ID, post.ID, "nested reply", &child.ID)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAddCommentRejectsWhitespaceOnlyContent(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "blank")
	post := createCorrectnessPost(t, db, user.ID)
	h := NewPostHandler(db, "https://example.test")

	w := invokeAddComment(t, h, user.ID, post.ID, "   \n\t  ", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAddCommentAcceptsTopLevelReply(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "reply")
	post := createCorrectnessPost(t, db, user.ID)

	parent := models.Comment{UserID: user.ID, PostID: post.ID, Content: "parent"}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("create parent comment: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	w := invokeAddComment(t, h, user.ID, post.ID, "reply", &parent.ID)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var replies []models.Comment
	if err := db.Where("post_id = ? AND parent_id = ?", post.ID, parent.ID).Find(&replies).Error; err != nil {
		t.Fatalf("load reply: %v", err)
	}
	if len(replies) != 1 || replies[0].Content != "reply" {
		t.Fatalf("unexpected replies: %+v", replies)
	}
}

func TestToggleLikeSerializesConcurrentRequests(t *testing.T) {
	db := newCorrectnessIntegrationDB(t)
	user := createCorrectnessUser(t, db, "like-race")
	post := createCorrectnessPost(t, db, user.ID)
	h := NewPostHandler(db, "https://example.test")

	gin.SetMode(gin.TestMode)
	start := make(chan struct{})
	responses := make(chan int, 2)
	liked := make(chan bool, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			r := gin.New()
			r.POST("/posts/:id/like", func(c *gin.Context) {
				c.Set("userId", user.ID)
				h.ToggleLike(c)
			})
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%d/like", post.ID), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			responses <- w.Code
			if w.Code == http.StatusOK {
				var payload struct {
					Liked bool `json:"liked"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Errorf("decode like response: %v", err)
					return
				}
				liked <- payload.Liked
			}
		}()
	}
	close(start)
	wg.Wait()
	close(responses)
	close(liked)

	for code := range responses {
		if code != http.StatusOK {
			t.Fatalf("concurrent like status = %d, want %d", code, http.StatusOK)
		}
	}
	var states []bool
	for state := range liked {
		states = append(states, state)
	}
	if len(states) != 2 || states[0] == states[1] {
		t.Fatalf("expected one like=true and one like=false, got %v", states)
	}

	var count int64
	if err := db.Model(&models.Like{}).Where("post_id = ? AND user_id = ?", post.ID, user.ID).Count(&count).Error; err != nil {
		t.Fatalf("count final likes: %v", err)
	}
	if count != 0 {
		t.Fatalf("final like count = %d, want 0", count)
	}

	_ = errors.Is(nil, gorm.ErrRecordNotFound)
}
