package middlewares

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewLoggingMiddleware(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	if middleware == nil {
		t.Fatal("Expected logging middleware to be created")
	}

	if middleware.Name() != "logging" {
		t.Errorf("Expected name to be 'logging', got %s", middleware.Name())
	}

	if middleware.Priority() != 1 {
		t.Errorf("Expected priority to be 1, got %d", middleware.Priority())
	}

	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled by default")
	}

	config := middleware.GetConfig()
	if !config.LogRequests {
		t.Error("Expected LogRequests to be true by default")
	}

	if !config.LogResponses {
		t.Error("Expected LogResponses to be true by default")
	}
}

func TestNewLoggingMiddlewareWithConfig(t *testing.T) {
	customConfig := LoggingConfig{
		LogRequests:  false,
		LogResponses: true,
		LogHeaders:   false,
		LogBody:      true,
		Format:       LogFormatText,
		MaxBodySize:  2048,
	}

	middleware := NewLoggingMiddlewareWithConfig(2, customConfig)

	if middleware == nil {
		t.Fatal("Expected logging middleware to be created")
	}

	config := middleware.GetConfig()
	if config.LogRequests {
		t.Error("Expected LogRequests to be false")
	}

	if !config.LogResponses {
		t.Error("Expected LogResponses to be true")
	}

	if config.LogHeaders {
		t.Error("Expected LogHeaders to be false")
	}

	if !config.LogBody {
		t.Error("Expected LogBody to be true")
	}

	if config.Format != LogFormatText {
		t.Errorf("Expected format to be LogFormatText, got %v", config.Format)
	}

	if config.MaxBodySize != 2048 {
		t.Errorf("Expected MaxBodySize to be 2048, got %d", config.MaxBodySize)
	}
}

func TestDefaultLoggingConfig(t *testing.T) {
	config := DefaultLoggingConfig()

	if !config.LogRequests {
		t.Error("Expected LogRequests to be true by default")
	}

	if !config.LogResponses {
		t.Error("Expected LogResponses to be true by default")
	}

	if !config.LogHeaders {
		t.Error("Expected LogHeaders to be true by default")
	}

	if config.LogBody {
		t.Error("Expected LogBody to be false by default")
	}

	if config.LogSensitiveData {
		t.Error("Expected LogSensitiveData to be false by default")
	}

	if config.Format != LogFormatJSON {
		t.Errorf("Expected format to be LogFormatJSON, got %v", config.Format)
	}

	if config.MaxBodySize != 1024 {
		t.Errorf("Expected MaxBodySize to be 1024, got %d", config.MaxBodySize)
	}

	expectedSkipPaths := []string{"/health", "/metrics", "/favicon.ico"}
	if len(config.SkipPaths) != len(expectedSkipPaths) {
		t.Errorf("Expected %d skip paths, got %d", len(expectedSkipPaths), len(config.SkipPaths))
	}

	expectedSensitiveHeaders := []string{"Authorization", "Cookie", "X-API-Key", "X-Auth-Token"}
	if len(config.SensitiveHeaders) != len(expectedSensitiveHeaders) {
		t.Errorf("Expected %d sensitive headers, got %d", len(expectedSensitiveHeaders), len(config.SensitiveHeaders))
	}
}

func TestLoggingMiddleware_Process(t *testing.T) {
	middleware := NewLoggingMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/test",
		"headers": map[string]string{
			"User-Agent": "test-agent",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		return map[string]interface{}{
			"status_code": 200,
			"body":        "test response",
		}, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Error("Expected response to not be nil")
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	// Check metrics
	metrics := middleware.GetMetrics()
	if metrics["logged_requests"].(int64) != 1 {
		t.Errorf("Expected 1 logged request, got %v", metrics["logged_requests"])
	}

	if metrics["logged_responses"].(int64) != 1 {
		t.Errorf("Expected 1 logged response, got %v", metrics["logged_responses"])
	}

	// Check that logs were written
	if len(testLogger.InfoMessages) < 2 {
		t.Errorf("Expected at least 2 info messages (request + response), got %d", len(testLogger.InfoMessages))
	}
}

func TestLoggingMiddleware_ProcessWithError(t *testing.T) {
	middleware := NewLoggingMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "POST",
		"path":   "/error",
	}

	testError := errors.New("test error")
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, testError
	}

	_, err := middleware.Process(ctx, req, next)
	if err == nil {
		t.Error("Expected error to be returned")
	}

	if err != testError {
		t.Errorf("Expected error to be %v, got %v", testError, err)
	}

	// Check metrics
	metrics := middleware.GetMetrics()
	if metrics["logged_errors"].(int64) != 1 {
		t.Errorf("Expected 1 logged error, got %v", metrics["logged_errors"])
	}

	// Check that error was logged
	if len(testLogger.ErrorMessages) == 0 && len(testLogger.InfoMessages) == 0 {
		t.Error("Expected error to be logged")
	}
}

