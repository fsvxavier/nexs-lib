package hooks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewStopHook(t *testing.T) {
	hook := NewStopHook("test-stop")

	if hook.GetName() != "test-stop" {
		t.Errorf("Expected hook name to be 'test-stop', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.shutdownEvents == nil {
		t.Error("Expected shutdown events slice to be initialized")
	}

	if hook.GetStopCount() != 0 {
		t.Errorf("Expected stop count to be 0 initially, got %d", hook.GetStopCount())
	}
}

func TestStopHook_SetMetricsEnabled(t *testing.T) {
	hook := NewStopHook("test-stop")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestStopHook_OnStart(t *testing.T) {
	hook := NewStopHook("test-stop")
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

	// Check that start time was recorded
	if hook.startTime.IsZero() {
		t.Error("Expected start time to be recorded")
	}
}

func TestStopHook_OnStop(t *testing.T) {
	hook := NewStopHook("test-stop")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Set a start time first
	startTime := time.Now().Add(-time.Hour) // 1 hour ago
	hook.SetStartTime(startTime)

	// Record stop time before calling OnStop
	beforeStop := time.Now()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Record time after calling OnStop
	afterStop := time.Now()

	// Check logging
	if len(testLogger.InfoMessages) < 1 {
		t.Errorf("Expected at least 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check basic properties
	if hook.GetStopCount() != 1 {
		t.Errorf("Expected stop count to be 1, got %d", hook.GetStopCount())
	}

	// Check stop time is within reasonable bounds
	stopTime := hook.GetStopTime()
	if stopTime.Before(beforeStop) || stopTime.After(afterStop) {
		t.Errorf("Expected stop time to be between %v and %v, got %v", beforeStop, afterStop, stopTime)
	}

	// Check uptime calculation
	lastUptime := hook.GetLastUptime()
	expectedUptime := beforeStop.Sub(startTime)
	if lastUptime < expectedUptime-time.Second || lastUptime > expectedUptime+time.Second {
		t.Errorf("Expected uptime around %v, got %v", expectedUptime, lastUptime)
	}

	// Check graceful stop count
	if hook.GetGracefulStopCount() != 1 {
		t.Errorf("Expected 1 graceful stop, got %d", hook.GetGracefulStopCount())
	}

	if hook.GetForcedStopCount() != 0 {
		t.Errorf("Expected 0 forced stops, got %d", hook.GetForcedStopCount())
	}

	// Check shutdown events
	events := hook.GetShutdownEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 shutdown event, got %d", len(events))
	}

	if !events[0].GracefulStop {
		t.Error("Expected shutdown event to be graceful")
	}

	if events[0].Error != nil {
		t.Errorf("Expected no error in shutdown event, got %v", events[0].Error)
	}
}

func TestStopHook_OnStopDisabled(t *testing.T) {
	hook := NewStopHook("test-stop")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that state was not updated
	if hook.GetStopCount() != 0 {
		t.Errorf("Expected stop count to be 0 for disabled hook, got %d", hook.GetStopCount())
	}
}

func TestStopHook_OnStopMetricsDisabled(t *testing.T) {
	hook := NewStopHook("test-stop")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) < 1 {
		t.Errorf("Expected at least 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Basic state should still be updated
	if hook.GetStopCount() != 1 {
		t.Errorf("Expected stop count to be 1, got %d", hook.GetStopCount())
	}

	// But shutdown events should not be collected
	events := hook.GetShutdownEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 shutdown events when metrics disabled, got %d", len(events))
	}
}

func TestStopHook_OnError(t *testing.T) {
	hook := NewStopHook("test-stop")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	testErr := errors.New("shutdown failed")

	// Set a start time first
	hook.SetStartTime(time.Now().Add(-time.Minute))

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.ErrorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(testLogger.ErrorMessages))
	}

	// Check that forced stop was recorded
	if hook.GetForcedStopCount() != 1 {
		t.Errorf("Expected 1 forced stop, got %d", hook.GetForcedStopCount())
	}

	// Check that failed shutdown event was recorded
	events := hook.GetShutdownEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 shutdown event, got %d", len(events))
	}

	if events[0].GracefulStop {
		t.Error("Expected shutdown event to be non-graceful")
	}

	if events[0].Error != testErr {
		t.Errorf("Expected event error to be %v, got %v", testErr, events[0].Error)
	}
}

func TestStopHook_MultipleStops(t *testing.T) {
	hook := NewStopHook("test-stop")

	ctx := context.Background()

	// Simulate multiple stop cycles
	for i := 0; i < 3; i++ {
		// Set start time
		hook.SetStartTime(time.Now().Add(-time.Hour))

		// Stop gracefully
		err := hook.OnStop(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Reset for next iteration
		hook.ResetStopTime()
	}

	if hook.GetStopCount() != 3 {
		t.Errorf("Expected stop count to be 3, got %d", hook.GetStopCount())
	}

	if hook.GetGracefulStopCount() != 3 {
		t.Errorf("Expected 3 graceful stops, got %d", hook.GetGracefulStopCount())
	}

	events := hook.GetShutdownEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 shutdown events, got %d", len(events))
	}
}

