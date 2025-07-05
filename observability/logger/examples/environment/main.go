package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	// _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
)

func loadEnvFile(filename string) error {
	// Simples loader de .env (em produção, use uma biblioteca como godotenv)
	// Para este exemplo, vamos definir as variáveis manualmente

	if filename == ".env.development" {
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("LOG_FORMAT", "console")
		os.Setenv("LOG_ADD_SOURCE", "true")
		os.Setenv("LOG_ADD_STACKTRACE", "false")
		os.Setenv("LOG_TIME_FORMAT", "15:04:05.000")
		os.Setenv("LOG_PROVIDER", "slog")
		os.Setenv("SERVICE_NAME", "isis-example")
		os.Setenv("SERVICE_VERSION", "1.0.0")
		os.Setenv("ENVIRONMENT", "development")
	} else if filename == ".env.production" {
		os.Setenv("LOG_LEVEL", "info")
		os.Setenv("LOG_FORMAT", "json")
		os.Setenv("LOG_ADD_SOURCE", "false")
		os.Setenv("LOG_ADD_STACKTRACE", "true")
		os.Setenv("LOG_TIME_FORMAT", time.RFC3339)
		os.Setenv("LOG_PROVIDER", "zap")
		os.Setenv("SERVICE_NAME", "isis-api")
		os.Setenv("SERVICE_VERSION", "2.1.0")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("LOG_SAMPLING_INITIAL", "1000")
		os.Setenv("LOG_SAMPLING_THEREAFTER", "100")
		os.Setenv("LOG_SAMPLING_TICK", "1s")
	}

	return nil
}

func main() {
	fmt.Println("=== Exemplo com Configuração por Ambiente ===\n")

	// Exemplo 1: Configuração de desenvolvimento
	fmt.Println("1. Configuração de Desenvolvimento:")
	loadEnvFile(".env.development")

	err := logger.ConfigFromEnvironment()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	logger.Info(ctx, "Aplicação iniciada em modo desenvolvimento",
		logger.String("config_source", "environment"),
	)

	logger.Debug(ctx, "Informação de debug visível",
		logger.String("detail", "configuração de debug ativa"),
	)

	// Exemplo 2: Mudança para produção
	fmt.Println("\n2. Mudança para Configuração de Produção:")
	loadEnvFile(".env.production")

	err = logger.ConfigFromEnvironment()
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "Aplicação reconfigurada para produção",
		logger.String("config_source", "environment"),
		logger.String("format", "json"),
	)

	// Debug não aparecerá porque o nível está em INFO
	logger.Debug(ctx, "Esta mensagem de debug não aparecerá")

	logger.Info(ctx, "Processamento de dados em produção",
		logger.Int("records_processed", 1500),
		logger.Bool("success", true),
	)

	// Exemplo 3: Configurações programáticas
	fmt.Println("\n3. Configurações Programáticas:")

	// Desenvolvimento
	devConfig := logger.DevelopmentConfig()
	devConfig.ServiceName = "isis-dev"
	devConfig.ServiceVersion = "dev-branch"

	logger.SetProvider("slog", devConfig)

	logger.Info(ctx, "Usando configuração de desenvolvimento programática",
		logger.String("method", "programmatic"),
	)

	// Produção
	prodConfig := logger.ProductionConfig()
	prodConfig.ServiceName = "isis-prod"
	prodConfig.ServiceVersion = "v2.1.0"
	prodConfig.Fields = map[string]any{
		"datacenter": "us-east-1",
		"cluster":    "production-01",
	}

	logger.SetProvider("zap", prodConfig)

	logger.Info(ctx, "Usando configuração de produção programática",
		logger.String("method", "programmatic"),
		logger.String("provider", "zap"),
	)

	// Exemplo 4: Teste
	fmt.Println("\n4. Configuração de Teste:")

	testConfig := logger.TestingConfig()
	testConfig.ServiceName = "isis-test"

	logger.SetProvider("slog", testConfig)

	logger.Info(ctx, "Executando em modo de teste",
		logger.String("test_suite", "integration"),
		logger.Bool("verbose", false),
	)

	// Exemplo 5: Configuração automática
	fmt.Println("\n5. Configuração Automática:")

	// Simula diferentes ambientes
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		os.Setenv("ENVIRONMENT", env)
		os.Setenv("SERVICE_NAME", "isis-auto")

		// Configuração específica por ambiente
		switch env {
		case "development":
			os.Setenv("LOG_LEVEL", "debug")
			os.Setenv("LOG_FORMAT", "console")
		case "staging":
			os.Setenv("LOG_LEVEL", "info")
			os.Setenv("LOG_FORMAT", "json")
		case "production":
			os.Setenv("LOG_LEVEL", "warn")
			os.Setenv("LOG_FORMAT", "json")
			os.Setenv("LOG_SAMPLING_INITIAL", "500")
		}

		err := logger.ConfigFromEnvironment()
		if err != nil {
			fmt.Printf("Erro ao configurar ambiente %s: %v\n", env, err)
			continue
		}

		logger.Info(ctx, "Ambiente configurado automaticamente",
			logger.String("environment", env),
			logger.String("auto_configured", "true"),
		)
	}

	fmt.Println("\n=== Variáveis de Ambiente Suportadas ===")
	envVars := []string{
		"LOG_LEVEL (debug|info|warn|error|fatal|panic)",
		"LOG_FORMAT (json|console|text)",
		"LOG_PROVIDER (slog|zap|zerolog)",
		"LOG_ADD_SOURCE (true|false)",
		"LOG_ADD_STACKTRACE (true|false)",
		"LOG_TIME_FORMAT (layout string)",
		"SERVICE_NAME (string)",
		"SERVICE_VERSION (string)",
		"ENVIRONMENT (string)",
		"LOG_SAMPLING_INITIAL (int)",
		"LOG_SAMPLING_THEREAFTER (int)",
		"LOG_SAMPLING_TICK (duration)",
	}

	for _, envVar := range envVars {
		fmt.Printf("- %s\n", envVar)
	}

	fmt.Println("\n=== Exemplo Concluído ===")
}
