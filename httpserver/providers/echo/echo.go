// Package echo provides an HTTP server implementation using the Echo web framework.
package echo

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server implements interfaces.HTTPServer using the Echo framework.
type Server struct {
	mu                sync.RWMutex
	echo              *echo.Echo
	config            *config.Config
	running           bool
	connections       int64
	startTime         time.Time
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	drainTimeout      time.Duration
}

// NewServer creates a new Echo server instance.
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

	// Create Echo instance
	e := echo.New()

	// Disable Echo's default banner
	e.HideBanner = true
	e.HidePort = true

	// Add default middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Configure Echo server
	e.Server.Addr = conf.Addr()
	e.Server.ReadTimeout = conf.ReadTimeout
	e.Server.WriteTimeout = conf.WriteTimeout
	e.Server.IdleTimeout = conf.IdleTimeout
	e.Server.MaxHeaderBytes = conf.MaxHeaderBytes

	server := &Server{
		config:            conf,
		echo:              e,
		drainTimeout:      30 * time.Second,
		preShutdownHooks:  make([]func() error, 0),
		postShutdownHooks: make([]func() error, 0),
	}

	return server, nil
}

// Start starts the Echo HTTP server.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

	// Channel to capture startup errors
	errChan := make(chan error, 1)

	go func() {
		var err error
		if s.config.TLSEnabled {
			if s.config.CertFile == "" || s.config.KeyFile == "" {
				errChan <- fmt.Errorf("TLS enabled but cert or key file not specified")
				return
			}
			err = s.echo.StartTLS(s.config.Addr(), s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.echo.Start(s.config.Addr())
		}

		if err != nil && err != http.ErrServerClosed {
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

// Stop gracefully stops the Echo HTTP server.
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

	err := s.echo.Shutdown(ctx)
	s.running = false
	return err
}

// SetHandler sets the HTTP handler for the server.
// For Echo, this wraps the handler as a catch-all route.
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove all existing routes
	s.echo.Routes()

	// Create a new Echo instance to clear routes
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Configure Echo server
	e.Server.Addr = s.config.Addr()
	e.Server.ReadTimeout = s.config.ReadTimeout
	e.Server.WriteTimeout = s.config.WriteTimeout
	e.Server.IdleTimeout = s.config.IdleTimeout
	e.Server.MaxHeaderBytes = s.config.MaxHeaderBytes

	// Add the handler as a catch-all route
	e.Any("/*", func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	s.echo = e
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.echo.Server.Addr
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

// GetHTTPServer returns the underlying http.Server.
func (s *Server) GetHTTPServer() *http.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.echo.Server
}

// GetEcho returns the underlying Echo instance for advanced configuration.
func (s *Server) GetEcho() *echo.Echo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.echo
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
	shutdownCtx, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	err := s.echo.Shutdown(shutdownCtx)
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

// Factory function for creating Echo servers.
var Factory interfaces.ServerFactory = NewServer
