// Package ip detection provides advanced IP detection capabilities including
// VPN/Proxy detection, ASN lookup, and reputation scoring.
package ip

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VPNProvider represents a known VPN service provider
type VPNProvider struct {
	Name        string
	Type        string // "commercial", "datacenter", "tor", "proxy"
	Reliability float64
}

// ASNInfo contains information about an Autonomous System Number
type ASNInfo struct {
	ASN             uint32
	Name            string
	Country         string
	Type            string // "isp", "hosting", "government", "enterprise"
	IsCloudProvider bool
	CloudProvider   string
}

// DetectionResult contains comprehensive IP detection results
type DetectionResult struct {
	IP              net.IP
	IsVPN           bool
	IsProxy         bool
	IsTor           bool
	IsDatacenter    bool
	IsCloudProvider bool
	VPNProvider     *VPNProvider
	ASNInfo         *ASNInfo
	TrustScore      float64 // 0.0 (untrusted) to 1.0 (trusted)
	RiskLevel       string  // "low", "medium", "high", "critical"
	DetectionTime   time.Duration
}

// AdvancedDetector provides advanced IP detection capabilities
type AdvancedDetector struct {
	vpnDatabase  map[string]*VPNProvider
	asnDatabase  map[uint32]*ASNInfo
	httpClient   *http.Client
	mu           sync.RWMutex
	cacheEnabled bool
	cache        map[string]*DetectionResult
	cacheTimeout time.Duration
	maxCacheSize int
	workerPool   *WorkerPool
}

// DetectorConfig contains configuration for the advanced detector
type DetectorConfig struct {
	CacheEnabled        bool
	CacheTimeout        time.Duration
	MaxCacheSize        int
	HTTPTimeout         time.Duration
	MaxWorkers          int
	VPNDatabaseURL      string
	ASNDatabaseURL      string
	EnableOnlineLookups bool
}

// DefaultDetectorConfig returns a default configuration
func DefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		CacheEnabled:        true,
		CacheTimeout:        time.Hour,
		MaxCacheSize:        10000,
		HTTPTimeout:         5 * time.Second,
		MaxWorkers:          10,
		EnableOnlineLookups: true,
	}
}

// NewAdvancedDetector creates a new advanced IP detector
func NewAdvancedDetector(config *DetectorConfig) *AdvancedDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}

	detector := &AdvancedDetector{
		vpnDatabase:  make(map[string]*VPNProvider),
		asnDatabase:  make(map[uint32]*ASNInfo),
		cacheEnabled: config.CacheEnabled,
		cache:        make(map[string]*DetectionResult),
		cacheTimeout: config.CacheTimeout,
		maxCacheSize: config.MaxCacheSize,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		workerPool: NewWorkerPool(config.MaxWorkers),
	}

	// Initialize with built-in VPN/Proxy detection data
	detector.initBuiltinData()

	return detector
}

// DetectAdvanced performs comprehensive IP detection analysis
func (d *AdvancedDetector) DetectAdvanced(ctx context.Context, ip net.IP) (*DetectionResult, error) {
	start := time.Now()

	// Check cache first
	if d.cacheEnabled {
		if cached := d.getCached(ip.String()); cached != nil {
			return cached, nil
		}
	}

	result := &DetectionResult{
		IP:            ip,
		DetectionTime: 0,
	}

	// Create context with timeout
	detectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Run detection operations concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	// VPN/Proxy detection
	wg.Add(1)
	go func() {
		defer wg.Done()
		vpnInfo := d.detectVPN(detectCtx, ip)
		mu.Lock()
		if vpnInfo != nil {
			result.IsVPN = vpnInfo.IsVPN
			result.IsProxy = vpnInfo.IsProxy
			result.IsTor = vpnInfo.IsTor
			result.VPNProvider = vpnInfo.Provider
		}
		mu.Unlock()
	}()

	// ASN lookup
	wg.Add(1)
	go func() {
		defer wg.Done()
		asnInfo := d.lookupASN(detectCtx, ip)
		mu.Lock()
		if asnInfo != nil {
			result.ASNInfo = asnInfo
			result.IsDatacenter = asnInfo.Type == "hosting"
			result.IsCloudProvider = asnInfo.IsCloudProvider
		}
		mu.Unlock()
	}()

	// Wait for all operations to complete
	wg.Wait()

	// Calculate trust score and risk level
	result.TrustScore = d.calculateTrustScore(result)
	result.RiskLevel = d.calculateRiskLevel(result)
	result.DetectionTime = time.Since(start)

	// Cache the result
	if d.cacheEnabled {
		d.setCached(ip.String(), result)
	}

	return result, nil
}

