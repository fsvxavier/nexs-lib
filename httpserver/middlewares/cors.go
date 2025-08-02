package middlewares

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// CORSMiddleware provides Cross-Origin Resource Sharing (CORS) functionality.
type CORSMiddleware struct {
	*BaseMiddleware

	// Configuration
	config CORSConfig

	// Metrics
	preflightRequests int64
	corsRequests      int64
	corsAllowed       int64
	corsBlocked       int64
	originChecks      int64

	// Internal state
	startTime time.Time
}

// CORSConfig defines configuration options for the CORS middleware.
type CORSConfig struct {
	// Origin configuration
	AllowedOrigins   []string
	AllowOriginFunc  func(string) bool
	AllowAllOrigins  bool
	AllowCredentials bool

	// Method configuration
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string

	// Preflight configuration
	MaxAge              int
	AllowPrivateNetwork bool

	// Request configuration
	SkipPaths          []string
	OptionsPassthrough bool

	// Security configuration
	VaryByOrigin bool
	Debug        bool
}

// CORSRequest represents a CORS request context.
type CORSRequest struct {
	Origin           string
	Method           string
	Path             string
	RequestedHeaders []string
	IsPreflightReq   bool
	IsCORSRequest    bool
}

// CORSResponse represents CORS response headers.
type CORSResponse struct {
	AllowOrigin      string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials string
	MaxAge           string
	VaryHeaders      []string
}

// NewCORSMiddleware creates a new CORS middleware with default configuration.
func NewCORSMiddleware(priority int) *CORSMiddleware {
	return &CORSMiddleware{
		BaseMiddleware: NewBaseMiddleware("cors", priority),
		config:         DefaultCORSConfig(),
		startTime:      time.Now(),
	}
}

// NewCORSMiddlewareWithConfig creates a new CORS middleware with custom configuration.
func NewCORSMiddlewareWithConfig(priority int, config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		BaseMiddleware: NewBaseMiddleware("cors", priority),
		config:         config,
		startTime:      time.Now(),
	}
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:      []string{"*"},
		AllowAllOrigins:     false,
		AllowCredentials:    false,
		AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:      []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:      []string{},
		MaxAge:              86400, // 24 hours
		AllowPrivateNetwork: false,
		SkipPaths:           []string{},
		OptionsPassthrough:  false,
		VaryByOrigin:        true,
		Debug:               false,
	}
}

// Process implements the Middleware interface for CORS handling.
func (cm *CORSMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !cm.IsEnabled() {
		return next(ctx, req)
	}

	// Extract CORS request information
	corsReq := cm.extractCORSRequest(req)

	// Check if path should skip CORS processing
	if cm.shouldSkipCORS(corsReq.Path) {
		return next(ctx, req)
	}

	atomic.AddInt64(&cm.corsRequests, 1)

	// Handle preflight requests
	if corsReq.IsPreflightReq {
		atomic.AddInt64(&cm.preflightRequests, 1)
		return cm.handlePreflightRequest(ctx, corsReq, req, next)
	}

	// Handle actual CORS requests
	if corsReq.IsCORSRequest {
		return cm.handleCORSRequest(ctx, corsReq, req, next)
	}

	// Non-CORS request, proceed normally
	return next(ctx, req)
}

// GetConfig returns the current CORS configuration.
func (cm *CORSMiddleware) GetConfig() CORSConfig {
	return cm.config
}

// SetConfig updates the CORS configuration.
func (cm *CORSMiddleware) SetConfig(config CORSConfig) {
	cm.config = config
	cm.GetLogger().Info("CORS middleware configuration updated")
}

// GetMetrics returns CORS metrics.
func (cm *CORSMiddleware) GetMetrics() map[string]interface{} {
	corsRequests := atomic.LoadInt64(&cm.corsRequests)
	corsAllowed := atomic.LoadInt64(&cm.corsAllowed)

	var allowedRate float64
	if corsRequests > 0 {
		allowedRate = float64(corsAllowed) / float64(corsRequests) * 100.0
	}

	return map[string]interface{}{
		"preflight_requests": atomic.LoadInt64(&cm.preflightRequests),
		"cors_requests":      corsRequests,
		"cors_allowed":       corsAllowed,
		"cors_blocked":       atomic.LoadInt64(&cm.corsBlocked),
		"origin_checks":      atomic.LoadInt64(&cm.originChecks),
		"allowed_rate":       allowedRate,
		"uptime":             time.Since(cm.startTime),
	}
}

