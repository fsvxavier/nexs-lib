package ip

import (
	"net/http"
	"testing"

	"github.com/fsvxavier/nexs-lib/ip/providers/nethttp"
)

func TestFactoryWithNetHTTP(t *testing.T) {
	// Create a mock HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set some headers
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	// Test GetRealIP
	clientIP := GetRealIP(req)
	if clientIP == "" {
		t.Error("Expected non-empty client IP")
	}
	t.Logf("Client IP: %s", clientIP)

	// Test GetRealIPInfo
	ipInfo := GetRealIPInfo(req)
	if ipInfo == nil {
		t.Error("Expected non-nil IP info")
	}
	if ipInfo != nil {
		t.Logf("IP Info: %s, Type: %s, Source: %s", ipInfo.IP, ipInfo.Type.String(), ipInfo.Source)
	}

	// Test GetIPChain
	ipChain := GetIPChain(req)
	if len(ipChain) == 0 {
		t.Error("Expected non-empty IP chain")
	}
	t.Logf("IP Chain: %v", ipChain)
}

func TestDirectAdapterUsage(t *testing.T) {
	// Create a mock HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("CF-Connecting-IP", "198.51.100.42")
	req.RemoteAddr = "172.16.0.1:8080"

	// Create adapter directly
	adapter := nethttp.NewAdapter(req)
	extractor := NewExtractor()

	// Test with direct adapter usage
	clientIP := extractor.GetRealIP(adapter)
	if clientIP != "198.51.100.42" {
		t.Errorf("Expected CF-Connecting-IP to be prioritized, got: %s", clientIP)
	}

	ipInfo := extractor.GetRealIPInfo(adapter)
	if ipInfo == nil {
		t.Error("Expected non-nil IP info")
	}
	if ipInfo != nil && ipInfo.Source != "CF-Connecting-IP" {
		t.Errorf("Expected source to be CF-Connecting-IP, got: %s", ipInfo.Source)
	}
}

func TestSupportedFrameworks(t *testing.T) {
	frameworks := GetSupportedFrameworks()
	if len(frameworks) == 0 {
		t.Error("Expected at least one supported framework")
	}

	t.Logf("Supported frameworks: %v", frameworks)

	// Check for expected frameworks
	expectedFrameworks := []string{"net/http", "fiber", "gin", "echo", "fasthttp", "atreugo"}
	for _, expected := range expectedFrameworks {
		found := false
		for _, framework := range frameworks {
			if framework == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected framework %s not found in supported list", expected)
		}
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that the old API still works
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("X-Real-IP", "203.0.113.100")
	req.RemoteAddr = "192.168.1.1:8080"

	// These should work exactly like before
	clientIP := GetRealIP(req)
	if clientIP != "203.0.113.100" {
		t.Errorf("Expected 203.0.113.100, got: %s", clientIP)
	}

	ipInfo := GetRealIPInfo(req)
	if ipInfo == nil {
		t.Error("Expected non-nil IP info")
	}

	// Test utility functions
	if !IsValidIP("192.168.1.1") {
		t.Error("Expected 192.168.1.1 to be valid")
	}

	if !IsPrivateIP(ipInfo.IP) && ipInfo.Type == IPTypePrivate {
		t.Error("Private IP detection mismatch")
	}
}

func TestErrorHandling(t *testing.T) {
	// Test with nil request
	clientIP := GetRealIP(nil)
	if clientIP != "" {
		t.Errorf("Expected empty string for nil request, got: %s", clientIP)
	}

	ipInfo := GetRealIPInfo(nil)
	if ipInfo != nil {
		t.Error("Expected nil IP info for nil request")
	}

	// Test with unsupported request type
	unsupportedReq := "not a supported request type"
	clientIP = GetRealIP(unsupportedReq)
	if clientIP != "" {
		t.Errorf("Expected empty string for unsupported request type, got: %s", clientIP)
	}
}
