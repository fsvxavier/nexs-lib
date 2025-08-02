package hooks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewStartHook(t *testing.T) {
	hook := NewStartHook("test-start")

	if hook.GetName() != "test-start" {
		t.Errorf("Expected hook name to be 'test-start', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.startupEvents == nil {
		t.Error("Expected startup events slice to be initialized")
	}

	if hook.GetStartCount() != 0 {
		t.Errorf("Expected start count to be 0 initially, got %d", hook.GetStartCount())
	}
}

func TestStartHook_SetMetricsEnabled(t *testing.T) {
	hook := NewStartHook("test-start")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestStartHook_OnStart(t *testing.T) {
	hook := NewStartHook("test-start")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	addr := "localhost:8080"

	// Record start time before calling OnStart
	beforeStart := time.Now()

	err := hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Record time after calling OnStart
	afterStart := time.Now()

	// Check logging
	if len(testLogger.InfoMessages) < 1 {
		t.Errorf("Expected at least 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check basic properties
	if hook.GetServerAddr() != addr {
		t.Errorf("Expected server addr to be %s, got %s", addr, hook.GetServerAddr())
	}

	if hook.GetStartCount() != 1 {
		t.Errorf("Expected start count to be 1, got %d", hook.GetStartCount())
	}

	// Check start time is within reasonable bounds
	startTime := hook.GetStartTime()
	if startTime.Before(beforeStart) || startTime.After(afterStart) {
		t.Errorf("Expected start time to be between %v and %v, got %v", beforeStart, afterStart, startTime)
	}

	// Check uptime is reasonable
	uptime := hook.GetUptime()
	if uptime < 0 || uptime > time.Second {
		t.Errorf("Expected uptime to be reasonable, got %v", uptime)
	}

	// Check startup events
	events := hook.GetStartupEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 startup event, got %d", len(events))
	}

	if !events[0].Success {
		t.Error("Expected startup event to be successful")
	}

	if events[0].Address != addr {
		t.Errorf("Expected event address to be %s, got %s", addr, events[0].Address)
	}

	if events[0].Error != nil {
		t.Errorf("Expected no error in startup event, got %v", events[0].Error)
	}
}

func TestStartHook_OnStartDisabled(t *testing.T) {
	hook := NewStartHook("test-start")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	addr := "localhost:8080"

	err := hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that state was not updated
	if hook.GetStartCount() != 0 {
		t.Errorf("Expected start count to be 0 for disabled hook, got %d", hook.GetStartCount())
	}

	if hook.GetServerAddr() != "" {
		t.Errorf("Expected server addr to be empty for disabled hook, got %s", hook.GetServerAddr())
	}
}

func TestStartHook_OnStartMetricsDisabled(t *testing.T) {
	hook := NewStartHook("test-start")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	addr := "localhost:8080"

	err := hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) < 1 {
		t.Errorf("Expected at least 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Basic state should still be updated
	if hook.GetStartCount() != 1 {
		t.Errorf("Expected start count to be 1, got %d", hook.GetStartCount())
	}

	// But startup events should not be collected
	events := hook.GetStartupEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 startup events when metrics disabled, got %d", len(events))
	}
}

func TestStartHook_OnStop(t *testing.T) {
	hook := NewStartHook("test-start")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Start the server first
	hook.OnStart(ctx, "localhost:8080")

	// Wait a bit to get some uptime
	time.Sleep(time.Millisecond * 10)

	// Clear previous messages
	testLogger.InfoMessages = nil

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}

func TestStartHook_OnError(t *testing.T) {
	hook := NewStartHook("test-start")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	testErr := errors.New("startup failed")

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.ErrorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(testLogger.ErrorMessages))
	}

	// Check that failed startup event was recorded
	events := hook.GetStartupEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 startup event, got %d", len(events))
	}

	if events[0].Success {
		t.Error("Expected startup event to be unsuccessful")
	}

	if events[0].Error != testErr {
		t.Errorf("Expected event error to be %v, got %v", testErr, events[0].Error)
	}
}

