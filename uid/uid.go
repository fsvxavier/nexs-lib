// Package uid provides a comprehensive UID generation and manipulation library.
// It supports multiple UID types including ULID and UUID variants with a unified interface.
package uid

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/config"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/internal"
	"github.com/fsvxavier/nexs-lib/uid/providers"
)

// Factory implements the Factory interface for creating UID providers.
// It provides a centralized way to create and manage UID providers with caching support.
type Factory struct {
	config     *config.FactoryConfig
	cache      map[interfaces.UIDType]interfaces.Provider
	cacheMutex sync.RWMutex
	name       string
	version    string
}

// NewFactory creates a new UID factory with the specified configuration.
func NewFactory(cfg *config.FactoryConfig) (*Factory, error) {
	if cfg == nil {
		cfg = config.DefaultFactoryConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid factory configuration: %w", err)
	}

	factory := &Factory{
		config:  cfg,
		cache:   make(map[interfaces.UIDType]interfaces.Provider),
		name:    "nexs-uid-factory",
		version: "1.0.0",
	}

	return factory, nil
}

// NewDefaultFactory creates a new UID factory with default configuration.
func NewDefaultFactory() (*Factory, error) {
	return NewFactory(config.DefaultFactoryConfig())
}

// CreateProvider creates a new provider instance for the specified UID type.
func (f *Factory) CreateProvider(ctx context.Context, uidType interfaces.UIDType) (interfaces.Provider, error) {
	if uidType == "" {
		uidType = f.config.DefaultType
	}

	// Check cache first if caching is enabled
	if f.config.EnableCaching {
		f.cacheMutex.RLock()
		if provider, exists := f.cache[uidType]; exists {
			f.cacheMutex.RUnlock()
			return provider, nil
		}
		f.cacheMutex.RUnlock()
	}

	// Create new provider
	provider, err := f.createProviderInstance(ctx, uidType)
	if err != nil {
		return nil, err
	}

	// Cache the provider if caching is enabled
	if f.config.EnableCaching {
		f.cacheMutex.Lock()
		defer f.cacheMutex.Unlock()

		// Check cache size limit
		if len(f.cache) >= f.config.MaxCacheSize {
			// Simple cache eviction: clear all (could be improved with LRU)
			f.cache = make(map[interfaces.UIDType]interfaces.Provider)
		}

		f.cache[uidType] = provider
	}

	return provider, nil
}

// createProviderInstance creates a new provider instance based on the UID type.
func (f *Factory) createProviderInstance(ctx context.Context, uidType interfaces.UIDType) (interfaces.Provider, error) {
	switch uidType {
	case interfaces.UIDTypeULID:
		return f.createULIDProvider()
	case interfaces.UIDTypeUUIDV1:
		return f.createUUIDProvider(1)
	case interfaces.UIDTypeUUIDV4:
		return f.createUUIDProvider(4)
	case interfaces.UIDTypeUUIDV6:
		return f.createUUIDProvider(6)
	case interfaces.UIDTypeUUIDV7:
		return f.createUUIDProvider(7)
	default:
		return nil, fmt.Errorf("unsupported UID type: %s", uidType)
	}
}

// createULIDProvider creates a new ULID provider with configuration.
func (f *Factory) createULIDProvider() (interfaces.Provider, error) {
	var cfg *config.ULIDConfig

	// Check if specific configuration exists
	if providerCfg, exists := f.config.ProviderConfigs[interfaces.UIDTypeULID]; exists {
		if ulidCfg, ok := providerCfg.(*config.ULIDConfig); ok {
			cfg = ulidCfg
		} else {
			return nil, fmt.Errorf("invalid ULID configuration type")
		}
	} else {
		cfg = config.DefaultULIDConfig()
	}

	return providers.NewULIDProvider(cfg)
}

// createUUIDProvider creates a new UUID provider with the specified version.
func (f *Factory) createUUIDProvider(version int) (interfaces.Provider, error) {
	var uidType interfaces.UIDType
	switch version {
	case 1:
		uidType = interfaces.UIDTypeUUIDV1
	case 4:
		uidType = interfaces.UIDTypeUUIDV4
	case 6:
		uidType = interfaces.UIDTypeUUIDV6
	case 7:
		uidType = interfaces.UIDTypeUUIDV7
	default:
		return nil, fmt.Errorf("unsupported UUID version: %d", version)
	}

	var cfg *config.UUIDConfig

	// Check if specific configuration exists
	if providerCfg, exists := f.config.ProviderConfigs[uidType]; exists {
		if uuidCfg, ok := providerCfg.(*config.UUIDConfig); ok {
			cfg = uuidCfg
		} else {
			return nil, fmt.Errorf("invalid UUID configuration type")
		}
	} else {
		// Create default configuration based on version
		switch version {
		case 4:
			cfg = config.DefaultUUIDV4Config()
		case 7:
			cfg = config.DefaultUUIDV7Config()
		default:
			cfg = &config.UUIDConfig{
				ProviderConfig: *config.DefaultProviderConfig(uidType),
				Version:        version,
			}
		}
	}

	return providers.NewUUIDProvider(cfg)
}

// GetSupportedTypes returns all UID types supported by this factory.
func (f *Factory) GetSupportedTypes() []interfaces.UIDType {
	return []interfaces.UIDType{
		interfaces.UIDTypeULID,
		interfaces.UIDTypeUUIDV1,
		interfaces.UIDTypeUUIDV4,
		interfaces.UIDTypeUUIDV6,
		interfaces.UIDTypeUUIDV7,
	}
}

