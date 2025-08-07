package advanced

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/performance"
)

func TestErrorAggregation(t *testing.T) {
	t.Run("ThresholdTrigger", func(t *testing.T) {
		aggregator := NewErrorAggregator(3, 5*time.Second)
		defer aggregator.Close()

		// Adicionar erros até atingir threshold
		for i := 0; i < 3; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "TEST_ERROR", "test error")
			if aggErr := aggregator.Add(err); aggErr != nil && i == 2 {
				// Deve disparar flush no terceiro erro
				t.Log("Flush triggered correctly")
			}
			err.Release()
		}

		// Verificar que não há erros pendentes após flush
		if aggregator.HasErrors() {
			t.Error("Expected no pending errors after flush")
		}
	})

	t.Run("WindowTrigger", func(t *testing.T) {
		aggregator := NewErrorAggregator(10, 100*time.Millisecond)
		defer aggregator.Close()

		// Adicionar um erro
		err := performance.NewPooledError(interfaces.ValidationError, "TEST_ERROR", "test error")
		aggregator.Add(err)
		err.Release()

		// Esperar window expirar
		time.Sleep(150 * time.Millisecond)

		// Deve ter feito flush automático
		if aggregator.Count() > 0 {
			t.Error("Expected automatic flush after window expiration")
		}
	})
}

func TestConditionalHooks(t *testing.T) {
	t.Run("ErrorTypeCondition", func(t *testing.T) {
		manager := NewConditionalHookManager()
		executed := false

		manager.RegisterConditionalErrorHook(
			"test_hook",
			5,
			ErrorTypeCondition(interfaces.SecurityError),
			func(ctx context.Context, err interfaces.DomainErrorInterface) error {
				executed = true
				return nil
			},
		)

		// Erro de segurança - deve executar hook
		secErr := performance.NewPooledError(interfaces.SecurityError, "TEST", "test")
		manager.ExecuteConditionalHooks(context.Background(), secErr)
		secErr.Release()

		if !executed {
			t.Error("Expected hook to be executed for security error")
		}

		// Erro de validação - não deve executar hook
		executed = false
		valErr := performance.NewPooledError(interfaces.ValidationError, "TEST", "test")
		manager.ExecuteConditionalHooks(context.Background(), valErr)
		valErr.Release()

		if executed {
			t.Error("Expected hook NOT to be executed for validation error")
		}
	})

	t.Run("CombinedConditions", func(t *testing.T) {
		manager := NewConditionalHookManager()
		executed := false

		condition := CombinedCondition(
			ErrorTypeCondition(interfaces.AuthorizationError),
			HTTPStatusCondition(403),
		)

		manager.RegisterConditionalErrorHook(
			"combined_hook",
			5,
			condition,
			func(ctx context.Context, err interfaces.DomainErrorInterface) error {
				executed = true
				return nil
			},
		)

		// Erro que satisfaz ambas condições
		err := performance.NewPooledError(interfaces.AuthorizationError, "FORBIDDEN", "access denied")
		manager.ExecuteConditionalHooks(context.Background(), err)
		err.Release()

		if !executed {
			t.Error("Expected hook to be executed for combined conditions")
		}
	})
}

func TestRetryMechanism(t *testing.T) {
	t.Run("SuccessAfterRetries", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:   3,
			InitialDelay:  10 * time.Millisecond,
			BackoffFactor: 2.0,
			RetryableErrors: []interfaces.ErrorType{
				interfaces.ExternalServiceError,
			},
		}

		handler := NewRetryHandler(config)
		attempts := 0

		operation := func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return performance.NewPooledError(interfaces.ExternalServiceError, "TIMEOUT", "service timeout")
			}
			return nil // Sucesso na terceira tentativa
		}

		err := handler.Execute(context.Background(), operation)

		if err != nil {
			t.Errorf("Expected success after retries, got: %v", err)
		}

		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got: %d", attempts)
		}
	})

	t.Run("NonRetryableError", func(t *testing.T) {
		handler := NewRetryHandler(DefaultRetryConfig())
		attempts := 0

		operation := func(ctx context.Context) error {
			attempts++
			return performance.NewPooledError(interfaces.ValidationError, "INVALID", "validation failed")
		}

		err := handler.Execute(context.Background(), operation)

		if err == nil {
			t.Error("Expected error for non-retryable operation")
		}

		if attempts != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got: %d", attempts)
		}
	})
}

