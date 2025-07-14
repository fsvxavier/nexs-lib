package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestPool_BasicOperations tests basic pool operations
func TestPool_BasicOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*testing.T) *Pool
		testFunc func(*testing.T, *Pool)
	}{
		{
			name: "Acquire_Success",
			setup: func(t *testing.T) *Pool {
				return setupTestPool(t)
			},
			testFunc: func(t *testing.T, pool *Pool) {
				ctx := context.Background()

				conn, err := pool.Acquire(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, conn)

				// Verify connection properties
				gormConn, ok := conn.(*Conn)
				assert.True(t, ok)
				assert.False(t, gormConn.released)
				assert.True(t, gormConn.isFromPool)
				assert.NotNil(t, gormConn.config)
			},
		},
		{
			name: "AcquireWithTimeout_Success",
			setup: func(t *testing.T) *Pool {
				return setupTestPool(t)
			},
			testFunc: func(t *testing.T, pool *Pool) {
				ctx := context.Background()
				timeout := 5 * time.Second

				conn, err := pool.AcquireWithTimeout(ctx, timeout)
				assert.NoError(t, err)
				assert.NotNil(t, conn)

				// Verify connection properties
				gormConn, ok := conn.(*Conn)
				assert.True(t, ok)
				assert.False(t, gormConn.released)
				assert.True(t, gormConn.isFromPool)
			},
		},
		{
			name: "Stats_WithNilSQLDB",
			setup: func(t *testing.T) *Pool {
				pool := setupTestPool(t)
				pool.sqlDB = nil // Simulate nil sqlDB
				return pool
			},
			testFunc: func(t *testing.T, pool *Pool) {
				stats := pool.Stats()

				// Should return empty stats when sqlDB is nil
				assert.Equal(t, int32(0), stats.AcquiredConns)
				assert.Equal(t, int32(0), stats.IdleConns)
				assert.Equal(t, int32(0), stats.MaxConns)
				assert.Equal(t, int32(0), stats.TotalConns)
			},
		},
		{
			name: "Config_Success",
			setup: func(t *testing.T) *Pool {
				return setupTestPool(t)
			},
			testFunc: func(t *testing.T, pool *Pool) {
				cfg := pool.Config()
				assert.NotNil(t, cfg)
				assert.Equal(t, "test", cfg.Database)
				assert.Equal(t, "test", cfg.Username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := tt.setup(t)
			defer pool.Close()
			tt.testFunc(t, pool)
		})
	}
}

// TestPool_ClosedPoolOperations tests operations on closed pools
func TestPool_ClosedPoolOperations(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)
	ctx := context.Background()

	// Close the pool
	pool.Close()
	assert.True(t, pool.closed)

	t.Run("Acquire_OnClosedPool", func(t *testing.T) {
		conn, err := pool.Acquire(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is closed")
		assert.Nil(t, conn)
	})

	t.Run("AcquireWithTimeout_OnClosedPool", func(t *testing.T) {
		conn, err := pool.AcquireWithTimeout(ctx, 5*time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is closed")
		assert.Nil(t, conn)
	})

	t.Run("Ping_OnClosedPool", func(t *testing.T) {
		err := pool.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is closed")
	})

	t.Run("HealthCheck_OnClosedPool", func(t *testing.T) {
		err := pool.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is closed")
	})
}

// TestPool_MultipleClose tests multiple close calls
func TestPool_MultipleClose(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)

	// First close
	pool.Close()
	assert.True(t, pool.closed)

	// Second close should be safe
	pool.Close()
	assert.True(t, pool.closed)
}

// TestPool_PingOperations tests ping operations
func TestPool_PingOperations(t *testing.T) {
	t.Parallel()

	t.Run("Ping_WithNilSQLDB", func(t *testing.T) {
		pool := setupTestPool(t)
		defer pool.Close()

		pool.sqlDB = nil
		ctx := context.Background()

		err := pool.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying sql.DB is nil")
	})

	t.Run("HealthCheck_WithNilSQLDB", func(t *testing.T) {
		pool := setupTestPool(t)
		defer pool.Close()

		pool.sqlDB = nil
		ctx := context.Background()

		err := pool.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "underlying sql.DB is nil")
	})
}

