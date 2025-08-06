package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// ProductionLogger simula um logger estruturado para produção
type ProductionLogger struct {
	logger *log.Logger
}

func NewProductionLogger() *ProductionLogger {
	return &ProductionLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (pl *ProductionLogger) LogError(level string, code string, message string, metadata map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"error": map[string]interface{}{
			"code":     code,
			"message":  message,
			"metadata": metadata,
		},
	}

	logJSON, _ := json.Marshal(logEntry)
	pl.logger.Printf("[%s] %s", strings.ToUpper(level), string(logJSON))
}

// MetricsCollector simula um coletor de métricas
type MetricsCollector struct {
	errorCounts map[string]int
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		errorCounts: make(map[string]int),
	}
}

func (mc *MetricsCollector) IncrementErrorCount(errorType string) {
	mc.errorCounts[errorType]++
	fmt.Printf("📊 Metrics: Error count for '%s': %d\n", errorType, mc.errorCounts[errorType])
}

func (mc *MetricsCollector) GetErrorCounts() map[string]int {
	return mc.errorCounts
}

// ErrorNotifier simula um sistema de notificação
type ErrorNotifier struct {
	criticalThreshold int
}

func NewErrorNotifier(threshold int) *ErrorNotifier {
	return &ErrorNotifier{criticalThreshold: threshold}
}

func (en *ErrorNotifier) NotifyIfCritical(errorType domainerrors.ErrorType, metadata map[string]interface{}) {
	if en.isCriticalError(errorType) {
		fmt.Printf("🚨 ALERT: Critical error detected - Type: %s\n", errorType)
		fmt.Printf("🚨 ALERT: Metadata: %+v\n", metadata)

		// Em produção, enviaria para Slack, email, PagerDuty, etc.
		fmt.Println("🚨 ALERT: Notification sent to on-call team")
	}
}

func (en *ErrorNotifier) isCriticalError(errorType domainerrors.ErrorType) bool {
	criticalTypes := []domainerrors.ErrorType{
		domainerrors.ErrorTypeServer,
		domainerrors.ErrorTypeDatabase,
		domainerrors.ErrorTypeSecurity,
		domainerrors.ErrorTypeServiceUnavailable,
	}

	for _, ct := range criticalTypes {
		if errorType == ct {
			return true
		}
	}
	return false
}

// Instâncias globais (normalmente injetadas via DI)
var (
	prodLogger       = NewProductionLogger()
	metricsCollector = NewMetricsCollector()
	errorNotifier    = NewErrorNotifier(3)
)

func runAdvancedExamples() {
	fmt.Println("=== Advanced Hooks and Middleware - Production Patterns ===")
	fmt.Println()

	// Setup production-ready hooks and middleware
	setupProductionHooksAndMiddleware()

	// Example 1: E-commerce order processing with comprehensive error handling
	ecommerceOrderExample()

	// Example 2: Financial transaction with high security requirements
	financialTransactionExample()

	// Example 3: Microservice integration with circuit breaker
	microserviceIntegrationExample()

	// Example 4: API rate limiting and abuse detection
	apiRateLimitingExample()

	// Example 5: Database operation with retry and circuit breaker
	databaseOperationExample()

	// Show collected metrics
	showCollectedMetrics()

	fmt.Println("=== Advanced Examples Complete ===")
}

