package echo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/echo/middleware"
)

// EchoServer implements the common.Server interface for Echo
type EchoServer struct {
	echo   *echo.Echo
	config *common.ServerConfig
}

// NewServer creates a new Echo server
func NewServer(options ...common.ServerOption) *EchoServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	s := &EchoServer{
		echo:   e,
		config: config,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware stack
func (s *EchoServer) setupMiddleware() {
	// Use Echo's recovery middleware
	s.echo.Use(echoMiddleware.Recover())

	// Add tracing if enabled
	if s.config.EnableTracing {
		s.echo.Use(middleware.Tracing)
	}

	// Add logging if enabled
	if s.config.EnableLogging {
		s.echo.Use(middleware.Logger)
	}

	// Add CORS middleware
	s.echo.Use(middleware.CORS)

	// Add custom error handler
	s.echo.HTTPErrorHandler = middleware.ErrorHandler

	// Handler para rotas não encontradas (404 Not Found)
	s.echo.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "Route not found",
		})
	})
}

// setupRoutes configures the routes
func (s *EchoServer) setupRoutes() {
	// Health check endpoints
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	s.echo.GET("/readyz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Add pprof if enabled
	if s.config.EnablePprof {
		middleware.RegisterPprof(s.echo)
	}

	// Add metrics if enabled
	if s.config.EnableMetrics {
		middleware.RegisterMetrics(s.echo)
	}

	// Add swagger if enabled
	if s.config.EnableSwagger {
		middleware.RegisterSwagger(s.echo)
	}
}

// Start starts the server
func (s *EchoServer) Start() error {
	// Configure server
	server := &http.Server{
		Addr:         s.Address(),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Starting Echo server on %s\n", s.Address())
		if err := s.echo.StartServer(server); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.Shutdown(ctx)
}

// Shutdown gracefully shuts down the server
func (s *EchoServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down Echo server...")
	return s.echo.Shutdown(ctx)
}

// Address returns the server's address
func (s *EchoServer) Address() string {
	return s.config.Host + ":" + s.config.Port
}

// Echo returns the underlying Echo instance
func (s *EchoServer) Echo() *echo.Echo {
	return s.echo
}

// Health returns a handler for health checks
func (s *EchoServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
