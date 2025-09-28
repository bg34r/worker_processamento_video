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

// Helper function para enviar notificação apenas se o serviço estiver disponível
func sendNotificationIfAvailable(ns *notification.NotificationService, fn func() error) {
	if ns != nil {
		if err := fn(); err != nil {
			log.Printf("⚠️ Erro ao enviar notificação: %v", err)
		}
	}
}

// Helper function para obter o email correto da mensagem ou usar o padrão
func getUserEmail(msg *queue.VideoMessage, defaultEmail string) string {
	if msg.Email != "" {
		return msg.Email
	}
	return defaultEmail
}

func main() {
	// Identificação do worker
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = "1"
	}

	workerName := os.Getenv("WORKER_NAME")
	if workerName == "" {
		workerName = "video-worker-" + workerID
	}

	log.Printf("🚀 %s (ID: %s) iniciando...", workerName, workerID)

	// Criar diretórios necessários com identificação do worker
	dirs := []string{
		"temp/worker-" + workerID,
		"outputs/worker-" + workerID,
	}
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

	log.Printf("📋 Configuração %s:", workerName)
	log.Printf("- Worker ID: %s", workerID)
	log.Printf("- Worker Name: %s", workerName)
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

	// Inicializar serviço de notificação Kafka (opcional)
	var notificationService *notification.NotificationService
	brokers := []string{kafkaBrokers}

	notificationService, err = notification.NewNotificationService(brokers, kafkaTopic)
	if err != nil {
		log.Printf("⚠️ Kafka não disponível - continuando sem notificações: %v", err)
		notificationService = nil
	} else {
		log.Printf("✅ Kafka conectado com sucesso!")
		defer notificationService.Close()
	}

	extractor := video.NewFFmpegExtractor()
	zipService := storage.NewZipService()

	log.Printf("✅ %s iniciado - monitorando fila SQS...", workerName)

	for {
		processMessages(s3Service, sqsService, extractor, zipService, notificationService, defaultUserEmail, workerID, workerName)
		time.Sleep(5 * time.Second) // Verifica a cada 5 segundos
	}
}

func processMessages(s3Service *storage.S3Service, sqsService *queue.SQSService,
	extractor video.FrameExtractor, zipService storage.ZipService, notificationService *notification.NotificationService,
	defaultUserEmail, workerID, workerName string) {

	messages, err := sqsService.ReceiveMessages()
	if err != nil {
		log.Printf("🔴 %s - Erro ao receber mensagens: %v", workerName, err)
		return
	}

	if len(messages) > 0 {
		log.Printf("📬 %s - Recebidas %d mensagens para processamento", workerName, len(messages))
	}

	for _, msg := range messages {
		// Deletar mensagem IMEDIATAMENTE para evitar duplicação entre workers
		if err := sqsService.DeleteMessage(msg.VideoID); err != nil {
			log.Printf("⚠️ %s - Erro ao deletar mensagem da fila: %v", workerName, err)
			continue
		}
		log.Printf("🔒 %s - Mensagem reservada para processamento: %s", workerName, msg.VideoKey)

		if err := processVideoMessage(msg, s3Service, sqsService, extractor, zipService, notificationService, defaultUserEmail, workerID, workerName); err != nil {
			log.Printf("🔴 %s - Erro ao processar mensagem %s: %v", workerName, msg.VideoKey, err)
		}
	}
}

