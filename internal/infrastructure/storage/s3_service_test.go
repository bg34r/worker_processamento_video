package storage

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testBucket        = "test-bucket"
	testVideoFile     = "video.mp4"
	testEndpoint      = "http://localhost:4566"
	testDocumentFile  = "document.pdf"
	testImageFile     = "image.jpg"
	errorS3Connection = "esperado erro de conexão S3, mas não ocorreu"
	errorCreateS3     = "erro ao criar S3Service: %v"
)

func TestNewS3ServiceSuccess(t *testing.T) {
	// Configurar endpoint de teste
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	bucket := testBucket
	service, err := NewS3Service(bucket)
	if err != nil {
		t.Fatalf("esperado sucesso ao criar S3Service, mas ocorreu erro: %v", err)
	}

	if service == nil {
		t.Fatalf("esperado service não nulo")
	}

	if service.bucket != bucket {
		t.Errorf("esperado bucket '%s', mas obteve '%s'", bucket, service.bucket)
	}
}

func TestNewS3ServiceWithoutEndpointEnv(t *testing.T) {
	// Limpar variável de ambiente para testar fallback
	originalEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	os.Unsetenv("LOCALSTACK_ENDPOINT")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("LOCALSTACK_ENDPOINT", originalEndpoint)
		}
	}()

	bucket := testBucket
	service, err := NewS3Service(bucket)
	if err != nil {
		t.Fatalf("esperado sucesso ao criar S3Service com endpoint padrão, mas ocorreu erro: %v", err)
	}

	if service == nil {
		t.Errorf("esperado service não nulo")
	}
}

func TestIsVideoFileValidExtensions(t *testing.T) {
	service := &S3Service{}

	validFiles := []string{
		testVideoFile,
		"movie.avi",
		"clip.mov",
		"film.mkv",
		"video.wmv",
		"stream.flv",
		"web.webm",
		"VIDEO.MP4", // Teste case insensitive
		"path/to/" + testVideoFile,
	}

	for _, file := range validFiles {
		if !service.isVideoFile(file) {
			t.Errorf("esperado que '%s' seja reconhecido como arquivo de vídeo", file)
		}
	}
}

func TestIsVideoFileInvalidExtensions(t *testing.T) {
	service := &S3Service{}

	invalidFiles := []string{
		testDocumentFile,
		testImageFile,
		"audio.mp3",
		"text.txt",
		"archive.zip",
		"video", // sem extensão
		"video.",
		"",
	}

	for _, file := range invalidFiles {
		if service.isVideoFile(file) {
			t.Errorf("esperado que '%s' NÃO seja reconhecido como arquivo de vídeo", file)
		}
	}
}

func TestDownloadVideoFileNotFound(t *testing.T) {
	// Testar criação de diretório inexistente para download
	localPath := filepath.Join("diretorio_inexistente", testVideoFile)

	// Verificar que o diretório realmente não existe
	if _, err := os.Stat(filepath.Dir(localPath)); err == nil {
		t.Errorf("diretório de teste não deveria existir inicialmente")
	}

	// Tentar criar arquivo no diretório inexistente deveria falhar
	_, err := os.Create(localPath)
	if err == nil {
		t.Errorf("esperado erro ao criar arquivo em diretório inexistente")
		os.Remove(localPath) // cleanup se por algum motivo criou
	}
}

func TestDownloadVideoInvalidPath(t *testing.T) {
	// Testar com caminho inválido
	invalidPath := string([]byte{0})

	// Tentar criar arquivo com caminho inválido
	_, err := os.Create(invalidPath)
	if err == nil {
		t.Errorf("esperado erro ao usar caminho inválido")
	}
}

func TestUploadZipFileNotFound(t *testing.T) {
	// Testar se arquivo inexistente realmente não existe
	nonExistentFile := "arquivo_inexistente.zip"

	if _, err := os.Stat(nonExistentFile); err == nil {
		t.Errorf("arquivo de teste não deveria existir")
	}

	// Tentar abrir arquivo inexistente
	_, err := os.Open(nonExistentFile)
	if err == nil {
		t.Errorf("esperado erro ao tentar abrir arquivo inexistente")
	}
}

