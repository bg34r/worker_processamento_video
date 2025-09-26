package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
	"worker/internal/infrastructure/notification"
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

	// Configurações via variáveis de ambiente
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = os.Getenv("S3_BUCKET") // Fallback para S3_BUCKET
	}

	defaultUserEmail := os.Getenv("DEFAULT_USER_EMAIL")
	if defaultUserEmail == "" {
		defaultUserEmail = "bruno@fiap.com.br" // Fallback padrão
	}
	if bucketName == "" {
		bucketName = "video-service-bucket"
	}

	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		queueURL = "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/video-processing-queue"
	}

	// Configurações Kafka para notificações
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "video-events"
	}

	localstackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if localstackEndpoint == "" {
		localstackEndpoint = "http://localhost:4566"
	}

	log.Printf("Configuração:")
	log.Printf("- Bucket S3: %s", bucketName)
	log.Printf("- Fila SQS: %s", queueURL)
	log.Printf("- Kafka Brokers: %s", kafkaBrokers)
	log.Printf("- Kafka Topic: %s", kafkaTopic)
	log.Printf("- LocalStack: %s", localstackEndpoint)

	// Inicializar serviços
	s3Service, err := storage.NewS3Service(bucketName)
	if err != nil {
		log.Fatalf("Erro ao inicializar S3: %v", err)
	}

	sqsService, err := queue.NewSQSService(queueURL)
	if err != nil {
		log.Fatalf("Erro ao inicializar SQS: %v", err)
	}

	// Inicializar serviço de notificação Kafka
	brokers := []string{kafkaBrokers}
	notificationService, err := notification.NewNotificationService(brokers, kafkaTopic)
	if err != nil {
		log.Fatalf("Erro ao inicializar serviço de notificação: %v", err)
	}
	defer notificationService.Close()

	extractor := video.NewFFmpegExtractor()
	zipService := storage.NewZipService()

	log.Println("Worker iniciado - monitorando fila SQS...")

	for {
		processMessages(s3Service, sqsService, extractor, zipService, notificationService, defaultUserEmail)
		time.Sleep(5 * time.Second) // Verifica a cada 5 segundos
	}
}

func processMessages(s3Service *storage.S3Service, sqsService *queue.SQSService,
	extractor video.FrameExtractor, zipService storage.ZipService, notificationService *notification.NotificationService, defaultUserEmail string) {

	messages, err := sqsService.ReceiveMessages()
	if err != nil {
		log.Printf("Erro ao receber mensagens: %v", err)
		return
	}

	for _, msg := range messages {
		if err := processVideoMessage(msg, s3Service, sqsService, extractor, zipService, notificationService, defaultUserEmail); err != nil {
			log.Printf("Erro ao processar mensagem %s: %v", msg.VideoKey, err)
		}
	}
}

func processVideoMessage(msg *queue.VideoMessage, s3Service *storage.S3Service,
	sqsService *queue.SQSService, extractor video.FrameExtractor, zipService storage.ZipService,
	notificationService *notification.NotificationService, defaultUserEmail string) error {

	log.Printf("Processando vídeo: %s", msg.VideoKey)

	// Baixar vídeo do S3
	localVideoPath := filepath.Join("temp", filepath.Base(msg.VideoKey))
	if err := s3Service.DownloadVideo(msg.VideoKey, localVideoPath); err != nil {
		notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao baixar vídeo do S3: "+err.Error(), msg.Autor, defaultUserEmail)
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
		notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao extrair frames: "+err.Error(), msg.Autor, defaultUserEmail)
		return err
	}

	if len(frames) == 0 {
		log.Printf("Nenhum frame extraído para %s", msg.VideoKey)
		notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Nenhum frame foi extraído do vídeo", msg.Autor, defaultUserEmail)
		return nil
	}

	// Criar ZIP
	zipName := filepath.Base(msg.VideoKey)
	zipName = zipName[:len(zipName)-len(filepath.Ext(zipName))] + "_frames.zip"
	localZipPath := filepath.Join("outputs", zipName)

	if err := zipService.CreateZipFile(frames, localZipPath); err != nil {
		notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao criar arquivo ZIP: "+err.Error(), msg.Autor, defaultUserEmail)
		return err
	}
	defer os.Remove(localZipPath)

	// Upload ZIP para S3
	s3ZipKey := "processed/" + zipName
	if err := s3Service.UploadZip(localZipPath, s3ZipKey); err != nil {
		notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao fazer upload do ZIP: "+err.Error(), msg.Autor, defaultUserEmail)
		return err
	}

	// Enviar notificação de sucesso
	videoURL := "s3://video-service-bucket/" + s3ZipKey // URL do arquivo processado
	if err := notificationService.SendProcessingCompleted(msg.IDVideo, msg.Titulo, videoURL, msg.Autor, defaultUserEmail); err != nil {
		log.Printf("Erro ao enviar notificação de sucesso: %v", err)
	}

	// Deletar mensagem da fila após processamento bem-sucedido
	if err := sqsService.DeleteMessage(msg.VideoID); err != nil {
		log.Printf("Erro ao deletar mensagem da fila: %v", err)
	}

	log.Printf("Processamento concluído: %s -> %s", msg.VideoKey, s3ZipKey)
	return nil
}
