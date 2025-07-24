package httpserver

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/graceful"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/echo"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/gin"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/nethttp"
)

func TestGracefulIntegration(t *testing.T) {
	// Create graceful manager
	manager := graceful.NewManager()
	manager.SetDrainTimeout(5 * time.Second)
	manager.SetShutdownTimeout(10 * time.Second)

	// Add health checks
	manager.AddHealthCheck("integration_test", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Integration test running",
			Duration:  time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	// Add hooks
	preHookCalled := false
	postHookCalled := false

	manager.AddPreShutdownHook(func() error {
		preHookCalled = true
		t.Log("Pre-shutdown hook executed")
		return nil
	})

	manager.AddPostShutdownHook(func() error {
		postHookCalled = true
		t.Log("Post-shutdown hook executed")
		return nil
	})

	// Create multiple servers
	cfg1 := &config.Config{Host: "localhost", Port: 8091}
	cfg2 := &config.Config{Host: "localhost", Port: 8092}
	cfg3 := &config.Config{Host: "localhost", Port: 8093}

	ginServer, err := gin.NewServer(cfg1)
	if err != nil {
		t.Fatalf("Failed to create Gin server: %v", err)
	}

	echoServer, err := echo.NewServer(cfg2)
	if err != nil {
		t.Fatalf("Failed to create Echo server: %v", err)
	}

	netHttpServer, err := nethttp.NewServer(cfg3)
	if err != nil {
		t.Fatalf("Failed to create net/http server: %v", err)
	}

	// Register servers
	manager.RegisterServer("gin-test", ginServer)
	manager.RegisterServer("echo-test", echoServer)
	manager.RegisterServer("nethttp-test", netHttpServer)

	// Test health status
	status := manager.GetHealthStatus()
	if status.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", status.Status)
	}
	if len(status.Checks) != 1 {
		t.Errorf("Expected 1 health check, got %d", len(status.Checks))
	}

	// Test connections management
	initialConnections := manager.GetConnectionsCount()
	manager.IncrementConnections()
	manager.IncrementConnections()
	if manager.GetConnectionsCount() != initialConnections+2 {
		t.Errorf("Expected connections to be incremented")
	}

	manager.DecrementConnections()
	if manager.GetConnectionsCount() != initialConnections+1 {
		t.Errorf("Expected connections to be decremented")
	}

	// Reset connections for clean shutdown test
	for manager.GetConnectionsCount() > 0 {
		manager.DecrementConnections()
	}

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = manager.GracefulShutdown(ctx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Verify hooks were called
	if !preHookCalled {
		t.Error("Pre-shutdown hook was not called")
	}
	if !postHookCalled {
		t.Error("Post-shutdown hook was not called")
	}

	// Test double shutdown protection
	err = manager.GracefulShutdown(ctx)
	if err == nil {
		t.Error("Expected error on second shutdown attempt")
	}
}

func TestGracefulMethodsAllProviders(t *testing.T) {
	providers := map[string]func(interface{}) (interfaces.HTTPServer, error){
		"gin":     gin.NewServer,
		"echo":    echo.NewServer,
		"nethttp": nethttp.NewServer,
	}

	for name, factory := range providers {
		t.Run(name, func(t *testing.T) {
			cfg := &config.Config{Host: "localhost", Port: 8080}
			server, err := factory(cfg)
			if err != nil {
				t.Fatalf("Failed to create %s server: %v", name, err)
			}

			// Test all graceful methods exist and work
			connections := server.GetConnectionsCount()
			if connections < 0 {
				t.Errorf("Invalid connections count: %d", connections)
			}

			status := server.GetHealthStatus()
			if status.Version == "" {
				t.Error("Health status should have version")
			}

			// Create context for testing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test hooks - need to cast to specific type for graceful methods
			if ginServer, ok := server.(*gin.Server); ok {
				ginServer.PreShutdownHook(func() error { return nil })
				ginServer.PostShutdownHook(func() error { return nil })
				ginServer.SetDrainTimeout(30 * time.Second)

				err = ginServer.WaitForConnections(ctx)
				if err != nil {
					t.Errorf("WaitForConnections failed for %s: %v", name, err)
				}
			} else if echoServer, ok := server.(*echo.Server); ok {
				echoServer.PreShutdownHook(func() error { return nil })
				echoServer.PostShutdownHook(func() error { return nil })
				echoServer.SetDrainTimeout(30 * time.Second)

				err = echoServer.WaitForConnections(ctx)
				if err != nil {
					t.Errorf("WaitForConnections failed for %s: %v", name, err)
				}
			} else if netHttpServer, ok := server.(*nethttp.Server); ok {
				netHttpServer.PreShutdownHook(func() error { return nil })
				netHttpServer.PostShutdownHook(func() error { return nil })
				netHttpServer.SetDrainTimeout(30 * time.Second)

				err = netHttpServer.WaitForConnections(ctx)
				if err != nil {
					t.Errorf("WaitForConnections failed for %s: %v", name, err)
				}
			}

			// Test graceful stop on non-running server
			err = server.GracefulStop(ctx, 5*time.Second)
			if err == nil {
				t.Errorf("Expected error calling GracefulStop on non-running %s server", name)
			}

			// Test restart on non-running server
			err = server.Restart(ctx)
			if err == nil {
				t.Errorf("Expected error calling Restart on non-running %s server", name)
			}
		})
	}
}

func TestGracefulManagerConfiguration(t *testing.T) {
	manager := graceful.NewManager()

	// Test timeout configuration
	drainTimeout := 45 * time.Second
	shutdownTimeout := 90 * time.Second

	manager.SetDrainTimeout(drainTimeout)
	manager.SetShutdownTimeout(shutdownTimeout)

	// Test health check management
	manager.AddHealthCheck("test1", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{Status: "healthy", Message: "Test 1 OK"}
	})

	manager.AddHealthCheck("test2", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{Status: "warning", Message: "Test 2 Warning"}
	})

	status := manager.GetHealthStatus()
	if len(status.Checks) != 2 {
		t.Errorf("Expected 2 health checks, got %d", len(status.Checks))
	}

	// Overall status should be warning due to one warning check
	if status.Status != "warning" {
		t.Errorf("Expected overall status 'warning', got '%s'", status.Status)
	}

	// Test server management
	cfg := &config.Config{Host: "localhost", Port: 8080}
	server, err := gin.NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	manager.RegisterServer("test", server)
	manager.UnregisterServer("test")

	// Should be able to register again after unregister
	manager.RegisterServer("test", server)
}
