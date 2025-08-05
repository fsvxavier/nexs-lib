package i18n

import (
	"context"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// MockProviderFactory is a mock implementation of ProviderFactory for testing.
type MockProviderFactory struct {
	name         string
	createFunc   func(config interface{}) (interfaces.I18n, error)
	validateFunc func(config interface{}) error
}

func NewMockProviderFactory(name string) *MockProviderFactory {
	return &MockProviderFactory{
		name: name,
		createFunc: func(config interface{}) (interfaces.I18n, error) {
			return &MockProvider{name: name}, nil
		},
		validateFunc: func(config interface{}) error {
			return nil
		},
	}
}

func (f *MockProviderFactory) Create(config interface{}) (interfaces.I18n, error) {
	return f.createFunc(config)
}

func (f *MockProviderFactory) Name() string {
	return f.name
}

func (f *MockProviderFactory) ValidateConfig(config interface{}) error {
	return f.validateFunc(config)
}

// MockProvider is a mock implementation of I18n for testing.
type MockProvider struct {
	name                   string
	supportedLanguages     []string
	defaultLanguage        string
	translations           map[string]map[string]string
	started                bool
	loadTranslationsCalled bool
}

func (p *MockProvider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	if !p.started {
		return "", fmt.Errorf("provider not started")
	}

	if p.translations == nil {
		p.translations = map[string]map[string]string{
			"en": {
				"hello": "Hello",
				"world": "World",
			},
			"es": {
				"hello": "Hola",
				"world": "Mundo",
			},
		}
	}

	if langTrans, exists := p.translations[lang]; exists {
		if translation, exists := langTrans[key]; exists {
			return translation, nil
		}
	}

	return "", fmt.Errorf("translation not found for key '%s' in language '%s'", key, lang)
}

func (p *MockProvider) LoadTranslations(ctx context.Context) error {
	p.loadTranslationsCalled = true
	return nil
}

func (p *MockProvider) GetSupportedLanguages() []string {
	if p.supportedLanguages == nil {
		return []string{"en", "es"}
	}
	return p.supportedLanguages
}

func (p *MockProvider) HasTranslation(key string, lang string) bool {
	if p.translations == nil {
		return false
	}
	if langTrans, exists := p.translations[lang]; exists {
		_, exists := langTrans[key]
		return exists
	}
	return false
}

func (p *MockProvider) GetDefaultLanguage() string {
	if p.defaultLanguage == "" {
		return "en"
	}
	return p.defaultLanguage
}

func (p *MockProvider) SetDefaultLanguage(lang string) {
	p.defaultLanguage = lang
}

func (p *MockProvider) Start(ctx context.Context) error {
	p.started = true
	return nil
}

func (p *MockProvider) Stop(ctx context.Context) error {
	p.started = false
	return nil
}

func (p *MockProvider) Health(ctx context.Context) error {
	if !p.started {
		return fmt.Errorf("provider not started")
	}
	return nil
}

// MockHook is a mock implementation of Hook for testing.
type MockHook struct {
	name     string
	priority int
	calls    []string
}

func NewMockHook(name string, priority int) *MockHook {
	return &MockHook{
		name:     name,
		priority: priority,
		calls:    make([]string, 0),
	}
}

func (h *MockHook) Name() string {
	return h.name
}

func (h *MockHook) Priority() int {
	return h.priority
}

func (h *MockHook) OnStart(ctx context.Context, providerName string) error {
	h.calls = append(h.calls, fmt.Sprintf("OnStart:%s", providerName))
	return nil
}

func (h *MockHook) OnStop(ctx context.Context, providerName string) error {
	h.calls = append(h.calls, fmt.Sprintf("OnStop:%s", providerName))
	return nil
}

func (h *MockHook) OnError(ctx context.Context, providerName string, err error) error {
	h.calls = append(h.calls, fmt.Sprintf("OnError:%s:%s", providerName, err.Error()))
	return nil
}

func (h *MockHook) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	h.calls = append(h.calls, fmt.Sprintf("OnTranslate:%s:%s:%s:%s", providerName, key, lang, result))
	return nil
}

func (h *MockHook) GetCalls() []string {
	return h.calls
}

func (h *MockHook) Reset() {
	h.calls = make([]string, 0)
}

// MockMiddleware is a mock implementation of Middleware for testing.
type MockMiddleware struct {
	name  string
	calls []string
}

func NewMockMiddleware(name string) *MockMiddleware {
	return &MockMiddleware{
		name:  name,
		calls: make([]string, 0),
	}
}

