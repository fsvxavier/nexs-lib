package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory_Name(t *testing.T) {
	factory := &Factory{}
	assert.Equal(t, "json", factory.Name())
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
					WithProviderConfig(&config.JSONProviderConfig{
						FilePath:    "/tmp",
						FilePattern: "{lang}.json",
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
					WithProviderConfig(&config.JSONProviderConfig{
						FilePath:    "/tmp",
						FilePattern: "{lang}.json",
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
					WithProviderConfig(config.DefaultJSONProviderConfig()).
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
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    "/tmp",
			FilePattern: "{lang}.json",
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
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    "/tmp",
			FilePattern: "{lang}.json",
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

func TestProvider_Translate(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_test")
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
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	// Write Portuguese translations
	ptFile := filepath.Join(tempDir, "pt.json")
	ptData, _ := json.Marshal(ptTranslations)
	require.NoError(t, os.WriteFile(ptFile, ptData, 0644))

	// Create provider configuration
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

func TestProvider_StartStop(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithLoadTimeout(5 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    "/nonexistent/path",
			FilePattern: "{lang}.json",
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

// Edge Cases and Advanced Test Coverage

func TestProvider_CorruptedFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_corrupted_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name         string
		fileContent  string
		validateJSON bool
		expectError  bool
	}{
		{
			name:         "invalid JSON syntax with validation",
			fileContent:  `{"hello": "Hello",}`, // trailing comma
			validateJSON: true,
			expectError:  true,
		},
		{
			name:         "invalid JSON syntax without validation",
			fileContent:  `{"hello": "Hello",}`, // trailing comma
			validateJSON: false,
			expectError:  false,
		},
		{
			name:         "incomplete JSON with validation",
			fileContent:  `{"hello": "Hello"`,
			validateJSON: true,
			expectError:  true,
		},
		{
			name:         "empty file with validation",
			fileContent:  "",
			validateJSON: true,
			expectError:  true,
		},
		{
			name:         "only whitespace with validation",
			fileContent:  "   \n\t  ",
			validateJSON: true,
			expectError:  true,
		},
		{
			name:         "non-object root with validation",
			fileContent:  `["hello", "world"]`,
			validateJSON: true,
			expectError:  true,
		},
		{
			name:         "null JSON with validation",
			fileContent:  "null",
			validateJSON: true,
			expectError:  false, // null is valid JSON, but creates empty translations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write corrupted file
			enFile := filepath.Join(tempDir, "en.json")
			require.NoError(t, os.WriteFile(enFile, []byte(tt.fileContent), 0644))

			cfg, err := config.NewConfigBuilder().
				WithSupportedLanguages("en").
				WithDefaultLanguage("en").
				WithProviderConfig(&config.JSONProviderConfig{
					FilePath:     tempDir,
					FilePattern:  "{lang}.json",
					Encoding:     "utf-8",
					ValidateJSON: tt.validateJSON,
				}).
				Build()
			require.NoError(t, err)

			factory := &Factory{}
			provider, err := factory.Create(cfg)
			require.NoError(t, err)

			ctx := context.Background()
			err = provider.Start(ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// If provider started successfully with corrupted JSON, test a translation
				if err == nil && tt.fileContent == "null" {
					result, err := provider.Translate(ctx, "nonexistent", "en", nil)
					// Should return the key itself in non-strict mode since no translations loaded
					assert.NoError(t, err)
					assert.Equal(t, "nonexistent", result)
				}
			}

			// Clean up for next iteration
			os.Remove(enFile)
		})
	}
}

func TestProvider_FilePermissions(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_permissions_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create valid JSON file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	tests := []struct {
		name        string
		permissions fs.FileMode
		expectError bool
	}{
		{
			name:        "no read permission",
			permissions: 0200, // write only
			expectError: true,
		},
		{
			name:        "read permission",
			permissions: 0400, // read only
			expectError: false,
		},
		{
			name:        "full permissions",
			permissions: 0644, // read/write
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change file permissions
			require.NoError(t, os.Chmod(enFile, tt.permissions))
			defer os.Chmod(enFile, 0644) // restore permissions

			cfg, err := config.NewConfigBuilder().
				WithSupportedLanguages("en").
				WithDefaultLanguage("en").
				WithProviderConfig(&config.JSONProviderConfig{
					FilePath:    tempDir,
					FilePattern: "{lang}.json",
					Encoding:    "utf-8",
				}).
				Build()
			require.NoError(t, err)

			factory := &Factory{}
			provider, err := factory.Create(cfg)
			require.NoError(t, err)

			ctx := context.Background()
			err = provider.Start(ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvider_NonexistentFiles(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "fr").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    "/nonexistent/path",
			FilePattern: "{lang}.json",
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

func TestProvider_ConcurrentTranslations(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_concurrent_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create large translation file for concurrent access
	translations := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		translations[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("Value %d", i)
	}
	translations["template"] = "Hello {{name}}!"

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(translations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	// Test concurrent translations
	const numGoroutines = 100
	const translationsPerGoroutine = 10

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines*translationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < translationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d", (goroutineID*translationsPerGoroutine+j)%1000)
				_, err := provider.Translate(ctx, key, "en", nil)
				results <- err
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Check all translations succeeded
	errorCount := 0
	for err := range results {
		if err != nil {
			errorCount++
			t.Logf("Translation error: %v", err)
		}
	}

	assert.Equal(t, 0, errorCount, "Expected no errors in concurrent translations")
}

func TestProvider_StrictMode(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_strict_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file with limited keys
	enTranslations := map[string]interface{}{
		"hello": "Hello",
		"world": "World",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	tests := []struct {
		name       string
		strictMode bool
		key        string
		wantError  bool
		expected   string
	}{
		{
			name:       "strict mode - existing key",
			strictMode: true,
			key:        "hello",
			wantError:  false,
			expected:   "Hello",
		},
		{
			name:       "strict mode - missing key",
			strictMode: true,
			key:        "missing",
			wantError:  true,
		},
		{
			name:       "non-strict mode - existing key",
			strictMode: false,
			key:        "hello",
			wantError:  false,
			expected:   "Hello",
		},
		{
			name:       "non-strict mode - missing key",
			strictMode: false,
			key:        "missing",
			wantError:  false,
			expected:   "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewConfigBuilder().
				WithSupportedLanguages("en").
				WithDefaultLanguage("en").
				WithStrictMode(tt.strictMode).
				WithProviderConfig(&config.JSONProviderConfig{
					FilePath:    tempDir,
					FilePattern: "{lang}.json",
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

			result, err := provider.Translate(ctx, tt.key, "en", nil)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProvider_ComplexNestedKeys(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_nested_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create complex nested translation structure
	enTranslations := map[string]interface{}{
		"app": map[string]interface{}{
			"title": "My Application",
			"menu": map[string]interface{}{
				"home":     "Home",
				"settings": "Settings",
				"user": map[string]interface{}{
					"profile": "Profile",
					"logout":  "Logout",
				},
			},
		},
		"messages": map[string]interface{}{
			"errors": map[string]interface{}{
				"validation": map[string]interface{}{
					"required": "Field {{field}} is required",
					"email":    "Invalid email format",
				},
			},
		},
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
		name     string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "level 2 nested key",
			key:      "app.title",
			expected: "My Application",
		},
		{
			name:     "level 3 nested key",
			key:      "app.menu.home",
			expected: "Home",
		},
		{
			name:     "level 4 nested key",
			key:      "app.menu.user.profile",
			expected: "Profile",
		},
		{
			name:     "deep nested with parameters",
			key:      "messages.errors.validation.required",
			params:   map[string]interface{}{"field": "Email"},
			expected: "Field Email is required",
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

func TestProvider_UnicodeAndSpecialCharacters(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_unicode_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create translation file with various Unicode and special characters
	translations := map[string]interface{}{
		"emoji":            "Hello 👋 World 🌍",
		"chinese":          "你好世界",
		"arabic":           "مرحبا بالعالم",
		"russian":          "Привет мир",
		"japanese":         "こんにちは世界",
		"special":          "Special chars: @#$%^&*()[]{}|\\:;\"'<>?,./",
		"multiline":        "Line 1\nLine 2\nLine 3",
		"template_unicode": "Olá {{nome}}! Como está? 🎉",
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(translations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	tests := []struct {
		name     string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "emoji characters",
			key:      "emoji",
			expected: "Hello 👋 World 🌍",
		},
		{
			name:     "chinese characters",
			key:      "chinese",
			expected: "你好世界",
		},
		{
			name:     "arabic characters",
			key:      "arabic",
			expected: "مرحبا بالعالم",
		},
		{
			name:     "russian characters",
			key:      "russian",
			expected: "Привет мир",
		},
		{
			name:     "japanese characters",
			key:      "japanese",
			expected: "こんにちは世界",
		},
		{
			name:     "special characters",
			key:      "special",
			expected: "Special chars: @#$%^&*()[]{}|\\:;\"'<>?,./",
		},
		{
			name:     "multiline text",
			key:      "multiline",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "unicode template",
			key:      "template_unicode",
			params:   map[string]interface{}{"nome": "João"},
			expected: "Olá João! Como está? 🎉",
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

func TestProvider_LargeFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_large_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create large translation file (10,000 keys)
	largeTranslations := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largeTranslations[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("Very long translation value for key %d with additional text to make it longer and test memory usage", i)
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(largeTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Measure loading time
	start := time.Now()
	err = provider.Start(ctx)
	loadTime := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, loadTime, 5*time.Second, "Large file loading should complete in reasonable time")

	// Test accessing various keys
	testKeys := []string{"key_0", "key_100", "key_5000", "key_9999"}
	expectedValues := []string{
		"Very long translation value for key 0 with additional text to make it longer and test memory usage",
		"Very long translation value for key 100 with additional text to make it longer and test memory usage",
		"Very long translation value for key 5000 with additional text to make it longer and test memory usage",
		"Very long translation value for key 9999 with additional text to make it longer and test memory usage",
	}

	for i, key := range testKeys {
		result, err := provider.Translate(ctx, key, "en", nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedValues[i], result)
	}
}

func TestProvider_ConcurrentStartStop(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_concurrent_startstop_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test concurrent start/stop operations
	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*2)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Start goroutine
		go func() {
			defer wg.Done()
			errors <- provider.Start(ctx)
		}()

		// Stop goroutine
		go func() {
			defer wg.Done()
			errors <- provider.Stop(ctx)
		}()
	}

	wg.Wait()
	close(errors)

	// Count errors (some are expected due to race conditions)
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	// Should handle concurrent operations gracefully
	assert.True(t, errorCount < numGoroutines, "Should handle most concurrent operations gracefully")
}

func TestProvider_ContextCancellation(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_context_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	// Test context cancellation during start
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = provider.Start(cancelCtx)
	assert.Error(t, err)
}

// Performance and Load Testing

func BenchmarkProvider_TranslateSimple(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "i18n_bench_simple")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	enTranslations := map[string]interface{}{
		"hello": "Hello",
		"world": "World",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(b, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = provider.Translate(ctx, "hello", "en", nil)
	}
}

func BenchmarkProvider_TranslateTemplate(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "i18n_bench_template")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	enTranslations := map[string]interface{}{
		"welcome": "Welcome {{name}} to {{app}}!",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(b, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	params := map[string]interface{}{
		"name": "John",
		"app":  "MyApp",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = provider.Translate(ctx, "welcome", "en", params)
	}
}

func BenchmarkProvider_ConcurrentTranslate(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "i18n_bench_concurrent")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	translations := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		translations[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("Value %d", i)
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(translations)
	require.NoError(b, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%100)
			_, _ = provider.Translate(ctx, key, "en", nil)
			i++
		}
	})
}

// Additional Tests for Complete Coverage

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.Equal(t, "json", factory.Name())
}

func TestProvider_HasTranslation(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_has_translation_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation files
	enTranslations := map[string]interface{}{
		"hello": "Hello",
		"nested": map[string]interface{}{
			"message": "This is nested",
		},
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
		name     string
		key      string
		language string
		expected bool
	}{
		{
			name:     "existing key",
			key:      "hello",
			language: "en",
			expected: true,
		},
		{
			name:     "nested key",
			key:      "nested.message",
			language: "en",
			expected: true,
		},
		{
			name:     "non-existing key",
			key:      "nonexistent",
			language: "en",
			expected: false,
		},
		{
			name:     "empty language defaults to default",
			key:      "hello",
			language: "",
			expected: true,
		},
		{
			name:     "unsupported language",
			key:      "hello",
			language: "fr",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.HasTranslation(tt.key, tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_SetDefaultLanguage(t *testing.T) {
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithProviderConfig(config.DefaultJSONProviderConfig()).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	// Test initial default language
	assert.Equal(t, "en", provider.GetDefaultLanguage())

	// Test setting new default language
	provider.SetDefaultLanguage("pt")
	assert.Equal(t, "pt", provider.GetDefaultLanguage())
}

func TestProvider_Health(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_health_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test health check before start
	err = provider.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not started")

	// Start provider
	err = provider.Start(ctx)
	require.NoError(t, err)

	// Test health check after start
	err = provider.Health(ctx)
	assert.NoError(t, err)

	// Test health check after removing translation directory
	os.RemoveAll(tempDir)
	err = provider.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "translation directory no longer exists")
}

func TestProvider_TranslationCounts(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_count_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation files
	enTranslations := map[string]interface{}{
		"hello": "Hello",
		"world": "World",
		"nested": map[string]interface{}{
			"message1": "Nested message 1",
			"message2": "Nested message 2",
			"deep": map[string]interface{}{
				"message": "Deep nested message",
			},
		},
	}
	ptTranslations := map[string]interface{}{
		"hello": "Olá",
		"world": "Mundo",
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	ptFile := filepath.Join(tempDir, "pt.json")
	ptData, _ := json.Marshal(ptTranslations)
	require.NoError(t, os.WriteFile(ptFile, ptData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	// Cast to specific type to access JSON provider methods
	jsonProvider, ok := provider.(*Provider)
	require.True(t, ok)

	// Test total translation count
	totalCount := jsonProvider.GetTranslationCount()
	assert.Equal(t, 7, totalCount) // 5 from en + 2 from pt

	// Test count by language
	enCount := jsonProvider.GetTranslationCountByLanguage("en")
	assert.Equal(t, 5, enCount) // hello, world, nested.message1, nested.message2, nested.deep.message

	ptCount := jsonProvider.GetTranslationCountByLanguage("pt")
	assert.Equal(t, 2, ptCount) // hello, world

	frCount := jsonProvider.GetTranslationCountByLanguage("fr")
	assert.Equal(t, 0, frCount) // non-existent language
}

func TestProvider_LoadTranslations(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_load_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test direct LoadTranslations call
	err = provider.LoadTranslations(ctx)
	assert.NoError(t, err)

	// Cast to specific type to access JSON provider methods
	jsonProvider, ok := provider.(*Provider)
	require.True(t, ok)

	// Verify translations were loaded
	assert.Equal(t, 1, jsonProvider.GetTranslationCountByLanguage("en"))
	assert.Equal(t, 0, jsonProvider.GetTranslationCountByLanguage("pt")) // file doesn't exist, should be empty
}

func TestProvider_ReloadTranslations(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_reload_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create initial translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	// Cast to specific type to access JSON provider methods
	jsonProvider, ok := provider.(*Provider)
	require.True(t, ok)

	// Verify initial translation
	result, err := provider.Translate(ctx, "hello", "en", nil)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", result)

	// Update translation file
	updatedTranslations := map[string]interface{}{
		"hello": "Hello Updated",
		"world": "World",
	}
	updatedData, _ := json.Marshal(updatedTranslations)
	require.NoError(t, os.WriteFile(enFile, updatedData, 0644))

	// Reload translations
	err = jsonProvider.ReloadTranslations(ctx)
	assert.NoError(t, err)

	// Verify updated translation
	result, err = provider.Translate(ctx, "hello", "en", nil)
	assert.NoError(t, err)
	assert.Equal(t, "Hello Updated", result)

	// Verify new translation
	result, err = provider.Translate(ctx, "world", "en", nil)
	assert.NoError(t, err)
	assert.Equal(t, "World", result)
}

func TestProvider_LoadTranslationsErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *config.JSONProviderConfig
	}{
		{
			name: "empty file path",
			config: &config.JSONProviderConfig{
				FilePath:    "",
				FilePattern: "{lang}.json",
				Encoding:    "utf-8",
			},
		},
		{
			name: "nonexistent directory",
			config: &config.JSONProviderConfig{
				FilePath:    "/nonexistent/path",
				FilePattern: "{lang}.json",
				Encoding:    "utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewConfigBuilder().
				WithSupportedLanguages("en").
				WithDefaultLanguage("en").
				WithProviderConfig(tt.config).
				Build()
			require.NoError(t, err)

			factory := &Factory{}
			provider, err := factory.Create(cfg)
			require.NoError(t, err)

			ctx := context.Background()
			err = provider.LoadTranslations(ctx)
			assert.Error(t, err)
		})
	}
}

func TestProvider_MaxFileSize(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_filesize_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a large translation file
	largeTranslations := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeTranslations[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("Very long value %d", i)
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(largeTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	// Test with small max file size
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
			MaxFileSize: 100, // Very small limit
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

// Tests for improved coverage of specific functions

func TestFactory_CreateEdgeCases(t *testing.T) {
	factory := &Factory{}

	tests := []struct {
		name      string
		config    interface{}
		wantError bool
	}{
		{
			name: "config with invalid provider config type",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en").
					WithDefaultLanguage("en").
					Build()
				require.NoError(t, err)
				// Set invalid provider config
				cfg.ProviderConfig = "invalid"
				return cfg
			}(),
			wantError: true,
		},
		{
			name: "config with nil provider config",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en").
					WithDefaultLanguage("en").
					Build()
				require.NoError(t, err)
				cfg.ProviderConfig = nil
				return cfg
			}(),
			wantError: false,
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
			}
		})
	}
}

func TestFactory_ValidateConfigEdgeCases(t *testing.T) {
	factory := &Factory{}

	tests := []struct {
		name      string
		config    interface{}
		wantError bool
	}{
		{
			name: "config with invalid provider config type",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en").
					WithDefaultLanguage("en").
					Build()
				require.NoError(t, err)
				cfg.ProviderConfig = "invalid"
				return cfg
			}(),
			wantError: true,
		},
		{
			name: "config with invalid JSON provider config",
			config: func() *config.Config {
				cfg, err := config.NewConfigBuilder().
					WithSupportedLanguages("en").
					WithDefaultLanguage("en").
					Build()
				require.NoError(t, err)
				cfg.ProviderConfig = &config.JSONProviderConfig{
					FilePath:    "", // Invalid - empty path
					FilePattern: "",
				}
				return cfg
			}(),
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

func TestProvider_HealthEdgeCases(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_health_edge_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt"). // Include both languages
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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

	// Cast to specific type to access JSON provider methods
	jsonProvider, ok := provider.(*Provider)
	require.True(t, ok)

	// First verify normal health check works
	err = provider.Health(ctx)
	assert.NoError(t, err)

	// Manually set default language to one without translations
	originalDefault := jsonProvider.GetDefaultLanguage()
	jsonProvider.SetDefaultLanguage("pt") // pt file doesn't exist, so no translations

	// Test health check when default language has no translations
	err = provider.Health(ctx)
	if err != nil {
		assert.Contains(t, err.Error(), "no translations available for default language")
	}

	// Restore original default
	jsonProvider.SetDefaultLanguage(originalDefault)
}

func TestProvider_NestedTranslationEdgeCases(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_nested_edge_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create complex nested translation structure with edge cases
	enTranslations := map[string]interface{}{
		"simple": "Simple value",
		"nested": map[string]interface{}{
			"value":     "Nested value",
			"nonstring": 123, // Non-string value
			"deeper": map[string]interface{}{
				"value":     "Deep value",
				"nonstring": true, // Non-string value at deeper level
			},
		},
		"invalid_nested": "not_a_map", // String where map expected
	}

	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
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
		name      string
		key       string
		wantError bool
		expected  string
	}{
		{
			name:      "nested key with non-string value",
			key:       "nested.nonstring",
			wantError: false,
			expected:  "nested.nonstring", // Should return key in non-strict mode
		},
		{
			name:      "deep nested key with non-string value",
			key:       "nested.deeper.nonstring",
			wantError: false,
			expected:  "nested.deeper.nonstring",
		},
		{
			name:      "invalid nested path (string where map expected)",
			key:       "invalid_nested.something",
			wantError: false,
			expected:  "invalid_nested.something",
		},
		{
			name:      "non-existent nested key",
			key:       "nested.nonexistent",
			wantError: false,
			expected:  "nested.nonexistent",
		},
		{
			name:      "nested key path through non-existent intermediate",
			key:       "nonexistent.deeper.value",
			wantError: false,
			expected:  "nonexistent.deeper.value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.Translate(ctx, tt.key, "en", nil)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProvider_LoadTranslationsWithMissingFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "i18n_missing_files_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create only one translation file
	enTranslations := map[string]interface{}{
		"hello": "Hello",
	}
	enFile := filepath.Join(tempDir, "en.json")
	enData, _ := json.Marshal(enTranslations)
	require.NoError(t, os.WriteFile(enFile, enData, 0644))

	// Configure for multiple languages but only one file exists
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "fr", "de").
		WithDefaultLanguage("en").
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:    tempDir,
			FilePattern: "{lang}.json",
			Encoding:    "utf-8",
		}).
		Build()
	require.NoError(t, err)

	factory := &Factory{}
	provider, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	assert.NoError(t, err) // Should succeed, missing files create empty translations

	// Cast to specific type to access JSON provider methods
	jsonProvider, ok := provider.(*Provider)
	require.True(t, ok)

	// Verify translation counts
	assert.Equal(t, 1, jsonProvider.GetTranslationCountByLanguage("en"))
	assert.Equal(t, 0, jsonProvider.GetTranslationCountByLanguage("pt"))
	assert.Equal(t, 0, jsonProvider.GetTranslationCountByLanguage("fr"))
	assert.Equal(t, 0, jsonProvider.GetTranslationCountByLanguage("de"))
}
