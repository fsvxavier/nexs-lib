package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog" // Descomente quando as dependências estiverem disponíveis
)

func main() {
	ctx := context.Background()

	// Exemplo usando Slog
	fmt.Println("=== Exemplo usando Slog ===")
	configSlog := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "isis-example",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		AddSource:      true,
		AddStacktrace:  true,
		Fields: map[string]any{
			"component": "example",
		},
	}

	err := logger.SetProvider("slog", configSlog)
	if err != nil {
		panic(err)
	}

	// Testes básicos
	logger.Info(ctx, "Aplicação iniciada",
		logger.String("status", "starting"),
		logger.Int("port", 8080),
	)

	logger.Infof(ctx, "Servidor iniciado na porta %d", 8080)

	// Teste com contexto
	ctxWithTrace := context.WithValue(ctx, "trace_id", "trace-123")
	ctxWithTrace = context.WithValue(ctxWithTrace, "user_id", "user-456")

	logger.WithContext(ctxWithTrace).Info(ctx, "Processando requisição",
		logger.String("method", "GET"),
		logger.String("path", "/api/users"),
	)

	// Teste com campos personalizados
	loggerWithFields := logger.WithFields(
		logger.String("module", "database"),
		logger.String("operation", "query"),
	)

	loggerWithFields.Info(ctx, "Executando consulta",
		logger.Duration("duration", 150*time.Millisecond),
		logger.Int("rows_affected", 5),
	)

	// Teste de erro
	logger.Error(ctx, "Erro ao conectar com o banco",
		logger.String("error", "connection timeout"),
		logger.String("host", "localhost:5432"),
	)

	// Exemplo usando Zap
	fmt.Println("\n=== Exemplo usando Zap ===")
	configZap := &logger.Config{
		Level:          logger.DebugLevel,
		Format:         logger.ConsoleFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "isis-example",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		AddSource:      true,
		AddStacktrace:  false,
		Fields: map[string]any{
			"component": "example",
		},
		SamplingConfig: &logger.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
			Tick:       time.Second,
		},
	}

	err = logger.SetProvider("zap", configZap)
	if err != nil {
		panic(err)
	}

	logger.Debug(ctx, "Debug message",
		logger.String("debug_info", "detailed information"),
	)

	logger.Info(ctx, "Application started with Zap",
		logger.String("provider", "zap"),
		logger.Bool("sampling", true),
	)

	// Teste performance logging
	start := time.Now()
	time.Sleep(10 * time.Millisecond) // Simula operação
	duration := time.Since(start)

	logger.Info(ctx, "Operação completada",
		logger.Duration("duration", duration),
		logger.String("status", "success"),
	)

	// Listar providers disponíveis
	fmt.Println("\n=== Providers Disponíveis ===")
	providers := logger.ListProviders()
	for _, provider := range providers {
		fmt.Printf("- %s\n", provider)
	}
}