func (m *MockMiddleware) Name() string {
	return m.name
}

func (m *MockMiddleware) WrapTranslate(next interfaces.TranslateFunc) interfaces.TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		m.calls = append(m.calls, fmt.Sprintf("WrapTranslate:%s:%s", key, lang))
		return next(ctx, key, lang, params)
	}
}

func (m *MockMiddleware) OnStart(ctx context.Context, providerName string) error {
	m.calls = append(m.calls, fmt.Sprintf("OnStart:%s", providerName))
	return nil
}

func (m *MockMiddleware) OnStop(ctx context.Context, providerName string) error {
	m.calls = append(m.calls, fmt.Sprintf("OnStop:%s", providerName))
	return nil
}

func (m *MockMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	m.calls = append(m.calls, fmt.Sprintf("OnError:%s:%s", providerName, err.Error()))
	return nil
}

func (m *MockMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	m.calls = append(m.calls, fmt.Sprintf("OnTranslate:%s:%s:%s:%s", providerName, key, lang, result))
	return nil
}

func (m *MockMiddleware) GetCalls() []string {
	return m.calls
}

func (m *MockMiddleware) Reset() {
	m.calls = make([]string, 0)
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Error("expected registry to be created, but got nil")
	}

	if len(registry.GetProviderNames()) != 0 {
		t.Errorf("expected empty provider names, got %v", registry.GetProviderNames())
	}

	if len(registry.GetHooks()) != 0 {
		t.Errorf("expected empty hooks, got %d", len(registry.GetHooks()))
	}

	if len(registry.GetMiddlewares()) != 0 {
		t.Errorf("expected empty middlewares, got %d", len(registry.GetMiddlewares()))
	}

	if registry.GetActiveInstances() != 0 {
		t.Errorf("expected 0 active instances, got %d", registry.GetActiveInstances())
	}
}

func TestRegistryRegisterProvider(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name        string
		factory     interfaces.ProviderFactory
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid factory",
			factory:     NewMockProviderFactory("test-provider"),
			expectError: false,
		},
		{
			name:        "nil factory",
			factory:     nil,
			expectError: true,
			errorMsg:    "factory cannot be nil",
		},
		{
			name:        "empty factory name",
			factory:     NewMockProviderFactory(""),
			expectError: true,
			errorMsg:    "factory name cannot be empty",
		},
		{
			name:        "duplicate factory name",
			factory:     NewMockProviderFactory("test-provider"), // same as first test
			expectError: true,
			errorMsg:    "provider factory 'test-provider' is already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterProvider(tt.factory)

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

				if tt.factory != nil {
					if !registry.HasProvider(tt.factory.Name()) {
						t.Errorf("expected provider '%s' to be registered", tt.factory.Name())
					}
				}
			}
		})
	}
}

func TestRegistryCreateProvider(t *testing.T) {
	registry := NewRegistry()
	factory := NewMockProviderFactory("test-provider")

	err := registry.RegisterProvider(factory)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	tests := []struct {
		name         string
		providerName string
		config       interface{}
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid provider creation",
			providerName: "test-provider",
			config:       map[string]interface{}{"test": "config"},
			expectError:  false,
		},
		{
			name:         "nonexistent provider",
			providerName: "nonexistent",
			config:       nil,
			expectError:  true,
			errorMsg:     "provider factory 'nonexistent' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := registry.CreateProvider(tt.providerName, tt.config)

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

				if provider == nil {
					t.Error("expected provider to be created, but got nil")
				}
			}
		})
	}
}

func TestRegistryGetProviderNames(t *testing.T) {
	registry := NewRegistry()

	// Should be empty initially
	names := registry.GetProviderNames()
	if len(names) != 0 {
		t.Errorf("expected empty provider names, got %v", names)
	}

	// Register some providers
	providers := []string{"provider-a", "provider-c", "provider-b"} // intentionally out of order
	for _, name := range providers {
		factory := NewMockProviderFactory(name)
		err := registry.RegisterProvider(factory)
		if err != nil {
			t.Fatalf("failed to register provider '%s': %v", name, err)
		}
	}

	// Should return sorted names
	names = registry.GetProviderNames()
	expected := []string{"provider-a", "provider-b", "provider-c"}

	if len(names) != len(expected) {
		t.Errorf("expected %d provider names, got %d", len(expected), len(names))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected provider name '%s' at index %d, got '%s'", expected[i], i, name)
		}
	}
}

