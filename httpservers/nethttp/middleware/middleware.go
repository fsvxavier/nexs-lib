package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// Logger middleware logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		rw := newResponseWriter(w)

		// Get or generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			w.Header().Set("X-Request-ID", requestID)
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate execution time
		duration := time.Since(start)

		// Log request
		fmt.Printf("[%s] %s %s - %d - %s - %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			requestID,
		)
	})
}

// Recover middleware recovers from panics
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				fmt.Printf("PANIC RECOVERED: %v\n%s\n", err, debug.Stack())

				// Return a 500 response
				apiError := common.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal Server Error")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(apiError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Tracing adds tracing capability
func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate trace ID if not present
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			w.Header().Set("X-Trace-ID", traceID)
		}

		// Set span ID
		spanID := uuid.New().String()
		w.Header().Set("X-Span-ID", spanID)

		next.ServeHTTP(w, r)
	})
}

// RegisterPprof registers pprof handlers
func RegisterPprof(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	mux.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
}

// RegisterMetrics registers metrics handlers
func RegisterMetrics(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Simple metrics endpoint - in a real implementation, you would use prometheus
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","metrics":{}}`))
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
