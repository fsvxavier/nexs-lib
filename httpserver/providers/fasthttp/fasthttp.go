// Package fasthttp provides HTTP server implementation using FastHTTP framework.
// FastHTTP is known for its high performance and low memory allocation.
package fasthttp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Factory implements the ProviderFactory interface for FastHTTP servers.
type Factory struct{}

// Create creates a new FastHTTP server instance.
func (f *Factory) Create(cfg interface{}) (interfaces.HTTPServer, error) {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	server := &Server{
		config:     baseConfig,
		observers:  hooks.NewObserverManager(),
		middleware: make([]fasthttp.RequestHandler, 0),
		routes:     make(map[string]map[string]fasthttp.RequestHandler),
		stats: &serverStats{
			provider: "fasthttp",
		},
		mutex: sync.RWMutex{},
	}

	// Initialize route map
	server.routes["GET"] = make(map[string]fasthttp.RequestHandler)
	server.routes["POST"] = make(map[string]fasthttp.RequestHandler)
	server.routes["PUT"] = make(map[string]fasthttp.RequestHandler)
	server.routes["DELETE"] = make(map[string]fasthttp.RequestHandler)
	server.routes["PATCH"] = make(map[string]fasthttp.RequestHandler)
	server.routes["OPTIONS"] = make(map[string]fasthttp.RequestHandler)
	server.routes["HEAD"] = make(map[string]fasthttp.RequestHandler)

	// Attach configured observers
	for _, observer := range baseConfig.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	// Register configured middlewares
	for _, middleware := range baseConfig.GetMiddlewares() {
		// Skip middleware registration here as they need to be FastHTTP-specific
		_ = middleware
	}

	return server, nil
}

// GetName returns the provider name.
func (f *Factory) GetName() string {
	return "fasthttp"
}

// GetDefaultConfig returns the default configuration for FastHTTP.
func (f *Factory) GetDefaultConfig() interface{} {
	return config.NewBaseConfig()
}

// ValidateConfig validates the configuration for FastHTTP.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}
	return baseConfig.Validate()
}

// Server implements the HTTPServer interface using FastHTTP.
type Server struct {
	config     *config.BaseConfig
	server     *fasthttp.Server
	observers  *hooks.ObserverManager
	middleware []fasthttp.RequestHandler
	routes     map[string]map[string]fasthttp.RequestHandler
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

// Start starts the FastHTTP server.
func (s *Server) Start(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 1 {
		return fmt.Errorf("server is already running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create FastHTTP server
	s.server = &fasthttp.Server{
		Handler:      s.buildHandler(),
		ReadTimeout:  s.config.GetReadTimeout(),
		WriteTimeout: s.config.GetWriteTimeout(),
		IdleTimeout:  s.config.GetIdleTimeout(),
		Name:         "fasthttp-nexs-lib",
	}

	// Update stats
	s.stats.mutex.Lock()
	s.stats.startTime = time.Now()
	s.stats.mutex.Unlock()

	// Notify observers
	if err := s.observers.NotifyObservers(interfaces.EventStart, ctx, s.config.GetFullAddr()); err != nil {
		return fmt.Errorf("observer notification failed: %w", err)
	}

	atomic.StoreInt32(&s.running, 1)

	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(s.config.GetFullAddr()); err != nil {
			atomic.StoreInt32(&s.running, 0)
			s.observers.NotifyObservers(interfaces.EventError, ctx, err)
		}
	}()

	return nil
}

// Stop gracefully stops the FastHTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 0 {
		return fmt.Errorf("server is not running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	atomic.StoreInt32(&s.running, 0)

	// Notify observers
	if err := s.observers.NotifyObservers(interfaces.EventStop, ctx, nil); err != nil {
		// Log error but continue with shutdown
		_ = err
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.GetShutdownTimeout())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.server.Shutdown()
	}()

	select {
	case err := <-done:
		return err
	case <-shutdownCtx.Done():
		return fmt.Errorf("server shutdown timed out")
	}
}

// RegisterRoute registers a new route with the server.
// The handler parameter should be a fasthttp.RequestHandler function.
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

	// Convert handler to FastHTTP handler
	fasthttpHandler, ok := handler.(fasthttp.RequestHandler)
	if !ok {
		return fmt.Errorf("handler must be a fasthttp.RequestHandler, got %T", handler)
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

	// Wrap handler with observability
	wrappedHandler := s.wrapHandler(method, path, fasthttpHandler)

	// Register route
	methodMap[path] = wrappedHandler

	return nil
}

// RegisterMiddleware registers middleware with the server.
// The middleware parameter should be a fasthttp.RequestHandler function.
func (s *Server) RegisterMiddleware(middleware interface{}) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert middleware to FastHTTP handler
	fasthttpMiddleware, ok := middleware.(fasthttp.RequestHandler)
	if !ok {
		return fmt.Errorf("middleware must be a fasthttp.RequestHandler, got %T", middleware)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, fasthttpMiddleware)
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

// buildHandler builds the main FastHTTP handler with middleware chain.
func (s *Server) buildHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Apply middleware chain
		handler := s.routeHandler
		for i := len(s.middleware) - 1; i >= 0; i-- {
			middleware := s.middleware[i]
			currentHandler := handler
			handler = func(ctx *fasthttp.RequestCtx) {
				// Execute middleware
				middleware(ctx)
				// Continue to next handler
				currentHandler(ctx)
			}
		}

		handler(ctx)
	}
}

