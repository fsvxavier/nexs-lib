package graceful

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// mockGracefulServer implements interfaces.GracefulServer for testing.
type mockGracefulServer struct {
	running            int32
	connections        int64
	stopCalled         bool
	gracefulStopCalled bool
	preShutdownHooks   []func() error
	postShutdownHooks  []func() error
	drainTimeout       time.Duration
}

func newMockGracefulServer() *mockGracefulServer {
	return &mockGracefulServer{
		drainTimeout: 5 * time.Second,
	}
}

func (m *mockGracefulServer) Start() error {
	atomic.StoreInt32(&m.running, 1)
	return nil
}

func (m *mockGracefulServer) Stop(ctx context.Context) error {
	m.stopCalled = true
	atomic.StoreInt32(&m.running, 0)
	return nil
}

func (m *mockGracefulServer) SetHandler(handler http.Handler) {}

func (m *mockGracefulServer) GetAddr() string {
	return "localhost:8080"
}

func (m *mockGracefulServer) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

func (m *mockGracefulServer) GracefulStop(ctx context.Context, drainTimeout time.Duration) error {
	m.gracefulStopCalled = true
	atomic.StoreInt32(&m.running, 0)
	return nil
}

func (m *mockGracefulServer) Restart(ctx context.Context) error {
	return nil
}

func (m *mockGracefulServer) GetConnectionsCount() int64 {
	return atomic.LoadInt64(&m.connections)
}

func (m *mockGracefulServer) GetHealthStatus() interfaces.HealthStatus {
	return interfaces.HealthStatus{
		Status:      "healthy",
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      time.Minute,
		Connections: m.GetConnectionsCount(),
		Checks:      make(map[string]interfaces.HealthCheck),
	}
}

func (m *mockGracefulServer) PreShutdownHook(hook func() error) {
	m.preShutdownHooks = append(m.preShutdownHooks, hook)
}

func (m *mockGracefulServer) PostShutdownHook(hook func() error) {
	m.postShutdownHooks = append(m.postShutdownHooks, hook)
}

func (m *mockGracefulServer) SetDrainTimeout(timeout time.Duration) {
	m.drainTimeout = timeout
}

func (m *mockGracefulServer) WaitForConnections(ctx context.Context) error {
	return nil
}

func (m *mockGracefulServer) addConnection() {
	atomic.AddInt64(&m.connections, 1)
}

func (m *mockGracefulServer) removeConnection() {
	atomic.AddInt64(&m.connections, -1)
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if len(manager.servers) != 0 {
		t.Error("Expected empty servers map")
	}

	if manager.drainTimeout != 30*time.Second {
		t.Errorf("Expected drain timeout 30s, got %v", manager.drainTimeout)
	}
}

func TestRegisterServer(t *testing.T) {
	manager := NewManager()
	server := newMockGracefulServer()

	manager.RegisterServer("test", server)

	if len(manager.servers) != 1 {
		t.Error("Expected one server registered")
	}

	registeredServer, exists := manager.servers["test"]
	if !exists {
		t.Error("Expected server to be registered")
	}

	if registeredServer == nil {
		t.Error("Expected server to be registered with correct name")
	}
}

func TestUnregisterServer(t *testing.T) {
	manager := NewManager()
	server := newMockGracefulServer()

	manager.RegisterServer("test", server)
	manager.UnregisterServer("test")

	if len(manager.servers) != 0 {
		t.Error("Expected server to be unregistered")
	}
}

func TestSetDrainTimeout(t *testing.T) {
	manager := NewManager()
	timeout := 45 * time.Second

	manager.SetDrainTimeout(timeout)

	if manager.drainTimeout != timeout {
		t.Errorf("Expected drain timeout %v, got %v", timeout, manager.drainTimeout)
	}
}

func TestSetShutdownTimeout(t *testing.T) {
	manager := NewManager()
	timeout := 90 * time.Second

	manager.SetShutdownTimeout(timeout)

	if manager.shutdownTimeout != timeout {
		t.Errorf("Expected shutdown timeout %v, got %v", timeout, manager.shutdownTimeout)
	}
}

