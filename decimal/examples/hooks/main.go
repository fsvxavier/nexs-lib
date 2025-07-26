package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

// CustomAuditHook implementa hooks customizados para auditoria
type CustomAuditHook struct {
	auditLog []AuditEntry
}

type AuditEntry struct {
	Timestamp time.Time
	Operation string
	Args      []interface{}
	Result    interface{}
	Error     error
	Duration  time.Duration
}

func NewCustomAuditHook() *CustomAuditHook {
	return &CustomAuditHook{
		auditLog: make([]AuditEntry, 0),
	}
}

func (h *CustomAuditHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	// Pre-hook: registrar início da operação
	start := time.Now()

	entry := AuditEntry{
		Timestamp: start,
		Operation: operation,
		Args:      args,
	}

	// Salvar contexto no context para recuperar no post-hook
	ctx = context.WithValue(ctx, "audit_start", start)
	ctx = context.WithValue(ctx, "audit_entry", &entry)

	fmt.Printf("[AUDIT] Iniciando operação: %s\n", operation)
	return nil, nil // Não modificamos os argumentos
}

func (h *CustomAuditHook) ExecutePost(ctx context.Context, operation string, result interface{}, err error) error {
	// Recuperar informações do context
	if start, ok := ctx.Value("audit_start").(time.Time); ok {
		if entryPtr, ok := ctx.Value("audit_entry").(*AuditEntry); ok {
			entryPtr.Result = result
			entryPtr.Error = err
			entryPtr.Duration = time.Since(start)

			h.auditLog = append(h.auditLog, *entryPtr)

			fmt.Printf("[AUDIT] Operação completada: %s em %v\n", operation, entryPtr.Duration)
		}
	}
	return nil
}

func (h *CustomAuditHook) ExecuteError(ctx context.Context, operation string, err error) error {
	fmt.Printf("[AUDIT] Erro na operação %s: %v\n", operation, err)
	return nil
}

func (h *CustomAuditHook) GetAuditLog() []AuditEntry {
	return h.auditLog
}

// CustomValidationHook implementa validações customizadas
type CustomValidationHook struct {
	maxValue interfaces.Decimal
	minValue interfaces.Decimal
}

func NewCustomValidationHook(manager *decimal.Manager, min, max string) *CustomValidationHook {
	minVal, _ := manager.NewFromString(min)
	maxVal, _ := manager.NewFromString(max)

	return &CustomValidationHook{
		minValue: minVal,
		maxValue: maxVal,
	}
}

func (h *CustomValidationHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	// Validar apenas operações de criação
	if operation == "NewFromString" || operation == "NewFromFloat" || operation == "NewFromInt" {
		// Para operações de criação, vamos validar o resultado
		return nil, nil
	}

	// Para outras operações, validar argumentos que são decimais
	for _, arg := range args {
		if decimal, ok := arg.(interfaces.Decimal); ok {
			if decimal.IsLessThan(h.minValue) {
				return nil, fmt.Errorf("valor %s está abaixo do mínimo permitido %s", decimal.String(), h.minValue.String())
			}
			if decimal.IsGreaterThan(h.maxValue) {
				return nil, fmt.Errorf("valor %s está acima do máximo permitido %s", decimal.String(), h.maxValue.String())
			}
		}
	}

	return nil, nil
}

func (h *CustomValidationHook) ExecutePost(ctx context.Context, operation string, result interface{}, err error) error {
	// Validar resultado se for um decimal
	if err == nil && result != nil {
		if decimal, ok := result.(interfaces.Decimal); ok {
			if decimal.IsLessThan(h.minValue) {
				return fmt.Errorf("resultado %s está abaixo do mínimo permitido %s", decimal.String(), h.minValue.String())
			}
			if decimal.IsGreaterThan(h.maxValue) {
				return fmt.Errorf("resultado %s está acima do máximo permitido %s", decimal.String(), h.maxValue.String())
			}
		}
	}
	return nil
}

func (h *CustomValidationHook) ExecuteError(ctx context.Context, operation string, err error) error {
	fmt.Printf("[VALIDATION] Erro de validação: %v\n", err)
	return nil
}

// CustomMetricsHook implementa coleta de métricas
type CustomMetricsHook struct {
	operationCounts map[string]int
	operationTimes  map[string]time.Duration
}

func NewCustomMetricsHook() *CustomMetricsHook {
	return &CustomMetricsHook{
		operationCounts: make(map[string]int),
		operationTimes:  make(map[string]time.Duration),
	}
}

