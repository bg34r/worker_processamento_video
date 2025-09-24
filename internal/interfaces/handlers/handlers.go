package handlers

import (
	"net/http"
	"worker/internal/domain/services"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService   services.VideoService
	storageService services.StorageService
}

func NewVideoHandler(videoService services.VideoService, storageService services.StorageService) *VideoHandler {
	return &VideoHandler{
		videoService:   videoService,
		storageService: storageService,
	}
}

func (h *VideoHandler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Erro ao receber arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	if !h.videoService.ValidateVideoFile(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Formato de arquivo não suportado. Use: mp4, avi, mov, mkv",
		})
		return
	}

	videoFile, err := h.storageService.SaveUploadedFile(file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erro ao salvar arquivo: " + err.Error(),
		})
		return
	}

	result, err := h.videoService.ProcessVideo(videoFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	if result.Success {
		h.storageService.DeleteFile(videoFile.Path)
	}

	c.JSON(http.StatusOK, result)
}

func (h *VideoHandler) GetStatus(c *gin.Context) {
	files, err := h.storageService.ListProcessedFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": len(files),
	})
}
