// Package hooks provides error hook implementations.
package hooks

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ErrorHook handles error events and provides comprehensive error tracking.
type ErrorHook struct {
	*BaseHook
	metricsEnabled  bool
	errorCounter    int64
	criticalErrors  int64
	warningErrors   int64
	infoErrors      int64
	errorByType     map[string]int64
	recentErrors    []ErrorEvent
	maxRecentErrors int
	errorThreshold  int64
	lastErrorTime   time.Time
	firstErrorTime  time.Time
}

// ErrorEvent represents an error event with detailed information.
type ErrorEvent struct {
	Timestamp   time.Time
	Error       error
	ErrorType   string
	Severity    ErrorSeverity
	Context     map[string]interface{}
	Recoverable bool
}

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity int

const (
	// SeverityInfo represents informational errors.
	SeverityInfo ErrorSeverity = iota
	// SeverityWarning represents warning errors.
	SeverityWarning
	// SeverityCritical represents critical errors.
	SeverityCritical
)

// String returns the string representation of ErrorSeverity.
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// NewErrorHook creates a new error hook.
func NewErrorHook(name string) *ErrorHook {
	return &ErrorHook{
		BaseHook:        NewBaseHook(name),
		metricsEnabled:  true,
		errorByType:     make(map[string]int64),
		recentErrors:    make([]ErrorEvent, 0),
		maxRecentErrors: 100, // Keep last 100 errors
		errorThreshold:  10,  // Alert after 10 errors
	}
}

// SetMetricsEnabled enables or disables metrics collection.
func (h *ErrorHook) SetMetricsEnabled(enabled bool) {
	h.metricsEnabled = enabled
}

// IsMetricsEnabled returns true if metrics collection is enabled.
func (h *ErrorHook) IsMetricsEnabled() bool {
	return h.metricsEnabled
}

// SetMaxRecentErrors sets the maximum number of recent errors to keep.
func (h *ErrorHook) SetMaxRecentErrors(max int) {
	h.maxRecentErrors = max
}

// SetErrorThreshold sets the error count threshold for alerts.
func (h *ErrorHook) SetErrorThreshold(threshold int64) {
	h.errorThreshold = threshold
}

// OnError handles error events with detailed tracking and classification.
func (h *ErrorHook) OnError(ctx context.Context, err error) error {
	if !h.IsEnabled() {
		return nil
	}

	now := time.Now()
	h.lastErrorTime = now

	// Set first error time if not set
	if h.firstErrorTime.IsZero() {
		h.firstErrorTime = now
	}

	// Classify error severity
	severity := h.classifyError(err)
	errorType := h.getErrorType(err)

	// Log with appropriate severity
	switch severity {
	case SeverityCritical:
		h.logger.Error("🔥 CRITICAL ERROR: %v (hook: %s)", err, h.GetName())
	case SeverityWarning:
		h.logger.Warn("⚠️  WARNING ERROR: %v (hook: %s)", err, h.GetName())
	case SeverityInfo:
		h.logger.Info("ℹ️  INFO ERROR: %v (hook: %s)", err, h.GetName())
	}

	if h.metricsEnabled {
		// Increment counters
		errorCount := atomic.AddInt64(&h.errorCounter, 1)

		switch severity {
		case SeverityCritical:
			atomic.AddInt64(&h.criticalErrors, 1)
		case SeverityWarning:
			atomic.AddInt64(&h.warningErrors, 1)
		case SeverityInfo:
			atomic.AddInt64(&h.infoErrors, 1)
		}

		// Track error by type
		h.incrementErrorType(errorType)

		// Create error event
		event := ErrorEvent{
			Timestamp:   now,
			Error:       err,
			ErrorType:   errorType,
			Severity:    severity,
			Context:     h.extractContextFromContext(ctx),
			Recoverable: h.isRecoverableError(err),
		}

		// Add to recent errors (with circular buffer behavior)
		h.addRecentError(event)

		// Check threshold and log alert if needed
		if errorCount > h.errorThreshold && (errorCount-1) == h.errorThreshold {
			h.logger.Error("🚨 ERROR THRESHOLD ALERT: %d errors reached! (hook: %s)", errorCount, h.GetName())
		}
	}

	return nil
}

