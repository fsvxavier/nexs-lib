package i18n

import (
	"fmt"
	"testing"
)

func TestNewProvider(t *testing.T) {
	tempDir, cleanup := createTestTranslations(t)
	defer cleanup()

	tests := []struct {
		name        string
		config      ProviderConfig
		expectError bool
	}{
		{
			name: "Basic provider",
			config: ProviderConfig{
				Type:               ProviderTypeBasic,
				TranslationsPath:   tempDir,
				TranslationsFormat: "json",
				Languages:          []string{"en", "pt-BR"},
			},
			expectError: false,
		},
		{
			name: "Cached provider",
			config: ProviderConfig{
				Type:               ProviderTypeCached,
				TranslationsPath:   tempDir,
				TranslationsFormat: "json",
				Languages:          []string{"en", "pt-BR"},
				CacheTTL:           60,
				CacheMaxSize:       1000,
			},
			expectError: false,
		},
		{
			name: "Invalid provider type",
			config: ProviderConfig{
				Type:               "invalid",
				TranslationsPath:   tempDir,
				TranslationsFormat: "json",
				Languages:          []string{"en", "pt-BR"},
			},
			expectError: true,
		},
		{
			name: "Invalid language",
			config: ProviderConfig{
				Type:               ProviderTypeBasic,
				TranslationsPath:   tempDir,
				TranslationsFormat: "json",
				Languages:          []string{"invalid"},
			},
			expectError: true,
		},
		{
			name: "Invalid translations path",
			config: ProviderConfig{
				Type:               ProviderTypeBasic,
				TranslationsPath:   "/invalid/path",
				TranslationsFormat: "json",
				Languages:          []string{"en"},
			},
			expectError: true,
		},
		{
			name: "Invalid format",
			config: ProviderConfig{
				Type:               ProviderTypeBasic,
				TranslationsPath:   tempDir,
				TranslationsFormat: "invalid",
				Languages:          []string{"en"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && provider == nil {
				t.Error("expected provider, got nil")
			}

			// Testa se o provider está funcionando
			if !tt.expectError && provider != nil {
				// Verifica se consegue traduzir
				translation, err := provider.Translate("hello", nil)
				if err != nil {
					t.Errorf("failed to translate: %v", err)
				}
				if translation != "Hello" {
					t.Errorf("got %q, want %q", translation, "Hello")
				}

				// Verifica os idiomas configurados
				languages := provider.GetLanguages()
				if len(languages) != len(tt.config.Languages) {
					t.Errorf("got %d languages, want %d", len(languages), len(tt.config.Languages))
				}
			}
		})
	}
}

func TestProviderConcurrency(t *testing.T) {
	tempDir, cleanup := createTestTranslations(t)
	defer cleanup()

	config := ProviderConfig{
		Type:               ProviderTypeCached,
		TranslationsPath:   tempDir,
		TranslationsFormat: "json",
		Languages:          []string{"en", "pt-BR"},
		CacheTTL:           60,
		CacheMaxSize:       1000,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatal(err)
	}

	const numGoroutines = 10
	const numRequests = 100
	errors := make(chan error, numGoroutines*numRequests)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRequests; j++ {
				// Tenta traduzir uma chave
				_, err := provider.Translate("hello", nil)
				if err != nil {
					errors <- fmt.Errorf("failed to translate: %v", err)
				}

				// Tenta traduzir uma chave plural
				_, err = provider.TranslatePlural("users", j%3, nil)
				if err != nil {
					errors <- fmt.Errorf("failed to translate plural: %v", err)
				}
			}
			done <- true
		}()
	}

	// Espera todas as goroutines terminarem
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verifica se houve algum erro
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
