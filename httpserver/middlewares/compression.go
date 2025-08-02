package middlewares

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// CompressionMiddleware provides response compression functionality.
type CompressionMiddleware struct {
	*BaseMiddleware

	// Configuration
	config CompressionConfig

	// Metrics
	totalRequests       int64
	compressedResponses int64
	compressionRatio    int64 // Stored as percentage * 100 for atomic operations
	bytesOriginal       int64
	bytesCompressed     int64
	compressionTime     int64 // Total compression time in nanoseconds

	// Internal state
	startTime time.Time
}

// CompressionConfig defines configuration options for the compression middleware.
type CompressionConfig struct {
	// Compression algorithms
	EnableGzip    bool
	EnableDeflate bool
	EnableBrotli  bool

	// Compression levels (1-9, where 9 is highest compression)
	GzipLevel    int
	DeflateLevel int
	BrotliLevel  int

	// Content type filtering
	CompressibleTypes []string
	SkipTypes         []string

	// Size filtering
	MinSize int // Minimum response size to compress (bytes)
	MaxSize int // Maximum response size to compress (bytes)

	// Request filtering
	SkipPaths   []string
	SkipMethods []string

	// Performance settings
	PoolSize   int
	BufferSize int
	ChunkSize  int

	// Behavior settings
	VaryHeader  bool
	NoTransform bool

	// Quality values for Accept-Encoding
	DefaultQuality float64
}

// CompressionContext represents compression context for a request.
type CompressionContext struct {
	SupportedEncodings []EncodingInfo
	SelectedEncoding   string
	OriginalSize       int64
	CompressedSize     int64
	CompressionRatio   float64
	ProcessingTime     time.Duration
	Skipped            bool
	SkipReason         string
}

// EncodingInfo represents information about a supported encoding.
type EncodingInfo struct {
	Name    string
	Quality float64
}

// NewCompressionMiddleware creates a new compression middleware with default configuration.
func NewCompressionMiddleware(priority int) *CompressionMiddleware {
	return &CompressionMiddleware{
		BaseMiddleware: NewBaseMiddleware("compression", priority),
		config:         DefaultCompressionConfig(),
		startTime:      time.Now(),
	}
}

// NewCompressionMiddlewareWithConfig creates a new compression middleware with custom configuration.
func NewCompressionMiddlewareWithConfig(priority int, config CompressionConfig) *CompressionMiddleware {
	return &CompressionMiddleware{
		BaseMiddleware: NewBaseMiddleware("compression", priority),
		config:         config,
		startTime:      time.Now(),
	}
}

// DefaultCompressionConfig returns a default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		EnableGzip:    true,
		EnableDeflate: true,
		EnableBrotli:  false, // Brotli support is optional
		GzipLevel:     6,     // Default compression level
		DeflateLevel:  6,
		BrotliLevel:   6,
		CompressibleTypes: []string{
			"text/html",
			"text/css",
			"text/javascript",
			"text/plain",
			"text/xml",
			"application/json",
			"application/javascript",
			"application/xml",
			"application/x-javascript",
			"application/rss+xml",
			"application/atom+xml",
			"image/svg+xml",
		},
		SkipTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"video/",
			"audio/",
			"application/pdf",
			"application/zip",
			"application/gzip",
		},
		MinSize:        1024,    // 1KB minimum
		MaxSize:        1048576, // 1MB maximum
		SkipPaths:      []string{"/health", "/metrics"},
		SkipMethods:    []string{"HEAD", "OPTIONS"},
		PoolSize:       10,
		BufferSize:     32768, // 32KB
		ChunkSize:      8192,  // 8KB
		VaryHeader:     true,
		NoTransform:    false,
		DefaultQuality: 1.0,
	}
}

