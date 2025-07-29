package valkey

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExponentialBackoffRetryPolicy(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		minBackoff time.Duration
		maxBackoff time.Duration
	}{
		{
			name:       "valid configuration",
			maxRetries: 3,
			minBackoff: 100 * time.Millisecond,
			maxBackoff: 5 * time.Second,
		},
		{
			name:       "zero retries",
			maxRetries: 0,
			minBackoff: 100 * time.Millisecond,
			maxBackoff: 5 * time.Second,
		},
		{
			name:       "large retry count",
			maxRetries: 10,
			minBackoff: 50 * time.Millisecond,
			maxBackoff: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := NewExponentialBackoffRetryPolicy(tt.maxRetries, tt.minBackoff, tt.maxBackoff)
			require.NotNil(t, policy)

			// Verificar se é do tipo correto
			backoffPolicy, ok := policy.(*ExponentialBackoffRetryPolicy)
			require.True(t, ok)
			assert.Equal(t, tt.maxRetries, backoffPolicy.maxRetries)
			assert.Equal(t, tt.minBackoff, backoffPolicy.minBackoff)
			assert.Equal(t, tt.maxBackoff, backoffPolicy.maxBackoff)
			assert.Equal(t, 2.0, backoffPolicy.multiplier)
			assert.True(t, backoffPolicy.jitterEnabled)
		})
	}
}

func TestExponentialBackoffRetryPolicy_ShouldRetry(t *testing.T) {
	policy := NewExponentialBackoffRetryPolicy(3, 100*time.Millisecond, 5*time.Second)

	tests := []struct {
		name        string
		attempt     int
		err         error
		shouldRetry bool
	}{
		{
			name:        "first attempt",
			attempt:     0,
			err:         errors.New("temporary error"),
			shouldRetry: true,
		},
		{
			name:        "second attempt",
			attempt:     1,
			err:         errors.New("temporary error"),
			shouldRetry: true,
		},
		{
			name:        "third attempt",
			attempt:     2,
			err:         errors.New("temporary error"),
			shouldRetry: true,
		},
		{
			name:        "max retries exceeded",
			attempt:     3,
			err:         errors.New("temporary error"),
			shouldRetry: false,
		},
		{
			name:        "authentication failed error",
			attempt:     1,
			err:         errors.New("authentication failed"),
			shouldRetry: false,
		},
		{
			name:        "context canceled error",
			attempt:     1,
			err:         context.Canceled,
			shouldRetry: false,
		},
		{
			name:        "context deadline exceeded",
			attempt:     1,
			err:         context.DeadlineExceeded,
			shouldRetry: false,
		},
		{
			name:        "nil error",
			attempt:     1,
			err:         nil,
			shouldRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.ShouldRetry(tt.attempt, tt.err)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}

func TestExponentialBackoffRetryPolicy_NextDelay(t *testing.T) {
	minBackoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second
	policy := NewExponentialBackoffRetryPolicy(5, minBackoff, maxBackoff)

	tests := []struct {
		name           string
		attempt        int
		expectedMinDur time.Duration
		expectedMaxDur time.Duration
	}{
		{
			name:           "zero attempt",
			attempt:        0,
			expectedMinDur: minBackoff,
			expectedMaxDur: minBackoff + 50*time.Millisecond, // com jitter
		},
		{
			name:           "first retry",
			attempt:        1,
			expectedMinDur: minBackoff,
			expectedMaxDur: 300 * time.Millisecond, // 100ms * 2 + jitter
		},
		{
			name:           "second retry",
			attempt:        2,
			expectedMinDur: 200 * time.Millisecond,
			expectedMaxDur: 600 * time.Millisecond, // 200ms * 2 + jitter
		},
		{
			name:           "third retry",
			attempt:        3,
			expectedMinDur: 400 * time.Millisecond,
			expectedMaxDur: 1200 * time.Millisecond, // 400ms * 2 + jitter
		},
		{
			name:           "max backoff reached",
			attempt:        10,
			expectedMinDur: maxBackoff - 1*time.Second, // deve estar próximo do máximo
			expectedMaxDur: maxBackoff + 1*time.Second, // tolerância maior para jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := policy.NextDelay(tt.attempt)
			assert.True(t, delay >= tt.expectedMinDur, "delay %v should be >= %v", delay, tt.expectedMinDur)
			assert.True(t, delay <= tt.expectedMaxDur, "delay %v should be <= %v", delay, tt.expectedMaxDur)
		})
	}
}

func TestExponentialBackoffRetryPolicy_NextDelay_ConsistentGrowth(t *testing.T) {
	policy := NewExponentialBackoffRetryPolicy(5, 100*time.Millisecond, 10*time.Second)

	var lastDelay time.Duration
	for attempt := 1; attempt < 5; attempt++ {
		delay := policy.NextDelay(attempt)
		if attempt > 1 {
			// O delay deve crescer (considerando que pode ter jitter)
			assert.True(t, delay >= lastDelay/2, "delay should generally increase, got %v after %v", delay, lastDelay)
		}
		lastDelay = delay
	}
}

func TestIsNonRetryableError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		isNonRetryable bool
	}{
		{
			name:           "nil error",
			err:            nil,
			isNonRetryable: false,
		},
		{
			name:           "context canceled",
			err:            context.Canceled,
			isNonRetryable: true,
		},
		{
			name:           "context deadline exceeded",
			err:            context.DeadlineExceeded,
			isNonRetryable: true,
		},
		{
			name:           "authentication failed error",
			err:            errors.New("authentication failed"),
			isNonRetryable: true,
		},
		{
			name:           "permission denied error",
			err:            errors.New("permission denied"),
			isNonRetryable: true,
		},
		{
			name:           "temporary network error",
			err:            errors.New("connection refused"),
			isNonRetryable: false,
		},
		{
			name:           "generic error",
			err:            errors.New("something went wrong"),
			isNonRetryable: false,
		},
		{
			name:           "wrapped context canceled",
			err:            fmt.Errorf("wrapped: %w", context.Canceled),
			isNonRetryable: true,
		},
		{
			name:           "wrapped authentication failed",
			err:            fmt.Errorf("wrapped: %w", errors.New("authentication failed")),
			isNonRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNonRetryableError(tt.err)
			assert.Equal(t, tt.isNonRetryable, result)
		})
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name                  string
		failureThreshold      int
		timeout               time.Duration
		maxConcurrentRequests int
	}{
		{
			name:                  "default configuration",
			failureThreshold:      5,
			timeout:               30 * time.Second,
			maxConcurrentRequests: 100,
		},
		{
			name:                  "high threshold",
			failureThreshold:      20,
			timeout:               60 * time.Second,
			maxConcurrentRequests: 200,
		},
		{
			name:                  "low threshold",
			failureThreshold:      1,
			timeout:               5 * time.Second,
			maxConcurrentRequests: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.failureThreshold, tt.timeout, tt.maxConcurrentRequests)
			require.NotNil(t, cb)

			circuitBreaker, ok := cb.(*CircuitBreaker)
			require.True(t, ok)
			assert.Equal(t, tt.failureThreshold, circuitBreaker.threshold)
			assert.Equal(t, tt.timeout, circuitBreaker.timeout)
			assert.Equal(t, tt.maxConcurrentRequests, circuitBreaker.maxRequests)
			assert.Equal(t, StateClosed, circuitBreaker.state)
			assert.Equal(t, 0, circuitBreaker.failureCount)
		})
	}
}

