// Package main demonstrates the usage of generic hooks in HTTP server.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func main() {
	fmt.Println("🚀 Generic Hooks Example")
	fmt.Println("========================")

	// Create hook registry
	registry := hooks.NewDefaultHookRegistry()

	// Create and set hook observer for tracing
	observer := hooks.NewTracingObserver("example-service")
	registry.SetObserver(observer)

	// Create and register various hooks
	setupHooks(registry)

	// Simulate server lifecycle events
	simulateServerLifecycle(registry)

	// Display metrics
	displayMetrics(registry)

	// Demonstrate hook chaining
	demonstrateHookChaining()

	// Demonstrate filtering
	demonstrateFiltering()

	fmt.Println("\n✅ Generic Hooks example completed successfully!")
}

func setupHooks(registry *hooks.DefaultHookRegistry) {
	fmt.Println("\n📋 Setting up hooks...")

	// 1. Logging Hook
	loggingHook := hooks.NewLoggingHook(nil)
	err := registry.Register(loggingHook)
	if err != nil {
		log.Fatalf("Failed to register logging hook: %v", err)
	}
	fmt.Println("✅ Registered logging hook")

	// 2. Metrics Hook
	metricsHook := hooks.NewMetricsHook()
	err = registry.Register(metricsHook)
	if err != nil {
		log.Fatalf("Failed to register metrics hook: %v", err)
	}
	fmt.Println("✅ Registered metrics hook")

	// 3. Security Hook
	securityHook := hooks.NewSecurityHook()
	securityHook.SetAllowedOrigins([]string{"http://localhost:3000", "https://example.com"})
	securityHook.SetBlockedIPs([]string{"192.168.1.100"})
	err = registry.Register(securityHook)
	if err != nil {
		log.Fatalf("Failed to register security hook: %v", err)
	}
	fmt.Println("✅ Registered security hook")

	// 4. Cache Hook
	cacheHook := hooks.NewCacheHook(5 * time.Minute)
	err = registry.Register(cacheHook)
	if err != nil {
		log.Fatalf("Failed to register cache hook: %v", err)
	}
	fmt.Println("✅ Registered cache hook")

	// 5. Health Check Hook
	healthHook := hooks.NewHealthCheckHook()
	healthHook.AddHealthCheck("database", func() error {
		// Simulate database health check
		return nil
	})
	healthHook.AddHealthCheck("redis", func() error {
		// Simulate redis health check
		return nil
	})
	err = registry.Register(healthHook)
	if err != nil {
		log.Fatalf("Failed to register health check hook: %v", err)
	}
	fmt.Println("✅ Registered health check hook")

	// 6. Custom async hook
	asyncHook := createCustomAsyncHook()
	err = registry.Register(asyncHook)
	if err != nil {
		log.Fatalf("Failed to register custom async hook: %v", err)
	}
	fmt.Println("✅ Registered custom async hook")

	fmt.Printf("📊 Total registered hooks: %d\n", len(registry.ListHooks()))
}

func simulateServerLifecycle(registry *hooks.DefaultHookRegistry) {
	fmt.Println("\n🔄 Simulating server lifecycle...")

	// Server start
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventServerStart,
		ServerName: "example-server",
		Timestamp:  time.Now(),
	}
	err := registry.Execute(ctx)
	if err != nil {
		log.Printf("Error executing server start hooks: %v", err)
	}

	// Simulate HTTP requests
	simulateRequests(registry)

	// Health check
	ctx = &interfaces.HookContext{
		Event:      interfaces.HookEventHealthCheck,
		ServerName: "example-server",
		Timestamp:  time.Now(),
	}
	err = registry.Execute(ctx)
	if err != nil {
		log.Printf("Error executing health check hooks: %v", err)
	}

	// Server stop
	ctx = &interfaces.HookContext{
		Event:      interfaces.HookEventServerStop,
		ServerName: "example-server",
		Timestamp:  time.Now(),
	}
	err = registry.Execute(ctx)
	if err != nil {
		log.Printf("Error executing server stop hooks: %v", err)
	}
}

func simulateRequests(registry *hooks.DefaultHookRegistry) {
	fmt.Println("\n📡 Simulating HTTP requests...")

	requests := []struct {
		method string
		path   string
		origin string
		status int
		delay  time.Duration
	}{
		{"GET", "/api/users", "http://localhost:3000", 200, 50 * time.Millisecond},
		{"POST", "/api/users", "http://localhost:3000", 201, 100 * time.Millisecond},
		{"GET", "/api/posts", "https://example.com", 200, 30 * time.Millisecond},
		{"DELETE", "/api/users/1", "http://malicious.com", 403, 10 * time.Millisecond},
		{"GET", "/api/health", "", 200, 5 * time.Millisecond},
	}

	for i, req := range requests {
		fmt.Printf("Processing request %d: %s %s\n", i+1, req.method, req.path)

		// Create HTTP request
		httpReq, _ := http.NewRequest(req.method, req.path, nil)
		httpReq.RemoteAddr = "127.0.0.1:8080"
		if req.origin != "" {
			httpReq.Header.Set("Origin", req.origin)
		}

		// Request start
		ctx := &interfaces.HookContext{
			Event:         interfaces.HookEventRequestStart,
			ServerName:    "example-server",
			Timestamp:     time.Now(),
			Request:       httpReq,
			TraceID:       fmt.Sprintf("trace-%d", i+1),
			CorrelationID: fmt.Sprintf("corr-%d", i+1),
		}

		err := registry.Execute(ctx)
		if err != nil {
			log.Printf("Request blocked by security: %v", err)
			continue
		}

		// Simulate processing delay
		time.Sleep(req.delay)

		// Request end
		ctx.Event = interfaces.HookEventRequestEnd
		ctx.StatusCode = req.status
		ctx.Duration = req.delay

		err = registry.Execute(ctx)
		if err != nil {
			log.Printf("Error executing request end hooks: %v", err)
		}

		fmt.Printf("✅ Request %d completed with status %d\n", i+1, req.status)
	}
}

