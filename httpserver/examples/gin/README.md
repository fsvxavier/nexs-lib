# Gin HTTP Server Example

This example demonstrates how to use the Gin HTTP server provider with the nexs-lib httpserver framework.

## Features

- **Gin Framework**: Uses the native Gin web framework with its router and middleware
- **Observer Pattern**: Demonstrates logging and metrics observers
- **Graceful Shutdown**: Implements proper server shutdown handling
- **Multiple Routes**: Shows different endpoint examples
- **JSON Responses**: Returns structured JSON responses

## Usage

```bash
# Run the example
go run main.go

# Test the endpoints
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/users
curl -X POST http://localhost:8080/api/users
```

## Endpoints

- `GET /` - Welcome message
- `GET /health` - Health check endpoint
- `GET /api/users` - List users
- `POST /api/users` - Create user

## Configuration

The server is configured with:
- Host: localhost
- Port: 8080
- Read Timeout: 30 seconds
- Write Timeout: 30 seconds
- Framework: Gin with native engine and middleware

## Framework Integration

This example showcases how the httpserver library:
1. **Adapts** standard `http.Handler` to work with Gin's native engine
2. **Observes** server lifecycle events through the Observer pattern
3. **Registers** framework-specific providers using the Factory pattern
4. **Manages** configuration through the extensible config system
