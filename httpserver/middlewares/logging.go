package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// LoggingMiddleware provides comprehensive request/response logging functionality.
type LoggingMiddleware struct {
	*BaseMiddleware

	// Configuration
	config LoggingConfig

	// Metrics
	loggedRequests  int64
	loggedResponses int64
	loggedErrors    int64
	failedLogs      int64

	// Internal state
	startTime time.Time
}

// LoggingConfig defines configuration options for the logging middleware.
type LoggingConfig struct {
	// Logging levels
	LogRequests      bool
	LogResponses     bool
	LogHeaders       bool
	LogBody          bool
	LogSensitiveData bool

	// Log format
	Format       LogFormat
	TimeFormat   string
	IncludeStack bool

	// Filtering
	SkipPaths       []string
	SkipMethods     []string
	SkipStatusCodes []int

	// Body logging
	MaxBodySize  int
	TruncateBody bool

	// Performance
	BufferSize    int
	FlushInterval time.Duration
	AsyncLogging  bool

	// Sanitization
	SensitiveHeaders []string
	SensitiveFields  []string
	MaskChar         string
}

// LogFormat defines the output format for logs.
type LogFormat int

const (
	LogFormatJSON LogFormat = iota
	LogFormatText
	LogFormatStructured
	LogFormatCustom
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	Level        string                 `json:"level"`
	Message      string                 `json:"message"`
	Method       string                 `json:"method,omitempty"`
	Path         string                 `json:"path,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	Duration     time.Duration          `json:"duration,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	Body         string                 `json:"body,omitempty"`
	ErrorMessage string                 `json:"error,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RemoteAddr   string                 `json:"remote_addr,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// RequestLogContext holds request-specific logging context.
type RequestLogContext struct {
	StartTime time.Time
	RequestID string
	Method    string
	Path      string
	Headers   map[string]string
	Body      string
	Extra     map[string]interface{}
}

// NewLoggingMiddleware creates a new logging middleware with default configuration.
func NewLoggingMiddleware(priority int) *LoggingMiddleware {
	return &LoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("logging", priority),
		config:         DefaultLoggingConfig(),
		startTime:      time.Now(),
	}
}

// NewLoggingMiddlewareWithConfig creates a new logging middleware with custom configuration.
func NewLoggingMiddlewareWithConfig(priority int, config LoggingConfig) *LoggingMiddleware {
	return &LoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("logging", priority),
		config:         config,
		startTime:      time.Now(),
	}
}

// DefaultLoggingConfig returns a default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		LogRequests:      true,
		LogResponses:     true,
		LogHeaders:       true,
		LogBody:          false,
		LogSensitiveData: false,
		Format:           LogFormatJSON,
		TimeFormat:       time.RFC3339,
		IncludeStack:     false,
		SkipPaths:        []string{"/health", "/metrics", "/favicon.ico"},
		SkipMethods:      []string{"OPTIONS"},
		SkipStatusCodes:  []int{},
		MaxBodySize:      1024,
		TruncateBody:     true,
		BufferSize:       100,
		FlushInterval:    time.Second * 5,
		AsyncLogging:     false,
		SensitiveHeaders: []string{"Authorization", "Cookie", "X-API-Key", "X-Auth-Token"},
		SensitiveFields:  []string{"password", "token", "secret", "key", "auth"},
		MaskChar:         "*",
	}
}

// Process implements the Middleware interface for request/response logging.
func (lm *LoggingMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !lm.IsEnabled() {
		return next(ctx, req)
	}

	startTime := time.Now()

	// Extract request context
	reqCtx := lm.extractRequestContext(ctx, req)

	// Check if request should be skipped
	if lm.shouldSkipRequest(reqCtx) {
		return next(ctx, req)
	}

	// Log request
	if lm.config.LogRequests {
		lm.logRequest(ctx, reqCtx)
		atomic.AddInt64(&lm.loggedRequests, 1)
	}

	// Process request
	resp, err := next(ctx, req)

	duration := time.Since(startTime)

	// Log response or error
	if err != nil {
		lm.logError(ctx, reqCtx, err, duration)
		atomic.AddInt64(&lm.loggedErrors, 1)
	} else if lm.config.LogResponses {
		lm.logResponse(ctx, reqCtx, resp, duration)
		atomic.AddInt64(&lm.loggedResponses, 1)
	}

	return resp, err
}

// GetConfig returns the current logging configuration.
func (lm *LoggingMiddleware) GetConfig() LoggingConfig {
	return lm.config
}

// SetConfig updates the logging configuration.
func (lm *LoggingMiddleware) SetConfig(config LoggingConfig) {
	lm.config = config
	lm.GetLogger().Info("Logging middleware configuration updated")
}

// GetMetrics returns logging metrics.
func (lm *LoggingMiddleware) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"logged_requests":  atomic.LoadInt64(&lm.loggedRequests),
		"logged_responses": atomic.LoadInt64(&lm.loggedResponses),
		"logged_errors":    atomic.LoadInt64(&lm.loggedErrors),
		"failed_logs":      atomic.LoadInt64(&lm.failedLogs),
		"uptime":           time.Since(lm.startTime),
	}
}

