// Package ip optimization provides zero-allocation optimizations for IP processing.
// This file contains performance optimizations including buffer pools, caching,
// and string manipulation optimizations to eliminate unnecessary allocations.
package ip

import (
	"net"
	"strings"
	"sync"
	"unsafe"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Buffer pool for string processing to avoid allocations
var (
	stringBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256) // Pre-allocate 256 bytes
		},
	}

	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16) // Pre-allocate for 16 IPs
		},
	}

	ipInfoPool = sync.Pool{
		New: func() interface{} {
			return &interfaces.IPInfo{}
		},
	}
)

// Cache for parsed IP results to avoid repeated parsing
type ipCache struct {
	mu      sync.RWMutex
	entries map[string]*interfaces.IPInfo
	maxSize int
}

var globalIPCache = &ipCache{
	entries: make(map[string]*interfaces.IPInfo),
	maxSize: 1000, // Cache up to 1000 entries
}

// getCachedIP retrieves a cached IP result
func (c *ipCache) get(ipStr string) *interfaces.IPInfo {
	c.mu.RLock()
	result := c.entries[ipStr]
	c.mu.RUnlock()
	return result
}

// setCachedIP stores an IP result in cache
func (c *ipCache) set(ipStr string, info *interfaces.IPInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: clear cache when it gets too large
	if len(c.entries) >= c.maxSize {
		// Clear half the cache randomly
		count := 0
		for k := range c.entries {
			delete(c.entries, k)
			count++
			if count >= c.maxSize/2 {
				break
			}
		}
	}

	c.entries[ipStr] = info
}

// OptimizedExtractor is a zero-allocation optimized version of the IP extractor
type OptimizedExtractor struct {
	*Extractor
}

// NewOptimizedExtractor creates a new optimized IP extractor
func NewOptimizedExtractor() *OptimizedExtractor {
	return &OptimizedExtractor{
		Extractor: NewExtractor(),
	}
}

// GetRealIPOptimized extracts the real client IP with zero-allocation optimizations
func (e *OptimizedExtractor) GetRealIPOptimized(adapter interfaces.RequestAdapter) string {
	ipInfo := e.GetRealIPInfoOptimized(adapter)
	if ipInfo != nil && ipInfo.IP != nil {
		return ipInfo.IP.String()
	}
	return ""
}

// GetRealIPInfoOptimized extracts detailed IP information with optimizations
func (e *OptimizedExtractor) GetRealIPInfoOptimized(adapter interfaces.RequestAdapter) *interfaces.IPInfo {
	if adapter == nil {
		return nil
	}

	// Try each header in order of preference
	for i := range realIPHeaders {
		header := realIPHeaders[i]
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeaderOptimized(header, headerValue)
		for j := range ips {
			ipStr := ips[j]
			if ipInfo := parseIPOptimized(ipStr); ipInfo != nil && ipInfo.IP != nil {
				// Prefer public IPs
				if IsPublicIP(ipInfo.IP) {
					ipInfo.Source = header
					return ipInfo
				}
			}
		}
		// Return slice to pool
		returnStringSlice(ips)
	}

	// If no public IP found in headers, try to find any valid IP in headers
	for i := range realIPHeaders {
		header := realIPHeaders[i]
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeaderOptimized(header, headerValue)
		for j := range ips {
			ipStr := ips[j]
			if ipInfo := parseIPOptimized(ipStr); ipInfo != nil && ipInfo.IP != nil {
				ipInfo.Source = header
				returnStringSlice(ips)
				return ipInfo
			}
		}
		returnStringSlice(ips)
	}

	// Fallback to remote address
	if remoteAddr := adapter.GetRemoteAddr(); remoteAddr != "" {
		if ip := getIPFromRemoteAddrOptimized(remoteAddr); ip != "" {
			if ipInfo := parseIPOptimized(ip); ipInfo != nil {
				ipInfo.Source = "RemoteAddr"
				return ipInfo
			}
		}
	}

	return nil
}

