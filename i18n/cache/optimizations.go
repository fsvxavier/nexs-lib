package cache

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// CacheStats mantém estatísticas do cache
type CacheStats struct {
	Hits   uint64
	Misses uint64
}

// CacheMetrics interface para coletar métricas do cache
type CacheMetrics interface {
	GetStats() CacheStats
	ResetStats()
}

// MetricsCollector coleta métricas do cache
type MetricsCollector struct {
	hits   uint64
	misses uint64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (m *MetricsCollector) RecordHit() {
	atomic.AddUint64(&m.hits, 1)
}

func (m *MetricsCollector) RecordMiss() {
	atomic.AddUint64(&m.misses, 1)
}

func (m *MetricsCollector) GetStats() CacheStats {
	return CacheStats{
		Hits:   atomic.LoadUint64(&m.hits),
		Misses: atomic.LoadUint64(&m.misses),
	}
}

func (m *MetricsCollector) ResetStats() {
	atomic.StoreUint64(&m.hits, 0)
	atomic.StoreUint64(&m.misses, 0)
}

// ObjectPool é um pool de objetos para reduzir alocações
type ObjectPool struct {
	pool sync.Pool
}

func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 10) // tamanho inicial otimizado
			},
		},
	}
}

func (p *ObjectPool) Get() map[string]interface{} {
	return p.pool.Get().(map[string]interface{})
}

func (p *ObjectPool) Put(m map[string]interface{}) {
	// Limpa o mapa antes de devolver ao pool
	for k := range m {
		delete(m, k)
	}
	p.pool.Put(m)
}

// LRUCache implementa um cache LRU thread-safe
type LRUCache struct {
	capacity int
	items    sync.Map
	lru      *sync.Map
	size     int64
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    sync.Map{},
		lru:      &sync.Map{},
		size:     0,
	}
}

func (c *LRUCache) Set(key, value interface{}) {
	// Incrementa o tamanho atomicamente
	if atomic.LoadInt64(&c.size) >= int64(c.capacity) {
		c.evict()
	}

	c.items.Store(key, value)
	c.lru.Store(key, time.Now().UnixNano())
	atomic.AddInt64(&c.size, 1)
}

func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	value, ok := c.items.Load(key)
	if ok {
		c.lru.Store(key, time.Now().UnixNano())
	}
	return value, ok
}

func (c *LRUCache) evict() {
	var oldestKey interface{}
	var oldestTime int64 = math.MaxInt64

	c.lru.Range(func(key, value interface{}) bool {
		timestamp := value.(int64)
		if timestamp < oldestTime {
			oldestTime = timestamp
			oldestKey = key
		}
		return true
	})

	if oldestKey != nil {
		c.items.Delete(oldestKey)
		c.lru.Delete(oldestKey)
		atomic.AddInt64(&c.size, -1)
	}
}

// KeyGenerator gera chaves otimizadas para o cache
type KeyGenerator struct {
	pool sync.Pool
}

func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 64) // tamanho inicial otimizado
			},
		},
	}
}

func (g *KeyGenerator) GenerateKey(parts ...string) string {
	buf := g.pool.Get().([]byte)
	defer g.pool.Put(buf)

	buf = buf[:0] // reset buffer
	for i, part := range parts {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, part...)
	}

	return string(buf)
}
