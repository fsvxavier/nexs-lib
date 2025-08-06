# Hooks and Middleware Examples

This directory contains comprehensive examples demonstrating how to use hooks and middleware with the Domain Errors library.

## Purpose

**Focus**: Advanced error handling with hooks and middleware systems
**Key Features**:
- Event-driven error processing with hooks
- Chain of responsibility pattern with middleware
- Real-world logging and audit scenarios
- Error enrichment pipelines
- Custom patterns for circuit breaker, rate limiting, and error transformation

**Best For**: Applications requiring sophisticated error processing, logging, monitoring, and transformation capabilities

## What You'll Learn

### 1. Hook System
- **Event-driven processing**: React to specific events in the error lifecycle
- **Hook types**: `before_error`, `after_error`, `before_metadata`, `after_metadata`, etc.
- **Side effects**: Logging, auditing, metrics collection, notifications
- **Non-intrusive**: Hooks observe and react but don't modify the error

### 2. Middleware System
- **Processing pipeline**: Transform and enrich errors through a chain
- **Chain of responsibility**: Sequential processing with `next()` function
- **Error transformation**: Modify error properties, add metadata, change messages
- **Order matters**: Middlewares execute in registration order

### 3. Combined Usage
- **Hooks + Middleware**: How they work together in the error lifecycle
- **Execution order**: Understanding when hooks and middleware execute
- **Best practices**: When to use hooks vs middleware

## Examples Overview

### Example 1: Basic Hook Registration
Demonstrates simple hook registration and execution:
```go
domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
    log.Printf("Error created: %s [%s]", err.Code, err.Type)
    return nil
})
```

### Example 2: Basic Middleware Chain
Shows middleware registration and error processing:
```go
domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
    // Enrich error with system information
    if err.Metadata == nil {
        err.Metadata = make(map[string]interface{})
    }
    err.Metadata["service"] = "user-service"
    err.Metadata["timestamp"] = time.Now()
    
    return next(err) // Call next middleware in chain
})
```

### Example 3: Complex Middleware Chain
Demonstrates multiple middlewares working together:
- **Validation middleware**: Checks error structure
- **Context enrichment**: Adds environment and request context
- **Logging middleware**: Records errors in structured format

### Example 4: Hooks and Middleware Combined
Shows how hooks and middleware work together:
- Hooks execute at specific lifecycle events
- Middleware processes errors in a chain
- Observe the execution order and interaction

### Example 5: Real-world Logging and Audit
Production-ready patterns:
- **Security audit hooks**: Special handling for authentication/authorization errors
- **Metrics middleware**: Automatic metrics collection
- **Severity classification**: Different handling based on error criticality

### Example 6: Error Enrichment Pipeline
Advanced error processing:
- **Request context**: HTTP method, path, IP address
- **User context**: User ID, roles, permissions
- **System context**: Hostname, version, environment, region

### Example 7: Custom Patterns
Advanced patterns for common scenarios:
- **Circuit breaker**: Track failures and open circuit when threshold reached
- **Error transformation**: Convert technical errors to user-friendly messages
- **Rate limiting**: Detect and handle error frequency

## Running the Examples

```bash
cd examples/hooks-middleware
go run main.go
```

## Key Concepts

### Hook Lifecycle Events
- `before_error` - Before error creation
- `after_error` - After error creation
- `before_metadata` - Before adding metadata
- `after_metadata` - After adding metadata
- `before_stack_trace` - Before capturing stack trace
- `after_stack_trace` - After capturing stack trace

### Middleware Chain Pattern
```
Error Creation → Middleware 1 → Middleware 2 → Middleware 3 → Final Error
                     ↓              ↓              ↓
                 Validation    Enrichment     Logging
```

### Execution Order
```
1. Error created
2. Middleware chain executes (transforms error)
3. before_* hooks execute (side effects)
4. Internal processing (metadata, stack trace, etc.)
5. after_* hooks execute (side effects)
6. Final error returned
```

## Real-world Use Cases

### 1. **API Error Handling**
- Enrich errors with request context
- Transform internal errors to API-friendly messages
- Log API errors for monitoring

### 2. **Security Auditing**
- Hook into authentication/authorization errors
- Log security events to SIEM systems
- Alert on suspicious patterns

### 3. **Monitoring & Observability**
- Collect metrics on error types and frequency
- Send errors to monitoring systems
- Create dashboards for error patterns

### 4. **Error Recovery**
- Circuit breaker pattern for external services
- Retry logic based on error types
- Graceful degradation

### 5. **Compliance & Logging**
- Structured logging for compliance
- PII sanitization in error messages
- Audit trails for error handling

## Best Practices

### When to Use Hooks:
- ✅ Logging and auditing
- ✅ Metrics collection
- ✅ Notifications and alerts
- ✅ Side effects that don't modify the error
- ✅ Event-driven reactions

### When to Use Middleware:
- ✅ Error enrichment and transformation
- ✅ Adding metadata or context
- ✅ Validation and sanitization
- ✅ Business logic processing
- ✅ Pipeline-based transformations

### Performance Considerations:
- 🔧 Hooks and middleware add overhead - use judiciously
- 🔧 Async hooks for non-critical operations
- 🔧 Avoid heavy processing in hooks/middleware
- 🔧 Consider pooling for high-frequency scenarios

## Integration Notes

This example shows the current implementation using the simplified hook/middleware system. In production:

1. **Setup once**: Register hooks and middleware during application startup
2. **Global state**: Hooks and middleware are typically global for the application
3. **Context passing**: Use context.Context for request-specific data
4. **Error handling**: Handle panics and errors within hooks/middleware
5. **Testing**: Mock or disable hooks/middleware in tests when needed

## Next Steps

After understanding hooks and middleware:
1. Explore other examples in the `examples/` directory
2. Check out the complete API documentation
3. Look at real-world integration patterns
4. Consider performance implications for your use case

---

**Note**: This example demonstrates the current API. Some functions referenced in comments may not be available in the current implementation and serve as documentation for future enhancements.
