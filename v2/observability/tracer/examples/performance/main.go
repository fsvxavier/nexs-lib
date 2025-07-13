// Package main demonstrates performance optimizations with span pooling and zero-allocation paths
package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

func main() {
	fmt.Println("=== Performance Optimizations Example ===")

	// Example 1: Span Pooling
	spanPoolingExample()

	// Example 2: Zero-allocation Fast Paths
	zeroAllocationExample()

	// Example 3: Memory Usage Comparison
	memoryUsageExample()

	// Example 4: Concurrent Performance
	concurrentPerformanceExample()

	fmt.Println("\n🎉 All performance optimization examples completed!")
}

func spanPoolingExample() {
	fmt.Println("\n--- Span Pooling Example ---")

	// Create performance configuration with span pooling enabled
	config := tracer.PerformanceConfig{
		EnableSpanPooling:    true,
		SpanPoolSize:         1000,
		EnableFastPaths:      true,
		MaxAttributesPerSpan: 64,
		MaxEventsPerSpan:     32,
		EnableZeroAlloc:      true,
	}

	// Create span pool
	pool := tracer.NewSpanPool(config)

	fmt.Printf("✅ Span pool created with size: %d\n", config.SpanPoolSize)

	// Demonstrate span reuse
	spans := make([]*tracer.PooledSpan, 10)

	// Get spans from pool
	start := time.Now()
	for i := 0; i < 10; i++ {
		span := pool.Get()
		span.SetOperationName(fmt.Sprintf("operation_%d", i))
		span.SetAttributeFast("iteration", i)
		span.SetAttributeFast("pool_demo", true)
		spans[i] = span
	}
	getTime := time.Since(start)

	// Return spans to pool
	start = time.Now()
	for _, span := range spans {
		pool.Put(span)
	}
	putTime := time.Since(start)

	metrics := pool.GetMetrics()
	fmt.Printf("📊 Pool metrics: Created=%d, Reused=%d, Destroyed=%d\n",
		metrics.SpansCreated, metrics.SpansReused, metrics.SpansDestroyed)
	fmt.Printf("⚡ Performance: Get=%.2fμs, Put=%.2fμs\n",
		float64(getTime.Nanoseconds())/1000, float64(putTime.Nanoseconds())/1000)
}

func zeroAllocationExample() {
	fmt.Println("\n--- Zero-allocation Fast Paths Example ---")

	config := tracer.PerformanceConfig{
		EnableSpanPooling: true,
		SpanPoolSize:      100,
		EnableFastPaths:   true,
		EnableZeroAlloc:   true,
	}

	pool := tracer.NewSpanPool(config)

	// Get a pooled span
	span := pool.Get()
	defer pool.Put(span)

	fmt.Println("🚀 Demonstrating zero-allocation operations:")

	// Fast attribute setting (zero allocation for basic types)
	start := time.Now()
	span.SetAttributeFast("user_id", "12345")
	span.SetAttributeFast("request_count", 42)
	span.SetAttributeFast("is_premium", true)
	span.SetAttributeFast("latency_ms", 123.45)
	fastTime := time.Since(start)

	// Regular attribute setting (with allocations)
	start = time.Now()
	span.SetAttributeRegular("user_profile", map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	})
	regularTime := time.Since(start)

	fmt.Printf("⚡ Fast path: %.2fns\n", float64(fastTime.Nanoseconds()))
	fmt.Printf("🐌 Regular path: %.2fns\n", float64(regularTime.Nanoseconds()))
	if regularTime.Nanoseconds() > 0 {
		fmt.Printf("📈 Speedup: %.2fx\n", float64(regularTime.Nanoseconds())/float64(fastTime.Nanoseconds()))
	}
}

