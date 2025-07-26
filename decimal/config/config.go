package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	// Default precision and exponent values
	DefaultMaxPrecision = 21
	DefaultMaxExponent  = 13
	DefaultMinExponent  = -8
	DefaultRounding     = "RoundDown"
	DefaultProvider     = "cockroach"
	DefaultTimeout      = 30 // seconds
)

// Config holds configuration for decimal operations
type Config struct {
	// Precision settings
	MaxPrecision uint32 `json:"max_precision" yaml:"max_precision"`
	MaxExponent  int32  `json:"max_exponent" yaml:"max_exponent"`
	MinExponent  int32  `json:"min_exponent" yaml:"min_exponent"`

	// Rounding mode
	DefaultRounding string `json:"default_rounding" yaml:"default_rounding"`

	// Provider settings
	ProviderName string `json:"provider_name" yaml:"provider_name"`

	// Hook settings
	HooksEnabled bool `json:"hooks_enabled" yaml:"hooks_enabled"`

	// Timeout settings (in seconds)
	Timeout int `json:"timeout" yaml:"timeout"`

	// Additional provider-specific settings
	ProviderConfig map[string]interface{} `json:"provider_config,omitempty" yaml:"provider_config,omitempty"`
}

// NewDefaultConfig creates a new config with default values
func NewDefaultConfig() *Config {
	return &Config{
		MaxPrecision:    DefaultMaxPrecision,
		MaxExponent:     DefaultMaxExponent,
		MinExponent:     DefaultMinExponent,
		DefaultRounding: DefaultRounding,
		ProviderName:    DefaultProvider,
		HooksEnabled:    false,
		Timeout:         DefaultTimeout,
		ProviderConfig:  make(map[string]interface{}),
	}
}

// NewConfig creates a new config with custom values
func NewConfig(opts ...ConfigOption) *Config {
	cfg := NewDefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// ConfigOption defines functional options for Config
type ConfigOption func(*Config)

// WithMaxPrecision sets the maximum precision
func WithMaxPrecision(precision uint32) ConfigOption {
	return func(c *Config) {
		c.MaxPrecision = precision
	}
}

// WithMaxExponent sets the maximum exponent
func WithMaxExponent(exponent int32) ConfigOption {
	return func(c *Config) {
		c.MaxExponent = exponent
	}
}

// WithMinExponent sets the minimum exponent
func WithMinExponent(exponent int32) ConfigOption {
	return func(c *Config) {
		c.MinExponent = exponent
	}
}

// WithRounding sets the default rounding mode
func WithRounding(rounding string) ConfigOption {
	return func(c *Config) {
		c.DefaultRounding = rounding
	}
}

// WithProvider sets the provider name
func WithProvider(provider string) ConfigOption {
	return func(c *Config) {
		c.ProviderName = provider
	}
}

// WithHooksEnabled enables or disables hooks
func WithHooksEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.HooksEnabled = enabled
	}
}

// WithTimeout sets the timeout in seconds
func WithTimeout(timeout int) ConfigOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithProviderConfig sets provider-specific configuration
func WithProviderConfig(key string, value interface{}) ConfigOption {
	return func(c *Config) {
		if c.ProviderConfig == nil {
			c.ProviderConfig = make(map[string]interface{})
		}
		c.ProviderConfig[key] = value
	}
}

// Getters implementing the Config interface

func (c *Config) GetMaxPrecision() uint32 {
	return c.MaxPrecision
}

func (c *Config) GetMaxExponent() int32 {
	return c.MaxExponent
}

func (c *Config) GetMinExponent() int32 {
	return c.MinExponent
}

func (c *Config) GetDefaultRounding() string {
	return c.DefaultRounding
}

func (c *Config) GetProviderName() string {
	return c.ProviderName
}

func (c *Config) IsHooksEnabled() bool {
	return c.HooksEnabled
}

func (c *Config) GetTimeout() int {
	return c.Timeout
}

func (c *Config) GetTimeoutDuration() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxPrecision == 0 {
		return errors.New("max_precision must be greater than 0")
	}

	if c.MaxPrecision > 1000 {
		return errors.New("max_precision cannot exceed 1000")
	}

	if c.MaxExponent <= c.MinExponent {
		return errors.New("max_exponent must be greater than min_exponent")
	}

	validRoundings := map[string]bool{
		"RoundDown":     true,
		"RoundUp":       true,
		"RoundHalfUp":   true,
		"RoundHalfDown": true,
		"RoundHalfEven": true,
		"RoundCeiling":  true,
		"RoundFloor":    true,
		"Round05Up":     true,
	}

	if !validRoundings[c.DefaultRounding] {
		return fmt.Errorf("invalid rounding mode: %s", c.DefaultRounding)
	}

	validProviders := map[string]bool{
		"cockroach":  true,
		"shopspring": true,
	}

	if !validProviders[c.ProviderName] {
		return fmt.Errorf("invalid provider: %s", c.ProviderName)
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.Timeout > 300 {
		return errors.New("timeout cannot exceed 300 seconds")
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c

	// Deep copy the provider config map
	if c.ProviderConfig != nil {
		clone.ProviderConfig = make(map[string]interface{})
		for k, v := range c.ProviderConfig {
			clone.ProviderConfig[k] = v
		}
	}

	return &clone
}

// String returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf("Config{MaxPrecision: %d, MaxExponent: %d, MinExponent: %d, Rounding: %s, Provider: %s, HooksEnabled: %t, Timeout: %d}",
		c.MaxPrecision, c.MaxExponent, c.MinExponent, c.DefaultRounding, c.ProviderName, c.HooksEnabled, c.Timeout)
}