func TestRegistryHasProvider(t *testing.T) {
	registry := NewRegistry()
	factory := NewMockProviderFactory("test-provider")

	// Should not exist initially
	if registry.HasProvider("test-provider") {
		t.Error("expected provider not to exist, but it was found")
	}

	// Register the provider
	err := registry.RegisterProvider(factory)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Should exist now
	if !registry.HasProvider("test-provider") {
		t.Error("expected provider to exist, but it was not found")
	}

	// Should not find nonexistent provider
	if registry.HasProvider("nonexistent") {
		t.Error("expected nonexistent provider not to be found, but it was")
	}
}

func TestRegistryAddHook(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name        string
		hook        interfaces.Hook
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid hook",
			hook:        NewMockHook("test-hook", 1),
			expectError: false,
		},
		{
			name:        "nil hook",
			hook:        nil,
			expectError: true,
			errorMsg:    "hook cannot be nil",
		},
		{
			name:        "empty hook name",
			hook:        NewMockHook("", 1),
			expectError: true,
			errorMsg:    "hook name cannot be empty",
		},
		{
			name:        "duplicate hook name",
			hook:        NewMockHook("test-hook", 2), // same name as first test
			expectError: true,
			errorMsg:    "hook 'test-hook' is already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.AddHook(tt.hook)

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

func TestRegistryHookPriority(t *testing.T) {
	registry := NewRegistry()

	// Add hooks with different priorities
	hook3 := NewMockHook("hook-priority-3", 3)
	hook1 := NewMockHook("hook-priority-1", 1)
	hook2 := NewMockHook("hook-priority-2", 2)

	registry.AddHook(hook3)
	registry.AddHook(hook1)
	registry.AddHook(hook2)

	hooks := registry.GetHooks()
	if len(hooks) != 3 {
		t.Errorf("expected 3 hooks, got %d", len(hooks))
	}

	// Should be sorted by priority
	expectedOrder := []string{"hook-priority-1", "hook-priority-2", "hook-priority-3"}
	for i, hook := range hooks {
		if hook.Name() != expectedOrder[i] {
			t.Errorf("expected hook '%s' at index %d, got '%s'", expectedOrder[i], i, hook.Name())
		}
	}
}

func TestRegistryRemoveHook(t *testing.T) {
	registry := NewRegistry()
	hook := NewMockHook("test-hook", 1)

	// Should fail to remove nonexistent hook
	err := registry.RemoveHook("nonexistent")
	if err == nil {
		t.Error("expected error removing nonexistent hook, but got none")
	}

	// Should fail with empty name
	err = registry.RemoveHook("")
	if err == nil {
		t.Error("expected error with empty hook name, but got none")
	}

	// Add hook
	err = registry.AddHook(hook)
	if err != nil {
		t.Fatalf("failed to add hook: %v", err)
	}

	if len(registry.GetHooks()) != 1 {
		t.Errorf("expected 1 hook, got %d", len(registry.GetHooks()))
	}

	// Remove hook
	err = registry.RemoveHook("test-hook")
	if err != nil {
		t.Errorf("expected no error removing hook, got: %v", err)
	}

	if len(registry.GetHooks()) != 0 {
		t.Errorf("expected 0 hooks after removal, got %d", len(registry.GetHooks()))
	}
}

func TestRegistryAddMiddleware(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name        string
		middleware  interfaces.Middleware
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid middleware",
			middleware:  NewMockMiddleware("test-middleware"),
			expectError: false,
		},
		{
			name:        "nil middleware",
			middleware:  nil,
			expectError: true,
			errorMsg:    "middleware cannot be nil",
		},
		{
			name:        "empty middleware name",
			middleware:  NewMockMiddleware(""),
			expectError: true,
			errorMsg:    "middleware name cannot be empty",
		},
		{
			name:        "duplicate middleware name",
			middleware:  NewMockMiddleware("test-middleware"), // same name as first test
			expectError: true,
			errorMsg:    "middleware 'test-middleware' is already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.AddMiddleware(tt.middleware)

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

func TestRegistryRemoveMiddleware(t *testing.T) {
	registry := NewRegistry()
	middleware := NewMockMiddleware("test-middleware")

	// Should fail to remove nonexistent middleware
	err := registry.RemoveMiddleware("nonexistent")
	if err == nil {
		t.Error("expected error removing nonexistent middleware, but got none")
	}

	// Should fail with empty name
	err = registry.RemoveMiddleware("")
	if err == nil {
		t.Error("expected error with empty middleware name, but got none")
	}

	// Add middleware
	err = registry.AddMiddleware(middleware)
	if err != nil {
		t.Fatalf("failed to add middleware: %v", err)
	}

	if len(registry.GetMiddlewares()) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(registry.GetMiddlewares()))
	}

	// Remove middleware
	err = registry.RemoveMiddleware("test-middleware")
	if err != nil {
		t.Errorf("expected no error removing middleware, got: %v", err)
	}

	if len(registry.GetMiddlewares()) != 0 {
		t.Errorf("expected 0 middlewares after removal, got %d", len(registry.GetMiddlewares()))
	}
}

