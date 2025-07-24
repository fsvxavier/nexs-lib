package nethttp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   NetHttpClientConfig
		validate func(*netHttpClientConfig) bool
	}{
		{
			name: "TLS Enabled",
			config: func(cfg *netHttpClientConfig) {
				cfg.tlsEnabled = true
			},
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.tlsEnabled
			},
		},
		{
			name: "Max Idle Conns",
			config: func(cfg *netHttpClientConfig) {
				cfg.maxIdleConns = 100
			},
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.maxIdleConns == 100
			},
		},
		{
			name: "Max Idle Conns Per Host",
			config: func(cfg *netHttpClientConfig) {
				cfg.maxIdleConnsPerHost = 10
			},
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.maxIdleConnsPerHost == 10
			},
		},
		{
			name: "Idle Conn Timeout",
			config: func(cfg *netHttpClientConfig) {
				cfg.idleConnTimeout = 30 * time.Second
			},
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.idleConnTimeout == 30*time.Second
			},
		},
		{
			name: "Client Tracer Enabled",
			config: func(cfg *netHttpClientConfig) {
				cfg.clientTracerEnabled = true
			},
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.clientTracerEnabled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &netHttpClientConfig{}
			tt.config(cfg)
			assert.True(t, tt.validate(cfg))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := &netHttpClientConfig{}
	defaultsClient(cfg)

	assert.False(t, cfg.tlsEnabled)
	assert.Greater(t, cfg.maxIdleConns, 0)
	assert.Greater(t, cfg.maxIdleConnsPerHost, 0)
	assert.Greater(t, cfg.maxConnsPerHost, 0)
	assert.NotZero(t, cfg.idleConnTimeout)
}
func TestGetClientConfig(t *testing.T) {
	config := GetClientConfig()

	// Test that GetClientConfig returns a zero-value netHttpClientConfig
	assert.Equal(t, netHttpClientConfig{}, config)

	// Verify all fields have zero values
	assert.Equal(t, 0, config.maxIdleConns)
	assert.Equal(t, 0, config.maxIdleConnsPerHost)
	assert.Equal(t, 0, config.maxConnsPerHost)
	assert.Equal(t, time.Duration(0), config.idleConnTimeout)
	assert.False(t, config.tlsEnabled)
	assert.False(t, config.disableKeepAlives)
	assert.False(t, config.clientTracerEnabled)
	assert.Equal(t, time.Duration(0), config.clientTimeout)
}
func TestConfigSetters(t *testing.T) {
	tests := []struct {
		name     string
		setter   NetHttpClientConfig
		validate func(*netHttpClientConfig) bool
	}{
		{
			name:   "SetMaxIdleConnsPerHost",
			setter: SetMaxIdleConnsPerHost(100),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.maxIdleConnsPerHost == 100
			},
		},
		{
			name:   "SetMaxIdleConns",
			setter: SetMaxIdleConns(200),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.maxIdleConns == 200
			},
		},
		{
			name:   "SetMaxConns",
			setter: SetMaxConns(300),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.maxConnsPerHost == 300
			},
		},
		{
			name:   "SetDisableKeepAlives",
			setter: SetDisableKeepAlives(true),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.disableKeepAlives
			},
		},
		{
			name:   "SetIdleConnTimeout",
			setter: SetIdleConnTimeout(5 * time.Minute),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.idleConnTimeout == 5*time.Minute
			},
		},
		{
			name:   "SetTracerEnabled",
			setter: SetTracerEnabled(true),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.clientTracerEnabled
			},
		},
		{
			name:   "SetClientTimeout",
			setter: SetClientTimeout(10 * time.Second),
			validate: func(cfg *netHttpClientConfig) bool {
				return cfg.clientTimeout == 10*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &netHttpClientConfig{}
			tt.setter(cfg)
			assert.True(t, tt.validate(cfg))
		})
	}
}
func TestSetTlsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "Enable TLS",
			enabled: true,
		},
		{
			name:    "Disable TLS",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &netHttpClientConfig{}
			SetTlsEnabled(tt.enabled)(cfg)
			assert.Equal(t, tt.enabled, cfg.tlsEnabled)
		})
	}
}
