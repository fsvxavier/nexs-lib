// Package config provides configuration structures and utilities for HTTP client.
package config

import (
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// DefaultConfig returns a default configuration for HTTP clients.
func DefaultConfig() *interfaces.Config {
	return &interfaces.Config{
		BaseURL:             "",
		Timeout:             30 * time.Second,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,
		InsecureSkipVerify:  false,
		Headers:             make(map[string]string),
		RetryConfig:         DefaultRetryConfig(),
		TracingEnabled:      true,
		MetricsEnabled:      true,
	}
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() *interfaces.RetryConfig {
	return &interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		RetryCondition:  DefaultRetryCondition,
	}
}

// DefaultRetryCondition defines the default condition for retrying requests.
func DefaultRetryCondition(resp *interfaces.Response, err error) bool {
	if err != nil {
		return true
	}

	// Retry on server errors (5xx) and specific client errors
	if resp != nil {
		switch resp.StatusCode {
		case 408, 429, 502, 503, 504:
			return true
		}

		// Retry on 5xx errors
		return resp.StatusCode >= 500 && resp.StatusCode < 600
	}

	return false
}

// Builder provides a fluent interface for building configurations.
type Builder struct {
	config *interfaces.Config
}

// NewBuilder creates a new configuration builder with default values.
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultConfig(),
	}
}

// WithBaseURL sets the base URL for the client.
func (b *Builder) WithBaseURL(baseURL string) *Builder {
	b.config.BaseURL = baseURL
	return b
}

// WithTimeout sets the request timeout.
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.config.Timeout = timeout
	return b
}

// WithMaxIdleConns sets the maximum number of idle connections.
func (b *Builder) WithMaxIdleConns(max int) *Builder {
	b.config.MaxIdleConns = max
	return b
}

// WithIdleConnTimeout sets the idle connection timeout.
func (b *Builder) WithIdleConnTimeout(timeout time.Duration) *Builder {
	b.config.IdleConnTimeout = timeout
	return b
}

// WithTLSHandshakeTimeout sets the TLS handshake timeout.
func (b *Builder) WithTLSHandshakeTimeout(timeout time.Duration) *Builder {
	b.config.TLSHandshakeTimeout = timeout
	return b
}

// WithDisableKeepAlives disables HTTP keep-alives.
func (b *Builder) WithDisableKeepAlives(disable bool) *Builder {
	b.config.DisableKeepAlives = disable
	return b
}

// WithDisableCompression disables HTTP compression.
func (b *Builder) WithDisableCompression(disable bool) *Builder {
	b.config.DisableCompression = disable
	return b
}

// WithInsecureSkipVerify skips TLS certificate verification.
func (b *Builder) WithInsecureSkipVerify(skip bool) *Builder {
	b.config.InsecureSkipVerify = skip
	return b
}

// WithHeaders sets default headers for all requests.
func (b *Builder) WithHeaders(headers map[string]string) *Builder {
	b.config.Headers = headers
	return b
}

// WithHeader adds a single header.
func (b *Builder) WithHeader(key, value string) *Builder {
	if b.config.Headers == nil {
		b.config.Headers = make(map[string]string)
	}
	b.config.Headers[key] = value
	return b
}

// WithRetryConfig sets the retry configuration.
func (b *Builder) WithRetryConfig(retryConfig *interfaces.RetryConfig) *Builder {
	b.config.RetryConfig = retryConfig
	return b
}

// WithMaxRetries sets the maximum number of retries.
func (b *Builder) WithMaxRetries(max int) *Builder {
	if b.config.RetryConfig == nil {
		b.config.RetryConfig = DefaultRetryConfig()
	}
	b.config.RetryConfig.MaxRetries = max
	return b
}

// WithRetryInterval sets the initial retry interval.
func (b *Builder) WithRetryInterval(interval time.Duration) *Builder {
	if b.config.RetryConfig == nil {
		b.config.RetryConfig = DefaultRetryConfig()
	}
	b.config.RetryConfig.InitialInterval = interval
	return b
}

// WithTracingEnabled enables or disables tracing.
func (b *Builder) WithTracingEnabled(enabled bool) *Builder {
	b.config.TracingEnabled = enabled
	return b
}

// WithMetricsEnabled enables or disables metrics collection.
func (b *Builder) WithMetricsEnabled(enabled bool) *Builder {
	b.config.MetricsEnabled = enabled
	return b
}

// Build returns the constructed configuration.
func (b *Builder) Build() *interfaces.Config {
	return b.config
}

// ValidateConfig validates the configuration and returns an error if invalid.
func ValidateConfig(c *interfaces.Config) error {
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}

	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 100
	}

	if c.IdleConnTimeout <= 0 {
		c.IdleConnTimeout = 90 * time.Second
	}

	if c.TLSHandshakeTimeout <= 0 {
		c.TLSHandshakeTimeout = 10 * time.Second
	}

	if c.Headers == nil {
		c.Headers = make(map[string]string)
	}

	if c.RetryConfig == nil {
		c.RetryConfig = DefaultRetryConfig()
	}

	return nil
}

// CloneConfig creates a deep copy of the configuration.
func CloneConfig(c *interfaces.Config) *interfaces.Config {
	clone := *c

	// Deep copy headers
	if c.Headers != nil {
		clone.Headers = make(map[string]string, len(c.Headers))
		for k, v := range c.Headers {
			clone.Headers[k] = v
		}
	}

	// Deep copy retry config
	if c.RetryConfig != nil {
		retryClone := *c.RetryConfig
		clone.RetryConfig = &retryClone
	}

	return &clone
}
