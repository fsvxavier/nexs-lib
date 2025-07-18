# HTTP Integration Example

This example demonstrates how to integrate domain errors with HTTP handlers and properly map them to HTTP status codes and error responses.

## Features

- **Error Mapping**: Converts domain errors to appropriate HTTP status codes
- **Structured Responses**: Returns consistent JSON error responses with detailed information
- **User Service**: Simulates a user service with various error scenarios
- **Multiple Error Types**: Demonstrates handling of validation, business, not found, conflict, timeout, and external service errors

## Running the Example

```bash
cd /home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/domainerrors/examples/http-integration
go run main.go
```

The server will start on port 8080. Visit `http://localhost:8080` for usage instructions.

## Error Scenarios

### 1. Validation Errors (400 Bad Request)
```bash
# Invalid user ID
curl http://localhost:8080/users/invalid

# Invalid JSON
curl -X POST http://localhost:8080/users -d 'invalid json'

# Missing required fields
curl -X POST http://localhost:8080/users -d '{"name":"","email":"","age":-1}'
```

### 2. Business Rule Errors (422 Unprocessable Entity)
```bash
# Underage user
curl -X POST http://localhost:8080/users -d '{"id":10,"name":"John","email":"john@test.com","age":15}'
```

### 3. Not Found Errors (404 Not Found)
```bash
# User not found
curl http://localhost:8080/users/404
```

### 4. Conflict Errors (409 Conflict)
```bash
# Try to create user with existing email
curl -X POST http://localhost:8080/users -d '{"id":20,"name":"John","email":"john@example.com","age":25}'
```

### 5. Timeout Errors (504 Gateway Timeout)
```bash
# Database timeout simulation
curl http://localhost:8080/users/999
```

### 6. External Service Errors (502 Bad Gateway)
```bash
# External service error
curl http://localhost:8080/users/888
```

### 7. Rate Limit Errors (429 Too Many Requests)
```bash
# Rate limit exceeded
curl http://localhost:8080/rate-limit
```

### 8. Authentication Errors (401 Unauthorized)
```bash
# Invalid token
curl http://localhost:8080/auth-error
```

### 9. Server Errors (500 Internal Server Error)
```bash
# Internal server error
curl http://localhost:8080/server-error
```

## Error Response Format

All errors follow a consistent JSON format:

```json
{
  "error": "domain_error",
  "code": "USER_NOT_FOUND",
  "message": "User not found",
  "details": {
    "resource_type": "user",
    "resource_id": "404"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Key Components

### UserService
- Simulates a user service with CRUD operations
- Demonstrates various error scenarios
- Shows proper error handling patterns

### Error Handler
- Converts domain errors to HTTP responses
- Maps error types to appropriate HTTP status codes
- Extracts and formats error details

### HTTP Handlers
- Properly handle request parsing
- Use domain errors for validation
- Return structured error responses

## Error Type Mapping

| Domain Error Type | HTTP Status Code | Description |
|------------------|------------------|-------------|
| ValidationError | 400 | Bad Request - Invalid input data |
| BusinessError | 422 | Unprocessable Entity - Business rule violation |
| NotFoundError | 404 | Not Found - Resource doesn't exist |
| ConflictError | 409 | Conflict - Resource already exists |
| TimeoutError | 504 | Gateway Timeout - Operation timed out |
| ExternalServiceError | 502 | Bad Gateway - External service failed |
| RateLimitError | 429 | Too Many Requests - Rate limit exceeded |
| AuthenticationError | 401 | Unauthorized - Authentication failed |
| AuthorizationError | 403 | Forbidden - Authorization failed |
| ServerError | 500 | Internal Server Error - Server-side error |

## Best Practices Demonstrated

1. **Consistent Error Handling**: All errors follow the same pattern
2. **Detailed Error Information**: Errors include relevant context
3. **Proper HTTP Status Codes**: Each error type maps to appropriate status
4. **Structured Responses**: JSON format with consistent fields
5. **Error Context**: Includes timestamps and request tracing information
6. **Separation of Concerns**: Business logic separate from HTTP handling

This example shows how to build robust HTTP APIs with proper error handling using domain errors.
