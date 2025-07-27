# Fiber Provider Example

This example demonstrates how to use the HTTP client with the Fiber provider, which uses the Fiber framework's HTTP client capabilities.

## Features Demonstrated

1. **Simple Client Creation**: Basic usage with Fiber backend
2. **JSON Operations**: Handling JSON requests and responses
3. **Performance Testing**: Measuring client performance
4. **Concurrent Requests**: Multi-threaded request handling
5. **Provider Metrics**: Performance monitoring and statistics

## Running the Example

```bash
cd examples/fiber
go run main.go
```

## Key Features of Fiber Provider

### Advantages
- **High Performance**: Built on FastHTTP for maximum speed
- **Low Memory Usage**: Efficient memory allocation
- **JSON Optimized**: Optimized JSON encoding/decoding
- **Fast Serialization**: Quick request/response processing
- **Lightweight**: Minimal overhead compared to standard HTTP

### Configuration Options
- **Timeout Settings**: Request and connection timeouts
- **JSON Handling**: Custom JSON marshaling/unmarshaling
- **Headers**: Default and custom headers
- **Performance Tuning**: Optimized for high throughput
- **Metrics**: Built-in performance monitoring

### Use Cases
- **High-Throughput APIs**: Excellent for high-volume API calls
- **Microservices**: Perfect for service-to-service communication
- **Real-time Applications**: Low latency requirements
- **JSON APIs**: Optimized for JSON-heavy workloads
- **Performance-Critical Systems**: When speed is paramount

## Example Configurations

### Basic Configuration
```go
client, err := httpclient.New(interfaces.ProviderFiber, "https://api.example.com")
```

### Performance-Optimized Configuration
```go
cfg := config.NewBuilder().
    WithBaseURL("https://api.example.com").
    WithTimeout(5 * time.Second).
    WithMetricsEnabled(true).
    WithTracingEnabled(false). // Disable for max performance
    WithHeader("Content-Type", "application/json").
    Build()

client, err := httpclient.NewWithConfig(interfaces.ProviderFiber, cfg)
```

### JSON Operations
```go
// Struct-based JSON
type User struct {
    Name   string   `json:"name"`
    Email  string   `json:"email"`
    Tags   []string `json:"tags"`
}

user := User{
    Name:  "John Doe",
    Email: "john@example.com",
    Tags:  []string{"admin", "user"},
}

resp, err := client.Post(ctx, "/users", user)
```

## Performance Characteristics

The Fiber provider offers superior performance characteristics:

- **Latency**: Ultra-low latency due to FastHTTP backend
- **Throughput**: Extremely high request throughput
- **Memory**: Minimal memory allocation and garbage collection
- **CPU**: Low CPU usage per request
- **Scalability**: Excellent for high-concurrency scenarios

## Performance Comparison

| Metric | Fiber | NetHTTP | FastHTTP |
|--------|-------|---------|----------|
| Latency | Ultra Low | Low | Ultra Low |
| Throughput | Very High | High | Very High |
| Memory Usage | Very Low | Moderate | Very Low |
| JSON Performance | Excellent | Good | Excellent |
| Ease of Use | High | High | Medium |

## Concurrent Processing

The Fiber provider excels at handling concurrent requests:

```go
// Example: 100 concurrent workers, 10 requests each
const numWorkers = 100
const requestsPerWorker = 10

for i := 0; i < numWorkers; i++ {
    go func(workerID int) {
        for j := 0; j < requestsPerWorker; j++ {
            resp, err := client.Get(ctx, "/endpoint")
            // Handle response
        }
    }(i)
}
```

## Best Practices

1. **JSON Optimization**: Use structs with JSON tags for best performance
2. **Connection Reuse**: The provider automatically handles connection pooling
3. **Timeout Configuration**: Set appropriate timeouts for your use case
4. **Metrics Monitoring**: Enable metrics to monitor performance
5. **Error Handling**: Implement proper error handling for network issues
6. **Concurrent Safety**: The client is safe for concurrent use
7. **Memory Management**: The provider handles memory efficiently

## When to Use Fiber Provider

Choose the Fiber provider when:
- **Performance is Critical**: Need maximum speed and minimal latency
- **High Concurrency**: Handling many simultaneous requests
- **JSON-Heavy Workloads**: Working primarily with JSON APIs
- **Resource Constraints**: Limited memory or CPU resources
- **Real-time Systems**: Building real-time or low-latency applications

## Integration with Fiber Web Framework

The Fiber provider works excellently when your application also uses the Fiber web framework, providing consistency across your HTTP stack.

```go
// In a Fiber web application
app := fiber.New()

// Use Fiber HTTP client for external API calls
client, err := httpclient.New(interfaces.ProviderFiber, "https://external-api.com")

app.Get("/proxy", func(c *fiber.Ctx) error {
    resp, err := client.Get(c.Context(), "/data")
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.Send(resp.Body)
})
```
