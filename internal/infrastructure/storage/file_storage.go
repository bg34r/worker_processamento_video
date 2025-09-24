package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"worker/internal/domain/entities"
	"worker/internal/domain/services"
)

type fileStorage struct{}

func NewFileStorage() services.StorageService {
	return &fileStorage{}
}

func (f *fileStorage) CreateDirectories() error {
	dirs := []string{"uploads", "outputs", "temp"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (f *fileStorage) SaveUploadedFile(file multipart.File, filename string) (*entities.VideoFile, error) {
	timestamp := time.Now().Format("20060102_150405")
	id := timestamp
	newFilename := fmt.Sprintf("%s_%s", timestamp, filename)
	videoPath := filepath.Join("uploads", newFilename)

	out, err := os.Create(videoPath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	size, err := io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	return &entities.VideoFile{
		ID:           id,
		Filename:     newFilename,
		OriginalName: filename,
		Path:         videoPath,
		Size:         size,
		CreatedAt:    time.Now(),
	}, nil
}

func (f *fileStorage) DeleteFile(path string) error {
	return os.Remove(path)
}

func (f *fileStorage) ListProcessedFiles() ([]*entities.ProcessedFile, error) {
	files, err := filepath.Glob(filepath.Join("outputs", "*.zip"))
	if err != nil {
		return nil, err
	}

	var results []*entities.ProcessedFile
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		results = append(results, &entities.ProcessedFile{
			Filename:    filepath.Base(file),
			Size:        info.Size(),
			CreatedAt:   info.ModTime(),
			DownloadURL: "/download/" + filepath.Base(file),
		})
	}

	return results, nil
}
