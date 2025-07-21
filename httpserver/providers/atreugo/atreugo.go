// Package atreugo provides an HTTP server implementation using the Atreugo web framework.
package atreugo

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/savsgio/atreugo/v11"
)

// Server implements interfaces.HTTPServer using the Atreugo framework.
type Server struct {
	mu                sync.RWMutex
	app               *atreugo.Atreugo
	config            *config.Config
	handler           http.Handler
	running           bool
	addr              string
	connections       int64
	startTime         time.Time
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	drainTimeout      time.Duration
}

// NewServer creates a new Atreugo server instance.
func NewServer(cfg interface{}) (interfaces.HTTPServer, error) {
	var conf *config.Config

	if cfg == nil {
		conf = config.DefaultConfig()
	} else {
		var ok bool
		conf, ok = cfg.(*config.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type, expected *config.Config")
		}
		if conf == nil {
			conf = config.DefaultConfig()
		}
	}

	// Create Atreugo configuration
	atreugoConfig := atreugo.Config{
		Addr:         conf.Addr(),
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	}

	// Configure TLS if enabled
	if conf.TLSEnabled {
		atreugoConfig.TLSEnable = true
		atreugoConfig.CertFile = conf.CertFile
		atreugoConfig.CertKey = conf.KeyFile
	}

	// Create Atreugo app
	app := atreugo.New(atreugoConfig)

	server := &Server{
		config:            conf,
		app:               app,
		addr:              conf.Addr(),
		drainTimeout:      30 * time.Second,
		preShutdownHooks:  make([]func() error, 0),
		postShutdownHooks: make([]func() error, 0),
	}

	return server, nil
}

// Start starts the Atreugo HTTP server.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

	// Set handler if available
	if s.handler != nil {
		s.setupHandler()
	}

	// Channel to capture startup errors
	errChan := make(chan error, 1)

	go func() {
		err := s.app.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	// Check for startup errors
	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start server: %w", err)
	default:
		s.running = true
		return nil
	}
}

// Stop gracefully stops the Atreugo HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.config.GracefulTimeout)
		defer cancel()
	}

	err := s.app.ShutdownWithContext(ctx)
	s.running = false
	return err
}

// SetHandler sets the HTTP handler for the server.
// For Atreugo, this adapts the handler to work with Atreugo's context.
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handler = handler
	if s.app != nil {
		s.setupHandler()
	}
}

// setupHandler configures the Atreugo app to use the http.Handler.
func (s *Server) setupHandler() {
	// Clear existing routes by creating a new app
	atreugoConfig := atreugo.Config{
		Addr:         s.config.Addr(),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	if s.config.TLSEnabled {
		atreugoConfig.TLSEnable = true
		atreugoConfig.CertFile = s.config.CertFile
		atreugoConfig.CertKey = s.config.KeyFile
	}

	s.app = atreugo.New(atreugoConfig)

	// Add catch-all route to handle all requests with the provided handler
	s.app.NetHTTPPath("*", "/{path:*}", s.handler)
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

// IsRunning returns true if the server is currently running.
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetConfig returns the server configuration.
func (s *Server) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetHTTPServer returns nil as Atreugo uses FastHTTP underneath.
func (s *Server) GetHTTPServer() *http.Server {
	return nil
}

// GetAtreugoApp returns the underlying Atreugo application for advanced configuration.
func (s *Server) GetAtreugoApp() *atreugo.Atreugo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.app
}

// GracefulStop performs a graceful shutdown with connection draining.
func (s *Server) GracefulStop(ctx context.Context, drainTimeout time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server is not running")
	}

	// Execute pre-shutdown hooks
	for _, hook := range s.preShutdownHooks {
		if err := hook(); err != nil {
			return fmt.Errorf("pre-shutdown hook failed: %v", err)
		}
	}

	// Stop accepting new connections and drain existing ones
	err := s.app.ShutdownWithContext(ctx)
	s.running = false

	// Execute post-shutdown hooks
	for _, hook := range s.postShutdownHooks {
		if hookErr := hook(); hookErr != nil {
			if err == nil {
				err = fmt.Errorf("post-shutdown hook failed: %v", hookErr)
			}
		}
	}

	return err
}

// Restart performs a zero-downtime restart.
func (s *Server) Restart(ctx context.Context) error {
	// For now, just perform a graceful stop and start
	// In a real implementation, this could involve more sophisticated logic
	if err := s.GracefulStop(ctx, s.drainTimeout); err != nil {
		return fmt.Errorf("restart failed during stop: %v", err)
	}

	return s.Start()
}

// GetConnectionsCount returns the number of active connections.
func (s *Server) GetConnectionsCount() int64 {
	return atomic.LoadInt64(&s.connections)
}

// GetHealthStatus returns the current health status.
func (s *Server) GetHealthStatus() interfaces.HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uptime := time.Duration(0)
	if !s.startTime.IsZero() {
		uptime = time.Since(s.startTime)
	}

	status := "healthy"
	if !s.running {
		status = "stopped"
	}

	return interfaces.HealthStatus{
		Status:      status,
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      uptime,
		Connections: s.GetConnectionsCount(),
		Checks:      make(map[string]interfaces.HealthCheck),
	}
}

// PreShutdownHook registers a function to be called before shutdown.
func (s *Server) PreShutdownHook(hook func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preShutdownHooks = append(s.preShutdownHooks, hook)
}

// PostShutdownHook registers a function to be called after shutdown.
func (s *Server) PostShutdownHook(hook func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.postShutdownHooks = append(s.postShutdownHooks, hook)
}

// SetDrainTimeout sets the timeout for connection draining.
func (s *Server) SetDrainTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drainTimeout = timeout
}

// WaitForConnections waits for all connections to finish or timeout.
func (s *Server) WaitForConnections(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if s.GetConnectionsCount() == 0 {
				return nil
			}
		}
	}
}

// Ensure Server implements interfaces.HTTPServer.
var _ interfaces.HTTPServer = (*Server)(nil)

// Factory function for creating Atreugo servers.
var Factory interfaces.ServerFactory = NewServer
