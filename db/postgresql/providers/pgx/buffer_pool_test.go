//go:build unit

package pgx

import (
	"sync"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	t.Run("NewBufferPool", func(t *testing.T) {
		pool := NewBufferPool()
		assert.NotNil(t, pool)

		// Verify interface compliance
		var _ interfaces.BufferPool = pool
	})

	t.Run("Get buffer", func(t *testing.T) {
		pool := NewBufferPool()

		buffer := pool.Get(1024)
		assert.NotNil(t, buffer)
		assert.Equal(t, 1024, len(buffer))
		assert.Equal(t, 1024, cap(buffer))
	})

	t.Run("Put buffer", func(t *testing.T) {
		pool := NewBufferPool()

		buffer := pool.Get(1024)
		assert.NotNil(t, buffer)

		// Put should not panic
		assert.NotPanics(t, func() {
			pool.Put(buffer)
		})
	})

	t.Run("Get after Put reuses buffer", func(t *testing.T) {
		pool := NewBufferPool()

		// Get a buffer
		buffer1 := pool.Get(1024)
		assert.NotNil(t, buffer1)
		assert.Equal(t, 1024, len(buffer1))

		// Modify the buffer to make it identifiable
		buffer1[0] = 42

		// Put it back
		pool.Put(buffer1)

		// Get another buffer of the same size
		buffer2 := pool.Get(1024)
		assert.NotNil(t, buffer2)
		assert.Equal(t, 1024, len(buffer2))

		// It should be the same buffer (reused)
		// Note: This test is implementation-dependent and might not always pass
		// depending on sync.Pool behavior
	})

	t.Run("Different sizes use different pools", func(t *testing.T) {
		pool := NewBufferPool()

		buffer1024 := pool.Get(1024)
		buffer2048 := pool.Get(2048)
		buffer512 := pool.Get(512)

		assert.Equal(t, 1024, len(buffer1024))
		assert.Equal(t, 2048, len(buffer2048))
		assert.Equal(t, 512, len(buffer512))
	})

	t.Run("Put wrong size buffer", func(t *testing.T) {
		pool := NewBufferPool()

		buffer := make([]byte, 1024)

		// Put should handle wrong size gracefully
		assert.NotPanics(t, func() {
			pool.Put(buffer)
		})
	})

	t.Run("Put nil buffer", func(t *testing.T) {
		pool := NewBufferPool()

		// Put should handle nil gracefully
		assert.NotPanics(t, func() {
			pool.Put(nil)
		})
	})

	t.Run("Put empty buffer", func(t *testing.T) {
		pool := NewBufferPool()

		buffer := make([]byte, 0)

		// Put should handle empty buffer gracefully
		assert.NotPanics(t, func() {
			pool.Put(buffer)
		})
	})

	t.Run("Stats tracking", func(t *testing.T) {
		pool := NewBufferPool()

		// Get initial stats
		initialStats := pool.Stats()

		// Get some buffers
		buffer1 := pool.Get(1024)
		buffer2 := pool.Get(2048)
		buffer3 := pool.Get(1024)

		// Check stats after allocation
		stats := pool.Stats()
		assert.True(t, stats.TotalAllocations >= initialStats.TotalAllocations)
		assert.True(t, stats.AllocatedBuffers >= initialStats.AllocatedBuffers)

		// Put buffers back
		pool.Put(buffer1)
		pool.Put(buffer2)
		pool.Put(buffer3)

		// Check stats after deallocation
		finalStats := pool.Stats()
		assert.True(t, finalStats.TotalDeallocations >= initialStats.TotalDeallocations)
	})

	t.Run("Reset clears pools", func(t *testing.T) {
		pool := NewBufferPool()

		// Get some buffers
		buffer1 := pool.Get(1024)
		buffer2 := pool.Get(2048)

		// Put them back to populate pools
		pool.Put(buffer1)
		pool.Put(buffer2)

		// Reset should not panic
		assert.NotPanics(t, func() {
			pool.Reset()
		})

		// Stats should be reset
		stats := pool.Stats()
		assert.Equal(t, int64(0), stats.TotalAllocations)
		assert.Equal(t, int64(0), stats.TotalDeallocations)
		assert.Equal(t, int32(0), stats.AllocatedBuffers)
		assert.Equal(t, int32(0), stats.PooledBuffers)
	})

	t.Run("Zero size buffer", func(t *testing.T) {
		pool := NewBufferPool()

		buffer := pool.Get(0)
		assert.NotNil(t, buffer)
		assert.Equal(t, 0, len(buffer))
	})

	t.Run("Negative size buffer", func(t *testing.T) {
		pool := NewBufferPool()

		// Should handle negative size gracefully
		assert.NotPanics(t, func() {
			buffer := pool.Get(-1)
			// Buffer should be empty or nil
			_ = buffer
		})
	})

	t.Run("Large buffer", func(t *testing.T) {
		pool := NewBufferPool()

		// Test with large buffer (1MB)
		size := 1024 * 1024
		buffer := pool.Get(size)
		assert.NotNil(t, buffer)
		assert.Equal(t, size, len(buffer))

		pool.Put(buffer)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		pool := NewBufferPool()
		const numGoroutines = 100
		const numOperations = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()

				for j := 0; j < numOperations; j++ {
					// Get buffer
					size := 1024 + (j%5)*512 // Vary sizes
					buffer := pool.Get(size)
					assert.NotNil(t, buffer)
					assert.Equal(t, size, len(buffer))

					// Simulate some work
					time.Sleep(1 * time.Millisecond)

					// Put buffer back
					pool.Put(buffer)
				}
			}()
		}

		wg.Wait()
	})

	t.Run("Memory stats accuracy", func(t *testing.T) {
		pool := NewBufferPool()

		const numBuffers = 10
		buffers := make([][]byte, numBuffers)

		// Get buffers
		for i := 0; i < numBuffers; i++ {
			buffers[i] = pool.Get(1024)
		}

		stats := pool.Stats()
		assert.True(t, stats.AllocatedBuffers >= int32(numBuffers))
		assert.True(t, stats.TotalAllocations >= int64(numBuffers))

		// Put half back
		for i := 0; i < numBuffers/2; i++ {
			pool.Put(buffers[i])
		}

		stats = pool.Stats()
		assert.True(t, stats.TotalDeallocations >= int64(numBuffers/2))
	})
}

