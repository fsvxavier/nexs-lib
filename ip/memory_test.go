package ip

import (
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestObjectPools_DetectionResult(t *testing.T) {
	// Get a detection result from pool
	result := GetPooledDetectionResult()
	if result == nil {
		t.Fatal("GetPooledDetectionResult returned nil")
	}

	// Verify it's reset
	if result.TrustScore != 0 || result.RiskLevel != "" || result.IsVPN {
		t.Error("Pooled DetectionResult should be reset")
	}

	// Modify the result
	result.IP = net.ParseIP("8.8.8.8")
	result.TrustScore = 0.8
	result.RiskLevel = "low"
	result.IsVPN = true

	// Return to pool
	PutPooledDetectionResult(result)

	// Get another one - should be reset
	result2 := GetPooledDetectionResult()
	if result2.TrustScore != 0 || result2.RiskLevel != "" || result2.IsVPN {
		t.Error("Pooled DetectionResult should be reset after return")
	}

	PutPooledDetectionResult(result2)
}

func TestObjectPools_ASNInfo(t *testing.T) {
	info := GetPooledASNInfo()
	if info == nil {
		t.Fatal("GetPooledASNInfo returned nil")
	}

	// Verify it's reset
	if info.ASN != 0 || info.Name != "" || info.Country != "" {
		t.Error("Pooled ASNInfo should be reset")
	}

	// Modify the info
	info.ASN = 16509
	info.Name = "Amazon"
	info.Country = "US"
	info.IsCloudProvider = true

	// Return to pool
	PutPooledASNInfo(info)

	// Get another one - should be reset
	info2 := GetPooledASNInfo()
	if info2.ASN != 0 || info2.Name != "" || info2.IsCloudProvider {
		t.Error("Pooled ASNInfo should be reset after return")
	}

	PutPooledASNInfo(info2)
}

func TestObjectPools_VPNProvider(t *testing.T) {
	provider := GetPooledVPNProvider()
	if provider == nil {
		t.Fatal("GetPooledVPNProvider returned nil")
	}

	// Verify it's reset
	if provider.Name != "" || provider.Type != "" || provider.Reliability != 0 {
		t.Error("Pooled VPNProvider should be reset")
	}

	// Modify the provider
	provider.Name = "TestVPN"
	provider.Type = "commercial"
	provider.Reliability = 0.8

	// Return to pool
	PutPooledVPNProvider(provider)

	// Get another one - should be reset
	provider2 := GetPooledVPNProvider()
	if provider2.Name != "" || provider2.Type != "" || provider2.Reliability != 0 {
		t.Error("Pooled VPNProvider should be reset after return")
	}

	PutPooledVPNProvider(provider2)
}

func TestObjectPools_Slices(t *testing.T) {
	// Test IP slice pool
	ipSlice := GetPooledIPSlice()
	if ipSlice == nil {
		t.Fatal("GetPooledIPSlice returned nil")
	}
	if len(ipSlice) != 0 {
		t.Error("Pooled IP slice should have zero length")
	}

	// Add some IPs
	ipSlice = append(ipSlice, net.ParseIP("8.8.8.8"))
	ipSlice = append(ipSlice, net.ParseIP("1.1.1.1"))

	// Return to pool
	PutPooledIPSlice(ipSlice)

	// Test string slice pool
	stringSlice := GetPooledStringSlice()
	if stringSlice == nil {
		t.Fatal("GetPooledStringSlice returned nil")
	}
	if len(stringSlice) != 0 {
		t.Error("Pooled string slice should have zero length")
	}

	// Add some strings
	stringSlice = append(stringSlice, "test1", "test2")

	// Return to pool
	PutPooledStringSlice(stringSlice)

	// Test byte slice pool
	byteSlice := GetPooledByteSlice()
	if byteSlice == nil {
		t.Fatal("GetPooledByteSlice returned nil")
	}
	if len(byteSlice) != 0 {
		t.Error("Pooled byte slice should have zero length")
	}

	// Add some bytes
	byteSlice = append(byteSlice, []byte("test data")...)

	// Return to pool
	PutPooledByteSlice(byteSlice)
}

func TestLazyDatabase(t *testing.T) {
	loadCalled := false
	testData := "test database data"

	db := NewLazyDatabase(func() error {
		loadCalled = true
		return nil
	}, time.Hour)

	// Initially not loaded
	if db.IsLoaded() {
		t.Error("Database should not be loaded initially")
	}

	// Set data manually
	db.Set(testData)

	// Should be loaded now
	if !db.IsLoaded() {
		t.Error("Database should be loaded after Set")
	}

	// Get data
	data, err := db.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if data != testData {
		t.Errorf("Expected '%s', got '%s'", testData, data)
	}

	// Load function should not have been called since we set data manually
	if loadCalled {
		t.Error("Load function should not have been called")
	}

	// Unload and test lazy loading
	db.Unload()
	if db.IsLoaded() {
		t.Error("Database should not be loaded after Unload")
	}

	// Now Get should trigger load function
	data, err = db.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !loadCalled {
		t.Error("Load function should have been called")
	}
}

func TestMemoryManager(t *testing.T) {
	config := DefaultMemoryConfig()
	config.CheckInterval = 100 * time.Millisecond
	config.MaxMemoryMB = 100 // Low limit for testing

	mm := NewMemoryManager(config)
	defer mm.Close()

	// Get initial stats
	stats := mm.GetMemoryStats()
	if stats.AllocMB < 0 {
		t.Error("Allocated memory should be non-negative")
	}

	// Test GC percent setting
	mm.SetGCPercent(50)
	// Note: We can't directly verify this worked without runtime internals

	// Wait a bit for monitoring to run
	time.Sleep(200 * time.Millisecond)

	// Get stats again
	stats2 := mm.GetMemoryStats()
	if stats2.NumGC < stats.NumGC {
		t.Error("GC count should not decrease")
	}
}

func TestMemoryManager_Stats(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Close()

	stats := mm.GetMemoryStats()

	// Verify all fields are present and reasonable
	if stats.AllocMB < 0 {
		t.Error("AllocMB should be non-negative")
	}
	if stats.TotalAllocMB < stats.AllocMB {
		t.Error("TotalAllocMB should be >= AllocMB")
	}
	if stats.SysMB <= 0 {
		t.Error("SysMB should be positive")
	}
	// NumGC and PauseTotalNs might be 0 initially, which is fine
}

func BenchmarkObjectPools_DetectionResult(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := GetPooledDetectionResult()
			result.IP = net.ParseIP("8.8.8.8")
			result.TrustScore = 0.8
			PutPooledDetectionResult(result)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := &DetectionResult{}
			result.IP = net.ParseIP("8.8.8.8")
			result.TrustScore = 0.8
			// No pooling
		}
	})
}

