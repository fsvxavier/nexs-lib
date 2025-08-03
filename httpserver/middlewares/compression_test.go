package middlewares

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestNewCompressionMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		want     *CompressionMiddleware
	}{
		{
			name:     "Create with priority 5",
			priority: 5,
		},
		{
			name:     "Create with priority 0",
			priority: 0,
		},
		{
			name:     "Create with negative priority",
			priority: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCompressionMiddleware(tt.priority)
			if got == nil {
				t.Error("NewCompressionMiddleware() returned nil")
				return
			}

			if got.Name() != "compression" {
				t.Errorf("NewCompressionMiddleware().Name() = %v, want %v", got.Name(), "compression")
			}

			if got.Priority() != tt.priority {
				t.Errorf("NewCompressionMiddleware().Priority() = %v, want %v", got.Priority(), tt.priority)
			}

			// Verify default config is applied
			defaultConfig := DefaultCompressionConfig()
			if !reflect.DeepEqual(got.config, defaultConfig) {
				t.Error("NewCompressionMiddleware() did not apply default config")
			}
		})
	}
}

func TestNewCompressionMiddlewareWithConfig(t *testing.T) {
	customConfig := CompressionConfig{
		EnableGzip:        true,
		EnableDeflate:     false,
		EnableBrotli:      true,
		GzipLevel:         9,
		DeflateLevel:      1,
		BrotliLevel:       8,
		CompressibleTypes: []string{"text/html", "application/json"},
		SkipTypes:         []string{"image/png"},
		MinSize:           2048,
		MaxSize:           512000,
		SkipPaths:         []string{"/api/health"},
		SkipMethods:       []string{"HEAD"},
		PoolSize:          5,
		BufferSize:        16384,
		ChunkSize:         4096,
		VaryHeader:        false,
		NoTransform:       true,
		DefaultQuality:    0.8,
	}

	tests := []struct {
		name     string
		priority int
		config   CompressionConfig
	}{
		{
			name:     "Create with custom config",
			priority: 3,
			config:   customConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCompressionMiddlewareWithConfig(tt.priority, tt.config)
			if got == nil {
				t.Error("NewCompressionMiddlewareWithConfig() returned nil")
				return
			}

			if !reflect.DeepEqual(got.config, tt.config) {
				t.Error("NewCompressionMiddlewareWithConfig() did not apply custom config correctly")
			}
		})
	}
}

func TestDefaultCompressionConfig(t *testing.T) {
	config := DefaultCompressionConfig()

	// Test basic compression settings
	if !config.EnableGzip {
		t.Error("DefaultCompressionConfig().EnableGzip should be true")
	}
	if !config.EnableDeflate {
		t.Error("DefaultCompressionConfig().EnableDeflate should be true")
	}
	if config.EnableBrotli {
		t.Error("DefaultCompressionConfig().EnableBrotli should be false")
	}

	// Test compression levels
	if config.GzipLevel != 6 {
		t.Errorf("DefaultCompressionConfig().GzipLevel = %v, want %v", config.GzipLevel, 6)
	}
	if config.DeflateLevel != 6 {
		t.Errorf("DefaultCompressionConfig().DeflateLevel = %v, want %v", config.DeflateLevel, 6)
	}
	if config.BrotliLevel != 6 {
		t.Errorf("DefaultCompressionConfig().BrotliLevel = %v, want %v", config.BrotliLevel, 6)
	}

	// Test compressible types
	expectedTypes := []string{
		"text/html", "text/css", "text/javascript", "text/plain", "text/xml",
		"application/json", "application/javascript", "application/xml",
		"application/x-javascript", "application/rss+xml", "application/atom+xml",
		"image/svg+xml",
	}
	if !reflect.DeepEqual(config.CompressibleTypes, expectedTypes) {
		t.Error("DefaultCompressionConfig().CompressibleTypes does not match expected types")
	}

	// Test skip types
	expectedSkipTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"video/", "audio/", "application/pdf", "application/zip", "application/gzip",
	}
	if !reflect.DeepEqual(config.SkipTypes, expectedSkipTypes) {
		t.Error("DefaultCompressionConfig().SkipTypes does not match expected types")
	}

	// Test size constraints
	if config.MinSize != 1024 {
		t.Errorf("DefaultCompressionConfig().MinSize = %v, want %v", config.MinSize, 1024)
	}
	if config.MaxSize != 1048576 {
		t.Errorf("DefaultCompressionConfig().MaxSize = %v, want %v", config.MaxSize, 1048576)
	}

	// Test default quality
	if config.DefaultQuality != 1.0 {
		t.Errorf("DefaultCompressionConfig().DefaultQuality = %v, want %v", config.DefaultQuality, 1.0)
	}
}