func displayMetrics(registry *hooks.DefaultHookRegistry) {
	fmt.Println("\n📈 Hook Execution Metrics:")
	fmt.Println("========================")

	metrics := registry.GetMetrics()
	fmt.Printf("Total Executions: %d\n", metrics.TotalExecutions)
	fmt.Printf("Successful Executions: %d\n", metrics.SuccessfulExecutions)
	fmt.Printf("Failed Executions: %d\n", metrics.FailedExecutions)
	fmt.Printf("Average Latency: %v\n", metrics.AverageLatency)
	fmt.Printf("Max Latency: %v\n", metrics.MaxLatency)
	fmt.Printf("Min Latency: %v\n", metrics.MinLatency)

	fmt.Println("\nExecutions by Event:")
	for event, count := range metrics.ExecutionsByEvent {
		fmt.Printf("  %s: %d\n", event, count)
	}

	fmt.Println("\nExecutions by Hook:")
	for hook, count := range metrics.ExecutionsByHook {
		fmt.Printf("  %s: %d\n", hook, count)
	}

	if len(metrics.ErrorsByHook) > 0 {
		fmt.Println("\nErrors by Hook:")
		for hook, count := range metrics.ErrorsByHook {
			fmt.Printf("  %s: %d\n", hook, count)
		}
	}
}

func demonstrateHookChaining() {
	fmt.Println("\n🔗 Demonstrating Hook Chaining:")
	fmt.Println("==============================")

	chain := hooks.NewHookChain()

	// Create simple hooks for demonstration
	hook1 := createSimpleHook("chain-hook-1", 1)
	hook2 := createSimpleHook("chain-hook-2", 2)
	hook3 := createSimpleHook("chain-hook-3", 3)

	// Add hooks to chain
	chain.Add(hook1).Add(hook2).Add(hook3)

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "chain-example",
		Timestamp:  time.Now(),
	}

	fmt.Println("Executing hook chain...")
	err := chain.Execute(ctx)
	if err != nil {
		log.Printf("Error executing hook chain: %v", err)
	}

	// Execute until condition
	fmt.Println("\nExecuting until condition (stop after 2 hooks)...")
	condition := func(ctx *interfaces.HookContext) bool {
		// Stop after 2 hooks (in real scenario, you might check response or context state)
		return ctx.Metadata != nil && len(ctx.Metadata) >= 2
	}

	ctx.Metadata = make(map[string]interface{})
	err = chain.ExecuteUntil(ctx, condition)
	if err != nil {
		log.Printf("Error executing hook chain until condition: %v", err)
	}
}

func demonstrateFiltering() {
	fmt.Println("\n🔍 Demonstrating Hook Filtering:")
	fmt.Println("===============================")

	// Create filtered hook for API endpoints only
	apiHook := hooks.NewFilteredBaseHook("api-only", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 10)

	// Set up path filter
	pathFilter := hooks.NewPathFilterBuilder().
		Include("/api/users", "/api/posts").
		Exclude("/api/internal").
		Build()
	apiHook.SetPathFilter(pathFilter)

	// Set up method filter
	methodFilter := hooks.NewMethodFilterBuilder().
		Allow("GET", "POST").
		Deny("DELETE").
		Build()
	apiHook.SetMethodFilter(methodFilter)

	// Test different requests
	testRequests := []struct {
		method   string
		path     string
		expected bool
	}{
		{"GET", "/api/users", true},
		{"POST", "/api/posts", true},
		{"DELETE", "/api/users", false},
		{"GET", "/api/internal", false},
		{"GET", "/static/css", false},
		{"PUT", "/api/users", false},
	}

	for _, test := range testRequests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		ctx := &interfaces.HookContext{
			Event:      interfaces.HookEventRequestStart,
			ServerName: "filter-test",
			Timestamp:  time.Now(),
			Request:    req,
		}

		shouldExecute := apiHook.ShouldExecute(ctx)
		result := "✅"
		if shouldExecute != test.expected {
			result = "❌"
		}

		fmt.Printf("%s %s %s -> Should execute: %v (expected: %v)\n",
			result, test.method, test.path, shouldExecute, test.expected)
	}
}

func createCustomAsyncHook() interfaces.AsyncHook {
	baseHook := hooks.NewAsyncBaseHook(
		"custom-async",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		80,
		5,
		1*time.Second,
	)

	return &customAsyncHook{
		AsyncBaseHook: baseHook,
	}
}

type customAsyncHook struct {
	*hooks.AsyncBaseHook
}

func (h *customAsyncHook) Execute(ctx *interfaces.HookContext) error {
	// Simulate some async work
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("📤 Custom async hook executed for %s\n", ctx.Event)
	return nil
}

func createSimpleHook(name string, priority int) interfaces.Hook {
	baseHook := hooks.NewBaseHook(name, []interfaces.HookEvent{interfaces.HookEventRequestStart}, priority)
	return &simpleHook{
		BaseHook: baseHook,
	}
}

type simpleHook struct {
	*hooks.BaseHook
}

func (h *simpleHook) Execute(ctx *interfaces.HookContext) error {
	fmt.Printf("🔧 Executing hook: %s (priority: %d)\n", h.Name(), h.Priority())

	// Add to metadata for demonstration
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]interface{})
	}
	ctx.Metadata[h.Name()] = time.Now()

	return nil
}

// Verify interface implementations
var _ interfaces.AsyncHook = (*customAsyncHook)(nil)
var _ interfaces.Hook = (*simpleHook)(nil)
