package atreugo

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/savsgio/atreugo/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Mock Observer for testing
type mockObserver struct {
	events []interfaces.ServerEvent
	mutex  sync.RWMutex
}

func (m *mockObserver) OnServerEvent(event interfaces.ServerEvent, ctx context.Context, data interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockObserver) GetEvents() []interfaces.ServerEvent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return append([]interfaces.ServerEvent{}, m.events...)
}

func (m *mockObserver) ClearEvents() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = nil
}

func TestFactory_Create(t *testing.T) {
	factory := &Factory{}

	cfg := config.NewBaseConfig()
	cfg.SetAddr("localhost")
	cfg.SetPort(8080)

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
	cfg := config.NewBaseConfig()
	cfg.SetPort(-1) // Invalid port

	_, err = factory.Create(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
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
	cfg := config.NewBaseConfig()
	cfg.SetPort(9999) // Use a different port to avoid conflicts

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
	cfg := config.NewBaseConfig()
	cfg.SetAddr("192.168.1.1")
	cfg.SetPort(3000)

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
	cfg := config.NewBaseConfig()
	cfg.AddObserver(observer)
	cfg.SetPort(8081) // Use different port

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
	events := observer.GetEvents()
	assert.Contains(t, events, interfaces.EventStart)
	assert.Contains(t, events, interfaces.EventStop)
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.Equal(t, "atreugo", factory.GetName())
}
