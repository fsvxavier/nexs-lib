package nethttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		provider, err := NewProvider(cfg)

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "nethttp", provider.Name())
		assert.Equal(t, "1.0.0", provider.Version())
		assert.True(t, provider.IsHealthy())
	})

	t.Run("nil config", func(t *testing.T) {
		provider, err := NewProvider(nil)

		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestProvider_Configure(t *testing.T) {
	t.Parallel()

	t.Run("successful configuration", func(t *testing.T) {
		provider := &Provider{}
		cfg := config.DefaultConfig()
		cfg.Timeout = 5 * time.Second
		cfg.MaxIdleConns = 50

		err := provider.Configure(cfg)

		require.NoError(t, err)
		assert.Equal(t, cfg, provider.config)
		assert.NotNil(t, provider.client)
		assert.Equal(t, 5*time.Second, provider.client.Timeout)
	})

	t.Run("nil config", func(t *testing.T) {
		provider := &Provider{}
		err := provider.Configure(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestProvider_DoRequest(t *testing.T) {
	t.Parallel()

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		case "/echo":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Echo-Method", r.Method)
			w.WriteHeader(http.StatusOK)
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			w.Write(body)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.BaseURL = server.URL
	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	t.Run("successful GET request", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/success",
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.False(t, resp.IsError)
		assert.Contains(t, string(resp.Body), "success")
		assert.Greater(t, resp.Latency, time.Duration(0))
		assert.Equal(t, "application/json", resp.Headers["Content-Type"])
	})

	t.Run("successful POST request with body", func(t *testing.T) {
		reqBody := map[string]string{"test": "data"}
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/echo",
			Body:   reqBody,
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.False(t, resp.IsError)
		assert.Contains(t, string(resp.Body), "test")
		assert.Contains(t, string(resp.Body), "data")
		assert.Equal(t, "POST", resp.Headers["X-Echo-Method"])
	})

	t.Run("request with string body", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/echo",
			Body:   "string body",
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "string body", string(resp.Body))
	})

	t.Run("request with byte body", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/echo",
			Body:   []byte("byte body"),
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "byte body", string(resp.Body))
	})

	t.Run("error response", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/error",
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.True(t, resp.IsError)
		assert.Contains(t, string(resp.Body), "internal server error")
	})

	t.Run("request with headers", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/success",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"X-Custom":      "value",
			},
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("nil request", func(t *testing.T) {
		resp, err := provider.DoRequest(context.Background(), nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("absolute URL override", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    server.URL + "/success",
		}

		resp, err := provider.DoRequest(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestProvider_DoRequest_WithoutBaseURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	// No base URL set
	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	req := &interfaces.Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp, err := provider.DoRequest(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProvider_Metrics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte("response"))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MetricsEnabled = true
	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	// Initial metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)

	// Successful request
	req := &interfaces.Request{Method: "GET", URL: "/success"}
	_, err = provider.DoRequest(context.Background(), req)
	require.NoError(t, err)

	// Failed request
	req = &interfaces.Request{Method: "GET", URL: "/error"}
	_, err = provider.DoRequest(context.Background(), req)
	require.NoError(t, err)

	// Check updated metrics
	metrics = provider.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.SuccessfulRequests)
	assert.Equal(t, int64(1), metrics.FailedRequests)
	assert.Greater(t, metrics.AverageLatency, time.Duration(0))
	assert.False(t, metrics.LastRequestTime.IsZero())
}

func TestProvider_MetricsDisabled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MetricsEnabled = false
	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	req := &interfaces.Request{Method: "GET", URL: "/success"}
	_, err = provider.DoRequest(context.Background(), req)
	require.NoError(t, err)

	// Metrics should remain at zero when disabled
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
}

func TestProvider_MarshalBody(t *testing.T) {
	t.Parallel()

	provider := &Provider{}

	tests := []struct {
		name     string
		body     interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "string body",
			body:     "test string",
			expected: "test string",
			wantErr:  false,
		},
		{
			name:     "byte slice body",
			body:     []byte("test bytes"),
			expected: "test bytes",
			wantErr:  false,
		},
		{
			name:     "struct body",
			body:     map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "reader body",
			body:     strings.NewReader("reader content"),
			expected: "reader content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.marshalBody(tt.body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

func TestProvider_SetHeaders(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Headers = map[string]string{
		"X-Default": "default-value",
		"Override":  "will-be-overridden",
	}

	provider := &Provider{config: cfg}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	reqHeaders := map[string]string{
		"Override": "overridden-value",
		"X-Custom": "custom-value",
	}

	provider.setHeaders(req, reqHeaders)

	assert.Equal(t, "default-value", req.Header.Get("X-Default"))
	assert.Equal(t, "overridden-value", req.Header.Get("Override"))
	assert.Equal(t, "custom-value", req.Header.Get("X-Custom"))
}

func TestProvider_SetHeaders_ContentType(t *testing.T) {
	t.Parallel()

	provider := &Provider{config: config.DefaultConfig()}

	t.Run("sets default content-type for request with body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader("body"))
		provider.setHeaders(req, nil)

		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	})

	t.Run("does not override existing content-type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader("body"))
		headers := map[string]string{"Content-Type": "text/plain"}
		provider.setHeaders(req, headers)

		assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))
	})

	t.Run("does not set content-type for request without body", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		provider.setHeaders(req, nil)

		assert.Empty(t, req.Header.Get("Content-Type"))
	})
}

func BenchmarkProvider_DoRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.MetricsEnabled = false // Disable metrics for cleaner benchmark
	provider, _ := NewProvider(cfg)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "/test",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := provider.DoRequest(context.Background(), req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestProvider_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	cfg := config.DefaultConfig()
	cfg.BaseURL = server.URL
	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	const numGoroutines = 10
	const requestsPerGoroutine = 5

	errChan := make(chan error, numGoroutines*requestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < requestsPerGoroutine; j++ {
				req := &interfaces.Request{
					Method: "GET",
					URL:    "/test",
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_, err := provider.DoRequest(ctx, req)
				cancel()

				errChan <- err
			}
		}()
	}

	// Collect all results
	for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Verify metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(numGoroutines*requestsPerGoroutine), metrics.TotalRequests)
	assert.Equal(t, int64(numGoroutines*requestsPerGoroutine), metrics.SuccessfulRequests)
	assert.Equal(t, int64(0), metrics.FailedRequests)
}