// GetDefaultType returns the default UID type for this factory.
func (f *Factory) GetDefaultType() interfaces.UIDType {
	return f.config.DefaultType
}

// ClearCache clears the provider cache.
func (f *Factory) ClearCache() {
	if !f.config.EnableCaching {
		return
	}

	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()

	f.cache = make(map[interfaces.UIDType]interfaces.Provider)
}

// GetCacheSize returns the current cache size.
func (f *Factory) GetCacheSize() int {
	if !f.config.EnableCaching {
		return 0
	}

	f.cacheMutex.RLock()
	defer f.cacheMutex.RUnlock()

	return len(f.cache)
}

// SetConfiguration updates the factory configuration.
func (f *Factory) SetConfiguration(cfg *config.FactoryConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid factory configuration: %w", err)
	}

	f.config = cfg

	// Clear cache if caching was disabled
	if !cfg.EnableCaching {
		f.ClearCache()
	}

	return nil
}

// GetConfiguration returns a copy of the current factory configuration.
func (f *Factory) GetConfiguration() *config.FactoryConfig {
	return &config.FactoryConfig{
		DefaultType:     f.config.DefaultType,
		EnableCaching:   f.config.EnableCaching,
		CacheTimeout:    f.config.CacheTimeout,
		MaxCacheSize:    f.config.MaxCacheSize,
		EnableMetrics:   f.config.EnableMetrics,
		ProviderConfigs: f.config.ProviderConfigs,
	}
}

// UIDManager provides a high-level interface for UID operations across multiple types.
type UIDManager struct {
	factory *Factory
}

// NewUIDManager creates a new UID manager with the specified factory.
func NewUIDManager(factory *Factory) *UIDManager {
	return &UIDManager{
		factory: factory,
	}
}

// NewDefaultUIDManager creates a new UID manager with default factory configuration.
func NewDefaultUIDManager() (*UIDManager, error) {
	factory, err := NewDefaultFactory()
	if err != nil {
		return nil, err
	}
	return NewUIDManager(factory), nil
}

// Generate generates a new UID of the specified type.
func (m *UIDManager) Generate(ctx context.Context, uidType interfaces.UIDType) (*interfaces.UIDData, error) {
	provider, err := m.factory.CreateProvider(ctx, uidType)
	if err != nil {
		return nil, err
	}
	return provider.Generate(ctx)
}

// GenerateWithTimestamp generates a new UID of the specified type with a custom timestamp.
func (m *UIDManager) GenerateWithTimestamp(ctx context.Context, uidType interfaces.UIDType, timestamp time.Time) (*interfaces.UIDData, error) {
	provider, err := m.factory.CreateProvider(ctx, uidType)
	if err != nil {
		return nil, err
	}
	return provider.GenerateWithTimestamp(ctx, timestamp)
}

// GenerateDefault generates a new UID using the factory's default type.
func (m *UIDManager) GenerateDefault(ctx context.Context) (*interfaces.UIDData, error) {
	return m.Generate(ctx, m.factory.GetDefaultType())
}

// Parse parses a UID string and attempts to determine its type automatically.
func (m *UIDManager) Parse(ctx context.Context, input string) (*interfaces.UIDData, error) {
	// Try to detect the format first
	detector := internal.NewFormatDetector()
	detectedType, err := detector.DetectFormat(input)
	if err != nil {
		return nil, fmt.Errorf("failed to detect UID format: %w", err)
	}

	// Create provider for detected type and parse
	provider, err := m.factory.CreateProvider(ctx, detectedType)
	if err != nil {
		return nil, err
	}

	return provider.Parse(ctx, input)
}

// ParseAs parses a UID string as the specified type.
func (m *UIDManager) ParseAs(ctx context.Context, input string, uidType interfaces.UIDType) (*interfaces.UIDData, error) {
	provider, err := m.factory.CreateProvider(ctx, uidType)
	if err != nil {
		return nil, err
	}
	return provider.Parse(ctx, input)
}

// Validate validates a UID string and attempts to determine its type automatically.
func (m *UIDManager) Validate(ctx context.Context, input string) error {
	// Try to detect the format first
	detector := internal.NewFormatDetector()
	detectedType, err := detector.DetectFormat(input)
	if err != nil {
		return fmt.Errorf("failed to detect UID format: %w", err)
	}

	// Create provider for detected type and validate
	provider, err := m.factory.CreateProvider(ctx, detectedType)
	if err != nil {
		return err
	}

	return provider.Validate(ctx, input)
}

// ValidateAs validates a UID string as the specified type.
func (m *UIDManager) ValidateAs(ctx context.Context, input string, uidType interfaces.UIDType) error {
	provider, err := m.factory.CreateProvider(ctx, uidType)
	if err != nil {
		return err
	}
	return provider.Validate(ctx, input)
}

// Convert converts a UID from one type to another (if supported).
func (m *UIDManager) Convert(ctx context.Context, uid *interfaces.UIDData, targetType interfaces.UIDType) (*interfaces.UIDData, error) {
	provider, err := m.factory.CreateProvider(ctx, uid.Type)
	if err != nil {
		return nil, err
	}
	return provider.ConvertType(ctx, uid, targetType)
}

// GetSupportedTypes returns all supported UID types.
func (m *UIDManager) GetSupportedTypes() []interfaces.UIDType {
	return m.factory.GetSupportedTypes()
}

// GetFactory returns the underlying factory instance.
func (m *UIDManager) GetFactory() *Factory {
	return m.factory
}
