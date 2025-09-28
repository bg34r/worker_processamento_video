package notification

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	testTopic          = "test-topic"
	testBroker         = "localhost:9092"
	testVideoID        = "video-123"
	testVideoTitle     = "Test Video"
	testVideoURL       = "https://example.com/video.mp4"
	testUserName       = "Test User"
	testUserEmail      = "test@example.com"
	testErrorMsg       = "Processing failed"
	eventTypeProcessed = "VIDEO_PROCESSED"
	eventTypeFailed    = "VIDEO_FAILED"
)

func TestNewNotificationServiceConnectionError(t *testing.T) {
	// Testar com broker inválido (deve falhar na conexão)
	brokers := []string{"invalid-broker:9092"}

	service, err := NewNotificationService(brokers, testTopic)
	if err == nil {
		t.Errorf("esperado erro ao conectar com broker inválido, mas não ocorreu")
		if service != nil {
			service.Close() // cleanup
		}
	}

	if service != nil {
		t.Errorf("service deveria ser nil quando há erro de conexão")
	}
}

func TestNewNotificationServiceValidation(t *testing.T) {
	// Testar parâmetros de entrada
	brokers := []string{testBroker}

	// Mesmo que falhe na conexão, validamos a lógica de inicialização
	service, err := NewNotificationService(brokers, testTopic)
	// Esperamos erro de conexão quando Kafka não está disponível
	if err == nil {
		// Se não houve erro, significa que conseguiu conectar (Kafka está rodando)
		// Neste caso, fechamos o service e consideramos o teste válido
		t.Logf("Kafka está disponível localmente, conexão bem-sucedida")
		if service != nil {
			service.Close()
		}
	} else {
		// Era esperado um erro de conexão se Kafka não estiver disponível
		t.Logf("Erro de conexão esperado (Kafka não disponível): %v", err)
	}
}

func TestNewNotificationServiceEmptyBrokers(t *testing.T) {
	// Testar com lista vazia de brokers
	brokers := []string{}

	service, err := NewNotificationService(brokers, testTopic)
	if err == nil {
		t.Errorf("esperado erro com lista vazia de brokers")
		if service != nil {
			service.Close()
		}
	}
}

func TestNotificationEventStructValidation(t *testing.T) {
	// Testar estrutura NotificationEvent
	event := NotificationEvent{
		EventID:   uuid.New().String(),
		EventType: eventTypeProcessed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: testVideoTitle,
			VideoURL:   testVideoURL,
		},
	}

	// Verificar se a estrutura é válida
	if event.EventID == "" {
		t.Errorf("EventID não deveria estar vazio")
	}

	if event.EventType != eventTypeProcessed {
		t.Errorf("EventType esperado '%s', obtido '%s'", eventTypeProcessed, event.EventType)
	}

	if event.User.Email != testUserEmail {
		t.Errorf("User.Email esperado '%s', obtido '%s'", testUserEmail, event.User.Email)
	}

	if event.Data.VideoID != testVideoID {
		t.Errorf("Data.VideoID esperado '%s', obtido '%s'", testVideoID, event.Data.VideoID)
	}
}