// vpnDetectionInfo contains VPN detection results
type vpnDetectionInfo struct {
	IsVPN    bool
	IsProxy  bool
	IsTor    bool
	Provider *VPNProvider
}

// detectVPN performs VPN/Proxy detection
func (d *AdvancedDetector) detectVPN(ctx context.Context, ip net.IP) *vpnDetectionInfo {
	info := &vpnDetectionInfo{}

	// Check built-in VPN database
	d.mu.RLock()
	if provider, exists := d.vpnDatabase[ip.String()]; exists {
		info.IsVPN = true
		info.Provider = provider
		if provider.Type == "tor" {
			info.IsTor = true
		} else if provider.Type == "proxy" {
			info.IsProxy = true
		}
		d.mu.RUnlock()
		return info
	}
	d.mu.RUnlock()

	// Check for known datacenter ranges (heuristic detection)
	if d.isDatacenterIP(ip) {
		info.IsProxy = true
		info.Provider = &VPNProvider{
			Name:        "Unknown Datacenter",
			Type:        "datacenter",
			Reliability: 0.7,
		}
	}

	// Check for suspicious patterns
	if d.hasSuspiciousPatterns(ip) {
		info.IsVPN = true
		info.Provider = &VPNProvider{
			Name:        "Suspicious Pattern",
			Type:        "proxy",
			Reliability: 0.6,
		}
	}

	return info
}

// lookupASN performs ASN lookup for the IP
func (d *AdvancedDetector) lookupASN(ctx context.Context, ip net.IP) *ASNInfo {
	// Check built-in ASN database first
	d.mu.RLock()
	for asn, info := range d.asnDatabase {
		// This is a simplified check - in production, you'd use proper CIDR matching
		if d.ipBelongsToASN(ip, asn) {
			d.mu.RUnlock()
			return info
		}
	}
	d.mu.RUnlock()

	// Fallback to heuristic detection
	return d.detectASNHeuristic(ip)
}

// calculateTrustScore calculates a trust score for the IP
func (d *AdvancedDetector) calculateTrustScore(result *DetectionResult) float64 {
	score := 1.0

	// Reduce score for VPN/Proxy
	if result.IsVPN {
		score -= 0.4
	}
	if result.IsProxy {
		score -= 0.3
	}
	if result.IsTor {
		score -= 0.6
	}

	// Reduce score for datacenter/cloud
	if result.IsDatacenter {
		score -= 0.2
	}
	if result.IsCloudProvider {
		score -= 0.1
	}

	// Apply VPN provider reliability
	if result.VPNProvider != nil {
		score *= result.VPNProvider.Reliability
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// calculateRiskLevel determines the risk level based on detection results
func (d *AdvancedDetector) calculateRiskLevel(result *DetectionResult) string {
	if result.IsTor {
		return "critical"
	}
	if result.IsVPN && result.TrustScore < 0.3 {
		return "high"
	}
	if result.IsProxy || (result.IsDatacenter && result.TrustScore < 0.5) {
		return "medium"
	}
	return "low"
}

// getCached retrieves a cached detection result
func (d *AdvancedDetector) getCached(ip string) *DetectionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if result, exists := d.cache[ip]; exists {
		// Check if cache entry is still valid
		if time.Since(time.Now().Add(-result.DetectionTime)) < d.cacheTimeout {
			return result
		}
		// Remove expired entry
		delete(d.cache, ip)
	}
	return nil
}

// setCached stores a detection result in cache
func (d *AdvancedDetector) setCached(ip string, result *DetectionResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clean cache if it's too large
	if len(d.cache) >= d.maxCacheSize {
		// Remove oldest entries (simple FIFO)
		count := 0
		for k := range d.cache {
			delete(d.cache, k)
			count++
			if count >= d.maxCacheSize/4 { // Remove 25% of entries
				break
			}
		}
	}

	d.cache[ip] = result
}

// initBuiltinData initializes built-in VPN and ASN databases
func (d *AdvancedDetector) initBuiltinData() {
	// Initialize with some common VPN/Proxy ranges and providers
	// This would typically be loaded from external databases

	// Common cloud providers
	d.asnDatabase[16509] = &ASNInfo{
		ASN:             16509,
		Name:            "Amazon Web Services",
		Country:         "US",
		Type:            "hosting",
		IsCloudProvider: true,
		CloudProvider:   "AWS",
	}

	d.asnDatabase[15169] = &ASNInfo{
		ASN:             15169,
		Name:            "Google LLC",
		Country:         "US",
		Type:            "hosting",
		IsCloudProvider: true,
		CloudProvider:   "Google Cloud",
	}

	d.asnDatabase[8075] = &ASNInfo{
		ASN:             8075,
		Name:            "Microsoft Corporation",
		Country:         "US",
		Type:            "hosting",
		IsCloudProvider: true,
		CloudProvider:   "Azure",
	}
}

// isDatacenterIP checks if an IP belongs to a known datacenter range
func (d *AdvancedDetector) isDatacenterIP(ip net.IP) bool {
	// Common datacenter IP patterns (simplified heuristic)
	ipStr := ip.String()

	// Check for common cloud provider ranges
	cloudRanges := []string{
		"52.", "54.", "3.", "18.", "34.", "35.", // AWS ranges (simplified)
		"104.", "130.", "146.", "162.", // Google Cloud ranges
		"13.", "40.", "52.", "104.", // Azure ranges
	}

	for _, prefix := range cloudRanges {
		if strings.HasPrefix(ipStr, prefix) {
			return true
		}
	}

	return false
}

// hasSuspiciousPatterns checks for suspicious IP patterns
func (d *AdvancedDetector) hasSuspiciousPatterns(ip net.IP) bool {
	// Check for sequential IPs, common VPN patterns, etc.
	// This is a simplified implementation

	if ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}

	// Check for common VPN/Proxy port patterns in IP
	ipStr := ip.String()

	// Some basic patterns that might indicate VPN/Proxy
	suspiciousPatterns := []string{
		".1.1", ".8.8", ".9.9", // Common DNS IPs sometimes used by VPNs
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(ipStr, pattern) {
			return true
		}
	}

	return false
}