func setupProductionHooksAndMiddleware() {
	fmt.Println("--- Setting up Production Hooks and Middleware ---")

	// Hook 1: Structured logging for all errors
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		level := getLogLevel(err.Type)
		prodLogger.LogError(level, err.Code, err.Message, err.Metadata)
		return nil
	})

	// Hook 2: Metrics collection
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		metricsCollector.IncrementErrorCount(string(err.Type))
		return nil
	})

	// Hook 3: Critical error notifications
	domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
		errorNotifier.NotifyIfCritical(err.Type, err.Metadata)
		return nil
	})

	// Middleware 1: Request context enrichment
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		// Add request context if available
		if requestID := ctx.Value("request_id"); requestID != nil {
			err.Metadata["request_id"] = requestID
		}
		if userID := ctx.Value("user_id"); userID != nil {
			err.Metadata["user_id"] = userID
		}
		if sessionID := ctx.Value("session_id"); sessionID != nil {
			err.Metadata["session_id"] = sessionID
		}

		return next(err)
	})

	// Middleware 2: Environment and service information
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["service"] = "nexs-lib-demo"
		err.Metadata["version"] = "2.1.0"
		err.Metadata["environment"] = "production"
		err.Metadata["hostname"] = "api-server-03"
		err.Metadata["region"] = "us-east-1"
		err.Metadata["timestamp"] = time.Now().Unix()

		return next(err)
	})

	// Middleware 3: Security sanitization
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		// Remove sensitive information from error messages
		err.Message = sanitizeSensitiveData(err.Message)

		// Sanitize metadata
		if err.Metadata != nil {
			err.Metadata = sanitizeMetadata(err.Metadata)
		}

		return next(err)
	})

	fmt.Println("✅ Production hooks and middleware configured")
	fmt.Println()
}

func ecommerceOrderExample() {
	fmt.Println("--- Example 1: E-commerce Order Processing ---")

	// Simulate processing an order with various potential errors
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req_order_12345")
	ctx = context.WithValue(ctx, "user_id", "user_789")
	ctx = context.WithValue(ctx, "session_id", "sess_abc123")

	// Simulate inventory check failure
	fmt.Println("Processing order - inventory check...")
	inventoryErr := domainerrors.NewWithType("INVENTORY_001", "Produto fora de estoque", domainerrors.ErrorTypeBusinessRule)
	inventoryErr.WithMetadata("product_id", "PROD_123")
	inventoryErr.WithMetadata("requested_quantity", 5)
	inventoryErr.WithMetadata("available_quantity", 2)
	inventoryErr.WithMetadata("operation", "inventory_check")

	// Simulate payment processing error
	fmt.Println("Processing payment...")
	paymentErr := domainerrors.NewWithType("PAYMENT_002", "Cartão de crédito recusado", domainerrors.ErrorTypeExternalService)
	paymentErr.WithMetadata("payment_method", "credit_card")
	paymentErr.WithMetadata("card_last_four", "1234")
	paymentErr.WithMetadata("amount", 199.99)
	paymentErr.WithMetadata("currency", "USD")
	paymentErr.WithMetadata("operation", "payment_processing")

	fmt.Println("Order processing completed with errors logged and metrics collected")
	fmt.Println()
}

func financialTransactionExample() {
	fmt.Println("--- Example 2: Financial Transaction (High Security) ---")

	// Simulate high-security financial operation
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req_transfer_67890")
	ctx = context.WithValue(ctx, "user_id", "user_456")

	// Simulate authentication failure
	fmt.Println("Authenticating financial transaction...")
	authErr := domainerrors.NewWithType("AUTH_FINANCIAL_001", "Falha na autenticação multifator", domainerrors.ErrorTypeAuthentication)
	authErr.WithMetadata("auth_method", "mfa")
	authErr.WithMetadata("attempt_count", 3)
	authErr.WithMetadata("client_ip", "203.0.113.1")
	authErr.WithMetadata("operation", "financial_transfer")
	authErr.WithMetadata("amount", 50000.00) // This will be sanitized

	// Simulate fraud detection
	fmt.Println("Running fraud detection...")
	fraudErr := domainerrors.NewWithType("FRAUD_001", "Transação suspeita detectada", domainerrors.ErrorTypeSecurity)
	fraudErr.WithMetadata("fraud_score", 85)
	fraudErr.WithMetadata("risk_factors", []string{"unusual_location", "large_amount", "off_hours"})
	fraudErr.WithMetadata("operation", "fraud_detection")

	fmt.Println("Financial transaction errors processed with enhanced security logging")
	fmt.Println()
}

