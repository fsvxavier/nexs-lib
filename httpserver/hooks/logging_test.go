package hooks

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewLoggingObserver(t *testing.T) {
	observer := NewLoggingObserver(nil)

	if observer == nil {
		t.Fatal("Expected observer to be created")
	}

	if observer.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestNewLoggingObserverWithCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	observer := NewLoggingObserver(logger)

	if observer == nil {
		t.Fatal("Expected observer to be created")
	}

	if observer.logger != logger {
		t.Error("Expected custom logger to be used")
	}
}

func TestLoggingObserverOnStart(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := NewLoggingObserver(logger)

	observer.OnStart("test-server")

	output := buf.String()
	if !strings.Contains(output, "Server test-server started") {
		t.Errorf("Expected start message in output, got: %s", output)
	}
}

func TestLoggingObserverOnStop(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := NewLoggingObserver(logger)

	observer.OnStop("test-server")

	output := buf.String()
	if !strings.Contains(output, "Server test-server stopped") {
		t.Errorf("Expected stop message in output, got: %s", output)
	}
}

func TestLoggingObserverOnRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := NewLoggingObserver(logger)

	req, _ := http.NewRequest("GET", "/api/test", nil)
	observer.OnRequest("test-server", req, 200, time.Millisecond*100)

	output := buf.String()
	expected := []string{
		"test-server",
		"GET",
		"/api/test",
		"Status: 200",
		"Duration:",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected '%s' in output, got: %s", exp, output)
		}
	}
}

func TestLoggingObserverOnBeforeRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := NewLoggingObserver(logger)

	req, _ := http.NewRequest("POST", "/api/users", nil)
	observer.OnBeforeRequest("test-server", req)

	output := buf.String()
	expected := []string{
		"test-server",
		"Processing",
		"POST",
		"/api/users",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected '%s' in output, got: %s", exp, output)
		}
	}
}

func TestLoggingObserverOnAfterRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	observer := NewLoggingObserver(logger)

	req, _ := http.NewRequest("DELETE", "/api/users/123", nil)
	observer.OnAfterRequest("test-server", req, 204, time.Millisecond*50)

	output := buf.String()
	expected := []string{
		"test-server",
		"Completed",
		"DELETE",
		"/api/users/123",
		"Status: 204",
		"Duration:",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected '%s' in output, got: %s", exp, output)
		}
	}
}
