package queue

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type VideoMessage struct {
	IDVideo     string `json:"id_video"`
	Titulo      string `json:"titulo"`
	Autor       string `json:"autor"`
	Status      string `json:"status"`
	FilePath    string `json:"file_path"`
	DataCriacao string `json:"data_criacao"`
	DataUpload  string `json:"data_upload"`
	VideoKey    string // Campo derivado do file_path
	VideoID     string // Campo para o ReceiptHandle
}

type SQSService struct {
	sqsClient *sqs.SQS
	queueURL  string
}

func NewSQSService(queueURL string) (*SQSService, error) {
	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}

	// Manter host.docker.internal quando estiver rodando dentro do container
	// (será substituído pela variável de ambiente LOCALSTACK_ENDPOINT)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(endpoint), // LocalStack endpoint
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	})
	if err != nil {
		return nil, err
	}

	return &SQSService{
		sqsClient: sqs.New(sess),
		queueURL:  queueURL,
	}, nil
}

func (s *SQSService) ReceiveMessages() ([]*VideoMessage, error) {
	result, err := s.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20), // Long polling
	})
	if err != nil {
		return nil, err
	}

	var messages []*VideoMessage
	for _, msg := range result.Messages {
		var videoMsg VideoMessage
		if err := json.Unmarshal([]byte(*msg.Body), &videoMsg); err != nil {
			log.Printf("Erro ao deserializar mensagem: %v", err)
			continue
		}

		// Extrair a chave do vídeo do file_path
		// file_path formato: "s3://video-service-bucket/videos/uuid/filename.mp4"
		// queremos apenas: "videos/uuid/filename.mp4"
		if videoMsg.FilePath != "" {
			bucketPrefix := "s3://video-service-bucket/"
			if strings.HasPrefix(videoMsg.FilePath, bucketPrefix) {
				videoMsg.VideoKey = videoMsg.FilePath[len(bucketPrefix):]
			} else {
				videoMsg.VideoKey = videoMsg.FilePath // fallback se o formato for diferente
			}
		}

		// Armazenar o ReceiptHandle para poder deletar a mensagem depois
		videoMsg.VideoID = *msg.ReceiptHandle
		messages = append(messages, &videoMsg)
	}

	return messages, nil
}

func (s *SQSService) DeleteMessage(receiptHandle string) error {
	_, err := s.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

func (s *SQSService) SendMessage(videoKey string) error {
	message := VideoMessage{
		VideoKey: videoKey,
		VideoID:  videoKey, // ou gere um ID único
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = s.sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}

// GetClient retorna o cliente SQS para reutilização
func (s *SQSService) GetClient() *sqs.SQS {
	return s.sqsClient
}
