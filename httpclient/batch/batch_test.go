package batch

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Mock client for testing
type mockClient struct {
	provider mockProvider
}

func (c *mockClient) Get(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Post(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Put(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Delete(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Patch(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Head(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Options(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) Execute(ctx context.Context, method, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}

func (c *mockClient) SetHeaders(headers map[string]string) interfaces.Client {
	return c
}

func (c *mockClient) SetTimeout(timeout time.Duration) interfaces.Client {
	return c
}

func (c *mockClient) SetErrorHandler(handler interfaces.ErrorHandler) interfaces.Client {
	return c
}

func (c *mockClient) SetRetryConfig(config *interfaces.RetryConfig) interfaces.Client {
	return c
}

func (c *mockClient) Unmarshal(v interface{}) interfaces.Client {
	return c
}

func (c *mockClient) UnmarshalResponse(resp *interfaces.Response, v interface{}) error {
	return nil
}

func (c *mockClient) AddMiddleware(middleware interfaces.Middleware) interfaces.Client {
	return c
}

func (c *mockClient) RemoveMiddleware(middleware interfaces.Middleware) interfaces.Client {
	return c
}

func (c *mockClient) AddHook(hook interfaces.Hook) interfaces.Client {
	return c
}

func (c *mockClient) RemoveHook(hook interfaces.Hook) interfaces.Client {
	return c
}

func (c *mockClient) Batch() interfaces.BatchRequestBuilder {
	return NewBuilder(c)
}

func (c *mockClient) Stream(ctx context.Context, method, endpoint string, handler interfaces.StreamHandler) error {
	return nil
}

func (c *mockClient) GetProvider() interfaces.Provider {
	return &c.provider
}

func (c *mockClient) GetConfig() *interfaces.Config {
	return &interfaces.Config{}
}

func (c *mockClient) GetID() string {
	return "mock-client"
}

func (c *mockClient) IsHealthy() bool {
	return true
}

func (c *mockClient) GetMetrics() *interfaces.ProviderMetrics {
	return &interfaces.ProviderMetrics{}
}

// Mock provider for testing
type mockProvider struct {
	responses []*interfaces.Response
	errors    []error
	callCount int
}

func (p *mockProvider) Name() string {
	return "mock"
}

func (p *mockProvider) Version() string {
	return "1.0.0"
}

func (p *mockProvider) DoRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	if p.callCount < len(p.errors) && p.errors[p.callCount] != nil {
		err := p.errors[p.callCount]
		p.callCount++
		return nil, err
	}

	if p.callCount < len(p.responses) {
		resp := p.responses[p.callCount]
		p.callCount++
		return resp, nil
	}

	// Default response
	p.callCount++
	return &interfaces.Response{
		StatusCode: 200,
		Body:       []byte("success"),
	}, nil
}

func (p *mockProvider) Configure(config *interfaces.Config) error {
	return nil
}

func (p *mockProvider) SetDefaults() {}

func (p *mockProvider) IsHealthy() bool {
	return true
}

func (p *mockProvider) GetMetrics() *interfaces.ProviderMetrics {
	return &interfaces.ProviderMetrics{}
}

func TestBuilder_Add(t *testing.T) {
	client := &mockClient{}
	builder := NewBuilder(client)

	builder.Add("GET", "http://test.com/1", nil)
	builder.Add("POST", "http://test.com/2", "body")

	if builder.Count() != 2 {
		t.Errorf("Expected 2 requests, got %d", builder.Count())
	}
}

func TestBuilder_AddRequest(t *testing.T) {
	client := &mockClient{}
	builder := NewBuilder(client)

	req1 := &interfaces.Request{Method: "GET", URL: "http://test.com/1"}
	req2 := &interfaces.Request{Method: "POST", URL: "http://test.com/2", Body: "data"}

	builder.AddRequest(req1).AddRequest(req2)

	if builder.Count() != 2 {
		t.Errorf("Expected 2 requests, got %d", builder.Count())
	}
}

func TestBuilder_Execute(t *testing.T) {
	client := &mockClient{}
	client.provider.responses = []*interfaces.Response{
		{StatusCode: 200, Body: []byte("response1")},
		{StatusCode: 201, Body: []byte("response2")},
	}

	builder := NewBuilder(client)
	builder.Add("GET", "http://test.com/1", nil)
	builder.Add("POST", "http://test.com/2", "body")

	ctx := context.Background()
	responses, err := builder.Execute(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}

	if responses[0].StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", responses[0].StatusCode)
	}

	if responses[1].StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", responses[1].StatusCode)
	}
}

func TestBuilder_ExecuteParallel(t *testing.T) {
	client := &mockClient{}
	client.provider.responses = []*interfaces.Response{
		{StatusCode: 200, Body: []byte("response1")},
		{StatusCode: 201, Body: []byte("response2")},
		{StatusCode: 202, Body: []byte("response3")},
	}

	builder := NewBuilder(client)
	builder.Add("GET", "http://test.com/1", nil)
	builder.Add("POST", "http://test.com/2", "body")
	builder.Add("PUT", "http://test.com/3", "data")

	ctx := context.Background()
	responses, err := builder.ExecuteParallel(ctx, 2)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}

	// Check that all responses are present (order may vary in parallel execution)
	statusCodes := make(map[int]bool)
	for _, response := range responses {
		statusCodes[response.StatusCode] = true
	}

	expectedCodes := []int{200, 201, 202}
	for _, expectedCode := range expectedCodes {
		if !statusCodes[expectedCode] {
			t.Errorf("Expected status code %d not found in responses", expectedCode)
		}
	}
}

func TestBuilder_ExecuteWithError(t *testing.T) {
	client := &mockClient{}
	client.provider.errors = []error{
		nil,
		errors.New("request failed"),
	}

	builder := NewBuilder(client)
	builder.Add("GET", "http://test.com/1", nil)
	builder.Add("POST", "http://test.com/2", "body")

	ctx := context.Background()
	_, err := builder.Execute(ctx)

	if err == nil {
		t.Error("Expected error from second request")
	}

	if !contains(err.Error(), "batch request 1 failed") {
		t.Errorf("Expected error message to contain 'batch request 1 failed', got %v", err)
	}
}

func TestBuilder_Clear(t *testing.T) {
	client := &mockClient{}
	builder := NewBuilder(client)

	builder.Add("GET", "http://test.com/1", nil)
	builder.Add("POST", "http://test.com/2", "body")

	if builder.Count() != 2 {
		t.Errorf("Expected 2 requests before clear, got %d", builder.Count())
	}

	builder.Clear()

	if builder.Count() != 0 {
		t.Errorf("Expected 0 requests after clear, got %d", builder.Count())
	}
}

func TestBatchExecutor_Configuration(t *testing.T) {
	client := &mockClient{}
	executor := NewBatchExecutor(client)

	executor.SetMaxBatchSize(50).
		SetBatchTimeout(15 * time.Second).
		SetMaxConcurrency(5).
		SetFailureThreshold(0.2)

	if executor.maxBatchSize != 50 {
		t.Errorf("Expected max batch size 50, got %d", executor.maxBatchSize)
	}

	if executor.batchTimeout != 15*time.Second {
		t.Errorf("Expected batch timeout 15s, got %v", executor.batchTimeout)
	}

	if executor.maxConcurrency != 5 {
		t.Errorf("Expected max concurrency 5, got %d", executor.maxConcurrency)
	}

	if executor.failureThreshold != 0.2 {
		t.Errorf("Expected failure threshold 0.2, got %f", executor.failureThreshold)
	}
}

func TestBatchExecutor_ExecuteWithStrategy(t *testing.T) {
	client := &mockClient{}
	client.provider.responses = []*interfaces.Response{
		{StatusCode: 200, Body: []byte("response1")},
		{StatusCode: 201, Body: []byte("response2")},
	}

	executor := NewBatchExecutor(client)

	requests := []*interfaces.Request{
		{Method: "GET", URL: "http://test.com/1"},
		{Method: "POST", URL: "http://test.com/2", Body: "data"},
	}

	ctx := context.Background()
	result, err := executor.ExecuteWithStrategy(ctx, requests, StrategyParallel)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsSuccess() {
		t.Error("Expected successful batch execution")
	}

	if result.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", result.TotalCount)
	}

	if result.SuccessCount != 2 {
		t.Errorf("Expected success count 2, got %d", result.SuccessCount)
	}

	if result.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", result.FailureCount)
	}
}

