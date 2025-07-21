package hooks

import (
	"net/http"
	"testing"
	"time"
)

func TestNewMetricsObserver(t *testing.T) {
	observer := NewMetricsObserver()

	if observer == nil {
		t.Fatal("Expected observer to be created")
	}

	if observer.metrics == nil {
		t.Error("Expected metrics map to be initialized")
	}
}

func TestMetricsObserverOnStart(t *testing.T) {
	observer := NewMetricsObserver()

	observer.OnStart("test-server")

	metrics := observer.GetMetrics("test-server")
	if metrics == nil {
		t.Fatal("Expected metrics to be initialized")
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
}

func TestMetricsObserverOnStop(t *testing.T) {
	observer := NewMetricsObserver()
	observer.OnStart("test-server")

	// Verify metrics exist before stop
	metrics := observer.GetMetrics("test-server")
	if metrics == nil {
		t.Fatal("Expected metrics to exist before stop")
	}

	observer.OnStop("test-server")

	// Verify metrics are cleaned up after stop
	metrics = observer.GetMetrics("test-server")
	if metrics != nil {
		t.Error("Expected metrics to be cleaned up after stop")
	}
}

func TestMetricsObserverOnRequest(t *testing.T) {
	observer := NewMetricsObserver()
	observer.OnStart("test-server")

	req, _ := http.NewRequest("GET", "/api/test", nil)

	// Test normal request
	observer.OnRequest("test-server", req, 200, time.Millisecond*100)

	metrics := observer.GetMetrics("test-server")
	if metrics == nil {
		t.Fatal("Expected metrics to exist")
	}

	if metrics.RequestCount != 1 {
		t.Errorf("Expected RequestCount to be 1, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0, got %d", metrics.ErrorCount)
	}

	// Test error request
	observer.OnRequest("test-server", req, 500, time.Millisecond*50)

	metrics = observer.GetMetrics("test-server")
	if metrics.RequestCount != 2 {
		t.Errorf("Expected RequestCount to be 2, got %d", metrics.RequestCount)
	}

	if metrics.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount to be 1, got %d", metrics.ErrorCount)
	}

	// Test that last request time is recorded
	lastReqTime := metrics.GetLastRequestTime()
	if lastReqTime.IsZero() {
		t.Error("Expected last request time to be set")
	}
}

func TestMetricsObserverOnBeforeAndAfterRequest(t *testing.T) {
	observer := NewMetricsObserver()
	observer.OnStart("test-server")

	req, _ := http.NewRequest("GET", "/api/test", nil)

	// Simulate request processing
	observer.OnBeforeRequest("test-server", req)

	metrics := observer.GetMetrics("test-server")
	if metrics.ActiveRequests != 1 {
		t.Errorf("Expected ActiveRequests to be 1, got %d", metrics.ActiveRequests)
	}

	observer.OnAfterRequest("test-server", req, 200, time.Millisecond*100)

	metrics = observer.GetMetrics("test-server")
	if metrics.ActiveRequests != 0 {
		t.Errorf("Expected ActiveRequests to be 0, got %d", metrics.ActiveRequests)
	}
}

func TestMetricsObserverGetMetricsNotFound(t *testing.T) {
	observer := NewMetricsObserver()

	metrics := observer.GetMetrics("nonexistent")
	if metrics != nil {
		t.Error("Expected nil metrics for nonexistent server")
	}
}

func TestMetricsObserverConcurrentRequests(t *testing.T) {
	observer := NewMetricsObserver()
	observer.OnStart("test-server")

	// Run concurrent requests to test race conditions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			req, _ := http.NewRequest("GET", "/api/test", nil)
			observer.OnBeforeRequest("test-server", req)
			time.Sleep(time.Millisecond * 10) // Simulate processing time
			observer.OnRequest("test-server", req, 200, time.Millisecond*10)
			observer.OnAfterRequest("test-server", req, 200, time.Millisecond*10)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := observer.GetMetrics("test-server")
	if metrics == nil {
		t.Fatal("Expected metrics to exist")
	}

	if metrics.RequestCount != 10 {
		t.Errorf("Expected RequestCount to be 10, got %d", metrics.RequestCount)
	}

	if metrics.ActiveRequests != 0 {
		t.Errorf("Expected ActiveRequests to be 0, got %d", metrics.ActiveRequests)
	}

	// Verify last request time is recorded
	lastReqTime := metrics.GetLastRequestTime()
	if lastReqTime.IsZero() {
		t.Error("Expected last request time to be set")
	}
}

func TestMetricsDataGetLastRequestTime(t *testing.T) {
	data := &MetricsData{}

	// Test zero time
	lastTime := data.GetLastRequestTime()
	if !lastTime.IsZero() {
		t.Error("Expected zero time for uninitialized LastRequestTimeNs")
	}

	// Test with actual time
	now := time.Now()
	data.LastRequestTimeNs = now.UnixNano()

	lastTime = data.GetLastRequestTime()
	if lastTime.Unix() != now.Unix() {
		t.Error("Expected GetLastRequestTime to return correct time")
	}
}
