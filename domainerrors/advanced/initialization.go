package advanced

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/performance"
)

// InitializeAdvancedFeatures inicializa todas as funcionalidades avançadas
func InitializeAdvancedFeatures() error {
	// Inicializar pools de performance
	initializePerformancePools()

	// Configurar agregação de erros
	initializeErrorAggregation()

	// Configurar hooks condicionais padrão
	initializeConditionalHooks()

	// Configurar sistema de retry
	initializeRetryMechanism()

	// Configurar recuperação de erros
	initializeErrorRecovery()

	// Configurar stack trace otimizado
	initializeStackTraceOptimization()

	return nil
}

// initializePerformancePools configura pools de performance
func initializePerformancePools() {
	// Configurar stack trace condicional
	performance.AddGlobalStackCondition(func() bool {
		// Captura stack trace apenas para erros críticos
		return true // Pode ser configurado baseado em nível de log, ambiente, etc.
	})

	// Adicionar strings comuns ao pool
	commonErrorCodes := []string{
		"VALIDATION_FAILED",
		"USER_NOT_FOUND",
		"PERMISSION_DENIED",
		"RATE_LIMIT_EXCEEDED",
		"SERVICE_UNAVAILABLE",
		"DATABASE_CONNECTION_FAILED",
		"EXTERNAL_API_TIMEOUT",
		"AUTHENTICATION_FAILED",
		"INVALID_TOKEN",
		"RESOURCE_NOT_FOUND",
	}

	for _, code := range commonErrorCodes {
		performance.AddCommonString(code)
	}
}

// initializeErrorAggregation configura agregação de erros
func initializeErrorAggregation() {
	// Configurar agregador global com threshold de 10 erros ou window de 5 segundos
	globalAggregator = NewErrorAggregator(10, 5*time.Second)
}

// initializeConditionalHooks configura hooks condicionais padrão
func initializeConditionalHooks() {
	// Hook para erros de segurança
	RegisterConditionalErrorHook(
		"security_alert",
		10, // Alta prioridade
		ErrorTypeCondition(interfaces.SecurityError),
		func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			// Implementar notificação de segurança
			fmt.Printf("[SECURITY ALERT] %s: %s\n", err.Code(), err.Error())
			return nil
		},
	)

	// Hook para erros críticos (5xx)
	RegisterConditionalErrorHook(
		"critical_error",
		9, // Alta prioridade
		func(err interfaces.DomainErrorInterface) bool {
			return err.HTTPStatus() >= 500
		},
		func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			// Implementar alertas críticos
			fmt.Printf("[CRITICAL] %s: %s (HTTP %d)\n", err.Code(), err.Error(), err.HTTPStatus())
			return nil
		},
	)

	// Hook para erros de rate limit
	RegisterConditionalErrorHook(
		"rate_limit_monitor",
		5, // Prioridade média
		ErrorTypeCondition(interfaces.RateLimitError),
		func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			// Implementar monitoramento de rate limit
			fmt.Printf("[RATE LIMIT] %s: %s\n", err.Code(), err.Error())
			return nil
		},
	)

	// Hook para erros de validação (log apenas, não agregar para evitar recursão)
	RegisterConditionalErrorHook(
		"validation_logger",
		3, // Prioridade baixa
		ErrorTypeCondition(interfaces.ValidationError),
		func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			// Apenas log, sem agregação para evitar recursão infinita
			fmt.Printf("[VALIDATION] %s: %s\n", err.Code(), err.Error())
			return nil
		},
	)
}

// initializeRetryMechanism configura sistema de retry
func initializeRetryMechanism() {
	// Configurar retry para serviços externos
	externalServiceConfig := &RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		JitterFactor:  0.2,
		RetryableErrors: []interfaces.ErrorType{
			interfaces.ExternalServiceError,
			interfaces.TimeoutError,
			interfaces.ServiceUnavailableError,
			interfaces.ResourceExhaustedError,
		},
	}

	SetGlobalRetryConfig(externalServiceConfig)
}

// initializeErrorRecovery configura recuperação de erros
func initializeErrorRecovery() {
	// Configurar estratégias específicas para diferentes tipos de erro

	// Timeout: Retry com timeout reduzido
	RegisterGlobalStrategy(interfaces.TimeoutError, &RecoveryConfig{
		Strategy:    RetryStrategy,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
	})

	// Cache errors: Fallback para fonte primária
	RegisterGlobalStrategy(interfaces.CacheError, &RecoveryConfig{
		Strategy:      FallbackStrategy,
		MaxAttempts:   2,
		Timeout:       2 * time.Second,
		FallbackValue: "cache_unavailable",
		FallbackFunc: func(ctx context.Context, err error) (interface{}, error) {
			return map[string]interface{}{
				"status":    "fallback",
				"message":   "Cache unavailable, using primary source",
				"timestamp": time.Now().Format(time.RFC3339),
			}, nil
		},
	})

	// Rate limit: Graceful degradation
	RegisterGlobalStrategy(interfaces.RateLimitError, &RecoveryConfig{
		Strategy:    GracefulDegradationStrategy,
		MaxAttempts: 1,
		Timeout:     1 * time.Second,
	})

	// External service: Circuit breaker
	RegisterGlobalStrategy(interfaces.ExternalServiceError, &RecoveryConfig{
		Strategy:    CircuitBreakerStrategy,
		MaxAttempts: 5,
		Timeout:     30 * time.Second,
	})
}

