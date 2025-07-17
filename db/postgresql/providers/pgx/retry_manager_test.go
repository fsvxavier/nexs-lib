//go:build unit

package pgx

import (
	"context"
	"errors"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPool for testing retry manager
type MockPool struct {
	mock.Mock
}

func (m *MockPool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.IConn), args.Error(1)
}

func (m *MockPool) AcquireFunc(ctx context.Context, f func(interfaces.IConn) error) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockPool) Close() {
	m.Called()
}

func (m *MockPool) Reset() {
	m.Called()
}

func (m *MockPool) Stats() interfaces.PoolStats {
	args := m.Called()
	return args.Get(0).(interfaces.PoolStats)
}

func (m *MockPool) Config() interfaces.PoolConfig {
	args := m.Called()
	return args.Get(0).(interfaces.PoolConfig)
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) GetHookManager() interfaces.HookManager {
	args := m.Called()
	return args.Get(0).(interfaces.HookManager)
}

func (m *MockPool) GetBufferPool() interfaces.BufferPool {
	args := m.Called()
	return args.Get(0).(interfaces.BufferPool)
}

func (m *MockPool) GetSafetyMonitor() interfaces.SafetyMonitor {
	args := m.Called()
	return args.Get(0).(interfaces.SafetyMonitor)
}

