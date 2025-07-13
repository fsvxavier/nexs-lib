package tracer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ErrorClassification represents different types of errors
type ErrorClassification int

const (
	ErrorClassificationUnknown ErrorClassification = iota
	ErrorClassificationNetwork
	ErrorClassificationTimeout
	ErrorClassificationAuth
	ErrorClassificationRateLimit
	ErrorClassificationInternal
	ErrorClassificationValidation
	ErrorClassificationResource
)

// String returns the string representation of the error classification
func (e ErrorClassification) String() string {
	switch e {
	case ErrorClassificationNetwork:
		return "NETWORK"
	case ErrorClassificationTimeout:
		return "TIMEOUT"
	case ErrorClassificationAuth:
		return "AUTH"
	case ErrorClassificationRateLimit:
		return "RATE_LIMIT"
	case ErrorClassificationInternal:
		return "INTERNAL"
	case ErrorClassificationValidation:
		return "VALIDATION"
	case ErrorClassificationResource:
		return "RESOURCE"
	default:
		return "UNKNOWN"
	}
}

// ErrorHandler provides enhanced error handling capabilities
type ErrorHandler interface {
	// ClassifyError classifies an error into predefined categories
	ClassifyError(err error) ErrorClassification

	// ShouldRetry determines if an operation should be retried based on error
	ShouldRetry(err error, attempt int) bool

	// HandleError processes an error and returns recovery actions
	HandleError(ctx context.Context, err error, operation string) error

	// GetRetryDelay calculates retry delay with exponential backoff and jitter
	GetRetryDelay(attempt int, baseDelay time.Duration) time.Duration
}

// DefaultErrorHandler implements ErrorHandler with comprehensive error handling
type DefaultErrorHandler struct {
	circuitBreaker CircuitBreaker
	retryConfig    RetryConfig
	mu             sync.RWMutex
	errorCounts    map[ErrorClassification]int64
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries      int                   `json:"max_retries"`
	BaseDelay       time.Duration         `json:"base_delay"`
	MaxDelay        time.Duration         `json:"max_delay"`
	JitterFactor    float64               `json:"jitter_factor"`
	BackoffFactor   float64               `json:"backoff_factor"`
	RetryableErrors []ErrorClassification `json:"retryable_errors"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		JitterFactor:  0.1,
		BackoffFactor: 2.0,
		RetryableErrors: []ErrorClassification{
			ErrorClassificationNetwork,
			ErrorClassificationTimeout,
			ErrorClassificationRateLimit,
			ErrorClassificationResource,
		},
	}
}

// CircuitBreaker interface for circuit breaker pattern
type CircuitBreaker interface {
	// Execute runs the operation with circuit breaker protection
	Execute(ctx context.Context, operation func() error) error

	// State returns current circuit breaker state
	State() CircuitBreakerState

	// IsOpen returns true if circuit breaker is open
	IsOpen() bool

	// Reset manually resets the circuit breaker
	Reset()

	// GetMetrics returns circuit breaker metrics
	GetMetrics() CircuitBreakerMetrics
}

// CircuitBreakerState represents the circuit breaker state
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerHalfOpen
	CircuitBreakerOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "CLOSED"
	case CircuitBreakerHalfOpen:
		return "HALF_OPEN"
	case CircuitBreakerOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
	HalfOpenTimeout  time.Duration `json:"half_open_timeout"`
	MaxConcurrent    int           `json:"max_concurrent"`
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
		HalfOpenTimeout:  30 * time.Second,
		MaxConcurrent:    100,
	}
}

// CircuitBreakerMetrics contains circuit breaker metrics
type CircuitBreakerMetrics struct {
	State           CircuitBreakerState `json:"state"`
	FailureCount    int64               `json:"failure_count"`
	SuccessCount    int64               `json:"success_count"`
	RequestCount    int64               `json:"request_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	LastSuccessTime time.Time           `json:"last_success_time"`
}

