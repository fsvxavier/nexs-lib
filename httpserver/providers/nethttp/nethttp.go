// Package nethttp provides HTTP server implementation using Go's standard net/http package.
// It serves as a reference implementation demonstrating the HTTP server interface.
package nethttp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Factory implements the ProviderFactory interface for net/http servers.
type Factory struct{}

// Create creates a new net/http server instance.
func (f *Factory) Create(cfg interface{}) (interfaces.HTTPServer, error) {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	server := &Server{
		config:     baseConfig,
		mux:        http.NewServeMux(),
		observers:  hooks.NewObserverManager(),
		middleware: make([]interface{}, 0),
		routes:     make(map[string]map[string]http.HandlerFunc),
		stats: &serverStats{
			provider: "nethttp",
		},
		mutex: sync.RWMutex{},
	}

	// Initialize route map
	server.routes["GET"] = make(map[string]http.HandlerFunc)
	server.routes["POST"] = make(map[string]http.HandlerFunc)
	server.routes["PUT"] = make(map[string]http.HandlerFunc)
	server.routes["DELETE"] = make(map[string]http.HandlerFunc)
	server.routes["PATCH"] = make(map[string]http.HandlerFunc)
	server.routes["OPTIONS"] = make(map[string]http.HandlerFunc)
	server.routes["HEAD"] = make(map[string]http.HandlerFunc)

	// Attach configured observers
	for _, observer := range baseConfig.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	// Register configured middlewares
	for _, middleware := range baseConfig.GetMiddlewares() {
		if err := server.RegisterMiddleware(middleware); err != nil {
			return nil, fmt.Errorf("failed to register middleware: %w", err)
		}
	}

	return server, nil
}

// GetName returns the provider name.
func (f *Factory) GetName() string {
	return "nethttp"
}

// GetDefaultConfig returns the default configuration for net/http.
func (f *Factory) GetDefaultConfig() interface{} {
	return config.NewBaseConfig()
}

// ValidateConfig validates the configuration for net/http.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}
	return baseConfig.Validate()
}

// Server implements the HTTPServer interface using Go's standard net/http.
type Server struct {
	config     *config.BaseConfig
	server     *http.Server
	mux        *http.ServeMux
	observers  *hooks.ObserverManager
	middleware []interface{}
	routes     map[string]map[string]http.HandlerFunc
	stats      *serverStats
	running    int32
	mutex      sync.RWMutex
}

