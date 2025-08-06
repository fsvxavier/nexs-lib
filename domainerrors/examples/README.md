# Domain Errors Examples

This directory contains comprehensive examples demonstrating how to use the Domain Errors library in various scenarios. Each example focuses on different aspects of error handling in domain-driven design.

## Available Examples

### 1. Basic Usage (`basic/`)
**Purpose**: Introduction to basic domain error usage
**Key Features**:
- Creating different types of domain errors
- Basic error handling patterns
- Error code and message usage
- Simple error type demonstrations

**Best For**: Getting started with the library, understanding basic concepts

### 2. Advanced Usage (`advanced/`)
**Purpose**: Complex error handling scenarios
**Key Features**:
- Error composition and chaining
- Context-aware error handling
- Custom error types and patterns
- Error aggregation and collection

**Best For**: Production applications requiring sophisticated error handling

### 3. Specific Errors (`specific-errors/`)
**Purpose**: Comprehensive demonstration of all 26 error types
**Key Features**:
- All error types with real-world examples
- Proper error construction with context
- Error-specific method usage
- Complete error type coverage

**Best For**: Understanding all available error types and their use cases

### 4. HTTP Integration (`http-integration/`)
**Purpose**: Web API error handling
**Key Features**:
- HTTP status code mapping
- JSON error responses
- REST API error handling
- Structured error responses

**Best For**: Web applications, REST APIs, HTTP services

### 5. Error Recovery (`error-recovery/`)
**Purpose**: Resilient error handling patterns
**Key Features**:
- Retry mechanisms with backoff
- Circuit breaker patterns
- Fallback strategies
- Bulk operation error handling

**Best For**: Microservices, distributed systems, resilient applications

### 7. **Hooks and Middleware (`hooks-middleware/`)**
**Purpose**: Advanced error processing with hooks and middleware systems
**Key Features**:
- Event-driven error processing with hooks
- Chain of responsibility pattern with middleware
- Real-world logging and audit scenarios
- Error enrichment pipelines
- Circuit breaker, rate limiting, and security patterns

**Best For**: Production applications requiring sophisticated error processing, monitoring, and transformation

### 8. Serialization (`serialization/`)
**Purpose**: Cross-system error communication
**Key Features**:
- JSON and XML serialization
- Error reconstruction
- Error collections
- Compact serialization formats

**Best For**: Cross-service communication, logging, monitoring

## Quick Start

1. **Clone the repository**:
```bash
git clone https://github.com/fsvxavier/nexs-lib.git
cd nexs-lib/domainerrors/examples
```

2. **Run any example**:
```bash
# Basic usage
cd basic && go run main.go

# HTTP integration
cd http-integration && go run main.go

# Error recovery patterns
cd error-recovery && go run main.go
```

3. **Run all examples**:
```bash
./run_all_examples.sh
```

## Example Structure

Each example directory contains:
- `main.go` - Main example code
- `README.md` - Detailed documentation
- Additional files as needed

## Error Types Covered

All examples demonstrate various combinations of these error types:

| Error Type | HTTP Status | Description |
|-----------|-------------|-------------|
| ValidationError | 400 | Input validation failures |
| BusinessError | 422 | Business rule violations |
| NotFoundError | 404 | Resource not found |
| ConflictError | 409 | Resource conflicts |
| TimeoutError | 504 | Operation timeouts |
| RateLimitError | 429 | Rate limit exceeded |
| ExternalServiceError | 502 | External service failures |
| DatabaseError | 500 | Database operation errors |
| AuthenticationError | 401 | Authentication failures |
| AuthorizationError | 403 | Authorization failures |
| ServerError | 500 | Internal server errors |
| NetworkError | 503 | Network-related errors |
| SecurityError | 403 | Security violations |
| ConfigurationError | 500 | Configuration errors |
| DependencyError | 503 | Dependency failures |
| ResourceExhaustedError | 429 | Resource exhaustion |
| CircuitBreakerError | 503 | Circuit breaker open |
| SerializationError | 400 | Serialization failures |
| MigrationError | 500 | Database migration errors |
| UnsupportedOperationError | 501 | Unsupported operations |
| PerformanceError | 503 | Performance degradation |
| DataIntegrityError | 409 | Data integrity violations |
| UnprocessableEntityError | 422 | Unprocessable entities |
| PreconditionFailedError | 412 | Precondition failures |
| ServiceUnavailableError | 503 | Service unavailable |
| CacheError | 503 | Cache operation errors |

## Common Patterns

### 1. Error Creation
```go
// Basic error
err := domainerrors.NewValidationError("INVALID_EMAIL", "Invalid email format", nil)

// With context
err.WithField("email", "must be a valid email address")
```

### 2. Error Handling
```go
if err != nil {
    if domainErr, ok := err.(*domainerrors.DomainError); ok {
        // Handle based on error type
        switch domainErr.Type {
        case domainerrors.ErrorTypeValidation:
            // Handle validation error
        case domainerrors.ErrorTypeNotFound:
            // Handle not found error
        }
    }
}
```

### 3. HTTP Response
```go
func handleError(w http.ResponseWriter, err error) {
    if domainErr, ok := err.(*domainerrors.DomainError); ok {
        statusCode := domainErr.HTTPStatus()
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": domainErr.Code,
            "message": domainErr.Message,
        })
    }
}
```

## Integration Guidelines

### 1. Web Applications
- Use HTTP integration example as base
- Map errors to appropriate HTTP status codes
- Return structured JSON error responses
- Include request tracing information

### 2. Microservices
- Use error recovery patterns
- Implement circuit breakers
- Use serialization for cross-service communication
- Include correlation IDs for tracing

### 3. Database Applications
- Use specific database error types
- Implement retry mechanisms
- Handle connection failures gracefully
- Use proper error context

### 4. API Clients
- Use external service error types
- Implement timeout handling
- Use rate limiting error handling
- Provide fallback mechanisms

## Testing

Each example includes test scenarios and validation:
- Error creation and validation
- HTTP response testing
- Serialization round-trip testing
- Error recovery validation

## Performance Considerations

- **Error Creation**: Minimal overhead for error creation
- **Serialization**: Efficient JSON/XML serialization
- **HTTP Integration**: Fast error response generation
- **Memory Usage**: Optimized error structures

## Best Practices

1. **Use Specific Error Types**: Choose the most appropriate error type
2. **Include Context**: Add relevant context information
3. **Consistent Codes**: Use consistent error codes across services
4. **Proper HTTP Status**: Map to appropriate HTTP status codes
5. **Structured Responses**: Use consistent error response format
6. **Logging**: Log errors with appropriate detail level
7. **Monitoring**: Track error rates and types
8. **Testing**: Test error scenarios thoroughly

## Contributing

To add new examples:
1. Create a new directory with descriptive name
2. Include `main.go` with working code
3. Add comprehensive `README.md`
4. Update this main README
5. Add to `run_all_examples.sh`

## Dependencies

All examples use:
- Go 1.23+
- Domain Errors library
- Standard library packages
- No external dependencies (except for specific examples)

## License

These examples are part of the nexs-lib project and follow the same license terms.
