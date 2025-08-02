package hooks

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewErrorHook(t *testing.T) {
	hook := NewErrorHook("test-error")

	if hook.GetName() != "test-error" {
		t.Errorf("Expected hook name to be 'test-error', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled by default")
	}

	if hook.errorByType == nil {
		t.Error("Expected error by type map to be initialized")
	}

	if hook.recentErrors == nil {
		t.Error("Expected recent errors slice to be initialized")
	}

	if hook.maxRecentErrors != 100 {
		t.Errorf("Expected max recent errors to be 100, got %d", hook.maxRecentErrors)
	}

	if hook.errorThreshold != 10 {
		t.Errorf("Expected error threshold to be 10, got %d", hook.errorThreshold)
	}
}

func TestErrorHook_SetMetricsEnabled(t *testing.T) {
	hook := NewErrorHook("test-error")

	hook.SetMetricsEnabled(false)
	if hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be disabled")
	}

	hook.SetMetricsEnabled(true)
	if !hook.IsMetricsEnabled() {
		t.Error("Expected metrics to be enabled")
	}
}

func TestErrorHook_SetMaxRecentErrors(t *testing.T) {
	hook := NewErrorHook("test-error")

	hook.SetMaxRecentErrors(50)
	if hook.maxRecentErrors != 50 {
		t.Errorf("Expected max recent errors to be 50, got %d", hook.maxRecentErrors)
	}
}

func TestErrorHook_SetErrorThreshold(t *testing.T) {
	hook := NewErrorHook("test-error")

	hook.SetErrorThreshold(5)
	if hook.errorThreshold != 5 {
		t.Errorf("Expected error threshold to be 5, got %d", hook.errorThreshold)
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityCritical, "CRITICAL"},
		{ErrorSeverity(999), "UNKNOWN"},
	}

	for _, test := range tests {
		actual := test.severity.String()
		if actual != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, actual)
		}
	}
}

func TestErrorHook_OnError(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Test first error
	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check logging (should be info level for generic error)
	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check metrics
	if hook.GetErrorCount() != 1 {
		t.Errorf("Expected error count to be 1, got %d", hook.GetErrorCount())
	}

	if hook.GetInfoErrorCount() != 1 {
		t.Errorf("Expected info error count to be 1, got %d", hook.GetInfoErrorCount())
	}

	// Check error time tracking
	if hook.GetFirstErrorTime().IsZero() {
		t.Error("Expected first error time to be set")
	}

	if hook.GetLastErrorTime().IsZero() {
		t.Error("Expected last error time to be set")
	}

	// Check recent errors
	recentErrors := hook.GetRecentErrors()
	if len(recentErrors) != 1 {
		t.Errorf("Expected 1 recent error, got %d", len(recentErrors))
	}

	if recentErrors[0].Error != testErr {
		t.Errorf("Expected recent error to be %v, got %v", testErr, recentErrors[0].Error)
	}
}

func TestErrorHook_OnErrorSeverityClassification(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	tests := []struct {
		error    error
		severity ErrorSeverity
		logType  string
	}{
		{errors.New("panic: runtime error"), SeverityCritical, "error"},
		{errors.New("network timeout"), SeverityWarning, "warn"},
		{errors.New("simple error"), SeverityInfo, "info"},
		{errors.New("database connection lost"), SeverityCritical, "error"},
		{errors.New("deprecated function used"), SeverityWarning, "warn"},
	}

	for _, test := range tests {
		// Clear previous messages
		testLogger.InfoMessages = nil
		testLogger.WarnMessages = nil
		testLogger.ErrorMessages = nil

		hook.OnError(ctx, test.error)

		// Verify classification
		switch test.logType {
		case "error":
			if len(testLogger.ErrorMessages) != 1 {
				t.Errorf("Expected 1 error message for %v, got %d", test.error, len(testLogger.ErrorMessages))
			}
		case "warn":
			if len(testLogger.WarnMessages) != 1 {
				t.Errorf("Expected 1 warn message for %v, got %d", test.error, len(testLogger.WarnMessages))
			}
		case "info":
			if len(testLogger.InfoMessages) != 1 {
				t.Errorf("Expected 1 info message for %v, got %d", test.error, len(testLogger.InfoMessages))
			}
		}
	}
}

func TestErrorHook_OnErrorDisabled(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetEnabled(false)

	ctx := context.Background()
	testErr := errors.New("test error")

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetErrorCount() != 0 {
		t.Errorf("Expected error count to be 0 for disabled hook, got %d", hook.GetErrorCount())
	}
}

