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

// Error categorization
type ErrorCategory int

const (
	ConnectionError ErrorCategory = iota
	SyntaxError
	ConstraintError
	TimeoutError
	UnknownError
)

// Custom metrics collector
type MetricsCollector struct {
	mu                 sync.RWMutex
	queryCount         int64
	totalDuration      time.Duration
	errorCount         int64
	slowQueries        int64
	slowQueryThreshold time.Duration
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		slowQueryThreshold: 100 * time.Millisecond,
	}
}

func (m *MetricsCollector) RecordQuery(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queryCount++
	m.totalDuration += duration

	if err != nil {
		m.errorCount++
	}

	if duration > m.slowQueryThreshold {
		m.slowQueries++
	}
}

func (m *MetricsCollector) GetStats() (int64, time.Duration, int64, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queryCount, m.totalDuration, m.errorCount, m.slowQueries
}

func main() {
	// Advanced PostgreSQL provider example with custom hooks and middlewares
	ctx := context.Background()

	// Create configuration with advanced features
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(10),
			postgresql.WithMinConns(2),
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

	// Example 1: Custom hooks demonstration
	if err := demonstrateCustomHooks(pool); err != nil {
		log.Printf("Custom hooks demo failed: %v", err)
	}

	// Example 2: Metrics and monitoring
	if err := demonstrateMetricsAndMonitoring(ctx, pool); err != nil {
		log.Printf("Metrics demo failed: %v", err)
	}

	// Example 3: Query performance analysis
	if err := demonstratePerformanceAnalysis(ctx, pool); err != nil {
		log.Printf("Performance analysis demo failed: %v", err)
	}

	// Example 4: Advanced error handling
	if err := demonstrateAdvancedErrorHandling(ctx, pool); err != nil {
		log.Printf("Error handling demo failed: %v", err)
	}

	// Example 5: Connection lifecycle hooks
	if err := demonstrateConnectionLifecycleHooks(ctx, pool); err != nil {
		log.Printf("Connection lifecycle demo failed: %v", err)
	}

	fmt.Println("Advanced example completed successfully!")
}

func demonstrateCustomHooks(pool interfaces.IPool) error {
	fmt.Println("\n=== Custom Hooks Demonstration ===")

	hookManager := pool.GetHookManager()
	if hookManager == nil {
		fmt.Println("Note: Hook manager not available in this implementation")
		return nil
	}

	// Create metrics collector
	metrics := NewMetricsCollector()

	// 1. Query logging hook
	queryLogHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Operation == "query" || ctx.Operation == "exec" {
			fmt.Printf("🪝 [QUERY] %s: %s\n", ctx.Operation, truncateQuery(ctx.Query, 60))
			if len(ctx.Args) > 0 {
				fmt.Printf("🪝 [ARGS] %v\n", ctx.Args)
			}
		}
		return &interfaces.HookResult{Continue: true}
	}

	// 2. Timing hook
	timingHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Duration > 0 {
			fmt.Printf("⏱️  [TIMING] %s took %v\n", ctx.Operation, ctx.Duration)
			metrics.RecordQuery(ctx.Duration, ctx.Error)
		}
		return &interfaces.HookResult{Continue: true}
	}

	// 3. Error logging hook
	errorHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Error != nil {
			fmt.Printf("❌ [ERROR] %s failed: %v\n", ctx.Operation, ctx.Error)
		}
		return &interfaces.HookResult{Continue: true}
	}

	// 4. Slow query detection hook
	slowQueryHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Duration > 50*time.Millisecond {
			fmt.Printf("🐌 [SLOW QUERY] %s took %v (threshold: 50ms)\n",
				ctx.Operation, ctx.Duration)
			fmt.Printf("🐌 [SLOW QUERY SQL] %s\n", truncateQuery(ctx.Query, 100))
		}
		return &interfaces.HookResult{Continue: true}
	}

	// 5. Security audit hook
	auditHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Log sensitive operations
		if containsSensitiveOperation(ctx.Query) {
			fmt.Printf("🔒 [AUDIT] Sensitive operation detected: %s\n",
				truncateQuery(ctx.Query, 80))
		}
		return &interfaces.HookResult{Continue: true}
	}

	// Register hooks
	hooks := []struct {
		hookType interfaces.HookType
		hook     interfaces.Hook
		name     string
	}{
		{interfaces.BeforeQueryHook, queryLogHook, "Query Logger"},
		{interfaces.AfterQueryHook, timingHook, "Timer"},
		{interfaces.OnErrorHook, errorHook, "Error Logger"},
		{interfaces.AfterQueryHook, slowQueryHook, "Slow Query Detector"},
		{interfaces.BeforeQueryHook, auditHook, "Security Auditor"},
	}

	for _, h := range hooks {
		if err := hookManager.RegisterHook(h.hookType, h.hook); err != nil {
			fmt.Printf("⚠️  Failed to register %s hook: %v\n", h.name, err)
		} else {
			fmt.Printf("✅ Registered %s hook\n", h.name)
		}
	}

	// Test hooks with some operations
	fmt.Println("\n📊 Testing hooks with sample operations...")

	ctx := context.Background()
	err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Test queries that will trigger various hooks
		queries := []string{
			"SELECT 1 as test",
			"SELECT COUNT(*) FROM information_schema.tables",
			"SELECT pg_sleep(0.1), 'slow query' as result", // This will be slow
			"SELECT * FROM non_existent_table",             // This will error
		}

		for i, query := range queries {
			fmt.Printf("\n🔍 Executing test query %d...\n", i+1)
			if i == 3 { // Skip the error query for this demo
				fmt.Printf("⏭️  Skipping error query for demo purposes\n")
				continue
			}

			row := conn.QueryRow(ctx, query)
			var result interface{}
			_ = row.Scan(&result) // Ignore errors for demo
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Note: Some operations would require actual database: %v\n", err)
	}

	// Show collected metrics
	queryCount, totalDuration, errorCount, slowQueries := metrics.GetStats()
	fmt.Printf("\n📈 Collected Metrics:\n")
	fmt.Printf("  - Total Queries: %d\n", queryCount)
	fmt.Printf("  - Total Duration: %v\n", totalDuration)
	fmt.Printf("  - Error Count: %d\n", errorCount)
	fmt.Printf("  - Slow Queries: %d\n", slowQueries)
	if queryCount > 0 {
		avgDuration := totalDuration / time.Duration(queryCount)
		fmt.Printf("  - Average Duration: %v\n", avgDuration)
	}

	return nil
}

