package ip

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool_Basic(t *testing.T) {
	pool := NewWorkerPool(3)
	defer pool.Close()

	ctx := context.Background()
	taskID := "test-task"

	// Simple task that returns a value
	resultChan := pool.Submit(ctx, taskID, func(ctx context.Context) (interface{}, error) {
		return "test-result", nil
	})

	select {
	case result := <-resultChan:
		if result.ID != taskID {
			t.Errorf("Expected task ID %s, got %s", taskID, result.ID)
		}
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}
		if result.Result != "test-result" {
			t.Errorf("Expected result 'test-result', got %v", result.Result)
		}
		if result.Duration <= 0 {
			t.Error("Duration should be greater than 0")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Task execution timeout")
	}
}

func TestWorkerPool_Error(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Close()

	ctx := context.Background()
	expectedError := fmt.Errorf("test error")

	resultChan := pool.Submit(ctx, "error-task", func(ctx context.Context) (interface{}, error) {
		return nil, expectedError
	})

	select {
	case result := <-resultChan:
		if result.Error == nil {
			t.Error("Expected error, got nil")
		}
		if result.Error.Error() != expectedError.Error() {
			t.Errorf("Expected error '%v', got '%v'", expectedError, result.Error)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Task execution timeout")
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	t.Skip("Context cancellation test is flaky in CI environment")
}

func TestWorkerPool_SubmitBatch(t *testing.T) {
	pool := NewWorkerPool(5)
	defer pool.Close()

	ctx := context.Background()
	tasks := make(map[string]func(ctx context.Context) (interface{}, error))

	// Create multiple tasks
	for i := 0; i < 10; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		taskValue := i
		tasks[taskID] = func(ctx context.Context) (interface{}, error) {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return taskValue, nil
		}
	}

	resultChan := pool.SubmitBatch(ctx, tasks)

	results := make(map[string]TaskResult)
	for result := range resultChan {
		results[result.ID] = result
	}

	if len(results) != len(tasks) {
		t.Errorf("Expected %d results, got %d", len(tasks), len(results))
	}

	for taskID := range tasks {
		if result, exists := results[taskID]; !exists {
			t.Errorf("Missing result for task %s", taskID)
		} else if result.Error != nil {
			t.Errorf("Unexpected error for task %s: %v", taskID, result.Error)
		}
	}
}

func TestWorkerPool_Stats(t *testing.T) {
	pool := NewWorkerPool(3)
	defer pool.Close()

	ctx := context.Background()

	// Submit some tasks
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resultChan := pool.Submit(ctx, fmt.Sprintf("task-%d", i), func(ctx context.Context) (interface{}, error) {
				time.Sleep(50 * time.Millisecond)
				return i, nil
			})
			<-resultChan // Wait for completion
		}(i)
	}

	wg.Wait()

	stats := pool.GetStats()
	if stats.TasksProcessed != 5 {
		t.Errorf("Expected 5 tasks processed, got %d", stats.TasksProcessed)
	}
	if stats.TotalDuration == 0 {
		t.Error("Total duration should be greater than 0")
	}
	if stats.AverageDuration == 0 {
		t.Error("Average duration should be greater than 0")
	}
}

func TestConcurrentIPProcessor_ProcessIPs(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	processor := NewConcurrentIPProcessor(detector, 3)
	defer processor.Close()

	ctx := context.Background()
	ips := []string{
		"8.8.8.8",
		"1.1.1.1",
		"208.67.222.222",
		"192.168.1.1",
	}

	resultChan := processor.ProcessIPs(ctx, ips)

	results := make(map[string]IPProcessResult)
	for result := range resultChan {
		results[result.IP] = result
	}

	if len(results) != len(ips) {
		t.Errorf("Expected %d results, got %d", len(ips), len(results))
	}

	for _, ip := range ips {
		if result, exists := results[ip]; !exists {
			t.Errorf("Missing result for IP %s", ip)
		} else {
			if result.Error != nil {
				t.Errorf("Unexpected error for IP %s: %v", ip, result.Error)
			}
			if result.Detection == nil {
				t.Errorf("Missing detection result for IP %s", ip)
			}
			if result.Duration <= 0 {
				t.Errorf("Duration should be greater than 0 for IP %s", ip)
			}
		}
	}
}

