// Package config provides configuration structures and utilities for the i18n system.
// It defines extensible configuration options that can be customized per provider.
package config

import (
	"fmt"
	"time"
)

// Config represents the main configuration structure for the i18n system.
// It provides common configuration options and can be extended by specific providers.
type Config struct {
	// DefaultLanguage is the fallback language when translation is not available
	DefaultLanguage string `json:"default_language" yaml:"default_language"`

	// SupportedLanguages is the list of languages supported by the provider
	SupportedLanguages []string `json:"supported_languages" yaml:"supported_languages"`

	// LoadTimeout is the maximum time to wait for loading translations
	LoadTimeout time.Duration `json:"load_timeout" yaml:"load_timeout"`

	// CacheEnabled determines if translation caching is enabled
	CacheEnabled bool `json:"cache_enabled" yaml:"cache_enabled"`

	// CacheTTL is the time-to-live for cached translations
	CacheTTL time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// ReloadOnChange determines if translations should be reloaded when source changes
	ReloadOnChange bool `json:"reload_on_change" yaml:"reload_on_change"`

	// FallbackToDefault determines if missing translations should fallback to default language
	FallbackToDefault bool `json:"fallback_to_default" yaml:"fallback_to_default"`

	// StrictMode determines if missing translations should return an error
	StrictMode bool `json:"strict_mode" yaml:"strict_mode"`

	// Provider-specific configuration can be stored here
	ProviderConfig interface{} `json:"provider_config,omitempty" yaml:"provider_config,omitempty"`
}

// JSONProviderConfig represents configuration specific to JSON-based translation providers.
type JSONProviderConfig struct {
	// FilePath is the path to the JSON translation files directory
	FilePath string `json:"file_path" yaml:"file_path"`

	// FilePattern is the pattern for translation files (e.g., "{lang}.json")
	FilePattern string `json:"file_pattern" yaml:"file_pattern"`

	// Encoding is the character encoding of the JSON files
	Encoding string `json:"encoding" yaml:"encoding"`

	// WatchFiles determines if file changes should be monitored
	WatchFiles bool `json:"watch_files" yaml:"watch_files"`

	// ValidateJSON determines if JSON structure should be validated on load
	ValidateJSON bool `json:"validate_json" yaml:"validate_json"`

	// MaxFileSize is the maximum allowed size for translation files
	MaxFileSize int64 `json:"max_file_size" yaml:"max_file_size"`

	// NestedKeys determines if nested key notation is supported (e.g., "user.name")
	NestedKeys bool `json:"nested_keys" yaml:"nested_keys"`
}

// YAMLProviderConfig represents configuration specific to YAML-based translation providers.
type YAMLProviderConfig struct {
	// FilePath is the path to the YAML translation files directory
	FilePath string `json:"file_path" yaml:"file_path"`

	// FilePattern is the pattern for translation files (e.g., "{lang}.yaml")
	FilePattern string `json:"file_pattern" yaml:"file_pattern"`

	// Encoding is the character encoding of the YAML files
	Encoding string `json:"encoding" yaml:"encoding"`

	// WatchFiles determines if file changes should be monitored
	WatchFiles bool `json:"watch_files" yaml:"watch_files"`

	// ValidateYAML determines if YAML structure should be validated on load
	ValidateYAML bool `json:"validate_yaml" yaml:"validate_yaml"`

	// MaxFileSize is the maximum allowed size for translation files
	MaxFileSize int64 `json:"max_file_size" yaml:"max_file_size"`

	// NestedKeys determines if nested key notation is supported (e.g., "user.name")
	NestedKeys bool `json:"nested_keys" yaml:"nested_keys"`

	// AllowDuplicateKeys determines if duplicate keys should be allowed
	AllowDuplicateKeys bool `json:"allow_duplicate_keys" yaml:"allow_duplicate_keys"`
}

// CurrencyConfig represents configuration for currency formatting.
type CurrencyConfig struct {
	// DefaultCurrency is the default currency code
	DefaultCurrency string `json:"default_currency" yaml:"default_currency"`

	// SupportedCurrencies is the list of supported currency codes
	SupportedCurrencies []string `json:"supported_currencies" yaml:"supported_currencies"`

	// ExchangeRateProvider is the provider for currency exchange rates
	ExchangeRateProvider string `json:"exchange_rate_provider" yaml:"exchange_rate_provider"`

	// UpdateInterval is the interval for updating exchange rates
	UpdateInterval time.Duration `json:"update_interval" yaml:"update_interval"`

	// CacheRates determines if exchange rates should be cached
	CacheRates bool `json:"cache_rates" yaml:"cache_rates"`
}

