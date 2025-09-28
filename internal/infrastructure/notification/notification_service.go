package notification

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// User representa as informações do usuário na notificação
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// EventData representa os dados específicos do evento
type EventData struct {
	VideoID      string `json:"videoId"`
	VideoTitle   string `json:"videoTitle"`
	VideoURL     string `json:"videoUrl,omitempty"`     // opcional para VIDEO_PROCESSED
	ErrorMessage string `json:"errorMessage,omitempty"` // apenas para VIDEO_FAILED
}

// NotificationEvent representa a estrutura completa da notificação para Kafka
type NotificationEvent struct {
	EventID   string    `json:"eventId"`
	EventType string    `json:"eventType"` // VIDEO_PROCESSED | VIDEO_FAILED
	Timestamp string    `json:"timestamp"`
	User      User      `json:"user"`
	Data      EventData `json:"data"`
}

// NotificationService gerencia o envio de notificações para Kafka
type NotificationService struct {
	producer sarama.SyncProducer
	topic    string
}

// NewNotificationService cria uma nova instância do serviço de notificação Kafka
func NewNotificationService(brokers []string, topic string) (*NotificationService, error) {
	log.Printf("🔧 Inicializando Kafka com brokers: %v", brokers)

	// Tentar resolver o problema de localhost->IPv6
	for i, broker := range brokers {
		log.Printf("🔍 Broker original: %s", broker)
		brokers[i] = broker // Manter como está por enquanto
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Version = sarama.V2_8_0_0

	// Configurações específicas para resolver o problema de redirecionamento
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.MaxMessageBytes = 1000000

	log.Printf("🔗 Tentando conectar aos brokers Kafka...")
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		producer: producer,
		topic:    topic,
	}, nil
}

// Close fecha o producer Kafka
func (ns *NotificationService) Close() error {
	return ns.producer.Close()
}

// SendEvent envia um evento para o Kafka
func (ns *NotificationService) SendEvent(event NotificationEvent) error {
	log.Printf("📤 Enviando evento %s para tópico %s", event.EventType, ns.topic)

	// Converter o evento para JSON
	messageBody, err := json.Marshal(event)
	if err != nil {
		log.Printf("Erro ao serializar evento: %v", err)
		return err
	}

	log.Printf("📋 Payload: %s", string(messageBody))

	// Criar mensagem Kafka
	msg := &sarama.ProducerMessage{
		Topic: ns.topic,
		Key:   sarama.StringEncoder(event.EventID),
		Value: sarama.StringEncoder(messageBody),
	}

	log.Printf("🚀 Tentando enviar para Kafka...")
	// Enviar mensagem
	partition, offset, err := ns.producer.SendMessage(msg)
	if err != nil {
		log.Printf("❌ Erro ao enviar evento para Kafka: %v", err)
		return err
	}

	log.Printf("✅ Evento enviado: %s -> partition=%d, offset=%d", event.EventType, partition, offset)
	return nil
}

// SendProcessingCompleted envia notificação de processamento concluído
func (ns *NotificationService) SendProcessingCompleted(videoID, videoTitle, videoURL, userName, userEmail string) error {
	event := NotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "VIDEO_PROCESSED",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    videoID, // Usando videoID como user ID (adapte conforme necessário)
			Name:  userName,
			Email: userEmail,
		},
		Data: EventData{
			VideoID:    videoID,
			VideoTitle: videoTitle,
			VideoURL:   videoURL,
		},
	}
	return ns.SendEvent(event)
}

// SendProcessingFailed envia notificação de erro no processamento
func (ns *NotificationService) SendProcessingFailed(videoID, videoTitle, errorMessage, userName, userEmail string) error {
	event := NotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "VIDEO_FAILED",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    videoID, // Usando videoID como user ID (adapte conforme necessário)
			Name:  userName,
			Email: userEmail,
		},
		Data: EventData{
			VideoID:      videoID,
			VideoTitle:   videoTitle,
			ErrorMessage: errorMessage,
		},
	}
	return ns.SendEvent(event)
}
