// Package cors provides CORS (Cross-Origin Resource Sharing) middleware implementation.
package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Config represents CORS configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass CORS.
	SkipPaths []string
	// AllowedOrigins is a list of allowed origins. Use ["*"] to allow all origins.
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods.
	AllowedMethods []string
	// AllowedHeaders is a list of allowed headers.
	AllowedHeaders []string
	// ExposedHeaders is a list of headers exposed to the browser.
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed.
	AllowCredentials bool
	// MaxAge indicates how long (in seconds) the browser can cache preflight results.
	MaxAge time.Duration
	// OptionsPassthrough passes preflight requests to the next handler.
	OptionsPassthrough bool
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

// DefaultConfig returns a default CORS configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders:     []string{},
		AllowCredentials:   false,
		MaxAge:             12 * time.Hour,
		OptionsPassthrough: false,
	}
}

// Middleware implements CORS middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new CORS middleware.
func NewMiddleware(config Config) *Middleware {
	return &Middleware{
		config: config,
	}
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

		origin := r.Header.Get("Origin")

		// Handle preflight requests
		if r.Method == http.MethodOptions && origin != "" {
			m.handlePreflight(w, r, origin)
			if !m.config.OptionsPassthrough {
				return
			}
		} else {
			// Handle simple requests
			m.handleSimpleRequest(w, r, origin)
		}

		next.ServeHTTP(w, r)
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "cors"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 100 // CORS should happen very early
}

// handlePreflight handles preflight requests.
func (m *Middleware) handlePreflight(w http.ResponseWriter, r *http.Request, origin string) {
	// Check if origin is allowed
	if !m.isOriginAllowed(origin) {
		return
	}

	// Set Access-Control-Allow-Origin
	w.Header().Set("Access-Control-Allow-Origin", m.getAllowedOrigin(origin))

	// Set Access-Control-Allow-Methods
	if len(m.config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
	}

	// Set Access-Control-Allow-Headers
	requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
	if requestedHeaders != "" && m.areHeadersAllowed(requestedHeaders) {
		w.Header().Set("Access-Control-Allow-Headers", requestedHeaders)
	} else if len(m.config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
	}

	// Set Access-Control-Allow-Credentials
	if m.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Set Access-Control-Max-Age
	if m.config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(m.config.MaxAge.Seconds())))
	}

	// Set Vary header
	w.Header().Add("Vary", "Origin")
	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")

	w.WriteHeader(http.StatusNoContent)
}

// handleSimpleRequest handles simple requests.
func (m *Middleware) handleSimpleRequest(w http.ResponseWriter, r *http.Request, origin string) {
	// Check if origin is allowed
	if origin != "" && !m.isOriginAllowed(origin) {
		return
	}

	// Set Access-Control-Allow-Origin
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", m.getAllowedOrigin(origin))
	}

	// Set Access-Control-Expose-Headers
	if len(m.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.config.ExposedHeaders, ", "))
	}

	// Set Access-Control-Allow-Credentials
	if m.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Set Vary header
	if origin != "" {
		w.Header().Add("Vary", "Origin")
	}
}

// isOriginAllowed checks if the origin is allowed.
func (m *Middleware) isOriginAllowed(origin string) bool {
	if len(m.config.AllowedOrigins) == 0 {
		return false
	}

	for _, allowedOrigin := range m.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// getAllowedOrigin returns the appropriate Access-Control-Allow-Origin value.
func (m *Middleware) getAllowedOrigin(origin string) string {
	if len(m.config.AllowedOrigins) == 1 && m.config.AllowedOrigins[0] == "*" {
		if m.config.AllowCredentials {
			// When credentials are allowed, we can't use "*"
			return origin
		}
		return "*"
	}
	return origin
}

// areHeadersAllowed checks if the requested headers are allowed.
func (m *Middleware) areHeadersAllowed(requestedHeaders string) bool {
	if len(m.config.AllowedHeaders) == 0 {
		return false
	}

	headers := strings.Split(requestedHeaders, ",")
	for _, header := range headers {
		header = strings.TrimSpace(header)
		header = strings.ToLower(header)

		allowed := false
		for _, allowedHeader := range m.config.AllowedHeaders {
			if strings.ToLower(allowedHeader) == header {
				allowed = true
				break
			}
		}

		if !allowed {
			return false
		}
	}

	return true
}
