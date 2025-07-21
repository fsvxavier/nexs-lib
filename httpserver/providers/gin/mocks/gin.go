// Package mocks provides mock implementations for testing the Gin HTTP server provider.
package mocks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// MockGinServer is a mock implementation of interfaces.HTTPServer for testing.
type MockGinServer struct {
	StartFunc               func() error
	StopFunc                func(ctx context.Context) error
	SetHandlerFunc          func(handler http.Handler)
	GetAddrFunc             func() string
	IsRunningFunc           func() bool
	GetConfigFunc           func() *config.Config
	GracefulStopFunc        func(ctx context.Context, drainTimeout time.Duration) error
	RestartFunc             func(ctx context.Context) error
	GetConnectionsCountFunc func() int64
	GetHealthStatusFunc     func() interfaces.HealthStatus
	PreShutdownHookFunc     func(hook func() error)
	PostShutdownHookFunc    func(hook func() error)
	SetDrainTimeoutFunc     func(timeout time.Duration)
	WaitForConnectionsFunc  func(ctx context.Context) error

	// State tracking
	Handler           http.Handler
	Running           bool
	Config            *config.Config
	Addr              string
	Connections       int64
	PreShutdownHooks  []func() error
	PostShutdownHooks []func() error
	DrainTimeout      time.Duration
}

// NewMockGinServer creates a new mock Gin server.
func NewMockGinServer() *MockGinServer {
	return &MockGinServer{
		Running:           false,
		Addr:              "localhost:8080",
		Config:            config.DefaultConfig(),
		Connections:       0,
		PreShutdownHooks:  make([]func() error, 0),
		PostShutdownHooks: make([]func() error, 0),
		DrainTimeout:      30 * time.Second,
	}
}

// Start starts the mock server.
func (m *MockGinServer) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	m.Running = true
	return nil
}

// Stop stops the mock server.
func (m *MockGinServer) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	m.Running = false
	return nil
}

// SetHandler sets the HTTP handler.
func (m *MockGinServer) SetHandler(handler http.Handler) {
	if m.SetHandlerFunc != nil {
		m.SetHandlerFunc(handler)
		return
	}
	m.Handler = handler
}

// GetAddr returns the server address.
func (m *MockGinServer) GetAddr() string {
	if m.GetAddrFunc != nil {
		return m.GetAddrFunc()
	}
	return m.Addr
}

// IsRunning returns whether the server is running.
func (m *MockGinServer) IsRunning() bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc()
	}
	return m.Running
}

// GetConfig returns the server configuration.
func (m *MockGinServer) GetConfig() *config.Config {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc()
	}
	return m.Config
}

// GracefulStop performs a graceful shutdown with connection draining.
func (m *MockGinServer) GracefulStop(ctx context.Context, drainTimeout time.Duration) error {
	if m.GracefulStopFunc != nil {
		return m.GracefulStopFunc(ctx, drainTimeout)
	}
	m.Running = false
	return nil
}

// Restart performs a zero-downtime restart.
func (m *MockGinServer) Restart(ctx context.Context) error {
	if m.RestartFunc != nil {
		return m.RestartFunc(ctx)
	}
	return nil
}

// GetConnectionsCount returns the number of active connections.
func (m *MockGinServer) GetConnectionsCount() int64 {
	if m.GetConnectionsCountFunc != nil {
		return m.GetConnectionsCountFunc()
	}
	return m.Connections
}

// GetHealthStatus returns the current health status.
func (m *MockGinServer) GetHealthStatus() interfaces.HealthStatus {
	if m.GetHealthStatusFunc != nil {
		return m.GetHealthStatusFunc()
	}
	status := "healthy"
	if !m.Running {
		status = "stopped"
	}
	return interfaces.HealthStatus{
		Status:      status,
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      time.Minute,
		Connections: m.Connections,
		Checks:      make(map[string]interfaces.HealthCheck),
	}
}

// PreShutdownHook registers a function to be called before shutdown.
func (m *MockGinServer) PreShutdownHook(hook func() error) {
	if m.PreShutdownHookFunc != nil {
		m.PreShutdownHookFunc(hook)
		return
	}
	m.PreShutdownHooks = append(m.PreShutdownHooks, hook)
}

// PostShutdownHook registers a function to be called after shutdown.
func (m *MockGinServer) PostShutdownHook(hook func() error) {
	if m.PostShutdownHookFunc != nil {
		m.PostShutdownHookFunc(hook)
		return
	}
	m.PostShutdownHooks = append(m.PostShutdownHooks, hook)
}

// SetDrainTimeout sets the timeout for connection draining.
func (m *MockGinServer) SetDrainTimeout(timeout time.Duration) {
	if m.SetDrainTimeoutFunc != nil {
		m.SetDrainTimeoutFunc(timeout)
		return
	}
	m.DrainTimeout = timeout
}

// WaitForConnections waits for all connections to finish or timeout.
func (m *MockGinServer) WaitForConnections(ctx context.Context) error {
	if m.WaitForConnectionsFunc != nil {
		return m.WaitForConnectionsFunc(ctx)
	}
	return nil
}

// Ensure MockGinServer implements interfaces.HTTPServer.
var _ interfaces.HTTPServer = (*MockGinServer)(nil)

// MockFactory is a factory function for creating mock Gin servers.
func MockFactory(cfg interface{}) (interfaces.HTTPServer, error) {
	conf, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	mock := NewMockGinServer()
	mock.Config = conf
	if conf != nil {
		mock.Addr = conf.Addr()
	}
	return mock, nil
}
