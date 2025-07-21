package echo

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
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	echoServer, ok := server.(*Server)
	if !ok {
		t.Fatal("Expected server to be of type *echo.Server")
	}

	if echoServer.GetEcho() == nil {
		t.Error("Expected Echo instance to be initialized")
	}
}

func TestNewServerWithInvalidConfig(t *testing.T) {
	server, err := NewServer("invalid")

	if err == nil {
		t.Error("Expected error with invalid config")
	}

	if server != nil {
		t.Error("Expected nil server with invalid config")
	}
}

func TestNewServerWithNilConfig(t *testing.T) {
	server, err := NewServer(nil)

	if err != nil {
		t.Fatalf("Expected no error with nil config, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created with default config")
	}
}

func TestFactory(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := Factory(cfg)

	if err != nil {
		t.Fatalf("Expected no error from factory, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server from factory")
	}

	_, ok := server.(*Server)
	if !ok {
		t.Error("Expected server to be of type *echo.Server")
	}
}

func TestServerStart(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0) // Random port
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	echoServer := server.(*Server)

	// Test starting server
	go func() {
		err := server.Start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Unexpected start error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	if !echoServer.IsRunning() {
		t.Error("Expected server to be running")
	}

	// Test stopping server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	if echoServer.IsRunning() {
		t.Error("Expected server to be stopped")
	}
}

func TestServerStartAlreadyRunning(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0)
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server first time
	go func() {
		server.Start()
	}()
	time.Sleep(100 * time.Millisecond)

	// Try to start again
	err = server.Start()
	if err == nil {
		t.Error("Expected error when starting already running server")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestServerSetHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	server.SetHandler(handler)

	// Test GetAddr
	addr := server.GetAddr()
	if addr == "" {
		t.Error("Expected non-empty address")
	}
}

func TestServerWithTLS(t *testing.T) {
	cfg := config.DefaultConfig().WithTLS("cert.pem", "key.pem")
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server with TLS: %v", err)
	}

	echoServer := server.(*Server)
	if !echoServer.config.TLSEnabled {
		t.Error("Expected TLS to be enabled")
	}
}

func TestServerGetAddr(t *testing.T) {
	cfg := config.DefaultConfig().WithHost("example.com").WithPort(9999)
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	addr := server.GetAddr()
	expected := "example.com:9999"
	if addr != expected {
		t.Errorf("Expected addr %s, got %s", expected, addr)
	}
}