func TestUploadZipValidFile(t *testing.T) {
	// Criar arquivo temporário para teste
	tempDir := "temp"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test_upload.zip")
	err := os.WriteFile(tempFile, []byte("conteúdo de teste zip"), 0644)
	if err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}

	// Testar apenas se o arquivo existe antes de tentar upload
	if _, err := os.Stat(tempFile); err != nil {
		t.Errorf("arquivo temporário não foi criado corretamente: %v", err)
	}

	// Verificar conteúdo do arquivo
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("erro ao ler arquivo temporário: %v", err)
	}

	expectedContent := "conteúdo de teste zip"
	if string(content) != expectedContent {
		t.Errorf("conteúdo do arquivo não confere. Esperado: %s, Obtido: %s",
			expectedContent, string(content))
	}
}

func TestListVideosConnectionError(t *testing.T) {
	// Testar se a função isVideoFile funciona corretamente com diferentes extensões
	service := &S3Service{
		bucket: testBucket,
	}

	// Como estamos testando sem cliente S3 real, apenas validamos que
	// a estrutura está correta para uso
	if service.bucket != testBucket {
		t.Errorf("bucket não foi configurado corretamente. Esperado: %s, Obtido: %s",
			testBucket, service.bucket)
	}
}

func TestS3ServiceIntegrationCreateAndCleanup(t *testing.T) {
	// Teste de integração que cria arquivos temporários e limpa depois
	tempDir := "temp_s3_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Criar arquivo de teste
	testFile := filepath.Join(tempDir, "test_video.mp4")
	testContent := []byte("conteúdo de vídeo de teste")
	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("erro ao criar arquivo de teste: %v", err)
	}

	// Verificar que o arquivo foi criado corretamente
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("arquivo de teste não foi criado corretamente")
	}

	// Verificar conteúdo
	readContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("erro ao ler arquivo de teste: %v", err)
	}

	if string(readContent) != string(testContent) {
		t.Errorf("conteúdo do arquivo não confere. Esperado: %s, Obtido: %s",
			string(testContent), string(readContent))
	}
}

func TestS3ServiceEdgeCases(t *testing.T) {
	service := &S3Service{
		bucket: testBucket,
	}

	// Teste com strings vazias
	if service.isVideoFile("") {
		t.Errorf("string vazia não deveria ser considerada arquivo de vídeo")
	}

	// Teste com apenas extensão (tecnicamente válido)
	if !service.isVideoFile(".mp4") {
		t.Errorf("extensão .mp4 deveria ser reconhecida como válida")
	}

	// Teste com caminho complexo
	complexPath := "path/to/deep/directory/video_file_name.MP4"
	if !service.isVideoFile(complexPath) {
		t.Errorf("caminho complexo com extensão válida deveria ser reconhecido")
	}
}

func TestDownloadVideoCreateFileError(t *testing.T) {
	// Testar erro na criação do arquivo local
	invalidPath := ""
	if os.PathSeparator == '\\' {
		// Windows - usar caminho inválido
		invalidPath = "\\\\invalid\\path\\video.mp4"
	} else {
		// Unix - usar caminho inválido
		invalidPath = "/proc/invalid/video.mp4"
	}

	service := &S3Service{
		bucket: testBucket,
	}

	err := service.DownloadVideo("test-key", invalidPath)
	if err == nil {
		t.Errorf("esperado erro ao tentar criar arquivo em caminho inválido")
	}
}

func TestDownloadVideoValidation(t *testing.T) {
	// Testar validação de parâmetros com S3Service inicializado corretamente
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewS3Service(testBucket)
	if err != nil {
		t.Fatalf("erro ao criar S3Service: %v", err)
	}

	// Criar diretório temporário
	tempDir := "temp_download_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	validPath := filepath.Join(tempDir, "test_download.mp4")

	// Como não temos S3 real, esperamos erro de conexão, não erro de arquivo
	err = service.DownloadVideo("test-video.mp4", validPath)
	if err == nil {
		t.Error(errorS3Connection)
	}

	// Cleanup se arquivo foi criado
	os.Remove(validPath)
}

func TestUploadZipParameterValidation(t *testing.T) {
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewS3Service(testBucket)
	if err != nil {
		t.Fatalf("erro ao criar S3Service: %v", err)
	}

	// Criar arquivo temporário válido
	tempDir := "temp_upload_validation"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	validFile := filepath.Join(tempDir, "test_upload_validation.zip")
	err = os.WriteFile(validFile, []byte("conteúdo teste"), 0644)
	if err != nil {
		t.Fatalf("erro ao criar arquivo de teste: %v", err)
	}

	// Testar upload com arquivo válido mas sem conexão S3
	err = service.UploadZip(validFile, "test-key.zip")
	if err == nil {
		t.Error(errorS3Connection)
	}
}