// DateTimeConfig represents configuration for date and time formatting.
type DateTimeConfig struct {
	// DefaultFormat is the default date/time format
	DefaultFormat string `json:"default_format" yaml:"default_format"`

	// DefaultDateFormat is the default date format
	DefaultDateFormat string `json:"default_date_format" yaml:"default_date_format"`

	// DefaultTimeFormat is the default time format
	DefaultTimeFormat string `json:"default_time_format" yaml:"default_time_format"`

	// Timezone is the default timezone
	Timezone string `json:"timezone" yaml:"timezone"`

	// Use24HourFormat determines if 24-hour format should be used by default
	Use24HourFormat bool `json:"use_24_hour_format" yaml:"use_24_hour_format"`
}

// HookConfig represents configuration for hooks.
type HookConfig struct {
	// Name is the unique name of the hook
	Name string `json:"name" yaml:"name"`

	// Enabled determines if the hook is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Priority is the execution priority (lower numbers execute first)
	Priority int `json:"priority" yaml:"priority"`

	// Config is hook-specific configuration
	Config interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// MiddlewareConfig represents configuration for middlewares.
type MiddlewareConfig struct {
	// Name is the unique name of the middleware
	Name string `json:"name" yaml:"name"`

	// Enabled determines if the middleware is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Order is the execution order (lower numbers execute first)
	Order int `json:"order" yaml:"order"`

	// Config is middleware-specific configuration
	Config interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// DefaultConfig returns a default configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en"},
		LoadTimeout:        30 * time.Second,
		CacheEnabled:       true,
		CacheTTL:           1 * time.Hour,
		ReloadOnChange:     false,
		FallbackToDefault:  true,
		StrictMode:         false,
	}
}

// DefaultJSONProviderConfig returns a default JSON provider configuration.
func DefaultJSONProviderConfig() *JSONProviderConfig {
	return &JSONProviderConfig{
		FilePath:     "./translations",
		FilePattern:  "{lang}.json",
		Encoding:     "utf-8",
		WatchFiles:   false,
		ValidateJSON: true,
		MaxFileSize:  10 * 1024 * 1024, // 10MB
		NestedKeys:   true,
	}
}

// DefaultYAMLProviderConfig returns a default YAML provider configuration.
func DefaultYAMLProviderConfig() *YAMLProviderConfig {
	return &YAMLProviderConfig{
		FilePath:           "./translations",
		FilePattern:        "{lang}.yaml",
		Encoding:           "utf-8",
		WatchFiles:         false,
		ValidateYAML:       true,
		MaxFileSize:        10 * 1024 * 1024, // 10MB
		NestedKeys:         true,
		AllowDuplicateKeys: false,
	}
}

// DefaultCurrencyConfig returns a default currency configuration.
func DefaultCurrencyConfig() *CurrencyConfig {
	return &CurrencyConfig{
		DefaultCurrency:      "USD",
		SupportedCurrencies:  []string{"USD", "EUR", "BRL", "GBP", "JPY"},
		ExchangeRateProvider: "mock",
		UpdateInterval:       24 * time.Hour,
		CacheRates:           true,
	}
}

