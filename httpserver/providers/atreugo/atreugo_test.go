package atreugo

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/savsgio/atreugo/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

// Mock Observer for testing
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

func TestFactory_Create(t *testing.T) {
	factory := &Factory{}

	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	require.NoError(t, err)
	assert.NotNil(t, server)

	atreugoServer, ok := server.(*Server)
	assert.True(t, ok)
	assert.NotNil(t, atreugoServer.atreugo)
}

func TestFactory_CreateInvalidConfig(t *testing.T) {
	factory := &Factory{}

	// Test with invalid config type
	_, err := factory.Create("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")

	// Test with invalid config values
	_, err = config.NewBuilder().Apply(config.WithPort(-1)).Build() // Invalid port
	assert.Error(t, err)
}

func TestFactory_GetName(t *testing.T) {
	factory := &Factory{}
	assert.Equal(t, "atreugo", factory.GetName())
}

func TestFactory_GetDefaultConfig(t *testing.T) {
	factory := &Factory{}
	cfg := factory.GetDefaultConfig()
	assert.NotNil(t, cfg)

	baseConfig, ok := cfg.(*config.BaseConfig)
	assert.True(t, ok)
	assert.NotNil(t, baseConfig)
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := &Factory{}

	// Valid config
	cfg := config.NewBaseConfig()
	err := factory.ValidateConfig(cfg)
	assert.NoError(t, err)

	// Invalid config type
	err = factory.ValidateConfig("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

func TestServer_StartStop(t *testing.T) {
	factory := &Factory{}
	cfg, err := config.NewBuilder().Apply(config.WithPort(9999)).Build()
	require.NoError(t, err)

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test starting the server
	err = server.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, server.IsRunning())

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping the server
	err = server.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, server.IsRunning())

	// Test starting already running server
	err = server.Start(ctx)
	assert.NoError(t, err)
	err = server.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stopping not running server
	server.Stop(ctx)
	err = server.Stop(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestServer_RegisterRoute(t *testing.T) {
	factory := &Factory{}
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	// Create a simple Atreugo handler
	handler := func(ctx *atreugo.RequestCtx) error {
		return ctx.TextResponse("Hello World", http.StatusOK)
	}

	// Test valid route registration
	err = server.RegisterRoute("GET", "/test", handler)
	assert.NoError(t, err)

	// Test duplicate route
	err = server.RegisterRoute("GET", "/test", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test invalid method
	err = server.RegisterRoute("", "/empty", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method cannot be empty")

	// Test invalid path
	err = server.RegisterRoute("GET", "", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path cannot be empty")

	// Test invalid handler
	err = server.RegisterRoute("GET", "/invalid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")

	// Test wrong handler type
	err = server.RegisterRoute("GET", "/wrong", "not-a-handler")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler must be an atreugo.View")

	// Test unsupported method
	err = server.RegisterRoute("INVALID", "/unsupported", handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported HTTP method")
}

func TestServer_RegisterMiddleware(t *testing.T) {
	factory := &Factory{}
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	// Create a simple Atreugo middleware
	middleware := func(ctx *atreugo.RequestCtx) error {
		ctx.Response.Header.Set("X-Test", "middleware")
		return ctx.Next()
	}

	// Test valid middleware registration
	err = server.RegisterMiddleware(middleware)
	assert.NoError(t, err)

	// Test nil middleware
	err = server.RegisterMiddleware(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "middleware cannot be nil")

	// Test wrong middleware type
	err = server.RegisterMiddleware("not-a-middleware")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "middleware must be an atreugo.Middleware")
}

func TestServer_ObserverManagement(t *testing.T) {
	factory := &Factory{}
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	observer := &mockObserver{}

	// Test attach observer
	err = server.AttachObserver(observer)
	assert.NoError(t, err)

	// Test detach observer
	err = server.DetachObserver(observer)
	assert.NoError(t, err)
}

func TestServer_GetAddr(t *testing.T) {
	factory := &Factory{}
	cfg, err := config.NewBuilder().Apply(
		config.WithAddr("192.168.1.1"),
		config.WithPort(3000),
	).Build()
	require.NoError(t, err)

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	addr := server.GetAddr()
	assert.Equal(t, "192.168.1.1:3000", addr)
}

func TestServer_GetStats(t *testing.T) {
	factory := &Factory{}
	cfg := config.NewBaseConfig()

	server, err := factory.Create(cfg)
	require.NoError(t, err)

	stats := server.GetStats()
	assert.Equal(t, "atreugo", stats.Provider)
	assert.Equal(t, int64(0), stats.RequestCount)
}

func TestServer_WithObservers(t *testing.T) {
	observer := &mockObserver{}
	cfg, err := config.NewBuilder().Apply(
		config.WithObserver(observer),
		config.WithPort(8081),
	).Build()
	require.NoError(t, err)

	factory := &Factory{}
	server, err := factory.Create(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Start server to trigger events
	err = server.Start(ctx)
	assert.NoError(t, err)

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	err = server.Stop(ctx)
	assert.NoError(t, err)

	// Check that events were triggered
	assert.True(t, observer.startCalled)
	assert.True(t, observer.stopCalled)
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.Equal(t, "atreugo", factory.GetName())
}
