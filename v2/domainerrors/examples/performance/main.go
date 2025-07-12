package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("⚡ Domain Errors v2 - Performance Examples")
	fmt.Println("=========================================")

	errorCreationBenchmarks()
	memoryUsageBenchmarks()
	concurrencyBenchmarks()
	serializationBenchmarks()
	errorChainBenchmarks()
	builderPatternBenchmarks()
	factoryBenchmarks()
	optimizationTechniques()
}

// errorCreationBenchmarks demonstrates performance of basic error creation
func errorCreationBenchmarks() {
	fmt.Println("\n🏗️ Error Creation Benchmarks:")

	benchmarks := []struct {
		name string
		fn   func() interfaces.DomainErrorInterface
	}{
		{
			"Simple Error",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("SIMPLE_ERROR").
					WithMessage("Simple error message").
					Build()
			},
		},
		{
			"Error with Type",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("TYPED_ERROR").
					WithMessage("Error with type information").
					WithType(string(types.ErrorTypeValidation)).
					Build()
			},
		},
		{
			"Error with Severity",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("SEVERITY_ERROR").
					WithMessage("Error with severity").
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					Build()
			},
		},
		{
			"Error with Details",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("DETAILED_ERROR").
					WithMessage("Error with details").
					WithType(string(types.ErrorTypeBusinessRule)).
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("user_id", "user123").
					WithDetail("operation", "create_order").
					WithDetail("amount", 99.99).
					Build()
			},
		},
		{
			"Complex Error",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("COMPLEX_ERROR").
					WithMessage("Complex error with all fields").
					WithType(string(types.ErrorTypeExternalService)).
					WithSeverity(interfaces.Severity(types.SeverityCritical)).
					WithDetail("service", "payment-gateway").
					WithDetail("endpoint", "/api/v1/payments").
					WithDetail("timeout", "30s").
					WithDetail("retry_count", 3).
					WithDetail("metadata", map[string]interface{}{
						"correlation_id": "corr123",
						"session_id":     "sess456",
						"user_agent":     "MyApp/1.0",
					}).
					WithTag("payment").
					WithTag("timeout").
					WithTag("retryable").
					Build()
			},
		},
	}

	iterations := 10000

	for _, benchmark := range benchmarks {
		fmt.Printf("  %s:\n", benchmark.name)

		// Warmup
		for i := 0; i < 100; i++ {
			_ = benchmark.fn()
		}

		// Measure performance
		start := time.Now()
		for i := 0; i < iterations; i++ {
			_ = benchmark.fn()
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(iterations)
		opsPerSecond := float64(iterations) / duration.Seconds()

		fmt.Printf("    Average: %v\n", avgDuration)
		fmt.Printf("    Ops/sec: %.0f\n", opsPerSecond)
		fmt.Printf("    Total: %v (%d iterations)\n", duration, iterations)

		// Performance classification
		if avgDuration <= 100*time.Nanosecond {
			fmt.Printf("    Performance: ⚡ EXCELLENT (≤100ns)\n")
		} else if avgDuration <= 1*time.Microsecond {
			fmt.Printf("    Performance: ✅ GOOD (≤1μs)\n")
		} else if avgDuration <= 10*time.Microsecond {
			fmt.Printf("    Performance: ⚠️ ACCEPTABLE (≤10μs)\n")
		} else {
			fmt.Printf("    Performance: ❌ SLOW (>10μs)\n")
		}
	}
}

