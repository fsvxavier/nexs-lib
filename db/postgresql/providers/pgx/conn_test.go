//go:build unit
// +build unit

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionStatsImpl(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"NewConnectionStats", testNewConnectionStats},
		{"IncrementQueries", testIncrementQueries},
		{"IncrementExecs", testIncrementExecs},
		{"IncrementTransactions", testIncrementTransactions},
		{"IncrementBatches", testIncrementBatches},
		{"IncrementFailedQueries", testIncrementFailedQueries},
		{"GetStats", testGetStats},
		{"GetAverageQueryTime", testGetAverageQueryTime},
		{"GetAverageExecTime", testGetAverageExecTime},
		{"UpdateLastActivity", testUpdateLastActivity},
		{"ThreadSafety", testConnectionStatsThreadSafety},
		{"IncrementFailedExecs", testIncrementFailedExecs},
		{"IncrementFailedTransactions", testIncrementFailedTransactions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func testNewConnectionStats(t *testing.T) {
	stats := NewConnectionStats()

	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.totalQueries)
	assert.Equal(t, int64(0), stats.totalExecs)
	assert.Equal(t, int64(0), stats.totalTransactions)
	assert.Equal(t, int64(0), stats.totalBatches)
	assert.Equal(t, int64(0), stats.failedQueries)
	assert.Equal(t, int64(0), stats.failedExecs)
	assert.Equal(t, int64(0), stats.failedTransactions)
	assert.Equal(t, time.Duration(0), stats.totalQueryTime)
	assert.Equal(t, time.Duration(0), stats.totalExecTime)
	assert.False(t, stats.createdAt.IsZero())
}

func testIncrementQueries(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementQueries()
	assert.Equal(t, int64(1), stats.totalQueries)

	stats.IncrementQueries()
	assert.Equal(t, int64(2), stats.totalQueries)
}

func testIncrementExecs(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementExecs()
	assert.Equal(t, int64(1), stats.totalExecs)

	stats.IncrementExecs()
	assert.Equal(t, int64(2), stats.totalExecs)
}

func testIncrementTransactions(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementTransactions()
	assert.Equal(t, int64(1), stats.totalTransactions)

	stats.IncrementTransactions()
	assert.Equal(t, int64(2), stats.totalTransactions)
}

func testIncrementBatches(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementBatches()
	assert.Equal(t, int64(1), stats.totalBatches)

	stats.IncrementBatches()
	assert.Equal(t, int64(2), stats.totalBatches)
}

func testIncrementFailedQueries(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementFailedQueries()
	assert.Equal(t, int64(1), stats.failedQueries)

	stats.IncrementFailedQueries()
	assert.Equal(t, int64(2), stats.failedQueries)
}

func testGetStats(t *testing.T) {
	stats := NewConnectionStats()

	// Increment some counters
	stats.IncrementQueries()
	stats.IncrementExecs()
	stats.AddQueryTime(time.Millisecond * 100)
	stats.AddExecTime(time.Millisecond * 50)

	result := stats.Stats()

	assert.Equal(t, int64(1), result.TotalQueries)
	assert.Equal(t, int64(1), result.TotalExecs)
	assert.Equal(t, int64(0), result.TotalTransactions)
	assert.Equal(t, int64(0), result.TotalBatches)
	assert.Equal(t, int64(0), result.FailedQueries)
	assert.Equal(t, int64(0), result.FailedExecs)
	assert.Equal(t, int64(0), result.FailedTransactions)
	assert.Equal(t, time.Millisecond*100, result.AverageQueryTime)
	assert.Equal(t, time.Millisecond*50, result.AverageExecTime)
	assert.False(t, result.CreatedAt.IsZero())
}

func testGetAverageQueryTime(t *testing.T) {
	stats := NewConnectionStats()

	// No queries yet
	assert.Equal(t, time.Duration(0), stats.GetAverageQueryTime())

	// Add some query times
	stats.AddQueryTime(time.Millisecond * 100)
	stats.IncrementQueries()
	assert.Equal(t, time.Millisecond*100, stats.GetAverageQueryTime())

	stats.AddQueryTime(time.Millisecond * 200)
	stats.IncrementQueries()
	assert.Equal(t, time.Millisecond*150, stats.GetAverageQueryTime())
}

