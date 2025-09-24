// cmd/api/main.go
package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Endpoint 1: Listar arquivos processados
	r.GET("/files", listFiles)

	// Endpoint 2: Download de arquivo
	r.GET("/download/:filename", downloadFile)

	r.Run(":8080")
}

func listFiles(c *gin.Context) {
	files, err := filepath.Glob("outputs/*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	var fileList []FileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileList = append(fileList, FileInfo{
			Name: filepath.Base(file),
			Size: info.Size(),
		})
	}

	c.JSON(http.StatusOK, fileList)
}

func downloadFile(c *gin.Context) {
	filename := c.Param("filename")

	// Validação básica de segurança
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome de arquivo inválido"})
		return
	}

	filePath := filepath.Join("outputs", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	c.File(filePath)
}