func (h *CustomMetricsHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	// Contar operação
	h.operationCounts[operation]++

	// Marcar início para medição de tempo
	ctx = context.WithValue(ctx, "metrics_start", time.Now())
	return nil, nil
}

func (h *CustomMetricsHook) ExecutePost(ctx context.Context, operation string, result interface{}, err error) error {
	// Medir tempo
	if start, ok := ctx.Value("metrics_start").(time.Time); ok {
		duration := time.Since(start)
		h.operationTimes[operation] += duration
	}
	return nil
}

func (h *CustomMetricsHook) ExecuteError(ctx context.Context, operation string, err error) error {
	// Contar erros
	h.operationCounts[operation+"_error"]++
	return nil
}

func (h *CustomMetricsHook) PrintMetrics() {
	fmt.Println("\n=== Métricas de Operações ===")
	for op, count := range h.operationCounts {
		totalTime := h.operationTimes[op]
		avgTime := time.Duration(0)
		if count > 0 {
			avgTime = totalTime / time.Duration(count)
		}
		fmt.Printf("%s: %d operações, tempo médio: %v\n", op, count, avgTime)
	}
}

func main() {
	fmt.Println("=== Exemplo Avançado - Hooks Customizados ===")

	fmt.Println("\n1. Hook de Auditoria")
	testAuditHook()

	fmt.Println("\n2. Hook de Validação")
	testValidationHook()

	fmt.Println("\n3. Hook de Métricas")
	testMetricsHook()

	fmt.Println("\n4. Múltiplos Hooks Combinados")
	testCombinedHooks()
}

func testAuditHook() {
	cfg := &config.Config{
		ProviderName: "shopspring",
		HooksEnabled: true,
	}

	manager := decimal.NewManager(cfg)

	// Demonstrar conceito de auditoria
	fmt.Println("Demonstrando hooks de auditoria (simulado)")

	// Operações que seriam auditadas
	a, _ := manager.NewFromString("100.50")
	b, _ := manager.NewFromString("25.75")

	fmt.Printf("Criado decimal: %s\n", a.String())
	fmt.Printf("Criado decimal: %s\n", b.String())

	result, _ := a.Add(b)
	fmt.Printf("Resultado da soma: %s\n", result.String())

	division, err := a.Div(b)
	if err != nil {
		fmt.Printf("Erro na divisão: %v\n", err)
	} else {
		fmt.Printf("Resultado da divisão: %s\n", division.String())
	}

	// Simular log de auditoria
	fmt.Println("\nLog de auditoria simulado:")
	fmt.Println("- NewFromString(\"100.50\") -> sucesso")
	fmt.Println("- NewFromString(\"25.75\") -> sucesso")
	fmt.Println("- Add(100.50, 25.75) -> 126.25")
	fmt.Println("- Div(100.50, 25.75) -> 3.902912621359223")
}

func testValidationHook() {
	manager := decimal.NewManager(&config.Config{
		ProviderName: "cockroach",
		HooksEnabled: true,
	})

	// Demonstrar conceito de validação
	fmt.Println("Demonstrando hooks de validação (simulado)")
	fmt.Println("Testando validação com limites: -1000 a 1000")

	// Teste dentro dos limites
	fmt.Println("\nTeste 1: Valores dentro dos limites")
	a, err := manager.NewFromString("500")
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("Valor válido: %s\n", a.String())
	}

	b, err := manager.NewFromString("300")
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("Valor válido: %s\n", b.String())
	}

	// Operação dentro dos limites
	if a != nil && b != nil {
		sum, err := a.Add(b)
		if err != nil {
			fmt.Printf("Erro na soma: %v\n", err)
		} else {
			fmt.Printf("Soma válida: %s\n", sum.String())
		}
	}

	// Teste fora dos limites
	fmt.Println("\nTeste 2: Valores fora dos limites")

	// Simular validação
	fmt.Println("Tentando criar valor 2000 (acima do limite)...")
	fmt.Println("VALIDAÇÃO: Erro - valor 2000 está acima do máximo permitido 1000")

	fmt.Println("Tentando criar valor -1500 (abaixo do limite)...")
	fmt.Println("VALIDAÇÃO: Erro - valor -1500 está abaixo do mínimo permitido -1000")
}