func microserviceIntegrationExample() {
	fmt.Println("--- Example 3: Microservice Integration with Circuit Breaker ---")

	// Simulate multiple service failures to trigger circuit breaker
	serviceFailureCount := 0

	for i := 1; i <= 5; i++ {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", fmt.Sprintf("req_service_%d", i))

		if i <= 3 {
			// First 3 calls fail - building up to circuit breaker
			fmt.Printf("Service call %d - timeout...\n", i)
			timeoutErr := domainerrors.NewWithType("SERVICE_TIMEOUT_001", "Serviço de usuários não respondeu", domainerrors.ErrorTypeTimeout)
			timeoutErr.WithMetadata("service_name", "user-service")
			timeoutErr.WithMetadata("endpoint", "/api/v1/users/profile")
			timeoutErr.WithMetadata("timeout_ms", 5000)
			timeoutErr.WithMetadata("attempt", i)

			serviceFailureCount++
		} else {
			// Circuit breaker should be open by now
			fmt.Printf("Service call %d - circuit breaker activated...\n", i)
			circuitErr := domainerrors.NewWithType("CIRCUIT_BREAKER_001", "Circuit breaker aberto para user-service", domainerrors.ErrorTypeCircuitBreaker)
			circuitErr.WithMetadata("service_name", "user-service")
			circuitErr.WithMetadata("failure_count", serviceFailureCount)
			circuitErr.WithMetadata("circuit_state", "OPEN")
		}
	}

	fmt.Println("Microservice integration errors handled with circuit breaker pattern")
	fmt.Println()
}

func apiRateLimitingExample() {
	fmt.Println("--- Example 4: API Rate Limiting and Abuse Detection ---")

	// Simulate rate limiting scenarios
	for i := 1; i <= 3; i++ {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", fmt.Sprintf("req_api_%d", i))
		ctx = context.WithValue(ctx, "user_id", "user_abuser_123")

		if i == 1 {
			// Normal rate limit exceeded
			rateLimitErr := domainerrors.NewWithType("RATE_LIMIT_001", "Limite de taxa excedido", domainerrors.ErrorTypeRateLimit)
			rateLimitErr.WithMetadata("limit_type", "per_user")
			rateLimitErr.WithMetadata("current_count", 1001)
			rateLimitErr.WithMetadata("limit", 1000)
			rateLimitErr.WithMetadata("window", "1h")
			rateLimitErr.WithMetadata("reset_time", time.Now().Add(time.Hour).Unix())
		} else if i == 2 {
			// Aggressive rate limiting - potential abuse
			abuseErr := domainerrors.NewWithType("ABUSE_DETECTED_001", "Possível abuso detectado", domainerrors.ErrorTypeSecurity)
			abuseErr.WithMetadata("abuse_type", "aggressive_requests")
			abuseErr.WithMetadata("requests_per_minute", 500)
			abuseErr.WithMetadata("client_ip", "198.51.100.1")
			abuseErr.WithMetadata("user_agent", "automated-bot/1.0")
		} else {
			// Account temporarily blocked
			blockErr := domainerrors.NewWithType("ACCOUNT_BLOCKED_001", "Conta temporariamente bloqueada", domainerrors.ErrorTypeAuthorization)
			blockErr.WithMetadata("block_reason", "suspected_abuse")
			blockErr.WithMetadata("block_duration", "24h")
			blockErr.WithMetadata("unblock_time", time.Now().Add(24*time.Hour).Unix())
		}
	}

	fmt.Println("API rate limiting and abuse detection examples processed")
	fmt.Println()
}

