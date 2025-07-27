// Package compression provides automatic request/response compression capabilities.
package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Compressor handles request and response compression.
type Compressor struct {
	enabledTypes     []interfaces.CompressionType
	compressionLevel int
	threshold        int // Minimum size to compress
}

// NewCompressor creates a new compressor with default settings.
func NewCompressor() *Compressor {
	return &Compressor{
		enabledTypes: []interfaces.CompressionType{
			interfaces.CompressionGzip,
			interfaces.CompressionDeflate,
		},
		compressionLevel: gzip.DefaultCompression,
		threshold:        1024, // 1KB threshold
	}
}

// SetEnabledTypes sets the enabled compression types.
func (c *Compressor) SetEnabledTypes(types ...interfaces.CompressionType) *Compressor {
	c.enabledTypes = types
	return c
}

// SetCompressionLevel sets the compression level (1-9 for gzip/deflate).
func (c *Compressor) SetCompressionLevel(level int) *Compressor {
	if level >= 1 && level <= 9 {
		c.compressionLevel = level
	}
	return c
}

// SetThreshold sets the minimum size threshold for compression.
func (c *Compressor) SetThreshold(threshold int) *Compressor {
	c.threshold = threshold
	return c
}

// CompressRequest compresses request body if applicable.
func (c *Compressor) CompressRequest(req *interfaces.Request) error {
	if req.Body == nil {
		return nil
	}

	// Convert body to bytes
	bodyBytes, err := c.bodyToBytes(req.Body)
	if err != nil {
		return fmt.Errorf("failed to convert body to bytes: %w", err)
	}

	// Check if compression is worthwhile
	if len(bodyBytes) < c.threshold {
		return nil
	}

	// Choose compression type
	compressionType := c.chooseCompressionType(req)
	if compressionType == "" {
		return nil
	}

	// Compress the body
	compressedBody, err := c.compressData(bodyBytes, compressionType)
	if err != nil {
		return fmt.Errorf("failed to compress body: %w", err)
	}

	// Only use compressed version if it's actually smaller
	if len(compressedBody) < len(bodyBytes) {
		req.Body = compressedBody
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Encoding"] = string(compressionType)
		req.Headers["Content-Length"] = fmt.Sprintf("%d", len(compressedBody))
	}

	return nil
}

// DecompressResponse decompresses response body if compressed.
func (c *Compressor) DecompressResponse(resp *interfaces.Response) error {
	if resp.Body == nil || len(resp.Body) == 0 {
		return nil
	}

	// Check Content-Encoding header
	encoding := c.getContentEncoding(resp)
	if encoding == "" {
		return nil
	}

	// Decompress based on encoding
	decompressed, err := c.decompressData(resp.Body, interfaces.CompressionType(encoding))
	if err != nil {
		return fmt.Errorf("failed to decompress response: %w", err)
	}

	resp.Body = decompressed
	resp.IsCompressed = false

	// Update headers
	if resp.Headers != nil {
		delete(resp.Headers, "Content-Encoding")
		resp.Headers["Content-Length"] = fmt.Sprintf("%d", len(decompressed))
	}

	return nil
}

// chooseCompressionType selects the best compression type for the request.
func (c *Compressor) chooseCompressionType(req *interfaces.Request) interfaces.CompressionType {
	if len(c.enabledTypes) == 0 {
		return ""
	}

	// Check Accept-Encoding header for server preferences
	if req.Headers != nil {
		if acceptEncoding, exists := req.Headers["Accept-Encoding"]; exists {
			for _, compressionType := range c.enabledTypes {
				if strings.Contains(acceptEncoding, string(compressionType)) {
					return compressionType
				}
			}
		}
	}

	// Default to first enabled type
	return c.enabledTypes[0]
}

// getContentEncoding extracts content encoding from response.
func (c *Compressor) getContentEncoding(resp *interfaces.Response) string {
	// Check response field first
	if resp.Headers != nil {
		if encoding, exists := resp.Headers["Content-Encoding"]; exists {
			return encoding
		}
	}

	return ""
}

// bodyToBytes converts various body types to byte slice.
func (c *Compressor) bodyToBytes(body interface{}) ([]byte, error) {
	switch v := body.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case *bytes.Buffer:
		return v.Bytes(), nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		return nil, fmt.Errorf("unsupported body type: %T", body)
	}
}

// compressData compresses data using the specified compression type.
func (c *Compressor) compressData(data []byte, compressionType interfaces.CompressionType) ([]byte, error) {
	var buf bytes.Buffer

	switch compressionType {
	case interfaces.CompressionGzip:
		writer, err := gzip.NewWriterLevel(&buf, c.compressionLevel)
		if err != nil {
			return nil, err
		}
		defer writer.Close()

		_, err = writer.Write(data)
		if err != nil {
			return nil, err
		}

		err = writer.Close()
		if err != nil {
			return nil, err
		}

	case interfaces.CompressionDeflate:
		writer, err := flate.NewWriter(&buf, c.compressionLevel)
		if err != nil {
			return nil, err
		}
		defer writer.Close()

		_, err = writer.Write(data)
		if err != nil {
			return nil, err
		}

		err = writer.Close()
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}

	return buf.Bytes(), nil
}

