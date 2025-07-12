// Package main demonstra o uso de logging assíncrono para alta performance
// Este exemplo mostra como configurar e usar o sistema de logging assíncrono
// com pools de workers e buffers para aplicações de alta escala.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func main() {
	fmt.Println("=== Logger v2 - Async Logging ===")

	// 1. Configuração para logging assíncrono
	config := logger.DefaultConfig()
	config.ServiceName = "async-example"
	config.ServiceVersion = "1.0.0"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.DebugLevel

	// Configuração assíncrona para alta performance
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    8192,                  // Buffer grande para alta throughput
		FlushInterval: 50 * time.Millisecond, // Flush frequente
		Workers:       4,                     // Múltiplos workers
		DropOnFull:    false,                 // Não descartar logs quando buffer estiver cheio
	}

	// Configuração de sampling para controle de volume
	config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    100,                                       // Primeiros 100 logs passam
		Thereafter: 10,                                        // Depois, 1 a cada 10
		Tick:       1 * time.Second,                           // Reset a cada segundo
		Levels:     []interfaces.Level{interfaces.DebugLevel}, // Aplica sampling apenas para DEBUG
	}

	// 2. Criação da factory e logger
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	asyncLogger, err := factory.CreateLogger("async", config)
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}

	ctx := context.Background()

	// 3. Demonstração de performance básica
	fmt.Println("\n--- Performance Básica ---")
	measurePerformance("Logging Síncrono", func() {
		// Simula logging síncrono (para comparação)
		syncConfig := logger.DefaultConfig()
		syncConfig.ServiceName = "sync-example"
		syncLogger, _ := factory.CreateLogger("sync", syncConfig)

		for i := 0; i < 1000; i++ {
			syncLogger.Info(ctx, "Mensagem síncrona",
				interfaces.Int("iteration", i),
				interfaces.String("type", "sync"),
			)
		}
		syncLogger.Flush()
	})

	measurePerformance("Logging Assíncrono", func() {
		for i := 0; i < 1000; i++ {
			asyncLogger.Info(ctx, "Mensagem assíncrona",
				interfaces.Int("iteration", i),
				interfaces.String("type", "async"),
			)
		}
		asyncLogger.Flush()
	})

	// 4. Teste de carga intensiva
	fmt.Println("\n--- Teste de Carga Intensiva ---")
	loadTestConcurrent(asyncLogger)

	// 5. Demonstração de sampling
	fmt.Println("\n--- Demonstração de Sampling ---")
	demonstrateSampling(asyncLogger)

	// 6. Monitoramento de buffer
	fmt.Println("\n--- Monitoramento de Performance ---")
	monitorPerformance(asyncLogger)

	// 7. Teste de failover
	fmt.Println("\n--- Teste de Failover ---")
	testFailover(asyncLogger)

	// 8. Cleanup
	fmt.Println("\n--- Finalizando ---")
	if err := asyncLogger.Flush(); err != nil {
		fmt.Printf("Erro ao fazer flush: %v\n", err)
	}

	if err := asyncLogger.Close(); err != nil {
		fmt.Printf("Erro ao fechar logger: %v\n", err)
	}

	fmt.Println("\n=== Async Logging Concluído ===")
}

// measurePerformance mede o tempo de execução de uma função
func measurePerformance(name string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	fmt.Printf("%s: %v\n", name, duration)
}

