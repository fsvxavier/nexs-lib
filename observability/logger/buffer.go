package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// CircularBuffer implementa um buffer circular thread-safe para alta performance
type CircularBuffer struct {
	config        *interfaces.BufferConfig
	entries       []*interfaces.LogEntry
	head          int64
	tail          int64
	size          int64
	capacity      int64
	writer        io.Writer
	mu            sync.RWMutex
	flushMu       sync.Mutex
	stats         interfaces.BufferStats
	stopCh        chan struct{}
	flushCh       chan struct{}
	wg            sync.WaitGroup
	closed        int32
	memoryUsage   int64
	lastFlushTime time.Time
}

// NewCircularBuffer cria um novo buffer circular
func NewCircularBuffer(config *interfaces.BufferConfig, writer io.Writer) *CircularBuffer {
	if config == nil {
		config = DefaultBufferConfig()
	}

	if config.Size <= 0 {
		config.Size = 1000
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}

	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 5 * time.Second
	}

	buffer := &CircularBuffer{
		config:        config,
		entries:       make([]*interfaces.LogEntry, config.Size),
		capacity:      int64(config.Size),
		writer:        writer,
		stopCh:        make(chan struct{}),
		flushCh:       make(chan struct{}, 1),
		lastFlushTime: time.Now(),
	}

	// Inicializa estatísticas
	buffer.stats = interfaces.BufferStats{
		BufferSize: config.Size,
		LastFlush:  time.Now(),
	}

	// Inicia goroutine de flush automático se habilitado
	if config.AutoFlush {
		buffer.wg.Add(1)
		go buffer.autoFlushWorker()
	}

	return buffer
}

// DefaultBufferConfig retorna configuração padrão do buffer
func DefaultBufferConfig() *interfaces.BufferConfig {
	return &interfaces.BufferConfig{
		Enabled:      true,
		Size:         1000,
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MemoryLimit:  50 * 1024 * 1024, // 50MB
		AutoFlush:    true,
		ForceSync:    false,
	}
}

// Write adiciona uma entrada no buffer
func (cb *CircularBuffer) Write(entry *interfaces.LogEntry) error {
	if atomic.LoadInt32(&cb.closed) == 1 {
		return fmt.Errorf("buffer is closed")
	}

	if !cb.config.Enabled {
		return cb.writeDirectly(entry)
	}

	// Calcula o tamanho estimado da entrada
	entry.Size = cb.estimateEntrySize(entry)

	// Verifica limite de memória
	if cb.config.MemoryLimit > 0 {
		currentMemory := atomic.LoadInt64(&cb.memoryUsage)
		if currentMemory+entry.Size > cb.config.MemoryLimit {
			// Força flush se próximo do limite
			cb.triggerFlush()
			// Espera um pouco para o flush completar
			time.Sleep(10 * time.Millisecond)
		}
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Se buffer está cheio, remove entrada mais antiga
	if atomic.LoadInt64(&cb.size) >= cb.capacity {
		oldEntry := cb.entries[cb.head%cb.capacity]
		if oldEntry != nil {
			atomic.AddInt64(&cb.memoryUsage, -oldEntry.Size)
			atomic.AddInt64(&cb.stats.DroppedEntries, 1)
		}
		atomic.AddInt64(&cb.head, 1)
	} else {
		atomic.AddInt64(&cb.size, 1)
	}

	// Adiciona nova entrada
	cb.entries[cb.tail%cb.capacity] = entry
	atomic.AddInt64(&cb.tail, 1)
	atomic.AddInt64(&cb.memoryUsage, entry.Size)
	atomic.AddInt64(&cb.stats.TotalEntries, 1)

	// Verifica se deve fazer flush por tamanho do batch
	if cb.config.AutoFlush && int(atomic.LoadInt64(&cb.size)) >= cb.config.BatchSize {
		cb.triggerFlush()
	}

	return nil
}

// Flush força o flush de todas as entradas pendentes
func (cb *CircularBuffer) Flush() error {
	if atomic.LoadInt32(&cb.closed) == 1 {
		return fmt.Errorf("buffer is closed")
	}

	cb.flushMu.Lock()
	defer cb.flushMu.Unlock()

	start := time.Now()
	defer func() {
		cb.stats.FlushDuration = time.Since(start)
		cb.stats.LastFlush = time.Now()
		cb.lastFlushTime = time.Now()
	}()

	cb.mu.Lock()
	size := atomic.LoadInt64(&cb.size)
	if size == 0 {
		cb.mu.Unlock()
		return nil
	}

	// Copia entradas para slice temporário
	entries := make([]*interfaces.LogEntry, size)
	head := atomic.LoadInt64(&cb.head)

	for i := int64(0); i < size; i++ {
		entries[i] = cb.entries[(head+i)%cb.capacity]
	}

	// Limpa o buffer
	atomic.StoreInt64(&cb.head, 0)
	atomic.StoreInt64(&cb.tail, 0)
	atomic.StoreInt64(&cb.size, 0)
	atomic.StoreInt64(&cb.memoryUsage, 0)

	// Limpa array
	for i := range cb.entries {
		cb.entries[i] = nil
	}

	cb.mu.Unlock()

	// Escreve entradas
	err := cb.writeEntries(entries)
	if err != nil {
		return fmt.Errorf("error writing entries: %w", err)
	}

	atomic.AddInt64(&cb.stats.FlushCount, 1)
	cb.stats.UsedSize = int(atomic.LoadInt64(&cb.size))

	// Força sincronização se configurado
	if cb.config.ForceSync {
		if syncer, ok := cb.writer.(interface{ Sync() error }); ok {
			if syncErr := syncer.Sync(); syncErr != nil {
				return fmt.Errorf("error syncing: %w", syncErr)
			}
		}
	}

	return nil
}

// Close fecha o buffer e faz flush final
func (cb *CircularBuffer) Close() error {
	if !atomic.CompareAndSwapInt32(&cb.closed, 0, 1) {
		return nil // Já fechado
	}

	// Para worker de auto flush
	close(cb.stopCh)
	cb.wg.Wait()

	// Flush final
	if atomic.LoadInt64(&cb.size) > 0 {
		return cb.Flush()
	}

	return nil
}

// Stats retorna estatísticas do buffer
func (cb *CircularBuffer) Stats() interfaces.BufferStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return interfaces.BufferStats{
		BufferSize:     cb.stats.BufferSize,
		UsedSize:       int(atomic.LoadInt64(&cb.size)),
		MemoryUsage:    atomic.LoadInt64(&cb.memoryUsage),
		TotalEntries:   cb.stats.TotalEntries,
		DroppedEntries: cb.stats.DroppedEntries,
		FlushCount:     cb.stats.FlushCount,
		LastFlush:      cb.stats.LastFlush,
		FlushDuration:  cb.stats.FlushDuration,
	}
}

