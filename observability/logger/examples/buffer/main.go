package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"

	// Importa providers para registrar automaticamente
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
)

func main() {
	fmt.Println("=== Demonstração do Sistema de Buffer ===")

	// Configuração com buffer habilitado
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		BufferConfig: &interfaces.BufferConfig{
			Enabled:      true,
			Size:         100,             // Buffer de 100 entradas
			BatchSize:    10,              // Flush a cada 10 entradas
			FlushTimeout: 2 * time.Second, // Flush automático a cada 2 segundos
			AutoFlush:    true,
			MemoryLimit:  1024 * 1024, // 1MB limit
		},
		ServiceName: "buffer-demo",
		Environment: "development",
	}

	// Configura provider slog com buffer
	err := logger.SetProvider("slog", config)
	if err != nil {
		panic(err)
	}

	lgr := logger.GetCurrentProvider()
	defer lgr.Close()

	ctx := context.Background()

	fmt.Println("\n1. Testando escrita com buffer")

	// Escreve algumas entradas - devem ficar no buffer
	for i := 0; i < 5; i++ {
		lgr.Info(ctx, "Mensagem no buffer",
			logger.Int("index", i),
			logger.String("status", "buffered"))
	}

	fmt.Println("   - 5 entradas adicionadas ao buffer")

	// Verifica estatísticas do buffer
	if provider, ok := lgr.(interface{ GetBufferStats() interfaces.BufferStats }); ok {
		stats := provider.GetBufferStats()
		fmt.Printf("   - Buffer stats: %d/%d entradas, %d bytes\n",
			stats.UsedSize, stats.BufferSize, stats.MemoryUsage)
	}

	fmt.Println("\n2. Testando flush automático por batch size")

	// Adiciona mais entradas para triggerar flush por batch size
	for i := 5; i < 15; i++ {
		lgr.Info(ctx, "Triggering batch flush",
			logger.Int("index", i),
			logger.Bool("batch_trigger", true))
	}

	// Espera um pouco para flush automático
	time.Sleep(100 * time.Millisecond)

	if provider, ok := lgr.(interface{ GetBufferStats() interfaces.BufferStats }); ok {
		stats := provider.GetBufferStats()
		fmt.Printf("   - Após batch flush: %d/%d entradas, flushes: %d\n",
			stats.UsedSize, stats.BufferSize, stats.FlushCount)
	}

	fmt.Println("\n3. Testando flush manual")

	// Adiciona mais algumas entradas
	for i := 0; i < 3; i++ {
		lgr.Warn(ctx, "Entrada para flush manual",
			logger.Int("manual_index", i),
			logger.String("type", "manual_flush"))
	}

	// Flush manual
	if provider, ok := lgr.(interface{ FlushBuffer() error }); ok {
		err := provider.FlushBuffer()
		if err != nil {
			fmt.Printf("   - Erro no flush manual: %v\n", err)
		} else {
			fmt.Println("   - Flush manual executado com sucesso")
		}
	}

	fmt.Println("\n4. Testando flush por timeout")

	// Adiciona uma entrada e espera timeout
	lgr.Error(ctx, "Entrada para timeout flush",
		logger.String("trigger", "timeout"))

	fmt.Println("   - Esperando timeout de 2 segundos...")
	time.Sleep(time.Duration(2.5 * float64(time.Second)))

	if provider, ok := lgr.(interface{ GetBufferStats() interfaces.BufferStats }); ok {
		stats := provider.GetBufferStats()
		fmt.Printf("   - Após timeout: %d entradas, flushes: %d\n",
			stats.UsedSize, stats.FlushCount)
	}

	fmt.Println("\n5. Testando alta carga")

	start := time.Now()

	// Simula alta carga
	for i := 0; i < 100; i++ {
		lgr.Info(ctx, "High load test",
			logger.Int("load_index", i),
			logger.Time("timestamp", time.Now()),
			logger.Duration("elapsed", time.Since(start)))
	}

	elapsed := time.Since(start)
	fmt.Printf("   - 100 entradas em %v (%.2f entradas/ms)\n",
		elapsed, float64(100)/float64(elapsed.Milliseconds()))

	// Estatísticas finais
	if provider, ok := lgr.(interface{ GetBufferStats() interfaces.BufferStats }); ok {
		stats := provider.GetBufferStats()
		fmt.Printf("\n=== Estatísticas Finais ===\n")
		fmt.Printf("Total de entradas: %d\n", stats.TotalEntries)
		fmt.Printf("Entradas perdidas: %d\n", stats.DroppedEntries)
		fmt.Printf("Total de flushes: %d\n", stats.FlushCount)
		fmt.Printf("Tamanho do buffer: %d/%d\n", stats.UsedSize, stats.BufferSize)
		fmt.Printf("Uso de memória: %d bytes\n", stats.MemoryUsage)
		fmt.Printf("Último flush: %v\n", stats.LastFlush.Format(time.RFC3339))
		fmt.Printf("Duração do último flush: %v\n", stats.FlushDuration)
	}

	fmt.Println("\n=== Demo finalizada ===")
}
