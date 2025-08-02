package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// HealthCheckMiddleware provides health check functionality.
type HealthCheckMiddleware struct {
	*BaseMiddleware

	// Configuration
	config HealthCheckConfig

	// Health checkers
	checkers map[string]HealthChecker
	mu       sync.RWMutex

	// Metrics
	totalChecks      int64
	successfulChecks int64
	failedChecks     int64
	lastCheckTime    int64 // Unix timestamp

	// Cache
	cachedResult *HealthResult
	cacheExpiry  time.Time
	cacheMutex   sync.RWMutex

	// Internal state
	startTime time.Time
}

// HealthCheckConfig defines configuration options for the health check middleware.
type HealthCheckConfig struct {
	// Endpoint configuration
	HealthPath    string
	LivenessPath  string
	ReadinessPath string

	// Check configuration
	CheckInterval time.Duration
	CheckTimeout  time.Duration
	CacheTimeout  time.Duration

	// Response configuration
	SuccessStatusCode int
	FailureStatusCode int
	DetailedResponse  bool
	IncludeMetrics    bool
	IncludeVersion    bool

	// Dependencies
	Dependencies   []string
	CriticalChecks []string

	// Behavior
	FailFast         bool
	ParallelChecks   bool
	GracefulShutdown bool

	// Custom data
	CustomData  map[string]interface{}
	Version     string
	ServiceName string
	Environment string
}

// HealthChecker defines the interface for health checkers.
type HealthChecker interface {
	Check(ctx context.Context) HealthCheckResult
	Name() string
	IsCritical() bool
	GetTimeout() time.Duration
}

// HealthCheckResult represents the result of a single health check.
type HealthCheckResult struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthResult represents the overall health status.
type HealthResult struct {
	Status      HealthStatus                 `json:"status"`
	Timestamp   time.Time                    `json:"timestamp"`
	Duration    time.Duration                `json:"duration"`
	Version     string                       `json:"version,omitempty"`
	ServiceName string                       `json:"service_name,omitempty"`
	Environment string                       `json:"environment,omitempty"`
	Checks      map[string]HealthCheckResult `json:"checks"`
	Metrics     map[string]interface{}       `json:"metrics,omitempty"`
	Custom      map[string]interface{}       `json:"custom,omitempty"`
}

// HealthStatus represents health status values.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// DatabaseHealthChecker checks database connectivity.
type DatabaseHealthChecker struct {
	name     string
	critical bool
	timeout  time.Duration
	// Database connection would be injected here
}

// ServiceHealthChecker checks external service connectivity.
type ServiceHealthChecker struct {
	name     string
	url      string
	critical bool
	timeout  time.Duration
}

// MemoryHealthChecker checks memory usage.
type MemoryHealthChecker struct {
	name      string
	critical  bool
	timeout   time.Duration
	threshold float64 // Memory usage threshold (0.0 - 1.0)
}

// NewHealthCheckMiddleware creates a new health check middleware with default configuration.
func NewHealthCheckMiddleware(priority int) *HealthCheckMiddleware {
	return &HealthCheckMiddleware{
		BaseMiddleware: NewBaseMiddleware("health_check", priority),
		config:         DefaultHealthCheckConfig(),
		checkers:       make(map[string]HealthChecker),
		startTime:      time.Now(),
	}
}

// NewHealthCheckMiddlewareWithConfig creates a new health check middleware with custom configuration.
func NewHealthCheckMiddlewareWithConfig(priority int, config HealthCheckConfig) *HealthCheckMiddleware {
	return &HealthCheckMiddleware{
		BaseMiddleware: NewBaseMiddleware("health_check", priority),
		config:         config,
		checkers:       make(map[string]HealthChecker),
		startTime:      time.Now(),
	}
}

