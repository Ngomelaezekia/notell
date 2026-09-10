package media

import (
	"fmt"
	"net/http"
	"os"
)

// Metadata contains metadata that can be safely collected without a transcoder.
type Metadata struct {
	ContentType string
	FileSize    int64
}

// DetectContentType provides a shared media type detector for processors.
func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}

// ExtractMetadata collects metadata available from the uploaded file itself.
// Video dimensions and duration remain unset until a media tool such as FFmpeg is configured.
func ExtractMetadata(path string, contentType string) (Metadata, error) {
	if path == "" {
		return Metadata{}, fmt.Errorf("media path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("stat media: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Metadata{}, fmt.Errorf("media path is not a regular file")
	}
	return Metadata{ContentType: contentType, FileSize: info.Size()}, nil
}
