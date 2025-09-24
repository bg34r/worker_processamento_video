package video

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFrames_FakeFrames(t *testing.T) {
	outputDir := "temp_frames"
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

	// Cria arquivos fake de frames
	for i := 1; i <= 3; i++ {
		name := filepath.Join(outputDir, fmt.Sprintf("frame_%04d.png", i))
		os.WriteFile(name, []byte("fake"), 0644)
	}

	extractor := NewFFmpegExtractor()
	frames, err := extractor.ExtractFrames("fake_video.mp4", outputDir)
	if err == nil {
		t.Errorf("esperado erro do ffmpeg, mas não ocorreu")
	}
	if len(frames) != 0 {
		t.Errorf("esperado 0 frames, mas encontrou %d", len(frames))
	}
}

func TestExtractFrames_ErrorOnGlob(t *testing.T) {
	extractor := NewFFmpegExtractor()
	// Usa um diretório que não existe para forçar erro no Glob
	frames, err := extractor.ExtractFrames("fake_video.mp4", string([]byte{0}))
	if err == nil {
		t.Errorf("esperado erro ao listar frames, mas não ocorreu")
	}
	if frames != nil {
		t.Errorf("esperado frames nil em caso de erro")
	}
}