// GetIPChainOptimized extracts the complete chain of IPs with optimizations
func (e *OptimizedExtractor) GetIPChainOptimized(adapter interfaces.RequestAdapter) []string {
	if adapter == nil {
		return nil
	}

	chain := getStringSlice()
	seen := make(map[string]bool)

	// Collect IPs from all headers
	for i := range realIPHeaders {
		header := realIPHeaders[i]
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeaderOptimized(header, headerValue)
		for j := range ips {
			ipStr := ips[j]
			if ipInfo := parseIPOptimized(ipStr); ipInfo != nil && ipInfo.IP != nil {
				ipString := ipInfo.IP.String()
				if !seen[ipString] {
					chain = append(chain, ipString)
					seen[ipString] = true
				}
			}
		}
		returnStringSlice(ips)
	}

	// Add remote address if not already present
	if remoteAddr := adapter.GetRemoteAddr(); remoteAddr != "" {
		if ip := getIPFromRemoteAddrOptimized(remoteAddr); ip != "" {
			if ipInfo := parseIPOptimized(ip); ipInfo != nil && ipInfo.IP != nil {
				ipString := ipInfo.IP.String()
				if !seen[ipString] {
					chain = append(chain, ipString)
				}
			}
		}
	}

	// Don't return the slice to pool as it's returned to caller
	return chain
}

// parseIPOptimized parses an IP string with caching and optimizations
func parseIPOptimized(ipStr string) *interfaces.IPInfo {
	// Fast path for empty strings
	if len(ipStr) == 0 {
		return nil
	}

	// Trim whitespace using unsafe string conversion to avoid allocation
	ipStr = trimSpaceOptimized(ipStr)
	if len(ipStr) == 0 {
		return nil
	}

	// Check cache first
	if cached := globalIPCache.get(ipStr); cached != nil {
		// Return a copy to avoid race conditions
		result := getIPInfo()
		*result = *cached
		return result
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		result := getIPInfo()
		result.Original = ipStr
		// Don't cache invalid IPs
		return result
	}

	result := getIPInfo()
	result.IP = ip
	result.Original = ipStr
	result.IsIPv4 = ip.To4() != nil
	result.IsIPv6 = ip.To4() == nil
	result.Type = classifyIP(ip)
	result.IsPublic = result.Type == interfaces.IPTypePublic
	result.IsPrivate = result.Type == interfaces.IPTypePrivate

	// Cache the result
	cachedResult := &interfaces.IPInfo{}
	*cachedResult = *result
	globalIPCache.set(ipStr, cachedResult)

	return result
}

// getIPsFromHeaderOptimized extracts IPs from a header value with optimizations
func getIPsFromHeaderOptimized(headerName, headerValue string) []string {
	if len(headerValue) == 0 {
		return nil
	}

	// Handle RFC 7239 Forwarded header specially
	if len(headerName) == 9 && strings.EqualFold(headerName, "forwarded") {
		return parseForwardedHeaderOptimized(headerValue)
	}

	// Handle comma-separated IP lists
	return parseCommaSeparatedIPsOptimized(headerValue)
}

// parseCommaSeparatedIPsOptimized parses comma-separated IP addresses with optimizations
func parseCommaSeparatedIPsOptimized(value string) []string {
	if len(value) == 0 {
		return nil
	}

	ips := getStringSlice()

	// Use optimized splitting to avoid allocations
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ',' {
			if i > start {
				part := trimSpaceOptimized(value[start:i])
				if len(part) > 0 {
					// Remove port if present
					if ip := getIPFromRemoteAddrOptimized(part); len(ip) > 0 {
						ips = append(ips, ip)
					}
				}
			}
			start = i + 1
		}
	}

	return ips
}

// parseForwardedHeaderOptimized parses RFC 7239 Forwarded header with optimizations
func parseForwardedHeaderOptimized(value string) []string {
	if len(value) == 0 {
		return nil
	}

	ips := getStringSlice()

	// Split by comma to handle multiple forwarded entries
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ',' {
			if i > start {
				entry := trimSpaceOptimized(value[start:i])
				if len(entry) > 0 {
					if ip := extractForParameterOptimized(entry); len(ip) > 0 {
						ips = append(ips, ip)
					}
				}
			}
			start = i + 1
		}
	}

	return ips
}