func TestCircuitBreaker_Execute_SuccessfulOperation(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second, 10)
	ctx := context.Background()

	result, err := cb.Execute(ctx, func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	circuitBreaker := cb.(*CircuitBreaker)
	assert.Equal(t, StateClosed, circuitBreaker.state)
	assert.Equal(t, 0, circuitBreaker.failureCount)
}

func TestCircuitBreaker_Execute_FailingOperation(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Second, 10)
	ctx := context.Background()
	testErr := errors.New("operation failed")

	// Primeira falha
	result, err := cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})

	assert.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Nil(t, result)

	circuitBreaker := cb.(*CircuitBreaker)
	assert.Equal(t, StateClosed, circuitBreaker.state)
	assert.Equal(t, 1, circuitBreaker.failureCount)

	// Segunda falha - deve abrir o circuit
	result, err = cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})

	assert.Error(t, err)
	assert.Equal(t, testErr, err)
	assert.Nil(t, result)

	assert.Equal(t, StateOpen, circuitBreaker.state)
	assert.Equal(t, 2, circuitBreaker.failureCount)
}

func TestCircuitBreaker_Execute_OpenState(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond, 10)
	ctx := context.Background()
	testErr := errors.New("operation failed")

	// Forçar abertura do circuit
	_, _ = cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})

	circuitBreaker := cb.(*CircuitBreaker)
	assert.Equal(t, StateOpen, circuitBreaker.state)

	// Tentar executar com circuit aberto
	result, err := cb.Execute(ctx, func() (interface{}, error) {
		return "should not execute", nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
	assert.Nil(t, result)
}

func TestCircuitBreaker_Execute_HalfOpenState(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond, 1)
	ctx := context.Background()
	testErr := errors.New("operation failed")

	// Forçar abertura do circuit
	_, _ = cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})

	circuitBreaker := cb.(*CircuitBreaker)
	assert.Equal(t, StateOpen, circuitBreaker.state)

	// Esperar timeout para transição para half-open
	time.Sleep(150 * time.Millisecond)

	// Primeira execução deve colocar em half-open
	result, err := cb.Execute(ctx, func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, circuitBreaker.state)
}

