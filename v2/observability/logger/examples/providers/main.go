// Package main demonstra o uso de diferentes providers de logging
// Este exemplo mostra como usar Zap, Slog e Zerolog com suas
// características específicas e configurações otimizadas.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func main() {
	fmt.Println("=== Logger v2 - Providers Comparison ===")

	ctx := context.Background()
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	// 1. Demonstração do provider Zap (Ultra-high performance)
	fmt.Println("\n--- Zap Provider (Ultra Performance) ---")
	demonstrateZapProvider(ctx, factory)

	// 2. Demonstração do provider Slog (Standard Library)
	fmt.Println("\n--- Slog Provider (Standard Library) ---")
	demonstrateSlogProvider(ctx, factory)

	// 3. Demonstração do provider Zerolog (Zero Allocation)
	fmt.Println("\n--- Zerolog Provider (Zero Allocation) ---")
	demonstrateZerologProvider(ctx, factory)

	// 4. Comparação de performance entre providers
	fmt.Println("\n--- Performance Comparison ---")
	compareProviderPerformance(factory)

	// 5. Configurações específicas por provider
	fmt.Println("\n--- Provider-Specific Configurations ---")
	demonstrateProviderConfigurations(factory)

	// 6. Hot swapping de providers
	fmt.Println("\n--- Hot Swapping Providers ---")
	demonstrateHotSwapping(factory)

	fmt.Println("\n=== Providers Comparison Concluído ===")
}

// demonstrateZapProvider mostra características específicas do Zap
func demonstrateZapProvider(ctx context.Context, factory *logger.Factory) {
	// Configuração otimizada para Zap
	config := logger.ProductionConfig()
	config.ServiceName = "zap-example"
	config.Level = interfaces.DebugLevel

	// Configuração assíncrona para máxima performance
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    8192,
		FlushInterval: 100 * time.Millisecond,
		Workers:       2,
		DropOnFull:    false,
	}

	// Criação específica do provider Zap
	zapProvider, err := factory.CreateProvider("zap", config)
	if err != nil {
		log.Printf("Erro ao criar provider Zap: %v", err)
		return
	}

	fmt.Printf("Provider: %s v%s\n", zapProvider.Name(), zapProvider.Version())

	// Health check do provider
	if err := zapProvider.HealthCheck(); err != nil {
		log.Printf("Health check falhou: %v", err)
		return
	}

	// Zap é otimizado para structured logging com tipos específicos
	zapProvider.Info(ctx, "Zap - Logging estruturado otimizado",
		interfaces.String("provider", "zap"),
		interfaces.Bool("high_performance", true),
		interfaces.Float64("benchmark_score", 9.8),
		interfaces.Duration("init_time", 15*time.Millisecond),
	)

	// Zap com sampling para alta frequência
	for i := 0; i < 100; i++ {
		zapProvider.Debug(ctx, "High frequency debug log",
			interfaces.Int("iteration", i),
			interfaces.Time("timestamp", time.Now()),
		)
	}

	zapProvider.Info(ctx, "Zap - 100 logs de alta frequência processados")
}

// demonstrateSlogProvider mostra características do Slog (Go standard)
func demonstrateSlogProvider(ctx context.Context, factory *logger.Factory) {
	// Configuração para Slog
	config := logger.DefaultConfig()
	config.ServiceName = "slog-example"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.InfoLevel

	// Criação específica do provider Slog
	slogProvider, err := factory.CreateProvider("slog", config)
	if err != nil {
		log.Printf("Erro ao criar provider Slog: %v", err)
		return
	}

	fmt.Printf("Provider: %s v%s\n", slogProvider.Name(), slogProvider.Version())

	// Slog é a implementação padrão do Go 1.21+
	slogProvider.Info(ctx, "Slog - Standard library logging",
		interfaces.String("provider", "slog"),
		interfaces.Bool("standard_library", true),
		interfaces.String("go_version", "1.21+"),
		interfaces.Bool("production_ready", true),
	)

	// Demonstração de diferentes níveis
	slogProvider.Debug(ctx, "Debug level - desenvolvimento")
	slogProvider.Info(ctx, "Info level - informacional")
	slogProvider.Warn(ctx, "Warn level - advertência")
	slogProvider.Error(ctx, "Error level - erro controlado")

	// Slog com contexto enriquecido
	enrichedLogger := slogProvider.WithFields(
		interfaces.String("component", "business-logic"),
		interfaces.String("operation", "user-registration"),
	)

	enrichedLogger.Info(ctx, "Usuário registrado com sucesso",
		interfaces.String("user_id", "user_789"),
		interfaces.String("email", "user@domain.com"),
		interfaces.Bool("email_verified", false),
	)
}

// demonstrateZerologProvider mostra características do Zerolog
func demonstrateZerologProvider(ctx context.Context, factory *logger.Factory) {
	// Configuração para Zerolog (zero allocation)
	config := logger.DefaultConfig()
	config.ServiceName = "zerolog-example"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.DebugLevel

	// Zerolog é otimizado para zero allocations
	config.BufferSize = 4096 // Buffer menor para demonstrar eficiência

	// Criação específica do provider Zerolog
	zerologProvider, err := factory.CreateProvider("zerolog", config)
	if err != nil {
		log.Printf("Erro ao criar provider Zerolog: %v", err)
		return
	}

	fmt.Printf("Provider: %s v%s\n", zerologProvider.Name(), zerologProvider.Version())

	// Zerolog é conhecido por zero allocations e alta performance
	zerologProvider.Info(ctx, "Zerolog - Zero allocation logging",
		interfaces.String("provider", "zerolog"),
		interfaces.Bool("zero_allocation", true),
		interfaces.Float64("memory_efficiency", 10.0),
		interfaces.String("specialty", "high_frequency_logging"),
	)

	// Teste de stress para demonstrar eficiência de memória
	start := time.Now()
	for i := 0; i < 1000; i++ {
		zerologProvider.Info(ctx, "High frequency message",
			interfaces.Int("message_id", i),
			interfaces.String("data", fmt.Sprintf("payload_%d", i)),
			interfaces.Bool("processed", true),
		)
	}
	duration := time.Since(start)

	zerologProvider.Info(ctx, "Stress test concluído",
		interfaces.Int("messages_logged", 1000),
		interfaces.Duration("total_time", duration),
		interfaces.Float64("messages_per_second", 1000.0/duration.Seconds()),
	)
}

