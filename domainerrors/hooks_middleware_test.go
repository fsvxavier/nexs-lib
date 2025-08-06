package domainerrors

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestHooksAndMiddlewares(t *testing.T) {
	// Limpa registros anteriores
	hooksRegistry = make(map[string][]HookFunc)
	middlewareChain = nil

	// Contador para verificar execução
	var hooksExecuted int
	var middlewaresExecuted int

	// Registra hooks
	RegisterHook("before_metadata", func(ctx context.Context, err *DomainError, operation string) error {
		hooksExecuted++
		t.Logf("Hook executado: %s para erro: %s", operation, err.Code)
		return nil
	})

	RegisterHook("after_error", func(ctx context.Context, err *DomainError, operation string) error {
		hooksExecuted++
		t.Logf("Hook executado: %s para erro: %s", operation, err.Code)
		return nil
	})

	// Registra middlewares
	RegisterMiddleware(func(ctx context.Context, err *DomainError, next func(*DomainError) *DomainError) *DomainError {
		middlewaresExecuted++
		t.Logf("Middleware executado para erro: %s", err.Code)

		// Adiciona metadados
		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}
		err.Metadata["processed_by_middleware"] = true
		err.Metadata["processed_at"] = time.Now()

		return next(err)
	})

	// Testa criação de erro
	err := New("TEST001", "Erro de teste")

	// Verifica se hooks foram executados
	if hooksExecuted == 0 {
		t.Error("Nenhum hook foi executado")
	}

	// Verifica se middlewares foram executados
	if middlewaresExecuted == 0 {
		t.Error("Nenhum middleware foi executado")
	}

	// Verifica se metadados foram adicionados pelo middleware
	if err.Metadata == nil {
		t.Error("Metadados não foram adicionados pelo middleware")
	} else {
		if processed, exists := err.Metadata["processed_by_middleware"]; !exists || !processed.(bool) {
			t.Error("Middleware não processou o erro corretamente")
		}
	}

	// Testa adição de metadados (deve executar hooks)
	hooksCountBefore := hooksExecuted
	err.WithMetadata("test_key", "test_value")

	if hooksExecuted <= hooksCountBefore {
		t.Error("Hooks não foram executados ao adicionar metadados")
	}

	t.Logf("Total de hooks executados: %d", hooksExecuted)
	t.Logf("Total de middlewares executados: %d", middlewaresExecuted)
}

func TestHookRegistration(t *testing.T) {
	// Limpa registros anteriores
	hooksRegistry = make(map[string][]HookFunc)

	// Registra múltiples hooks para a mesma operação
	RegisterHook("test_operation", func(ctx context.Context, err *DomainError, operation string) error {
		return nil
	})

	RegisterHook("test_operation", func(ctx context.Context, err *DomainError, operation string) error {
		return nil
	})

	// Verifica se os hooks foram registrados
	hooks := hooksRegistry["test_operation"]
	if len(hooks) != 2 {
		t.Errorf("Esperado 2 hooks, encontrado %d", len(hooks))
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Limpa cadeia anterior
	middlewareChain = nil

	var executionOrder []string

	// Registra middlewares em ordem
	RegisterMiddleware(func(ctx context.Context, err *DomainError, next func(*DomainError) *DomainError) *DomainError {
		executionOrder = append(executionOrder, "middleware1_before")
		result := next(err)
		executionOrder = append(executionOrder, "middleware1_after")
		return result
	})

	RegisterMiddleware(func(ctx context.Context, err *DomainError, next func(*DomainError) *DomainError) *DomainError {
		executionOrder = append(executionOrder, "middleware2_before")
		result := next(err)
		executionOrder = append(executionOrder, "middleware2_after")
		return result
	})

	// Cria erro para testar
	err := &DomainError{
		Code:    "TEST002",
		Message: "Teste middleware chain",
		Type:    ErrorTypeServer,
	}

	// Executa a cadeia
	ctx := context.Background()
	executeMiddleware(ctx, err)

	// Verifica ordem de execução
	expectedOrder := []string{
		"middleware1_before",
		"middleware2_before",
		"middleware2_after",
		"middleware1_after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Ordem de execução incorreta. Esperado %v, obtido %v", expectedOrder, executionOrder)
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Ordem de execução incorreta no índice %d. Esperado %s, obtido %s",
				i, expected, executionOrder[i])
		}
	}
}

func TestErrorWithContext(t *testing.T) {
	// Limpa registros anteriores
	hooksRegistry = make(map[string][]HookFunc)

	var contextReceived context.Context

	// Registra hook para capturar contexto
	RegisterHook("after_stack_trace", func(ctx context.Context, err *DomainError, operation string) error {
		contextReceived = ctx
		return nil
	})

	// Cria erro com contexto
	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	err := New("TEST003", "Erro com contexto")
	err.WithContext(ctx, "contexto adicionado")

	// Verifica se o contexto foi passado para o hook
	if contextReceived == nil {
		t.Error("Contexto não foi recebido pelo hook")
	}

	if value := contextReceived.Value("test_key"); value == nil || value.(string) != "test_value" {
		t.Error("Valor do contexto não foi preservado")
	}
}

func TestHookError(t *testing.T) {
	// Limpa registros anteriores
	hooksRegistry = make(map[string][]HookFunc)

	var hookExecuted bool

	// Registra hook que retorna erro (mas não cria outro DomainError)
	RegisterHook("after_error", func(ctx context.Context, err *DomainError, operation string) error {
		hookExecuted = true
		t.Logf("Hook executado com erro: %s", err.Code)
		// Retorna um erro simples, não um DomainError para evitar recursão
		return fmt.Errorf("erro no hook")
	})

	// Testa se o erro é criado mesmo com falha no hook
	err := New("TEST004", "Teste erro no hook")

	// O erro deve ser criado mesmo com falha no hook
	if err == nil {
		t.Error("Erro não foi criado")
	}

	if err.Code != "TEST004" {
		t.Errorf("Código do erro incorreto. Esperado TEST004, obtido %s", err.Code)
	}

	// Verifica se o hook foi executado
	if !hookExecuted {
		t.Error("Hook não foi executado")
	}
}

func BenchmarkErrorCreationWithHooksAndMiddlewares(b *testing.B) {
	// Setup hooks and middlewares
	hooksRegistry = make(map[string][]HookFunc)
	middlewareChain = nil

	RegisterHook("after_error", func(ctx context.Context, err *DomainError, operation string) error {
		return nil
	})

	RegisterMiddleware(func(ctx context.Context, err *DomainError, next func(*DomainError) *DomainError) *DomainError {
		return next(err)
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		New("BENCH001", "Benchmark error")
	}
}

func BenchmarkErrorCreationWithoutHooksAndMiddlewares(b *testing.B) {
	// Clear hooks and middlewares
	hooksRegistry = make(map[string][]HookFunc)
	middlewareChain = nil

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := &DomainError{
			Code:      "BENCH002",
			Message:   "Benchmark error without hooks",
			Type:      ErrorTypeServer,
			Timestamp: time.Now(),
		}
		err.captureStackFrame("error created")
	}
}
