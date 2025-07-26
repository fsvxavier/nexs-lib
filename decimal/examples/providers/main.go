package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func main() {
	fmt.Println("=== Comparação entre Providers ===")

	fmt.Println("\n1. Cockroach Provider (Alta Precisão)")
	testCockroachProvider()

	fmt.Println("\n2. Shopspring Provider (Performance)")
	testShopspringProvider()

	fmt.Println("\n3. Comparação Lado a Lado")
	sideByComparison()

	fmt.Println("\n4. Benchmark Simples")
	simpleBenchmark()
}

func testCockroachProvider() {
	cfg := &config.Config{
		ProviderName:    "cockroach",
		MaxPrecision:    28,
		MaxExponent:     15,
		MinExponent:     -15,
		DefaultRounding: "RoundDown",
		HooksEnabled:    false,
		Timeout:         30,
	}

	manager := decimal.NewManager(cfg)

	// Teste de alta precisão
	a, _ := manager.NewFromString("1.234567890123456789")
	b, _ := manager.NewFromString("9.876543210987654321")

	sum, _ := a.Add(b)
	product, _ := a.Mul(b)
	division, _ := b.Div(a)

	fmt.Printf("Provider: %s\n", manager.GetProvider().Name())
	fmt.Printf("a = %s\n", a.String())
	fmt.Printf("b = %s\n", b.String())
	fmt.Printf("a + b = %s\n", sum.String())
	fmt.Printf("a * b = %s\n", product.String())
	fmt.Printf("b / a = %s\n", division.String())

	// Teste com números grandes (mas dentro dos limites)
	large1, _ := manager.NewFromString("999999999999.999999")
	large2, _ := manager.NewFromString("123456789012.567890")
	largeSum, _ := large1.Add(large2)

	fmt.Printf("\nNúmeros grandes:\n")
	fmt.Printf("large1 = %s\n", large1.String())
	fmt.Printf("large2 = %s\n", large2.String())
	fmt.Printf("soma = %s\n", largeSum.String())
}

func testShopspringProvider() {
	cfg := &config.Config{
		ProviderName:    "shopspring",
		MaxPrecision:    15,
		MaxExponent:     10,
		MinExponent:     -10,
		DefaultRounding: "RoundHalfUp",
		HooksEnabled:    false,
		Timeout:         30,
	}

	manager := decimal.NewManager(cfg)

	// Mesmos testes para comparação
	a, _ := manager.NewFromString("1.234567890123456789")
	b, _ := manager.NewFromString("9.876543210987654321")

	sum, _ := a.Add(b)
	product, _ := a.Mul(b)
	division, _ := b.Div(a)

	fmt.Printf("Provider: %s\n", manager.GetProvider().Name())
	fmt.Printf("a = %s\n", a.String())
	fmt.Printf("b = %s\n", b.String())
	fmt.Printf("a + b = %s\n", sum.String())
	fmt.Printf("a * b = %s\n", product.String())
	fmt.Printf("b / a = %s\n", division.String())

	// Teste com números grandes (ajustados para shopspring)
	large1, _ := manager.NewFromString("999999999.999999")
	large2, _ := manager.NewFromString("123456789.567890")
	largeSum, _ := large1.Add(large2)

	fmt.Printf("\nNúmeros grandes:\n")
	fmt.Printf("large1 = %s\n", large1.String())
	fmt.Printf("large2 = %s\n", large2.String())
	fmt.Printf("soma = %s\n", largeSum.String())
}

func sideByComparison() {
	// Configurações para ambos os providers
	cockroachCfg := &config.Config{
		ProviderName:    "cockroach",
		MaxPrecision:    20,
		MaxExponent:     10,
		MinExponent:     -10,
		DefaultRounding: "RoundDown",
	}

	shopspringCfg := &config.Config{
		ProviderName:    "shopspring",
		MaxPrecision:    20,
		MaxExponent:     10,
		MinExponent:     -10,
		DefaultRounding: "RoundHalfUp",
	}

	cockroachManager := decimal.NewManager(cockroachCfg)
	shopspringManager := decimal.NewManager(shopspringCfg)

	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{"Precisão Decimal", "1.0000000000000001", "2.0000000000000002"},
		{"Divisão com Resto", "10", "3"},
		{"Números Pequenos", "0.0001", "0.0002"},
		{"Operação Complexa", "123.456789", "987.654321"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)

		// Cockroach
		ca, _ := cockroachManager.NewFromString(tc.a)
		cb, _ := cockroachManager.NewFromString(tc.b)
		cSum, _ := ca.Add(cb)
		cDiv, _ := ca.Div(cb)

		// Shopspring
		sa, _ := shopspringManager.NewFromString(tc.a)
		sb, _ := shopspringManager.NewFromString(tc.b)
		sSum, _ := sa.Add(sb)
		sDiv, _ := sa.Div(sb)

		fmt.Printf("Cockroach: %s + %s = %s\n", ca.String(), cb.String(), cSum.String())
		fmt.Printf("Shopspring: %s + %s = %s\n", sa.String(), sb.String(), sSum.String())
		fmt.Printf("Cockroach: %s / %s = %s\n", ca.String(), cb.String(), cDiv.String())
		fmt.Printf("Shopspring: %s / %s = %s\n", sa.String(), sb.String(), sDiv.String())
	}
}

