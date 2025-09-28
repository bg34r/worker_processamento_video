package queue

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const (
	testQueueURL      = "http://localhost:4566/000000000000/test-queue"
	testEndpoint      = "http://localhost:4566"
	testVideoKey      = "videos/test-uuid/test-video.mp4"
	testReceiptHandle = "test-receipt-handle-12345"
	testFilePath      = "s3://video-service-bucket/videos/test-uuid/test-video.mp4"
	testVideoTitle    = "Test Video Title"
	testAuthor        = "Test Author"
	testEmail         = "test@example.com"
	testUsername      = "testuser"
	testDataCriacao   = "2023-12-01T10:00:00Z"
	testDataUpload    = "2023-12-01T10:05:00Z"
	errorCreateSQS    = "erro ao criar SQSService: %v"
)

func TestNewSQSServiceSuccess(t *testing.T) {
	// Configurar endpoint de teste
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("esperado sucesso ao criar SQSService, mas ocorreu erro: %v", err)
	}

	if service == nil {
		t.Fatalf("esperado service não nulo")
	}

	if service.queueURL != testQueueURL {
		t.Errorf("esperado queueURL '%s', mas obteve '%s'", testQueueURL, service.queueURL)
	}

	if service.sqsClient == nil {
		t.Errorf("sqsClient não foi inicializado")
	}
}

func TestNewSQSServiceWithoutEndpointEnv(t *testing.T) {
	// Limpar variável de ambiente para testar fallback
	originalEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	os.Unsetenv("LOCALSTACK_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("LOCALSTACK_ENDPOINT", originalEndpoint)
		}
	}()

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("esperado sucesso ao criar SQSService com endpoint padrão, mas ocorreu erro: %v", err)
	}

	if service == nil {
		t.Errorf("esperado service não nulo")
	}
}

func TestVideoMessageStructValidation(t *testing.T) {
	// Testar estrutura VideoMessage
	videoMsg := VideoMessage{
		IDVideo:     "video-123",
		Titulo:      testVideoTitle,
		Autor:       testAuthor,
		Status:      "PENDING",
		FilePath:    testFilePath,
		DataCriacao: "2023-12-01T10:00:00Z",
		DataUpload:  "2023-12-01T10:05:00Z",
		Email:       testEmail,
		Username:    testUsername,
		ID:          1,
		VideoKey:    testVideoKey,
		VideoID:     testReceiptHandle,
	}

	// Verificar se todos os campos foram configurados corretamente
	if videoMsg.IDVideo != "video-123" {
		t.Errorf("IDVideo esperado 'video-123', obtido '%s'", videoMsg.IDVideo)
	}

	if videoMsg.Titulo != testVideoTitle {
		t.Errorf("Titulo esperado '%s', obtido '%s'", testVideoTitle, videoMsg.Titulo)
	}

	if videoMsg.Email != testEmail {
		t.Errorf("Email esperado '%s', obtido '%s'", testEmail, videoMsg.Email)
	}

	if videoMsg.Username != testUsername {
		t.Errorf("Username esperado '%s', obtido '%s'", testUsername, videoMsg.Username)
	}

	if videoMsg.VideoKey != testVideoKey {
		t.Errorf("VideoKey esperado '%s', obtido '%s'", testVideoKey, videoMsg.VideoKey)
	}

	if videoMsg.ID != 1 {
		t.Errorf("ID esperado 1, obtido %d", videoMsg.ID)
	}
}

