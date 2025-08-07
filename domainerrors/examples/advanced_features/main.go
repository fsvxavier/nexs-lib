package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/advanced"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/performance"
)

func main() {
	fmt.Println("🚀 Demonstração das Funcionalidades Avançadas e Melhorias de Performance")
	fmt.Println(strings.Repeat("=", 80))

	// Inicializar todas as funcionalidades avançadas
	if err := advanced.InitializeAdvancedFeatures(); err != nil {
		log.Fatalf("Falha ao inicializar funcionalidades avançadas: %v", err)
	}
	defer func() {
		if err := advanced.ShutdownAdvancedFeatures(); err != nil {
			log.Printf("Erro ao finalizar funcionalidades: %v", err)
		}
	}()

	// Executar demonstrações
	demonstrateErrorAggregation()
	demonstrateConditionalHooks()
	demonstrateRetryMechanism()
	demonstrateErrorRecovery()
	demonstratePerformanceOptimizations()
	demonstrateBenchmarkComparison()

	fmt.Println("✅ Demonstração concluída com sucesso!")
}

func demonstrateErrorAggregation() {
	fmt.Println("\n📊 1. DEMONSTRAÇÃO: Error Aggregation")
	fmt.Println(strings.Repeat("-", 50))

	// Criar agregador para coletar erros de validação
	aggregator := advanced.NewErrorAggregator(5, 2*time.Second)
	defer aggregator.Close()

	// Simular múltiplos erros de validação em um formulário
	validationErrors := []struct {
		field   string
		code    string
		message string
	}{
		{"email", "INVALID_EMAIL", "Formato de email inválido"},
		{"password", "WEAK_PASSWORD", "Senha muito fraca"},
		{"age", "INVALID_AGE", "Idade deve ser maior que 0"},
		{"name", "EMPTY_NAME", "Nome é obrigatório"},
		{"phone", "INVALID_PHONE", "Formato de telefone inválido"},
	}

	fmt.Printf("Adicionando %d erros de validação ao agregador...\n", len(validationErrors))

	for i, errInfo := range validationErrors {
		err := performance.NewPooledError(
			interfaces.ValidationError,
			errInfo.code,
			errInfo.message,
		)
		err.WithMetadata("field", errInfo.field)
		err.WithMetadata("form", "user_registration")

		if aggErr := aggregator.Add(err); aggErr != nil {
			fmt.Printf("✅ Flush automático disparado após %d erros\n", i+1)
		}

		err.Release()
		time.Sleep(50 * time.Millisecond) // Simular tempo entre erros
	}

	fmt.Printf("📈 Erros restantes no agregador: %d\n", aggregator.Count())

	// Força flush dos erros restantes
	aggregator.Flush()
}

func demonstrateConditionalHooks() {
	fmt.Println("\n🎯 2. DEMONSTRAÇÃO: Conditional Hooks")
	fmt.Println(strings.Repeat("-", 50))

	// Os hooks condicionais foram registrados na inicialização
	// Vamos disparar diferentes tipos de erro para demonstrar

	fmt.Println("Disparando erro de segurança (deve ativar hook de alta prioridade)...")
	securityErr := performance.NewPooledError(
		interfaces.SecurityError,
		"UNAUTHORIZED_ACCESS",
		"Tentativa de acesso não autorizado detectada",
	)
	securityErr.WithMetadata("ip", "192.168.1.100")
	securityErr.WithMetadata("user_agent", "Suspicious-Bot/1.0")
	securityErr.Release()

	fmt.Println("Disparando erro interno (deve ativar hook crítico)...")
	internalErr := performance.NewPooledError(
		interfaces.ServerError,
		"DATABASE_CONNECTION_FAILED",
		"Falha na conexão com banco de dados",
	)
	internalErr.WithMetadata("database", "postgres")
	internalErr.WithMetadata("retry_count", 3)
	internalErr.Release()

	fmt.Println("Disparando erro de rate limit (deve ativar hook de monitoramento)...")
	rateLimitErr := performance.NewPooledError(
		interfaces.RateLimitError,
		"RATE_LIMIT_EXCEEDED",
		"Limite de requisições excedido",
	)
	rateLimitErr.WithMetadata("client_ip", "10.0.0.1")
	rateLimitErr.WithMetadata("requests_per_minute", 1000)
	rateLimitErr.Release()
}

