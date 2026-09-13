package media

import "testing"

func TestValidateUploadMetadataAcceptsWebM(t *testing.T) {
	if err := ValidateUploadMetadata("video/webm", 1024, "clip.webm"); err != nil {
		t.Fatalf("ValidateUploadMetadata() error = %v", err)
	}
}

func TestValidateUploadMetadataRejectsWebMWithWrongExtension(t *testing.T) {
	if err := ValidateUploadMetadata("video/webm", 1024, "clip.mp4"); err == nil {
		t.Fatal("ValidateUploadMetadata() expected WebM extension error")
	}
}

func TestValidateUploadMetadataRejectsOversizedWebM(t *testing.T) {
	if err := ValidateUploadMetadata("video/webm", MaxMediaSize+1, "clip.webm"); err == nil {
		t.Fatal("ValidateUploadMetadata() expected size error")
	}
}
