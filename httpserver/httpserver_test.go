package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// mockServer implements interfaces.HTTPServer for testing
type mockServer struct {
	addr        string
	running     bool
	handler     http.Handler
	startErr    error
	stopErr     error
	connections int64
	startTime   time.Time
}

func newMockServer(cfg interface{}) (interfaces.HTTPServer, error) {
	conf, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}
	return &mockServer{
		addr: conf.Addr(),
	}, nil
}

func newFailingMockServer(cfg interface{}) (interfaces.HTTPServer, error) {
	return nil, fmt.Errorf("factory error")
}

func (m *mockServer) Start() error {
	if m.startErr != nil {
		return m.startErr
	}
	m.running = true
	return nil
}

func (m *mockServer) Stop(ctx context.Context) error {
	m.running = false
	return m.stopErr
}

func (m *mockServer) SetHandler(handler http.Handler) {
	m.handler = handler
}

func (m *mockServer) GetAddr() string {
	return m.addr
}

func (m *mockServer) IsRunning() bool {
	return m.running
}

func (m *mockServer) GetConfig() *config.Config {
	return &config.Config{}
}

// GracefulStop performs a graceful shutdown with connection draining.
func (m *mockServer) GracefulStop(ctx context.Context, drainTimeout time.Duration) error {
	m.running = false
	return nil
}

// Restart performs a zero-downtime restart.
func (m *mockServer) Restart(ctx context.Context) error {
	return nil
}

// GetConnectionsCount returns the number of active connections.
func (m *mockServer) GetConnectionsCount() int64 {
	return m.connections
}

// GetHealthStatus returns the current health status.
func (m *mockServer) GetHealthStatus() interfaces.HealthStatus {
	uptime := time.Duration(0)
	if !m.startTime.IsZero() {
		uptime = time.Since(m.startTime)
	}

	status := "healthy"
	if !m.running {
		status = "stopped"
	}

	return interfaces.HealthStatus{
		Status:      status,
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      uptime,
		Connections: m.connections,
		Checks:      make(map[string]interfaces.HealthCheck),
	}
}

// PreShutdownHook registers a function to be called before shutdown.
func (m *mockServer) PreShutdownHook(hook func()) error {
	// Mock implementation - do nothing
	return nil
}

// PostShutdownHook registers a function to be called after shutdown.
func (m *mockServer) PostShutdownHook(hook func()) error {
	// Mock implementation - do nothing
	return nil
}

// SetDrainTimeout sets the timeout for connection draining.
func (m *mockServer) SetDrainTimeout(timeout time.Duration) {
	// Mock implementation - do nothing
}

// WaitForConnections waits for all connections to finish or timeout.
func (m *mockServer) WaitForConnections(ctx context.Context) error {
	return nil
}

// mockObserver implements interfaces.ServerObserver for testing
type mockObserver struct {
	startCalls []string
	stopCalls  []string
}

func newMockObserver() *mockObserver {
	return &mockObserver{
		startCalls: make([]string, 0),
		stopCalls:  make([]string, 0),
	}
}

func (m *mockObserver) OnStart(name string) {
	m.startCalls = append(m.startCalls, name)
}

func (m *mockObserver) OnStop(name string) {
	m.stopCalls = append(m.stopCalls, name)
}

