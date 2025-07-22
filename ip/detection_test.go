package ip

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestAdvancedDetector_DetectAdvanced(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected DetectionResult
		wantErr  bool
	}{
		{
			name: "valid_public_ip",
			ip:   "8.8.8.8",
			expected: DetectionResult{
				IsVPN:        false,
				IsProxy:      false,
				IsTor:        false,
				IsDatacenter: false,
				RiskLevel:    "low",
			},
			wantErr: false,
		},
		{
			name: "datacenter_ip",
			ip:   "52.86.85.143", // AWS IP range
			expected: DetectionResult{
				IsDatacenter:    true,
				IsCloudProvider: true,
				RiskLevel:       "medium",
			},
			wantErr: false,
		},
		{
			name: "private_ip",
			ip:   "192.168.1.1",
			expected: DetectionResult{
				IsVPN:     false,
				IsProxy:   false,
				RiskLevel: "low",
			},
			wantErr: false,
		},
	}

	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("invalid IP address: %s", tt.ip)
			}

			result, err := detector.DetectAdvanced(ctx, ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectAdvanced() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("DetectAdvanced() returned nil result")
				return
			}

			// Verify IP is set correctly
			if !result.IP.Equal(ip) {
				t.Errorf("IP mismatch: got %v, want %v", result.IP, ip)
			}

			// Verify trust score is within valid range
			if result.TrustScore < 0 || result.TrustScore > 1 {
				t.Errorf("TrustScore out of range: got %f, want 0.0-1.0", result.TrustScore)
			}

			// Verify risk level is valid
			validRiskLevels := []string{"low", "medium", "high", "critical"}
			found := false
			for _, level := range validRiskLevels {
				if result.RiskLevel == level {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Invalid risk level: %s", result.RiskLevel)
			}

			// Verify detection time is recorded
			if result.DetectionTime <= 0 {
				t.Error("DetectionTime should be greater than 0")
			}
		})
	}
}

func TestAdvancedDetector_Cache(t *testing.T) {
	config := DefaultDetectorConfig()
	config.CacheEnabled = true
	config.CacheTimeout = time.Hour

	detector := NewAdvancedDetector(config)
	defer detector.Close()

	ctx := context.Background()
	ip := net.ParseIP("8.8.8.8")

	// First call should perform detection
	start1 := time.Now()
	result1, err := detector.DetectAdvanced(ctx, ip)
	duration1 := time.Since(start1)
	if err != nil {
		t.Fatalf("First detection failed: %v", err)
	}

	// Second call should use cache and be faster
	start2 := time.Now()
	result2, err := detector.DetectAdvanced(ctx, ip)
	duration2 := time.Since(start2)
	if err != nil {
		t.Fatalf("Second detection failed: %v", err)
	}

	// Cache should make second call faster
	if duration2 >= duration1 {
		t.Log("Warning: Cached call was not significantly faster")
	}

	// Results should be identical
	if !result1.IP.Equal(result2.IP) {
		t.Error("Cached result IP mismatch")
	}
	if result1.TrustScore != result2.TrustScore {
		t.Error("Cached result TrustScore mismatch")
	}
	if result1.RiskLevel != result2.RiskLevel {
		t.Error("Cached result RiskLevel mismatch")
	}

	// Check cache stats
	size, _ := detector.GetCacheStats()
	if size == 0 {
		t.Error("Cache should contain at least one entry")
	}

	// Clear cache and verify
	detector.ClearCache()
	size, _ = detector.GetCacheStats()
	if size != 0 {
		t.Error("Cache should be empty after clear")
	}
}

func TestAdvancedDetector_ConcurrentAccess(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	ctx := context.Background()
	ips := []string{
		"8.8.8.8",
		"1.1.1.1",
		"208.67.222.222",
		"52.86.85.143",
		"192.168.1.1",
	}

	// Run multiple goroutines concurrently
	done := make(chan bool, len(ips))
	for _, ipStr := range ips {
		go func(ip string) {
			defer func() { done <- true }()

			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				t.Errorf("Invalid IP: %s", ip)
				return
			}

			result, err := detector.DetectAdvanced(ctx, parsedIP)
			if err != nil {
				t.Errorf("Detection failed for %s: %v", ip, err)
				return
			}
			if result == nil {
				t.Errorf("Nil result for %s", ip)
				return
			}
		}(ipStr)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(ips); i++ {
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent detection")
		}
	}
}

func TestAdvancedDetector_LoadVPNDatabase(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	// Create test VPN database CSV data
	csvData := `ip,name,type,reliability
1.2.3.4,TestVPN,commercial,0.8
5.6.7.8,TestProxy,proxy,0.6
9.10.11.12,TestTor,tor,0.9`

	reader := strings.NewReader(csvData)
	err := detector.LoadVPNDatabase(reader)
	if err != nil {
		t.Fatalf("Failed to load VPN database: %v", err)
	}

	// Test detection with loaded data
	ctx := context.Background()
	ip := net.ParseIP("1.2.3.4")
	result, err := detector.DetectAdvanced(ctx, ip)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if !result.IsVPN {
		t.Error("IP should be detected as VPN")
	}
	if result.VPNProvider == nil {
		t.Error("VPN provider should be set")
	}
	if result.VPNProvider != nil && result.VPNProvider.Name != "TestVPN" {
		t.Errorf("Expected VPN provider name 'TestVPN', got '%s'", result.VPNProvider.Name)
	}
}

