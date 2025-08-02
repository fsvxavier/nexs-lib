// Package hooks provides stop hook implementations.
package hooks

import (
	"context"
	"fmt"
	"time"
)

// StopHook handles server stop events and provides shutdown tracking.
type StopHook struct {
	*BaseHook
	metricsEnabled bool
	stopTime       time.Time
	stopCount      int64
	gracefulStops  int64
	forcedStops    int64
	lastUptime     time.Duration
	shutdownEvents []ShutdownEvent
	startTime      time.Time
}

// ShutdownEvent represents a server shutdown event.
type ShutdownEvent struct {
	Timestamp     time.Time
	Uptime        time.Duration
	GracefulStop  bool
	Error         error
	FinalRequests int64
}

// NewStopHook creates a new stop hook.
func NewStopHook(name string) *StopHook {
	return &StopHook{
		BaseHook:       NewBaseHook(name),
		metricsEnabled: true,
		shutdownEvents: make([]ShutdownEvent, 0),
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *StopHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *StopHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// OnStart tracks server start time for uptime calculation.
func (h *StopHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}

	h.startTime = time.Now()
	h.logger.Info("StopHook %s: Server started on %s (tracking for uptime)", h.GetName(), addr)
	return nil
}

// OnStop handles server stop events with detailed tracking.
func (h *StopHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	now := time.Now()
	h.stopTime = now
	h.stopCount++

	// Calculate uptime
	var uptime time.Duration
	if !h.startTime.IsZero() {
		uptime = now.Sub(h.startTime)
		h.lastUptime = uptime
	}

	h.logger.Info("🛑 Server STOPPED gracefully (hook: %s)", h.GetName())
	h.logger.Info("   Shutdown time: %s", now.Format(time.RFC3339))
	h.logger.Info("   Server uptime: %v", uptime)
	h.logger.Info("   Stop count: %d", h.stopCount)

	if h.metricsEnabled {
		// Assume graceful stop by default (in OnError we'll track forced stops)
		h.gracefulStops++

		event := ShutdownEvent{
			Timestamp:     now,
			Uptime:        uptime,
			GracefulStop:  true,
			Error:         nil,
			FinalRequests: 0, // Could be enhanced to track active requests
		}
		h.shutdownEvents = append(h.shutdownEvents, event)
	}

	return nil
}

// OnError logs error events and tracks forced shutdowns.
func (h *StopHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Error("StopHook %s: Error during shutdown: %v", h.GetName(), err)

	if h.metricsEnabled {
		now := time.Now()

		// Track as forced stop if we haven't recorded this shutdown yet
		h.forcedStops++

		var uptime time.Duration
		if !h.startTime.IsZero() {
			uptime = now.Sub(h.startTime)
		}

		event := ShutdownEvent{
			Timestamp:     now,
			Uptime:        uptime,
			GracefulStop:  false,
			Error:         err,
			FinalRequests: 0,
		}
		h.shutdownEvents = append(h.shutdownEvents, event)
	}

	return nil
}

// OnRequest provides basic request logging.
func (h *StopHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("StopHook %s: Request received", h.GetName())
	return nil
}

// OnResponse provides basic response logging.
func (h *StopHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("StopHook %s: Response sent in %v", h.GetName(), duration)
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *StopHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("StopHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// OnRouteExit provides basic route exit logging.
func (h *StopHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("StopHook %s: Exiting route %s (duration: %v)", h.GetName(), routeKey, duration)
	return nil
}

// GetStopTime returns the time when the server was last stopped.
func (h *StopHook) GetStopTime() time.Time {
	return h.stopTime
}

// GetStopCount returns the number of times the server has been stopped.
func (h *StopHook) GetStopCount() int64 {
	return h.stopCount
}

// GetGracefulStopCount returns the number of graceful stops.
func (h *StopHook) GetGracefulStopCount() int64 {
	return h.gracefulStops
}

// GetForcedStopCount returns the number of forced stops.
func (h *StopHook) GetForcedStopCount() int64 {
	return h.forcedStops
}

// GetLastUptime returns the uptime from the last server run.
func (h *StopHook) GetLastUptime() time.Duration {
	return h.lastUptime
}

// GetShutdownEvents returns a copy of all shutdown events.
func (h *StopHook) GetShutdownEvents() []ShutdownEvent {
	// Return a copy to prevent external modification
	copy := make([]ShutdownEvent, len(h.shutdownEvents))
	for i, event := range h.shutdownEvents {
		copy[i] = ShutdownEvent{
			Timestamp:     event.Timestamp,
			Uptime:        event.Uptime,
			GracefulStop:  event.GracefulStop,
			Error:         event.Error,
			FinalRequests: event.FinalRequests,
		}
	}
	return copy
}

// GetGracefulStopRate returns the rate of graceful stops as a percentage.
func (h *StopHook) GetGracefulStopRate() float64 {
	total := h.gracefulStops + h.forcedStops
	if total == 0 {
		return 0.0
	}
	return float64(h.gracefulStops) / float64(total) * 100.0
}

// GetAverageUptime returns the average uptime across all server runs.
func (h *StopHook) GetAverageUptime() time.Duration {
	if len(h.shutdownEvents) == 0 {
		return 0
	}

	var totalUptime time.Duration
	for _, event := range h.shutdownEvents {
		totalUptime += event.Uptime
	}

	return time.Duration(int64(totalUptime) / int64(len(h.shutdownEvents)))
}

// GetTotalUptime returns the cumulative uptime across all server runs.
func (h *StopHook) GetTotalUptime() time.Duration {
	var totalUptime time.Duration
	for _, event := range h.shutdownEvents {
		totalUptime += event.Uptime
	}
	return totalUptime
}

// GetMaxUptime returns the longest uptime recorded.
func (h *StopHook) GetMaxUptime() time.Duration {
	var maxUptime time.Duration
	for _, event := range h.shutdownEvents {
		if event.Uptime > maxUptime {
			maxUptime = event.Uptime
		}
	}
	return maxUptime
}

// GetMinUptime returns the shortest uptime recorded.
func (h *StopHook) GetMinUptime() time.Duration {
	if len(h.shutdownEvents) == 0 {
		return 0
	}

	minUptime := h.shutdownEvents[0].Uptime
	for _, event := range h.shutdownEvents {
		if event.Uptime < minUptime {
			minUptime = event.Uptime
		}
	}
	return minUptime
}

// IsServerStopped returns true if the server has been stopped.
func (h *StopHook) IsServerStopped() bool {
	return !h.stopTime.IsZero()
}

// ClearMetrics clears all collected shutdown metrics.
func (h *StopHook) ClearMetrics() {
	h.shutdownEvents = make([]ShutdownEvent, 0)
	h.stopCount = 0
	h.gracefulStops = 0
	h.forcedStops = 0
}

// ResetStopTime resets the stop time (useful for testing).
func (h *StopHook) ResetStopTime() {
	h.stopTime = time.Time{}
}

// SetStartTime sets the start time manually (useful for testing).
func (h *StopHook) SetStartTime(startTime time.Time) {
	h.startTime = startTime
}
