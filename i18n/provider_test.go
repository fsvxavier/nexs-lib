package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"
)

func createTestTranslations(t *testing.T) (string, func()) {
	tempDir := t.TempDir()

	// Create test translation files
	translations := map[string]map[string]string{
		"en": {
			"hello":         "Hello",
			"welcome":       "Welcome, {{name}}!",
			"users.1":       "One user",
			"users.0":       "No users",
			"users.other":   "{{count}} users",
			"nested.key":    "Nested translation",
			"with.plural.1": "One item",
			"with.plural.2": "Two items",
		},
		"pt-BR": {
			"hello":         "Olá",
			"welcome":       "Bem-vindo, {{name}}!",
			"users.1":       "Um usuário",
			"users.0":       "Nenhum usuário",
			"users.other":   "{{count}} usuários",
			"nested.key":    "Tradução aninhada",
			"with.plural.1": "Um item",
			"with.plural.2": "Dois itens",
		},
	}

	// Write translation files
	for lang, trans := range translations {
		content, err := json.MarshalIndent(trans, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(tempDir, lang+".json"), content, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestBasicProvider_LoadTranslations(t *testing.T) {
	tempDir, cleanup := createTestTranslations(t)
	defer cleanup()

	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "Valid JSON format",
			format:      "json",
			expectError: false,
		},
		{
			name:        "Unsupported format",
			format:      "yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewBasicProvider()
			err := p.LoadTranslations(tempDir, tt.format)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBasicProvider_Translate(t *testing.T) {
	tempDir, cleanup := createTestTranslations(t)
	defer cleanup()

	tests := []struct {
		name        string
		languages   []string
		key         string
		data        map[string]interface{}
		want        string
		expectError bool
	}{
		{
			name:        "Simple translation en",
			languages:   []string{"en"},
			key:         "hello",
			data:        nil,
			want:        "Hello",
			expectError: false,
		},
		{
			name:        "Simple translation pt-BR",
			languages:   []string{"pt-BR"},
			key:         "hello",
			data:        nil,
			want:        "Olá",
			expectError: false,
		},
		{
			name:        "Missing key",
			languages:   []string{"en"},
			key:         "missing.key",
			data:        nil,
			want:        "",
			expectError: true,
		},
		{
			name:        "No languages configured",
			languages:   []string{},
			key:         "hello",
			data:        nil,
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewBasicProvider()
			err := p.LoadTranslations(tempDir, "json")
			if err != nil {
				t.Fatal(err)
			}

			if len(tt.languages) > 0 {
				err = p.SetLanguages(tt.languages...)
				if err != nil {
					t.Fatal(err)
				}
			}

			got, err := p.Translate(tt.key, tt.data)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBasicProvider_TranslatePlural(t *testing.T) {
	tempDir, cleanup := createTestTranslations(t)
	defer cleanup()

	tests := []struct {
		name        string
		languages   []string
		key         string
		count       interface{}
		data        map[string]interface{}
		want        string
		expectError bool
	}{
		{
			name:        "Single user",
			languages:   []string{"en"},
			key:         "users",
			count:       1,
			data:        nil,
			want:        "One user",
			expectError: false,
		},
		{
			name:        "No users",
			languages:   []string{"en"},
			key:         "users",
			count:       0,
			data:        nil,
			want:        "No users",
			expectError: false,
		},
		{
			name:        "Multiple users",
			languages:   []string{"en"},
			key:         "users",
			count:       5,
			data:        nil,
			want:        "5 users",
			expectError: false,
		},
		{
			name:        "Missing plural key",
			languages:   []string{"en"},
			key:         "missing",
			count:       1,
			data:        nil,
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewBasicProvider()
			err := p.LoadTranslations(tempDir, "json")
			if err != nil {
				t.Fatal(err)
			}

			err = p.SetLanguages(tt.languages...)
			if err != nil {
				t.Fatal(err)
			}

			got, err := p.TranslatePlural(tt.key, tt.count, tt.data)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("TranslatePlural() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBasicProvider_SetLanguages(t *testing.T) {
	tests := []struct {
		name        string
		languages   []string
		want        []language.Tag
		expectError bool
	}{
		{
			name:        "Valid languages",
			languages:   []string{"en", "pt-BR", "es"},
			expectError: false,
		},
		{
			name:        "Invalid language code",
			languages:   []string{"invalid"},
			expectError: true,
		},
		{
			name:        "Empty languages",
			languages:   []string{},
			expectError: false,
		},
		{
			name:        "Mixed valid and invalid",
			languages:   []string{"en", "invalid", "pt-BR"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewBasicProvider()
			err := p.SetLanguages(tt.languages...)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				got := p.GetLanguages()
				if len(got) != len(tt.languages) {
					t.Errorf("SetLanguages() resulted in %d languages, want %d", len(got), len(tt.languages))
				}
			}
		})
	}
}
