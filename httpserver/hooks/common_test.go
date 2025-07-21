package hooks

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestLoggingHook(t *testing.T) {
	// Capture log output
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	hook := NewLoggingHook(nil)

	if hook.Name() != "logging" {
		t.Fatalf("Expected name 'logging', got '%s'", hook.Name())
	}

	if hook.Priority() != 100 {
		t.Fatalf("Expected priority 100, got %d", hook.Priority())
	}

	// Test server start event
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventServerStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(buf.String(), "Server test-server started") {
		t.Fatalf("Expected server start log, got: %s", buf.String())
	}

	// Reset buffer
	buf.Reset()

	// Test request start event
	req, _ := http.NewRequest("GET", "/api/users", nil)
	ctx = &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    req,
		TraceID:    "trace-123",
	}

	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Request started") {
		t.Fatalf("Expected request start log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "GET /api/users") {
		t.Fatalf("Expected method and path in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "trace-123") {
		t.Fatalf("Expected trace ID in log, got: %s", logOutput)
	}

	// Reset buffer
	buf.Reset()

	// Test request end event
	ctx = &interfaces.HookContext{
		Event:      interfaces.HookEventRequestEnd,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    req,
		StatusCode: 200,
		Duration:   50 * time.Millisecond,
	}

	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	logOutput = buf.String()
	if !strings.Contains(logOutput, "Request completed") {
		t.Fatalf("Expected request completed log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Status: 200") {
		t.Fatalf("Expected status code in log, got: %s", logOutput)
	}
}

func TestMetricsHook(t *testing.T) {
	hook := NewMetricsHook()

	if hook.Name() != "metrics" {
		t.Fatalf("Expected name 'metrics', got '%s'", hook.Name())
	}

	if hook.Priority() != 50 {
		t.Fatalf("Expected priority 50, got %d", hook.Priority())
	}

	// Test initial metrics
	metrics := hook.GetMetrics()
	if metrics["request_count"] != int64(0) {
		t.Fatalf("Expected initial request count 0, got %v", metrics["request_count"])
	}

	// Test request end event
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestEnd,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		StatusCode: 200,
		Duration:   100 * time.Millisecond,
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	metrics = hook.GetMetrics()
	if metrics["request_count"] != int64(1) {
		t.Fatalf("Expected request count 1, got %v", metrics["request_count"])
	}
	if metrics["total_duration"] != 100*time.Millisecond {
		t.Fatalf("Expected total duration 100ms, got %v", metrics["total_duration"])
	}
	if metrics["average_duration"] != 100*time.Millisecond {
		t.Fatalf("Expected average duration 100ms, got %v", metrics["average_duration"])
	}

	// Add another request
	ctx.Duration = 200 * time.Millisecond
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	metrics = hook.GetMetrics()
	if metrics["request_count"] != int64(2) {
		t.Fatalf("Expected request count 2, got %v", metrics["request_count"])
	}
	if metrics["average_duration"] != 150*time.Millisecond {
		t.Fatalf("Expected average duration 150ms, got %v", metrics["average_duration"])
	}

	// Test error event
	ctx = &interfaces.HookContext{
		Event:      interfaces.HookEventRequestError,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	metrics = hook.GetMetrics()
	if metrics["error_count"] != int64(1) {
		t.Fatalf("Expected error count 1, got %v", metrics["error_count"])
	}
	if metrics["error_rate"] != float64(0.5) {
		t.Fatalf("Expected error rate 0.5, got %v", metrics["error_rate"])
	}
}

func TestSecurityHook(t *testing.T) {
	hook := NewSecurityHook()

	if hook.Name() != "security" {
		t.Fatalf("Expected name 'security', got '%s'", hook.Name())
	}

	if hook.Priority() != 10 {
		t.Fatalf("Expected priority 10, got %d", hook.Priority())
	}

	// Test allowed request
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Origin", "http://localhost:3000")

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    req,
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error for allowed request, got %v", err)
	}

	// Test blocked IP
	hook.SetBlockedIPs([]string{"127.0.0.1:8080"})
	err = hook.Execute(ctx)
	if err == nil {
		t.Fatal("Expected error for blocked IP")
	}
	if !strings.Contains(err.Error(), "blocked IP address") {
		t.Fatalf("Expected blocked IP error, got %v", err)
	}

	// Reset blocked IPs
	hook.SetBlockedIPs([]string{})

	// Test CORS origin restriction
	hook.SetAllowedOrigins([]string{"http://example.com"})
	err = hook.Execute(ctx)
	if err == nil {
		t.Fatal("Expected error for disallowed origin")
	}
	if !strings.Contains(err.Error(), "origin not allowed") {
		t.Fatalf("Expected origin not allowed error, got %v", err)
	}

	// Test allowed origin
	req.Header.Set("Origin", "http://example.com")
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error for allowed origin, got %v", err)
	}

	// Test wildcard origin
	hook.SetAllowedOrigins([]string{"*"})
	req.Header.Set("Origin", "http://any-origin.com")
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error for wildcard origin, got %v", err)
	}

	// Test request without origin header
	req.Header.Del("Origin")
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error for request without origin, got %v", err)
	}
}