func TestErrorRecovery(t *testing.T) {
	t.Run("FallbackRecovery", func(t *testing.T) {
		handler := NewRecoveryHandler()

		// Registrar estratégia de fallback
		handler.RegisterStrategy(interfaces.CacheError, &RecoveryConfig{
			Strategy:      FallbackStrategy,
			FallbackValue: "fallback_data",
		})

		operation := func(ctx context.Context) (interface{}, error) {
			return nil, performance.NewPooledError(interfaces.CacheError, "CACHE_MISS", "cache miss")
		}

		result, err := handler.Recover(
			context.Background(),
			performance.NewPooledError(interfaces.CacheError, "CACHE_MISS", "cache miss"),
			operation,
		)

		if err != nil {
			t.Errorf("Expected successful recovery, got error: %v", err)
		}

		if result != "fallback_data" {
			t.Errorf("Expected fallback_data, got: %v", result)
		}
	})

	t.Run("NoRecoveryStrategy", func(t *testing.T) {
		handler := NewRecoveryHandler()

		operation := func(ctx context.Context) (interface{}, error) {
			return nil, performance.NewPooledError(interfaces.BusinessError, "BUSINESS_RULE", "business rule violation")
		}

		originalErr := performance.NewPooledError(interfaces.BusinessError, "BUSINESS_RULE", "business rule violation")
		result, err := handler.Recover(context.Background(), originalErr, operation)
		originalErr.Release()

		if err == nil {
			t.Error("Expected error when no recovery strategy is available")
		}

		if result != nil {
			t.Error("Expected nil result when recovery fails")
		}
	})
}

func TestPerformanceOptimizations(t *testing.T) {
	t.Run("ErrorPooling", func(t *testing.T) {
		pool := performance.NewErrorPool()

		// Obter erro do pool
		err1 := pool.GetDomainError()
		if err1 == nil {
			t.Fatal("Expected non-nil error from pool")
		}

		// Inicializar e usar
		err1.Initialize(interfaces.ValidationError, "TEST", "test error")
		err1.WithMetadata("key", "value")

		// Retornar ao pool
		pool.PutDomainError(err1)

		// Obter novamente - deve ser a mesma instância reutilizada
		err2 := pool.GetDomainError()
		if err2 == nil {
			t.Fatal("Expected non-nil error from pool")
		}

		// Deve estar limpa para reutilização
		if err2.Code() != "" || err2.Error() != "" {
			t.Error("Expected clean error instance from pool")
		}

		pool.PutDomainError(err2)
	})

	t.Run("LazyStackTrace", func(t *testing.T) {
		lst := performance.NewLazyStackTrace(1)

		// Verificar se tem frames sem capturar detalhes
		if !lst.HasFrames() {
			t.Error("Expected stack trace to have frames")
		}

		// Verificar que ainda não capturou detalhes
		if lst.IsCaptured() {
			t.Error("Expected stack trace to not be captured yet")
		}

		// Forçar captura de detalhes
		frames := lst.GetFrames()
		if len(frames) == 0 {
			t.Error("Expected non-empty frames after capture")
		}

		// Agora deve estar capturado
		if !lst.IsCaptured() {
			t.Error("Expected stack trace to be captured after GetFrames()")
		}
	})

	t.Run("StringInterning", func(t *testing.T) {
		// Strings comuns devem retornar mesma referência
		str1 := performance.InternString("VALIDATION_ERROR")
		str2 := performance.InternString("VALIDATION_ERROR")

		if str1 != str2 {
			t.Error("Expected same reference for common strings")
		}

		// Strings não comuns devem retornar referências diferentes
		str3 := performance.InternString("UNCOMMON_ERROR_123")
		str4 := performance.InternString("UNCOMMON_ERROR_456")

		if str3 == str4 {
			t.Error("Expected different references for uncommon strings")
		}
	})
}

func TestInitialization(t *testing.T) {
	t.Run("InitializeAdvancedFeatures", func(t *testing.T) {
		err := InitializeAdvancedFeatures()
		if err != nil {
			t.Errorf("Failed to initialize advanced features: %v", err)
		}

		// Verificar que agregador global foi criado
		if GetGlobalAggregator() == nil {
			t.Error("Expected global aggregator to be initialized")
		}

		// Limpar
		err = ShutdownAdvancedFeatures()
		if err != nil {
			t.Errorf("Failed to shutdown advanced features: %v", err)
		}
	})
}

// BenchmarkAdvancedFeatures testa performance das funcionalidades avançadas
func BenchmarkAdvancedFeatures(b *testing.B) {
	b.Run("ErrorAggregation", func(b *testing.B) {
		aggregator := NewErrorAggregator(1000, 1*time.Second) // High threshold para evitar flush durante benchmark
		defer aggregator.Close()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "BENCH_ERROR", "benchmark error")
			aggregator.Add(err)
			err.Release()
		}
	})

	b.Run("ConditionalHooks", func(b *testing.B) {
		manager := NewConditionalHookManager()
		manager.RegisterConditionalErrorHook(
			"bench_hook",
			5,
			ErrorTypeCondition(interfaces.ValidationError),
			func(ctx context.Context, err interfaces.DomainErrorInterface) error {
				return nil
			},
		)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "BENCH_ERROR", "benchmark error")
			manager.ExecuteConditionalHooks(context.Background(), err)
			err.Release()
		}
	})

	b.Run("PerformanceOptimizations", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "BENCH_ERROR", "benchmark error")
			err.WithMetadata("key", "value")
			lst := performance.CaptureStackTrace(1)

			// Simular uso
			_ = err.Error()
			_ = lst.HasFrames()

			// Limpar
			err.Release()
			performance.ReleaseStackTrace(lst)
		}
	})
}
