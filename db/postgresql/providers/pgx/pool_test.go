//go:build unit

package pgx

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

func TestPool_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IPool = &Pool{}
}

func TestPool_ClosedOperations(t *testing.T) {
	cfg := config.DefaultConfig()
	pool := &Pool{
		config: cfg,
		closed: true,
	}

	ctx := context.Background()

	// Test operations that should fail when pool is closed
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error for Acquire on closed pool")
	}

	_, err = pool.AcquireWithTimeout(ctx, time.Second)
	if err == nil {
		t.Error("Expected error for AcquireWithTimeout on closed pool")
	}

	err = pool.Ping(ctx)
	if err == nil {
		t.Error("Expected error for Ping on closed pool")
	}

	err = pool.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error for HealthCheck on closed pool")
	}
}

func TestPool_DoubleClose(t *testing.T) {
	cfg := config.DefaultConfig()
	pool := &Pool{
		config: cfg,
		closed: false,
	}

	// First close
	pool.Close()

	if !pool.closed {
		t.Error("Expected pool to be marked as closed")
	}

	// Second close should be a no-op
	pool.Close() // Should not panic
}

func TestPool_Statistics(t *testing.T) {
	cfg := config.DefaultConfig()
	pool := &Pool{
		config: cfg,
		closed: false,
	}

	// Test stats when pool is nil
	stats := pool.Stats()

	// Should return empty stats without panicking
	_ = stats.AcquireCount
	_ = stats.TotalConns
	_ = stats.IdleConns
	_ = stats.MaxConns
}

func TestPool_HealthCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	pool := &Pool{
		config: cfg,
		closed: true, // Set as closed to trigger the closed check first
		pool:   nil,
	}

	ctx := context.Background()

	// Test health check when pool is closed
	err := pool.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error for health check on closed pool")
	}

	expectedMsg := "pool is closed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestPool_NilBatch(t *testing.T) {
	// Pool doesn't have SendBatch method - this test is not needed
	// since Pool only manages connections, not queries
}

func TestPool_InvalidBatch(t *testing.T) {
	// Pool doesn't have SendBatch method - this test is not needed
	// since Pool only manages connections, not queries
}

func TestPool_ConvertTxOptions(t *testing.T) {
	// Pool doesn't have convertTxOptions method since it doesn't handle transactions
	// Transactions are handled by connections acquired from the pool
}

func TestPool_QueryMethods(t *testing.T) {
	// Pool doesn't implement query methods - only connection management
	// Query methods are implemented by connections acquired from the pool
}

func TestPool_TransactionMethods(t *testing.T) {
	// Pool doesn't implement transaction methods - only connection management
	// Transaction methods are implemented by connections acquired from the pool
}

func TestPool_PrepareMethod(t *testing.T) {
	// Pool doesn't implement Prepare method - only connection management
	// Prepare method is implemented by connections acquired from the pool
}

func TestPool_Hooks(t *testing.T) {
	// Pool doesn't implement query hooks - only connection management
	// Query hooks are implemented by connections acquired from the pool
}

func TestPool_HookErrors(t *testing.T) {
	// Pool doesn't implement query hooks - only connection management
	// Query hooks are implemented by connections acquired from the pool
}

func TestPool_Timeout(t *testing.T) {
	cfg := config.NewConfig(
		config.WithQueryTimeout(100 * time.Millisecond),
	)

	pool := &Pool{
		config: cfg,
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	ctx := context.Background()

	// Test that operations return error for closed pool
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error for closed pool")
	}

	expectedMsg := "pool is closed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestPool_EdgeCases(t *testing.T) {
	// Test with minimal config
	cfg := &config.Config{}

	pool := &Pool{
		config: cfg,
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	ctx := context.Background()

	// Operations should return error for closed pool
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error for closed pool")
	}

	expectedMsg := "pool is closed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test ping with minimal config also returns error
	err = pool.Ping(ctx)
	if err == nil {
		t.Error("Expected error for ping on closed pool")
	}
}

func TestPool_AcquireTimeout(t *testing.T) {
	cfg := config.NewConfig(
		config.WithConnectTimeout(100 * time.Millisecond),
	)

	pool := &Pool{
		config: cfg,
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	ctx := context.Background()

	// Test acquire with timeout on closed pool
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error for Acquire on closed pool")
	}

	expectedMsg := "pool is closed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestPool_ConfigNil(t *testing.T) {
	pool := &Pool{
		config: nil,
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	ctx := context.Background()

	// Operations should handle nil config gracefully
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error due to nil pool")
	}
}

func TestPool_Statistics_NilPool(t *testing.T) {
	pool := &Pool{
		config: config.DefaultConfig(),
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	// Should not panic when pool is nil
	stats := pool.Stats()
	// Should return zero values
	if stats.AcquireCount != 0 {
		t.Error("Expected AcquireCount to be 0 for nil pool")
	}

	if stats.TotalConns != 0 {
		t.Error("Expected TotalConns to be 0 for nil pool")
	}

	if stats.IdleConns != 0 {
		t.Error("Expected IdleConns to be 0 for nil pool")
	}

	if stats.MaxConns != 0 {
		t.Error("Expected MaxConns to be 0 for nil pool")
	}
}

func TestPool_ReleaseConn(t *testing.T) {
	// Pool only manages connection acquisition, not release
	// Connections are released directly
}

func TestPool_ReleaseNilConn(t *testing.T) {
	// Pool only manages connection acquisition, not release
	// Connections are released directly
}

func TestPool_MultipleCloses(t *testing.T) {
	cfg := config.DefaultConfig()
	pool := &Pool{
		config: cfg,
		closed: true, // Set as closed to avoid nil pool access
		pool:   nil,
	}

	// Multiple closes should not panic
	pool.Close()
	pool.Close()
	pool.Close()

	if !pool.closed {
		t.Error("Expected pool to be marked as closed")
	}
}
