package tracer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorClassification_String(t *testing.T) {
	tests := []struct {
		classification ErrorClassification
		expected       string
	}{
		{ErrorClassificationNetwork, "NETWORK"},
		{ErrorClassificationTimeout, "TIMEOUT"},
		{ErrorClassificationAuth, "AUTH"},
		{ErrorClassificationRateLimit, "RATE_LIMIT"},
		{ErrorClassificationInternal, "INTERNAL"},
		{ErrorClassificationValidation, "VALIDATION"},
		{ErrorClassificationResource, "RESOURCE"},
		{ErrorClassificationUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.classification.String())
		})
	}
}

func TestDefaultErrorHandler_ClassifyError(t *testing.T) {
	handler := NewDefaultErrorHandler(DefaultRetryConfig(), DefaultCircuitBreakerConfig())

	tests := []struct {
		name          string
		err           error
		expectedClass ErrorClassification
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedClass: ErrorClassificationUnknown,
		},
		{
			name:          "network error",
			err:           errors.New("connection refused"),
			expectedClass: ErrorClassificationNetwork,
		},
		{
			name:          "dns error",
			err:           errors.New("dns lookup failed"),
			expectedClass: ErrorClassificationNetwork,
		},
		{
			name:          "timeout error",
			err:           errors.New("context deadline exceeded"),
			expectedClass: ErrorClassificationTimeout,
		},
		{
			name:          "timeout error 2",
			err:           errors.New("operation timeout"),
			expectedClass: ErrorClassificationTimeout,
		},
		{
			name:          "auth error",
			err:           errors.New("unauthorized access"),
			expectedClass: ErrorClassificationAuth,
		},
		{
			name:          "rate limit error",
			err:           errors.New("too many requests"),
			expectedClass: ErrorClassificationRateLimit,
		},
		{
			name:          "resource error",
			err:           errors.New("out of memory"),
			expectedClass: ErrorClassificationResource,
		},
		{
			name:          "validation error",
			err:           errors.New("invalid request format"),
			expectedClass: ErrorClassificationValidation,
		},
		{
			name:          "internal error",
			err:           errors.New("internal server error"),
			expectedClass: ErrorClassificationInternal,
		},
		{
			name:          "unknown error",
			err:           errors.New("some random error"),
			expectedClass: ErrorClassificationUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.ClassifyError(tt.err)
			assert.Equal(t, tt.expectedClass, result)
		})
	}
}

func TestDefaultErrorHandler_ShouldRetry(t *testing.T) {
	retryConfig := DefaultRetryConfig()
	retryConfig.MaxRetries = 3
	handler := NewDefaultErrorHandler(retryConfig, DefaultCircuitBreakerConfig())

	tests := []struct {
		name        string
		err         error
		attempt     int
		shouldRetry bool
	}{
		{
			name:        "network error - first attempt",
			err:         errors.New("connection refused"),
			attempt:     1,
			shouldRetry: true,
		},
		{
			name:        "network error - max attempts reached",
			err:         errors.New("connection refused"),
			attempt:     3,
			shouldRetry: false,
		},
		{
			name:        "auth error - not retryable",
			err:         errors.New("unauthorized"),
			attempt:     1,
			shouldRetry: false,
		},
		{
			name:        "timeout error - retryable",
			err:         errors.New("timeout"),
			attempt:     1,
			shouldRetry: true,
		},
		{
			name:        "unknown error - not retryable",
			err:         errors.New("unknown error"),
			attempt:     1,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.ShouldRetry(tt.err, tt.attempt)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}

func TestDefaultErrorHandler_HandleError(t *testing.T) {
	handler := NewDefaultErrorHandler(DefaultRetryConfig(), DefaultCircuitBreakerConfig())
	ctx := context.Background()

	t.Run("nil error", func(t *testing.T) {
		result := handler.HandleError(ctx, nil, "test-operation")
		assert.NoError(t, result)
	})

	t.Run("network error", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		result := handler.HandleError(ctx, originalErr, "test-operation")

		assert.Error(t, result)
		assert.Contains(t, result.Error(), "test-operation")
		assert.Contains(t, result.Error(), "NETWORK")
		assert.Contains(t, result.Error(), "connection refused")
	})

	t.Run("error counts incremented", func(t *testing.T) {
		initialCounts := handler.GetErrorCounts()
		networkErr := errors.New("network failure")

		handler.HandleError(ctx, networkErr, "test-op")

		newCounts := handler.GetErrorCounts()
		assert.Greater(t, newCounts[ErrorClassificationNetwork],
			initialCounts[ErrorClassificationNetwork])
	})
}

func TestDefaultErrorHandler_GetRetryDelay(t *testing.T) {
	retryConfig := RetryConfig{
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}
	handler := NewDefaultErrorHandler(retryConfig, DefaultCircuitBreakerConfig())

	t.Run("exponential backoff", func(t *testing.T) {
		delay1 := handler.GetRetryDelay(0, 100*time.Millisecond)
		delay2 := handler.GetRetryDelay(1, 100*time.Millisecond)
		delay3 := handler.GetRetryDelay(2, 100*time.Millisecond)

		// Due to jitter, we test within ranges
		assert.InDelta(t, 100*time.Millisecond, delay1, float64(20*time.Millisecond))
		assert.InDelta(t, 200*time.Millisecond, delay2, float64(40*time.Millisecond))
		assert.InDelta(t, 400*time.Millisecond, delay3, float64(80*time.Millisecond))
	})

	t.Run("max delay constraint", func(t *testing.T) {
		delay := handler.GetRetryDelay(10, 100*time.Millisecond)
		assert.LessOrEqual(t, delay, retryConfig.MaxDelay+time.Duration(float64(retryConfig.MaxDelay)*retryConfig.JitterFactor))
	})

	t.Run("use default base delay", func(t *testing.T) {
		delay := handler.GetRetryDelay(0, 0)
		assert.InDelta(t, retryConfig.BaseDelay, delay, float64(20*time.Millisecond))
	})
}

func TestDefaultCircuitBreaker_BasicOperations(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.FailureThreshold = 3
	config.SuccessThreshold = 2
	cb := NewDefaultCircuitBreaker(config)

	ctx := context.Background()

	t.Run("initial state is closed", func(t *testing.T) {
		assert.Equal(t, CircuitBreakerClosed, cb.State())
		assert.False(t, cb.IsOpen())
	})

	t.Run("successful operations", func(t *testing.T) {
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.State())
	})

	t.Run("failed operations open circuit", func(t *testing.T) {
		// Create enough failures to open circuit
		for i := 0; i < config.FailureThreshold; i++ {
			err := cb.Execute(ctx, func() error {
				return errors.New("simulated failure")
			})
			assert.Error(t, err)
		}

		assert.Equal(t, CircuitBreakerOpen, cb.State())
		assert.True(t, cb.IsOpen())
	})

	t.Run("operations rejected when open", func(t *testing.T) {
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})
}