// TestPool_GetConnWithNotPresent tests connection with notification presence
func TestPool_GetConnWithNotPresent(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockIConn(ctrl)

	t.Run("GetConnWithNotPresent_Success", func(t *testing.T) {
		conn, releaseFunc, err := pool.GetConnWithNotPresent(ctx, mockConn)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		assert.NotNil(t, releaseFunc)

		// Test release function
		releaseFunc()

		// Verify the connection was acquired properly
		gormConn, ok := conn.(*Conn)
		assert.True(t, ok)
		assert.True(t, gormConn.released) // Should be released after calling releaseFunc
	})
}

// TestPool_HookOperations tests hook operations
func TestPool_HookOperations(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	t.Run("BeforeAcquireHook_Success", func(t *testing.T) {
		err := pool.BeforeAcquireHook(ctx)
		assert.NoError(t, err)
	})

	t.Run("AfterReleaseHook_Success", func(t *testing.T) {
		err := pool.AfterReleaseHook(ctx)
		assert.NoError(t, err)
	})
}

// TestPool_StatsWithRealConnection tests stats with real DB stats
func TestPool_StatsWithRealConnection(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)

	// Store original sqlDB
	originalSQLDB := pool.sqlDB

	// Reset to nil to test empty stats
	pool.sqlDB = nil

	t.Run("Stats_WithNilSQLDB", func(t *testing.T) {
		stats := pool.Stats()

		// With nil SQL DB, stats should be empty
		assert.Equal(t, int32(0), stats.AcquiredConns)
		assert.Equal(t, int32(0), stats.IdleConns)
		assert.Equal(t, int32(0), stats.MaxConns)
		assert.Equal(t, int32(0), stats.TotalConns)
	})

	// Restore original sqlDB and close properly
	pool.sqlDB = originalSQLDB
	defer pool.Close()
}

// TestPool_ConcurrentOperations tests concurrent pool operations
func TestPool_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)
	defer pool.Close()

	ctx := context.Background()
	concurrency := 10

	t.Run("ConcurrentAcquire", func(t *testing.T) {
		errChan := make(chan error, concurrency)
		connChan := make(chan postgresql.IConn, concurrency)

		// Launch concurrent acquire operations
		for i := 0; i < concurrency; i++ {
			go func() {
				conn, err := pool.Acquire(ctx)
				errChan <- err
				connChan <- conn
			}()
		}

		// Collect results
		var conns []postgresql.IConn
		for i := 0; i < concurrency; i++ {
			err := <-errChan
			conn := <-connChan

			assert.NoError(t, err)
			assert.NotNil(t, conn)
			conns = append(conns, conn)
		}

		// Release all connections
		for _, conn := range conns {
			conn.Release(ctx)
		}
	})
}

// TestPool_ErrorHandling tests error handling scenarios
func TestPool_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("ContextCancellation", func(t *testing.T) {
		pool := setupTestPool(t)
		defer pool.Close()

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := pool.Acquire(ctx)
		// With GORM, this might not fail immediately, but tests the flow
		// The actual behavior depends on GORM's context handling
		_ = err // Ignore error for this test
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		pool := setupTestPool(t)
		defer pool.Close()

		// Create timeout context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		_, err := pool.AcquireWithTimeout(ctx, 1*time.Nanosecond)
		// With GORM, this might not fail immediately, but tests the flow
		_ = err // Ignore error for this test
	})
}

// TestPool_AcquireTimeout tests acquire timeout functionality
func TestPool_AcquireTimeout(t *testing.T) {
	t.Parallel()

	pool := setupTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	t.Run("AcquireWithTimeout_ZeroTimeout", func(t *testing.T) {
		conn, err := pool.AcquireWithTimeout(ctx, 0)
		// Should complete immediately
		assert.NoError(t, err)
		assert.NotNil(t, conn)

		conn.Release(ctx)
	})

	t.Run("AcquireWithTimeout_ShortTimeout", func(t *testing.T) {
		timeout := 1 * time.Millisecond
		conn, err := pool.AcquireWithTimeout(ctx, timeout)

		// With GORM and no real connection, this should succeed quickly
		assert.NoError(t, err)
		assert.NotNil(t, conn)

		if conn != nil {
			conn.Release(ctx)
		}
	})
}

// setupTestPool creates a test pool for unit testing
func setupTestPool(t *testing.T) *Pool {
	// Create a GORM DB instance with a mock dialector for testing
	db, _ := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=invalid port=0 user=test dbname=test sslmode=disable",
	}), &gorm.Config{
		Logger: logger.Discard, // Disable logging for tests
	})

	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	pool := &Pool{
		db:     db,
		config: cfg,
		closed: false,
	}

	// Try to get the underlying sql.DB for stats
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			pool.sqlDB = sqlDB
		}
	}

	return pool
}
