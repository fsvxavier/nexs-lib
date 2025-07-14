package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

// PerformanceMonitor monitora performance das operações
type PerformanceMonitor struct {
	mu         sync.RWMutex
	operations map[string]*OperationStats
	startTime  time.Time
}

// OperationStats estatísticas de uma operação
type OperationStats struct {
	Name         string
	TotalRecords int
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	AvgTime      time.Duration
	Operations   int
}

// NewPerformanceMonitor cria um novo monitor de performance
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		operations: make(map[string]*OperationStats),
		startTime:  time.Now(),
	}
}

// AddOperation adiciona uma operação ao monitoramento
func (pm *PerformanceMonitor) AddOperation(name string, records int, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stats, exists := pm.operations[name]
	if !exists {
		stats = &OperationStats{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
		}
		pm.operations[name] = stats
	}

	stats.TotalRecords += records
	stats.TotalTime += duration
	stats.Operations++

	if duration < stats.MinTime {
		stats.MinTime = duration
	}
	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}

	stats.AvgTime = stats.TotalTime / time.Duration(stats.Operations)
}

// PrintSummary imprime resumo das estatísticas
func (pm *PerformanceMonitor) PrintSummary() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalElapsed := time.Since(pm.startTime)

	fmt.Printf("\n📊 Resumo de Performance (Tempo total: %v):\n", totalElapsed)
	fmt.Println("┌────────────────────────────┬─────────┬────────────┬─────────────┬─────────────┬─────────────┐")
	fmt.Println("│ Operação                   │ Records │ Tempo Total│ Tempo Médio │ Min         │ Max         │")
	fmt.Println("├────────────────────────────┼─────────┼────────────┼─────────────┼─────────────┼─────────────┤")

	totalRecords := 0
	totalTime := time.Duration(0)

	for _, stats := range pm.operations {
		totalRecords += stats.TotalRecords
		totalTime += stats.TotalTime

		fmt.Printf("│ %-26s │ %7d │ %10v │ %11v │ %11v │ %11v │\n",
			stats.Name, stats.TotalRecords, stats.TotalTime,
			stats.AvgTime, stats.MinTime, stats.MaxTime)
	}

	fmt.Println("├────────────────────────────┼─────────┼────────────┼─────────────┼─────────────┼─────────────┤")

	avgRate := float64(0)
	if totalTime > 0 {
		avgRate = float64(totalRecords) / totalTime.Seconds()
	}

	fmt.Printf("│ TOTAL                      │ %7d │ %10v │ %11.2f/s │             │             │\n",
		totalRecords, totalTime, avgRate)
	fmt.Println("└────────────────────────────┴─────────┴────────────┴─────────────┴─────────────┴─────────────┘")
}

// insertCustomersIndividual insere clientes individualmente (para comparação)
func insertCustomersIndividual(ctx context.Context, pool postgresql.IPool, customers []Customer) (time.Duration, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	for _, customer := range customers {
		if err := conn.Exec(ctx,
			"INSERT INTO batch_customers (name, email, phone, city, country) VALUES ($1, $2, $3, $4, $5)",
			customer.Name, customer.Email, customer.Phone, customer.City, customer.Country); err != nil {
			return 0, fmt.Errorf("erro ao inserir cliente: %w", err)
		}
	}

	duration := time.Since(start)
	fmt.Printf("   ✅ %d clientes inseridos individualmente\n", len(customers))

	return duration, nil
}

// insertCustomersBatchOptimized insere clientes usando batch otimizado
func insertCustomersBatchOptimized(ctx context.Context, pool postgresql.IPool, customers []Customer) (time.Duration, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	// Usar batch size otimizado
	batchSize := 500
	totalBatches := (len(customers) + batchSize - 1) / batchSize

	for i := 0; i < totalBatches; i++ {
		startIdx := i * batchSize
		endIdx := startIdx + batchSize
		if endIdx > len(customers) {
			endIdx = len(customers)
		}

		batch := &simpleBatch{}
		for j := startIdx; j < endIdx; j++ {
			customer := customers[j]
			batch.Queue(
				"INSERT INTO batch_customers (name, email, phone, city, country) VALUES ($1, $2, $3, $4, $5)",
				customer.Name, customer.Email, customer.Phone, customer.City, customer.Country)
		}

		results, err := conn.SendBatch(ctx, batch)
		if err != nil {
			return 0, fmt.Errorf("erro ao executar batch %d: %w", i+1, err)
		}

		for j := 0; j < batch.Len(); j++ {
			if err := results.Exec(); err != nil {
				results.Close()
				return 0, fmt.Errorf("erro no item %d do batch %d: %w", j, i+1, err)
			}
		}
		results.Close()
	}

	duration := time.Since(start)
	fmt.Printf("   ✅ %d clientes inseridos em %d batches\n", len(customers), totalBatches)

	return duration, nil
}