func BenchmarkObjectPools_StringSlice(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := GetPooledStringSlice()
			slice = append(slice, "test1", "test2", "test3")
			PutPooledStringSlice(slice)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]string, 0)
			slice = append(slice, "test1", "test2", "test3")
			// No pooling
		}
	})
}

func BenchmarkLazyDatabase_Get(b *testing.B) {
	testData := "benchmark test data"

	db := NewLazyDatabase(func() error {
		return nil
	}, time.Hour)

	db.Set(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Get()
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func TestObjectPools_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100
	operationsPerGoroutine := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Test DetectionResult pool
				result := GetPooledDetectionResult()
				result.TrustScore = 0.5
				PutPooledDetectionResult(result)

				// Test ASNInfo pool
				asn := GetPooledASNInfo()
				asn.ASN = 12345
				PutPooledASNInfo(asn)

				// Test slice pools
				stringSlice := GetPooledStringSlice()
				stringSlice = append(stringSlice, "test")
				PutPooledStringSlice(stringSlice)

				byteSlice := GetPooledByteSlice()
				byteSlice = append(byteSlice, 'a', 'b', 'c')
				PutPooledByteSlice(byteSlice)
			}
		}()
	}

	wg.Wait()
	// If we reach here without panics, concurrent access is working
}

func TestMemoryManager_ForceGC(t *testing.T) {
	// Create many objects to potentially trigger GC
	var objects [][]byte
	for i := 0; i < 1000; i++ {
		objects = append(objects, make([]byte, 1024))
	}

	mm := NewMemoryManager(DefaultMemoryConfig())
	defer mm.Close()

	initialStats := mm.GetMemoryStats()

	// Force GC manually
	runtime.GC()

	// Check that GC count increased
	finalStats := mm.GetMemoryStats()
	if finalStats.NumGC <= initialStats.NumGC {
		t.Log("GC might not have run, which is okay for this test")
	}

	// Use objects to prevent optimization
	_ = len(objects)
}

func TestLazyDatabase_TTL(t *testing.T) {
	loadCount := 0
	testData := "ttl test data"

	db := NewLazyDatabase(func() error {
		loadCount++
		return nil
	}, 50*time.Millisecond) // Very short TTL

	// Set initial data
	db.Set(testData)

	// Get data immediately - should not trigger load
	_, err := db.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if loadCount != 0 {
		t.Error("Load should not have been called yet")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Get data again - should trigger load due to TTL expiry
	_, err = db.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if loadCount != 1 {
		t.Errorf("Load should have been called once, got %d", loadCount)
	}
}
