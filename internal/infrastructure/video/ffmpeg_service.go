package video

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type FrameExtractor interface {
	ExtractFrames(videoPath, outputDir string) ([]string, error)
}

type ffmpegExtractor struct{}

func NewFFmpegExtractor() FrameExtractor {
	return &ffmpegExtractor{}
}

func (f *ffmpegExtractor) ExtractFrames(videoPath string, outputDir string) ([]string, error) {
	framePattern := filepath.Join(outputDir, "frame_%04d.png")

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro no ffmpeg: %s\nOutput: %s", err.Error(), string(output))
	}

	// Listar arquivos gerados
	frames, err := filepath.Glob(filepath.Join(outputDir, "*.png"))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar frames: %s", err.Error())
	}

	return frames, nil
}