// memoryUsageBenchmarks demonstrates memory allocation patterns
func memoryUsageBenchmarks() {
	fmt.Println("\n💾 Memory Usage Benchmarks:")

	scenarios := []struct {
		name string
		fn   func()
	}{
		{
			"Simple Error Creation",
			func() {
				_ = factory.GetDefaultFactory().Builder().
					WithCode("MEMORY_TEST").
					WithMessage("Memory test error").
					Build()
			},
		},
		{
			"Error with Multiple Details",
			func() {
				builder := factory.GetDefaultFactory().Builder().
					WithCode("DETAILED_MEMORY_TEST").
					WithMessage("Detailed memory test error")

				for i := 0; i < 10; i++ {
					builder.WithDetail(fmt.Sprintf("detail_%d", i), fmt.Sprintf("value_%d", i))
				}

				_ = builder.Build()
			},
		},
		{
			"Error Chain Creation",
			func() {
				rootErr := factory.GetDefaultFactory().Builder().
					WithCode("ROOT_ERROR").
					WithMessage("Root error").
					Build()

				_ = factory.GetDefaultFactory().Builder().
					WithCode("WRAPPED_ERROR").
					WithMessage("Wrapped error").
					WithDetail("wrapped_error", rootErr).
					Build()
			},
		},
	}

	iterations := 1000

	for _, scenario := range scenarios {
		fmt.Printf("  %s:\n", scenario.name)

		// Force garbage collection to get clean baseline
		runtime.GC()
		runtime.GC()

		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Run scenario
		for i := 0; i < iterations; i++ {
			scenario.fn()
		}

		runtime.ReadMemStats(&m2)

		// Calculate memory usage
		allocations := m2.Mallocs - m1.Mallocs
		deallocations := m2.Frees - m1.Frees
		netAllocations := allocations - deallocations
		bytesAllocated := m2.TotalAlloc - m1.TotalAlloc

		fmt.Printf("    Allocations: %d\n", allocations)
		fmt.Printf("    Deallocations: %d\n", deallocations)
		fmt.Printf("    Net Allocations: %d\n", netAllocations)
		fmt.Printf("    Bytes Allocated: %d\n", bytesAllocated)
		fmt.Printf("    Avg Bytes/Operation: %.2f\n", float64(bytesAllocated)/float64(iterations))
		fmt.Printf("    Allocs/Operation: %.2f\n", float64(allocations)/float64(iterations))

		// Memory efficiency classification
		avgBytesPerOp := float64(bytesAllocated) / float64(iterations)
		if avgBytesPerOp <= 500 {
			fmt.Printf("    Memory Efficiency: ⚡ EXCELLENT (≤500 bytes/op)\n")
		} else if avgBytesPerOp <= 1000 {
			fmt.Printf("    Memory Efficiency: ✅ GOOD (≤1KB/op)\n")
		} else if avgBytesPerOp <= 5000 {
			fmt.Printf("    Memory Efficiency: ⚠️ ACCEPTABLE (≤5KB/op)\n")
		} else {
			fmt.Printf("    Memory Efficiency: ❌ HIGH USAGE (>5KB/op)\n")
		}
	}

	// Show current memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\n  Current Memory Stats:\n")
	fmt.Printf("    Allocated: %d KB\n", m.Alloc/1024)
	fmt.Printf("    Total Allocations: %d\n", m.Mallocs)
	fmt.Printf("    GC Cycles: %d\n", m.NumGC)
	fmt.Printf("    Heap Objects: %d\n", m.HeapObjects)
}

// concurrencyBenchmarks demonstrates thread-safety and concurrent performance
func concurrencyBenchmarks() {
	fmt.Println("\n🔄 Concurrency Benchmarks:")

	scenarios := []struct {
		name        string
		goroutines  int
		iterations  int
		description string
	}{
		{
			"Low Concurrency",
			10,
			1000,
			"10 goroutines, 1000 operations each",
		},
		{
			"Medium Concurrency",
			100,
			500,
			"100 goroutines, 500 operations each",
		},
		{
			"High Concurrency",
			1000,
			100,
			"1000 goroutines, 100 operations each",
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("  %s (%s):\n", scenario.name, scenario.description)

		var wg sync.WaitGroup
		start := time.Now()

		// Launch goroutines
		for i := 0; i < scenario.goroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < scenario.iterations; j++ {
					_ = factory.GetDefaultFactory().Builder().
						WithCode("CONCURRENT_ERROR").
						WithMessage("Concurrent error test").
						WithType(string(types.ErrorTypeValidation)).
						WithSeverity(interfaces.Severity(types.SeverityMedium)).
						WithDetail("goroutine_id", goroutineID).
						WithDetail("iteration", j).
						WithTag("concurrency").
						Build()
				}
			}(i)
		}

		// Wait for completion
		wg.Wait()
		duration := time.Since(start)

		totalOperations := scenario.goroutines * scenario.iterations
		opsPerSecond := float64(totalOperations) / duration.Seconds()
		avgDuration := duration / time.Duration(totalOperations)

		fmt.Printf("    Total Operations: %d\n", totalOperations)
		fmt.Printf("    Total Duration: %v\n", duration)
		fmt.Printf("    Ops/sec: %.0f\n", opsPerSecond)
		fmt.Printf("    Avg Duration: %v\n", avgDuration)

		// Concurrency performance classification
		if opsPerSecond >= 100000 {
			fmt.Printf("    Concurrency Performance: ⚡ EXCELLENT (≥100K ops/sec)\n")
		} else if opsPerSecond >= 50000 {
			fmt.Printf("    Concurrency Performance: ✅ GOOD (≥50K ops/sec)\n")
		} else if opsPerSecond >= 10000 {
			fmt.Printf("    Concurrency Performance: ⚠️ ACCEPTABLE (≥10K ops/sec)\n")
		} else {
			fmt.Printf("    Concurrency Performance: ❌ SLOW (<10K ops/sec)\n")
		}
	}

	// Test thread safety
	fmt.Printf("\n  Thread Safety Test:\n")
	testThreadSafety()
}

