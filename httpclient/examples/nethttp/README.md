# NetHTTP Provider Example

This example demonstrates how to use the HTTP client with the NetHTTP provider, which uses Go's standard `net/http` library.

## Features Demonstrated

1. **Simple Client Creation**: Basic usage with minimal configuration
2. **Advanced Configuration**: Custom settings including timeouts, connection pools, headers
3. **Error Handling**: Custom error handlers and retry mechanisms
4. **Custom Headers**: Setting default and request-specific headers
5. **Different HTTP Methods**: GET, POST, PUT, DELETE, HEAD, OPTIONS
6. **Metrics and Tracing**: Provider performance metrics and distributed tracing

## Running the Example

```bash
cd examples/nethttp
go run main.go
```

## Key Features of NetHTTP Provider

### Advantages
- **Standard Library**: Uses Go's built-in `net/http` package
- **Mature and Stable**: Well-tested and widely used
- **Full HTTP/2 Support**: Native HTTP/2 support
- **Connection Pooling**: Efficient connection reuse
- **Detailed Tracing**: Support for request/response tracing
- **Context Support**: Full context.Context integration

### Configuration Options
- **Timeout Settings**: Request, TLS handshake, idle connection timeouts
- **Connection Pool**: Max idle connections, connection reuse
- **TLS Configuration**: Certificate verification, custom TLS settings
- **Headers**: Default headers for all requests
- **Retry Logic**: Configurable retry attempts with exponential backoff
- **Tracing**: Distributed tracing support
- **Metrics**: Request/response metrics collection

### Use Cases
- **General Purpose**: Suitable for most HTTP client needs
- **High Performance**: Good for high-throughput applications
- **Enterprise Applications**: Reliable for production systems
- **Microservices**: Excellent for service-to-service communication
- **API Integration**: Perfect for REST API consumption

## Example Configurations

### Basic Configuration
```go
client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://api.example.com")
```

### Advanced Configuration
```go
cfg := config.NewBuilder().
    WithBaseURL("https://api.example.com").
    WithTimeout(30 * time.Second).
    WithMaxIdleConns(100).
    WithHeader("User-Agent", "MyApp/1.0").
    WithMaxRetries(3).
    WithTracingEnabled(true).
    Build()

client, err := httpclient.NewWithConfig(interfaces.ProviderNetHTTP, cfg)
```

### Custom Error Handling
```go
client.SetErrorHandler(func(resp *interfaces.Response) error {
    if resp.StatusCode >= 400 {
        return fmt.Errorf("API error %d: %s", resp.StatusCode, string(resp.Body))
    }
    return nil
})
```

## Performance Characteristics

The NetHTTP provider offers excellent performance characteristics:

- **Latency**: Low latency for most requests
- **Throughput**: High throughput with connection pooling
- **Memory**: Efficient memory usage
- **CPU**: Moderate CPU usage
- **Scalability**: Excellent horizontal scaling

## Best Practices

1. **Connection Pooling**: Configure appropriate pool sizes for your workload
2. **Timeouts**: Set reasonable timeouts to prevent hanging requests
3. **Retry Logic**: Use exponential backoff for retries
4. **Headers**: Set common headers at the client level
5. **Context**: Always use context.Context for request cancellation
6. **Metrics**: Monitor request metrics for performance insights
7. **Error Handling**: Implement custom error handlers for better error management