// DefaultCircuitBreaker implements CircuitBreaker interface
type DefaultCircuitBreaker struct {
	config          CircuitBreakerConfig
	state           int32 // atomic access
	failureCount    int64 // atomic access
	successCount    int64 // atomic access
	requestCount    int64 // atomic access
	lastFailureTime int64 // atomic access (unix nano)
	lastSuccessTime int64 // atomic access (unix nano)
	nextAttemptTime int64 // atomic access (unix nano)
	mu              sync.RWMutex
}

// NewDefaultCircuitBreaker creates a new default circuit breaker
func NewDefaultCircuitBreaker(config CircuitBreakerConfig) *DefaultCircuitBreaker {
	return &DefaultCircuitBreaker{
		config: config,
		state:  int32(CircuitBreakerClosed),
	}
}

// Execute runs the operation with circuit breaker protection
func (cb *DefaultCircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if cb.IsOpen() {
		return errors.New("circuit breaker is open")
	}

	// Increment request count
	atomic.AddInt64(&cb.requestCount, 1)

	// Execute operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// State returns current circuit breaker state
func (cb *DefaultCircuitBreaker) State() CircuitBreakerState {
	return CircuitBreakerState(atomic.LoadInt32(&cb.state))
}

// IsOpen returns true if circuit breaker is open
func (cb *DefaultCircuitBreaker) IsOpen() bool {
	state := cb.State()

	if state == CircuitBreakerOpen {
		// Check if we should transition to half-open
		nextAttempt := time.Unix(0, atomic.LoadInt64(&cb.nextAttemptTime))
		if time.Now().After(nextAttempt) {
			atomic.StoreInt32(&cb.state, int32(CircuitBreakerHalfOpen))
			return false
		}
		return true
	}

	return false
}

// Reset manually resets the circuit breaker
func (cb *DefaultCircuitBreaker) Reset() {
	atomic.StoreInt32(&cb.state, int32(CircuitBreakerClosed))
	atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt64(&cb.successCount, 0)
}

// GetMetrics returns circuit breaker metrics
func (cb *DefaultCircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	return CircuitBreakerMetrics{
		State:           cb.State(),
		FailureCount:    atomic.LoadInt64(&cb.failureCount),
		SuccessCount:    atomic.LoadInt64(&cb.successCount),
		RequestCount:    atomic.LoadInt64(&cb.requestCount),
		LastFailureTime: time.Unix(0, atomic.LoadInt64(&cb.lastFailureTime)),
		LastSuccessTime: time.Unix(0, atomic.LoadInt64(&cb.lastSuccessTime)),
	}
}

// recordFailure records a failure and updates circuit breaker state
func (cb *DefaultCircuitBreaker) recordFailure() {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&cb.lastFailureTime, now)

	failureCount := atomic.AddInt64(&cb.failureCount, 1)

	// Check if we should open the circuit
	if cb.State() == CircuitBreakerClosed && failureCount >= int64(cb.config.FailureThreshold) {
		atomic.StoreInt32(&cb.state, int32(CircuitBreakerOpen))
		atomic.StoreInt64(&cb.nextAttemptTime, now+cb.config.Timeout.Nanoseconds())
	}
}

// recordSuccess records a success and updates circuit breaker state
func (cb *DefaultCircuitBreaker) recordSuccess() {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&cb.lastSuccessTime, now)

	successCount := atomic.AddInt64(&cb.successCount, 1)

	// Reset failure count on success
	atomic.StoreInt64(&cb.failureCount, 0)

	// Check if we should close the circuit (from half-open)
	if cb.State() == CircuitBreakerHalfOpen && successCount >= int64(cb.config.SuccessThreshold) {
		atomic.StoreInt32(&cb.state, int32(CircuitBreakerClosed))
		atomic.StoreInt64(&cb.successCount, 0)
	}
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler(retryConfig RetryConfig, circuitBreakerConfig CircuitBreakerConfig) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		circuitBreaker: NewDefaultCircuitBreaker(circuitBreakerConfig),
		retryConfig:    retryConfig,
		errorCounts:    make(map[ErrorClassification]int64),
	}
}

