// Package contenttype provides content type validation middleware implementation.
package contenttype

import (
	"net/http"
	"strings"
)

// Config represents content type validation configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass content type validation.
	SkipPaths []string
	// AllowedContentTypes contains the list of allowed content types.
	AllowedContentTypes []string
	// RestrictMethods contains HTTP methods that should be validated.
	RestrictMethods []string
	// IgnoreMethods contains HTTP methods that should be ignored.
	IgnoreMethods []string
	// RequireContentType indicates if Content-Type header is required for restricted methods.
	RequireContentType bool
	// CaseSensitive indicates if content type matching should be case sensitive.
	CaseSensitive bool
	// StrictMatching requires exact content type match (no partial matching).
	StrictMatching bool
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

// ShouldValidateMethod returns true if the given method should be validated.
func (c Config) ShouldValidateMethod(method string) bool {
	// Check ignore methods first
	for _, ignoreMethod := range c.IgnoreMethods {
		if strings.EqualFold(method, ignoreMethod) {
			return false
		}
	}

	// If restrict methods is empty, validate all methods (except ignored)
	if len(c.RestrictMethods) == 0 {
		return true
	}

	// Check if method is in restrict list
	for _, restrictMethod := range c.RestrictMethods {
		if strings.EqualFold(method, restrictMethod) {
			return true
		}
	}

	return false
}

// DefaultConfig returns a default content type validation configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		AllowedContentTypes: []string{
			"application/json",
			"application/json; charset=utf-8",
			"application/xml",
			"application/xml; charset=utf-8",
			"text/xml",
			"text/xml; charset=utf-8",
		},
		RestrictMethods: []string{
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
		},
		IgnoreMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodDelete,
		},
		RequireContentType: true,
		CaseSensitive:      false,
		StrictMatching:     false,
	}
}

// Middleware implements content type validation middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new content type validation middleware.
func NewMiddleware(config Config) *Middleware {
	return &Middleware{
		config: config,
	}
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "contenttype"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 150 // Medium-high priority, should run before business logic
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

		if !m.config.ShouldValidateMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")

		// Check if content type is required
		if m.config.RequireContentType && contentType == "" {
			http.Error(w, "Content-Type header is required", http.StatusBadRequest)
			return
		}

		// Skip validation if no content type and not required
		if contentType == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Validate content type
		if !m.isAllowedContentType(contentType) {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowedContentType checks if the given content type is allowed.
func (m *Middleware) isAllowedContentType(contentType string) bool {
	if len(m.config.AllowedContentTypes) == 0 {
		return true // Allow all if no restrictions
	}

	// Normalize content type for comparison
	normalizedContentType := contentType
	if !m.config.CaseSensitive {
		normalizedContentType = strings.ToLower(strings.TrimSpace(contentType))
	} else {
		normalizedContentType = strings.TrimSpace(contentType)
	}

	for _, allowedType := range m.config.AllowedContentTypes {
		normalizedAllowedType := allowedType
		if !m.config.CaseSensitive {
			normalizedAllowedType = strings.ToLower(strings.TrimSpace(allowedType))
		} else {
			normalizedAllowedType = strings.TrimSpace(allowedType)
		}

		// Check for match based on strict matching setting
		if m.config.StrictMatching {
			if normalizedContentType == normalizedAllowedType {
				return true
			}
		} else {
			// Partial matching - check if content type starts with or contains allowed type
			if strings.Contains(normalizedContentType, normalizedAllowedType) ||
				strings.HasPrefix(normalizedContentType, normalizedAllowedType) {
				return true
			}
		}
	}

	return false
}

// CreateForMethods creates a content type middleware for specific HTTP methods.
func CreateForMethods(allowedContentTypes []string, methods ...string) *Middleware {
	config := DefaultConfig()
	config.AllowedContentTypes = allowedContentTypes
	config.RestrictMethods = methods
	config.IgnoreMethods = nil // Clear ignore methods

	return NewMiddleware(config)
}

// CreateJSONOnly creates a content type middleware that only allows JSON.
func CreateJSONOnly(methods ...string) *Middleware {
	return CreateForMethods([]string{
		"application/json",
		"application/json; charset=utf-8",
	}, methods...)
}

// CreateXMLOnly creates a content type middleware that only allows XML.
func CreateXMLOnly(methods ...string) *Middleware {
	return CreateForMethods([]string{
		"application/xml",
		"application/xml; charset=utf-8",
		"text/xml",
		"text/xml; charset=utf-8",
	}, methods...)
}