func demonstrateRetryMechanism() {
	fmt.Println("\n🔄 3. DEMONSTRAÇÃO: Retry Mechanism")
	fmt.Println(strings.Repeat("-", 50))

	ctx := context.Background()

	// Simular operação que falha algumas vezes antes de ter sucesso
	attempts := 0
	simulatedExternalService := func(ctx context.Context) error {
		attempts++
		fmt.Printf("  Tentativa %d...\n", attempts)

		if attempts < 3 {
			// Falha nas primeiras tentativas
			return performance.NewPooledError(
				interfaces.ExternalServiceError,
				"API_TIMEOUT",
				"Timeout na API externa",
			)
		}

		// Sucesso na terceira tentativa
		fmt.Printf("  ✅ Sucesso na tentativa %d\n", attempts)
		return nil
	}

	fmt.Println("Executando operação com retry automático...")
	start := time.Now()

	if err := advanced.ExecuteWithRetry(ctx, simulatedExternalService); err != nil {
		fmt.Printf("❌ Operação falhou após todas as tentativas: %v\n", err)
	} else {
		fmt.Printf("✅ Operação bem-sucedida após %d tentativas em %v\n", attempts, time.Since(start))
	}

	// Demonstrar operação com resultado
	fmt.Println("\nExecutando operação com retry que retorna resultado...")
	attempts = 0

	operationWithResult := func(ctx context.Context) (interface{}, error) {
		attempts++
		fmt.Printf("  Tentativa %d para obter dados...\n", attempts)

		if attempts < 2 {
			return nil, performance.NewPooledError(
				interfaces.ExternalServiceError,
				"SERVICE_UNAVAILABLE",
				"Serviço temporariamente indisponível",
			)
		}

		return map[string]interface{}{
			"data":      "dados importantes",
			"timestamp": time.Now().Format(time.RFC3339),
			"attempts":  attempts,
		}, nil
	}

	result, err := advanced.ExecuteWithRetryAndResult(ctx, operationWithResult)
	if err != nil {
		fmt.Printf("❌ Falha ao obter dados: %v\n", err)
	} else {
		fmt.Printf("✅ Dados obtidos com sucesso: %+v\n", result)
	}

	// Mostrar estatísticas de retry
	stats := advanced.GetGlobalRetryStats()
	fmt.Printf("📊 Estatísticas de Retry: %+v\n", stats)
}

func demonstrateErrorRecovery() {
	fmt.Println("\n🛠️ 4. DEMONSTRAÇÃO: Error Recovery")
	fmt.Println(strings.Repeat("-", 50))

	ctx := context.Background()

	// Exemplo 1: Recuperação com Fallback
	fmt.Println("Testando recuperação com fallback para erro de cache...")

	cacheOperation := func(ctx context.Context) (interface{}, error) {
		return nil, performance.NewPooledError(
			interfaces.CacheError,
			"CACHE_MISS",
			"Dados não encontrados no cache",
		)
	}

	cacheErr := performance.NewPooledError(interfaces.CacheError, "CACHE_MISS", "Cache miss")
	result, err := advanced.Recover(ctx, cacheErr, cacheOperation)
	cacheErr.Release()

	if err != nil {
		fmt.Printf("❌ Recuperação falhou: %v\n", err)
	} else {
		fmt.Printf("✅ Recuperação bem-sucedida com fallback: %+v\n", result)
	}

	// Exemplo 2: Recuperação com Retry
	fmt.Println("\nTestando recuperação com retry para timeout...")

	timeoutAttempts := 0
	timeoutOperation := func(ctx context.Context) (interface{}, error) {
		timeoutAttempts++
		if timeoutAttempts < 2 {
			return nil, performance.NewPooledError(
				interfaces.TimeoutError,
				"REQUEST_TIMEOUT",
				"Timeout na requisição",
			)
		}
		return "dados recuperados", nil
	}

	timeoutErr := performance.NewPooledError(interfaces.TimeoutError, "REQUEST_TIMEOUT", "Timeout")
	result, err = advanced.Recover(ctx, timeoutErr, timeoutOperation)
	timeoutErr.Release()

	if err != nil {
		fmt.Printf("❌ Recuperação por retry falhou: %v\n", err)
	} else {
		fmt.Printf("✅ Recuperação por retry bem-sucedida: %v\n", result)
	}

	// Exemplo 3: Degradação Graciosa
	fmt.Println("\nTestando degradação graciosa para resource exhausted...")

	resourceOperation := func(ctx context.Context) (interface{}, error) {
		return nil, performance.NewPooledError(
			interfaces.ResourceExhaustedError,
			"MEMORY_EXHAUSTED",
			"Memória insuficiente",
		)
	}

	resourceErr := performance.NewPooledError(interfaces.ResourceExhaustedError, "MEMORY_EXHAUSTED", "Memory exhausted")
	result, err = advanced.Recover(ctx, resourceErr, resourceOperation)
	resourceErr.Release()

	if err != nil {
		fmt.Printf("❌ Degradação graciosa falhou: %v\n", err)
	} else {
		fmt.Printf("✅ Degradação graciosa ativada: %+v\n", result)
	}
}

