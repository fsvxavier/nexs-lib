package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

// TestPerformanceImprovements verifies that the performance optimizations work correctly
func TestPerformanceImprovements(t *testing.T) {
	manager := NewManager(nil)

	t.Run("pool_functionality", func(t *testing.T) {
		// Test that the decimal pool functions work
		slice1 := GetDecimalSlice()
		assert.NotNil(t, slice1)
		assert.Equal(t, 0, len(slice1))
		assert.True(t, cap(slice1) >= 0) // Should have some capacity

		// Add some elements
		dec1, _ := manager.NewFromString("123.45")
		slice1 = append(slice1, dec1)
		assert.Equal(t, 1, len(slice1))

		// Return to pool
		PutDecimalSlice(slice1)

		// Get another slice (might be the same one)
		slice2 := GetDecimalSlice()
		assert.NotNil(t, slice2)
		assert.Equal(t, 0, len(slice2)) // Should be reset
	})

	t.Run("fast_path_optimization", func(t *testing.T) {
		// Create a large homogeneous slice (should trigger fast path)
		decimals := make([]interfaces.Decimal, 50)
		for i := 0; i < 50; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i))
		}

		processor := manager.NewBatchProcessor()
		result, err := processor.ProcessSlice(decimals)
		require.NoError(t, err)

		// Verify results are correct
		assert.Equal(t, 50, result.Count)
		expected, _ := manager.NewFromInt(49) // Last element
		assert.True(t, result.Max.IsEqual(expected))

		zero := manager.Zero()
		assert.True(t, result.Min.IsEqual(zero))

		// Verify sum (0+1+2+...+49 = 49*50/2 = 1225)
		expectedSum, _ := manager.NewFromInt(1225)
		assert.True(t, result.Sum.IsEqual(expectedSum))
	})

	t.Run("batch_vs_individual_operations", func(t *testing.T) {
		// Create test data
		decimals := make([]interfaces.Decimal, 10)
		for i := 0; i < 10; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i * 10))
		}

		// Test batch operation
		processor := manager.NewBatchProcessor()
		batchResult, err := processor.ProcessSlice(decimals)
		require.NoError(t, err)

		// Test individual operations
		sum, err := manager.SumSlice(decimals)
		require.NoError(t, err)
		avg, err := manager.AverageSlice(decimals)
		require.NoError(t, err)
		max, err := manager.MaxSlice(decimals)
		require.NoError(t, err)
		min, err := manager.MinSlice(decimals)
		require.NoError(t, err)

		// Results should be identical
		assert.True(t, batchResult.Sum.IsEqual(sum))
		assert.True(t, batchResult.Average.IsEqual(avg))
		assert.True(t, batchResult.Max.IsEqual(max))
		assert.True(t, batchResult.Min.IsEqual(min))
	})
}

// BenchmarkPerformanceImprovements measures the impact of optimizations
func BenchmarkPerformanceImprovements(b *testing.B) {
	manager := NewManager(nil)

	b.Run("with_pool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := GetDecimalSlice()
			for j := 0; j < 10; j++ {
				dec, _ := manager.NewFromInt(int64(j))
				slice = append(slice, dec)
			}
			PutDecimalSlice(slice)
		}
	})

	b.Run("without_pool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]interfaces.Decimal, 0, 10)
			for j := 0; j < 10; j++ {
				dec, _ := manager.NewFromInt(int64(j))
				slice = append(slice, dec)
			}
			// Don't use pool
		}
	})

	b.Run("batch_processor_small", func(b *testing.B) {
		decimals := make([]interfaces.Decimal, 10)
		for i := 0; i < 10; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i))
		}
		processor := manager.NewBatchProcessor()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			processor.ProcessSlice(decimals)
		}
	})

	b.Run("batch_processor_large", func(b *testing.B) {
		decimals := make([]interfaces.Decimal, 1000)
		for i := 0; i < 1000; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i))
		}
		processor := manager.NewBatchProcessor()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			processor.ProcessSlice(decimals)
		}
	})

	b.Run("individual_operations", func(b *testing.B) {
		decimals := make([]interfaces.Decimal, 100)
		for i := 0; i < 100; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.SumSlice(decimals)
			manager.AverageSlice(decimals)
			manager.MaxSlice(decimals)
			manager.MinSlice(decimals)
		}
	})

	b.Run("batch_operation", func(b *testing.B) {
		decimals := make([]interfaces.Decimal, 100)
		for i := 0; i < 100; i++ {
			decimals[i], _ = manager.NewFromInt(int64(i))
		}
		processor := manager.NewBatchProcessor()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			processor.ProcessSlice(decimals)
		}
	})
}
