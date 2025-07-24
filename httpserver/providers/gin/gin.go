// Package gin provides an HTTP server implementation using the Gin web framework.
package gin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/gin-gonic/gin"
)

// Server implements interfaces.HTTPServer using the Gin framework.
type Server struct {
	mu                sync.RWMutex
	engine            *gin.Engine
	server            *http.Server
	config            *config.Config
	running           bool
	connections       int64
	startTime         time.Time
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	drainTimeout      time.Duration
}

// NewServer creates a new Gin server instance.
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

	// Set Gin mode based on configuration
	gin.SetMode(gin.ReleaseMode)

	// Create Gin engine
	engine := gin.New()

	// Add default middleware
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	server := &Server{
		config: conf,
		engine: engine,
		server: &http.Server{
			Addr:           conf.Addr(),
			Handler:        engine,
			ReadTimeout:    conf.ReadTimeout,
			WriteTimeout:   conf.WriteTimeout,
			IdleTimeout:    conf.IdleTimeout,
			MaxHeaderBytes: conf.MaxHeaderBytes,
		},
		drainTimeout:      30 * time.Second,
		preShutdownHooks:  make([]func() error, 0),
		postShutdownHooks: make([]func() error, 0),
	}

	return server, nil
}

// Start starts the Gin HTTP server.
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
			err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.server.ListenAndServe()
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

// Stop gracefully stops the Gin HTTP server.
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

	err := s.server.Shutdown(ctx)
	s.running = false
	return err
}

// SetHandler sets the HTTP handler for the server.
// For Gin, this wraps the handler in the Gin engine.
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear all routes first
	s.engine = gin.New()
	s.engine.Use(gin.Recovery())
	s.engine.Use(gin.Logger())

	// Wrap the external handler to work with Gin
	s.engine.NoRoute(func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	s.server.Handler = s.engine
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server.Addr
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
	return s.server
}

// GetGinEngine returns the underlying Gin engine for advanced configuration.
func (s *Server) GetGinEngine() *gin.Engine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine
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

	err := s.server.Shutdown(shutdownCtx)
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

// Factory function for creating Gin servers.
var Factory interfaces.ServerFactory = NewServer