// DefaultHealthCheckConfig returns a default health check configuration.
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		HealthPath:        "/health",
		LivenessPath:      "/health/live",
		ReadinessPath:     "/health/ready",
		CheckInterval:     time.Second * 30,
		CheckTimeout:      time.Second * 10,
		CacheTimeout:      time.Second * 5,
		SuccessStatusCode: 200,
		FailureStatusCode: 503, // Service Unavailable
		DetailedResponse:  true,
		IncludeMetrics:    true,
		IncludeVersion:    true,
		Dependencies:      []string{},
		CriticalChecks:    []string{},
		FailFast:          false,
		ParallelChecks:    true,
		GracefulShutdown:  true,
		CustomData:        make(map[string]interface{}),
		Version:           "1.0.0",
		ServiceName:       "http-server",
		Environment:       "development",
	}
}

// Process implements the Middleware interface for health check handling.
func (hcm *HealthCheckMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !hcm.IsEnabled() {
		return next(ctx, req)
	}

	// Extract request path
	path := hcm.extractPath(req)

	// Check if this is a health check request
	if hcm.isHealthCheckPath(path) {
		return hcm.handleHealthCheck(ctx, path)
	}

	// Not a health check request, proceed normally
	return next(ctx, req)
}

// GetConfig returns the current health check configuration.
func (hcm *HealthCheckMiddleware) GetConfig() HealthCheckConfig {
	return hcm.config
}

// SetConfig updates the health check configuration.
func (hcm *HealthCheckMiddleware) SetConfig(config HealthCheckConfig) {
	hcm.config = config
	hcm.GetLogger().Info("Health check middleware configuration updated")
}

// AddChecker adds a health checker.
func (hcm *HealthCheckMiddleware) AddChecker(checker HealthChecker) error {
	if checker == nil {
		return fmt.Errorf("checker cannot be nil")
	}

	hcm.mu.Lock()
	hcm.checkers[checker.Name()] = checker
	hcm.mu.Unlock()

	hcm.GetLogger().Info("Health checker '%s' added", checker.Name())
	return nil
}

// RemoveChecker removes a health checker.
func (hcm *HealthCheckMiddleware) RemoveChecker(name string) error {
	hcm.mu.Lock()
	defer hcm.mu.Unlock()

	if _, exists := hcm.checkers[name]; !exists {
		return fmt.Errorf("checker '%s' not found", name)
	}

	delete(hcm.checkers, name)
	hcm.GetLogger().Info("Health checker '%s' removed", name)
	return nil
}

// GetChecker retrieves a health checker by name.
func (hcm *HealthCheckMiddleware) GetChecker(name string) (HealthChecker, error) {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	checker, exists := hcm.checkers[name]
	if !exists {
		return nil, fmt.Errorf("checker '%s' not found", name)
	}
	return checker, nil
}

// GetMetrics returns health check metrics.
func (hcm *HealthCheckMiddleware) GetMetrics() map[string]interface{} {
	totalChecks := atomic.LoadInt64(&hcm.totalChecks)
	successfulChecks := atomic.LoadInt64(&hcm.successfulChecks)
	failedChecks := atomic.LoadInt64(&hcm.failedChecks)

	var successRate float64
	if totalChecks > 0 {
		successRate = float64(successfulChecks) / float64(totalChecks) * 100.0
	}

	lastCheckTime := atomic.LoadInt64(&hcm.lastCheckTime)
	var timeSinceLastCheck time.Duration
	if lastCheckTime > 0 {
		timeSinceLastCheck = time.Since(time.Unix(lastCheckTime, 0))
	}

	hcm.mu.RLock()
	checkerCount := len(hcm.checkers)
	hcm.mu.RUnlock()

	return map[string]interface{}{
		"total_checks":          totalChecks,
		"successful_checks":     successfulChecks,
		"failed_checks":         failedChecks,
		"success_rate":          successRate,
		"time_since_last_check": timeSinceLastCheck,
		"active_checkers":       checkerCount,
		"uptime":                time.Since(hcm.startTime),
	}
}

// extractPath extracts the request path.
func (hcm *HealthCheckMiddleware) extractPath(req interface{}) string {
	if httpReq, ok := req.(map[string]interface{}); ok {
		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				return p
			}
		}
	}
	return ""
}

