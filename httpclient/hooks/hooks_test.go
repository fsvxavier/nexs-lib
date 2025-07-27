package hooks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Mock hook for testing
type testHook struct {
	name           string
	beforeCalls    []string
	afterCalls     []string
	errorCalls     []string
	beforeError    error
	afterError     error
	errorHookError error
}

func (h *testHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	h.beforeCalls = append(h.beforeCalls, h.name)
	return h.beforeError
}

func (h *testHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	h.afterCalls = append(h.afterCalls, h.name)
	return h.afterError
}

func (h *testHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	h.errorCalls = append(h.errorCalls, h.name)
	return h.errorHookError
}

func TestManager_Add(t *testing.T) {
	manager := NewManager()
	hook := &testHook{name: "test"}

	manager.Add(hook)

	if manager.Count() != 1 {
		t.Errorf("Expected 1 hook, got %d", manager.Count())
	}
}

func TestManager_Remove(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1"}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)
	manager.Remove(hook1)

	if manager.Count() != 1 {
		t.Errorf("Expected 1 hook after removal, got %d", manager.Count())
	}
}

func TestManager_Clear(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1"}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)
	manager.Clear()

	if manager.Count() != 0 {
		t.Errorf("Expected 0 hooks after clear, got %d", manager.Count())
	}
}

func TestManager_ExecuteBeforeRequest(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1"}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	err := manager.ExecuteBeforeRequest(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(hook1.beforeCalls) != 1 || hook1.beforeCalls[0] != "hook1" {
		t.Errorf("Hook1 before not called correctly: %v", hook1.beforeCalls)
	}

	if len(hook2.beforeCalls) != 1 || hook2.beforeCalls[0] != "hook2" {
		t.Errorf("Hook2 before not called correctly: %v", hook2.beforeCalls)
	}
}

func TestManager_ExecuteBeforeRequest_WithError(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1", beforeError: errors.New("hook1 error")}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	err := manager.ExecuteBeforeRequest(ctx, req)

	if err == nil {
		t.Error("Expected error from hook1")
	}

	if err.Error() != "hook1 error" {
		t.Errorf("Expected 'hook1 error', got %v", err)
	}

	// Hook2 should not be called due to early return
	if len(hook2.beforeCalls) != 0 {
		t.Errorf("Hook2 should not be called when hook1 fails: %v", hook2.beforeCalls)
	}
}

func TestManager_ExecuteAfterResponse(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1"}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	resp := &interfaces.Response{StatusCode: 200}
	ctx := context.Background()

	err := manager.ExecuteAfterResponse(ctx, req, resp)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(hook1.afterCalls) != 1 || hook1.afterCalls[0] != "hook1" {
		t.Errorf("Hook1 after not called correctly: %v", hook1.afterCalls)
	}

	if len(hook2.afterCalls) != 1 || hook2.afterCalls[0] != "hook2" {
		t.Errorf("Hook2 after not called correctly: %v", hook2.afterCalls)
	}
}

func TestManager_ExecuteOnError(t *testing.T) {
	manager := NewManager()
	hook1 := &testHook{name: "hook1"}
	hook2 := &testHook{name: "hook2"}

	manager.Add(hook1).Add(hook2)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	originalErr := errors.New("request error")
	ctx := context.Background()

	err := manager.ExecuteOnError(ctx, req, originalErr)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(hook1.errorCalls) != 1 || hook1.errorCalls[0] != "hook1" {
		t.Errorf("Hook1 error not called correctly: %v", hook1.errorCalls)
	}

	if len(hook2.errorCalls) != 1 || hook2.errorCalls[0] != "hook2" {
		t.Errorf("Hook2 error not called correctly: %v", hook2.errorCalls)
	}
}

func TestTimingHook(t *testing.T) {
	var capturedMethod, capturedURL string
	var capturedDuration int64

	callback := func(method, url string, duration int64) {
		capturedMethod = method
		capturedURL = url
		capturedDuration = duration
	}

	hook := NewTimingHook(callback)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	resp := &interfaces.Response{
		StatusCode: 200,
		Latency:    100 * time.Millisecond,
	}
	ctx := context.Background()

	err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	err = hook.AfterResponse(ctx, req, resp)
	if err != nil {
		t.Fatalf("AfterResponse failed: %v", err)
	}

	if capturedMethod != "GET" {
		t.Errorf("Expected method GET, got %s", capturedMethod)
	}

	if capturedURL != "http://test.com" {
		t.Errorf("Expected URL http://test.com, got %s", capturedURL)
	}

	expectedDuration := (100 * time.Millisecond).Nanoseconds()
	if capturedDuration != expectedDuration {
		t.Errorf("Expected duration %d, got %d", expectedDuration, capturedDuration)
	}
}

