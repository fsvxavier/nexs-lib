package config_test

import (
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"golang.org/x/text/language"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	// Test default values
	ptBR := language.MustParse("pt-BR")
	if cfg.DefaultLanguage != ptBR {
		t.Errorf("Expected default language to be pt-BR, got %v", cfg.DefaultLanguage)
	}

	if len(cfg.FallbackLanguages) != 2 {
		t.Errorf("Expected 2 fallback languages, got %d", len(cfg.FallbackLanguages))
	}

	if cfg.FallbackLanguages[0] != ptBR {
		t.Errorf("Expected first fallback language to be pt-BR, got %v", cfg.FallbackLanguages[0])
	}

	if cfg.FallbackLanguages[1] != language.English {
		t.Errorf("Expected second fallback language to be English, got %v", cfg.FallbackLanguages[1])
	}

	if cfg.TranslationsPath != "translations" {
		t.Errorf("Expected translations path to be 'translations', got %s", cfg.TranslationsPath)
	}

	if cfg.TranslationsFormat != "json" {
		t.Errorf("Expected translations format to be 'json', got %s", cfg.TranslationsFormat)
	}
}

func TestConfigBuilder(t *testing.T) {
	cfg := config.DefaultConfig().
		WithDefaultLanguage(language.English).
		WithFallbackLanguages(language.English, language.French).
		WithTranslationsPath("/custom/path").
		WithTranslationsFormat("yaml")

	// Test custom values
	if cfg.DefaultLanguage != language.English {
		t.Errorf("Expected default language to be English, got %v", cfg.DefaultLanguage)
	}

	if len(cfg.FallbackLanguages) != 2 {
		t.Errorf("Expected 2 fallback languages, got %d", len(cfg.FallbackLanguages))
	}

	if cfg.FallbackLanguages[0] != language.English {
		t.Errorf("Expected first fallback language to be English, got %v", cfg.FallbackLanguages[0])
	}

	if cfg.FallbackLanguages[1] != language.French {
		t.Errorf("Expected second fallback language to be French, got %v", cfg.FallbackLanguages[1])
	}

	if cfg.TranslationsPath != "/custom/path" {
		t.Errorf("Expected translations path to be '/custom/path', got %s", cfg.TranslationsPath)
	}

	if cfg.TranslationsFormat != "yaml" {
		t.Errorf("Expected translations format to be 'yaml', got %s", cfg.TranslationsFormat)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name:        "Valid config",
			config:      config.DefaultConfig(),
			expectError: false,
		},
		{
			name: "Invalid default language",
			config: &config.Config{
				DefaultLanguage:    language.Und,
				FallbackLanguages:  []language.Tag{language.English},
				TranslationsPath:   "translations",
				TranslationsFormat: "json",
			},
			expectError: true,
		},
		{
			name: "No fallback languages",
			config: &config.Config{
				DefaultLanguage:    language.English,
				FallbackLanguages:  []language.Tag{},
				TranslationsPath:   "translations",
				TranslationsFormat: "json",
			},
			expectError: true,
		},
		{
			name: "Empty translations path",
			config: &config.Config{
				DefaultLanguage:    language.English,
				FallbackLanguages:  []language.Tag{language.English},
				TranslationsPath:   "",
				TranslationsFormat: "json",
			},
			expectError: true,
		},
		{
			name: "Invalid format",
			config: &config.Config{
				DefaultLanguage:    language.English,
				FallbackLanguages:  []language.Tag{language.English},
				TranslationsPath:   "translations",
				TranslationsFormat: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetTranslationFilePath(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		lang         string
		expectedPath string
	}{
		{
			name: "JSON format",
			config: &config.Config{
				TranslationsPath:   "/path/to/translations",
				TranslationsFormat: "json",
			},
			lang:         "pt-BR",
			expectedPath: filepath.Join("/path/to/translations", "translations.pt-BR.json"),
		},
		{
			name: "YAML format",
			config: &config.Config{
				TranslationsPath:   "/path/to/translations",
				TranslationsFormat: "yaml",
			},
			lang:         "en",
			expectedPath: filepath.Join("/path/to/translations", "translations.en.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.config.GetTranslationFilePath(tt.lang)
			if path != tt.expectedPath {
				t.Errorf("Expected path %q, got %q", tt.expectedPath, path)
			}
		})
	}
}
