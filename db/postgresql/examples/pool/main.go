package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func main() {
	// Pool management example
	ctx := context.Background()

	// Create configuration for pool
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	// Configure pool settings
	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(20),
			postgresql.WithMinConns(5),
			postgresql.WithMaxConnLifetime(30*time.Minute),
			postgresql.WithMaxConnIdleTime(5*time.Minute),
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

	// Example 1: Basic pool usage
	if err := basicPoolExample(ctx, provider, cfg); err != nil {
		log.Printf("Basic pool example failed: %v", err)
	}

	// Example 2: Pool statistics monitoring
	if err := poolStatsExample(ctx, provider, cfg); err != nil {
		log.Printf("Pool stats example failed: %v", err)
	}

	// Example 3: Concurrent operations
	if err := concurrentOperationsExample(ctx, provider, cfg); err != nil {
		log.Printf("Concurrent operations example failed: %v", err)
	}

	// Example 4: Pool lifecycle management
	if err := poolLifecycleExample(ctx, provider, cfg); err != nil {
		log.Printf("Pool lifecycle example failed: %v", err)
	}

	fmt.Println("Pool examples completed!")
}

func basicPoolExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("=== Basic Pool Example ===")

	// Create connection pool
	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Pool creation would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Get pool statistics
	stats := pool.Stats()
	fmt.Printf("📊 Initial Pool Stats:\n")
	fmt.Printf("  - Total Connections: %d\n", stats.TotalConns)
	fmt.Printf("  - Acquired Connections: %d\n", stats.AcquiredConns)
	fmt.Printf("  - Idle Connections: %d\n", stats.IdleConns)
	fmt.Printf("  - Max Connections: %d\n", stats.MaxConns)

	// Use AcquireFunc for automatic resource management
	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		fmt.Println("🔌 Connection acquired from pool")

		// Perform a simple query
		var result int
		row := conn.QueryRow(ctx, "SELECT 1")
		if err := row.Scan(&result); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		fmt.Printf("✅ Query executed successfully, result: %d\n", result)
		return nil
	})

	if err != nil {
		fmt.Printf("Note: Pool operations would require actual database connection: %v\n", err)
	} else {
		fmt.Println("✅ Pool operations completed successfully")
	}

	// Show updated statistics
	updatedStats := pool.Stats()
	fmt.Printf("📊 Updated Pool Stats:\n")
	fmt.Printf("  - Total Connections: %d\n", updatedStats.TotalConns)
	fmt.Printf("  - Acquired Connections: %d\n", updatedStats.AcquiredConns)
	fmt.Printf("  - Idle Connections: %d\n", updatedStats.IdleConns)

	return nil
}

func poolStatsExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Pool Statistics Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Pool stats example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Monitor pool statistics over time
	fmt.Println("📈 Monitoring pool statistics for 5 seconds...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			stats := pool.Stats()
			fmt.Printf("⏰ [%s] Total: %d, Acquired: %d, Idle: %d, Constructing: %d\n",
				time.Now().Format("15:04:05"),
				stats.TotalConns,
				stats.AcquiredConns,
				stats.IdleConns,
				stats.ConstructingConns,
			)
		case <-timeout:
			fmt.Println("✅ Pool monitoring completed")
			return nil
		}
	}
}

func concurrentOperationsExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Concurrent Operations Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Concurrent operations example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Number of concurrent operations
	const numWorkers = 10
	const operationsPerWorker = 5

	var wg sync.WaitGroup
	var successCount, errorCount int64
	var mu sync.Mutex

	fmt.Printf("🚀 Starting %d workers with %d operations each\n", numWorkers, operationsPerWorker)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
					// Simulate work
					time.Sleep(100 * time.Millisecond)

					// Execute a query
					var result int
					row := conn.QueryRow(ctx, "SELECT $1", workerID*100+j)
					return row.Scan(&result)
				})

				mu.Lock()
				if err != nil {
					errorCount++
					fmt.Printf("❌ Worker %d operation %d failed: %v\n", workerID, j, err)
				} else {
					successCount++
					fmt.Printf("✅ Worker %d operation %d completed\n", workerID, j)
				}
				mu.Unlock()
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()

	fmt.Printf("📊 Concurrent operations summary:\n")
	fmt.Printf("  - Successful operations: %d\n", successCount)
	fmt.Printf("  - Failed operations: %d\n", errorCount)

	// Final pool statistics
	stats := pool.Stats()
	fmt.Printf("📊 Final Pool Stats:\n")
	fmt.Printf("  - Total Connections: %d\n", stats.TotalConns)
	fmt.Printf("  - Acquired Connections: %d\n", stats.AcquiredConns)
	fmt.Printf("  - Idle Connections: %d\n", stats.IdleConns)

	return nil
}

func poolLifecycleExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Pool Lifecycle Example ===")

	fmt.Println("🔄 Creating pool...")
	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Pool lifecycle example would require actual database: %v\n", err)
		return nil
	}

	// Show initial state
	stats := pool.Stats()
	fmt.Printf("📊 Pool created - Initial connections: %d\n", stats.TotalConns)

	// Perform some operations to warm up the pool
	fmt.Println("🔥 Warming up pool with operations...")
	for i := 0; i < 3; i++ {
		err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		if err != nil {
			fmt.Printf("Warmup operation %d failed: %v\n", i, err)
		} else {
			fmt.Printf("✅ Warmup operation %d completed\n", i)
		}
	}

	// Show warmed up state
	stats = pool.Stats()
	fmt.Printf("📊 Pool warmed up - Total connections: %d, Idle: %d\n",
		stats.TotalConns, stats.IdleConns)

	// Health check
	fmt.Println("🏥 Performing pool health check...")
	if healthChecker, ok := pool.(interface{ HealthCheck(context.Context) error }); ok {
		if err := healthChecker.HealthCheck(ctx); err != nil {
			fmt.Printf("❌ Pool health check failed: %v\n", err)
		} else {
			fmt.Println("✅ Pool health check passed")
		}
	}

	// Graceful shutdown
	fmt.Println("🔄 Closing pool gracefully...")
	pool.Close()
	fmt.Println("✅ Pool closed successfully")

	return nil
}