// ClassifyError classifies an error into predefined categories
func (eh *DefaultErrorHandler) ClassifyError(err error) ErrorClassification {
	if err == nil {
		return ErrorClassificationUnknown
	}

	errStr := err.Error()

	// Network errors
	if contains(errStr, []string{"connection", "network", "dns", "no route"}) {
		return ErrorClassificationNetwork
	}

	// Timeout errors
	if contains(errStr, []string{"timeout", "deadline", "context deadline exceeded"}) {
		return ErrorClassificationTimeout
	}

	// Authentication errors
	if contains(errStr, []string{"unauthorized", "authentication", "invalid credentials", "401"}) {
		return ErrorClassificationAuth
	}

	// Rate limit errors
	if contains(errStr, []string{"rate limit", "too many requests", "429"}) {
		return ErrorClassificationRateLimit
	}

	// Resource errors
	if contains(errStr, []string{"resource", "memory", "disk", "quota", "503"}) {
		return ErrorClassificationResource
	}

	// Validation errors
	if contains(errStr, []string{"validation", "invalid", "bad request", "400"}) {
		return ErrorClassificationValidation
	}

	// Internal errors
	if contains(errStr, []string{"internal", "server error", "500"}) {
		return ErrorClassificationInternal
	}

	return ErrorClassificationUnknown
}

// ShouldRetry determines if an operation should be retried based on error
func (eh *DefaultErrorHandler) ShouldRetry(err error, attempt int) bool {
	if attempt >= eh.retryConfig.MaxRetries {
		return false
	}

	classification := eh.ClassifyError(err)

	for _, retryableError := range eh.retryConfig.RetryableErrors {
		if classification == retryableError {
			return true
		}
	}

	return false
}

// HandleError processes an error and returns recovery actions
func (eh *DefaultErrorHandler) HandleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	classification := eh.ClassifyError(err)

	// Increment error count
	eh.mu.Lock()
	eh.errorCounts[classification]++
	eh.mu.Unlock()

	// Wrap error with additional context
	return fmt.Errorf("operation '%s' failed with %s error: %w", operation, classification.String(), err)
}

// GetRetryDelay calculates retry delay with exponential backoff and jitter
func (eh *DefaultErrorHandler) GetRetryDelay(attempt int, baseDelay time.Duration) time.Duration {
	if baseDelay == 0 {
		baseDelay = eh.retryConfig.BaseDelay
	}

	// Calculate exponential backoff
	delay := float64(baseDelay) * math.Pow(eh.retryConfig.BackoffFactor, float64(attempt))

	// Apply maximum delay constraint
	if delay > float64(eh.retryConfig.MaxDelay) {
		delay = float64(eh.retryConfig.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	jitter := delay * eh.retryConfig.JitterFactor * (rand.Float64()*2 - 1)
	delay += jitter

	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(baseDelay)
	}

	return time.Duration(delay)
}

// GetErrorCounts returns error counts by classification
func (eh *DefaultErrorHandler) GetErrorCounts() map[ErrorClassification]int64 {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	counts := make(map[ErrorClassification]int64)
	for k, v := range eh.errorCounts {
		counts[k] = v
	}

	return counts
}

// RetryWithBackoff executes an operation with retry and exponential backoff
func RetryWithBackoff(ctx context.Context, operation func() error, errorHandler ErrorHandler, operationName string) error {
	var lastErr error

	for attempt := 0; attempt < errorHandler.(*DefaultErrorHandler).retryConfig.MaxRetries+1; attempt++ {
		// Execute operation through circuit breaker
		err := errorHandler.(*DefaultErrorHandler).circuitBreaker.Execute(ctx, operation)

		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry
		if !errorHandler.ShouldRetry(err, attempt) {
			break
		}

		// Calculate delay for next attempt
		delay := errorHandler.GetRetryDelay(attempt, 0)

		// Handle the error
		handledErr := errorHandler.HandleError(ctx, err, operationName)
		if handledErr != nil {
			lastErr = handledErr
		}

		// Wait before retry (with context cancellation support)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// Helper function to check if error message contains any of the keywords
func contains(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) >= len(keyword) {
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}
