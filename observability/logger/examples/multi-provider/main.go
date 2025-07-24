package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	// Importa os providers para auto-registração
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
	fmt.Println("=== Demonstração Multi-Provider ===")
	fmt.Println()

	// Cria contexto com informações de rastreamento
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-12345")
	ctx = context.WithValue(ctx, logger.SpanIDKey, "span-67890")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-456")

	// Configurações para cada provider
	configs := map[string]*logger.Config{
		"slog": {
			Level:          logger.InfoLevel,
			Format:         logger.JSONFormat,
			Output:         os.Stdout,
			AddSource:      false,
			AddStacktrace:  false,
			TimeFormat:     time.RFC3339,
			ServiceName:    "demo-app",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			Fields: map[string]any{
				"component": "demo",
			},
		},
		"zap": {
			Level:          logger.InfoLevel,
			Format:         logger.JSONFormat,
			Output:         os.Stdout,
			AddSource:      true,
			AddStacktrace:  true,
			TimeFormat:     time.RFC3339,
			ServiceName:    "demo-app",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			Fields: map[string]any{
				"component": "demo",
			},
		},
		"zerolog": {
			Level:          logger.InfoLevel,
			Format:         logger.JSONFormat,
			Output:         os.Stdout,
			AddSource:      false,
			AddStacktrace:  false,
			TimeFormat:     time.RFC3339,
			ServiceName:    "demo-app",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			Fields: map[string]any{
				"component": "demo",
			},
		},
	}

	// Demonstra cada provider
	providers := []string{"slog", "zap", "zerolog"}
	for _, providerName := range providers {
		fmt.Printf("=== Provider: %s ===\n", providerName)

		// Configura o provider
		err := logger.ConfigureProvider(providerName, configs[providerName])
		if err != nil {
			fmt.Printf("Erro ao configurar provider %s: %v\n", providerName, err)
			continue
		}

		// Alterna para o provider
		err = logger.SetActiveProvider(providerName)
		if err != nil {
			fmt.Printf("Erro ao definir provider %s: %v\n", providerName, err)
			continue
		}

		// Testa diferentes níveis de log
		logger.Info(ctx, "Informação básica")
		logger.Warn(ctx, "Aviso importante")
		logger.Error(ctx, "Erro simulado")

		// Testa com campos estruturados
		logger.Info(ctx, "Informação com campos",
			logger.String("operation", "test"),
			logger.Int("attempt", 1),
			logger.Bool("success", true),
			logger.Duration("elapsed", 100*time.Millisecond),
		)

		// Testa logs formatados
		logger.Infof(ctx, "Log formatado: %s = %d", "count", 42)

		// Testa com código de erro
		logger.ErrorWithCode(ctx, "E001", "Erro com código específico",
			logger.String("details", "Detalhes adicionais"),
		)

		// Testa WithFields
		contextLogger := logger.WithFields(
			logger.String("module", "auth"),
			logger.String("action", "login"),
		)
		contextLogger.Info(ctx, "Log com contexto pré-definido")

		// Testa WithContext
		ctxLogger := logger.WithContext(ctx)
		ctxLogger.Info(context.Background(), "Log com contexto extraído")

		fmt.Println()
	}

	fmt.Println("=== Demonstração de Benchmark Completo ===")
	fmt.Println()

	// Testa performance básica
	benchmarkProviders(ctx, configs)
}

// BenchmarkResult armazena resultados de benchmark
type BenchmarkResult struct {
	Provider        string
	TestName        string
	Iterations      int
	TotalTime       time.Duration
	TimePerLog      time.Duration
	LogsPerSecond   float64
	MemoryAllocated uint64
}