func TestLoggingMiddleware_SkipPaths(t *testing.T) {
	config := DefaultLoggingConfig()
	middleware := NewLoggingMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/health", // This path should be skipped
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"status_code": 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that no logs were written for skipped path
	metrics := middleware.GetMetrics()
	if metrics["logged_requests"].(int64) != 0 {
		t.Errorf("Expected 0 logged requests for skipped path, got %v", metrics["logged_requests"])
	}
}

func TestLoggingMiddleware_SkipMethods(t *testing.T) {
	config := DefaultLoggingConfig()
	middleware := NewLoggingMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "OPTIONS", // This method should be skipped
		"path":   "/test",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"status_code": 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that no logs were written for skipped method
	metrics := middleware.GetMetrics()
	if metrics["logged_requests"].(int64) != 0 {
		t.Errorf("Expected 0 logged requests for skipped method, got %v", metrics["logged_requests"])
	}
}

func TestLoggingMiddleware_DisabledMiddleware(t *testing.T) {
	middleware := NewLoggingMiddleware(1)
	middleware.SetEnabled(false)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/test",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"status_code": 200}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that no logs were written when disabled
	metrics := middleware.GetMetrics()
	if metrics["logged_requests"].(int64) != 0 {
		t.Errorf("Expected 0 logged requests when disabled, got %v", metrics["logged_requests"])
	}
}

func TestLoggingMiddleware_SanitizeHeaders(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	headers := map[string]string{
		"Authorization": "Bearer secret-token",
		"Cookie":        "session=abc123",
		"Content-Type":  "application/json",
		"User-Agent":    "test-agent",
	}

	sanitized := middleware.sanitizeHeaders(headers)

	// Sensitive headers should be masked
	if !strings.Contains(sanitized["Authorization"], "*") {
		t.Error("Expected Authorization header to be masked")
	}

	if !strings.Contains(sanitized["Cookie"], "*") {
		t.Error("Expected Cookie header to be masked")
	}

	// Non-sensitive headers should remain unchanged
	if sanitized["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type to remain unchanged, got %s", sanitized["Content-Type"])
	}

	if sanitized["User-Agent"] != "test-agent" {
		t.Errorf("Expected User-Agent to remain unchanged, got %s", sanitized["User-Agent"])
	}
}

func TestLoggingMiddleware_SanitizeBody(t *testing.T) {
	config := DefaultLoggingConfig()
	config.MaxBodySize = 10
	config.TruncateBody = true
	middleware := NewLoggingMiddlewareWithConfig(1, config)

	// Test body truncation
	longBody := "this is a very long body that should be truncated"
	sanitized := middleware.sanitizeBody(longBody)

	if len(sanitized) <= config.MaxBodySize {
		// Should be truncated
		if !strings.HasSuffix(sanitized, "...") {
			t.Error("Expected truncated body to end with '...'")
		}
	}

	// Test sensitive field masking
	sensitiveBody := `{"password": "secret123", "username": "user"}`
	sanitized = middleware.sanitizeBody(sensitiveBody)

	if strings.Contains(sanitized, "secret123") {
		t.Error("Expected sensitive field to be masked")
	}
}

