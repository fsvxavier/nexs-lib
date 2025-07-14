package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Worker representa um worker que executa operações no banco
type Worker struct {
	ID     int
	pool   postgresql.IPool
	stats  *WorkerStats
	ctx    context.Context
	cancel context.CancelFunc
}

// WorkerStats estatísticas do worker
type WorkerStats struct {
	mu              sync.RWMutex
	QueriesExecuted int64
	ErrorsCount     int64
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	LastQueryTime   time.Time
}

// NewWorker cria um novo worker
func NewWorker(id int, pool postgresql.IPool) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		ID:     id,
		pool:   pool,
		stats:  &WorkerStats{},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start inicia o worker
func (w *Worker) Start(wg *sync.WaitGroup, duration time.Duration) {
	defer wg.Done()

	fmt.Printf("🚀 Worker %d iniciado\n", w.ID)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-w.ctx.Done():
			fmt.Printf("⏹️ Worker %d parado por cancelamento\n", w.ID)
			return
		case <-timeout:
			fmt.Printf("⏰ Worker %d concluído por timeout\n", w.ID)
			return
		case <-ticker.C:
			if err := w.executeQuery(); err != nil {
				w.incrementError()
				log.Printf("❌ Worker %d erro: %v", w.ID, err)
			}
		}
	}
}

// Stop para o worker
func (w *Worker) Stop() {
	w.cancel()
}

// executeQuery executa uma query de exemplo
func (w *Worker) executeQuery() error {
	start := time.Now()

	// Usar timeout específico para a query
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
	defer cancel()

	conn, err := w.pool.AcquireWithTimeout(ctx, 2*time.Second)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	// Executar query simples
	var result int
	row := conn.QueryRow(ctx, "SELECT $1::int + $2::int", w.ID, time.Now().Unix()%100)
	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("erro ao executar query: %w", err)
	}

	duration := time.Since(start)
	w.updateStats(duration)

	return nil
}

// updateStats atualiza as estatísticas do worker
func (w *Worker) updateStats(duration time.Duration) {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()

	w.stats.QueriesExecuted++
	w.stats.TotalDuration += duration
	w.stats.AvgResponseTime = w.stats.TotalDuration / time.Duration(w.stats.QueriesExecuted)
	w.stats.LastQueryTime = time.Now()
}

// incrementError incrementa contador de erros
func (w *Worker) incrementError() {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()
	w.stats.ErrorsCount++
}

// GetStats retorna estatísticas do worker
func (w *Worker) GetStats() WorkerStats {
	w.stats.mu.RLock()
	defer w.stats.mu.RUnlock()
	return *w.stats
}

// WorkerManager gerencia múltiplos workers
type WorkerManager struct {
	workers []*Worker
	pool    postgresql.IPool
}

// NewWorkerManager cria um novo gerenciador de workers
func NewWorkerManager(pool postgresql.IPool) *WorkerManager {
	return &WorkerManager{
		pool: pool,
	}
}

// StartWorkers inicia múltiplos workers
func (wm *WorkerManager) StartWorkers(count int, duration time.Duration) {
	var wg sync.WaitGroup

	fmt.Printf("🏭 Iniciando %d workers por %v...\n", count, duration)

	// Criar e iniciar workers
	for i := 0; i < count; i++ {
		worker := NewWorker(i+1, wm.pool)
		wm.workers = append(wm.workers, worker)

		wg.Add(1)
		go worker.Start(&wg, duration)
	}

	// Aguardar conclusão
	wg.Wait()
	fmt.Printf("✅ Todos os workers finalizaram\n")
}

// StopAllWorkers para todos os workers
func (wm *WorkerManager) StopAllWorkers() {
	fmt.Printf("🛑 Parando todos os workers...\n")
	for _, worker := range wm.workers {
		worker.Stop()
	}
}

// GetAggregatedStats retorna estatísticas agregadas
func (wm *WorkerManager) GetAggregatedStats() map[string]interface{} {
	totalQueries := int64(0)
	totalErrors := int64(0)
	totalDuration := time.Duration(0)

	for _, worker := range wm.workers {
		stats := worker.GetStats()
		totalQueries += stats.QueriesExecuted
		totalErrors += stats.ErrorsCount
		totalDuration += stats.TotalDuration
	}

	avgResponseTime := time.Duration(0)
	if totalQueries > 0 {
		avgResponseTime = totalDuration / time.Duration(totalQueries)
	}

	return map[string]interface{}{
		"workers_count":     len(wm.workers),
		"total_queries":     totalQueries,
		"total_errors":      totalErrors,
		"error_rate":        float64(totalErrors) / float64(totalQueries) * 100,
		"avg_response_time": avgResponseTime,
		"queries_per_sec":   float64(totalQueries) / totalDuration.Seconds(),
	}
}

// PrintWorkerStats imprime estatísticas detalhadas dos workers
func (wm *WorkerManager) PrintWorkerStats() {
	fmt.Println("\n📊 Estatísticas Detalhadas dos Workers:")
	fmt.Println("─────────────────────────────────────────")

	for _, worker := range wm.workers {
		stats := worker.GetStats()
		fmt.Printf("Worker %d:\n", worker.ID)
		fmt.Printf("  Queries: %d | Erros: %d | Tempo Médio: %v\n",
			stats.QueriesExecuted, stats.ErrorsCount, stats.AvgResponseTime)
		fmt.Printf("  Última Query: %v\n", stats.LastQueryTime.Format("15:04:05"))
		fmt.Println()
	}
}