// classifyError determines the severity of an error.
func (h *ErrorHook) classifyError(err error) ErrorSeverity {
	if err == nil {
		return SeverityInfo
	}

	errorStr := err.Error()

	// Critical errors
	if containsAny(errorStr, []string{
		"panic", "fatal", "crash", "out of memory", "segmentation fault",
		"database connection lost", "network unreachable", "timeout exceeded",
	}) {
		return SeverityCritical
	}

	// Warning errors
	if containsAny(errorStr, []string{
		"deprecated", "slow", "retry", "fallback", "degraded",
		"high latency", "connection refused", "temporary failure", "timeout",
	}) {
		return SeverityWarning
	}

	// Default to info
	return SeverityInfo
}

// getErrorType extracts the type of error.
func (h *ErrorHook) getErrorType(err error) string {
	if err == nil {
		return "unknown"
	}

	errorStr := err.Error()

	// Common error patterns
	if containsAny(errorStr, []string{"connection", "network", "tcp", "http"}) {
		return "network"
	}
	if containsAny(errorStr, []string{"database", "sql", "query", "transaction"}) {
		return "database"
	}
	if containsAny(errorStr, []string{"permission", "unauthorized", "forbidden", "access"}) {
		return "authorization"
	}
	if containsAny(errorStr, []string{"timeout", "deadline", "context canceled"}) {
		return "timeout"
	}
	if containsAny(errorStr, []string{"validation", "invalid", "bad request", "malformed"}) {
		return "validation"
	}

	return "general"
}

// isRecoverableError determines if an error is recoverable.
func (h *ErrorHook) isRecoverableError(err error) bool {
	if err == nil {
		return true
	}

	errorStr := err.Error()

	// Non-recoverable errors
	if containsAny(errorStr, []string{
		"panic", "fatal", "out of memory", "segmentation fault",
	}) {
		return false
	}

	// Most errors are considered recoverable
	return true
}

// extractContextFromContext extracts relevant context information.
func (h *ErrorHook) extractContextFromContext(ctx context.Context) map[string]interface{} {
	contextData := make(map[string]interface{})

	if ctx == nil {
		return contextData
	}

	// Extract common context values if they exist
	if deadline, ok := ctx.Deadline(); ok {
		contextData["deadline"] = deadline
	}

	if ctx.Err() != nil {
		contextData["context_error"] = ctx.Err().Error()
	}

	return contextData
}

// incrementErrorType safely increments the count for an error type.
func (h *ErrorHook) incrementErrorType(errorType string) {
	if count, exists := h.errorByType[errorType]; exists {
		h.errorByType[errorType] = count + 1
	} else {
		h.errorByType[errorType] = 1
	}
}

// addRecentError adds an error to the recent errors list with circular buffer behavior.
func (h *ErrorHook) addRecentError(event ErrorEvent) {
	if len(h.recentErrors) >= h.maxRecentErrors {
		// Remove oldest error (shift left)
		copy(h.recentErrors, h.recentErrors[1:])
		h.recentErrors[len(h.recentErrors)-1] = event
	} else {
		h.recentErrors = append(h.recentErrors, event)
	}
}

// containsAny checks if the string contains any of the given substrings.
func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(str, substr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(str, substr string) bool {
	// Simple case-insensitive contains check
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				anySubstring(str, substr)))
}

// anySubstring checks if any substring of str matches substr.
func anySubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// OnStart logs server start events.
func (h *ErrorHook) OnStart(ctx context.Context, addr string) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Info("ErrorHook %s: Server started on %s", h.GetName(), addr)
	return nil
}

// OnStop logs server stop events and error summary.
func (h *ErrorHook) OnStop(ctx context.Context) error {
	if !h.IsEnabled() {
		return nil
	}

	h.logger.Info("ErrorHook %s: Server stopped", h.GetName())

	if h.metricsEnabled {
		h.logErrorSummary()
	}

	return nil
}

// logErrorSummary logs a comprehensive error summary.
func (h *ErrorHook) logErrorSummary() {
	totalErrors := atomic.LoadInt64(&h.errorCounter)
	criticalErrors := atomic.LoadInt64(&h.criticalErrors)
	warningErrors := atomic.LoadInt64(&h.warningErrors)
	infoErrors := atomic.LoadInt64(&h.infoErrors)

	h.logger.Info("Error Summary:")
	h.logger.Info("  Total errors: %d", totalErrors)
	h.logger.Info("  Critical errors: %d", criticalErrors)
	h.logger.Info("  Warning errors: %d", warningErrors)
	h.logger.Info("  Info errors: %d", infoErrors)

	if totalErrors > 0 {
		criticalRate := float64(criticalErrors) / float64(totalErrors) * 100
		h.logger.Info("  Critical error rate: %.2f%%", criticalRate)

		// Error frequency
		if !h.firstErrorTime.IsZero() && !h.lastErrorTime.IsZero() {
			duration := h.lastErrorTime.Sub(h.firstErrorTime)
			if duration > 0 {
				frequency := float64(totalErrors) / duration.Hours()
				h.logger.Info("  Error frequency: %.2f errors/hour", frequency)
			}
		}
	}

	// Error type distribution
	if len(h.errorByType) > 0 {
		h.logger.Info("  Error type distribution:")
		for errorType, count := range h.errorByType {
			percentage := float64(count) / float64(totalErrors) * 100
			h.logger.Info("    %s: %d (%.1f%%)", errorType, count, percentage)
		}
	}
}