func TestStartHook_MultipleStarts(t *testing.T) {
	hook := NewStartHook("test-start")

	ctx := context.Background()

	// Start multiple times
	addrs := []string{"localhost:8080", "localhost:8081", "localhost:8082"}
	for _, addr := range addrs {
		err := hook.OnStart(ctx, addr)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	if hook.GetStartCount() != int64(len(addrs)) {
		t.Errorf("Expected start count to be %d, got %d", len(addrs), hook.GetStartCount())
	}

	// Should track the last address
	if hook.GetServerAddr() != addrs[len(addrs)-1] {
		t.Errorf("Expected server addr to be %s, got %s", addrs[len(addrs)-1], hook.GetServerAddr())
	}

	events := hook.GetStartupEvents()
	if len(events) != len(addrs) {
		t.Errorf("Expected %d startup events, got %d", len(addrs), len(events))
	}
}

func TestStartHook_GetStartupEventsCounts(t *testing.T) {
	hook := NewStartHook("test-start")

	ctx := context.Background()

	// Initially should be 0
	if hook.GetSuccessfulStartCount() != 0 {
		t.Errorf("Expected 0 successful starts initially, got %d", hook.GetSuccessfulStartCount())
	}

	if hook.GetFailedStartCount() != 0 {
		t.Errorf("Expected 0 failed starts initially, got %d", hook.GetFailedStartCount())
	}

	// Add successful start
	hook.OnStart(ctx, "localhost:8080")

	if hook.GetSuccessfulStartCount() != 1 {
		t.Errorf("Expected 1 successful start, got %d", hook.GetSuccessfulStartCount())
	}

	if hook.GetFailedStartCount() != 0 {
		t.Errorf("Expected 0 failed starts, got %d", hook.GetFailedStartCount())
	}

	// Add failed start
	hook.OnError(ctx, errors.New("start failed"))

	if hook.GetSuccessfulStartCount() != 1 {
		t.Errorf("Expected 1 successful start, got %d", hook.GetSuccessfulStartCount())
	}

	if hook.GetFailedStartCount() != 1 {
		t.Errorf("Expected 1 failed start, got %d", hook.GetFailedStartCount())
	}
}

func TestStartHook_GetStartSuccessRate(t *testing.T) {
	hook := NewStartHook("test-start")

	// Initially should be 0 (no events)
	if hook.GetStartSuccessRate() != 0.0 {
		t.Errorf("Expected 0.0 success rate initially, got %.2f", hook.GetStartSuccessRate())
	}

	ctx := context.Background()

	// Add 3 successful starts and 1 failed start
	hook.OnStart(ctx, "localhost:8080")
	hook.OnStart(ctx, "localhost:8081")
	hook.OnStart(ctx, "localhost:8082")
	hook.OnError(ctx, errors.New("start failed"))

	expectedRate := 75.0 // 3 successful out of 4 total = 75%
	actualRate := hook.GetStartSuccessRate()
	if actualRate != expectedRate {
		t.Errorf("Expected success rate %.1f%%, got %.1f%%", expectedRate, actualRate)
	}
}

func TestStartHook_IsServerRunning(t *testing.T) {
	hook := NewStartHook("test-start")

	// Initially should not be running
	if hook.IsServerRunning() {
		t.Error("Expected server to not be running initially")
	}

	ctx := context.Background()

	// After start should be running
	hook.OnStart(ctx, "localhost:8080")
	if !hook.IsServerRunning() {
		t.Error("Expected server to be running after start")
	}

	// After reset should not be running
	hook.ResetStartTime()
	if hook.IsServerRunning() {
		t.Error("Expected server to not be running after reset")
	}
}

func TestStartHook_ClearMetrics(t *testing.T) {
	hook := NewStartHook("test-start")

	ctx := context.Background()

	// Add some events
	hook.OnStart(ctx, "localhost:8080")
	hook.OnError(ctx, errors.New("test error"))

	// Verify events exist
	if len(hook.GetStartupEvents()) == 0 {
		t.Error("Expected startup events to exist before clear")
	}

	if hook.GetStartCount() == 0 {
		t.Error("Expected start count to be non-zero before clear")
	}

	// Clear metrics
	hook.ClearMetrics()

	// Verify everything is cleared
	if len(hook.GetStartupEvents()) != 0 {
		t.Errorf("Expected 0 startup events after clear, got %d", len(hook.GetStartupEvents()))
	}

	if hook.GetStartCount() != 0 {
		t.Errorf("Expected start count to be 0 after clear, got %d", hook.GetStartCount())
	}
}

func TestStartHook_ResetStartTime(t *testing.T) {
	hook := NewStartHook("test-start")

	ctx := context.Background()

	// Start the server
	hook.OnStart(ctx, "localhost:8080")

	// Verify start time and addr are set
	if hook.GetStartTime().IsZero() {
		t.Error("Expected start time to be set")
	}

	if hook.GetServerAddr() == "" {
		t.Error("Expected server addr to be set")
	}

	// Reset start time
	hook.ResetStartTime()

	// Verify start time and addr are cleared
	if !hook.GetStartTime().IsZero() {
		t.Error("Expected start time to be zero after reset")
	}

	if hook.GetServerAddr() != "" {
		t.Error("Expected server addr to be empty after reset")
	}
}

func TestStartHook_GetStartupEventsCopy(t *testing.T) {
	hook := NewStartHook("test-start")

	ctx := context.Background()
	hook.OnStart(ctx, "localhost:8080")

	events := hook.GetStartupEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Modify the returned slice
	events[0].Success = false
	events[0].Address = "modified"

	// Verify original is not affected
	originalEvents := hook.GetStartupEvents()
	if !originalEvents[0].Success {
		t.Error("Expected original event to still be successful")
	}

	if originalEvents[0].Address != "localhost:8080" {
		t.Error("Expected original event address to be unchanged")
	}
}