func testGetAverageExecTime(t *testing.T) {
	stats := NewConnectionStats()

	// No execs yet
	assert.Equal(t, time.Duration(0), stats.GetAverageExecTime())

	// Add some exec times
	stats.AddExecTime(time.Millisecond * 50)
	stats.IncrementExecs()
	assert.Equal(t, time.Millisecond*50, stats.GetAverageExecTime())

	stats.AddExecTime(time.Millisecond * 150)
	stats.IncrementExecs()
	assert.Equal(t, time.Millisecond*100, stats.GetAverageExecTime())
}

func testUpdateLastActivity(t *testing.T) {
	stats := NewConnectionStats()

	before := time.Now()
	stats.UpdateLastActivity()
	after := time.Now()

	assert.True(t, stats.lastActivity.After(before) || stats.lastActivity.Equal(before))
	assert.True(t, stats.lastActivity.Before(after) || stats.lastActivity.Equal(after))
}

func testConnectionStatsThreadSafety(t *testing.T) {
	stats := NewConnectionStats()
	const goroutines = 100
	const iterations = 100

	done := make(chan bool, goroutines)

	// Run concurrent operations
	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				stats.IncrementQueries()
				stats.IncrementExecs()
				stats.IncrementTransactions()
				stats.IncrementBatches()
				stats.IncrementFailedQueries()
				stats.AddQueryTime(time.Millisecond)
				stats.AddExecTime(time.Millisecond)
				stats.UpdateLastActivity()
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

	assert.Equal(t, expectedCount, result.TotalQueries)
	assert.Equal(t, expectedCount, result.TotalExecs)
	assert.Equal(t, expectedCount, result.TotalTransactions)
	assert.Equal(t, expectedCount, result.TotalBatches)
	assert.Equal(t, expectedCount, result.FailedQueries)
}

func TestNewConn(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_config", func(t *testing.T) {
		conn, err := NewConn(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		conn, err := NewConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("invalid_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = "invalid://connection/string"

		conn, err := NewConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}

func TestNewListenConn(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_config", func(t *testing.T) {
		conn, err := NewListenConn(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		conn, err := NewListenConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("invalid_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = "invalid://connection/string"

		conn, err := NewListenConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}

func TestNewConnWithManagers(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()

	t.Run("nil_config", func(t *testing.T) {
		conn, err := NewConnWithManagers(ctx, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("nil_hook_manager", func(t *testing.T) {
		conn, err := NewConnWithManagers(ctx, config, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "hook manager cannot be nil")
	})

}

// Benchmark tests for connection stats
func BenchmarkConnectionStats(b *testing.B) {
	stats := NewConnectionStats()

	b.Run("IncrementQueries", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.IncrementQueries()
		}
	})

	b.Run("IncrementExecs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.IncrementExecs()
		}
	})

	b.Run("AddQueryTime", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.AddQueryTime(time.Millisecond)
		}
	})

	b.Run("GetStats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stats.Stats()
		}
	})

	b.Run("UpdateLastActivity", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats.UpdateLastActivity()
		}
	})
}

func BenchmarkConcurrentConnectionStats(b *testing.B) {
	stats := NewConnectionStats()

	b.Run("ConcurrentIncrements", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				stats.IncrementQueries()
				stats.IncrementExecs()
				stats.AddQueryTime(time.Millisecond)
			}
		})
	})

	b.Run("ConcurrentReads", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = stats.Stats()
				_ = stats.GetAverageQueryTime()
				_ = stats.GetAverageExecTime()
			}
		})
	})
}

func testIncrementFailedExecs(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementFailedExecs()
	assert.Equal(t, int64(1), stats.failedExecs)

	stats.IncrementFailedExecs()
	assert.Equal(t, int64(2), stats.failedExecs)
}

func testIncrementFailedTransactions(t *testing.T) {
	stats := NewConnectionStats()

	stats.IncrementFailedTransactions()
	assert.Equal(t, int64(1), stats.failedTransactions)

	stats.IncrementFailedTransactions()
	assert.Equal(t, int64(2), stats.failedTransactions)
}
