package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// PerformanceMetrics tracks performance data
type PerformanceMetrics struct {
	mu                 sync.RWMutex
	TotalQueries       int64
	TotalDuration      time.Duration
	MinDuration        time.Duration
	MaxDuration        time.Duration
	ErrorCount         int64
	SlowQueries        int64
	slowQueryThreshold time.Duration
	QueryTimes         []time.Duration
}

func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		slowQueryThreshold: 100 * time.Millisecond,
		MinDuration:        time.Hour, // Start with high value
		QueryTimes:         make([]time.Duration, 0),
	}
}

func (pm *PerformanceMetrics) RecordQuery(duration time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.TotalQueries++
	pm.TotalDuration += duration
	pm.QueryTimes = append(pm.QueryTimes, duration)

	if duration < pm.MinDuration {
		pm.MinDuration = duration
	}
	if duration > pm.MaxDuration {
		pm.MaxDuration = duration
	}

	if err != nil {
		pm.ErrorCount++
	}

	if duration > pm.slowQueryThreshold {
		pm.SlowQueries++
	}
}

func (pm *PerformanceMetrics) GetStats() (int64, time.Duration, time.Duration, time.Duration, int64, int64, time.Duration) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	avgDuration := time.Duration(0)
	if pm.TotalQueries > 0 {
		avgDuration = pm.TotalDuration / time.Duration(pm.TotalQueries)
	}

	return pm.TotalQueries, pm.TotalDuration, pm.MinDuration, pm.MaxDuration, pm.ErrorCount, pm.SlowQueries, avgDuration
}

func (pm *PerformanceMetrics) GetPercentile(percentile float64) time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.QueryTimes) == 0 {
		return 0
	}

	// Simple percentile calculation (should use proper sorting for production)
	index := int(float64(len(pm.QueryTimes)) * percentile / 100.0)
	if index >= len(pm.QueryTimes) {
		index = len(pm.QueryTimes) - 1
	}

	return pm.QueryTimes[index]
}

// ConnectionPoolBenchmark benchmarks pool performance
type ConnectionPoolBenchmark struct {
	metrics *PerformanceMetrics
}

func NewConnectionPoolBenchmark() *ConnectionPoolBenchmark {
	return &ConnectionPoolBenchmark{
		metrics: NewPerformanceMetrics(),
	}
}

func main() {
	// Performance optimization and benchmarking example
	ctx := context.Background()

	// Create optimized configuration
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			// Optimized pool settings
			postgresql.WithMaxConns(int32(50)),
			postgresql.WithMinConns(int32(10)),
			postgresql.WithMaxConnLifetime(1*time.Hour),
			postgresql.WithMaxConnIdleTime(10*time.Minute),
		)
		if err != nil {
			log.Fatalf("Failed to apply configuration: %v", err)
		}
	}

	// Create provider
	provider, err := postgresql.NewPGXProvider()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Println("🚀 Starting PostgreSQL Performance Examples")
	fmt.Println("💡 Note: These examples require a running PostgreSQL database")
	fmt.Println("🔧 If database is not available, examples will run in simulation mode")
	fmt.Println()

	// Example 1: Connection pool performance
	if err := demonstratePoolPerformance(ctx, provider, cfg); err != nil {
		log.Printf("Pool performance example failed: %v", err)
	}

	// Example 2: Query optimization
	if err := demonstrateQueryOptimization(ctx, provider, cfg); err != nil {
		log.Printf("Query optimization example failed: %v", err)
	}

	// Example 3: Batch operations performance
	if err := demonstrateBatchPerformance(ctx, provider, cfg); err != nil {
		log.Printf("Batch performance example failed: %v", err)
	}

	// Example 4: Concurrent operations benchmark
	if err := demonstrateConcurrentBenchmark(ctx, provider, cfg); err != nil {
		log.Printf("Concurrent benchmark failed: %v", err)
	}

	// Example 5: Memory optimization
	if err := demonstrateMemoryOptimization(ctx, provider, cfg); err != nil {
		log.Printf("Memory optimization example failed: %v", err)
	}

	// Example 6: Connection lifecycle optimization
	if err := demonstrateConnectionLifecycleOptimization(ctx, provider, cfg); err != nil {
		log.Printf("Connection lifecycle optimization failed: %v", err)
	}

	fmt.Println("Performance examples completed!")
}

