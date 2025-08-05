package providers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

func TestJSONProvider(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test JSON files
	ptBRContent := `{
		"hello": "Olá",
		"welcome": "Bem-vindo {{.Name}}",
		"users_count": {
			"one": "{{.Count}} usuário",
			"other": "{{.Count}} usuários"
		}
	}`

	enContent := `{
		"hello": "Hello",
		"welcome": "Welcome {{.Name}}",
		"users_count": {
			"one": "{{.Count}} user",
			"other": "{{.Count}} users"
		}
	}`

	ptBRFile := filepath.Join(tmpDir, "translations.pt-BR.json")
	enFile := filepath.Join(tmpDir, "translations.en.json")

	if err := os.WriteFile(ptBRFile, []byte(ptBRContent), 0644); err != nil {
		t.Fatalf("Failed to write pt-BR file: %v", err)
	}
	if err := os.WriteFile(enFile, []byte(enContent), 0644); err != nil {
		t.Fatalf("Failed to write en file: %v", err)
	}

	// Create provider
	p := providers.NewJSONProvider()

	// Test loading translations
	t.Run("LoadTranslations", func(t *testing.T) {
		tests := []struct {
			name        string
			file        string
			format      string
			expectError bool
		}{
			{
				name:        "Valid JSON pt-BR",
				file:        ptBRFile,
				format:      "json",
				expectError: false,
			},
			{
				name:        "Valid JSON en",
				file:        enFile,
				format:      "json",
				expectError: false,
			},
			{
				name:        "Invalid format",
				file:        ptBRFile,
				format:      "yaml",
				expectError: true,
			},
			{
				name:        "Non-existent file",
				file:        filepath.Join(tmpDir, "non-existent.json"),
				format:      "json",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := p.LoadTranslations(tt.file, tt.format)
				if tt.expectError && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	// Set languages and test translations
	if err := p.SetLanguages("pt-BR", "en"); err != nil {
		t.Fatalf("Failed to set languages: %v", err)
	}

	// Test simple translations
	t.Run("SimpleTranslations", func(t *testing.T) {
		tests := []struct {
			name     string
			key      string
			expected string
		}{
			{
				name:     "Portuguese hello",
				key:      "hello",
				expected: "Olá",
			},
			{
				name:     "Missing key",
				key:      "missing",
				expected: "missing",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := p.Translate(tt.key, nil)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	// Test template translations
	t.Run("TemplateTranslations", func(t *testing.T) {
		result, err := p.Translate("welcome", map[string]interface{}{
			"Name": "John",
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "Bem-vindo John"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	// Test plural translations
	t.Run("PluralTranslations", func(t *testing.T) {
		tests := []struct {
			name     string
			count    interface{}
			expected string
		}{
			{
				name:     "One user",
				count:    1,
				expected: "1 usuário",
			},
			{
				name:     "Multiple users",
				count:    2,
				expected: "2 usuários",
			},
			{
				name:     "Zero users",
				count:    0,
				expected: "0 usuários",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := p.TranslatePlural("users_count", tt.count, map[string]interface{}{
					"Count": tt.count,
				})
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	// Test fallback
	t.Run("Fallback", func(t *testing.T) {
		// Set only English and try to translate
		if err := p.SetLanguages("en"); err != nil {
			t.Fatalf("Failed to set languages: %v", err)
		}

		result, err := p.Translate("hello", nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "Hello"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}
