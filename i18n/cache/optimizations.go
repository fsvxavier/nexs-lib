package cache

import (
	"sync"
	"sync/atomic"
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
