package postaccess

import (
    "errors"

    "notell/models"

    "gorm.io/gorm"
 )

var (
    ErrAccessDenied = errors.New("post access denied")
    ErrPostNotFound = errors.New("post not found")
)

func CanViewPost(db *gorm.DB, postID, viewerID uint) error {
    if db == nil || postID == 0 { return ErrPostNotFound }
    var post models.Post
    if err := db.Select("id, user_id, visibility").First(&post, postID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { return ErrPostNotFound }
        return err
    }
    return CanViewRecordWithDB(db, post, viewerID)
}

func CanViewRecord(post models.Post, viewerID uint) error {
    if post.UserID == viewerID || post.Visibility == "public" { return nil }
    return ErrAccessDenied
}

func CanViewRecordWithDB(db *gorm.DB, post models.Post, viewerID uint) error {
    if post.UserID == viewerID { return nil }
    switch post.Visibility {
    case "public": return nil
    case "followers":
        if viewerID == 0 || db == nil { return ErrAccessDenied }
        var count int64
        if err := db.Model(&models.Relationship{}).Where("follower_id = ? AND following_id = ? AND status = ?", viewerID, post.UserID, "accepted").Count(&count).Error; err != nil { return err }
        if count > 0 { return nil }
    case "selected":
        if viewerID == 0 || db == nil { return ErrAccessDenied }
        var count int64
        if err := db.Model(&models.PostAudience{}).Where("post_id = ? AND user_id = ?", post.ID, viewerID).Count(&count).Error; err != nil { return err }
        if count > 0 { return nil }
    case "subscribers":
        // Subscriber state is owned by the payment/entitlement service. Fail closed until explicitly wired.
        return ErrAccessDenied
    }
    return ErrAccessDenied
}

func VisibleTo(db *gorm.DB, viewerID uint) *gorm.DB {
    if db == nil { return db }
    return db.Where("("+"posts.visibility = ? OR posts.user_id = ? OR "+"(posts.visibility = ? AND EXISTS (SELECT 1 FROM user_relationships ur WHERE ur.follower_id = ? AND ur.following_id = posts.user_id AND ur.status = ?)) OR "+"(posts.visibility = ? AND EXISTS (SELECT 1 FROM post_audiences pa WHERE pa.post_id = posts.id AND pa.user_id = ?))"+")", "public", viewerID, "followers", viewerID, "accepted", "selected", viewerID)
}

func CanEditPost(db *gorm.DB, postID, viewerID uint) error {
    if db == nil || postID == 0 || viewerID == 0 { return ErrAccessDenied }
    var post models.Post
    if err := db.Select("id, user_id").First(&post, postID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { return ErrPostNotFound }
        return err
    }
    if post.UserID != viewerID { return ErrAccessDenied }
    return nil
}

func CanDeletePost(db *gorm.DB, postID, viewerID uint) error { return CanEditPost(db, postID, viewerID) }
func CanManageMusic(db *gorm.DB, postID, viewerID uint) error { return CanEditPost(db, postID, viewerID) }
func CanManageMedia(db *gorm.DB, postID, viewerID uint) error { return CanEditPost(db, postID, viewerID) }