func demonstratePoolPerformance(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("=== Connection Pool Performance Example ===")

	// Test different pool configurations
	poolConfigs := []struct {
		name     string
		maxConns int
		minConns int
	}{
		{"Small Pool", 5, 2},
		{"Medium Pool", 20, 5},
		{"Large Pool", 50, 10},
	}

	for _, poolConfig := range poolConfigs {
		fmt.Printf("\n🏊 Testing %s (Max: %d, Min: %d)\n",
			poolConfig.name, poolConfig.maxConns, poolConfig.minConns)

		// Create pool with specific configuration
		testCfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
		if defaultCfg, ok := testCfg.(*config.DefaultConfig); ok {
			err := defaultCfg.ApplyOptions(
				postgresql.WithMaxConns(int32(poolConfig.maxConns)),
				postgresql.WithMinConns(int32(poolConfig.minConns)),
			)
			if err != nil {
				fmt.Printf("  ❌ Failed to configure pool: %v\n", err)
				continue
			}
		}

		pool, err := provider.NewPool(ctx, testCfg)
		if err != nil {
			fmt.Printf("  💡 Pool performance test would require actual database: %v\n", err)
			fmt.Printf("  📊 Simulating pool performance metrics...\n")

			// Simulate performance metrics
			numOperations := 100
			simulatedAvgTime := time.Duration(1+poolConfig.maxConns/10) * time.Millisecond

			fmt.Printf("  📈 Simulated Results for %s:\n", poolConfig.name)
			fmt.Printf("    - Total Operations: %d\n", numOperations)
			fmt.Printf("    - Simulated Average Time: %v\n", simulatedAvgTime)
			fmt.Printf("    - Estimated Operations/sec: %.2f\n", float64(numOperations)/simulatedAvgTime.Seconds()*1000)
			fmt.Printf("    - Pool Configuration: Max=%d, Min=%d\n", poolConfig.maxConns, poolConfig.minConns)
			continue
		}

		// Test pool connectivity before proceeding
		fmt.Printf("  🔍 Testing pool connectivity...\n")

		var testErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					testErr = fmt.Errorf("connection test failed with panic: %v", r)
				}
			}()

			testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
				// Simple connectivity test
				return conn.HealthCheck(ctx)
			})
		}()

		if testErr != nil {
			fmt.Printf("  💡 Pool created but database connection failed: %v\n", testErr)
			fmt.Printf("  📊 Simulating pool performance metrics...\n")

			// Simulate performance metrics
			numOperations := 100
			simulatedAvgTime := time.Duration(1+poolConfig.maxConns/10) * time.Millisecond

			fmt.Printf("  📈 Simulated Results for %s:\n", poolConfig.name)
			fmt.Printf("    - Total Operations: %d\n", numOperations)
			fmt.Printf("    - Simulated Average Time: %v\n", simulatedAvgTime)
			fmt.Printf("    - Estimated Operations/sec: %.2f\n", float64(numOperations)/simulatedAvgTime.Seconds()*1000)
			fmt.Printf("    - Pool Configuration: Max=%d, Min=%d\n", poolConfig.maxConns, poolConfig.minConns)

			pool.Close()
			continue
		} // Benchmark pool acquisition
		metrics := NewPerformanceMetrics()
		numOperations := 100

		fmt.Printf("  📊 Running %d acquisition operations...\n", numOperations)
		start := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				opStart := time.Now()
				err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
					// Simulate some work
					time.Sleep(1 * time.Millisecond)
					return nil
				})
				opDuration := time.Since(opStart)

				metrics.RecordQuery(opDuration, err)

				if index%20 == 0 {
					fmt.Printf("    ⏱️  Operation %d: %v\n", index, opDuration)
				}
			}(i)
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Display results
		totalQueries, _, minTime, maxTime, errorCount, slowQueries, avgTime := metrics.GetStats()
		fmt.Printf("  📈 Results for %s:\n", poolConfig.name)
		fmt.Printf("    - Total Operations: %d\n", totalQueries)
		fmt.Printf("    - Total Time: %v\n", totalDuration)
		fmt.Printf("    - Average Time: %v\n", avgTime)
		fmt.Printf("    - Min Time: %v\n", minTime)
		fmt.Printf("    - Max Time: %v\n", maxTime)
		fmt.Printf("    - Error Count: %d\n", errorCount)
		fmt.Printf("    - Slow Operations: %d\n", slowQueries)
		fmt.Printf("    - Operations/sec: %.2f\n", float64(numOperations)/totalDuration.Seconds())

		// Pool stats
		stats := pool.Stats()
		fmt.Printf("    - Final Pool Size: %d\n", stats.TotalConns)
		fmt.Printf("    - Idle Connections: %d\n", stats.IdleConns)

		pool.Close()
	}

	return nil
}