func (m *mockObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
}
func (m *mockObserver) OnBeforeRequest(name string, req *http.Request) {}
func (m *mockObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	if len(registry.ListFactories()) != 0 {
		t.Error("Expected no factories in new registry")
	}

	if len(registry.GetObservers()) != 0 {
		t.Error("Expected no observers in new registry")
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register("test", newMockServer)
	if err != nil {
		t.Fatalf("Expected no error registering factory, got: %v", err)
	}

	factories := registry.ListFactories()
	if len(factories) != 1 {
		t.Fatalf("Expected 1 factory, got %d", len(factories))
	}

	if factories[0] != "test" {
		t.Errorf("Expected factory name 'test', got '%s'", factories[0])
	}
}

func TestRegistryRegisterErrors(t *testing.T) {
	registry := NewRegistry()

	// Test empty name
	err := registry.Register("", newMockServer)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test nil factory
	err = registry.Register("test", nil)
	if err == nil {
		t.Error("Expected error for nil factory")
	}

	// Test duplicate registration
	registry.Register("test", newMockServer)
	err = registry.Register("test", newMockServer)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestRegistryUnregister(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test", newMockServer)

	err := registry.Unregister("test")
	if err != nil {
		t.Fatalf("Expected no error unregistering factory, got: %v", err)
	}

	if len(registry.ListFactories()) != 0 {
		t.Error("Expected no factories after unregistering")
	}
}

func TestRegistryUnregisterNotFound(t *testing.T) {
	registry := NewRegistry()

	err := registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Expected error for unregistering nonexistent factory")
	}
}

func TestRegistryCreate(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test", newMockServer)

	cfg := config.DefaultConfig().WithHost("localhost").WithPort(8080)
	server, err := registry.Create("test", cfg)

	if err != nil {
		t.Fatalf("Expected no error creating server, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.GetAddr() != "localhost:8080" {
		t.Errorf("Expected addr 'localhost:8080', got '%s'", server.GetAddr())
	}
}

func TestRegistryCreateWithNilConfig(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test", newMockServer)

	server, err := registry.Create("test", nil)

	if err != nil {
		t.Fatalf("Expected no error creating server with nil config, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}
}

func TestRegistryCreateNotFound(t *testing.T) {
	registry := NewRegistry()

	cfg := config.DefaultConfig()
	server, err := registry.Create("nonexistent", cfg)

	if err == nil {
		t.Error("Expected error for nonexistent factory")
	}

	if server != nil {
		t.Error("Expected nil server for nonexistent factory")
	}
}

func TestRegistryCreateFactoryError(t *testing.T) {
	registry := NewRegistry()
	registry.Register("failing", newFailingMockServer)

	observer := newMockObserver()
	registry.AttachObserver(observer)

	cfg := config.DefaultConfig()
	server, err := registry.Create("failing", cfg)

	if err == nil {
		t.Error("Expected error from failing factory")
	}

	if server != nil {
		t.Error("Expected nil server from failing factory")
	}

	// Check that error observer was called - error notifications removed in new interface
	// if len(observer.errorCalls) != 1 {
	//	t.Errorf("Expected 1 error call, got %d", len(observer.errorCalls))
	// }
}

func TestRegistryAttachObserver(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()

	registry.AttachObserver(observer)

	observers := registry.GetObservers()
	if len(observers) != 1 {
		t.Fatalf("Expected 1 observer, got %d", len(observers))
	}

	if observers[0] != observer {
		t.Error("Expected attached observer to be in list")
	}
}

func TestRegistryAttachNilObserver(t *testing.T) {
	registry := NewRegistry()

	registry.AttachObserver(nil)

	if len(registry.GetObservers()) != 0 {
		t.Error("Expected no observers after attaching nil")
	}
}

func TestRegistryDetachObserver(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()

	registry.AttachObserver(observer)
	registry.DetachObserver(observer)

	if len(registry.GetObservers()) != 0 {
		t.Error("Expected no observers after detaching")
	}
}

func TestRegistryDetachNilObserver(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()

	registry.AttachObserver(observer)
	registry.DetachObserver(nil)

	if len(registry.GetObservers()) != 1 {
		t.Error("Expected observer to remain after detaching nil")
	}
}

func TestObservableServerStart(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()
	registry.AttachObserver(observer)
	registry.Register("test", newMockServer)

	cfg := config.DefaultConfig()
	server, _ := registry.Create("test", cfg)

	err := server.Start()
	if err != nil {
		t.Fatalf("Expected no error starting server, got: %v", err)
	}

	if len(observer.startCalls) != 1 {
		t.Errorf("Expected 1 start call, got %d", len(observer.startCalls))
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running")
	}
}

func TestObservableServerStop(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()
	registry.AttachObserver(observer)
	registry.Register("test", newMockServer)

	cfg := config.DefaultConfig()
	server, _ := registry.Create("test", cfg)

	server.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Fatalf("Expected no error stopping server, got: %v", err)
	}

	if len(observer.stopCalls) != 1 {
		t.Errorf("Expected 1 stop call, got %d", len(observer.stopCalls))
	}

	if server.IsRunning() {
		t.Error("Expected server to be stopped")
	}
}

func TestObservableServerStartError(t *testing.T) {
	registry := NewRegistry()
	observer := newMockObserver()
	registry.AttachObserver(observer)

	// Create a mock server that fails on start
	failingFactory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		server := &mockServer{
			startErr: fmt.Errorf("start error"),
		}
		return server, nil
	}

	registry.Register("failing", failingFactory)

	cfg := config.DefaultConfig()
	server, _ := registry.Create("failing", cfg)

	err := server.Start()
	if err == nil {
		t.Error("Expected error starting server")
	}

	// Error notifications removed in new interface
	// if len(observer.errorCalls) != 1 {
	//	t.Errorf("Expected 1 error call, got %d", len(observer.errorCalls))
	// }
}

func TestGlobalRegistryFunctions(t *testing.T) {
	// Test Register
	err := Register("global-test", newMockServer)
	if err != nil {
		t.Fatalf("Expected no error registering with global registry, got: %v", err)
	}

	// Test ListFactories
	factories := ListFactories()
	found := false
	for _, factory := range factories {
		if factory == "global-test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'global-test' factory in global registry")
	}

	// Test Create
	cfg := config.DefaultConfig()
	server, err := Create("global-test", cfg)
	if err != nil {
		t.Fatalf("Expected no error creating server with global registry, got: %v", err)
	}
	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// Test AttachObserver
	observer := newMockObserver()
	AttachObserver(observer)

	// Test DetachObserver
	DetachObserver(observer)

	// Test Unregister
	err = Unregister("global-test")
	if err != nil {
		t.Fatalf("Expected no error unregistering with global registry, got: %v", err)
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	registry := GetDefaultRegistry()
	if registry == nil {
		t.Fatal("Expected default registry to exist")
	}

	// Test that it's the same instance
	registry2 := GetDefaultRegistry()
	if registry != registry2 {
		t.Error("Expected same default registry instance")
	}
}
