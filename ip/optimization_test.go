package ip

import (
	"net/http"
	"testing"

	"github.com/fsvxavier/nexs-lib/ip/providers"
)

func TestOptimizedExtractor_GetRealIP(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		remoteAddr  string
		expectedIP  string
		description string
	}{
		{
			name: "X-Forwarded-For with public IP",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 192.168.1.1",
			},
			remoteAddr:  "192.168.1.1:8080",
			expectedIP:  "203.0.113.195",
			description: "Should extract public IP from X-Forwarded-For header",
		},
		{
			name: "CF-Connecting-IP priority",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.100",
				"X-Forwarded-For":  "203.0.113.195, 192.168.1.1",
			},
			remoteAddr:  "192.168.1.1:8080",
			expectedIP:  "203.0.113.100",
			description: "Should prioritize CF-Connecting-IP over X-Forwarded-For",
		},
		{
			name: "Forwarded header RFC 7239",
			headers: map[string]string{
				"Forwarded": `for="203.0.113.195:8080"`,
			},
			remoteAddr:  "192.168.1.1:8080",
			expectedIP:  "203.0.113.195",
			description: "Should parse RFC 7239 Forwarded header",
		},
		{
			name: "IPv6 address",
			headers: map[string]string{
				"X-Real-IP": "2001:db8::1",
			},
			remoteAddr:  "[::1]:8080",
			expectedIP:  "2001:db8::1",
			description: "Should handle IPv6 addresses",
		},
		{
			name: "Private IP fallback",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
			},
			remoteAddr:  "192.168.1.1:8080",
			expectedIP:  "192.168.1.100",
			description: "Should return private IP when no public IP available",
		},
		{
			name:        "Remote address fallback",
			headers:     map[string]string{},
			remoteAddr:  "203.0.113.195:8080",
			expectedIP:  "203.0.113.195",
			description: "Should fallback to remote address when no headers",
		},
	}

	extractor := NewOptimizedExtractor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			adapter, err := providers.CreateAdapter(req)
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			result := extractor.GetRealIPOptimized(adapter)

			if result != tt.expectedIP {
				t.Errorf("%s: expected IP %s, got %s", tt.description, tt.expectedIP, result)
			}
		})
	}
}

func TestOptimizedExtractor_GetRealIPInfo(t *testing.T) {
	extractor := NewOptimizedExtractor()

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	adapter, err := providers.CreateAdapter(req)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ipInfo := extractor.GetRealIPInfoOptimized(adapter)

	if ipInfo == nil {
		t.Fatal("Expected non-nil IP info")
	}

	if ipInfo.IP.String() != "203.0.113.195" {
		t.Errorf("Expected IP 203.0.113.195, got %s", ipInfo.IP.String())
	}

	if !ipInfo.IsPublic {
		t.Error("Expected IP to be marked as public")
	}

	if ipInfo.Source != "X-Real-IP" {
		t.Errorf("Expected source to be X-Real-IP, got %s", ipInfo.Source)
	}
}

func TestOptimizedExtractor_GetIPChain(t *testing.T) {
	extractor := NewOptimizedExtractor()

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1")
	req.Header.Set("X-Real-IP", "203.0.113.100")
	req.RemoteAddr = "10.0.0.1:8080"

	adapter, err := providers.CreateAdapter(req)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	chain := extractor.GetIPChainOptimized(adapter)

	expectedIPs := []string{"203.0.113.195", "192.168.1.1", "203.0.113.100", "10.0.0.1"}

	if len(chain) != len(expectedIPs) {
		t.Errorf("Expected %d IPs in chain, got %d", len(expectedIPs), len(chain))
	}

	// Check that all expected IPs are present (order may vary due to header processing)
	found := make(map[string]bool)
	for _, ip := range chain {
		found[ip] = true
	}

	for _, expectedIP := range expectedIPs {
		if !found[expectedIP] {
			t.Errorf("Expected IP %s not found in chain %v", expectedIP, chain)
		}
	}
}

func TestParseIPOptimized(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isPublic bool
		isValid  bool
	}{
		{"203.0.113.195", "203.0.113.195", true, true},
		{"192.168.1.1", "192.168.1.1", false, true},
		{"127.0.0.1", "127.0.0.1", false, true},
		{"2001:db8::1", "2001:db8::1", true, true},
		{"::1", "::1", false, true},
		{"", "", false, false},
		{"   ", "", false, false},
		{"invalid", "", false, false},
		{" 203.0.113.195 ", "203.0.113.195", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseIPOptimized(tt.input)

			if !tt.isValid {
				if result != nil && result.IP == nil {
					// This is acceptable - invalid IP with original string
					return
				}
				if result == nil {
					// This is also acceptable for empty strings
					return
				}
				if result.IP != nil {
					t.Errorf("Expected invalid result for input %q, got valid IP %v", tt.input, result.IP)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected valid result for input %q, got nil", tt.input)
				return
			}

			if result.IP == nil {
				t.Errorf("Expected valid IP for input %q, got nil IP", tt.input)
				return
			}

			if result.IP.String() != tt.expected {
				t.Errorf("Expected IP %s for input %q, got %s", tt.expected, tt.input, result.IP.String())
			}

			if result.IsPublic != tt.isPublic {
				t.Errorf("Expected IsPublic %v for input %q, got %v", tt.isPublic, tt.input, result.IsPublic)
			}
		})
	}
}

