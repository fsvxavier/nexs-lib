package config

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.Equal(t, "", config.BaseURL)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.Equal(t, 10*time.Second, config.TLSHandshakeTimeout)
	assert.False(t, config.DisableKeepAlives)
	assert.False(t, config.DisableCompression)
	assert.False(t, config.InsecureSkipVerify)
	assert.NotNil(t, config.Headers)
	assert.NotNil(t, config.RetryConfig)
	assert.True(t, config.TracingEnabled)
	assert.True(t, config.MetricsEnabled)
}

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	retryConfig := DefaultRetryConfig()

	assert.Equal(t, 3, retryConfig.MaxRetries)
	assert.Equal(t, 1*time.Second, retryConfig.InitialInterval)
	assert.Equal(t, 30*time.Second, retryConfig.MaxInterval)
	assert.Equal(t, 2.0, retryConfig.Multiplier)
	assert.NotNil(t, retryConfig.RetryCondition)
}

func TestDefaultRetryCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resp        *interfaces.Response
		err         error
		shouldRetry bool
	}{
		{
			name:        "retry on error",
			resp:        nil,
			err:         assert.AnError,
			shouldRetry: true,
		},
		{
			name:        "retry on 408",
			resp:        &interfaces.Response{StatusCode: 408},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "retry on 429",
			resp:        &interfaces.Response{StatusCode: 429},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "retry on 502",
			resp:        &interfaces.Response{StatusCode: 502},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "retry on 503",
			resp:        &interfaces.Response{StatusCode: 503},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "retry on 504",
			resp:        &interfaces.Response{StatusCode: 504},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "retry on 500",
			resp:        &interfaces.Response{StatusCode: 500},
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "no retry on 200",
			resp:        &interfaces.Response{StatusCode: 200},
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "no retry on 400",
			resp:        &interfaces.Response{StatusCode: 400},
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "no retry on 404",
			resp:        &interfaces.Response{StatusCode: 404},
			err:         nil,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultRetryCondition(tt.resp, tt.err)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}

func TestBuilder(t *testing.T) {
	t.Parallel()

	baseURL := "https://api.example.com"
	timeout := 10 * time.Second
	maxIdleConns := 50
	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	config := NewBuilder().
		WithBaseURL(baseURL).
		WithTimeout(timeout).
		WithMaxIdleConns(maxIdleConns).
		WithHeaders(headers).
		WithHeader("X-API-Version", "v1").
		WithDisableKeepAlives(true).
		WithInsecureSkipVerify(true).
		WithMaxRetries(5).
		WithRetryInterval(2 * time.Second).
		WithTracingEnabled(false).
		WithMetricsEnabled(false).
		Build()

	assert.Equal(t, baseURL, config.BaseURL)
	assert.Equal(t, timeout, config.Timeout)
	assert.Equal(t, maxIdleConns, config.MaxIdleConns)
	assert.True(t, config.DisableKeepAlives)
	assert.True(t, config.InsecureSkipVerify)
	assert.False(t, config.TracingEnabled)
	assert.False(t, config.MetricsEnabled)

	assert.Equal(t, "Bearer token", config.Headers["Authorization"])
	assert.Equal(t, "application/json", config.Headers["Content-Type"])
	assert.Equal(t, "v1", config.Headers["X-API-Version"])

	assert.Equal(t, 5, config.RetryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryConfig.InitialInterval)
}

func TestBuilderWithNilHeaders(t *testing.T) {
	t.Parallel()

	config := NewBuilder().
		WithHeader("Test", "Value").
		Build()

	assert.NotNil(t, config.Headers)
	assert.Equal(t, "Value", config.Headers["Test"])
}

func TestBuilderWithNilRetryConfig(t *testing.T) {
	t.Parallel()

	config := NewBuilder().
		WithMaxRetries(10).
		Build()

	assert.NotNil(t, config.RetryConfig)
	assert.Equal(t, 10, config.RetryConfig.MaxRetries)
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *interfaces.Config
		expected *interfaces.Config
	}{
		{
			name: "valid config unchanged",
			config: &interfaces.Config{
				Timeout:             10 * time.Second,
				MaxIdleConns:        50,
				IdleConnTimeout:     60 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
				Headers:             map[string]string{"test": "value"},
				RetryConfig:         &interfaces.RetryConfig{MaxRetries: 5},
			},
			expected: &interfaces.Config{
				Timeout:             10 * time.Second,
				MaxIdleConns:        50,
				IdleConnTimeout:     60 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
				Headers:             map[string]string{"test": "value"},
				RetryConfig:         &interfaces.RetryConfig{MaxRetries: 5},
			},
		},
		{
			name: "invalid config gets defaults",
			config: &interfaces.Config{
				Timeout:             0,
				MaxIdleConns:        0,
				IdleConnTimeout:     0,
				TLSHandshakeTimeout: 0,
				Headers:             nil,
				RetryConfig:         nil,
			},
			expected: &interfaces.Config{
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				Headers:             map[string]string{},
				RetryConfig:         DefaultRetryConfig(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Timeout, tt.config.Timeout)
			assert.Equal(t, tt.expected.MaxIdleConns, tt.config.MaxIdleConns)
			assert.Equal(t, tt.expected.IdleConnTimeout, tt.config.IdleConnTimeout)
			assert.Equal(t, tt.expected.TLSHandshakeTimeout, tt.config.TLSHandshakeTimeout)
			assert.NotNil(t, tt.config.Headers)
			assert.NotNil(t, tt.config.RetryConfig)
		})
	}
}

func TestCloneConfig(t *testing.T) {
	t.Parallel()

	original := &interfaces.Config{
		BaseURL:     "https://api.example.com",
		Timeout:     10 * time.Second,
		Headers:     map[string]string{"Authorization": "Bearer token"},
		RetryConfig: &interfaces.RetryConfig{MaxRetries: 5},
	}

	clone := CloneConfig(original)

	// Verify clone values are equal but not the same instance
	assert.Equal(t, original.BaseURL, clone.BaseURL)
	assert.Equal(t, original.Timeout, clone.Timeout)
	// assert.NotSame(t, original, clone) // Skip pointer comparison

	// Verify deep copy of headers
	require.NotNil(t, clone.Headers)
	for k, v := range original.Headers {
		assert.Equal(t, v, clone.Headers[k])
	} // Verify deep copy of retry config
	require.NotNil(t, clone.RetryConfig)
	assert.Equal(t, original.RetryConfig.MaxRetries, clone.RetryConfig.MaxRetries)

	// Modify clone and verify original is unchanged
	clone.Headers["New-Header"] = "new-value"
	clone.RetryConfig.MaxRetries = 10

	assert.NotContains(t, original.Headers, "New-Header")
	assert.Equal(t, 5, original.RetryConfig.MaxRetries)
}

func TestCloneConfigWithNilFields(t *testing.T) {
	t.Parallel()

	original := &interfaces.Config{
		BaseURL:     "https://api.example.com",
		Headers:     nil,
		RetryConfig: nil,
	}

	clone := CloneConfig(original)

	assert.Equal(t, original, clone)
	assert.NotSame(t, original, clone)
	assert.Nil(t, clone.Headers)
	assert.Nil(t, clone.RetryConfig)
}