// serializationBenchmarks demonstrates serialization performance
func serializationBenchmarks() {
	fmt.Println("\n📦 Serialization Benchmarks:")

	// Create test errors of different complexities
	testErrors := []struct {
		name string
		err  interfaces.DomainErrorInterface
	}{
		{
			"Simple Error",
			factory.GetDefaultFactory().Builder().
				WithCode("SIMPLE").
				WithMessage("Simple error").
				Build(),
		},
		{
			"Medium Complexity Error",
			factory.GetDefaultFactory().Builder().
				WithCode("MEDIUM_COMPLEX").
				WithMessage("Medium complexity error").
				WithType(string(types.ErrorTypeBusinessRule)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("user_id", "user123").
				WithDetail("operation", "update_profile").
				WithTag("validation").
				Build(),
		},
		{
			"High Complexity Error",
			func() interfaces.DomainErrorInterface {
				builder := factory.GetDefaultFactory().Builder().
					WithCode("HIGH_COMPLEX").
					WithMessage("High complexity error with extensive metadata").
					WithType(string(types.ErrorTypeExternalService)).
					WithSeverity(interfaces.Severity(types.SeverityCritical))

				// Add many details
				for i := 0; i < 20; i++ {
					builder.WithDetail(fmt.Sprintf("detail_%d", i), fmt.Sprintf("value_%d", i))
				}

				// Add many tags
				for i := 0; i < 10; i++ {
					builder.WithTag(fmt.Sprintf("tag_%d", i))
				}

				return builder.Build()
			}(),
		},
	}

	iterations := 1000

	for _, testError := range testErrors {
		fmt.Printf("  %s:\n", testError.name)

		// Test JSON serialization
		start := time.Now()
		var totalSize int64
		for i := 0; i < iterations; i++ {
			data, err := json.Marshal(testError.err)
			if err != nil {
				fmt.Printf("    JSON Serialization Error: %v\n", err)
				continue
			}
			totalSize += int64(len(data))
		}
		jsonDuration := time.Since(start)

		avgJsonDuration := jsonDuration / time.Duration(iterations)
		avgJsonSize := float64(totalSize) / float64(iterations)
		jsonOpsPerSec := float64(iterations) / jsonDuration.Seconds()

		fmt.Printf("    JSON Serialization:\n")
		fmt.Printf("      Average Duration: %v\n", avgJsonDuration)
		fmt.Printf("      Average Size: %.0f bytes\n", avgJsonSize)
		fmt.Printf("      Ops/sec: %.0f\n", jsonOpsPerSec)

		// Test String conversion
		start = time.Now()
		for i := 0; i < iterations; i++ {
			_ = testError.err.Error()
		}
		stringDuration := time.Since(start)

		avgStringDuration := stringDuration / time.Duration(iterations)
		stringOpsPerSec := float64(iterations) / stringDuration.Seconds()

		fmt.Printf("    String Conversion:\n")
		fmt.Printf("      Average Duration: %v\n", avgStringDuration)
		fmt.Printf("      Ops/sec: %.0f\n", stringOpsPerSec)

		// Performance classification
		if avgJsonDuration <= 1*time.Microsecond {
			fmt.Printf("    Serialization Performance: ⚡ EXCELLENT (≤1μs)\n")
		} else if avgJsonDuration <= 10*time.Microsecond {
			fmt.Printf("    Serialization Performance: ✅ GOOD (≤10μs)\n")
		} else if avgJsonDuration <= 100*time.Microsecond {
			fmt.Printf("    Serialization Performance: ⚠️ ACCEPTABLE (≤100μs)\n")
		} else {
			fmt.Printf("    Serialization Performance: ❌ SLOW (>100μs)\n")
		}
	}
}

// errorChainBenchmarks demonstrates error chaining performance
func errorChainBenchmarks() {
	fmt.Println("\n🔗 Error Chain Benchmarks:")

	chainLengths := []int{1, 5, 10, 20, 50}
	iterations := 1000

	for _, chainLength := range chainLengths {
		fmt.Printf("  Chain Length %d:\n", chainLength)

		// Benchmark chain creation
		start := time.Now()
		for i := 0; i < iterations; i++ {
			createErrorChain(chainLength)
		}
		creationDuration := time.Since(start)

		avgCreationDuration := creationDuration / time.Duration(iterations)
		creationOpsPerSec := float64(iterations) / creationDuration.Seconds()

		fmt.Printf("    Chain Creation:\n")
		fmt.Printf("      Average Duration: %v\n", avgCreationDuration)
		fmt.Printf("      Ops/sec: %.0f\n", creationOpsPerSec)

		// Benchmark chain traversal
		testChain := createErrorChain(chainLength)
		start = time.Now()
		for i := 0; i < iterations; i++ {
			traverseErrorChain(testChain)
		}
		traversalDuration := time.Since(start)

		avgTraversalDuration := traversalDuration / time.Duration(iterations)
		traversalOpsPerSec := float64(iterations) / traversalDuration.Seconds()

		fmt.Printf("    Chain Traversal:\n")
		fmt.Printf("      Average Duration: %v\n", avgTraversalDuration)
		fmt.Printf("      Ops/sec: %.0f\n", traversalOpsPerSec)

		// Performance classification
		if avgCreationDuration <= time.Duration(chainLength)*time.Microsecond {
			fmt.Printf("    Chain Performance: ⚡ EXCELLENT (≤1μs per link)\n")
		} else if avgCreationDuration <= time.Duration(chainLength)*5*time.Microsecond {
			fmt.Printf("    Chain Performance: ✅ GOOD (≤5μs per link)\n")
		} else if avgCreationDuration <= time.Duration(chainLength)*10*time.Microsecond {
			fmt.Printf("    Chain Performance: ⚠️ ACCEPTABLE (≤10μs per link)\n")
		} else {
			fmt.Printf("    Chain Performance: ❌ SLOW (>10μs per link)\n")
		}
	}
}

// builderPatternBenchmarks demonstrates builder pattern performance
func builderPatternBenchmarks() {
	fmt.Println("\n🏗️ Builder Pattern Benchmarks:")

	scenarios := []struct {
		name      string
		builderFn func() interfaces.DomainErrorInterface
	}{
		{
			"Minimal Builder",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("MINIMAL").
					Build()
			},
		},
		{
			"Fluent Builder",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("FLUENT").
					WithMessage("Fluent builder test").
					WithType(string(types.ErrorTypeValidation)).
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					Build()
			},
		},
		{
			"Step-by-Step Builder",
			func() interfaces.DomainErrorInterface {
				builder := factory.GetDefaultFactory().Builder()
				builder.WithCode("STEP_BY_STEP")
				builder.WithMessage("Step by step builder test")
				builder.WithType(string(types.ErrorTypeBusinessRule))
				builder.WithSeverity(interfaces.Severity(types.SeverityHigh))
				builder.WithDetail("key1", "value1")
				builder.WithDetail("key2", "value2")
				builder.WithTag("tag1")
				builder.WithTag("tag2")
				return builder.Build()
			},
		},
		{
			"Reused Builder",
			func() interfaces.DomainErrorInterface {
				// Simulate builder reuse (not recommended but tested for performance)
				builder := factory.GetDefaultFactory().Builder()
				return builder.
					WithCode("REUSED").
					WithMessage("Reused builder test").
					Build()
			},
		},
	}

	iterations := 10000

	for _, scenario := range scenarios {
		fmt.Printf("  %s:\n", scenario.name)

		// Warmup
		for i := 0; i < 100; i++ {
			_ = scenario.builderFn()
		}

		// Benchmark
		start := time.Now()
		for i := 0; i < iterations; i++ {
			_ = scenario.builderFn()
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(iterations)
		opsPerSec := float64(iterations) / duration.Seconds()

		fmt.Printf("    Average Duration: %v\n", avgDuration)
		fmt.Printf("    Ops/sec: %.0f\n", opsPerSec)

		// Performance classification
		if avgDuration <= 500*time.Nanosecond {
			fmt.Printf("    Builder Performance: ⚡ EXCELLENT (≤500ns)\n")
		} else if avgDuration <= 2*time.Microsecond {
			fmt.Printf("    Builder Performance: ✅ GOOD (≤2μs)\n")
		} else if avgDuration <= 10*time.Microsecond {
			fmt.Printf("    Builder Performance: ⚠️ ACCEPTABLE (≤10μs)\n")
		} else {
			fmt.Printf("    Builder Performance: ❌ SLOW (>10μs)\n")
		}
	}
}

