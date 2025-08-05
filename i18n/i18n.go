// Package i18n provides a comprehensive internationalization (i18n) system for Go applications.
// It implements multiple design patterns including Factory, Observer, Hook, Middleware, and Registry
// to provide a flexible and extensible translation system.
//
// The main components are:
// - Registry: Manages provider factories and instances
// - Factory: Creates provider instances with configuration
// - Observer: Notifies hooks and middlewares of events
// - Hook: Executes at specific lifecycle points
// - Middleware: Wraps translation operations with additional functionality
//
// Example usage:
//
//	// Create a registry
//	registry := i18n.NewRegistry()
//
//	// Register providers
//	registry.RegisterProvider(json.NewFactory())
//
//	// Create a provider
//	config := &config.Config{
//		DefaultLanguage: "en",
//		SupportedLanguages: []string{"en", "es", "pt"},
//	}
//	provider, err := registry.CreateProvider("json", config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use the provider
//	result, err := provider.Translate(ctx, "hello.world", "en", nil)
package i18n

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// Registry implements the Registry pattern to manage translation providers,
// hooks, and middlewares. It serves as the central orchestrator for the i18n system.
type Registry struct {
	factories   map[string]interfaces.ProviderFactory
	hooks       []interfaces.Hook
	middlewares []interfaces.Middleware
	instances   map[string]interfaces.I18n
	mu          sync.RWMutex
}

// NewRegistry creates a new registry instance.
func NewRegistry() *Registry {
	return &Registry{
		factories:   make(map[string]interfaces.ProviderFactory),
		hooks:       make([]interfaces.Hook, 0),
		middlewares: make([]interfaces.Middleware, 0),
		instances:   make(map[string]interfaces.I18n),
	}
}

