package fiber

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestNewServerWithValidConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := NewServer(cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	fiberServer, ok := server.(*Server)
	if !ok {
		t.Fatal("Expected server to be of type *fiber.Server")
	}

	if fiberServer.GetConfig() != cfg {
		t.Error("Expected server config to match provided config")
	}
}

func TestNewServerWithInvalidConfig(t *testing.T) {
	server, err := NewServer("invalid")

	if err == nil {
		t.Error("Expected error with invalid config type")
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
	})

	server.SetHandler(handler)

	fiberServer := server.(*Server)
	if fiberServer.handler == nil {
		t.Error("Expected handler to be set")
	}
}

func TestServerGetAddr(t *testing.T) {
	cfg := &config.Config{
		Host: "localhost",
		Port: 8080,
	}
	server, _ := NewServer(cfg)

	addr := server.GetAddr()
	expected := "localhost:8080"

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

	fiberServer := server.(*Server)
	if fiberServer.GetConfig() != cfg {
		t.Error("Expected GetConfig to return the same config")
	}
}

func TestServerGetHTTPServer(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	fiberServer := server.(*Server)
	httpServer := fiberServer.GetHTTPServer()
	if httpServer != nil {
		t.Error("Expected GetHTTPServer to return nil for Fiber")
	}
}

func TestServerGetFiberApp(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	fiberServer := server.(*Server)
	app := fiberServer.GetFiberApp()

	if app == nil {
		t.Error("Expected GetFiberApp to return a valid Fiber app")
	}
}

func TestServerStartStop(t *testing.T) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            0, // Random port
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     60 * time.Second,
		GracefulTimeout: 5 * time.Second,
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Set a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})
	server.SetHandler(handler)

	// Test starting the server
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Give server time to start
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	case <-time.After(2 * time.Second):
		// Server started successfully
	}

	// Check if server is running
	if !server.IsRunning() {
		t.Error("Expected server to be running after Start()")
	}

	// Test stopping the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Check if server is stopped
	if server.IsRunning() {
		t.Error("Expected server to be stopped after Stop()")
	}
}

func TestServerStartAlreadyRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Port = 0 // Random port

	server, _ := NewServer(cfg)
	fiberServer := server.(*Server)

	// Manually set running to true
	fiberServer.running = true

	err := server.Start()
	if err == nil {
		t.Error("Expected error when starting already running server")
	}
}

func TestServerStopNotRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	ctx := context.Background()
	err := server.Stop(ctx)

	if err != nil {
		t.Errorf("Expected no error when stopping non-running server, got %v", err)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	cfg := config.DefaultConfig()
	server, _ := NewServer(cfg)

	// Test that server implements interfaces.HTTPServer
	var _ interfaces.HTTPServer = server
}

func TestFactory(t *testing.T) {
	cfg := config.DefaultConfig()
	server, err := Factory(cfg)

	if err != nil {
		t.Fatalf("Factory failed: %v", err)
	}

	_, ok := server.(*Server)
	if !ok {
		t.Error("Expected factory to return fiber.Server")
	}
}