// factoryBenchmarks demonstrates factory pattern performance
func factoryBenchmarks() {
	fmt.Println("\n🏭 Factory Benchmarks:")

	scenarios := []struct {
		name      string
		factoryFn func() interfaces.DomainErrorInterface
	}{
		{
			"Default Factory",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("DEFAULT_FACTORY").
					WithMessage("Default factory test").
					Build()
			},
		},
		{
			"Cached Factory Access",
			func() interfaces.DomainErrorInterface {
				f := factory.GetDefaultFactory() // Cache reference
				return f.Builder().
					WithCode("CACHED_FACTORY").
					WithMessage("Cached factory test").
					Build()
			},
		},
		{
			"Direct Builder Creation",
			func() interfaces.DomainErrorInterface {
				// Direct instantiation (if available)
				return factory.GetDefaultFactory().Builder().
					WithCode("DIRECT_BUILDER").
					WithMessage("Direct builder test").
					Build()
			},
		},
	}

	iterations := 10000

	for _, scenario := range scenarios {
		fmt.Printf("  %s:\n", scenario.name)

		// Warmup
		for i := 0; i < 100; i++ {
			_ = scenario.factoryFn()
		}

		// Benchmark
		start := time.Now()
		for i := 0; i < iterations; i++ {
			_ = scenario.factoryFn()
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(iterations)
		opsPerSec := float64(iterations) / duration.Seconds()

		fmt.Printf("    Average Duration: %v\n", avgDuration)
		fmt.Printf("    Ops/sec: %.0f\n", opsPerSec)

		// Performance classification
		if avgDuration <= 1*time.Microsecond {
			fmt.Printf("    Factory Performance: ⚡ EXCELLENT (≤1μs)\n")
		} else if avgDuration <= 5*time.Microsecond {
			fmt.Printf("    Factory Performance: ✅ GOOD (≤5μs)\n")
		} else if avgDuration <= 20*time.Microsecond {
			fmt.Printf("    Factory Performance: ⚠️ ACCEPTABLE (≤20μs)\n")
		} else {
			fmt.Printf("    Factory Performance: ❌ SLOW (>20μs)\n")
		}
	}
}