func TestRegistryWithHooksAndMiddlewares(t *testing.T) {
	registry := NewRegistry()
	factory := NewMockProviderFactory("test-provider")
	hook := NewMockHook("test-hook", 1)
	middleware := NewMockMiddleware("test-middleware")

	// Register factory, hook, and middleware
	err := registry.RegisterProvider(factory)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	err = registry.AddHook(hook)
	if err != nil {
		t.Fatalf("failed to add hook: %v", err)
	}

	err = registry.AddMiddleware(middleware)
	if err != nil {
		t.Fatalf("failed to add middleware: %v", err)
	}

	// Create provider instance
	provider, err := registry.CreateProvider("test-provider", nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Start provider (should trigger hooks and middlewares)
	err = provider.Start(ctx)
	if err != nil {
		t.Errorf("expected no error starting provider, got: %v", err)
	}

	// Translate (should trigger middleware and hook)
	result, err := provider.Translate(ctx, "hello", "en", nil)
	if err != nil {
		t.Errorf("expected no error translating, got: %v", err)
	}
	if result != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", result)
	}

	// Stop provider (should trigger hooks and middlewares)
	err = provider.Stop(ctx)
	if err != nil {
		t.Errorf("expected no error stopping provider, got: %v", err)
	}

	// Verify hook was called
	hookCalls := hook.GetCalls()
	if len(hookCalls) < 3 { // OnStart, OnTranslate, OnStop
		t.Errorf("expected at least 3 hook calls, got %d: %v", len(hookCalls), hookCalls)
	}

	// Verify middleware was called
	middlewareCalls := middleware.GetCalls()
	if len(middlewareCalls) < 3 { // OnStart, WrapTranslate, OnStop
		t.Errorf("expected at least 3 middleware calls, got %d: %v", len(middlewareCalls), middlewareCalls)
	}
}

func TestRegistryShutdown(t *testing.T) {
	registry := NewRegistry()
	factory := NewMockProviderFactory("test-provider")

	err := registry.RegisterProvider(factory)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Create multiple provider instances
	provider1, err := registry.CreateProvider("test-provider", nil)
	if err != nil {
		t.Fatalf("failed to create provider 1: %v", err)
	}

	provider2, err := registry.CreateProvider("test-provider", nil)
	if err != nil {
		t.Fatalf("failed to create provider 2: %v", err)
	}

	ctx := context.Background()

	// Start providers
	provider1.Start(ctx)
	provider2.Start(ctx)

	// Verify active instances
	if registry.GetActiveInstances() != 2 {
		t.Errorf("expected 2 active instances, got %d", registry.GetActiveInstances())
	}

	// Shutdown registry
	err = registry.Shutdown(ctx)
	if err != nil {
		t.Errorf("expected no error during shutdown, got: %v", err)
	}

	// Verify instances are cleared
	if registry.GetActiveInstances() != 0 {
		t.Errorf("expected 0 active instances after shutdown, got %d", registry.GetActiveInstances())
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test that the default registry functions work
	factory := NewMockProviderFactory("default-test-provider")

	err := RegisterProvider(factory)
	if err != nil {
		t.Errorf("expected no error registering with default registry, got: %v", err)
	}

	if !HasProvider("default-test-provider") {
		t.Error("expected provider to be registered with default registry")
	}

	names := GetProviderNames()
	found := false
	for _, name := range names {
		if name == "default-test-provider" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected provider name in default registry provider names")
	}

	hook := NewMockHook("default-test-hook", 1)
	err = AddHook(hook)
	if err != nil {
		t.Errorf("expected no error adding hook to default registry, got: %v", err)
	}

	middleware := NewMockMiddleware("default-test-middleware")
	err = AddMiddleware(middleware)
	if err != nil {
		t.Errorf("expected no error adding middleware to default registry, got: %v", err)
	}

	provider, err := CreateProvider("default-test-provider", nil)
	if err != nil {
		t.Errorf("expected no error creating provider with default registry, got: %v", err)
	}

	if provider == nil {
		t.Error("expected provider to be created with default registry")
	}
}
