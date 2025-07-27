package http2

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func TestHTTP2Provider_BasicCreation(t *testing.T) {
	config := &interfaces.Config{}
	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be initialized")
	}

	if provider.Name() != "http2" {
		t.Errorf("Expected provider name to be 'http2', got '%s'", provider.Name())
	}

	if provider.Version() == "" {
		t.Error("Expected provider version to be set")
	}
}

func TestHTTP2Provider_Configuration(t *testing.T) {
	config := &interfaces.Config{
		Timeout:         10 * time.Second,
		MaxIdleConns:    50,
		IdleConnTimeout: 60 * time.Second,
	}

	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	if provider.config.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10s, got %v", provider.config.Timeout)
	}

	if provider.config.MaxIdleConns != 50 {
		t.Errorf("Expected MaxIdleConns to be 50, got %d", provider.config.MaxIdleConns)
	}
}

func TestHTTP2Provider_IsHealthy(t *testing.T) {
	config := &interfaces.Config{}
	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	if !provider.IsHealthy() {
		t.Error("Expected provider to be healthy initially")
	}
}

func TestHTTP2Provider_GetMetrics(t *testing.T) {
	config := &interfaces.Config{}
	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	metrics := provider.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be initialized")
	}

	if metrics.TotalRequests != 0 {
		t.Errorf("Expected initial total requests to be 0, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 0 {
		t.Errorf("Expected initial successful requests to be 0, got %d", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 0 {
		t.Errorf("Expected initial failed requests to be 0, got %d", metrics.FailedRequests)
	}
}

func TestMultiplexedClient_Creation(t *testing.T) {
	config := &interfaces.Config{}
	multiplexed, err := NewMultiplexedClient(config)
	if err != nil {
		t.Fatalf("Failed to create multiplexed client: %v", err)
	}

	if multiplexed == nil {
		t.Fatal("Expected multiplexed client to be created")
	}
}

func TestMultiplexedClient_EmptyRequests(t *testing.T) {
	config := &interfaces.Config{}
	multiplexed, err := NewMultiplexedClient(config)
	if err != nil {
		t.Fatalf("Failed to create multiplexed client: %v", err)
	}

	responses, err := multiplexed.ExecuteConcurrent(context.Background(), []*interfaces.Request{})
	if err != nil {
		t.Fatalf("Expected no error for empty requests, got %v", err)
	}

	if len(responses) != 0 {
		t.Errorf("Expected 0 responses for empty requests, got %d", len(responses))
	}
}

func TestConnectionMonitor_Basic(t *testing.T) {
	// Test basic connection monitor functionality
	config := &interfaces.Config{}
	client, err := NewHTTP2Client(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 client: %v", err)
	}

	monitor := NewConnectionMonitor(client, time.Second)
	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	// Test start/stop functionality
	ctx, cancel := context.WithCancel(context.Background())

	// Start monitoring in background
	go monitor.Start(ctx, "http://example.com/health")

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop monitoring
	cancel()
	monitor.Stop()

	// Should not panic or error
}

func TestHTTP2Provider_InvalidRequest(t *testing.T) {
	config := &interfaces.Config{}
	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	// Test with invalid URL
	request := &interfaces.Request{
		Method:  "GET",
		URL:     "invalid-url",
		Headers: map[string]string{"Accept": "application/json"},
		Timeout: 5 * time.Second,
	}

	_, err = provider.DoRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestHTTP2Provider_ContextCancellation(t *testing.T) {
	config := &interfaces.Config{}
	provider, err := NewHTTP2Provider(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 provider: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	request := &interfaces.Request{
		Method:  "GET",
		URL:     "http://httpbin.org/delay/5", // This would take time
		Headers: map[string]string{"Accept": "application/json"},
		Timeout: 5 * time.Second,
	}

	_, err = provider.DoRequest(ctx, request)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestHTTP2Client_Creation(t *testing.T) {
	config := &interfaces.Config{}
	client, err := NewHTTP2Client(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP/2 client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created")
	}
}
