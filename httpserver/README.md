# HTTP Server Library

A extensible HTTP server library for Go that implements multiple design patterns to provide a unified interface for different HTTP frameworks with **comprehensive graceful operations and generic hooks support**.

## 🚀 Key Features

- **6 HTTP Framework Providers**: Gin, Echo, Fiber, FastHTTP, Atreugo, NetHTTP
- **Generic Hooks Interface**: Framework-agnostic hooks working across all providers
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
- **Chain of Responsibility**: Hook chaining for complex request processing workflows

## Supported Providers

| Provider | Framework | Status | Graceful Ops | Hooks | Description |
|----------|-----------|--------|--------------|-------|-------------|
| **gin** | [Gin](https://github.com/gin-gonic/gin) | ✅ Native | ✅ Full | ✅ Generic | High-performance HTTP framework with native engine |
| **echo** | [Echo](https://github.com/labstack/echo) | ✅ Native | ✅ Full | ✅ Generic | High performance, extensible, minimalist web framework |
| **fiber** | [Fiber](https://github.com/gofiber/fiber) | ✅ Native | ✅ Full | ✅ Generic | Express inspired web framework built on Fasthttp |
| **fasthttp** | [FastHTTP](https://github.com/valyala/fasthttp) | ✅ Native | ✅ Full | ✅ Generic | Fast HTTP package for Go, 10x faster than net/http |
| **atreugo** | [Atreugo](https://github.com/savsgio/atreugo) | ✅ Native | ✅ Full | ✅ Generic | High performance fiber for Fasthttp framework |
| **nethttp** | Standard library | ✅ Native | ✅ Full | ✅ Generic | Go standard net/http server |

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

## 🎣 Generic Hooks Interface

All providers support a comprehensive generic hooks system that allows framework-agnostic request interception and processing:

### Basic Hook Usage

```go
import (
    "github.com/fsvxavier/nexs-lib/httpserver/hooks"
    "github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Create hook registry
registry := hooks.NewDefaultHookRegistry()

// Create and register logging hook
loggingHook := hooks.NewLoggingHook(log.Default())
err := registry.Register(loggingHook)

// Create and register metrics hook
metricsHook := hooks.NewMetricsHook()
err = registry.Register(metricsHook)

// Create and register security hook with CORS
securityHook := hooks.NewSecurityHook()
securityHook.EnableCORS("*", []string{"GET", "POST"}, []string{"Content-Type"})
err = registry.Register(securityHook)
```

### Hook Types Available

```go
// Basic Hook - Simple request interception
type Hook interface {
    Execute(ctx *HookContext) error
    Name() string
    Events() []HookEvent
    Priority() int
    IsEnabled() bool
    ShouldExecute(ctx *HookContext) bool
}

// Async Hook - Non-blocking execution
type AsyncHook interface {
    Hook
    ExecuteAsync(ctx *HookContext) <-chan error
    BufferSize() int
    Timeout() time.Duration
}

// Conditional Hook - Execute based on conditions
type ConditionalHook interface {
    Hook
    Condition() func(ctx *HookContext) bool
}

// Filtered Hook - Request filtering capabilities
type FilteredHook interface {
    Hook
    PathFilter() func(path string) bool
    MethodFilter() func(method string) bool
    HeaderFilter() func(headers http.Header) bool
}
```

### Built-in Hooks

| Hook | Purpose | Features |
|------|---------|----------|
| **LoggingHook** | Request/Response logging | Structured logging, configurable levels |
| **MetricsHook** | Performance metrics | Request timing, throughput, error rates |
| **SecurityHook** | Security enforcement | CORS, IP filtering, header validation |
| **CacheHook** | Response caching | TTL-based caching, cache invalidation |
| **HealthCheckHook** | Health monitoring | Custom health checks, status reporting |

### Advanced Hook Features

```go
// Hook Chaining
chain := hooks.NewHookChain()
chain.Add(loggingHook).Add(securityHook).Add(metricsHook)

// Execute with conditions
err := chain.ExecuteIf(ctx, func(ctx *interfaces.HookContext) bool {
    return ctx.Request.Method == "POST"
})

// Execute until condition is met
err = chain.ExecuteUntil(ctx, func(ctx *interfaces.HookContext) bool {
    return ctx.Metadata["processed"] == true
})

// Filtered execution with builders
filteredHook := hooks.NewFilteredBaseHook("api-only", events, 10)
filteredHook.SetPathFilter(
    hooks.NewPathFilterBuilder().
        Include("/api/*").
        Exclude("/api/health").
        Build(),
)
```

### Hook Events

30+ event types covering complete server lifecycle:

```go
const (
    // Server Events
    ServerStart HookEvent = "server.start"
    ServerStop  HookEvent = "server.stop"
    
    // Request Events  
    RequestReceived   HookEvent = "request.received"
    RequestProcessing HookEvent = "request.processing"
    RequestCompleted  HookEvent = "request.completed"
    
    // Response Events
    ResponseSending HookEvent = "response.sending"
    ResponseSent    HookEvent = "response.sent"
    
    // Middleware Events
    MiddlewareExecuting HookEvent = "middleware.executing"
    MiddlewareCompleted HookEvent = "middleware.completed"
    
    // Security Events
    SecurityCheck   HookEvent = "security.check"
    SecurityBlocked HookEvent = "security.blocked"
    
    // And many more...
)
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
func (s *Server) PreShutdownHook(hook func() error) { /* ... */ }
func (s *Server) PostShutdownHook(hook func() error) { /* ... */ }
func (s *Server) SetDrainTimeout(timeout time.Duration) { /* ... */ }
func (s *Server) WaitForConnections(ctx context.Context) error { /* ... */ }
```

### 2. Register Provider Factory

```go
// Package init
func init() {
    httpserver.RegisterProvider("myframework", func(cfg interface{}) (interfaces.HTTPServer, error) {
        return NewServer(cfg)
    })
}
```

## 📊 Testing & Examples

### Test Coverage

The library maintains high test coverage across all components:

- **HTTP Server Core**: 64 Go files with comprehensive test suites
- **All Providers**: Complete test coverage for all 6 providers (Gin 79.8%, Echo, Fiber, etc.)
- **Graceful Operations**: Full integration testing across all providers
- **Generic Hooks**: 87.8% test coverage with 60+ test cases
- **Mock Implementations**: Complete mocks available for all providers

### Running Tests

```bash
# Run all tests
cd httpserver
go test ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...

# Run specific provider tests
go test ./providers/gin/
go test ./providers/echo/
go test ./hooks/
```

### Available Examples

| Example | Description | Location |
|---------|-------------|----------|
| **Basic Provider Examples** | Simple server setup for each provider | `examples/gin/`, `examples/echo/`, etc. |
| **Graceful Operations** | Complete graceful shutdown and restart examples | `examples/graceful/` |
| **Generic Hooks** | Comprehensive hooks demonstration with all features | `examples/hooks/` |
| **Middleware Integration** | Custom middleware examples | `examples/middleware/` |

### Hook Example Output

```bash
cd examples/hooks
go run main.go

🚀 Generic Hooks Example
========================

📋 Setting up hooks...
✅ Logging hook registered successfully
✅ Metrics hook registered successfully  
✅ Security hook registered successfully
✅ Cache hook registered successfully
✅ Health check hook registered successfully

🔄 Simulating server lifecycle...
[ServerStart] Server starting on :8080
[RequestReceived] GET /api/users - 127.0.0.1
[SecurityCheck] CORS validation passed
[RequestProcessing] Processing request
[ResponseSending] Sending 200 response
[ResponseSent] Response sent successfully

📊 Hook Execution Metrics:
- Total executions: 28
- Sequential executions: 20  
- Parallel executions: 8
- Average execution time: 1.2ms
- Hook registry observer events: 15

✅ Generic Hooks example completed successfully!
```

## 🏗️ Architecture Benefits

- **Framework Independence**: Write once, deploy on any supported framework
- **Production Ready**: Built-in graceful operations and comprehensive testing
- **Extensible**: Easy to add new providers and hooks
- **Observability**: Complete lifecycle visibility through hooks and observers
- **Type Safety**: Full compile-time checking with Go interfaces
- **Performance**: Minimal overhead with efficient hook execution
- **Standards Compliant**: Follows Go best practices and idioms

## 📚 Learn More

- [NEXT_STEPS.md](NEXT_STEPS.md) - Detailed roadmap and implementation status
- [examples/](examples/) - Working examples for all features
- [hooks/](hooks/) - Generic hooks implementation details
- [graceful/](graceful/) - Graceful operations implementation
- [providers/](providers/) - Individual provider implementations
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
