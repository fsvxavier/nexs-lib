// Package hooks provides request hook implementations.
package hooks

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RequestHook handles request events and provides request tracking and logging.
type RequestHook struct {
	*BaseHook
	mu                 sync.Mutex
	metricsEnabled     bool
	requestCounter     int64
	activeRequests     int64
	maxActiveRequests  int64
	totalRequestsSize  int64
	requestStartTimes  map[interface{}]time.Time
	requestSizeTracker map[interface{}]int64
}

// NewRequestHook creates a new request hook.
func NewRequestHook(name string) *RequestHook {
	return &RequestHook{
		BaseHook:           NewBaseHook(name),
		metricsEnabled:     true,
		requestStartTimes:  make(map[interface{}]time.Time),
		requestSizeTracker: make(map[interface{}]int64),
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *RequestHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *RequestHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// OnRequest handles incoming request events with detailed tracking.
func (h *RequestHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}

	requestID := fmt.Sprintf("%p", req)
	now := time.Now()

	h.logger.Info("Request received (ID: %s) (hook: %s)", requestID, h.GetName())

	if h.metricsEnabled {
		// Increment counters
		atomic.AddInt64(&h.requestCounter, 1)
		currentActive := atomic.AddInt64(&h.activeRequests, 1)

		// Update max active requests
		for {
			maxActive := atomic.LoadInt64(&h.maxActiveRequests)
			if currentActive <= maxActive || atomic.CompareAndSwapInt64(&h.maxActiveRequests, maxActive, currentActive) {
				break
			}
		}

		// Store request start time
		h.mu.Lock()
		h.requestStartTimes[req] = now
		h.mu.Unlock()

		// Track request size if possible
		if size := h.extractRequestSize(req); size > 0 {
			h.mu.Lock()
			h.requestSizeTracker[req] = size
			h.mu.Unlock()
			atomic.AddInt64(&h.totalRequestsSize, size)
		}
	}

	return nil
}

// OnResponse handles response events and calculates request duration.
func (h *RequestHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}

	requestID := fmt.Sprintf("%p", req)
	h.logger.Debug("Request completed (ID: %s, duration: %v) (hook: %s)", requestID, duration, h.GetName())

	if h.metricsEnabled {
		// Decrement active requests
		atomic.AddInt64(&h.activeRequests, -1)

		// Clean up tracking data
		h.mu.Lock()
		delete(h.requestStartTimes, req)
		delete(h.requestSizeTracker, req)
		h.mu.Unlock()
	}

	return nil
}

// extractRequestSize attempts to extract the size of the request.
// This is a simple implementation that can be extended based on request types.
func (h *RequestHook) extractRequestSize(req interface{}) int64 {
	// This is a placeholder implementation
	// In real scenarios, you would examine the request type and extract content length
	// For example, for HTTP requests you might look at Content-Length header

	if req == nil {
		return 0
	}

	// Simple size estimation based on string representation
	reqStr := fmt.Sprintf("%v", req)
	return int64(len(reqStr))
}

// OnStart logs server start events.
func (h *RequestHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Info("RequestHook %s: Server started on %s", h.GetName(), addr)
	return nil
}

// OnStop logs server stop events and request metrics summary.
func (h *RequestHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Info("RequestHook %s: Server stopped", h.GetName())

	if h.metricsEnabled {
		totalRequests := atomic.LoadInt64(&h.requestCounter)
		activeRequests := atomic.LoadInt64(&h.activeRequests)
		maxActive := atomic.LoadInt64(&h.maxActiveRequests)
		totalSize := atomic.LoadInt64(&h.totalRequestsSize)

		h.logger.Info("Request Metrics Summary:")
		h.logger.Info("  Total requests: %d", totalRequests)
		h.logger.Info("  Active requests: %d", activeRequests)
		h.logger.Info("  Max concurrent requests: %d", maxActive)
		h.logger.Info("  Total request size: %d bytes", totalSize)

		if totalRequests > 0 {
			avgSize := totalSize / totalRequests
			h.logger.Info("  Average request size: %d bytes", avgSize)
		}
	}

	return nil
}

// OnError logs error events.
func (h *RequestHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Error("RequestHook %s: Error occurred: %v", h.GetName(), err)
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *RequestHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("RequestHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// OnRouteExit provides basic route exit logging.
func (h *RequestHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("RequestHook %s: Exiting route %s (duration: %v)", h.GetName(), routeKey, duration)
	return nil
}

// GetRequestCount returns the total number of requests processed.
func (h *RequestHook) GetRequestCount() int64 {
	return atomic.LoadInt64(&h.requestCounter)
}

// GetActiveRequestCount returns the current number of active requests.
func (h *RequestHook) GetActiveRequestCount() int64 {
	return atomic.LoadInt64(&h.activeRequests)
}

// GetMaxActiveRequestCount returns the maximum number of concurrent requests seen.
func (h *RequestHook) GetMaxActiveRequestCount() int64 {
	return atomic.LoadInt64(&h.maxActiveRequests)
}

// GetTotalRequestSize returns the total size of all requests processed.
func (h *RequestHook) GetTotalRequestSize() int64 {
	return atomic.LoadInt64(&h.totalRequestsSize)
}

// GetAverageRequestSize returns the average size of requests processed.
func (h *RequestHook) GetAverageRequestSize() int64 {
	total := atomic.LoadInt64(&h.requestCounter)
	if total == 0 {
		return 0
	}
	totalSize := atomic.LoadInt64(&h.totalRequestsSize)
	return totalSize / total
}

// ResetMetrics resets all collected metrics.
func (h *RequestHook) ResetMetrics() {
	atomic.StoreInt64(&h.requestCounter, 0)
	atomic.StoreInt64(&h.activeRequests, 0)
	atomic.StoreInt64(&h.maxActiveRequests, 0)
	atomic.StoreInt64(&h.totalRequestsSize, 0)
	h.mu.Lock()
	h.requestStartTimes = make(map[interface{}]time.Time)
	h.requestSizeTracker = make(map[interface{}]int64)
	h.mu.Unlock()
}
