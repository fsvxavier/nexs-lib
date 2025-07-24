package nethttp

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg)

	if err != nil {
		t.Fatalf("Expected no error creating server, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.GetAddr() != cfg.Addr() {
		t.Errorf("Expected addr '%s', got '%s'", cfg.Addr(), server.GetAddr())
	}
}

func TestNewServerWithInvalidConfig(t *testing.T) {
	server, err := NewServer("invalid config")

	if err == nil {
		t.Error("Expected error for invalid config type")
	}

	if server != nil {
		t.Error("Expected nil server for invalid config")
	}
}

func TestNewServerWithNilConfig(t *testing.T) {
	server, err := NewServer(nil)

	if err != nil {
		t.Fatalf("Expected no error with nil config, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}
}

func TestServerSetHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server.SetHandler(handler)

	netServer := server.(*Server)
	if netServer.handler == nil {
		t.Error("Expected handler to be set")
	}
}

func TestServerGetAddr(t *testing.T) {
	cfg := config.DefaultConfig().WithHost("example.com").WithPort(9090)
	server, _ := NewServer(cfg)

	expectedAddr := "example.com:9090"
	if server.GetAddr() != expectedAddr {
		t.Errorf("Expected addr '%s', got '%s'", expectedAddr, server.GetAddr())
	}
}

func TestServerIsRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	if server.IsRunning() {
		t.Error("Expected server not to be running initially")
	}
}

func TestServerGetConfig(t *testing.T) {
	cfg := config.DefaultConfig().WithHost("test.host")
	server, _ := NewServer(cfg)

	netServer := server.(*Server)
	retrievedConfig := netServer.GetConfig()

	if retrievedConfig.Host != cfg.Host {
		t.Errorf("Expected config host '%s', got '%s'", cfg.Host, retrievedConfig.Host)
	}
}

func TestServerGetHTTPServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	netServer := server.(*Server)
	httpServer := netServer.GetHTTPServer()

	if httpServer == nil {
		t.Error("Expected HTTP server to be available")
	}

	if httpServer.Addr != cfg.Addr() {
		t.Errorf("Expected HTTP server addr '%s', got '%s'", cfg.Addr(), httpServer.Addr)
	}
}

func TestServerStartAndStop(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0) // Use random port
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(handler)

	// Test Start
	err := server.Start()
	if err != nil {
		t.Fatalf("Expected no error starting server, got: %v", err)
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running after start")
	}

	// Test Stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Fatalf("Expected no error stopping server, got: %v", err)
	}

	if server.IsRunning() {
		t.Error("Expected server not to be running after stop")
	}
}

func TestServerStartAlreadyRunning(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0)
	server, _ := NewServer(cfg)

	server.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Try to start again
	err := server.Start()
	if err == nil {
		t.Error("Expected error when starting already running server")
	}
}

func TestServerStopNotRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Expected no error stopping non-running server, got: %v", err)
	}
}

func TestServerStopWithNilContext(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0)
	server, _ := NewServer(cfg)

	server.Start()

	err := server.Stop(nil)
	if err != nil {
		t.Errorf("Expected no error stopping with nil context, got: %v", err)
	}

	if server.IsRunning() {
		t.Error("Expected server not to be running after stop")
	}
}

func TestServerWithTLS(t *testing.T) {
	cfg := config.DefaultConfig().
		WithTLS("/nonexistent/cert.pem", "/nonexistent/key.pem").
		WithPort(0)

	server, _ := NewServer(cfg)

	// This should fail because the cert/key files don't exist
	err := server.Start()
	if err == nil {
		t.Error("Expected error starting TLS server with nonexistent files")
	}
}

func TestServerWithTLSMissingFiles(t *testing.T) {
	cfg := config.DefaultConfig().WithTLS("", "").WithPort(0)
	server, _ := NewServer(cfg)

	err := server.Start()
	if err == nil {
		t.Error("Expected error starting TLS server with missing cert/key files")
	}
}

func TestFactory(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := Factory(cfg)

	if err != nil {
		t.Fatalf("Expected no error from factory, got: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server from factory")
	}

	// Verify it's actually a nethttp server
	_, ok := server.(*Server)
	if !ok {
		t.Error("Expected factory to return nethttp.Server")
	}
}