func TestRetryManager(t *testing.T) {
	defaultConfig := interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		RandomizeWait:   false,
	}

	t.Run("NewRetryManager", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		assert.NotNil(t, rm)

		// Verify interface compliance
		var _ interfaces.RetryManager = rm

		stats := rm.GetStats()
		assert.Equal(t, int64(0), stats.TotalAttempts)
		assert.Equal(t, int64(0), stats.SuccessfulOps)
		assert.Equal(t, int64(0), stats.FailedOps)
		assert.Equal(t, int64(0), stats.TotalRetries)
		assert.Equal(t, float64(0), stats.AverageRetries)
	})

	t.Run("Execute successful operation", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		called := 0
		operation := func() error {
			called++
			return nil
		}

		err := rm.Execute(ctx, operation)
		assert.NoError(t, err)
		assert.Equal(t, 1, called)

		stats := rm.GetStats()
		assert.Equal(t, int64(1), stats.TotalAttempts)
		assert.Equal(t, int64(1), stats.SuccessfulOps)
		assert.Equal(t, int64(0), stats.FailedOps)
		assert.Equal(t, int64(0), stats.TotalRetries)
	})

	t.Run("Execute operation that fails then succeeds", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		called := 0
		operation := func() error {
			called++
			if called == 1 {
				// Simulate a retryable connection error
				return &pgconn.PgError{Code: "08001", Message: "connection refused"}
			}
			return nil
		}

		err := rm.Execute(ctx, operation)
		assert.NoError(t, err)
		assert.Equal(t, 2, called)

		stats := rm.GetStats()
		assert.Equal(t, int64(1), stats.TotalAttempts)
		assert.Equal(t, int64(1), stats.SuccessfulOps)
		assert.Equal(t, int64(0), stats.FailedOps)
		assert.Equal(t, int64(1), stats.TotalRetries)
	})

	t.Run("Execute operation that always fails with retryable error", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		called := 0
		operation := func() error {
			called++
			return &pgconn.PgError{Code: "40001", Message: "serialization failure"}
		}

		err := rm.Execute(ctx, operation)
		assert.Error(t, err)
		assert.Equal(t, 4, called) // Initial + 3 retries
		assert.Contains(t, err.Error(), "operation failed after 4 attempts")

		stats := rm.GetStats()
		assert.Equal(t, int64(1), stats.TotalAttempts)
		assert.Equal(t, int64(0), stats.SuccessfulOps)
		assert.Equal(t, int64(1), stats.FailedOps)
		assert.Equal(t, int64(3), stats.TotalRetries)
	})

	t.Run("Execute operation that fails with non-retryable error", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		called := 0
		operation := func() error {
			called++
			return &pgconn.PgError{Code: "23505", Message: "duplicate key"}
		}

		err := rm.Execute(ctx, operation)
		assert.Error(t, err)
		assert.Equal(t, 1, called) // Only initial attempt
		assert.Contains(t, err.Error(), "non-retryable error")

		stats := rm.GetStats()
		assert.Equal(t, int64(1), stats.TotalAttempts)
		assert.Equal(t, int64(0), stats.SuccessfulOps)
		assert.Equal(t, int64(1), stats.FailedOps)
		assert.Equal(t, int64(0), stats.TotalRetries)
	})

	t.Run("Execute operation with context cancellation", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx, cancel := context.WithCancel(context.Background())

		called := 0
		operation := func() error {
			called++
			if called == 1 {
				cancel() // Cancel context after first call
				return &pgconn.PgError{Code: "08001", Message: "connection refused"}
			}
			return nil
		}

		err := rm.Execute(ctx, operation)
		assert.Error(t, err)
		assert.Equal(t, 1, called)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("Execute operation with context timeout", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		operation := func() error {
			time.Sleep(10 * time.Millisecond) // Longer than context timeout
			return &pgconn.PgError{Code: "08001", Message: "connection refused"}
		}

		err := rm.Execute(ctx, operation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("ExecuteWithConn success", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		mockPool := new(MockPool)
		mockPool.On("AcquireFunc", ctx, mock.AnythingOfType("func(interfaces.IConn) error")).Return(nil)

		operation := func(conn interfaces.IConn) error {
			return nil
		}

		err := rm.ExecuteWithConn(ctx, mockPool, operation)
		assert.NoError(t, err)

		mockPool.AssertExpectations(t)
	})

	t.Run("ExecuteWithConn with pool error", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		mockPool := new(MockPool)
		poolErr := &pgconn.PgError{Code: "08001", Message: "connection refused"}
		mockPool.On("AcquireFunc", ctx, mock.AnythingOfType("func(interfaces.IConn) error")).Return(poolErr).Times(4) // Initial + 3 retries

		operation := func(conn interfaces.IConn) error {
			return nil
		}

		err := rm.ExecuteWithConn(ctx, mockPool, operation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation failed after 4 attempts")

		mockPool.AssertExpectations(t)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)

		newConfig := interfaces.RetryConfig{
			MaxRetries:      5,
			InitialInterval: 20 * time.Millisecond,
			MaxInterval:     200 * time.Millisecond,
			Multiplier:      3.0,
			RandomizeWait:   true,
		}

		err := rm.UpdateConfig(newConfig)
		assert.NoError(t, err)

		// Verify config is updated by testing with new max retries
		ctx := context.Background()
		called := 0
		operation := func() error {
			called++
			return &pgconn.PgError{Code: "40001", Message: "serialization failure"}
		}

		err = rm.Execute(ctx, operation)
		assert.Error(t, err)
		assert.Equal(t, 6, called) // Initial + 5 retries
	})

	t.Run("UpdateConfig with invalid config", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)

		invalidConfig := interfaces.RetryConfig{
			MaxRetries:      -1,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		err := rm.UpdateConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries cannot be negative")
	})

	t.Run("IsRetryableError tests", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig).(*RetryManagerImpl)

		// Nil error
		assert.False(t, rm.isRetryableError(nil))

		// Context errors
		assert.False(t, rm.isRetryableError(context.Canceled))
		assert.False(t, rm.isRetryableError(context.DeadlineExceeded))

		// PGX errors
		assert.False(t, rm.isRetryableError(pgx.ErrNoRows))

		// Retryable PostgreSQL errors
		retryableErrors := []string{"08001", "40001", "40P01", "57014", "57P01"}
		for _, code := range retryableErrors {
			pgErr := &pgconn.PgError{Code: code, Message: "test"}
			assert.True(t, rm.isRetryableError(pgErr), "Error code %s should be retryable", code)
		}

		// Non-retryable PostgreSQL errors
		nonRetryableErrors := []string{"23505", "42601", "42501"}
		for _, code := range nonRetryableErrors {
			pgErr := &pgconn.PgError{Code: code, Message: "test"}
			assert.False(t, rm.isRetryableError(pgErr), "Error code %s should not be retryable", code)
		}

		// Network errors (retryable)
		networkErrors := []error{
			errors.New("connection refused"),
			errors.New("connection reset"),
			errors.New("network is unreachable"),
			errors.New("timeout occurred"),
		}
		for _, err := range networkErrors {
			assert.True(t, rm.isRetryableError(err), "Network error should be retryable: %v", err)
		}
	})

	t.Run("Calculate wait time", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig).(*RetryManagerImpl)

		config := interfaces.RetryConfig{
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1000 * time.Millisecond,
			Multiplier:      2.0,
			RandomizeWait:   false,
		}

		// Test exponential backoff
		wait1 := rm.calculateWaitTime(0, config)
		wait2 := rm.calculateWaitTime(1, config)
		wait3 := rm.calculateWaitTime(2, config)

		assert.Equal(t, 100*time.Millisecond, wait1)
		assert.Equal(t, 200*time.Millisecond, wait2)
		assert.Equal(t, 400*time.Millisecond, wait3)

		// Test max interval cap
		wait10 := rm.calculateWaitTime(10, config)
		assert.Equal(t, 1000*time.Millisecond, wait10)
	})

	t.Run("Calculate wait time with randomization", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig).(*RetryManagerImpl)

		config := interfaces.RetryConfig{
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1000 * time.Millisecond,
			Multiplier:      2.0,
			RandomizeWait:   true,
		}

		// Test that randomization produces different results
		wait1 := rm.calculateWaitTime(1, config)
		wait2 := rm.calculateWaitTime(1, config)

		// They might be the same due to randomness, but at least one should be in a reasonable range
		expectedBase := 200 * time.Millisecond
		minWait := time.Duration(float64(expectedBase) * 0.75) // 25% jitter downward
		maxWait := time.Duration(float64(expectedBase) * 1.25) // 25% jitter upward

		assert.True(t, wait1 >= minWait && wait1 <= maxWait, "Wait time should be within jitter range")
		assert.True(t, wait2 >= minWait && wait2 <= maxWait, "Wait time should be within jitter range")
	})

	t.Run("Stats tracking", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig)
		ctx := context.Background()

		// Successful operation
		successOp := func() error { return nil }
		rm.Execute(ctx, successOp)

		// Failed operation (non-retryable)
		failOp := func() error {
			return &pgconn.PgError{Code: "23505", Message: "duplicate key"}
		}
		rm.Execute(ctx, failOp)

		// Failed operation with retries
		retryOp := func() error {
			return &pgconn.PgError{Code: "40001", Message: "serialization failure"}
		}
		rm.Execute(ctx, retryOp)

		stats := rm.GetStats()
		assert.Equal(t, int64(3), stats.TotalAttempts)
		assert.Equal(t, int64(1), stats.SuccessfulOps)
		assert.Equal(t, int64(2), stats.FailedOps)
		assert.Equal(t, int64(3), stats.TotalRetries) // Only from the retryable failure
		assert.True(t, stats.AverageRetries > 0)
		assert.False(t, stats.LastRetryTime.IsZero())
	})

	t.Run("Config validation", func(t *testing.T) {
		rm := NewRetryManager(defaultConfig).(*RetryManagerImpl)

		validationTests := []struct {
			name        string
			config      interfaces.RetryConfig
			expectError bool
			errorMsg    string
		}{
			{
				name: "valid config",
				config: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 10 * time.Millisecond,
					MaxInterval:     100 * time.Millisecond,
					Multiplier:      2.0,
				},
				expectError: false,
			},
			{
				name: "negative max retries",
				config: interfaces.RetryConfig{
					MaxRetries:      -1,
					InitialInterval: 10 * time.Millisecond,
					MaxInterval:     100 * time.Millisecond,
					Multiplier:      2.0,
				},
				expectError: true,
				errorMsg:    "max retries cannot be negative",
			},
			{
				name: "zero initial interval",
				config: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 0,
					MaxInterval:     100 * time.Millisecond,
					Multiplier:      2.0,
				},
				expectError: true,
				errorMsg:    "initial interval must be positive",
			},
			{
				name: "zero max interval",
				config: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 10 * time.Millisecond,
					MaxInterval:     0,
					Multiplier:      2.0,
				},
				expectError: true,
				errorMsg:    "max interval must be positive",
			},
			{
				name: "max interval less than initial",
				config: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 100 * time.Millisecond,
					MaxInterval:     50 * time.Millisecond,
					Multiplier:      2.0,
				},
				expectError: true,
				errorMsg:    "max interval cannot be less than initial interval",
			},
			{
				name: "multiplier too small",
				config: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 10 * time.Millisecond,
					MaxInterval:     100 * time.Millisecond,
					Multiplier:      1.0,
				},
				expectError: true,
				errorMsg:    "multiplier must be greater than 1.0",
			},
		}

		for _, tt := range validationTests {
			t.Run(tt.name, func(t *testing.T) {
				err := rm.validateConfig(tt.config)
				if tt.expectError {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMsg)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// Benchmark tests
func BenchmarkRetryManager_Execute_Success(b *testing.B) {
	rm := NewRetryManager(interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
		Multiplier:      2.0,
	})
	ctx := context.Background()

	operation := func() error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.Execute(ctx, operation)
	}
}

func BenchmarkRetryManager_Execute_WithRetries(b *testing.B) {
	rm := NewRetryManager(interfaces.RetryConfig{
		MaxRetries:      2,
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     5 * time.Millisecond,
		Multiplier:      2.0,
	})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		called := 0
		operation := func() error {
			called++
			if called <= 2 {
				return &pgconn.PgError{Code: "40001", Message: "serialization failure"}
			}
			return nil
		}
		rm.Execute(ctx, operation)
	}
}

func BenchmarkRetryManager_IsRetryableError(b *testing.B) {
	rm := NewRetryManager(interfaces.RetryConfig{}).(*RetryManagerImpl)
	pgErr := &pgconn.PgError{Code: "40001", Message: "serialization failure"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.isRetryableError(pgErr)
	}
}

func BenchmarkRetryManager_CalculateWaitTime(b *testing.B) {
	rm := NewRetryManager(interfaces.RetryConfig{}).(*RetryManagerImpl)
	config := interfaces.RetryConfig{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		RandomizeWait:   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.calculateWaitTime(i%10, config)
	}
}
