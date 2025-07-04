package common

import (
	"context"
	"net/http"
	"time"
)

// ServerOption represents a function to configure a Server
type ServerOption func(*ServerConfig)

// ServerConfig represents the configuration for HTTP servers
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
	EnableMetrics   bool
	EnablePprof     bool
	EnableSwagger   bool
	EnableTracing   bool
	EnableLogging   bool
	Prefork         bool
}

// DefaultServerConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:            "0.0.0.0",
		Port:            "8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1 MB
		EnableMetrics:   false,
		EnablePprof:     false,
		EnableSwagger:   false,
		EnableTracing:   false,
		EnableLogging:   true,
		Prefork:         false,
	}
}

// WithHost sets the server host
func WithHost(host string) ServerOption {
	return func(c *ServerConfig) {
		c.Host = host
	}
}

// WithPort sets the server port
func WithPort(port string) ServerOption {
	return func(c *ServerConfig) {
		c.Port = port
	}
}

// WithReadTimeout sets the read timeout
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the idle timeout
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.IdleTimeout = timeout
	}
}

// WithShutdownTimeout sets the shutdown timeout
func WithShutdownTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.ShutdownTimeout = timeout
	}
}

// WithMaxHeaderBytes sets the max header bytes
func WithMaxHeaderBytes(maxHeaderBytes int) ServerOption {
	return func(c *ServerConfig) {
		c.MaxHeaderBytes = maxHeaderBytes
	}
}

// WithMetrics enables metrics
func WithMetrics(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.EnableMetrics = enable
	}
}

// WithPprof enables pprof
func WithPprof(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.EnablePprof = enable
	}
}

// WithSwagger enables swagger
func WithSwagger(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.EnableSwagger = enable
	}
}

// WithTracing enables tracing
func WithTracing(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.EnableTracing = enable
	}
}

// WithLogging enables logging
func WithLogging(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.EnableLogging = enable
	}
}

// WithPrefork enables prefork mode (for servers that support it)
func WithPrefork(enable bool) ServerOption {
	return func(c *ServerConfig) {
		c.Prefork = enable
	}
}

// Server defines the common interface for HTTP servers
type Server interface {
	// Start starts the HTTP server
	Start() error

	// Shutdown gracefully shuts down the server
	Shutdown(ctx context.Context) error

	// Address returns the server's address (host:port)
	Address() string

	// Health returns the server's health status
	Health() http.Handler
}