func demonstrateQueryOptimization(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Query Optimization Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("💡 Query optimization would require actual database: %v\n", err)
		fmt.Printf("📊 Simulating query optimization patterns...\n")

		// Simulate different query patterns
		queryPatterns := []string{
			"Simple Select - Simulated: ~0.5ms",
			"Parameterized Query - Simulated: ~1.2ms",
			"Complex Calculation - Simulated: ~5.8ms",
			"System Query - Simulated: ~12.3ms",
		}

		for _, pattern := range queryPatterns {
			fmt.Printf("  🔍 %s\n", pattern)
		}

		fmt.Printf("\n  💡 Performance Tips:\n")
		fmt.Printf("    - Use parameterized queries for security and performance\n")
		fmt.Printf("    - Prepare statements for repeated queries\n")
		fmt.Printf("    - Index frequently queried columns\n")
		fmt.Printf("    - Monitor slow query logs\n")

		return nil
	}
	defer pool.Close()

	// Test pool connectivity before proceeding
	var testErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				testErr = fmt.Errorf("connection test failed with panic: %v", r)
			}
		}()

		testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple connectivity test
			return conn.HealthCheck(ctx)
		})
	}()

	if testErr != nil {
		fmt.Printf("💡 Query optimization would require actual database: %v\n", testErr)
		return nil
	}

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Test different query patterns
		queryTests := []struct {
			name  string
			query string
			args  []interface{}
		}{
			{
				"Simple Select",
				"SELECT 1 as result",
				nil,
			},
			{
				"Parameterized Query",
				"SELECT $1 + $2 as sum",
				[]interface{}{10, 20},
			},
			{
				"Complex Calculation",
				"SELECT generate_series(1, $1) as num",
				[]interface{}{1000},
			},
			{
				"System Query",
				"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1",
				[]interface{}{"public"},
			},
		}

		fmt.Println("📊 Testing different query patterns...")

		for _, test := range queryTests {
			fmt.Printf("\n  🔍 Testing: %s\n", test.name)

			// Run multiple iterations to get average
			iterations := 10
			metrics := NewPerformanceMetrics()

			for i := 0; i < iterations; i++ {
				start := time.Now()

				if test.args != nil {
					row := conn.QueryRow(ctx, test.query, test.args...)
					var result interface{}
					err := row.Scan(&result)

					duration := time.Since(start)
					metrics.RecordQuery(duration, err)
				} else {
					row := conn.QueryRow(ctx, test.query)
					var result interface{}
					err := row.Scan(&result)

					duration := time.Since(start)
					metrics.RecordQuery(duration, err)
				}
			}

			// Display results
			_, _, minTime, maxTime, errorCount, _, avgTime := metrics.GetStats()
			fmt.Printf("    - Average: %v\n", avgTime)
			fmt.Printf("    - Min: %v\n", minTime)
			fmt.Printf("    - Max: %v\n", maxTime)
			fmt.Printf("    - Errors: %d\n", errorCount)

			// Performance rating
			if avgTime < 1*time.Millisecond {
				fmt.Printf("    - Rating: ⚡ Excellent\n")
			} else if avgTime < 10*time.Millisecond {
				fmt.Printf("    - Rating: ✅ Good\n")
			} else if avgTime < 100*time.Millisecond {
				fmt.Printf("    - Rating: ⚠️  Acceptable\n")
			} else {
				fmt.Printf("    - Rating: 🐌 Needs Optimization\n")
			}
		}

		// Demonstrate prepared statement optimization
		fmt.Println("\n📝 Testing prepared statement vs regular query...")

		// Regular query benchmark
		regularMetrics := NewPerformanceMetrics()
		iterations := 50

		fmt.Printf("  🔄 Running %d regular queries...\n", iterations)
		for i := 0; i < iterations; i++ {
			start := time.Now()
			row := conn.QueryRow(ctx, "SELECT $1 * $2 as product", i, i+1)
			var result int
			err := row.Scan(&result)
			duration := time.Since(start)
			regularMetrics.RecordQuery(duration, err)
		}

		_, _, _, _, _, _, regularAvg := regularMetrics.GetStats()
		fmt.Printf("    Regular Query Average: %v\n", regularAvg)

		// Note: Prepared statements would require actual implementation
		fmt.Printf("    ℹ️  Prepared statement optimization would require driver-specific implementation\n")
		fmt.Printf("    💡 Generally 2-5x faster for repeated queries\n")

		return nil
	})
}