func memoryUsageExample() {
	fmt.Println("\n--- Memory Usage Comparison ---")

	// Measure memory without pooling
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Create spans without pooling
	normalSpans := make([]*tracer.PooledSpan, 1000)
	for i := 0; i < 1000; i++ {
		// Simulate span creation without pooling
		config := tracer.PerformanceConfig{EnableSpanPooling: false}
		pool := tracer.NewSpanPool(config)
		span := pool.Get()
		span.SetAttributeFast("id", i)
		span.SetAttributeFast("type", "normal")
		normalSpans[i] = span
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Measure memory with pooling
	config := tracer.PerformanceConfig{
		EnableSpanPooling: true,
		SpanPoolSize:      1000,
		EnableFastPaths:   true,
	}

	pool := tracer.NewSpanPool(config)

	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	pooledSpans := make([]*tracer.PooledSpan, 1000)
	for i := 0; i < 1000; i++ {
		span := pool.Get()
		span.SetAttributeFast("id", i)
		span.SetAttributeFast("type", "pooled")
		pooledSpans[i] = span
	}

	runtime.GC()
	var m4 runtime.MemStats
	runtime.ReadMemStats(&m4)

	// Return spans to pool
	for _, span := range pooledSpans {
		pool.Put(span)
	}

	normalMemory := m2.TotalAlloc - m1.TotalAlloc
	pooledMemory := m4.TotalAlloc - m3.TotalAlloc

	fmt.Printf("💾 Memory usage comparison (1000 spans):\n")
	fmt.Printf("   Normal spans: %d bytes\n", normalMemory)
	fmt.Printf("   Pooled spans: %d bytes\n", pooledMemory)
	if normalMemory > pooledMemory {
		savings := float64(normalMemory-pooledMemory) / float64(normalMemory) * 100
		fmt.Printf("   💰 Memory savings: %.1f%%\n", savings)
	}

	poolMetrics := pool.GetMetrics()
	fmt.Printf("📊 Pool metrics: Created=%d, Reused=%d\n",
		poolMetrics.SpansCreated, poolMetrics.SpansReused)
}

func concurrentPerformanceExample() {
	fmt.Println("\n--- Concurrent Performance Example ---")

	config := tracer.PerformanceConfig{
		EnableSpanPooling:    true,
		SpanPoolSize:         1000,
		EnableFastPaths:      true,
		MaxAttributesPerSpan: 16,
		EnableZeroAlloc:      true,
	}

	pool := tracer.NewSpanPool(config)

	const numGoroutines = 10
	const spansPerGoroutine = 100

	start := time.Now()

	// Channel to collect results
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < spansPerGoroutine; i++ {
				span := pool.Get()

				// Perform concurrent operations
				span.SetAttributeFast("goroutine_id", goroutineID)
				span.SetAttributeFast("span_index", i)
				span.SetAttributeFast("timestamp", time.Now().Unix())

				span.AddEventFast("span_created", map[string]interface{}{
					"g": goroutineID,
					"i": i,
				})

				// Simulate some work
				time.Sleep(1 * time.Millisecond)

				pool.Put(span)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	duration := time.Since(start)
	totalSpans := numGoroutines * spansPerGoroutine

	fmt.Printf("⚡ Concurrent performance results:\n")
	fmt.Printf("   Goroutines: %d\n", numGoroutines)
	fmt.Printf("   Spans per goroutine: %d\n", spansPerGoroutine)
	fmt.Printf("   Total spans: %d\n", totalSpans)
	fmt.Printf("   Total time: %v\n", duration)
	fmt.Printf("   Spans/second: %.0f\n", float64(totalSpans)/duration.Seconds())
	fmt.Printf("   Avg time per span: %.2fμs\n", float64(duration.Nanoseconds())/float64(totalSpans)/1000)

	poolMetrics := pool.GetMetrics()
	fmt.Printf("📊 Final pool metrics: Created=%d, Reused=%d, Destroyed=%d\n",
		poolMetrics.SpansCreated, poolMetrics.SpansReused, poolMetrics.SpansDestroyed)
}