// isHealthCheckPath determines if the path is a health check endpoint.
func (hcm *HealthCheckMiddleware) isHealthCheckPath(path string) bool {
	return path == hcm.config.HealthPath ||
		path == hcm.config.LivenessPath ||
		path == hcm.config.ReadinessPath
}

// handleHealthCheck handles health check requests.
func (hcm *HealthCheckMiddleware) handleHealthCheck(ctx context.Context, path string) (interface{}, error) {
	hcm.GetLogger().Debug("Handling health check request for path: %s", path)

	// Check cache first
	if cachedResult := hcm.getCachedResult(); cachedResult != nil {
		return hcm.createHealthResponse(cachedResult), nil
	}

	// Perform health checks
	result := hcm.performHealthChecks(ctx, path)

	// Cache the result
	hcm.setCachedResult(result)

	return hcm.createHealthResponse(result), nil
}

// getCachedResult retrieves cached health check result if valid.
func (hcm *HealthCheckMiddleware) getCachedResult() *HealthResult {
	hcm.cacheMutex.RLock()
	defer hcm.cacheMutex.RUnlock()

	if hcm.cachedResult != nil && time.Now().Before(hcm.cacheExpiry) {
		return hcm.cachedResult
	}

	return nil
}

// setCachedResult caches the health check result.
func (hcm *HealthCheckMiddleware) setCachedResult(result *HealthResult) {
	hcm.cacheMutex.Lock()
	defer hcm.cacheMutex.Unlock()

	hcm.cachedResult = result
	hcm.cacheExpiry = time.Now().Add(hcm.config.CacheTimeout)
}

// performHealthChecks performs all health checks.
func (hcm *HealthCheckMiddleware) performHealthChecks(ctx context.Context, path string) *HealthResult {
	startTime := time.Now()
	atomic.AddInt64(&hcm.totalChecks, 1)
	atomic.StoreInt64(&hcm.lastCheckTime, startTime.Unix())

	result := &HealthResult{
		Timestamp:   startTime,
		Version:     hcm.config.Version,
		ServiceName: hcm.config.ServiceName,
		Environment: hcm.config.Environment,
		Checks:      make(map[string]HealthCheckResult),
		Custom:      hcm.config.CustomData,
	}

	// Get checkers to run
	checkersToRun := hcm.getCheckersForPath(path)

	if len(checkersToRun) == 0 {
		result.Status = HealthStatusHealthy
		result.Duration = time.Since(startTime)
		atomic.AddInt64(&hcm.successfulChecks, 1)
		return result
	}

	// Run health checks
	if hcm.config.ParallelChecks {
		hcm.runParallelChecks(ctx, checkersToRun, result)
	} else {
		hcm.runSequentialChecks(ctx, checkersToRun, result)
	}

	// Determine overall status
	result.Status = hcm.determineOverallStatus(result.Checks)
	result.Duration = time.Since(startTime)

	// Include metrics if configured
	if hcm.config.IncludeMetrics {
		result.Metrics = hcm.GetMetrics()
	}

	// Update metrics
	if result.Status == HealthStatusHealthy {
		atomic.AddInt64(&hcm.successfulChecks, 1)
	} else {
		atomic.AddInt64(&hcm.failedChecks, 1)
	}

	hcm.GetLogger().Debug("Health check completed: %s (duration: %v)", result.Status, result.Duration)
	return result
}

// getCheckersForPath returns the appropriate checkers for the given path.
func (hcm *HealthCheckMiddleware) getCheckersForPath(path string) []HealthChecker {
	hcm.mu.RLock()
	defer hcm.mu.RUnlock()

	var checkers []HealthChecker

	switch path {
	case hcm.config.LivenessPath:
		// Liveness checks - typically just critical system checks
		for _, checker := range hcm.checkers {
			if checker.IsCritical() {
				checkers = append(checkers, checker)
			}
		}
	case hcm.config.ReadinessPath:
		// Readiness checks - all dependencies must be ready
		for _, checker := range hcm.checkers {
			checkers = append(checkers, checker)
		}
	default:
		// General health check - all checkers
		for _, checker := range hcm.checkers {
			checkers = append(checkers, checker)
		}
	}

	return checkers
}

