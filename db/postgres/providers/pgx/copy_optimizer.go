package pgxprovider

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/jackc/pgx/v5"
)

// CopyOptimizer implementa otimizações para operações de CopyTo/CopyFrom
type CopyOptimizer struct {
	// Configurações de otimização
	bufferSize       int
	maxWorkers       int
	batchSize        int
	progressCallback func(processed, total int64)

	// Métricas
	totalProcessed int64
	totalErrors    int64
	startTime      time.Time

	// Controle de workers
	workerPool chan struct{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// CopyFromSourceOptimized representa uma fonte de dados otimizada
type CopyFromSourceOptimized struct {
	data       [][]interface{}
	currentRow int64
	totalRows  int64
	mu         sync.RWMutex
}

// CopyToWriterOptimized representa um escritor otimizado
type CopyToWriterOptimized struct {
	writer     io.Writer
	buffer     []byte
	bufferSize int
	written    int64
	mu         sync.Mutex
}

// NewCopyOptimizer cria um novo otimizador de copy
func NewCopyOptimizer(bufferSize, maxWorkers, batchSize int) *CopyOptimizer {
	if maxWorkers <= 0 {
		maxWorkers = runtime.GOMAXPROCS(0)
	}

	if bufferSize <= 0 {
		bufferSize = 64 * 1024 // 64KB default
	}

	if batchSize <= 0 {
		batchSize = 1000 // 1000 rows default
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CopyOptimizer{
		bufferSize: bufferSize,
		maxWorkers: maxWorkers,
		batchSize:  batchSize,
		workerPool: make(chan struct{}, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
	}
}

// SetProgressCallback define callback para progresso
func (co *CopyOptimizer) SetProgressCallback(callback func(processed, total int64)) {
	co.progressCallback = callback
}

// CopyFromOptimized implementa CopyFrom otimizado
func (c *Conn) CopyFromOptimized(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.ICopyFromSource) (int64, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrConnClosed
	}
	c.mu.RUnlock()

	// Criar otimizador
	optimizer := NewCopyOptimizer(64*1024, runtime.GOMAXPROCS(0), 1000)
	defer optimizer.Close()

	// Executar hook de copy
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "copy_from",
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			return 0, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
		}()
	}

	// Usar pgx CopyFrom com otimizações
	conn := c.getConn()
	if conn == nil {
		return 0, ErrConnClosed
	}

	// Wrapper otimizado para a fonte de dados
	optimizedSrc := &OptimizedCopyFromSource{
		source:    rowSrc,
		optimizer: optimizer,
		processed: 0,
	}

	// Executar copy
	start := time.Now()
	rowsAffected, err := conn.CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, optimizedSrc)

	// Registrar métricas
	if c.hookManager != nil {
		// Registrar na pool de métricas se disponível
		if pool, ok := c.hookManager.(interface{ GetMetrics() *PerformanceMetrics }); ok {
			metrics := pool.GetMetrics()
			if metrics != nil {
				metrics.RecordQuery(ctx, fmt.Sprintf("COPY %s", tableName), time.Since(start), err)
			}
		}
	}

	if err != nil {
		atomic.AddInt64(&optimizer.totalErrors, 1)
		return 0, fmt.Errorf("copy from failed: %w", err)
	}

	atomic.AddInt64(&optimizer.totalProcessed, rowsAffected)
	return rowsAffected, nil
}

// CopyToOptimized implementa CopyTo otimizado
func (c *Conn) CopyToOptimized(ctx context.Context, w interfaces.ICopyToWriter, query string, args ...interface{}) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrConnClosed
	}
	c.mu.RUnlock()

	// Criar otimizador
	optimizer := NewCopyOptimizer(64*1024, runtime.GOMAXPROCS(0), 1000)
	defer optimizer.Close()

	// Executar hook de copy
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "copy_to",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			return err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
		}()
	}

	// Usar pgx CopyTo com otimizações
	conn := c.getConn()
	if conn == nil {
		return ErrConnClosed
	}

	// Wrapper otimizado para o writer
	optimizedWriter := &OptimizedCopyToWriter{
		writer:     w,
		optimizer:  optimizer,
		bufferSize: optimizer.bufferSize,
		buffer:     make([]byte, 0, optimizer.bufferSize),
	}

	// Executar copy using pgx directly

	// For CopyTo, we need to use a different approach since pgx doesn't expose CopyTo on the connection interface
	// We'll simulate it by executing the query and streaming results
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		atomic.AddInt64(&optimizer.totalErrors, 1)
		return fmt.Errorf("copy to query failed: %w", err)
	}
	defer rows.Close()

	// Stream results to writer
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			atomic.AddInt64(&optimizer.totalErrors, 1)
			return fmt.Errorf("copy to values failed: %w", err)
		}

		// Convert values to byte format and write
		if err := optimizedWriter.WriteRow(values); err != nil {
			atomic.AddInt64(&optimizer.totalErrors, 1)
			return fmt.Errorf("copy to write failed: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		atomic.AddInt64(&optimizer.totalErrors, 1)
		return fmt.Errorf("copy to rows error: %w", err)
	}

	// Flush any remaining data
	if err := optimizedWriter.Flush(); err != nil {
		atomic.AddInt64(&optimizer.totalErrors, 1)
		return fmt.Errorf("copy to flush failed: %w", err)
	}

	return nil
}