func databaseOperationExample() {
	fmt.Println("--- Example 5: Database Operation with Retry and Circuit Breaker ---")

	// Simulate database operation failures
	for attempt := 1; attempt <= 4; attempt++ {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", fmt.Sprintf("req_db_%d", attempt))

		if attempt <= 3 {
			// Connection/timeout errors that might be retryable
			dbErr := domainerrors.NewWithType("DB_CONNECTION_001", "Falha na conexão com banco de dados", domainerrors.ErrorTypeDatabase)
			dbErr.WithMetadata("database", "postgresql")
			dbErr.WithMetadata("host", "db-cluster-primary.internal")
			dbErr.WithMetadata("port", 5432)
			dbErr.WithMetadata("operation", "SELECT")
			dbErr.WithMetadata("table", "users")
			dbErr.WithMetadata("attempt", attempt)
			dbErr.WithMetadata("max_retries", 3)
			dbErr.WithMetadata("backoff_ms", 1000*attempt) // Exponential backoff

			fmt.Printf("Database attempt %d failed - will retry\n", attempt)
		} else {
			// Final failure - circuit breaker opens
			circuitErr := domainerrors.NewWithType("DB_CIRCUIT_BREAKER_001", "Circuit breaker aberto para banco de dados", domainerrors.ErrorTypeCircuitBreaker)
			circuitErr.WithMetadata("database", "postgresql")
			circuitErr.WithMetadata("total_attempts", attempt)
			circuitErr.WithMetadata("circuit_state", "OPEN")
			circuitErr.WithMetadata("estimated_recovery", time.Now().Add(5*time.Minute).Unix())

			fmt.Printf("Database circuit breaker activated after %d attempts\n", attempt)
		}
	}

	fmt.Println("Database operation errors handled with retry and circuit breaker")
	fmt.Println()
}

func showCollectedMetrics() {
	fmt.Println("--- Collected Metrics Summary ---")

	errorCounts := metricsCollector.GetErrorCounts()
	if len(errorCounts) == 0 {
		fmt.Println("No metrics collected")
		return
	}

	fmt.Println("Error counts by type:")
	for errorType, count := range errorCounts {
		fmt.Printf("  %s: %d\n", errorType, count)
	}

	// Calculate totals
	totalErrors := 0
	for _, count := range errorCounts {
		totalErrors += count
	}

	fmt.Printf("Total errors processed: %d\n", totalErrors)
	fmt.Println()
}

// Helper functions
func getLogLevel(errorType domainerrors.ErrorType) string {
	switch errorType {
	case domainerrors.ErrorTypeServer, domainerrors.ErrorTypeDatabase, domainerrors.ErrorTypeSecurity:
		return "error"
	case domainerrors.ErrorTypeAuthentication, domainerrors.ErrorTypeAuthorization, domainerrors.ErrorTypeTimeout:
		return "warn"
	case domainerrors.ErrorTypeValidation, domainerrors.ErrorTypeNotFound:
		return "info"
	default:
		return "warn"
	}
}

func sanitizeSensitiveData(message string) string {
	// Em produção, usaria regex mais sofisticados para detectar e mascarar:
	// - Números de cartão de crédito
	// - CPFs, CNPJs
	// - Senhas, tokens
	// - Informações pessoais

	// Exemplo simples
	if strings.Contains(strings.ToLower(message), "password") {
		return strings.ReplaceAll(message, "password", "****")
	}
	if strings.Contains(strings.ToLower(message), "token") {
		return strings.ReplaceAll(message, "token", "****")
	}

	return message
}

func sanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range metadata {
		lowerKey := strings.ToLower(key)

		// Sanitize sensitive fields
		if lowerKey == "password" || lowerKey == "token" || lowerKey == "secret" {
			sanitized[key] = "****"
		} else if lowerKey == "amount" && value != nil {
			// Mask large financial amounts
			if amount, ok := value.(float64); ok && amount > 10000 {
				sanitized[key] = "****"
			} else {
				sanitized[key] = value
			}
		} else if lowerKey == "card_number" {
			sanitized[key] = "****"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}
