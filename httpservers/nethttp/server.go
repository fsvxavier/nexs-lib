package nethttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/nethttp/middleware"
)

// NetHTTPServer implements the common.Server interface using the standard net/http package
type NetHTTPServer struct {
	server *http.Server
	config *common.ServerConfig
	router *http.ServeMux
}

// NewServer creates a new server based on net/http
func NewServer(options ...common.ServerOption) *NetHTTPServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	router := http.NewServeMux()

	// Create server
	server := &http.Server{
		Addr:           config.Host + ":" + config.Port,
		Handler:        router,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	s := &NetHTTPServer{
		server: server,
		config: config,
		router: router,
	}

	// Setup default routes
	s.setupRoutes()

	// Apply middlewares
	var handler http.Handler = router

	// Add logging middleware
	if config.EnableLogging {
		handler = middleware.Logger(handler)
	}

	// Add tracing middleware
	if config.EnableTracing {
		handler = middleware.Tracing(handler)
	}

	// Add recovery middleware (always first)
	handler = middleware.Recover(handler)

	// Set the final handler
	s.server.Handler = handler

	return s
}

// setupRoutes sets up the default routes
func (s *NetHTTPServer) setupRoutes() {
	// Health check endpoints
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	s.router.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add pprof endpoints if enabled
	if s.config.EnablePprof {
		middleware.RegisterPprof(s.router)
	}

	// Add metrics endpoints if enabled
	if s.config.EnableMetrics {
		middleware.RegisterMetrics(s.router)
	}

	// Add a 404 handler for any other route
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Route not found"}`))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})
}

// Router returns the underlying router
func (s *NetHTTPServer) Router() *http.ServeMux {
	return s.router
}

// Start starts the server
func (s *NetHTTPServer) Start() error {
	// Start server in a goroutine
	go func() {
		addr := s.Address()
		fmt.Printf("Starting HTTP server on %s\n", addr)

		// Create a custom listener
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("Error creating listener: %v\n", err)
			return
		}

		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
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
func (s *NetHTTPServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// Address returns the server's address
func (s *NetHTTPServer) Address() string {
	return s.server.Addr
}

// Health returns a handler for health checks
func (s *NetHTTPServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