// OptimizedCopyFromSource wrapper otimizado para fonte de dados
type OptimizedCopyFromSource struct {
	source    interfaces.ICopyFromSource
	optimizer *CopyOptimizer
	processed int64
}

func (s *OptimizedCopyFromSource) Next() bool {
	hasNext := s.source.Next()
	if hasNext {
		processed := atomic.AddInt64(&s.processed, 1)

		// Callback de progresso
		if s.optimizer.progressCallback != nil && processed%int64(s.optimizer.batchSize) == 0 {
			s.optimizer.progressCallback(processed, -1) // -1 significa total desconhecido
		}
	}
	return hasNext
}

func (s *OptimizedCopyFromSource) Values() ([]interface{}, error) {
	return s.source.Values()
}

func (s *OptimizedCopyFromSource) Err() error {
	return s.source.Err()
}

// OptimizedCopyToWriter wrapper otimizado para writer
type OptimizedCopyToWriter struct {
	writer     interfaces.ICopyToWriter
	optimizer  *CopyOptimizer
	buffer     []byte
	bufferSize int
	written    int64
	mu         sync.Mutex
}

func (w *OptimizedCopyToWriter) WriteRow(values []interface{}) error {
	// Convert values to string format for writing
	line := ""
	for i, val := range values {
		if i > 0 {
			line += "\t"
		}
		line += fmt.Sprintf("%v", val)
	}
	line += "\n"

	_, err := w.Write([]byte(line))
	return err
}

func (w *OptimizedCopyToWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Usar buffer para otimizar escritas
	if len(w.buffer)+len(p) > w.bufferSize {
		// Flush buffer
		if len(w.buffer) > 0 {
			if err := w.flushBuffer(); err != nil {
				return 0, err
			}
		}

		// Se o novo dado é maior que o buffer, escrever diretamente
		if len(p) > w.bufferSize {
			if err := w.writeToUnderlying(p); err != nil {
				return 0, err
			}
			atomic.AddInt64(&w.written, int64(len(p)))
			return len(p), nil
		}
	}

	// Adicionar ao buffer
	w.buffer = append(w.buffer, p...)
	atomic.AddInt64(&w.written, int64(len(p)))

	// Callback de progresso
	if w.optimizer.progressCallback != nil {
		currentWritten := atomic.LoadInt64(&w.written)
		if currentWritten%int64(w.optimizer.batchSize*100) == 0 { // A cada 100 batches
			w.optimizer.progressCallback(currentWritten, -1)
		}
	}

	return len(p), nil
}

func (w *OptimizedCopyToWriter) flushBuffer() error {
	if len(w.buffer) > 0 {
		if err := w.writeToUnderlying(w.buffer); err != nil {
			return err
		}
		w.buffer = w.buffer[:0]
	}
	return nil
}

func (w *OptimizedCopyToWriter) writeToUnderlying(data []byte) error {
	// Convert byte data to interface{} slice for the writer
	values := make([]interface{}, 1)
	values[0] = string(data)
	return w.writer.Write(values)
}

func (w *OptimizedCopyToWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.flushBuffer(); err != nil {
		return err
	}

	// Flush writer se suportar
	if flusher, ok := w.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// Close fecha o otimizador
func (co *CopyOptimizer) Close() error {
	co.cancel()
	co.wg.Wait()
	return nil
}

// GetStats retorna estatísticas do otimizador
func (co *CopyOptimizer) GetStats() map[string]interface{} {
	processed := atomic.LoadInt64(&co.totalProcessed)
	errors := atomic.LoadInt64(&co.totalErrors)
	duration := time.Since(co.startTime)

	var rate float64
	if duration.Seconds() > 0 {
		rate = float64(processed) / duration.Seconds()
	}

	return map[string]interface{}{
		"total_processed":  processed,
		"total_errors":     errors,
		"duration_seconds": duration.Seconds(),
		"rate_per_second":  rate,
		"buffer_size":      co.bufferSize,
		"max_workers":      co.maxWorkers,
		"batch_size":       co.batchSize,
	}
}