func demonstrateMetricsAndMonitoring(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n=== Metrics and Monitoring Demonstration ===")

	// Custom monitoring hook
	var operationStats = struct {
		sync.RWMutex
		operations map[string]int64
		durations  map[string]time.Duration
	}{
		operations: make(map[string]int64),
		durations:  make(map[string]time.Duration),
	}

	hookManager := pool.GetHookManager()
	if hookManager != nil {
		// Metrics collection hook
		metricsHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
			operationStats.Lock()
			operationStats.operations[ctx.Operation]++
			operationStats.durations[ctx.Operation] += ctx.Duration
			operationStats.Unlock()
			return &interfaces.HookResult{Continue: true}
		}

		hookManager.RegisterHook(interfaces.AfterQueryHook, metricsHook)
		hookManager.RegisterHook(interfaces.AfterExecHook, metricsHook)
	}

	// Simulate various operations
	fmt.Println("🔄 Simulating various database operations...")

	err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		operations := []struct {
			name  string
			query string
		}{
			{"metadata_query", "SELECT version()"},
			{"count_query", "SELECT 1"},
			{"info_query", "SELECT current_database()"},
		}

		for _, op := range operations {
			fmt.Printf("  Executing %s...\n", op.name)
			row := conn.QueryRow(ctx, op.query)
			var result string
			_ = row.Scan(&result)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Note: Monitoring operations would require actual database: %v\n", err)
	}

	// Display collected metrics
	fmt.Println("\n📊 Operation Metrics:")
	operationStats.RLock()
	for operation, count := range operationStats.operations {
		duration := operationStats.durations[operation]
		avgDuration := time.Duration(0)
		if count > 0 {
			avgDuration = duration / time.Duration(count)
		}
		fmt.Printf("  %s: %d operations, avg duration: %v\n",
			operation, count, avgDuration)
	}
	operationStats.RUnlock()

	// Pool statistics
	fmt.Println("\n📈 Pool Statistics:")
	stats := pool.Stats()
	fmt.Printf("  - Total Connections: %d\n", stats.TotalConns)
	fmt.Printf("  - Acquired Connections: %d\n", stats.AcquiredConns)
	fmt.Printf("  - Idle Connections: %d\n", stats.IdleConns)
	fmt.Printf("  - Max Connections: %d\n", stats.MaxConns)

	return nil
}