// runParallelChecks runs health checks in parallel.
func (hcm *HealthCheckMiddleware) runParallelChecks(ctx context.Context, checkers []HealthChecker, result *HealthResult) {
	var wg sync.WaitGroup
	resultChan := make(chan HealthCheckResult, len(checkers))

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, c.GetTimeout())
			defer cancel()

			checkResult := c.Check(checkCtx)
			resultChan <- checkResult

			if hcm.config.FailFast && checkResult.Status != HealthStatusHealthy && c.IsCritical() {
				// Cancel other checks if this critical check fails
				// Note: This is a simplified implementation
				return
			}
		}(checker)
	}

	wg.Wait()
	close(resultChan)

	// Collect results
	for checkResult := range resultChan {
		result.Checks[checkResult.Name] = checkResult
	}
}

// runSequentialChecks runs health checks sequentially.
func (hcm *HealthCheckMiddleware) runSequentialChecks(ctx context.Context, checkers []HealthChecker, result *HealthResult) {
	for _, checker := range checkers {
		checkCtx, cancel := context.WithTimeout(ctx, checker.GetTimeout())

		checkResult := checker.Check(checkCtx)
		result.Checks[checkResult.Name] = checkResult

		cancel()

		if hcm.config.FailFast && checkResult.Status != HealthStatusHealthy && checker.IsCritical() {
			hcm.GetLogger().Warn("Critical health check failed, stopping remaining checks: %s", checker.Name())
			break
		}
	}
}

// determineOverallStatus determines the overall health status.
func (hcm *HealthCheckMiddleware) determineOverallStatus(checks map[string]HealthCheckResult) HealthStatus {
	if len(checks) == 0 {
		return HealthStatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		case HealthStatusUnknown:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// createHealthResponse creates an HTTP response for health check.
func (hcm *HealthCheckMiddleware) createHealthResponse(result *HealthResult) interface{} {
	var statusCode int
	if result.Status == HealthStatusHealthy {
		statusCode = hcm.config.SuccessStatusCode
	} else {
		statusCode = hcm.config.FailureStatusCode
	}

	var body string
	if hcm.config.DetailedResponse {
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		body = string(jsonData)
	} else {
		simpleResult := map[string]interface{}{
			"status":    result.Status,
			"timestamp": result.Timestamp,
		}
		jsonData, _ := json.Marshal(simpleResult)
		body = string(jsonData)
	}

	return map[string]interface{}{
		"status_code": statusCode,
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache, no-store, must-revalidate",
		},
		"body": body,
	}
}

// Reset resets all metrics.
func (hcm *HealthCheckMiddleware) Reset() {
	atomic.StoreInt64(&hcm.totalChecks, 0)
	atomic.StoreInt64(&hcm.successfulChecks, 0)
	atomic.StoreInt64(&hcm.failedChecks, 0)
	atomic.StoreInt64(&hcm.lastCheckTime, 0)

	hcm.cacheMutex.Lock()
	hcm.cachedResult = nil
	hcm.cacheExpiry = time.Time{}
	hcm.cacheMutex.Unlock()

	hcm.startTime = time.Now()
	hcm.GetLogger().Info("Health check middleware metrics reset")
}

// NewDatabaseHealthChecker creates a new database health checker.
func NewDatabaseHealthChecker(name string, critical bool) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:     name,
		critical: critical,
		timeout:  time.Second * 5,
	}
}

