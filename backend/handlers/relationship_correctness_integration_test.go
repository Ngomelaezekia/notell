package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"notell/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newRelationshipCorrectnessDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("NOTELL_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTELL_TEST_DATABASE_URL is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Relationship{}, &models.Notification{}); err != nil {
		t.Fatalf("migrate relationship test database: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func createRelationshipTestUser(t *testing.T, db *gorm.DB, suffix string, allowFollowers bool) models.User {
	t.Helper()
	stamp := time.Now().UnixNano()
	user := models.User{
		Username: fmt.Sprintf("relationship_%s_%d", suffix, stamp),
		Email: fmt.Sprintf("relationship_%s_%d@example.test", suffix, stamp),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create relationship test user: %v", err)
	}
	// The User model has a gorm default:true tag, so a false value supplied
	// during Create is replaced by the database/model default. Use an explicit
	// update to make the false case deterministic.
	if err := db.Model(&user).Update("allow_followers", allowFollowers).Error; err != nil {
		t.Fatalf("set allow_followers: %v", err)
	}
	user.AllowFollowers = allowFollowers
	t.Cleanup(func() {
		_ = db.Where("follower_id = ? OR following_id = ?", user.ID, user.ID).Delete(&models.Relationship{}).Error
		_ = db.Where("user_id = ? OR actor_id = ?", user.ID, user.ID).Delete(&models.Notification{}).Error
		_ = db.Delete(&user).Error
	})
	return user
}

func invokeRelationship(t *testing.T, h *RelationshipHandler, method, path string, userID uint, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Handle(method, "/users/:id/*action", func(c *gin.Context) {
		c.Set("userId", userID)
		handler(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestFollowUserRejectsSelfFollow(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	user := createRelationshipTestUser(t, db, "self", true)
	h := NewRelationshipHandler(db)

	w := invokeRelationship(t, h, http.MethodPost, fmt.Sprintf("/users/%d/follow", user.ID), user.ID, h.FollowUser)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestFollowUserRejectsMissingTarget(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	user := createRelationshipTestUser(t, db, "missing-target", true)
	h := NewRelationshipHandler(db)

	w := invokeRelationship(t, h, http.MethodPost, "/users/999999999/follow", user.ID, h.FollowUser)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestFollowUserRespectsFollowerSetting(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	follower := createRelationshipTestUser(t, db, "blocked-follower", true)
	target := createRelationshipTestUser(t, db, "blocked-target", false)
	h := NewRelationshipHandler(db)

	w := invokeRelationship(t, h, http.MethodPost, fmt.Sprintf("/users/%d/follow", target.ID), follower.ID, h.FollowUser)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusForbidden, w.Body.String())
	}

	var count int64
	if err := db.Model(&models.Relationship{}).Where("follower_id = ? AND following_id = ?", follower.ID, target.ID).Count(&count).Error; err != nil {
		t.Fatalf("count relationship: %v", err)
	}
	if count != 0 {
		t.Fatalf("relationship count = %d, want 0", count)
	}
}

func TestFollowUserRejectsDuplicateWithoutCreatingNotification(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	follower := createRelationshipTestUser(t, db, "duplicate-follower", true)
	target := createRelationshipTestUser(t, db, "duplicate-target", true)
	h := NewRelationshipHandler(db)

	first := invokeRelationship(t, h, http.MethodPost, fmt.Sprintf("/users/%d/follow", target.ID), follower.ID, h.FollowUser)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d; body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	second := invokeRelationship(t, h, http.MethodPost, fmt.Sprintf("/users/%d/follow", target.ID), follower.ID, h.FollowUser)
	if second.Code != http.StatusConflict {
		t.Fatalf("second status = %d, want %d; body=%s", second.Code, http.StatusConflict, second.Body.String())
	}

	var relationshipCount, notificationCount int64
	if err := db.Model(&models.Relationship{}).Where("follower_id = ? AND following_id = ?", follower.ID, target.ID).Count(&relationshipCount).Error; err != nil {
		t.Fatalf("count relationship: %v", err)
	}
	if err := db.Model(&models.Notification{}).Where("user_id = ? AND actor_id = ? AND type = ?", target.ID, follower.ID, "follow").Count(&notificationCount).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if relationshipCount != 1 || notificationCount != 1 {
		t.Fatalf("relationshipCount=%d notificationCount=%d, want 1 and 1", relationshipCount, notificationCount)
	}
}

func TestUnfollowUserRemovesExistingRelationship(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	follower := createRelationshipTestUser(t, db, "unfollow-follower", true)
	target := createRelationshipTestUser(t, db, "unfollow-target", true)
	h := NewRelationshipHandler(db)

	if err := db.Create(&models.Relationship{FollowerID: follower.ID, FollowingID: target.ID, Status: "accepted"}).Error; err != nil {
		t.Fatalf("create relationship: %v", err)
	}

	w := invokeRelationship(t, h, http.MethodDelete, fmt.Sprintf("/users/%d/unfollow", target.ID), follower.ID, h.UnfollowUser)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var count int64
	if err := db.Model(&models.Relationship{}).Where("follower_id = ? AND following_id = ?", follower.ID, target.ID).Count(&count).Error; err != nil {
		t.Fatalf("count relationship: %v", err)
	}
	if count != 0 {
		t.Fatalf("relationship count = %d, want 0", count)
	}
}

func TestUnfollowUserReportsMissingRelationship(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	follower := createRelationshipTestUser(t, db, "unfollow-missing-follower", true)
	target := createRelationshipTestUser(t, db, "unfollow-missing-target", true)
	h := NewRelationshipHandler(db)

	w := invokeRelationship(t, h, http.MethodDelete, fmt.Sprintf("/users/%d/unfollow", target.ID), follower.ID, h.UnfollowUser)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestGetRelationshipStatusReportsBothDirectionsAndCounts(t *testing.T) {
	db := newRelationshipCorrectnessDB(t)
	viewer := createRelationshipTestUser(t, db, "status-viewer", true)
	target := createRelationshipTestUser(t, db, "status-target", true)
	other := createRelationshipTestUser(t, db, "status-other", true)
	h := NewRelationshipHandler(db)

	for _, rel := range []models.Relationship{
		{FollowerID: viewer.ID, FollowingID: target.ID, Status: "accepted"},
		{FollowerID: other.ID, FollowingID: target.ID, Status: "accepted"},
		{FollowerID: target.ID, FollowingID: viewer.ID, Status: "accepted"},
	} {
		if err := db.Create(&rel).Error; err != nil {
			t.Fatalf("create relationship: %v", err)
		}
	}

	w := invokeRelationship(t, h, http.MethodGet, fmt.Sprintf("/users/%d/relationship", target.ID), viewer.ID, h.GetRelationshipStatus)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var payload struct {
		Data struct {
			Following bool `json:"following"`
			Follower bool `json:"follower"`
			FollowerCount int64 `json:"followerCount"`
			FollowingCount int64 `json:"followingCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode relationship response: %v", err)
	}
	if !payload.Data.Following || !payload.Data.Follower || payload.Data.FollowerCount != 2 || payload.Data.FollowingCount != 1 {
		t.Fatalf("unexpected relationship payload: %+v", payload.Data)
	}
}