// extractCORSRequest extracts CORS-related information from the request.
func (cm *CORSMiddleware) extractCORSRequest(req interface{}) *CORSRequest {
	corsReq := &CORSRequest{}

	if httpReq, ok := req.(map[string]interface{}); ok {
		// Extract method
		if method, exists := httpReq["method"]; exists {
			if m, ok := method.(string); ok {
				corsReq.Method = strings.ToUpper(m)
			}
		}

		// Extract path
		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				corsReq.Path = p
			}
		}

		// Extract headers
		if headers, exists := httpReq["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				// Check for Origin header
				if origin, hasOrigin := h["Origin"]; hasOrigin {
					corsReq.Origin = origin
					corsReq.IsCORSRequest = true
				}

				// Check for preflight request
				if corsReq.Method == "OPTIONS" && corsReq.IsCORSRequest {
					if _, hasRequestMethod := h["Access-Control-Request-Method"]; hasRequestMethod {
						corsReq.IsPreflightReq = true

						// Extract requested headers
						if requestHeaders, hasRequestHeaders := h["Access-Control-Request-Headers"]; hasRequestHeaders {
							corsReq.RequestedHeaders = cm.parseHeaderList(requestHeaders)
						}
					}
				}
			}
		}
	}

	return corsReq
}

// shouldSkipCORS determines if CORS processing should be skipped for a path.
func (cm *CORSMiddleware) shouldSkipCORS(path string) bool {
	for _, skipPath := range cm.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// handlePreflightRequest handles CORS preflight (OPTIONS) requests.
func (cm *CORSMiddleware) handlePreflightRequest(ctx context.Context, corsReq *CORSRequest, req interface{}, next MiddlewareNext) (interface{}, error) {
	cm.GetLogger().Debug("Handling CORS preflight request from origin: %s", corsReq.Origin)

	// Check if origin is allowed
	if !cm.isOriginAllowed(corsReq.Origin) {
		atomic.AddInt64(&cm.corsBlocked, 1)
		cm.GetLogger().Warn("CORS preflight blocked for origin: %s", corsReq.Origin)
		return cm.createCORSErrorResponse("Origin not allowed")
	}

	// Check if method is allowed
	if !cm.isMethodAllowed(corsReq.Method) {
		atomic.AddInt64(&cm.corsBlocked, 1)
		cm.GetLogger().Warn("CORS preflight blocked for method: %s", corsReq.Method)
		return cm.createCORSErrorResponse("Method not allowed")
	}

	// Check if requested headers are allowed
	if !cm.areHeadersAllowed(corsReq.RequestedHeaders) {
		atomic.AddInt64(&cm.corsBlocked, 1)
		cm.GetLogger().Warn("CORS preflight blocked for headers: %v", corsReq.RequestedHeaders)
		return cm.createCORSErrorResponse("Headers not allowed")
	}

	atomic.AddInt64(&cm.corsAllowed, 1)

	// Create CORS response
	corsResp := cm.createCORSResponse(corsReq)

	// Create response with CORS headers
	response := map[string]interface{}{
		"status_code": 204, // No Content for preflight
		"headers":     cm.corsResponseToHeaders(corsResp),
		"body":        "",
	}

	cm.GetLogger().Debug("CORS preflight approved for origin: %s", corsReq.Origin)

	// If OptionsPassthrough is enabled, continue to next middleware
	if cm.config.OptionsPassthrough {
		// Add CORS headers to context for next middleware
		ctxWithCORS := context.WithValue(ctx, "cors_headers", corsResp)
		return next(ctxWithCORS, req)
	}

	return response, nil
}

// handleCORSRequest handles actual CORS requests (non-preflight).
func (cm *CORSMiddleware) handleCORSRequest(ctx context.Context, corsReq *CORSRequest, req interface{}, next MiddlewareNext) (interface{}, error) {
	cm.GetLogger().Debug("Handling CORS request from origin: %s", corsReq.Origin)

	// Check if origin is allowed
	if !cm.isOriginAllowed(corsReq.Origin) {
		atomic.AddInt64(&cm.corsBlocked, 1)
		cm.GetLogger().Warn("CORS request blocked for origin: %s", corsReq.Origin)
		return cm.createCORSErrorResponse("Origin not allowed")
	}

	atomic.AddInt64(&cm.corsAllowed, 1)

	// Create CORS response
	corsResp := cm.createCORSResponse(corsReq)

	// Add CORS headers to context for response modification
	ctxWithCORS := context.WithValue(ctx, "cors_headers", corsResp)

	// Process request with next middleware
	resp, err := next(ctxWithCORS, req)
	if err != nil {
		return resp, err
	}

	// Add CORS headers to response
	if httpResp, ok := resp.(map[string]interface{}); ok {
		if headers, exists := httpResp["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				corsHeaders := cm.corsResponseToHeaders(corsResp)
				for key, value := range corsHeaders {
					h[key] = value
				}
			}
		} else {
			httpResp["headers"] = cm.corsResponseToHeaders(corsResp)
		}
	}

	cm.GetLogger().Debug("CORS request processed for origin: %s", corsReq.Origin)
	return resp, nil
}

