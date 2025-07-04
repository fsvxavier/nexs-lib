package atreugo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/savsgio/atreugo/v11"

	"github.com/fsvxavier/nexs-lib/httpservers/atreugo/middleware"
	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// AtreugoServer implements the common.Server interface for Atreugo
type AtreugoServer struct {
	server *atreugo.Atreugo
	config *common.ServerConfig
}

// NewServer creates a new Atreugo server
func NewServer(options ...common.ServerOption) *AtreugoServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	// Configure Atreugo
	atreugoConfig := atreugo.Config{
		Addr:               config.Host + ":" + config.Port,
		Name:               "AtreugoServer",
		Compress:           true,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		IdleTimeout:        config.IdleTimeout,
		Logger:             nil, // Usa o logger default
		GracefulShutdown:   true,
		MaxRequestBodySize: config.MaxHeaderBytes,
		// NotFoundView é configurado diretamente na configuração
		NotFoundView: func(ctx *atreugo.RequestCtx) error {
			return ctx.JSONResponse(map[string]interface{}{
				"error": "Route not found",
			}, 404)
		},
	}

	server := atreugo.New(atreugoConfig)

	s := &AtreugoServer{
		server: server,
		config: config,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware stack
func (s *AtreugoServer) setupMiddleware() {
	// Recovery middleware
	s.server.UseBefore(middleware.Recover)

	// Add CORS middleware
	s.server.UseBefore(middleware.CORS)

	// Add tracing if enabled
	if s.config.EnableTracing {
		s.server.UseBefore(middleware.Tracing)
	}

	// Add logging if enabled
	if s.config.EnableLogging {
		s.server.UseBefore(middleware.Logger)
	}
}

// setupRoutes configures the routes
func (s *AtreugoServer) setupRoutes() {
	// Health check endpoints
	s.server.GET("/health", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		}, 200)
	})

	s.server.GET("/readyz", func(ctx *atreugo.RequestCtx) error {
		return ctx.JSONResponse(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		}, 200)
	})

	// Add metrics if enabled
	if s.config.EnableMetrics {
		s.server.GET("/metrics", func(ctx *atreugo.RequestCtx) error {
			return ctx.TextResponse("Metrics endpoint enabled", 200)
		})
	}

	// Add pprof if enabled
	if s.config.EnablePprof {
		// In a real implementation, this would use pprof endpoints
		s.server.GET("/debug/pprof", func(ctx *atreugo.RequestCtx) error {
			return ctx.TextResponse("Pprof endpoint enabled", 200)
		})
	}

	// Add swagger if enabled
	if s.config.EnableSwagger {
		s.server.GET("/swagger/*", func(ctx *atreugo.RequestCtx) error {
			return ctx.TextResponse("Swagger endpoint enabled", 200)
		})
	}

	// O Atreugo não possui método SetNotFoundHandler,
	// NotFoundView deve ser configurado diretamente na configuração
}

// Start starts the server
func (s *AtreugoServer) Start() error {
	// Start the server in a goroutine
	go func() {
		fmt.Printf("Starting Atreugo server on %s\n", s.Address())
		if err := s.server.ListenAndServe(); err != nil {
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
func (s *AtreugoServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down Atreugo server...")
	return s.server.ShutdownWithContext(ctx)
}

// Address returns the server's address
func (s *AtreugoServer) Address() string {
	return s.config.Host + ":" + s.config.Port
}

// Server returns the underlying Atreugo instance
func (s *AtreugoServer) Server() *atreugo.Atreugo {
	return s.server
}

// Health returns a handler for health checks
func (s *AtreugoServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
