package handlers

import (
	"fmt"
	"os"
	"testing"
	"time"

	"notell/models"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openNotificationLifecycleDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("NOTELL_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTELL_TEST_DATABASE_URL is not set")
	}

	db, err := openTestPostgres(dsn, logger.Error)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Notification{},
	); err != nil {
		t.Fatalf("migrate notification lifecycle schema: %v", err)
	}
	return db
}

func createNotificationLifecycleUser(t *testing.T, db *gorm.DB, suffix string) models.User {
	t.Helper()

	user := models.User{
		Username: fmt.Sprintf("notification_lifecycle_%s_%d", suffix, time.Now().UnixNano()),
		Email:    fmt.Sprintf("notification_lifecycle_%s_%d@example.test", suffix, time.Now().UnixNano()),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create %s user: %v", suffix, err)
	}
	return user
}

func TestNotificationForeignKeysHandleOwnerDeletion(t *testing.T) {
	db := openNotificationLifecycleDB(t)

	recipient := createNotificationLifecycleUser(t, db, "recipient")
	actor := createNotificationLifecycleUser(t, db, "actor")
	owner := createNotificationLifecycleUser(t, db, "owner")

	post := models.Post{
		UserID:      owner.ID,
		ContentType: "image",
		ContentURL:  "https://example.test/uploads/notification-lifecycle.jpg",
		Caption:     "notification lifecycle",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	comment := models.Comment{
		UserID:  actor.ID,
		PostID:  post.ID,
		Content: "notification lifecycle comment",
	}
	if err := db.Create(&comment).Error; err != nil {
		t.Fatalf("create comment: %v", err)
	}

	notification := models.Notification{
		UserID:    recipient.ID,
		ActorID:   actor.ID,
		Type:      "comment",
		PostID:    &post.ID,
		CommentID: &comment.ID,
	}
	if err := db.Create(&notification).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if err := db.Delete(&owner).Error; err != nil {
		t.Fatalf("delete post owner: %v", err)
	}

	var remaining models.Notification
	if err := db.First(&remaining, notification.ID).Error; err != nil {
		t.Fatalf("notification should survive post/comment deletion: %v", err)
	}
	if remaining.PostID != nil || remaining.CommentID != nil {
		t.Fatalf("notification references deleted content: post=%v comment=%v", remaining.PostID, remaining.CommentID)
	}

	if err := db.Delete(&recipient).Error; err != nil {
		t.Fatalf("delete recipient: %v", err)
	}
	var deleted models.Notification
	if err := db.First(&deleted, notification.ID).Error; err == nil {
		t.Fatal("notification should be deleted with its recipient")
	}

	if err := db.Delete(&actor).Error; err != nil {
		t.Fatalf("delete actor: %v", err)
	}
}
