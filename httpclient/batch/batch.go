// Package batch provides batch request processing capabilities.
package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Builder implements the BatchRequestBuilder interface.
type Builder struct {
	requests []batchRequest
	client   interfaces.Client
	mu       sync.Mutex
}

// batchRequest represents a single request in a batch.
type batchRequest struct {
	request *interfaces.Request
	index   int
}

// NewBuilder creates a new batch request builder.
func NewBuilder(client interfaces.Client) *Builder {
	return &Builder{
		requests: make([]batchRequest, 0),
		client:   client,
	}
}

// Add adds a new request to the batch.
func (b *Builder) Add(method, endpoint string, body interface{}) interfaces.BatchRequestBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()

	req := &interfaces.Request{
		Method: method,
		URL:    endpoint,
		Body:   body,
	}

	b.requests = append(b.requests, batchRequest{
		request: req,
		index:   len(b.requests),
	})

	return b
}

// AddRequest adds a prepared request to the batch.
func (b *Builder) AddRequest(req *interfaces.Request) interfaces.BatchRequestBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.requests = append(b.requests, batchRequest{
		request: req,
		index:   len(b.requests),
	})

	return b
}

// Execute executes all requests in the batch sequentially.
func (b *Builder) Execute(ctx context.Context) ([]*interfaces.Response, error) {
	b.mu.Lock()
	requests := make([]batchRequest, len(b.requests))
	copy(requests, b.requests)
	b.mu.Unlock()

	responses := make([]*interfaces.Response, len(requests))

	for i, batchReq := range requests {
		req := batchReq.request
		if req.Context == nil {
			req.Context = ctx
		}

		resp, err := b.executeRequest(ctx, req)
		if err != nil {
			return responses, fmt.Errorf("batch request %d failed: %w", i, err)
		}

		responses[i] = resp
	}

	return responses, nil
}

// ExecuteParallel executes all requests in the batch in parallel with limited concurrency.
func (b *Builder) ExecuteParallel(ctx context.Context, maxConcurrency int) ([]*interfaces.Response, error) {
	b.mu.Lock()
	requests := make([]batchRequest, len(b.requests))
	copy(requests, b.requests)
	b.mu.Unlock()

	if maxConcurrency <= 0 {
		maxConcurrency = 10 // Default concurrency
	}

	responses := make([]*interfaces.Response, len(requests))
	errors := make([]error, len(requests))

	// Create semaphore for limiting concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, batchReq := range requests {
		wg.Add(1)
		go func(index int, req *interfaces.Request) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if req.Context == nil {
				req.Context = ctx
			}

			resp, err := b.executeRequest(ctx, req)
			responses[index] = resp
			errors[index] = err
		}(i, batchReq.request)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return responses, fmt.Errorf("batch request %d failed: %w", i, err)
		}
	}

	return responses, nil
}

// executeRequest executes a single request using the client's provider.
func (b *Builder) executeRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	provider := b.client.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("no provider available")
	}

	return provider.DoRequest(ctx, req)
}

// Count returns the number of requests in the batch.
func (b *Builder) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.requests)
}

// Clear removes all requests from the batch.
func (b *Builder) Clear() *Builder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requests = b.requests[:0]
	return b
}

// BatchExecutor provides advanced batch execution capabilities.
type BatchExecutor struct {
	client           interfaces.Client
	maxBatchSize     int
	batchTimeout     time.Duration
	maxConcurrency   int
	retryConfig      *interfaces.RetryConfig
	failureThreshold float64 // Percentage of failures that cause batch to fail
}

// NewBatchExecutor creates a new batch executor with configuration.
func NewBatchExecutor(client interfaces.Client) *BatchExecutor {
	return &BatchExecutor{
		client:           client,
		maxBatchSize:     100,
		batchTimeout:     30 * time.Second,
		maxConcurrency:   10,
		failureThreshold: 0.1, // 10% failure threshold
	}
}

// SetMaxBatchSize sets the maximum batch size.
func (e *BatchExecutor) SetMaxBatchSize(size int) *BatchExecutor {
	e.maxBatchSize = size
	return e
}

// SetBatchTimeout sets the batch execution timeout.
func (e *BatchExecutor) SetBatchTimeout(timeout time.Duration) *BatchExecutor {
	e.batchTimeout = timeout
	return e
}

// SetMaxConcurrency sets the maximum concurrency for parallel execution.
func (e *BatchExecutor) SetMaxConcurrency(concurrency int) *BatchExecutor {
	e.maxConcurrency = concurrency
	return e
}

// SetFailureThreshold sets the failure threshold (0.0 to 1.0).
func (e *BatchExecutor) SetFailureThreshold(threshold float64) *BatchExecutor {
	e.failureThreshold = threshold
	return e
}

