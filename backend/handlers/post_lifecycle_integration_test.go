package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// These tests require an explicitly configured PostgreSQL test database.
// Example:
//   NOTELL_TEST_DATABASE_URL="host=localhost user=postgres password=postgres dbname=notell_test port=5432 sslmode=disable" go test ./handlers -run Lifecycle
//
// The test database must be disposable. The suite migrates only the tables it
// exercises and removes its rows during cleanup; it never falls back to the
// application's development database configuration.
func newPostLifecycleIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("NOTELL_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTELL_TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Upload{},
		&models.Post{},
		&models.Notification{},
	); err != nil {
		_ = closePostLifecycleDB(db)
		t.Fatalf("migrate PostgreSQL test database: %v", err)
	}

	t.Cleanup(func() {
		if err := closePostLifecycleDB(db); err != nil {
			t.Errorf("close PostgreSQL test database: %v", err)
		}
	})

	return db
}

func closePostLifecycleDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func createLifecycleUser(t *testing.T, db *gorm.DB, suffix string) models.User {
	t.Helper()
	stamp := time.Now().UnixNano()
	user := models.User{
		Username: fmt.Sprintf("lifecycle_%s_%d", suffix, stamp),
		Email:    fmt.Sprintf("%s_%d@example.test", suffix, stamp),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Delete(&user).Error
	})
	return user
}

func createLifecycleUpload(t *testing.T, db *gorm.DB, userID uint, filename, mediaType string) models.Upload {
	t.Helper()
	upload := models.Upload{
		UserID:    userID,
		Filename:  filename,
		Path:      filepath.ToSlash(filepath.Join("uploads", filename)),
		MediaType: mediaType,
	}
	if err := db.Create(&upload).Error; err != nil {
		t.Fatalf("create test upload: %v", err)
	}
	return upload
}

func invokeCreatePost(t *testing.T, h *PostHandler, userID uint, contentType, contentURL string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/posts", func(c *gin.Context) {
		c.Set("userId", userID)
		h.CreatePost(c)
	})

	payload, err := json.Marshal(map[string]string{
		"contentType": contentType,
		"contentUrl":  contentURL,
		"caption":     "lifecycle test",
	})
	if err != nil {
		t.Fatalf("marshal create-post payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPostMediaLifecycle_CreateClaimsUpload(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "claim")
	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, user.ID, filename, "image/jpeg")

	h := NewPostHandler(db, "https://example.test")
	w := invokeCreatePost(t, h, user.ID, "image", "https://example.test/uploads/"+filename)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreatePost status = %d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	var claimed models.Upload
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("reload upload: %v", err)
	}
	if claimed.PostID == nil {
		t.Fatal("expected upload to be claimed by created post")
	}
	if claimed.ClaimedAt == nil {
		t.Fatal("expected upload claimed_at to be populated")
	}

	var post models.Post
	if err := db.First(&post, *claimed.PostID).Error; err != nil {
		t.Fatalf("reload created post: %v", err)
	}
	if post.UserID != user.ID || post.ContentType != "image" || post.ContentURL != "https://example.test/uploads/"+filename {
		t.Fatalf("created post contract mismatch: %+v", post)
	}
}

func TestPostMediaLifecycle_RejectsDuplicateClaim(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "duplicate")
	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, user.ID, filename, "image/jpeg")

	h := NewPostHandler(db, "https://example.test")
	contentURL := "https://example.test/uploads/" + filename

	first := invokeCreatePost(t, h, user.ID, "image", contentURL)
	if first.Code != http.StatusCreated {
		t.Fatalf("first CreatePost status = %d, want %d; body=%s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := invokeCreatePost(t, h, user.ID, "image", contentURL)
	if second.Code != http.StatusBadRequest {
		t.Fatalf("duplicate CreatePost status = %d, want %d; body=%s", second.Code, http.StatusBadRequest, second.Body.String())
	}

	var count int64
	if err := db.Model(&models.Post{}).Where("content_url = ?", contentURL).Count(&count).Error; err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 1 {
		t.Fatalf("post count = %d, want 1", count)
	}

	var claimed models.Upload
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("reload upload: %v", err)
	}
	if claimed.PostID == nil {
		t.Fatal("expected upload to remain claimed")
	}
}

func TestPostMediaLifecycle_RejectsCrossUserClaim(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	owner := createLifecycleUser(t, db, "owner")
	attacker := createLifecycleUser(t, db, "attacker")
	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, owner.ID, filename, "image/jpeg")

	h := NewPostHandler(db, "https://example.test")
	w := invokeCreatePost(t, h, attacker.ID, "image", "https://example.test/uploads/"+filename)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("cross-user CreatePost status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	var claimed models.Upload
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("reload upload: %v", err)
	}
	if claimed.PostID != nil || claimed.ClaimedAt != nil {
		t.Fatalf("cross-user request changed upload claim: %+v", claimed)
	}
}

func TestPostMediaLifecycle_DeleteReleasesUploadAndRemovesFile(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "delete")
	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, user.ID, filename, "image/jpeg")

	if err := os.MkdirAll("uploads", 0755); err != nil {
		t.Fatalf("create uploads directory: %v", err)
	}
	mediaPath := filepath.Join("uploads", filename)
	if err := os.WriteFile(mediaPath, []byte("lifecycle-test"), 0600); err != nil {
		t.Fatalf("create test media file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(mediaPath) })

	// Exercise the real create path so the delete test covers the complete
	// claim lifecycle: upload row + physical file -> post -> deletion.
	h := NewPostHandler(db, "https://example.test")
	contentURL := "https://example.test/uploads/" + filename
	createResponse := invokeCreatePost(t, h, user.ID, "image", contentURL)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("CreatePost status = %d, want %d; body=%s", createResponse.Code, http.StatusCreated, createResponse.Body.String())
	}

	var claimed models.Upload
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("reload claimed upload: %v", err)
	}
	if claimed.PostID == nil {
		t.Fatal("expected upload to be claimed")
	}
	postID := *claimed.PostID

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set("userId", user.ID)
		h.DeletePost(c)
	})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%d", postID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("DeletePost status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var deletedPost models.Post
	if err := db.First(&deletedPost, postID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected post to be deleted, err=%v", err)
	}

	var deletedUpload models.Upload
	if err := db.First(&deletedUpload, upload.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected upload to be deleted, err=%v", err)
	}

	if _, err := os.Stat(mediaPath); !os.IsNotExist(err) {
		t.Fatalf("expected media file to be removed, stat err=%v", err)
	}
}
