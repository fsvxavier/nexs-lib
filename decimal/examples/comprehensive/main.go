package main

import (
	"encoding/json"
	"fmt"

	dec "github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func main() {
	fmt.Println("=== Demonstração Completa do Provider Decimal ===")

	// 1. Exemplo básico usando factory
	fmt.Println("\n1. Exemplo básico usando factory:")
	basicFactoryExample()

	// 2. Exemplo de operações matemáticas
	fmt.Println("\n2. Operações matemáticas:")
	mathOperationsExample()

	// 3. Exemplo de comparações
	fmt.Println("\n3. Operações de comparação:")
	comparisonExample()

	// 4. Exemplo de serialização JSON
	fmt.Println("\n4. Serialização JSON:")
	jsonExample()

	// 5. Exemplo de formatação e arredondamento
	fmt.Println("\n5. Formatação e arredondamento:")
	formattingExample()

	// 6. Exemplo de tratamento de erros
	fmt.Println("\n6. Tratamento de erros:")
	errorHandlingExample()
}

func basicFactoryExample() {
	// Criar providers usando factory
	shopProvider := dec.NewProvider(dec.ShopSpring)
	apdProvider := dec.NewProvider(dec.APD)

	// Criar decimais
	shopDecimal, _ := shopProvider.NewFromString("100.50")
	apdDecimal, _ := apdProvider.NewFromString("100.50")

	fmt.Printf("ShopSpring: %s\n", shopDecimal.String())
	fmt.Printf("APD: %s\n", apdDecimal.String())
}

func mathOperationsExample() {
	provider := dec.NewProvider(dec.ShopSpring)

	a, _ := provider.NewFromString("100.25")
	b, _ := provider.NewFromString("50.75")

	fmt.Printf("a = %s, b = %s\n", a.String(), b.String())
	fmt.Printf("a + b = %s\n", a.Add(b).String())
	fmt.Printf("a - b = %s\n", a.Sub(b).String())
	fmt.Printf("a * b = %s\n", a.Mul(b).String())

	result, err := a.Div(b)
	if err != nil {
		fmt.Printf("Erro na divisão: %v\n", err)
	} else {
		fmt.Printf("a / b = %s\n", result.String())
	}

	fmt.Printf("abs(-a) = %s\n", a.Sub(a).Sub(a).Abs().String())
}

func comparisonExample() {
	provider := dec.NewProvider(dec.APD)

	a, _ := provider.NewFromString("100.25")
	b, _ := provider.NewFromString("100.25")
	c, _ := provider.NewFromString("200.50")

	fmt.Printf("a = %s, b = %s, c = %s\n", a.String(), b.String(), c.String())
	fmt.Printf("a == b: %t\n", a.Equals(b))
	fmt.Printf("a > c: %t\n", a.GreaterThan(c))
	fmt.Printf("c > a: %t\n", c.GreaterThan(a))
	fmt.Printf("a <= b: %t\n", a.LessThanOrEqual(b))
	fmt.Printf("a >= b: %t\n", a.GreaterThanOrEqual(b))
}

func jsonExample() {
	provider := dec.NewProvider(dec.ShopSpring)

	// Criar um decimal
	decimal, _ := provider.NewFromString("123.456789")

	// Serializar para JSON
	jsonData, err := json.Marshal(decimal)
	if err != nil {
		fmt.Printf("Erro ao serializar: %v\n", err)
		return
	}

	fmt.Printf("JSON: %s\n", string(jsonData))

	// Deserializar de JSON
	newDecimal, _ := provider.NewFromString("0")
	err = json.Unmarshal(jsonData, newDecimal)
	if err != nil {
		fmt.Printf("Erro ao deserializar: %v\n", err)
		return
	}

	fmt.Printf("Deserializado: %s\n", newDecimal.String())
}

func formattingExample() {
	provider := dec.NewProvider(dec.APD)

	pi, _ := provider.NewFromString("3.141592653589793238462643383279502884197169399375105820974944")

	fmt.Printf("Pi completo: %s\n", pi.String())
	fmt.Printf("Pi (2 casas): %s\n", pi.Round(2).String())
	fmt.Printf("Pi (5 casas): %s\n", pi.Round(5).String())
	fmt.Printf("Pi truncado (3 casas): %s\n", pi.Truncate(3).String())

	// Teste com números grandes
	bigNumber, _ := provider.NewFromString("123456789.987654321")
	fmt.Printf("Número grande: %s\n", bigNumber.String())
	fmt.Printf("Arredondado (2 casas): %s\n", bigNumber.Round(2).String())
}

func errorHandlingExample() {
	provider := dec.NewProvider(dec.ShopSpring)

	// Teste com string inválida
	_, err := provider.NewFromString("not-a-number")
	if err != nil {
		fmt.Printf("Erro esperado ao criar decimal de string inválida: %v\n", err)
	}

	// Teste de divisão por zero
	a, _ := provider.NewFromString("100")
	zero, _ := provider.NewFromString("0")

	_, err = a.Div(zero)
	if err != nil {
		fmt.Printf("Erro esperado na divisão por zero: %v\n", err)
	}

	// Teste com valores extremos
	maxValue, _ := provider.NewFromString("999999999999999999999999999999.999999999999999999999999999")
	fmt.Printf("Valor máximo testado: %s\n", maxValue.String())
}

// Função auxiliar para demonstrar uso com interface genérica
func demonstrateProvider(provider interfaces.Provider, name string) {
	fmt.Printf("\n--- Demonstrando %s ---\n", name)

	decimal, _ := provider.NewFromString("42.42")
	doubled := decimal.Add(decimal)

	fmt.Printf("Valor: %s\n", decimal.String())
	fmt.Printf("Dobrado: %s\n", doubled.String())
	fmt.Printf("É zero? %t\n", decimal.IsZero())
	fmt.Printf("É positivo? %t\n", decimal.IsPositive())
}
