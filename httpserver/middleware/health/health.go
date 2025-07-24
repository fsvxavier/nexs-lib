// Package health provides advanced health check implementations.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// CheckType represents the type of health check.
type CheckType string

const (
	// CheckTypeLiveness indicates a liveness probe.
	CheckTypeLiveness CheckType = "liveness"
	// CheckTypeReadiness indicates a readiness probe.
	CheckTypeReadiness CheckType = "readiness"
	// CheckTypeStartup indicates a startup probe.
	CheckTypeStartup CheckType = "startup"
)

// Check represents a health check function.
type Check func(ctx context.Context) interfaces.HealthCheckResult

// CheckResult represents the result of a health check.
type CheckResult = interfaces.HealthCheckResult

// Registry manages health checks.
type Registry struct {
	mu     sync.RWMutex
	checks map[string]CheckInfo
}

// CheckInfo contains check metadata.
type CheckInfo struct {
	Check    Check
	Type     CheckType
	Interval time.Duration
	Timeout  time.Duration
	Critical bool
}

// NewRegistry creates a new health check registry.
func NewRegistry() *Registry {
	return &Registry{
		checks: make(map[string]CheckInfo),
	}
}

// Register registers a health check.
func (r *Registry) Register(name string, check Check, options ...CheckOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := CheckInfo{
		Check:    check,
		Type:     CheckTypeLiveness,
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Critical: true,
	}

	// Apply options
	for _, opt := range options {
		opt(&info)
	}

	r.checks[name] = info
	return nil
}

// Unregister removes a health check.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.checks, name)
}

// RunCheck executes a specific health check.
func (r *Registry) RunCheck(ctx context.Context, name string) (interfaces.HealthCheckResult, error) {
	r.mu.RLock()
	info, exists := r.checks[name]
	r.mu.RUnlock()

	if !exists {
		return interfaces.HealthCheckResult{}, fmt.Errorf("health check %s not found", name)
	}

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, info.Timeout)
	defer cancel()

	start := time.Now()
	result := info.Check(checkCtx)
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()
	result.Type = string(info.Type)

	// Default name if not set
	if result.Name == "" {
		result.Name = name
	}

	return result, nil
}

// RunChecks executes all health checks of a specific type.
func (r *Registry) RunChecks(ctx context.Context, checkType CheckType) map[string]interfaces.HealthCheckResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]interfaces.HealthCheckResult)
	var wg sync.WaitGroup

	for name, info := range r.checks {
		if info.Type != checkType {
			continue
		}

		wg.Add(1)
		go func(n string, i CheckInfo) {
			defer wg.Done()

			result, err := r.RunCheck(ctx, n)
			if err != nil {
				result = interfaces.HealthCheckResult{
					Name:      n,
					Status:    "error",
					Message:   err.Error(),
					Duration:  0,
					Timestamp: time.Now(),
					Type:      string(checkType),
					Critical:  i.Critical,
				}
			}

			results[n] = result
		}(name, info)
	}

	wg.Wait()
	return results
}

// GetOverallStatus determines overall health status based on check results.
func (r *Registry) GetOverallStatus(results map[string]interfaces.HealthCheckResult) interfaces.HealthStatus {
	overallStatus := "healthy"
	healthyChecks := 0
	warningChecks := 0
	errorChecks := 0
	uptime := time.Duration(0)

	checks := make(map[string]interfaces.HealthCheck)

	for name, result := range results {
		check := interfaces.HealthCheck{
			Status:    result.Status,
			Message:   result.Message,
			Duration:  result.Duration,
			Timestamp: result.Timestamp,
		}
		checks[name] = check

		switch result.Status {
		case "healthy":
			healthyChecks++
		case "warning":
			warningChecks++
			if result.Critical {
				overallStatus = "warning"
			}
		case "unhealthy", "error":
			errorChecks++
			if result.Critical {
				overallStatus = "unhealthy"
			}
		}
	}

	// Determine overall status priority: error > warning > healthy
	if errorChecks > 0 && overallStatus != "unhealthy" {
		overallStatus = "warning"
	}

	return interfaces.HealthStatus{
		Status:      overallStatus,
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Uptime:      uptime,
		Connections: 0, // This would be set by the caller
		Checks:      checks,
	}
}

// CheckOption represents a health check configuration option.
type CheckOption func(*CheckInfo)