// loadTestConcurrent executa teste de carga com múltiplas goroutines
func loadTestConcurrent(logger interfaces.Logger) {
	const (
		numGoroutines    = 10
		logsPerGoroutine = 500
	)

	ctx := context.Background()
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			goroutineLogger := logger.WithFields(
				interfaces.Int("goroutine_id", goroutineID),
				interfaces.String("operation", "load_test"),
			)

			for j := 0; j < logsPerGoroutine; j++ {
				goroutineLogger.Info(ctx, "Mensagem de carga",
					interfaces.Int("message_id", j),
					interfaces.Time("timestamp", time.Now()),
					interfaces.String("data", generateTestData()),
				)

				// Mix de níveis para teste
				if j%10 == 0 {
					goroutineLogger.Debug(ctx, "Debug message",
						interfaces.Int("debug_id", j),
					)
				}
				if j%50 == 0 {
					goroutineLogger.Warn(ctx, "Warning message",
						interfaces.Int("warn_id", j),
					)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalLogs := numGoroutines * logsPerGoroutine
	throughput := float64(totalLogs) / duration.Seconds()

	fmt.Printf("Teste de carga concluído:\n")
	fmt.Printf("  Goroutines: %d\n", numGoroutines)
	fmt.Printf("  Logs por goroutine: %d\n", logsPerGoroutine)
	fmt.Printf("  Total de logs: %d\n", totalLogs)
	fmt.Printf("  Duração: %v\n", duration)
	fmt.Printf("  Throughput: %.2f logs/segundo\n", throughput)
}

// demonstrateSampling mostra o funcionamento do sampling
func demonstrateSampling(logger interfaces.Logger) {
	ctx := context.Background()

	fmt.Println("Gerando 200 logs DEBUG para demonstrar sampling...")

	for i := 0; i < 200; i++ {
		logger.Debug(ctx, "Debug message com sampling",
			interfaces.Int("sequence", i),
			interfaces.String("level", "debug"),
		)

		// Pequena pausa para simular processamento
		time.Sleep(1 * time.Millisecond)
	}

	fmt.Println("Logs DEBUG gerados (observe que apenas alguns passaram pelo sampling)")
}

// monitorPerformance monitora métricas de performance
func monitorPerformance(logger interfaces.Logger) {
	ctx := context.Background()

	// Simula operações de diferentes complexidades
	operations := []struct {
		name     string
		duration time.Duration
		fields   int
	}{
		{"fast_operation", 5 * time.Millisecond, 3},
		{"medium_operation", 25 * time.Millisecond, 8},
		{"slow_operation", 100 * time.Millisecond, 15},
	}

	for _, op := range operations {
		start := time.Now()

		// Simula operação
		time.Sleep(op.duration)

		actualDuration := time.Since(start)

		// Cria campos dinamicamente
		fields := []interfaces.Field{
			interfaces.String("operation", op.name),
			interfaces.Duration("duration", actualDuration),
			interfaces.Duration("expected_duration", op.duration),
			interfaces.Int("field_count", op.fields),
		}

		// Adiciona campos extras baseados na configuração
		for i := 0; i < op.fields; i++ {
			fields = append(fields, interfaces.String(
				fmt.Sprintf("field_%d", i),
				fmt.Sprintf("value_%d", i),
			))
		}

		logger.Info(ctx, "Operação monitorada", fields...)
	}
}

// testFailover testa o comportamento em situações de falha
func testFailover(logger interfaces.Logger) {
	ctx := context.Background()

	// Simula situação de alta carga que pode encher o buffer
	fmt.Println("Testando comportamento sob alta carga...")

	var wg sync.WaitGroup
	const burst = 5000

	// Burst de logs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < burst; i++ {
			logger.Error(ctx, "Burst message",
				interfaces.Int("burst_id", i),
				interfaces.String("data", generateLargeTestData()),
			)
		}
	}()

	// Logs normais continuam funcionando
	for i := 0; i < 10; i++ {
		logger.Info(ctx, "Normal operation durante burst",
			interfaces.Int("normal_id", i),
		)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("Teste de failover concluído")
}

// generateTestData gera dados de teste
func generateTestData() string {
	return fmt.Sprintf("test_data_%d", time.Now().UnixNano()%10000)
}

// generateLargeTestData gera dados de teste maiores
func generateLargeTestData() string {
	data := "large_data_"
	for i := 0; i < 50; i++ {
		data += fmt.Sprintf("chunk_%d_", i)
	}
	return data
}
