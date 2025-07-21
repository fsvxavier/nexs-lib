// Package config provides configuration structures for HTTP servers.
package config

import (
	"fmt"
	"time"
)

// Config holds the base configuration for HTTP servers.
type Config struct {
	// Host is the host address to bind to.
	Host string
	// Port is the port to bind to.
	Port int
	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request.
	IdleTimeout time.Duration
	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header.
	MaxHeaderBytes int
	// TLSEnabled enables TLS for the server.
	TLSEnabled bool
	// CertFile is the path to the certificate file for TLS.
	CertFile string
	// KeyFile is the path to the key file for TLS.
	KeyFile string
	// GracefulTimeout is the timeout for graceful shutdown.
	GracefulTimeout time.Duration
	// Extensions allows providers to add custom configuration fields.
	Extensions map[string]interface{}
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1024 * 1024, // 1MB
		TLSEnabled:      false,
		GracefulTimeout: 30 * time.Second,
		Extensions:      make(map[string]interface{}),
	}
}

// WithHost sets the host address.
func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

// WithPort sets the port.
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

// WithReadTimeout sets the read timeout.
func (c *Config) WithReadTimeout(timeout time.Duration) *Config {
	c.ReadTimeout = timeout
	return c
}

// WithWriteTimeout sets the write timeout.
func (c *Config) WithWriteTimeout(timeout time.Duration) *Config {
	c.WriteTimeout = timeout
	return c
}

// WithIdleTimeout sets the idle timeout.
func (c *Config) WithIdleTimeout(timeout time.Duration) *Config {
	c.IdleTimeout = timeout
	return c
}

// WithMaxHeaderBytes sets the maximum header bytes.
func (c *Config) WithMaxHeaderBytes(bytes int) *Config {
	c.MaxHeaderBytes = bytes
	return c
}

// WithTLS enables TLS with the given certificate and key files.
func (c *Config) WithTLS(certFile, keyFile string) *Config {
	c.TLSEnabled = true
	c.CertFile = certFile
	c.KeyFile = keyFile
	return c
}

// WithGracefulTimeout sets the graceful shutdown timeout.
func (c *Config) WithGracefulTimeout(timeout time.Duration) *Config {
	c.GracefulTimeout = timeout
	return c
}

// WithExtension adds a custom extension to the configuration.
func (c *Config) WithExtension(key string, value interface{}) *Config {
	if c.Extensions == nil {
		c.Extensions = make(map[string]interface{})
	}
	c.Extensions[key] = value
	return c
}

// GetExtension retrieves a custom extension from the configuration.
func (c *Config) GetExtension(key string) (interface{}, bool) {
	if c.Extensions == nil {
		return nil, false
	}
	value, exists := c.Extensions[key]
	return value, exists
}

// Addr returns the full address string.
func (c *Config) Addr() string {
	host := c.Host
	port := c.Port

	if host == "" {
		host = "localhost"
	}
	// Don't change port if it's 0 (allows random port assignment)
	return fmt.Sprintf("%s:%d", host, port)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	clone := *c
	if c.Extensions != nil {
		clone.Extensions = make(map[string]interface{})
		for k, v := range c.Extensions {
			clone.Extensions[k] = v
		}
	}
	return &clone
}
