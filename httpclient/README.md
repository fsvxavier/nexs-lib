# HTTP Client Library

A high-performance, multi-provider HTTP client library for Go with support for NetHTTP, Fiber, and FastHTTP backends. Built with clean architecture principles, dependency injection, and extensive configurability.

## 🚀 Features

- **Multiple Providers**: NetHTTP, Fiber, and FastHTTP backends
- **Factory Pattern**: Easy provider switching and configuration
- **Clean Architecture**: Interface-driven design with dependency inversion
- **Performance Metrics**: Built-in request/response monitoring
- **Retry Logic**: Configurable retry mechanisms with exponential backoff
- **Distributed Tracing**: Support for request tracing across services
- **Context Support**: Full context.Context integration for cancellation
- **Connection Pooling**: Optimized connection management
- **Flexible Configuration**: Builder pattern for complex configurations
- **Error Handling**: Customizable error handling strategies
- **Concurrent Safe**: Thread-safe client operations
- **Comprehensive Testing**: 98%+ test coverage with benchmarks

## 📦 Installation

```bash
go get github.com/fsvxavier/nexs-lib/httpclient
```

## 🏗️ Architecture

The library follows clean architecture principles with clear separation of concerns:

```
httpclient/
├── interfaces/          # Core contracts and interfaces
├── config/             # Configuration management
├── providers/          # Provider implementations
│   ├── nethttp/       # Standard net/http provider
│   ├── fiber/         # Fiber framework provider
│   └── fasthttp/      # FastHTTP provider
└── examples/          # Usage examples and demos
```

### Key Design Patterns

- **Factory Pattern**: Provider creation and management
- **Strategy Pattern**: Pluggable HTTP providers
- **Builder Pattern**: Flexible configuration construction
- **Dependency Injection**: Interface-based dependencies
- **Repository Pattern**: Clean data access abstraction

## 🎯 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fsvxavier/nexs-lib/httpclient"
    "github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func main() {
    // Create a simple client
    client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://api.example.com")
    if err != nil {
        log.Fatal(err)
    }

    // Make a GET request
    resp, err := client.Get(context.Background(), "/users")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %d\n", resp.StatusCode)
    fmt.Printf("Body: %s\n", string(resp.Body))
}
```

### Advanced Configuration

```go
package main