func TestVideoMessageJSONSerialization(t *testing.T) {
	// Testar serialização JSON da estrutura VideoMessage
	videoMsg := VideoMessage{
		IDVideo:     "video-456",
		Titulo:      testVideoTitle,
		Autor:       testAuthor,
		Status:      "PROCESSING",
		FilePath:    testFilePath,
		DataCriacao: "2023-12-01T10:00:00Z",
		DataUpload:  "2023-12-01T10:05:00Z",
		Email:       testEmail,
		Username:    testUsername,
		ID:          2,
		VideoKey:    testVideoKey,
		VideoID:     testReceiptHandle,
	}

	// Serializar para JSON
	jsonData, err := json.Marshal(videoMsg)
	if err != nil {
		t.Fatalf("erro ao serializar VideoMessage para JSON: %v", err)
	}

	// Verificar se contém campos esperados
	jsonStr := string(jsonData)
	expectedFields := []string{
		`"id_video":"video-456"`,
		`"titulo":"` + testVideoTitle + `"`,
		`"email":"` + testEmail + `"`,
		`"username":"` + testUsername + `"`,
		`"id":2`,
		`"status":"PROCESSING"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON não contém campo esperado: %s\nJSON: %s", field, jsonStr)
		}
	}

	// Deserializar de volta
	var deserializedMsg VideoMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	if err != nil {
		t.Fatalf("erro ao deserializar JSON: %v", err)
	}

	// Verificar se os dados são consistentes
	if deserializedMsg.IDVideo != videoMsg.IDVideo {
		t.Errorf("IDVideo após deserialização não confere")
	}

	if deserializedMsg.Email != videoMsg.Email {
		t.Errorf("Email após deserialização não confere")
	}

	if deserializedMsg.ID != videoMsg.ID {
		t.Errorf("ID após deserialização não confere")
	}
}

func TestVideoKeyExtractionLogic(t *testing.T) {
	// Testar a lógica de extração de VideoKey do FilePath
	testCases := []struct {
		name         string
		filePath     string
		expectedKey  string
	}{
		{
			name:        "Caminho S3 padrão",
			filePath:    "s3://video-service-bucket/videos/uuid-123/video.mp4",
			expectedKey: "videos/uuid-123/video.mp4",
		},
		{
			name:        "Caminho S3 com subdiretórios",
			filePath:    "s3://video-service-bucket/videos/2023/12/uuid-456/video.avi",
			expectedKey: "videos/2023/12/uuid-456/video.avi",
		},
		{
			name:        "Caminho sem prefixo S3",
			filePath:    "videos/direct/path/video.mkv",
			expectedKey: "videos/direct/path/video.mkv",
		},
		{
			name:        "Caminho vazio",
			filePath:    "",
			expectedKey: "",
		},
		{
			name:        "Bucket diferente",
			filePath:    "s3://other-bucket/videos/test/video.mp4",
			expectedKey: "s3://other-bucket/videos/test/video.mp4", // fallback
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simular lógica de ReceiveMessages
			videoMsg := VideoMessage{
				FilePath: tc.filePath,
			}

			// Aplicar lógica de extração
			if videoMsg.FilePath != "" {
				bucketPrefix := "s3://video-service-bucket/"
				if strings.HasPrefix(videoMsg.FilePath, bucketPrefix) {
					videoMsg.VideoKey = videoMsg.FilePath[len(bucketPrefix):]
				} else {
					videoMsg.VideoKey = videoMsg.FilePath // fallback
				}
			}

			if videoMsg.VideoKey != tc.expectedKey {
				t.Errorf("VideoKey esperado '%s', obtido '%s'", tc.expectedKey, videoMsg.VideoKey)
			}
		})
	}
}

func TestReceiveMessagesConnectionError(t *testing.T) {
	// Testar erro de conexão no ReceiveMessages
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	// Sem SQS real, esperamos erro de conexão
	messages, err := service.ReceiveMessages()
	if err == nil {
		t.Errorf("esperado erro de conexão SQS, mas não ocorreu")
	}

	if messages != nil {
		t.Errorf("esperado messages nil em caso de erro, mas obteve: %v", messages)
	}
}

func TestDeleteMessageConnectionError(t *testing.T) {
	// Testar erro de conexão no DeleteMessage
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	// Sem SQS real, esperamos erro de conexão
	err = service.DeleteMessage(testReceiptHandle)
	if err == nil {
		t.Errorf("esperado erro de conexão SQS para DeleteMessage, mas não ocorreu")
	}
}

func TestSendMessageValidation(t *testing.T) {
	// Testar validação no SendMessage
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	// Testar SendMessage com videoKey válido
	err = service.SendMessage(testVideoKey)
	if err == nil {
		t.Errorf("esperado erro de conexão SQS para SendMessage, mas não ocorreu")
	}
}

func TestSendMessageJSONCreation(t *testing.T) {
	// Testar criação do JSON no SendMessage
	videoKey := testVideoKey

	// Simular o que SendMessage faz internamente
	message := VideoMessage{
		VideoKey: videoKey,
		VideoID:  videoKey,
	}

	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("erro ao serializar mensagem: %v", err)
	}

	// Verificar se o JSON contém os campos esperados
	jsonStr := string(body)
	if !strings.Contains(jsonStr, videoKey) {
		t.Errorf("JSON não contém videoKey esperado: %s\nJSON: %s", videoKey, jsonStr)
	}

	// Verificar se pode ser deserializado
	var deserializedMsg VideoMessage
	err = json.Unmarshal(body, &deserializedMsg)
	if err != nil {
		t.Fatalf("erro ao deserializar mensagem: %v", err)
	}

	if deserializedMsg.VideoKey != videoKey {
		t.Errorf("VideoKey após deserialização não confere")
	}

	if deserializedMsg.VideoID != videoKey {
		t.Errorf("VideoID após deserialização não confere")
	}
}

func TestGetClient(t *testing.T) {
	// Testar GetClient
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	client := service.GetClient()
	if client == nil {
		t.Errorf("GetClient retornou nil")
	}

	// Verificar se é o mesmo client
	if client != service.sqsClient {
		t.Errorf("GetClient não retornou o mesmo cliente interno")
	}
}

func TestSQSServiceStructureValidation(t *testing.T) {
	// Testar estrutura do SQSService
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	// Verificar se todos os campos foram inicializados
	if service.sqsClient == nil {
		t.Errorf("sqsClient não foi inicializado")
	}

	if service.queueURL != testQueueURL {
		t.Errorf("queueURL não foi configurado corretamente")
	}
}

func TestVideoMessageWithSpecialCharacters(t *testing.T) {
	// Testar VideoMessage com caracteres especiais
	videoMsg := VideoMessage{
		IDVideo:     "video-special-chars",
		Titulo:      "Título com acentos áéíóú e símbolos #@$%",
		Autor:       "Autor com \"aspas\" e \n quebras de linha",
		Status:      "PENDING",
		FilePath:    testFilePath,
		Email:       "test+special@example.com",
		Username:    "user_with-special.chars",
		ID:          999,
	}

	// Deve conseguir serializar mesmo com caracteres especiais
	jsonData, err := json.Marshal(videoMsg)
	if err != nil {
		t.Fatalf("erro inesperado ao serializar VideoMessage com caracteres especiais: %v", err)
	}

	// Verificar se pode ser deserializado de volta
	var deserializedMsg VideoMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	if err != nil {
		t.Fatalf("erro ao deserializar VideoMessage com caracteres especiais: %v", err)
	}

	// Verificar preservação dos dados
	if deserializedMsg.Titulo != videoMsg.Titulo {
		t.Errorf("Título não foi preservado na serialização")
	}

	if deserializedMsg.Autor != videoMsg.Autor {
		t.Errorf("Autor não foi preservado na serialização")
	}
}

func TestVideoMessageEmptyFields(t *testing.T) {
	// Testar VideoMessage com campos vazios
	videoMsg := VideoMessage{
		IDVideo:  "", // vazio
		Titulo:   "", // vazio
		Autor:    "", // vazio
		Status:   "PENDING",
		FilePath: "", // vazio
		Email:    "", // vazio
		Username: "", // vazio
		ID:       0,
	}

	// Deve conseguir serializar mesmo com campos vazios
	jsonData, err := json.Marshal(videoMsg)
	if err != nil {
		t.Fatalf("erro ao serializar VideoMessage com campos vazios: %v", err)
	}

	// Verificar se pode ser deserializado
	var deserializedMsg VideoMessage
	err = json.Unmarshal(jsonData, &deserializedMsg)
	if err != nil {
		t.Fatalf("erro ao deserializar VideoMessage com campos vazios: %v", err)
	}

	// Verificar se campos vazios foram preservados
	if deserializedMsg.IDVideo != "" {
		t.Errorf("IDVideo vazio não foi preservado")
	}

	if deserializedMsg.ID != 0 {
		t.Errorf("ID zero não foi preservado")
	}
}

func TestSendMessageEmptyVideoKey(t *testing.T) {
	// Testar SendMessage com videoKey vazio
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewSQSService(testQueueURL)
	if err != nil {
		t.Fatalf("erro ao criar SQSService: %v", err)
	}

	// Testar com videoKey vazio - deve funcionar mas gerar JSON com campos vazios
	err = service.SendMessage("")
	if err == nil {
		t.Errorf("esperado erro de conexão SQS mesmo com videoKey vazio")
	}
}

// Benchmark para serialização JSON
func BenchmarkVideoMessageSerialization(b *testing.B) {
	videoMsg := VideoMessage{
		IDVideo:     "video-benchmark",
		Titulo:      testVideoTitle,
		Autor:       testAuthor,
		Status:      "PROCESSING",
		FilePath:    testFilePath,
		DataCriacao: "2023-12-01T10:00:00Z",
		DataUpload:  "2023-12-01T10:05:00Z",
		Email:       testEmail,
		Username:    testUsername,
		ID:          100,
		VideoKey:    testVideoKey,
		VideoID:     testReceiptHandle,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(videoMsg)
		if err != nil {
			b.Fatalf("Erro na serialização: %v", err)
		}
	}
}

// Benchmark para extração de VideoKey
func BenchmarkVideoKeyExtraction(b *testing.B) {
	filePath := testFilePath
	bucketPrefix := "s3://video-service-bucket/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var videoKey string
		if strings.HasPrefix(filePath, bucketPrefix) {
			videoKey = filePath[len(bucketPrefix):]
		} else {
			videoKey = filePath
		}
		_ = videoKey // evitar otimização do compilador
	}
}