package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Domain Errors - Hooks and Middleware Examples ===")
	fmt.Println()

	// Example 1: Basic hook registration and usage
	basicHooksExample()

	// Example 2: Basic middleware chain
	basicMiddlewareExample()

	// Example 3: Complex middleware chain with transformations
	complexMiddlewareExample()

	// Example 4: Hooks and middleware working together
	combinedHooksMiddlewareExample()

	// Example 5: Real-world logging and audit scenario
	loggingAuditExample()

	// Example 6: Error enrichment pipeline
	errorEnrichmentExample()

	// Example 7: Custom patterns demonstration
	customPatternsExample()

	fmt.Println("=== Examples Complete! ===")
	fmt.Println("Note: In production, you would typically set up hooks and middlewares once during application startup.")
	fmt.Println()

	fmt.Println("Now running advanced production examples...")
	fmt.Println()
	runAdvancedExamples()
}

// Example 1: Basic hook registration and usage
func basicHooksExample() {
	fmt.Println("--- Example 1: Basic Hook Registration ---")

	// Register a simple after_error hook for logging
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		fmt.Printf("🎣 Hook executado: Erro '%s' criado com tipo '%s'\n", err.Code, err.Type)
		return nil
	})

	// Create an error - hook will be executed automatically
	err := domainerrors.New("USER_001", "Usuário não encontrado")
	fmt.Printf("Erro criado: %s\n", err.Error())

	fmt.Println()
}

// Example 2: Basic middleware chain
func basicMiddlewareExample() {
	fmt.Println("--- Example 2: Basic Middleware Chain ---")

	// Clear previous middlewares to start fresh for this example
	clearMiddlewaresManually()

	// Register a simple enrichment middleware
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware: Enriquecendo erro com informações do sistema")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["service"] = "user-service"
		err.Metadata["version"] = "1.0.0"
		err.Metadata["timestamp"] = time.Now().Format(time.RFC3339)

		return next(err)
	})

	// Create error - middleware will be executed automatically
	err := domainerrors.NewWithType("AUTH_001", "Falha na autenticação", domainerrors.ErrorTypeAuthentication)
	fmt.Printf("Erro enriquecido: %+v\n", err.Metadata)

	fmt.Println()
}

// Example 3: Complex middleware chain with transformations
func complexMiddlewareExample() {
	fmt.Println("--- Example 3: Complex Middleware Chain ---")

	// Clear previous middlewares
	clearMiddlewaresManually()

	// Middleware 1: Validation
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 1: Validando estrutura do erro")

		if err.Code == "" {
			fmt.Println("  ⚠️  Erro sem código detectado")
		}
		if err.Message == "" {
			fmt.Println("  ⚠️  Erro sem mensagem detectado")
		}

		return next(err)
	})

	// Middleware 2: Context enrichment
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 2: Enriquecendo contexto")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["environment"] = "production"
		err.Metadata["request_id"] = "req-" + fmt.Sprintf("%d", time.Now().UnixNano())

		return next(err)
	})

	// Middleware 3: Logging
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 3: Registrando erro no sistema de logs")

		logEntry := map[string]interface{}{
			"code":     err.Code,
			"type":     err.Type,
			"message":  err.Message,
			"metadata": err.Metadata,
		}

		logJSON, _ := json.Marshal(logEntry)
		fmt.Printf("  📝 Log: %s\n", string(logJSON))

		return next(err)
	})

	// Create error - all middlewares will execute in chain
	businessErr := domainerrors.NewWithType("BUSINESS_001", "Saldo insuficiente", domainerrors.ErrorTypeBusinessRule)
	fmt.Printf("Erro final processado: %s\n", businessErr.Code)

	fmt.Println()
}

