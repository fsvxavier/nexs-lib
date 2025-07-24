# Error Recovery Patterns Example

This example demonstrates comprehensive error recovery patterns and strategies using domain errors. It shows how to handle failures gracefully with retry mechanisms, circuit breakers, fallback strategies, and alternative storage options.

## Features

- **Retry with Backoff**: Implements exponential, linear, and jitter backoff strategies
- **Circuit Breaker**: Prevents cascading failures by temporarily stopping requests to failing services
- **Fallback Mechanisms**: Provides alternative data sources when primary sources fail
- **Alternative Storage**: Uses backup storage options when primary storage fails
- **Context-based Timeouts**: Proper timeout handling with context cancellation
- **Bulk Operation Handling**: Demonstrates error recovery in bulk operations

## Running the Example

```bash
cd /home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/domainerrors/examples/error-recovery
go run main.go
```

## Key Components

### 1. Retry Configuration
```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    BackoffType BackoffType
}
```

### 2. Backoff Strategies
- **Linear Backoff**: Increases delay linearly with each attempt
- **Exponential Backoff**: Doubles delay with each attempt
- **Jitter Backoff**: Adds randomness to prevent thundering herd

### 3. Circuit Breaker
- **Closed State**: Normal operation, allows all requests
- **Open State**: Fails fast, prevents requests to failing service
- **Half-Open State**: Allows limited requests to test if service recovered

### 4. Error Recovery Patterns

#### Retry with Backoff
```go
err := RetryWithBackoff(ctx, retryConfig, func() error {
    return riskyOperation()
})
```

#### Circuit Breaker Protection
```go
err := circuitBreaker.Execute(func() error {
    return externalServiceCall()
})
```

#### Fallback Strategy
```go
data, err := primarySource.GetData(id)
if err != nil {
    data, err = fallbackSource.GetData(id)
}
```

## Error Scenarios Demonstrated

### 1. External API Failures
- Network timeouts
- Service unavailable (5xx errors)
- Rate limiting
- Connection refused

### 2. Database Failures
- Connection failures
- Query timeouts
- Deadlocks
- Constraint violations

### 3. Recovery Strategies
- Automatic retries with backoff
- Circuit breaker protection
- Fallback data sources
- Alternative storage options

## Example Output

```
=== Error Recovery Patterns Example ===

Processing data for ID: user1
Used fallback data for ID: user1
✅ Successfully processed user1 (took 234ms)

Processing data for ID: user2
❌ Failed to process user2 (took 1.2s): DATA_SAVE_FAILED: Failed to save data
   Error Code: DATA_SAVE_FAILED
   Error Type: database
   HTTP Status: 500
   Cause: Query execution failed

Processing data for ID: user3
Used alternative storage for ID: user3
✅ Successfully processed user3 (took 456ms)
```

## Best Practices Demonstrated

### 1. Resilient Error Handling
- Multiple recovery strategies
- Graceful degradation
- Proper error propagation

### 2. Performance Optimization
- Exponential backoff to reduce load
- Circuit breaker to prevent cascading failures
- Timeout management to prevent hanging

### 3. Observability
- Detailed error logging
- Performance metrics
- Success/failure rates

### 4. Configuration Management
- Configurable retry policies
- Adjustable timeout values
- Flexible backoff strategies

## Error Recovery Strategies

### 1. Immediate Retry
- Quick retry for transient failures
- Useful for network blips
- Limited attempts to prevent infinite loops

### 2. Exponential Backoff
- Increases delay between retries
- Reduces load on failing services
- Prevents thundering herd problems

### 3. Circuit Breaker
- Temporarily disables failing services
- Allows system to recover
- Prevents resource exhaustion

### 4. Fallback Data
- Alternative data sources
- Cached/stale data
- Default values

### 5. Alternative Storage
- Backup storage systems
- Message queues for later processing
- File-based storage

## Integration Points

This example can be integrated with:
- HTTP APIs (see http-integration example)
- Logging systems for error tracking
- Monitoring systems for alerting
- Configuration management for tuning
- Testing frameworks for resilience testing

## Performance Considerations

- **Timeout Management**: Proper timeouts prevent resource leaks
- **Backoff Strategies**: Prevent overwhelming recovering services
- **Circuit Breaker**: Reduces load on failing dependencies
- **Fallback Performance**: Ensure fallbacks are faster than primary
- **Resource Cleanup**: Proper context cancellation and cleanup

This example demonstrates production-ready error recovery patterns that can significantly improve system resilience and user experience.