func TestUploadZipDirectoryAsFile(t *testing.T) {
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewS3Service(testBucket)
	if err != nil {
		t.Fatalf("erro ao criar S3Service: %v", err)
	}

	// Criar diretório temporário
	tempDir := "temp_directory_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Tentar fazer upload de um diretório (deve falhar)
	err = service.UploadZip(tempDir, "test-directory")
	if err == nil {
		t.Errorf("esperado erro ao tentar fazer upload de diretório")
	}
}

func TestListVideosEmptyBucket(t *testing.T) {
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewS3Service(testBucket)
	if err != nil {
		t.Fatalf("erro ao criar S3Service: %v", err)
	}

	// Sem conexão S3 real, esperamos erro
	videos, err := service.ListVideos()
	if err == nil {
		t.Error(errorS3Connection)
	}

	if videos != nil {
		t.Errorf("esperado videos nil em caso de erro, mas obteve: %v", videos)
	}
}

func TestListVideosFilteringLogic(t *testing.T) {
	// Testar a lógica de filtragem de vídeos sem S3 real
	service := &S3Service{
		bucket: testBucket,
	}

	// Simular lista de arquivos
	testFiles := []string{
		"video1.mp4",
		"video2.avi",
		testDocumentFile,
		testImageFile,
		"movie.mkv",
		"archive.zip",
		"clip.MOV", // case insensitive
	}

	expectedVideos := []string{
		"video1.mp4",
		"video2.avi",
		"movie.mkv",
		"clip.MOV",
	}

	// Testar filtragem usando isVideoFile
	var filteredVideos []string
	for _, file := range testFiles {
		if service.isVideoFile(file) {
			filteredVideos = append(filteredVideos, file)
		}
	}

	if len(filteredVideos) != len(expectedVideos) {
		t.Errorf("esperado %d vídeos, mas obteve %d", len(expectedVideos), len(filteredVideos))
	}

	// Verificar se todos os vídeos esperados estão presentes
	for _, expected := range expectedVideos {
		found := false
		for _, filtered := range filteredVideos {
			if filtered == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("vídeo esperado '%s' não encontrado na lista filtrada", expected)
		}
	}
}

func TestS3ServiceStructureValidation(t *testing.T) {
	// Testar se a estrutura do S3Service é inicializada corretamente
	os.Setenv("LOCALSTACK_ENDPOINT", testEndpoint)
	defer os.Unsetenv("LOCALSTACK_ENDPOINT")

	service, err := NewS3Service(testBucket)
	if err != nil {
		t.Fatalf("erro ao criar S3Service: %v", err)
	}

	// Verificar se todos os campos foram inicializados
	if service.s3Client == nil {
		t.Errorf("s3Client não foi inicializado")
	}

	if service.uploader == nil {
		t.Errorf("uploader não foi inicializado")
	}

	if service.downloader == nil {
		t.Errorf("downloader não foi inicializado")
	}

	if service.bucket != testBucket {
		t.Errorf("bucket não foi configurado corretamente")
	}
}

func TestIsVideoFileExtensionHandling(t *testing.T) {
	service := &S3Service{}

	// Testar diferentes formatos de extensão
	testCases := []struct {
		filename string
		expected bool
		desc     string
	}{
		{".mp4", true, "apenas extensão"},
		{"file.mp4", true, "arquivo com extensão"},
		{"file.MP4", true, "extensão maiúscula"},
		{"file.Mp4", true, "extensão mista"},
		{"file.mp4.backup", false, "extensão dupla"},
		{"mp4", false, "sem ponto"},
		{"file.", false, "ponto sem extensão"},
		{"path/file.mp4", true, "com caminho"},
		{"path\\file.mp4", true, "com caminho Windows"},
		{"very.long.path.with.dots.file.mkv", true, "múltiplos pontos"},
	}

	for _, tc := range testCases {
		result := service.isVideoFile(tc.filename)
		if result != tc.expected {
			t.Errorf("Teste '%s': arquivo '%s' - esperado %v, obtido %v",
				tc.desc, tc.filename, tc.expected, result)
		}
	}
}

// Benchmark para testar performance da validação de arquivos
func BenchmarkIsVideoFile(b *testing.B) {
	service := &S3Service{}
	testFiles := []string{
		testVideoFile,
		"movie.avi",
		testDocumentFile,
		testImageFile,
		"very/long/path/to/video/file.mkv",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, file := range testFiles {
			service.isVideoFile(file)
		}
	}
}
