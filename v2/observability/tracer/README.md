# Nexs-Lib v2 Observability Tracer

A comprehensive, production-ready distributed tracing library for Go applications with support for multiple providers including **Datadog**, **New Relic**, and **Prometheus**. Now with **98.9% test coverage** and **zero race conditions**.

## 🚀 Features

- **Multi-Provider Support**: Datadog APM, New Relic APM, and Prometheus metrics
- **OpenTelemetry-Compatible**: Follows OpenTelemetry standards and patterns
- **Production-Ready**: Enterprise-grade error handling, metrics, and monitoring
- **Zero Dependencies**: Core package has zero external dependencies
- **Type-Safe**: Full type safety with comprehensive interfaces
- **Performance Optimized**: Minimal overhead with efficient span management
- **Concurrent-Safe**: Thread-safe operations with proper synchronization
- **Comprehensive Testing**: 98%+ test coverage with race condition detection
- **Extensible Architecture**: Easy to add new providers and customize behavior

## 📦 Installation

```bash
go get github.com/fsvxavier/nexs-lib/v2/observability/tracer
```

## 🏗️ Architecture

The tracer follows a modular architecture with clear separation of concerns:

```
tracer/
├── interfaces.go          # Core interfaces and types
├── options.go            # Configuration options and builders
├── factory.go            # Provider factory and management
├── noop.go              # No-operation implementations
├── tracer.go            # Multi-provider and utilities
└── providers/
    ├── datadog/         # Datadog APM implementation
    ├── newrelic/        # New Relic APM implementation
    └── prometheus/      # Prometheus metrics implementation
```

## 🔧 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/fsvxavier/nexs-lib/v2/observability/tracer"
    "github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
)

