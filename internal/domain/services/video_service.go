package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"worker/internal/domain/entities"
)

type videoService struct {
	frameExtractor FrameExtractor
	zipService     ZipService
	storageService StorageService
}

func NewVideoService(frameExtractor FrameExtractor, zipService ZipService, storageService StorageService) VideoService {
	return &videoService{
		frameExtractor: frameExtractor,
		zipService:     zipService,
		storageService: storageService,
	}
}

func (s *videoService) ProcessVideo(videoFile *entities.VideoFile) (*entities.ProcessingResult, error) {
	fmt.Printf("Iniciando processamento: %s\n", videoFile.Path)

	// Criar diretório temporário
	tempDir := filepath.Join("temp", videoFile.ID)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Extrair frames
	frames, err := s.frameExtractor.ExtractFrames(videoFile.Path, tempDir)
	if err != nil {
		return &entities.ProcessingResult{
			Success: false,
			Message: fmt.Sprintf("Erro na extração de frames: %s", err.Error()),
		}, err
	}

	if len(frames) == 0 {
		return &entities.ProcessingResult{
			Success: false,
			Message: "Nenhum frame foi extraído do vídeo",
		}, fmt.Errorf("nenhum frame extraído")
	}

	fmt.Printf("📸 Extraídos %d frames\n", len(frames))

	// Criar ZIP
	zipFilename := fmt.Sprintf("frames_%s.zip", videoFile.ID)
	zipPath := filepath.Join("outputs", zipFilename)

	err = s.zipService.CreateZipFile(frames, zipPath)
	if err != nil {
		return &entities.ProcessingResult{
			Success: false,
			Message: "Erro ao criar arquivo ZIP: " + err.Error(),
		}, err
	}

	fmt.Printf("✅ ZIP criado: %s\n", zipPath)

	// Preparar nomes das imagens
	imageNames := make([]string, len(frames))
	for i, frame := range frames {
		imageNames[i] = filepath.Base(frame)
	}

	return &entities.ProcessingResult{
		Success:    true,
		Message:    fmt.Sprintf("Processamento concluído! %d frames extraídos.", len(frames)),
		ZipPath:    zipFilename,
		FrameCount: len(frames),
		Images:     imageNames,
		VideoID:    videoFile.ID,
	}, nil
}

func (s *videoService) ValidateVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
