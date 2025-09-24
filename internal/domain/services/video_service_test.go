package services

import (
	"errors"
	"mime/multipart"
	"testing"
	"worker/internal/domain/entities"
)

// Mocks
type mockFrameExtractor struct {
	frames []string
	err    error
}

func (m *mockFrameExtractor) ExtractFrames(path, outDir string) ([]string, error) {
	return m.frames, m.err
}

type mockZipService struct {
	err error
}

func (m *mockZipService) CreateZipFile(files []string, dest string) error {
	return m.err
}

type mockStorageService struct{}

// DeleteFile implements StorageService.
func (m *mockStorageService) DeleteFile(path string) error {
	panic("unimplemented")
}

// ListProcessedFiles implements StorageService.
func (m *mockStorageService) ListProcessedFiles() ([]*entities.ProcessedFile, error) {
	panic("unimplemented")
}

// SaveUploadedFile implements StorageService.
func (m *mockStorageService) SaveUploadedFile(file multipart.File, filename string) (*entities.VideoFile, error) {
	panic("unimplemented")
}

func (m *mockStorageService) CreateDirectories() error {
	return nil
}

func TestValidateVideoFile(t *testing.T) {
	svc := NewVideoService(nil, nil, nil)

	valid := []string{"video.mp4", "movie.avi", "clip.mkv"}
	invalid := []string{"image.jpg", "doc.pdf", "audio.mp3"}

	for _, f := range valid {
		if !svc.ValidateVideoFile(f) {
			t.Errorf("expected '%s' to be valid", f)
		}
	}
	for _, f := range invalid {
		if svc.ValidateVideoFile(f) {
			t.Errorf("expected '%s' to be invalid", f)
		}
	}
}

func TestProcessVideo_Success(t *testing.T) {
	video := &entities.VideoFile{ID: "123", Path: "video.mp4"}
	frameExtractor := &mockFrameExtractor{frames: []string{"frame1.jpg", "frame2.jpg"}, err: nil}
	zipService := &mockZipService{err: nil}
	storageService := &mockStorageService{}

	svc := NewVideoService(frameExtractor, zipService, storageService)
	result, err := svc.ProcessVideo(video)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success true, got false")
	}
	if result.FrameCount != 2 {
		t.Errorf("expected FrameCount 2, got %d", result.FrameCount)
	}
}

func TestProcessVideo_FrameExtractorError(t *testing.T) {
	video := &entities.VideoFile{ID: "123", Path: "video.mp4"}
	frameExtractor := &mockFrameExtractor{frames: nil, err: errors.New("extract error")}
	zipService := &mockZipService{err: nil}
	storageService := &mockStorageService{}

	svc := NewVideoService(frameExtractor, zipService, storageService)
	result, err := svc.ProcessVideo(video)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result.Success {
		t.Errorf("expected Success false, got true")
	}
}
