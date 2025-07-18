package main

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/observability/logger"

	// Importa os providers para auto-registração
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
	fmt.Println("=== Teste de Provider Padrão ===")

	// Contexto simples
	ctx := context.Background()

	// Verifica qual provider está sendo usado
	providers := logger.ListProviders()
	currentProvider := logger.GetCurrentProviderName()
	fmt.Printf("Providers disponíveis: %v\n", providers)
	fmt.Printf("Provider atual (padrão): %s\n", currentProvider)
	fmt.Println()

	// Usa o logger sem configurar explicitamente (deve usar zap como padrão)
	logger.Info(ctx, "Mensagem de teste com provider padrão")

	// Teste com campos estruturados
	logger.Info(ctx, "Teste com campos estruturados",
		logger.String("provider", currentProvider),
		logger.Bool("zap_default", currentProvider == "zap"),
		logger.Int("test_number", 42),
	)

	// Teste com diferentes níveis
	logger.Debug(ctx, "Mensagem de debug (pode não aparecer se nível for INFO)")
	logger.Warn(ctx, "Mensagem de warning")
	logger.Error(ctx, "Mensagem de erro")

	fmt.Println("\n=== Teste concluído ===")
}
