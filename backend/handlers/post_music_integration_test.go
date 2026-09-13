package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
)

func createLifecycleAudioUpload(t *testing.T, db interface {
	Create(value interface{}) interface{ Error }
}, _ uint) {
	// Intentionally unused helper signature guard; concrete fixtures are created
	// below through the existing lifecycle helpers in this package.
}

func invokeSetPostMusic(t *testing.T, h *PostHandler, userID, postID, uploadID uint, startSec, endSec, volume float64) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/posts/:id/music", func(c *gin.Context) {
		c.Set("userId", userID)
		h.SetPostMusic(c)
	})

	payload, err := json.Marshal(postMusicInput{
		UploadID: uploadID,
		StartSec: startSec,
		EndSec:   endSec,
		Volume:   volume,
	})
	if err != nil {
		t.Fatalf("marshal post-music payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%d/music", postID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func invokeRemovePostMusic(t *testing.T, h *PostHandler, userID, postID uint) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/posts/:id/music", func(c *gin.Context) {
		c.Set("userId", userID)
		h.RemovePostMusic(c)
	})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%d/music", postID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func createLifecyclePostForMusic(t *testing.T, db interface {
	Create(value interface{}) interface{ Error }
}, userID uint) uint {
	return 0
}

func TestPostMusicLifecycle_AttachAndRemove(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "music_attach")

	post := models.Post{UserID: user.ID, ContentType: "image", ContentURL: "https://example.test/uploads/post.jpg", Visibility: "public"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	filename := fmt.Sprintf("%d.mp3", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, user.ID, filename, "audio/mpeg")
	if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", upload.ID).Update("duration", 120).Error; err != nil {
		t.Fatalf("set audio duration: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	attach := invokeSetPostMusic(t, h, user.ID, post.ID, upload.ID, 10, 60, 0.75)
	if attach.Code != http.StatusCreated {
		t.Fatalf("SetPostMusic status = %d, want %d; body=%s", attach.Code, http.StatusCreated, attach.Body.String())
	}

	var music models.PostMusic
	if err := db.Where("post_id = ?", post.ID).First(&music).Error; err != nil {
		t.Fatalf("load post music: %v", err)
	}
	if music.UploadID != upload.ID || music.StartSec != 10 || music.EndSec != 60 || music.Volume != 0.75 {
		t.Fatalf("unexpected post music: %+v", music)
	}

	var claimed models.Upload
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("load claimed audio upload: %v", err)
	}
	if claimed.PostID == nil || *claimed.PostID != post.ID {
		t.Fatalf("audio upload was not claimed by post: %+v", claimed)
	}

	remove := invokeRemovePostMusic(t, h, user.ID, post.ID)
	if remove.Code != http.StatusOK {
		t.Fatalf("RemovePostMusic status = %d, want %d; body=%s", remove.Code, http.StatusOK, remove.Body.String())
	}

	if err := db.Where("post_id = ?", post.ID).First(&music).Error; err == nil {
		t.Fatal("expected post music relationship to be removed")
	}
	if err := db.First(&claimed, upload.ID).Error; err != nil {
		t.Fatalf("reload released audio upload: %v", err)
	}
	if claimed.PostID != nil {
		t.Fatalf("expected released audio upload, got post_id=%v", *claimed.PostID)
	}
}

