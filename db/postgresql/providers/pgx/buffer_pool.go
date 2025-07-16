package pgx

import (
	"sync"
	"sync/atomic"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// BufferPoolImpl implements the BufferPool interface
type BufferPoolImpl struct {
	pools map[int]*sync.Pool
	stats *MemoryStatsImpl
	mu    sync.RWMutex
}

// MemoryStatsImpl implements memory statistics tracking
type MemoryStatsImpl struct {
	bufferSize         int64
	allocatedBuffers   int32
	pooledBuffers      int32
	totalAllocations   int64
	totalDeallocations int64
	mu                 sync.RWMutex
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() interfaces.BufferPool {
	return &BufferPoolImpl{
		pools: make(map[int]*sync.Pool),
		stats: &MemoryStatsImpl{},
	}
}

// Get retrieves a buffer of the specified size
func (bp *BufferPoolImpl) Get(size int) []byte {
	bp.mu.RLock()
	pool, exists := bp.pools[size]
	bp.mu.RUnlock()

	if !exists {
		bp.mu.Lock()
		// Double-check pattern
		if pool, exists = bp.pools[size]; !exists {
			pool = &sync.Pool{
				New: func() interface{} {
					atomic.AddInt64(&bp.stats.totalAllocations, 1)
					atomic.AddInt32(&bp.stats.allocatedBuffers, 1)
					return make([]byte, size)
				},
			}
			bp.pools[size] = pool
		}
		bp.mu.Unlock()
	}

	buf := pool.Get().([]byte)
	atomic.AddInt32(&bp.stats.pooledBuffers, -1)

	// Reset buffer
	for i := range buf {
		buf[i] = 0
	}

	return buf
}

// Put returns a buffer to the pool
func (bp *BufferPoolImpl) Put(buf []byte) {
	if buf == nil {
		return
	}

	size := len(buf)
	bp.mu.RLock()
	pool, exists := bp.pools[size]
	bp.mu.RUnlock()

	if exists {
		pool.Put(buf)
		atomic.AddInt32(&bp.stats.pooledBuffers, 1)
		atomic.AddInt64(&bp.stats.totalDeallocations, 1)
	}
}

// Stats returns current memory statistics
func (bp *BufferPoolImpl) Stats() interfaces.MemoryStats {
	bp.stats.mu.RLock()
	defer bp.stats.mu.RUnlock()

	return interfaces.MemoryStats{
		BufferSize:         atomic.LoadInt64(&bp.stats.bufferSize),
		AllocatedBuffers:   atomic.LoadInt32(&bp.stats.allocatedBuffers),
		PooledBuffers:      atomic.LoadInt32(&bp.stats.pooledBuffers),
		TotalAllocations:   atomic.LoadInt64(&bp.stats.totalAllocations),
		TotalDeallocations: atomic.LoadInt64(&bp.stats.totalDeallocations),
	}
}

// Reset clears all buffers from the pool
func (bp *BufferPoolImpl) Reset() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.pools = make(map[int]*sync.Pool)

	// Reset stats
	atomic.StoreInt64(&bp.stats.bufferSize, 0)
	atomic.StoreInt32(&bp.stats.allocatedBuffers, 0)
	atomic.StoreInt32(&bp.stats.pooledBuffers, 0)
	atomic.StoreInt64(&bp.stats.totalAllocations, 0)
	atomic.StoreInt64(&bp.stats.totalDeallocations, 0)
}

// SafetyMonitorImpl implements thread-safety monitoring
type SafetyMonitorImpl struct {
	deadlocks       []interfaces.DeadlockInfo
	raceConditions  []interfaces.RaceConditionInfo
	leaks           []interfaces.LeakInfo
	isHealthy       bool
	lastHealthCheck time.Time
	mu              sync.RWMutex
}

// NewSafetyMonitor creates a new safety monitor
func NewSafetyMonitor() interfaces.SafetyMonitor {
	return &SafetyMonitorImpl{
		deadlocks:       make([]interfaces.DeadlockInfo, 0),
		raceConditions:  make([]interfaces.RaceConditionInfo, 0),
		leaks:           make([]interfaces.LeakInfo, 0),
		isHealthy:       true,
		lastHealthCheck: time.Now(),
	}
}

// CheckDeadlocks checks for potential deadlocks
func (sm *SafetyMonitorImpl) CheckDeadlocks() []interfaces.DeadlockInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to avoid race conditions
	deadlocks := make([]interfaces.DeadlockInfo, len(sm.deadlocks))
	copy(deadlocks, sm.deadlocks)
	return deadlocks
}