func TestBatchExecutor_FailFastStrategy(t *testing.T) {
	client := &mockClient{}
	client.provider.responses = []*interfaces.Response{
		{StatusCode: 200, Body: []byte("response1")},
		{StatusCode: 500, Body: []byte("error"), IsError: true},
	}

	executor := NewBatchExecutor(client)

	requests := []*interfaces.Request{
		{Method: "GET", URL: "http://test.com/1"},
		{Method: "POST", URL: "http://test.com/2", Body: "data"},
	}

	ctx := context.Background()
	_, err := executor.ExecuteWithStrategy(ctx, requests, StrategyFailFast)

	if err == nil {
		t.Error("Expected error from fail-fast strategy")
	}

	if !contains(err.Error(), "fail-fast") {
		t.Errorf("Expected error message to contain 'fail-fast', got %v", err)
	}
}

func TestBatchResult_Metrics(t *testing.T) {
	result := &BatchResult{
		TotalCount:   10,
		SuccessCount: 8,
		FailureCount: 2,
	}

	if result.SuccessRate() != 0.8 {
		t.Errorf("Expected success rate 0.8, got %f", result.SuccessRate())
	}

	if result.FailureRate() != 0.2 {
		t.Errorf("Expected failure rate 0.2, got %f", result.FailureRate())
	}

	if result.IsSuccess() {
		t.Error("Expected batch to not be successful with failures")
	}
}

func TestBatchResult_EmptyBatch(t *testing.T) {
	result := &BatchResult{
		TotalCount:   0,
		SuccessCount: 0,
		FailureCount: 0,
	}

	if result.SuccessRate() != 0.0 {
		t.Errorf("Expected success rate 0.0 for empty batch, got %f", result.SuccessRate())
	}

	if result.FailureRate() != 0.0 {
		t.Errorf("Expected failure rate 0.0 for empty batch, got %f", result.FailureRate())
	}
}

func TestBatchExecutor_SplitIntoBatches(t *testing.T) {
	client := &mockClient{}
	executor := NewBatchExecutor(client).SetMaxBatchSize(3)

	requests := make([]*interfaces.Request, 7)
	for i := range requests {
		requests[i] = &interfaces.Request{Method: "GET", URL: "http://test.com"}
	}

	batches := executor.splitIntoBatches(requests)

	if len(batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(batches))
	}

	if len(batches[0]) != 3 {
		t.Errorf("Expected first batch size 3, got %d", len(batches[0]))
	}

	if len(batches[1]) != 3 {
		t.Errorf("Expected second batch size 3, got %d", len(batches[1]))
	}

	if len(batches[2]) != 1 {
		t.Errorf("Expected third batch size 1, got %d", len(batches[2]))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(s, substr))))
}

// Simple implementation of strings.Contains for testing
func strings_Contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
