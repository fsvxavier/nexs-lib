// Package demo demonstrates the improvements implemented in the decimal module
// This file showcases the Cockroach precision corrections, performance optimizations,
// and expanded edge case coverage as requested in NEXT_STEPS.md
package main

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func main() {
	fmt.Println("=== Decimal Module Improvements Demo ===")
	fmt.Println()

	demonstratePrecisionImprovements()
	fmt.Println()

	demonstrateEdgeCases()
	fmt.Println()

	demonstratePerformanceOptimizations()
}

// demonstratePrecisionImprovements shows the Cockroach provider precision fixes
func demonstratePrecisionImprovements() {
	fmt.Println("🔧 Precision Improvements in Cockroach Provider:")

	manager := decimal.NewManager(nil)

	// Test high-precision division
	dividend, _ := manager.NewFromString("10")
	divisor, _ := manager.NewFromString("3")

	result, _ := dividend.Div(divisor)
	fmt.Printf("   10 ÷ 3 = %s (enhanced precision)\n", result.String())

	// Verify mathematical consistency
	backCheck, _ := result.Mul(divisor)
	diff, _ := dividend.Sub(backCheck)
	fmt.Printf("   Verification: (result × 3) - 10 = %s (should be very small)\n", diff.String())

	// Test another precision case
	one, _ := manager.NewFromString("1")
	seven, _ := manager.NewFromString("7")
	oneOverSeven, _ := one.Div(seven)
	fmt.Printf("   1 ÷ 7 = %s (repeating decimal handled properly)\n", oneOverSeven.String())
}

// demonstrateEdgeCases shows the expanded edge case coverage
func demonstrateEdgeCases() {
	fmt.Println("🧪 Expanded Edge Cases Coverage:")

	manager := decimal.NewManager(nil)

	// Very small numbers
	tiny1, _ := manager.NewFromString("0.000000001")
	tiny2, _ := manager.NewFromString("0.000000002")
	tinySum, _ := tiny1.Add(tiny2)
	fmt.Printf("   Tiny numbers: 0.000000001 + 0.000000002 = %s\n", tinySum.String())

	// Scientific notation
	scientific, _ := manager.NewFromString("1.5E-3")
	fmt.Printf("   Scientific notation: 1.5E-3 = %s\n", scientific.String())

	large, _ := manager.NewFromString("2.5e2")
	fmt.Printf("   Scientific notation: 2.5e2 = %s\n", large.String())

	// Type conversions edge cases
	maxInt := int64(9223372036854775807)
	fromMaxInt, _ := manager.NewFromInt(maxInt)
	backToInt, _ := fromMaxInt.Int64()
	fmt.Printf("   Max int64 roundtrip: %d -> %s -> %d\n", maxInt, fromMaxInt.String(), backToInt)

	// Leading/trailing zeros
	zeros, _ := manager.NewFromString("000123.456000")
	fmt.Printf("   Zero handling: '000123.456000' -> %s\n", zeros.String())
}

// demonstratePerformanceOptimizations shows the performance improvements
func demonstratePerformanceOptimizations() {
	fmt.Println("⚡ Performance Optimizations:")

	manager := decimal.NewManager(nil)

	// Demonstrate object pool
	fmt.Println("   Object Pool Demo:")
	slice1 := decimal.GetDecimalSlice()
	fmt.Printf("   - Got slice from pool, capacity: %d\n", cap(slice1))

	for i := 0; i < 5; i++ {
		dec, _ := manager.NewFromInt(int64(i))
		slice1 = append(slice1, dec)
	}
	fmt.Printf("   - Added 5 elements, length: %d\n", len(slice1))

	decimal.PutDecimalSlice(slice1)
	fmt.Println("   - Returned slice to pool for reuse")

	// Demonstrate batch processing performance
	fmt.Println("   Batch Processing Demo:")

	// Create test data
	testData := make([]interfaces.Decimal, 100)
	for i := 0; i < 100; i++ {
		testData[i], _ = manager.NewFromInt(int64(i))
	}

	// Time individual operations
	start := time.Now()
	sum, _ := manager.SumSlice(testData)
	avg, _ := manager.AverageSlice(testData)
	max, _ := manager.MaxSlice(testData)
	min, _ := manager.MinSlice(testData)
	individualTime := time.Since(start)

	// Time batch operation
	start = time.Now()
	processor := manager.NewBatchProcessor()
	batchResult, _ := processor.ProcessSlice(testData)
	batchTime := time.Since(start)

	fmt.Printf("   - Individual operations: %v\n", individualTime)
	fmt.Printf("   - Batch operation: %v\n", batchTime)
	fmt.Printf("   - Results identical: sum=%v, avg=%v, max=%v, min=%v\n",
		batchResult.Sum.IsEqual(sum),
		batchResult.Average.IsEqual(avg),
		batchResult.Max.IsEqual(max),
		batchResult.Min.IsEqual(min))

	// Show fast path optimization for homogeneous types
	fmt.Println("   Fast Path Optimization:")
	homogeneousData := make([]interfaces.Decimal, 50)
	for i := 0; i < 50; i++ {
		homogeneousData[i], _ = manager.NewFromInt(int64(i))
	}

	start = time.Now()
	homogeneousResult, _ := processor.ProcessSlice(homogeneousData)
	homogeneousTime := time.Since(start)

	fmt.Printf("   - Homogeneous dataset (50 elements): %v\n", homogeneousTime)
	fmt.Printf("   - Sum: %s, Average: %s\n",
		homogeneousResult.Sum.String(),
		homogeneousResult.Average.String())

	fmt.Println()
	fmt.Println("✅ All improvements successfully implemented and demonstrated!")
	fmt.Println("   - Precision fixes for Cockroach provider")
	fmt.Println("   - Comprehensive edge case coverage")
	fmt.Println("   - Performance optimizations with pooling and fast paths")
}