// CheckRaceConditions checks for potential race conditions
func (sm *SafetyMonitorImpl) CheckRaceConditions() []interfaces.RaceConditionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to avoid race conditions
	raceConditions := make([]interfaces.RaceConditionInfo, len(sm.raceConditions))
	copy(raceConditions, sm.raceConditions)
	return raceConditions
}

// CheckLeaks checks for resource leaks
func (sm *SafetyMonitorImpl) CheckLeaks() []interfaces.LeakInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to avoid race conditions
	leaks := make([]interfaces.LeakInfo, len(sm.leaks))
	copy(leaks, sm.leaks)
	return leaks
}

// IsHealthy returns the current health status
func (sm *SafetyMonitorImpl) IsHealthy() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isHealthy
}

// recordDeadlock records a deadlock incident
func (sm *SafetyMonitorImpl) recordDeadlock(goroutines []string, stackTraces map[string]string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	deadlock := interfaces.DeadlockInfo{
		Timestamp:   time.Now(),
		Goroutines:  goroutines,
		StackTraces: stackTraces,
	}

	sm.deadlocks = append(sm.deadlocks, deadlock)
	sm.isHealthy = false
}

// recordRaceCondition records a race condition incident
func (sm *SafetyMonitorImpl) recordRaceCondition(location, details string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	raceCondition := interfaces.RaceConditionInfo{
		Timestamp: time.Now(),
		Location:  location,
		Details:   details,
	}

	sm.raceConditions = append(sm.raceConditions, raceCondition)
	sm.isHealthy = false
}

// recordLeak records a resource leak incident
func (sm *SafetyMonitorImpl) recordLeak(resource string, count int64, details string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	leak := interfaces.LeakInfo{
		Timestamp: time.Now(),
		Resource:  resource,
		Count:     count,
		Details:   details,
	}

	sm.leaks = append(sm.leaks, leak)
	sm.isHealthy = false
}

// PerformHealthCheck performs a comprehensive health check
func (sm *SafetyMonitorImpl) PerformHealthCheck() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lastHealthCheck = time.Now()

	// Check if there are recent issues (within last 5 minutes)
	cutoff := time.Now().Add(-5 * time.Minute)
	hasRecentIssues := false

	for _, deadlock := range sm.deadlocks {
		if deadlock.Timestamp.After(cutoff) {
			hasRecentIssues = true
			break
		}
	}

	if !hasRecentIssues {
		for _, raceCondition := range sm.raceConditions {
			if raceCondition.Timestamp.After(cutoff) {
				hasRecentIssues = true
				break
			}
		}
	}

	if !hasRecentIssues {
		for _, leak := range sm.leaks {
			if leak.Timestamp.After(cutoff) {
				hasRecentIssues = true
				break
			}
		}
	}

	sm.isHealthy = !hasRecentIssues
}

// ClearOldRecords clears records older than the specified duration
func (sm *SafetyMonitorImpl) ClearOldRecords(maxAge time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	// Filter deadlocks
	filteredDeadlocks := make([]interfaces.DeadlockInfo, 0)
	for _, deadlock := range sm.deadlocks {
		if deadlock.Timestamp.After(cutoff) {
			filteredDeadlocks = append(filteredDeadlocks, deadlock)
		}
	}
	sm.deadlocks = filteredDeadlocks

	// Filter race conditions
	filteredRaceConditions := make([]interfaces.RaceConditionInfo, 0)
	for _, raceCondition := range sm.raceConditions {
		if raceCondition.Timestamp.After(cutoff) {
			filteredRaceConditions = append(filteredRaceConditions, raceCondition)
		}
	}
	sm.raceConditions = filteredRaceConditions

	// Filter leaks
	filteredLeaks := make([]interfaces.LeakInfo, 0)
	for _, leak := range sm.leaks {
		if leak.Timestamp.After(cutoff) {
			filteredLeaks = append(filteredLeaks, leak)
		}
	}
	sm.leaks = filteredLeaks
}
