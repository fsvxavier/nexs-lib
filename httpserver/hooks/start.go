// Package hooks provides start hook implementations.
package hooks

import (
	"context"
	"fmt"
	"time"
)

// StartHook handles server start events and provides startup tracking.
type StartHook struct {
	*BaseHook
	metricsEnabled bool
	startTime      time.Time
	startCount     int64
	serverAddr     string
	startupEvents  []StartupEvent
}

// StartupEvent represents a server startup event.
type StartupEvent struct {
	Timestamp time.Time
	Address   string
	Success   bool
	Error     error
}

// NewStartHook creates a new start hook.
func NewStartHook(name string) *StartHook {
	return &StartHook{
		BaseHook:       NewBaseHook(name),
		metricsEnabled: true,
		startupEvents:  make([]StartupEvent, 0),
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *StartHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *StartHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// OnStart handles server start events with detailed tracking.
func (h *StartHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}

	now := time.Now()
	h.startTime = now
	h.serverAddr = addr
	h.startCount++

	h.logger.Info("🚀 Server STARTED successfully on %s (hook: %s)", addr, h.GetName())
	h.logger.Info("   Server startup time: %s", now.Format(time.RFC3339))
	h.logger.Info("   Start count: %d", h.startCount)

	if h.metricsEnabled {
		event := StartupEvent{
			Timestamp: now,
			Address:   addr,
			Success:   true,
			Error:     nil,
		}
		h.startupEvents = append(h.startupEvents, event)
	}

	return nil
}

// OnStop logs server stop events.
func (h *StartHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	uptime := time.Since(h.startTime)
	h.logger.Info("StartHook %s: Server stopped (uptime: %v)", h.GetName(), uptime)
	return nil
}

// OnError logs error events and tracks startup failures.
func (h *StartHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Error("StartHook %s: Error during startup: %v", h.GetName(), err)

	if h.metricsEnabled {
		event := StartupEvent{
			Timestamp: time.Now(),
			Address:   h.serverAddr,
			Success:   false,
			Error:     err,
		}
		h.startupEvents = append(h.startupEvents, event)
	}

	return nil
}

// OnRequest provides basic request logging.
func (h *StartHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("StartHook %s: Request received", h.GetName())
	return nil
}

// OnResponse provides basic response logging.
func (h *StartHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("StartHook %s: Response sent in %v", h.GetName(), duration)
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *StartHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("StartHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// OnRouteExit provides basic route exit logging.
func (h *StartHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("StartHook %s: Exiting route %s (duration: %v)", h.GetName(), routeKey, duration)
	return nil
}

// GetStartTime returns the time when the server was started.
func (h *StartHook) GetStartTime() time.Time {
	return h.startTime
}

// GetUptime returns the duration since the server was started.
func (h *StartHook) GetUptime() time.Duration {
	if h.startTime.IsZero() {
		return 0
	}
	return time.Since(h.startTime)
}

// GetStartCount returns the number of times the server has been started.
func (h *StartHook) GetStartCount() int64 {
	return h.startCount
}

// GetServerAddr returns the address the server is listening on.
func (h *StartHook) GetServerAddr() string {
	return h.serverAddr
}

// GetStartupEvents returns a copy of all startup events.
func (h *StartHook) GetStartupEvents() []StartupEvent {
	// Return a copy to prevent external modification
	copy := make([]StartupEvent, len(h.startupEvents))
	for i, event := range h.startupEvents {
		copy[i] = StartupEvent{
			Timestamp: event.Timestamp,
			Address:   event.Address,
			Success:   event.Success,
			Error:     event.Error,
		}
	}
	return copy
}

// GetSuccessfulStartCount returns the number of successful starts.
func (h *StartHook) GetSuccessfulStartCount() int {
	count := 0
	for _, event := range h.startupEvents {
		if event.Success {
			count++
		}
	}
	return count
}

// GetFailedStartCount returns the number of failed starts.
func (h *StartHook) GetFailedStartCount() int {
	count := 0
	for _, event := range h.startupEvents {
		if !event.Success {
			count++
		}
	}
	return count
}

// GetStartSuccessRate returns the success rate of server starts as a percentage.
func (h *StartHook) GetStartSuccessRate() float64 {
	totalEvents := len(h.startupEvents)
	if totalEvents == 0 {
		return 0.0
	}

	successCount := h.GetSuccessfulStartCount()
	return float64(successCount) / float64(totalEvents) * 100.0
}

// IsServerRunning returns true if the server is currently running.
func (h *StartHook) IsServerRunning() bool {
	return !h.startTime.IsZero() && h.serverAddr != ""
}

// ClearMetrics clears all collected startup metrics.
func (h *StartHook) ClearMetrics() {
	h.startupEvents = make([]StartupEvent, 0)
	h.startCount = 0
}

// ResetStartTime resets the start time (useful for testing).
func (h *StartHook) ResetStartTime() {
	h.startTime = time.Time{}
	h.serverAddr = ""
}