func TestAddPreShutdownHook(t *testing.T) {
	manager := NewManager()
	called := false

	hook := func() error {
		called = true
		return nil
	}

	manager.AddPreShutdownHook(hook)

	if len(manager.preShutdownHooks) != 1 {
		t.Error("Expected one pre-shutdown hook")
	}

	// Execute hook to test
	err := manager.preShutdownHooks[0]()
	if err != nil {
		t.Errorf("Hook execution failed: %v", err)
	}

	if !called {
		t.Error("Expected hook to be called")
	}
}

func TestAddPostShutdownHook(t *testing.T) {
	manager := NewManager()
	called := false

	hook := func() error {
		called = true
		return nil
	}

	manager.AddPostShutdownHook(hook)

	if len(manager.postShutdownHooks) != 1 {
		t.Error("Expected one post-shutdown hook")
	}

	// Execute hook to test
	err := manager.postShutdownHooks[0]()
	if err != nil {
		t.Errorf("Hook execution failed: %v", err)
	}

	if !called {
		t.Error("Expected hook to be called")
	}
}

func TestAddHealthCheck(t *testing.T) {
	manager := NewManager()

	healthCheck := func() interfaces.HealthCheck {
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "OK",
			Duration:  time.Millisecond,
			Timestamp: time.Now(),
		}
	}

	manager.AddHealthCheck("database", healthCheck)

	if len(manager.healthChecks) != 1 {
		t.Error("Expected one health check")
	}

	// Test health check execution
	check := manager.healthChecks["database"]()
	if check.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", check.Status)
	}
}

func TestGetHealthStatus(t *testing.T) {
	manager := NewManager()

	// Add a health check
	manager.AddHealthCheck("test", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Test OK",
			Duration:  time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	status := manager.GetHealthStatus()

	if status.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", status.Status)
	}

	if status.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", status.Version)
	}

	if len(status.Checks) != 1 {
		t.Errorf("Expected 1 health check, got %d", len(status.Checks))
	}
}

func TestGracefulShutdown(t *testing.T) {
	manager := NewManager()
	server := newMockGracefulServer()

	// Start the mock server
	server.Start()
	manager.RegisterServer("test", server)

	// Add hooks to test execution
	preHookCalled := false
	postHookCalled := false

	manager.AddPreShutdownHook(func() error {
		preHookCalled = true
		return nil
	})

	manager.AddPostShutdownHook(func() error {
		postHookCalled = true
		return nil
	})

	// Perform graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := manager.GracefulShutdown(ctx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Verify server was stopped
	if !server.gracefulStopCalled {
		t.Error("Expected graceful stop to be called")
	}

	if !preHookCalled {
		t.Error("Expected pre-shutdown hook to be called")
	}

	if !postHookCalled {
		t.Error("Expected post-shutdown hook to be called")
	}
}

func TestConnectionsManagement(t *testing.T) {
	manager := NewManager()

	// Test initial count
	if manager.GetConnectionsCount() != 0 {
		t.Error("Expected 0 initial connections")
	}

	// Increment connections
	manager.IncrementConnections()
	manager.IncrementConnections()

	if manager.GetConnectionsCount() != 2 {
		t.Errorf("Expected 2 connections, got %d", manager.GetConnectionsCount())
	}

	// Decrement connections
	manager.DecrementConnections()

	if manager.GetConnectionsCount() != 1 {
		t.Errorf("Expected 1 connection, got %d", manager.GetConnectionsCount())
	}
}

func TestDoubleShutdown(t *testing.T) {
	manager := NewManager()
	server := newMockGracefulServer()
	server.Start()
	manager.RegisterServer("test", server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First shutdown should succeed
	err := manager.GracefulShutdown(ctx)
	if err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// Second shutdown should fail
	err = manager.GracefulShutdown(ctx)
	if err == nil {
		t.Error("Expected second shutdown to fail")
	}

	if err.Error() != "shutdown already in progress" {
		t.Errorf("Expected 'shutdown already in progress' error, got: %v", err)
	}
}
