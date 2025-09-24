package handlers

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"worker/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// Mocks
type mockVideoService struct{}

func (m *mockVideoService) ValidateVideoFile(filename string) bool {
	return filename == "video.mp4"
}
func (m *mockVideoService) ProcessVideo(videoFile *entities.VideoFile) (*entities.ProcessingResult, error) {
	if videoFile.Filename == "video.mp4" {
		return &entities.ProcessingResult{Success: true, Message: "OK"}, nil
	}
	return &entities.ProcessingResult{Success: false, Message: "Erro"}, errors.New("erro")
}

type mockStorageService struct{}

func (m *mockStorageService) SaveUploadedFile(file multipart.File, filename string) (*entities.VideoFile, error) {
	if filename == "video.mp4" {
		return &entities.VideoFile{Filename: filename, Path: "uploads/video.mp4"}, nil
	}
	return nil, errors.New("erro ao salvar")
}
func (m *mockStorageService) DeleteFile(path string) error { return nil }
func (m *mockStorageService) CreateDirectories() error     { return nil }
func (m *mockStorageService) ListProcessedFiles() ([]*entities.ProcessedFile, error) {
	return []*entities.ProcessedFile{
		{Filename: "file1.zip", Size: 123, DownloadURL: "/outputs/file1.zip"},
	}, nil
}

func TestUploadVideo_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &mockStorageService{})
	router.POST("/upload", handler.UploadVideo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "video.mp4")
	io.Copy(part, bytes.NewReader([]byte("fake content")))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", resp.Code)
	}
}

func TestUploadVideo_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &mockStorageService{})
	router.POST("/upload", handler.UploadVideo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "video.txt")
	io.Copy(part, bytes.NewReader([]byte("fake content")))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", resp.Code)
	}
}

func TestUploadVideo_SaveError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &mockStorageService{})
	router.POST("/upload", handler.UploadVideo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "erro.mp4")
	io.Copy(part, bytes.NewReader([]byte("fake content")))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", resp.Code)
	}
}

func TestGetStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &mockStorageService{})
	router.GET("/status", handler.GetStatus)

	req := httptest.NewRequest("GET", "/status", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", resp.Code)
	}
}

func TestUploadVideo_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &mockStorageService{})
	router.POST("/upload", handler.UploadVideo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Não adiciona arquivo
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", resp.Code)
	}
}

type errorStorageService struct {
	mockStorageService
}

func (m *errorStorageService) ListProcessedFiles() ([]*entities.ProcessedFile, error) {
	return nil, errors.New("erro ao listar")
}

func TestGetStatus_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewVideoHandler(&mockVideoService{}, &errorStorageService{})
	router.GET("/status", handler.GetStatus)

	req := httptest.NewRequest("GET", "/status", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("esperado status 500, obteve %d", resp.Code)
	}
}
