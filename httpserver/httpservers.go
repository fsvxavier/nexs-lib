// Package httpserver provides a flexible HTTP server abstraction with support for multiple providers.
// It implements Factory, Registry, and Observer patterns to enable extensible HTTP server management.
package httpserver

import (
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Registry manages HTTP server providers and creates server instances.
// It implements the Registry and Factory patterns for extensible server creation.
type Registry struct {
	providers map[string]interfaces.ProviderFactory
	mutex     sync.RWMutex
}

// NewRegistry creates a new server registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]interfaces.ProviderFactory),
		mutex:     sync.RWMutex{},
	}
}

// Register adds a new provider factory to the registry.
func (r *Registry) Register(name string, factory interfaces.ProviderFactory) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("provider factory cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider '%s' already registered", name)
	}

	r.providers[name] = factory
	return nil
}

// Unregister removes a provider factory from the registry.
func (r *Registry) Unregister(name string) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	delete(r.providers, name)
	return nil
}

// Create creates a new HTTP server instance using the specified provider.
func (r *Registry) Create(providerName string, config interface{}) (interfaces.HTTPServer, error) {
	if providerName == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}

	r.mutex.RLock()
	factory, exists := r.providers[providerName]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", providerName)
	}

	return factory.Create(config)
}

// List returns a list of all registered provider names.
func (r *Registry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks if a provider is registered.
func (r *Registry) IsRegistered(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.providers[name]
	return exists
}

// GetProvider returns the factory for the specified provider.
func (r *Registry) GetProvider(name string) (interfaces.ProviderFactory, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return factory, nil
}

// GetProviderCount returns the number of registered providers.
func (r *Registry) GetProviderCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.providers)
}

// Clear removes all registered providers.
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.providers = make(map[string]interfaces.ProviderFactory)
}

// Manager provides high-level HTTP server management functionality.
// It orchestrates the registry, configuration, and observer management.
type Manager struct {
	registry *Registry
	observer *hooks.ObserverManager
}

// NewManager creates a new HTTP server manager.
func NewManager() *Manager {
	return &Manager{
		registry: NewRegistry(),
		observer: hooks.NewObserverManager(),
	}
}

// RegisterProvider registers a new HTTP server provider.
func (m *Manager) RegisterProvider(name string, factory interfaces.ProviderFactory) error {
	return m.registry.Register(name, factory)
}

// UnregisterProvider removes a provider from the registry.
func (m *Manager) UnregisterProvider(name string) error {
	return m.registry.Unregister(name)
}

// CreateServer creates a new HTTP server with the specified provider and configuration.
func (m *Manager) CreateServer(providerName string, options ...config.Option) (interfaces.HTTPServer, error) {
	// Build configuration
	builder := config.NewBuilder()
	if len(options) > 0 {
		builder = builder.Apply(options...)
	}

	cfg, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	// Set provider name in config
	cfg.SetCustom("provider", providerName)

	// Create server
	server, err := m.registry.Create(providerName, cfg)
	if err != nil {
		return nil, fmt.Errorf("server creation error: %w", err)
	}

	// Attach configured observers
	for _, observer := range cfg.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	return server, nil
}

// CreateServerWithConfig creates a new HTTP server with a pre-built configuration.
func (m *Manager) CreateServerWithConfig(providerName string, cfg *config.BaseConfig) (interfaces.HTTPServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set provider name in config
	cfg.SetCustom("provider", providerName)

	// Create server
	server, err := m.registry.Create(providerName, cfg)
	if err != nil {
		return nil, fmt.Errorf("server creation error: %w", err)
	}

	// Attach configured observers
	for _, observer := range cfg.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	return server, nil
}

// ListProviders returns a list of all registered provider names.
func (m *Manager) ListProviders() []string {
	return m.registry.List()
}

// IsProviderRegistered checks if a provider is registered.
func (m *Manager) IsProviderRegistered(name string) bool {
	return m.registry.IsRegistered(name)
}

// GetProvider returns the factory for the specified provider.
func (m *Manager) GetProvider(name string) (interfaces.ProviderFactory, error) {
	return m.registry.GetProvider(name)
}

// GetProviderDefaultConfig returns the default configuration for a provider.
func (m *Manager) GetProviderDefaultConfig(name string) (interface{}, error) {
	factory, err := m.registry.GetProvider(name)
	if err != nil {
		return nil, err
	}
	return factory.GetDefaultConfig(), nil
}

// ValidateProviderConfig validates a configuration for a specific provider.
func (m *Manager) ValidateProviderConfig(name string, config interface{}) error {
	factory, err := m.registry.GetProvider(name)
	if err != nil {
		return err
	}
	return factory.ValidateConfig(config)
}

// AttachGlobalObserver attaches an observer that will be notified of events from all servers.
func (m *Manager) AttachGlobalObserver(observer interfaces.ServerObserver) error {
	return m.observer.AttachObserver(observer)
}

// DetachGlobalObserver removes a global observer.
func (m *Manager) DetachGlobalObserver(observer interfaces.ServerObserver) error {
	return m.observer.DetachObserver(observer)
}

// AttachGlobalHook attaches a hook that will be executed for events from all servers.
func (m *Manager) AttachGlobalHook(eventType interfaces.EventType, hook interfaces.HookFunc) error {
	return m.observer.AttachHook(eventType, hook)
}

// DetachGlobalHook removes all global hooks for a specific event type.
func (m *Manager) DetachGlobalHook(eventType interfaces.EventType) error {
	return m.observer.DetachHook(eventType)
}

// GetObserverManager returns the observer manager for advanced operations.
func (m *Manager) GetObserverManager() *hooks.ObserverManager {
	return m.observer
}

// GetRegistry returns the provider registry for advanced operations.
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// Default global manager instance
var defaultManager = NewManager()

// RegisterProvider registers a provider with the default manager.
func RegisterProvider(name string, factory interfaces.ProviderFactory) error {
	return defaultManager.RegisterProvider(name, factory)
}

// UnregisterProvider removes a provider from the default manager.
func UnregisterProvider(name string) error {
	return defaultManager.UnregisterProvider(name)
}

// CreateServer creates a server using the default manager.
func CreateServer(providerName string, options ...config.Option) (interfaces.HTTPServer, error) {
	return defaultManager.CreateServer(providerName, options...)
}

// CreateServerWithConfig creates a server with pre-built config using the default manager.
func CreateServerWithConfig(providerName string, cfg *config.BaseConfig) (interfaces.HTTPServer, error) {
	return defaultManager.CreateServerWithConfig(providerName, cfg)
}

// ListProviders lists all providers in the default manager.
func ListProviders() []string {
	return defaultManager.ListProviders()
}

// IsProviderRegistered checks if a provider is registered in the default manager.
func IsProviderRegistered(name string) bool {
	return defaultManager.IsProviderRegistered(name)
}

// GetProvider gets a provider from the default manager.
func GetProvider(name string) (interfaces.ProviderFactory, error) {
	return defaultManager.GetProvider(name)
}

// GetDefaultManager returns the default manager instance.
func GetDefaultManager() *Manager {
	return defaultManager
}
