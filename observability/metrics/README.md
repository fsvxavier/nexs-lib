# Metrics Module

A comprehensive, provider-agnostic metrics library for Go applications with support for Prometheus, DataDog, and NewRelic.

## Features

- **Multiple Providers**: Support for Prometheus, DataDog, and NewRelic
- **Standard Metric Types**: Counter, Histogram, Gauge, and Summary metrics
- **Labels Support**: Dynamic and constant labels for all metric types
- **Thread-Safe**: All operations are thread-safe and optimized for concurrent use
- **Timing Utilities**: Built-in timing functions for measuring durations
- **Graceful Shutdown**: Proper resource cleanup and data flushing
- **Extensible**: Easy to add new metric providers
- **Testing Support**: Comprehensive mocks and test utilities

## Metric Types

### Counter
Monotonically increasing metric, ideal for counting events like requests, errors, or completed tasks.

```go
counter, _ := provider.CreateCounter(metrics.CounterOptions{
    Name:      "http_requests_total",
    Help:      "Total number of HTTP requests",
    Labels:    []string{"method", "status"},
    Namespace: "http",
})

counter.Inc("GET", "200")
counter.Add(5.0, "POST", "201")
```

### Histogram
Samples observations and counts them in configurable buckets, perfect for request durations, response sizes, etc.

```go
histogram, _ := provider.CreateHistogram(metrics.HistogramOptions{
    Name:    "http_request_duration_seconds",
    Help:    "HTTP request duration in seconds",
    Labels:  []string{"method", "endpoint"},
    Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
})

histogram.Observe(0.5, "GET", "/api/users")
histogram.Time(func() {
    // Your code here
}, "POST", "/api/orders")
```

### Gauge
Metric that can go up and down, suitable for current values like memory usage, queue size, or temperature.

```go
gauge, _ := provider.CreateGauge(metrics.GaugeOptions{
    Name:   "memory_usage_bytes",
    Help:   "Current memory usage in bytes",
    Labels: []string{"type"},
})

gauge.Set(1024000, "heap")
gauge.Inc("stack")
gauge.Add(500000, "heap")
```

### Summary
Similar to histogram but calculates quantiles over a sliding time window.

```go
summary, _ := provider.CreateSummary(metrics.SummaryOptions{
    Name:       "response_time_seconds",
    Help:       "Response time summary",
    Labels:     []string{"service"},
    Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
})

summary.Observe(0.25, "user-service")
timer := summary.StartTimer("order-service")
// ... do work ...
timer()
```

## Providers

### Prometheus

```go
import "github.com/fsvxavier/nexs-lib/observability/metrics/providers/prometheus"

config := &prometheus.Config{
    Registry:  prometheus.NewRegistry(),
    Namespace: "myapp",
    Subsystem: "http",
    ConstLabels: map[string]string{
        "version":     "1.0.0",
        "environment": "production",
    },
}

provider := prometheus.NewProvider(config)
```

**Features:**
- Full Prometheus client library integration
- Custom registries support
- Namespace and subsystem organization
- Default bucket configurations
- HTTP handler for `/metrics` endpoint

### DataDog

```go
import "github.com/fsvxavier/nexs-lib/observability/metrics/providers/datadog"

config := &datadog.Config{
    APIKey:      "your-api-key",
    AppKey:      "your-app-key",
    Environment: "production",
    Service:     "myapp",
    Version:     "1.0.0",
    Tags:        []string{"env:prod", "team:backend"},
    Namespace:   "myapp",
    FlushPeriod: 10 * time.Second,
}

provider := datadog.NewProvider(config, nil) // nil uses mock client
```

**Features:**
- DataDog APM integration
- Custom tags and attributes
- Automatic metric name formatting
- Configurable flush periods
- Mock client for testing

### NewRelic

```go
import "github.com/fsvxavier/nexs-lib/observability/metrics/providers/newrelic"

config := &newrelic.Config{
    LicenseKey:  "your-license-key",
    AppName:     "MyApp",
    Environment: "production",
    Namespace:   "myapp",
    Attributes: map[string]interface{}{
        "version": "1.0.0",
        "region":  "us-east-1",
    },
}

provider := newrelic.NewProvider(config, nil) // nil uses mock client
```

**Features:**
- NewRelic APM integration
- Custom attributes and metadata
- Event recording capabilities
- Automatic metric naming
- Mock client for testing

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "time"
    
    "github.com/fsvxavier/nexs-lib/observability/metrics"
    "github.com/fsvxavier/nexs-lib/observability/metrics/providers/prometheus"
)

