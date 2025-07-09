package example

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// Exemplo de uso do padrão Builder
func exampleBuilderPattern() {
	fmt.Println("=== Exemplo do Padrão Builder ===")

	// Uso básico do builder
	err1 := domainerrors.NewErrorBuilder().
		Type(domainerrors.ErrorTypeValidation).
		Code("INVALID_EMAIL").
		Message("O email fornecido não é válido").
		ValidationField("email", "Formato de email inválido").
		WithTimestamp().
		Build()

	fmt.Printf("Erro 1: %s\n", err1.Error())

	// Builder com múltiplas validações
	err2 := domainerrors.NewErrorBuilder().
		Type(domainerrors.ErrorTypeValidation).
		Code("MULTIPLE_VALIDATION_ERRORS").
		MessageF("Validação falhou para %d campos", 3).
		ValidationField("name", "Nome é obrigatório").
		ValidationField("age", "Idade deve ser maior que 0").
		ValidationField("email", "Email deve ter formato válido").
		WithTimestamp().
		Build()

	fmt.Printf("Erro 2: %s\n", err2.Error())

	// Builder para erro de negócio
	err3 := domainerrors.NewErrorBuilder().
		Type(domainerrors.ErrorTypeBusinessRule).
		Code("INSUFFICIENT_BALANCE").
		Message("Saldo insuficiente para realizar a transação").
		Entity("User").
		Detail("requested_amount", "1000.00").
		Detail("current_balance", "500.00").
		WithRequestID("req-123").
		WithUserID("user-456").
		WithTimestamp().
		Build()

	fmt.Printf("Erro 3: %s\n", err3.Error())

	// Builder clonado para variações
	baseBuilder := domainerrors.NewErrorBuilder().
		Type(domainerrors.ErrorTypeExternalService).
		Code("API_ERROR").
		WithTimestamp()

	// Clone para diferentes serviços
	err4 := baseBuilder.Clone().
		Message("Falha ao consultar serviço de pagamento").
		ExternalService("payment-api", 503).
		Build()

	err5 := baseBuilder.Clone().
		Message("Falha ao consultar serviço de usuários").
		ExternalService("user-api", 500).
		Build()

	fmt.Printf("Erro 4: %s\n", err4.Error())
	fmt.Printf("Erro 5: %s\n", err5.Error())
}

// Exemplo de uso do padrão Factory
func exampleFactoryPattern() {
	fmt.Println("\n=== Exemplo do Padrão Factory ===")

	// Factory básica
	factory := domainerrors.NewDomainErrorFactory()

	// Criação de diferentes tipos de erro
	validationErr := factory.CreateValidationError(
		"Dados de entrada inválidos",
		map[string][]string{
			"email": {"Formato inválido", "Já está em uso"},
			"age":   {"Deve ser maior que 18"},
		},
	)
	fmt.Printf("Validation Error: %s\n", validationErr.Error())

	notFoundErr := factory.CreateNotFoundError(
		"Usuário não encontrado",
		"User",
		"user-123",
	)
	fmt.Printf("Not Found Error: %s\n", notFoundErr.Error())

	businessErr := factory.CreateBusinessError(
		"INSUFFICIENT_FUNDS",
		"Saldo insuficiente para a operação",
	)
	fmt.Printf("Business Error: %s\n", businessErr.Error())

	timeoutErr := factory.CreateTimeoutError(
		"Operação expirou",
		30*time.Second,
	)
	fmt.Printf("Timeout Error: %s\n", timeoutErr.Error())
}

// Exemplo com observers (logging e métricas)
func exampleWithObservers() {
	fmt.Println("\n=== Exemplo com Observers ===")

	// Logger simples
	logger := func(level, message string, fields map[string]interface{}) {
		fmt.Printf("[%s] %s | Fields: %+v\n", level, message, fields)
	}

	// Contador de métricas simples
	metricsCounter := func(name string, tags map[string]string) {
		fmt.Printf("METRIC: %s | Tags: %+v\n", name, tags)
	}

	// Factory com observers
	factory := domainerrors.NewDomainErrorFactory(
		domainerrors.WithErrorObserver(domainerrors.NewLoggingErrorObserver(logger)),
		domainerrors.WithErrorObserver(domainerrors.NewMetricsErrorObserver(metricsCounter)),
	)

	// Criar erro - irá acionar os observers
	err := factory.CreateBusinessError("ORDER_LIMIT_EXCEEDED", "Limite de pedidos excedido")
	fmt.Printf("Error: %s\n", err.Error())
}

// Exemplo com context enricher
func exampleWithContextEnricher() {
	fmt.Println("\n=== Exemplo com Context Enricher ===")

	// Factory com enricher
	factory := domainerrors.NewDomainErrorFactory(
		domainerrors.WithContextEnricher(domainerrors.NewDefaultContextEnricher()),
	)

	// Contexto com informações
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-789")
	ctx = context.WithValue(ctx, "user_id", "user-999")
	ctx = context.WithValue(ctx, "operation_id", "op-abc")

	// Criar erro com contexto - será enriquecido automaticamente
	builder := domainerrors.NewErrorBuilder().
		Type(domainerrors.ErrorTypeAuthentication).
		Code("INVALID_TOKEN").
		Message("Token de acesso inválido")

	err := factory.CreateWithContext(ctx, builder)

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		fmt.Printf("Error: %s\n", err.Error())
		fmt.Printf("Metadata: %+v\n", domainErr.Metadata)
	}
}

// Exemplo de factory customizada
func exampleCustomFactory() {
	fmt.Println("\n=== Exemplo de Factory Customizada ===")

	// Builder factory customizada que sempre adiciona timestamp e status trace
	customBuilderFactory := func() *domainerrors.ErrorBuilder {
		return domainerrors.NewErrorBuilder().
			WithTimestamp().
			Detail("trace_id", fmt.Sprintf("trace-%d", time.Now().UnixNano()))
	}

	// Factory com configuração customizada
	factory := domainerrors.NewDomainErrorFactory(
		domainerrors.WithBuilderFactory(customBuilderFactory),
	)

	err := factory.CreateBusinessError("CUSTOM_ERROR", "Erro com configuração customizada")

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		fmt.Printf("Error: %s\n", err.Error())
		fmt.Printf("Timestamp: %v\n", domainErr.Metadata["timestamp"])
		fmt.Printf("Trace ID: %v\n", domainErr.Metadata["trace_id"])
	}
}

// Exemplo de error recovery e wrapping
func exampleErrorRecovery() {
	fmt.Println("\n=== Exemplo de Error Recovery ===")

	// Simular uma função que pode entrar em panic
	defer func() {
		if r := recover(); r != nil {
			// Recuperar panic e converter em domain error
			recoveredErr := domainerrors.NewErrorBuilder().
				Type(domainerrors.ErrorTypeInternal).
				Code("PANIC_RECOVERED").
				MessageF("Panic recuperado: %v", r).
				WithTimestamp().
				Build()

			fmt.Printf("Recovered Error: %s\n", recoveredErr.Error())
		}
	}()

	// Simular panic
	panic("Algo deu muito errado!")
}

// RunBuilderFactoryExamples executa exemplos dos padrões Builder e Factory
func RunBuilderFactoryExamples() {
	fmt.Println("Demonstração dos padrões Builder e Factory para Domain Errors")
	fmt.Println(strings.Repeat("=", 60))

	exampleBuilderPattern()
	exampleFactoryPattern()
	exampleWithObservers()
	exampleWithContextEnricher()
	exampleCustomFactory()
	exampleErrorRecovery()

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Demonstração completa!")
}