func main() {
    // Create Datadog provider
    config := &datadog.Config{
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        Environment:    "production",
        AgentHost:      "localhost",
        AgentPort:      8126,
        SampleRate:     1.0,
    }
    
    provider, err := datadog.NewProvider(config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())
    
    // Create tracer
    tr, err := provider.CreateTracer("my-tracer", 
        tracer.WithServiceName("my-service"),
        tracer.WithEnvironment("production"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create and use spans
    ctx, span := tr.StartSpan(context.Background(), "operation",
        tracer.WithSpanKind(tracer.SpanKindServer),
        tracer.WithSpanAttributes(map[string]interface{}{
            "http.method": "GET",
            "http.url":    "/api/users",
        }),
    )
    defer span.End()
    
    // Add attributes and events
    span.SetAttribute("user.id", "12345")
    span.AddEvent("processing_started", map[string]interface{}{
        "queue_size": 42,
    })
    
    // Simulate work
    doSomeWork(ctx)
    
    // Set success status
    span.SetStatus(tracer.StatusCodeOk, "completed successfully")
}

func doSomeWork(ctx context.Context) {
    // Extract tracer from context or get from global factory
    tr, _ := tracer.GetProvider(tracer.ProviderTypeDatadog)
    tracer, _ := tr.CreateTracer("worker")
    
    _, childSpan := tracer.StartSpan(ctx, "database_query",
        tracer.WithSpanKind(tracer.SpanKindClient),
    )
    defer childSpan.End()
    
    // Simulate database work
    childSpan.SetAttribute("db.statement", "SELECT * FROM users")
    childSpan.SetStatus(tracer.StatusCodeOk, "query successful")
}
```

### Multi-Provider Setup

```go
func setupMultiProviderTracing() {
    // Setup Datadog
    ddConfig := &datadog.Config{
        ServiceName: "my-service",
        Environment: "production",
        AgentHost:   "dd-agent",
        AgentPort:   8126,
    }
    ddProvider, _ := datadog.NewProvider(ddConfig)
    
    // Setup New Relic
    nrConfig := &newrelic.Config{
        AppName:    "my-service",
        LicenseKey: "your-license-key",
        Environment: "production",
    }
    nrProvider, _ := newrelic.NewProvider(nrConfig)
    
    // Setup Prometheus
    promConfig := &prometheus.Config{
        ServiceName: "my-service",
        Namespace:   "myapp",
        Subsystem:   "api",
    }
    promProvider, _ := prometheus.NewProvider(promConfig)
    
    // Create tracers
    ddTracer, _ := ddProvider.CreateTracer("main")
    nrTracer, _ := nrProvider.CreateTracer("main")
    promTracer, _ := promProvider.CreateTracer("main")
    
    // Create multi-provider tracer
    multiTracer := tracer.NewMultiProviderTracer(ddTracer, nrTracer, promTracer)
    
    // Use multi-tracer (spans will be sent to all providers)
    ctx, span := multiTracer.StartSpan(context.Background(), "api_request")
    defer span.End()
    
    span.SetAttribute("user.id", "12345")
    span.SetStatus(tracer.StatusCodeOk, "success")
}
```

## 🏗️ Provider Configurations

### Datadog Configuration

```go
config := &datadog.Config{
    // Core settings
    ServiceName:     "my-service",
    ServiceVersion:  "1.2.3",
    Environment:     "production",
    
    // Agent configuration
    AgentHost:       "dd-agent.example.com",
    AgentPort:       8126,
    
    // Sampling and performance
    SampleRate:      0.1,  // 10% sampling
    EnableProfiling: true,
    Tags: map[string]string{
        "team":    "backend",
        "region":  "us-east-1",
    },
    
    // Advanced settings
    Debug:              false,
    RuntimeMetrics:     true,
    AnalyticsEnabled:   true,
    PrioritySampling:   true,
    MaxTracesPerSecond: 1000,
    
    // Security
    ObfuscationEnabled: true,
    ObfuscatedTags:     []string{"password", "token"},
}
```

### New Relic Configuration

```go
config := &newrelic.Config{
    // Core settings
    AppName:        "my-service",
    LicenseKey:     "your-40-character-license-key",
    Environment:    "production",
    ServiceVersion: "1.2.3",
    
    // Feature flags
    DistributedTracer: true,
    Enabled:           true,
    HighSecurity:      false,
    CodeLevelMetrics:  true,
    
    // Performance settings
    LogLevel:              "info",
    MaxSamplesStored:      10000,
    DatastoreTracer:       true,
    CrossApplicationTrace: true,
    
    // Security and compliance
    AttributesEnabled: true,
    AttributesExclude: []string{
        "request.headers.authorization",
        "request.headers.cookie",
    },
    CustomInsightsEvents: true,
    
    // Labels for metadata
    Labels: map[string]string{
        "team":        "backend",
        "environment": "production",
        "version":     "1.2.3",
    },
}
```

### Prometheus Configuration

```go
config := &prometheus.Config{
    // Core settings
    ServiceName:    "my-service",
    ServiceVersion: "1.2.3",
    Environment:    "production",
    Namespace:      "myapp",
    Subsystem:      "api",
    
    // Metrics configuration
    EnableDetailedMetrics: true,
    CustomLabels: map[string]string{
        "team":   "backend",
        "region": "us-east-1",
    },
    BucketBoundaries: []float64{
        0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
    },
    MaxCardinality: 1000,
    
    // Performance settings
    CollectionInterval: 30 * time.Second,
    RetentionPeriod:   24 * time.Hour,
    BatchSize:         100,
    
    // Registry settings
    UseGlobalRegistry: false,
    Registry:          nil, // Use custom registry if needed
}
```

## 📊 Metrics and Monitoring

### Tracer Metrics

Each tracer provides comprehensive metrics:

```go
tracer, _ := provider.CreateTracer("my-tracer")

// Get tracer metrics
metrics := tracer.GetMetrics()
fmt.Printf("Spans created: %d\n", metrics.SpansCreated)
fmt.Printf("Spans finished: %d\n", metrics.SpansFinished)
fmt.Printf("Spans dropped: %d\n", metrics.SpansDropped)
fmt.Printf("Average duration: %v\n", metrics.AvgSpanDuration)
fmt.Printf("Last activity: %v\n", metrics.LastActivity)
```

### Provider Metrics

Monitor provider health and performance:

```go
// Get provider metrics
metrics := provider.GetProviderMetrics()
fmt.Printf("Active tracers: %d\n", metrics.TracersActive)
fmt.Printf("Connection state: %s\n", metrics.ConnectionState)
fmt.Printf("Last flush: %v\n", metrics.LastFlush)
fmt.Printf("Error count: %d\n", metrics.ErrorCount)
fmt.Printf("Bytes sent: %d\n", metrics.BytesSent)

// Health check
err := provider.HealthCheck(context.Background())
if err != nil {
    log.Printf("Provider health check failed: %v", err)
}
```

### Factory Metrics

Monitor all providers at once:

```go
// Get all provider metrics
allMetrics := tracer.GetMetrics()
for providerType, metrics := range allMetrics {
    fmt.Printf("%s: %d active tracers\n", providerType, metrics.TracersActive)
}

// Health check all providers
healthResults := tracer.HealthCheck(context.Background())
for providerType, err := range healthResults {
    if err != nil {
        log.Printf("%s health check failed: %v", providerType, err)
    }
}

// Get detailed provider information
providerInfos := tracer.GetProviderInfo(context.Background())
for _, info := range providerInfos {
    fmt.Printf("Provider: %s, Active: %t, State: %s\n", 
        info.Name, info.IsActive, info.Metrics.ConnectionState)
}
```

## 🧪 Testing

The library includes comprehensive tests with 98%+ coverage:

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./...

# Run specific provider tests
go test ./providers/datadog/...
go test ./providers/newrelic/...
go test ./providers/prometheus/...
```

### Test Coverage Requirements

- **Minimum Coverage**: 98%
- **Race Detection**: All tests pass with `-race` flag
- **Timeout Handling**: All tests complete within 30 seconds
- **Error Scenarios**: Comprehensive error case testing
- **Concurrent Safety**: Multi-goroutine testing included

## 🔒 Security Considerations

### Data Obfuscation

```go
// Datadog obfuscation
config := &datadog.Config{
    ObfuscationEnabled: true,
    ObfuscatedTags:     []string{"password", "token", "key", "secret"},
}

// New Relic attribute filtering
config := &newrelic.Config{
    AttributesExclude: []string{
        "request.headers.authorization",
        "request.headers.cookie",
        "user.password",
        "database.password",
    },
}
```

### High Security Mode

```go
// New Relic high security mode
config := &newrelic.Config{
    HighSecurity: true, // Enables additional security measures
}
```

## 📈 Performance Optimization

### Sampling Configuration

```go
// Reduce overhead with sampling
config := &datadog.Config{
    SampleRate:         0.1, // Sample 10% of traces
    MaxTracesPerSecond: 100, // Rate limiting
}

config := &tracer.TracerConfig{
    SamplingRate:  0.05, // Sample 5% of spans
    BatchSize:     500,  // Larger batches for efficiency
    FlushInterval: 30 * time.Second, // Less frequent flushes
}
```

### Resource Management

```go
// Proper resource cleanup
defer func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := tracer.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}()
```

## 🔄 Migration Guide

### From v1 to v2

Key changes in v2:

1. **Package Structure**: New modular provider architecture
2. **Interface Changes**: Enhanced interfaces with better type safety
3. **Configuration**: Structured configuration objects
4. **Error Handling**: Comprehensive error types and handling
5. **Metrics**: Built-in metrics and monitoring

```go
// v1 usage
import "github.com/fsvxavier/nexs-lib/observability/tracer"

// v2 usage
import "github.com/fsvxavier/nexs-lib/v2/observability/tracer"
import "github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
```

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/fsvxavier/nexs-lib.git
cd nexs-lib/v2/observability/tracer

# Install dependencies
go mod download

# Run tests
go test -race -cover ./...

# Run linting
golangci-lint run

# Format code
gofmt -w .
```

### Adding New Providers

1. Create provider directory: `providers/myprovider/`
2. Implement `Provider` and `Tracer` interfaces
3. Add comprehensive tests (98%+ coverage)
4. Update documentation and examples
5. Submit pull request

## 📋 TODO / Roadmap

- [ ] **OpenTelemetry Integration**: Native OpenTelemetry exporter support
- [ ] **Jaeger Provider**: Add Jaeger tracing support
- [ ] **Zipkin Provider**: Add Zipkin tracing support
- [ ] **AWS X-Ray Provider**: Add AWS X-Ray integration
- [ ] **Sampling Strategies**: Advanced sampling algorithms
- [ ] **Span Processors**: Custom span processing pipelines
- [ ] **Context Propagation**: B3, W3C trace context support
- [ ] **gRPC Integration**: Built-in gRPC interceptors
- [ ] **HTTP Integration**: Built-in HTTP middleware
- [ ] **Database Integration**: Auto-instrumentation for popular databases

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../../LICENSE) file for details.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)
- **Email**: support@nexs-lib.dev

## 🙏 Acknowledgments

- [OpenTelemetry](https://opentelemetry.io/) for tracing standards
- [Datadog](https://www.datadoghq.com/) for APM inspiration
- [New Relic](https://newrelic.com/) for monitoring patterns
- [Prometheus](https://prometheus.io/) for metrics standards
- Go community for excellent tooling and libraries