// serverStats tracks server runtime statistics.
type serverStats struct {
	startTime         time.Time
	requestCount      int64
	errorCount        int64
	totalResponseTime int64
	activeConnections int64
	provider          string
	mutex             sync.RWMutex
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 1 {
		return fmt.Errorf("server is already running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.config.GetFullAddr(),
		Handler:      s.buildHandler(),
		ReadTimeout:  s.config.GetReadTimeout(),
		WriteTimeout: s.config.GetWriteTimeout(),
		IdleTimeout:  s.config.GetIdleTimeout(),
	}

	// Update stats
	s.stats.mutex.Lock()
	s.stats.startTime = time.Now()
	s.stats.mutex.Unlock()

	// Notify observers
	if err := s.observers.NotifyObservers(interfaces.EventStart, ctx, s.server.Addr); err != nil {
		return fmt.Errorf("observer notification failed: %w", err)
	}

	atomic.StoreInt32(&s.running, 1)

	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			atomic.StoreInt32(&s.running, 0)
			s.observers.NotifyObservers(interfaces.EventError, ctx, err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 0 {
		return fmt.Errorf("server is not running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	// Create shutdown context
	shutdownCtx := ctx
	if s.config.IsGracefulShutdown() {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(ctx, s.config.GetShutdownTimeout())
		defer cancel()
	}

	// Shutdown server
	err := s.server.Shutdown(shutdownCtx)
	atomic.StoreInt32(&s.running, 0)

	// Notify observers
	if notifyErr := s.observers.NotifyObservers(interfaces.EventStop, ctx, nil); notifyErr != nil {
		if err == nil {
			err = fmt.Errorf("observer notification failed: %w", notifyErr)
		}
	}

	return err
}

// RegisterRoute registers a new route with the server.
// The handler parameter should be an http.HandlerFunc function.
func (s *Server) RegisterRoute(method, path string, handler interface{}) error {
	if method == "" {
		return fmt.Errorf("method cannot be empty")
	}
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Convert handler to net/http handler
	httpHandler, ok := handler.(http.HandlerFunc)
	if !ok {
		return fmt.Errorf("handler must be an http.HandlerFunc, got %T", handler)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate method
	methodMap, exists := s.routes[method]
	if !exists {
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Check if route already exists
	if _, exists := methodMap[path]; exists {
		return fmt.Errorf("route %s %s already registered", method, path)
	}

	// Wrap handler with middleware and observability
	wrappedHandler := s.wrapHandler(method, path, httpHandler)

	// Register route
	methodMap[path] = wrappedHandler
	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			wrappedHandler(w, r)
			return
		}
		// Method not allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return nil
}

// RegisterMiddleware registers middleware with the server.
func (s *Server) RegisterMiddleware(middleware interface{}) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, middleware)
	return nil
}

// AttachObserver attaches an observer to the server.
func (s *Server) AttachObserver(observer interfaces.ServerObserver) error {
	return s.observers.AttachObserver(observer)
}

// DetachObserver detaches an observer from the server.
func (s *Server) DetachObserver(observer interfaces.ServerObserver) error {
	return s.observers.DetachObserver(observer)
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	if s.server != nil {
		return s.server.Addr
	}
	return s.config.GetFullAddr()
}

// IsRunning returns true if the server is running.
func (s *Server) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// GetStats returns server statistics.
func (s *Server) GetStats() interfaces.ServerStats {
	s.stats.mutex.RLock()
	defer s.stats.mutex.RUnlock()

	var avgResponseTime time.Duration
	if s.stats.requestCount > 0 {
		avgResponseTime = time.Duration(s.stats.totalResponseTime / s.stats.requestCount)
	}

	return interfaces.ServerStats{
		StartTime:           s.stats.startTime,
		RequestCount:        s.stats.requestCount,
		ErrorCount:          s.stats.errorCount,
		AverageResponseTime: avgResponseTime,
		ActiveConnections:   s.stats.activeConnections,
		Provider:            s.stats.provider,
	}
}

// buildHandler builds the main HTTP handler with middleware chain.
func (s *Server) buildHandler() http.Handler {
	handler := http.Handler(s.mux)

	// Apply middleware in reverse order (last registered, first executed)
	for i := len(s.middleware) - 1; i >= 0; i-- {
		middleware := s.middleware[i]
		// Type assert to MiddlewareFunc for net/http provider
		if middlewareFunc, ok := middleware.(func(http.HandlerFunc) http.HandlerFunc); ok {
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareFunc(handler.ServeHTTP)(w, r)
			})
		}
	}

	return handler
}

// wrapHandler wraps a route handler with observability and statistics.
func (s *Server) wrapHandler(method, path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Update active connections
		atomic.AddInt64(&s.stats.activeConnections, 1)
		defer atomic.AddInt64(&s.stats.activeConnections, -1)

		// Increment request count
		atomic.AddInt64(&s.stats.requestCount, 1)

		// Notify route enter
		ctx := r.Context()
		routeData := hooks.RouteEventData{
			Method:  method,
			Path:    path,
			Request: r,
		}
		s.observers.NotifyObservers(interfaces.EventRouteEnter, ctx, routeData)

		// Notify request
		s.observers.NotifyObservers(interfaces.EventRequest, ctx, r)

		// Create response wrapper
		rw := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute handler
		func() {
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt64(&s.stats.errorCount, 1)
					err := fmt.Errorf("panic in handler: %v", r)
					s.observers.NotifyObservers(interfaces.EventError, ctx, err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			handler(rw, r)
		}()

		duration := time.Since(startTime)

		// Update response time stats
		atomic.AddInt64(&s.stats.totalResponseTime, int64(duration))

		// Notify response
		respData := hooks.ResponseEventData{
			Request:  r,
			Response: &http.Response{StatusCode: rw.statusCode},
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventResponse, ctx, respData)

		// Notify route exit
		routeExitData := hooks.RouteExitEventData{
			Method:   method,
			Path:     path,
			Request:  r,
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventRouteExit, ctx, routeExitData)
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code.
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewFactory creates a new net/http factory.
func NewFactory() *Factory {
	return &Factory{}
}