func TestAdvancedDetector_LoadASNDatabase(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	// Create test ASN database CSV data
	csvData := `asn,name,country,type,is_cloud_provider,cloud_provider
16509,Amazon Web Services,US,hosting,true,AWS
15169,Google LLC,US,hosting,true,Google Cloud
1234,Test ISP,US,isp,false,`

	reader := strings.NewReader(csvData)
	err := detector.LoadASNDatabase(reader)
	if err != nil {
		t.Fatalf("Failed to load ASN database: %v", err)
	}

	// Verify database was loaded
	detector.mu.RLock()
	asnInfo := detector.asnDatabase[16509]
	detector.mu.RUnlock()

	if asnInfo == nil {
		t.Error("ASN 16509 should be loaded")
	}
	if asnInfo != nil && asnInfo.Name != "Amazon Web Services" {
		t.Errorf("Expected ASN name 'Amazon Web Services', got '%s'", asnInfo.Name)
	}
}

func TestAdvancedDetector_TrustScoreCalculation(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	tests := []struct {
		name     string
		result   DetectionResult
		expected float64
	}{
		{
			name: "clean_ip",
			result: DetectionResult{
				IsVPN:           false,
				IsProxy:         false,
				IsTor:           false,
				IsDatacenter:    false,
				IsCloudProvider: false,
			},
			expected: 1.0,
		},
		{
			name: "vpn_ip",
			result: DetectionResult{
				IsVPN:   true,
				IsProxy: false,
				IsTor:   false,
			},
			expected: 0.6, // 1.0 - 0.4
		},
		{
			name: "tor_ip",
			result: DetectionResult{
				IsVPN:   false,
				IsProxy: false,
				IsTor:   true,
			},
			expected: 0.4, // 1.0 - 0.6
		},
		{
			name: "datacenter_ip",
			result: DetectionResult{
				IsVPN:        false,
				IsProxy:      false,
				IsTor:        false,
				IsDatacenter: true,
			},
			expected: 0.8, // 1.0 - 0.2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateTrustScore(&tt.result)
			if score != tt.expected {
				t.Errorf("Trust score mismatch: got %f, want %f", score, tt.expected)
			}
		})
	}
}

func TestAdvancedDetector_RiskLevelCalculation(t *testing.T) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	tests := []struct {
		name     string
		result   DetectionResult
		expected string
	}{
		{
			name: "tor_critical",
			result: DetectionResult{
				IsTor: true,
			},
			expected: "critical",
		},
		{
			name: "low_trust_vpn_high",
			result: DetectionResult{
				IsVPN:      true,
				TrustScore: 0.2,
			},
			expected: "high",
		},
		{
			name: "proxy_medium",
			result: DetectionResult{
				IsProxy:    true,
				TrustScore: 0.7,
			},
			expected: "medium",
		},
		{
			name: "clean_low",
			result: DetectionResult{
				IsVPN:      false,
				IsProxy:    false,
				IsTor:      false,
				TrustScore: 0.9,
			},
			expected: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := detector.calculateRiskLevel(&tt.result)
			if level != tt.expected {
				t.Errorf("Risk level mismatch: got %s, want %s", level, tt.expected)
			}
		})
	}
}

func BenchmarkAdvancedDetector_DetectAdvanced(b *testing.B) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	ctx := context.Background()
	ip := net.ParseIP("8.8.8.8")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := detector.DetectAdvanced(ctx, ip)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}
	}
}

func BenchmarkAdvancedDetector_DetectAdvanced_Cached(b *testing.B) {
	config := DefaultDetectorConfig()
	config.CacheEnabled = true
	detector := NewAdvancedDetector(config)
	defer detector.Close()

	ctx := context.Background()
	ip := net.ParseIP("8.8.8.8")

	// Prime the cache
	_, err := detector.DetectAdvanced(ctx, ip)
	if err != nil {
		b.Fatalf("Initial detection failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := detector.DetectAdvanced(ctx, ip)
		if err != nil {
			b.Fatalf("Detection failed: %v", err)
		}
	}
}

func BenchmarkAdvancedDetector_ConcurrentDetection(b *testing.B) {
	detector := NewAdvancedDetector(DefaultDetectorConfig())
	defer detector.Close()

	ctx := context.Background()
	ips := []net.IP{
		net.ParseIP("8.8.8.8"),
		net.ParseIP("1.1.1.1"),
		net.ParseIP("208.67.222.222"),
		net.ParseIP("52.86.85.143"),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ip := ips[i%len(ips)]
			_, err := detector.DetectAdvanced(ctx, ip)
			if err != nil {
				b.Fatalf("Detection failed: %v", err)
			}
			i++
		}
	})
}
