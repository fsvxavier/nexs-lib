//go:build unit
// +build unit

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_config", func(t *testing.T) {
		pool, err := NewPool(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		pool, err := NewPool(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("invalid_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = "invalid://connection/string"

		pool, err := NewPool(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("empty_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = ""

		pool, err := NewPool(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})
}

func TestNewPoolWithManagers(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	t.Run("nil_config", func(t *testing.T) {
		pool, err := NewPoolWithManagers(ctx, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("nil_hook_manager", func(t *testing.T) {
		pool, err := NewPoolWithManagers(ctx, config, nil)
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "hook manager cannot be nil")
	})
}

func TestPoolStatsImpl(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"NewPoolStats", testNewPoolStats},
		{"IncrementAcquireCount", testIncrementAcquireCount},
		{"AddAcquireDuration", testAddAcquireDuration},
		{"IncrementCanceledAcquireCount", testIncrementCanceledAcquireCount},
		{"IncrementEmptyAcquireCount", testIncrementEmptyAcquireCount},
		{"IncrementNewConnsCount", testIncrementNewConnsCount},
		{"GetStats", testGetPoolStats},
		{"ThreadSafety", testPoolStatsThreadSafety},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func testNewPoolStats(t *testing.T) {
	stats := NewPoolStats()

	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.acquireCount)
	assert.Equal(t, time.Duration(0), stats.acquireDuration)
	assert.Equal(t, int32(0), stats.acquiredConns)
	assert.Equal(t, int64(0), stats.canceledAcquireCount)
	assert.Equal(t, int32(0), stats.constructingConns)
	assert.Equal(t, int64(0), stats.emptyAcquireCount)
	assert.Equal(t, int32(0), stats.idleConns)
	assert.Equal(t, int32(0), stats.maxConns)
	assert.Equal(t, int32(0), stats.totalConns)
	assert.Equal(t, int64(0), stats.newConnsCount)
	assert.Equal(t, int64(0), stats.maxLifetimeDestroyCount)
	assert.Equal(t, int64(0), stats.maxIdleDestroyCount)
}

func testIncrementAcquireCount(t *testing.T) {
	stats := NewPoolStats()

	stats.IncrementAcquireCount()
	assert.Equal(t, int64(1), stats.acquireCount)

	stats.IncrementAcquireCount()
	assert.Equal(t, int64(2), stats.acquireCount)
}

func testAddAcquireDuration(t *testing.T) {
	stats := NewPoolStats()

	stats.AddAcquireDuration(time.Millisecond * 100)
	assert.Equal(t, time.Millisecond*100, stats.acquireDuration)

	stats.AddAcquireDuration(time.Millisecond * 200)
	assert.Equal(t, time.Millisecond*300, stats.acquireDuration)
}

func testIncrementCanceledAcquireCount(t *testing.T) {
	stats := NewPoolStats()

	stats.IncrementCanceledAcquireCount()
	assert.Equal(t, int64(1), stats.canceledAcquireCount)

	stats.IncrementCanceledAcquireCount()
	assert.Equal(t, int64(2), stats.canceledAcquireCount)
}

func testIncrementEmptyAcquireCount(t *testing.T) {
	stats := NewPoolStats()

	stats.IncrementEmptyAcquireCount()
	assert.Equal(t, int64(1), stats.emptyAcquireCount)

	stats.IncrementEmptyAcquireCount()
	assert.Equal(t, int64(2), stats.emptyAcquireCount)
}

func testIncrementNewConnsCount(t *testing.T) {
	stats := NewPoolStats()

	stats.IncrementNewConnsCount()
	assert.Equal(t, int64(1), stats.newConnsCount)

	stats.IncrementNewConnsCount()
	assert.Equal(t, int64(2), stats.newConnsCount)
}

func testGetPoolStats(t *testing.T) {
	stats := NewPoolStats()

	// Increment some counters
	stats.IncrementAcquireCount()
	stats.AddAcquireDuration(time.Millisecond * 100)
	stats.IncrementNewConnsCount()
	stats.SetAcquiredConns(5)
	stats.SetIdleConns(3)
	stats.SetTotalConns(8)
	stats.SetMaxConns(10)

	result := stats.Stats()

	assert.Equal(t, int64(1), result.AcquireCount)
	assert.Equal(t, time.Millisecond*100, result.AcquireDuration)
	assert.Equal(t, int32(5), result.AcquiredConns)
	assert.Equal(t, int64(0), result.CanceledAcquireCount)
	assert.Equal(t, int32(0), result.ConstructingConns)
	assert.Equal(t, int64(0), result.EmptyAcquireCount)
	assert.Equal(t, int32(3), result.IdleConns)
	assert.Equal(t, int32(10), result.MaxConns)
	assert.Equal(t, int32(8), result.TotalConns)
	assert.Equal(t, int64(1), result.NewConnsCount)
	assert.Equal(t, int64(0), result.MaxLifetimeDestroyCount)
	assert.Equal(t, int64(0), result.MaxIdleDestroyCount)
}

func testPoolStatsThreadSafety(t *testing.T) {
	stats := NewPoolStats()
	const goroutines = 100
	const iterations = 100

	done := make(chan bool, goroutines)

	// Run concurrent operations
	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				stats.IncrementAcquireCount()
				stats.AddAcquireDuration(time.Millisecond)
				stats.IncrementCanceledAcquireCount()
				stats.IncrementEmptyAcquireCount()
				stats.IncrementNewConnsCount()
				stats.IncrementMaxLifetimeDestroyCount()
				stats.IncrementMaxIdleDestroyCount()
				_ = stats.Stats()
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify final counts
	result := stats.Stats()
	expectedCount := int64(goroutines * iterations)

	assert.Equal(t, expectedCount, result.AcquireCount)
	assert.Equal(t, expectedCount, result.CanceledAcquireCount)
	assert.Equal(t, expectedCount, result.EmptyAcquireCount)
	assert.Equal(t, expectedCount, result.NewConnsCount)
	assert.Equal(t, expectedCount, result.MaxLifetimeDestroyCount)
	assert.Equal(t, expectedCount, result.MaxIdleDestroyCount)
}

func TestPoolConfigConversion(t *testing.T) {
	t.Run("valid_pool_config", func(t *testing.T) {
		config := createTestConfig()
		poolConfig := config.GetPoolConfig()

		assert.Equal(t, int32(10), poolConfig.MaxConns)
		assert.Equal(t, int32(1), poolConfig.MinConns)
		assert.Equal(t, time.Hour, poolConfig.MaxConnLifetime)
		assert.Equal(t, time.Minute*30, poolConfig.MaxConnIdleTime)
		assert.Equal(t, time.Minute*5, poolConfig.HealthCheckPeriod)
		assert.Equal(t, time.Second*30, poolConfig.ConnectTimeout)
		assert.False(t, poolConfig.LazyConnect)
	})

	t.Run("zero_values", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.poolConfig = interfaces.PoolConfig{}

		poolConfig := config.GetPoolConfig()

		assert.Equal(t, int32(0), poolConfig.MaxConns)
		assert.Equal(t, int32(0), poolConfig.MinConns)
		assert.Equal(t, time.Duration(0), poolConfig.MaxConnLifetime)
		assert.Equal(t, time.Duration(0), poolConfig.MaxConnIdleTime)
		assert.Equal(t, time.Duration(0), poolConfig.HealthCheckPeriod)
		assert.Equal(t, time.Duration(0), poolConfig.ConnectTimeout)
		assert.False(t, poolConfig.LazyConnect)
	})
}

func TestPoolTimeout(t *testing.T) {
	// Test with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	config := createTestConfig().(*MockConfig)
	// Use an invalid connection string that will take time to fail
	config.connectionString = "postgres://nonexistent:badpass@192.0.2.1:5432/testdb?sslmode=disable&connect_timeout=10"

	t.Run("NewPool_with_timeout", func(t *testing.T) {
		pool, err := NewPool(ctx, config)
		if err != nil {
			// The timeout might be caught or the connection might fail
			assert.Contains(t, err.Error(), "context")
			assert.Nil(t, pool)
		} else {
			// If pool was created, close it
			assert.NotNil(t, pool)
			pool.Close()
		}
	})
}

func TestPoolEdgeCases(t *testing.T) {
	t.Run("invalid_max_conns", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.poolConfig.MaxConns = -1

		pool, err := NewPool(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("invalid_min_conns", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.poolConfig.MinConns = -1

		pool, err := NewPool(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("min_conns_greater_than_max_conns", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.poolConfig.MinConns = 10
		config.poolConfig.MaxConns = 5

		pool, err := NewPool(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})
}

// Benchmark tests for pool stats
func BenchmarkPoolStats(b *testing.B) {
	stats := NewPoolStats()

	b.Run("IncrementAcquireCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.IncrementAcquireCount()
		}
	})

	b.Run("AddAcquireDuration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.AddAcquireDuration(time.Millisecond)
		}
	})

	b.Run("GetStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stats.Stats()
		}
	})

	b.Run("SetAcquiredConns", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.SetAcquiredConns(int32(i % 100))
		}
	})
}

func BenchmarkConcurrentPoolStats(b *testing.B) {
	stats := NewPoolStats()

	b.Run("ConcurrentIncrements", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				stats.IncrementAcquireCount()
				stats.IncrementNewConnsCount()
				stats.AddAcquireDuration(time.Millisecond)
			}
		})
	})

	b.Run("ConcurrentReads", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = stats.Stats()
			}
		})
	})
}