func TestErrorHook_OnErrorMetricsDisabled(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetMetricsEnabled(false)

	ctx := context.Background()
	testErr := errors.New("test error")

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Check that metrics were not collected
	if hook.GetErrorCount() != 0 {
		t.Errorf("Expected error count to be 0 when metrics disabled, got %d", hook.GetErrorCount())
	}
}

func TestErrorHook_ErrorTypeClassification(t *testing.T) {
	hook := NewErrorHook("test-error")

	tests := []struct {
		error    error
		expected string
	}{
		{errors.New("network connection failed"), "network"},
		{errors.New("database query error"), "database"},
		{errors.New("unauthorized access"), "authorization"},
		{errors.New("request timeout"), "timeout"},
		{errors.New("invalid input"), "validation"},
		{errors.New("some other error"), "general"},
	}

	for _, test := range tests {
		errorType := hook.getErrorType(test.error)
		if errorType != test.expected {
			t.Errorf("Expected error type %s for %v, got %s", test.expected, test.error, errorType)
		}
	}
}

func TestErrorHook_ErrorThreshold(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)
	hook.SetErrorThreshold(3) // Low threshold for testing

	ctx := context.Background()

	// Add errors up to threshold
	for i := 0; i < 3; i++ {
		hook.OnError(ctx, errors.New("test error"))
	}

	// Clear previous messages to isolate threshold message
	testLogger.ErrorMessages = nil

	// This should trigger threshold alert
	hook.OnError(ctx, errors.New("threshold error"))

	// Should have threshold alert
	found := false
	for _, msg := range testLogger.ErrorMessages {
		if strings.Contains(msg, "THRESHOLD ALERT") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected threshold alert message in error logs")
	}
}

func TestErrorHook_RecentErrorsCircularBuffer(t *testing.T) {
	hook := NewErrorHook("test-error")
	hook.SetMaxRecentErrors(3) // Small buffer for testing

	ctx := context.Background()

	// Add more errors than buffer size
	errors := []error{
		errors.New("error 1"),
		errors.New("error 2"),
		errors.New("error 3"),
		errors.New("error 4"),
		errors.New("error 5"),
	}

	for _, err := range errors {
		hook.OnError(ctx, err)
	}

	recentErrors := hook.GetRecentErrors()
	if len(recentErrors) != 3 {
		t.Errorf("Expected 3 recent errors, got %d", len(recentErrors))
	}

	// Should contain the last 3 errors
	expectedErrors := []string{"error 3", "error 4", "error 5"}
	for i, expected := range expectedErrors {
		if recentErrors[i].Error.Error() != expected {
			t.Errorf("Expected recent error %d to be %s, got %s", i, expected, recentErrors[i].Error.Error())
		}
	}
}

func TestErrorHook_OnStop(t *testing.T) {
	hook := NewErrorHook("test-error")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	// Add some errors first
	hook.OnError(ctx, errors.New("test error 1"))
	hook.OnError(ctx, errors.New("panic: critical error"))

	// Clear test messages
	testLogger.InfoMessages = nil

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should have stop message plus error summary
	if len(testLogger.InfoMessages) < 2 {
		t.Errorf("Expected at least 2 info messages (stop + summary), got %d", len(testLogger.InfoMessages))
	}
}

func TestErrorHook_GetErrorByType(t *testing.T) {
	hook := NewErrorHook("test-error")

	ctx := context.Background()

	// Add errors of different types
	hook.OnError(ctx, errors.New("network error"))
	hook.OnError(ctx, errors.New("database error"))
	hook.OnError(ctx, errors.New("another network error"))

	errorByType := hook.GetErrorByType()

	if errorByType["network"] != 2 {
		t.Errorf("Expected 2 network errors, got %d", errorByType["network"])
	}

	if errorByType["database"] != 1 {
		t.Errorf("Expected 1 database error, got %d", errorByType["database"])
	}

	// Verify it's a copy (modifications shouldn't affect original)
	errorByType["network"] = 999
	originalErrorByType := hook.GetErrorByType()
	if originalErrorByType["network"] == 999 {
		t.Error("Expected returned error by type to be a copy, not reference")
	}
}

func TestErrorHook_GetCriticalErrorRate(t *testing.T) {
	hook := NewErrorHook("test-error")

	// Initially should be 0 (no errors)
	if hook.GetCriticalErrorRate() != 0.0 {
		t.Errorf("Expected 0.0 critical error rate initially, got %.2f", hook.GetCriticalErrorRate())
	}

	ctx := context.Background()

	// Add 2 critical and 2 non-critical errors
	hook.OnError(ctx, errors.New("panic: critical"))
	hook.OnError(ctx, errors.New("simple error"))
	hook.OnError(ctx, errors.New("fatal error"))
	hook.OnError(ctx, errors.New("another simple error"))

	expectedRate := 50.0 // 2 critical out of 4 total = 50%
	actualRate := hook.GetCriticalErrorRate()
	if actualRate != expectedRate {
		t.Errorf("Expected critical error rate %.1f%%, got %.1f%%", expectedRate, actualRate)
	}
}

