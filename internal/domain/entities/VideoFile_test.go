package entities

import (
	"testing"
	"time"
)

func TestVideoFileInitialization(t *testing.T) {
	now := time.Now()
	video := VideoFile{
		ID:           "123",
		Filename:     "video.mp4",
		OriginalName: "original.mp4",
		Path:         "/uploads/video.mp4",
		Size:         1024,
		CreatedAt:    now,
	}

	if video.ID != "123" {
		t.Errorf("expected ID '123', got '%s'", video.ID)
	}
	if video.Filename != "video.mp4" {
		t.Errorf("expected Filename 'video.mp4', got '%s'", video.Filename)
	}
	if video.OriginalName != "original.mp4" {
		t.Errorf("expected OriginalName 'original.mp4', got '%s'", video.OriginalName)
	}
	if video.Path != "/uploads/video.mp4" {
		t.Errorf("expected Path '/uploads/video.mp4', got '%s'", video.Path)
	}
	if video.Size != 1024 {
		t.Errorf("expected Size 1024, got %d", video.Size)
	}
	if !video.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt '%v', got '%v'", now, video.CreatedAt)
	}
}
