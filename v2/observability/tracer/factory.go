package tracer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ProviderType represents the type of tracing provider
type ProviderType string

const (
	// ProviderTypeDatadog represents Datadog APM provider
	ProviderTypeDatadog ProviderType = "datadog"
	// ProviderTypeNewRelic represents New Relic APM provider
	ProviderTypeNewRelic ProviderType = "newrelic"
	// ProviderTypePrometheus represents Prometheus metrics provider
	ProviderTypePrometheus ProviderType = "prometheus"
	// ProviderTypeNoop represents no-operation provider
	ProviderTypeNoop ProviderType = "noop"
)

// String returns string representation of provider type
func (p ProviderType) String() string {
	return string(p)
}

// Factory manages tracer provider creation and lifecycle
type Factory struct {
	providers    map[ProviderType]Provider
	constructors map[ProviderType]ProviderConstructor
	mu           sync.RWMutex
}

// ProviderConstructor creates a new provider instance
type ProviderConstructor func(config interface{}) (Provider, error)

// NewFactory creates a new tracer factory
func NewFactory() *Factory {
	return &Factory{
		providers:    make(map[ProviderType]Provider),
		constructors: make(map[ProviderType]ProviderConstructor),
	}
}

// RegisterProvider registers a provider constructor
func (f *Factory) RegisterProvider(providerType ProviderType, constructor ProviderConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.constructors[providerType] = constructor
}

// CreateProvider creates a new provider instance
func (f *Factory) CreateProvider(providerType ProviderType, config interface{}) (Provider, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if provider already exists
	if provider, exists := f.providers[providerType]; exists {
		return provider, nil
	}

	// Check if constructor is registered
	constructor, exists := f.constructors[providerType]
	if !exists {
		return nil, fmt.Errorf("provider type %s is not registered", providerType)
	}

	// Create provider
	provider, err := constructor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	// Cache provider
	f.providers[providerType] = provider
	return provider, nil
}

// GetProvider returns an existing provider
func (f *Factory) GetProvider(providerType ProviderType) (Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider type %s not found", providerType)
	}

	return provider, nil
}

// ListProviders returns all registered provider types
func (f *Factory) ListProviders() []ProviderType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]ProviderType, 0, len(f.constructors))
	for providerType := range f.constructors {
		types = append(types, providerType)
	}
	return types
}

// GetActiveProviders returns all active provider instances
func (f *Factory) GetActiveProviders() map[ProviderType]Provider {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[ProviderType]Provider)
	for providerType, provider := range f.providers {
		result[providerType] = provider
	}
	return result
}

// Shutdown gracefully shuts down all providers
func (f *Factory) Shutdown(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errors []string
	for providerType, provider := range f.providers {
		if err := provider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", providerType, err))
		}
	}

	// Clear providers
	f.providers = make(map[ProviderType]Provider)

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// HealthCheck performs health check on all active providers
func (f *Factory) HealthCheck(ctx context.Context) map[ProviderType]error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	results := make(map[ProviderType]error)
	for providerType, provider := range f.providers {
		results[providerType] = provider.HealthCheck(ctx)
	}
	return results
}

// GetMetrics returns metrics from all active providers
func (f *Factory) GetMetrics() map[ProviderType]ProviderMetrics {
	f.mu.RLock()
	defer f.mu.RUnlock()

	results := make(map[ProviderType]ProviderMetrics)
	for providerType, provider := range f.providers {
		results[providerType] = provider.GetProviderMetrics()
	}
	return results
}

// TracerManager manages multiple tracers across providers
type TracerManager struct {
	factory  *Factory
	tracers  map[string]Tracer
	provider Provider
	mu       sync.RWMutex
}

// NewTracerManager creates a new tracer manager
func NewTracerManager(factory *Factory, providerType ProviderType, config interface{}) (*TracerManager, error) {
	provider, err := factory.CreateProvider(providerType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &TracerManager{
		factory:  factory,
		tracers:  make(map[string]Tracer),
		provider: provider,
	}, nil
}

// GetTracer returns a tracer by name, creating it if necessary
func (tm *TracerManager) GetTracer(name string, options ...TracerOption) (Tracer, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tracer, exists := tm.tracers[name]; exists {
		return tracer, nil
	}

	tracer, err := tm.provider.CreateTracer(name, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer %s: %w", name, err)
	}

	tm.tracers[name] = tracer
	return tracer, nil
}

// ListTracers returns all active tracer names
func (tm *TracerManager) ListTracers() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	names := make([]string, 0, len(tm.tracers))
	for name := range tm.tracers {
		names = append(names, name)
	}
	return names
}

// GetAllMetrics returns metrics from all tracers
func (tm *TracerManager) GetAllMetrics() map[string]TracerMetrics {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	results := make(map[string]TracerMetrics)
	for name, tracer := range tm.tracers {
		results[name] = tracer.GetMetrics()
	}
	return results
}

// Shutdown gracefully shuts down all tracers and the provider
func (tm *TracerManager) Shutdown(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Close all tracers
	var errors []string
	for name, tracer := range tm.tracers {
		if err := tracer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("tracer %s: %v", name, err))
		}
	}

	// Shutdown provider
	if err := tm.provider.Shutdown(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("provider: %v", err))
	}

	// Clear tracers
	tm.tracers = make(map[string]Tracer)

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetProvider returns the underlying provider
func (tm *TracerManager) GetProvider() Provider {
	return tm.provider
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Type        ProviderType    `json:"type"`
	Name        string          `json:"name"`
	IsActive    bool            `json:"is_active"`
	Metrics     ProviderMetrics `json:"metrics,omitempty"`
	HealthCheck error           `json:"health_check,omitempty"`
}

// GetProviderInfo returns detailed information about providers
func (f *Factory) GetProviderInfo(ctx context.Context) []ProviderInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var infos []ProviderInfo

	// Add active providers
	for providerType, provider := range f.providers {
		info := ProviderInfo{
			Type:        providerType,
			Name:        provider.Name(),
			IsActive:    true,
			Metrics:     provider.GetProviderMetrics(),
			HealthCheck: provider.HealthCheck(ctx),
		}
		infos = append(infos, info)
	}

	// Add registered but inactive providers
	for providerType := range f.constructors {
		if _, exists := f.providers[providerType]; !exists {
			info := ProviderInfo{
				Type:     providerType,
				Name:     string(providerType),
				IsActive: false,
			}
			infos = append(infos, info)
		}
	}

	return infos
}