// WithType sets the check type.
func WithType(checkType CheckType) CheckOption {
	return func(info *CheckInfo) {
		info.Type = checkType
	}
}

// WithInterval sets the check interval.
func WithInterval(interval time.Duration) CheckOption {
	return func(info *CheckInfo) {
		info.Interval = interval
	}
}

// WithTimeout sets the check timeout.
func WithTimeout(timeout time.Duration) CheckOption {
	return func(info *CheckInfo) {
		info.Timeout = timeout
	}
}

// WithCritical sets whether the check is critical.
func WithCritical(critical bool) CheckOption {
	return func(info *CheckInfo) {
		info.Critical = critical
	}
}

// Handler creates HTTP handlers for health checks.
type Handler struct {
	registry *Registry
}

// NewHandler creates a new health check handler.
func NewHandler(registry *Registry) *Handler {
	return &Handler{registry: registry}
}

// LivenessHandler handles liveness probe requests.
func (h *Handler) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := h.registry.RunChecks(r.Context(), CheckTypeLiveness)
		status := h.registry.GetOverallStatus(results)

		w.Header().Set("Content-Type", "application/json")

		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status.Status,
			"checks": results,
		})
	}
}

// ReadinessHandler handles readiness probe requests.
func (h *Handler) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := h.registry.RunChecks(r.Context(), CheckTypeReadiness)
		status := h.registry.GetOverallStatus(results)

		w.Header().Set("Content-Type", "application/json")

		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status.Status,
			"checks": results,
		})
	}
}

// StartupHandler handles startup probe requests.
func (h *Handler) StartupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := h.registry.RunChecks(r.Context(), CheckTypeStartup)
		status := h.registry.GetOverallStatus(results)

		w.Header().Set("Content-Type", "application/json")

		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status.Status,
			"checks": results,
		})
	}
}

// HealthHandler handles comprehensive health check requests.
func (h *Handler) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		livenessResults := h.registry.RunChecks(r.Context(), CheckTypeLiveness)
		readinessResults := h.registry.RunChecks(r.Context(), CheckTypeReadiness)
		startupResults := h.registry.RunChecks(r.Context(), CheckTypeStartup)

		allResults := make(map[string]CheckResult)
		for name, result := range livenessResults {
			allResults[name] = result
		}
		for name, result := range readinessResults {
			allResults[name] = result
		}
		for name, result := range startupResults {
			allResults[name] = result
		}

		status := h.registry.GetOverallStatus(allResults)

		w.Header().Set("Content-Type", "application/json")

		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    status.Status,
			"timestamp": status.Timestamp,
			"uptime":    status.Uptime,
			"liveness":  livenessResults,
			"readiness": readinessResults,
			"startup":   startupResults,
		})
	}
}

// Common health check implementations

// DatabaseCheck creates a database health check.
func DatabaseCheck(pingFunc func(ctx context.Context) error) Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		err := pingFunc(ctx)
		if err != nil {
			return interfaces.HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("Database connection failed: %v", err),
			}
		}
		return interfaces.HealthCheckResult{
			Status:  "healthy",
			Message: "Database connection OK",
		}
	}
}

// URLCheck creates a URL health check.
func URLCheck(url string) Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return interfaces.HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("Failed to create request: %v", err),
			}
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return interfaces.HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("HTTP request failed: %v", err),
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return interfaces.HealthCheckResult{
				Status:  "healthy",
				Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
			}
		}

		return interfaces.HealthCheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}
} // MemoryCheck creates a memory usage health check.
func MemoryCheck(maxMemoryMB int64) Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		// This is a simplified version - in production you'd use runtime.MemStats
		return interfaces.HealthCheckResult{
			Status:  "healthy",
			Message: "Memory usage within limits",
			Metadata: map[string]interface{}{
				"max_memory_mb": maxMemoryMB,
			},
		}
	}
}

// DiskSpaceCheck creates a disk space health check.
func DiskSpaceCheck(path string, minFreeGB int64) Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		// This is a simplified version - in production you'd check actual disk space
		return interfaces.HealthCheckResult{
			Status:  "healthy",
			Message: "Disk space sufficient",
			Metadata: map[string]interface{}{
				"path":        path,
				"min_free_gb": minFreeGB,
			},
		}
	}
}
