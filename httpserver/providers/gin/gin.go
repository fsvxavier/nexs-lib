package gin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Factory implements the Factory pattern for creating Gin HTTP servers.
type Factory struct{}

// Server represents a Gin HTTP server implementation.
type Server struct {
	engine     *gin.Engine
	server     *http.Server
	config     *config.BaseConfig
	observers  *hooks.ObserverManager
	middleware []gin.HandlerFunc
	routes     map[string]map[string]gin.HandlerFunc // method -> path -> handler
	mutex      sync.RWMutex
	running    int32
	stats      interfaces.ServerStats
	startTime  time.Time
}

// Create creates a new Gin server instance.
func (f *Factory) Create(cfg interface{}) (interfaces.HTTPServer, error) {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	if err := f.ValidateConfig(baseConfig); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Add default middleware
	engine.Use(gin.Recovery())

	server := &Server{
		engine:     engine,
		config:     baseConfig,
		observers:  hooks.NewObserverManager(),
		middleware: make([]gin.HandlerFunc, 0),
		routes:     make(map[string]map[string]gin.HandlerFunc),
		stats: interfaces.ServerStats{
			Provider:     "gin",
			RequestCount: 0,
		},
	}

	// Initialize route maps for supported methods
	supportedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range supportedMethods {
		server.routes[method] = make(map[string]gin.HandlerFunc)
	}

	// Register observers from config
	for _, observer := range baseConfig.GetObservers() {
		if err := server.AttachObserver(observer); err != nil {
			return nil, fmt.Errorf("failed to attach observer: %w", err)
		}
	}

	// Register configured middlewares
	for _, middleware := range baseConfig.GetMiddlewares() {
		// Skip middleware registration here as they need to be Gin-specific
		_ = middleware
	}

	return server, nil
}

// GetName returns the provider name.
func (f *Factory) GetName() string {
	return "gin"
}

// GetDefaultConfig returns the default configuration for Gin.
func (f *Factory) GetDefaultConfig() interface{} {
	return config.NewBaseConfig()
}

// ValidateConfig validates the configuration for Gin.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	baseConfig, ok := cfg.(*config.BaseConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.BaseConfig, got %T", cfg)
	}

	return baseConfig.Validate()
}

// Start starts the Gin HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

	// Build the final handler with middleware chain
	s.buildHandler()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.GetAddr(), s.config.GetPort())
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.GetReadTimeout(),
		WriteTimeout: s.config.GetWriteTimeout(),
		IdleTimeout:  s.config.GetIdleTimeout(),
	}

	// Notify observers of start event
	if err := s.observers.NotifyObservers(interfaces.EventStart, ctx, addr); err != nil {
		atomic.StoreInt32(&s.running, 0)
		return fmt.Errorf("failed to notify start observers: %w", err)
	}

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Notify observers of error
			s.observers.NotifyObservers(interfaces.EventError, ctx, err)
		}
	}()

	return nil
}

// Stop stops the Gin HTTP server gracefully.
func (s *Server) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		return fmt.Errorf("server is not running")
	}

	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	// Create context with shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.GetShutdownTimeout())
	defer cancel()

	// Graceful shutdown
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		// Force close if graceful shutdown fails
		if closeErr := s.server.Close(); closeErr != nil {
			return fmt.Errorf("failed to force close server: %w", closeErr)
		}
		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
	}

	// Notify observers of stop event
	if err := s.observers.NotifyObservers(interfaces.EventStop, ctx, "graceful shutdown"); err != nil {
		return fmt.Errorf("failed to notify stop observers: %w", err)
	}

	return nil
}

// RegisterRoute registers a route with the Gin server.
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

	// Convert handler to Gin handler
	ginHandler, ok := handler.(gin.HandlerFunc)
	if !ok {
		return fmt.Errorf("handler must be a gin.HandlerFunc, got %T", handler)
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
	methodMap[path] = ginHandler

	// Register route with Gin engine
	switch strings.ToUpper(method) {
	case "GET":
		s.engine.GET(path, s.wrapHandler(method, path, ginHandler))
	case "POST":
		s.engine.POST(path, s.wrapHandler(method, path, ginHandler))
	case "PUT":
		s.engine.PUT(path, s.wrapHandler(method, path, ginHandler))
	case "DELETE":
		s.engine.DELETE(path, s.wrapHandler(method, path, ginHandler))
	case "PATCH":
		s.engine.PATCH(path, s.wrapHandler(method, path, ginHandler))
	case "HEAD":
		s.engine.HEAD(path, s.wrapHandler(method, path, ginHandler))
	case "OPTIONS":
		s.engine.OPTIONS(path, s.wrapHandler(method, path, ginHandler))
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	return nil
}

// RegisterMiddleware registers a middleware with the Gin server.
func (s *Server) RegisterMiddleware(middleware interface{}) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	ginMiddleware, ok := middleware.(gin.HandlerFunc)
	if !ok {
		return fmt.Errorf("middleware must be a gin.HandlerFunc, got %T", middleware)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, ginMiddleware)
	s.engine.Use(ginMiddleware)

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
		s.engine.Use(mw)
	}
}

// wrapHandler wraps a Gin handler with observability and statistics.
func (s *Server) wrapHandler(method, path string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment request count
		atomic.AddInt64(&s.stats.RequestCount, 1)

		// Notify observers of route enter
		s.observers.NotifyObservers(interfaces.EventRouteEnter, c.Request.Context(), c.Request)

		// Notify observers of request
		s.observers.NotifyObservers(interfaces.EventRequest, c.Request.Context(), c.Request)

		// Call the actual handler
		handler(c)

		// Notify observers of response
		s.observers.NotifyObservers(interfaces.EventResponse, c.Request.Context(), c.Request)

		// Notify observers of route exit
		s.observers.NotifyObservers(interfaces.EventRouteExit, c.Request.Context(), c.Request)
	}
}

// NewFactory creates a new Gin factory.
func NewFactory() interfaces.ProviderFactory {
	return &Factory{}
}
