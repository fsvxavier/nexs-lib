package hooks

import (
	"context"
	"testing"
	"time"
)

func TestNewRouteOutHook(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	if hook.GetName() != "test-route-out" {
		t.Errorf("Expected hook name to be 'test-route-out', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.performanceTrack == nil {
		t.Error("Expected performance track map to be initialized")
	}

	if hook.GetSlowThreshold() != time.Second {
		t.Errorf("Expected default slow threshold to be 1 second, got %v", hook.GetSlowThreshold())
	}
}

func TestRouteOutHook_SetMetricsEnabled(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestRouteOutHook_SetSlowThreshold(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	newThreshold := time.Millisecond * 500
	hook.SetSlowThreshold(newThreshold)

	if hook.GetSlowThreshold() != newThreshold {
		t.Errorf("Expected slow threshold to be %v, got %v", newThreshold, hook.GetSlowThreshold())
	}
}

func TestRouteOutHook_OnRouteExit(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"
	duration := time.Millisecond * 150

	// Test normal request
	err := hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check metrics
	metrics, exists := hook.GetPerformanceMetrics(method, path)
	if !exists {
		t.Error("Expected performance metrics to exist")
	}

	if metrics.ExitCount != 1 {
		t.Errorf("Expected exit count to be 1, got %d", metrics.ExitCount)
	}

	if metrics.TotalDuration != duration {
		t.Errorf("Expected total duration to be %v, got %v", duration, metrics.TotalDuration)
	}

	if metrics.AvgDuration != duration {
		t.Errorf("Expected average duration to be %v, got %v", duration, metrics.AvgDuration)
	}

	if metrics.MinDuration != duration {
		t.Errorf("Expected min duration to be %v, got %v", duration, metrics.MinDuration)
	}

	if metrics.MaxDuration != duration {
		t.Errorf("Expected max duration to be %v, got %v", duration, metrics.MaxDuration)
	}

	if metrics.SlowRequests != 0 {
		t.Errorf("Expected 0 slow requests, got %d", metrics.SlowRequests)
	}
}

func TestRouteOutHook_OnRouteExitSlowRequest(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetSlowThreshold(time.Millisecond * 100) // Set low threshold

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"
	duration := time.Millisecond * 200 // Above threshold

	err := hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should log as warning for slow request
	if len(testLogger.WarnMessages) != 1 {
		t.Errorf("Expected 1 warn message for slow request, got %d", len(testLogger.WarnMessages))
	}

	// Check metrics show slow request
	metrics, exists := hook.GetPerformanceMetrics(method, path)
	if !exists {
		t.Error("Expected performance metrics to exist")
	}

	if metrics.SlowRequests != 1 {
		t.Errorf("Expected 1 slow request, got %d", metrics.SlowRequests)
	}
}

func TestRouteOutHook_OnRouteExitMultipleRequests(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"

	durations := []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 150,
	}

	// Process multiple requests
	for _, duration := range durations {
		err := hook.OnRouteExit(ctx, method, path, req, duration)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Check aggregated metrics
	metrics, exists := hook.GetPerformanceMetrics(method, path)
	if !exists {
		t.Error("Expected performance metrics to exist")
	}

	if metrics.ExitCount != 3 {
		t.Errorf("Expected exit count to be 3, got %d", metrics.ExitCount)
	}

	expectedTotal := time.Millisecond * 450
	if metrics.TotalDuration != expectedTotal {
		t.Errorf("Expected total duration to be %v, got %v", expectedTotal, metrics.TotalDuration)
	}

	expectedAvg := time.Millisecond * 150
	if metrics.AvgDuration != expectedAvg {
		t.Errorf("Expected average duration to be %v, got %v", expectedAvg, metrics.AvgDuration)
	}

	expectedMin := time.Millisecond * 100
	if metrics.MinDuration != expectedMin {
		t.Errorf("Expected min duration to be %v, got %v", expectedMin, metrics.MinDuration)
	}

	expectedMax := time.Millisecond * 200
	if metrics.MaxDuration != expectedMax {
		t.Errorf("Expected max duration to be %v, got %v", expectedMax, metrics.MaxDuration)
	}
}

func TestRouteOutHook_OnRouteExitDisabled(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"
	duration := time.Millisecond * 150

	err := hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	_, exists := hook.GetPerformanceMetrics(method, path)
	if exists {
		t.Error("Expected no performance metrics to exist for disabled hook")
	}
}

func TestRouteOutHook_OnRouteExitMetricsDisabled(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"
	duration := time.Millisecond * 150

	err := hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	_, exists := hook.GetPerformanceMetrics(method, path)
	if exists {
		t.Error("Expected no performance metrics to exist when metrics are disabled")
	}
}

func TestRouteOutHook_OnStart(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	addr := "localhost:8080"

	err := hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}

func TestRouteOutHook_OnStop(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Add some metrics first
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100)
	hook.OnRouteExit(ctx, "POST", "/api/users", nil, time.Millisecond*200)

	// Clear test messages
	testLogger.InfoMessages = nil

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have stop message plus performance summary
	if len(testLogger.InfoMessages) < 2 {
		t.Errorf("Expected at least 2 info messages (stop + performance), got %d", len(testLogger.InfoMessages))
	}
}

func TestRouteOutHook_OnStopNoMetrics(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should only have stop message, no performance summary
	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}

func TestRouteOutHook_GetAllPerformanceMetrics(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	// Test empty metrics
	metrics := hook.GetAllPerformanceMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics, got %d entries", len(metrics))
	}

	// Add some metrics
	ctx := context.Background()
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100)
	hook.OnRouteExit(ctx, "POST", "/api/users", nil, time.Millisecond*200)
	hook.OnRouteExit(ctx, "GET", "/api/posts", nil, time.Millisecond*150)

	metrics = hook.GetAllPerformanceMetrics()
	if len(metrics) != 3 {
		t.Errorf("Expected 3 metric entries, got %d", len(metrics))
	}

	// Verify it's a copy (modifications shouldn't affect original)
	for key, metric := range metrics {
		metric.ExitCount = 999
		originalMetric, _ := hook.GetPerformanceMetrics("GET", "/api/users")
		if key == "GET /api/users" && originalMetric.ExitCount == 999 {
			t.Error("Expected returned metrics to be a copy, not reference")
		}
	}
}

