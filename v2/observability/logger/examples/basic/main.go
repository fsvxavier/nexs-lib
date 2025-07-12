// Package main demonstra o uso básico do sistema de logging v2
// Este exemplo mostra as funcionalidades fundamentais: níveis de log,
// configuração simples e uso direto dos métodos de logging.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func main() {
	fmt.Println("=== Logger v2 - Exemplo Básico ===")

	// 1. Configuração básica com valores padrão
	config := logger.DefaultConfig()
	config.ServiceName = "basic-example"
	config.ServiceVersion = "1.0.0"
	config.Format = interfaces.ConsoleFormat
	config.Level = interfaces.DebugLevel

	// 2. Criação da factory e registro do provider
	factory := logger.NewFactory()

	// Registrar providers padrão
	factory.RegisterDefaultProviders()

	// 3. Criação do logger
	logger, err := factory.CreateLogger("basic", config)
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}

	ctx := context.Background()

	// 4. Demonstração de diferentes níveis de log
	fmt.Println("\n--- Níveis de Log ---")
	logger.Debug(ctx, "Mensagem de debug para desenvolvimento")
	logger.Info(ctx, "Aplicação iniciada com sucesso")
	logger.Warn(ctx, "Esta é uma advertência")
	logger.Error(ctx, "Erro simulado para demonstração")

	// 5. Logging com formatação
	fmt.Println("\n--- Logging com Formatação ---")
	userID := 12345
	operation := "login"
	logger.Infof(ctx, "Usuário %d executou operação: %s", userID, operation)
	logger.Debugf(ctx, "Processando %d registros", 100)

	// 6. Verificação de nível
	fmt.Println("\n--- Verificação de Nível ---")
	if logger.IsLevelEnabled(interfaces.DebugLevel) {
		fmt.Println("Debug está habilitado")
		logger.Debug(ctx, "Esta mensagem será processada")
	}

	if !logger.IsLevelEnabled(interfaces.TraceLevel) {
		fmt.Println("Trace não está habilitado")
	}

	// 7. Mudança dinâmica de nível
	fmt.Println("\n--- Mudança Dinâmica de Nível ---")
	logger.SetLevel(interfaces.WarnLevel)
	fmt.Printf("Nível atual: %s\n", logger.GetLevel().String())

	logger.Info(ctx, "Esta mensagem NÃO será exibida (nível muito baixo)")
	logger.Warn(ctx, "Esta mensagem SERÁ exibida")
	logger.Error(ctx, "Esta mensagem SERÁ exibida")

	// 8. Cleanup
	if err := logger.Flush(); err != nil {
		fmt.Printf("Erro ao fazer flush: %v\n", err)
	}

	if err := logger.Close(); err != nil {
		fmt.Printf("Erro ao fechar logger: %v\n", err)
	}

	fmt.Println("\n=== Exemplo Básico Concluído ===")
}
