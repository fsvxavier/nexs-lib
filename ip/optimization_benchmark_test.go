package ip

import (
	"net/http"
	"testing"

	"github.com/fsvxavier/nexs-lib/ip/providers"
)

// BenchmarkGetRealIP_NetHTTP benchmarks the standard GetRealIP function with net/http
func BenchmarkGetRealIP_NetHTTP(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetRealIP(req)
	}
}

// BenchmarkGetRealIP_Optimized benchmarks the optimized GetRealIP function
func BenchmarkGetRealIP_Optimized(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	adapter, _ := providers.CreateAdapter(req)
	extractor := NewOptimizedExtractor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = extractor.GetRealIPOptimized(adapter)
	}
}

// BenchmarkGetRealIPInfo_NetHTTP benchmarks the standard GetRealIPInfo function
func BenchmarkGetRealIPInfo_NetHTTP(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetRealIPInfo(req)
	}
}

// BenchmarkGetRealIPInfo_Optimized benchmarks the optimized GetRealIPInfo function
func BenchmarkGetRealIPInfo_Optimized(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	adapter, _ := providers.CreateAdapter(req)
	extractor := NewOptimizedExtractor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = extractor.GetRealIPInfoOptimized(adapter)
	}
}

// BenchmarkGetIPChain_NetHTTP benchmarks the standard GetIPChain function
func BenchmarkGetIPChain_NetHTTP(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.Header.Set("CF-Connecting-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetIPChain(req)
	}
}

// BenchmarkGetIPChain_Optimized benchmarks the optimized GetIPChain function
func BenchmarkGetIPChain_Optimized(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.Header.Set("CF-Connecting-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	adapter, _ := providers.CreateAdapter(req)
	extractor := NewOptimizedExtractor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = extractor.GetIPChainOptimized(adapter)
	}
}

// BenchmarkParseIP_Standard benchmarks the standard ParseIP function
func BenchmarkParseIP_Standard(b *testing.B) {
	ipStr := "203.0.113.195"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ParseIP(ipStr)
	}
}

// BenchmarkParseIP_Optimized benchmarks the optimized parseIPOptimized function
func BenchmarkParseIP_Optimized(b *testing.B) {
	ipStr := "203.0.113.195"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = parseIPOptimized(ipStr)
	}
}

// BenchmarkParseIP_Cached benchmarks parsing with warm cache
func BenchmarkParseIP_Cached(b *testing.B) {
	ipStr := "203.0.113.195"

	// Warm up cache
	_ = parseIPOptimized(ipStr)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = parseIPOptimized(ipStr)
	}
}

// BenchmarkStringOperations_Standard benchmarks standard string operations
func BenchmarkStringOperations_Standard(b *testing.B) {
	header := "X-Forwarded-For"
	value := "203.0.113.195, 192.168.1.1, 10.0.0.1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getIPsFromHeaderOptimized(header, value)
	}
}

// BenchmarkStringOperations_Optimized benchmarks optimized string operations
func BenchmarkStringOperations_Optimized(b *testing.B) {
	header := "X-Forwarded-For"
	value := "203.0.113.195, 192.168.1.1, 10.0.0.1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ips := getIPsFromHeaderOptimized(header, value)
		returnStringSlice(ips)
	}
}

// BenchmarkForwardedHeader_Standard benchmarks standard Forwarded header parsing
func BenchmarkForwardedHeader_Standard(b *testing.B) {
	value := `for="203.0.113.195:8080", for="192.168.1.1:3128"`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = parseForwardedHeaderOptimized(value)
	}
}

// BenchmarkForwardedHeader_Optimized benchmarks optimized Forwarded header parsing
func BenchmarkForwardedHeader_Optimized(b *testing.B) {
	value := `for="203.0.113.195:8080", for="192.168.1.1:3128"`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ips := parseForwardedHeaderOptimized(value)
		returnStringSlice(ips)
	}
}

// BenchmarkRemoteAddr_Standard benchmarks standard remote address parsing
func BenchmarkRemoteAddr_Standard(b *testing.B) {
	remoteAddr := "203.0.113.195:8080"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getIPFromRemoteAddrOptimized(remoteAddr)
	}
}

// BenchmarkRemoteAddr_Optimized benchmarks optimized remote address parsing
func BenchmarkRemoteAddr_Optimized(b *testing.B) {
	remoteAddr := "203.0.113.195:8080"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getIPFromRemoteAddrOptimized(remoteAddr)
	}
}

// BenchmarkConcurrentAccess tests concurrent access to optimized functions
func BenchmarkConcurrentAccess(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.RemoteAddr = "192.168.1.1:8080"

	adapter, _ := providers.CreateAdapter(req)
	extractor := NewOptimizedExtractor()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = extractor.GetRealIPOptimized(adapter)
		}
	})
}

// BenchmarkMemoryPooling tests memory pool efficiency
func BenchmarkMemoryPooling(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slice := getStringSlice()
		slice = append(slice, "test1", "test2", "test3")
		returnStringSlice(slice)

		info := getIPInfo()
		returnIPInfo(info)
	}
}

// BenchmarkCacheEfficiency tests cache hit rates
func BenchmarkCacheEfficiency(b *testing.B) {
	ips := []string{
		"203.0.113.195",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
	}

	// Warm up cache
	for _, ip := range ips {
		_ = parseIPOptimized(ip)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ip := ips[i%len(ips)]
		_ = parseIPOptimized(ip)
	}
}
