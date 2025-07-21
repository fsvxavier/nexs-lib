// Package tenantid provides tenant ID middleware implementation for multi-tenant applications.
package tenantid

import (
	"context"
	"net/http"
	"strings"
)

// Config represents tenant ID configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass tenant ID extraction.
	SkipPaths []string
	// HeaderName is the primary header name to use for tenant ID.
	HeaderName string
	// ContextKey is the key to use when storing tenant ID in context.
	ContextKey string
	// AlternativeHeaders are alternative header names to check for existing tenant ID.
	AlternativeHeaders []string
	// QueryParam is the query parameter name to check for tenant ID.
	QueryParam string
	// DefaultTenant is the default tenant ID to use if none is found.
	DefaultTenant string
	// Required indicates if tenant ID is required (return error if missing).
	Required bool
	// CaseSensitive indicates if tenant ID should be case sensitive.
	CaseSensitive bool
}

// IsEnabled returns true if the middleware is enabled.
func (c Config) IsEnabled() bool {
	return c.Enabled
}

// ShouldSkip returns true if the given path should be skipped.
func (c Config) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// DefaultConfig returns a default tenant ID configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:    true,
		HeaderName: "X-Tenant-Id",
		ContextKey: "tenant_id",
		QueryParam: "tenant_id",
		AlternativeHeaders: []string{
			"Client-Id",
			"client-id",
			"Client-ID",
			"X-Client-Id",
			"x-client-id",
			"Tenant-Id",
			"tenant-id",
		},
		Required:      false,
		CaseSensitive: true,
	}
}

// Middleware implements tenant ID middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new tenant ID middleware.
func NewMiddleware(config Config) *Middleware {
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-Id"
	}
	if config.ContextKey == "" {
		config.ContextKey = "tenant_id"
	}

	return &Middleware{
		config: config,
	}
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "tenantid"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 100 // High priority, should run early
}

// Wrap implements the interfaces.Middleware interface.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if m.config.ShouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		tenantID := m.extractTenantID(r)

		// Normalize tenant ID case if not case sensitive
		if !m.config.CaseSensitive && tenantID != "" {
			tenantID = strings.ToLower(tenantID)
		}

		// Use default tenant if none found
		if tenantID == "" && m.config.DefaultTenant != "" {
			tenantID = m.config.DefaultTenant
		}

		// Check if tenant ID is required
		if m.config.Required && tenantID == "" {
			http.Error(w, "Tenant ID is required", http.StatusBadRequest)
			return
		}

		// Add tenant ID to request context
		if tenantID != "" {
			ctx := context.WithValue(r.Context(), m.config.ContextKey, tenantID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// extractTenantID extracts tenant ID from request headers and query parameters.
func (m *Middleware) extractTenantID(r *http.Request) string {
	// Try primary header first
	if tenantID := r.Header.Get(m.config.HeaderName); tenantID != "" {
		return strings.TrimSpace(tenantID)
	}

	// Try alternative headers
	for _, altHeader := range m.config.AlternativeHeaders {
		if tenantID := r.Header.Get(altHeader); tenantID != "" {
			return strings.TrimSpace(tenantID)
		}
	}

	// Try query parameter
	if m.config.QueryParam != "" {
		if tenantID := r.URL.Query().Get(m.config.QueryParam); tenantID != "" {
			return strings.TrimSpace(tenantID)
		}
	}

	return ""
}

// GetTenantIDFromContext extracts tenant ID from context.
func GetTenantIDFromContext(ctx context.Context, contextKey string) string {
	if contextKey == "" {
		contextKey = "tenant_id"
	}

	if tenantID, ok := ctx.Value(contextKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetTenantIDFromRequest extracts tenant ID from request context or directly from request.
func GetTenantIDFromRequest(r *http.Request, config Config) string {
	// Try context first
	if tenantID := GetTenantIDFromContext(r.Context(), config.ContextKey); tenantID != "" {
		return tenantID
	}

	// Create middleware instance to use extraction logic
	middleware := NewMiddleware(config)
	return middleware.extractTenantID(r)
}