// decompressData decompresses data using the specified compression type.
func (c *Compressor) decompressData(data []byte, compressionType interfaces.CompressionType) ([]byte, error) {
	reader := bytes.NewReader(data)

	switch compressionType {
	case interfaces.CompressionGzip:
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()

		return io.ReadAll(gzipReader)

	case interfaces.CompressionDeflate:
		flateReader := flate.NewReader(reader)
		defer flateReader.Close()

		return io.ReadAll(flateReader)

	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
}

// GetCompressionRatio returns the compression ratio for given data.
func (c *Compressor) GetCompressionRatio(original, compressed []byte) float64 {
	if len(original) == 0 {
		return 0.0
	}
	return float64(len(compressed)) / float64(len(original))
}

// GetSavings returns the space savings percentage.
func (c *Compressor) GetSavings(original, compressed []byte) float64 {
	ratio := c.GetCompressionRatio(original, compressed)
	return (1.0 - ratio) * 100.0
}

// IsCompressed checks if data appears to be compressed.
func IsCompressed(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// Check for gzip magic number
	if data[0] == 0x1f && data[1] == 0x8b {
		return true
	}

	// Check for deflate/zlib magic number
	if data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x9c || data[1] == 0xda) {
		return true
	}

	return false
}

// CompressionMiddleware provides middleware for automatic compression.
type CompressionMiddleware struct {
	compressor *Compressor
}

// NewCompressionMiddleware creates a new compression middleware.
func NewCompressionMiddleware(compressor *Compressor) *CompressionMiddleware {
	if compressor == nil {
		compressor = NewCompressor()
	}
	return &CompressionMiddleware{
		compressor: compressor,
	}
}

// Process implements the Middleware interface.
func (m *CompressionMiddleware) Process(ctx interface{}, req *interfaces.Request, next func(interface{}, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	// Compress request if applicable
	if err := m.compressor.CompressRequest(req); err != nil {
		return nil, fmt.Errorf("request compression failed: %w", err)
	}

	// Add Accept-Encoding header for response compression
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	if _, exists := req.Headers["Accept-Encoding"]; !exists {
		encodings := make([]string, len(m.compressor.enabledTypes))
		for i, t := range m.compressor.enabledTypes {
			encodings[i] = string(t)
		}
		req.Headers["Accept-Encoding"] = strings.Join(encodings, ", ")
	}

	// Execute request
	resp, err := next(ctx, req)
	if err != nil {
		return resp, err
	}

	// Decompress response if compressed
	if resp != nil {
		if decompErr := m.compressor.DecompressResponse(resp); decompErr != nil {
			return resp, fmt.Errorf("response decompression failed: %w", decompErr)
		}
	}

	return resp, nil
}

// AutoCompressionConfig provides configuration for automatic compression.
type AutoCompressionConfig struct {
	RequestCompression  bool
	ResponseCompression bool
	MinSize             int
	CompressionLevel    int
	EnabledTypes        []interfaces.CompressionType
}

// DefaultAutoCompressionConfig returns default configuration.
func DefaultAutoCompressionConfig() *AutoCompressionConfig {
	return &AutoCompressionConfig{
		RequestCompression:  true,
		ResponseCompression: true,
		MinSize:             1024,
		CompressionLevel:    gzip.DefaultCompression,
		EnabledTypes: []interfaces.CompressionType{
			interfaces.CompressionGzip,
			interfaces.CompressionDeflate,
		},
	}
}

// Statistics tracks compression statistics.
type Statistics struct {
	RequestsCompressed    int64
	ResponsesDecompressed int64
	BytesSavedRequest     int64
	BytesSavedResponse    int64
	CompressionRatio      float64
}

// AddRequestCompression adds request compression statistics.
func (s *Statistics) AddRequestCompression(originalSize, compressedSize int) {
	s.RequestsCompressed++
	s.BytesSavedRequest += int64(originalSize - compressedSize)
	s.updateCompressionRatio(originalSize, compressedSize)
}

// AddResponseDecompression adds response decompression statistics.
func (s *Statistics) AddResponseDecompression(compressedSize, decompressedSize int) {
	s.ResponsesDecompressed++
	s.BytesSavedResponse += int64(decompressedSize - compressedSize)
	s.updateCompressionRatio(decompressedSize, compressedSize)
}

// updateCompressionRatio updates the average compression ratio.
func (s *Statistics) updateCompressionRatio(originalSize, compressedSize int) {
	if originalSize == 0 {
		return
	}

	newRatio := float64(compressedSize) / float64(originalSize)
	totalOps := s.RequestsCompressed + s.ResponsesDecompressed

	if totalOps == 1 {
		s.CompressionRatio = newRatio
	} else {
		// Moving average
		s.CompressionRatio = (s.CompressionRatio*float64(totalOps-1) + newRatio) / float64(totalOps)
	}
}

// GetSavingsPercentage returns the overall savings percentage.
func (s *Statistics) GetSavingsPercentage() float64 {
	return (1.0 - s.CompressionRatio) * 100.0
}