// Process implements the Middleware interface for response compression.
func (cm *CompressionMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !cm.IsEnabled() {
		return next(ctx, req)
	}

	atomic.AddInt64(&cm.totalRequests, 1)
	startTime := time.Now()

	// Extract request information
	reqInfo := cm.extractRequestInfo(req)

	// Check if compression should be skipped
	compCtx := &CompressionContext{}
	if cm.shouldSkipCompression(reqInfo, compCtx) {
		cm.GetLogger().Debug("Compression skipped for %s %s: %s", reqInfo.Method, reqInfo.Path, compCtx.SkipReason)
		return next(ctx, req)
	}

	// Parse Accept-Encoding header
	compCtx.SupportedEncodings = cm.parseAcceptEncoding(reqInfo.AcceptEncoding)
	compCtx.SelectedEncoding = cm.selectBestEncoding(compCtx.SupportedEncodings)

	if compCtx.SelectedEncoding == "" {
		compCtx.Skipped = true
		compCtx.SkipReason = "No supported encoding found"
		cm.GetLogger().Debug("No supported compression encoding found for request")
		return next(ctx, req)
	}

	// Add compression context to request context
	ctxWithCompression := context.WithValue(ctx, "compression", compCtx)

	// Process request
	resp, err := next(ctxWithCompression, req)
	if err != nil {
		return resp, err
	}

	// Apply compression to response
	compressedResp, compressionErr := cm.compressResponse(resp, compCtx)
	if compressionErr != nil {
		cm.GetLogger().Error("Compression failed: %v", compressionErr)
		return resp, nil // Return original response on compression error
	}

	// Update metrics
	processingTime := time.Since(startTime)
	compCtx.ProcessingTime = processingTime
	atomic.AddInt64(&cm.compressionTime, processingTime.Nanoseconds())

	if compCtx.CompressedSize > 0 {
		atomic.AddInt64(&cm.compressedResponses, 1)
		atomic.AddInt64(&cm.bytesOriginal, compCtx.OriginalSize)
		atomic.AddInt64(&cm.bytesCompressed, compCtx.CompressedSize)

		// Update compression ratio (stored as percentage * 100)
		if compCtx.OriginalSize > 0 {
			ratio := (1.0 - float64(compCtx.CompressedSize)/float64(compCtx.OriginalSize)) * 10000
			atomic.StoreInt64(&cm.compressionRatio, int64(ratio))
		}

		cm.GetLogger().Debug("Response compressed: %s, %d -> %d bytes (%.1f%% reduction)",
			compCtx.SelectedEncoding, compCtx.OriginalSize, compCtx.CompressedSize, compCtx.CompressionRatio)
	}

	return compressedResp, nil
}

// GetConfig returns the current compression configuration.
func (cm *CompressionMiddleware) GetConfig() CompressionConfig {
	return cm.config
}

// SetConfig updates the compression configuration.
func (cm *CompressionMiddleware) SetConfig(config CompressionConfig) {
	cm.config = config
	cm.GetLogger().Info("Compression middleware configuration updated")
}

// GetMetrics returns compression metrics.
func (cm *CompressionMiddleware) GetMetrics() map[string]interface{} {
	totalRequests := atomic.LoadInt64(&cm.totalRequests)
	compressedResponses := atomic.LoadInt64(&cm.compressedResponses)
	bytesOriginal := atomic.LoadInt64(&cm.bytesOriginal)
	bytesCompressed := atomic.LoadInt64(&cm.bytesCompressed)

	var compressionRate float64
	if totalRequests > 0 {
		compressionRate = float64(compressedResponses) / float64(totalRequests) * 100.0
	}

	var avgCompressionRatio float64
	if bytesOriginal > 0 {
		avgCompressionRatio = (1.0 - float64(bytesCompressed)/float64(bytesOriginal)) * 100.0
	}

	var avgProcessingTime float64
	if compressedResponses > 0 {
		avgProcessingTime = float64(atomic.LoadInt64(&cm.compressionTime)) / float64(compressedResponses) / 1e6 // Convert to milliseconds
	}

	return map[string]interface{}{
		"total_requests":        totalRequests,
		"compressed_responses":  compressedResponses,
		"compression_rate":      compressionRate,
		"bytes_original":        bytesOriginal,
		"bytes_compressed":      bytesCompressed,
		"bytes_saved":           bytesOriginal - bytesCompressed,
		"avg_compression_ratio": avgCompressionRatio,
		"avg_processing_time":   avgProcessingTime,
		"uptime":                time.Since(cm.startTime),
	}
}

// CompressionRequestInfo holds request information for compression analysis.
type CompressionRequestInfo struct {
	Method         string
	Path           string
	AcceptEncoding string
	ContentType    string
	ContentLength  int64
}

// extractRequestInfo extracts relevant information from the request.
func (cm *CompressionMiddleware) extractRequestInfo(req interface{}) *CompressionRequestInfo {
	info := &CompressionRequestInfo{}

	if httpReq, ok := req.(map[string]interface{}); ok {
		if method, exists := httpReq["method"]; exists {
			if m, ok := method.(string); ok {
				info.Method = strings.ToUpper(m)
			}
		}

		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				info.Path = p
			}
		}

		if headers, exists := httpReq["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				info.AcceptEncoding = h["Accept-Encoding"]
				info.ContentType = h["Content-Type"]

				if contentLength, exists := h["Content-Length"]; exists {
					if length, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
						info.ContentLength = length
					}
				}
			}
		}
	}

	return info
}

