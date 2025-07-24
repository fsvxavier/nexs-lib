package main

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	// Importa os providers para auto-registração
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

// BenchmarkResult armazena resultados detalhados de benchmark
type BenchmarkResult struct {
	Provider           string
	TestName           string
	Iterations         int
	TotalTime          time.Duration
	TimePerLog         time.Duration
	LogsPerSecond      float64
	MemoryAllocsBefore uint64
	MemoryAllocsAfter  uint64
	MemoryAllocsDelta  uint64
	GCBefore           uint32
	GCAfter            uint32
	GCDelta            uint32
}

// BenchmarkSuite define um conjunto de testes de benchmark
type BenchmarkSuite struct {
	Name       string
	Iterations int
	TestFunc   func(context.Context, string, int) BenchmarkResult
}

func main() {
	fmt.Println("=== Benchmark Completo dos Providers de Logging ===")
	fmt.Println()

	// Informações do sistema
	printSystemInfo()

	// Configuração base para todos os providers (com output silencioso para benchmark)
	baseConfig := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         io.Discard, // Não escreve nada - apenas para benchmark
		AddSource:      false,
		AddStacktrace:  false,
		TimeFormat:     time.RFC3339,
		ServiceName:    "benchmark-app",
		ServiceVersion: "1.0.0",
		Environment:    "benchmark",
	}

	// Cria contexto com informações de rastreamento
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-benchmark-123")
	ctx = context.WithValue(ctx, logger.SpanIDKey, "span-benchmark-456")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-benchmark-789")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-benchmark-101")

	// Define os benchmarks
	benchmarks := []BenchmarkSuite{
		{
			Name:       "Logs Simples",
			Iterations: 10000,
			TestFunc:   benchmarkSimpleLogs,
		},
		{
			Name:       "Logs com Campos Estruturados",
			Iterations: 5000,
			TestFunc:   benchmarkStructuredLogs,
		},
		{
			Name:       "Logs com Contexto Rico",
			Iterations: 3000,
			TestFunc:   benchmarkContextLogs,
		},
		{
			Name:       "Logs de Erro",
			Iterations: 2000,
			TestFunc:   benchmarkErrorLogs,
		},
		{
			Name:       "Logs Formatados",
			Iterations: 2000,
			TestFunc:   benchmarkFormattedLogs,
		},
		{
			Name:       "Logs com Campos Complexos",
			Iterations: 1000,
			TestFunc:   benchmarkComplexLogs,
		},
	}

	providers := []string{"slog", "zap", "zerolog"}
	allResults := []BenchmarkResult{}

	// Executa benchmarks
	for _, benchmark := range benchmarks {
		fmt.Printf("🔥 Executando: %s (%d iterações)\n", benchmark.Name, benchmark.Iterations)
		fmt.Printf("%-10s | %-12s | %-12s | %-15s | %-12s | %-10s | %-8s\n",
			"Provider", "Iterações", "Tempo Total", "Tempo/Log", "Logs/Seg", "Mem Delta", "GC Delta")
		fmt.Println(strings.Repeat("-", 95))

		for _, provider := range providers {
			// Configura o provider
			err := logger.ConfigureProvider(provider, baseConfig)
			if err != nil {
				fmt.Printf("❌ Erro ao configurar %s: %v\n", provider, err)
				continue
			}

			err = logger.SetActiveProvider(provider)
			if err != nil {
				fmt.Printf("❌ Erro ao ativar %s: %v\n", provider, err)
				continue
			}

			// Executa benchmark
			result := benchmark.TestFunc(ctx, provider, benchmark.Iterations)
			allResults = append(allResults, result)

			// Exibe resultado
			fmt.Printf("%-10s | %-12d | %-12v | %-15v | %-12.0f | %-10d | %-8d\n",
				provider, result.Iterations, result.TotalTime, result.TimePerLog,
				result.LogsPerSecond, result.MemoryAllocsDelta, result.GCDelta)
		}
		fmt.Println()
	}

	// Análise comparativa
	fmt.Println("=== Análise Comparativa Completa ===")
	analyzeResults(allResults)

	// Recomendações
	fmt.Println("\n=== Recomendações de Uso ===")
	printRecommendations(allResults)
}

