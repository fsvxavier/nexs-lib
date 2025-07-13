# Nexs-Lib v2 Tracer Examples

This directory contains comprehensive examples demonstrating how to use the Nexs-Lib v2 observability tracer with different providers and configurations.

## Available Examples

### 📊 [Datadog Provider](./datadog/)
Complete HTTP server example using Datadog APM for distributed tracing.

**Features Demonstrated:**
- HTTP request tracing with span hierarchy
- Error handling and classification
- Business metrics integration
- Database operation tracking
- Authentication flow tracing
- Health check monitoring

**Best For:** APM-focused monitoring, service maps, real user monitoring

### 📈 [New Relic Provider](./newrelic/)
Comprehensive business transaction monitoring with New Relic APM.

**Features Demonstrated:**
- Business transaction tracking
- Custom insights and events
- Error analytics and classification
- Performance monitoring
- User behavior tracking
- Revenue impact analysis

**Best For:** Business intelligence, application insights, deployment tracking

### 🎯 [Prometheus Provider](./prometheus/)
Metrics-based monitoring and alerting with Prometheus.

**Features Demonstrated:**
- Custom metrics collection
- Histogram and counter tracking
- Business metrics for alerting
- Performance SLA monitoring
- Inventory level tracking
- Resource utilization monitoring

**Best For:** Metrics collection, alerting, cost-effective monitoring

### 🔄 [Multi-Provider](./multi-provider/)
Advanced example using multiple providers simultaneously.

**Features Demonstrated:**
- Simultaneous tracing to multiple providers
- Provider failover and redundancy
- Cross-provider correlation
- Performance comparison
- Cost optimization strategies

**Best For:** High availability, vendor independence, cost optimization

### 🏗️ [Microservices](./microservices/)
Complete microservices architecture with distributed tracing.

**Features Demonstrated:**
- Service-to-service communication
- Context propagation across services
- Load balancing and service discovery
- Authentication and authorization flows
- Database integration patterns
- Health check monitoring

**Best For:** Distributed systems, service mesh monitoring, DevOps

### 🚀 [gRPC & Message Queue](./grpc-messagequeue/)
Advanced integration with gRPC services and message queues.

**Features Demonstrated:**
- gRPC client and server tracing
- Asynchronous message processing
- Producer and consumer patterns
- Retry logic and error handling
- Dead letter queue management
- Background job processing

**Best For:** Event-driven architectures, async processing, gRPC services

### ⚡ [Performance & Benchmark](./performance-benchmark/)
Performance monitoring and benchmarking with detailed metrics.

**Features Demonstrated:**
- Load testing scenarios
- Performance SLA monitoring
- Latency percentile tracking (P50, P95, P99)
- Throughput optimization
- Memory and resource monitoring
- Stress testing patterns

**Best For:** Performance optimization, SLA compliance, capacity planning

### 🛡️ [Edge Cases & Error Handling](./edge-cases-error-handling/)
Comprehensive error handling and edge case scenarios.

**Features Demonstrated:**
- Network failure simulation
- Resource exhaustion scenarios
- Retry patterns and circuit breakers
- Graceful degradation
- Error recovery mechanisms
- Resilience testing

**Best For:** Production readiness, fault tolerance, reliability engineering

### 🔥 [OpenTelemetry Integration](./opentelemetry/)
Native OpenTelemetry SDK integration with OTLP exporters and W3C trace context.

**Features Demonstrated:**
- OTLP exporters (HTTP/gRPC)
- W3C trace context propagation
- Resource detection and configuration
- Batch processing optimization
- Sampling strategies
- Context propagation patterns
- Semantic conventions
- Error recording with context

**Best For:** OpenTelemetry adoption, vendor neutrality, standardization

### ⚠️ [Enhanced Error Handling](./error_handling_example/)
Advanced error handling patterns with circuit breakers and intelligent retry mechanisms.

**Features Demonstrated:**
- Circuit breaker pattern implementation
- Exponential backoff with jitter
- Intelligent error classification
- Retry strategies for different error types
- Context-aware error handling
- Metrics and monitoring for error patterns
- Graceful degradation patterns

