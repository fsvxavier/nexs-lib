// Package ip memory provides advanced memory optimization for IP processing.
// This file implements object pooling, lazy loading, and garbage collection tuning
// to minimize memory footprint and improve performance.
package ip

import (
	"net"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// ObjectPools manages various object pools for memory optimization
type ObjectPools struct {
	detectionResultPool sync.Pool
	asnInfoPool         sync.Pool
	vpnProviderPool     sync.Pool
	ipSlicePool         sync.Pool
	stringSlicePool     sync.Pool
	byteSlicePool       sync.Pool
}

// Global object pools instance
var globalPools = &ObjectPools{}

// init initializes the global object pools
func init() {
	globalPools.initPools()
}

// initPools initializes all object pools with appropriate constructors
func (p *ObjectPools) initPools() {
	// DetectionResult pool
	p.detectionResultPool = sync.Pool{
		New: func() interface{} {
			return &DetectionResult{}
		},
	}

	// ASNInfo pool
	p.asnInfoPool = sync.Pool{
		New: func() interface{} {
			return &ASNInfo{}
		},
	}

	// VPNProvider pool
	p.vpnProviderPool = sync.Pool{
		New: func() interface{} {
			return &VPNProvider{}
		},
	}

	// IP slice pool
	p.ipSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]net.IP, 0, 16) // Pre-allocate for 16 IPs
		},
	}

	// String slice pool
	p.stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 32) // Pre-allocate for 32 strings
		},
	}

	// Byte slice pool for string processing
	p.byteSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024) // Pre-allocate 1KB buffer
		},
	}
}

// GetDetectionResult returns a pooled DetectionResult
func (p *ObjectPools) GetDetectionResult() *DetectionResult {
	result := p.detectionResultPool.Get().(*DetectionResult)
	result.reset()
	return result
}

// PutDetectionResult returns a DetectionResult to the pool
func (p *ObjectPools) PutDetectionResult(result *DetectionResult) {
	if result != nil {
		result.reset()
		p.detectionResultPool.Put(result)
	}
}

// GetASNInfo returns a pooled ASNInfo
func (p *ObjectPools) GetASNInfo() *ASNInfo {
	info := p.asnInfoPool.Get().(*ASNInfo)
	info.reset()
	return info
}

// PutASNInfo returns an ASNInfo to the pool
func (p *ObjectPools) PutASNInfo(info *ASNInfo) {
	if info != nil {
		info.reset()
		p.asnInfoPool.Put(info)
	}
}

// GetVPNProvider returns a pooled VPNProvider
func (p *ObjectPools) GetVPNProvider() *VPNProvider {
	provider := p.vpnProviderPool.Get().(*VPNProvider)
	provider.reset()
	return provider
}

// PutVPNProvider returns a VPNProvider to the pool
func (p *ObjectPools) PutVPNProvider(provider *VPNProvider) {
	if provider != nil {
		provider.reset()
		p.vpnProviderPool.Put(provider)
	}
}

// GetIPSlice returns a pooled IP slice
func (p *ObjectPools) GetIPSlice() []net.IP {
	slice := p.ipSlicePool.Get().([]net.IP)
	return slice[:0] // Reset length but keep capacity
}

// PutIPSlice returns an IP slice to the pool
func (p *ObjectPools) PutIPSlice(slice []net.IP) {
	if slice != nil && cap(slice) > 0 {
		p.ipSlicePool.Put(slice)
	}
}