// optimizationTechniques demonstrates various optimization techniques
func optimizationTechniques() {
	fmt.Println("\n🚀 Optimization Techniques:")

	fmt.Printf("  Object Pooling Simulation:\n")
	objectPoolingBenchmark()

	fmt.Printf("\n  String Interning Simulation:\n")
	stringInterningBenchmark()

	fmt.Printf("\n  Lazy Initialization:\n")
	lazyInitializationBenchmark()

	fmt.Printf("\n  Memory Pre-allocation:\n")
	preallocationBenchmark()
}

// Helper functions

func createErrorChain(length int) interfaces.DomainErrorInterface {
	var currentError interfaces.DomainErrorInterface

	for i := 0; i < length; i++ {
		builder := factory.GetDefaultFactory().Builder().
			WithCode(fmt.Sprintf("CHAIN_ERROR_%d", i)).
			WithMessage(fmt.Sprintf("Chain error level %d", i))

		if currentError != nil {
			builder.WithDetail("previous_error", currentError)
		}

		currentError = builder.Build()
	}

	return currentError
}

func traverseErrorChain(err interfaces.DomainErrorInterface) int {
	count := 0
	current := err

	for current != nil {
		count++
		// Simulate traversal by accessing error properties
		_ = current.Code()
		_ = current.Error()
		_ = current.Details()

		// Try to get next error in chain (simplified)
		if details := current.Details(); details != nil {
			if nextErr, ok := details["previous_error"].(interfaces.DomainErrorInterface); ok {
				current = nextErr
			} else {
				break
			}
		} else {
			break
		}
	}

	return count
}

