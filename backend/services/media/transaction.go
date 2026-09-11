package media

import (
	"context"
	"errors"
)

// mediaDeleter is the narrow storage contract needed by transaction rollback.
// Keeping this interface local avoids importing the parent services package and
// creating an import cycle between services and services/media.
type mediaDeleter interface {
	Delete(ctx context.Context, key string) error
}

// RollbackUploadedObject removes a storage object after a failed database commit.
// Upload handlers can pass any storage implementation that satisfies mediaDeleter.
func RollbackUploadedObject(ctx context.Context, storage mediaDeleter, key string) error {
	if storage == nil {
		return errors.New("media storage is nil")
	}
	if key == "" {
		return errors.New("media key is empty")
	}
	return storage.Delete(ctx, key)
}
