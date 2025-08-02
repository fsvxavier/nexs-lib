package hooks

import (
	"context"
	"testing"
	"time"
)

func TestNewResponseHook(t *testing.T) {
	hook := NewResponseHook("test-response")

	if hook.GetName() != "test-response" {
		t.Errorf("Expected hook name to be 'test-response', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.statusCodeCounts == nil {
		t.Error("Expected status code counts map to be initialized")
	}

	if hook.responseSizeTracker == nil {
		t.Error("Expected response size tracker map to be initialized")
	}
}

func TestResponseHook_SetMetricsEnabled(t *testing.T) {
	hook := NewResponseHook("test-response")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestResponseHook_OnResponse(t *testing.T) {
	hook := NewResponseHook("test-response")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	// Test first response
	err := hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check metrics
	if hook.GetResponseCount() != 1 {
		t.Errorf("Expected response count to be 1, got %d", hook.GetResponseCount())
	}

	if hook.GetAverageResponseTime() != duration {
		t.Errorf("Expected average response time to be %v, got %v", duration, hook.GetAverageResponseTime())
	}

	// Test second response
	resp2 := "test response 2"
	duration2 := time.Millisecond * 200
	err = hook.OnResponse(ctx, req, resp2, duration2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if hook.GetResponseCount() != 2 {
		t.Errorf("Expected response count to be 2, got %d", hook.GetResponseCount())
	}

	expectedAvg := time.Millisecond * 150 // (100 + 200) / 2
	if hook.GetAverageResponseTime() != expectedAvg {
		t.Errorf("Expected average response time to be %v, got %v", expectedAvg, hook.GetAverageResponseTime())
	}
}

func TestResponseHook_OnResponseDisabled(t *testing.T) {
	hook := NewResponseHook("test-response")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	err := hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetResponseCount() != 0 {
		t.Errorf("Expected response count to be 0 for disabled hook, got %d", hook.GetResponseCount())
	}
}

func TestResponseHook_OnResponseMetricsDisabled(t *testing.T) {
	hook := NewResponseHook("test-response")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	err := hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetResponseCount() != 0 {
		t.Errorf("Expected response count to be 0 when metrics disabled, got %d", hook.GetResponseCount())
	}
}

func TestResponseHook_OnStart(t *testing.T) {
	hook := NewResponseHook("test-response")
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

func TestResponseHook_OnStop(t *testing.T) {
	hook := NewResponseHook("test-response")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Add some responses first
	hook.OnResponse(ctx, "req1", "resp1", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "resp2", time.Millisecond*200)

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

func TestResponseHook_OnStopNoMetrics(t *testing.T) {
	hook := NewResponseHook("test-response")
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

func TestResponseHook_GetTotalResponseSize(t *testing.T) {
	hook := NewResponseHook("test-response")

	ctx := context.Background()

	// Initially should be 0
	if hook.GetTotalResponseSize() != 0 {
		t.Errorf("Expected total response size to be 0 initially, got %d", hook.GetTotalResponseSize())
	}

	// Add some responses
	hook.OnResponse(ctx, "req1", "small", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "larger response", time.Millisecond*200)

	totalSize := hook.GetTotalResponseSize()
	if totalSize <= 0 {
		t.Errorf("Expected total response size to be positive, got %d", totalSize)
	}
}

func TestResponseHook_GetAverageResponseSize(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Initially should be 0 (no responses)
	if hook.GetAverageResponseSize() != 0 {
		t.Errorf("Expected average response size to be 0 initially, got %d", hook.GetAverageResponseSize())
	}

	ctx := context.Background()

	// Add some responses
	hook.OnResponse(ctx, "req1", "test", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "test", time.Millisecond*200)

	avgSize := hook.GetAverageResponseSize()
	if avgSize <= 0 {
		t.Errorf("Expected average response size to be positive, got %d", avgSize)
	}
}

func TestResponseHook_GetAverageResponseTime(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Initially should be 0 (no responses)
	if hook.GetAverageResponseTime() != 0 {
		t.Errorf("Expected average response time to be 0 initially, got %v", hook.GetAverageResponseTime())
	}

	ctx := context.Background()

	// Add responses with known durations
	duration1 := time.Millisecond * 100
	duration2 := time.Millisecond * 200
	hook.OnResponse(ctx, "req1", "resp1", duration1)
	hook.OnResponse(ctx, "req2", "resp2", duration2)

	expectedAvg := time.Millisecond * 150 // (100 + 200) / 2
	avgTime := hook.GetAverageResponseTime()
	if avgTime != expectedAvg {
		t.Errorf("Expected average response time to be %v, got %v", expectedAvg, avgTime)
	}
}

func TestResponseHook_GetErrorRate(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Initially should be 0 (no responses)
	if hook.GetErrorRate() != 0.0 {
		t.Errorf("Expected error rate to be 0.0 initially, got %.2f", hook.GetErrorRate())
	}

	ctx := context.Background()

	// Add responses (all should be successful with default status 200)
	hook.OnResponse(ctx, "req1", "resp1", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "resp2", time.Millisecond*200)

	errorRate := hook.GetErrorRate()
	if errorRate != 0.0 {
		t.Errorf("Expected error rate to be 0.0 for successful responses, got %.2f", errorRate)
	}
}

func TestResponseHook_GetStatusCodeCounts(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Initially should be empty
	counts := hook.GetStatusCodeCounts()
	if len(counts) != 0 {
		t.Errorf("Expected empty status code counts initially, got %d entries", len(counts))
	}

	ctx := context.Background()

	// Add some responses
	hook.OnResponse(ctx, "req1", "resp1", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "resp2", time.Millisecond*200)

	counts = hook.GetStatusCodeCounts()
	if len(counts) == 0 {
		t.Error("Expected non-empty status code counts after responses")
	}

	// Verify it's a copy (modifications shouldn't affect original)
	for code := range counts {
		counts[code] = 999
		originalCounts := hook.GetStatusCodeCounts()
		if originalCounts[code] == 999 {
			t.Error("Expected returned counts to be a copy, not reference")
		}
	}
}

func TestResponseHook_ResetMetrics(t *testing.T) {
	hook := NewResponseHook("test-response")

	ctx := context.Background()

	// Add some responses
	hook.OnResponse(ctx, "req1", "resp1", time.Millisecond*100)
	hook.OnResponse(ctx, "req2", "resp2", time.Millisecond*200)

	// Verify metrics are non-zero
	if hook.GetResponseCount() == 0 {
		t.Error("Expected response count to be non-zero before reset")
	}

	// Reset metrics
	hook.ResetMetrics()

	// Verify all metrics are reset
	if hook.GetResponseCount() != 0 {
		t.Errorf("Expected response count to be 0 after reset, got %d", hook.GetResponseCount())
	}

	if hook.GetErrorResponseCount() != 0 {
		t.Errorf("Expected error response count to be 0 after reset, got %d", hook.GetErrorResponseCount())
	}

	if hook.GetTotalResponseSize() != 0 {
		t.Errorf("Expected total response size to be 0 after reset, got %d", hook.GetTotalResponseSize())
	}

	if hook.GetAverageResponseTime() != 0 {
		t.Errorf("Expected average response time to be 0 after reset, got %v", hook.GetAverageResponseTime())
	}

	if hook.GetErrorRate() != 0.0 {
		t.Errorf("Expected error rate to be 0.0 after reset, got %.2f", hook.GetErrorRate())
	}

	counts := hook.GetStatusCodeCounts()
	if len(counts) != 0 {
		t.Errorf("Expected empty status code counts after reset, got %d entries", len(counts))
	}
}

func TestResponseHook_extractResponseSize(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Test nil response
	size := hook.extractResponseSize(nil)
	if size != 0 {
		t.Errorf("Expected size 0 for nil response, got %d", size)
	}

	// Test string response
	resp := "test response"
	size = hook.extractResponseSize(resp)
	expectedSize := int64(len("test response"))
	if size != expectedSize {
		t.Errorf("Expected size %d for string response, got %d", expectedSize, size)
	}
}

func TestResponseHook_extractStatusCode(t *testing.T) {
	hook := NewResponseHook("test-response")

	// Test with any response (should return default 200)
	statusCode := hook.extractStatusCode("any response")
	if statusCode != 200 {
		t.Errorf("Expected default status code 200, got %d", statusCode)
	}
}