// Check implements HealthChecker interface for database.
func (dhc *DatabaseHealthChecker) Check(ctx context.Context) HealthCheckResult {
	startTime := time.Now()

	// Simulate database connectivity check
	// In real implementation, this would ping the database
	time.Sleep(time.Millisecond * 10)

	return HealthCheckResult{
		Name:      dhc.name,
		Status:    HealthStatusHealthy,
		Message:   "Database connection is healthy",
		Duration:  time.Since(startTime),
		Timestamp: startTime,
		Metadata: map[string]interface{}{
			"type":     "database",
			"critical": dhc.critical,
		},
	}
}

// Name returns the checker name.
func (dhc *DatabaseHealthChecker) Name() string {
	return dhc.name
}

// IsCritical returns whether this checker is critical.
func (dhc *DatabaseHealthChecker) IsCritical() bool {
	return dhc.critical
}

// GetTimeout returns the check timeout.
func (dhc *DatabaseHealthChecker) GetTimeout() time.Duration {
	return dhc.timeout
}

// NewServiceHealthChecker creates a new service health checker.
func NewServiceHealthChecker(name, url string, critical bool) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		name:     name,
		url:      url,
		critical: critical,
		timeout:  time.Second * 10,
	}
}

// Check implements HealthChecker interface for external service.
func (shc *ServiceHealthChecker) Check(ctx context.Context) HealthCheckResult {
	startTime := time.Now()

	// Simulate external service check
	// In real implementation, this would make an HTTP request
	time.Sleep(time.Millisecond * 50)

	status := HealthStatusHealthy
	message := fmt.Sprintf("Service %s is reachable", shc.url)

	// Simulate occasional service issues
	if time.Now().Unix()%10 == 0 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Service %s is responding slowly", shc.url)
	}

	return HealthCheckResult{
		Name:      shc.name,
		Status:    status,
		Message:   message,
		Duration:  time.Since(startTime),
		Timestamp: startTime,
		Metadata: map[string]interface{}{
			"type":     "service",
			"url":      shc.url,
			"critical": shc.critical,
		},
	}
}

// Name returns the checker name.
func (shc *ServiceHealthChecker) Name() string {
	return shc.name
}

// IsCritical returns whether this checker is critical.
func (shc *ServiceHealthChecker) IsCritical() bool {
	return shc.critical
}

// GetTimeout returns the check timeout.
func (shc *ServiceHealthChecker) GetTimeout() time.Duration {
	return shc.timeout
}

// NewMemoryHealthChecker creates a new memory health checker.
func NewMemoryHealthChecker(name string, threshold float64, critical bool) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		name:      name,
		critical:  critical,
		timeout:   time.Second * 2,
		threshold: threshold,
	}
}

// Check implements HealthChecker interface for memory usage.
func (mhc *MemoryHealthChecker) Check(ctx context.Context) HealthCheckResult {
	startTime := time.Now()

	// Simulate memory usage check
	// In real implementation, this would check actual memory usage
	currentUsage := 0.65 // 65% memory usage

	var status HealthStatus
	var message string

	if currentUsage < mhc.threshold {
		status = HealthStatusHealthy
		message = fmt.Sprintf("Memory usage is normal (%.1f%%)", currentUsage*100)
	} else if currentUsage < mhc.threshold*1.2 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Memory usage is high (%.1f%%)", currentUsage*100)
	} else {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Memory usage is critical (%.1f%%)", currentUsage*100)
	}

	return HealthCheckResult{
		Name:      mhc.name,
		Status:    status,
		Message:   message,
		Duration:  time.Since(startTime),
		Timestamp: startTime,
		Metadata: map[string]interface{}{
			"type":      "memory",
			"usage":     currentUsage,
			"threshold": mhc.threshold,
			"critical":  mhc.critical,
		},
	}
}

// Name returns the checker name.
func (mhc *MemoryHealthChecker) Name() string {
	return mhc.name
}

// IsCritical returns whether this checker is critical.
func (mhc *MemoryHealthChecker) IsCritical() bool {
	return mhc.critical
}

// GetTimeout returns the check timeout.
func (mhc *MemoryHealthChecker) GetTimeout() time.Duration {
	return mhc.timeout
}