func demonstratePerformanceAnalysis(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n=== Performance Analysis Demonstration ===")

	// Performance tracking
	type QueryPerformance struct {
		Query     string
		Duration  time.Duration
		Timestamp time.Time
	}

	var performanceLog []QueryPerformance
	var mu sync.Mutex

	hookManager := pool.GetHookManager()
	if hookManager != nil {
		// Performance analysis hook
		perfHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
			if ctx.Duration > 0 {
				mu.Lock()
				performanceLog = append(performanceLog, QueryPerformance{
					Query:     truncateQuery(ctx.Query, 50),
					Duration:  ctx.Duration,
					Timestamp: time.Now(),
				})
				mu.Unlock()
			}
			return &interfaces.HookResult{Continue: true}
		}

		hookManager.RegisterHook(interfaces.AfterQueryHook, perfHook)
	}

	// Execute test queries with different complexities
	fmt.Println("🔍 Executing queries with different performance characteristics...")

	err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		testQueries := []struct {
			name  string
			query string
		}{
			{"simple", "SELECT 1"},
			{"metadata", "SELECT schemaname, tablename FROM pg_tables LIMIT 5"},
			{"calculation", "SELECT generate_series(1,100) as num"},
		}

		for _, test := range testQueries {
			fmt.Printf("  📝 Executing %s query...\n", test.name)

			start := time.Now()
			row := conn.QueryRow(ctx, test.query)
			var result interface{}
			_ = row.Scan(&result)

			duration := time.Since(start)
			fmt.Printf("    ⏱️  Completed in %v\n", duration)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Note: Performance analysis would require actual database: %v\n", err)
	}

	// Analyze performance data
	fmt.Println("\n📊 Performance Analysis Results:")
	mu.Lock()
	if len(performanceLog) == 0 {
		fmt.Println("  No performance data collected (hooks may not be implemented)")
	} else {
		var totalDuration time.Duration
		var slowQueries int
		const slowThreshold = 10 * time.Millisecond

		for _, perf := range performanceLog {
			totalDuration += perf.Duration
			if perf.Duration > slowThreshold {
				slowQueries++
			}
			fmt.Printf("  - %s: %v\n", perf.Query, perf.Duration)
		}

		avgDuration := totalDuration / time.Duration(len(performanceLog))
		fmt.Printf("\n📈 Summary:\n")
		fmt.Printf("  - Total Queries: %d\n", len(performanceLog))
		fmt.Printf("  - Average Duration: %v\n", avgDuration)
		fmt.Printf("  - Slow Queries (>%v): %d\n", slowThreshold, slowQueries)
		fmt.Printf("  - Performance Score: %.1f%%\n",
			float64(len(performanceLog)-slowQueries)/float64(len(performanceLog))*100)
	}
	mu.Unlock()

	return nil
}

func demonstrateAdvancedErrorHandling(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n=== Advanced Error Handling Demonstration ===")

	var errorStats = struct {
		sync.RWMutex
		categories map[ErrorCategory]int
		errors     []string
	}{
		categories: make(map[ErrorCategory]int),
		errors:     make([]string, 0),
	}

	hookManager := pool.GetHookManager()
	if hookManager != nil {
		// Error categorization hook
		errorAnalysisHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
			if ctx.Error != nil {
				category := categorizeError(ctx.Error)

				errorStats.Lock()
				errorStats.categories[category]++
				errorStats.errors = append(errorStats.errors, ctx.Error.Error())
				errorStats.Unlock()

				fmt.Printf("🔍 [ERROR ANALYSIS] Category: %s, Error: %v\n",
					getCategoryName(category), ctx.Error)
			}
			return &interfaces.HookResult{Continue: true}
		}

		hookManager.RegisterHook(interfaces.OnErrorHook, errorAnalysisHook)
	}

	// Test different error scenarios
	fmt.Println("🔍 Testing various error scenarios...")

	err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		errorTests := []struct {
			name  string
			query string
		}{
			{"syntax_error", "SELEC 1"}, // Intentional syntax error
			{"missing_table", "SELECT * FROM non_existent_table"},
			{"invalid_function", "SELECT unknown_function()"},
		}

		for _, test := range errorTests {
			fmt.Printf("  🧪 Testing %s...\n", test.name)

			row := conn.QueryRow(ctx, test.query)
			var result interface{}
			err := row.Scan(&result)

			if err != nil {
				fmt.Printf("    ❌ Expected error occurred: %v\n", err)
			} else {
				fmt.Printf("    ✅ Unexpected success\n")
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Note: Error handling tests would require actual database: %v\n", err)
	}

	// Display error analysis
	fmt.Println("\n📊 Error Analysis Summary:")
	errorStats.RLock()
	if len(errorStats.categories) == 0 {
		fmt.Println("  No errors recorded (database operations may be mocked)")
	} else {
		for category, count := range errorStats.categories {
			fmt.Printf("  - %s: %d occurrences\n", getCategoryName(category), count)
		}
	}
	errorStats.RUnlock()

	return nil
}