func TestDefaultCircuitBreaker_StateTransitions(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.FailureThreshold = 2
	config.SuccessThreshold = 2
	config.Timeout = 100 * time.Millisecond
	cb := NewDefaultCircuitBreaker(config)

	ctx := context.Background()

	// Open the circuit
	for i := 0; i < config.FailureThreshold; i++ {
		cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}
	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Wait for timeout to transition to half-open
	time.Sleep(config.Timeout + 10*time.Millisecond)

	// First call should transition to half-open
	assert.False(t, cb.IsOpen()) // This should update state to half-open

	// Successful operations should close the circuit
	for i := 0; i < config.SuccessThreshold; i++ {
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
	}

	// Circuit should be closed again
	assert.Equal(t, CircuitBreakerClosed, cb.State())
}

func TestDefaultCircuitBreaker_Metrics(t *testing.T) {
	cb := NewDefaultCircuitBreaker(DefaultCircuitBreakerConfig())
	ctx := context.Background()

	// Execute some operations
	cb.Execute(ctx, func() error { return nil })
	cb.Execute(ctx, func() error { return errors.New("failure") })
	cb.Execute(ctx, func() error { return nil })

	metrics := cb.GetMetrics()
	assert.Equal(t, int64(3), metrics.RequestCount)
	assert.Equal(t, int64(2), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount) // Reset on success
	assert.NotZero(t, metrics.LastSuccessTime)
}