// GetStringSlice returns a pooled string slice
func (p *ObjectPools) GetStringSlice() []string {
	slice := p.stringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// PutStringSlice returns a string slice to the pool
func (p *ObjectPools) PutStringSlice(slice []string) {
	if slice != nil && cap(slice) > 0 {
		p.stringSlicePool.Put(slice)
	}
}

// GetByteSlice returns a pooled byte slice
func (p *ObjectPools) GetByteSlice() []byte {
	slice := p.byteSlicePool.Get().([]byte)
	return slice[:0] // Reset length but keep capacity
}

// PutByteSlice returns a byte slice to the pool
func (p *ObjectPools) PutByteSlice(slice []byte) {
	if slice != nil && cap(slice) > 0 {
		p.byteSlicePool.Put(slice)
	}
}

// reset methods for pooled objects

// reset clears all fields of DetectionResult
func (d *DetectionResult) reset() {
	d.IP = nil
	d.IsVPN = false
	d.IsProxy = false
	d.IsTor = false
	d.IsDatacenter = false
	d.IsCloudProvider = false
	d.VPNProvider = nil
	d.ASNInfo = nil
	d.TrustScore = 0
	d.RiskLevel = ""
	d.DetectionTime = 0
}

// reset clears all fields of ASNInfo
func (a *ASNInfo) reset() {
	a.ASN = 0
	a.Name = ""
	a.Country = ""
	a.Type = ""
	a.IsCloudProvider = false
	a.CloudProvider = ""
}

// reset clears all fields of VPNProvider
func (v *VPNProvider) reset() {
	v.Name = ""
	v.Type = ""
	v.Reliability = 0
}

// LazyDatabase provides lazy loading for large databases
type LazyDatabase struct {
	mu           sync.RWMutex
	loaded       bool
	loadFunc     func() error
	data         interface{}
	lastAccessed time.Time
	ttl          time.Duration
}

// NewLazyDatabase creates a new lazy-loaded database
func NewLazyDatabase(loadFunc func() error, ttl time.Duration) *LazyDatabase {
	return &LazyDatabase{
		loadFunc: loadFunc,
		ttl:      ttl,
	}
}

// Get returns the database data, loading it if necessary
func (db *LazyDatabase) Get() (interface{}, error) {
	db.mu.RLock()
	if db.loaded && (db.ttl == 0 || time.Since(db.lastAccessed) < db.ttl) {
		db.lastAccessed = time.Now()
		data := db.data
		db.mu.RUnlock()
		return data, nil
	}
	db.mu.RUnlock()

	db.mu.Lock()
	defer db.mu.Unlock()

	// Double-check pattern
	if db.loaded && (db.ttl == 0 || time.Since(db.lastAccessed) < db.ttl) {
		db.lastAccessed = time.Now()
		return db.data, nil
	}

	if err := db.loadFunc(); err != nil {
		return nil, err
	}

	db.loaded = true
	db.lastAccessed = time.Now()
	return db.data, nil
}

// Set sets the database data
func (db *LazyDatabase) Set(data interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data = data
	db.loaded = true
	db.lastAccessed = time.Now()
}

// Unload unloads the database data to free memory
func (db *LazyDatabase) Unload() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data = nil
	db.loaded = false
}

// IsLoaded returns whether the database is currently loaded
func (db *LazyDatabase) IsLoaded() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.loaded
}

// MemoryManager provides garbage collection tuning and memory monitoring
type MemoryManager struct {
	gcPercent        int
	maxMemoryMB      int64
	checkInterval    time.Duration
	lastGC           time.Time
	ticker           *time.Ticker
	stopChan         chan bool
	forceGCThreshold float64
}

// MemoryConfig contains configuration for memory management
type MemoryConfig struct {
	GCPercent        int           // GOGC percentage (default: 100)
	MaxMemoryMB      int64         // Maximum memory usage in MB (0 = no limit)
	CheckInterval    time.Duration // Memory check interval
	ForceGCThreshold float64       // Force GC when memory usage exceeds this ratio (0.0-1.0)
}

// DefaultMemoryConfig returns a default memory configuration
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		GCPercent:        100,
		MaxMemoryMB:      0,
		CheckInterval:    30 * time.Second,
		ForceGCThreshold: 0.8,
	}
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(config *MemoryConfig) *MemoryManager {
	if config == nil {
		config = DefaultMemoryConfig()
	}

	mm := &MemoryManager{
		gcPercent:        config.GCPercent,
		maxMemoryMB:      config.MaxMemoryMB,
		checkInterval:    config.CheckInterval,
		forceGCThreshold: config.ForceGCThreshold,
		stopChan:         make(chan bool),
	}

	// Set initial GC percent
	debug.SetGCPercent(mm.gcPercent)

	// Start memory monitoring if enabled
	if mm.checkInterval > 0 {
		mm.startMonitoring()
	}

	return mm
}

// startMonitoring starts the memory monitoring goroutine
func (mm *MemoryManager) startMonitoring() {
	mm.ticker = time.NewTicker(mm.checkInterval)

	go func() {
		for {
			select {
			case <-mm.ticker.C:
				mm.checkMemory()
			case <-mm.stopChan:
				return
			}
		}
	}()
}

