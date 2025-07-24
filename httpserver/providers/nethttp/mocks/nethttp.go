// Package mocks provides mock implementations for testing the net/http server provider.
package mocks

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// MockNetHTTPServer is a mock implementation of interfaces.HTTPServer for testing.
type MockNetHTTPServer struct {
	StartFunc      func() error
	StopFunc       func(ctx context.Context) error
	SetHandlerFunc func(handler http.Handler)
	GetAddrFunc    func() string
	IsRunningFunc  func() bool
	GetConfigFunc  func() *config.Config

	// Graceful operation functions
	GracefulStopFunc        func(ctx context.Context, drainTimeout time.Duration) error
	RestartFunc             func(ctx context.Context) error
	GetConnectionsCountFunc func() int64
	GetHealthStatusFunc     func() interfaces.HealthStatus
	PreShutdownHookFunc     func(func()) error
	PostShutdownHookFunc    func(func()) error
	SetDrainTimeoutFunc     func(timeout time.Duration)
	WaitForConnectionsFunc  func(ctx context.Context) error

	// State tracking
	Handler http.Handler
	Running bool
	Config  *config.Config
	Addr    string

	// Graceful state
	connections  int64
	startTime    time.Time
	preHooks     []func()
	postHooks    []func()
	drainTimeout time.Duration
}

// NewMockNetHTTPServer creates a new mock net/http server.
func NewMockNetHTTPServer() *MockNetHTTPServer {
	return &MockNetHTTPServer{
		Running:      false,
		Addr:         "localhost:8080",
		Config:       config.DefaultConfig(),
		connections:  0,
		startTime:    time.Now(),
		preHooks:     make([]func(), 0),
		postHooks:    make([]func(), 0),
		drainTimeout: 30 * time.Second,
	}
}

// Start starts the mock server.
func (m *MockNetHTTPServer) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	m.Running = true
	return nil
}

// Stop stops the mock server.
func (m *MockNetHTTPServer) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	m.Running = false
	return nil
}

// SetHandler sets the HTTP handler.
func (m *MockNetHTTPServer) SetHandler(handler http.Handler) {
	if m.SetHandlerFunc != nil {
		m.SetHandlerFunc(handler)
		return
	}
	m.Handler = handler
}

// GetAddr returns the server address.
func (m *MockNetHTTPServer) GetAddr() string {
	if m.GetAddrFunc != nil {
		return m.GetAddrFunc()
	}
	return m.Addr
}

// IsRunning returns whether the server is running.
func (m *MockNetHTTPServer) IsRunning() bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc()
	}
	return m.Running
}

// GetConfig returns the server configuration.
func (m *MockNetHTTPServer) GetConfig() *config.Config {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc()
	}
	return m.Config
}

// GracefulStop gracefully stops the server.
func (m *MockNetHTTPServer) GracefulStop(ctx context.Context, drainTimeout time.Duration) error {
	if m.GracefulStopFunc != nil {
		return m.GracefulStopFunc(ctx, drainTimeout)
	}

	// Execute pre-shutdown hooks
	for _, hook := range m.preHooks {
		hook()
	}

	m.Running = false

	// Execute post-shutdown hooks
	for _, hook := range m.postHooks {
		hook()
	}

	return nil
}

// Restart restarts the server.
func (m *MockNetHTTPServer) Restart(ctx context.Context) error {
	if m.RestartFunc != nil {
		return m.RestartFunc(ctx)
	}

	if err := m.GracefulStop(ctx, m.drainTimeout); err != nil {
		return err
	}

	return m.Start()
}

// GetConnectionsCount returns the number of active connections.
func (m *MockNetHTTPServer) GetConnectionsCount() int64 {
	if m.GetConnectionsCountFunc != nil {
		return m.GetConnectionsCountFunc()
	}
	return atomic.LoadInt64(&m.connections)
}

// GetHealthStatus returns the server health status.
func (m *MockNetHTTPServer) GetHealthStatus() interfaces.HealthStatus {
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
		Uptime:      time.Since(m.startTime),
		Connections: atomic.LoadInt64(&m.connections),
		Checks: map[string]interfaces.HealthCheck{
			"server": {
				Status:    status,
				Message:   "Server is " + status,
				Duration:  time.Millisecond,
				Timestamp: time.Now(),
			},
		},
	}
}

// PreShutdownHook adds a pre-shutdown hook.
func (m *MockNetHTTPServer) PreShutdownHook(hook func()) error {
	if m.PreShutdownHookFunc != nil {
		return m.PreShutdownHookFunc(hook)
	}
	m.preHooks = append(m.preHooks, hook)
	return nil
}

// PostShutdownHook adds a post-shutdown hook.
func (m *MockNetHTTPServer) PostShutdownHook(hook func()) error {
	if m.PostShutdownHookFunc != nil {
		return m.PostShutdownHookFunc(hook)
	}
	m.postHooks = append(m.postHooks, hook)
	return nil
}

// SetDrainTimeout sets the connection drain timeout.
func (m *MockNetHTTPServer) SetDrainTimeout(timeout time.Duration) {
	if m.SetDrainTimeoutFunc != nil {
		m.SetDrainTimeoutFunc(timeout)
		return
	}
	m.drainTimeout = timeout
}

// WaitForConnections waits for all connections to close.
func (m *MockNetHTTPServer) WaitForConnections(ctx context.Context) error {
	if m.WaitForConnectionsFunc != nil {
		return m.WaitForConnectionsFunc(ctx)
	}

	// Simulate waiting for connections to drain
	for atomic.LoadInt64(&m.connections) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}

	return nil
}

// Ensure MockNetHTTPServer implements interfaces.HTTPServer.
var _ interfaces.HTTPServer = (*MockNetHTTPServer)(nil)

// MockFactory is a factory function for creating mock net/http servers.
func MockFactory(cfg interface{}) (interfaces.HTTPServer, error) {
	conf, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	mock := NewMockNetHTTPServer()
	mock.Config = conf
	if conf != nil {
		mock.Addr = conf.Addr()
	}
	return mock, nil
}
