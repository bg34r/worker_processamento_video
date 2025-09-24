package entities

import (
	"testing"
	"time"
)

func TestProcessedFileInitialization(t *testing.T) {
	now := time.Now()
	file := ProcessedFile{
		Filename:    "frames_123.zip",
		Size:        2048,
		CreatedAt:   now,
		DownloadURL: "http://localhost:8080/download/frames_123.zip",
	}

	if file.Filename != "frames_123.zip" {
		t.Errorf("expected Filename 'frames_123.zip', got '%s'", file.Filename)
	}
	if file.Size != 2048 {
		t.Errorf("expected Size 2048, got %d", file.Size)
	}
	if !file.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt '%v', got '%v'", now, file.CreatedAt)
	}
	if file.DownloadURL != "http://localhost:8080/download/frames_123.zip" {
		t.Errorf("expected DownloadURL 'http://localhost:8080/download/frames_123.zip', got '%s'", file.DownloadURL)
	}
}
