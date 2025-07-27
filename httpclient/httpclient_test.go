package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	t.Parallel()

	factory := NewFactory()

	assert.NotNil(t, factory)

	providers := factory.GetAvailableProviders()
	assert.Contains(t, providers, interfaces.ProviderNetHTTP)
	assert.Contains(t, providers, interfaces.ProviderFiber)
	assert.Contains(t, providers, interfaces.ProviderFastHTTP)
	assert.Len(t, providers, 3)
}

func TestFactory_CreateClient(t *testing.T) {
	t.Parallel()

	factory := NewFactory()
	cfg := config.DefaultConfig()
	cfg.BaseURL = "https://api.example.com"

	t.Run("create nethttp client", func(t *testing.T) {
		client, err := factory.CreateClient(interfaces.ProviderNetHTTP, cfg)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "nethttp", client.GetProvider().Name())
	})

	t.Run("create fiber client", func(t *testing.T) {
		client, err := factory.CreateClient(interfaces.ProviderFiber, cfg)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "fiber", client.GetProvider().Name())
	})

	t.Run("create fasthttp client", func(t *testing.T) {
		client, err := factory.CreateClient(interfaces.ProviderFastHTTP, cfg)

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "fasthttp", client.GetProvider().Name())
	})

	t.Run("unsupported provider", func(t *testing.T) {
		client, err := factory.CreateClient("unsupported", cfg)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})

	t.Run("nil config uses default", func(t *testing.T) {
		client, err := factory.CreateClient(interfaces.ProviderNetHTTP, nil)

		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestFactory_RegisterProvider(t *testing.T) {
	t.Parallel()

	factory := NewFactory()

	t.Run("register custom provider", func(t *testing.T) {
		customType := interfaces.ProviderType("custom")
		constructor := func(config *interfaces.Config) (interfaces.Provider, error) {
			return nil, errors.New("mock provider")
		}

		err := factory.RegisterProvider(customType, constructor)

		require.NoError(t, err)

		providers := factory.GetAvailableProviders()
		assert.Contains(t, providers, customType)
	})

	t.Run("nil constructor", func(t *testing.T) {
		err := factory.RegisterProvider("test", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "constructor cannot be nil")
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		client, err := New(interfaces.ProviderNetHTTP, "https://api.example.com")

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "nethttp", client.GetProvider().Name())
	})

	t.Run("unsupported provider", func(t *testing.T) {
		client, err := New("unsupported", "https://api.example.com")

		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestNewWithConfig(t *testing.T) {
	t.Parallel()

	cfg := config.NewBuilder().
		WithBaseURL("https://api.example.com").
		WithTimeout(5 * time.Second).
		Build()

	client, err := NewWithConfig(interfaces.ProviderNetHTTP, cfg)

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClient_HTTPMethods(t *testing.T) {
	t.Parallel()

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{"method": "%s", "path": "%s"}`, r.Method, r.URL.Path)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := New(interfaces.ProviderNetHTTP, server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("GET request", func(t *testing.T) {
		resp, err := client.Get(ctx, "/test")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "GET")
		assert.Contains(t, string(resp.Body), "/test")
	})

	t.Run("POST request", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		resp, err := client.Post(ctx, "/test", body)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "POST")
	})

	t.Run("PUT request", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		resp, err := client.Put(ctx, "/test", body)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "PUT")
	})

	t.Run("DELETE request", func(t *testing.T) {
		resp, err := client.Delete(ctx, "/test")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "DELETE")
	})

	t.Run("PATCH request", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		resp, err := client.Patch(ctx, "/test", body)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "PATCH")
	})

	t.Run("HEAD request", func(t *testing.T) {
		resp, err := client.Head(ctx, "/test")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		resp, err := client.Options(ctx, "/test")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "OPTIONS")
	})
}

func TestClient_SetHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Auth", r.Header.Get("Authorization"))
		w.Header().Set("X-Received-Custom", r.Header.Get("X-Custom"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(interfaces.ProviderNetHTTP, server.URL)
	require.NoError(t, err)

	headers := map[string]string{
		"Authorization": "Bearer token",
		"X-Custom":      "custom-value",
	}
	client.SetHeaders(headers)

	resp, err := client.Get(context.Background(), "/test")

	require.NoError(t, err)
	assert.Equal(t, "Bearer token", resp.Headers["X-Received-Auth"])
	assert.Equal(t, "custom-value", resp.Headers["X-Received-Custom"])
}

func TestClient_SetTimeout(t *testing.T) {
	t.Parallel()

	client, err := New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	require.NoError(t, err)

	// Set a short timeout
	client.SetTimeout(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Get(ctx, "/delay/1")
	assert.Error(t, err) // Should timeout
}

func TestClient_SetErrorHandler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	client, err := New(interfaces.ProviderNetHTTP, server.URL)
	require.NoError(t, err)

	customError := errors.New("custom error handling")
	client.SetErrorHandler(func(resp *interfaces.Response) error {
		if resp.StatusCode >= 500 {
			return customError
		}
		return nil
	})

	_, err = client.Get(context.Background(), "/error")

	assert.Error(t, err)
	assert.Equal(t, customError, err)
}

func TestClient_RetryLogic(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	cfg := config.NewBuilder().
		WithBaseURL(server.URL).
		WithMaxRetries(3).
		WithRetryInterval(10 * time.Millisecond).
		Build()

	client, err := NewWithConfig(interfaces.ProviderNetHTTP, cfg)
	require.NoError(t, err)

	resp, err := client.Get(context.Background(), "/test")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "success", string(resp.Body))
	assert.Equal(t, 3, attempts) // Should have retried 2 times + initial attempt
}

func TestClient_CalculateRetryDelay(t *testing.T) {
	t.Parallel()

	cfg := config.NewBuilder().
		WithRetryInterval(100 * time.Millisecond).
		Build()
	cfg.RetryConfig.Multiplier = 2.0
	cfg.RetryConfig.MaxInterval = 1 * time.Second

	client := &Client{
		config:      cfg,
		retryConfig: cfg.RetryConfig,
	}

	tests := []struct {
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{1, 0, 200 * time.Millisecond},
		{2, 100 * time.Millisecond, 300 * time.Millisecond},
		{3, 200 * time.Millisecond, 500 * time.Millisecond},
		{10, 1 * time.Second, 1 * time.Second}, // Should be capped at max
	}

	for _, tt := range tests {
		delay := client.calculateRetryDelay(tt.attempt)
		assert.GreaterOrEqual(t, delay, tt.expectedMin, "attempt %d", tt.attempt)
		assert.LessOrEqual(t, delay, tt.expectedMax, "attempt %d", tt.attempt)
	}
}

func BenchmarkClient_Get(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client, _ := New(interfaces.ProviderNetHTTP, server.URL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Get(context.Background(), "/test")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