// ExecuteWithStrategy executes requests with different strategies.
func (e *BatchExecutor) ExecuteWithStrategy(ctx context.Context, requests []*interfaces.Request, strategy ExecutionStrategy) (*BatchResult, error) {
	if len(requests) == 0 {
		return &BatchResult{}, nil
	}

	// Split into chunks if needed
	batches := e.splitIntoBatches(requests)
	allResults := make([]*interfaces.Response, 0, len(requests))
	allErrors := make([]error, 0)

	for _, batch := range batches {
		batchCtx, cancel := context.WithTimeout(ctx, e.batchTimeout)

		var responses []*interfaces.Response
		var err error

		switch strategy {
		case StrategySequential:
			responses, err = e.executeSequential(batchCtx, batch)
		case StrategyParallel:
			responses, err = e.executeParallel(batchCtx, batch)
		case StrategyFailFast:
			responses, err = e.executeFailFast(batchCtx, batch)
		default:
			responses, err = e.executeParallel(batchCtx, batch)
		}

		cancel()

		if err != nil {
			return e.buildResult(allResults, allErrors, err), err
		}

		allResults = append(allResults, responses...)
	}

	return e.buildResult(allResults, allErrors, nil), nil
}

// splitIntoBatches splits requests into smaller batches.
func (e *BatchExecutor) splitIntoBatches(requests []*interfaces.Request) [][]*interfaces.Request {
	if len(requests) <= e.maxBatchSize {
		return [][]*interfaces.Request{requests}
	}

	batches := make([][]*interfaces.Request, 0)
	for i := 0; i < len(requests); i += e.maxBatchSize {
		end := i + e.maxBatchSize
		if end > len(requests) {
			end = len(requests)
		}
		batches = append(batches, requests[i:end])
	}

	return batches
}

// executeSequential executes requests sequentially.
func (e *BatchExecutor) executeSequential(ctx context.Context, requests []*interfaces.Request) ([]*interfaces.Response, error) {
	builder := NewBuilder(e.client)
	for _, req := range requests {
		builder.AddRequest(req)
	}
	return builder.Execute(ctx)
}

// executeParallel executes requests in parallel.
func (e *BatchExecutor) executeParallel(ctx context.Context, requests []*interfaces.Request) ([]*interfaces.Response, error) {
	builder := NewBuilder(e.client)
	for _, req := range requests {
		builder.AddRequest(req)
	}
	return builder.ExecuteParallel(ctx, e.maxConcurrency)
}

// executeFailFast executes requests and fails fast on first error.
func (e *BatchExecutor) executeFailFast(ctx context.Context, requests []*interfaces.Request) ([]*interfaces.Response, error) {
	responses := make([]*interfaces.Response, len(requests))

	for i, req := range requests {
		if req.Context == nil {
			req.Context = ctx
		}

		provider := e.client.GetProvider()
		if provider == nil {
			return responses, fmt.Errorf("no provider available")
		}

		resp, err := provider.DoRequest(ctx, req)
		if err != nil {
			return responses, fmt.Errorf("request %d failed (fail-fast): %w", i, err)
		}

		responses[i] = resp

		// Check for HTTP errors in fail-fast mode
		if resp.IsError || resp.StatusCode >= 400 {
			return responses, fmt.Errorf("request %d returned error status %d (fail-fast)", i, resp.StatusCode)
		}
	}

	return responses, nil
}

// buildResult builds the final batch result.
func (e *BatchExecutor) buildResult(responses []*interfaces.Response, errors []error, finalError error) *BatchResult {
	result := &BatchResult{
		Responses:    responses,
		Errors:       errors,
		TotalCount:   len(responses),
		SuccessCount: 0,
		FailureCount: 0,
		FinalError:   finalError,
	}

	for _, resp := range responses {
		if resp != nil && !resp.IsError && resp.StatusCode < 400 {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}

	return result
}

// ExecutionStrategy defines batch execution strategies.
type ExecutionStrategy int

const (
	StrategySequential ExecutionStrategy = iota
	StrategyParallel
	StrategyFailFast
)

// BatchResult contains the results of batch execution.
type BatchResult struct {
	Responses    []*interfaces.Response
	Errors       []error
	TotalCount   int
	SuccessCount int
	FailureCount int
	FinalError   error
}

// IsSuccess returns whether the batch execution was successful.
func (r *BatchResult) IsSuccess() bool {
	return r.FinalError == nil && r.FailureCount == 0
}

// SuccessRate returns the success rate as a percentage.
func (r *BatchResult) SuccessRate() float64 {
	if r.TotalCount == 0 {
		return 0.0
	}
	return float64(r.SuccessCount) / float64(r.TotalCount)
}

// FailureRate returns the failure rate as a percentage.
func (r *BatchResult) FailureRate() float64 {
	if r.TotalCount == 0 {
		return 0.0
	}
	return float64(r.FailureCount) / float64(r.TotalCount)
}
