// Package fiber provides an HTTP server implementation using the Fiber web framework.
package fiber

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Server implements interfaces.HTTPServer using the Fiber framework.
type Server struct {
	mu                sync.RWMutex
	app               *fiber.App
	config            *config.Config
	handler           http.Handler
	running           bool
	connections       int64
	startTime         time.Time
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	drainTimeout      time.Duration
}

// NewServer creates a new Fiber server instance.
func NewServer(cfg interface{}) (interfaces.HTTPServer, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	conf, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type, expected *config.Config")
	}

	// Create Fiber configuration
	fiberConfig := fiber.Config{
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
		Prefork:      false, // Usually disabled for better compatibility
	}

	// Create Fiber app
	app := fiber.New(fiberConfig)

	// Add default middleware
	app.Use(recover.New())
	app.Use(logger.New())

	server := &Server{
		config:            conf,
		app:               app,
		drainTimeout:      30 * time.Second,
		preShutdownHooks:  make([]func() error, 0),
		postShutdownHooks: make([]func() error, 0),
	}

	return server, nil
}

// Start starts the Fiber HTTP server.
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
		var err error
		if s.config.TLSEnabled {
			if s.config.CertFile == "" || s.config.KeyFile == "" {
				errChan <- fmt.Errorf("TLS enabled but cert or key file not specified")
				return
			}
			err = s.app.ListenTLS(s.config.Addr(), s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.app.Listen(s.config.Addr())
		}

		if err != nil {
			errChan <- err
		}
	}()

	// Give the server a moment to start
	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start server: %w", err)
	default:
		s.running = true
		return nil
	}
}

// Stop gracefully stops the Fiber HTTP server.
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
// For Fiber, this adapts the handler to work with Fiber's context.
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handler = handler
	if s.app != nil {
		s.setupHandler()
	}
}

// setupHandler configures the Fiber app to use the http.Handler.
func (s *Server) setupHandler() {
	// Clear existing routes by creating a new app
	fiberConfig := fiber.Config{
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
		Prefork:      false,
	}

	s.app = fiber.New(fiberConfig)
	s.app.Use(recover.New())
	s.app.Use(logger.New())

	// Add catch-all route to handle all requests with the provided handler
	s.app.All("*", adaptor.HTTPHandler(s.handler))
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Addr()
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

// GetHTTPServer returns nil as Fiber uses its own server implementation.
func (s *Server) GetHTTPServer() *http.Server {
	return nil
}

// GetFiberApp returns the underlying Fiber application for advanced configuration.
func (s *Server) GetFiberApp() *fiber.App {
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
	err := s.app.ShutdownWithTimeout(drainTimeout)
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

// Factory function for creating Fiber servers.
var Factory interfaces.ServerFactory = NewServer
