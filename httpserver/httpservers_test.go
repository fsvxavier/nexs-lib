package httpserver

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.GetProviderCount() != 0 {
		t.Errorf("Expected 0 providers, got %d", registry.GetProviderCount())
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	factory := &mockFactory{name: "test"}

	// Test successful registration
	err := registry.Register("test", factory)
	if err != nil {
		t.Errorf("Register error = %v, want nil", err)
	}

	if registry.GetProviderCount() != 1 {
		t.Errorf("Expected 1 provider, got %d", registry.GetProviderCount())
	}

	// Test empty name
	err = registry.Register("", factory)
	if err == nil {
		t.Error("Register with empty name should return error")
	}

	// Test nil factory
	err = registry.Register("test2", nil)
	if err == nil {
		t.Error("Register with nil factory should return error")
	}

	// Test duplicate registration
	err = registry.Register("test", factory)
	if err == nil {
		t.Error("Register with duplicate name should return error")
	}
}

func TestRegistryUnregister(t *testing.T) {
	registry := NewRegistry()
	factory := &mockFactory{name: "test"}

	// Register a provider
	registry.Register("test", factory)
	if registry.GetProviderCount() != 1 {
		t.Errorf("Expected 1 provider, got %d", registry.GetProviderCount())
	}

	// Test successful unregistration
	err := registry.Unregister("test")
	if err != nil {
		t.Errorf("Unregister error = %v, want nil", err)
	}

	if registry.GetProviderCount() != 0 {
		t.Errorf("Expected 0 providers after unregister, got %d", registry.GetProviderCount())
	}

	// Test empty name
	err = registry.Unregister("")
	if err == nil {
		t.Error("Unregister with empty name should return error")
	}

	// Test non-existent provider
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Unregister with non-existent name should return error")
	}
}

func TestRegistryCreate(t *testing.T) {
	registry := NewRegistry()
	factory := &mockFactory{name: "test"}
	config := &mockConfig{}

	// Register a provider
	registry.Register("test", factory)

	// Test successful creation
	server, err := registry.Create("test", config)
	if err != nil {
		t.Errorf("Create error = %v, want nil", err)
	}

	if server == nil {
		t.Error("Create returned nil server")
	}

	// Test empty name
	_, err = registry.Create("", config)
	if err == nil {
		t.Error("Create with empty name should return error")
	}

	// Test non-existent provider
	_, err = registry.Create("nonexistent", config)
	if err == nil {
		t.Error("Create with non-existent provider should return error")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()
	factory1 := &mockFactory{name: "test1"}
	factory2 := &mockFactory{name: "test2"}

	// Test empty list
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Register providers
	registry.Register("test1", factory1)
	registry.Register("test2", factory2)

	// Test populated list
	list = registry.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(list))
	}

	// Check that both providers are in the list
	found1, found2 := false, false
	for _, name := range list {
		if name == "test1" {
			found1 = true
		}
		if name == "test2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("List does not contain expected providers: %v", list)
	}
}

func TestRegistryIsRegistered(t *testing.T) {
	registry := NewRegistry()
	factory := &mockFactory{name: "test"}

	// Test non-existent provider
	if registry.IsRegistered("test") {
		t.Error("IsRegistered should return false for non-existent provider")
	}

	// Register provider
	registry.Register("test", factory)

	// Test existing provider
	if !registry.IsRegistered("test") {
		t.Error("IsRegistered should return true for existing provider")
	}
}

func TestRegistryGetProvider(t *testing.T) {
	registry := NewRegistry()
	factory := &mockFactory{name: "test"}

	// Test non-existent provider
	_, err := registry.GetProvider("test")
	if err == nil {
		t.Error("GetProvider should return error for non-existent provider")
	}

	// Test empty name
	_, err = registry.GetProvider("")
	if err == nil {
		t.Error("GetProvider should return error for empty name")
	}

	// Register provider
	registry.Register("test", factory)

	// Test existing provider
	retrievedFactory, err := registry.GetProvider("test")
	if err != nil {
		t.Errorf("GetProvider error = %v, want nil", err)
	}

	if retrievedFactory != factory {
		t.Error("GetProvider returned different factory")
	}
}