func TestIPCache(t *testing.T) {
	// Clear cache before test
	ClearCache()

	ip := "203.0.113.195"

	// First parse should cache the result
	result1 := parseIPOptimized(ip)
	if result1 == nil {
		t.Fatal("Expected non-nil result")
	}

	// Second parse should use cached result
	result2 := parseIPOptimized(ip)
	if result2 == nil {
		t.Fatal("Expected non-nil result from cache")
	}

	// Results should be equivalent but different instances
	if result1.IP.String() != result2.IP.String() {
		t.Errorf("Expected same IP from cache: %s != %s", result1.IP.String(), result2.IP.String())
	}

	// Check cache stats
	size, maxSize := GetCacheStats()
	if size == 0 {
		t.Error("Expected cache to contain entries")
	}
	if maxSize <= 0 {
		t.Error("Expected positive max cache size")
	}

	// Test cache size setting
	SetCacheSize(500)
	_, newMaxSize := GetCacheStats()
	if newMaxSize != 500 {
		t.Errorf("Expected max cache size 500, got %d", newMaxSize)
	}
}

func TestStringOperationsOptimized(t *testing.T) {
	tests := []struct {
		name   string
		header string
		value  string
		expect []string
	}{
		{
			name:   "Comma separated IPs",
			header: "X-Forwarded-For",
			value:  "203.0.113.195, 192.168.1.1, 10.0.0.1",
			expect: []string{"203.0.113.195", "192.168.1.1", "10.0.0.1"},
		},
		{
			name:   "IPs with ports",
			header: "X-Forwarded-For",
			value:  "203.0.113.195:8080, 192.168.1.1:3128",
			expect: []string{"203.0.113.195", "192.168.1.1"},
		},
		{
			name:   "Forwarded header",
			header: "Forwarded",
			value:  `for="203.0.113.195:8080", for="192.168.1.1:3128"`,
			expect: []string{"203.0.113.195", "192.168.1.1"},
		},
		{
			name:   "IPv6 with brackets",
			header: "X-Forwarded-For",
			value:  "[2001:db8::1]:8080, [::1]:8080",
			expect: []string{"2001:db8::1", "::1"},
		},
		{
			name:   "Empty value",
			header: "X-Forwarded-For",
			value:  "",
			expect: nil,
		},
		{
			name:   "Single IP",
			header: "X-Real-IP",
			value:  "203.0.113.195",
			expect: []string{"203.0.113.195"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIPsFromHeaderOptimized(tt.header, tt.value)

			if len(result) != len(tt.expect) {
				t.Errorf("Expected %d IPs, got %d: %v", len(tt.expect), len(result), result)
				return
			}

			for i, expected := range tt.expect {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected IP[%d] = %s, got %s", i, expected, result[i])
				}
			}

			// Clean up
			returnStringSlice(result)
		})
	}
}

func TestRemoteAddrOptimized(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"203.0.113.195:8080", "203.0.113.195"},
		{"192.168.1.1", "192.168.1.1"},
		{"[2001:db8::1]:8080", "2001:db8::1"},
		{"[::1]:8080", "::1"},
		{"[2001:db8::1]", "2001:db8::1"},
		{"2001:db8::1", "2001:db8::1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getIPFromRemoteAddrOptimized(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMemoryPools(t *testing.T) {
	// Test string slice pool
	slice1 := getStringSlice()
	if slice1 == nil {
		t.Error("Expected non-nil slice from pool")
	}

	slice1 = append(slice1, "test1", "test2")
	returnStringSlice(slice1)

	slice2 := getStringSlice()
	if len(slice2) != 0 {
		t.Error("Expected empty slice from pool after return")
	}

	// Test IPInfo pool
	info1 := getIPInfo()
	if info1 == nil {
		t.Error("Expected non-nil IPInfo from pool")
	}

	info1.Original = "test"
	returnIPInfo(info1)

	info2 := getIPInfo()
	if info2.Original != "" {
		t.Error("Expected reset IPInfo from pool after return")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test trimSpaceOptimized
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"", ""},
		{"   ", ""},
		{"\t\n test \r\n", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimSpaceOptimized(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}

	// Test indexCaseInsensitive
	if indexCaseInsensitive("Hello World", "WORLD") != 6 {
		t.Error("Expected case insensitive index to work")
	}

	if indexCaseInsensitive("test", "notfound") != -1 {
		t.Error("Expected -1 for not found substring")
	}

	// Test byte functions
	if indexByte("hello", 'l') != 2 {
		t.Error("Expected indexByte to find first occurrence")
	}

	if lastIndexByte("hello", 'l') != 3 {
		t.Error("Expected lastIndexByte to find last occurrence")
	}

	if countByte("hello", 'l') != 2 {
		t.Error("Expected countByte to count all occurrences")
	}
}

func TestNilSafety(t *testing.T) {
	extractor := NewOptimizedExtractor()

	// Test with nil adapter
	if result := extractor.GetRealIPOptimized(nil); result != "" {
		t.Error("Expected empty string for nil adapter")
	}

	if result := extractor.GetRealIPInfoOptimized(nil); result != nil {
		t.Error("Expected nil result for nil adapter")
	}

	if result := extractor.GetIPChainOptimized(nil); result != nil {
		t.Error("Expected nil result for nil adapter")
	}
}
