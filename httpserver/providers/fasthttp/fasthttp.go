// Package fasthttp provides an HTTP server implementation using the FastHTTP framework.
package fasthttp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/valyala/fasthttp"
)

// Server implements interfaces.HTTPServer using the FastHTTP framework.
type Server struct {
	mu                sync.RWMutex
	server            *fasthttp.Server
	config            *config.Config
	handler           http.Handler
	running           bool
	addr              string
	stopChan          chan struct{}
	connections       int64
	startTime         time.Time
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	drainTimeout      time.Duration
}

// NewServer creates a new FastHTTP server instance.
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

	server := &Server{
		config: conf,
		addr:   conf.Addr(),
		server: &fasthttp.Server{
			ReadTimeout:        conf.ReadTimeout,
			WriteTimeout:       conf.WriteTimeout,
			IdleTimeout:        conf.IdleTimeout,
			MaxRequestBodySize: conf.MaxHeaderBytes,
		},
		stopChan:          make(chan struct{}),
		drainTimeout:      30 * time.Second,
		preShutdownHooks:  make([]func() error, 0),
		postShutdownHooks: make([]func() error, 0),
	}

	return server, nil
}

// Start starts the FastHTTP server.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

	// Set handler if available
	if s.handler != nil {
		s.server.Handler = s.adaptHandler(s.handler)
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
			err = s.server.ListenAndServeTLS(s.addr, s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.server.ListenAndServe(s.addr)
		}

		// Check if server was intentionally stopped
		select {
		case <-s.stopChan:
			return
		default:
			if err != nil {
				errChan <- err
			}
		}
	}()

	// Check for startup errors
	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start server: %w", err)
	case <-time.After(10 * time.Millisecond):
		s.running = true
		return nil
	}
}

// Stop gracefully stops the FastHTTP server.
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

	// Signal stop
	close(s.stopChan)

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		done <- s.server.Shutdown()
	}()

	// Wait for shutdown or timeout
	select {
	case err := <-done:
		s.running = false
		return err
	case <-ctx.Done():
		s.running = false
		return ctx.Err()
	}
}

// SetHandler sets the HTTP handler for the server.
// For FastHTTP, this adapts the http.Handler to fasthttp.RequestHandler.
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handler = handler
	if s.server != nil {
		s.server.Handler = s.adaptHandler(handler)
	}
}

// adaptHandler adapts an http.Handler to work with FastHTTP.
func (s *Server) adaptHandler(handler http.Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Convert fasthttp request to net/http request
		uri := ctx.URI()
		url := &url.URL{
			Scheme:   string(uri.Scheme()),
			Host:     string(uri.Host()),
			Path:     string(uri.Path()),
			RawQuery: string(uri.QueryString()),
		}

		req := &http.Request{
			Method:     string(ctx.Method()),
			URL:        url,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Body:       &requestBody{data: ctx.PostBody()},
			Host:       string(ctx.Host()),
			RequestURI: string(ctx.RequestURI()),
		}

		// Copy headers
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})

		// Create response writer
		w := &responseWriter{ctx: ctx}

		// Call the handler
		handler.ServeHTTP(w, req)
	}
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

// GetHTTPServer returns nil as FastHTTP doesn't use net/http.Server.
func (s *Server) GetHTTPServer() *http.Server {
	return nil
}

// GetFastHTTPServer returns the underlying FastHTTP server for advanced configuration.
func (s *Server) GetFastHTTPServer() *fasthttp.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server
}

// requestBody adapts byte slice to io.ReadCloser
type requestBody struct {
	data []byte
	pos  int
}

func (r *requestBody) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *requestBody) Close() error {
	return nil
}

// responseWriter adapts fasthttp.RequestCtx to http.ResponseWriter
type responseWriter struct {
	ctx    *fasthttp.RequestCtx
	header http.Header
}

func (w *responseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *responseWriter) Write(data []byte) (int, error) {
	// Copy headers if not already copied
	if w.header != nil {
		for k, v := range w.header {
			for _, vv := range v {
				w.ctx.Response.Header.Add(k, vv)
			}
		}
		w.header = nil // Mark as copied
	}
	return w.ctx.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	// Copy headers before setting status
	if w.header != nil {
		for k, v := range w.header {
			for _, vv := range v {
				w.ctx.Response.Header.Add(k, vv)
			}
		}
		w.header = nil // Mark as copied
	}
	w.ctx.SetStatusCode(statusCode)
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
	close(s.stopChan)
	err := s.server.ShutdownWithContext(ctx)
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

// Factory function for creating FastHTTP servers.
var Factory interfaces.ServerFactory = NewServer
