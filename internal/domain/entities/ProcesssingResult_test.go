package entities

import (
	"testing"
)

func TestProcessingResultInitialization(t *testing.T) {
	result := ProcessingResult{
		Success:    true,
		Message:    "Processamento concluído!",
		ZipPath:    "frames_123.zip",
		FrameCount: 10,
		Images:     []string{"frame1.jpg", "frame2.jpg"},
		VideoID:    "123",
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}
	if result.Message != "Processamento concluído!" {
		t.Errorf("expected Message 'Processamento concluído!', got '%s'", result.Message)
	}
	if result.ZipPath != "frames_123.zip" {
		t.Errorf("expected ZipPath 'frames_123.zip', got '%s'", result.ZipPath)
	}
	if result.FrameCount != 10 {
		t.Errorf("expected FrameCount 10, got %d", result.FrameCount)
	}
	if len(result.Images) != 2 || result.Images[0] != "frame1.jpg" || result.Images[1] != "frame2.jpg" {
		t.Errorf("expected Images ['frame1.jpg', 'frame2.jpg'], got %v", result.Images)
	}
	if result.VideoID != "123" {
		t.Errorf("expected VideoID '123', got '%s'", result.VideoID)
	}
}