// routeHandler handles routing to registered routes.
func (s *Server) routeHandler(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())
	path := string(ctx.Path())

	s.mutex.RLock()
	methodMap, exists := s.routes[method]
	if !exists {
		s.mutex.RUnlock()
		ctx.Error("Method Not Allowed", fasthttp.StatusMethodNotAllowed)
		return
	}

	handler, exists := methodMap[path]
	s.mutex.RUnlock()

	if !exists {
		ctx.Error("Not Found", fasthttp.StatusNotFound)
		return
	}

	handler(ctx)
}

// wrapHandler wraps a FastHTTP handler with observability and statistics.
func (s *Server) wrapHandler(method, path string, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		startTime := time.Now()

		// Update active connections
		atomic.AddInt64(&s.stats.activeConnections, 1)
		defer atomic.AddInt64(&s.stats.activeConnections, -1)

		// Increment request count
		atomic.AddInt64(&s.stats.requestCount, 1)

		reqCtx := context.Background()

		// Notify route enter - using FastHTTP context directly
		routeData := hooks.RouteEventData{
			Method:  method,
			Path:    path,
			Request: ctx, // Use FastHTTP context instead of http.Request
		}
		s.observers.NotifyObservers(interfaces.EventRouteEnter, reqCtx, routeData)

		// Notify request - using FastHTTP context directly
		s.observers.NotifyObservers(interfaces.EventRequest, reqCtx, ctx)

		// Execute handler with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt64(&s.stats.errorCount, 1)
					err := fmt.Errorf("panic in handler: %v", r)
					s.observers.NotifyObservers(interfaces.EventError, reqCtx, err)
					ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
				}
			}()
			handler(ctx)
		}()

		duration := time.Since(startTime)

		// Update response time stats
		atomic.AddInt64(&s.stats.totalResponseTime, int64(duration))

		// Create response data with FastHTTP-specific response info
		respData := hooks.ResponseEventData{
			Request: ctx, // FastHTTP context
			Response: map[string]interface{}{ // Generic response data
				"statusCode": ctx.Response.StatusCode(),
				"headers":    extractHeaders(ctx),
			},
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventResponse, reqCtx, respData)

		// Notify route exit
		routeExitData := hooks.RouteExitEventData{
			Method:   method,
			Path:     path,
			Request:  ctx, // FastHTTP context
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventRouteExit, reqCtx, routeExitData)
	}
}

// extractHeaders extracts headers from FastHTTP response as a map
func extractHeaders(ctx *fasthttp.RequestCtx) map[string]string {
	headers := make(map[string]string)
	ctx.Response.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// NewFactory creates a new FastHTTP factory.
func NewFactory() *Factory {
	return &Factory{}
}