// extractForParameterOptimized extracts IP from for= parameter in Forwarded header
func extractForParameterOptimized(entry string) string {
	// Look for for= parameter
	forPrefix := "for="
	start := indexCaseInsensitive(entry, forPrefix)
	if start == -1 {
		return ""
	}

	start += len(forPrefix)
	end := start

	// Find end of parameter value
	for end < len(entry) && entry[end] != ';' {
		end++
	}

	if end <= start {
		return ""
	}

	forValue := trimSpaceOptimized(entry[start:end])

	// Remove quotes if present
	if len(forValue) >= 2 && forValue[0] == '"' && forValue[len(forValue)-1] == '"' {
		forValue = forValue[1 : len(forValue)-1]
	}

	// Remove brackets for IPv6
	if len(forValue) >= 2 && forValue[0] == '[' && forValue[len(forValue)-1] == ']' {
		forValue = forValue[1 : len(forValue)-1]
	}

	// Remove port if present
	return getIPFromRemoteAddrOptimized(forValue)
}

// getIPFromRemoteAddrOptimized extracts IP from remote address with optimizations
func getIPFromRemoteAddrOptimized(remoteAddr string) string {
	if len(remoteAddr) == 0 {
		return ""
	}

	// Handle IPv6 with brackets and port: [::1]:8080
	if remoteAddr[0] == '[' {
		if idx := indexByte(remoteAddr, ']'); idx != -1 {
			if idx+1 < len(remoteAddr) && remoteAddr[idx+1] == ':' {
				return remoteAddr[1:idx]
			}
			// Just brackets without port: [::1]
			if idx == len(remoteAddr)-1 {
				return remoteAddr[1:idx]
			}
		}
	}

	// Handle IPv4 with port: 192.168.1.1:8080
	if idx := lastIndexByte(remoteAddr, ':'); idx != -1 {
		// Make sure it's not IPv6 without brackets
		if countByte(remoteAddr, ':') == 1 {
			return remoteAddr[:idx]
		}
	}

	// Return as-is if no port found
	return remoteAddr
}

// Optimized helper functions

// trimSpaceOptimized trims whitespace without allocations
func trimSpaceOptimized(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && isSpace(s[start]) {
		start++
	}

	// Trim trailing spaces
	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if character is whitespace
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// indexCaseInsensitive finds index of substring ignoring case
func indexCaseInsensitive(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32 // Convert to lowercase
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32 // Convert to lowercase
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// indexByte finds first occurrence of byte
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// lastIndexByte finds last occurrence of byte
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// countByte counts occurrences of byte
func countByte(s string, c byte) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			count++
		}
	}
	return count
}

// Pool management functions

// getStringSlice gets a string slice from pool
func getStringSlice() []string {
	slice := stringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// returnStringSlice returns a string slice to pool
func returnStringSlice(slice []string) {
	if slice != nil && cap(slice) <= 32 { // Only return reasonably sized slices
		stringSlicePool.Put(slice)
	}
}

// getIPInfo gets an IPInfo from pool
func getIPInfo() *interfaces.IPInfo {
	info := ipInfoPool.Get().(*interfaces.IPInfo)
	// Reset all fields
	*info = interfaces.IPInfo{}
	return info
}

// returnIPInfo returns an IPInfo to pool
func returnIPInfo(info *interfaces.IPInfo) {
	if info != nil {
		ipInfoPool.Put(info)
	}
}

// stringToBytes converts string to byte slice without allocation using unsafe
func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{s, len(s)}))
}

// bytesToString converts byte slice to string without allocation using unsafe
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ClearCache clears the IP parsing cache (useful for testing or memory management)
func ClearCache() {
	globalIPCache.mu.Lock()
	globalIPCache.entries = make(map[string]*interfaces.IPInfo)
	globalIPCache.mu.Unlock()
}

// GetCacheStats returns cache statistics
func GetCacheStats() (size int, maxSize int) {
	globalIPCache.mu.RLock()
	size = len(globalIPCache.entries)
	maxSize = globalIPCache.maxSize
	globalIPCache.mu.RUnlock()
	return
}

// SetCacheSize sets the maximum cache size
func SetCacheSize(size int) {
	globalIPCache.mu.Lock()
	globalIPCache.maxSize = size
	if len(globalIPCache.entries) > size {
		// Clear cache if current size exceeds new limit
		globalIPCache.entries = make(map[string]*interfaces.IPInfo)
	}
	globalIPCache.mu.Unlock()
}
