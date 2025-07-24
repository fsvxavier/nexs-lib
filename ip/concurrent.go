// Package ip concurrent provides concurrent processing capabilities for IP operations.
// This file implements worker pools and goroutine management for heavy IP operations.
package ip

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Task represents a unit of work that can be executed by the worker pool
type Task struct {
	ID       string
	Function func(ctx context.Context) (interface{}, error)
	Context  context.Context
	Result   chan TaskResult
}

// TaskResult contains the result of a task execution
type TaskResult struct {
	ID       string
	Result   interface{}
	Error    error
	Duration time.Duration
}

// WorkerPool manages a pool of goroutines for concurrent IP operations
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	quit       chan bool
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	stats      *PoolStats
	statsMutex sync.RWMutex
}

// PoolStats contains statistics about the worker pool
type PoolStats struct {
	TasksProcessed  int64
	TasksInQueue    int64
	ActiveWorkers   int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	ErrorCount      int64
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, workers*2), // Buffer for better throughput
		quit:      make(chan bool),
		ctx:       ctx,
		cancel:    cancel,
		stats:     &PoolStats{},
	}

	pool.start()
	return pool
}

// start initializes and starts the worker goroutines
func (p *WorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is the main worker goroutine function
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	p.incrementActiveWorkers()
	defer p.decrementActiveWorkers()

	for {
		select {
		case task := <-p.taskQueue:
			p.processTask(task)
		case <-p.quit:
			return
		case <-p.ctx.Done():
			return
		}
	}
}

// processTask executes a single task
func (p *WorkerPool) processTask(task Task) {
	start := time.Now()

	result, err := task.Function(task.Context)
	duration := time.Since(start)

	taskResult := TaskResult{
		ID:       task.ID,
		Result:   result,
		Error:    err,
		Duration: duration,
	}

	// Update statistics
	p.updateStats(duration, err != nil)

	// Send result back
	select {
	case task.Result <- taskResult:
	case <-task.Context.Done():
	case <-p.ctx.Done():
	}

	p.decrementTasksInQueue()
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(ctx context.Context, id string, fn func(ctx context.Context) (interface{}, error)) <-chan TaskResult {
	resultChan := make(chan TaskResult, 1)

	task := Task{
		ID:       id,
		Function: fn,
		Context:  ctx,
		Result:   resultChan,
	}

	p.incrementTasksInQueue()

	select {
	case p.taskQueue <- task:
		return resultChan
	case <-ctx.Done():
		p.decrementTasksInQueue()
		go func() {
			resultChan <- TaskResult{
				ID:    id,
				Error: ctx.Err(),
			}
		}()
		return resultChan
	case <-p.ctx.Done():
		p.decrementTasksInQueue()
		go func() {
			resultChan <- TaskResult{
				ID:    id,
				Error: p.ctx.Err(),
			}
		}()
		return resultChan
	}
}

// SubmitBatch submits multiple tasks and returns a channel for results
func (p *WorkerPool) SubmitBatch(ctx context.Context, tasks map[string]func(ctx context.Context) (interface{}, error)) <-chan TaskResult {
	resultChan := make(chan TaskResult, len(tasks))

	var wg sync.WaitGroup

	for id, fn := range tasks {
		wg.Add(1)
		go func(taskID string, taskFn func(ctx context.Context) (interface{}, error)) {
			defer wg.Done()

			taskResultChan := p.Submit(ctx, taskID, taskFn)
			select {
			case result := <-taskResultChan:
				resultChan <- result
			case <-ctx.Done():
				resultChan <- TaskResult{
					ID:    taskID,
					Error: ctx.Err(),
				}
			}
		}(id, fn)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// GetStats returns current pool statistics
func (p *WorkerPool) GetStats() PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()
	return *p.stats
}

// GetQueueSize returns the current number of tasks in the queue
func (p *WorkerPool) GetQueueSize() int {
	return len(p.taskQueue)
}

// GetActiveWorkers returns the number of currently active workers
func (p *WorkerPool) GetActiveWorkers() int64 {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()
	return p.stats.ActiveWorkers
}

// Close gracefully shuts down the worker pool
func (p *WorkerPool) Close() error {
	p.cancel()
	close(p.quit)
	p.wg.Wait()
	close(p.taskQueue)
	return nil
}

// updateStats updates the pool statistics
func (p *WorkerPool) updateStats(duration time.Duration, isError bool) {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()

	p.stats.TasksProcessed++
	p.stats.TotalDuration += duration
	p.stats.AverageDuration = time.Duration(int64(p.stats.TotalDuration) / p.stats.TasksProcessed)

	if isError {
		p.stats.ErrorCount++
	}
}

// incrementActiveWorkers increments the active worker count
func (p *WorkerPool) incrementActiveWorkers() {
	p.statsMutex.Lock()
	p.stats.ActiveWorkers++
	p.statsMutex.Unlock()
}

// decrementActiveWorkers decrements the active worker count
func (p *WorkerPool) decrementActiveWorkers() {
	p.statsMutex.Lock()
	p.stats.ActiveWorkers--
	p.statsMutex.Unlock()
}

// incrementTasksInQueue increments the tasks in queue count
func (p *WorkerPool) incrementTasksInQueue() {
	p.statsMutex.Lock()
	p.stats.TasksInQueue++
	p.statsMutex.Unlock()
}

// decrementTasksInQueue decrements the tasks in queue count
func (p *WorkerPool) decrementTasksInQueue() {
	p.statsMutex.Lock()
	p.stats.TasksInQueue--
	p.statsMutex.Unlock()
}

// ConcurrentIPProcessor provides concurrent processing for multiple IP operations
type ConcurrentIPProcessor struct {
	detector   *AdvancedDetector
	workerPool *WorkerPool
}

// NewConcurrentIPProcessor creates a new concurrent IP processor
func NewConcurrentIPProcessor(detector *AdvancedDetector, workers int) *ConcurrentIPProcessor {
	return &ConcurrentIPProcessor{
		detector:   detector,
		workerPool: NewWorkerPool(workers),
	}
}

// ProcessIPs processes multiple IPs concurrently
func (p *ConcurrentIPProcessor) ProcessIPs(ctx context.Context, ips []string) <-chan IPProcessResult {
	resultChan := make(chan IPProcessResult, len(ips))

	tasks := make(map[string]func(ctx context.Context) (interface{}, error))

	for _, ipStr := range ips {
		ip := ipStr // Capture loop variable
		tasks[ip] = func(ctx context.Context) (interface{}, error) {
			parsedIP := parseIP(ip)
			if parsedIP == nil {
				return nil, fmt.Errorf("invalid IP address: %s", ip)
			}
			return p.detector.DetectAdvanced(ctx, parsedIP)
		}
	}

	taskResults := p.workerPool.SubmitBatch(ctx, tasks)

	go func() {
		defer close(resultChan)
		for taskResult := range taskResults {
			ipResult := IPProcessResult{
				IP:       taskResult.ID,
				Duration: taskResult.Duration,
				Error:    taskResult.Error,
			}

			if taskResult.Result != nil {
				if detection, ok := taskResult.Result.(*DetectionResult); ok {
					ipResult.Detection = detection
				}
			}

			resultChan <- ipResult
		}
	}()

	return resultChan
}

// IPProcessResult contains the result of processing a single IP
type IPProcessResult struct {
	IP        string
	Detection *DetectionResult
	Duration  time.Duration
	Error     error
}

// Close closes the concurrent processor
func (p *ConcurrentIPProcessor) Close() error {
	return p.workerPool.Close()
}

// parseIP safely parses an IP address string
func parseIP(ipStr string) net.IP {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}
	return ip
}