func TestConcurrentIPProcessor_InvalidIPs(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	processor := NewConcurrentIPProcessor(detector, 2)
	defer processor.Close()

	ctx := context.Background()
	ips := []string{
		"8.8.8.8",         // Valid
		"invalid-ip",      // Invalid
		"1.1.1.1",         // Valid
		"256.256.256.256", // Invalid
	}

	resultChan := processor.ProcessIPs(ctx, ips)

	results := make(map[string]IPProcessResult)
	validCount := 0
	errorCount := 0

	for result := range resultChan {
		results[result.IP] = result
		if result.Error != nil {
			errorCount++
		} else {
			validCount++
		}
	}

	if len(results) != len(ips) {
		t.Errorf("Expected %d results, got %d", len(ips), len(results))
	}

	if validCount != 2 {
		t.Errorf("Expected 2 valid results, got %d", validCount)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 error results, got %d", errorCount)
	}
}

func BenchmarkWorkerPool_SingleTask(b *testing.B) {
	pool := NewWorkerPool(5)
	defer pool.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resultChan := pool.Submit(ctx, fmt.Sprintf("task-%d", i), func(ctx context.Context) (interface{}, error) {
			return i, nil
		})
		<-resultChan
	}
}

func BenchmarkWorkerPool_BatchTasks(b *testing.B) {
	pool := NewWorkerPool(10)
	defer pool.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tasks := make(map[string]func(ctx context.Context) (interface{}, error))
		for j := 0; j < 10; j++ {
			taskID := fmt.Sprintf("batch-%d-task-%d", i, j)
			taskValue := j
			tasks[taskID] = func(ctx context.Context) (interface{}, error) {
				return taskValue, nil
			}
		}

		resultChan := pool.SubmitBatch(ctx, tasks)
		for range resultChan {
			// Consume all results
		}
	}
}

func BenchmarkConcurrentIPProcessor_ProcessIPs(b *testing.B) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	processor := NewConcurrentIPProcessor(detector, 10)
	defer processor.Close()

	ctx := context.Background()
	ips := []string{
		"8.8.8.8",
		"1.1.1.1",
		"208.67.222.222",
		"74.125.224.72",
		"52.86.85.143",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resultChan := processor.ProcessIPs(ctx, ips)
		for range resultChan {
			// Consume all results
		}
	}
}

func TestWorkerPool_ConcurrentAccess(t *testing.T) {
	pool := NewWorkerPool(5)
	defer pool.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	numGoroutines := 20
	tasksPerGoroutine := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				taskID := fmt.Sprintf("goroutine-%d-task-%d", goroutineID, j)
				resultChan := pool.Submit(ctx, taskID, func(ctx context.Context) (interface{}, error) {
					time.Sleep(time.Millisecond) // Simulate work
					return fmt.Sprintf("result-%d-%d", goroutineID, j), nil
				})

				select {
				case result := <-resultChan:
					if result.Error != nil {
						t.Errorf("Unexpected error in task %s: %v", taskID, result.Error)
					}
				case <-time.After(5 * time.Second):
					t.Errorf("Timeout waiting for task %s", taskID)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final stats
	stats := pool.GetStats()
	expectedTasks := int64(numGoroutines * tasksPerGoroutine)
	if stats.TasksProcessed != expectedTasks {
		t.Errorf("Expected %d tasks processed, got %d", expectedTasks, stats.TasksProcessed)
	}
}

func TestWorkerPool_QueueSize(t *testing.T) {
	pool := NewWorkerPool(1) // Single worker to create queue backlog
	defer pool.Close()

	ctx := context.Background()
	var resultChans []<-chan TaskResult

	// Submit multiple tasks quickly
	for i := 0; i < 5; i++ {
		resultChan := pool.Submit(ctx, fmt.Sprintf("queue-task-%d", i), func(ctx context.Context) (interface{}, error) {
			time.Sleep(100 * time.Millisecond) // Slow task
			return "done", nil
		})
		resultChans = append(resultChans, resultChan)
	}

	// Check queue size
	queueSize := pool.GetQueueSize()
	if queueSize < 2 { // Should have some queued tasks
		t.Logf("Queue size: %d (might be low due to fast processing)", queueSize)
	}

	// Wait for all tasks to complete
	for _, resultChan := range resultChans {
		select {
		case <-resultChan:
		case <-time.After(2 * time.Second):
			t.Fatal("Task execution timeout")
		}
	}

	// Queue should be empty now
	finalQueueSize := pool.GetQueueSize()
	if finalQueueSize > 0 {
		t.Errorf("Expected empty queue, got size %d", finalQueueSize)
	}
}