func TestLoggingHook(t *testing.T) {
	logs := []struct {
		level  string
		format string
		args   []interface{}
	}{}

	logger := func(level string, format string, args ...interface{}) {
		logs = append(logs, struct {
			level  string
			format string
			args   []interface{}
		}{level, format, args})
	}

	hook := NewLoggingHook(logger)

	req := &interfaces.Request{Method: "POST", URL: "http://api.test.com"}
	resp := &interfaces.Response{
		StatusCode: 201,
		Latency:    50 * time.Millisecond,
	}
	ctx := context.Background()

	err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	err = hook.AfterResponse(ctx, req, resp)
	if err != nil {
		t.Fatalf("AfterResponse failed: %v", err)
	}

	testErr := errors.New("test error")
	err = hook.OnError(ctx, req, testErr)
	if err != nil {
		t.Fatalf("OnError failed: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}

	if logs[0].level != "INFO" {
		t.Errorf("Expected first log level INFO, got %s", logs[0].level)
	}

	if logs[2].level != "ERROR" {
		t.Errorf("Expected third log level ERROR, got %s", logs[2].level)
	}
}

func TestValidationHook(t *testing.T) {
	validator := func(req *interfaces.Request) error {
		if req.Method == "" {
			return errors.New("method is required")
		}
		return nil
	}

	hook := NewValidationHook(validator)

	// Test valid request
	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error for valid request, got %v", err)
	}

	// Test invalid request
	invalidReq := &interfaces.Request{URL: "http://test.com"}
	err = hook.BeforeRequest(ctx, invalidReq)
	if err == nil {
		t.Error("Expected error for invalid request")
	}

	if err.Error() != "method is required" {
		t.Errorf("Expected 'method is required', got %v", err)
	}
}

func TestCircuitBreakerHook(t *testing.T) {
	hook := NewCircuitBreakerHook(2) // Open after 2 failures

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	// First request should pass
	err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("First request should pass, got %v", err)
	}

	// Simulate first failure
	errorResp := &interfaces.Response{StatusCode: 500, IsError: true}
	err = hook.AfterResponse(ctx, req, errorResp)
	if err != nil {
		t.Fatalf("AfterResponse should not error, got %v", err)
	}

	if hook.IsOpen() {
		t.Error("Circuit breaker should not be open after 1 failure")
	}

	// Simulate second failure
	err = hook.AfterResponse(ctx, req, errorResp)
	if err != nil {
		t.Fatalf("AfterResponse should not error, got %v", err)
	}

	if !hook.IsOpen() {
		t.Error("Circuit breaker should be open after 2 failures")
	}

	// Next request should be blocked
	err = hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Error("Request should be blocked when circuit breaker is open")
	}

	// Reset circuit breaker
	hook.Reset()
	if hook.IsOpen() {
		t.Error("Circuit breaker should be closed after reset")
	}

	// Request should pass after reset
	err = hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("Request should pass after reset, got %v", err)
	}
}

func TestHookFunc(t *testing.T) {
	var beforeCalled, afterCalled, errorCalled bool

	hook := NewHookFunc(
		func(ctx context.Context, req *interfaces.Request) error {
			beforeCalled = true
			return nil
		},
		func(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
			afterCalled = true
			return nil
		},
		func(ctx context.Context, req *interfaces.Request, err error) error {
			errorCalled = true
			return nil
		},
	)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	resp := &interfaces.Response{StatusCode: 200}
	ctx := context.Background()
	testErr := errors.New("test error")

	err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	err = hook.AfterResponse(ctx, req, resp)
	if err != nil {
		t.Fatalf("AfterResponse failed: %v", err)
	}

	err = hook.OnError(ctx, req, testErr)
	if err != nil {
		t.Fatalf("OnError failed: %v", err)
	}

	if !beforeCalled {
		t.Error("Before function should be called")
	}

	if !afterCalled {
		t.Error("After function should be called")
	}

	if !errorCalled {
		t.Error("Error function should be called")
	}
}
