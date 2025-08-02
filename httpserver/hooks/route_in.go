// Package hooks provides route entry hook implementations.
package hooks

import (
	"context"
	"fmt"
	"time"
)

// RouteInHook handles route entry events and provides detailed logging and metrics.
type RouteInHook struct {
	*BaseHook
	metricsEnabled bool
	routeMetrics   map[string]*RouteMetrics
}

// RouteMetrics contains metrics for a specific route.
type RouteMetrics struct {
	EntryCount    int64
	LastEntryTime time.Time
	TotalDuration time.Duration
	AvgDuration   time.Duration
}

// NewRouteInHook creates a new route entry hook.
func NewRouteInHook(name string) *RouteInHook {
	return &RouteInHook{
		BaseHook:       NewBaseHook(name),
		metricsEnabled: true,
		routeMetrics:   make(map[string]*RouteMetrics),
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *RouteInHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *RouteInHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// OnRouteEnter handles route entry events with enhanced logging and metrics.
func (h *RouteInHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}

	routeKey := fmt.Sprintf("%s %s", method, path)
	now := time.Now()

	h.logger.Info("Entering route: %s (hook: %s)", routeKey, h.GetName())

	// Collect metrics if enabled
	if h.metricsEnabled {
		if metrics, exists := h.routeMetrics[routeKey]; exists {
			metrics.EntryCount++
			metrics.LastEntryTime = now
		} else {
			h.routeMetrics[routeKey] = &RouteMetrics{
				EntryCount:    1,
				LastEntryTime: now,
			}
		}
	}

	// Store entry time in context for duration calculation
	if ctx != nil {
		ctx = context.WithValue(ctx, "route_entry_time", now)
	}

	return nil
}

// OnStart logs server start events.
func (h *RouteInHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Info("RouteInHook %s: Server started on %s", h.GetName(), addr)
	return nil
}

// OnStop logs server stop events and optionally prints metrics.
func (h *RouteInHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Info("RouteInHook %s: Server stopped", h.GetName())

	if h.metricsEnabled && len(h.routeMetrics) > 0 {
		h.logger.Info("Route Entry Metrics Summary:")
		for route, metrics := range h.routeMetrics {
			h.logger.Info("  %s: %d entries, last entry: %v",
				route, metrics.EntryCount, metrics.LastEntryTime.Format(time.RFC3339))
		}
	}

	return nil
}

// OnError logs error events.
func (h *RouteInHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Error("RouteInHook %s: Error occurred: %v", h.GetName(), err)
	return nil
}

// OnRequest provides basic request logging.
func (h *RouteInHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("RouteInHook %s: Request received", h.GetName())
	return nil
}

// OnResponse provides basic response logging.
func (h *RouteInHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("RouteInHook %s: Response sent in %v", h.GetName(), duration)
	return nil
}

// OnRouteExit handles route exit events but focuses on entry metrics.
func (h *RouteInHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}

	routeKey := fmt.Sprintf("%s %s", method, path)

	// Update metrics with route duration if metrics are enabled
	if h.metricsEnabled {
		if metrics, exists := h.routeMetrics[routeKey]; exists {
			metrics.TotalDuration += duration
			metrics.AvgDuration = time.Duration(int64(metrics.TotalDuration) / metrics.EntryCount)
		}
	}

	h.logger.Debug("RouteInHook %s: Route %s completed in %v", h.GetName(), routeKey, duration)
	return nil
}

// GetRouteMetrics returns the metrics for a specific route.
func (h *RouteInHook) GetRouteMetrics(method, path string) (*RouteMetrics, bool) {
	routeKey := fmt.Sprintf("%s %s", method, path)
	metrics, exists := h.routeMetrics[routeKey]
	return metrics, exists
}

// GetAllRouteMetrics returns all route metrics.
func (h *RouteInHook) GetAllRouteMetrics() map[string]*RouteMetrics {
	// Return a copy to prevent external modification
	copy := make(map[string]*RouteMetrics)
	for key, metrics := range h.routeMetrics {
		copy[key] = &RouteMetrics{
			EntryCount:    metrics.EntryCount,
			LastEntryTime: metrics.LastEntryTime,
			TotalDuration: metrics.TotalDuration,
			AvgDuration:   metrics.AvgDuration,
		}
	}
	return copy
}

// ClearMetrics clears all collected metrics.
func (h *RouteInHook) ClearMetrics() {
	h.routeMetrics = make(map[string]*RouteMetrics)
}

// GetMetricsCount returns the number of routes being tracked.
func (h *RouteInHook) GetMetricsCount() int {
	return len(h.routeMetrics)
}