func TestCompressionMiddleware_GetConfig(t *testing.T) {
	customConfig := CompressionConfig{
		EnableGzip:     true,
		EnableDeflate:  false,
		GzipLevel:      9,
		MinSize:        2048,
		DefaultQuality: 0.9,
	}

	cm := NewCompressionMiddlewareWithConfig(1, customConfig)
	got := cm.GetConfig()

	if !reflect.DeepEqual(got, customConfig) {
		t.Error("GetConfig() did not return the correct configuration")
	}
}

func TestCompressionMiddleware_SetConfig(t *testing.T) {
	cm := NewCompressionMiddleware(1)
	newConfig := CompressionConfig{
		EnableGzip:     false,
		EnableDeflate:  true,
		GzipLevel:      3,
		MinSize:        512,
		DefaultQuality: 0.7,
	}

	cm.SetConfig(newConfig)
	got := cm.GetConfig()

	if !reflect.DeepEqual(got, newConfig) {
		t.Error("SetConfig() did not update the configuration correctly")
	}
}

func TestCompressionMiddleware_extractRequestInfo(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	tests := []struct {
		name string
		req  interface{}
		want *CompressionRequestInfo
	}{
		{
			name: "Valid HTTP request",
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"Accept-Encoding": "gzip, deflate, br",
					"Content-Type":    "application/json",
					"Content-Length":  "1024",
				},
			},
			want: &CompressionRequestInfo{
				Method:         "GET",
				Path:           "/api/data",
				AcceptEncoding: "gzip, deflate, br",
				ContentType:    "application/json",
				ContentLength:  1024,
			},
		},
		{
			name: "Request without headers",
			req: map[string]interface{}{
				"method": "POST",
				"path":   "/api/upload",
			},
			want: &CompressionRequestInfo{
				Method:         "POST",
				Path:           "/api/upload",
				AcceptEncoding: "",
				ContentType:    "",
				ContentLength:  0,
			},
		},
		{
			name: "Invalid request type",
			req:  "invalid",
			want: &CompressionRequestInfo{},
		},
		{
			name: "Request with invalid content length",
			req: map[string]interface{}{
				"method": "PUT",
				"path":   "/api/update",
				"headers": map[string]string{
					"Accept-Encoding": "gzip",
					"Content-Length":  "invalid",
				},
			},
			want: &CompressionRequestInfo{
				Method:         "PUT",
				Path:           "/api/update",
				AcceptEncoding: "gzip",
				ContentType:    "",
				ContentLength:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.extractRequestInfo(tt.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractRequestInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionMiddleware_shouldSkipCompression(t *testing.T) {
	config := DefaultCompressionConfig()
	config.SkipPaths = []string{"/health", "/metrics"}
	config.SkipMethods = []string{"HEAD", "OPTIONS"}

	cm := NewCompressionMiddlewareWithConfig(1, config)

	tests := []struct {
		name     string
		reqInfo  *CompressionRequestInfo
		wantSkip bool
		reason   string
	}{
		{
			name: "Skip path in skip list",
			reqInfo: &CompressionRequestInfo{
				Method: "GET",
				Path:   "/health",
			},
			wantSkip: true,
			reason:   "Path in skip list",
		},
		{
			name: "Skip method in skip list",
			reqInfo: &CompressionRequestInfo{
				Method: "HEAD",
				Path:   "/api/data",
			},
			wantSkip: true,
			reason:   "Method in skip list",
		},
		{
			name: "Skip no Accept-Encoding header",
			reqInfo: &CompressionRequestInfo{
				Method:         "GET",
				Path:           "/api/data",
				AcceptEncoding: "",
			},
			wantSkip: true,
			reason:   "No Accept-Encoding header",
		},
		{
			name: "Do not skip valid request",
			reqInfo: &CompressionRequestInfo{
				Method:         "GET",
				Path:           "/api/data",
				AcceptEncoding: "gzip",
			},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compCtx := &CompressionContext{}
			got := cm.shouldSkipCompression(tt.reqInfo, compCtx)

			if got != tt.wantSkip {
				t.Errorf("shouldSkipCompression() = %v, want %v", got, tt.wantSkip)
			}

			if tt.wantSkip && compCtx.SkipReason != tt.reason {
				t.Errorf("shouldSkipCompression() skip reason = %v, want %v", compCtx.SkipReason, tt.reason)
			}
		})
	}
}

func TestCompressionMiddleware_parseAcceptEncoding(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	tests := []struct {
		name           string
		acceptEncoding string
		want           []EncodingInfo
	}{
		{
			name:           "Empty Accept-Encoding",
			acceptEncoding: "",
			want:           []EncodingInfo{},
		},
		{
			name:           "Single encoding without quality",
			acceptEncoding: "gzip",
			want:           []EncodingInfo{{Name: "gzip", Quality: 1.0}},
		},
		{
			name:           "Single encoding with quality",
			acceptEncoding: "gzip;q=0.8",
			want:           []EncodingInfo{{Name: "gzip", Quality: 0.8}},
		},
		{
			name:           "Multiple encodings",
			acceptEncoding: "gzip, deflate;q=0.9, br;q=0.7",
			want: []EncodingInfo{
				{Name: "gzip", Quality: 1.0},
				{Name: "deflate", Quality: 0.9},
				{Name: "br", Quality: 0.7},
			},
		},
		{
			name:           "Complex encoding with invalid quality",
			acceptEncoding: "gzip;q=invalid, deflate",
			want: []EncodingInfo{
				{Name: "gzip", Quality: 1.0},
				{Name: "deflate", Quality: 1.0},
			},
		},
		{
			name:           "Encoding with spaces",
			acceptEncoding: " gzip ; q=0.5 , deflate ",
			want: []EncodingInfo{
				{Name: "gzip", Quality: 0.5},
				{Name: "deflate", Quality: 1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.parseAcceptEncoding(tt.acceptEncoding)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAcceptEncoding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionMiddleware_selectBestEncoding(t *testing.T) {
	tests := []struct {
		name      string
		config    CompressionConfig
		encodings []EncodingInfo
		want      string
	}{
		{
			name: "Select gzip with highest quality",
			config: CompressionConfig{
				EnableGzip:    true,
				EnableDeflate: true,
			},
			encodings: []EncodingInfo{
				{Name: "gzip", Quality: 1.0},
				{Name: "deflate", Quality: 0.8},
			},
			want: "gzip",
		},
		{
			name: "Select deflate when gzip disabled",
			config: CompressionConfig{
				EnableGzip:    false,
				EnableDeflate: true,
			},
			encodings: []EncodingInfo{
				{Name: "gzip", Quality: 1.0},
				{Name: "deflate", Quality: 0.8},
			},
			want: "deflate",
		},
		{
			name: "Select brotli with highest quality",
			config: CompressionConfig{
				EnableGzip:    true,
				EnableDeflate: true,
				EnableBrotli:  true,
			},
			encodings: []EncodingInfo{
				{Name: "gzip", Quality: 0.8},
				{Name: "deflate", Quality: 0.7},
				{Name: "br", Quality: 0.9},
			},
			want: "br",
		},
		{
			name: "No supported encodings",
			config: CompressionConfig{
				EnableGzip:    false,
				EnableDeflate: false,
				EnableBrotli:  false,
			},
			encodings: []EncodingInfo{
				{Name: "gzip", Quality: 1.0},
			},
			want: "",
		},
		{
			name: "Skip zero quality encoding",
			config: CompressionConfig{
				EnableGzip:    true,
				EnableDeflate: true,
			},
			encodings: []EncodingInfo{
				{Name: "gzip", Quality: 0.0},
				{Name: "deflate", Quality: 0.8},
			},
			want: "deflate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCompressionMiddlewareWithConfig(1, tt.config)
			got := cm.selectBestEncoding(tt.encodings)
			if got != tt.want {
				t.Errorf("selectBestEncoding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionMiddleware_shouldCompressContent(t *testing.T) {
	config := DefaultCompressionConfig()
	cm := NewCompressionMiddlewareWithConfig(1, config)

	tests := []struct {
		name     string
		httpResp map[string]interface{}
		want     bool
	}{
		{
			name: "Compress JSON content",
			httpResp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			want: true,
		},
		{
			name: "Skip image content",
			httpResp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "image/png",
				},
			},
			want: false,
		},
		{
			name: "No headers",
			httpResp: map[string]interface{}{
				"body": "test",
			},
			want: false,
		},
		{
			name: "No content type",
			httpResp: map[string]interface{}{
				"headers": map[string]string{},
			},
			want: false,
		},
		{
			name: "Invalid headers type",
			httpResp: map[string]interface{}{
				"headers": "invalid",
			},
			want: false,
		},
		{
			name: "Compress HTML content",
			httpResp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "text/html; charset=utf-8",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.shouldCompressContent(tt.httpResp)
			if got != tt.want {
				t.Errorf("shouldCompressContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionMiddleware_performCompression(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	tests := []struct {
		name     string
		data     string
		encoding string
		want     string
		wantErr  bool
	}{
		{
			name:     "Compress with gzip",
			data:     "test data",
			encoding: "gzip",
			want:     "GZIP[test data]",
			wantErr:  false,
		},
		{
			name:     "Compress with deflate",
			data:     "test data",
			encoding: "deflate",
			want:     "DEFLATE[test data]",
			wantErr:  false,
		},
		{
			name:     "Compress with brotli",
			data:     "test data",
			encoding: "br",
			want:     "BROTLI[test data]",
			wantErr:  false,
		},
		{
			name:     "Unsupported encoding",
			data:     "test data",
			encoding: "unknown",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cm.performCompression(tt.data, tt.encoding)
			if (err != nil) != tt.wantErr {
				t.Errorf("performCompression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("performCompression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompressionMiddleware_SetCompressionLevel(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	tests := []struct {
		name      string
		algorithm string
		level     int
		wantErr   bool
	}{
		{
			name:      "Set valid gzip level",
			algorithm: "gzip",
			level:     9,
			wantErr:   false,
		},
		{
			name:      "Set valid deflate level",
			algorithm: "deflate",
			level:     1,
			wantErr:   false,
		},
		{
			name:      "Set valid brotli level",
			algorithm: "brotli",
			level:     5,
			wantErr:   false,
		},
		{
			name:      "Set valid br level",
			algorithm: "br",
			level:     7,
			wantErr:   false,
		},
		{
			name:      "Invalid level too low",
			algorithm: "gzip",
			level:     0,
			wantErr:   true,
		},
		{
			name:      "Invalid level too high",
			algorithm: "gzip",
			level:     10,
			wantErr:   true,
		},
		{
			name:      "Unsupported algorithm",
			algorithm: "unknown",
			level:     5,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.SetCompressionLevel(tt.algorithm, tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCompressionLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				config := cm.GetConfig()
				switch strings.ToLower(tt.algorithm) {
				case "gzip":
					if config.GzipLevel != tt.level {
						t.Errorf("GzipLevel = %v, want %v", config.GzipLevel, tt.level)
					}
				case "deflate":
					if config.DeflateLevel != tt.level {
						t.Errorf("DeflateLevel = %v, want %v", config.DeflateLevel, tt.level)
					}
				case "brotli", "br":
					if config.BrotliLevel != tt.level {
						t.Errorf("BrotliLevel = %v, want %v", config.BrotliLevel, tt.level)
					}
				}
			}
		})
	}
}

func TestCompressionMiddleware_AddCompressibleType(t *testing.T) {
	cm := NewCompressionMiddleware(1)
	originalTypes := len(cm.GetConfig().CompressibleTypes)

	cm.AddCompressibleType("text/custom")

	config := cm.GetConfig()
	if len(config.CompressibleTypes) != originalTypes+1 {
		t.Errorf("AddCompressibleType() did not add type, got %d types, want %d", len(config.CompressibleTypes), originalTypes+1)
	}

	found := false
	for _, cType := range config.CompressibleTypes {
		if cType == "text/custom" {
			found = true
			break
		}
	}

	if !found {
		t.Error("AddCompressibleType() did not add the specified type")
	}
}

func TestCompressionMiddleware_RemoveCompressibleType(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	// Add a type first
	cm.AddCompressibleType("text/custom")
	originalTypes := len(cm.GetConfig().CompressibleTypes)

	// Remove it
	cm.RemoveCompressibleType("text/custom")

	config := cm.GetConfig()
	if len(config.CompressibleTypes) != originalTypes-1 {
		t.Errorf("RemoveCompressibleType() did not remove type, got %d types, want %d", len(config.CompressibleTypes), originalTypes-1)
	}

	// Verify it's not in the list
	for _, cType := range config.CompressibleTypes {
		if cType == "text/custom" {
			t.Error("RemoveCompressibleType() did not remove the specified type")
		}
	}

	// Test removing non-existent type
	originalTypes = len(cm.GetConfig().CompressibleTypes)
	cm.RemoveCompressibleType("non/existent")

	if len(cm.GetConfig().CompressibleTypes) != originalTypes {
		t.Error("RemoveCompressibleType() should not change list when removing non-existent type")
	}
}

func TestCompressionMiddleware_Reset(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	// Simulate some metrics
	cm.totalRequests = 100
	cm.compressedResponses = 80
	cm.bytesOriginal = 10000
	cm.bytesCompressed = 5000
	cm.compressionTime = 1000000

	cm.Reset()

	metrics := cm.GetMetrics()
	if metrics["total_requests"].(int64) != 0 {
		t.Error("Reset() did not reset total_requests")
	}
	if metrics["compressed_responses"].(int64) != 0 {
		t.Error("Reset() did not reset compressed_responses")
	}
	if metrics["bytes_original"].(int64) != 0 {
		t.Error("Reset() did not reset bytes_original")
	}
	if metrics["bytes_compressed"].(int64) != 0 {
		t.Error("Reset() did not reset bytes_compressed")
	}
}

func TestCompressionMiddleware_GetMetrics(t *testing.T) {
	cm := NewCompressionMiddleware(1)

	// Initial metrics should be zero
	metrics := cm.GetMetrics()

	expectedKeys := []string{
		"total_requests", "compressed_responses", "compression_rate",
		"bytes_original", "bytes_compressed", "bytes_saved",
		"avg_compression_ratio", "avg_processing_time", "uptime",
	}

	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("GetMetrics() missing key: %s", key)
		}
	}

	// Test with some data
	cm.totalRequests = 100
	cm.compressedResponses = 80
	cm.bytesOriginal = 10000
	cm.bytesCompressed = 5000
	cm.compressionTime = 80000000 // 80ms total = 1ms average for 80 requests

	metrics = cm.GetMetrics()

	if metrics["total_requests"].(int64) != 100 {
		t.Errorf("GetMetrics() total_requests = %v, want %v", metrics["total_requests"], 100)
	}

	compressionRate := metrics["compression_rate"].(float64)
	if compressionRate != 80.0 {
		t.Errorf("GetMetrics() compression_rate = %v, want %v", compressionRate, 80.0)
	}

	bytesSaved := metrics["bytes_saved"].(int64)
	if bytesSaved != 5000 {
		t.Errorf("GetMetrics() bytes_saved = %v, want %v", bytesSaved, 5000)
	}

	avgCompressionRatio := metrics["avg_compression_ratio"].(float64)
	if avgCompressionRatio != 50.0 {
		t.Errorf("GetMetrics() avg_compression_ratio = %v, want %v", avgCompressionRatio, 50.0)
	}

	avgProcessingTime := metrics["avg_processing_time"].(float64)
	if avgProcessingTime != 1.0 { // 1ms
		t.Errorf("GetMetrics() avg_processing_time = %v, want %v", avgProcessingTime, 1.0)
	}
}

func TestCompressionMiddleware_Process(t *testing.T) {
	tests := []struct {
		name           string
		config         CompressionConfig
		req            interface{}
		mockResponse   interface{}
		expectError    bool
		expectSkipped  bool
		expectDisabled bool
	}{
		{
			name:   "Process with disabled middleware",
			config: DefaultCompressionConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"Accept-Encoding": "gzip",
				},
			},
			mockResponse:   map[string]interface{}{"body": "test data"},
			expectDisabled: true,
		},
		{
			name:   "Process skip for health check path",
			config: DefaultCompressionConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/health",
				"headers": map[string]string{
					"Accept-Encoding": "gzip",
				},
			},
			mockResponse:  map[string]interface{}{"body": "OK"},
			expectSkipped: true,
		},
		{
			name:   "Process skip for HEAD method",
			config: DefaultCompressionConfig(),
			req: map[string]interface{}{
				"method": "HEAD",
				"path":   "/api/data",
				"headers": map[string]string{
					"Accept-Encoding": "gzip",
				},
			},
			mockResponse:  map[string]interface{}{"body": "test"},
			expectSkipped: true,
		},
		{
			name:   "Process skip for no Accept-Encoding",
			config: DefaultCompressionConfig(),
			req: map[string]interface{}{
				"method":  "GET",
				"path":    "/api/data",
				"headers": map[string]string{},
			},
			mockResponse:  map[string]interface{}{"body": "test"},
			expectSkipped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCompressionMiddlewareWithConfig(1, tt.config)

			if tt.expectDisabled {
				cm.SetEnabled(false)
			} else {
				cm.SetEnabled(true)
			}

			ctx := context.Background()
			called := false
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return tt.mockResponse, nil
			}

			resp, err := cm.Process(ctx, tt.req, next)

			if (err != nil) != tt.expectError {
				t.Errorf("Process() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !called {
				t.Error("Process() did not call next middleware")
			}

			if resp == nil {
				t.Error("Process() returned nil response")
			}
		})
	}
}

func TestCompressionMiddleware_compressResponse(t *testing.T) {
	config := DefaultCompressionConfig()
	cm := NewCompressionMiddlewareWithConfig(1, config)

	tests := []struct {
		name           string
		resp           interface{}
		compCtx        *CompressionContext
		expectError    bool
		expectSkipped  bool
		expectModified bool
	}{
		{
			name: "Invalid response format",
			resp: "invalid",
			compCtx: &CompressionContext{
				SelectedEncoding: "gzip",
			},
			expectError: true,
		},
		{
			name: "Compress JSON response",
			resp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"body": strings.Repeat("test data ", 200), // Make it large enough
			},
			compCtx: &CompressionContext{
				SelectedEncoding: "gzip",
			},
			expectModified: true,
		},
		{
			name: "Skip image response",
			resp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "image/png",
				},
				"body": "binary image data",
			},
			compCtx: &CompressionContext{
				SelectedEncoding: "gzip",
			},
			expectSkipped: true,
		},
		{
			name: "Skip small response",
			resp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"body": "small",
			},
			compCtx: &CompressionContext{
				SelectedEncoding: "gzip",
			},
			expectSkipped: true,
		},
		{
			name: "Response without body",
			resp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			compCtx: &CompressionContext{
				SelectedEncoding: "gzip",
			},
			expectSkipped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := cm.compressResponse(tt.resp, tt.compCtx)

			if (err != nil) != tt.expectError {
				t.Errorf("compressResponse() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError {
				return
			}

			if tt.expectSkipped && !tt.compCtx.Skipped {
				t.Error("compressResponse() should have been skipped")
			}

			if tt.expectModified && resp != nil {
				if httpResp, ok := resp.(map[string]interface{}); ok {
					if headers, exists := httpResp["headers"]; exists {
						if h, ok := headers.(map[string]string); ok {
							if h["Content-Encoding"] != tt.compCtx.SelectedEncoding {
								t.Errorf("compressResponse() Content-Encoding = %v, want %v", h["Content-Encoding"], tt.compCtx.SelectedEncoding)
							}
						}
					}
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkCompressionMiddleware_parseAcceptEncoding(b *testing.B) {
	cm := NewCompressionMiddleware(1)
	acceptEncoding := "gzip;q=0.9, deflate;q=0.8, br;q=0.7, *;q=0.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.parseAcceptEncoding(acceptEncoding)
	}
}

func BenchmarkCompressionMiddleware_selectBestEncoding(b *testing.B) {
	cm := NewCompressionMiddleware(1)
	encodings := []EncodingInfo{
		{Name: "gzip", Quality: 0.9},
		{Name: "deflate", Quality: 0.8},
		{Name: "br", Quality: 0.7},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.selectBestEncoding(encodings)
	}
}

func BenchmarkCompressionMiddleware_performCompression(b *testing.B) {
	cm := NewCompressionMiddleware(1)
	data := strings.Repeat("test data ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.performCompression(data, "gzip")
	}
}