// ipBelongsToASN checks if an IP belongs to a specific ASN
func (d *AdvancedDetector) ipBelongsToASN(ip net.IP, asn uint32) bool {
	// This is a simplified implementation
	// In production, you'd use proper CIDR matching with ASN databases
	return false
}

// detectASNHeuristic performs heuristic ASN detection
func (d *AdvancedDetector) detectASNHeuristic(ip net.IP) *ASNInfo {
	// Simplified heuristic detection based on IP patterns
	if d.isDatacenterIP(ip) {
		return &ASNInfo{
			ASN:             0,
			Name:            "Unknown Hosting Provider",
			Country:         "Unknown",
			Type:            "hosting",
			IsCloudProvider: true,
			CloudProvider:   "Unknown",
		}
	}

	return &ASNInfo{
		ASN:             0,
		Name:            "Unknown ISP",
		Country:         "Unknown",
		Type:            "isp",
		IsCloudProvider: false,
	}
}

// LoadVPNDatabase loads VPN database from CSV data
func (d *AdvancedDetector) LoadVPNDatabase(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read VPN database: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(record) < 4 {
			continue
		}

		reliability, _ := strconv.ParseFloat(record[3], 64)
		provider := &VPNProvider{
			Name:        record[1],
			Type:        record[2],
			Reliability: reliability,
		}

		d.vpnDatabase[record[0]] = provider
	}

	return nil
}

// LoadASNDatabase loads ASN database from CSV data
func (d *AdvancedDetector) LoadASNDatabase(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read ASN database: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(record) < 6 {
			continue
		}

		asn, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			continue
		}

		isCloudProvider, _ := strconv.ParseBool(record[4])

		info := &ASNInfo{
			ASN:             uint32(asn),
			Name:            record[1],
			Country:         record[2],
			Type:            record[3],
			IsCloudProvider: isCloudProvider,
			CloudProvider:   record[5],
		}

		d.asnDatabase[uint32(asn)] = info
	}

	return nil
}

// GetCacheStats returns cache statistics
func (d *AdvancedDetector) GetCacheStats() (size int, hitRate float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.cache), 0.0 // Hit rate calculation would require tracking hits/misses
}

// ClearCache clears the detection cache
func (d *AdvancedDetector) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache = make(map[string]*DetectionResult)
}

// Close closes the detector and cleans up resources
func (d *AdvancedDetector) Close() error {
	if d.workerPool != nil {
		d.workerPool.Close()
	}
	d.ClearCache()
	return nil
}