// OnRequest provides basic request logging.
func (h *ErrorHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("ErrorHook %s: Request received", h.GetName())
	return nil
}

// OnResponse provides basic response logging.
func (h *ErrorHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	h.logger.Debug("ErrorHook %s: Response sent in %v", h.GetName(), duration)
	return nil
}

// OnRouteEnter provides basic route entry logging.
func (h *ErrorHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("ErrorHook %s: Entering route %s", h.GetName(), routeKey)
	return nil
}

// OnRouteExit provides basic route exit logging.
func (h *ErrorHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.IsEnabled() {
		return nil
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	h.logger.Debug("ErrorHook %s: Exiting route %s (duration: %v)", h.GetName(), routeKey, duration)
	return nil
}

// Metrics getters

// GetErrorCount returns the total number of errors.
func (h *ErrorHook) GetErrorCount() int64 {
	return atomic.LoadInt64(&h.errorCounter)
}

// GetCriticalErrorCount returns the number of critical errors.
func (h *ErrorHook) GetCriticalErrorCount() int64 {
	return atomic.LoadInt64(&h.criticalErrors)
}

// GetWarningErrorCount returns the number of warning errors.
func (h *ErrorHook) GetWarningErrorCount() int64 {
	return atomic.LoadInt64(&h.warningErrors)
}

// GetInfoErrorCount returns the number of info errors.
func (h *ErrorHook) GetInfoErrorCount() int64 {
	return atomic.LoadInt64(&h.infoErrors)
}

// GetErrorByType returns a copy of error counts by type.
func (h *ErrorHook) GetErrorByType() map[string]int64 {
	copy := make(map[string]int64)
	for errorType, count := range h.errorByType {
		copy[errorType] = count
	}
	return copy
}

// GetRecentErrors returns a copy of recent errors.
func (h *ErrorHook) GetRecentErrors() []ErrorEvent {
	copy := make([]ErrorEvent, len(h.recentErrors))
	for i, event := range h.recentErrors {
		copy[i] = ErrorEvent{
			Timestamp:   event.Timestamp,
			Error:       event.Error,
			ErrorType:   event.ErrorType,
			Severity:    event.Severity,
			Context:     event.Context,
			Recoverable: event.Recoverable,
		}
	}
	return copy
}

// GetCriticalErrorRate returns the critical error rate as a percentage.
func (h *ErrorHook) GetCriticalErrorRate() float64 {
	total := atomic.LoadInt64(&h.errorCounter)
	if total == 0 {
		return 0.0
	}
	critical := atomic.LoadInt64(&h.criticalErrors)
	return float64(critical) / float64(total) * 100.0
}

// GetErrorFrequency returns the error frequency in errors per hour.
func (h *ErrorHook) GetErrorFrequency() float64 {
	if h.firstErrorTime.IsZero() || h.lastErrorTime.IsZero() {
		return 0.0
	}

	duration := h.lastErrorTime.Sub(h.firstErrorTime)
	if duration <= 0 {
		return 0.0
	}

	totalErrors := atomic.LoadInt64(&h.errorCounter)
	return float64(totalErrors) / duration.Hours()
}

// GetLastErrorTime returns the time of the last error.
func (h *ErrorHook) GetLastErrorTime() time.Time {
	return h.lastErrorTime
}

// GetFirstErrorTime returns the time of the first error.
func (h *ErrorHook) GetFirstErrorTime() time.Time {
	return h.firstErrorTime
}

// ResetMetrics resets all collected error metrics.
func (h *ErrorHook) ResetMetrics() {
	atomic.StoreInt64(&h.errorCounter, 0)
	atomic.StoreInt64(&h.criticalErrors, 0)
	atomic.StoreInt64(&h.warningErrors, 0)
	atomic.StoreInt64(&h.infoErrors, 0)
	h.errorByType = make(map[string]int64)
	h.recentErrors = make([]ErrorEvent, 0)
	h.lastErrorTime = time.Time{}
	h.firstErrorTime = time.Time{}
}