func TestRouteOutHook_ClearMetrics(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	// Add some metrics
	ctx := context.Background()
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100)
	hook.OnRouteExit(ctx, "POST", "/api/users", nil, time.Millisecond*200)

	if hook.GetMetricsCount() != 2 {
		t.Errorf("Expected 2 metrics before clear, got %d", hook.GetMetricsCount())
	}

	hook.ClearMetrics()

	if hook.GetMetricsCount() != 0 {
		t.Errorf("Expected 0 metrics after clear, got %d", hook.GetMetricsCount())
	}

	metrics := hook.GetAllPerformanceMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics after clear, got %d entries", len(metrics))
	}
}

func TestRouteOutHook_GetSlowRequestsCount(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	hook.SetSlowThreshold(time.Millisecond * 150)

	ctx := context.Background()

	// Add mix of slow and fast requests
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100)  // Fast
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*200)  // Slow
	hook.OnRouteExit(ctx, "POST", "/api/users", nil, time.Millisecond*300) // Slow

	slowCount := hook.GetSlowRequestsCount()
	if slowCount != 2 {
		t.Errorf("Expected 2 slow requests, got %d", slowCount)
	}
}

func TestRouteOutHook_GetTotalRequestsCount(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")

	ctx := context.Background()
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100)
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*200)
	hook.OnRouteExit(ctx, "POST", "/api/users", nil, time.Millisecond*150)

	totalCount := hook.GetTotalRequestsCount()
	if totalCount != 3 {
		t.Errorf("Expected 3 total requests, got %d", totalCount)
	}
}

func TestRouteOutHook_GetSlowRequestPercentage(t *testing.T) {
	hook := NewRouteOutHook("test-route-out")
	hook.SetSlowThreshold(time.Millisecond * 150)

	// Test with no requests
	percentage := hook.GetSlowRequestPercentage()
	if percentage != 0.0 {
		t.Errorf("Expected 0%% for no requests, got %.1f%%", percentage)
	}

	ctx := context.Background()

	// Add requests: 1 fast, 1 slow
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*100) // Fast
	hook.OnRouteExit(ctx, "GET", "/api/users", nil, time.Millisecond*200) // Slow

	percentage = hook.GetSlowRequestPercentage()
	expectedPercentage := 50.0
	if percentage != expectedPercentage {
		t.Errorf("Expected %.1f%% slow requests, got %.1f%%", expectedPercentage, percentage)
	}
}
