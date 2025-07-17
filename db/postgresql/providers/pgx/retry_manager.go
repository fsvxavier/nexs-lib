package pgx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// RetryManagerImpl implements the RetryManager interface
type RetryManagerImpl struct {
	config interfaces.RetryConfig
	stats  *RetryStatsImpl
	mu     sync.RWMutex
}

// RetryStatsImpl implements retry statistics tracking
type RetryStatsImpl struct {
	totalAttempts int64
	successfulOps int64
	failedOps     int64
	totalRetries  int64
	lastRetryTime time.Time
	mu            sync.RWMutex
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config interfaces.RetryConfig) interfaces.RetryManager {
	return &RetryManagerImpl{
		config: config,
		stats:  &RetryStatsImpl{},
	}
}

// Execute performs retry logic for operations
func (rm *RetryManagerImpl) Execute(ctx context.Context, operation func() error) error {
	return rm.executeWithRetry(ctx, operation, nil, nil)
}

// ExecuteWithConn performs retry logic for operations that need a connection
func (rm *RetryManagerImpl) ExecuteWithConn(ctx context.Context, pool interfaces.IPool, operation func(conn interfaces.IConn) error) error {
	wrappedOp := func() error {
		return pool.AcquireFunc(ctx, operation)
	}
	return rm.executeWithRetry(ctx, wrappedOp, pool, operation)
}

// executeWithRetry implements the core retry logic
func (rm *RetryManagerImpl) executeWithRetry(ctx context.Context, operation func() error, pool interfaces.IPool, connOp func(interfaces.IConn) error) error {
	config := rm.getConfig()
	stats := rm.stats

	atomic.AddInt64(&stats.totalAttempts, 1)

	var lastErr error
	attempt := 0
	maxAttempts := config.MaxRetries + 1 // Include initial attempt

	for attempt < maxAttempts {
		// Check context cancellation
		select {
		case <-ctx.Done():
			atomic.AddInt64(&stats.failedOps, 1)
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Execute operation
		startTime := time.Now()
		err := operation()
		_ = time.Since(startTime) // Track duration for potential future use

		if err == nil {
			atomic.AddInt64(&stats.successfulOps, 1)
			return nil
		}

		lastErr = err
		attempt++

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			atomic.AddInt64(&stats.failedOps, 1)
			return fmt.Errorf("non-retryable error after %d attempts: %w", attempt, err)
		}

		// If this was the last attempt, don't wait
		if attempt >= maxAttempts {
			break
		}

		// Calculate wait time with exponential backoff
		waitTime := rm.calculateWaitTime(attempt-1, config)

		// Record retry attempt
		atomic.AddInt64(&stats.totalRetries, 1)
		stats.mu.Lock()
		stats.lastRetryTime = time.Now()
		stats.mu.Unlock()

		// Wait with context cancellation support
		timer := time.NewTimer(waitTime)
		select {
		case <-ctx.Done():
			timer.Stop()
			atomic.AddInt64(&stats.failedOps, 1)
			return fmt.Errorf("operation cancelled during retry wait: %w", ctx.Err())
		case <-timer.C:
			// Continue to next retry
		}
	}

	atomic.AddInt64(&stats.failedOps, 1)
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}

// calculateWaitTime calculates the wait time for a retry attempt
func (rm *RetryManagerImpl) calculateWaitTime(attempt int, config interfaces.RetryConfig) time.Duration {
	// Exponential backoff: initialInterval * (multiplier ^ attempt)
	waitTime := float64(config.InitialInterval) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max interval
	if waitTime > float64(config.MaxInterval) {
		waitTime = float64(config.MaxInterval)
	}

	duration := time.Duration(waitTime)

	// Add randomization if enabled
	if config.RandomizeWait {
		// Add jitter: ±25% of the wait time
		jitter := float64(duration) * 0.25
		randomOffset := (rand.Float64() * 2 * jitter) - jitter
		duration = time.Duration(float64(duration) + randomOffset)

		// Ensure duration is not negative
		if duration < 0 {
			duration = config.InitialInterval
		}
	}

	return duration
}