**Best For:** Resilient systems, fault tolerance, production reliability

### 🎯 [Complete Integration](./complete-integration/)
All Critical Improvements working together in production scenarios.

**Features Demonstrated:**
- OpenTelemetry + Error Handling + Performance
- Production-ready configuration
- Real-world usage patterns
- Comprehensive monitoring
- Full observability stack
- Enterprise-grade deployment

**Best For:** Production deployment, comprehensive observability, reference implementation

### ⚡ [Performance Optimizations](./performance/)
High-performance span operations with object pooling and zero-allocation paths.

**Features Demonstrated:**
- Span object pooling (138x improvement)
- Zero-allocation fast paths
- Memory usage optimization
- Concurrent performance testing
- Throughput benchmarking
- Resource efficiency patterns

**Best For:** High-throughput systems, performance optimization, resource efficiency

### 🎛️ [Multi-Provider](./multi-provider/)
Using multiple providers simultaneously for redundancy and feature comparison.

**Best For:** Provider comparison, redundancy, migration scenarios

## Quick Start

### 1. Choose Your Provider

```bash
# For Datadog APM
cd datadog && go run main.go

# For New Relic APM  
cd newrelic && go run main.go

# For Prometheus metrics
cd prometheus && go run main.go

# For multiple providers
cd multi-provider && go run main.go
```

### 2. Test the Endpoints

Each example includes these common endpoints:
- `/` - Home page with basic tracing
- `/api/*` - API endpoints with business logic
- `/health` - Health check monitoring
- `/metrics` - Prometheus metrics (where applicable)

### 3. View Your Data

