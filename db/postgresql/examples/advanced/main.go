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
	// Advanced PostgreSQL provider example with custom hooks and middlewares
	ctx := context.Background()

	// Create configuration with advanced features
	cfg := postgresql.NewDefaultConfig("postgres://postgres:password@localhost:5432/testdb")

	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(50),
			postgresql.WithMinConns(10),
			postgresql.WithMaxConnLifetime(1*time.Hour),
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

	// Create pool
	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Demonstrate custom hooks
	if err := demonstrateCustomHooks(pool); err != nil {
		log.Fatalf("Custom hooks demo failed: %v", err)
	}

	// Demonstrate concurrent operations
	if err := demonstrateConcurrentOperations(ctx, pool); err != nil {
		log.Fatalf("Concurrent operations demo failed: %v", err)
	}

	// Demonstrate advanced querying
	if err := demonstrateAdvancedQuerying(ctx, pool); err != nil {
		log.Fatalf("Advanced querying demo failed: %v", err)
	}

	// Show detailed statistics
	showDetailedStatistics(pool)

	fmt.Println("Advanced example completed successfully!")
}

func demonstrateCustomHooks(pool interfaces.IPool) error {
	fmt.Println("\n--- Custom Hooks Demo ---")

	hookManager := pool.GetHookManager()

	// Register custom query logging hook
	queryLogHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Operation == "query" || ctx.Operation == "exec" {
			fmt.Printf("🪝 [HOOK] %s: %s\n", ctx.Operation, ctx.Query)
		}
		return &interfaces.HookResult{Continue: true}
	}

	// Register timing hook
	timingHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Duration > 0 {
			fmt.Printf("⏱️  [TIMING] %s took %v\n", ctx.Operation, ctx.Duration)
		}
		return &interfaces.HookResult{Continue: true}
	}

	// Register error logging hook
	errorHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Error != nil {
			fmt.Printf("❌ [ERROR] %s failed: %v\n", ctx.Operation, ctx.Error)
		}
		return &interfaces.HookResult{Continue: true}
	}

	// Register hooks
	if err := hookManager.RegisterHook(interfaces.BeforeQueryHook, queryLogHook); err != nil {
		return fmt.Errorf("failed to register query hook: %w", err)
	}

	if err := hookManager.RegisterHook(interfaces.AfterQueryHook, timingHook); err != nil {
		return fmt.Errorf("failed to register timing hook: %w", err)
	}

	if err := hookManager.RegisterHook(interfaces.OnErrorHook, errorHook); err != nil {
		return fmt.Errorf("failed to register error hook: %w", err)
	}

	// Register custom performance monitoring hook
	performanceHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Duration > 1*time.Second {
			fmt.Printf("🐌 [SLOW QUERY] Operation %s took %v - Query: %s\n",
				ctx.Operation, ctx.Duration, ctx.Query)
		}
		return &interfaces.HookResult{Continue: true}
	}

	if err := hookManager.RegisterCustomHook(interfaces.CustomHookBase+1, "performance_monitor", performanceHook); err != nil {
		return fmt.Errorf("failed to register performance hook: %w", err)
	}

	fmt.Println("✓ Custom hooks registered successfully")
	return nil
}

func demonstrateConcurrentOperations(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n--- Concurrent Operations Demo ---")

	// Setup test table
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS concurrent_test (
			id SERIAL PRIMARY KEY,
			value INTEGER,
			worker_id INTEGER,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	conn.Release()
	if err != nil {
		return err
	}

	// Run concurrent workers
	const numWorkers = 10
	const operationsPerWorker = 20

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
					// Insert with worker ID
					_, err := conn.Exec(ctx,
						"INSERT INTO concurrent_test (value, worker_id) VALUES ($1, $2)",
						j, workerID)
					return err
				})

				if err != nil {
					errors <- fmt.Errorf("worker %d operation %d failed: %w", workerID, j, err)
					return
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errors)

	// Check for errors
	var errorCount int
	for err := range errors {
		log.Printf("Concurrent operation error: %v", err)
		errorCount++
	}

	// Verify results
	conn, err = pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM concurrent_test")
	if err != nil {
		return err
	}

	fmt.Printf("✓ Concurrent operations completed: %d errors, %d total records\n", errorCount, count)
	return nil
}

func demonstrateAdvancedQuerying(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n--- Advanced Querying Demo ---")

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Prepared statement example
		err := conn.Prepare(ctx, "get_user_by_id", "SELECT id, name, email FROM users WHERE id = $1")
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer conn.Deallocate(ctx, "get_user_by_id")

		// Complex query with CTEs
		complexQuery := `
			WITH user_stats AS (
				SELECT 
					worker_id,
					COUNT(*) as operation_count,
					AVG(value) as avg_value
				FROM concurrent_test 
				GROUP BY worker_id
			)
			SELECT 
				worker_id,
				operation_count,
				ROUND(avg_value, 2) as avg_value
			FROM user_stats 
			ORDER BY operation_count DESC
			LIMIT 5
		`

		var stats []struct {
			WorkerID       int     `db:"worker_id"`
			OperationCount int64   `db:"operation_count"`
			AvgValue       float64 `db:"avg_value"`
		}

		err = conn.QueryAll(ctx, &stats, complexQuery)
		if err != nil {
			return fmt.Errorf("failed to execute complex query: %w", err)
		}

		fmt.Printf("✓ Top 5 workers by operation count:\n")
		for _, stat := range stats {
			fmt.Printf("  Worker %d: %d operations, avg value: %.2f\n",
				stat.WorkerID, stat.OperationCount, stat.AvgValue)
		}

		return nil
	})
}

