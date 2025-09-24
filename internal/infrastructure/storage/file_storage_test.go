package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirectories(t *testing.T) {
	fs := NewFileStorage()
	err := fs.CreateDirectories()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Verifica se os diretórios foram criados
	for _, dir := range []string{"uploads", "outputs", "temp"} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory '%s' to exist", dir)
		}
	}
}

type mockMultipartFile struct {
	*bytes.Reader
	closed bool
}

func (m *mockMultipartFile) Close() error {
	m.closed = true
	return nil
}

func TestSaveUploadedFile(t *testing.T) {
	fs := NewFileStorage()
	fs.CreateDirectories()

	content := []byte("video content")
	file := &mockMultipartFile{Reader: bytes.NewReader(content)}
	filename := "test_video.mp4"

	videoFile, err := fs.SaveUploadedFile(file, filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if videoFile.OriginalName != filename {
		t.Errorf("expected OriginalName '%s', got '%s'", filename, videoFile.OriginalName)
	}
	if videoFile.Size != int64(len(content)) {
		t.Errorf("expected Size %d, got %d", len(content), videoFile.Size)
	}
	// Limpa o arquivo criado
	os.Remove(videoFile.Path)
}

func TestDeleteFile(t *testing.T) {
	fs := NewFileStorage()
	fs.CreateDirectories()

	// Cria um arquivo temporário
	tmpFile := filepath.Join("uploads", "delete_me.txt")
	os.WriteFile(tmpFile, []byte("delete me"), 0644)

	err := fs.DeleteFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted")
	}
}

func TestListProcessedFiles(t *testing.T) {
	fs := NewFileStorage()
	fs.CreateDirectories()

	// Cria um arquivo ZIP fictício
	zipPath := filepath.Join("outputs", "test.zip")
	os.WriteFile(zipPath, []byte("zip content"), 0644)

	files, err := fs.ListProcessedFiles()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	found := false
	for _, f := range files {
		if f.Filename == "test.zip" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find 'test.zip' in processed files")
	}
	// Limpa o arquivo criado
	os.Remove(zipPath)
}

func TestSaveUploadedFile_ErrorOnCreate(t *testing.T) {
	fs := NewFileStorage()
	fs.CreateDirectories()

	content := []byte("video content")
	file := &mockMultipartFile{Reader: bytes.NewReader(content)}
	// Nome inválido no Windows
	filename := "test<>file.mp4"

	_, err := fs.SaveUploadedFile(file, filename)
	if err == nil {
		t.Errorf("expected error when saving file with invalid filename")
	}
}
