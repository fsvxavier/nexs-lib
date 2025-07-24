package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"

	// Importa todos os providers para auto-registração
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
	fmt.Println("=== Exemplo Básico Multi-Provider ===")
	fmt.Println()

	// Cria contexto com informações de rastreamento
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

	// Demonstra cada provider
	providers := []string{"slog", "zap", "zerolog"}
	for _, providerName := range providers {
		fmt.Printf("=== Provider: %s ===\n", providerName)

		// Configuração básica
		config := logger.DefaultConfig()
		config.Level = logger.InfoLevel
		config.Format = logger.JSONFormat
		config.ServiceName = "logger-example"
		config.ServiceVersion = "1.0.0"
		config.Environment = "development"
		config.AddSource = true
		config.Fields = map[string]any{
			"component": "example",
		}

		// Configura o provider
		err := logger.ConfigureProvider(providerName, config)
		if err != nil {
			fmt.Printf("Erro ao configurar provider %s: %v\n", providerName, err)
			continue
		}

		// Define como provider ativo
		err = logger.SetActiveProvider(providerName)
		if err != nil {
			fmt.Printf("Erro ao definir provider %s: %v\n", providerName, err)
			continue
		}

		// Logs básicos
		logger.Info(ctx, "Aplicação iniciada",
			logger.String("status", "starting"),
			logger.Int("port", 8080),
		)

		// Log formatado
		logger.Infof(ctx, "Servidor iniciado na porta %d", 8080)

		// Log com contexto
		logger.Info(ctx, "Processando requisição",
			logger.String("method", "GET"),
			logger.String("path", "/api/users"),
		)

		// Log com campos estruturados
		logger.Info(ctx, "Executando consulta",
			logger.String("module", "database"),
			logger.String("operation", "query"),
			logger.Duration("duration", 150*time.Millisecond),
			logger.Int("rows_affected", 5),
		)

		// Log de erro
		logger.Error(ctx, "Erro ao conectar com o banco",
			logger.String("error", "connection timeout"),
			logger.String("host", "localhost:5432"),
			logger.String("stacktrace", "stack_trace"),
		)

		// Log de warning
		logger.Warn(ctx, "Warning message",
			logger.String("warning_type", "performance"),
		)

		// Log com código
		logger.InfoWithCode(ctx, "USER_CREATED", "Usuário criado com sucesso",
			logger.String("user_id", "123"),
			logger.String("email", "user@example.com"),
		)

		// Logger com contexto pré-definido
		contextLogger := logger.WithFields(
			logger.String("module", "auth"),
			logger.String("operation", "login"),
		)
		contextLogger.Info(ctx, "Tentativa de login",
			logger.String("user_id", "456"),
			logger.Bool("success", true),
		)

		fmt.Println()
	}

	fmt.Println("=== Providers Disponíveis ===")
	for _, provider := range logger.ListProviders() {
		fmt.Printf("- %s\n", provider)
	}
	fmt.Println()

	fmt.Println("=== Comparação de Formatos ===")

	// Configuração para formato console
	consoleConfig := logger.DefaultConfig()
	consoleConfig.Format = logger.ConsoleFormat
	consoleConfig.ServiceName = "logger-example"
	consoleConfig.Environment = "development"
	consoleConfig.AddSource = true

	// Exemplo com formato console (slog)
	fmt.Println("--- Formato Console (slog) ---")
	logger.ConfigureProvider("slog", consoleConfig)
	logger.SetActiveProvider("slog")
	logger.Info(ctx, "Log em formato console",
		logger.String("key", "value"),
		logger.Int("number", 42),
	)

	// Exemplo com formato JSON (zerolog)
	jsonConfig := logger.DefaultConfig()
	jsonConfig.Format = logger.JSONFormat
	jsonConfig.ServiceName = "logger-example"
	jsonConfig.Environment = "development"

	fmt.Println("--- Formato JSON (zerolog) ---")
	logger.ConfigureProvider("zerolog", jsonConfig)
	logger.SetActiveProvider("zerolog")
	logger.Info(ctx, "Log em formato JSON",
		logger.String("key", "value"),
		logger.Int("number", 42),
	)

	fmt.Println()
	fmt.Println("=== Exemplo Concluído ===")
}