func showDetailedStatistics(pool interfaces.IPool) {
	fmt.Println("\n--- Detailed Statistics ---")

	// Pool statistics
	poolStats := pool.Stats()
	fmt.Printf("Pool Statistics:\n")
	fmt.Printf("  Acquired Connections: %d\n", poolStats.AcquiredConns)
	fmt.Printf("  Total Connections: %d\n", poolStats.TotalConns)
	fmt.Printf("  Idle Connections: %d\n", poolStats.IdleConns)
	fmt.Printf("  Max Connections: %d\n", poolStats.MaxConns)
	fmt.Printf("  Acquire Count: %d\n", poolStats.AcquireCount)
	fmt.Printf("  Average Acquire Duration: %v\n", poolStats.AcquireDuration)

	// Buffer pool statistics (if available)
	if bufferPool := pool.GetBufferPool(); bufferPool != nil {
		memStats := bufferPool.Stats()
		fmt.Printf("Buffer Pool Statistics:\n")
		fmt.Printf("  Allocated Buffers: %d\n", memStats.AllocatedBuffers)
		fmt.Printf("  Pooled Buffers: %d\n", memStats.PooledBuffers)
		fmt.Printf("  Total Allocations: %d\n", memStats.TotalAllocations)
		fmt.Printf("  Total Deallocations: %d\n", memStats.TotalDeallocations)
	}

	// Safety monitor (if available)
	if safetyMonitor := pool.GetSafetyMonitor(); safetyMonitor != nil {
		fmt.Printf("Safety Monitor:\n")
		fmt.Printf("  Is Healthy: %t\n", safetyMonitor.IsHealthy())
		fmt.Printf("  Deadlocks: %d\n", len(safetyMonitor.CheckDeadlocks()))
		fmt.Printf("  Race Conditions: %d\n", len(safetyMonitor.CheckRaceConditions()))
		fmt.Printf("  Leaks: %d\n", len(safetyMonitor.CheckLeaks()))
	}
}

// Custom Audit Middleware
type AuditMiddleware struct {
	name     string
	priority int
	auditLog []AuditEntry
	mu       sync.RWMutex
}

type AuditEntry struct {
	Timestamp time.Time
	Operation string
	Query     string
	Duration  time.Duration
	Success   bool
	UserID    string // Could be extracted from context
}

func (am *AuditMiddleware) Name() string {
	return am.name
}

func (am *AuditMiddleware) Priority() int {
	return am.priority
}

func (am *AuditMiddleware) Before(ctx *interfaces.ExecutionContext) error {
	// Could extract user information from context here
	return nil
}

func (am *AuditMiddleware) After(ctx *interfaces.ExecutionContext) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	entry := AuditEntry{
		Timestamp: time.Now(),
		Operation: ctx.Operation,
		Query:     ctx.Query,
		Duration:  ctx.Duration,
		Success:   ctx.Error == nil,
		UserID:    "system", // Could be extracted from context
	}

	am.auditLog = append(am.auditLog, entry)
	fmt.Printf("📋 [AUDIT] %s operation %s in %v\n",
		entry.Operation,
		map[bool]string{true: "succeeded", false: "failed"}[entry.Success],
		entry.Duration)

	return nil
}

func (am *AuditMiddleware) OnError(ctx *interfaces.ExecutionContext) error {
	// Error handling is done in After() method
	return nil
}

// Custom Rate Limiting Middleware
type RateLimitMiddleware struct {
	name     string
	priority int
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	mu       sync.RWMutex
}

func (rlm *RateLimitMiddleware) Name() string {
	return rlm.name
}

func (rlm *RateLimitMiddleware) Priority() int {
	return rlm.priority
}

func (rlm *RateLimitMiddleware) Before(ctx *interfaces.ExecutionContext) error {
	// For demo purposes, use operation type as key
	// In real implementation, you might use user ID or IP address
	key := ctx.Operation

	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	now := time.Now()

	// Initialize if not exists
	if rlm.requests[key] == nil {
		rlm.requests[key] = make([]time.Time, 0)
	}

	// Clean old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range rlm.requests[key] {
		if now.Sub(reqTime) <= rlm.window {
			validRequests = append(validRequests, reqTime)
		}
	}
	rlm.requests[key] = validRequests

	// Check rate limit
	if len(rlm.requests[key]) >= rlm.limit {
		return fmt.Errorf("rate limit exceeded for operation %s: %d requests in %v",
			key, len(rlm.requests[key]), rlm.window)
	}

	// Record this request
	rlm.requests[key] = append(rlm.requests[key], now)

	return nil
}

func (rlm *RateLimitMiddleware) After(ctx *interfaces.ExecutionContext) error {
	return nil
}

func (rlm *RateLimitMiddleware) OnError(ctx *interfaces.ExecutionContext) error {
	return nil
}