func TestStopHook_GetGracefulStopRate(t *testing.T) {
	hook := NewStopHook("test-stop")

	// Initially should be 0 (no events)
	if hook.GetGracefulStopRate() != 0.0 {
		t.Errorf("Expected 0.0 graceful stop rate initially, got %.2f", hook.GetGracefulStopRate())
	}

	ctx := context.Background()

	// Add 3 graceful stops and 1 forced stop
	hook.SetStartTime(time.Now().Add(-time.Hour))
	hook.OnStop(ctx)
	hook.OnStop(ctx)
	hook.OnStop(ctx)
	hook.OnError(ctx, errors.New("forced stop"))

	expectedRate := 75.0 // 3 graceful out of 4 total = 75%
	actualRate := hook.GetGracefulStopRate()
	if actualRate != expectedRate {
		t.Errorf("Expected graceful stop rate %.1f%%, got %.1f%%", expectedRate, actualRate)
	}
}

func TestStopHook_UptimeCalculations(t *testing.T) {
	hook := NewStopHook("test-stop")

	ctx := context.Background()

	// Add shutdown events with known uptimes
	uptimes := []time.Duration{
		time.Hour,
		time.Hour * 2,
		time.Hour * 3,
	}

	for _, uptime := range uptimes {
		startTime := time.Now().Add(-uptime)
		hook.SetStartTime(startTime)
		hook.OnStop(ctx)
	}

	// Test average uptime
	expectedAvg := time.Hour * 2 // (1 + 2 + 3) / 3 = 2 hours
	actualAvg := hook.GetAverageUptime()
	tolerance := time.Millisecond * 100 // 100ms tolerance
	if actualAvg < expectedAvg-tolerance || actualAvg > expectedAvg+tolerance {
		t.Errorf("Expected average uptime around %v, got %v", expectedAvg, actualAvg)
	}

	// Test total uptime
	expectedTotal := time.Hour * 6 // 1 + 2 + 3 = 6 hours
	actualTotal := hook.GetTotalUptime()
	if actualTotal < expectedTotal-tolerance || actualTotal > expectedTotal+tolerance {
		t.Errorf("Expected total uptime around %v, got %v", expectedTotal, actualTotal)
	}

	// Test max uptime
	expectedMax := time.Hour * 3
	actualMax := hook.GetMaxUptime()
	if actualMax < expectedMax-tolerance || actualMax > expectedMax+tolerance {
		t.Errorf("Expected max uptime around %v, got %v", expectedMax, actualMax)
	}

	// Test min uptime
	expectedMin := time.Hour
	actualMin := hook.GetMinUptime()
	if actualMin < expectedMin-tolerance || actualMin > expectedMin+tolerance {
		t.Errorf("Expected min uptime around %v, got %v", expectedMin, actualMin)
	}
}

func TestStopHook_IsServerStopped(t *testing.T) {
	hook := NewStopHook("test-stop")

	// Initially should not be stopped
	if hook.IsServerStopped() {
		t.Error("Expected server to not be stopped initially")
	}

	ctx := context.Background()

	// After stop should be stopped
	hook.OnStop(ctx)
	if !hook.IsServerStopped() {
		t.Error("Expected server to be stopped after stop")
	}

	// After reset should not be stopped
	hook.ResetStopTime()
	if hook.IsServerStopped() {
		t.Error("Expected server to not be stopped after reset")
	}
}

func TestStopHook_ClearMetrics(t *testing.T) {
	hook := NewStopHook("test-stop")

	ctx := context.Background()

	// Add some events
	hook.OnStop(ctx)
	hook.OnError(ctx, errors.New("test error"))

	// Verify events exist
	if len(hook.GetShutdownEvents()) == 0 {
		t.Error("Expected shutdown events to exist before clear")
	}

	if hook.GetStopCount() == 0 {
		t.Error("Expected stop count to be non-zero before clear")
	}

	// Clear metrics
	hook.ClearMetrics()

	// Verify everything is cleared
	if len(hook.GetShutdownEvents()) != 0 {
		t.Errorf("Expected 0 shutdown events after clear, got %d", len(hook.GetShutdownEvents()))
	}

	if hook.GetStopCount() != 0 {
		t.Errorf("Expected stop count to be 0 after clear, got %d", hook.GetStopCount())
	}

	if hook.GetGracefulStopCount() != 0 {
		t.Errorf("Expected graceful stop count to be 0 after clear, got %d", hook.GetGracefulStopCount())
	}

	if hook.GetForcedStopCount() != 0 {
		t.Errorf("Expected forced stop count to be 0 after clear, got %d", hook.GetForcedStopCount())
	}
}

func TestStopHook_GetShutdownEventsCopy(t *testing.T) {
	hook := NewStopHook("test-stop")

	ctx := context.Background()
	hook.OnStop(ctx)

	events := hook.GetShutdownEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Modify the returned slice
	events[0].GracefulStop = false
	events[0].Uptime = time.Hour * 999

	// Verify original is not affected
	originalEvents := hook.GetShutdownEvents()
	if !originalEvents[0].GracefulStop {
		t.Error("Expected original event to still be graceful")
	}

	if originalEvents[0].Uptime == time.Hour*999 {
		t.Error("Expected original event uptime to be unchanged")
	}
}

func TestStopHook_NoUptimeCalculations(t *testing.T) {
	hook := NewStopHook("test-stop")

	// Test uptime calculations with no events
	if hook.GetAverageUptime() != 0 {
		t.Errorf("Expected 0 average uptime with no events, got %v", hook.GetAverageUptime())
	}

	if hook.GetTotalUptime() != 0 {
		t.Errorf("Expected 0 total uptime with no events, got %v", hook.GetTotalUptime())
	}

	if hook.GetMaxUptime() != 0 {
		t.Errorf("Expected 0 max uptime with no events, got %v", hook.GetMaxUptime())
	}

	if hook.GetMinUptime() != 0 {
		t.Errorf("Expected 0 min uptime with no events, got %v", hook.GetMinUptime())
	}
}
