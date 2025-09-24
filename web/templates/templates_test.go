package templates

import (
	"strings"
	"testing"
)

func TestIndexHTML_ContainsUploadForm(t *testing.T) {
	if !strings.Contains(IndexHTML, `<form id="uploadForm"`) {
		t.Errorf("O template não contém o formulário de upload")
	}
	if !strings.Contains(IndexHTML, `id="filesList"`) {
		t.Errorf("O template não contém a lista de arquivos processados")
	}
	if !strings.Contains(IndexHTML, `<script>`) {
		t.Errorf("O template não contém o script JS")
	}
}