func demonstratePerformanceOptimizations() {
	fmt.Println("\n⚡ 5. DEMONSTRAÇÃO: Performance Optimizations")
	fmt.Println(strings.Repeat("-", 50))

	// Demonstrar pooling de erros
	fmt.Println("Testando performance com pooling de erros...")

	const iterations = 1000

	// Medir criação tradicional de erros
	performance.MeasureGlobal("traditional_errors", func() {
		for i := 0; i < iterations; i++ {
			err := fmt.Errorf("erro tradicional %d com metadados: %s", i, "dados extras")
			_ = err.Error() // Usar o erro
		}
	})

	// Medir criação com pool
	performance.MeasureGlobal("pooled_errors", func() {
		for i := 0; i < iterations; i++ {
			err := performance.NewPooledError(
				interfaces.ValidationError,
				"POOLED_ERROR",
				fmt.Sprintf("erro pooled %d", i),
			)
			err.WithMetadata("iteration", i)
			err.WithMetadata("extra_data", "dados extras")

			_ = err.Error() // Usar o erro
			_ = err.HTTPStatus()
			_ = err.Metadata()

			err.Release() // Importante: retornar ao pool
		}
	})

	// Demonstrar lazy stack trace
	fmt.Println("Testando lazy stack trace...")

	performance.MeasureGlobal("lazy_stacktrace", func() {
		for i := 0; i < 100; i++ {
			lst := performance.CaptureStackTrace(1)

			// Na maioria dos casos, apenas verificamos se existe
			if lst.HasFrames() {
				// Só capturamos detalhes quando realmente necessário
				if i%10 == 0 { // 10% das vezes
					frames := lst.GetFrames()
					_ = len(frames)
				}
			}

			performance.ReleaseStackTrace(lst)
		}
	})

	// Demonstrar string interning
	fmt.Println("Testando string interning...")

	performance.MeasureGlobal("string_interning", func() {
		commonCodes := []string{
			"VALIDATION_ERROR",
			"NOT_FOUND",
			"INTERNAL_ERROR",
			"VALIDATION_ERROR", // Repetida
			"NOT_FOUND",        // Repetida
		}

		for _, code := range commonCodes {
			internedStr := performance.InternString(code)
			_ = internedStr
		}
	})

	// Exibir estatísticas de performance
	fmt.Println("\n📊 Estatísticas de Performance:")
	stats := performance.GlobalProfiler.GetStats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func demonstrateBenchmarkComparison() {
	fmt.Println("\n🏁 6. DEMONSTRAÇÃO: Comparação de Performance")
	fmt.Println(strings.Repeat("-", 50))

	// Função helper para medir tempo e alocações
	measureOperation := func(name string, iterations int, operation func()) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		start := time.Now()
		operation()
		duration := time.Since(start)

		runtime.ReadMemStats(&m2)

		allocations := m2.TotalAlloc - m1.TotalAlloc
		mallocCount := m2.Mallocs - m1.Mallocs

		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Duração: %v\n", duration)
		fmt.Printf("    Alocações: %d bytes (%d objetos)\n", allocations, mallocCount)
		fmt.Printf("    Média por operação: %v (%d bytes/op)\n",
			duration/time.Duration(iterations), allocations/uint64(iterations))
	}

	const iterations = 10000

	// Comparar criação de erros
	fmt.Println("Comparando criação de erros:")

	measureOperation("Erros tradicionais", iterations, func() {
		for i := 0; i < iterations; i++ {
			err := fmt.Errorf("erro %d", i)
			_ = err.Error()
		}
	})

	measureOperation("Erros pooled", iterations, func() {
		for i := 0; i < iterations; i++ {
			err := performance.NewPooledError(interfaces.ValidationError, "TEST", fmt.Sprintf("erro %d", i))
			_ = err.Error()
			err.Release()
		}
	})

	// Comparar stack trace
	fmt.Println("\nComparando captura de stack trace:")

	measureOperation("Stack trace tradicional", 100, func() {
		for i := 0; i < 100; i++ {
			buf := make([]byte, 1024)
			runtime.Stack(buf, false)
		}
	})

	measureOperation("Lazy stack trace", 100, func() {
		for i := 0; i < 100; i++ {
			lst := performance.CaptureStackTrace(1)
			_ = lst.HasFrames()
			performance.ReleaseStackTrace(lst)
		}
	})

	fmt.Println("\n🎯 Resumo:")
	fmt.Println("  • Error pooling reduz significativamente alocações de memória")
	fmt.Println("  • Lazy stack trace evita overhead desnecessário")
	fmt.Println("  • String interning otimiza uso de memória para códigos comuns")
	fmt.Println("  • Funcionalidades avançadas mantêm performance alta")
}
