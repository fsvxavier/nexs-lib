// Package retry provides retry policy middleware implementation.
package retry

import (
	"math"
	"net/http"
	"time"
)

// Config represents retry configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass retry.
	SkipPaths []string
	// MaxRetries is the maximum number of retries.
	MaxRetries int
	// InitialDelay is the initial delay between retries.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff.
	BackoffMultiplier float64
	// RetryableStatus contains HTTP status codes that should trigger a retry.
	RetryableStatus []int
	// ShouldRetry is a custom function to determine if a request should be retried.
	ShouldRetry func(*http.Request, int, error) bool
	// OnRetry is called before each retry attempt.
	OnRetry func(*http.Request, int, time.Duration)
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

// DefaultConfig returns a default retry configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatus: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		ShouldRetry: nil, // Will use default logic
		OnRetry:     nil, // No-op by default
	}
}

// Middleware implements retry middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new retry middleware.
func NewMiddleware(config Config) *Middleware {
	if config.ShouldRetry == nil {
		config.ShouldRetry = defaultShouldRetry(config.RetryableStatus)
	}
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

		// Only retry for idempotent methods
		if !isIdempotentMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		var lastStatusCode int
		var lastError error

		for attempt := 0; attempt <= m.config.MaxRetries; attempt++ {
			if attempt > 0 {
				// Calculate delay for retry
				delay := m.calculateDelay(attempt)

				// Call retry callback
				if m.config.OnRetry != nil {
					m.config.OnRetry(r, attempt, delay)
				}

				// Wait before retry
				time.Sleep(delay)
			}

			// Create a response recorder to capture the response
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				headerWritten:  false,
			}

			// Execute the request
			next.ServeHTTP(recorder, r)

			lastStatusCode = recorder.statusCode
			lastError = nil

			// Check if we should retry
			if !m.config.ShouldRetry(r, lastStatusCode, lastError) {
				// Success or non-retryable error, write the response
				if !recorder.headerWritten {
					recorder.WriteHeader(lastStatusCode)
				}
				return
			}

			// This was a retryable error, continue to next attempt
			// Reset the recorder for next attempt
			recorder.reset()
		}

		// All retries exhausted, return the last response
		w.WriteHeader(lastStatusCode)
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "retry"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 400 // Retry should happen after most other middleware
}

// calculateDelay calculates the delay for the given attempt using exponential backoff.
func (m *Middleware) calculateDelay(attempt int) time.Duration {
	delay := float64(m.config.InitialDelay) * math.Pow(m.config.BackoffMultiplier, float64(attempt-1))

	if delay > float64(m.config.MaxDelay) {
		delay = float64(m.config.MaxDelay)
	}

	return time.Duration(delay)
}

// isIdempotentMethod checks if the HTTP method is idempotent.
func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// defaultShouldRetry returns a default retry function based on status codes.
func defaultShouldRetry(retryableStatus []int) func(*http.Request, int, error) bool {
	return func(r *http.Request, statusCode int, err error) bool {
		// Retry on error
		if err != nil {
			return true
		}

		// Check if status code is retryable
		for _, code := range retryableStatus {
			if statusCode == code {
				return true
			}
		}

		return false
	}
}

// responseRecorder records response data without writing to the actual response.
type responseRecorder struct {
	http.ResponseWriter
	statusCode    int
	body          []byte
	headerWritten bool
}

// WriteHeader records the status code.
func (rr *responseRecorder) WriteHeader(code int) {
	if rr.headerWritten {
		return
	}
	rr.statusCode = code
	rr.headerWritten = true
}

// Write records the response body.
func (rr *responseRecorder) Write(data []byte) (int, error) {
	if !rr.headerWritten {
		rr.WriteHeader(http.StatusOK)
	}
	rr.body = append(rr.body, data...)
	return len(data), nil
}

// reset resets the recorder for a new attempt.
func (rr *responseRecorder) reset() {
	rr.statusCode = http.StatusOK
	rr.body = nil
	rr.headerWritten = false
}
