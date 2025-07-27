# FastHTTP Provider Example

This example demonstrates how to use the HTTP client with the FastHTTP provider, which uses the high-performance FastHTTP library for maximum speed and efficiency.

## Features Demonstrated

1. **Simple Client Creation**: Basic FastHTTP client usage
2. **High-Performance Configuration**: Optimized settings for maximum throughput
3. **Stress Testing**: High-concurrency load testing capabilities
4. **Memory Efficiency**: Demonstration of FastHTTP's memory optimization

## Running the Example

```bash
cd examples/fasthttp
go run main.go
```

## Key Features of FastHTTP Provider

### Advantages
- **Extreme Performance**: Up to 10x faster than net/http in some scenarios
- **Zero Memory Allocations**: Optimized to minimize garbage collection
- **High Concurrency**: Excellent for handling thousands of concurrent requests
- **Low Latency**: Minimal overhead and processing time
- **Memory Efficiency**: Reuses objects and minimizes allocations

### Performance Characteristics
- **Throughput**: Can handle 100,000+ requests/second
- **Latency**: Sub-millisecond response times for simple requests
- **Memory**: Minimal memory footprint and GC pressure
- **CPU**: Low CPU usage per request
- **Scalability**: Linear scaling with available resources

### Configuration Options
- **Connection Pooling**: Aggressive connection reuse
- **Timeout Settings**: Fine-grained timeout controls
- **Buffer Management**: Optimized buffer reuse
- **Header Handling**: Efficient header processing
- **Compression**: Built-in compression support

## Use Cases

### Perfect For
- **High-Frequency Trading**: Ultra-low latency requirements
- **API Gateways**: High-throughput proxy services
- **Load Testing Tools**: Generating massive request loads
- **Real-time Systems**: Time-critical applications
- **Microservices**: High-performance service communication

### Consider Alternatives When
- **HTTP/2 Required**: FastHTTP doesn't support HTTP/2
- **Complex TLS**: Advanced TLS features may be limited
- **Standard Compliance**: Some edge cases of HTTP spec
- **Debugging**: Less debugging tools compared to net/http

## Performance Benchmarks

Typical performance improvements over net/http:

| Metric | FastHTTP | net/http | Improvement |
|--------|----------|----------|-------------|
| Requests/sec | 100,000+ | 20,000+ | 5-10x |
| Memory Usage | Very Low | Moderate | 3-5x less |
| Latency | Ultra Low | Low | 2-3x faster |
| CPU Usage | Very Low | Moderate | 2-4x less |
| GC Pressure | Minimal | Moderate | 10x less |

## Example Configurations

### Maximum Performance
```go
cfg := config.NewBuilder().
    WithBaseURL("https://api.example.com").
    WithTimeout(1 * time.Second).        // Aggressive timeout
    WithMaxIdleConns(1000).              // Large connection pool
    WithIdleConnTimeout(60 * time.Second). // Keep connections alive
    WithMetricsEnabled(true).
    WithTracingEnabled(false).           // Disable for max speed
    Build()

client, err := httpclient.NewWithConfig(interfaces.ProviderFastHTTP, cfg)
```

### High Concurrency
```go
// Handle 10,000 concurrent requests
const numWorkers = 100
const requestsPerWorker = 100

var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := 0; j < requestsPerWorker; j++ {
            resp, err := client.Get(ctx, "/endpoint")
            // Process response
        }
    }()
}
wg.Wait()
```

## Memory Management

FastHTTP's memory efficiency comes from:

1. **Object Pooling**: Reuses request/response objects
2. **Zero Copy**: Avoids unnecessary data copying
3. **Buffer Reuse**: Reuses internal buffers
4. **Minimal Allocations**: Designed to minimize heap allocations

### Memory Usage Example
```go
// This creates minimal garbage
for i := 0; i < 10000; i++ {
    resp, err := client.Get(ctx, "/endpoint")
    if err != nil {
        continue
    }
    // Process resp.Body (avoid creating strings unnecessarily)
    processBytes(resp.Body)
}
```

## Best Practices

### DO
1. **Reuse Clients**: Create once, use many times
2. **Use Connection Pooling**: Configure appropriate pool sizes
3. **Handle Errors Gracefully**: Network errors are common at high scale
4. **Monitor Metrics**: Track performance and errors
5. **Test Under Load**: Validate performance under realistic conditions
6. **Avoid String Conversions**: Work with []byte when possible

### DON'T
1. **Create Many Clients**: Stick to one client per service
2. **Ignore Timeouts**: Always set reasonable timeouts
3. **Convert []byte to string**: Avoid unnecessary allocations
4. **Use for HTTP/2**: FastHTTP doesn't support HTTP/2
5. **Expect net/http Compatibility**: Some behaviors differ

## Monitoring and Debugging

### Built-in Metrics
```go
metrics := client.GetProvider().GetMetrics()
fmt.Printf("Requests: %d\n", metrics.TotalRequests)
fmt.Printf("Success Rate: %.2f%%\n", 
    float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
fmt.Printf("Avg Latency: %v\n", metrics.AverageLatency)
```

### Performance Monitoring
- Monitor connection pool utilization
- Track request/response sizes
- Watch for timeout patterns
- Observe memory usage and GC frequency

## Common Patterns

### Streaming Large Responses
```go
resp, err := client.Get(ctx, "/large-file")
if err != nil {
    return err
}

// Process response body in chunks to avoid loading all into memory
// (Implementation depends on your specific needs)
```

### Batch Operations
```go
// Process multiple requests efficiently
urls := []string{"/endpoint1", "/endpoint2", "/endpoint3"}
results := make(chan *interfaces.Response, len(urls))

for _, url := range urls {
    go func(u string) {
        resp, err := client.Get(ctx, u)
        if err == nil {
            results <- resp
        }
    }(url)
}

// Collect results
for i := 0; i < len(urls); i++ {
    select {
    case resp := <-results:
        // Process response
    case <-time.After(5 * time.Second):
        // Handle timeout
    }
}
```

## Troubleshooting

### Common Issues
1. **Connection Pool Exhaustion**: Increase MaxIdleConns
2. **Timeout Errors**: Adjust timeout settings
3. **Memory Leaks**: Ensure proper response handling
4. **High CPU**: Check for excessive string conversions

### Performance Tuning
1. **Profile Your Application**: Use Go's built-in profiler
2. **Benchmark Different Configurations**: Find optimal settings
3. **Monitor Resource Usage**: Track CPU, memory, and network
4. **Load Test**: Validate performance under realistic conditions

The FastHTTP provider is the best choice when maximum performance is required and you can accept some limitations compared to the standard net/http package.
