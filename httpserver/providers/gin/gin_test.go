package gin

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

	ginServer, ok := server.(*Server)
	if !ok {
		t.Fatal("Expected server to be of type *gin.Server")
	}

	if ginServer.GetGinEngine() == nil {
		t.Error("Expected Gin engine to be initialized")
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

func TestServerSetHandler(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	server.SetHandler(handler)

	ginServer := server.(*Server)
	if ginServer.GetGinEngine() == nil {
		t.Error("Expected Gin engine to be set after SetHandler")
	}
}

func TestServerGetAddr(t *testing.T) {
	cfg := config.DefaultConfig().WithHost("localhost").WithPort(9090)
	server, _ := NewServer(cfg)

	addr := server.GetAddr()
	expected := "localhost:9090"

	if addr != expected {
		t.Errorf("Expected addr %s, got %s", expected, addr)
	}
}

func TestServerIsRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	if server.IsRunning() {
		t.Error("Expected server to not be running initially")
	}
}

func TestServerGetConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	returnedConfig := server.(*Server).GetConfig()
	if returnedConfig != cfg {
		t.Error("Expected same config instance")
	}
}

func TestServerGetHTTPServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	httpServer := server.(*Server).GetHTTPServer()
	if httpServer == nil {
		t.Error("Expected HTTP server to be available")
	}
}

func TestServerStartAndStop(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0) // Use random port
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(handler)

	// Test start
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running after start")
	}

	// Test stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsRunning() {
		t.Error("Expected server to not be running after stop")
	}
}

func TestServerStartAlreadyRunning(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0)
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(handler)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
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
		t.Errorf("Expected no error stopping non-running server, got %v", err)
	}
}

func TestServerStopWithNilContext(t *testing.T) {
	cfg := config.DefaultConfig().WithPort(0)
	server, _ := NewServer(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(handler)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Stop with nil context (should use default timeout)
	err := server.Stop(nil)
	if err != nil {
		t.Errorf("Expected no error stopping with nil context, got %v", err)
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
		t.Error("Expected server to be of type *gin.Server")
	}
}

func TestGracefulMethods(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ginServer := server.(*Server)

	// Test GetConnectionsCount
	count := ginServer.GetConnectionsCount()
	if count != 0 {
		t.Errorf("Expected 0 connections, got %d", count)
	}

	// Test GetHealthStatus
	status := ginServer.GetHealthStatus()
	if status.Status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", status.Status)
	}
	if status.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", status.Version)
	}

	// Test hooks
	hook := func() error {
		return nil
	}

	ginServer.PreShutdownHook(hook)
	ginServer.PostShutdownHook(hook)

	// Test SetDrainTimeout
	timeout := 45 * time.Second
	ginServer.SetDrainTimeout(timeout)

	// Test WaitForConnections
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = ginServer.WaitForConnections(ctx)
	if err != nil {
		t.Errorf("Expected no error from WaitForConnections, got %v", err)
	}
}

func TestGracefulStopNotRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ginServer := server.(*Server)
	ctx := context.Background()

	err = ginServer.GracefulStop(ctx, 5*time.Second)
	if err == nil {
		t.Error("Expected error when calling GracefulStop on non-running server")
	}

	err = ginServer.Restart(ctx)
	if err == nil {
		t.Error("Expected error when calling Restart on non-running server")
	}
}