func testThreadSafety() {
	const numGoroutines = 100
	const numOperationsPerGoroutine = 100

	var wg sync.WaitGroup
	errorChannel := make(chan interfaces.DomainErrorInterface, numGoroutines*numOperationsPerGoroutine)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				err := factory.GetDefaultFactory().Builder().
					WithCode("THREAD_SAFETY_TEST").
					WithMessage("Thread safety test").
					WithDetail("goroutine_id", goroutineID).
					WithDetail("operation_id", j).
					Build()

				errorChannel <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorChannel)

	// Verify all errors were created correctly
	count := 0
	for err := range errorChannel {
		count++
		// Verify error integrity
		if err.Code() != "THREAD_SAFETY_TEST" {
			fmt.Printf("    ❌ Thread Safety Issue: Invalid error code\n")
			return
		}
	}

	expectedCount := numGoroutines * numOperationsPerGoroutine
	if count == expectedCount {
		fmt.Printf("    ✅ Thread Safety: PASSED (%d/%d errors created correctly)\n", count, expectedCount)
	} else {
		fmt.Printf("    ❌ Thread Safety: FAILED (%d/%d errors created)\n", count, expectedCount)
	}
}

func objectPoolingBenchmark() {
	// Simulate object pooling benefits
	pool := sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}

	iterations := 10000

	// Without pooling
	start := time.Now()
	for i := 0; i < iterations; i++ {
		details := make(map[string]interface{})
		details["key"] = "value"
		// Simulate usage
		_ = details
	}
	withoutPooling := time.Since(start)

	// With pooling
	start = time.Now()
	for i := 0; i < iterations; i++ {
		details := pool.Get().(map[string]interface{})
		details["key"] = "value"
		// Clear and return to pool
		for k := range details {
			delete(details, k)
		}
		pool.Put(details)
	}
	withPooling := time.Since(start)

	fmt.Printf("    Without Pooling: %v\n", withoutPooling)
	fmt.Printf("    With Pooling: %v\n", withPooling)
	improvement := float64(withoutPooling-withPooling) / float64(withoutPooling) * 100
	fmt.Printf("    Improvement: %.1f%%\n", improvement)
}

func stringInterningBenchmark() {
	// Simulate string interning benefits
	commonStrings := []string{
		"VALIDATION_ERROR",
		"NOT_FOUND",
		"INTERNAL_ERROR",
		"TIMEOUT",
		"UNAUTHORIZED",
	}

	internPool := make(map[string]string)
	for _, s := range commonStrings {
		internPool[s] = s
	}

	iterations := 10000

	// Without interning
	start := time.Now()
	for i := 0; i < iterations; i++ {
		code := commonStrings[i%len(commonStrings)]
		_ = code
	}
	withoutInterning := time.Since(start)

	// With interning
	start = time.Now()
	for i := 0; i < iterations; i++ {
		originalCode := commonStrings[i%len(commonStrings)]
		code := internPool[originalCode]
		_ = code
	}
	withInterning := time.Since(start)

	fmt.Printf("    Without Interning: %v\n", withoutInterning)
	fmt.Printf("    With Interning: %v\n", withInterning)
	improvement := float64(withoutInterning-withInterning) / float64(withoutInterning) * 100
	fmt.Printf("    Improvement: %.1f%%\n", improvement)
}