func demonstrateConnectionLifecycleHooks(ctx context.Context, pool interfaces.IPool) error {
	fmt.Println("\n=== Connection Lifecycle Hooks Demonstration ===")

	var lifecycleEvents = struct {
		sync.RWMutex
		events []string
	}{
		events: make([]string, 0),
	}

	hookManager := pool.GetHookManager()
	if hookManager != nil {
		// Connection lifecycle hooks
		connectionHooks := map[interfaces.HookType]func(string){
			interfaces.BeforeConnectionHook: func(event string) {
				lifecycleEvents.Lock()
				lifecycleEvents.events = append(lifecycleEvents.events,
					fmt.Sprintf("🔗 [BEFORE CONNECTION] %s", event))
				lifecycleEvents.Unlock()
				fmt.Printf("🔗 [LIFECYCLE] Before connection: %s\n", event)
			},
			interfaces.AfterConnectionHook: func(event string) {
				lifecycleEvents.Lock()
				lifecycleEvents.events = append(lifecycleEvents.events,
					fmt.Sprintf("✅ [AFTER CONNECTION] %s", event))
				lifecycleEvents.Unlock()
				fmt.Printf("✅ [LIFECYCLE] After connection: %s\n", event)
			},
			interfaces.BeforeReleaseHook: func(event string) {
				lifecycleEvents.Lock()
				lifecycleEvents.events = append(lifecycleEvents.events,
					fmt.Sprintf("🔓 [BEFORE RELEASE] %s", event))
				lifecycleEvents.Unlock()
				fmt.Printf("🔓 [LIFECYCLE] Before release: %s\n", event)
			},
			interfaces.AfterReleaseHook: func(event string) {
				lifecycleEvents.Lock()
				lifecycleEvents.events = append(lifecycleEvents.events,
					fmt.Sprintf("🆓 [AFTER RELEASE] %s", event))
				lifecycleEvents.Unlock()
				fmt.Printf("🆓 [LIFECYCLE] After release: %s\n", event)
			},
		}

		for hookType, eventFunc := range connectionHooks {
			hook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
				eventFunc(fmt.Sprintf("Operation: %s", ctx.Operation))
				return &interfaces.HookResult{Continue: true}
			}

			if err := hookManager.RegisterHook(hookType, hook); err != nil {
				fmt.Printf("⚠️  Failed to register lifecycle hook: %v\n", err)
			}
		}
	}

	// Test connection lifecycle
	fmt.Println("🔄 Testing connection lifecycle...")

	for i := 0; i < 3; i++ {
		fmt.Printf("\n  🔄 Connection cycle %d:\n", i+1)

		err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Simple operation to trigger lifecycle events
			row := conn.QueryRow(ctx, "SELECT 1")
			var result int
			return row.Scan(&result)
		})

		if err != nil {
			fmt.Printf("    ❌ Connection cycle %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("    ✅ Connection cycle %d completed\n", i+1)
		}
	}

	// Display lifecycle events
	fmt.Println("\n📋 Connection Lifecycle Events:")
	lifecycleEvents.RLock()
	if len(lifecycleEvents.events) == 0 {
		fmt.Println("  No lifecycle events recorded (hooks may not be implemented)")
	} else {
		for i, event := range lifecycleEvents.events {
			fmt.Printf("  %d. %s\n", i+1, event)
		}
	}
	lifecycleEvents.RUnlock()

	return nil
}

// Helper functions

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen-3] + "..."
}

func containsSensitiveOperation(query string) bool {
	sensitiveKeywords := []string{"DROP", "DELETE", "UPDATE", "ALTER", "GRANT", "REVOKE"}
	upperQuery := fmt.Sprintf("%s", query) // Convert to uppercase
	for _, keyword := range sensitiveKeywords {
		if len(upperQuery) > len(keyword) &&
			upperQuery[:len(keyword)] == keyword {
			return true
		}
	}
	return false
}

func categorizeError(err error) ErrorCategory {
	errStr := err.Error()

	if len(errStr) > 10 && errStr[:6] == "syntax" {
		return SyntaxError
	}
	if len(errStr) > 15 && errStr[:10] == "connection" {
		return ConnectionError
	}
	if len(errStr) > 10 && errStr[:7] == "timeout" {
		return TimeoutError
	}
	if len(errStr) > 15 && errStr[:10] == "constraint" {
		return ConstraintError
	}

	return UnknownError
}

func getCategoryName(category ErrorCategory) string {
	switch category {
	case ConnectionError:
		return "Connection Error"
	case SyntaxError:
		return "Syntax Error"
	case ConstraintError:
		return "Constraint Error"
	case TimeoutError:
		return "Timeout Error"
	default:
		return "Unknown Error"
	}
}
