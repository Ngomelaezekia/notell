package media

import (
	"context"
	"errors"

	"notell/services"
)

// RollbackUploadedObject removes a storage object after a failed database commit.
// Upload handlers should call this from every failure path after a successful Put.
func RollbackUploadedObject(ctx context.Context, storage services.MediaStorage, key string) error {
	if storage == nil {
		return errors.New("media storage is nil")
	}
	if key == "" {
		return errors.New("media key is empty")
	}
	return storage.Delete(ctx, key)
}