func TestDefaultCircuitBreaker_Reset(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.FailureThreshold = 2
	cb := NewDefaultCircuitBreaker(config)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < config.FailureThreshold; i++ {
		cb.Execute(ctx, func() error {
			return errors.New("failure")
		})
	}
	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Reset the circuit breaker
	cb.Reset()
	assert.Equal(t, CircuitBreakerClosed, cb.State())

	// Should be able to execute operations again
	err := cb.Execute(ctx, func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestRetryWithBackoff(t *testing.T) {
	retryConfig := RetryConfig{
		MaxRetries:      2,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		RetryableErrors: []ErrorClassification{ErrorClassificationNetwork},
	}

	handler := NewDefaultErrorHandler(retryConfig, DefaultCircuitBreakerConfig())
	ctx := context.Background()

	t.Run("successful operation", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(ctx, func() error {
			attempts++
			return nil
		}, handler, "test-operation")

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retryable error", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(ctx, func() error {
			attempts++
			if attempts < 3 {
				return errors.New("connection refused")
			}
			return nil
		}, handler, "test-operation")

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(ctx, func() error {
			attempts++
			return errors.New("unauthorized")
		}, handler, "test-operation")

		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(ctx, func() error {
			attempts++
			return errors.New("connection refused")
		}, handler, "test-operation")

		assert.Error(t, err)
		assert.Equal(t, retryConfig.MaxRetries+1, attempts)
	})
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		attempts := 0

		err := RetryWithBackoff(ctx, func() error {
			attempts++
			time.Sleep(30 * time.Millisecond) // Ensure we hit timeout
			return errors.New("connection refused")
		}, handler, "test-operation")

		assert.Error(t, err)
		// Could be either context.DeadlineExceeded or circuit breaker error
		assert.True(t, err == context.DeadlineExceeded ||
			err.Error() == "circuit breaker is open" ||
			err.Error() == "context deadline exceeded")
		assert.LessOrEqual(t, attempts, retryConfig.MaxRetries+1)
	})
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keywords []string
		expected bool
	}{
		{
			name:     "contains keyword",
			text:     "connection refused error",
			keywords: []string{"connection", "timeout"},
			expected: true,
		},
		{
			name:     "does not contain keyword",
			text:     "validation error",
			keywords: []string{"connection", "timeout"},
			expected: false,
		},
		{
			name:     "empty keywords",
			text:     "some text",
			keywords: []string{},
			expected: false,
		},
		{
			name:     "empty text",
			text:     "",
			keywords: []string{"keyword"},
			expected: false,
		},
		{
			name:     "exact match",
			text:     "timeout",
			keywords: []string{"timeout"},
			expected: true,
		},
		{
			name:     "partial match",
			text:     "operation timeout occurred",
			keywords: []string{"timeout"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.text, tt.keywords)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkDefaultErrorHandler_ClassifyError(b *testing.B) {
	handler := NewDefaultErrorHandler(DefaultRetryConfig(), DefaultCircuitBreakerConfig())
	err := errors.New("connection refused")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ClassifyError(err)
	}
}

func BenchmarkDefaultCircuitBreaker_Execute(b *testing.B) {
	cb := NewDefaultCircuitBreaker(DefaultCircuitBreakerConfig())
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(ctx, func() error {
				return nil
			})
		}
	})
}

func BenchmarkRetryWithBackoff(b *testing.B) {
	handler := NewDefaultErrorHandler(DefaultRetryConfig(), DefaultCircuitBreakerConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RetryWithBackoff(ctx, func() error {
			return nil
		}, handler, "benchmark-operation")
	}
}

// Race condition tests
func TestConcurrentCircuitBreakerAccess(t *testing.T) {
	cb := NewDefaultCircuitBreaker(DefaultCircuitBreakerConfig())
	ctx := context.Background()
	numGoroutines := 100
	numOperationsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				cb.Execute(ctx, func() error {
					if (id+j)%5 == 0 {
						return errors.New("simulated failure")
					}
					return nil
				})

				// Test concurrent access to metrics
				_ = cb.GetMetrics()
				_ = cb.State()
				_ = cb.IsOpen()
			}
		}(i)
	}

	wg.Wait()

	// Verify final metrics
	metrics := cb.GetMetrics()
	expectedRequests := int64(numGoroutines * numOperationsPerGoroutine)
	assert.Equal(t, expectedRequests, metrics.RequestCount)
}

func TestConcurrentErrorHandlerAccess(t *testing.T) {
	handler := NewDefaultErrorHandler(DefaultRetryConfig(), DefaultCircuitBreakerConfig())
	ctx := context.Background()
	numGoroutines := 50
	numOperationsPerGoroutine := 20

	errors := []error{
		errors.New("connection refused"),
		errors.New("timeout occurred"),
		errors.New("unauthorized"),
		errors.New("rate limit exceeded"),
		errors.New("internal server error"),
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				err := errors[j%len(errors)]

				// Test concurrent classification
				_ = handler.ClassifyError(err)

				// Test concurrent error handling
				_ = handler.HandleError(ctx, err, "concurrent-test")

				// Test concurrent retry decision
				_ = handler.ShouldRetry(err, j%3)

				// Test concurrent delay calculation
				_ = handler.GetRetryDelay(j%3, 0)

				// Test concurrent error counts access
				_ = handler.GetErrorCounts()
			}
		}(i)
	}

	wg.Wait()

	// Verify error counts
	counts := handler.GetErrorCounts()
	var totalErrors int64
	for _, count := range counts {
		totalErrors += count
	}

	expectedErrors := int64(numGoroutines * numOperationsPerGoroutine)
	assert.Equal(t, expectedErrors, totalErrors)
}