func printSystemInfo() {
	fmt.Printf("🖥️  Sistema: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("🔧 Go: %s\n", runtime.Version())
	fmt.Printf("⚙️  CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("💾 Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

func benchmarkSimpleLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	// Força garbage collection antes do teste
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(ctx, "Mensagem de benchmark simples")
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs Simples",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func benchmarkStructuredLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(ctx, "Mensagem com campos estruturados",
			logger.String("provider", provider),
			logger.Int("iteration", i),
			logger.Bool("benchmark", true),
			logger.Duration("elapsed", time.Duration(i)*time.Microsecond),
			logger.Float64("rate", 1.234),
		)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs com Campos Estruturados",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func benchmarkContextLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Cria contexto com múltiplos valores
	enrichedCtx := context.WithValue(ctx, logger.TraceIDKey, "trace-benchmark-context-123")
	enrichedCtx = context.WithValue(enrichedCtx, logger.SpanIDKey, "span-benchmark-context-456")
	enrichedCtx = context.WithValue(enrichedCtx, logger.UserIDKey, "user-benchmark-context-789")
	enrichedCtx = context.WithValue(enrichedCtx, logger.RequestIDKey, "req-benchmark-context-101")

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(enrichedCtx, "Mensagem com contexto enriquecido",
			logger.String("operation", "benchmark"),
			logger.Int("step", i),
			logger.String("component", "logger"),
		)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs com Contexto Rico",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func benchmarkErrorLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Error(ctx, "Erro de benchmark",
			logger.String("provider", provider),
			logger.Int("iteration", i),
			logger.String("error_type", "benchmark_error"),
			logger.String("details", "Detalhes do erro de benchmark"),
			logger.String("stack_trace", "main.go:123 -> benchmark.go:456"),
		)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs de Erro",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func benchmarkFormattedLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Infof(ctx, "Mensagem formatada %s: iteração %d de %d com taxa %.2f",
			provider, i, iterations, float64(i)/float64(iterations)*100)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs Formatados",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func benchmarkComplexLogs(ctx context.Context, provider string, iterations int) BenchmarkResult {
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(ctx, "Mensagem com campos complexos",
			logger.String("provider", provider),
			logger.Int("iteration", i),
			logger.Bool("benchmark", true),
			logger.Duration("elapsed", time.Duration(i)*time.Microsecond),
			logger.Float64("rate", 1.234),
			logger.Int64("timestamp", time.Now().UnixNano()),
			logger.String("user_agent", "Mozilla/5.0 (compatible; Benchmark/1.0)"),
			logger.String("ip_address", "192.168.1.100"),
			logger.String("method", "POST"),
			logger.String("url", "/api/v1/benchmark"),
			logger.Int("status_code", 200),
			logger.String("response_time", "123ms"),
			logger.String("request_id", fmt.Sprintf("req-%d", i)),
		)
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Provider:           provider,
		TestName:           "Logs com Campos Complexos",
		Iterations:         iterations,
		TotalTime:          elapsed,
		TimePerLog:         elapsed / time.Duration(iterations),
		LogsPerSecond:      float64(iterations) / elapsed.Seconds(),
		MemoryAllocsBefore: m1.Alloc,
		MemoryAllocsAfter:  m2.Alloc,
		MemoryAllocsDelta:  m2.Alloc - m1.Alloc,
		GCBefore:           m1.NumGC,
		GCAfter:            m2.NumGC,
		GCDelta:            m2.NumGC - m1.NumGC,
	}
}

