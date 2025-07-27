package fiber

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		cfg := &interfaces.Config{
			BaseURL:        "https://api.example.com",
			Timeout:        30 * time.Second,
			MetricsEnabled: true,
		}
		provider, err := NewProvider(cfg)

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "fiber", provider.Name())
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
		cfg := &interfaces.Config{
			BaseURL: "https://api.example.com",
		}
		provider, err := NewProvider(cfg)
		require.NoError(t, err)

		newCfg := &interfaces.Config{
			BaseURL: "https://api2.example.com",
		}

		err = provider.Configure(newCfg)
		require.NoError(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		cfg := &interfaces.Config{
			BaseURL: "https://api.example.com",
		}
		provider, err := NewProvider(cfg)
		require.NoError(t, err)

		err = provider.Configure(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestProvider_DoRequest(t *testing.T) {
	t.Parallel()

	cfg := &interfaces.Config{
		BaseURL:        "https://httpbin.org",
		Timeout:        30 * time.Second,
		MetricsEnabled: true,
		TracingEnabled: false,
	}

	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	t.Run("successful GET request", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/get",
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
		assert.NotEmpty(t, resp.Body)
	})

	t.Run("successful POST request with body", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/post",
			Body:   map[string]string{"key": "value"},
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
		assert.NotEmpty(t, resp.Body)
	})

	t.Run("request with string body", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/post",
			Body:   "test string body",
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
	})

	t.Run("request with byte body", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "POST",
			URL:    "/post",
			Body:   []byte("test byte body"),
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
	})

	t.Run("error response", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/status/500",
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 500, resp.StatusCode)
		assert.True(t, resp.IsError)
	})

	t.Run("request with headers", func(t *testing.T) {
		req := &interfaces.Request{
			Method: "GET",
			URL:    "/headers",
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
			},
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
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
			URL:    "https://httpbin.org/get",
		}

		resp, err := provider.DoRequest(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.False(t, resp.IsError)
	})
}

func TestProvider_DoRequest_WithoutBaseURL(t *testing.T) {
	t.Parallel()

	cfg := &interfaces.Config{
		Timeout:        30 * time.Second,
		MetricsEnabled: false,
		TracingEnabled: false,
	}

	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "https://httpbin.org/get",
	}

	resp, err := provider.DoRequest(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.False(t, resp.IsError)
}

func TestProvider_Metrics(t *testing.T) {
	t.Parallel()

	cfg := &interfaces.Config{
		BaseURL:        "https://httpbin.org",
		Timeout:        30 * time.Second,
		MetricsEnabled: true,
	}

	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "/get",
	}

	// Get initial metrics
	initialMetrics := provider.GetMetrics()

	// Make request
	_, err = provider.DoRequest(context.Background(), req)
	require.NoError(t, err)

	// Check metrics were updated
	updatedMetrics := provider.GetMetrics()
	assert.Greater(t, updatedMetrics.TotalRequests, initialMetrics.TotalRequests)
	assert.Greater(t, updatedMetrics.SuccessfulRequests, initialMetrics.SuccessfulRequests)
}

func TestProvider_MetricsDisabled(t *testing.T) {
	t.Parallel()

	cfg := &interfaces.Config{
		BaseURL:        "https://httpbin.org",
		Timeout:        30 * time.Second,
		MetricsEnabled: false,
	}

	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	req := &interfaces.Request{
		Method: "GET",
		URL:    "/get",
	}

	// Make request
	_, err = provider.DoRequest(context.Background(), req)
	require.NoError(t, err)

	// Metrics should remain at zero
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.SuccessfulRequests)
}

func TestProvider_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	cfg := &interfaces.Config{
		BaseURL:        "https://httpbin.org",
		Timeout:        30 * time.Second,
		MetricsEnabled: true,
	}

	provider, err := NewProvider(cfg)
	require.NoError(t, err)

	// Number of concurrent requests
	numRequests := 10
	results := make(chan error, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req := &interfaces.Request{
				Method: "GET",
				URL:    "/get",
			}

			_, err := provider.DoRequest(context.Background(), req)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Check metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(numRequests), metrics.TotalRequests)
	assert.Equal(t, int64(numRequests), metrics.SuccessfulRequests)
}
