package httpclient

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// TestMiddleware is a test middleware implementation
type TestMiddleware struct {
	name     string
	executed bool
}

func (m *TestMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	m.executed = true
	// Add a header to track middleware execution
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	req.Headers["X-Middleware-"+m.name] = "executed"
	return next(ctx, req)
}

// TestHook is a test hook implementation
type TestHook struct {
	name               string
	beforeRequestCalls int
	afterResponseCalls int
	onErrorCalls       int
}

func (h *TestHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	h.beforeRequestCalls++
	// Add a header to track hook execution
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	req.Headers["X-Hook-"+h.name] = "before"
	return nil
}

func (h *TestHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	h.afterResponseCalls++
	return nil
}

func (h *TestHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	h.onErrorCalls++
	return nil
}

// TestStreamHandler is a test stream handler implementation
type TestStreamHandler struct {
	chunks    [][]byte
	completed bool
	errored   bool
}

func (h *TestStreamHandler) OnData(data []byte) error {
	h.chunks = append(h.chunks, make([]byte, len(data)))
	copy(h.chunks[len(h.chunks)-1], data)
	return nil
}

func (h *TestStreamHandler) OnError(err error) {
	h.errored = true
}

func (h *TestStreamHandler) OnComplete() {
	h.completed = true
}

func TestClientMiddleware(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get concrete client for testing internal state
	concreteClient := client.(*Client)

	// Test adding middleware
	middleware1 := &TestMiddleware{name: "test1"}
	middleware2 := &TestMiddleware{name: "test2"}

	client.AddMiddleware(middleware1)
	client.AddMiddleware(middleware2)

	// Test duplicate middleware (should not be added)
	initialCount := len(concreteClient.middlewares)
	client.AddMiddleware(middleware1)
	if len(concreteClient.middlewares) != initialCount {
		t.Errorf("Expected middleware count to remain %d, got %d", initialCount, len(concreteClient.middlewares))
	}

	// Test removing middleware
	client.RemoveMiddleware(middleware1)
	found := false
	for _, m := range concreteClient.middlewares {
		if m == middleware1 {
			found = true
			break
		}
	}
	if found {
		t.Error("Middleware should have been removed")
	}

	// Test removing non-existent middleware (should not error)
	client.RemoveMiddleware(&TestMiddleware{name: "nonexistent"})
}

func TestClientHooks(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get concrete client for testing internal state
	concreteClient := client.(*Client)

	// Test adding hooks
	hook1 := &TestHook{name: "test1"}
	hook2 := &TestHook{name: "test2"}

	client.AddHook(hook1)
	client.AddHook(hook2)

	// Test duplicate hook (should not be added)
	initialCount := len(concreteClient.hooks)
	client.AddHook(hook1)
	if len(concreteClient.hooks) != initialCount {
		t.Errorf("Expected hook count to remain %d, got %d", initialCount, len(concreteClient.hooks))
	}

	// Test removing hook
	client.RemoveHook(hook1)
	found := false
	for _, h := range concreteClient.hooks {
		testHook, ok := h.(*TestHook)
		if ok && testHook == hook1 {
			found = true
			break
		}
	}
	if found {
		t.Error("Hook should have been removed")
	}

	// Test removing non-existent hook (should not error)
	client.RemoveHook(&TestHook{name: "nonexistent"})
}

func TestClientBatch(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test batch creation
	batch := client.Batch()
	if batch == nil {
		t.Error("Batch should not be nil")
	}

	// Test that batch uses the same client
	// Note: This is a basic test - full batch functionality is tested in batch package
}

func TestClientStream(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test stream with nil handler (should error)
	ctx := context.Background()
	err = client.Stream(ctx, "GET", "http://httpbin.org/get", nil)
	if err == nil {
		t.Error("Expected error for nil stream handler")
	}

	// Test stream with valid handler
	handler := &TestStreamHandler{}
	err = client.Stream(ctx, "GET", "http://httpbin.org/get", handler)
	// This might fail due to network, but we're testing the API
	// The actual streaming functionality is tested in the streaming package
}

func TestClientUnmarshalResponse(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with nil response
	var target map[string]interface{}
	err = client.UnmarshalResponse(nil, &target)
	if err == nil {
		t.Error("Expected error for nil response")
	}

	// Test with nil target
	response := &interfaces.Response{
		StatusCode: 200,
		Body:       []byte(`{"test": "value"}`),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	err = client.UnmarshalResponse(response, nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}

	// Test with valid response and target
	err = client.UnmarshalResponse(response, &target)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify unmarshaling worked
	if target["test"] != "value" {
		t.Errorf("Expected target['test'] to be 'value', got %v", target["test"])
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	// Create a test client
	client, err := NewWithConfig(interfaces.ProviderNetHTTP, config.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Add middleware and hooks
	middleware := &TestMiddleware{name: "integration"}
	hook := &TestHook{name: "integration"}

	client.AddMiddleware(middleware)
	client.AddHook(hook)

	// Make a request to test integration
	ctx := context.Background()
	resp, err := client.Execute(ctx, "GET", "http://httpbin.org/get", nil)

	// Check if middleware was executed
	if !middleware.executed {
		t.Error("Middleware should have been executed")
	}

	// Check if hooks were called
	if hook.beforeRequestCalls == 0 {
		t.Error("BeforeRequest hook should have been called")
	}

	if hook.afterResponseCalls == 0 {
		t.Error("AfterResponse hook should have been called")
	}

	// If network call succeeded, verify response
	if err == nil && resp != nil {
		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	}
	// Note: We don't fail the test on network errors since they're environment dependent
}

func TestConcurrentMiddlewareAccess(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test concurrent access to middlewares
	done := make(chan bool, 10)

	// Start multiple goroutines adding/removing middlewares
	for i := 0; i < 5; i++ {
		go func(id int) {
			middleware := &TestMiddleware{name: fmt.Sprintf("concurrent%d", id)}
			client.AddMiddleware(middleware)
			time.Sleep(time.Millisecond * 10)
			client.RemoveMiddleware(middleware)
			done <- true
		}(i)
	}

	// Start multiple goroutines adding/removing hooks
	for i := 0; i < 5; i++ {
		go func(id int) {
			hook := &TestHook{name: fmt.Sprintf("concurrent%d", id)}
			client.AddHook(hook)
			time.Sleep(time.Millisecond * 10)
			client.RemoveHook(hook)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test should complete without deadlocks or race conditions
}

func TestMethodChaining(t *testing.T) {
	// Create a test client
	client, err := New(interfaces.ProviderNetHTTP, "http://httpbin.org")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get concrete client for testing internal state
	concreteClient := client.(*Client)

	middleware := &TestMiddleware{name: "chain"}
	hook := &TestHook{name: "chain"}

	// Test method chaining
	result := client.
		AddMiddleware(middleware).
		AddHook(hook).
		SetTimeout(10 * time.Second)

	if result != client {
		t.Error("Method chaining should return the same client instance")
	}

	// Verify middleware and hook were added
	found := false
	for _, m := range concreteClient.middlewares {
		if m == middleware {
			found = true
			break
		}
	}
	if !found {
		t.Error("Middleware should have been added through method chaining")
	}

	found = false
	for _, h := range concreteClient.hooks {
		testHook, ok := h.(*TestHook)
		if ok && testHook == hook {
			found = true
			break
		}
	}
	if !found {
		t.Error("Hook should have been added through method chaining")
	}
}