func TestRegistryClear(t *testing.T) {
	registry := NewRegistry()
	factory1 := &mockFactory{name: "test1"}
	factory2 := &mockFactory{name: "test2"}

	// Register providers
	registry.Register("test1", factory1)
	registry.Register("test2", factory2)

	if registry.GetProviderCount() != 2 {
		t.Errorf("Expected 2 providers, got %d", registry.GetProviderCount())
	}

	// Clear registry
	registry.Clear()

	if registry.GetProviderCount() != 0 {
		t.Errorf("Expected 0 providers after clear, got %d", registry.GetProviderCount())
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.GetRegistry() == nil {
		t.Error("Manager registry is nil")
	}

	if manager.GetObserverManager() == nil {
		t.Error("Manager observer manager is nil")
	}
}

func TestManagerRegisterProvider(t *testing.T) {
	manager := NewManager()
	factory := &mockFactory{name: "test"}

	// Test successful registration
	err := manager.RegisterProvider("test", factory)
	if err != nil {
		t.Errorf("RegisterProvider error = %v, want nil", err)
	}

	// Check registration
	if !manager.IsProviderRegistered("test") {
		t.Error("Provider not registered")
	}
}

func TestManagerCreateServer(t *testing.T) {
	manager := NewManager()
	factory := &mockFactory{name: "test"}

	// Register provider
	manager.RegisterProvider("test", factory)

	// Test server creation
	server, err := manager.CreateServer("test", config.WithPort(9000))
	if err != nil {
		t.Errorf("CreateServer error = %v, want nil", err)
	}

	if server == nil {
		t.Error("CreateServer returned nil server")
	}
}

func TestManagerCreateServerWithConfig(t *testing.T) {
	manager := NewManager()
	factory := &mockFactory{name: "test"}

	// Register provider
	manager.RegisterProvider("test", factory)

	// Create config
	cfg := config.NewBaseConfig()

	// Test server creation with config
	server, err := manager.CreateServerWithConfig("test", cfg)
	if err != nil {
		t.Errorf("CreateServerWithConfig error = %v, want nil", err)
	}

	if server == nil {
		t.Error("CreateServerWithConfig returned nil server")
	}

	// Test with nil config
	_, err = manager.CreateServerWithConfig("test", nil)
	if err == nil {
		t.Error("CreateServerWithConfig with nil config should return error")
	}
}

func TestManagerListProviders(t *testing.T) {
	manager := NewManager()
	factory := &mockFactory{name: "test"}

	// Test empty list
	list := manager.ListProviders()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Register provider
	manager.RegisterProvider("test", factory)

	// Test populated list
	list = manager.ListProviders()
	if len(list) != 1 || list[0] != "test" {
		t.Errorf("Expected [test], got %v", list)
	}
}

func TestManagerGetProviderDefaultConfig(t *testing.T) {
	manager := NewManager()
	factory := &mockFactory{name: "test"}

	// Register provider
	manager.RegisterProvider("test", factory)

	// Test get default config
	config, err := manager.GetProviderDefaultConfig("test")
	if err != nil {
		t.Errorf("GetProviderDefaultConfig error = %v, want nil", err)
	}

	if config == nil {
		t.Error("GetProviderDefaultConfig returned nil config")
	}

	// Test non-existent provider
	_, err = manager.GetProviderDefaultConfig("nonexistent")
	if err == nil {
		t.Error("GetProviderDefaultConfig should return error for non-existent provider")
	}
}

func TestDefaultManager(t *testing.T) {
	// Test default manager functions
	factory := &mockFactory{name: "test"}

	// Register provider
	err := RegisterProvider("test", factory)
	if err != nil {
		t.Errorf("RegisterProvider error = %v, want nil", err)
	}

	// Check registration
	if !IsProviderRegistered("test") {
		t.Error("Provider not registered in default manager")
	}

	// List providers
	list := ListProviders()
	if len(list) == 0 {
		t.Error("Expected providers in default manager")
	}

	// Create server
	server, err := CreateServer("test", config.WithPort(9001))
	if err != nil {
		t.Errorf("CreateServer error = %v, want nil", err)
	}

	if server == nil {
		t.Error("CreateServer returned nil server")
	}

	// Get provider
	_, err = GetProvider("test")
	if err != nil {
		t.Errorf("GetProvider error = %v, want nil", err)
	}

	// Cleanup
	UnregisterProvider("test")
}

// Mock factory for testing
type mockFactory struct {
	name string
}

func (m *mockFactory) Create(config interface{}) (interfaces.HTTPServer, error) {
	return &mockServer{}, nil
}

func (m *mockFactory) GetName() string {
	return m.name
}

func (m *mockFactory) GetDefaultConfig() interface{} {
	return &mockConfig{}
}

func (m *mockFactory) ValidateConfig(config interface{}) error {
	return nil
}

// Mock server for testing
type mockServer struct{}

func (m *mockServer) Start(ctx context.Context) error {
	return nil
}

func (m *mockServer) Stop(ctx context.Context) error {
	return nil
}

func (m *mockServer) RegisterRoute(method, path string, handler interface{}) error {
	return nil
}

func (m *mockServer) RegisterMiddleware(middleware interface{}) error {
	return nil
}

func (m *mockServer) AttachObserver(observer interfaces.ServerObserver) error {
	return nil
}

func (m *mockServer) DetachObserver(observer interfaces.ServerObserver) error {
	return nil
}

func (m *mockServer) GetAddr() string {
	return "127.0.0.1:8080"
}

func (m *mockServer) IsRunning() bool {
	return false
}

func (m *mockServer) GetStats() interfaces.ServerStats {
	return interfaces.ServerStats{}
}

// Mock config for testing
type mockConfig struct{}