func TestErrorHook_GetErrorFrequency(t *testing.T) {
	hook := NewErrorHook("test-error")

	// Initially should be 0 (no errors)
	if hook.GetErrorFrequency() != 0.0 {
		t.Errorf("Expected 0.0 error frequency initially, got %.2f", hook.GetErrorFrequency())
	}

	ctx := context.Background()

	// Manually set times for predictable testing
	now := time.Now()
	hook.firstErrorTime = now.Add(-time.Hour) // 1 hour ago
	hook.lastErrorTime = now

	// Add 4 errors to simulate frequency
	for i := 0; i < 4; i++ {
		hook.OnError(ctx, errors.New("test error"))
	}

	// Should be approximately 4 errors per hour
	frequency := hook.GetErrorFrequency()
	if frequency < 3.9 || frequency > 4.1 {
		t.Errorf("Expected error frequency around 4.0 errors/hour, got %.2f", frequency)
	}
}

func TestErrorHook_ResetMetrics(t *testing.T) {
	hook := NewErrorHook("test-error")

	ctx := context.Background()

	// Add some errors
	hook.OnError(ctx, errors.New("test error 1"))
	hook.OnError(ctx, errors.New("panic: critical error"))

	// Verify metrics are non-zero
	if hook.GetErrorCount() == 0 {
		t.Error("Expected error count to be non-zero before reset")
	}

	if len(hook.GetRecentErrors()) == 0 {
		t.Error("Expected recent errors to be non-empty before reset")
	}

	// Reset metrics
	hook.ResetMetrics()

	// Verify all metrics are reset
	if hook.GetErrorCount() != 0 {
		t.Errorf("Expected error count to be 0 after reset, got %d", hook.GetErrorCount())
	}

	if hook.GetCriticalErrorCount() != 0 {
		t.Errorf("Expected critical error count to be 0 after reset, got %d", hook.GetCriticalErrorCount())
	}

	if hook.GetWarningErrorCount() != 0 {
		t.Errorf("Expected warning error count to be 0 after reset, got %d", hook.GetWarningErrorCount())
	}

	if hook.GetInfoErrorCount() != 0 {
		t.Errorf("Expected info error count to be 0 after reset, got %d", hook.GetInfoErrorCount())
	}

	if len(hook.GetErrorByType()) != 0 {
		t.Errorf("Expected empty error by type after reset, got %d entries", len(hook.GetErrorByType()))
	}

	if len(hook.GetRecentErrors()) != 0 {
		t.Errorf("Expected empty recent errors after reset, got %d entries", len(hook.GetRecentErrors()))
	}

	if !hook.GetFirstErrorTime().IsZero() {
		t.Error("Expected first error time to be zero after reset")
	}

	if !hook.GetLastErrorTime().IsZero() {
		t.Error("Expected last error time to be zero after reset")
	}
}

func TestErrorHook_IsRecoverableError(t *testing.T) {
	hook := NewErrorHook("test-error")

	tests := []struct {
		error       error
		recoverable bool
	}{
		{errors.New("panic: runtime error"), false},
		{errors.New("fatal error"), false},
		{errors.New("out of memory"), false},
		{errors.New("network timeout"), true},
		{errors.New("simple error"), true},
		{nil, true},
	}

	for _, test := range tests {
		recoverable := hook.isRecoverableError(test.error)
		if recoverable != test.recoverable {
			t.Errorf("Expected error %v to be recoverable=%v, got %v", test.error, test.recoverable, recoverable)
		}
	}
}

func TestErrorHook_GetRecentErrorsCopy(t *testing.T) {
	hook := NewErrorHook("test-error")

	ctx := context.Background()
	testErr := errors.New("test error")
	hook.OnError(ctx, testErr)

	recentErrors := hook.GetRecentErrors()
	if len(recentErrors) != 1 {
		t.Errorf("Expected 1 recent error, got %d", len(recentErrors))
	}

	// Modify the returned slice
	recentErrors[0].Severity = SeverityCritical
	recentErrors[0].Recoverable = false

	// Verify original is not affected
	originalRecentErrors := hook.GetRecentErrors()
	if originalRecentErrors[0].Severity == SeverityCritical {
		t.Error("Expected original recent error severity to be unchanged")
	}

	if !originalRecentErrors[0].Recoverable {
		t.Error("Expected original recent error to still be recoverable")
	}
}
