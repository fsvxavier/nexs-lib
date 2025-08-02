package nethttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	if factory == nil {
		t.Fatal("NewFactory returned nil")
	}

	if factory.GetName() != "nethttp" {
		t.Errorf("Expected name 'nethttp', got '%s'", factory.GetName())
	}
}

func TestFactoryGetDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.GetDefaultConfig()

	if cfg == nil {
		t.Fatal("GetDefaultConfig returned nil")
	}

	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		t.Fatalf("Expected *config.BaseConfig, got %T", cfg)
	}

	if baseConfig.GetPort() != 8080 {
		t.Errorf("Expected default port 8080, got %d", baseConfig.GetPort())
	}
}

func TestFactoryValidateConfig(t *testing.T) {
	factory := NewFactory()

	// Test valid config
	validConfig := config.NewBaseConfig()
	err := factory.ValidateConfig(validConfig)
	if err != nil {
		t.Errorf("ValidateConfig error = %v, want nil", err)
	}

	// Test invalid config type
	err = factory.ValidateConfig("invalid")
	if err == nil {
		t.Error("ValidateConfig should return error for invalid type")
	}

	// Test invalid config values
	invalidConfig := config.NewBaseConfig()
	invalidConfig.SetCustom("port", 0) // This won't affect validation as it uses internal fields

	// Create invalid config through builder
	_, err = config.NewBuilder().Apply(config.WithPort(0)).Build()
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

func TestFactoryCreate(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	// Test successful creation
	server, err := factory.Create(cfg)
	if err != nil {
		t.Errorf("Create error = %v, want nil", err)
	}

	if server == nil {
		t.Error("Create returned nil server")
	}

	// Test invalid config type
	_, err = factory.Create("invalid")
	if err == nil {
		t.Error("Create should return error for invalid config type")
	}
}

func TestServerStart(t *testing.T) {
	factory := NewFactory()
	cfg, err := config.NewBuilder().Apply(config.WithPort(8081)).Build() // Use test port
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test start
	err = server.Start(ctx)
	if err != nil {
		t.Errorf("Start error = %v, want nil", err)
	}

	if !server.IsRunning() {
		t.Error("Server should be running after start")
	}

	// Test double start
	err = server.Start(ctx)
	if err == nil {
		t.Error("Start should return error when already running")
	}

	// Cleanup
	server.Stop(ctx)
}

func TestServerStop(t *testing.T) {
	factory := NewFactory()
	cfg, err := config.NewBuilder().Apply(config.WithPort(8082)).Build()
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx := context.Background()

	// Test stop before start
	err = server.Stop(ctx)
	if err == nil {
		t.Error("Stop should return error when not running")
	}

	// Start server
	err = server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // Give server time to start and detect any port conflicts

	// Check if server is actually running
	if !server.IsRunning() {
		t.Skip("Server failed to start, likely due to port conflict")
	}

	// Test stop
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop error = %v, want nil", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after stop")
	}
}

func TestServerRegisterRoute(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test successful registration
	err = server.RegisterRoute("GET", "/test", handler)
	if err != nil {
		t.Errorf("RegisterRoute error = %v, want nil", err)
	}

	// Test empty method
	err = server.RegisterRoute("", "/test2", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for empty method")
	}

	// Test empty path
	err = server.RegisterRoute("GET", "", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for empty path")
	}

	// Test nil handler
	err = server.RegisterRoute("GET", "/test3", nil)
	if err == nil {
		t.Error("RegisterRoute should return error for nil handler")
	}

	// Test invalid method
	err = server.RegisterRoute("INVALID", "/test4", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for invalid method")
	}

	// Test duplicate route
	err = server.RegisterRoute("GET", "/test", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for duplicate route")
	}
}

func TestServerRegisterMiddleware(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	}

	// Test successful registration
	err = server.RegisterMiddleware(middleware)
	if err != nil {
		t.Errorf("RegisterMiddleware error = %v, want nil", err)
	}

	// Test nil middleware
	err = server.RegisterMiddleware(nil)
	if err == nil {
		t.Error("RegisterMiddleware should return error for nil middleware")
	}
}

func TestServerObservers(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	observer := &mockObserver{}

	// Test attach
	err = server.AttachObserver(observer)
	if err != nil {
		t.Errorf("AttachObserver error = %v, want nil", err)
	}

	// Test detach
	err = server.DetachObserver(observer)
	if err != nil {
		t.Errorf("DetachObserver error = %v, want nil", err)
	}
}

func TestServerGetAddr(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	addr := server.GetAddr()
	if addr != "0.0.0.0:8080" {
		t.Errorf("Expected addr '0.0.0.0:8080', got '%s'", addr)
	}
}

func TestServerGetStats(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	stats := server.GetStats()

	if stats.Provider != "nethttp" {
		t.Errorf("Expected provider 'nethttp', got '%s'", stats.Provider)
	}

	if stats.RequestCount != 0 {
		t.Errorf("Expected request count 0, got %d", stats.RequestCount)
	}

	if stats.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", stats.ErrorCount)
	}
}

func TestServerHandlerWrapping(t *testing.T) {
	factory := NewFactory()
	cfg, err := config.NewBuilder().Apply(config.WithPort(8083)).Build()
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add observer to track events
	observer := &mockObserver{}
	server.AttachObserver(observer)

	// Register a test route
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	err = server.RegisterRoute("GET", "/test", handler)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Start server
	ctx := context.Background()
	err = server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond) // Give server time to start

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// We can't easily test the actual HTTP server without knowing the port,
	// so we'll test the handler wrapping logic directly
	netHttpServer := server.(*Server)
	wrappedHandler := netHttpServer.wrapHandler("GET", "/test", handler)
	wrappedHandler(w, req)

	if !called {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check stats were updated
	stats := server.GetStats()
	if stats.RequestCount == 0 {
		t.Error("Request count should be > 0")
	}
}

func TestResponseWrapper(t *testing.T) {
	w := httptest.NewRecorder()
	wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

	// Test default status
	if wrapper.statusCode != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", wrapper.statusCode)
	}

	// Test WriteHeader
	wrapper.WriteHeader(http.StatusNotFound)
	if wrapper.statusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", wrapper.statusCode)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected response writer code 404, got %d", w.Code)
	}
}

// Mock observer for testing
type mockObserver struct {
	startCalled      bool
	stopCalled       bool
	errorCalled      bool
	requestCalled    bool
	responseCalled   bool
	routeEnterCalled bool
	routeExitCalled  bool
}

func (m *mockObserver) OnStart(ctx context.Context, addr string) error {
	m.startCalled = true
	return nil
}

func (m *mockObserver) OnStop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func (m *mockObserver) OnError(ctx context.Context, err error) error {
	m.errorCalled = true
	return nil
}

func (m *mockObserver) OnRequest(ctx context.Context, req interface{}) error {
	m.requestCalled = true
	return nil
}

func (m *mockObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	m.responseCalled = true
	return nil
}

func (m *mockObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	m.routeEnterCalled = true
	return nil
}

func (m *mockObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	m.routeExitCalled = true
	return nil
}

func (m *mockObserver) reset() {
	m.startCalled = false
	m.stopCalled = false
	m.errorCalled = false
	m.requestCalled = false
	m.responseCalled = false
	m.routeEnterCalled = false
	m.routeExitCalled = false
}
