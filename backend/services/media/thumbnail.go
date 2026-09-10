package media

import "context"

// ThumbnailGenerator defines the thumbnail step without coupling media processing
// to a specific image/video tool. FFmpeg can implement this boundary later.
type ThumbnailGenerator interface {
	Generate(ctx context.Context, sourcePath, destinationPath string) error
}

// NoopThumbnailGenerator is used when thumbnail tooling is not configured yet.
// It deliberately does not claim a thumbnail was generated.
type NoopThumbnailGenerator struct{}

func (NoopThumbnailGenerator) Generate(context.Context, string, string) error { return nil }
