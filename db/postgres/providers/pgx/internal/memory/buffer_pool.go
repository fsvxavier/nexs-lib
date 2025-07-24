package memory

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// BufferPool implementa IBufferPool com otimizações de memória
type BufferPool struct {
	// Pools para diferentes tamanhos de buffer
	pools map[int]*sync.Pool
	mu    sync.RWMutex

	// Estatísticas atômicas para performance
	stats atomic.Value // interfaces.MemoryStats

	// Configurações
	maxBufferSize int
	minBufferSize int

	// GC helper
	gcTicker *time.Ticker
	gcStop   chan struct{}
}

// NewBufferPool cria um novo pool de buffers otimizado
func NewBufferPool(minSize, maxSize int) interfaces.IBufferPool {
	if minSize <= 0 {
		minSize = 1024 // 1KB default
	}
	if maxSize <= 0 {
		maxSize = 1024 * 1024 // 1MB default
	}

	pool := &BufferPool{
		pools:         make(map[int]*sync.Pool),
		maxBufferSize: maxSize,
		minBufferSize: minSize,
		gcStop:        make(chan struct{}),
	}

	// Inicializar stats
	pool.stats.Store(interfaces.MemoryStats{})

	// Configurar GC automático a cada 5 minutos
	pool.gcTicker = time.NewTicker(5 * time.Minute)
	go pool.gcRoutine()

	return pool
}

// Get obtém um buffer do pool
func (p *BufferPool) Get(size int) []byte {
	// Normalizar tamanho para próxima potência de 2
	normalizedSize := normalizeSize(size)

	// Verificar limites
	if normalizedSize > p.maxBufferSize {
		// Para buffers muito grandes, alocar diretamente
		p.incrementStats(func(s *interfaces.MemoryStats) {
			s.TotalAllocations++
		})
		return make([]byte, size)
	}

	if normalizedSize < p.minBufferSize {
		normalizedSize = p.minBufferSize
	}

	// Obter pool para o tamanho
	pool := p.getPool(normalizedSize)

	// Tentar obter do pool
	if buf := pool.Get(); buf != nil {
		buffer := buf.([]byte)
		p.incrementStats(func(s *interfaces.MemoryStats) {
			s.PooledBuffers--
			s.AllocatedBuffers++
		})
		return buffer[:size] // Retornar o tamanho exato solicitado
	}

	// Alocar novo buffer
	buffer := make([]byte, normalizedSize)
	p.incrementStats(func(s *interfaces.MemoryStats) {
		s.TotalAllocations++
		s.AllocatedBuffers++
	})

	return buffer[:size]
}

// Put retorna um buffer para o pool
func (p *BufferPool) Put(buf []byte) {
	if buf == nil {
		return
	}

	capacity := cap(buf)

	// Verificar se o buffer é elegível para pooling
	if capacity > p.maxBufferSize || capacity < p.minBufferSize {
		p.incrementStats(func(s *interfaces.MemoryStats) {
			s.TotalDeallocations++
		})
		return
	}

	// Normalizar tamanho
	normalizedSize := normalizeSize(capacity)

	// Limpar buffer (security measure)
	for i := range buf {
		buf[i] = 0
	}

	// Restaurar capacidade total
	buf = buf[:normalizedSize]

	// Retornar ao pool
	pool := p.getPool(normalizedSize)
	pool.Put(buf)

	p.incrementStats(func(s *interfaces.MemoryStats) {
		s.AllocatedBuffers--
		s.PooledBuffers++
		s.TotalDeallocations++
	})
}

// Stats retorna estatísticas do pool
func (p *BufferPool) Stats() interfaces.MemoryStats {
	stats := p.stats.Load().(interfaces.MemoryStats)

	// Atualizar buffer size atual
	p.mu.RLock()
	totalBufferSize := int64(0)
	for size, pool := range p.pools {
		// Estimativa baseada no tamanho do pool
		totalBufferSize += int64(size * p.estimatePoolSize(pool))
	}
	p.mu.RUnlock()

	stats.BufferSize = totalBufferSize
	return stats
}

// Reset limpa todos os pools
func (p *BufferPool) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Limpar todos os pools
	for _, pool := range p.pools {
		// Forçar GC dos buffers
		*pool = sync.Pool{}
	}

	// Resetar estatísticas
	p.stats.Store(interfaces.MemoryStats{})
}

// Close fecha o pool e para o GC
func (p *BufferPool) Close() {
	if p.gcTicker != nil {
		p.gcTicker.Stop()
	}

	close(p.gcStop)
	p.Reset()
}

// getPool obtém ou cria um pool para o tamanho especificado
func (p *BufferPool) getPool(size int) *sync.Pool {
	p.mu.RLock()
	pool, exists := p.pools[size]
	p.mu.RUnlock()

	if exists {
		return pool
	}

	// Criar novo pool
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := p.pools[size]; exists {
		return pool
	}

	pool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, size)
		},
	}

	p.pools[size] = pool
	return pool
}

// incrementStats incrementa estatísticas de forma thread-safe
func (p *BufferPool) incrementStats(update func(*interfaces.MemoryStats)) {
	for {
		oldStats := p.stats.Load().(interfaces.MemoryStats)
		newStats := oldStats // Copy
		update(&newStats)

		if p.stats.CompareAndSwap(oldStats, newStats) {
			break
		}
	}
}

// estimatePoolSize estima o tamanho do pool (não é exato, mas suficiente para stats)
func (p *BufferPool) estimatePoolSize(pool *sync.Pool) int {
	// Isso é uma estimativa simples
	// Em produção, você pode implementar um mecanismo mais preciso
	count := 0
	for i := 0; i < 10; i++ {
		if buf := pool.Get(); buf != nil {
			count++
			pool.Put(buf)
		} else {
			break
		}
	}
	return count
}

// gcRoutine executa garbage collection periódico
func (p *BufferPool) gcRoutine() {
	for {
		select {
		case <-p.gcTicker.C:
			p.performGC()
		case <-p.gcStop:
			return
		}
	}
}

// performGC executa limpeza de memória
func (p *BufferPool) performGC() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Limpar pools não utilizados recentemente
	for size, pool := range p.pools {
		// Verificar se o pool está sendo usado
		if p.estimatePoolSize(pool) == 0 {
			// Pool vazio, pode ser removido
			delete(p.pools, size)
		}
	}
}

// normalizeSize normaliza o tamanho para próxima potência de 2
func normalizeSize(size int) int {
	if size <= 0 {
		return 1024 // 1KB minimum
	}

	// Encontrar próxima potência de 2
	power := 1
	for power < size {
		power <<= 1
	}

	return power
}

// SafeBufferPool é um wrapper thread-safe adicional (se necessário)
type SafeBufferPool struct {
	pool interfaces.IBufferPool
	mu   sync.RWMutex
}

// NewSafeBufferPool cria um wrapper thread-safe
func NewSafeBufferPool(pool interfaces.IBufferPool) interfaces.IBufferPool {
	return &SafeBufferPool{
		pool: pool,
	}
}

// Get thread-safe
func (s *SafeBufferPool) Get(size int) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pool.Get(size)
}

// Put thread-safe
func (s *SafeBufferPool) Put(buf []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pool.Put(buf)
}

// Stats thread-safe
func (s *SafeBufferPool) Stats() interfaces.MemoryStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pool.Stats()
}

// Reset thread-safe
func (s *SafeBufferPool) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pool.Reset()
}