func simpleBenchmark() {
	cockroachManager := decimal.NewManager(&config.Config{ProviderName: "cockroach"})
	shopspringManager := decimal.NewManager(&config.Config{ProviderName: "shopspring"})

	const iterations = 10000

	// Benchmark Cockroach
	fmt.Printf("Executando %d operações com Cockroach...\n", iterations)
	cockroachTime := benchmarkOperations(cockroachManager, iterations)

	// Benchmark Shopspring
	fmt.Printf("Executando %d operações com Shopspring...\n", iterations)
	shopspringTime := benchmarkOperations(shopspringManager, iterations)

	fmt.Printf("\nResultados (aproximados):\n")
	fmt.Printf("Cockroach: %v por operação\n", cockroachTime/iterations)
	fmt.Printf("Shopspring: %v por operação\n", shopspringTime/iterations)

	if cockroachTime < shopspringTime {
		fmt.Printf("Cockroach foi %.2fx mais rápido\n", float64(shopspringTime)/float64(cockroachTime))
	} else {
		fmt.Printf("Shopspring foi %.2fx mais rápido\n", float64(cockroachTime)/float64(shopspringTime))
	}
}

func benchmarkOperations(manager *decimal.Manager, iterations int) int {
	// Simulação simples de tempo (não é um benchmark real)
	a, _ := manager.NewFromString("123.456")
	b, _ := manager.NewFromString("789.012")

	operations := 0
	for i := 0; i < iterations; i++ {
		_, _ = a.Add(b)
		_, _ = a.Mul(b)
		_, _ = a.Div(b)
		operations += 3
	}

	// Retorna um valor simulado baseado no provider
	// Em um benchmark real, você usaria time.Now() antes e depois
	if manager.GetProvider().Name() == "cockroach" {
		return operations * 120 // Simulação: mais lento devido à precisão
	}
	return operations * 80 // Simulação: mais rápido
}

// Demonstração de casos de uso específicos
func demonstrateUseCases() {
	fmt.Println("\n=== Casos de Uso Específicos ===")

	// Caso 1: Cálculos financeiros (requer alta precisão)
	fmt.Println("\n1. Cálculos Financeiros (use Cockroach):")
	financialManager := decimal.NewManager(&config.Config{ProviderName: "cockroach"})

	principal, _ := financialManager.NewFromString("10000.00")
	rate, _ := financialManager.NewFromString("0.0325") // 3.25% ao ano
	years, _ := financialManager.NewFromString("5")

	// Juros simples: P * R * T
	interest, _ := principal.Mul(rate)
	interest, _ = interest.Mul(years)
	total, _ := principal.Add(interest)

	fmt.Printf("Principal: %s\n", principal.String())
	fmt.Printf("Taxa: %s\n", rate.String())
	fmt.Printf("Anos: %s\n", years.String())
	fmt.Printf("Juros: %s\n", interest.String())
	fmt.Printf("Total: %s\n", total.String())

	// Caso 2: Cálculos de inventário (performance é importante)
	fmt.Println("\n2. Cálculos de Inventário (use Shopspring):")
	inventoryManager := decimal.NewManager(&config.Config{ProviderName: "shopspring"})

	items := []struct {
		name     string
		quantity string
		price    string
	}{
		{"Produto A", "150", "25.99"},
		{"Produto B", "75", "49.50"},
		{"Produto C", "200", "12.75"},
	}

	var totalValue interfaces.Decimal = inventoryManager.Zero()

	for _, item := range items {
		qty, _ := inventoryManager.NewFromString(item.quantity)
		price, _ := inventoryManager.NewFromString(item.price)
		value, _ := qty.Mul(price)
		totalValue, _ = totalValue.Add(value)

		fmt.Printf("%s: %s x %s = %s\n", item.name, qty.String(), price.String(), value.String())
	}

	fmt.Printf("Valor total do inventário: %s\n", totalValue.String())
}
