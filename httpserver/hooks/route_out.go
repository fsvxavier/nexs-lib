// Package hooks provides route exit hook implementations.
package hooks

import (
	"context"
	"fmt"
	"time"
)

// RouteOutHook handles route exit events and provides detailed performance tracking.
type RouteOutHook struct {
	*BaseHook
	metricsEnabled   bool
	performanceTrack map[string]*PerformanceMetrics
	slowThreshold    time.Duration
}

// PerformanceMetrics contains performance metrics for a specific route.
type PerformanceMetrics struct {
	ExitCount     int64
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	AvgDuration   time.Duration
	SlowRequests  int64
	LastExitTime  time.Time
}

// NewRouteOutHook creates a new route exit hook.
func NewRouteOutHook(name string) *RouteOutHook {
	return &RouteOutHook{
		BaseHook:         NewBaseHook(name),
		metricsEnabled:   true,
		performanceTrack: make(map[string]*PerformanceMetrics),
		slowThreshold:    time.Second * 1, // Default 1 second threshold
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *RouteOutHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *RouteOutHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// SetSlowThreshold sets the duration threshold for considering a request as slow.
func (h *RouteOutHook) SetSlowThreshold(threshold time.Duration) {
	h.slowThreshold = threshold
}

// GetSlowThreshold returns the current slow request threshold.
func (h *RouteOutHook) GetSlowThreshold() time.Duration {
	return h.slowThreshold
}

// OnRouteExit handles route exit events with enhanced performance tracking.
func (h *RouteOutHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}

	routeKey := fmt.Sprintf("%s %s", method, path)
	now := time.Now()

	// Log route exit
	if duration > h.slowThreshold {
		h.logger.Warn("SLOW REQUEST: Route %s completed in %v (threshold: %v) (hook: %s)",
			routeKey, duration, h.slowThreshold, h.GetName())
	} else {
		h.logger.Info("Route %s completed in %v (hook: %s)", routeKey, duration, h.GetName())
	}

	// Collect performance metrics if enabled
	if h.metricsEnabled {
		h.updatePerformanceMetrics(routeKey, duration, now)
	}

	return nil
}

// updatePerformanceMetrics updates the performance metrics for a route.
func (h *RouteOutHook) updatePerformanceMetrics(routeKey string, duration time.Duration, exitTime time.Time) {
	metrics, exists := h.performanceTrack[routeKey]
	if !exists {
		// Initialize new metrics
		h.performanceTrack[routeKey] = &PerformanceMetrics{
			ExitCount:     1,
			TotalDuration: duration,
			MinDuration:   duration,
			MaxDuration:   duration,
			AvgDuration:   duration,
			SlowRequests: func() int64 {
				if duration > h.slowThreshold {
					return 1
				}
				return 0
			}(),
			LastExitTime: exitTime,
		}
		return
	}

	// Update existing metrics
	metrics.ExitCount++
	metrics.TotalDuration += duration
	metrics.LastExitTime = exitTime

	// Update min/max durations
	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}

	// Update average duration
	metrics.AvgDuration = time.Duration(int64(metrics.TotalDuration) / metrics.ExitCount)

	// Track slow requests
	if duration > h.slowThreshold {
		metrics.SlowRequests++
	}
}

// OnStart logs server start events.
func (h *RouteOutHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Info("RouteOutHook %s: Server started on %s", h.GetName(), addr)
	return nil
}

// OnStop logs server stop events and performance summary.
func (h *RouteOutHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Info("RouteOutHook %s: Server stopped", h.GetName())

	if h.metricsEnabled && len(h.performanceTrack) > 0 {
		h.logger.Info("Performance Metrics Summary:")
		for route, metrics := range h.performanceTrack {
			slowPercent := float64(metrics.SlowRequests) / float64(metrics.ExitCount) * 100
			h.logger.Info("  %s: %d exits, avg: %v, min: %v, max: %v, slow: %.1f%%",
				route, metrics.ExitCount, metrics.AvgDuration, metrics.MinDuration,
				metrics.MaxDuration, slowPercent)
		}
	}

	return nil
}

// OnError logs error events.
func (h *RouteOutHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Error("RouteOutHook %s: Error occurred: %v", h.GetName(), err)
	return nil
}

// OnRequest provides basic request logging.
func (h *RouteOutHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("RouteOutHook %s: Request received", h.GetName())
	return nil
}

// OnResponse provides basic response logging.
func (h *RouteOutHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("RouteOutHook %s: Response sent in %v", h.GetName(), duration)
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *RouteOutHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("RouteOutHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// GetPerformanceMetrics returns the performance metrics for a specific route.
func (h *RouteOutHook) GetPerformanceMetrics(method, path string) (*PerformanceMetrics, bool) {
	routeKey := fmt.Sprintf("%s %s", method, path)
	metrics, exists := h.performanceTrack[routeKey]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	return &PerformanceMetrics{
		ExitCount:     metrics.ExitCount,
		TotalDuration: metrics.TotalDuration,
		MinDuration:   metrics.MinDuration,
		MaxDuration:   metrics.MaxDuration,
		AvgDuration:   metrics.AvgDuration,
		SlowRequests:  metrics.SlowRequests,
		LastExitTime:  metrics.LastExitTime,
	}, true
}

// GetAllPerformanceMetrics returns all performance metrics.
func (h *RouteOutHook) GetAllPerformanceMetrics() map[string]*PerformanceMetrics {
	// Return a copy to prevent external modification
	copy := make(map[string]*PerformanceMetrics)
	for key, metrics := range h.performanceTrack {
		copy[key] = &PerformanceMetrics{
			ExitCount:     metrics.ExitCount,
			TotalDuration: metrics.TotalDuration,
			MinDuration:   metrics.MinDuration,
			MaxDuration:   metrics.MaxDuration,
			AvgDuration:   metrics.AvgDuration,
			SlowRequests:  metrics.SlowRequests,
			LastExitTime:  metrics.LastExitTime,
		}
	}
	return copy
}

// ClearMetrics clears all collected performance metrics.
func (h *RouteOutHook) ClearMetrics() {
	h.performanceTrack = make(map[string]*PerformanceMetrics)
}

// GetMetricsCount returns the number of routes being tracked.
func (h *RouteOutHook) GetMetricsCount() int {
	return len(h.performanceTrack)
}

// GetSlowRequestsCount returns the total number of slow requests across all routes.
func (h *RouteOutHook) GetSlowRequestsCount() int64 {
	var total int64
	for _, metrics := range h.performanceTrack {
		total += metrics.SlowRequests
	}
	return total
}

// GetTotalRequestsCount returns the total number of requests across all routes.
func (h *RouteOutHook) GetTotalRequestsCount() int64 {
	var total int64
	for _, metrics := range h.performanceTrack {
		total += metrics.ExitCount
	}
	return total
}

// GetSlowRequestPercentage returns the percentage of slow requests across all routes.
func (h *RouteOutHook) GetSlowRequestPercentage() float64 {
	total := h.GetTotalRequestsCount()
	if total == 0 {
		return 0.0
	}
	slow := h.GetSlowRequestsCount()
	return float64(slow) / float64(total) * 100.0
}
