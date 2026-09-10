package media

import "context"

// Transcoder defines the media stream generation boundary.
// Implementations can later use FFmpeg or external processing services without
// changing upload or worker contracts.
type Transcoder interface {
	GenerateStreams(ctx context.Context, inputPath string) (*StreamOutput, error)
}

// StreamOutput represents generated playback assets.
type StreamOutput struct {
	MasterPlaylistURL string
	Variants          []StreamVariant
}

type StreamVariant struct {
	Name      string
	Width     int
	Height    int
	URL       string
	Bitrate   int64
}

// NoopTranscoder keeps the current MP4 flow working until real transcoding is enabled.
type NoopTranscoder struct{}

func NewNoopTranscoder() *NoopTranscoder {
	return &NoopTranscoder{}
}

func (t *NoopTranscoder) GenerateStreams(ctx context.Context, inputPath string) (*StreamOutput, error) {
	return &StreamOutput{}, nil
}
