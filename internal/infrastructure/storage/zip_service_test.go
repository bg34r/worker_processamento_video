package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCreateZipFile_Success(t *testing.T) {
	files := []string{}
	os.MkdirAll("temp", 0755)
	for i := 0; i < 2; i++ {
		name := filepath.Join("temp", "file_test_zip_"+strconv.Itoa(i)+".txt")
		err := os.WriteFile(name, []byte("conteúdo de teste"), 0644)
		if err != nil {
			t.Fatalf("erro ao criar arquivo temporário: %v", err)
		}
		files = append(files, name)
		defer os.Remove(name)
	}

	zipPath := filepath.Join("temp", "test.zip")
	defer os.Remove(zipPath)

	z := NewZipService()
	err := z.CreateZipFile(files, zipPath)
	if err != nil {
		t.Fatalf("esperado sucesso ao criar zip, mas ocorreu erro: %v", err)
	}

	if _, err := os.Stat(zipPath); err != nil {
		t.Errorf("esperado arquivo zip criado, mas não encontrado")
	}
}

func TestCreateZipFile_FileNotFound(t *testing.T) {
	z := NewZipService()
	zipPath := filepath.Join("temp", "test_fail.zip")
	defer os.Remove(zipPath)

	files := []string{"arquivo_inexistente.txt"}
	err := z.CreateZipFile(files, zipPath)
	if err == nil {
		t.Errorf("esperado erro ao adicionar arquivo inexistente ao zip")
	}
}

func TestCreateZipFile_ErrorOnCreateZip(t *testing.T) {
	z := NewZipService()
	// Caminho inválido para forçar erro
	zipPath := string([]byte{0})
	files := []string{}
	err := z.CreateZipFile(files, zipPath)
	if err == nil {
		t.Errorf("esperado erro ao criar arquivo zip em caminho inválido")
	}
}
