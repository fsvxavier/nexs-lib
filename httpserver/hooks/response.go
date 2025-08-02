// Package hooks provides response hook implementations.
package hooks

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ResponseHook handles response events and provides response tracking and logging.
type ResponseHook struct {
	*BaseHook
	metricsEnabled      bool
	responseCounter     int64
	totalResponseTime   int64 // in nanoseconds
	totalResponseSize   int64
	errorResponseCount  int64
	statusCodeCounts    map[int]int64
	responseSizeTracker map[interface{}]int64
}

// NewResponseHook creates a new response hook.
func NewResponseHook(name string) *ResponseHook {
	return &ResponseHook{
		BaseHook:            NewBaseHook(name),
		metricsEnabled:      true,
		statusCodeCounts:    make(map[int]int64),
		responseSizeTracker: make(map[interface{}]int64),
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *ResponseHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *ResponseHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// OnResponse handles response events with detailed tracking.
func (h *ResponseHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}

	requestID := fmt.Sprintf("%p", req)
	responseID := fmt.Sprintf("%p", resp)

	h.logger.Info("Response sent (Request ID: %s, Response ID: %s, Duration: %v) (hook: %s)",
		requestID, responseID, duration, h.GetName())

	if h.metricsEnabled {
		// Increment response counter
		atomic.AddInt64(&h.responseCounter, 1)

		// Add to total response time
		atomic.AddInt64(&h.totalResponseTime, int64(duration))

		// Track response size if possible
		if size := h.extractResponseSize(resp); size > 0 {
			h.responseSizeTracker[resp] = size
			atomic.AddInt64(&h.totalResponseSize, size)
		}

		// Track status code if possible
		if statusCode := h.extractStatusCode(resp); statusCode > 0 {
			h.incrementStatusCode(statusCode)

			// Track error responses (4xx and 5xx)
			if statusCode >= 400 {
				atomic.AddInt64(&h.errorResponseCount, 1)
			}
		}
	}

	return nil
}

// extractResponseSize attempts to extract the size of the response.
func (h *ResponseHook) extractResponseSize(resp interface{}) int64 {
	if resp == nil {
		return 0
	}

	// Simple size estimation based on string representation
	respStr := fmt.Sprintf("%v", resp)
	return int64(len(respStr))
}

// extractStatusCode attempts to extract the HTTP status code from the response.
func (h *ResponseHook) extractStatusCode(resp interface{}) int {
	// This is a placeholder implementation
	// In real scenarios, you would examine the response type and extract the status code
	// For example, for HTTP responses you might access StatusCode field

	// For demonstration, we'll return a default status code
	// In practice, this would be provider-specific
	return 200 // Default OK status
}

// incrementStatusCode safely increments the count for a status code.
func (h *ResponseHook) incrementStatusCode(statusCode int) {
	// Since we're using a map, we need to ensure thread safety
	// For simplicity, we'll use a basic approach here
	// In production, you might want to use sync.Map or other concurrent-safe structures
	if count, exists := h.statusCodeCounts[statusCode]; exists {
		h.statusCodeCounts[statusCode] = count + 1
	} else {
		h.statusCodeCounts[statusCode] = 1
	}
}

// OnStart logs server start events.
func (h *ResponseHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Info("ResponseHook %s: Server started on %s", h.GetName(), addr)
	return nil
}

// OnStop logs server stop events and response metrics summary.
func (h *ResponseHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Info("ResponseHook %s: Server stopped", h.GetName())

	if h.metricsEnabled {
		totalResponses := atomic.LoadInt64(&h.responseCounter)
		totalTime := atomic.LoadInt64(&h.totalResponseTime)
		totalSize := atomic.LoadInt64(&h.totalResponseSize)
		errorCount := atomic.LoadInt64(&h.errorResponseCount)

		h.logger.Info("Response Metrics Summary:")
		h.logger.Info("  Total responses: %d", totalResponses)
		h.logger.Info("  Error responses: %d", errorCount)
		h.logger.Info("  Total response size: %d bytes", totalSize)

		if totalResponses > 0 {
			avgTime := time.Duration(totalTime / totalResponses)
			avgSize := totalSize / totalResponses
			errorRate := float64(errorCount) / float64(totalResponses) * 100

			h.logger.Info("  Average response time: %v", avgTime)
			h.logger.Info("  Average response size: %d bytes", avgSize)
			h.logger.Info("  Error rate: %.2f%%", errorRate)
		}

		// Log status code distribution
		if len(h.statusCodeCounts) > 0 {
			h.logger.Info("  Status code distribution:")
			for code, count := range h.statusCodeCounts {
				percentage := float64(count) / float64(totalResponses) * 100
				h.logger.Info("    %d: %d (%.1f%%)", code, count, percentage)
			}
		}
	}

	return nil
}

// OnError logs error events.
func (h *ResponseHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Error("ResponseHook %s: Error occurred: %v", h.GetName(), err)
	return nil
}

// OnRequest provides basic request logging.
func (h *ResponseHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("ResponseHook %s: Request received", h.GetName())
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *ResponseHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("ResponseHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// OnRouteExit provides basic route exit logging.
func (h *ResponseHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("ResponseHook %s: Exiting route %s (duration: %v)", h.GetName(), routeKey, duration)
	return nil
}

// GetResponseCount returns the total number of responses sent.
func (h *ResponseHook) GetResponseCount() int64 {
	return atomic.LoadInt64(&h.responseCounter)
}

// GetErrorResponseCount returns the total number of error responses (4xx and 5xx).
func (h *ResponseHook) GetErrorResponseCount() int64 {
	return atomic.LoadInt64(&h.errorResponseCount)
}

// GetTotalResponseSize returns the total size of all responses sent.
func (h *ResponseHook) GetTotalResponseSize() int64 {
	return atomic.LoadInt64(&h.totalResponseSize)
}

// GetAverageResponseSize returns the average size of responses sent.
func (h *ResponseHook) GetAverageResponseSize() int64 {
	total := atomic.LoadInt64(&h.responseCounter)
	if total == 0 {
		return 0
	}
	totalSize := atomic.LoadInt64(&h.totalResponseSize)
	return totalSize / total
}

// GetAverageResponseTime returns the average response time.
func (h *ResponseHook) GetAverageResponseTime() time.Duration {
	total := atomic.LoadInt64(&h.responseCounter)
	if total == 0 {
		return 0
	}
	totalTime := atomic.LoadInt64(&h.totalResponseTime)
	return time.Duration(totalTime / total)
}

// GetErrorRate returns the error rate as a percentage.
func (h *ResponseHook) GetErrorRate() float64 {
	total := atomic.LoadInt64(&h.responseCounter)
	if total == 0 {
		return 0.0
	}
	errors := atomic.LoadInt64(&h.errorResponseCount)
	return float64(errors) / float64(total) * 100.0
}

// GetStatusCodeCounts returns a copy of the status code counts.
func (h *ResponseHook) GetStatusCodeCounts() map[int]int64 {
	// Return a copy to prevent external modification
	copy := make(map[int]int64)
	for code, count := range h.statusCodeCounts {
		copy[code] = count
	}
	return copy
}

// ResetMetrics resets all collected metrics.
func (h *ResponseHook) ResetMetrics() {
	atomic.StoreInt64(&h.responseCounter, 0)
	atomic.StoreInt64(&h.totalResponseTime, 0)
	atomic.StoreInt64(&h.totalResponseSize, 0)
	atomic.StoreInt64(&h.errorResponseCount, 0)
	h.statusCodeCounts = make(map[int]int64)
	h.responseSizeTracker = make(map[interface{}]int64)
}