func TestCircuitBreaker_Execute_MaxConcurrentRequests(t *testing.T) {
	cb := NewCircuitBreaker(5, 1*time.Second, 1)
	ctx := context.Background()

	// Testar múltiplas requisições concorrentes
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := cb.Execute(ctx, func() (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return "success", nil
			})
			results <- err
		}()
	}

	// Coletar resultados
	var errorCount int
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			errorCount++
		}
	}

	// Com circuit breaker fechado, não deveria haver erros
	// Este teste verifica se o circuit breaker lida bem com concorrência
	assert.True(t, errorCount <= numRequests/2, "most requests should succeed when circuit is closed")
}

func TestCircuitBreaker_Execute_ContextCancellation(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := cb.Execute(ctx, func() (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return "success", nil
		}
	})

	// Deve falhar devido ao timeout do contexto
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
	assert.Nil(t, result)
}

func TestCircuitBreaker_States(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond, 10)
	ctx := context.Background()
	testErr := errors.New("operation failed")

	circuitBreaker := cb.(*CircuitBreaker)

	// Estado inicial: Closed
	assert.Equal(t, StateClosed, circuitBreaker.state)

	// Primeira falha: ainda Closed
	cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})
	assert.Equal(t, StateClosed, circuitBreaker.state)
	assert.Equal(t, 1, circuitBreaker.failureCount)

	// Segunda falha: transição para Open
	cb.Execute(ctx, func() (interface{}, error) {
		return nil, testErr
	})
	assert.Equal(t, StateOpen, circuitBreaker.state)
	assert.Equal(t, 2, circuitBreaker.failureCount)

	// Esperar timeout
	time.Sleep(60 * time.Millisecond)

	// Próxima execução deve transicionar para HalfOpen
	cb.Execute(ctx, func() (interface{}, error) {
		return "success", nil
	})

	// Deve estar Closed novamente após sucesso
	assert.Equal(t, StateClosed, circuitBreaker.state)
	assert.Equal(t, 0, circuitBreaker.failureCount)
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(1, 1*time.Second, 10)
	ctx := context.Background()

	// Forçar falha
	cb.Execute(ctx, func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	circuitBreaker := cb.(*CircuitBreaker)
	assert.Equal(t, StateOpen, circuitBreaker.state)
	assert.Equal(t, 1, circuitBreaker.failureCount)

	// Reset
	circuitBreaker.Reset()

	assert.Equal(t, StateClosed, circuitBreaker.state)
	assert.Equal(t, 0, circuitBreaker.failureCount)
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond, 50)
	ctx := context.Background()

	const numGoroutines = 100
	results := make(chan error, numGoroutines)

	// Executar operações concorrentes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_, err := cb.Execute(ctx, func() (interface{}, error) {
				if id%10 == 0 {
					return nil, errors.New("fail")
				}
				return "success", nil
			})
			results <- err
		}(i)
	}

	// Coletar resultados
	var successCount, errorCount int
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Verificar que houve sucessos e falhas
	assert.True(t, successCount > 0, "expected some successful operations")
	assert.True(t, errorCount > 0, "expected some failed operations")

	// O circuit breaker deve ainda estar funcionando corretamente
	circuitBreaker := cb.(*CircuitBreaker)
	assert.True(t, circuitBreaker.state == StateClosed || circuitBreaker.state == StateOpen)
}

func BenchmarkExponentialBackoffRetryPolicy_NextDelay(b *testing.B) {
	policy := NewExponentialBackoffRetryPolicy(5, 100*time.Millisecond, 5*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy.NextDelay(i % 5)
	}
}

func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	cb := NewCircuitBreaker(100, 1*time.Second, 1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(ctx, func() (interface{}, error) {
			return "success", nil
		})
	}
}

func BenchmarkCircuitBreaker_ExecuteConcurrent(b *testing.B) {
	cb := NewCircuitBreaker(100, 1*time.Second, 1000)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(ctx, func() (interface{}, error) {
				return "success", nil
			})
		}
	})
}
