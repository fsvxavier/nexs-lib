package hooks

import (
	"net/http"
	"testing"
	"time"
)

func TestNewTracingObserver(t *testing.T) {
	observer := NewTracingObserver("")

	if observer == nil {
		t.Fatal("Expected observer to be created")
	}

	if observer.serviceName != "httpserver" {
		t.Errorf("Expected default service name to be 'httpserver', got '%s'", observer.serviceName)
	}
}

func TestNewTracingObserverWithCustomName(t *testing.T) {
	observer := NewTracingObserver("my-service")

	if observer == nil {
		t.Fatal("Expected observer to be created")
	}

	if observer.serviceName != "my-service" {
		t.Errorf("Expected service name to be 'my-service', got '%s'", observer.serviceName)
	}
}

func TestTracingObserverOnStart(t *testing.T) {
	observer := NewTracingObserver("test-service")

	// OnStart should not panic
	observer.OnStart("test-server")
}

func TestTracingObserverOnStop(t *testing.T) {
	observer := NewTracingObserver("test-service")

	// OnStop should not panic
	observer.OnStop("test-server")
}

func TestTracingObserverOnRequest(t *testing.T) {
	observer := NewTracingObserver("test-service")

	req, _ := http.NewRequest("GET", "/api/test", nil)

	// OnRequest should not panic
	observer.OnRequest("test-server", req, 200, time.Millisecond*100)
}

func TestTracingObserverOnBeforeRequest(t *testing.T) {
	observer := NewTracingObserver("test-service")

	req, _ := http.NewRequest("POST", "/api/users", nil)

	// OnBeforeRequest should not panic
	observer.OnBeforeRequest("test-server", req)
}

func TestTracingObserverOnAfterRequest(t *testing.T) {
	observer := NewTracingObserver("test-service")

	req, _ := http.NewRequest("DELETE", "/api/users/123", nil)

	// OnAfterRequest should not panic
	observer.OnAfterRequest("test-server", req, 204, time.Millisecond*50)
}

func TestTracingObserverInterfaceCompliance(t *testing.T) {
	observer := NewTracingObserver("test-service")

	// Test all interface methods exist and can be called
	req, _ := http.NewRequest("GET", "/test", nil)

	observer.OnStart("test")
	observer.OnBeforeRequest("test", req)
	observer.OnRequest("test", req, 200, time.Millisecond)
	observer.OnAfterRequest("test", req, 200, time.Millisecond)
	observer.OnStop("test")
}