func lazyInitializationBenchmark() {
	// Simulate lazy initialization benefits
	type LazyError struct {
		code        string
		message     string
		details     map[string]interface{}
		detailsInit sync.Once
	}

	getDetails := func(le *LazyError) map[string]interface{} {
		le.detailsInit.Do(func() {
			le.details = map[string]interface{}{
				"timestamp": time.Now(),
				"source":    "lazy_benchmark",
			}
		})
		return le.details
	}

	iterations := 10000

	// Eager initialization
	start := time.Now()
	for i := 0; i < iterations; i++ {
		err := &LazyError{
			code:    "EAGER_ERROR",
			message: "Eager initialization",
			details: map[string]interface{}{
				"timestamp": time.Now(),
				"source":    "eager_benchmark",
			},
		}
		_ = err
	}
	eagerTime := time.Since(start)

	// Lazy initialization (but access details)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		err := &LazyError{
			code:    "LAZY_ERROR",
			message: "Lazy initialization",
		}
		_ = getDetails(err) // Force initialization
	}
	lazyTime := time.Since(start)

	fmt.Printf("    Eager Initialization: %v\n", eagerTime)
	fmt.Printf("    Lazy Initialization: %v\n", lazyTime)
	if lazyTime < eagerTime {
		improvement := float64(eagerTime-lazyTime) / float64(eagerTime) * 100
		fmt.Printf("    Improvement: %.1f%%\n", improvement)
	} else {
		overhead := float64(lazyTime-eagerTime) / float64(eagerTime) * 100
		fmt.Printf("    Overhead: %.1f%%\n", overhead)
	}
}

func preallocationBenchmark() {
	iterations := 1000

	// Without pre-allocation
	start := time.Now()
	for i := 0; i < iterations; i++ {
		details := make(map[string]interface{})
		for j := 0; j < 10; j++ {
			details[fmt.Sprintf("key_%d", j)] = fmt.Sprintf("value_%d", j)
		}
	}
	withoutPrealloc := time.Since(start)

	// With pre-allocation
	start = time.Now()
	for i := 0; i < iterations; i++ {
		details := make(map[string]interface{}, 10) // Pre-allocate capacity
		for j := 0; j < 10; j++ {
			details[fmt.Sprintf("key_%d", j)] = fmt.Sprintf("value_%d", j)
		}
	}
	withPrealloc := time.Since(start)

	fmt.Printf("    Without Pre-allocation: %v\n", withoutPrealloc)
	fmt.Printf("    With Pre-allocation: %v\n", withPrealloc)
	improvement := float64(withoutPrealloc-withPrealloc) / float64(withoutPrealloc) * 100
	fmt.Printf("    Improvement: %.1f%%\n", improvement)
}

// Benchmark tests (for use with go test -bench=.)

func BenchmarkSimpleErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = factory.GetDefaultFactory().Builder().
			WithCode("BENCHMARK_ERROR").
			WithMessage("Benchmark test error").
			Build()
	}
}

func BenchmarkComplexErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = factory.GetDefaultFactory().Builder().
			WithCode("COMPLEX_BENCHMARK_ERROR").
			WithMessage("Complex benchmark test error").
			WithType(string(types.ErrorTypeBusinessRule)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("user_id", "user123").
			WithDetail("operation", "benchmark_test").
			WithDetail("iteration", i).
			WithTag("benchmark").
			WithTag("performance").
			Build()
	}
}

func BenchmarkErrorSerialization(b *testing.B) {
	err := factory.GetDefaultFactory().Builder().
		WithCode("SERIALIZATION_BENCHMARK").
		WithMessage("Serialization benchmark test").
		WithType(string(types.ErrorTypeValidation)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("benchmark", true).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err)
	}
}

func BenchmarkConcurrentErrorCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = factory.GetDefaultFactory().Builder().
				WithCode("CONCURRENT_BENCHMARK").
				WithMessage("Concurrent benchmark test").
				Build()
		}
	})
}