func main() {
    // Create provider
    provider := prometheus.NewProvider(nil)
    defer provider.Shutdown(context.Background())
    
    // Create metrics
    requests, _ := provider.CreateCounter(metrics.CounterOptions{
        Name:   "requests_total",
        Help:   "Total requests",
        Labels: []string{"method", "status"},
    })
    
    duration, _ := provider.CreateHistogram(metrics.HistogramOptions{
        Name:   "request_duration_seconds",
        Help:   "Request duration",
        Labels: []string{"method"},
    })
    
    // Use metrics
    requests.Inc("GET", "200")
    
    duration.Time(func() {
        time.Sleep(100 * time.Millisecond)
    }, "GET")
}
```

### HTTP Middleware Example

```go
func MetricsMiddleware(collector *MetricsCollector) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Increment active connections
            collector.activeConnections.Inc()
            defer collector.activeConnections.Dec()
            
            // Wrap response writer to capture status
            ww := &responseWriter{ResponseWriter: w, statusCode: 200}
            
            // Process request
            next.ServeHTTP(ww, r)
            
            // Record metrics
            duration := time.Since(start).Seconds()
            method := r.Method
            status := fmt.Sprintf("%d", ww.statusCode)
            endpoint := r.URL.Path
            
            collector.httpRequests.Inc(method, status, endpoint)
            collector.httpDuration.Observe(duration, method, endpoint)
            collector.responseTime.Observe(duration, method, endpoint)
        })
    }
}
```

### Multiple Providers

```go
func setupMetrics() ([]metrics.Provider, error) {
    var providers []metrics.Provider
    
    // Prometheus for internal monitoring
    promProvider := prometheus.NewProvider(&prometheus.Config{
        Namespace: "myapp",
    })
    providers = append(providers, promProvider)
    
    // DataDog for external monitoring
    ddProvider := datadog.NewProvider(&datadog.Config{
        Environment: "production",
        Service:     "myapp",
    }, nil)
    providers = append(providers, ddProvider)
    
    // NewRelic for APM
    nrProvider := newrelic.NewProvider(&newrelic.Config{
        AppName: "MyApp",
    }, nil)
    providers = append(providers, nrProvider)
    
    return providers, nil
}
```

## Testing

### Using Mocks

```go
import "github.com/fsvxavier/nexs-lib/observability/metrics/mocks"

func TestMetrics(t *testing.T) {
    provider := mocks.NewMockProvider("test")
    
    counter, _ := provider.CreateCounter(metrics.CounterOptions{
        Name:   "test_counter",
        Labels: []string{"method"},
    })
    
    // Record some metrics
    counter.Inc("GET")
    counter.Inc("GET")
    counter.Add(5.0, "POST")
    
    // Verify using mock methods
    mockCounter := provider.GetCounters()["counter__test_counter"]
    if mockCounter.GetIncCalls("GET") != 2 {
        t.Errorf("expected 2 inc calls for GET, got %d", 
            mockCounter.GetIncCalls("GET"))
    }
    
    if mockCounter.Get("POST") != 5.0 {
        t.Errorf("expected 5.0 for POST counter, got %f", 
            mockCounter.Get("POST"))
    }
}
```

### Provider-Specific Testing

```go
func TestPrometheusIntegration(t *testing.T) {
    provider := prometheus.NewProvider(nil)
    defer provider.Shutdown(context.Background())
    
    counter, _ := provider.CreateCounter(metrics.CounterOptions{
        Name: "test_counter",
    })
    
    counter.Inc()
    
    // Gather metrics from registry
    registry := provider.GetRegistry().(*prometheus.Registry)
    families, _ := registry.Gather()
    
    // Verify metrics were recorded
    // ... assertions ...
}
```

## Performance

The library is optimized for high-throughput applications:

- **Lock-free operations** where possible
- **Efficient label handling** with pre-allocated slices
- **Minimal allocations** in hot paths
- **Concurrent-safe** operations across all providers

### Benchmarks

```
BenchmarkCounterInc-8           10000000    150 ns/op    0 allocs/op
BenchmarkHistogramObserve-8      5000000    280 ns/op    0 allocs/op
BenchmarkGaugeSet-8             10000000    120 ns/op    0 allocs/op
BenchmarkSummaryObserve-8        3000000    450 ns/op    1 allocs/op
```

## Best Practices

### Metric Naming

- Use clear, descriptive names: `http_requests_total` instead of `requests`
- Follow provider conventions (Prometheus: snake_case, DataDog: dots)
- Include units in the name: `_seconds`, `_bytes`, `_total`

### Label Usage

- Keep labels cardinality low (< 1000 combinations)
- Use labels for dimensions, not values
- Prefer constant labels for static metadata

```go
// Good
counter.Inc("GET", "200", "/api/users")

// Bad - high cardinality
counter.Inc("GET", "200", "/api/users/12345")
```

### Error Handling

```go
counter, err := provider.CreateCounter(opts)
if err != nil {
    log.Printf("Failed to create counter: %v", err)
    // Use noop counter or handle gracefully
    counter = &metrics.NoopCounter{}
}
```

### Graceful Shutdown

```go
func (app *App) Shutdown() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    for _, provider := range app.providers {
        if err := provider.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down provider: %v", err)
        }
    }
}
```

## Configuration

### Environment Variables

```bash
# Prometheus
PROMETHEUS_NAMESPACE=myapp
PROMETHEUS_SUBSYSTEM=http

# DataDog
DATADOG_API_KEY=your-api-key
DATADOG_APP_KEY=your-app-key
DATADOG_ENV=production

# NewRelic
NEWRELIC_LICENSE_KEY=your-license-key
NEWRELIC_APP_NAME=MyApp
```

### YAML Configuration

```yaml
metrics:
  providers:
    prometheus:
      enabled: true
      namespace: myapp
      endpoint: ":8080/metrics"
    
    datadog:
      enabled: true
      api_key: ${DATADOG_API_KEY}
      environment: production
      flush_period: 10s
    
    newrelic:
      enabled: true
      license_key: ${NEWRELIC_LICENSE_KEY}
      app_name: MyApp
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Run benchmarks: `go test -bench=. ./...`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