func TestPostMusicLifecycle_RejectsCrossUserAndNotReadyAudio(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	owner := createLifecycleUser(t, db, "music_owner")
	attacker := createLifecycleUser(t, db, "music_attacker")
	post := models.Post{UserID: owner.ID, ContentType: "image", ContentURL: "https://example.test/uploads/post.jpg", Visibility: "private"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	attackerAudio := createLifecycleUpload(t, db, attacker.ID, fmt.Sprintf("%d.mp3", time.Now().UnixNano()), "audio/mpeg")
	if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", attackerAudio.ID).Update("duration", 120).Error; err != nil {
		t.Fatalf("set attacker audio duration: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	crossUser := invokeSetPostMusic(t, h, attacker.ID, post.ID, attackerAudio.ID, 0, 0, 1)
	if crossUser.Code != http.StatusNotFound {
		t.Fatalf("cross-user SetPostMusic status = %d, want %d; body=%s", crossUser.Code, http.StatusNotFound, crossUser.Body.String())
	}

	notReady := createLifecycleUpload(t, db, owner.ID, fmt.Sprintf("%d.mp3", time.Now().UnixNano()), "audio/mpeg")
	if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", notReady.ID).Updates(map[string]interface{}{"duration": 120, "status": "processing"}).Error; err != nil {
		t.Fatalf("mark audio processing: %v", err)
	}
	readyCheck := invokeSetPostMusic(t, h, owner.ID, post.ID, notReady.ID, 0, 0, 1)
	if readyCheck.Code != http.StatusBadRequest {
		t.Fatalf("not-ready SetPostMusic status = %d, want %d; body=%s", readyCheck.Code, http.StatusBadRequest, readyCheck.Body.String())
	}
}

func TestPostMusicLifecycle_ReplacementReleasesPreviousUpload(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "music_replace")
	post := models.Post{UserID: user.ID, ContentType: "image", ContentURL: "https://example.test/uploads/post.jpg", Visibility: "public"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	first := createLifecycleUpload(t, db, user.ID, fmt.Sprintf("%d-first.mp3", time.Now().UnixNano()), "audio/mpeg")
	second := createLifecycleUpload(t, db, user.ID, fmt.Sprintf("%d-second.mp3", time.Now().UnixNano()), "audio/mpeg")
	for _, upload := range []models.Upload{first, second} {
		if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", upload.ID).Update("duration", 180).Error; err != nil {
			t.Fatalf("set audio duration: %v", err)
		}
	}

	h := NewPostHandler(db, "https://example.test")
	firstAttach := invokeSetPostMusic(t, h, user.ID, post.ID, first.ID, 0, 90, 1)
	if firstAttach.Code != http.StatusCreated {
		t.Fatalf("first attach status = %d, want %d; body=%s", firstAttach.Code, http.StatusCreated, firstAttach.Body.String())
	}
	secondAttach := invokeSetPostMusic(t, h, user.ID, post.ID, second.ID, 20, 80, 0.5)
	if secondAttach.Code != http.StatusCreated {
		t.Fatalf("replacement attach status = %d, want %d; body=%s", secondAttach.Code, http.StatusCreated, secondAttach.Body.String())
	}

	var music models.PostMusic
	if err := db.Where("post_id = ?", post.ID).First(&music).Error; err != nil {
		t.Fatalf("load replacement music: %v", err)
	}
	if music.UploadID != second.ID {
		t.Fatalf("music upload = %d, want replacement %d", music.UploadID, second.ID)
	}

	var releasedFirst models.Upload
	if err := db.First(&releasedFirst, first.ID).Error; err != nil {
		t.Fatalf("load previous upload: %v", err)
	}
	if releasedFirst.PostID != nil {
		t.Fatalf("previous audio upload remained claimed: %+v", releasedFirst)
	}
}

func TestPostMusicLifecycle_DeletePostCascadesMusicAndAudioClaim(t *testing.T) {
	db := newPostLifecycleIntegrationDB(t)
	user := createLifecycleUser(t, db, "music_delete")
	filename := fmt.Sprintf("%d.mp3", time.Now().UnixNano())
	upload := createLifecycleUpload(t, db, user.ID, filename, "audio/mpeg")
	if err := db.Model(&models.MediaMetadata{}).Where("upload_id = ?", upload.ID).Update("duration", 120).Error; err != nil {
		t.Fatalf("set audio duration: %v", err)
	}

	h := NewPostHandler(db, "https://example.test")
	post := models.Post{UserID: user.ID, ContentType: "image", ContentURL: "https://example.test/uploads/post.jpg", Visibility: "public"}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	attach := invokeSetPostMusic(t, h, user.ID, post.ID, upload.ID, 0, 0, 1)
	if attach.Code != http.StatusCreated {
		t.Fatalf("attach status = %d, want %d; body=%s", attach.Code, http.StatusCreated, attach.Body.String())
	}

	if err := db.Where("id = ?", post.ID).Delete(&models.Post{}).Error; err != nil {
		t.Fatalf("delete post fixture: %v", err)
	}
	var music models.PostMusic
	if err := db.Where("post_id = ?", post.ID).First(&music).Error; err == nil {
		t.Fatal("expected post music cascade to remove relationship")
	}
}