// insertCustomersTransaction insere clientes usando transação com prepared statement
func insertCustomersTransaction(ctx context.Context, pool postgresql.IPool, customers []Customer) (time.Duration, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	// Iniciar transação
	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Preparar statement
	stmtName := "insert_customer"
	if err := tx.Prepare(ctx, stmtName,
		"INSERT INTO batch_customers (name, email, phone, city, country) VALUES ($1, $2, $3, $4, $5)"); err != nil {
		_ = tx.Rollback(ctx)
		return 0, fmt.Errorf("erro ao preparar statement: %w", err)
	}

	// Executar inserções usando prepared statement
	for _, customer := range customers {
		if err := tx.Exec(ctx, stmtName,
			customer.Name, customer.Email, customer.Phone, customer.City, customer.Country); err != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("erro ao inserir cliente: %w", err)
		}
	}

	// Confirmar transação
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("   ✅ %d clientes inseridos com prepared statement\n", len(customers))

	return duration, nil
}

// MemoryUsageMonitor monitora uso de memória
type MemoryUsageMonitor struct {
	samples    []MemorySample
	mu         sync.RWMutex
	monitoring bool
	stop       chan bool
}

// MemorySample amostra de uso de memória
type MemorySample struct {
	Timestamp time.Time
	HeapAlloc uint64
	HeapSys   uint64
	NumGC     uint32
}

// NewMemoryUsageMonitor cria um novo monitor de memória
func NewMemoryUsageMonitor() *MemoryUsageMonitor {
	return &MemoryUsageMonitor{
		samples: make([]MemorySample, 0),
		stop:    make(chan bool),
	}
}

// StartMonitoring inicia o monitoramento de memória
func (m *MemoryUsageMonitor) StartMonitoring(interval time.Duration) {
	m.mu.Lock()
	m.monitoring = true
	m.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.stop:
				return
			case <-ticker.C:
				m.takeSample()
			}
		}
	}()
}

// StopMonitoring para o monitoramento
func (m *MemoryUsageMonitor) StopMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.monitoring {
		close(m.stop)
		m.monitoring = false
	}
}

// takeSample coleta uma amostra de memória
func (m *MemoryUsageMonitor) takeSample() {
	// Note: Em um exemplo real, usaríamos runtime.ReadMemStats()
	// Aqui simulamos os dados para evitar dependências extras
	sample := MemorySample{
		Timestamp: time.Now(),
		HeapAlloc: uint64(50 * 1024 * 1024),  // 50MB simulado
		HeapSys:   uint64(100 * 1024 * 1024), // 100MB simulado
		NumGC:     10,                        // Simulado
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, sample)
}

// GetPeakUsage retorna o pico de uso de memória
func (m *MemoryUsageMonitor) GetPeakUsage() MemorySample {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.samples) == 0 {
		return MemorySample{}
	}

	peak := m.samples[0]
	for _, sample := range m.samples {
		if sample.HeapAlloc > peak.HeapAlloc {
			peak = sample
		}
	}

	return peak
}

// PrintMemoryReport imprime relatório de memória
func (m *MemoryUsageMonitor) PrintMemoryReport() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.samples) == 0 {
		fmt.Println("📊 Nenhuma amostra de memória coletada")
		return
	}

	peak := m.GetPeakUsage()
	current := m.samples[len(m.samples)-1]

	fmt.Printf("\n💾 Relatório de Uso de Memória:\n")
	fmt.Printf("   Pico de uso: %.2f MB\n", float64(peak.HeapAlloc)/(1024*1024))
	fmt.Printf("   Uso atual: %.2f MB\n", float64(current.HeapAlloc)/(1024*1024))
	fmt.Printf("   Total de amostras: %d\n", len(m.samples))
}

// BenchmarkResult resultado de benchmark
type BenchmarkResult struct {
	Name          string
	TotalTime     time.Duration
	RecordsPerSec float64
	MemoryUsage   uint64
	ErrorRate     float64
}

// RunBenchmark executa um benchmark completo
func RunBenchmark(ctx context.Context, pool postgresql.IPool, name string,
	testFunc func(context.Context, postgresql.IPool, []Customer) (time.Duration, error),
	customers []Customer) BenchmarkResult {

	fmt.Printf("🏁 Executando benchmark: %s\n", name)

	// Iniciar monitoramento de memória
	memMonitor := NewMemoryUsageMonitor()
	memMonitor.StartMonitoring(100 * time.Millisecond)
	defer memMonitor.StopMonitoring()

	// Executar teste
	duration, err := testFunc(ctx, pool, customers)

	errorRate := 0.0
	if err != nil {
		errorRate = 100.0
		fmt.Printf("   ❌ Erro: %v\n", err)
	}

	// Coletar estatísticas de memória
	peak := memMonitor.GetPeakUsage()

	// Calcular taxa de registros por segundo
	recordsPerSec := float64(len(customers)) / duration.Seconds()

	result := BenchmarkResult{
		Name:          name,
		TotalTime:     duration,
		RecordsPerSec: recordsPerSec,
		MemoryUsage:   peak.HeapAlloc,
		ErrorRate:     errorRate,
	}

	fmt.Printf("   ⏱️ Tempo: %v\n", duration)
	fmt.Printf("   📊 Taxa: %.2f registros/segundo\n", recordsPerSec)
	fmt.Printf("   💾 Memória: %.2f MB\n", float64(peak.HeapAlloc)/(1024*1024))

	return result
}