func processVideoMessage(msg *queue.VideoMessage, s3Service *storage.S3Service,
	sqsService *queue.SQSService, extractor video.FrameExtractor, zipService storage.ZipService,
	notificationService *notification.NotificationService, defaultUserEmail, workerID, workerName string) error {

	log.Printf("🎬 %s - Processando vídeo: %s", workerName, msg.VideoKey)

	// Criar caminhos únicos para evitar conflitos entre workers
	workerTempDir := filepath.Join("temp", "worker-"+workerID)
	workerOutputDir := filepath.Join("outputs", "worker-"+workerID)

	// Garantir que os diretórios existem
	os.MkdirAll(workerTempDir, 0755)
	os.MkdirAll(workerOutputDir, 0755)

	// Baixar vídeo do S3 em diretório específico do worker
	localVideoPath := filepath.Join(workerTempDir, filepath.Base(msg.VideoKey))
	if err := s3Service.DownloadVideo(msg.VideoKey, localVideoPath); err != nil {
		log.Printf("🔴 %s - Erro ao baixar vídeo do S3: %v", workerName, err)
		sendNotificationIfAvailable(notificationService, func() error {
			return notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao baixar vídeo do S3: "+err.Error(), msg.Autor, getUserEmail(msg, defaultUserEmail))
		})
		return err
	}
	defer os.Remove(localVideoPath)

	// Criar diretório temporário para frames específico do worker
	framesDir := filepath.Join(workerTempDir, "frames_"+msg.VideoID)
	os.MkdirAll(framesDir, 0755)
	defer os.RemoveAll(framesDir)

	// Extrair frames
	log.Printf("⚙️ %s - Extraindo frames do vídeo: %s", workerName, msg.VideoKey)
	frames, err := extractor.ExtractFrames(localVideoPath, framesDir)
	if err != nil {
		log.Printf("🔴 %s - Erro ao extrair frames: %v", workerName, err)
		sendNotificationIfAvailable(notificationService, func() error {
			return notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao extrair frames: "+err.Error(), msg.Autor, getUserEmail(msg, defaultUserEmail))
		})
		return err
	}

	if len(frames) == 0 {
		log.Printf("⚠️ %s - Nenhum frame extraído para %s", workerName, msg.VideoKey)
		sendNotificationIfAvailable(notificationService, func() error {
			return notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Nenhum frame foi extraído do vídeo", msg.Autor, getUserEmail(msg, defaultUserEmail))
		})
		return nil
	}

	log.Printf("✅ %s - Extraídos %d frames do vídeo: %s", workerName, len(frames), msg.VideoKey)

	// Criar ZIP em diretório específico do worker
	zipName := filepath.Base(msg.VideoKey)
	zipName = zipName[:len(zipName)-len(filepath.Ext(zipName))] + "_frames.zip"
	localZipPath := filepath.Join(workerOutputDir, zipName)

	log.Printf("📦 %s - Criando arquivo ZIP: %s", workerName, zipName)
	if err := zipService.CreateZipFile(frames, localZipPath); err != nil {
		log.Printf("🔴 %s - Erro ao criar arquivo ZIP: %v", workerName, err)
		sendNotificationIfAvailable(notificationService, func() error {
			return notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao criar arquivo ZIP: "+err.Error(), msg.Autor, getUserEmail(msg, defaultUserEmail))
		})
		return err
	}
	defer os.Remove(localZipPath)

	// Upload ZIP para S3
	s3ZipKey := "processed/" + zipName
	log.Printf("📤 %s - Fazendo upload do ZIP para S3: %s", workerName, s3ZipKey)
	if err := s3Service.UploadZip(localZipPath, s3ZipKey); err != nil {
		log.Printf("🔴 %s - Erro ao fazer upload do ZIP: %v", workerName, err)
		sendNotificationIfAvailable(notificationService, func() error {
			return notificationService.SendProcessingFailed(msg.IDVideo, msg.Titulo, "Erro ao fazer upload do ZIP: "+err.Error(), msg.Autor, getUserEmail(msg, defaultUserEmail))
		})
		return err
	}

	// Enviar notificação de sucesso
	videoURL := "s3://video-service-bucket/" + s3ZipKey // URL do arquivo processado
	log.Printf("📢 %s - Enviando notificação de sucesso para: %s", workerName, msg.VideoKey)
	sendNotificationIfAvailable(notificationService, func() error {
		// Usar msg.Email da fila SQS, com fallback para defaultUserEmail se vazio
		userEmail := msg.Email
		if userEmail == "" {
			userEmail = defaultUserEmail
		}
		return notificationService.SendProcessingCompleted(msg.IDVideo, msg.Titulo, videoURL, msg.Autor, userEmail)
	})

	log.Printf("🎉 %s - Processamento concluído: %s -> %s", workerName, msg.VideoKey, s3ZipKey)
	return nil
}