// Example 4: Hooks and middleware working together
func combinedHooksMiddlewareExample() {
	fmt.Println("--- Example 4: Hooks and Middleware Working Together ---")

	// Clear previous middlewares
	clearMiddlewaresManually()
	clearHooksManually()

	// Register hooks
	domainerrors.RegisterHook("before_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		fmt.Println("🎣 Hook: ANTES da criação do erro")
		return nil
	})

	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		fmt.Println("🎣 Hook: APÓS criação do erro")
		return nil
	})

	domainerrors.RegisterHook("before_metadata", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		fmt.Println("🎣 Hook: ANTES de adicionar metadados")
		return nil
	})

	domainerrors.RegisterHook("after_metadata", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		fmt.Println("🎣 Hook: APÓS adicionar metadados")
		return nil
	})

	// Register middleware
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware: Processando erro na cadeia")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}
		err.Metadata["processed_by"] = "combined_example"

		return next(err)
	})

	// Create error and add metadata - observe the execution order
	fmt.Println("Criando erro...")
	err := domainerrors.New("COMBINED_001", "Exemplo combinado de hooks e middleware")

	fmt.Println("Adicionando metadados...")
	err.WithMetadata("example", "combined_hooks_middleware")

	fmt.Printf("Resultado final: %s com metadados %+v\n", err.Code, err.Metadata)

	fmt.Println()
}

// Example 5: Real-world logging and audit scenario
func loggingAuditExample() {
	fmt.Println("--- Example 5: Real-world Logging and Audit ---")

	// Clear previous state
	clearMiddlewaresManually()
	clearHooksManually()

	// Create a simple logger
	logger := log.New(os.Stdout, "[AUDIT] ", log.LstdFlags)

	// Audit hook for security-related errors
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		if err.Type == domainerrors.ErrorTypeAuthentication || err.Type == domainerrors.ErrorTypeAuthorization {
			logger.Printf("SECURITY ALERT: %s - %s", err.Code, err.Message)
		}
		return nil
	})

	// Metrics middleware
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["metric_timestamp"] = time.Now().Unix()
		err.Metadata["metric_type"] = "domain_error"
		err.Metadata["severity"] = getSeverityLevel(err.Type)

		fmt.Printf("📊 Métricas registradas para erro tipo: %s\n", err.Type)

		return next(err)
	})

	// Simulate security errors
	fmt.Println("Criando erro de autenticação...")
	authErr := domainerrors.NewWithType("AUTH_FAILED", "Credenciais inválidas", domainerrors.ErrorTypeAuthentication)
	authErr.WithMetadata("user_ip", "192.168.1.100")
	authErr.WithMetadata("user_agent", "curl/7.68.0")

	fmt.Println("Criando erro de autorização...")
	permissionErr := domainerrors.NewWithType("ACCESS_DENIED", "Acesso negado ao recurso", domainerrors.ErrorTypeAuthorization)
	permissionErr.WithMetadata("user_id", "user123")
	permissionErr.WithMetadata("resource", "/admin/users")

	fmt.Printf("Erros de segurança processados com auditoria completa\n")

	fmt.Println()
}

// Helper function for severity levels
func getSeverityLevel(errorType domainerrors.ErrorType) string {
	switch errorType {
	case domainerrors.ErrorTypeAuthentication, domainerrors.ErrorTypeAuthorization, domainerrors.ErrorTypeSecurity:
		return "high"
	case domainerrors.ErrorTypeServer, domainerrors.ErrorTypeDatabase:
		return "critical"
	case domainerrors.ErrorTypeValidation, domainerrors.ErrorTypeNotFound:
		return "low"
	default:
		return "medium"
	}
}