// extractRequestContext extracts relevant information from the request context.
func (lm *LoggingMiddleware) extractRequestContext(ctx context.Context, req interface{}) *RequestLogContext {
	reqCtx := &RequestLogContext{
		StartTime: time.Now(),
		Headers:   make(map[string]string),
		Extra:     make(map[string]interface{}),
	}

	// Extract request ID from context if available
	if reqID := ctx.Value("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			reqCtx.RequestID = id
		}
	}

	// Try to extract HTTP-specific information
	if httpReq, ok := req.(map[string]interface{}); ok {
		if method, exists := httpReq["method"]; exists {
			if m, ok := method.(string); ok {
				reqCtx.Method = m
			}
		}

		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				reqCtx.Path = p
			}
		}

		if headers, exists := httpReq["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				reqCtx.Headers = lm.sanitizeHeaders(h)
			}
		}

		if body, exists := httpReq["body"]; exists {
			if lm.config.LogBody {
				reqCtx.Body = lm.sanitizeBody(fmt.Sprintf("%v", body))
			}
		}
	}

	return reqCtx
}

// shouldSkipRequest determines if a request should be skipped from logging.
func (lm *LoggingMiddleware) shouldSkipRequest(reqCtx *RequestLogContext) bool {
	// Check skip paths
	for _, skipPath := range lm.config.SkipPaths {
		if reqCtx.Path == skipPath {
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range lm.config.SkipMethods {
		if reqCtx.Method == skipMethod {
			return true
		}
	}

	return false
}

// logRequest logs incoming request details.
func (lm *LoggingMiddleware) logRequest(ctx context.Context, reqCtx *RequestLogContext) {
	entry := LogEntry{
		Timestamp: reqCtx.StartTime,
		Level:     "INFO",
		Message:   "Incoming request",
		Method:    reqCtx.Method,
		Path:      reqCtx.Path,
		Headers:   reqCtx.Headers,
		Body:      reqCtx.Body,
		RequestID: reqCtx.RequestID,
		Extra:     reqCtx.Extra,
	}

	// Extract additional context information
	if userAgent, ok := reqCtx.Headers["User-Agent"]; ok {
		entry.UserAgent = userAgent
	}

	if remoteAddr := ctx.Value("remote_addr"); remoteAddr != nil {
		if addr, ok := remoteAddr.(string); ok {
			entry.RemoteAddr = addr
		}
	}

	lm.writeLog(entry)
}

// logResponse logs response details.
func (lm *LoggingMiddleware) logResponse(ctx context.Context, reqCtx *RequestLogContext, resp interface{}, duration time.Duration) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "Response sent",
		Method:    reqCtx.Method,
		Path:      reqCtx.Path,
		Duration:  duration,
		RequestID: reqCtx.RequestID,
		Extra:     reqCtx.Extra,
	}

	// Extract response details
	if httpResp, ok := resp.(map[string]interface{}); ok {
		if statusCode, exists := httpResp["status_code"]; exists {
			if code, ok := statusCode.(int); ok {
				entry.StatusCode = code
				entry.Level = lm.getLogLevelForStatusCode(code)
			}
		}

		if lm.config.LogBody {
			if body, exists := httpResp["body"]; exists {
				entry.Body = lm.sanitizeBody(fmt.Sprintf("%v", body))
			}
		}

		if headers, exists := httpResp["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				entry.Headers = lm.sanitizeHeaders(h)
			}
		}
	}

	lm.writeLog(entry)
}

// logError logs error details.
func (lm *LoggingMiddleware) logError(ctx context.Context, reqCtx *RequestLogContext, err error, duration time.Duration) {
	entry := LogEntry{
		Timestamp:    time.Now(),
		Level:        "ERROR",
		Message:      "Request processing error",
		Method:       reqCtx.Method,
		Path:         reqCtx.Path,
		Duration:     duration,
		ErrorMessage: err.Error(),
		RequestID:    reqCtx.RequestID,
		Extra:        reqCtx.Extra,
	}

	lm.writeLog(entry)
}

// writeLog writes the log entry in the configured format.
func (lm *LoggingMiddleware) writeLog(entry LogEntry) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&lm.failedLogs, 1)
			lm.GetLogger().Error("Failed to write log entry: %v", r)
		}
	}()

	switch lm.config.Format {
	case LogFormatJSON:
		lm.writeJSONLog(entry)
	case LogFormatText:
		lm.writeTextLog(entry)
	case LogFormatStructured:
		lm.writeStructuredLog(entry)
	default:
		lm.writeJSONLog(entry)
	}
}