// isRetryableError checks if an error is retryable
func (rm *RetryManagerImpl) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors (not retryable)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for PGX specific errors
	if errors.Is(err, pgx.ErrNoRows) {
		return false // Not retryable
	}

	// Check for PostgreSQL errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		// Connection errors (retryable)
		case "08000", "08001", "08003", "08004", "08006", "08P01":
			return true

		// Serialization failure (retryable)
		case "40001":
			return true

		// Deadlock detected (retryable)
		case "40P01":
			return true

		// Statement timeout (retryable)
		case "57014":
			return true

		// Query canceled (retryable)
		case "57P01":
			return true

		// Constraint violations (not retryable)
		case "23000", "23001", "23502", "23503", "23505", "23514":
			return false

		// Syntax errors (not retryable)
		case "42000", "42601", "42602", "42622", "42701", "42702":
			return false

		// Permission errors (not retryable)
		case "42501":
			return false

		// Default: consider unknown errors as potentially retryable
		default:
			return true
		}
	}

	// Network-related errors are usually retryable
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"network is unreachable",
		"timeout",
		"temporary failure",
		"service unavailable",
		"too many connections",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// UpdateConfig updates the retry configuration
func (rm *RetryManagerImpl) UpdateConfig(config interfaces.RetryConfig) error {
	if err := rm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid retry config: %w", err)
	}

	rm.mu.Lock()
	rm.config = config
	rm.mu.Unlock()

	return nil
}

// validateConfig validates retry configuration
func (rm *RetryManagerImpl) validateConfig(config interfaces.RetryConfig) error {
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.InitialInterval <= 0 {
		return fmt.Errorf("initial interval must be positive")
	}

	if config.MaxInterval <= 0 {
		return fmt.Errorf("max interval must be positive")
	}

	if config.MaxInterval < config.InitialInterval {
		return fmt.Errorf("max interval cannot be less than initial interval")
	}

	if config.Multiplier <= 1.0 {
		return fmt.Errorf("multiplier must be greater than 1.0")
	}

	return nil
}

// GetStats returns retry statistics
func (rm *RetryManagerImpl) GetStats() interfaces.RetryStats {
	stats := rm.stats

	stats.mu.RLock()
	lastRetryTime := stats.lastRetryTime
	stats.mu.RUnlock()

	totalAttempts := atomic.LoadInt64(&stats.totalAttempts)
	totalRetries := atomic.LoadInt64(&stats.totalRetries)

	var averageRetries float64
	if totalAttempts > 0 {
		averageRetries = float64(totalRetries) / float64(totalAttempts)
	}

	return interfaces.RetryStats{
		TotalAttempts:  totalAttempts,
		SuccessfulOps:  atomic.LoadInt64(&stats.successfulOps),
		FailedOps:      atomic.LoadInt64(&stats.failedOps),
		TotalRetries:   totalRetries,
		AverageRetries: averageRetries,
		LastRetryTime:  lastRetryTime,
	}
}

// getConfig safely returns the current configuration
func (rm *RetryManagerImpl) getConfig() interfaces.RetryConfig {
	rm.mu.RLock()
	config := rm.config
	rm.mu.RUnlock()
	return config
}

// ResetStats resets all retry statistics
func (rm *RetryManagerImpl) ResetStats() {
	stats := rm.stats
	atomic.StoreInt64(&stats.totalAttempts, 0)
	atomic.StoreInt64(&stats.successfulOps, 0)
	atomic.StoreInt64(&stats.failedOps, 0)
	atomic.StoreInt64(&stats.totalRetries, 0)

	stats.mu.Lock()
	stats.lastRetryTime = time.Time{}
	stats.mu.Unlock()
}
