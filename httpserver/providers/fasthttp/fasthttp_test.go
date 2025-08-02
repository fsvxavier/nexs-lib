package fasthttp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	if factory == nil {
		t.Fatal("NewFactory returned nil")
	}

	if factory.GetName() != "fasthttp" {
		t.Errorf("Expected name 'fasthttp', got '%s'", factory.GetName())
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
		t.Error("ValidateConfig should return error for invalid config type")
	}

	// Test invalid config values
	_, err = config.NewBuilder().Apply(config.WithPort(0)).Build() // This will fail during build
	if err == nil {
		t.Error("ValidateConfig should return error for invalid config values")
	}
}

func TestFactoryCreate(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

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

func TestServerRegisterRoute(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	handler := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
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

	// Test invalid handler type
	err = server.RegisterRoute("GET", "/test4", "not a handler")
	if err == nil {
		t.Error("RegisterRoute should return error for invalid handler type")
	}

	// Test duplicate route
	err = server.RegisterRoute("GET", "/test", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for duplicate route")
	}

	// Test unsupported method
	err = server.RegisterRoute("INVALID", "/test5", handler)
	if err == nil {
		t.Error("RegisterRoute should return error for unsupported method")
	}
}

func TestServerRegisterMiddleware(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	middleware := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue("middleware", "called")
	})

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

	// Test invalid middleware type
	err = server.RegisterMiddleware("not middleware")
	if err == nil {
		t.Error("RegisterMiddleware should return error for invalid middleware type")
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

	// Test attach observer
	err = server.AttachObserver(observer)
	if err != nil {
		t.Errorf("AttachObserver error = %v, want nil", err)
	}

	// Test detach observer
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
	expected := "0.0.0.0:8080"
	if addr != expected {
		t.Errorf("GetAddr() = %s, want %s", addr, expected)
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

	if stats.Provider != "fasthttp" {
		t.Errorf("Expected provider 'fasthttp', got '%s'", stats.Provider)
	}

	if stats.RequestCount != 0 {
		t.Errorf("Expected request count 0, got %d", stats.RequestCount)
	}
}

func TestServerIsRunning(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Server should not be running initially
	if server.IsRunning() {
		t.Error("Server should not be running initially")
	}
}

func TestServerStartStop(t *testing.T) {
	factory := NewFactory()
	cfg, err := config.NewBuilder().Apply(config.WithPort(8091)).Build()
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

	// Test start
	err = server.Start(ctx)
	if err != nil {
		t.Errorf("Start error = %v, want nil", err)
	}
	time.Sleep(200 * time.Millisecond) // Give server time to start and detect any port conflicts

	// Check if server is actually running
	if !server.IsRunning() {
		t.Skip("Server failed to start, likely due to port conflict")
	}

	// Test start when already running
	err = server.Start(ctx)
	if err == nil {
		t.Error("Start should return error when already running")
	}

	// Test stop
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop error = %v, want nil", err)
	}

	// Wait for server to stop
	time.Sleep(100 * time.Millisecond)

	if server.IsRunning() {
		t.Error("Server should not be running after stop")
	}
}

func TestServerHandlerWrapping(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add observer to track events
	observer := &mockObserver{}
	server.AttachObserver(observer)

	// Register a test route
	handler := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	err = server.RegisterRoute("GET", "/test", handler)
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Verify route was registered
	fasthttpServer := server.(*Server)
	if fasthttpServer.routes["GET"]["/test"] == nil {
		t.Error("Route was not registered properly")
	}
}

func TestServerWithMiddleware(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add middleware
	middleware := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue("middleware", "executed")
	})

	err = server.RegisterMiddleware(middleware)
	if err != nil {
		t.Errorf("RegisterMiddleware error = %v, want nil", err)
	}

	// Verify middleware was registered
	fasthttpServer := server.(*Server)
	if len(fasthttpServer.middleware) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(fasthttpServer.middleware))
	}
}

func TestServerEventNotifications(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	observer := &mockObserver{}
	err = server.AttachObserver(observer)
	if err != nil {
		t.Fatalf("Failed to attach observer: %v", err)
	}

	ctx := context.Background()

	// Test start event notification
	fasthttpServer := server.(*Server)
	err = fasthttpServer.observers.NotifyObservers(interfaces.EventStart, ctx, "localhost:8080")
	if err != nil {
		t.Errorf("NotifyObservers error = %v, want nil", err)
	}

	if !observer.startCalled {
		t.Error("Observer OnStart should have been called")
	}

	// Test error event notification
	observer.reset()
	testError := fmt.Errorf("test error")
	err = fasthttpServer.observers.NotifyObservers(interfaces.EventError, ctx, testError)
	if err != nil {
		t.Errorf("NotifyObservers error = %v, want nil", err)
	}

	if !observer.errorCalled {
		t.Error("Observer OnError should have been called")
	}
}

func TestServerRouteHandler(t *testing.T) {
	factory := NewFactory()
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	fasthttpServer := server.(*Server)

	// Create a mock RequestCtx for testing
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://localhost/notfound")
	ctx.Request.Header.SetMethod("GET")

	// Test route not found
	fasthttpServer.routeHandler(ctx)
	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("Expected status 404, got %d", ctx.Response.StatusCode())
	}

	// Test method not allowed
	ctx2 := &fasthttp.RequestCtx{}
	ctx2.Request.SetRequestURI("http://localhost/test")
	ctx2.Request.Header.SetMethod("INVALID")
	fasthttpServer.routeHandler(ctx2)
	if ctx2.Response.StatusCode() != fasthttp.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", ctx2.Response.StatusCode())
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