// initializeStackTraceOptimization configura otimização de stack trace
func initializeStackTraceOptimization() {
	// Desabilitar stack trace em produção para performance
	// Em desenvolvimento, manter habilitado
	isDevelopment := true // Seria lido de configuração

	performance.SetStackTraceEnabled(isDevelopment)

	// Adicionar condições para captura seletiva
	performance.AddGlobalStackCondition(func() bool {
		// Captura apenas para erros críticos (5xx)
		return true // Seria implementada lógica específica
	})
}

// Variáveis globais
var globalAggregator *ErrorAggregator

// GetGlobalAggregator retorna agregador global
func GetGlobalAggregator() *ErrorAggregator {
	return globalAggregator
}

// ShutdownAdvancedFeatures finaliza funcionalidades avançadas
func ShutdownAdvancedFeatures() error {
	// Flush final do agregador
	if globalAggregator != nil {
		if err := globalAggregator.Close(); err != nil {
			return fmt.Errorf("failed to close error aggregator: %w", err)
		}
	}

	return nil
}

// ExampleUsage demonstra uso das funcionalidades avançadas
func ExampleUsage() {
	ctx := context.Background()

	// Exemplo 1: Error Aggregation
	fmt.Println("=== Error Aggregation Example ===")
	aggregator := NewErrorAggregator(3, 2*time.Second)

	// Adicionar alguns erros
	err1 := performance.NewPooledError(interfaces.ValidationError, "INVALID_EMAIL", "Invalid email format")
	err2 := performance.NewPooledError(interfaces.ValidationError, "INVALID_NAME", "Name too short")
	err3 := performance.NewPooledError(interfaces.ValidationError, "INVALID_AGE", "Age must be positive")

	aggregator.Add(err1)
	aggregator.Add(err2)
	aggregator.Add(err3) // Isso deve disparar o flush

	// Limpar recursos
	err1.Release()
	err2.Release()
	err3.Release()
	aggregator.Close()

	// Exemplo 2: Conditional Hooks
	fmt.Println("=== Conditional Hooks Example ===")

	// Criar erro de segurança - deve disparar hook condicional
	securityErr := performance.NewPooledError(interfaces.SecurityError, "UNAUTHORIZED_ACCESS", "Unauthorized access attempt")
	// Os hooks condicionais registrados serão executados automaticamente
	securityErr.Release()

	// Exemplo 3: Retry Mechanism
	fmt.Println("=== Retry Mechanism Example ===")

	retryableOperation := func(ctx context.Context) error {
		// Simula operação que pode falhar
		return performance.NewPooledError(interfaces.ExternalServiceError, "API_TIMEOUT", "API timeout")
	}

	if err := ExecuteWithRetry(ctx, retryableOperation); err != nil {
		fmt.Printf("Operation failed after retries: %v\n", err)
	}

	// Exemplo 4: Error Recovery
	fmt.Println("=== Error Recovery Example ===")

	failingOperation := func(ctx context.Context) (interface{}, error) {
		return nil, performance.NewPooledError(interfaces.CacheError, "CACHE_MISS", "Cache miss")
	}

	result, err := Recover(ctx, performance.NewPooledError(interfaces.CacheError, "CACHE_MISS", "Cache miss"), failingOperation)
	if err == nil {
		fmt.Printf("Recovered with result: %v\n", result)
	} else {
		fmt.Printf("Recovery failed: %v\n", err)
	}

	// Exemplo 5: Performance Optimizations
	fmt.Println("=== Performance Optimizations Example ===")

	// Usar pools para criação eficiente de erros
	perfErr := performance.NewPooledError(interfaces.DatabaseError, "CONNECTION_FAILED", "Database connection failed")
	perfErr.WithMetadata("database", "postgres")
	perfErr.WithMetadata("retry_count", 3)

	// Stack trace lazy (só captura detalhes se necessário)
	stackTrace := performance.CaptureStackTrace(1)
	if stackTrace.HasFrames() {
		fmt.Printf("Stack trace captured with %d frames\n", stackTrace.FrameCount())
	}

	// Limpeza
	perfErr.Release()
	performance.ReleaseStackTrace(stackTrace)

	fmt.Println("=== Example completed ===")
}

// PerformanceDemo demonstra melhorias de performance
func PerformanceDemo() {
	fmt.Println("=== Performance Demonstration ===")

	// Medir criação tradicional vs pooled
	performance.MeasureGlobal("traditional_error_creation", func() {
		for i := 0; i < 1000; i++ {
			_ = fmt.Errorf("error %d", i)
		}
	})

	performance.MeasureGlobal("pooled_error_creation", func() {
		for i := 0; i < 1000; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "TEST", fmt.Sprintf("error %d", i))
			err.Release()
		}
	})

	// Exibir estatísticas
	stats := performance.GlobalProfiler.GetStats()
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	performance.GlobalProfiler.Clear()
}
