package yaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFactory_Name(t *testing.T) {
	factory := &Factory{}
	assert.Equal(t, "yaml", factory.Name())
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := &Factory{}

	tests := []struct {
		name      string
		config    interface{}
		wantError bool
	}{
		{
			name: "valid config",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en", "pt").
					WithDefaultLanguage("en").
					WithProviderConfig(&config.YAMLProviderConfig{
						FilePath:    "/tmp",
						FilePattern: "{lang}.yaml",
						Encoding:    "utf-8",
					}).
					Build()
				require.NoError(t, err)
				return cfg
			}(),
			wantError: false,
		},
		{
			name:      "invalid config type",
			config:    "invalid",
			wantError: true,
		},
		{
			name: "invalid base config",
			config: &config.Config{
				SupportedLanguages: []string{},
				DefaultLanguage:    "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFactory_Create(t *testing.T) {
	factory := &Factory{}

	tests := []struct {
		name      string
		config    interface{}
		wantError bool
	}{
		{
			name: "valid config",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en", "pt").
					WithDefaultLanguage("en").
					WithProviderConfig(&config.YAMLProviderConfig{
						FilePath:    "/tmp",
						FilePattern: "{lang}.yaml",
						Encoding:    "utf-8",
					}).
					Build()
				require.NoError(t, err)
				return cfg
			}(),
			wantError: false,
		},
		{
			name: "valid config with provider config",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en", "pt").
					WithDefaultLanguage("en").
					WithProviderConfig(config.DefaultYAMLProviderConfig()).
					Build()
				require.NoError(t, err)
				return cfg
			}(),
			wantError: false,
		},
		{
			name:      "invalid config type",
			config:    "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.Create(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Implements(t, (*interfaces.I18n)(nil), provider)
			}
		})
	}
}

func TestProvider_GetSupportedLanguages(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    "/tmp",
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	languages := provider.GetSupportedLanguages()
	assert.Equal(t, []string{"en", "pt", "es"}, languages)
}

func TestProvider_GetDefaultLanguage(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("pt").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    "/tmp",
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	defaultLang := provider.GetDefaultLanguage()
	assert.Equal(t, "pt", defaultLang)
}

func TestProvider_SetDefaultLanguage(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    "/tmp",
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	// Initially should be "en"
	assert.Equal(t, "en", provider.GetDefaultLanguage())

	// Change to "pt"
	provider.SetDefaultLanguage("pt")
	assert.Equal(t, "pt", provider.GetDefaultLanguage())
}