// DefaultDateTimeConfig returns a default date/time configuration.
func DefaultDateTimeConfig() *DateTimeConfig {
	return &DateTimeConfig{
		DefaultFormat:     "2006-01-02 15:04:05",
		DefaultDateFormat: "2006-01-02",
		DefaultTimeFormat: "15:04:05",
		Timezone:          "UTC",
		Use24HourFormat:   true,
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.DefaultLanguage == "" {
		return fmt.Errorf("default_language cannot be empty")
	}

	if len(c.SupportedLanguages) == 0 {
		return fmt.Errorf("supported_languages cannot be empty")
	}

	// Check if default language is in supported languages
	found := false
	for _, lang := range c.SupportedLanguages {
		if lang == c.DefaultLanguage {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default_language '%s' must be included in supported_languages", c.DefaultLanguage)
	}

	if c.LoadTimeout <= 0 {
		return fmt.Errorf("load_timeout must be positive")
	}

	if c.CacheEnabled && c.CacheTTL <= 0 {
		return fmt.Errorf("cache_ttl must be positive when cache is enabled")
	}

	return nil
}

// Validate validates the JSON provider configuration.
func (c *JSONProviderConfig) Validate() error {
	if c.FilePath == "" {
		return fmt.Errorf("file_path cannot be empty")
	}

	if c.FilePattern == "" {
		return fmt.Errorf("file_pattern cannot be empty")
	}

	if c.Encoding == "" {
		c.Encoding = "utf-8"
	}

	if c.MaxFileSize <= 0 {
		c.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	return nil
}

// Validate validates the YAML provider configuration.
func (c *YAMLProviderConfig) Validate() error {
	if c.FilePath == "" {
		return fmt.Errorf("file_path cannot be empty")
	}

	if c.FilePattern == "" {
		return fmt.Errorf("file_pattern cannot be empty")
	}

	if c.Encoding == "" {
		c.Encoding = "utf-8"
	}

	if c.MaxFileSize <= 0 {
		c.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	return nil
}

// Validate validates the currency configuration.
func (c *CurrencyConfig) Validate() error {
	if c.DefaultCurrency == "" {
		return fmt.Errorf("default_currency cannot be empty")
	}

	if len(c.SupportedCurrencies) == 0 {
		return fmt.Errorf("supported_currencies cannot be empty")
	}

	// Check if default currency is in supported currencies
	found := false
	for _, currency := range c.SupportedCurrencies {
		if currency == c.DefaultCurrency {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default_currency '%s' must be included in supported_currencies", c.DefaultCurrency)
	}

	if c.UpdateInterval <= 0 {
		return fmt.Errorf("update_interval must be positive")
	}

	return nil
}

// Validate validates the date/time configuration.
func (c *DateTimeConfig) Validate() error {
	if c.DefaultFormat == "" {
		return fmt.Errorf("default_format cannot be empty")
	}

	if c.DefaultDateFormat == "" {
		return fmt.Errorf("default_date_format cannot be empty")
	}

	if c.DefaultTimeFormat == "" {
		return fmt.Errorf("default_time_format cannot be empty")
	}

	if c.Timezone == "" {
		c.Timezone = "UTC"
	}

	return nil
}

// ConfigBuilder provides a fluent interface for building configurations.
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder with default values.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// WithDefaultLanguage sets the default language.
func (b *ConfigBuilder) WithDefaultLanguage(lang string) *ConfigBuilder {
	b.config.DefaultLanguage = lang
	return b
}

// WithSupportedLanguages sets the supported languages.
func (b *ConfigBuilder) WithSupportedLanguages(langs ...string) *ConfigBuilder {
	b.config.SupportedLanguages = langs
	return b
}

// WithLoadTimeout sets the load timeout.
func (b *ConfigBuilder) WithLoadTimeout(timeout time.Duration) *ConfigBuilder {
	b.config.LoadTimeout = timeout
	return b
}

// WithCache enables or disables caching with the specified TTL.
func (b *ConfigBuilder) WithCache(enabled bool, ttl time.Duration) *ConfigBuilder {
	b.config.CacheEnabled = enabled
	b.config.CacheTTL = ttl
	return b
}

// WithReloadOnChange enables or disables automatic reloading.
func (b *ConfigBuilder) WithReloadOnChange(enabled bool) *ConfigBuilder {
	b.config.ReloadOnChange = enabled
	return b
}

// WithFallbackToDefault enables or disables fallback to default language.
func (b *ConfigBuilder) WithFallbackToDefault(enabled bool) *ConfigBuilder {
	b.config.FallbackToDefault = enabled
	return b
}

// WithStrictMode enables or disables strict mode.
func (b *ConfigBuilder) WithStrictMode(enabled bool) *ConfigBuilder {
	b.config.StrictMode = enabled
	return b
}

// WithProviderConfig sets the provider-specific configuration.
func (b *ConfigBuilder) WithProviderConfig(config interface{}) *ConfigBuilder {
	b.config.ProviderConfig = config
	return b
}

// Build builds and validates the configuration.
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return b.config, nil
}
