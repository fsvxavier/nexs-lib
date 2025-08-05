package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	// Test concurrent access to metrics
	wg := sync.WaitGroup{}
	ops := 1000

	for i := 0; i < ops; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			collector.RecordHit()
		}()
		go func() {
			defer wg.Done()
			collector.RecordMiss()
		}()
	}

	wg.Wait()

	stats := collector.GetStats()
	assert.Equal(t, uint64(ops), stats.Hits)
	assert.Equal(t, uint64(ops), stats.Misses)

	collector.ResetStats()
	stats = collector.GetStats()
	assert.Equal(t, uint64(0), stats.Hits)
	assert.Equal(t, uint64(0), stats.Misses)
}

func TestObjectPool(t *testing.T) {
	pool := NewObjectPool()

	// Test concurrent access to pool
	wg := sync.WaitGroup{}
	ops := 1000

	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := pool.Get()
			assert.NotNil(t, obj)
			assert.Len(t, obj, 0)
			obj["test"] = "value"
			pool.Put(obj)
		}()
	}

	wg.Wait()

	// Verify object reuse
	obj := pool.Get()
	assert.Len(t, obj, 0)
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3)

	// Test basic operations
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify values
	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Test eviction
	cache.Set("key4", "value4")
	_, ok = cache.Get("key1")
	assert.True(t, ok) // key1 was recently accessed

	// Test concurrent access
	wg := sync.WaitGroup{}
	ops := 1000

	for i := 0; i < ops; i++ {
		wg.Add(2)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key_%d", index)
			cache.Set(key, index)
		}(i)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key_%d", index)
			_, _ = cache.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestKeyGenerator(t *testing.T) {
	gen := NewKeyGenerator()

	// Test basic key generation
	key := gen.GenerateKey("part1", "part2")
	assert.Equal(t, "part1:part2", key)

	// Test concurrent key generation
	wg := sync.WaitGroup{}
	ops := 1000

	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := gen.GenerateKey(
				fmt.Sprintf("prefix_%d", index),
				fmt.Sprintf("suffix_%d", index),
			)
			assert.Contains(t, key, "prefix")
			assert.Contains(t, key, "suffix")
		}(i)
	}

	wg.Wait()
}

func BenchmarkMetricsCollector(b *testing.B) {
	collector := NewMetricsCollector()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.RecordHit()
			collector.RecordMiss()
			collector.GetStats()
		}
	})
}

func BenchmarkObjectPool(b *testing.B) {
	pool := NewObjectPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			obj["test"] = "value"
			pool.Put(obj)
		}
	})
}

func BenchmarkLRUCache(b *testing.B) {
	cache := NewLRUCache(1000)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%100)
			if i%2 == 0 {
				cache.Set(key, i)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

func BenchmarkKeyGenerator(b *testing.B) {
	gen := NewKeyGenerator()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			gen.GenerateKey(
				fmt.Sprintf("prefix_%d", i%100),
				fmt.Sprintf("suffix_%d", i%100),
			)
			i++
		}
	})
}
