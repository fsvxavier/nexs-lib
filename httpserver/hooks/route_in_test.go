package hooks

import (
	"context"
	"testing"
	"time"
)

func TestNewRouteInHook(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	if hook.GetName() != "test-route-in" {
		t.Errorf("Expected hook name to be 'test-route-in', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.routeMetrics == nil {
		t.Error("Expected route metrics map to be initialized")
	}
}

func TestRouteInHook_SetMetricsEnabled(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestRouteInHook_OnRouteEnter(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"

	// Test first entry
	err := hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check metrics
	metrics, exists := hook.GetRouteMetrics(method, path)
	if !exists {
		t.Error("Expected route metrics to exist")
	}

	if metrics.EntryCount != 1 {
		t.Errorf("Expected entry count to be 1, got %d", metrics.EntryCount)
	}

	// Test second entry
	err = hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	metrics, exists = hook.GetRouteMetrics(method, path)
	if !exists {
		t.Error("Expected route metrics to exist")
	}

	if metrics.EntryCount != 2 {
		t.Errorf("Expected entry count to be 2, got %d", metrics.EntryCount)
	}
}

func TestRouteInHook_OnRouteEnterDisabled(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"

	err := hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	_, exists := hook.GetRouteMetrics(method, path)
	if exists {
		t.Error("Expected no route metrics to exist for disabled hook")
	}
}

func TestRouteInHook_OnRouteEnterMetricsDisabled(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"

	err := hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	_, exists := hook.GetRouteMetrics(method, path)
	if exists {
		t.Error("Expected no route metrics to exist when metrics are disabled")
	}
}

func TestRouteInHook_OnStart(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
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

func TestRouteInHook_OnStop(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Add some metrics first
	hook.OnRouteEnter(ctx, "GET", "/api/users", nil)
	hook.OnRouteEnter(ctx, "POST", "/api/users", nil)

	// Clear test messages
	testLogger.InfoMessages = nil

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have stop message plus metrics summary
	if len(testLogger.InfoMessages) < 2 {
		t.Errorf("Expected at least 2 info messages (stop + metrics), got %d", len(testLogger.InfoMessages))
	}
}

func TestRouteInHook_OnStopNoMetrics(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should only have stop message, no metrics summary
	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}

func TestRouteInHook_OnError(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	testErr := &TestError{Message: "test error"}

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.ErrorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(testLogger.ErrorMessages))
	}
}

func TestRouteInHook_OnRequest(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"

	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestRouteInHook_OnResponse(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	err := hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestRouteInHook_OnRouteExit(t *testing.T) {
	hook := NewRouteInHook("test-route-in")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/api/users"
	req := "test request"
	duration := time.Millisecond * 150

	// First, enter the route to create metrics
	err := hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error during route enter, got %v", err)
	}

	// Now exit the route
	err = hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}

	// Check that duration was recorded in metrics
	metrics, exists := hook.GetRouteMetrics(method, path)
	if !exists {
		t.Error("Expected route metrics to exist")
	}

	if metrics.TotalDuration != duration {
		t.Errorf("Expected total duration to be %v, got %v", duration, metrics.TotalDuration)
	}

	if metrics.AvgDuration != duration {
		t.Errorf("Expected average duration to be %v, got %v", duration, metrics.AvgDuration)
	}
}

func TestRouteInHook_GetRouteMetrics(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	method := "GET"
	path := "/api/users"

	// Test non-existent metrics
	_, exists := hook.GetRouteMetrics(method, path)
	if exists {
		t.Error("Expected no route metrics to exist initially")
	}

	// Create metrics
	ctx := context.Background()
	hook.OnRouteEnter(ctx, method, path, nil)

	// Test existing metrics
	metrics, exists := hook.GetRouteMetrics(method, path)
	if !exists {
		t.Error("Expected route metrics to exist after route enter")
	}

	if metrics.EntryCount != 1 {
		t.Errorf("Expected entry count to be 1, got %d", metrics.EntryCount)
	}
}

func TestRouteInHook_GetAllRouteMetrics(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	// Test empty metrics
	metrics := hook.GetAllRouteMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics, got %d entries", len(metrics))
	}

	// Add some metrics
	ctx := context.Background()
	hook.OnRouteEnter(ctx, "GET", "/api/users", nil)
	hook.OnRouteEnter(ctx, "POST", "/api/users", nil)
	hook.OnRouteEnter(ctx, "GET", "/api/posts", nil)

	metrics = hook.GetAllRouteMetrics()
	if len(metrics) != 3 {
		t.Errorf("Expected 3 metric entries, got %d", len(metrics))
	}

	// Verify it's a copy (modifications shouldn't affect original)
	for key, metric := range metrics {
		metric.EntryCount = 999
		originalMetric, _ := hook.GetRouteMetrics("GET", "/api/users")
		if key == "GET /api/users" && originalMetric.EntryCount == 999 {
			t.Error("Expected returned metrics to be a copy, not reference")
		}
	}
}

func TestRouteInHook_ClearMetrics(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	// Add some metrics
	ctx := context.Background()
	hook.OnRouteEnter(ctx, "GET", "/api/users", nil)
	hook.OnRouteEnter(ctx, "POST", "/api/users", nil)

	if hook.GetMetricsCount() != 2 {
		t.Errorf("Expected 2 metrics before clear, got %d", hook.GetMetricsCount())
	}

	hook.ClearMetrics()

	if hook.GetMetricsCount() != 0 {
		t.Errorf("Expected 0 metrics after clear, got %d", hook.GetMetricsCount())
	}

	metrics := hook.GetAllRouteMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected empty metrics after clear, got %d entries", len(metrics))
	}
}

func TestRouteInHook_GetMetricsCount(t *testing.T) {
	hook := NewRouteInHook("test-route-in")

	if hook.GetMetricsCount() != 0 {
		t.Errorf("Expected 0 metrics initially, got %d", hook.GetMetricsCount())
	}

	ctx := context.Background()
	hook.OnRouteEnter(ctx, "GET", "/api/users", nil)
	hook.OnRouteEnter(ctx, "POST", "/api/users", nil)
	hook.OnRouteEnter(ctx, "GET", "/api/users", nil) // Same route, should not increase count

	if hook.GetMetricsCount() != 2 {
		t.Errorf("Expected 2 unique routes, got %d", hook.GetMetricsCount())
	}
}

// TestError is a simple error implementation for testing.
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}