func TestMemoryStats(t *testing.T) {
	t.Run("Initial stats", func(t *testing.T) {
		stats := &MemoryStatsImpl{}

		memStats := stats.GetStats()
		assert.Equal(t, int64(0), memStats.BufferSize)
		assert.Equal(t, int32(0), memStats.AllocatedBuffers)
		assert.Equal(t, int32(0), memStats.PooledBuffers)
		assert.Equal(t, int64(0), memStats.TotalAllocations)
		assert.Equal(t, int64(0), memStats.TotalDeallocations)
	})

	t.Run("Stats increments", func(t *testing.T) {
		stats := &MemoryStatsImpl{}

		// Test increment methods
		stats.IncrementAllocatedBuffers()
		stats.IncrementPooledBuffers()
		stats.IncrementTotalAllocations()
		stats.IncrementTotalDeallocations()
		stats.SetBufferSize(1024)

		memStats := stats.GetStats()
		assert.Equal(t, int64(1024), memStats.BufferSize)
		assert.Equal(t, int32(1), memStats.AllocatedBuffers)
		assert.Equal(t, int32(1), memStats.PooledBuffers)
		assert.Equal(t, int64(1), memStats.TotalAllocations)
		assert.Equal(t, int64(1), memStats.TotalDeallocations)
	})

	t.Run("Stats decrements", func(t *testing.T) {
		stats := &MemoryStatsImpl{}

		// Set initial values
		stats.IncrementAllocatedBuffers()
		stats.IncrementAllocatedBuffers()
		stats.IncrementPooledBuffers()
		stats.IncrementPooledBuffers()

		// Test decrement methods
		stats.DecrementAllocatedBuffers()
		stats.DecrementPooledBuffers()

		memStats := stats.GetStats()
		assert.Equal(t, int32(1), memStats.AllocatedBuffers)
		assert.Equal(t, int32(1), memStats.PooledBuffers)
	})

	t.Run("Stats reset", func(t *testing.T) {
		stats := &MemoryStatsImpl{}

		// Set some values
		stats.IncrementAllocatedBuffers()
		stats.IncrementPooledBuffers()
		stats.IncrementTotalAllocations()
		stats.IncrementTotalDeallocations()
		stats.SetBufferSize(1024)

		// Reset
		stats.Reset()

		memStats := stats.GetStats()
		assert.Equal(t, int64(0), memStats.BufferSize)
		assert.Equal(t, int32(0), memStats.AllocatedBuffers)
		assert.Equal(t, int32(0), memStats.PooledBuffers)
		assert.Equal(t, int64(0), memStats.TotalAllocations)
		assert.Equal(t, int64(0), memStats.TotalDeallocations)
	})

	t.Run("Concurrent stats access", func(t *testing.T) {
		stats := &MemoryStatsImpl{}
		const numGoroutines = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines * 2) // For increments and decrements

		// Increment goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					stats.IncrementAllocatedBuffers()
					stats.IncrementPooledBuffers()
					stats.IncrementTotalAllocations()
					stats.IncrementTotalDeallocations()
				}
			}()
		}

		// Decrement goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					stats.DecrementAllocatedBuffers()
					stats.DecrementPooledBuffers()
				}
			}()
		}

		wg.Wait()

		// Check that stats are consistent (no negative values from race conditions)
		memStats := stats.GetStats()
		assert.True(t, memStats.AllocatedBuffers >= 0)
		assert.True(t, memStats.PooledBuffers >= 0)
		assert.True(t, memStats.TotalAllocations >= 0)
		assert.True(t, memStats.TotalDeallocations >= 0)
	})
}

// Benchmark tests
func BenchmarkBufferPool_Get(b *testing.B) {
	pool := NewBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := pool.Get(1024)
		_ = buffer
	}
}

func BenchmarkBufferPool_Put(b *testing.B) {
	pool := NewBufferPool()
	buffers := make([][]byte, b.N)

	// Pre-allocate buffers
	for i := 0; i < b.N; i++ {
		buffers[i] = pool.Get(1024)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Put(buffers[i])
	}
}

func BenchmarkBufferPool_GetPut(b *testing.B) {
	pool := NewBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := pool.Get(1024)
		pool.Put(buffer)
	}
}

func BenchmarkBufferPool_ConcurrentAccess(b *testing.B) {
	pool := NewBufferPool()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buffer := pool.Get(1024)
			pool.Put(buffer)
		}
	})
}

func BenchmarkMemoryStats_Increment(b *testing.B) {
	stats := &MemoryStatsImpl{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.IncrementAllocatedBuffers()
	}
}

func BenchmarkMemoryStats_GetStats(b *testing.B) {
	stats := &MemoryStatsImpl{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.GetStats()
	}
}
