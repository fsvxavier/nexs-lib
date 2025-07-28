package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

// PaginationConfig holds pagination middleware configuration
type PaginationConfig struct {
	// Service is the pagination service to use
	Service *pagination.PaginationService

	// SortableFields maps route patterns to allowed sortable fields
	SortableFields map[string][]string

	// DefaultSortableFields used when no specific fields are configured
	DefaultSortableFields []string

	// ContextKey is the key used to store pagination params in request context
	ContextKey string

	// ErrorHandler handles pagination validation errors
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// SkipPaths defines paths that should skip pagination processing
	SkipPaths []string
}

// DefaultPaginationConfig returns a default pagination middleware configuration
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		Service:               pagination.NewPaginationService(nil),
		SortableFields:        make(map[string][]string),
		DefaultSortableFields: []string{"id", "created_at", "updated_at"},
		ContextKey:            "pagination_params",
		ErrorHandler:          defaultErrorHandler,
		SkipPaths:             []string{"/health", "/metrics", "/favicon.ico"},
	}
}

// PaginationMiddleware creates a middleware that automatically parses and validates pagination parameters
func PaginationMiddleware(cfg *PaginationConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultPaginationConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip processing for certain paths
			if shouldSkipPath(r.URL.Path, cfg.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Parse pagination parameters from query string
			params, err := cfg.Service.ParseRequest(r.URL.Query(), getSortableFields(r.URL.Path, cfg)...)
			if err != nil {
				cfg.ErrorHandler(w, r, err)
				return
			}

			// Store pagination parameters in request context
			ctx := context.WithValue(r.Context(), cfg.ContextKey, params)
			r = r.WithContext(ctx)

			// Add pagination headers for convenience
			w.Header().Set("X-Pagination-Page", strconv.Itoa(params.Page))
			w.Header().Set("X-Pagination-Limit", strconv.Itoa(params.Limit))
			if params.SortField != "" {
				w.Header().Set("X-Pagination-Sort", params.SortField)
				w.Header().Set("X-Pagination-Order", params.SortOrder)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetPaginationParams extracts pagination parameters from request context
func GetPaginationParams(r *http.Request, contextKey ...string) *interfaces.PaginationParams {
	key := "pagination_params"
	if len(contextKey) > 0 {
		key = contextKey[0]
	}

	if params, ok := r.Context().Value(key).(*interfaces.PaginationParams); ok {
		return params
	}
	return nil
}

// PaginatedResponseMiddleware creates a middleware that wraps responses in pagination format
func PaginatedResponseMiddleware(service *pagination.PaginationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom ResponseWriter to capture the response
			wrapper := &responseWrapper{
				ResponseWriter: w,
				body:           make([]byte, 0),
				statusCode:     http.StatusOK,
			}

			// Execute the handler
			next.ServeHTTP(wrapper, r)

			// Get pagination parameters from context
			params := GetPaginationParams(r)
			if params == nil {
				// No pagination params, write original response
				w.WriteHeader(wrapper.statusCode)
				w.Write(wrapper.body)
				return
			}

			// Try to parse the response body as JSON
			var content interface{}
			if err := json.Unmarshal(wrapper.body, &content); err != nil {
				// Not JSON, write original response
				w.WriteHeader(wrapper.statusCode)
				w.Write(wrapper.body)
				return
			}

			// Check if we have a total count header
			totalStr := w.Header().Get("X-Total-Count")
			total := 0
			if totalStr != "" {
				if parsed, err := strconv.Atoi(totalStr); err == nil {
					total = parsed
				}
			}

			// Create paginated response
			paginatedResponse := service.CreateResponse(content, params, total)

			// Write the paginated response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(wrapper.statusCode)
			json.NewEncoder(w).Encode(paginatedResponse)
		})
	}
}

// shouldSkipPath checks if a path should skip pagination processing
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getSortableFields returns the sortable fields for a given path
func getSortableFields(path string, cfg *PaginationConfig) []string {
	// Try to find exact match first
	if fields, exists := cfg.SortableFields[path]; exists {
		return fields
	}

	// Try pattern matching
	for pattern, fields := range cfg.SortableFields {
		if matchesPattern(path, pattern) {
			return fields
		}
	}

	// Return default fields
	return cfg.DefaultSortableFields
}

// matchesPattern performs simple pattern matching (supports wildcards)
func matchesPattern(path, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1])
		}
	}

	return path == pattern
}

// defaultErrorHandler handles pagination validation errors
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "pagination_validation_error",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// responseWrapper captures HTTP response for processing
type responseWrapper struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// ConfigureRoute configures pagination for a specific route
func (cfg *PaginationConfig) ConfigureRoute(pattern string, sortableFields []string) {
	cfg.SortableFields[pattern] = sortableFields
}

// SetErrorHandler sets a custom error handler
func (cfg *PaginationConfig) SetErrorHandler(handler func(http.ResponseWriter, *http.Request, error)) {
	cfg.ErrorHandler = handler
}

// AddSkipPath adds a path to skip pagination processing
func (cfg *PaginationConfig) AddSkipPath(path string) {
	cfg.SkipPaths = append(cfg.SkipPaths, path)
}

// WithHooks configures hooks for the pagination service
func (cfg *PaginationConfig) WithHooks() *HookConfig {
	return &HookConfig{cfg: cfg}
}

// HookConfig provides a fluent interface for configuring pagination hooks
type HookConfig struct {
	cfg *PaginationConfig
}

// PreValidation adds a pre-validation hook
func (hc *HookConfig) PreValidation(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("pre-validation", hook)
	return hc
}

// PostValidation adds a post-validation hook
func (hc *HookConfig) PostValidation(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("post-validation", hook)
	return hc
}

// PreQuery adds a pre-query hook
func (hc *HookConfig) PreQuery(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("pre-query", hook)
	return hc
}

// PostQuery adds a post-query hook
func (hc *HookConfig) PostQuery(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("post-query", hook)
	return hc
}

// PreResponse adds a pre-response hook
func (hc *HookConfig) PreResponse(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("pre-response", hook)
	return hc
}

// PostResponse adds a post-response hook
func (hc *HookConfig) PostResponse(hook interfaces.Hook) *HookConfig {
	hc.cfg.Service.AddHook("post-response", hook)
	return hc
}

// Done returns the pagination configuration
func (hc *HookConfig) Done() *PaginationConfig {
	return hc.cfg
}
