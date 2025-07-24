// Package compression provides HTTP response compression middleware implementation.
package compression

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Config represents compression configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass compression.
	SkipPaths []string
	// Level is the compression level (0-9 for gzip, -2 to 9 for deflate).
	Level int
	// MinSize is the minimum response size to compress (in bytes).
	MinSize int64
	// Types is the list of MIME types to compress.
	Types []string
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

// DefaultConfig returns a default compression configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Level:   gzip.DefaultCompression,
		MinSize: 1024, // 1KB
		Types: []string{
			"text/html",
			"text/css",
			"text/javascript",
			"text/plain",
			"application/json",
			"application/javascript",
			"application/xml",
			"application/rss+xml",
			"application/atom+xml",
			"image/svg+xml",
		},
	}
}

// Middleware implements compression middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new compression middleware.
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

		// Check if client accepts compression
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if acceptEncoding == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Find the best compression algorithm
		algorithm := m.getBestEncoding(acceptEncoding)
		if algorithm == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap response writer with compression
		cw := &compressionWriter{
			ResponseWriter: w,
			algorithm:      algorithm,
			config:         m.config,
			headerWritten:  false,
		}

		// Set Vary header
		w.Header().Add("Vary", "Accept-Encoding")

		next.ServeHTTP(cw, r)

		// Close the compressor if it was used
		if cw.compressor != nil {
			cw.compressor.Close()
		}
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "compression"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 800 // Compression should happen late in the chain
}

// getBestEncoding returns the best compression encoding from Accept-Encoding header.
func (m *Middleware) getBestEncoding(acceptEncoding string) string {
	// Simple implementation that supports gzip and deflate
	supportedAlgorithms := []string{"gzip", "deflate"}

	encodings := strings.Split(acceptEncoding, ",")
	for _, encoding := range encodings {
		encoding = strings.TrimSpace(encoding)

		// Handle quality values (e.g., "gzip;q=0.8")
		parts := strings.Split(encoding, ";")
		algorithm := strings.TrimSpace(parts[0])

		// Check if we support this algorithm
		for _, supported := range supportedAlgorithms {
			if algorithm == supported {
				return algorithm
			}
		}
	}

	return ""
}

// compressionWriter wraps http.ResponseWriter to compress responses.
type compressionWriter struct {
	http.ResponseWriter
	algorithm     string
	config        Config
	compressor    io.WriteCloser
	headerWritten bool
	size          int64
}

// WriteHeader writes the status code and decides whether to compress.
func (cw *compressionWriter) WriteHeader(code int) {
	if cw.headerWritten {
		return
	}
	cw.headerWritten = true

	// Don't compress certain status codes
	if code < 200 || code >= 300 {
		cw.ResponseWriter.WriteHeader(code)
		return
	}

	// Check content type
	contentType := cw.Header().Get("Content-Type")
	if !cw.shouldCompress(contentType) {
		cw.ResponseWriter.WriteHeader(code)
		return
	}

	// Check content length
	if contentLength := cw.Header().Get("Content-Length"); contentLength != "" {
		if length, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			if length < cw.config.MinSize {
				cw.ResponseWriter.WriteHeader(code)
				return
			}
		}
	}

	// Set compression headers
	cw.Header().Set("Content-Encoding", cw.algorithm)
	cw.Header().Del("Content-Length") // We don't know the compressed length

	cw.ResponseWriter.WriteHeader(code)
}

// Write compresses and writes data.
func (cw *compressionWriter) Write(data []byte) (int, error) {
	if !cw.headerWritten {
		cw.WriteHeader(http.StatusOK)
	}

	// If no compression is being used, write directly
	if cw.compressor == nil {
		// Check if we should start compressing based on accumulated size
		cw.size += int64(len(data))
		if cw.size >= cw.config.MinSize && cw.shouldStartCompression() {
			// Too late to start compression here, write directly
		}
		return cw.ResponseWriter.Write(data)
	}

	// Write to compressor
	return cw.compressor.Write(data)
}

// shouldCompress checks if the content type should be compressed.
func (cw *compressionWriter) shouldCompress(contentType string) bool {
	if contentType == "" {
		return false
	}

	// Remove charset and other parameters
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(ct)

	for _, compressibleType := range cw.config.Types {
		if ct == compressibleType {
			return true
		}
	}

	return false
}

// shouldStartCompression checks if compression should start.
func (cw *compressionWriter) shouldStartCompression() bool {
	contentType := cw.Header().Get("Content-Type")
	return cw.shouldCompress(contentType)
}

// Hijack implements http.Hijacker interface.
func (cw *compressionWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := cw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Flush implements http.Flusher interface.
func (cw *compressionWriter) Flush() {
	if cw.compressor != nil {
		if flusher, ok := cw.compressor.(flusher); ok {
			flusher.Flush()
		}
	}
	if flusher, ok := cw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// flusher interface for flushable compressors.
type flusher interface {
	Flush() error
}

// GzipCompressor implements gzip compression.
type GzipCompressor struct {
	level int
}

// NewGzipCompressor creates a new gzip compressor.
func NewGzipCompressor() *GzipCompressor {
	return &GzipCompressor{
		level: gzip.DefaultCompression,
	}
}

// Compress compresses data using gzip.
func (gc *GzipCompressor) Compress(algorithm string, data []byte) ([]byte, error) {
	if algorithm != "gzip" {
		return nil, http.ErrNotSupported
	}

	// This is a simplified implementation
	// In practice, you'd use a streaming approach
	return data, nil
}

// SupportedAlgorithms returns supported compression algorithms.
func (gc *GzipCompressor) SupportedAlgorithms() []string {
	return []string{"gzip", "deflate"}
}

// DeflateCompressor implements deflate compression.
type DeflateCompressor struct {
	level int
}

// NewDeflateCompressor creates a new deflate compressor.
func NewDeflateCompressor() *DeflateCompressor {
	return &DeflateCompressor{
		level: flate.DefaultCompression,
	}
}

// Compress compresses data using deflate.
func (dc *DeflateCompressor) Compress(algorithm string, data []byte) ([]byte, error) {
	if algorithm != "deflate" {
		return nil, http.ErrNotSupported
	}

	// This is a simplified implementation
	// In practice, you'd use a streaming approach
	return data, nil
}

// SupportedAlgorithms returns supported compression algorithms.
func (dc *DeflateCompressor) SupportedAlgorithms() []string {
	return []string{"deflate"}
}