func demonstrateBatchPerformance(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Batch Operations Performance Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("💡 Batch performance would require actual database: %v\n", err)
		fmt.Printf("📊 Simulating batch operation comparison...\n")

		fmt.Printf("  📈 Simulated Performance Comparison (1000 records):\n")
		fmt.Printf("    - Individual inserts: ~15.2s (65.8 ops/sec)\n")
		fmt.Printf("    - Batch inserts: ~2.1s (476.2 ops/sec)\n")
		fmt.Printf("    - Speedup: 7.2x faster with batching\n")
		fmt.Printf("    - Rating: ⚡ Excellent optimization\n")

		fmt.Printf("\n  💡 Batch Operation Tips:\n")
		fmt.Printf("    - Use VALUES clause for multiple inserts\n")
		fmt.Printf("    - Optimal batch size: 100-1000 records\n")
		fmt.Printf("    - Use transactions for consistency\n")
		fmt.Printf("    - Consider COPY for very large datasets\n")

		return nil
	}
	defer pool.Close()

	// Test pool connectivity before proceeding
	var testErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				testErr = fmt.Errorf("connection test failed with panic: %v", r)
			}
		}()

		testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple connectivity test
			return conn.HealthCheck(ctx)
		})
	}()

	if testErr != nil {
		fmt.Printf("💡 Batch performance would require actual database: %v\n", testErr)
		return nil
	}

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Create test table
		_, err := conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS performance_test (
				id SERIAL PRIMARY KEY,
				data TEXT NOT NULL,
				value INTEGER NOT NULL,
				created_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			fmt.Printf("  ❌ Failed to create test table: %v\n", err)
			return nil // Don't fail the example
		}

		// Clean existing data
		_, err = conn.Exec(ctx, "TRUNCATE TABLE performance_test")
		if err != nil {
			fmt.Printf("  ⚠️  Failed to clean test table: %v\n", err)
		}

		fmt.Println("📊 Comparing individual inserts vs batch operations...")

		recordCount := 1000

		// Test 1: Individual inserts
		fmt.Printf("\n  🔄 Testing %d individual inserts...\n", recordCount)
		start := time.Now()

		for i := 0; i < recordCount; i++ {
			_, err := conn.Exec(ctx,
				"INSERT INTO performance_test (data, value) VALUES ($1, $2)",
				fmt.Sprintf("data_%d", i), i)
			if err != nil {
				fmt.Printf("    ❌ Insert %d failed: %v\n", i, err)
				break
			}

			if i%100 == 0 && i > 0 {
				fmt.Printf("    ⏱️  Inserted %d records...\n", i)
			}
		}

		individualDuration := time.Since(start)
		fmt.Printf("    ✅ Individual inserts completed in: %v\n", individualDuration)
		fmt.Printf("    📈 Rate: %.2f inserts/sec\n", float64(recordCount)/individualDuration.Seconds())

		// Clean for next test
		_, err = conn.Exec(ctx, "TRUNCATE TABLE performance_test")
		if err != nil {
			fmt.Printf("  ⚠️  Failed to clean for batch test: %v\n", err)
		}

		// Test 2: Batch insert using VALUES clause
		fmt.Printf("\n  📦 Testing batch insert with VALUES clause...\n")
		start = time.Now()

		batchSize := 100
		for i := 0; i < recordCount; i += batchSize {
			// Build batch VALUES clause
			values := ""
			args := make([]interface{}, 0, batchSize*2)

			for j := 0; j < batchSize && i+j < recordCount; j++ {
				if j > 0 {
					values += ","
				}
				argIndex := j * 2
				values += fmt.Sprintf("($%d, $%d)", argIndex+1, argIndex+2)
				args = append(args, fmt.Sprintf("data_%d", i+j), i+j)
			}

			query := fmt.Sprintf("INSERT INTO performance_test (data, value) VALUES %s", values)
			_, err := conn.Exec(ctx, query, args...)
			if err != nil {
				fmt.Printf("    ❌ Batch insert failed: %v\n", err)
				break
			}

			if i%500 == 0 && i > 0 {
				fmt.Printf("    ⏱️  Batch inserted %d records...\n", i)
			}
		}

		batchDuration := time.Since(start)
		fmt.Printf("    ✅ Batch inserts completed in: %v\n", batchDuration)
		fmt.Printf("    📈 Rate: %.2f inserts/sec\n", float64(recordCount)/batchDuration.Seconds())

		// Performance comparison
		if individualDuration > 0 && batchDuration > 0 {
			speedup := float64(individualDuration) / float64(batchDuration)
			fmt.Printf("\n  📊 Performance Comparison:\n")
			fmt.Printf("    - Individual: %v (%.2f ops/sec)\n",
				individualDuration, float64(recordCount)/individualDuration.Seconds())
			fmt.Printf("    - Batch: %v (%.2f ops/sec)\n",
				batchDuration, float64(recordCount)/batchDuration.Seconds())
			fmt.Printf("    - Speedup: %.2fx faster\n", speedup)

			if speedup > 5 {
				fmt.Printf("    - Rating: ⚡ Excellent optimization\n")
			} else if speedup > 2 {
				fmt.Printf("    - Rating: ✅ Good optimization\n")
			} else {
				fmt.Printf("    - Rating: ⚠️  Minimal improvement\n")
			}
		}

		// Verify data
		var count int
		row := conn.QueryRow(ctx, "SELECT COUNT(*) FROM performance_test")
		err = row.Scan(&count)
		if err != nil {
			fmt.Printf("  ❌ Failed to count records: %v\n", err)
		} else {
			fmt.Printf("    ✅ Verified %d records in table\n", count)
		}

		return nil
	})
}

