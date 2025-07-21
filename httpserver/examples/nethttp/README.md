# Net/HTTP Server Example

This example demonstrates how to use the nexs-lib httpserver with the net/http provider.

## Features Demonstrated

- **Factory Pattern**: Registering and creating servers by name
- **Observer Pattern**: Logging and metrics collection
- **Configuration**: Flexible server configuration
- **Graceful Shutdown**: Proper server lifecycle management

## Running the Example

```bash
lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "Port 8080 is free"
cd examples/nethttp
go run main.go
```

## Testing the Server

Once the server is running, you can test it:

```bash
# Basic endpoint
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Metrics endpoint
curl http://localhost:8080/metrics
```

## Code Walkthrough

1. **Provider Registration**: The net/http provider is registered with the httpserver registry
2. **Observer Attachment**: Logging and metrics observers are attached to monitor server events
3. **Configuration**: Server is configured with host, port, and timeouts
4. **Handler Setup**: HTTP handlers are created and set on the server
5. **Server Lifecycle**: Server is started and gracefully shut down on interrupt

## Key Components

- `httpserver.Register()`: Registers a server provider
- `httpserver.AttachObserver()`: Attaches lifecycle observers
- `config.DefaultConfig()`: Creates default configuration
- `httpserver.Create()`: Creates a server instance
- `server.Start()` / `server.Stop()`: Server lifecycle management

This example showcases the extensible architecture that allows different HTTP frameworks to be used interchangeably while maintaining consistent interfaces and observability.
