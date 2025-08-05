package yaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