func demonstrateConcurrentBenchmark(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Concurrent Operations Benchmark ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("💡 Concurrent benchmark would require actual database: %v\n", err)
		fmt.Printf("📊 Simulating concurrent performance analysis...\n")

		concurrencyLevels := []int{1, 5, 10, 20, 50}

		for _, workers := range concurrencyLevels {
			fmt.Printf("\n  🚀 Simulated results for %d concurrent workers:\n", workers)

			// Simulate performance degradation with more workers
			baseOpsPerSec := 500.0
			efficiency := 1.0 - (float64(workers)-1)*0.02 // 2% degradation per worker
			if efficiency < 0.3 {
				efficiency = 0.3 // minimum efficiency
			}

			simulatedOpsPerSec := baseOpsPerSec * efficiency * float64(workers)
			avgLatency := time.Duration(float64(time.Millisecond) / efficiency)

			fmt.Printf("      - Throughput: %.2f ops/sec\n", simulatedOpsPerSec)
			fmt.Printf("      - Average Latency: %v\n", avgLatency)
			fmt.Printf("      - Error Rate: 0.00%%\n")

			if simulatedOpsPerSec > 1000 {
				fmt.Printf("      - Rating: ⚡ Excellent\n")
			} else if simulatedOpsPerSec > 500 {
				fmt.Printf("      - Rating: ✅ Good\n")
			} else {
				fmt.Printf("      - Rating: ⚠️  Acceptable\n")
			}
		}

		fmt.Printf("\n  💡 Concurrency Tips:\n")
		fmt.Printf("    - Optimal concurrency = CPU cores × 2\n")
		fmt.Printf("    - Monitor pool exhaustion\n")
		fmt.Printf("    - Use connection pooling\n")
		fmt.Printf("    - Consider async operations\n")

		return nil
	}
	defer pool.Close()

	// Test pool connectivity before proceeding
	var testErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				testErr = fmt.Errorf("connection test failed with panic: %v", r)
			}
		}()

		testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple connectivity test
			return conn.HealthCheck(ctx)
		})
	}()

	if testErr != nil {
		fmt.Printf("💡 Concurrent benchmark would require actual database: %v\n", testErr)
		return nil
	}

	// Test different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 20, 50}
	operationsPerWorker := 50

	fmt.Printf("📊 Testing concurrency levels with %d operations per worker...\n", operationsPerWorker)

	for _, workers := range concurrencyLevels {
		fmt.Printf("\n  🚀 Testing %d concurrent workers...\n", workers)

		var wg sync.WaitGroup
		metrics := NewPerformanceMetrics()
		start := time.Now()

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operationsPerWorker; j++ {
					opStart := time.Now()

					err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
						// Simulate work
						row := conn.QueryRow(ctx, "SELECT $1 * $2 as result", workerID, j)
						var result int
						return row.Scan(&result)
					})

					opDuration := time.Since(opStart)
					metrics.RecordQuery(opDuration, err)
				}
			}(i)
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Calculate results
		totalOps := workers * operationsPerWorker
		totalQueries, _, minTime, maxTime, errorCount, _, avgTime := metrics.GetStats()

		fmt.Printf("    📈 Results for %d workers:\n", workers)
		fmt.Printf("      - Total Operations: %d\n", totalOps)
		fmt.Printf("      - Total Time: %v\n", totalDuration)
		fmt.Printf("      - Throughput: %.2f ops/sec\n", float64(totalOps)/totalDuration.Seconds())
		fmt.Printf("      - Average Latency: %v\n", avgTime)
		fmt.Printf("      - Min Latency: %v\n", minTime)
		fmt.Printf("      - Max Latency: %v\n", maxTime)
		fmt.Printf("      - Error Rate: %.2f%%\n", float64(errorCount)/float64(totalQueries)*100)

		// Pool statistics
		stats := pool.Stats()
		fmt.Printf("      - Pool Connections: %d\n", stats.TotalConns)
		fmt.Printf("      - Active Connections: %d\n", stats.AcquiredConns)

		// Performance rating
		opsPerSec := float64(totalOps) / totalDuration.Seconds()
		if opsPerSec > 1000 {
			fmt.Printf("      - Rating: ⚡ Excellent\n")
		} else if opsPerSec > 500 {
			fmt.Printf("      - Rating: ✅ Good\n")
		} else if opsPerSec > 100 {
			fmt.Printf("      - Rating: ⚠️  Acceptable\n")
		} else {
			fmt.Printf("      - Rating: 🐌 Needs Optimization\n")
		}

		// Brief pause between tests
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func demonstrateMemoryOptimization(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Memory Optimization Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("💡 Memory optimization would require actual database: %v\n", err)
		fmt.Printf("📊 Simulating memory usage analysis...\n")

		// Get actual memory stats for demonstration
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)

		fmt.Printf("  📈 Current Memory Statistics:\n")
		fmt.Printf("    - Current allocation: %d KB\n", m.Alloc/1024)
		fmt.Printf("    - Total allocations: %d KB\n", m.TotalAlloc/1024)
		fmt.Printf("    - System memory: %d KB\n", m.Sys/1024)
		fmt.Printf("    - GC cycles: %d\n", m.NumGC)

		fmt.Printf("\n  💡 Memory Optimization Tips:\n")
		fmt.Printf("    - Use connection pooling to reuse connections\n")
		fmt.Printf("    - Process large result sets in batches\n")
		fmt.Printf("    - Close rows and statements explicitly\n")
		fmt.Printf("    - Use prepared statements for repeated queries\n")
		fmt.Printf("    - Monitor and tune pool sizes\n")
		fmt.Printf("    - Consider streaming for large datasets\n")

		return nil
	}
	defer pool.Close()

	// Test pool connectivity before proceeding
	var testErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				testErr = fmt.Errorf("connection test failed with panic: %v", r)
			}
		}()

		testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple connectivity test
			return conn.HealthCheck(ctx)
		})
	}()

	if testErr != nil {
		fmt.Printf("💡 Memory optimization would require actual database: %v\n", testErr)
		return nil
	}

	// Memory usage tracking
	var m runtime.MemStats

	// Baseline memory
	runtime.GC()
	runtime.ReadMemStats(&m)
	baselineAlloc := m.Alloc

	fmt.Printf("📊 Memory usage analysis...\n")
	fmt.Printf("  📈 Baseline memory: %d KB\n", baselineAlloc/1024)

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Test 1: Large result set handling
		fmt.Printf("\n  🔍 Testing large result set handling...\n")

		// Simulate large query result
		start := time.Now()
		rows, err := conn.Query(ctx, "SELECT generate_series(1, 10000) as num, md5(random()::text) as hash")
		if err != nil {
			fmt.Printf("    ❌ Failed to execute large query: %v\n", err)
			return nil
		}
		defer rows.Close()

		// Process results
		var processed int
		for rows.Next() {
			var num int
			var hash string
			if err := rows.Scan(&num, &hash); err != nil {
				fmt.Printf("    ❌ Failed to scan row: %v\n", err)
				break
			}
			processed++
		}

		duration := time.Since(start)

		// Check memory after large operation
		runtime.GC()
		runtime.ReadMemStats(&m)
		afterLargeQuery := m.Alloc

		fmt.Printf("    ✅ Processed %d rows in %v\n", processed, duration)
		fmt.Printf("    📈 Memory after large query: %d KB (delta: +%d KB)\n",
			afterLargeQuery/1024, (afterLargeQuery-baselineAlloc)/1024)

		// Test 2: Connection pooling memory efficiency
		fmt.Printf("\n  🏊 Testing connection pool memory efficiency...\n")

		// Create multiple connections rapidly
		numConnTests := 20
		var wg sync.WaitGroup

		start = time.Now()
		for i := 0; i < numConnTests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
					// Simple operation
					row := conn.QueryRow(ctx, "SELECT $1", index)
					var result int
					return row.Scan(&result)
				})
				if err != nil {
					fmt.Printf("    ⚠️  Connection test %d failed: %v\n", index, err)
				}
			}(i)
		}

		wg.Wait()
		duration = time.Since(start)

		// Check memory after pool operations
		runtime.GC()
		runtime.ReadMemStats(&m)
		afterPoolTest := m.Alloc

		fmt.Printf("    ✅ Completed %d pool operations in %v\n", numConnTests, duration)
		fmt.Printf("    📈 Memory after pool test: %d KB (delta: +%d KB)\n",
			afterPoolTest/1024, (afterPoolTest-baselineAlloc)/1024)

		// Memory optimization recommendations
		fmt.Printf("\n  💡 Memory Optimization Tips:\n")
		fmt.Printf("    - Use connection pooling to reuse connections\n")
		fmt.Printf("    - Process large result sets in batches\n")
		fmt.Printf("    - Close rows and statements explicitly\n")
		fmt.Printf("    - Use prepared statements for repeated queries\n")
		fmt.Printf("    - Monitor and tune pool sizes\n")

		// Pool memory stats
		stats := pool.Stats()
		fmt.Printf("\n  📊 Current Pool Statistics:\n")
		fmt.Printf("    - Total Connections: %d\n", stats.TotalConns)
		fmt.Printf("    - Idle Connections: %d\n", stats.IdleConns)
		fmt.Printf("    - Acquired Connections: %d\n", stats.AcquiredConns)
		fmt.Printf("    - Max Connections: %d\n", stats.MaxConns)

		return nil
	})
}

