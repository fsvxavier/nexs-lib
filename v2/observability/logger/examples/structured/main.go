// Package main demonstra o uso de structured logging com campos tipados
// Este exemplo mostra como usar campos estruturados para criar logs
// ricos em contexto e facilmente consultáveis.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func main() {
	fmt.Println("=== Logger v2 - Structured Logging ===")

	// 1. Configuração para JSON estruturado
	config := logger.DefaultConfig()
	config.ServiceName = "structured-example"
	config.ServiceVersion = "1.0.0"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.DebugLevel

	// 2. Criação da factory e logger
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	logger, err := factory.CreateLogger("structured", config)
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}

	ctx := context.Background()

	// 3. Logging com campos estruturados básicos
	fmt.Println("\n--- Campos Estruturados Básicos ---")
	logger.Info(ctx, "Usuário autenticado",
		interfaces.String("user_id", "user123"),
		interfaces.Int("age", 30),
		interfaces.String("email", "user@example.com"),
		interfaces.Bool("is_admin", false),
	)

	// 4. Logging de operações comerciais
	fmt.Println("\n--- Operações Comerciais ---")
	logger.Info(ctx, "Pedido criado",
		interfaces.String("order_id", "ord_789"),
		interfaces.String("customer_id", "cust_456"),
		interfaces.Float64("total_amount", 299.99),
		interfaces.String("currency", "BRL"),
		interfaces.Int("item_count", 3),
		interfaces.Time("created_at", time.Now()),
	)

	// 5. Logging de métricas de performance
	fmt.Println("\n--- Métricas de Performance ---")
	duration := 150 * time.Millisecond
	logger.Info(ctx, "Operação de banco de dados concluída",
		interfaces.String("operation", "SELECT"),
		interfaces.String("table", "users"),
		interfaces.Duration("duration", duration),
		interfaces.Int("rows_affected", 25),
		interfaces.Bool("cache_hit", false),
	)

	// 6. Logging de erros com contexto
	fmt.Println("\n--- Erros com Contexto ---")
	businessErr := errors.New("saldo insuficiente")
	logger.Error(ctx, "Falha no processamento do pagamento",
		interfaces.String("payment_id", "pay_123"),
		interfaces.String("user_id", "user789"),
		interfaces.Float64("amount", 150.00),
		interfaces.Float64("balance", 75.50),
		interfaces.ErrorNamed("error", businessErr),
		interfaces.String("error_code", "INSUFFICIENT_FUNDS"),
	)

	// 7. Logging com arrays e objetos
	fmt.Println("\n--- Arrays e Objetos ---")
	tags := []string{"premium", "verified", "vip"}
	userProfile := map[string]interface{}{
		"name":     "João Silva",
		"country":  "Brazil",
		"timezone": "America/Sao_Paulo",
	}

	logger.Info(ctx, "Perfil de usuário atualizado",
		interfaces.String("user_id", "user456"),
		interfaces.Array("tags", tags),
		interfaces.Object("profile", userProfile),
	)

	// 8. Uso do WithFields para contexto comum
	fmt.Println("\n--- Logger com Campos Comuns ---")
	requestLogger := logger.WithFields(
		interfaces.String("request_id", "req_abc123"),
		interfaces.String("method", "POST"),
		interfaces.String("endpoint", "/api/users"),
		interfaces.String("client_ip", "192.168.1.100"),
	)

	requestLogger.Info(ctx, "Requisição recebida")
	requestLogger.Debug(ctx, "Validando dados de entrada",
		interfaces.Int("payload_size", 1024),
	)
	requestLogger.Info(ctx, "Processamento concluído",
		interfaces.Int("status_code", 201),
		interfaces.Duration("response_time", 45*time.Millisecond),
	)

	// 9. Logging hierárquico para microserviços
	fmt.Println("\n--- Logging Hierárquico ---")
	serviceLogger := logger.WithFields(
		interfaces.String("service", "user-service"),
		interfaces.String("version", "v2.1.0"),
		interfaces.String("environment", "production"),
	)

	componentLogger := serviceLogger.WithFields(
		interfaces.String("component", "authentication"),
		interfaces.String("module", "jwt-validator"),
	)

	componentLogger.Info(ctx, "Token JWT validado",
		interfaces.String("token_type", "bearer"),
		interfaces.Int64("expires_in", 3600),
		interfaces.String("scope", "read write"),
	)

	// 10. Cleanup
	if err := logger.Flush(); err != nil {
		fmt.Printf("Erro ao fazer flush: %v\n", err)
	}

	if err := logger.Close(); err != nil {
		fmt.Printf("Erro ao fechar logger: %v\n", err)
	}

	fmt.Println("\n=== Structured Logging Concluído ===")
}