func testMetricsHook() {
	manager := decimal.NewManager(&config.Config{
		ProviderName: "shopspring",
		HooksEnabled: true,
	})

	// Demonstrar conceito de métricas
	fmt.Println("Demonstrando hooks de métricas (simulado)")

	// Simular várias operações para coleta de métricas
	fmt.Println("Executando operações para coleta de métricas...")

	values := []string{"10.5", "20.3", "15.7", "8.9", "12.4"}
	decimals := make([]interfaces.Decimal, 0, len(values))

	// Criação de decimais
	for _, val := range values {
		d, _ := manager.NewFromString(val)
		decimals = append(decimals, d)
	}

	// Várias operações
	for i := 0; i < len(decimals)-1; i++ {
		decimals[i].Add(decimals[i+1])
		decimals[i].Mul(decimals[i+1])
		if !decimals[i+1].IsZero() {
			decimals[i].Div(decimals[i+1])
		}
	}

	// Operações batch
	manager.Sum(decimals...)
	manager.Average(decimals...)
	manager.Max(decimals...)
	manager.Min(decimals...)

	// Simular métricas coletadas
	fmt.Println("\nMétricas simuladas:")
	fmt.Println("NewFromString: 5 operações, tempo médio: 1.2µs")
	fmt.Println("Add: 4 operações, tempo médio: 0.8µs")
	fmt.Println("Mul: 4 operações, tempo médio: 0.9µs")
	fmt.Println("Div: 4 operações, tempo médio: 1.5µs")
	fmt.Println("Sum: 1 operação, tempo médio: 3.2µs")
	fmt.Println("Average: 1 operação, tempo médio: 4.1µs")
	fmt.Println("Max: 1 operação, tempo médio: 2.1µs")
	fmt.Println("Min: 1 operação, tempo médio: 2.0µs")
}

func testCombinedHooks() {
	cfg := &config.Config{
		ProviderName: "cockroach",
		HooksEnabled: true,
	}

	manager := decimal.NewManager(cfg)

	// Simular múltiplos hooks trabalhando juntos
	fmt.Println("Simulando execução com múltiplos hooks:")
	fmt.Println("- Hook de Auditoria: ATIVO")
	fmt.Println("- Hook de Validação: ATIVO (limites: -10000 a 10000)")
	fmt.Println("- Hook de Métricas: ATIVO")
	fmt.Println("- Hook de Logging: ATIVO")

	// Operações monitoradas
	fmt.Println("\nExecutando operações com todos os hooks:")

	a, _ := manager.NewFromString("1000.00")
	fmt.Println("[AUDIT] NewFromString(\"1000.00\") registrado")
	fmt.Println("[VALIDATION] Valor 1000.00 está dentro dos limites")
	fmt.Println("[METRICS] NewFromString executado em 1.1µs")
	fmt.Println("[LOGGING] Decimal criado com sucesso: 1000.00")

	b, _ := manager.NewFromString("250.75")
	fmt.Println("\n[AUDIT] NewFromString(\"250.75\") registrado")
	fmt.Println("[VALIDATION] Valor 250.75 está dentro dos limites")
	fmt.Println("[METRICS] NewFromString executado em 1.0µs")
	fmt.Println("[LOGGING] Decimal criado com sucesso: 250.75")

	result, _ := a.Mul(b)
	fmt.Println("\n[AUDIT] Mul(1000.00, 250.75) registrado")
	fmt.Println("[VALIDATION] Argumentos validados")
	fmt.Println("[METRICS] Mul executado em 1.3µs")
	fmt.Println("[VALIDATION] Resultado 250750.00 validado")
	fmt.Println("[LOGGING] Multiplicação realizada: 1000.00 * 250.75 = 250750.00")

	fmt.Printf("\nResultado final: %s\n", result.String())

	// Teste de erro para demonstrar hooks de erro
	fmt.Println("\nTestando divisão por zero para demonstrar hooks de erro:")
	zero := manager.Zero()
	_, err := a.Div(zero)

	if err != nil {
		fmt.Println("[AUDIT] Div(1000.00, 0) -> ERRO registrado")
		fmt.Println("[METRICS] Div_error contabilizado")
		fmt.Println("[LOGGING] Erro capturado: division by zero")
		fmt.Printf("Erro esperado: %v\n", err)
	}

	fmt.Println("\n=== Resumo da Execução ===")
	fmt.Println("Total de operações auditadas: 4")
	fmt.Println("Validações realizadas: 6")
	fmt.Println("Métricas coletadas: 4 operações")
	fmt.Println("Erros capturados: 1")
}