func demonstrateConnectionLifecycleOptimization(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Connection Lifecycle Optimization Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("💡 Connection lifecycle optimization would require actual database: %v\n", err)
		fmt.Printf("📊 Simulating connection lifecycle analysis...\n")

		// Simulate different connection patterns
		patterns := []struct {
			name        string
			description string
			performance string
		}{
			{
				"Rapid Acquire/Release",
				"Many short-lived connections",
				"50 operations in ~125ms (400 ops/sec)",
			},
			{
				"Long-held Connection",
				"One longer-lived connection",
				"10 operations in ~110ms (90.9 ops/sec)",
			},
		}

		for _, pattern := range patterns {
			fmt.Printf("\n  🔍 Testing: %s\n", pattern.name)
			fmt.Printf("    Description: %s\n", pattern.description)
			fmt.Printf("    Simulated Performance: %s\n", pattern.performance)
		}

		fmt.Printf("\n  💡 Connection Lifecycle Optimization Tips:\n")
		fmt.Printf("    - Use AcquireFunc for automatic connection management\n")
		fmt.Printf("    - Configure appropriate MaxConnLifetime\n")
		fmt.Printf("    - Set reasonable MaxConnIdleTime\n")
		fmt.Printf("    - Monitor connection pool metrics\n")
		fmt.Printf("    - Implement health checks for long-lived connections\n")
		fmt.Printf("    - Use prepared statements to reduce connection overhead\n")

		return nil
	}
	defer pool.Close()

	// Test pool connectivity before proceeding
	var testErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				testErr = fmt.Errorf("connection test failed with panic: %v", r)
			}
		}()

		testErr = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple connectivity test
			return conn.HealthCheck(ctx)
		})
	}()

	if testErr != nil {
		fmt.Printf("💡 Connection lifecycle optimization would require actual database: %v\n", testErr)
		return nil
	}

	fmt.Printf("📊 Analyzing connection lifecycle performance...\n")

	// Test connection acquisition patterns
	acquisitionTests := []struct {
		name        string
		pattern     func() error
		description string
	}{
		{
			"Rapid Acquire/Release",
			func() error {
				for i := 0; i < 50; i++ {
					err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
						// Quick operation
						row := conn.QueryRow(ctx, "SELECT 1")
						var result int
						return row.Scan(&result)
					})
					if err != nil {
						return err
					}
				}
				return nil
			},
			"Many short-lived connections",
		},
		{
			"Long-held Connection",
			func() error {
				return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
					// Simulate longer operation
					for i := 0; i < 10; i++ {
						row := conn.QueryRow(ctx, "SELECT $1", i)
						var result int
						if err := row.Scan(&result); err != nil {
							return err
						}
						time.Sleep(10 * time.Millisecond)
					}
					return nil
				})
			},
			"One longer-lived connection",
		},
	}

	for _, test := range acquisitionTests {
		fmt.Printf("\n  🔍 Testing: %s\n", test.name)
		fmt.Printf("    Description: %s\n", test.description)

		// Get initial pool stats
		initialStats := pool.Stats()

		start := time.Now()
		err := test.pattern()
		duration := time.Since(start)

		// Get final pool stats
		finalStats := pool.Stats()

		if err != nil {
			fmt.Printf("    ❌ Test failed: %v\n", err)
		} else {
			fmt.Printf("    ✅ Test completed in: %v\n", duration)
		}

		fmt.Printf("    📈 Pool Changes:\n")
		fmt.Printf("      - Initial connections: %d\n", initialStats.TotalConns)
		fmt.Printf("      - Final connections: %d\n", finalStats.TotalConns)
		fmt.Printf("      - Connection delta: %+d\n", int(finalStats.TotalConns)-int(initialStats.TotalConns))
		fmt.Printf("      - Final idle: %d\n", finalStats.IdleConns)

		// Brief pause between tests
		time.Sleep(500 * time.Millisecond)
	}

	// Demonstrate health checking
	fmt.Printf("\n  🏥 Testing connection health checking...\n")

	healthCheckStart := time.Now()
	var healthyConns int

	// Test multiple connections for health
	for i := 0; i < 5; i++ {
		err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Health check operation
			if err := conn.HealthCheck(ctx); err != nil {
				return err
			}
			healthyConns++
			return nil
		})
		if err != nil {
			fmt.Printf("    ⚠️  Health check %d failed: %v\n", i, err)
		}
	}

	healthCheckDuration := time.Since(healthCheckStart)
	fmt.Printf("    ✅ Health checked %d connections in %v\n", healthyConns, healthCheckDuration)

	// Final optimization recommendations
	fmt.Printf("\n  💡 Connection Lifecycle Optimization Tips:\n")
	fmt.Printf("    - Use AcquireFunc for automatic connection management\n")
	fmt.Printf("    - Configure appropriate MaxConnLifetime\n")
	fmt.Printf("    - Set reasonable MaxConnIdleTime\n")
	fmt.Printf("    - Monitor connection pool metrics\n")
	fmt.Printf("    - Implement health checks for long-lived connections\n")
	fmt.Printf("    - Use prepared statements to reduce connection overhead\n")

	return nil
}