// shouldSkipCompression determines if compression should be skipped.
func (cm *CompressionMiddleware) shouldSkipCompression(reqInfo *CompressionRequestInfo, compCtx *CompressionContext) bool {
	// Check skip paths
	for _, skipPath := range cm.config.SkipPaths {
		if reqInfo.Path == skipPath {
			compCtx.Skipped = true
			compCtx.SkipReason = "Path in skip list"
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range cm.config.SkipMethods {
		if reqInfo.Method == skipMethod {
			compCtx.Skipped = true
			compCtx.SkipReason = "Method in skip list"
			return true
		}
	}

	// Check if client supports compression
	if reqInfo.AcceptEncoding == "" {
		compCtx.Skipped = true
		compCtx.SkipReason = "No Accept-Encoding header"
		return true
	}

	return false
}

// parseAcceptEncoding parses the Accept-Encoding header.
func (cm *CompressionMiddleware) parseAcceptEncoding(acceptEncoding string) []EncodingInfo {
	if acceptEncoding == "" {
		return []EncodingInfo{}
	}

	var encodings []EncodingInfo
	parts := strings.Split(acceptEncoding, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse encoding and quality value
		encoding := EncodingInfo{Quality: cm.config.DefaultQuality}

		if strings.Contains(part, ";") {
			subParts := strings.Split(part, ";")
			encoding.Name = strings.TrimSpace(subParts[0])

			// Parse quality value
			for _, subPart := range subParts[1:] {
				subPart = strings.TrimSpace(subPart)
				if strings.HasPrefix(subPart, "q=") {
					if quality, err := strconv.ParseFloat(subPart[2:], 64); err == nil {
						encoding.Quality = quality
					}
				}
			}
		} else {
			encoding.Name = part
		}

		encodings = append(encodings, encoding)
	}

	return encodings
}

// selectBestEncoding selects the best compression encoding based on configuration and client support.
func (cm *CompressionMiddleware) selectBestEncoding(encodings []EncodingInfo) string {
	var bestEncoding string
	var bestQuality float64

	for _, encoding := range encodings {
		if encoding.Quality <= 0 {
			continue // Client explicitly rejects this encoding
		}

		var supported bool
		switch strings.ToLower(encoding.Name) {
		case "gzip":
			supported = cm.config.EnableGzip
		case "deflate":
			supported = cm.config.EnableDeflate
		case "br":
			supported = cm.config.EnableBrotli
		}

		if supported && encoding.Quality > bestQuality {
			bestEncoding = encoding.Name
			bestQuality = encoding.Quality
		}
	}

	return bestEncoding
}

// compressResponse compresses the response based on the selected encoding.
func (cm *CompressionMiddleware) compressResponse(resp interface{}, compCtx *CompressionContext) (interface{}, error) {
	httpResp, ok := resp.(map[string]interface{})
	if !ok {
		return resp, fmt.Errorf("invalid response format")
	}

	// Check if response should be compressed based on content type
	if !cm.shouldCompressContent(httpResp) {
		compCtx.Skipped = true
		compCtx.SkipReason = "Content type not compressible"
		return resp, nil
	}

	// Get response body
	body, exists := httpResp["body"]
	if !exists {
		return resp, nil
	}

	bodyStr := fmt.Sprintf("%v", body)
	compCtx.OriginalSize = int64(len(bodyStr))

	// Check size constraints
	if compCtx.OriginalSize < int64(cm.config.MinSize) {
		compCtx.Skipped = true
		compCtx.SkipReason = "Response too small"
		return resp, nil
	}

	if cm.config.MaxSize > 0 && compCtx.OriginalSize > int64(cm.config.MaxSize) {
		compCtx.Skipped = true
		compCtx.SkipReason = "Response too large"
		return resp, nil
	}

	// Compress the response body
	compressedBody, err := cm.performCompression(bodyStr, compCtx.SelectedEncoding)
	if err != nil {
		return resp, fmt.Errorf("compression failed: %w", err)
	}

	compCtx.CompressedSize = int64(len(compressedBody))
	if compCtx.OriginalSize > 0 {
		compCtx.CompressionRatio = (1.0 - float64(compCtx.CompressedSize)/float64(compCtx.OriginalSize)) * 100.0
	}

	// Update response with compressed content
	httpResp["body"] = compressedBody

	// Update headers
	headers, exists := httpResp["headers"]
	if !exists {
		headers = make(map[string]string)
		httpResp["headers"] = headers
	}

	if h, ok := headers.(map[string]string); ok {
		h["Content-Encoding"] = compCtx.SelectedEncoding
		h["Content-Length"] = fmt.Sprintf("%d", compCtx.CompressedSize)

		if cm.config.VaryHeader {
			if vary, exists := h["Vary"]; exists {
				if !strings.Contains(vary, "Accept-Encoding") {
					h["Vary"] = vary + ", Accept-Encoding"
				}
			} else {
				h["Vary"] = "Accept-Encoding"
			}
		}

		if cm.config.NoTransform {
			h["Cache-Control"] = "no-transform"
		}
	}

	return httpResp, nil
}

// shouldCompressContent determines if the response content should be compressed.
func (cm *CompressionMiddleware) shouldCompressContent(httpResp map[string]interface{}) bool {
	headers, exists := httpResp["headers"]
	if !exists {
		return false
	}

	h, ok := headers.(map[string]string)
	if !ok {
		return false
	}

	contentType := h["Content-Type"]
	if contentType == "" {
		return false
	}

	// Check if content type is in skip list
	for _, skipType := range cm.config.SkipTypes {
		if strings.HasPrefix(contentType, skipType) {
			return false
		}
	}

	// Check if content type is compressible
	for _, compressibleType := range cm.config.CompressibleTypes {
		if strings.HasPrefix(contentType, compressibleType) {
			return true
		}
	}

	return false
}

// performCompression performs the actual compression based on the selected encoding.
func (cm *CompressionMiddleware) performCompression(data, encoding string) (string, error) {
	switch strings.ToLower(encoding) {
	case "gzip":
		return cm.compressGzip(data)
	case "deflate":
		return cm.compressDeflate(data)
	case "br":
		return cm.compressBrotli(data)
	default:
		return "", fmt.Errorf("unsupported compression encoding: %s", encoding)
	}
}

// compressGzip compresses data using gzip.
func (cm *CompressionMiddleware) compressGzip(data string) (string, error) {
	// Simulated gzip compression (in real implementation, use compress/gzip)
	compressed := fmt.Sprintf("GZIP[%s]", data)
	return compressed, nil
}

// compressDeflate compresses data using deflate.
func (cm *CompressionMiddleware) compressDeflate(data string) (string, error) {
	// Simulated deflate compression (in real implementation, use compress/flate)
	compressed := fmt.Sprintf("DEFLATE[%s]", data)
	return compressed, nil
}

// compressBrotli compresses data using Brotli.
func (cm *CompressionMiddleware) compressBrotli(data string) (string, error) {
	// Simulated Brotli compression (in real implementation, use external Brotli library)
	compressed := fmt.Sprintf("BROTLI[%s]", data)
	return compressed, nil
}

// Reset resets all metrics.
func (cm *CompressionMiddleware) Reset() {
	atomic.StoreInt64(&cm.totalRequests, 0)
	atomic.StoreInt64(&cm.compressedResponses, 0)
	atomic.StoreInt64(&cm.compressionRatio, 0)
	atomic.StoreInt64(&cm.bytesOriginal, 0)
	atomic.StoreInt64(&cm.bytesCompressed, 0)
	atomic.StoreInt64(&cm.compressionTime, 0)
	cm.startTime = time.Now()
	cm.GetLogger().Info("Compression middleware metrics reset")
}

// SetCompressionLevel sets the compression level for a specific algorithm.
func (cm *CompressionMiddleware) SetCompressionLevel(algorithm string, level int) error {
	if level < 1 || level > 9 {
		return fmt.Errorf("compression level must be between 1 and 9")
	}

	switch strings.ToLower(algorithm) {
	case "gzip":
		cm.config.GzipLevel = level
	case "deflate":
		cm.config.DeflateLevel = level
	case "brotli", "br":
		cm.config.BrotliLevel = level
	default:
		return fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}

	cm.GetLogger().Info("Compression level for %s set to %d", algorithm, level)
	return nil
}

// AddCompressibleType adds a content type to the compressible types list.
func (cm *CompressionMiddleware) AddCompressibleType(contentType string) {
	cm.config.CompressibleTypes = append(cm.config.CompressibleTypes, contentType)
	cm.GetLogger().Info("Added compressible content type: %s", contentType)
}

// RemoveCompressibleType removes a content type from the compressible types list.
func (cm *CompressionMiddleware) RemoveCompressibleType(contentType string) {
	for i, compType := range cm.config.CompressibleTypes {
		if compType == contentType {
			cm.config.CompressibleTypes = append(cm.config.CompressibleTypes[:i], cm.config.CompressibleTypes[i+1:]...)
			cm.GetLogger().Info("Removed compressible content type: %s", contentType)
			return
		}
	}
}
