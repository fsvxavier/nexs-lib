package atreugo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/savsgio/atreugo/v11"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Factory implements the Factory pattern for creating Atreugo HTTP servers.
type Factory struct{}

// Server represents an Atreugo HTTP server implementation.
type Server struct {
	atreugo    *atreugo.Atreugo
	config     *config.BaseConfig
	observers  *hooks.ObserverManager
	middleware []atreugo.Middleware
	routes     map[string]map[string]atreugo.View // method -> path -> handler
	mutex      sync.RWMutex
	running    int32
	stats      interfaces.ServerStats
	startTime  time.Time
}

// Create creates a new Atreugo server instance.
func (f *Factory) Create(cfg interface{}) (interfaces.HTTPServer, error) {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	if err := f.ValidateConfig(baseConfig); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Configure Atreugo
	atreugoConfig := atreugo.Config{
		Addr:         fmt.Sprintf("%s:%d", baseConfig.GetAddr(), baseConfig.GetPort()),
		ReadTimeout:  baseConfig.GetReadTimeout(),
		WriteTimeout: baseConfig.GetWriteTimeout(),
		IdleTimeout:  baseConfig.GetIdleTimeout(),
	}

	app := atreugo.New(atreugoConfig)

	server := &Server{
		atreugo:    app,
		config:     baseConfig,
		observers:  hooks.NewObserverManager(),
		middleware: make([]atreugo.Middleware, 0),
		routes:     make(map[string]map[string]atreugo.View),
		stats: interfaces.ServerStats{
			Provider:     "atreugo",
			RequestCount: 0,
		},
	}

	// Initialize route maps for supported methods
	supportedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range supportedMethods {
		server.routes[method] = make(map[string]atreugo.View)
	}

	// Register observers from config
	for _, observer := range baseConfig.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	// Register configured middlewares
	for _, middleware := range baseConfig.GetMiddlewares() {
		// Skip middleware registration here as they need to be Atreugo-specific
		_ = middleware
	}

	return server, nil
}

// GetName returns the provider name.
func (f *Factory) GetName() string {
	return "atreugo"
}

// GetDefaultConfig returns the default configuration for Atreugo.
func (f *Factory) GetDefaultConfig() interface{} {
	return config.NewBaseConfig()
}

// ValidateConfig validates the configuration for Atreugo.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	return baseConfig.Validate()
}

// Start starts the Atreugo HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

	// Build the final handler with middleware chain
	s.buildHandler()

	addr := s.GetAddr()

	// Notify observers of start event
	if err := s.observers.NotifyObservers(interfaces.EventStart, ctx, addr); err != nil {
		atomic.StoreInt32(&s.running, 0)
		return fmt.Errorf("failed to notify start observers: %w", err)
	}

	// Start server in a goroutine
	go func() {
		if err := s.atreugo.ListenAndServe(); err != nil {
			// Notify observers of error
			s.observers.NotifyObservers(interfaces.EventError, ctx, err)
		}
	}()

	return nil
}

// Stop stops the Atreugo HTTP server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		return fmt.Errorf("server is not running")
	}

	// Create context with shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.GetShutdownTimeout())
	defer cancel()

	// Graceful shutdown
	if err := s.atreugo.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}

	// Notify observers of stop event
	if err := s.observers.NotifyObservers(interfaces.EventStop, ctx, "graceful shutdown"); err != nil {
		return fmt.Errorf("failed to notify stop observers: %w", err)
	}

	return nil
}

// RegisterRoute registers a route with the Atreugo server.
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

	// Convert handler to Atreugo handler
	atreugoHandler, ok := handler.(atreugo.View)
	if !ok {
		return fmt.Errorf("handler must be an atreugo.View, got %T", handler)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate method
	methodMap, exists := s.routes[method]
	if !exists {
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Check for duplicate routes
	if _, exists := methodMap[path]; exists {
		return fmt.Errorf("route %s %s already registered", method, path)
	}

	// Store route
	methodMap[path] = atreugoHandler

	// Register route with Atreugo
	switch strings.ToUpper(method) {
	case "GET":
		s.atreugo.GET(path, s.wrapHandler(method, path, atreugoHandler))
	case "POST":
		s.atreugo.POST(path, s.wrapHandler(method, path, atreugoHandler))
	case "PUT":
		s.atreugo.PUT(path, s.wrapHandler(method, path, atreugoHandler))
	case "DELETE":
		s.atreugo.DELETE(path, s.wrapHandler(method, path, atreugoHandler))
	case "PATCH":
		s.atreugo.PATCH(path, s.wrapHandler(method, path, atreugoHandler))
	case "HEAD":
		s.atreugo.HEAD(path, s.wrapHandler(method, path, atreugoHandler))
	case "OPTIONS":
		s.atreugo.OPTIONS(path, s.wrapHandler(method, path, atreugoHandler))
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	return nil
}

// RegisterMiddleware registers a middleware with the Atreugo server.
func (s *Server) RegisterMiddleware(middleware interface{}) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	atreugoMiddleware, ok := middleware.(atreugo.Middleware)
	if !ok {
		return fmt.Errorf("middleware must be an atreugo.Middleware, got %T", middleware)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, atreugoMiddleware)
	s.atreugo.UseBefore(atreugoMiddleware)

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
	return fmt.Sprintf("%s:%d", s.config.GetAddr(), s.config.GetPort())
}

// IsRunning returns true if the server is currently running.
func (s *Server) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// GetStats returns server statistics.
func (s *Server) GetStats() interfaces.ServerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := s.stats
	// Note: Uptime field doesn't exist in ServerStats interface
	// Keeping it simple for now

	return stats
}

// buildHandler builds the final handler with middleware chain.
func (s *Server) buildHandler() {
	// Apply global middleware that was registered
	for _, mw := range s.middleware {
		s.atreugo.UseBefore(mw)
	}
}

// wrapHandler wraps an Atreugo handler with observability and statistics.
func (s *Server) wrapHandler(method, path string, handler atreugo.View) atreugo.View {
	return func(ctx *atreugo.RequestCtx) error {
		// Increment request count
		atomic.AddInt64(&s.stats.RequestCount, 1)

		// Convert fasthttp.RequestCtx to standard request for observers
		req := &ctx.RequestCtx

		// Notify observers of route enter
		s.observers.NotifyObservers(interfaces.EventRouteEnter, context.Background(), req)

		// Notify observers of request
		s.observers.NotifyObservers(interfaces.EventRequest, context.Background(), req)

		// Call the actual handler
		err := handler(ctx)

		// Notify observers of response
		s.observers.NotifyObservers(interfaces.EventResponse, context.Background(), req)

		// Notify observers of route exit
		s.observers.NotifyObservers(interfaces.EventRouteExit, context.Background(), req)

		return err
	}
}

// NewFactory creates a new Atreugo factory.
func NewFactory() interfaces.ProviderFactory {
	return &Factory{}
}
