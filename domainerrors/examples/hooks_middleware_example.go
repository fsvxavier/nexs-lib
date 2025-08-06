package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Exemplo de Hooks e Middlewares para Domain Errors ===\n")

	// 1. Registrando hooks
	fmt.Println("1. Registrando hooks...")

	// Hook para logging de metadados
	domainerrors.RegisterHook("before_metadata", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		log.Printf("[HOOK] Antes de adicionar metadados ao erro: %s", err.Code)
		return nil
	})

	domainerrors.RegisterHook("after_metadata", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		log.Printf("[HOOK] Após adicionar metadados ao erro: %s (Total metadados: %d)",
			err.Code, len(err.Metadata))
		return nil
	})

	// Hook para auditoria
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		log.Printf("[AUDIT] Erro criado - Código: %s, Tipo: %s, Timestamp: %s",
			err.Code, err.Type, err.Timestamp.Format(time.RFC3339))
		return nil
	})

	// 2. Registrando middlewares
	fmt.Println("2. Registrando middlewares...")

	// Middleware de enriquecimento
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		log.Printf("[MIDDLEWARE] Enriquecendo erro: %s", err.Code)

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}
		err.Metadata["enriched_at"] = time.Now()
		err.Metadata["service"] = "example-service"
		err.Metadata["version"] = "1.0.0"

		return next(err)
	})

	// Middleware de validação
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		log.Printf("[MIDDLEWARE] Validando erro: %s", err.Code)

		// Valida se o erro tem código e mensagem
		if err.Code == "" {
			log.Println("[MIDDLEWARE] WARN: Erro sem código!")
		}
		if err.Message == "" {
			log.Println("[MIDDLEWARE] WARN: Erro sem mensagem!")
		}

		return next(err)
	})

	// Middleware de logging
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		start := time.Now()
		log.Printf("[MIDDLEWARE] Processando erro: %s", err.Code)

		result := next(err)

		duration := time.Since(start)
		log.Printf("[MIDDLEWARE] Erro processado em %v", duration)

		return result
	})

	fmt.Println("\n3. Criando erros de exemplo...")

	// 3. Criando diferentes tipos de erros

	// Erro de validação
	fmt.Println("\n--- Erro de Validação ---")
	validationErr := domainerrors.NewWithType("VALIDATION_001", "Campo obrigatório não informado", domainerrors.ErrorTypeValidation)
	validationErr.WithMetadata("field", "email")
	validationErr.WithMetadata("value", "")

	fmt.Printf("Erro criado: %s\n", validationErr.Error())
	fmt.Printf("Metadados: %+v\n", validationErr.Metadata)

	// Erro de negócio
	fmt.Println("\n--- Erro de Negócio ---")
	businessErr := domainerrors.NewWithType("BUSINESS_001", "Saldo insuficiente para realizar a operação", domainerrors.ErrorTypeBusinessRule)
	businessErr.WithMetadata("account_id", "12345")
	businessErr.WithMetadata("requested_amount", 1500.00)
	businessErr.WithMetadata("current_balance", 800.00)

	fmt.Printf("Erro criado: %s\n", businessErr.Error())
	fmt.Printf("Status HTTP: %d\n", businessErr.HTTPStatus())

	// Erro com contexto
	fmt.Println("\n--- Erro com Contexto ---")
	ctx := context.WithValue(context.Background(), "user_id", "user123")
	ctx = context.WithValue(ctx, "trace_id", "trace456")

	contextErr := domainerrors.New("CONTEXT_001", "Erro com contexto personalizado")
	contextErr.WithContext(ctx, "Contexto adicionado ao erro")
	contextErr.WithMetadata("operation", "transfer_funds")
	contextErr.WithMetadata("amount", 250.50)

	fmt.Printf("Erro criado: %s\n", contextErr.Error())
	fmt.Printf("Stack trace: %s\n", contextErr.StackTrace())

	// 4. Demonstrando encadeamento de erros
	fmt.Println("\n--- Encadeamento de Erros ---")

	originalErr := fmt.Errorf("conexão com banco de dados falhou")
	wrappedErr := domainerrors.NewWithCause("DB_CONNECTION_001", "Falha ao conectar com o banco de dados", originalErr)
	wrappedErr.WithMetadata("database", "users_db")
	wrappedErr.WithMetadata("host", "localhost:5432")

	fmt.Printf("Erro encadeado: %s\n", wrappedErr.Error())
	fmt.Printf("Causa original: %s\n", wrappedErr.Unwrap().Error())

	fmt.Println("\n5. Resumo final...")
	fmt.Printf("Total de tipos de erro suportados: %d\n", 29) // Baseado nas constantes definidas
	fmt.Println("Hooks e middlewares integrados com sucesso!")

	fmt.Println("\n=== Exemplo concluído ===")
}
