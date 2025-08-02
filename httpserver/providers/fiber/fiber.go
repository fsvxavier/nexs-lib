// Package fiber provides HTTP server implementation using Fiber v2 framework.
// This is the default provider for the HTTP server abstraction, offering high performance
// and extensive middleware support.
package fiber

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Factory implements the ProviderFactory interface for Fiber servers.
type Factory struct{}

// Create creates a new Fiber server instance.
func (f *Factory) Create(cfg interface{}) (interfaces.HTTPServer, error) {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	// Create Fiber config
	fiberConfig := fiber.Config{
		ReadTimeout:  baseConfig.GetReadTimeout(),
		WriteTimeout: baseConfig.GetWriteTimeout(),
		IdleTimeout:  baseConfig.GetIdleTimeout(),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
		DisableStartupMessage: true,
	}

	app := fiber.New(fiberConfig)

	server := &Server{
		config:     baseConfig,
		app:        app,
		observers:  hooks.NewObserverManager(),
		middleware: make([]fiber.Handler, 0),
		routes:     make(map[string]map[string]fiber.Handler),
		stats: &serverStats{
			provider: "fiber",
		},
		mutex: sync.RWMutex{},
	}

	// Initialize route map
	server.routes["GET"] = make(map[string]fiber.Handler)
	server.routes["POST"] = make(map[string]fiber.Handler)
	server.routes["PUT"] = make(map[string]fiber.Handler)
	server.routes["DELETE"] = make(map[string]fiber.Handler)
	server.routes["PATCH"] = make(map[string]fiber.Handler)
	server.routes["OPTIONS"] = make(map[string]fiber.Handler)
	server.routes["HEAD"] = make(map[string]fiber.Handler)

	// Attach configured observers
	for _, observer := range baseConfig.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	// Register configured middlewares
	for _, middleware := range baseConfig.GetMiddlewares() {
		// Skip middleware registration here as they need to be Fiber-specific
		_ = middleware
	}

	return server, nil
}

// GetName returns the provider name.
func (f *Factory) GetName() string {
	return "fiber"
}

// GetDefaultConfig returns the default configuration for Fiber.
func (f *Factory) GetDefaultConfig() interface{} {
	return config.NewBaseConfig()
}

// ValidateConfig validates the configuration for Fiber.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}
	return baseConfig.Validate()
}

// Server implements the HTTPServer interface using Fiber.
type Server struct {
	config     *config.BaseConfig
	app        *fiber.App
	observers  *hooks.ObserverManager
	middleware []fiber.Handler
	routes     map[string]map[string]fiber.Handler
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

// Start starts the Fiber server.
func (s *Server) Start(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 1 {
		return fmt.Errorf("server is already running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	addr := s.config.GetFullAddr()

	// Update stats
	s.stats.mutex.Lock()
	s.stats.startTime = time.Now()
	s.stats.mutex.Unlock()

	// Notify observers
	if err := s.observers.NotifyObservers(interfaces.EventStart, ctx, addr); err != nil {
		return fmt.Errorf("observer notification failed: %w", err)
	}

	atomic.StoreInt32(&s.running, 1)

	// Start server in background
	go func() {
		if err := s.app.Listen(addr); err != nil {
			atomic.StoreInt32(&s.running, 0)
			s.observers.NotifyObservers(interfaces.EventError, ctx, err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop gracefully stops the Fiber server.
func (s *Server) Stop(ctx context.Context) error {
	if atomic.LoadInt32(&s.running) == 0 {
		return fmt.Errorf("server is not running")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Shutdown server
	err := s.app.Shutdown()
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
// The handler parameter should be a fiber.Handler function.
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

	// Convert handler to Fiber handler
	fiberHandler, ok := handler.(fiber.Handler)
	if !ok {
		return fmt.Errorf("handler must be a fiber.Handler, got %T", handler)
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
	wrappedHandler := s.wrapHandler(method, path, fiberHandler)

	// Register route
	methodMap[path] = wrappedHandler

	switch method {
	case "GET":
		s.app.Get(path, wrappedHandler)
	case "POST":
		s.app.Post(path, wrappedHandler)
	case "PUT":
		s.app.Put(path, wrappedHandler)
	case "DELETE":
		s.app.Delete(path, wrappedHandler)
	case "PATCH":
		s.app.Patch(path, wrappedHandler)
	case "OPTIONS":
		s.app.Options(path, wrappedHandler)
	case "HEAD":
		s.app.Head(path, wrappedHandler)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	return nil
}

// RegisterMiddleware registers middleware with the server.
// The middleware parameter should be a fiber.Handler function.
func (s *Server) RegisterMiddleware(middleware interface{}) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	// Convert middleware to Fiber handler
	fiberMiddleware, ok := middleware.(fiber.Handler)
	if !ok {
		return fmt.Errorf("middleware must be a fiber.Handler, got %T", middleware)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, fiberMiddleware)
	s.app.Use(fiberMiddleware)
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

// wrapHandler wraps a Fiber handler with observability and statistics.
func (s *Server) wrapHandler(method, path string, handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()

		// Update active connections
		atomic.AddInt64(&s.stats.activeConnections, 1)
		defer atomic.AddInt64(&s.stats.activeConnections, -1)

		// Increment request count
		atomic.AddInt64(&s.stats.requestCount, 1)

		ctx := context.Background()

		// Notify route enter - using Fiber context directly
		routeData := hooks.RouteEventData{
			Method:  method,
			Path:    path,
			Request: c, // Use Fiber context instead of http.Request
		}
		s.observers.NotifyObservers(interfaces.EventRouteEnter, ctx, routeData)

		// Notify request - using Fiber context directly
		s.observers.NotifyObservers(interfaces.EventRequest, ctx, c)

		// Execute handler with panic recovery
		var handlerErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt64(&s.stats.errorCount, 1)
					err := fmt.Errorf("panic in handler: %v", r)
					s.observers.NotifyObservers(interfaces.EventError, ctx, err)
					handlerErr = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Internal Server Error",
					})
				}
			}()
			handlerErr = handler(c)
		}()

		duration := time.Since(startTime)

		// Update response time stats
		atomic.AddInt64(&s.stats.totalResponseTime, int64(duration))

		// Create response data with Fiber-specific response info
		respData := hooks.ResponseEventData{
			Request: c, // Fiber context
			Response: map[string]interface{}{ // Generic response data
				"statusCode": c.Response().StatusCode(),
				"headers":    extractHeaders(c),
			},
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventResponse, ctx, respData)

		// Notify route exit
		routeExitData := hooks.RouteExitEventData{
			Method:   method,
			Path:     path,
			Request:  c, // Fiber context
			Duration: duration,
		}
		s.observers.NotifyObservers(interfaces.EventRouteExit, ctx, routeExitData)

		return handlerErr
	}
}

// extractHeaders extracts headers from Fiber response as a map
func extractHeaders(c *fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	c.Response().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// NewFactory creates a new Fiber factory.
func NewFactory() *Factory {
	return &Factory{}
}
