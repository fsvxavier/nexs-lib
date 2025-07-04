package fiber

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/fiber/middleware"
)

// FiberServer implements the common.Server interface for Fiber
type FiberServer struct {
	app    *fiber.App
	config *common.ServerConfig
}

// NewServer creates a new Fiber server
func NewServer(options ...common.ServerOption) *FiberServer {
	config := common.DefaultServerConfig()
	for _, opt := range options {
		opt(config)
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:           config.ReadTimeout,
		WriteTimeout:          config.WriteTimeout,
		IdleTimeout:           config.IdleTimeout,
		DisableStartupMessage: true,
		Prefork:               config.Prefork,
		ErrorHandler:          middleware.ErrorHandler,
	})

	s := &FiberServer{
		app:    app,
		config: config,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware stack
func (s *FiberServer) setupMiddleware() {
	// Recovery middleware
	s.app.Use(recover.New())

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Add tracing if enabled
	if s.config.EnableTracing {
		s.app.Use(middleware.Tracing())
	}

	// Add logging if enabled
	if s.config.EnableLogging {
		s.app.Use(middleware.Logger())
	}

	// Add metrics if enabled
	if s.config.EnableMetrics {
		s.app.Get("/metrics", monitor.New())
	}

	// Add pprof if enabled
	if s.config.EnablePprof {
		s.app.Use(pprof.New())
	}
}

// setupRoutes configures the routes
func (s *FiberServer) setupRoutes() {
	// Health check endpoints
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	s.app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Add swagger if enabled
	if s.config.EnableSwagger {
		s.app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Add a 404 handler
	s.app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	})
}

// Start starts the server
func (s *FiberServer) Start() error {
	// Start the server in a goroutine
	go func() {
		addr := s.Address()
		fmt.Printf("Starting Fiber server on %s\n", addr)
		if err := s.app.Listen(addr); err != nil && err != http.ErrServerClosed {
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
func (s *FiberServer) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down Fiber server...")
	return s.app.ShutdownWithContext(ctx)
}

// Address returns the server's address
func (s *FiberServer) Address() string {
	return s.config.Host + ":" + s.config.Port
}

// App returns the underlying Fiber app instance
func (s *FiberServer) App() *fiber.App {
	return s.app
}

// Health returns a handler for health checks
func (s *FiberServer) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