// checkMemory checks current memory usage and triggers GC if necessary
func (mm *MemoryManager) checkMemory() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	allocMB := int64(stats.Alloc) / 1024 / 1024

	// Force GC if memory usage is too high
	if mm.maxMemoryMB > 0 && allocMB > mm.maxMemoryMB {
		mm.forceGC()
		return
	}

	// Force GC if threshold is exceeded
	if mm.forceGCThreshold > 0 && mm.maxMemoryMB > 0 {
		ratio := float64(allocMB) / float64(mm.maxMemoryMB)
		if ratio > mm.forceGCThreshold {
			mm.forceGC()
		}
	}
}

// forceGC forces garbage collection
func (mm *MemoryManager) forceGC() {
	now := time.Now()
	if now.Sub(mm.lastGC) > time.Second { // Avoid too frequent GC calls
		runtime.GC()
		mm.lastGC = now
	}
}

// GetMemoryStats returns current memory statistics
func (mm *MemoryManager) GetMemoryStats() MemoryStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return MemoryStats{
		AllocMB:      int64(stats.Alloc) / 1024 / 1024,
		TotalAllocMB: int64(stats.TotalAlloc) / 1024 / 1024,
		SysMB:        int64(stats.Sys) / 1024 / 1024,
		NumGC:        stats.NumGC,
		PauseTotalNs: stats.PauseTotalNs,
		LastGC:       time.Unix(0, int64(stats.LastGC)),
	}
}

// MemoryStats contains memory statistics
type MemoryStats struct {
	AllocMB      int64     // Currently allocated memory in MB
	TotalAllocMB int64     // Total allocated memory in MB
	SysMB        int64     // System memory in MB
	NumGC        uint32    // Number of GC cycles
	PauseTotalNs uint64    // Total GC pause time in nanoseconds
	LastGC       time.Time // Last GC time
}

// SetGCPercent sets the garbage collection percentage
func (mm *MemoryManager) SetGCPercent(percent int) {
	mm.gcPercent = percent
	debug.SetGCPercent(percent)
}

// Close stops the memory manager
func (mm *MemoryManager) Close() {
	if mm.ticker != nil {
		mm.ticker.Stop()
	}
	close(mm.stopChan)
}

// Global convenience functions using the global pools

// GetPooledDetectionResult returns a pooled DetectionResult from global pools
func GetPooledDetectionResult() *DetectionResult {
	return globalPools.GetDetectionResult()
}

// PutPooledDetectionResult returns a DetectionResult to global pools
func PutPooledDetectionResult(result *DetectionResult) {
	globalPools.PutDetectionResult(result)
}

// GetPooledASNInfo returns a pooled ASNInfo from global pools
func GetPooledASNInfo() *ASNInfo {
	return globalPools.GetASNInfo()
}

// PutPooledASNInfo returns an ASNInfo to global pools
func PutPooledASNInfo(info *ASNInfo) {
	globalPools.PutASNInfo(info)
}

// GetPooledVPNProvider returns a pooled VPNProvider from global pools
func GetPooledVPNProvider() *VPNProvider {
	return globalPools.GetVPNProvider()
}

// PutPooledVPNProvider returns a VPNProvider to global pools
func PutPooledVPNProvider(provider *VPNProvider) {
	globalPools.PutVPNProvider(provider)
}

// GetPooledIPSlice returns a pooled IP slice from global pools
func GetPooledIPSlice() []net.IP {
	return globalPools.GetIPSlice()
}

// PutPooledIPSlice returns an IP slice to global pools
func PutPooledIPSlice(slice []net.IP) {
	globalPools.PutIPSlice(slice)
}

// GetPooledStringSlice returns a pooled string slice from global pools
func GetPooledStringSlice() []string {
	return globalPools.GetStringSlice()
}

// PutPooledStringSlice returns a string slice to global pools
func PutPooledStringSlice(slice []string) {
	globalPools.PutStringSlice(slice)
}

// GetPooledByteSlice returns a pooled byte slice from global pools
func GetPooledByteSlice() []byte {
	return globalPools.GetByteSlice()
}

// PutPooledByteSlice returns a byte slice to global pools
func PutPooledByteSlice(slice []byte) {
	globalPools.PutByteSlice(slice)
}