func TestProvider_Translate(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_yaml_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation files
	enTranslations := map[string]interface{}{
		"hello":   "Hello",
		"welcome": "Welcome, {{name}}!",
		"nested": map[string]interface{}{
			"message": "This is nested",
		},
	}
	ptTranslations := map[string]interface{}{
		"hello":   "Olá",
		"welcome": "Bem-vindo, {{name}}!",
	}

	// Write English translations
	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	// Write Portuguese translations
	ptFile := filepath.Join(tempDir, "pt.yaml")
	ptData, _ := yaml.Marshal(ptTranslations)
	require.NoError(t, os.WriteFile(ptFile, ptData, 0644))

	// Create provider configuration
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
			NestedKeys:  true, // Enable nested key support
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	// Start the provider
	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		language  string
		params    map[string]interface{}
		expected  string
		wantError bool
	}{
		{
			name:     "simple translation",
			key:      "hello",
			language: "en",
			expected: "Hello",
		},
		{
			name:     "translation with parameters",
			key:      "welcome",
			language: "en",
			params:   map[string]interface{}{"name": "John"},
			expected: "Welcome, John!",
		},
		{
			name:     "translation in Portuguese",
			key:      "hello",
			language: "pt",
			expected: "Olá",
		},
		{
			name:     "fallback to default language",
			key:      "nested.message",
			language: "pt",
			expected: "This is nested",
		},
		{
			name:      "missing key without fallback",
			key:       "nonexistent",
			language:  "en",
			wantError: false, // In non-strict mode, returns the key itself
			expected:  "nonexistent",
		},
		{
			name:      "empty key",
			key:       "",
			language:  "en",
			wantError: true,
		},
		{
			name:      "empty language",
			key:       "hello",
			language:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Translate(ctx, tt.key, tt.language, tt.params)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	// Test translation before start
	newProvider, err := factory.Create(cfg)
	require.NoError(t, err)

	_, err = newProvider.Translate(ctx, "hello", "en", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not started")
}

func TestProvider_HasTranslation(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_yaml_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
		"nested": map[string]interface{}{
			"message": "This is nested",
		},
	}
	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
			NestedKeys:  true,
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(t, err)

	tests := []struct {
		key      string
		language string
		expected bool
	}{
		{"hello", "en", true},
		{"nested.message", "en", true},
		{"nonexistent", "en", false},
		{"hello", "pt", false}, // pt translations not loaded
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.key, tt.language), func(t *testing.T) {
			result := provider.HasTranslation(tt.key, tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_Health(t *testing.T) {
	// Create temporary directory with a valid translation file
	tempDir, err := os.MkdirTemp("", "i18n_yaml_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Health check before start should fail
	err = provider.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not started")

	// Start provider successfully
	err = provider.Start(ctx)
	require.NoError(t, err)

	// Health check after successful start should pass
	err = provider.Health(ctx)
	assert.NoError(t, err)
}

func TestProvider_StartStop(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_yaml_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithLoadTimeout(5 * time.Second).
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test start
	err = provider.Start(ctx)
	assert.NoError(t, err)

	// Test double start (should not error)
	err = provider.Start(ctx)
	assert.NoError(t, err)

	// Test stop
	err = provider.Stop(ctx)
	assert.NoError(t, err)

	// Test translation after stop (should error)
	_, err = provider.Translate(ctx, "hello", "en", nil)
	assert.Error(t, err)
}

func TestProvider_StartTimeout(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithLoadTimeout(10 * time.Millisecond). // Very short timeout
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    "/nonexistent/path",
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load translations")
}

// Simple benchmark tests
func BenchmarkProvider_Translate(b *testing.B) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_yaml_bench")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	translations := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		translations[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("Value %d", i)
	}
	translations["template"] = "Hello {{name}}!"

	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(translations)
	require.NoError(b, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(b, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(b, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(b, err)

	b.ResetTimer()

	b.Run("simple_translation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.Translate(ctx, "key_0", "en", nil)
		}
	})

	b.Run("template_translation", func(b *testing.B) {
		params := map[string]interface{}{"name": "John"}
		for i := 0; i < b.N; i++ {
			_, _ = provider.Translate(ctx, "template", "en", params)
		}
	})
}

// ===== TESTES AVANÇADOS PARA EXPANSÃO DE COBERTURA =====

// TestProvider_EdgeCases - Testes de casos extremos
func TestProvider_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("empty_yaml_file", func(t *testing.T) {
		// Criar arquivo YAML vazio
		emptyFile := filepath.Join(tempDir, "empty.yaml")
		err := os.WriteFile(emptyFile, []byte(""), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "empty.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Deve retornar chave não encontrada
		result, err := provider.Translate(ctx, "any.key", "en", nil)
		assert.NoError(t, err)
		assert.Equal(t, "any.key", result)
	})

	t.Run("malformed_yaml", func(t *testing.T) {
		// YAML malformado com indentação incorreta
		malformedContent := `
key1: value1
  key2: value2  # indentação incorreta
key3:
  - item1
  - item2
    - nested_item  # indentação incorreta
`
		malformedFile := filepath.Join(tempDir, "malformed.yaml")
		err := os.WriteFile(malformedFile, []byte(malformedContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "malformed.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Deve falhar ao carregar YAML malformado
		err = provider.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("corrupted_yaml_file", func(t *testing.T) {
		// Arquivo com bytes inválidos
		corruptedContent := []byte{0xFF, 0xFE, 0x00, 0x00} // BOM + null bytes
		corruptedFile := filepath.Join(tempDir, "corrupted.yaml")
		err := os.WriteFile(corruptedFile, corruptedContent, 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "corrupted.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Deve falhar ao ler arquivo corrompido
		err = provider.Start(ctx)
		assert.Error(t, err)
	})

	t.Run("extremely_nested_yaml", func(t *testing.T) {
		// YAML com múltiplos níveis de aninhamento
		deepContent := `
level1:
  level2:
    level3:
      level4:
        level5:
          level6:
            level7:
              level8:
                level9:
                  level10:
                    deep_key: "very deep value"
                    array:
                      - item1
                      - item2:
                          nested: "array nested value"
`
		deepFile := filepath.Join(tempDir, "deep.yaml")
		err := os.WriteFile(deepFile, []byte(deepContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "deep.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Deve retornar chave não encontrada (comportamento atual do provider)
		result, err := provider.Translate(ctx, "level1.level2.level3.level4.level5.level6.level7.level8.level9.level10.deep_key", "en", nil)
		assert.NoError(t, err)
		assert.Equal(t, "level1.level2.level3.level4.level5.level6.level7.level8.level9.level10.deep_key", result)
	})

	t.Run("special_yaml_characters", func(t *testing.T) {
		// YAML com caracteres especiais e edge cases
		specialContent := `
special_chars: "Special: @#$%^&*()_+-=[]{}|;':\",./<>?"
multiline: |
  This is a multiline string
  with multiple lines
  and special characters: !@#$%
folded: >
  This is a folded string
  that should be on one line
  when parsed
unicode: "Unicode: 你好世界 🌍 Olá mundo"
null_value: null
empty_string: ""
boolean_true: true
boolean_false: false
number_int: 42
number_float: 3.14159
array:
  - "item with spaces"
  - "item:with:colons"
  - "item-with-dashes"
  - "item_with_underscores"
quotes:
  single: 'Single quotes'
  double: "Double quotes"
  mixed: 'Mixed "quotes" inside'
`
		specialFile := filepath.Join(tempDir, "special.yaml")
		err := os.WriteFile(specialFile, []byte(specialContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "special.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Testar diferentes tipos de valores
		tests := []struct {
			key      string
			expected string
		}{
			{"special_chars", "Special: @#$%^&*()_+-=[]{}|;':\",./<>?"},
			{"multiline", "This is a multiline string\nwith multiple lines\nand special characters: !@#$%\n"},
			{"folded", "This is a folded string that should be on one line when parsed\n"},
			{"unicode", "Unicode: 你好世界 🌍 Olá mundo"},
			{"null_value", "null_value"}, // Null values devem retornar a chave
			{"empty_string", ""},
			{"boolean_true", "boolean_true"},   // Valores não-string retornam a chave
			{"boolean_false", "boolean_false"}, // Valores não-string retornam a chave
			{"number_int", "number_int"},       // Valores não-string retornam a chave
			{"number_float", "number_float"},   // Valores não-string retornam a chave
			{"quotes.single", "quotes.single"}, // Chaves aninhadas não encontradas retornam a chave
			{"quotes.double", "quotes.double"}, // Chaves aninhadas não encontradas retornam a chave
			{"quotes.mixed", "quotes.mixed"},   // Chaves aninhadas não encontradas retornam a chave
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("key_%s", strings.ReplaceAll(tt.key, ".", "_")), func(t *testing.T) {
				result, err := provider.Translate(ctx, tt.key, "en", nil)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestProvider_FilePermissions - Testes de permissões de arquivo
func TestProvider_FilePermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping file permission tests when running as root")
	}

	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "restricted.yaml")

	// Criar arquivo YAML válido
	content := `key: "value"`
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(t, err)

	// Remover permissões de leitura
	err = os.Chmod(yamlFile, 0000)
	require.NoError(t, err)

	// Restaurar permissões após o teste
	defer func() {
		os.Chmod(yamlFile, 0644)
	}()

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "restricted.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Deve falhar devido a permissões
	err = provider.Start(ctx)
	assert.Error(t, err)
	assert.True(t, os.IsPermission(err) || strings.Contains(err.Error(), "permission denied"))
}

// TestProvider_ConcurrentAccess - Testes de concorrência
func TestProvider_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()

	// Criar arquivo YAML com múltiplas chaves
	content := `
greeting: "Hello, {{.name}}!"
farewell: "Goodbye, {{.name}}!"
`
	for i := 0; i < 100; i++ {
		content += fmt.Sprintf("key_%d: \"Value %d\"\n", i, i)
	}

	yamlFile := filepath.Join(tempDir, "concurrent.yaml")
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "concurrent.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = provider.Start(ctx)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	// Teste de concorrência com múltiplas goroutines
	const numGoroutines = 100
	const translationsPerGoroutine = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*translationsPerGoroutine)
	results := make(chan string, numGoroutines*translationsPerGoroutine)

	// Executar traduções concorrentes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < translationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d", (goroutineID*translationsPerGoroutine+j)%100)
				result, err := provider.Translate(ctx, key, "en", nil)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}
		}(i)
	}

	// Aguardar conclusão
	wg.Wait()
	close(errors)
	close(results)

	// Verificar se não houve erros
	for err := range errors {
		t.Errorf("Concurrent translation error: %v", err)
	}

	// Verificar se recebemos todos os resultados
	resultCount := 0
	for range results {
		resultCount++
	}
	assert.Equal(t, numGoroutines*translationsPerGoroutine, resultCount)
}

// TestProvider_TemplateProcessing - Testes avançados de templates
func TestProvider_TemplateProcessing(t *testing.T) {
	tempDir := t.TempDir()

	content := `
simple_template: "Hello, {{name}}!"
complex_template: "Welcome {{user.name}}, you have {{user.messages}} messages"
template_with_special_chars: "Price: ${{price}} ({{currency}})"
template_missing_param: "Hello, {{missing}}!"
template_with_html: "<h1>{{title}}</h1><p>{{content}}</p>"
template_with_numbers: "Count: {{count}} | Total: {{total}}"
empty_template: ""
no_template: "This has no template variables"
multiple_same_param: "{{name}} and {{name}} are friends"
conditional_like: "Status: {{if .active}}Active{{else}}Inactive{{end}}"
`

	yamlFile := filepath.Join(tempDir, "templates.yaml")
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "templates.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = provider.Start(ctx)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	tests := []struct {
		name     string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "simple_template_success",
			key:      "simple_template",
			params:   map[string]interface{}{"name": "John"},
			expected: "Hello, John!",
		},
		{
			name: "complex_nested_params",
			key:  "complex_template",
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "Alice",
					"messages": 5,
				},
			},
			expected: "Welcome {{user.name}}, you have {{user.messages}} messages", // Comportamento atual: não suporta params aninhados
		},
		{
			name:     "special_characters_in_template",
			key:      "template_with_special_chars",
			params:   map[string]interface{}{"price": "99.99", "currency": "USD"},
			expected: "Price: $99.99 (USD)",
		},
		{
			name:     "missing_parameter_handling",
			key:      "template_missing_param",
			params:   map[string]interface{}{"name": "John"},
			expected: "Hello, {{missing}}!", // Comportamento atual: mantém placeholder
		},
		{
			name:     "html_content_template",
			key:      "template_with_html",
			params:   map[string]interface{}{"title": "Welcome", "content": "This is a test"},
			expected: "<h1>Welcome</h1><p>This is a test</p>",
		},
		{
			name:     "numeric_parameters",
			key:      "template_with_numbers",
			params:   map[string]interface{}{"count": 42, "total": 100.50},
			expected: "Count: 42 | Total: 100.5",
		},
		{
			name:     "empty_template",
			key:      "empty_template",
			params:   map[string]interface{}{"name": "John"},
			expected: "",
		},
		{
			name:     "no_template_variables",
			key:      "no_template",
			params:   map[string]interface{}{"name": "John"},
			expected: "This has no template variables",
		},
		{
			name:     "multiple_same_parameter",
			key:      "multiple_same_param",
			params:   map[string]interface{}{"name": "Alice"},
			expected: "Alice and Alice are friends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Translate(ctx, tt.key, "en", tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProvider_LargeFiles - Teste com arquivos grandes
func TestProvider_LargeFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Criar arquivo YAML grande
	var content strings.Builder
	content.WriteString("# Large YAML file for performance testing\n")

	// Adicionar muitas chaves
	for i := 0; i < 5000; i++ {
		content.WriteString(fmt.Sprintf("key_%d: \"This is value number %d with some extra text to make it longer\"\n", i, i))
	}

	// Adicionar estruturas aninhadas
	content.WriteString("nested:\n")
	for i := 0; i < 1000; i++ {
		content.WriteString(fmt.Sprintf("  item_%d:\n", i))
		content.WriteString(fmt.Sprintf("    name: \"Item %d\"\n", i))
		content.WriteString(fmt.Sprintf("    value: %d\n", i*10))
		content.WriteString(fmt.Sprintf("    description: \"Description for item %d with more text\"\n", i))
	}

	largeFile := filepath.Join(tempDir, "large.yaml")
	err := os.WriteFile(largeFile, []byte(content.String()), 0644)
	require.NoError(t, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "large.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Medir tempo de carregamento
	startTime := time.Now()
	err = provider.Start(ctx)
	loadTime := time.Since(startTime)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	// Verificar se carregou em tempo razoável (deve ser < 2 segundos)
	assert.Less(t, loadTime, 2*time.Second, "Large file loading should be fast")

	// Testar acesso a diferentes partes do arquivo
	tests := []struct {
		key      string
		expected string
	}{
		{"key_0", "This is value number 0 with some extra text to make it longer"},
		{"key_1000", "This is value number 1000 with some extra text to make it longer"},
		{"key_4999", "This is value number 4999 with some extra text to make it longer"},
		{"nested.item_0.name", "nested.item_0.name"},                   // Chaves aninhadas não funcionam no provider atual
		{"nested.item_500.value", "nested.item_500.value"},             // Chaves aninhadas não funcionam no provider atual
		{"nested.item_999.description", "nested.item_999.description"}, // Chaves aninhadas não funcionam no provider atual
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("access_%s", strings.ReplaceAll(tt.key, ".", "_")), func(t *testing.T) {
			startTime := time.Now()
			result, err := provider.Translate(ctx, tt.key, "en", nil)
			accessTime := time.Since(startTime)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Less(t, accessTime, 10*time.Millisecond, "Translation access should be fast")
		})
	}
}

// TestProvider_UnicodeHandling - Testes de suporte Unicode
func TestProvider_UnicodeHandling(t *testing.T) {
	tempDir := t.TempDir()

	content := `
# Unicode test file
chinese: "你好世界"
japanese: "こんにちは世界"
korean: "안녕하세요 세계"
arabic: "مرحبا بالعالم"
russian: "Привет мир"
emoji: "Hello 👋 World 🌍"
mixed: "English + 中文 + 日本語 + 한국어"
template_unicode: "Hello {{name}}, welcome to {{place}}!"
special_unicode: "Symbols: ∀∂∈ℝ∧∨∩∪≈≠≤≥"
zero_width: "Before‌After"  # Contains zero-width non-joiner
rtl_text: "العربية"
combining: "é à ñ ç"  # Combining characters
`

	unicodeFile := filepath.Join(tempDir, "unicode.yaml")
	err := os.WriteFile(unicodeFile, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "unicode.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = provider.Start(ctx)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	tests := []struct {
		name     string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{"chinese", "chinese", nil, "你好世界"},
		{"japanese", "japanese", nil, "こんにちは世界"},
		{"korean", "korean", nil, "안녕하세요 세계"},
		{"arabic", "arabic", nil, "مرحبا بالعالم"},
		{"russian", "russian", nil, "Привет мир"},
		{"emoji", "emoji", nil, "Hello 👋 World 🌍"},
		{"mixed_languages", "mixed", nil, "English + 中文 + 日本語 + 한국어"},
		{
			"template_with_unicode",
			"template_unicode",
			map[string]interface{}{"name": "José", "place": "São Paulo"},
			"Hello José, welcome to São Paulo!",
		},
		{"mathematical_symbols", "special_unicode", nil, "Symbols: ∀∂∈ℝ∧∨∩∪≈≠≤≥"},
		{"zero_width_characters", "zero_width", nil, "Before‌After"},
		{"rtl_text", "rtl_text", nil, "العربية"},
		{"combining_characters", "combining", nil, "é à ñ ç"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Translate(ctx, tt.key, "en", tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			// Verificar que o comprimento da string está correto
			assert.Greater(t, len([]rune(result)), 0, "Result should contain unicode characters")
		})
	}
}

// TestProvider_CountingMethods - Testes para métodos de contagem e listagem
func TestProvider_CountingMethods(t *testing.T) {
	tempDir := t.TempDir()

	// Criar arquivos de tradução para múltiplas linguagens
	enTranslations := map[string]interface{}{
		"simple":  "Hello",
		"welcome": "Welcome, {{name}}!",
		"nested": map[string]interface{}{
			"greeting": "Good morning",
			"farewell": "Goodbye",
			"deep": map[string]interface{}{
				"message": "Deep nested message",
			},
		},
		"empty": "",
	}

	ptTranslations := map[string]interface{}{
		"simple":  "Olá",
		"welcome": "Bem-vindo, {{name}}!",
		"other":   "Outra mensagem",
	}

	esTranslations := map[string]interface{}{
		"simple": "Hola",
		"nested": map[string]interface{}{
			"greeting": "Buenos días",
		},
	}

	// Escrever arquivos YAML
	enFile := filepath.Join(tempDir, "en.yaml")
	enData, _ := yaml.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	ptFile := filepath.Join(tempDir, "pt.yaml")
	ptData, _ := yaml.Marshal(ptTranslations)
	require.NoError(t, os.WriteFile(ptFile, ptData, 0644))

	esFile := filepath.Join(tempDir, "es.yaml")
	esData, _ := yaml.Marshal(esTranslations)
	require.NoError(t, os.WriteFile(esFile, esData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.yaml",
			Encoding:    "utf-8",
			NestedKeys:  true,
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = provider.Start(ctx)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	t.Run("GetTranslationCount", func(t *testing.T) {
		// Contagem total esperada:
		// en: simple(1) + welcome(1) + nested.greeting(1) + nested.farewell(1) + nested.deep.message(1) + empty(1) = 6
		// pt: simple(1) + welcome(1) + other(1) = 3
		// es: simple(1) + nested.greeting(1) = 2
		// Total: 11
		total := provider.GetTranslationCount()
		assert.Equal(t, 11, total)
	})

	t.Run("GetTranslationCountByLanguage", func(t *testing.T) {
		// Testar contagem por linguagem
		enCount := provider.GetTranslationCountByLanguage("en")
		assert.Equal(t, 6, enCount, "English should have 6 translations")

		ptCount := provider.GetTranslationCountByLanguage("pt")
		assert.Equal(t, 3, ptCount, "Portuguese should have 3 translations")

		esCount := provider.GetTranslationCountByLanguage("es")
		assert.Equal(t, 2, esCount, "Spanish should have 2 translations")

		// Testar linguagem não existente
		nonExistentCount := provider.GetTranslationCountByLanguage("fr")
		assert.Equal(t, 0, nonExistentCount, "Non-existent language should return 0")
	})

	t.Run("GetLoadedLanguages", func(t *testing.T) {
		loadedLanguages := provider.GetLoadedLanguages()

		// Verificar que todas as linguagens foram carregadas
		assert.Len(t, loadedLanguages, 3, "Should have loaded 3 languages")

		// Verificar que contém todas as linguagens esperadas
		expectedLangs := []string{"en", "pt", "es"}
		for _, expectedLang := range expectedLangs {
			assert.Contains(t, loadedLanguages, expectedLang, "Should contain language %s", expectedLang)
		}
	})
}

func TestProvider_CountingEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("EmptyTranslations", func(t *testing.T) {
		// Criar arquivo YAML vazio
		emptyFile := filepath.Join(tempDir, "empty.yaml")
		err := os.WriteFile(emptyFile, []byte(""), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "empty.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		assert.Equal(t, 0, provider.GetTranslationCount())
		assert.Equal(t, 0, provider.GetTranslationCountByLanguage("en"))

		loadedLanguages := provider.GetLoadedLanguages()
		assert.Len(t, loadedLanguages, 1)
		assert.Contains(t, loadedLanguages, "en")
	})

	t.Run("MixedDataTypes", func(t *testing.T) {
		// Criar arquivo YAML com tipos mistos
		mixedContent := `string_value: "This is a string"
number_value: 42
boolean_value: true
array_value:
  - "item1"
  - "item2"
null_value: null
empty_string: ""
nested_with_mixed:
  string_in_nested: "Nested string"
  number_in_nested: 123
  array_in_nested:
    - "nested_array_item"
  deeply_nested:
    final_string: "Deep string"
    final_number: 456
`
		mixedFile := filepath.Join(tempDir, "mixed.yaml")
		err := os.WriteFile(mixedFile, []byte(mixedContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "mixed.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Apenas valores string devem ser contados
		// string_value(1) + empty_string(1) + nested_with_mixed.string_in_nested(1) + nested_with_mixed.deeply_nested.final_string(1) = 4
		totalCount := provider.GetTranslationCount()
		assert.Equal(t, 4, totalCount, "Only string values should be counted")

		enCount := provider.GetTranslationCountByLanguage("en")
		assert.Equal(t, 4, enCount, "Only string values should be counted for 'en'")
	})

	t.Run("DeeplyNestedStructure", func(t *testing.T) {
		// Criar estrutura profundamente aninhada
		deepContent := `level1:
  level2:
    level3:
      level4:
        level5:
          translation1: "Deep translation 1"
          translation2: "Deep translation 2"
          level6:
            translation3: "Very deep translation"
            level7:
              translation4: "Extremely deep translation"
    another_branch:
      translation5: "Branch translation"
root_translation: "Root level translation"
`
		deepFile := filepath.Join(tempDir, "deep.yaml")
		err := os.WriteFile(deepFile, []byte(deepContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "deep.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Contar todas as strings recursivamente: 6 strings totais
		totalCount := provider.GetTranslationCount()
		assert.Equal(t, 6, totalCount, "Should count all nested string translations")

		enCount := provider.GetTranslationCountByLanguage("en")
		assert.Equal(t, 6, enCount, "Should count all nested string translations for 'en'")
	})

	t.Run("BeforeStart", func(t *testing.T) {
		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "{lang}.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		// Testar métodos antes do Start()
		assert.Equal(t, 0, provider.GetTranslationCount())
		assert.Equal(t, 0, provider.GetTranslationCountByLanguage("en"))
		assert.Empty(t, provider.GetLoadedLanguages())
	})
}

func TestProvider_ConcurrentCounting(t *testing.T) {
	tempDir := t.TempDir()

	// Criar arquivo com muitas traduções
	var content strings.Builder
	for i := 0; i < 1000; i++ {
		content.WriteString(fmt.Sprintf("key_%d: \"Translation %d\"\n", i, i))
	}

	// Adicionar estrutura aninhada
	content.WriteString("nested:\n")
	for i := 0; i < 500; i++ {
		content.WriteString(fmt.Sprintf("  nested_key_%d: \"Nested translation %d\"\n", i, i))
	}

	largeFile := filepath.Join(tempDir, "large.yaml")
	err := os.WriteFile(largeFile, []byte(content.String()), 0644)
	require.NoError(t, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "large.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = provider.Start(ctx)
	require.NoError(t, err)
	defer provider.Stop(ctx)

	// Executar operações de contagem concorrentemente
	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make(chan int, numGoroutines*3) // 3 operações por goroutine

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Executar as três operações de contagem
			totalCount := provider.GetTranslationCount()
			enCount := provider.GetTranslationCountByLanguage("en")
			loadedLangs := len(provider.GetLoadedLanguages())

			results <- totalCount
			results <- enCount
			results <- loadedLangs
		}()
	}

	wg.Wait()
	close(results)

	// Verificar que todos os resultados são consistentes
	expectedTotal := 1500 // 1000 + 500 traduções
	expectedEnCount := 1500
	expectedLangCount := 1

	totalResults := make([]int, 0, numGoroutines*3)
	for result := range results {
		totalResults = append(totalResults, result)
	}

	// Verificar que temos o número correto de resultados
	assert.Len(t, totalResults, numGoroutines*3)

	// Verificar que os resultados estão corretos (devem ser alternados entre os 3 tipos)
	for i := 0; i < len(totalResults); i += 3 {
		assert.Equal(t, expectedTotal, totalResults[i], "Total count should be consistent")
		assert.Equal(t, expectedEnCount, totalResults[i+1], "Language count should be consistent")
		assert.Equal(t, expectedLangCount, totalResults[i+2], "Loaded languages count should be consistent")
	}
}

// TestProvider_GetTranslationCount - Comprehensive tests for GetTranslationCount method
func TestProvider_GetTranslationCount(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("SingleLanguageSimpleTranslations", func(t *testing.T) {
		// Create YAML with simple key-value pairs
		content := `
hello: "Hello"
goodbye: "Goodbye"
welcome: "Welcome"
`
		yamlFile := filepath.Join(tempDir, "simple.yaml")
		err := os.WriteFile(yamlFile, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "simple.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		assert.Equal(t, 3, count, "Should count 3 simple translations")
	})

	t.Run("MultipleLanguagesSimpleTranslations", func(t *testing.T) {
		// Create multiple language files
		enContent := `
hello: "Hello"
goodbye: "Goodbye"
`
		ptContent := `
hello: "Olá"
goodbye: "Tchau"
welcome: "Bem-vindo"
`
		esContent := `
hello: "Hola"
`

		enFile := filepath.Join(tempDir, "en.yaml")
		ptFile := filepath.Join(tempDir, "pt.yaml")
		esFile := filepath.Join(tempDir, "es.yaml")

		require.NoError(t, os.WriteFile(enFile, []byte(enContent), 0644))
		require.NoError(t, os.WriteFile(ptFile, []byte(ptContent), 0644))
		require.NoError(t, os.WriteFile(esFile, []byte(esContent), 0644))

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en", "pt", "es").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "{lang}.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		// en: 2, pt: 3, es: 1 = total 6
		assert.Equal(t, 6, count, "Should count translations across all languages")
	})

	t.Run("NestedTranslations", func(t *testing.T) {
		content := `simple: "Simple value"
level1:
  nested1: "Nested value 1"
  nested2: "Nested value 2"
  level2:
    deep1: "Deep value 1"
    deep2: "Deep value 2"
    level3:
      verydeep: "Very deep value"
`
		yamlFile := filepath.Join(tempDir, "nested.yaml")
		err := os.WriteFile(yamlFile, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "nested.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		// simple(1) + nested1(1) + nested2(1) + deep1(1) + deep2(1) + verydeep(1) = 6
		assert.Equal(t, 6, count, "Should count all nested string translations")
	})

	t.Run("MixedDataTypes", func(t *testing.T) {
		content := `string_value: "This is a string"
number_value: 42
boolean_value: true
array_value:
  - "array item 1"
  - "array item 2"
null_value: null
empty_string: ""
nested_mixed:
  string_in_nested: "Nested string"
  number_in_nested: 123
  boolean_in_nested: false
  array_in_nested:
    - "nested array item"
  deeply_nested:
    final_string: "Final string"
    final_number: 456
    final_boolean: true
`
		yamlFile := filepath.Join(tempDir, "mixed.yaml")
		err := os.WriteFile(yamlFile, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "mixed.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		// Only string values: string_value(1) + empty_string(1) + string_in_nested(1) + final_string(1) = 4
		assert.Equal(t, 4, count, "Should only count string values, not numbers, booleans, or arrays")
	})

	t.Run("EmptyTranslations", func(t *testing.T) {
		// Empty YAML file
		yamlFile := filepath.Join(tempDir, "empty.yaml")
		err := os.WriteFile(yamlFile, []byte(""), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "empty.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		assert.Equal(t, 0, count, "Empty translations should return count of 0")
	})

	t.Run("OnlyNullAndEmptyFile", func(t *testing.T) {
		content := `
null_value: null
empty_map: {}
empty_array: []
`
		yamlFile := filepath.Join(tempDir, "nulls.yaml")
		err := os.WriteFile(yamlFile, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "nulls.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		assert.Equal(t, 0, count, "Non-string values should not be counted")
	})

	t.Run("BeforeStart", func(t *testing.T) {
		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "{lang}.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		// Test before Start() is called
		count := provider.GetTranslationCount()
		assert.Equal(t, 0, count, "Should return 0 before translations are loaded")
	})

	t.Run("AfterStop", func(t *testing.T) {
		content := `hello: "Hello"`
		yamlFile := filepath.Join(tempDir, "stop_test.yaml")
		err := os.WriteFile(yamlFile, []byte(content), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "stop_test.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)

		// Verify count after start
		count := provider.GetTranslationCount()
		assert.Equal(t, 1, count, "Should count 1 translation after start")

		// Stop the provider
		err = provider.Stop(ctx)
		require.NoError(t, err)

		// Count should still work after stop (translations remain in memory)
		count = provider.GetTranslationCount()
		assert.Equal(t, 1, count, "Should still count translations after stop")
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		// Create large translation file
		var content strings.Builder
		for i := 0; i < 1000; i++ {
			content.WriteString(fmt.Sprintf("key_%d: \"Translation %d\"\n", i, i))
		}

		yamlFile := filepath.Join(tempDir, "concurrent.yaml")
		err := os.WriteFile(yamlFile, []byte(content.String()), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "concurrent.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Test concurrent access
		const numGoroutines = 50
		var wg sync.WaitGroup
		results := make(chan int, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				count := provider.GetTranslationCount()
				results <- count
			}()
		}

		wg.Wait()
		close(results)

		// All results should be the same
		expectedCount := 1000
		for count := range results {
			assert.Equal(t, expectedCount, count, "All concurrent calls should return the same count")
		}
	})

	t.Run("LargeNestedStructure", func(t *testing.T) {
		var content strings.Builder
		content.WriteString("root_key: \"Root value\"\n")

		// Create deeply nested structure
		for level := 0; level < 5; level++ {
			content.WriteString(fmt.Sprintf("level_%d:\n", level))
			for i := 0; i < 10; i++ {
				indent := strings.Repeat("  ", level+1)
				content.WriteString(fmt.Sprintf("%snested_key_%d_%d: \"Nested value %d at level %d\"\n", indent, level, i, i, level))
			}
		}

		yamlFile := filepath.Join(tempDir, "large_nested.yaml")
		err := os.WriteFile(yamlFile, []byte(content.String()), 0644)
		require.NoError(t, err)

		cfg, err := config.NewConfigBuilder().
			WithSupportedLanguages("en").
			WithDefaultLanguage("en").
			WithProviderConfig(&config.YAMLProviderConfig{
				FilePath:    tempDir,
				FilePattern: "large_nested.yaml",
				Encoding:    "utf-8",
			}).
			Build()
		require.NoError(t, err)

		factory := &Factory{}
		provider, err := factory.Create(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		count := provider.GetTranslationCount()
		// root_key(1) + 5 levels * 10 items per level = 1 + 50 = 51
		assert.Equal(t, 51, count, "Should count all nested translations in large structure")
	})
}

func BenchmarkProvider_CountingMethods(b *testing.B) {
	tempDir := b.TempDir()

	// Criar arquivo com muitas traduções para benchmark
	var content strings.Builder
	for i := 0; i < 10000; i++ {
		content.WriteString(fmt.Sprintf("key_%d: \"Translation %d\"\n", i, i))
	}

	// Adicionar estruturas aninhadas
	for level := 0; level < 5; level++ {
		content.WriteString(fmt.Sprintf("level_%d:\n", level))
		for i := 0; i < 1000; i++ {
			content.WriteString(fmt.Sprintf("  nested_key_%d_%d: \"Nested translation %d at level %d\"\n", level, i, i, level))
		}
	}

	benchFile := filepath.Join(tempDir, "bench.yaml")
	err := os.WriteFile(benchFile, []byte(content.String()), 0644)
	require.NoError(b, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "bench.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(b, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(b, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(b, err)
	defer provider.Stop(ctx)

	b.ResetTimer()

	b.Run("GetTranslationCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.GetTranslationCount()
		}
	})

	b.Run("GetTranslationCountByLanguage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.GetTranslationCountByLanguage("en")
		}
	})

	b.Run("GetLoadedLanguages", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.GetLoadedLanguages()
		}
	})

	b.Run("ConcurrentCounting", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = provider.GetTranslationCount()
				_ = provider.GetTranslationCountByLanguage("en")
				_ = provider.GetLoadedLanguages()
			}
		})
	})
} // Benchmarks para medir performance
func BenchmarkYAMLProvider(b *testing.B) {
	tempDir := b.TempDir()

	// Criar arquivo de benchmark
	content := `
greeting: "Hello World"
template: "Hello, {{.name}}!"
`
	for i := 0; i < 100; i++ {
		content += fmt.Sprintf("key_%d: \"Benchmark value %d\"\n", i, i)
	}

	yamlFile := filepath.Join(tempDir, "bench.yaml")
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(b, err)

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:    tempDir,
			FilePattern: "bench.yaml",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(b, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(b, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(b, err)
	defer provider.Stop(ctx)

	b.ResetTimer()

	b.Run("simple_translation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.Translate(ctx, "greeting", "en", nil)
		}
	})

	b.Run("template_translation", func(b *testing.B) {
		params := map[string]interface{}{"name": "John"}
		for i := 0; i < b.N; i++ {
			_, _ = provider.Translate(ctx, "template", "en", params)
		}
	})

	b.Run("concurrent_translation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = provider.Translate(ctx, "greeting", "en", nil)
			}
		})
	})
}