// RegisterProvider registers a new provider factory with the registry.
// The factory will be used to create instances of the provider type.
func (r *Registry) RegisterProvider(factory interfaces.ProviderFactory) error {
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	name := factory.Name()
	if name == "" {
		return fmt.Errorf("factory name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider factory '%s' is already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// CreateProvider creates a new provider instance using the specified factory and configuration.
// The created provider will have all registered hooks and middlewares applied.
func (r *Registry) CreateProvider(name string, config interface{}) (interfaces.I18n, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	if !exists {
		r.mu.RUnlock()
		return nil, fmt.Errorf("provider factory '%s' not found", name)
	}

	// Copy hooks and middlewares to avoid holding the lock during creation
	hooks := make([]interfaces.Hook, len(r.hooks))
	copy(hooks, r.hooks)
	middlewares := make([]interfaces.Middleware, len(r.middlewares))
	copy(middlewares, r.middlewares)
	r.mu.RUnlock()

	// Validate configuration
	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration for provider '%s': %w", name, err)
	}

	// Create the provider instance
	provider, err := factory.Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider '%s': %w", name, err)
	}

	// Wrap the provider with hooks and middlewares
	wrappedProvider := r.wrapProvider(provider, name, hooks, middlewares)

	// Store the instance for management
	r.mu.Lock()
	instanceKey := fmt.Sprintf("%s-%p", name, wrappedProvider)
	r.instances[instanceKey] = wrappedProvider
	r.mu.Unlock()

	return wrappedProvider, nil
}

// GetProviderNames returns a list of all registered provider factory names.
func (r *Registry) GetProviderNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// HasProvider checks if a provider factory with the specified name is registered.
func (r *Registry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[name]
	return exists
}

// AddHook adds a hook to be executed for all providers created by this registry.
// Hooks are executed in priority order (lower numbers first).
func (r *Registry) AddHook(hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	name := hook.Name()
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate hook names
	for _, existingHook := range r.hooks {
		if existingHook.Name() == name {
			return fmt.Errorf("hook '%s' is already registered", name)
		}
	}

	r.hooks = append(r.hooks, hook)

	// Sort hooks by priority
	sort.Slice(r.hooks, func(i, j int) bool {
		return r.hooks[i].Priority() < r.hooks[j].Priority()
	})

	return nil
}

// RemoveHook removes a hook by name from the registry.
func (r *Registry) RemoveHook(name string) error {
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, hook := range r.hooks {
		if hook.Name() == name {
			// Remove the hook from the slice
			r.hooks = append(r.hooks[:i], r.hooks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("hook '%s' not found", name)
}

// AddMiddleware adds a middleware to be applied to all providers created by this registry.
// Middlewares are applied in the order they are added.
func (r *Registry) AddMiddleware(middleware interfaces.Middleware) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	name := middleware.Name()
	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate middleware names
	for _, existingMiddleware := range r.middlewares {
		if existingMiddleware.Name() == name {
			return fmt.Errorf("middleware '%s' is already registered", name)
		}
	}

	r.middlewares = append(r.middlewares, middleware)
	return nil
}

// RemoveMiddleware removes a middleware by name from the registry.
func (r *Registry) RemoveMiddleware(name string) error {
	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, middleware := range r.middlewares {
		if middleware.Name() == name {
			// Remove the middleware from the slice
			r.middlewares = append(r.middlewares[:i], r.middlewares[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("middleware '%s' not found", name)
}

// GetHooks returns a copy of all registered hooks.
func (r *Registry) GetHooks() []interfaces.Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hooks := make([]interfaces.Hook, len(r.hooks))
	copy(hooks, r.hooks)
	return hooks
}

// GetMiddlewares returns a copy of all registered middlewares.
func (r *Registry) GetMiddlewares() []interfaces.Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	middlewares := make([]interfaces.Middleware, len(r.middlewares))
	copy(middlewares, r.middlewares)
	return middlewares
}

// GetActiveInstances returns the number of active provider instances.
func (r *Registry) GetActiveInstances() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.instances)
}

// Shutdown gracefully shuts down all active provider instances.
func (r *Registry) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	instances := make([]interfaces.I18n, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}
	r.instances = make(map[string]interfaces.I18n) // Clear instances
	r.mu.Unlock()

	var errors []error
	for _, instance := range instances {
		if err := instance.Stop(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}

// wrapProvider wraps a provider with hooks and middlewares.
func (r *Registry) wrapProvider(provider interfaces.I18n, providerName string, hooks []interfaces.Hook, middlewares []interfaces.Middleware) interfaces.I18n {
	return &wrappedProvider{
		provider:     provider,
		providerName: providerName,
		hooks:        hooks,
		middlewares:  middlewares,
	}
}

// wrappedProvider is a wrapper that applies hooks and middlewares to a provider.
type wrappedProvider struct {
	provider     interfaces.I18n
	providerName string
	hooks        []interfaces.Hook
	middlewares  []interfaces.Middleware
}

// Translate performs translation with middleware and hook support.
func (wp *wrappedProvider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	// Create the base translation function
	translateFunc := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		return wp.provider.Translate(ctx, key, lang, params)
	}

	// Apply middlewares in reverse order (last registered middleware is outermost)
	for i := len(wp.middlewares) - 1; i >= 0; i-- {
		translateFunc = wp.middlewares[i].WrapTranslate(translateFunc)
	}

	// Execute the wrapped translation
	result, err := translateFunc(ctx, key, lang, params)

	// Notify hooks about the translation
	if err == nil {
		for _, hook := range wp.hooks {
			if hookErr := hook.OnTranslate(ctx, wp.providerName, key, lang, result); hookErr != nil {
				// Log hook error but don't fail the translation
				// In a real implementation, you might want to use a logger here
			}
		}
	}

	return result, err
}

// LoadTranslations loads translation data from the configured source.
func (wp *wrappedProvider) LoadTranslations(ctx context.Context) error {
	return wp.provider.LoadTranslations(ctx)
}

// GetSupportedLanguages returns a list of supported language codes.
func (wp *wrappedProvider) GetSupportedLanguages() []string {
	return wp.provider.GetSupportedLanguages()
}

// HasTranslation checks if a translation exists for the given key and language.
func (wp *wrappedProvider) HasTranslation(key string, lang string) bool {
	return wp.provider.HasTranslation(key, lang)
}

// GetDefaultLanguage returns the default language code.
func (wp *wrappedProvider) GetDefaultLanguage() string {
	return wp.provider.GetDefaultLanguage()
}

// SetDefaultLanguage sets the default language code.
func (wp *wrappedProvider) SetDefaultLanguage(lang string) {
	wp.provider.SetDefaultLanguage(lang)
}

// Start initializes the translation provider and notifies hooks.
func (wp *wrappedProvider) Start(ctx context.Context) error {
	// Notify middlewares about start
	for _, middleware := range wp.middlewares {
		if err := middleware.OnStart(ctx, wp.providerName); err != nil {
			return fmt.Errorf("middleware '%s' start failed: %w", middleware.Name(), err)
		}
	}

	// Start the actual provider
	if err := wp.provider.Start(ctx); err != nil {
		// Notify hooks about error
		for _, hook := range wp.hooks {
			hook.OnError(ctx, wp.providerName, err)
		}

		// Notify middlewares about error
		for _, middleware := range wp.middlewares {
			middleware.OnError(ctx, wp.providerName, err)
		}

		return err
	}

	// Notify hooks about successful start
	for _, hook := range wp.hooks {
		if err := hook.OnStart(ctx, wp.providerName); err != nil {
			// Log hook error but don't fail the start
			// In a real implementation, you might want to use a logger here
		}
	}

	return nil
}

// Stop gracefully shuts down the translation provider and notifies hooks.
func (wp *wrappedProvider) Stop(ctx context.Context) error {
	// Stop the actual provider
	err := wp.provider.Stop(ctx)

	// Notify hooks about stop (regardless of error)
	for _, hook := range wp.hooks {
		if hookErr := hook.OnStop(ctx, wp.providerName); hookErr != nil {
			// Log hook error but don't fail the stop
			// In a real implementation, you might want to use a logger here
		}
	}

	// Notify middlewares about stop (regardless of error)
	for _, middleware := range wp.middlewares {
		if mwErr := middleware.OnStop(ctx, wp.providerName); mwErr != nil {
			// Log middleware error but don't fail the stop
			// In a real implementation, you might want to use a logger here
		}
	}

	if err != nil {
		// Notify hooks about error
		for _, hook := range wp.hooks {
			hook.OnError(ctx, wp.providerName, err)
		}

		// Notify middlewares about error
		for _, middleware := range wp.middlewares {
			middleware.OnError(ctx, wp.providerName, err)
		}
	}

	return err
}

// Health returns the health status of the translation provider.
func (wp *wrappedProvider) Health(ctx context.Context) error {
	return wp.provider.Health(ctx)
}

// DefaultRegistry is a global registry instance for convenience.
var DefaultRegistry = NewRegistry()

// RegisterProvider registers a provider factory with the default registry.
func RegisterProvider(factory interfaces.ProviderFactory) error {
	return DefaultRegistry.RegisterProvider(factory)
}

// CreateProvider creates a provider instance using the default registry.
func CreateProvider(name string, config interface{}) (interfaces.I18n, error) {
	return DefaultRegistry.CreateProvider(name, config)
}

// AddHook adds a hook to the default registry.
func AddHook(hook interfaces.Hook) error {
	return DefaultRegistry.AddHook(hook)
}

// AddMiddleware adds a middleware to the default registry.
func AddMiddleware(middleware interfaces.Middleware) error {
	return DefaultRegistry.AddMiddleware(middleware)
}

// GetProviderNames returns provider names from the default registry.
func GetProviderNames() []string {
	return DefaultRegistry.GetProviderNames()
}

// HasProvider checks if a provider exists in the default registry.
func HasProvider(name string) bool {
	return DefaultRegistry.HasProvider(name)
}

// Shutdown shuts down the default registry.
func Shutdown(ctx context.Context) error {
	return DefaultRegistry.Shutdown(ctx)
}
