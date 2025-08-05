package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DefaultLanguage != "en" {
		t.Errorf("expected default language 'en', got '%s'", config.DefaultLanguage)
	}

	if len(config.SupportedLanguages) != 1 || config.SupportedLanguages[0] != "en" {
		t.Errorf("expected supported languages ['en'], got %v", config.SupportedLanguages)
	}

	if config.LoadTimeout != 30*time.Second {
		t.Errorf("expected load timeout 30s, got %v", config.LoadTimeout)
	}

	if !config.CacheEnabled {
		t.Error("expected cache to be enabled by default")
	}

	if config.CacheTTL != 1*time.Hour {
		t.Errorf("expected cache TTL 1h, got %v", config.CacheTTL)
	}

	if config.ReloadOnChange {
		t.Error("expected reload on change to be disabled by default")
	}

	if !config.FallbackToDefault {
		t.Error("expected fallback to default to be enabled by default")
	}

	if config.StrictMode {
		t.Error("expected strict mode to be disabled by default")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "empty default language",
			config: &Config{
				DefaultLanguage:    "",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "default_language cannot be empty",
		},
		{
			name: "empty supported languages",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "supported_languages cannot be empty",
		},
		{
			name: "default language not in supported languages",
			config: &Config{
				DefaultLanguage:    "fr",
				SupportedLanguages: []string{"en", "es"},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "default_language 'fr' must be included in supported_languages",
		},
		{
			name: "zero load timeout",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        0,
				CacheEnabled:       true,
				CacheTTL:           1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "load_timeout must be positive",
		},
		{
			name: "negative load timeout",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        -1 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "load_timeout must be positive",
		},
		{
			name: "cache enabled with zero TTL",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           0,
			},
			expectError: true,
			errorMsg:    "cache_ttl must be positive when cache is enabled",
		},
		{
			name: "cache enabled with negative TTL",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       true,
				CacheTTL:           -1 * time.Hour,
			},
			expectError: true,
			errorMsg:    "cache_ttl must be positive when cache is enabled",
		},
		{
			name: "cache disabled with zero TTL",
			config: &Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				LoadTimeout:        30 * time.Second,
				CacheEnabled:       false,
				CacheTTL:           0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestDefaultJSONProviderConfig(t *testing.T) {
	config := DefaultJSONProviderConfig()

	if config.FilePath != "./translations" {
		t.Errorf("expected file path './translations', got '%s'", config.FilePath)
	}

	if config.FilePattern != "{lang}.json" {
		t.Errorf("expected file pattern '{lang}.json', got '%s'", config.FilePattern)
	}

	if config.Encoding != "utf-8" {
		t.Errorf("expected encoding 'utf-8', got '%s'", config.Encoding)
	}

	if config.WatchFiles {
		t.Error("expected watch files to be disabled by default")
	}

	if !config.ValidateJSON {
		t.Error("expected JSON validation to be enabled by default")
	}

	if config.MaxFileSize != 10*1024*1024 {
		t.Errorf("expected max file size 10MB, got %d", config.MaxFileSize)
	}

	if !config.NestedKeys {
		t.Error("expected nested keys to be enabled by default")
	}
}

func TestJSONProviderConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *JSONProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultJSONProviderConfig(),
			expectError: false,
		},
		{
			name: "empty file path",
			config: &JSONProviderConfig{
				FilePath:    "",
				FilePattern: "{lang}.json",
			},
			expectError: true,
			errorMsg:    "file_path cannot be empty",
		},
		{
			name: "empty file pattern",
			config: &JSONProviderConfig{
				FilePath:    "./translations",
				FilePattern: "",
			},
			expectError: true,
			errorMsg:    "file_pattern cannot be empty",
		},
		{
			name: "empty encoding gets default",
			config: &JSONProviderConfig{
				FilePath:    "./translations",
				FilePattern: "{lang}.json",
				Encoding:    "",
			},
			expectError: false,
		},
		{
			name: "zero max file size gets default",
			config: &JSONProviderConfig{
				FilePath:    "./translations",
				FilePattern: "{lang}.json",
				MaxFileSize: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestDefaultYAMLProviderConfig(t *testing.T) {
	config := DefaultYAMLProviderConfig()

	if config.FilePath != "./translations" {
		t.Errorf("expected file path './translations', got '%s'", config.FilePath)
	}

	if config.FilePattern != "{lang}.yaml" {
		t.Errorf("expected file pattern '{lang}.yaml', got '%s'", config.FilePattern)
	}

	if config.Encoding != "utf-8" {
		t.Errorf("expected encoding 'utf-8', got '%s'", config.Encoding)
	}

	if config.WatchFiles {
		t.Error("expected watch files to be disabled by default")
	}

	if !config.ValidateYAML {
		t.Error("expected YAML validation to be enabled by default")
	}

	if config.MaxFileSize != 10*1024*1024 {
		t.Errorf("expected max file size 10MB, got %d", config.MaxFileSize)
	}

	if !config.NestedKeys {
		t.Error("expected nested keys to be enabled by default")
	}

	if config.AllowDuplicateKeys {
		t.Error("expected duplicate keys to be disabled by default")
	}
}

func TestYAMLProviderConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *YAMLProviderConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultYAMLProviderConfig(),
			expectError: false,
		},
		{
			name: "empty file path",
			config: &YAMLProviderConfig{
				FilePath:    "",
				FilePattern: "{lang}.yaml",
			},
			expectError: true,
			errorMsg:    "file_path cannot be empty",
		},
		{
			name: "empty file pattern",
			config: &YAMLProviderConfig{
				FilePath:    "./translations",
				FilePattern: "",
			},
			expectError: true,
			errorMsg:    "file_pattern cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()

	config, err := builder.
		WithDefaultLanguage("pt").
		WithSupportedLanguages("pt", "en", "es").
		WithLoadTimeout(45*time.Second).
		WithCache(true, 2*time.Hour).
		WithReloadOnChange(true).
		WithFallbackToDefault(false).
		WithStrictMode(true).
		WithProviderConfig(DefaultJSONProviderConfig()).
		Build()

	if err != nil {
		t.Errorf("expected no error building config, got: %v", err)
	}

	if config.DefaultLanguage != "pt" {
		t.Errorf("expected default language 'pt', got '%s'", config.DefaultLanguage)
	}

	expectedLangs := []string{"pt", "en", "es"}
	if len(config.SupportedLanguages) != len(expectedLangs) {
		t.Errorf("expected %d supported languages, got %d", len(expectedLangs), len(config.SupportedLanguages))
	}

	for i, lang := range expectedLangs {
		if config.SupportedLanguages[i] != lang {
			t.Errorf("expected supported language '%s' at index %d, got '%s'", lang, i, config.SupportedLanguages[i])
		}
	}

	if config.LoadTimeout != 45*time.Second {
		t.Errorf("expected load timeout 45s, got %v", config.LoadTimeout)
	}

	if !config.CacheEnabled {
		t.Error("expected cache to be enabled")
	}

	if config.CacheTTL != 2*time.Hour {
		t.Errorf("expected cache TTL 2h, got %v", config.CacheTTL)
	}

	if !config.ReloadOnChange {
		t.Error("expected reload on change to be enabled")
	}

	if config.FallbackToDefault {
		t.Error("expected fallback to default to be disabled")
	}

	if !config.StrictMode {
		t.Error("expected strict mode to be enabled")
	}

	if config.ProviderConfig == nil {
		t.Error("expected provider config to be set")
	}
}

func TestConfigBuilderInvalidConfig(t *testing.T) {
	builder := NewConfigBuilder()

	_, err := builder.
		WithDefaultLanguage("").
		Build()

	if err == nil {
		t.Error("expected error for invalid config, but got none")
	}
}

func TestDefaultCurrencyConfig(t *testing.T) {
	config := DefaultCurrencyConfig()

	if config.DefaultCurrency != "USD" {
		t.Errorf("expected default currency 'USD', got '%s'", config.DefaultCurrency)
	}

	expectedCurrencies := []string{"USD", "EUR", "BRL", "GBP", "JPY"}
	if len(config.SupportedCurrencies) != len(expectedCurrencies) {
		t.Errorf("expected %d supported currencies, got %d", len(expectedCurrencies), len(config.SupportedCurrencies))
	}

	if config.ExchangeRateProvider != "mock" {
		t.Errorf("expected exchange rate provider 'mock', got '%s'", config.ExchangeRateProvider)
	}

	if config.UpdateInterval != 24*time.Hour {
		t.Errorf("expected update interval 24h, got %v", config.UpdateInterval)
	}

	if !config.CacheRates {
		t.Error("expected cache rates to be enabled by default")
	}
}

func TestCurrencyConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *CurrencyConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultCurrencyConfig(),
			expectError: false,
		},
		{
			name: "empty default currency",
			config: &CurrencyConfig{
				DefaultCurrency:     "",
				SupportedCurrencies: []string{"USD"},
				UpdateInterval:      24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "default_currency cannot be empty",
		},
		{
			name: "empty supported currencies",
			config: &CurrencyConfig{
				DefaultCurrency:     "USD",
				SupportedCurrencies: []string{},
				UpdateInterval:      24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "supported_currencies cannot be empty",
		},
		{
			name: "default currency not in supported",
			config: &CurrencyConfig{
				DefaultCurrency:     "CHF",
				SupportedCurrencies: []string{"USD", "EUR"},
				UpdateInterval:      24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "default_currency 'CHF' must be included in supported_currencies",
		},
		{
			name: "zero update interval",
			config: &CurrencyConfig{
				DefaultCurrency:     "USD",
				SupportedCurrencies: []string{"USD"},
				UpdateInterval:      0,
			},
			expectError: true,
			errorMsg:    "update_interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestDefaultDateTimeConfig(t *testing.T) {
	config := DefaultDateTimeConfig()

	if config.DefaultFormat != "2006-01-02 15:04:05" {
		t.Errorf("expected default format '2006-01-02 15:04:05', got '%s'", config.DefaultFormat)
	}

	if config.DefaultDateFormat != "2006-01-02" {
		t.Errorf("expected default date format '2006-01-02', got '%s'", config.DefaultDateFormat)
	}

	if config.DefaultTimeFormat != "15:04:05" {
		t.Errorf("expected default time format '15:04:05', got '%s'", config.DefaultTimeFormat)
	}

	if config.Timezone != "UTC" {
		t.Errorf("expected timezone 'UTC', got '%s'", config.Timezone)
	}

	if !config.Use24HourFormat {
		t.Error("expected 24-hour format to be enabled by default")
	}
}

func TestDateTimeConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *DateTimeConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultDateTimeConfig(),
			expectError: false,
		},
		{
			name: "empty default format",
			config: &DateTimeConfig{
				DefaultFormat:     "",
				DefaultDateFormat: "2006-01-02",
				DefaultTimeFormat: "15:04:05",
			},
			expectError: true,
			errorMsg:    "default_format cannot be empty",
		},
		{
			name: "empty default date format",
			config: &DateTimeConfig{
				DefaultFormat:     "2006-01-02 15:04:05",
				DefaultDateFormat: "",
				DefaultTimeFormat: "15:04:05",
			},
			expectError: true,
			errorMsg:    "default_date_format cannot be empty",
		},
		{
			name: "empty default time format",
			config: &DateTimeConfig{
				DefaultFormat:     "2006-01-02 15:04:05",
				DefaultDateFormat: "2006-01-02",
				DefaultTimeFormat: "",
			},
			expectError: true,
			errorMsg:    "default_time_format cannot be empty",
		},
		{
			name: "empty timezone gets default",
			config: &DateTimeConfig{
				DefaultFormat:     "2006-01-02 15:04:05",
				DefaultDateFormat: "2006-01-02",
				DefaultTimeFormat: "15:04:05",
				Timezone:          "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				// Check if empty timezone was set to default
				if tt.config.Timezone == "" {
					if tt.config.Timezone != "UTC" {
						t.Error("expected empty timezone to be set to 'UTC'")
					}
				}
			}
		})
	}
}