// compareProviderPerformance compara a performance entre providers
func compareProviderPerformance(factory *logger.Factory) {
	const iterations = 5000
	ctx := context.Background()

	providers := []string{"zap", "slog", "zerolog"}
	results := make(map[string]time.Duration)

	for _, providerName := range providers {
		config := logger.DefaultConfig()
		config.ServiceName = fmt.Sprintf("%s-benchmark", providerName)
		config.Level = interfaces.InfoLevel

		provider, err := factory.CreateProvider(providerName, config)
		if err != nil {
			log.Printf("Erro ao criar provider %s: %v", providerName, err)
			continue
		}

		// Benchmark
		start := time.Now()
		for i := 0; i < iterations; i++ {
			provider.Info(ctx, "Benchmark message",
				interfaces.String("provider", providerName),
				interfaces.Int("iteration", i),
				interfaces.Time("timestamp", time.Now()),
				interfaces.Bool("benchmark", true),
			)
		}
		provider.Flush()
		duration := time.Since(start)
		results[providerName] = duration

		fmt.Printf("%s: %d logs em %v (%.2f logs/s)\n",
			providerName,
			iterations,
			duration,
			float64(iterations)/duration.Seconds(),
		)
	}

	// Encontra o mais rápido
	var fastest string
	var fastestTime time.Duration = time.Hour
	for provider, duration := range results {
		if duration < fastestTime {
			fastestTime = duration
			fastest = provider
		}
	}

	fmt.Printf("\nProvider mais rápido: %s (%v)\n", fastest, fastestTime)
}

// demonstrateProviderConfigurations mostra configurações específicas
func demonstrateProviderConfigurations(factory *logger.Factory) {
	ctx := context.Background()

	// Configuração específica para desenvolvimento (Slog + Console)
	devConfig := logger.DefaultConfig()
	devConfig.ServiceName = "dev-app"
	devConfig.Format = interfaces.ConsoleFormat
	devConfig.Level = interfaces.DebugLevel
	devConfig.AddCaller = true
	devConfig.AddSource = true

	devLogger, _ := factory.CreateProvider("slog", devConfig)
	devLogger.Info(ctx, "Configuração de desenvolvimento ativa")

	// Configuração específica para produção (Zap + JSON + Async)
	prodConfig := logger.ProductionConfig()
	prodConfig.ServiceName = "prod-app"
	prodConfig.Format = interfaces.JSONFormat
	prodConfig.Level = interfaces.InfoLevel
	prodConfig.EnableMetrics = true

	prodLogger, _ := factory.CreateProvider("zap", prodConfig)
	prodLogger.Info(ctx, "Configuração de produção ativa",
		interfaces.String("environment", "production"),
		interfaces.Bool("metrics_enabled", true),
	)

	// Configuração específica para alta frequência (Zerolog)
	highFreqConfig := logger.DefaultConfig()
	highFreqConfig.ServiceName = "high-freq-app"
	highFreqConfig.Format = interfaces.JSONFormat
	highFreqConfig.Level = interfaces.WarnLevel // Apenas warnings e errors
	highFreqConfig.BufferSize = 16384           // Buffer grande
	highFreqConfig.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    1000,
		Thereafter: 100,
		Tick:       10 * time.Second,
	}

	highFreqLogger, _ := factory.CreateProvider("zerolog", highFreqConfig)
	highFreqLogger.Warn(ctx, "Configuração de alta frequência ativa",
		interfaces.Bool("sampling_enabled", true),
		interfaces.Int("buffer_size", 16384),
	)
}

// demonstrateHotSwapping mostra troca de provider em tempo de execução
func demonstrateHotSwapping(factory *logger.Factory) {
	ctx := context.Background()

	// Cria logger inicial com Slog
	config := logger.DefaultConfig()
	config.ServiceName = "hot-swap-demo"

	currentLogger, _ := factory.CreateProvider("slog", config)
	currentLogger.Info(ctx, "Iniciado com provider Slog")

	// Simula necessidade de trocar para Zap para melhor performance
	time.Sleep(100 * time.Millisecond)

	// Flush do logger atual
	currentLogger.Flush()

	// Cria novo logger com Zap
	zapConfig := logger.ProductionConfig()
	zapConfig.ServiceName = "hot-swap-demo"

	newLogger, _ := factory.CreateProvider("zap", zapConfig)
	newLogger.Info(ctx, "Trocado para provider Zap para melhor performance")

	// Simula operação crítica que precisa de máxima performance
	start := time.Now()
	for i := 0; i < 1000; i++ {
		newLogger.Debug(ctx, "Operação crítica",
			interfaces.Int("operation_id", i),
		)
	}
	duration := time.Since(start)

	newLogger.Info(ctx, "Operação crítica concluída",
		interfaces.Duration("total_time", duration),
		interfaces.String("provider_used", "zap"),
	)

	// Cleanup
	currentLogger.Close()
	newLogger.Close()
}
