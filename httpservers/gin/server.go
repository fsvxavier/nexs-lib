package gin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/gin/middleware"
)

// GinServer implements the common.Server interface for Gin
type GinServer struct {
	router *gin.Engine
	server *http.Server
	config *common.ServerConfig
}

// NewServer creates a new Gin server
func NewServer(options ...common.ServerOption) *GinServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	// Set gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	s := &GinServer{
		router: router,
		config: config,
		server: &http.Server{
			Addr:           config.Host + ":" + config.Port,
			Handler:        router,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			IdleTimeout:    config.IdleTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes,
		},
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware stack
func (s *GinServer) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Add tracing if enabled
	if s.config.EnableTracing {
		s.router.Use(middleware.Tracing())
	}

	// Add logging if enabled
	if s.config.EnableLogging {
		s.router.Use(middleware.Logger())
	}

	// Add CORS middleware
	s.router.Use(middleware.CORS())

	// Add custom error handler
	s.router.Use(middleware.ErrorHandler())
}

// setupRoutes configures the routes
func (s *GinServer) setupRoutes() {
	// Health check endpoints
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	s.router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Add pprof if enabled
	if s.config.EnablePprof {
		middleware.RegisterPprof(s.router)
	}

	// Add metrics if enabled
	if s.config.EnableMetrics {
		middleware.RegisterMetrics(s.router)
	}

	// Add swagger if enabled
	if s.config.EnableSwagger {
		middleware.RegisterSwagger(s.router)
	}

	// Add a 404 handler
	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
		})
	})
}

// Start starts the server
func (s *GinServer) Start() error {
	// Start the server in a goroutine
	go func() {
		addr := s.Address()
		fmt.Printf("Starting Gin server on %s\n", addr)

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
func (s *GinServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down Gin server...")
	return s.server.Shutdown(ctx)
}

// Address returns the server's address
func (s *GinServer) Address() string {
	return s.server.Addr
}

// Router returns the underlying Gin router
func (s *GinServer) Router() *gin.Engine {
	return s.router
}

// Health returns a handler for health checks
func (s *GinServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
