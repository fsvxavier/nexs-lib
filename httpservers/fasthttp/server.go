package fasthttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/fasthttp/middleware"
)

// FastHTTPServer implements the common.Server interface for FastHTTP
type FastHTTPServer struct {
	server *fasthttp.Server
	config *common.ServerConfig
	router *fasthttp.RequestHandler
	ln     net.Listener
}

// NewServer creates a new FastHTTP server
func NewServer(options ...common.ServerOption) *FastHTTPServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	s := &FastHTTPServer{
		config: config,
	}

	// Create request handler with middleware
	var handler fasthttp.RequestHandler = s.defaultHandler

	// Add middlewares (in reverse order)
	if config.EnableLogging {
		handler = middleware.Logger(handler)
	}

	if config.EnableTracing {
		handler = middleware.Tracing(handler)
	}

	// Add recovery middleware (always first)
	handler = middleware.Recover(handler)

	s.router = &handler

	// Configure server
	s.server = &fasthttp.Server{
		Handler:            *s.router,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		IdleTimeout:        config.IdleTimeout,
		MaxRequestBodySize: config.MaxHeaderBytes,
		Name:               "FastHTTPServer",
	}

	return s
}

// defaultHandler is the default request handler
func (s *FastHTTPServer) defaultHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())

	switch {
	case path == "/health" || path == "/readyz":
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.WriteString(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`)
		return
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.WriteString(`{"error":"Route not found"}`)
	}
}

// SetHandler sets a custom request handler
func (s *FastHTTPServer) SetHandler(handler fasthttp.RequestHandler) {
	s.router = &handler
	s.server.Handler = handler
}

// Start starts the server
func (s *FastHTTPServer) Start() error {
	// Create listener
	addr := s.Address()
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	s.ln = ln

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Starting FastHTTP server on %s\n", addr)
		if err := s.server.Serve(ln); err != nil {
			fmt.Printf("Error in FastHTTP server: %v\n", err)
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
func (s *FastHTTPServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down FastHTTP server...")

	// Create a channel to signal shutdown completion
	done := make(chan struct{})
	go func() {
		s.server.Shutdown()
		close(done)
	}()

	// Wait for shutdown or timeout
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Address returns the server's address
func (s *FastHTTPServer) Address() string {
	return s.config.Host + ":" + s.config.Port
}

// Health returns a handler for health checks
func (s *FastHTTPServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