func TestLoggingMiddleware_MaskValue(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	tests := []struct {
		input    string
		expected string
	}{
		{"ab", "**"},
		{"abc", "***"},
		{"abcd", "****"},
		{"abcde", "ab*de"},
		{"secret123", "se*****23"},
	}

	for _, test := range tests {
		result := middleware.maskValue(test.input)
		if result != test.expected {
			t.Errorf("Expected maskValue(%s) to be %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestLoggingMiddleware_GetLogLevelForStatusCode(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	tests := []struct {
		statusCode int
		expected   string
	}{
		{200, "INFO"},
		{201, "INFO"},
		{301, "INFO"},
		{400, "WARN"},
		{404, "WARN"},
		{500, "ERROR"},
		{503, "ERROR"},
	}

	for _, test := range tests {
		result := middleware.getLogLevelForStatusCode(test.statusCode)
		if result != test.expected {
			t.Errorf("Expected log level for status %d to be %s, got %s", test.statusCode, test.expected, result)
		}
	}
}

func TestLoggingMiddleware_SetConfig(t *testing.T) {
	middleware := NewLoggingMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	newConfig := LoggingConfig{
		LogRequests:  false,
		LogResponses: false,
		LogHeaders:   false,
		LogBody:      true,
		Format:       LogFormatText,
	}

	middleware.SetConfig(newConfig)

	config := middleware.GetConfig()
	if config.LogRequests {
		t.Error("Expected LogRequests to be false after config update")
	}

	if config.LogBody != true {
		t.Error("Expected LogBody to be true after config update")
	}

	if config.Format != LogFormatText {
		t.Errorf("Expected format to be LogFormatText after config update, got %v", config.Format)
	}

	// Check that a log message was written about config update
	if len(testLogger.InfoMessages) == 0 {
		t.Error("Expected info message about config update")
	}
}

func TestLoggingMiddleware_Reset(t *testing.T) {
	middleware := NewLoggingMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	// Simulate some activity
	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/test",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"status_code": 200}, nil
	}

	middleware.Process(ctx, req, next)

	// Check that metrics are not zero
	metrics := middleware.GetMetrics()
	if metrics["logged_requests"].(int64) == 0 {
		t.Error("Expected logged_requests to be greater than 0 before reset")
	}

	// Reset metrics
	middleware.Reset()

	// Check that metrics are reset
	metrics = middleware.GetMetrics()
	if metrics["logged_requests"].(int64) != 0 {
		t.Errorf("Expected logged_requests to be 0 after reset, got %v", metrics["logged_requests"])
	}

	if metrics["logged_responses"].(int64) != 0 {
		t.Errorf("Expected logged_responses to be 0 after reset, got %v", metrics["logged_responses"])
	}

	if metrics["logged_errors"].(int64) != 0 {
		t.Errorf("Expected logged_errors to be 0 after reset, got %v", metrics["logged_errors"])
	}

	if metrics["failed_logs"].(int64) != 0 {
		t.Errorf("Expected failed_logs to be 0 after reset, got %v", metrics["failed_logs"])
	}
}

func TestLoggingMiddleware_IsSensitiveHeader(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	tests := []struct {
		header   string
		expected bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"AUTHORIZATION", true},
		{"Cookie", true},
		{"X-API-Key", true},
		{"Content-Type", false},
		{"User-Agent", false},
		{"Accept", false},
	}

	for _, test := range tests {
		result := middleware.isSensitiveHeader(test.header)
		if result != test.expected {
			t.Errorf("Expected isSensitiveHeader(%s) to be %v, got %v", test.header, test.expected, result)
		}
	}
}

func TestLoggingMiddleware_ExtractRequestContext(t *testing.T) {
	middleware := NewLoggingMiddleware(1)

	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	req := map[string]interface{}{
		"method": "POST",
		"path":   "/api/test",
		"headers": map[string]string{
			"Content-Type": "application/json",
			"User-Agent":   "test-client",
		},
		"body": `{"test": "data"}`,
	}

	reqCtx := middleware.extractRequestContext(ctx, req)

	if reqCtx.RequestID != "test-123" {
		t.Errorf("Expected request ID to be 'test-123', got %s", reqCtx.RequestID)
	}

	if reqCtx.Method != "POST" {
		t.Errorf("Expected method to be 'POST', got %s", reqCtx.Method)
	}

	if reqCtx.Path != "/api/test" {
		t.Errorf("Expected path to be '/api/test', got %s", reqCtx.Path)
	}

	if len(reqCtx.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(reqCtx.Headers))
	}

	if reqCtx.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
}
