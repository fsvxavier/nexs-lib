# HTTP Server Library

A extensible HTTP server library for Go that implements multiple design patterns to provide a unified interface for different HTTP frameworks with **comprehensive graceful operations support**.

## 🚀 Key Features

- **6 HTTP Framework Providers**: Gin, Echo, Fiber, FastHTTP, Atreugo, NetHTTP
- **Graceful Operations**: Complete graceful shutdown, restart, and health monitoring
- **Production Ready**: Signal handling, connection draining, zero-downtime operations
- **Framework Agnostic**: Switch between frameworks without code changes
- **Observability Built-in**: Hooks for monitoring, logging, and metrics
- **Type Safe**: Full compile-time type checking with interfaces

## Architecture

This library implements several design patterns:

- **Factory Pattern**: Register and instantiate HTTP servers by name (gin, echo, fiber, fasthttp, atreugo, nethttp)
- **Adapter Pattern**: Adapt standard `http.Handler` to framework-specific handlers for each framework
- **Observer Pattern**: Propagate lifecycle events (start/stop/request/response) to external hooks for monitoring/logging
- **Registry Pattern**: Maintain a registry of available HTTP servers and configurations

## Supported Providers

| Provider | Framework | Status | Graceful Ops | Description |
|----------|-----------|--------|--------------|-------------|
| **gin** | [Gin](https://github.com/gin-gonic/gin) | ✅ Native | ✅ Full | High-performance HTTP framework with native engine |
| **echo** | [Echo](https://github.com/labstack/echo) | ✅ Native | ✅ Full | High performance, extensible, minimalist web framework |
| **fiber** | [Fiber](https://github.com/gofiber/fiber) | ✅ Native | ✅ Full | Express inspired web framework built on Fasthttp |
| **fasthttp** | [FastHTTP](https://github.com/valyala/fasthttp) | ✅ Native | ✅ Full | Fast HTTP package for Go, 10x faster than net/http |
| **atreugo** | [Atreugo](https://github.com/savsgio/atreugo) | ✅ Native | ✅ Full | High performance fiber for Fasthttp framework |
| **nethttp** | Standard library | ✅ Native | ✅ Full | Go standard net/http server |

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/fsvxavier/nexs-lib/httpserver"
    "github.com/fsvxavier/nexs-lib/httpserver/config"
    "github.com/fsvxavier/nexs-lib/httpserver/hooks"
)

func main() {
    // Providers are auto-registered - no manual registration needed!
    
    // Attach observers for monitoring
    httpserver.AttachObserver(hooks.NewLoggingObserver(log.Default()))
    httpserver.AttachObserver(hooks.NewMetricsObserver())

    // Configure server
    cfg := config.DefaultConfig().
        WithHost("localhost").
        WithPort(8080).
        WithReadTimeout(30 * time.Second)

    // Create server (any supported provider)
    server, err := httpserver.Create("gin", cfg) // or "echo", "fiber", etc.
    if err != nil {
        log.Fatal(err)
    }

    // Set your handler
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from %s server!", "gin")
    })
    server.SetHandler(mux)

    // Start server
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }

    // Graceful shutdown with connection draining
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.GracefulStop(ctx, 10*time.Second); err != nil {
            log.Printf("Graceful shutdown failed: %v", err)
        }
    }()
}
```

## 🔄 Graceful Operations

All providers support comprehensive graceful operations for production environments:

### Basic Graceful Operations

```go
// Graceful shutdown with connection draining
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := server.GracefulStop(ctx, 10*time.Second) // drain timeout
if err != nil {
    log.Printf("Graceful shutdown failed: %v", err)
}

// Zero-downtime restart
err = server.Restart(context.Background())
if err != nil {
    log.Printf("Restart failed: %v", err)
}

// Health monitoring
status := server.GetHealthStatus()
fmt.Printf("Status: %s, Uptime: %s, Connections: %d\n", 
    status.Status, status.Uptime, status.Connections)

// Connection monitoring
activeConns := server.GetConnectionsCount()
fmt.Printf("Active connections: %d\n", activeConns)
```

### Advanced Graceful Management

```go
import "github.com/fsvxavier/nexs-lib/httpserver/graceful"

// Create graceful manager for multiple servers
manager := graceful.NewManager()

// Register servers
manager.RegisterServer("api", apiServer)
manager.RegisterServer("metrics", metricsServer)

// Add health checks
manager.AddHealthCheck("database", func() interfaces.HealthCheck {
    return interfaces.HealthCheck{
        Status: "healthy",
        Message: "Database connection OK",
    }
})

// Add hooks
manager.AddPreShutdownHook(func() error {
    log.Println("Preparing for shutdown...")
    return nil
})

manager.AddPostShutdownHook(func() error {
    log.Println("Cleanup completed")
    return nil
})

// Graceful shutdown all servers
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

err := manager.GracefulShutdown(ctx)
if err != nil {
    log.Printf("Graceful shutdown failed: %v", err)
}
```

### Signal Handling

```go
import (
    "os"
    "os/signal"
    "syscall"
)

// Setup graceful shutdown on signals
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

