package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
	"worker/internal/infrastructure/queue"
	"worker/internal/infrastructure/storage"
	"worker/internal/infrastructure/video"
)

func main() {
	// Criar diretórios necessários
	dirs := []string{"temp", "outputs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Erro ao criar diretório %s: %v", dir, err)
		}
	}

	// Inicializar serviços
	s3Service, err := storage.NewS3Service("video-bucket")
	if err != nil {
		log.Fatalf("Erro ao inicializar S3: %v", err)
	}

	sqsService, err := queue.NewSQSService("http://localhost:4566/000000000000/video-processing-queue")
	if err != nil {
		log.Fatalf("Erro ao inicializar SQS: %v", err)
	}

	extractor := video.NewFFmpegExtractor()
	zipService := storage.NewZipService()

	log.Println("Worker iniciado - monitorando fila SQS...")

	for {
		processMessages(s3Service, sqsService, extractor, zipService)
		time.Sleep(5 * time.Second) // Verifica a cada 5 segundos
	}
}

func processMessages(s3Service *storage.S3Service, sqsService *queue.SQSService,
	extractor video.FrameExtractor, zipService storage.ZipService) {

	messages, err := sqsService.ReceiveMessages()
	if err != nil {
		log.Printf("Erro ao receber mensagens: %v", err)
		return
	}

	for _, msg := range messages {
		if err := processVideoMessage(msg, s3Service, sqsService, extractor, zipService); err != nil {
			log.Printf("Erro ao processar mensagem %s: %v", msg.VideoKey, err)
		}
	}
}

func processVideoMessage(msg *queue.VideoMessage, s3Service *storage.S3Service,
	sqsService *queue.SQSService, extractor video.FrameExtractor, zipService storage.ZipService) error {

	log.Printf("Processando vídeo: %s", msg.VideoKey)

	// Baixar vídeo do S3
	localVideoPath := filepath.Join("temp", filepath.Base(msg.VideoKey))
	if err := s3Service.DownloadVideo(msg.VideoKey, localVideoPath); err != nil {
		return err
	}
	defer os.Remove(localVideoPath)

	// Criar diretório temporário para frames
	framesDir := filepath.Join("temp", "frames_"+msg.VideoID)
	os.MkdirAll(framesDir, 0755)
	defer os.RemoveAll(framesDir)

	// Extrair frames
	frames, err := extractor.ExtractFrames(localVideoPath, framesDir)
	if err != nil {
		return err
	}

	if len(frames) == 0 {
		log.Printf("Nenhum frame extraído para %s", msg.VideoKey)
		return nil
	}

	// Criar ZIP
	zipName := filepath.Base(msg.VideoKey)
	zipName = zipName[:len(zipName)-len(filepath.Ext(zipName))] + "_frames.zip"
	localZipPath := filepath.Join("outputs", zipName)

	if err := zipService.CreateZipFile(frames, localZipPath); err != nil {
		return err
	}
	defer os.Remove(localZipPath)

	// Upload ZIP para S3
	s3ZipKey := "processed/" + zipName
	if err := s3Service.UploadZip(localZipPath, s3ZipKey); err != nil {
		return err
	}

	// Deletar mensagem da fila após processamento bem-sucedido
	if err := sqsService.DeleteMessage(msg.VideoID); err != nil {
		log.Printf("Erro ao deletar mensagem da fila: %v", err)
	}

	log.Printf("Processamento concluído: %s -> %s", msg.VideoKey, s3ZipKey)
	return nil
}