// writeJSONLog writes log entry in JSON format.
func (lm *LoggingMiddleware) writeJSONLog(entry LogEntry) {
	entry.Timestamp = entry.Timestamp.UTC()
	if lm.config.TimeFormat != "" {
		// Custom time format handling would go here
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		atomic.AddInt64(&lm.failedLogs, 1)
		lm.GetLogger().Error("Failed to marshal log entry to JSON: %v", err)
		return
	}

	lm.GetLogger().Info(string(jsonData))
}

// writeTextLog writes log entry in text format.
func (lm *LoggingMiddleware) writeTextLog(entry LogEntry) {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("[%s] %s", entry.Timestamp.Format(lm.config.TimeFormat), entry.Level))

	if entry.RequestID != "" {
		msg.WriteString(fmt.Sprintf(" [%s]", entry.RequestID))
	}

	msg.WriteString(fmt.Sprintf(" %s", entry.Message))

	if entry.Method != "" && entry.Path != "" {
		msg.WriteString(fmt.Sprintf(" %s %s", entry.Method, entry.Path))
	}

	if entry.StatusCode > 0 {
		msg.WriteString(fmt.Sprintf(" %d", entry.StatusCode))
	}

	if entry.Duration > 0 {
		msg.WriteString(fmt.Sprintf(" %v", entry.Duration))
	}

	if entry.ErrorMessage != "" {
		msg.WriteString(fmt.Sprintf(" error: %s", entry.ErrorMessage))
	}

	lm.GetLogger().Info(msg.String())
}

// writeStructuredLog writes log entry in structured format.
func (lm *LoggingMiddleware) writeStructuredLog(entry LogEntry) {
	level := entry.Level
	msg := entry.Message

	args := []interface{}{
		"timestamp", entry.Timestamp,
		"request_id", entry.RequestID,
		"method", entry.Method,
		"path", entry.Path,
	}

	if entry.StatusCode > 0 {
		args = append(args, "status_code", entry.StatusCode)
	}

	if entry.Duration > 0 {
		args = append(args, "duration", entry.Duration)
	}

	if entry.ErrorMessage != "" {
		args = append(args, "error", entry.ErrorMessage)
	}

	switch level {
	case "ERROR":
		lm.GetLogger().Error(msg, args...)
	case "WARN":
		lm.GetLogger().Warn(msg, args...)
	case "DEBUG":
		lm.GetLogger().Debug(msg, args...)
	default:
		lm.GetLogger().Info(msg, args...)
	}
}

// sanitizeHeaders removes or masks sensitive header information.
func (lm *LoggingMiddleware) sanitizeHeaders(headers map[string]string) map[string]string {
	if !lm.config.LogSensitiveData {
		sanitized := make(map[string]string)
		for key, value := range headers {
			if lm.isSensitiveHeader(key) {
				sanitized[key] = lm.maskValue(value)
			} else {
				sanitized[key] = value
			}
		}
		return sanitized
	}
	return headers
}

// sanitizeBody removes or masks sensitive body information.
func (lm *LoggingMiddleware) sanitizeBody(body string) string {
	if len(body) > lm.config.MaxBodySize && lm.config.TruncateBody {
		body = body[:lm.config.MaxBodySize] + "..."
	}

	if !lm.config.LogSensitiveData {
		// Simple field masking for JSON-like bodies
		for _, field := range lm.config.SensitiveFields {
			// This is a simplified approach - in production, you'd want proper JSON parsing
			if strings.Contains(strings.ToLower(body), strings.ToLower(field)) {
				// Mask the field value
				body = strings.ReplaceAll(body, field, lm.maskValue(field))
			}
		}
	}

	return body
}

// isSensitiveHeader checks if a header contains sensitive information.
func (lm *LoggingMiddleware) isSensitiveHeader(header string) bool {
	headerLower := strings.ToLower(header)
	for _, sensitive := range lm.config.SensitiveHeaders {
		if strings.ToLower(sensitive) == headerLower {
			return true
		}
	}
	return false
}

// maskValue masks a sensitive value.
func (lm *LoggingMiddleware) maskValue(value string) string {
	if len(value) <= 4 {
		return strings.Repeat(lm.config.MaskChar, len(value))
	}
	return value[:2] + strings.Repeat(lm.config.MaskChar, len(value)-4) + value[len(value)-2:]
}

// getLogLevelForStatusCode returns appropriate log level based on HTTP status code.
func (lm *LoggingMiddleware) getLogLevelForStatusCode(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "ERROR"
	case statusCode >= 400:
		return "WARN"
	case statusCode >= 300:
		return "INFO"
	default:
		return "INFO"
	}
}

// Reset resets all metrics.
func (lm *LoggingMiddleware) Reset() {
	atomic.StoreInt64(&lm.loggedRequests, 0)
	atomic.StoreInt64(&lm.loggedResponses, 0)
	atomic.StoreInt64(&lm.loggedErrors, 0)
	atomic.StoreInt64(&lm.failedLogs, 0)
	lm.startTime = time.Now()
	lm.GetLogger().Info("Logging middleware metrics reset")
}