func analyzeResults(results []BenchmarkResult) {
	// Agrupa por tipo de teste
	testGroups := make(map[string][]BenchmarkResult)
	for _, result := range results {
		testGroups[result.TestName] = append(testGroups[result.TestName], result)
	}

	// Análise por tipo de teste
	for testName, group := range testGroups {
		fmt.Printf("\n📊 %s:\n", testName)

		// Encontra o mais rápido
		fastest := group[0]
		for _, result := range group {
			if result.LogsPerSecond > fastest.LogsPerSecond {
				fastest = result
			}
		}

		// Encontra o mais eficiente em memória
		mostMemEfficient := group[0]
		for _, result := range group {
			if result.MemoryAllocsDelta < mostMemEfficient.MemoryAllocsDelta {
				mostMemEfficient = result
			}
		}

		fmt.Printf("  🏆 Mais rápido: %s (%.0f logs/seg)\n", fastest.Provider, fastest.LogsPerSecond)
		fmt.Printf("  💾 Mais eficiente em memória: %s (%d bytes)\n",
			mostMemEfficient.Provider, mostMemEfficient.MemoryAllocsDelta)

		// Comparação detalhada
		for _, result := range group {
			if result.Provider != fastest.Provider {
				speedRatio := fastest.LogsPerSecond / result.LogsPerSecond
				memRatio := float64(result.MemoryAllocsDelta) / float64(mostMemEfficient.MemoryAllocsDelta)
				fmt.Printf("  📈 %s: %.0f logs/seg (%.1fx mais lento), %d bytes (%.1fx mais memória)\n",
					result.Provider, result.LogsPerSecond, speedRatio, result.MemoryAllocsDelta, memRatio)
			}
		}
	}

	// Ranking geral
	fmt.Println("\n🏆 Ranking Geral de Performance:")
	providerStats := make(map[string]struct {
		totalLogs   float64
		totalMemory uint64
		totalGC     uint32
		testCount   int
	})

	for _, result := range results {
		stats := providerStats[result.Provider]
		stats.totalLogs += result.LogsPerSecond
		stats.totalMemory += result.MemoryAllocsDelta
		stats.totalGC += result.GCDelta
		stats.testCount++
		providerStats[result.Provider] = stats
	}

	type ProviderRank struct {
		Provider  string
		AvgLogs   float64
		AvgMemory uint64
		AvgGC     uint32
	}

	var ranking []ProviderRank
	for provider, stats := range providerStats {
		ranking = append(ranking, ProviderRank{
			Provider:  provider,
			AvgLogs:   stats.totalLogs / float64(stats.testCount),
			AvgMemory: stats.totalMemory / uint64(stats.testCount),
			AvgGC:     stats.totalGC / uint32(stats.testCount),
		})
	}

	// Ordena por performance
	for i := 0; i < len(ranking); i++ {
		for j := i + 1; j < len(ranking); j++ {
			if ranking[j].AvgLogs > ranking[i].AvgLogs {
				ranking[i], ranking[j] = ranking[j], ranking[i]
			}
		}
	}

	for i, rank := range ranking {
		medal := "🥉"
		if i == 0 {
			medal = "🥇"
		} else if i == 1 {
			medal = "🥈"
		}
		fmt.Printf("%s %d. %s: %.0f logs/seg, %d bytes/teste, %d GC/teste\n",
			medal, i+1, rank.Provider, rank.AvgLogs, rank.AvgMemory, rank.AvgGC)
	}
}

func printRecommendations(results []BenchmarkResult) {
	fmt.Println("💡 Performance Máxima:")
	fmt.Println("   - Use zap para aplicações de alta performance")
	fmt.Println("   - Evite logs formatados em hot paths")
	fmt.Println("   - Prefira campos estruturados over logs simples")
	fmt.Println()

	fmt.Println("⚖️  Balanceamento:")
	fmt.Println("   - Use slog para compatibilidade com padrões Go")
	fmt.Println("   - Boa performance com menos complexidade")
	fmt.Println("   - Ideal para aplicações corporativas")
	fmt.Println()

	fmt.Println("🔧 Funcionalidades:")
	fmt.Println("   - Use zerolog para logs JSON nativos")
	fmt.Println("   - Melhor para pipelines de processamento")
	fmt.Println("   - Boa integração com sistemas de observabilidade")
	fmt.Println()

	fmt.Println("🚀 Dicas de Otimização:")
	fmt.Println("   - Evite alocações desnecessárias")
	fmt.Println("   - Use sampling em ambientes de produção")
	fmt.Println("   - Configure níveis de log apropriados")
	fmt.Println("   - Monitore uso de memória e GC")
}