// Example 6: Error enrichment pipeline
func errorEnrichmentExample() {
	fmt.Println("--- Example 6: Error Enrichment Pipeline ---")

	// Clear previous state
	clearMiddlewaresManually()

	// Middleware 1: Add request context
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["request_context"] = map[string]interface{}{
			"method": "POST",
			"path":   "/api/v1/users",
			"ip":     "10.0.0.1",
		}

		fmt.Println("🔧 Pipeline Step 1: Request context added")
		return next(err)
	})

	// Middleware 2: Add user context
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}
		err.Metadata["user_context"] = map[string]interface{}{
			"user_id": "user_12345",
			"role":    "admin",
		}

		fmt.Println("🔧 Pipeline Step 2: User context added")
		return next(err)
	})

	// Middleware 3: Add system information
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["system_info"] = map[string]interface{}{
			"hostname":    "api-server-01",
			"version":     "2.1.0",
			"environment": "production",
			"region":      "us-east-1",
		}

		fmt.Println("🔧 Pipeline Step 3: System information added")
		return next(err)
	})

	// The error will go through the enrichment pipeline
	enrichedErr := domainerrors.NewWithType("DATA_VALIDATION", "Dados inválidos fornecidos", domainerrors.ErrorTypeValidation)

	// Print the enriched error
	enrichedJSON, _ := json.MarshalIndent(enrichedErr.Metadata, "", "  ")
	fmt.Printf("Erro enriquecido com pipeline completo:\n%s\n", string(enrichedJSON))

	fmt.Println()
}

// Example 7: Custom patterns demonstration
func customPatternsExample() {
	fmt.Println("--- Example 7: Custom Patterns ---")

	// Clear previous state
	clearMiddlewaresManually()
	clearHooksManually()

	// Pattern 1: Circuit breaker middleware
	var failureCount int
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Type == domainerrors.ErrorTypeExternalService || err.Type == domainerrors.ErrorTypeTimeout {
			failureCount++
			fmt.Printf("🔧 Circuit Breaker: Failure count increased to %d\n", failureCount)

			if failureCount >= 3 {
				fmt.Println("🔧 Circuit Breaker: OPENED - Too many failures!")
			}
		}

		return next(err)
	})

	// Pattern 2: Error transformation hook
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		// Transform internal errors to user-friendly messages
		if err.Type == domainerrors.ErrorTypeDatabase {
			fmt.Println("🎣 Hook: Transforming technical error to user-friendly message")
			originalMessage := err.Message
			err.Message = "Serviço temporariamente indisponível. Tente novamente em alguns instantes."
			if err.Metadata == nil {
				err.Metadata = make(map[string]interface{})
			}
			err.Metadata["original_message"] = originalMessage
		}
		return nil
	})

	// Pattern 3: Rate limiting hook
	lastErrorTime := time.Now().Add(-time.Minute) // Initialize in the past
	domainerrors.RegisterHook("before_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		if time.Since(lastErrorTime) < time.Second {
			fmt.Println("🎣 Hook: Rate limiting - Similar errors too frequent!")
		}
		lastErrorTime = time.Now()
		return nil
	})

	// Test the patterns
	fmt.Println("\nTesting custom patterns:")

	// Test circuit breaker
	for i := 1; i <= 4; i++ {
		fmt.Printf("Creating timeout error #%d\n", i)
		_ = domainerrors.NewWithType(fmt.Sprintf("TIMEOUT_%d", i), "Service timeout", domainerrors.ErrorTypeTimeout)
	}

	// Test error transformation
	fmt.Println("\nTesting error transformation:")
	dbErr := domainerrors.NewWithType("DB_CONNECTION", "Connection pool exhausted", domainerrors.ErrorTypeDatabase)
	fmt.Printf("Database error message after transformation: %s\n", dbErr.Message)
	fmt.Printf("Original message stored in metadata: %v\n", dbErr.Metadata["original_message"])

	// Test rate limiting
	fmt.Println("\nTesting rate limiting:")
	for i := 1; i <= 3; i++ {
		_ = domainerrors.New(fmt.Sprintf("RATE_%d", i), "Rate limited")
		time.Sleep(500 * time.Millisecond) // Small delay between errors
	}

	fmt.Println()

	fmt.Println("=== Custom Patterns Complete ===")
}

// Helper functions to simulate clearing since the API doesn't expose these directly
func clearMiddlewaresManually() {
	// Since we don't have a clear function, we'll work with what we have
	// In a real application, you might implement this or design around it
	fmt.Println("Note: Middlewares from previous examples may still be active")
}

func clearHooksManually() {
	// Since we don't have a clear function, we'll work with what we have
	// In a real application, you might implement this or design around it
	fmt.Println("Note: Hooks from previous examples may still be active")
}
