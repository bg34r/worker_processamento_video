package services

import (
	"mime/multipart"
	"worker/internal/domain/entities"
)

type VideoService interface {
	ProcessVideo(videoFile *entities.VideoFile) (*entities.ProcessingResult, error)
	ValidateVideoFile(filename string) bool
}

type StorageService interface {
	SaveUploadedFile(file multipart.File, filename string) (*entities.VideoFile, error)
	DeleteFile(path string) error
	CreateDirectories() error
	ListProcessedFiles() ([]*entities.ProcessedFile, error)
}

type FrameExtractor interface {
	ExtractFrames(videoPath string, outputDir string) ([]string, error)
}

type ZipService interface {
	CreateZipFile(files []string, zipPath string) error
}
