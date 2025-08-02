package hooks

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewRequestHook(t *testing.T) {
	hook := NewRequestHook("test-request")

	if hook.GetName() != "test-request" {
		t.Errorf("Expected hook name to be 'test-request', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.requestStartTimes == nil {
		t.Error("Expected request start times map to be initialized")
	}

	if hook.requestSizeTracker == nil {
		t.Error("Expected request size tracker map to be initialized")
	}
}

func TestRequestHook_SetMetricsEnabled(t *testing.T) {
	hook := NewRequestHook("test-request")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestRequestHook_OnRequest(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"

	// Test first request
	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check metrics
	if hook.GetRequestCount() != 1 {
		t.Errorf("Expected request count to be 1, got %d", hook.GetRequestCount())
	}

	if hook.GetActiveRequestCount() != 1 {
		t.Errorf("Expected active request count to be 1, got %d", hook.GetActiveRequestCount())
	}

	if hook.GetMaxActiveRequestCount() != 1 {
		t.Errorf("Expected max active request count to be 1, got %d", hook.GetMaxActiveRequestCount())
	}

	// Test second request
	req2 := "test request 2"
	err = hook.OnRequest(ctx, req2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if hook.GetRequestCount() != 2 {
		t.Errorf("Expected request count to be 2, got %d", hook.GetRequestCount())
	}

	if hook.GetActiveRequestCount() != 2 {
		t.Errorf("Expected active request count to be 2, got %d", hook.GetActiveRequestCount())
	}

	if hook.GetMaxActiveRequestCount() != 2 {
		t.Errorf("Expected max active request count to be 2, got %d", hook.GetMaxActiveRequestCount())
	}
}

func TestRequestHook_OnRequestDisabled(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	req := "test request"

	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetRequestCount() != 0 {
		t.Errorf("Expected request count to be 0 for disabled hook, got %d", hook.GetRequestCount())
	}
}

func TestRequestHook_OnRequestMetricsDisabled(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	req := "test request"

	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetRequestCount() != 0 {
		t.Errorf("Expected request count to be 0 when metrics disabled, got %d", hook.GetRequestCount())
	}
}

func TestRequestHook_OnResponse(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	// First, add a request
	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error during request, got %v", err)
	}

	// Then handle response
	err = hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}

	// Active requests should decrease
	if hook.GetActiveRequestCount() != 0 {
		t.Errorf("Expected active request count to be 0 after response, got %d", hook.GetActiveRequestCount())
	}

	// Total requests should remain
	if hook.GetRequestCount() != 1 {
		t.Errorf("Expected total request count to be 1, got %d", hook.GetRequestCount())
	}
}

func TestRequestHook_OnStart(t *testing.T) {
	hook := NewRequestHook("test-request")
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

func TestRequestHook_OnStop(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Add some requests first
	hook.OnRequest(ctx, "request1")
	hook.OnRequest(ctx, "request2")
	hook.OnResponse(ctx, "request1", "response1", time.Millisecond*100)

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

func TestRequestHook_OnStopNoMetrics(t *testing.T) {
	hook := NewRequestHook("test-request")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have stop message and metrics summary (even if zero)
	if len(testLogger.InfoMessages) < 2 {
		t.Errorf("Expected at least 2 info messages (stop + metrics), got %d", len(testLogger.InfoMessages))
	}
}

func TestRequestHook_GetTotalRequestSize(t *testing.T) {
	hook := NewRequestHook("test-request")

	ctx := context.Background()

	// Initially should be 0
	if hook.GetTotalRequestSize() != 0 {
		t.Errorf("Expected total request size to be 0 initially, got %d", hook.GetTotalRequestSize())
	}

	// Add some requests
	hook.OnRequest(ctx, "small")
	hook.OnRequest(ctx, "larger request")

	totalSize := hook.GetTotalRequestSize()
	if totalSize <= 0 {
		t.Errorf("Expected total request size to be positive, got %d", totalSize)
	}
}

func TestRequestHook_GetAverageRequestSize(t *testing.T) {
	hook := NewRequestHook("test-request")

	// Initially should be 0 (no requests)
	if hook.GetAverageRequestSize() != 0 {
		t.Errorf("Expected average request size to be 0 initially, got %d", hook.GetAverageRequestSize())
	}

	ctx := context.Background()

	// Add some requests
	hook.OnRequest(ctx, "test")
	hook.OnRequest(ctx, "test")

	avgSize := hook.GetAverageRequestSize()
	if avgSize <= 0 {
		t.Errorf("Expected average request size to be positive, got %d", avgSize)
	}
}

func TestRequestHook_ResetMetrics(t *testing.T) {
	hook := NewRequestHook("test-request")

	ctx := context.Background()

	// Add some requests
	hook.OnRequest(ctx, "request1")
	hook.OnRequest(ctx, "request2")

	// Verify metrics are non-zero
	if hook.GetRequestCount() == 0 {
		t.Error("Expected request count to be non-zero before reset")
	}

	// Reset metrics
	hook.ResetMetrics()

	// Verify all metrics are reset
	if hook.GetRequestCount() != 0 {
		t.Errorf("Expected request count to be 0 after reset, got %d", hook.GetRequestCount())
	}

	if hook.GetActiveRequestCount() != 0 {
		t.Errorf("Expected active request count to be 0 after reset, got %d", hook.GetActiveRequestCount())
	}

	if hook.GetMaxActiveRequestCount() != 0 {
		t.Errorf("Expected max active request count to be 0 after reset, got %d", hook.GetMaxActiveRequestCount())
	}

	if hook.GetTotalRequestSize() != 0 {
		t.Errorf("Expected total request size to be 0 after reset, got %d", hook.GetTotalRequestSize())
	}
}

func TestRequestHook_ConcurrentRequests(t *testing.T) {
	hook := NewRequestHook("test-request")

	ctx := context.Background()
	numRequests := 100
	var wg sync.WaitGroup

	// Simulate concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := fmt.Sprintf("request-%d", i)
			hook.OnRequest(ctx, req)

			// Simulate some processing time
			time.Sleep(time.Millisecond)

			hook.OnResponse(ctx, req, "response", time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Check final metrics
	if hook.GetRequestCount() != int64(numRequests) {
		t.Errorf("Expected request count to be %d, got %d", numRequests, hook.GetRequestCount())
	}

	if hook.GetActiveRequestCount() != 0 {
		t.Errorf("Expected active request count to be 0 after all completed, got %d", hook.GetActiveRequestCount())
	}

	if hook.GetMaxActiveRequestCount() <= 0 {
		t.Errorf("Expected max active request count to be positive, got %d", hook.GetMaxActiveRequestCount())
	}
}