go func() {
    <-stop
    log.Println("Shutting down gracefully...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := manager.GracefulShutdown(ctx); err != nil {
        log.Printf("Forced shutdown: %v", err)
        os.Exit(1)
    }
    
    log.Println("Server stopped gracefully")
    os.Exit(0)
}()
```

## Registering New Providers

To register a new HTTP framework provider:

### 1. Implement the HTTPServer Interface

```go
// Package myframework provides HTTP server implementation
package myframework

import (
    "context"
    "net/http"
    "github.com/fsvxavier/nexs-lib/httpserver/interfaces"
    "github.com/fsvxavier/nexs-lib/httpserver/config"
)

type Server struct {
    // Your framework-specific fields
}

func NewServer(cfg interface{}) (interfaces.HTTPServer, error) {
    // Implementation
}

func (s *Server) Start() error { /* ... */ }
func (s *Server) Stop(ctx context.Context) error { /* ... */ }
func (s *Server) SetHandler(handler http.Handler) { /* ... */ }
func (s *Server) GetAddr() string { /* ... */ }
func (s *Server) IsRunning() bool { /* ... */ }
func (s *Server) GetConfig() *config.Config { /* ... */ }

// Graceful operations methods
func (s *Server) GracefulStop(ctx context.Context, drainTimeout time.Duration) error { /* ... */ }
func (s *Server) Restart(ctx context.Context) error { /* ... */ }
func (s *Server) GetConnectionsCount() int64 { /* ... */ }
func (s *Server) GetHealthStatus() interfaces.HealthStatus { /* ... */ }
func (s *Server) PreShutdownHook(hook func()) error { /* ... */ }
func (s *Server) PostShutdownHook(hook func()) error { /* ... */ }
func (s *Server) SetDrainTimeout(timeout time.Duration) { /* ... */ }
func (s *Server) WaitForConnections(ctx context.Context) error { /* ... */ }

var Factory interfaces.ServerFactory = NewServer
```

### 2. Register the Provider

```go
httpserver.Register("myframework", myframework.Factory)
```

### 3. Use Your Provider

```go
server, err := httpserver.Create("myframework", cfg)
```

## Using Hooks/Observers

The library supports the Observer pattern for monitoring server lifecycle:

### Available Hooks

- **LoggingObserver**: Logs all server events
- **MetricsObserver**: Collects performance metrics  
- **TracingObserver**: Distributed tracing integration

### Custom Observers

```go
type MyObserver struct{}

func (o *MyObserver) OnStart(name string) {
    // Server started
}

func (o *MyObserver) OnStop(name string) {
    // Server stopped
}

func (o *MyObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
    // Request completed
}

func (o *MyObserver) OnBeforeRequest(name string, req *http.Request) {
    // Before request processing
}

func (o *MyObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
    // After request processing
}

// Register observer
httpserver.AttachObserver(&MyObserver{})
```

## Configuration

Extensible configuration with framework-specific options:

```go
cfg := config.DefaultConfig().
    WithHost("0.0.0.0").
    WithPort(8080).
    WithReadTimeout(30 * time.Second).
    WithWriteTimeout(30 * time.Second).
    WithIdleTimeout(60 * time.Second).
    WithTLS("cert.pem", "key.pem")
```

## Examples

See the `examples/` directory for complete working examples:

- [Gin Example](examples/gin/) - Gin web framework with graceful operations
- [Echo Example](examples/echo/) - Echo framework with graceful shutdown  
- [Fiber Example](examples/fiber/) - Fiber framework with health monitoring
- [FastHTTP Example](examples/fasthttp/) - FastHTTP framework with restart capability
- [Atreugo Example](examples/atreugo/) - Atreugo framework with connection tracking
- [NetHTTP Example](examples/nethttp/) - Standard library with graceful manager
- [Graceful Example](examples/graceful/) - Multi-server graceful operations

## Testing

```bash
# Run all tests
go test -race -timeout 30s -coverprofile=coverage.out ./...

# Integration tests
go test -tags=integration -v ./...

# Benchmarks
go test -bench=. -benchmem ./...

# Linting
golangci-lint run
go vet ./...
go mod tidy
go mod verify
```

## Design Benefits

1. **Production Ready**: Complete graceful operations with connection draining and health monitoring
2. **Zero-Downtime Operations**: Graceful restart and shutdown without dropping connections
3. **Framework Agnostic**: Switch between HTTP frameworks without changing your application code
4. **Consistent Interface**: All frameworks expose the same interface including graceful operations
5. **Observability**: Built-in hooks for monitoring, logging, and metrics
6. **Extensible**: Easy to add new frameworks and observers
7. **Type Safe**: Full compile-time type checking
8. **Performance**: Native framework implementations for optimal performance

## Production Features

- ✅ **Graceful Shutdown**: Connection draining with configurable timeouts
- ✅ **Zero-Downtime Restart**: Hot restart without dropping connections  
- ✅ **Health Monitoring**: Comprehensive health status and checks
- ✅ **Connection Tracking**: Real-time active connection monitoring
- ✅ **Signal Handling**: SIGTERM/SIGINT graceful shutdown support
- ✅ **Hook System**: Pre/post shutdown hooks for cleanup operations
- ✅ **Multi-Server Management**: Centralized graceful operations manager
- ✅ **Testing Support**: Complete mocks with graceful operations

## License

MIT License
