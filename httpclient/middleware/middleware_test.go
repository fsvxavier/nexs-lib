package middleware

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Mock middleware for testing
type testMiddleware struct {
	name  string
	calls []string
}

func (m *testMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	m.calls = append(m.calls, "before-"+m.name)
	resp, err := next(ctx, req)
	m.calls = append(m.calls, "after-"+m.name)
	return resp, err
}

func TestChain_BasicExecution(t *testing.T) {
	chain := NewChain()

	middleware1 := &testMiddleware{name: "1"}
	middleware2 := &testMiddleware{name: "2"}

	chain.Add(middleware1).Add(middleware2)

	if chain.Count() != 2 {
		t.Errorf("Expected 2 middlewares, got %d", chain.Count())
	}

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	called := false
	finalNext := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		called = true
		return &interfaces.Response{StatusCode: 200}, nil
	}

	resp, err := chain.Execute(ctx, req, finalNext)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Final next function was not called")
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check middleware execution order
	// Since execution is interleaved, we need to check the actual pattern
	// middleware1: before, then middleware2: before, then middleware2: after, then middleware1: after
	if len(middleware1.calls) != 2 || len(middleware2.calls) != 2 {
		t.Errorf("Expected 2 calls each, got middleware1: %d, middleware2: %d", len(middleware1.calls), len(middleware2.calls))
	}

	if middleware1.calls[0] != "before-1" {
		t.Errorf("Expected first call to be before-1, got %v", middleware1.calls[0])
	}

	if middleware2.calls[0] != "before-2" {
		t.Errorf("Expected middleware2 first call to be before-2, got %v", middleware2.calls[0])
	}

	if middleware2.calls[1] != "after-2" {
		t.Errorf("Expected middleware2 second call to be after-2, got %v", middleware2.calls[1])
	}

	if middleware1.calls[1] != "after-1" {
		t.Errorf("Expected middleware1 second call to be after-1, got %v", middleware1.calls[1])
	}
}

func TestChain_RemoveMiddleware(t *testing.T) {
	chain := NewChain()

	middleware1 := &testMiddleware{name: "1"}
	middleware2 := &testMiddleware{name: "2"}

	chain.Add(middleware1).Add(middleware2)
	chain.Remove(middleware1)

	if chain.Count() != 1 {
		t.Errorf("Expected 1 middleware after removal, got %d", chain.Count())
	}
}

func TestChain_Clear(t *testing.T) {
	chain := NewChain()

	middleware1 := &testMiddleware{name: "1"}
	middleware2 := &testMiddleware{name: "2"}

	chain.Add(middleware1).Add(middleware2)
	chain.Clear()

	if chain.Count() != 0 {
		t.Errorf("Expected 0 middlewares after clear, got %d", chain.Count())
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logs := []string{}
	logger := func(format string, args ...interface{}) {
		logs = append(logs, format)
	}

	middleware := NewLoggingMiddleware(logger)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		return &interfaces.Response{StatusCode: 200}, nil
	}

	resp, err := middleware.Process(ctx, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}
}

func TestRetryMiddleware(t *testing.T) {
	attempts := 0
	middleware := NewRetryMiddleware(2, DefaultRetryCondition)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		attempts++
		if attempts < 3 {
			return &interfaces.Response{StatusCode: 500}, nil
		}
		return &interfaces.Response{StatusCode: 200}, nil
	}

	resp, err := middleware.Process(ctx, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryMiddleware_MaxRetriesExceeded(t *testing.T) {
	middleware := NewRetryMiddleware(1, DefaultRetryCondition)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		return &interfaces.Response{StatusCode: 500}, nil
	}

	resp, err := middleware.Process(ctx, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware(t *testing.T) {
	middleware := NewAuthMiddleware("Authorization", "Bearer token123")

	req := &interfaces.Request{
		Method: "GET",
		URL:    "http://test.com",
	}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		if req.Headers == nil {
			t.Error("Headers should not be nil")
		}

		auth := req.Headers["Authorization"]
		if auth != "Bearer token123" {
			t.Errorf("Expected Authorization header to be 'Bearer token123', got %s", auth)
		}

		return &interfaces.Response{StatusCode: 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDynamicAuthMiddleware(t *testing.T) {
	provider := func() string {
		return "Bearer dynamic-token"
	}

	middleware := NewDynamicAuthMiddleware("Authorization", provider)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "http://test.com",
	}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		auth := req.Headers["Authorization"]
		if auth != "Bearer dynamic-token" {
			t.Errorf("Expected Authorization header to be 'Bearer dynamic-token', got %s", auth)
		}

		return &interfaces.Response{StatusCode: 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCompressionMiddleware(t *testing.T) {
	middleware := NewCompressionMiddleware(interfaces.CompressionGzip, interfaces.CompressionDeflate)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "http://test.com",
	}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		encoding := req.Headers["Accept-Encoding"]
		if !strings.Contains(encoding, "gzip") {
			t.Errorf("Expected Accept-Encoding to contain gzip, got %s", encoding)
		}

		return &interfaces.Response{StatusCode: 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	var collectedMethod, collectedURL string
	var collectedStatusCode int
	var collectedDuration time.Duration
	var collectedErr error

	collector := func(method, url string, statusCode int, duration time.Duration, err error) {
		collectedMethod = method
		collectedURL = url
		collectedStatusCode = statusCode
		collectedDuration = duration
		collectedErr = err
	}

	middleware := NewMetricsMiddleware(collector)

	req := &interfaces.Request{Method: "GET", URL: "http://test.com"}
	ctx := context.Background()

	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return &interfaces.Response{StatusCode: 200}, nil
	}

	resp, err := middleware.Process(ctx, req, next)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if collectedMethod != "GET" {
		t.Errorf("Expected method GET, got %s", collectedMethod)
	}

	if collectedURL != "http://test.com" {
		t.Errorf("Expected URL http://test.com, got %s", collectedURL)
	}

	if collectedStatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", collectedStatusCode)
	}

	if collectedDuration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", collectedDuration)
	}

	if collectedErr != nil {
		t.Errorf("Expected no error, got %v", collectedErr)
	}
}

func TestDefaultRetryCondition(t *testing.T) {
	tests := []struct {
		name        string
		resp        *interfaces.Response
		err         error
		shouldRetry bool
	}{
		{"Error", nil, errors.New("network error"), true},
		{"Nil response", nil, nil, true},
		{"500 Internal Server Error", &interfaces.Response{StatusCode: 500}, nil, true},
		{"502 Bad Gateway", &interfaces.Response{StatusCode: 502}, nil, true},
		{"429 Too Many Requests", &interfaces.Response{StatusCode: 429}, nil, true},
		{"408 Request Timeout", &interfaces.Response{StatusCode: 408}, nil, true},
		{"200 OK", &interfaces.Response{StatusCode: 200}, nil, false},
		{"400 Bad Request", &interfaces.Response{StatusCode: 400}, nil, false},
		{"404 Not Found", &interfaces.Response{StatusCode: 404}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultRetryCondition(tt.resp, tt.err)
			if result != tt.shouldRetry {
				t.Errorf("Expected %v, got %v", tt.shouldRetry, result)
			}
		})
	}
}