#### Datadog
- **APM**: [app.datadoghq.com/apm](https://app.datadoghq.com/apm)
- **Service Map**: Automatic service dependency mapping
- **Traces**: Detailed request tracing with performance metrics

#### New Relic
- **APM**: [one.newrelic.com](https://one.newrelic.com)
- **Transactions**: Business transaction monitoring
- **Insights**: Custom dashboards and analytics

#### Prometheus
- **Metrics**: `http://localhost:8080/metrics`
- **Prometheus UI**: `http://localhost:9090` (if running Prometheus server)
- **Grafana**: `http://localhost:3000` (if running Grafana)

## Example Comparison

| Feature | Datadog | New Relic | Prometheus | Multi-Provider |
|---------|---------|-----------|------------|----------------|
| **Complexity** | Medium | Medium | Medium | High |
| **Setup Time** | 5 min | 5 min | 2 min | 10 min |
| **External Deps** | Agent | License Key | None | Variable |
| **Best Use Case** | APM | Business | Metrics | Migration |

## Configuration Examples

### Basic Configuration
```go
// Minimal setup for development
config := &provider.Config{
    ServiceName: "my-service",
    Environment: "development",
}
```

### Production Configuration
```go
// Production-ready setup
config := &provider.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.2.3",
    Environment:    "production",
    SampleRate:     0.1,  // 10% sampling
    // Provider-specific settings...
}
```

### High-Volume Configuration
```go
// Optimized for high-traffic applications
config := &provider.Config{
    ServiceName:        "high-traffic-service",
    SampleRate:         0.01,  // 1% sampling
    MaxTracesPerSecond: 100,
    BatchSize:          1000,
    FlushInterval:      30 * time.Second,
}
```

## Integration Patterns

### HTTP Middleware
```go
func tracingMiddleware(tr tracer.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := tr.StartSpan(c.Request.Context(), c.FullPath())
        defer span.End()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        
        span.SetAttribute("http.status_code", c.Writer.Status())
    }
}
```

### Database Integration
```go
func queryWithTracing(ctx context.Context, tr tracer.Tracer, query string) error {
    _, span := tr.StartSpan(ctx, "db.query",
        tracer.WithSpanKind(tracer.SpanKindClient),
        tracer.WithSpanAttributes(map[string]interface{}{
            "db.statement": query,
        }),
    )
    defer span.End()
    
    // Execute query...
    return nil
}
```

### Business Logic Tracking
```go
func processOrder(ctx context.Context, tr tracer.Tracer, order Order) error {
    _, span := tr.StartSpan(ctx, "business.process_order",
        tracer.WithSpanAttributes(map[string]interface{}{
            "order.id":    order.ID,
            "order.value": order.Total,
            "user.id":     order.UserID,
        }),
    )
    defer span.End()
    
    // Business logic...
    span.SetAttribute("order.status", "processed")
    return nil
}
```

## Testing and Validation

### Load Testing
```bash
# Generate load to test tracing overhead
hey -n 1000 -c 10 http://localhost:8080/api/orders
```

### Trace Validation
```bash
# Check trace data in each provider
curl "http://localhost:8080/api/orders?trace_id=test123"
```

### Performance Monitoring
```bash
# Monitor resource usage during tracing
top -p $(pgrep -f "go run main.go")
```

## Common Issues and Solutions

### 1. High Overhead
**Problem:** Tracing adds significant latency
**Solutions:**
- Reduce sampling rate
- Use asynchronous span processing
- Optimize attribute collection

### 2. Missing Traces
**Problem:** Traces not appearing in provider
**Solutions:**
- Check provider connectivity
- Verify authentication credentials
- Increase flush frequency

### 3. Inconsistent Data
**Problem:** Different data across providers
**Solutions:**
- Use consistent trace IDs
- Standardize attribute naming
- Synchronize timestamps

## Development Workflow

### 1. Local Development
```bash
# Start with Prometheus for immediate feedback
cd prometheus
go run main.go
curl http://localhost:8080/metrics
```

### 2. Integration Testing
```bash
# Test with your target provider
cd datadog  # or newrelic
go run main.go
# Verify data in provider UI
```

### 3. Production Deployment
```bash
# Use multi-provider for redundancy
cd multi-provider
go run main.go
# Monitor all providers for consistency
```

## Best Practices

### 1. Start Simple
- Begin with single provider
- Add complexity gradually
- Monitor performance impact

### 2. Consistent Naming
```go
// Use consistent attribute names
span.SetAttribute("user.id", userID)        // ✅ Good
span.SetAttribute("userId", userID)         // ❌ Inconsistent
span.SetAttribute("user_identifier", userID) // ❌ Inconsistent
```

### 3. Meaningful Spans
```go
// Create spans for meaningful operations
_, span := tr.StartSpan(ctx, "payment.process")     // ✅ Good
_, span := tr.StartSpan(ctx, "function_call")       // ❌ Too generic
```

### 4. Error Handling
```go
if err != nil {
    span.SetStatus(tracer.StatusCodeError, err.Error())
    span.SetAttribute("error", true)
    span.SetAttribute("error.type", "validation_error")
    return err
}
```

### 5. Resource Cleanup
```go
defer func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := provider.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}()
```

## Performance Benchmarks

| Scenario | Requests/sec | Latency (p95) | Memory Usage |
|----------|--------------|---------------|--------------|
| No Tracing | 10,000 | 50ms | 25MB |
| Single Provider | 8,500 | 65ms | 45MB |
| Multi-Provider | 6,000 | 85ms | 75MB |

*Benchmarks run on 4-core, 8GB RAM environment*

## Next Steps

1. **Choose Your Provider**: Start with the provider that matches your observability goals
2. **Integrate Gradually**: Begin with critical endpoints, expand coverage over time
3. **Monitor Impact**: Track performance overhead and adjust sampling as needed
4. **Build Dashboards**: Create meaningful dashboards and alerts
5. **Scale Up**: Move to multi-provider setup for production redundancy

## Support and Resources

- **Main Documentation**: [../../README.md](../../README.md)
- **API Reference**: [../../interfaces.go](../../interfaces.go)
- **Issue Reporting**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

## Contributing

Found an issue or want to add an example?
1. Fork the repository
2. Create a new example in this directory
3. Follow the existing patterns and documentation style
4. Submit a pull request with comprehensive tests and documentation
