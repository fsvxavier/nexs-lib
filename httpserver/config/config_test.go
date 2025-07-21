package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("Expected ReadTimeout 30s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("Expected WriteTimeout 30s, got %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", cfg.IdleTimeout)
	}
	if cfg.MaxHeaderBytes != 1024*1024 {
		t.Errorf("Expected MaxHeaderBytes 1MB, got %d", cfg.MaxHeaderBytes)
	}
	if cfg.TLSEnabled {
		t.Error("Expected TLSEnabled false, got true")
	}
	if cfg.GracefulTimeout != 30*time.Second {
		t.Errorf("Expected GracefulTimeout 30s, got %v", cfg.GracefulTimeout)
	}
	if cfg.Extensions == nil {
		t.Error("Expected Extensions to be initialized")
	}
}

func TestConfigWithHost(t *testing.T) {
	cfg := DefaultConfig().WithHost("0.0.0.0")
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Host)
	}
}

func TestConfigWithPort(t *testing.T) {
	cfg := DefaultConfig().WithPort(9090)
	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}
}

func TestConfigWithReadTimeout(t *testing.T) {
	timeout := 45 * time.Second
	cfg := DefaultConfig().WithReadTimeout(timeout)
	if cfg.ReadTimeout != timeout {
		t.Errorf("Expected ReadTimeout %v, got %v", timeout, cfg.ReadTimeout)
	}
}

func TestConfigWithWriteTimeout(t *testing.T) {
	timeout := 45 * time.Second
	cfg := DefaultConfig().WithWriteTimeout(timeout)
	if cfg.WriteTimeout != timeout {
		t.Errorf("Expected WriteTimeout %v, got %v", timeout, cfg.WriteTimeout)
	}
}

func TestConfigWithIdleTimeout(t *testing.T) {
	timeout := 120 * time.Second
	cfg := DefaultConfig().WithIdleTimeout(timeout)
	if cfg.IdleTimeout != timeout {
		t.Errorf("Expected IdleTimeout %v, got %v", timeout, cfg.IdleTimeout)
	}
}

func TestConfigWithMaxHeaderBytes(t *testing.T) {
	bytes := 2048 * 1024
	cfg := DefaultConfig().WithMaxHeaderBytes(bytes)
	if cfg.MaxHeaderBytes != bytes {
		t.Errorf("Expected MaxHeaderBytes %d, got %d", bytes, cfg.MaxHeaderBytes)
	}
}

func TestConfigWithTLS(t *testing.T) {
	certFile := "/path/to/cert.pem"
	keyFile := "/path/to/key.pem"
	cfg := DefaultConfig().WithTLS(certFile, keyFile)

	if !cfg.TLSEnabled {
		t.Error("Expected TLSEnabled true, got false")
	}
	if cfg.CertFile != certFile {
		t.Errorf("Expected CertFile '%s', got '%s'", certFile, cfg.CertFile)
	}
	if cfg.KeyFile != keyFile {
		t.Errorf("Expected KeyFile '%s', got '%s'", keyFile, cfg.KeyFile)
	}
}

func TestConfigWithGracefulTimeout(t *testing.T) {
	timeout := 60 * time.Second
	cfg := DefaultConfig().WithGracefulTimeout(timeout)
	if cfg.GracefulTimeout != timeout {
		t.Errorf("Expected GracefulTimeout %v, got %v", timeout, cfg.GracefulTimeout)
	}
}

func TestConfigWithExtension(t *testing.T) {
	cfg := DefaultConfig().WithExtension("test.key", "test.value")

	value, exists := cfg.GetExtension("test.key")
	if !exists {
		t.Error("Expected extension to exist")
	}
	if value != "test.value" {
		t.Errorf("Expected extension value 'test.value', got '%v'", value)
	}
}

func TestConfigGetExtensionNotFound(t *testing.T) {
	cfg := DefaultConfig()

	value, exists := cfg.GetExtension("nonexistent")
	if exists {
		t.Error("Expected extension not to exist")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

func TestConfigAddr(t *testing.T) {
	cfg := DefaultConfig()
	expected := "localhost:8080"
	if cfg.Addr() != expected {
		t.Errorf("Expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestConfigAddrWithCustomValues(t *testing.T) {
	cfg := DefaultConfig().WithHost("0.0.0.0").WithPort(9090)
	expected := "0.0.0.0:9090"
	if cfg.Addr() != expected {
		t.Errorf("Expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestConfigAddrWithEmptyHost(t *testing.T) {
	cfg := &Config{Port: 8080}
	expected := "localhost:8080"
	if cfg.Addr() != expected {
		t.Errorf("Expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestConfigAddrWithZeroPort(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 0}
	expected := "localhost:0"
	if cfg.Addr() != expected {
		t.Errorf("Expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestConfigClone(t *testing.T) {
	original := DefaultConfig().
		WithHost("test.host").
		WithPort(9999).
		WithExtension("test.key", "test.value")

	clone := original.Clone()

	// Verify clone has same values
	if clone.Host != original.Host {
		t.Errorf("Expected cloned host '%s', got '%s'", original.Host, clone.Host)
	}
	if clone.Port != original.Port {
		t.Errorf("Expected cloned port %d, got %d", original.Port, clone.Port)
	}

	value, exists := clone.GetExtension("test.key")
	if !exists {
		t.Error("Expected extension to exist in clone")
	}
	if value != "test.value" {
		t.Errorf("Expected cloned extension value 'test.value', got '%v'", value)
	}

	// Verify they are different instances
	clone.WithHost("modified.host")
	if original.Host == clone.Host {
		t.Error("Expected original and clone to be independent")
	}

	// Verify extensions are independent
	clone.WithExtension("test.key", "modified.value")
	originalValue, _ := original.GetExtension("test.key")
	if originalValue == "modified.value" {
		t.Error("Expected original and clone extensions to be independent")
	}
}

func TestConfigChaining(t *testing.T) {
	cfg := DefaultConfig().
		WithHost("example.com").
		WithPort(3000).
		WithReadTimeout(15*time.Second).
		WithWriteTimeout(15*time.Second).
		WithTLS("/cert.pem", "/key.pem").
		WithExtension("env", "production")

	if cfg.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", cfg.WriteTimeout)
	}
	if !cfg.TLSEnabled {
		t.Error("Expected TLS to be enabled")
	}
	if cfg.CertFile != "/cert.pem" {
		t.Errorf("Expected CertFile '/cert.pem', got '%s'", cfg.CertFile)
	}
	if cfg.KeyFile != "/key.pem" {
		t.Errorf("Expected KeyFile '/key.pem', got '%s'", cfg.KeyFile)
	}

	env, exists := cfg.GetExtension("env")
	if !exists {
		t.Error("Expected env extension to exist")
	}
	if env != "production" {
		t.Errorf("Expected env 'production', got '%v'", env)
	}
}