// IsFull verifica se o buffer está cheio
func (cb *CircularBuffer) IsFull() bool {
	return atomic.LoadInt64(&cb.size) >= cb.capacity
}

// Size retorna o número atual de entradas no buffer
func (cb *CircularBuffer) Size() int {
	return int(atomic.LoadInt64(&cb.size))
}

// Clear limpa o buffer sem fazer flush
func (cb *CircularBuffer) Clear() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.StoreInt64(&cb.head, 0)
	atomic.StoreInt64(&cb.tail, 0)
	atomic.StoreInt64(&cb.size, 0)
	atomic.StoreInt64(&cb.memoryUsage, 0)

	for i := range cb.entries {
		cb.entries[i] = nil
	}
}

// triggerFlush sinaliza para fazer flush (non-blocking)
func (cb *CircularBuffer) triggerFlush() {
	select {
	case cb.flushCh <- struct{}{}:
	default:
		// Canal está cheio, flush já foi solicitado
	}
}

// autoFlushWorker executa flush automático baseado em timeout
func (cb *CircularBuffer) autoFlushWorker() {
	defer cb.wg.Done()

	ticker := time.NewTicker(cb.config.FlushTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-cb.stopCh:
			return

		case <-ticker.C:
			// Flush por timeout se houver entradas
			if atomic.LoadInt64(&cb.size) > 0 {
				if err := cb.Flush(); err != nil {
					// Log do erro (pode usar um logger simples aqui)
					fmt.Printf("Error during auto flush: %v\n", err)
				}
			}

		case <-cb.flushCh:
			// Flush solicitado
			if err := cb.Flush(); err != nil {
				fmt.Printf("Error during requested flush: %v\n", err)
			}
		}
	}
}

// writeDirectly escreve diretamente sem buffer
func (cb *CircularBuffer) writeDirectly(entry *interfaces.LogEntry) error {
	return cb.writeEntries([]*interfaces.LogEntry{entry})
}

// writeEntries escreve um slice de entradas para o writer
func (cb *CircularBuffer) writeEntries(entries []*interfaces.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		// Serializa entrada como JSON
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("error marshaling log entry: %w", err)
		}

		// Adiciona newline
		data = append(data, '\n')

		// Escreve para o writer
		if _, err := cb.writer.Write(data); err != nil {
			return fmt.Errorf("error writing to output: %w", err)
		}
	}

	return nil
}

// estimateEntrySize estima o tamanho de uma entrada em bytes
func (cb *CircularBuffer) estimateEntrySize(entry *interfaces.LogEntry) int64 {
	size := int64(0)

	// Tamanhos base
	size += int64(len(entry.Message))
	size += int64(len(entry.Code))
	size += int64(len(entry.Source))
	size += int64(len(entry.Stack))
	size += 64 // Timestamp e outras estruturas

	// Estima campos
	for key, value := range entry.Fields {
		size += int64(len(key))

		switch v := value.(type) {
		case string:
			size += int64(len(v))
		case []byte:
			size += int64(len(v))
		default:
			size += 32 // Estimativa para outros tipos
		}
	}

	return size
}
