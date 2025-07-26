package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func main() {
	fmt.Println("=== Exemplo Básico - Provider Cockroach (Padrão) ===")
	basicExample()

	fmt.Println("\n=== Exemplo com Configuração Customizada ===")
	customConfigExample()

	fmt.Println("\n=== Exemplo de Operações Batch ===")
	batchOperationsExample()

	fmt.Println("\n=== Exemplo com Hooks ===")
	hooksExample()
}

// Exemplo básico usando o provider padrão (cockroach)
func basicExample() {
	// Usar configuração padrão
	manager := decimal.NewManager(nil) // nil = usa configuração padrão

	// Criar decimais
	a, err := manager.NewFromString("10.50")
	if err != nil {
		log.Fatalf("Erro ao criar decimal: %v", err)
	}

	b, err := manager.NewFromFloat(3.25)
	if err != nil {
		log.Fatalf("Erro ao criar decimal: %v", err)
	}

	// Operações aritméticas
	sum, err := a.Add(b)
	if err != nil {
		log.Fatalf("Erro na soma: %v", err)
	}

	product, err := a.Mul(b)
	if err != nil {
		log.Fatalf("Erro na multiplicação: %v", err)
	}

	fmt.Printf("a = %s\n", a.String())
	fmt.Printf("b = %s\n", b.String())
	fmt.Printf("a + b = %s\n", sum.String())
	fmt.Printf("a * b = %s\n", product.String())

	// Comparações
	if a.IsGreaterThan(b) {
		fmt.Printf("%s é maior que %s\n", a.String(), b.String())
	}
}

// Exemplo com configuração customizada
func customConfigExample() {
	// Configuração customizada com provider shopspring
	cfg := &config.Config{
		ProviderName:    "shopspring",
		MaxPrecision:    10,
		MaxExponent:     8,
		MinExponent:     -4,
		DefaultRounding: "RoundHalfUp",
		HooksEnabled:    false,
		Timeout:         60,
	}

	// Criar manager com configuração customizada
	manager := decimal.NewManager(cfg)

	// Criar decimal com alta precisão
	value, err := manager.NewFromString("123.4567")
	if err != nil {
		log.Fatalf("Erro ao criar decimal: %v", err)
	}

	// Operação de divisão com alta precisão
	divisor, err := manager.NewFromInt(7)
	if err != nil {
		log.Fatalf("Erro ao criar divisor: %v", err)
	}

	result, err := value.Div(divisor)
	if err != nil {
		log.Fatalf("Erro na divisão: %v", err)
	}

	fmt.Printf("Provider: %s\n", cfg.ProviderName)
	fmt.Printf("Valor: %s\n", value.String())
	fmt.Printf("Divisor: %s\n", divisor.String())
	fmt.Printf("Resultado (alta precisão): %s\n", result.String())

	// Teste de mudança de provider
	fmt.Println("\nMudando para provider cockroach:")
	err = manager.SwitchProvider("cockroach")
	if err != nil {
		log.Printf("Erro ao mudar provider: %v", err)
	} else {
		fmt.Printf("Provider atual: %s\n", manager.GetProvider().Name())

		// Mesma operação com provider diferente
		result2, err := value.Div(divisor)
		if err != nil {
			log.Fatalf("Erro na divisão: %v", err)
		}
		fmt.Printf("Resultado com cockroach: %s\n", result2.String())
	}
}

// Exemplo de operações em lote
func batchOperationsExample() {
	manager := decimal.NewManager(nil)

	// Criar lista de valores
	values := []string{"10.50", "25.75", "8.90", "12.30", "45.60"}
	decimals := make([]interfaces.Decimal, 0, len(values))

	fmt.Println("Valores originais:")
	for _, val := range values {
		d, err := manager.NewFromString(val)
		if err != nil {
			log.Fatalf("Erro ao criar decimal: %v", err)
		}
		decimals = append(decimals, d)
		fmt.Printf("  %s\n", d.String())
	}

	// Operação batch: somar todos
	total, err := manager.Sum(decimals...)
	if err != nil {
		log.Fatalf("Erro na soma batch: %v", err)
	}

	fmt.Printf("\nSoma total: %s\n", total.String())

	// Operação batch: calcular média
	average, err := manager.Average(decimals...)
	if err != nil {
		log.Fatalf("Erro no cálculo da média: %v", err)
	}

	fmt.Printf("Média: %s\n", average.String())

	// Encontrar máximo e mínimo
	maximum, err := manager.Max(decimals...)
	if err != nil {
		log.Fatalf("Erro ao encontrar máximo: %v", err)
	}

	minimum, err := manager.Min(decimals...)
	if err != nil {
		log.Fatalf("Erro ao encontrar mínimo: %v", err)
	}

	fmt.Printf("Máximo: %s\n", maximum.String())
	fmt.Printf("Mínimo: %s\n", minimum.String())
}

// Exemplo usando hooks para logging e validação
func hooksExample() {
	// Criar configuração com hooks habilitados
	cfg := &config.Config{
		ProviderName:    "shopspring",
		MaxPrecision:    8,
		MaxExponent:     4,
		MinExponent:     -2,
		DefaultRounding: "RoundHalfUp",
		HooksEnabled:    true,
		Timeout:         30,
	}

	manager := decimal.NewManager(cfg)

	// O manager já vem com hooks básicos de logging quando habilitados
	fmt.Println("Executando operações com hooks habilitados:")

	a, err := manager.NewFromString("100.00")
	if err != nil {
		log.Fatalf("Erro ao criar decimal: %v", err)
	}

	b, err := manager.NewFromString("25.50")
	if err != nil {
		log.Fatalf("Erro ao criar decimal: %v", err)
	}

	// As operações podem ser interceptadas pelos hooks
	result, err := a.Add(b)
	if err != nil {
		log.Fatalf("Erro na operação: %v", err)
	}

	fmt.Printf("Resultado final: %s\n", result.String())

	// Exemplo de operação que pode falhar (divisão por zero)
	fmt.Println("\nTestando divisão por zero:")
	zero := manager.Zero()

	_, err = a.Div(zero)
	if err != nil {
		fmt.Printf("Erro esperado capturado: %v\n", err)
	}

	// Demonstrar métodos de comparação
	fmt.Println("\nExemplos de comparação:")
	if a.IsGreaterThan(b) {
		fmt.Printf("%s > %s\n", a.String(), b.String())
	}

	if a.IsEqual(a) {
		fmt.Printf("%s == %s\n", a.String(), a.String())
	}

	if !b.IsEqual(zero) {
		fmt.Printf("%s != %s\n", b.String(), zero.String())
	}
}
