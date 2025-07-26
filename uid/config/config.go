// Package config provides configuration structures and validation for UID providers.
// This package centralizes all configuration logic and ensures type safety.
package config

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
)

// ProviderConfig represents the configuration for a UID provider.
type ProviderConfig struct {
	// Type specifies the UID type to generate.
	Type interfaces.UIDType `json:"type"`

	// Name is a human-readable name for the provider instance.
	Name string `json:"name"`

	// ThreadSafe indicates if the provider should be thread-safe.
	ThreadSafe bool `json:"thread_safe"`

	// CacheSize specifies the number of pre-generated UIDs to cache (0 = no cache).
	CacheSize int `json:"cache_size"`

	// ValidationEnabled enables strict validation on all operations.
	ValidationEnabled bool `json:"validation_enabled"`

	// MetricsEnabled enables performance metrics collection.
	MetricsEnabled bool `json:"metrics_enabled"`
}

// ULIDConfig represents specific configuration for ULID providers.
type ULIDConfig struct {
	ProviderConfig

	// MonotonicMode enables monotonic mode for same-timestamp ULIDs.
	MonotonicMode bool `json:"monotonic_mode"`

	// CustomEntropy allows setting a custom entropy source.
	CustomEntropy bool `json:"custom_entropy"`

	// EntropySize specifies the entropy size in bytes (default: 10).
	EntropySize int `json:"entropy_size"`
}

// UUIDConfig represents specific configuration for UUID providers.
type UUIDConfig struct {
	ProviderConfig

	// Version specifies the UUID version (1, 4, 6, 7).
	Version int `json:"version"`

	// NodeID for UUID v1/v6 (if not provided, will be generated).
	NodeID []byte `json:"node_id,omitempty"`

	// ClockSequence for UUID v1/v6 (if not provided, will be generated).
	ClockSequence *int `json:"clock_sequence,omitempty"`

	// Namespace for UUID v3/v5 operations.
	Namespace string `json:"namespace,omitempty"`

	// EnableSorting for UUID v6/v7 to ensure lexicographic ordering.
	EnableSorting bool `json:"enable_sorting"`
}

// FactoryConfig represents configuration for the UID factory.
type FactoryConfig struct {
	// DefaultType specifies the default UID type when none is specified.
	DefaultType interfaces.UIDType `json:"default_type"`

	// EnableCaching enables provider instance caching.
	EnableCaching bool `json:"enable_caching"`

	// CacheTimeout specifies how long to cache provider instances.
	CacheTimeout time.Duration `json:"cache_timeout"`

	// MaxCacheSize limits the number of cached provider instances.
	MaxCacheSize int `json:"max_cache_size"`

	// EnableMetrics enables factory-level metrics collection.
	EnableMetrics bool `json:"enable_metrics"`

	// ProviderConfigs contains specific configurations for each UID type.
	ProviderConfigs map[interfaces.UIDType]interface{} `json:"provider_configs"`
}

// DefaultProviderConfig returns a default configuration for a UID provider.
func DefaultProviderConfig(uidType interfaces.UIDType) *ProviderConfig {
	return &ProviderConfig{
		Type:              uidType,
		Name:              fmt.Sprintf("%s-provider", uidType),
		ThreadSafe:        true,
		CacheSize:         0,
		ValidationEnabled: true,
		MetricsEnabled:    false,
	}
}

// DefaultULIDConfig returns a default configuration for ULID providers.
func DefaultULIDConfig() *ULIDConfig {
	return &ULIDConfig{
		ProviderConfig: *DefaultProviderConfig(interfaces.UIDTypeULID),
		MonotonicMode:  true,
		CustomEntropy:  false,
		EntropySize:    10,
	}
}

// DefaultUUIDV4Config returns a default configuration for UUID v4 providers.
func DefaultUUIDV4Config() *UUIDConfig {
	return &UUIDConfig{
		ProviderConfig: *DefaultProviderConfig(interfaces.UIDTypeUUIDV4),
		Version:        4,
		EnableSorting:  false,
	}
}

// DefaultUUIDV7Config returns a default configuration for UUID v7 providers.
func DefaultUUIDV7Config() *UUIDConfig {
	return &UUIDConfig{
		ProviderConfig: *DefaultProviderConfig(interfaces.UIDTypeUUIDV7),
		Version:        7,
		EnableSorting:  true,
	}
}

// DefaultFactoryConfig returns a default configuration for the UID factory.
func DefaultFactoryConfig() *FactoryConfig {
	return &FactoryConfig{
		DefaultType:     interfaces.UIDTypeULID,
		EnableCaching:   true,
		CacheTimeout:    30 * time.Minute,
		MaxCacheSize:    100,
		EnableMetrics:   false,
		ProviderConfigs: make(map[interfaces.UIDType]interface{}),
	}
}

// Validate validates the provider configuration.
func (c *ProviderConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("UID type cannot be empty")
	}

	if c.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if c.CacheSize < 0 {
		return fmt.Errorf("cache size cannot be negative")
	}

	return nil
}

// Validate validates the ULID configuration.
func (c *ULIDConfig) Validate() error {
	if err := c.ProviderConfig.Validate(); err != nil {
		return err
	}

	if c.Type != interfaces.UIDTypeULID {
		return fmt.Errorf("ULID config must have ULID type")
	}

	if c.EntropySize < 1 || c.EntropySize > 16 {
		return fmt.Errorf("entropy size must be between 1 and 16 bytes")
	}

	return nil
}

// Validate validates the UUID configuration.
func (c *UUIDConfig) Validate() error {
	if err := c.ProviderConfig.Validate(); err != nil {
		return err
	}

	if c.Version < 1 || c.Version > 7 {
		return fmt.Errorf("UUID version must be between 1 and 7")
	}

	if c.NodeID != nil && len(c.NodeID) != 6 {
		return fmt.Errorf("node ID must be 6 bytes")
	}

	if c.ClockSequence != nil && (*c.ClockSequence < 0 || *c.ClockSequence > 16383) {
		return fmt.Errorf("clock sequence must be between 0 and 16383")
	}

	return nil
}

// Validate validates the factory configuration.
func (c *FactoryConfig) Validate() error {
	if c.DefaultType == "" {
		return fmt.Errorf("default UID type cannot be empty")
	}

	if c.CacheTimeout < 0 {
		return fmt.Errorf("cache timeout cannot be negative")
	}

	if c.MaxCacheSize < 0 {
		return fmt.Errorf("max cache size cannot be negative")
	}

	return nil
}