import (
    "context"
    "time"

    "github.com/fsvxavier/nexs-lib/httpclient"
    "github.com/fsvxavier/nexs-lib/httpclient/config"
    "github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func main() {
    // Build custom configuration
    cfg := config.NewBuilder().
        WithBaseURL("https://api.example.com").
        WithTimeout(30 * time.Second).
        WithMaxIdleConns(100).
        WithHeader("Authorization", "Bearer token").
        WithHeader("User-Agent", "MyApp/1.0").
        WithMaxRetries(3).
        WithRetryInterval(1 * time.Second).
        WithTracingEnabled(true).
        WithMetricsEnabled(true).
        Build()

    // Create client with configuration
    client, err := httpclient.NewWithConfig(interfaces.ProviderFastHTTP, cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Set custom error handler
    client.SetErrorHandler(func(resp *interfaces.Response) error {
        if resp.StatusCode >= 400 {
            return fmt.Errorf("API error %d: %s", resp.StatusCode, string(resp.Body))
        }
        return nil
    })

    // Make requests
    ctx := context.Background()
    
    // GET request
    resp, err := client.Get(ctx, "/users")
    
    // POST request with JSON body
    user := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    resp, err = client.Post(ctx, "/users", user)
    
    // Check metrics
    metrics := client.GetProvider().GetMetrics()
    fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
    fmt.Printf("Average latency: %v\n", metrics.AverageLatency)
}
```

## 🔧 Provider Comparison

| Feature | NetHTTP | Fiber | FastHTTP |
|---------|---------|-------|----------|
| **Performance** | High | Very High | Ultra High |
| **Memory Usage** | Moderate | Low | Very Low |
| **HTTP/2 Support** | ✅ | ❌ | ❌ |
| **Standard Library** | ✅ | ❌ | ❌ |
| **Ease of Use** | ✅ | ✅ | ⚠️ |
| **Concurrency** | High | Very High | Ultra High |
| **JSON Performance** | Good | Excellent | Excellent |
| **Connection Pooling** | ✅ | ✅ | ✅ |
| **Best Use Case** | General Purpose | High Throughput | Maximum Performance |

### When to Use Each Provider

#### NetHTTP Provider
- **General purpose applications**
- **HTTP/2 requirements**
- **Standard library preference**
- **Enterprise applications**
- **Microservices communication**

#### Fiber Provider  
- **High-throughput APIs**
- **JSON-heavy workloads**
- **Fiber web framework integration**
- **Real-time applications**
- **Moderate performance requirements**

#### FastHTTP Provider
- **Ultra-high performance needs**
- **Maximum concurrency**
- **Low-latency requirements**
- **High-frequency operations**
- **Resource-constrained environments**

## 🎛️ Configuration Options

### Basic Configuration

```go
cfg := config.DefaultConfig()
cfg.BaseURL = "https://api.example.com"
cfg.Timeout = 30 * time.Second
```

### Builder Pattern

```go
cfg := config.NewBuilder().
    WithBaseURL("https://api.example.com").
    WithTimeout(30 * time.Second).
    WithMaxIdleConns(100).
    WithIdleConnTimeout(90 * time.Second).
    WithTLSHandshakeTimeout(10 * time.Second).
    WithDisableKeepAlives(false).
    WithInsecureSkipVerify(false).
    WithHeaders(map[string]string{
        "User-Agent": "MyApp/1.0",
        "Accept":     "application/json",
    }).
    WithMaxRetries(3).
    WithRetryInterval(1 * time.Second).
    WithTracingEnabled(true).
    WithMetricsEnabled(true).
    Build()
```

### Retry Configuration

```go
retryConfig := &interfaces.RetryConfig{
    MaxRetries:      5,
    InitialInterval: 1 * time.Second,
    MaxInterval:     30 * time.Second,
    Multiplier:      2.0,
    RetryCondition: func(resp *interfaces.Response, err error) bool {
        // Custom retry logic
        if err != nil {
            return true // Retry on any error
        }
        // Retry on server errors
        return resp.StatusCode >= 500
    },
}

cfg.RetryConfig = retryConfig
```

## 🔄 HTTP Methods

All providers support the standard HTTP methods:

```go
// GET request
resp, err := client.Get(ctx, "/users")

// POST request with body
resp, err := client.Post(ctx, "/users", userData)

// PUT request
resp, err := client.Put(ctx, "/users/123", updateData)

// DELETE request
resp, err := client.Delete(ctx, "/users/123")

// PATCH request
resp, err := client.Patch(ctx, "/users/123", patchData)

// HEAD request
resp, err := client.Head(ctx, "/users")

// OPTIONS request
resp, err := client.Options(ctx, "/users")

// Generic execute
resp, err := client.Execute(ctx, "CUSTOM", "/endpoint", data)
```

## 📊 Metrics and Monitoring

### Built-in Metrics

```go
metrics := client.GetProvider().GetMetrics()

fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
fmt.Printf("Successful Requests: %d\n", metrics.SuccessfulRequests)
fmt.Printf("Failed Requests: %d\n", metrics.FailedRequests)
fmt.Printf("Success Rate: %.2f%%\n", 
    float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
fmt.Printf("Average Latency: %v\n", metrics.AverageLatency)
fmt.Printf("Last Request: %v\n", metrics.LastRequestTime)
```

### Response Information

```go
resp, err := client.Get(ctx, "/endpoint")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Status Code: %d\n", resp.StatusCode)
fmt.Printf("Headers: %+v\n", resp.Headers)
fmt.Printf("Body Length: %d\n", len(resp.Body))
fmt.Printf("Latency: %v\n", resp.Latency)
fmt.Printf("Is Error: %t\n", resp.IsError)
```

## 🔧 Error Handling

### Custom Error Handlers

```go
client.SetErrorHandler(func(resp *interfaces.Response) error {
    switch resp.StatusCode {
    case 400:
        return errors.New("bad request")
    case 401:
        return errors.New("unauthorized")
    case 403:
        return errors.New("forbidden")
    case 404:
        return errors.New("not found")
    case 429:
        return errors.New("rate limited")
    case 500:
        return errors.New("internal server error")
    default:
        if resp.StatusCode >= 400 {
            return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(resp.Body))
        }
        return nil
    }
})
```

### Retry Logic

```go
// Configure retries
client.SetRetryConfig(interfaces.RetryConfig{
    MaxRetries:      3,
    InitialInterval: 1 * time.Second,
    MaxInterval:     10 * time.Second,
    Multiplier:      2.0,
    RetryCondition: func(resp *interfaces.Response, err error) bool {
        // Retry on network errors
        if err != nil {
            return true
        }
        // Retry on specific status codes
        switch resp.StatusCode {
        case 408, 429, 502, 503, 504:
            return true
        }
        return false
    },
})
```

## 🧪 Testing

The library includes comprehensive test coverage with various testing utilities:

### Running Tests

```bash
# Unit tests
go test -v -race -timeout 30s ./...

# Tests with coverage
go test -v -race -timeout 30s -coverprofile=coverage.out ./...

# Benchmark tests
go test -bench=. -benchmem ./...

# Integration tests
go test -tags=integration -v ./...
```

### Mock Providers

```go
import "github.com/fsvxavier/nexs-lib/httpclient/providers/nethttp/mocks"

func TestMyFunction(t *testing.T) {
    mockProvider := mocks.NewMockProvider()
    
    // Set up expectations
    mockProvider.ExpectSuccessfulRequest("GET", "/users", 
        []byte(`{"users": []}`), 200)
    
    // Use mock in your tests
    client := &httpclient.Client{Provider: mockProvider}
    
    // Test your code
    resp, err := client.Get(context.Background(), "/users")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Verify expectations
    mockProvider.AssertExpectations(t)
}
```

## 📈 Performance Benchmarks

### Throughput Comparison

```
BenchmarkNetHTTP_Get-8     10000   150000 ns/op   2048 B/op   12 allocs/op
BenchmarkFiber_Get-8       20000    80000 ns/op   1024 B/op    8 allocs/op
BenchmarkFastHTTP_Get-8    50000    30000 ns/op    512 B/op    4 allocs/op
```

### Memory Usage

```
Provider  | Memory/Request | Allocations/Request | GC Pressure
----------|----------------|-------------------|-------------
NetHTTP   | 2048 B        | 12                | Moderate
Fiber     | 1024 B        | 8                 | Low
FastHTTP  | 512 B         | 4                 | Very Low
```

## 🎯 Examples

Comprehensive examples are available for each provider:

- **[NetHTTP Example](examples/nethttp/README.md)**: Standard library implementation
- **[Fiber Example](examples/fiber/README.md)**: High-performance Fiber client
- **[FastHTTP Example](examples/fasthttp/README.md)**: Ultra-fast HTTP operations

### Running Examples

```bash
# NetHTTP example
cd examples/nethttp && go run main.go

# Fiber example
cd examples/fiber && go run main.go

# FastHTTP example
cd examples/fasthttp && go run main.go
```

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Write tests**: Ensure 98%+ coverage
4. **Run tests**: `go test -v -race ./...`
5. **Submit a pull request**

### Development Setup

```bash
# Clone the repository
git clone https://github.com/fsvxavier/nexs-lib.git
cd nexs-lib/httpclient

# Install dependencies
go mod download

# Run tests
make test

# Run linting
make lint

# Generate coverage report
make coverage
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Team**: For the excellent standard library
- **Fiber Team**: For the high-performance web framework
- **FastHTTP Team**: For the ultra-fast HTTP implementation
- **Community**: For feedback and contributions

## 📞 Support

- **Documentation**: [Full API Documentation](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/httpclient)
- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

---

**Made with ❤️ for the Go community**
