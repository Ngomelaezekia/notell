package media

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

const MaxMediaSize int64 = 100 << 20

var allowedMediaTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"video/mp4":       {},
	"video/quicktime": {},
	"video/webm":      {},
	"audio/mpeg":      {},
}

// ValidateUploadMetadata validates the same media contract enforced by the upload endpoint.
func ValidateUploadMetadata(contentType string, size int64, filename string) error {
	if size <= 0 {
		return fmt.Errorf("media file is empty")
	}
	if size > MaxMediaSize {
		return fmt.Errorf("media file exceeds the %d MiB limit", MaxMediaSize>>20)
	}
	if _, ok := allowedMediaTypes[contentType]; !ok {
		return fmt.Errorf("unsupported media type %q", contentType)
	}
	if ext := strings.ToLower(filepath.Ext(filename)); ext == "" {
		return fmt.Errorf("media filename has no extension")
	}
	if contentType == "audio/mpeg" && strings.ToLower(filepath.Ext(filename)) != ".mp3" {
		return fmt.Errorf("audio media must use an .mp3 filename")
	}
	if contentType == "video/webm" && strings.ToLower(filepath.Ext(filename)) != ".webm" {
		return fmt.Errorf("WebM video must use a .webm filename")
	}
	return nil
}

// DetectAndValidateContentType validates the supplied bytes against supported media types.
func DetectAndValidateContentType(data []byte) (string, error) {
	contentType := http.DetectContentType(data)
	if _, ok := allowedMediaTypes[contentType]; !ok {
		return "", fmt.Errorf("unsupported media type %q", contentType)
	}
	return contentType, nil
}