func TestCacheHook(t *testing.T) {
	hook := NewCacheHook(100 * time.Millisecond)

	if hook.Name() != "cache" {
		t.Fatalf("Expected name 'cache', got '%s'", hook.Name())
	}

	if hook.Priority() != 30 {
		t.Fatalf("Expected priority 30, got %d", hook.Priority())
	}

	// Test condition (should only execute for GET requests)
	req, _ := http.NewRequest("POST", "/api/users", nil)
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    req,
	}

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook not to execute for POST request")
	}

	// Test with GET request
	req.Method = "GET"
	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute for GET request")
	}

	// Test caching flow
	ctx.Event = interfaces.HookEventRequestStart
	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should not have cache hit yet
	if ctx.Metadata != nil && ctx.Metadata["cache_hit"] == true {
		t.Fatal("Expected no cache hit for first request")
	}

	// Simulate successful response
	ctx.Event = interfaces.HookEventRequestEnd
	ctx.StatusCode = 200
	ctx.Duration = 50 * time.Millisecond
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test cache hit
	ctx.Event = interfaces.HookEventRequestStart
	ctx.Metadata = nil
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.Metadata == nil || ctx.Metadata["cache_hit"] != true {
		t.Fatal("Expected cache hit for second request")
	}

	// Test cache expiration
	time.Sleep(150 * time.Millisecond) // Wait for cache to expire

	ctx.Event = interfaces.HookEventRequestStart
	ctx.Metadata = nil
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.Metadata != nil && ctx.Metadata["cache_hit"] == true {
		t.Fatal("Expected no cache hit after expiration")
	}

	// Test cache stats
	stats := hook.GetCacheStats()
	if stats["cache_ttl"] != 100*time.Millisecond {
		t.Fatalf("Expected cache TTL 100ms, got %v", stats["cache_ttl"])
	}
}

func TestHealthCheckHook(t *testing.T) {
	hook := NewHealthCheckHook()

	if hook.Name() != "healthcheck" {
		t.Fatalf("Expected name 'healthcheck', got '%s'", hook.Name())
	}

	if hook.Priority() != 20 {
		t.Fatalf("Expected priority 20, got %d", hook.Priority())
	}

	// Test initial health status (should be healthy with no checks)
	if !hook.IsHealthy() {
		t.Fatal("Expected hook to be healthy with no checks")
	}

	// Add health checks
	dbHealthy := true
	hook.AddHealthCheck("database", func() error {
		if !dbHealthy {
			return errors.New("database connection failed")
		}
		return nil
	})

	hook.AddHealthCheck("redis", func() error {
		return nil // Always healthy
	})

	// Execute health check
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventHealthCheck,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be healthy
	if !hook.IsHealthy() {
		t.Fatal("Expected hook to be healthy")
	}

	status := hook.GetHealthStatus()
	if !status["database"] {
		t.Fatal("Expected database to be healthy")
	}
	if !status["redis"] {
		t.Fatal("Expected redis to be healthy")
	}

	// Make database unhealthy
	dbHealthy = false
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be unhealthy now
	if hook.IsHealthy() {
		t.Fatal("Expected hook to be unhealthy")
	}

	status = hook.GetHealthStatus()
	if status["database"] {
		t.Fatal("Expected database to be unhealthy")
	}
	if !status["redis"] {
		t.Fatal("Expected redis to still be healthy")
	}
}

func TestLoggingHook_WithNilRequest(t *testing.T) {
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	hook := NewLoggingHook(nil)

	// Test with nil request
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    nil, // nil request
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error with nil request, got %v", err)
	}

	// Should handle nil request gracefully (log should be empty or minimal)
}

func TestSecurityHook_WithNilRequest(t *testing.T) {
	hook := NewSecurityHook()

	// Test with nil request
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    nil, // nil request
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error with nil request, got %v", err)
	}
}

func TestCacheHook_WithNilRequest(t *testing.T) {
	hook := NewCacheHook(100 * time.Millisecond)

	// Test with nil request
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    nil, // nil request
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error with nil request, got %v", err)
	}
}

func TestCacheHook_NonSuccessfulResponse(t *testing.T) {
	hook := NewCacheHook(100 * time.Millisecond)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestEnd,
		ServerName: "test-server",
		Timestamp:  time.Now(),
		Request:    req,
		StatusCode: 500, // Error status
		Duration:   50 * time.Millisecond,
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that error responses are not cached
	ctx.Event = interfaces.HookEventRequestStart
	ctx.Metadata = nil
	err = hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ctx.Metadata != nil && ctx.Metadata["cache_hit"] == true {
		t.Fatal("Expected no cache hit for error response")
	}
}