func benchmarkProviders(ctx context.Context, configs map[string]*logger.Config) {
	providers := []string{"slog", "zap", "zerolog"}
	results := []BenchmarkResult{}

	// Diferentes cenários de benchmark
	benchmarks := map[string]struct {
		iterations int
		testFunc   func(context.Context, string, int) time.Duration
	}{
		"Logs Simples": {
			iterations: 1000,
			testFunc:   benchmarkSimpleLogs,
		},
		"Logs com Campos": {
			iterations: 1000,
			testFunc:   benchmarkStructuredLogs,
		},
		"Logs com Contexto": {
			iterations: 500,
			testFunc:   benchmarkContextLogs,
		},
		"Logs com Erros": {
			iterations: 500,
			testFunc:   benchmarkErrorLogs,
		},
		"Logs Formatados": {
			iterations: 500,
			testFunc:   benchmarkFormattedLogs,
		},
	}

	for testName, benchmark := range benchmarks {
		fmt.Printf("=== Benchmark: %s ===\n", testName)
		fmt.Printf("%-10s | %-12s | %-12s | %-15s | %-12s\n",
			"Provider", "Iterações", "Tempo Total", "Tempo/Log", "Logs/Seg")
		fmt.Println(strings.Repeat("-", 75))

		for _, providerName := range providers {
			// Configura o provider
			err := logger.ConfigureProvider(providerName, configs[providerName])
			if err != nil {
				fmt.Printf("Erro ao configurar provider %s: %v\n", providerName, err)
				continue
			}

			err = logger.SetActiveProvider(providerName)
			if err != nil {
				fmt.Printf("Erro ao definir provider %s: %v\n", providerName, err)
				continue
			}

			// Executa benchmark
			elapsed := benchmark.testFunc(ctx, providerName, benchmark.iterations)
			timePerLog := elapsed / time.Duration(benchmark.iterations)
			logsPerSecond := float64(benchmark.iterations) / elapsed.Seconds()

			// Armazena resultado
			result := BenchmarkResult{
				Provider:      providerName,
				TestName:      testName,
				Iterations:    benchmark.iterations,
				TotalTime:     elapsed,
				TimePerLog:    timePerLog,
				LogsPerSecond: logsPerSecond,
			}
			results = append(results, result)

			// Exibe resultado
			fmt.Printf("%-10s | %-12d | %-12v | %-15v | %-12.0f\n",
				providerName, benchmark.iterations, elapsed, timePerLog, logsPerSecond)
		}
		fmt.Println()
	}

	// Exibe resumo comparativo
	fmt.Println("=== Resumo Comparativo ===")
	displayBenchmarkSummary(results)
}

func benchmarkSimpleLogs(ctx context.Context, provider string, iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(ctx, "Mensagem de benchmark simples")
	}
	return time.Since(start)
}

func benchmarkStructuredLogs(ctx context.Context, provider string, iterations int) time.Duration {
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
	return time.Since(start)
}

func benchmarkContextLogs(ctx context.Context, provider string, iterations int) time.Duration {
	// Cria contexto com múltiplos valores
	enrichedCtx := context.WithValue(ctx, logger.TraceIDKey, "trace-benchmark-123")
	enrichedCtx = context.WithValue(enrichedCtx, logger.SpanIDKey, "span-benchmark-456")
	enrichedCtx = context.WithValue(enrichedCtx, logger.UserIDKey, "user-benchmark-789")
	enrichedCtx = context.WithValue(enrichedCtx, logger.RequestIDKey, "req-benchmark-101")

	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Info(enrichedCtx, "Mensagem com contexto enriquecido",
			logger.String("operation", "benchmark"),
			logger.Int("step", i),
		)
	}
	return time.Since(start)
}

func benchmarkErrorLogs(ctx context.Context, provider string, iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Error(ctx, "Erro de benchmark",
			logger.String("provider", provider),
			logger.Int("iteration", i),
			logger.String("error_type", "benchmark_error"),
			logger.String("details", "Detalhes do erro de benchmark"),
		)
	}
	return time.Since(start)
}

func benchmarkFormattedLogs(ctx context.Context, provider string, iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		logger.Infof(ctx, "Mensagem formatada %s: iteração %d de %d", provider, i, iterations)
	}
	return time.Since(start)
}

func displayBenchmarkSummary(results []BenchmarkResult) {
	// Agrupa por tipo de teste
	testGroups := make(map[string][]BenchmarkResult)
	for _, result := range results {
		testGroups[result.TestName] = append(testGroups[result.TestName], result)
	}

	for testName, group := range testGroups {
		fmt.Printf("\n--- %s ---\n", testName)

		// Encontra o mais rápido
		fastest := group[0]
		for _, result := range group {
			if result.LogsPerSecond > fastest.LogsPerSecond {
				fastest = result
			}
		}

		// Exibe comparação
		fmt.Printf("🏆 Mais rápido: %s (%.0f logs/seg)\n", fastest.Provider, fastest.LogsPerSecond)

		for _, result := range group {
			if result.Provider != fastest.Provider {
				ratio := fastest.LogsPerSecond / result.LogsPerSecond
				fmt.Printf("📊 %s: %.0f logs/seg (%.1fx mais lento)\n",
					result.Provider, result.LogsPerSecond, ratio)
			}
		}
	}

	// Ranking geral
	fmt.Println("\n=== Ranking Geral de Performance ===")
	providerStats := make(map[string]struct {
		totalLogs float64
		testCount int
	})

	for _, result := range results {
		stats := providerStats[result.Provider]
		stats.totalLogs += result.LogsPerSecond
		stats.testCount++
		providerStats[result.Provider] = stats
	}

	type ProviderRank struct {
		Provider string
		AvgLogs  float64
	}

	var ranking []ProviderRank
	for provider, stats := range providerStats {
		ranking = append(ranking, ProviderRank{
			Provider: provider,
			AvgLogs:  stats.totalLogs / float64(stats.testCount),
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
		fmt.Printf("%s %d. %s: %.0f logs/seg (média)\n",
			medal, i+1, rank.Provider, rank.AvgLogs)
	}
}