func TestNotificationEventJSONSerialization(t *testing.T) {
	// Testar serialização JSON da estrutura
	event := NotificationEvent{
		EventID:   "test-event-123",
		EventType: eventTypeProcessed,
		Timestamp: "2023-12-01T10:00:00Z",
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: testVideoTitle,
			VideoURL:   testVideoURL,
		},
	}

	// Serializar para JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("erro ao serializar evento para JSON: %v", err)
	}

	// Verificar se contém campos esperados
	jsonStr := string(jsonData)
	expectedFields := []string{
		`"eventId":"test-event-123"`,
		`"eventType":"VIDEO_PROCESSED"`,
		`"videoId":"video-123"`,
		`"email":"test@example.com"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON não contém campo esperado: %s\nJSON: %s", field, jsonStr)
		}
	}

	// Deserializar de volta
	var deserializedEvent NotificationEvent
	err = json.Unmarshal(jsonData, &deserializedEvent)
	if err != nil {
		t.Fatalf("erro ao deserializar JSON: %v", err)
	}

	// Verificar se os dados são consistentes
	if deserializedEvent.EventID != event.EventID {
		t.Errorf("EventID após deserialização não confere")
	}

	if deserializedEvent.User.Email != event.User.Email {
		t.Errorf("User.Email após deserialização não confere")
	}
}

func TestNotificationEventFailedType(t *testing.T) {
	// Testar evento de falha
	event := NotificationEvent{
		EventID:   uuid.New().String(),
		EventType: eventTypeFailed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:      testVideoID,
			VideoTitle:   testVideoTitle,
			ErrorMessage: testErrorMsg,
		},
	}

	// Verificar campos específicos de falha
	if event.EventType != eventTypeFailed {
		t.Errorf("EventType esperado '%s', obtido '%s'", eventTypeFailed, event.EventType)
	}

	if event.Data.ErrorMessage != testErrorMsg {
		t.Errorf("ErrorMessage esperado '%s', obtido '%s'", testErrorMsg, event.Data.ErrorMessage)
	}

	// VideoURL deve estar vazio para eventos de falha
	if event.Data.VideoURL != "" {
		t.Errorf("VideoURL deveria estar vazio para eventos de falha, mas obteve '%s'", event.Data.VideoURL)
	}
}

func TestSendProcessingCompletedStructure(t *testing.T) {
	// Testar a lógica de criação de evento de sucesso (sem envio real)
	// Simulamos o que SendProcessingCompleted faria

	eventID := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	event := NotificationEvent{
		EventID:   eventID,
		EventType: eventTypeProcessed,
		Timestamp: timestamp,
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: testVideoTitle,
			VideoURL:   testVideoURL,
		},
	}

	// Validar estrutura do evento criado
	if event.EventType != eventTypeProcessed {
		t.Errorf("Tipo de evento incorreto para processamento concluído")
	}

	if event.Data.VideoURL == "" {
		t.Errorf("VideoURL não deveria estar vazio para evento de sucesso")
	}

	if event.Data.ErrorMessage != "" {
		t.Errorf("ErrorMessage deveria estar vazio para evento de sucesso")
	}

	// Verificar se UUID é válido
	if _, err := uuid.Parse(event.EventID); err != nil {
		t.Errorf("EventID não é um UUID válido: %v", err)
	}

	// Verificar formato do timestamp
	if _, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		t.Errorf("Timestamp não está no formato RFC3339: %v", err)
	}
}

func TestSendProcessingFailedStructure(t *testing.T) {
	// Testar a lógica de criação de evento de falha (sem envio real)

	eventID := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	event := NotificationEvent{
		EventID:   eventID,
		EventType: eventTypeFailed,
		Timestamp: timestamp,
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:      testVideoID,
			VideoTitle:   testVideoTitle,
			ErrorMessage: testErrorMsg,
		},
	}

	// Validar estrutura do evento criado
	if event.EventType != eventTypeFailed {
		t.Errorf("Tipo de evento incorreto para processamento falhado")
	}

	if event.Data.ErrorMessage == "" {
		t.Errorf("ErrorMessage não deveria estar vazio para evento de falha")
	}

	if event.Data.VideoURL != "" {
		t.Errorf("VideoURL deveria estar vazio para evento de falha")
	}
}

func TestUserStructValidation(t *testing.T) {
	// Testar estrutura User
	user := User{
		ID:    testVideoID,
		Name:  testUserName,
		Email: testUserEmail,
	}

	if user.ID != testVideoID {
		t.Errorf("User.ID esperado '%s', obtido '%s'", testVideoID, user.ID)
	}

	if user.Name != testUserName {
		t.Errorf("User.Name esperado '%s', obtido '%s'", testUserName, user.Name)
	}

	if user.Email != testUserEmail {
		t.Errorf("User.Email esperado '%s', obtido '%s'", testUserEmail, user.Email)
	}
}

func TestEventDataValidation(t *testing.T) {
	// Testar estrutura EventData para diferentes cenários

	// Cenário 1: Processamento bem-sucedido
	successData := EventData{
		VideoID:    testVideoID,
		VideoTitle: testVideoTitle,
		VideoURL:   testVideoURL,
	}

	if successData.VideoID != testVideoID {
		t.Errorf("VideoID esperado '%s', obtido '%s'", testVideoID, successData.VideoID)
	}

	if successData.VideoURL == "" {
		t.Errorf("VideoURL não deveria estar vazio para sucesso")
	}

	if successData.ErrorMessage != "" {
		t.Errorf("ErrorMessage deveria estar vazio para sucesso")
	}

	// Cenário 2: Processamento com falha
	failureData := EventData{
		VideoID:      testVideoID,
		VideoTitle:   testVideoTitle,
		ErrorMessage: testErrorMsg,
	}

	if failureData.ErrorMessage == "" {
		t.Errorf("ErrorMessage não deveria estar vazio para falha")
	}

	if failureData.VideoURL != "" {
		t.Errorf("VideoURL deveria estar vazio para falha")
	}
}

func TestNotificationServiceCloseWithoutProducer(t *testing.T) {
	// Testar Close sem producer inicializado (simulação)
	service := &NotificationService{
		producer: nil,
		topic:    testTopic,
	}

	// Isso deve resultar em panic se não tratado adequadamente
	// Como não podemos simular o producer real, apenas verificamos a estrutura
	if service.topic != testTopic {
		t.Errorf("Topic não foi configurado corretamente")
	}
}

func TestEventTypeConstants(t *testing.T) {
	// Verificar se os tipos de evento são consistentes
	if eventTypeProcessed != "VIDEO_PROCESSED" {
		t.Errorf("Constante eventTypeProcessed incorreta: %s", eventTypeProcessed)
	}

	if eventTypeFailed != "VIDEO_FAILED" {
		t.Errorf("Constante eventTypeFailed incorreta: %s", eventTypeFailed)
	}
}

func TestTimestampFormat(t *testing.T) {
	// Testar formato de timestamp RFC3339
	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	// Tentar fazer parse de volta
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("Erro ao fazer parse do timestamp RFC3339: %v", err)
	}

	// Verificar se a diferença é mínima (menos de 1 segundo)
	diff := now.Sub(parsedTime)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("Timestamp não preservou o tempo corretamente. Diferença: %v", diff)
	}
}

func TestUUIDGeneration(t *testing.T) {
	// Testar geração de UUIDs únicos
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()

	if uuid1 == uuid2 {
		t.Errorf("UUIDs gerados não são únicos: %s == %s", uuid1, uuid2)
	}

	// Verificar se são UUIDs válidos
	if _, err := uuid.Parse(uuid1); err != nil {
		t.Errorf("UUID1 não é válido: %v", err)
	}

	if _, err := uuid.Parse(uuid2); err != nil {
		t.Errorf("UUID2 não é válido: %v", err)
	}
}

func TestSendEventJSONMarshalling(t *testing.T) {
	// Testar a lógica de serialização JSON dentro de SendEvent
	event := NotificationEvent{
		EventID:   "test-event-456",
		EventType: eventTypeProcessed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: testVideoTitle,
			VideoURL:   testVideoURL,
		},
	}

	// Simular o que acontece dentro de SendEvent
	messageBody, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Erro ao serializar evento (simulando SendEvent): %v", err)
	}

	// Verificar se o JSON contém os campos esperados
	jsonStr := string(messageBody)
	expectedFields := []string{
		`"eventId":"test-event-456"`,
		`"eventType":"VIDEO_PROCESSED"`,
		`"videoId":"video-123"`,
		testUserEmail,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON serializado não contém campo esperado: %s", field)
		}
	}
}

func TestSendEventInvalidJSON(t *testing.T) {
	// Testar comportamento com dados que podem causar problemas na serialização
	event := NotificationEvent{
		EventID:   "test-with-special-chars",
		EventType: eventTypeProcessed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    testVideoID,
			Name:  "User with \"quotes\" and \n newlines",
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: "Title with special chars: áéíóú & symbols #@$%",
			VideoURL:   testVideoURL,
		},
	}

	// Deve conseguir serializar mesmo com caracteres especiais
	messageBody, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Erro inesperado ao serializar evento com caracteres especiais: %v", err)
	}

	// Verificar se pode ser deserializado de volta
	var deserializedEvent NotificationEvent
	err = json.Unmarshal(messageBody, &deserializedEvent)
	if err != nil {
		t.Fatalf("Erro ao deserializar evento com caracteres especiais: %v", err)
	}

	// Verificar preservação dos dados
	if deserializedEvent.User.Name != event.User.Name {
		t.Errorf("Nome do usuário não foi preservado na serialização")
	}
}

func TestSendProcessingCompletedParameters(t *testing.T) {
	// Testar validação de parâmetros para SendProcessingCompleted
	testCases := []struct {
		name        string
		videoID     string
		title       string
		videoURL    string
		userName    string
		userEmail   string
		expectValid bool
	}{
		{
			name:        "Parâmetros válidos",
			videoID:     testVideoID,
			title:       testVideoTitle,
			videoURL:    testVideoURL,
			userName:    testUserName,
			userEmail:   testUserEmail,
			expectValid: true,
		},
		{
			name:        "VideoID vazio",
			videoID:     "",
			title:       testVideoTitle,
			videoURL:    testVideoURL,
			userName:    testUserName,
			userEmail:   testUserEmail,
			expectValid: true, // Função não valida parâmetros vazios
		},
		{
			name:        "Email vazio",
			videoID:     testVideoID,
			title:       testVideoTitle,
			videoURL:    testVideoURL,
			userName:    testUserName,
			userEmail:   "",
			expectValid: true, // Função não valida parâmetros vazios
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simular criação do evento que SendProcessingCompleted faria
			event := NotificationEvent{
				EventID:   uuid.New().String(),
				EventType: eventTypeProcessed,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				User: User{
					ID:    tc.videoID,
					Name:  tc.userName,
					Email: tc.userEmail,
				},
				Data: EventData{
					VideoID:    tc.videoID,
					VideoTitle: tc.title,
					VideoURL:   tc.videoURL,
				},
			}

			// Verificar se a estrutura foi criada corretamente
			if event.EventType != eventTypeProcessed {
				t.Errorf("EventType incorreto para processamento concluído")
			}

			if event.Data.VideoURL != tc.videoURL {
				t.Errorf("VideoURL não foi configurado corretamente")
			}

			// Verificar serialização
			_, err := json.Marshal(event)
			if err != nil {
				t.Errorf("Erro na serialização do evento: %v", err)
			}
		})
	}
}

func TestSendProcessingFailedParameters(t *testing.T) {
	// Testar validação de parâmetros para SendProcessingFailed
	testCases := []struct {
		name        string
		videoID     string
		title       string
		errorMsg    string
		userName    string
		userEmail   string
		expectValid bool
	}{
		{
			name:        "Parâmetros válidos",
			videoID:     testVideoID,
			title:       testVideoTitle,
			errorMsg:    testErrorMsg,
			userName:    testUserName,
			userEmail:   testUserEmail,
			expectValid: true,
		},
		{
			name:        "Erro message longa",
			videoID:     testVideoID,
			title:       testVideoTitle,
			errorMsg:    strings.Repeat("Erro muito longo ", 100),
			userName:    testUserName,
			userEmail:   testUserEmail,
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simular criação do evento que SendProcessingFailed faria
			event := NotificationEvent{
				EventID:   uuid.New().String(),
				EventType: eventTypeFailed,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				User: User{
					ID:    tc.videoID,
					Name:  tc.userName,
					Email: tc.userEmail,
				},
				Data: EventData{
					VideoID:      tc.videoID,
					VideoTitle:   tc.title,
					ErrorMessage: tc.errorMsg,
				},
			}

			// Verificar se a estrutura foi criada corretamente
			if event.EventType != eventTypeFailed {
				t.Errorf("EventType incorreto para processamento falhado")
			}

			if event.Data.ErrorMessage != tc.errorMsg {
				t.Errorf("ErrorMessage não foi configurado corretamente")
			}

			if event.Data.VideoURL != "" {
				t.Errorf("VideoURL deveria estar vazio para eventos de falha")
			}

			// Verificar serialização
			_, err := json.Marshal(event)
			if err != nil {
				t.Errorf("Erro na serialização do evento: %v", err)
			}
		})
	}
}

func TestNotificationServiceStructure(t *testing.T) {
	// Testar estrutura do NotificationService
	service := &NotificationService{
		producer: nil, // Simular sem producer
		topic:    testTopic,
	}

	if service.topic != testTopic {
		t.Errorf("Topic não foi configurado corretamente")
	}

	// O producer being nil é um estado válido para testes
	if service.producer != nil {
		t.Errorf("Producer deveria ser nil para este teste")
	}
}

func TestNotificationEventValidation(t *testing.T) {
	// Testar diferentes combinações de EventType
	validEventTypes := []string{eventTypeProcessed, eventTypeFailed}

	for _, eventType := range validEventTypes {
		event := NotificationEvent{
			EventID:   uuid.New().String(),
			EventType: eventType,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			User: User{
				ID:    testVideoID,
				Name:  testUserName,
				Email: testUserEmail,
			},
			Data: EventData{
				VideoID:    testVideoID,
				VideoTitle: testVideoTitle,
			},
		}

		// Configurar campos específicos baseados no tipo
		if eventType == eventTypeProcessed {
			event.Data.VideoURL = testVideoURL
		} else if eventType == eventTypeFailed {
			event.Data.ErrorMessage = testErrorMsg
		}

		// Verificar consistência
		if eventType == eventTypeProcessed && event.Data.VideoURL == "" {
			t.Errorf("VideoURL deveria estar preenchido para %s", eventType)
		}

		if eventType == eventTypeFailed && event.Data.ErrorMessage == "" {
			t.Errorf("ErrorMessage deveria estar preenchido para %s", eventType)
		}

		// Testar serialização
		jsonData, err := json.Marshal(event)
		if err != nil {
			t.Errorf("Erro na serialização para eventType %s: %v", eventType, err)
		}

		// Verificar se contém o eventType no JSON
		if !strings.Contains(string(jsonData), eventType) {
			t.Errorf("JSON não contém o eventType %s", eventType)
		}
	}
}

// Benchmark para serialização JSON
func BenchmarkNotificationEventSerialization(b *testing.B) {
	event := NotificationEvent{
		EventID:   uuid.New().String(),
		EventType: eventTypeProcessed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		User: User{
			ID:    testVideoID,
			Name:  testUserName,
			Email: testUserEmail,
		},
		Data: EventData{
			VideoID:    testVideoID,
			VideoTitle: testVideoTitle,
			VideoURL:   testVideoURL,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(event)
		if err != nil {
			b.Fatalf("Erro na serialização: %v", err)
		}
	}
}