// isOriginAllowed checks if the given origin is allowed.
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	atomic.AddInt64(&cm.originChecks, 1)

	if cm.config.AllowAllOrigins {
		return true
	}

	// Use custom origin function if provided
	if cm.config.AllowOriginFunc != nil {
		return cm.config.AllowOriginFunc(origin)
	}

	// Check against allowed origins list
	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if cm.matchOrigin(origin, allowedOrigin) {
			return true
		}
	}

	return false
}

// matchOrigin checks if the origin matches the allowed origin pattern.
func (cm *CORSMiddleware) matchOrigin(origin, allowedOrigin string) bool {
	if allowedOrigin == "*" {
		return true
	}

	if allowedOrigin == origin {
		return true
	}

	// Support for wildcard subdomains (e.g., *.example.com)
	if strings.HasPrefix(allowedOrigin, "*.") {
		domain := strings.TrimPrefix(allowedOrigin, "*.")
		if strings.HasSuffix(origin, "."+domain) {
			return true
		}
	}

	return false
}

// isMethodAllowed checks if the given method is allowed.
func (cm *CORSMiddleware) isMethodAllowed(method string) bool {
	for _, allowedMethod := range cm.config.AllowedMethods {
		if strings.ToUpper(allowedMethod) == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

// areHeadersAllowed checks if all requested headers are allowed.
func (cm *CORSMiddleware) areHeadersAllowed(requestedHeaders []string) bool {
	for _, requestedHeader := range requestedHeaders {
		if !cm.isHeaderAllowed(requestedHeader) {
			return false
		}
	}
	return true
}

// isHeaderAllowed checks if a specific header is allowed.
func (cm *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(strings.TrimSpace(header))

	// Always allow simple headers
	simpleHeaders := []string{
		"accept", "accept-language", "content-language", "content-type",
	}

	for _, simpleHeader := range simpleHeaders {
		if header == simpleHeader {
			return true
		}
	}

	// Check against configured allowed headers
	for _, allowedHeader := range cm.config.AllowedHeaders {
		if strings.ToLower(allowedHeader) == header {
			return true
		}
	}

	return false
}

// createCORSResponse creates a CORS response based on the request and configuration.
func (cm *CORSMiddleware) createCORSResponse(corsReq *CORSRequest) *CORSResponse {
	corsResp := &CORSResponse{}

	// Set Allow-Origin
	if cm.config.AllowAllOrigins && !cm.config.AllowCredentials {
		corsResp.AllowOrigin = "*"
	} else {
		corsResp.AllowOrigin = corsReq.Origin
	}

	// Set Allow-Methods for preflight
	if corsReq.IsPreflightReq {
		corsResp.AllowMethods = strings.Join(cm.config.AllowedMethods, ", ")
	}

	// Set Allow-Headers for preflight
	if corsReq.IsPreflightReq && len(cm.config.AllowedHeaders) > 0 {
		corsResp.AllowHeaders = strings.Join(cm.config.AllowedHeaders, ", ")
	}

	// Set Expose-Headers
	if len(cm.config.ExposedHeaders) > 0 {
		corsResp.ExposeHeaders = strings.Join(cm.config.ExposedHeaders, ", ")
	}

	// Set Allow-Credentials
	if cm.config.AllowCredentials {
		corsResp.AllowCredentials = "true"
	}

	// Set Max-Age for preflight
	if corsReq.IsPreflightReq && cm.config.MaxAge > 0 {
		corsResp.MaxAge = fmt.Sprintf("%d", cm.config.MaxAge)
	}

	// Set Vary headers
	if cm.config.VaryByOrigin {
		corsResp.VaryHeaders = append(corsResp.VaryHeaders, "Origin")
	}

	return corsResp
}

// corsResponseToHeaders converts CORS response to HTTP headers map.
func (cm *CORSMiddleware) corsResponseToHeaders(corsResp *CORSResponse) map[string]string {
	headers := make(map[string]string)

	if corsResp.AllowOrigin != "" {
		headers["Access-Control-Allow-Origin"] = corsResp.AllowOrigin
	}

	if corsResp.AllowMethods != "" {
		headers["Access-Control-Allow-Methods"] = corsResp.AllowMethods
	}

	if corsResp.AllowHeaders != "" {
		headers["Access-Control-Allow-Headers"] = corsResp.AllowHeaders
	}

	if corsResp.ExposeHeaders != "" {
		headers["Access-Control-Expose-Headers"] = corsResp.ExposeHeaders
	}

	if corsResp.AllowCredentials != "" {
		headers["Access-Control-Allow-Credentials"] = corsResp.AllowCredentials
	}

	if corsResp.MaxAge != "" {
		headers["Access-Control-Max-Age"] = corsResp.MaxAge
	}

	if len(corsResp.VaryHeaders) > 0 {
		headers["Vary"] = strings.Join(corsResp.VaryHeaders, ", ")
	}

	return headers
}

// parseHeaderList parses a comma-separated list of headers.
func (cm *CORSMiddleware) parseHeaderList(headerStr string) []string {
	if headerStr == "" {
		return []string{}
	}

	headers := strings.Split(headerStr, ",")
	var result []string

	for _, header := range headers {
		header = strings.TrimSpace(header)
		if header != "" {
			result = append(result, header)
		}
	}

	return result
}

// createCORSErrorResponse creates an error response for CORS violations.
func (cm *CORSMiddleware) createCORSErrorResponse(message string) (interface{}, error) {
	return map[string]interface{}{
		"status_code": 403, // Forbidden
		"headers": map[string]string{
			"Content-Type": "application/json",
		},
		"body": fmt.Sprintf(`{"error": "CORS policy violation", "message": "%s"}`, message),
	}, nil
}

// Reset resets all metrics.
func (cm *CORSMiddleware) Reset() {
	atomic.StoreInt64(&cm.preflightRequests, 0)
	atomic.StoreInt64(&cm.corsRequests, 0)
	atomic.StoreInt64(&cm.corsAllowed, 0)
	atomic.StoreInt64(&cm.corsBlocked, 0)
	atomic.StoreInt64(&cm.originChecks, 0)
	cm.startTime = time.Now()
	cm.GetLogger().Info("CORS middleware metrics reset")
}

// AddAllowedOrigin adds an origin to the allowed origins list.
func (cm *CORSMiddleware) AddAllowedOrigin(origin string) {
	cm.config.AllowedOrigins = append(cm.config.AllowedOrigins, origin)
	cm.GetLogger().Info("Added allowed origin: %s", origin)
}

// RemoveAllowedOrigin removes an origin from the allowed origins list.
func (cm *CORSMiddleware) RemoveAllowedOrigin(origin string) {
	for i, allowedOrigin := range cm.config.AllowedOrigins {
		if allowedOrigin == origin {
			cm.config.AllowedOrigins = append(cm.config.AllowedOrigins[:i], cm.config.AllowedOrigins[i+1:]...)
			cm.GetLogger().Info("Removed allowed origin: %s", origin)
			return
		}
	}
}

// SetAllowAllOrigins enables or disables allowing all origins.
func (cm *CORSMiddleware) SetAllowAllOrigins(allow bool) {
	cm.config.AllowAllOrigins = allow
	if allow {
		cm.GetLogger().Info("Enabled allow all origins")
	} else {
		cm.GetLogger().Info("Disabled allow all origins")
	}
}

// SetAllowCredentials enables or disables credentials support.
func (cm *CORSMiddleware) SetAllowCredentials(allow bool) {
	cm.config.AllowCredentials = allow
	if allow {
		cm.GetLogger().Info("Enabled credentials support")
	} else {
		cm.GetLogger().Info("Disabled credentials support")
	}
}
