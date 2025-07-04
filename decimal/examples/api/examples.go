package api

import (
	"fmt"

	dec "github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

// RunExamples executa todos os exemplos
func RunExamples() {
	// Exemplo com ShopSpring usando Factory
	fmt.Println("=== Usando ShopSpring Provider via Factory ===")
	factoryShopSpringExample()

	fmt.Println()

	// Exemplo com APD usando Factory
	fmt.Println("=== Usando APD Provider via Factory ===")
	factoryApdExample()

	fmt.Println()

	// Exemplo usando interface genérica
	fmt.Println("=== Usando Interface Genérica ===")
	useFactoryDecimalProvider(dec.NewProvider(dec.ShopSpring))
	useFactoryDecimalProvider(dec.NewProvider(dec.APD))

	fmt.Println()

	// Exemplo usando helpers do pacote principal
	fmt.Println("=== Usando Helpers do Pacote Principal ===")
	helpersAPIExample()
}

func factoryShopSpringExample() {
	// Criar um novo provider
	provider := dec.NewProvider(dec.ShopSpring)

	// Criar decimais
	dec1, _ := provider.NewFromString("123.456")
	fmt.Printf("dec1 (from string): %s\n", dec1.String())

	dec2, _ := provider.NewFromFloat(789.123)
	fmt.Printf("dec2 (from float): %s\n", dec2.String())

	dec3, _ := provider.NewFromInt(42)
	fmt.Printf("dec3 (from int): %s\n", dec3.String())

	// Operações aritméticas
	sum := dec1.Add(dec2)
	fmt.Printf("dec1 + dec2 = %s\n", sum.String())

	diff := dec2.Sub(dec1)
	fmt.Printf("dec2 - dec1 = %s\n", diff.String())

	product := dec1.Mul(dec3)
	fmt.Printf("dec1 * dec3 = %s\n", product.String())

	quotient, _ := dec2.Div(dec3)
	fmt.Printf("dec2 / dec3 = %s\n", quotient.String())

	// Operações de comparação
	fmt.Printf("dec1 == dec2: %t\n", dec1.Equals(dec2))
	fmt.Printf("dec1 > dec3: %t\n", dec1.GreaterThan(dec3))
	fmt.Printf("dec3 < dec2: %t\n", dec3.LessThan(dec2))

	// Formatação e arredondamento
	pi, _ := provider.NewFromString("3.1415926535897932384626433832795028841971693993751058")
	fmt.Printf("pi: %s\n", pi.String())
	fmt.Printf("pi (rounded to 5 places): %s\n", pi.Round(5).String())
	fmt.Printf("pi (truncated to 5 places): %s\n", pi.Truncate(5).String())
}

func factoryApdExample() {
	// Criar um novo provider
	provider := dec.NewProvider(dec.APD)

	// Criar decimais
	dec1, _ := provider.NewFromString("123.456")
	fmt.Printf("dec1 (from string): %s\n", dec1.String())

	dec2, _ := provider.NewFromFloat(789.123)
	fmt.Printf("dec2 (from float): %s\n", dec2.String())

	dec3, _ := provider.NewFromInt(42)
	fmt.Printf("dec3 (from int): %s\n", dec3.String())

	// Operações aritméticas
	sum := dec1.Add(dec2)
	fmt.Printf("dec1 + dec2 = %s\n", sum.String())

	diff := dec2.Sub(dec1)
	fmt.Printf("dec2 - dec1 = %s\n", diff.String())

	product := dec1.Mul(dec3)
	fmt.Printf("dec1 * dec3 = %s\n", product.String())

	quotient, _ := dec2.Div(dec3)
	fmt.Printf("dec2 / dec3 = %s\n", quotient.String())

	// Operações de comparação
	fmt.Printf("dec1 == dec2: %t\n", dec1.Equals(dec2))
	fmt.Printf("dec1 > dec3: %t\n", dec1.GreaterThan(dec3))
	fmt.Printf("dec3 < dec2: %t\n", dec3.LessThan(dec2))

	// Formatação e arredondamento
	pi, _ := provider.NewFromString("3.1415926535897932384626433832795028841971693993751058")
	fmt.Printf("pi: %s\n", pi.String())
	fmt.Printf("pi (rounded to 5 places): %s\n", pi.Round(5).String())
	fmt.Printf("pi (truncated to 5 places): %s\n", pi.Truncate(5).String())
}

// Exemplo de como usar a interface genérica com Factory
func useFactoryDecimalProvider(provider interfaces.Provider) {
	// Cria um decimal usando o provider
	d, _ := provider.NewFromString("123.456")

	// Faz operações independentes da implementação subjacente
	result := d.Add(d).Mul(d)
	fmt.Println("Resultado usando interface genérica:", result.String())
}

// Exemplo usando helpers para criar decimais com API
func helpersAPIExample() {
	// Usando helpers do pacote principal para criar decimais com API
	dec1, _ := dec.NewDecimal("123.456")
	fmt.Printf("dec1 (helper padrão): %s\n", dec1.String())

	dec2, _ := dec.ShopSpringDecimal("789.123")
	fmt.Printf("dec2 (helper shopspring): %s\n", dec2.String())

	dec3, _ := dec.APDDecimal("42.0")
	fmt.Printf("dec3 (helper apd): %s\n", dec3.String())

	// Operações usando API
	sum := dec1.Add(dec2)
	fmt.Printf("dec1 + dec2 = %s\n", sum.String())

	// Demonstrar criação com provider específico
	dec4, _ := dec.NewDecimalWithProvider("123.456", dec.ShopSpring)
	dec5, _ := dec.NewDecimalWithProvider("123.456", dec.APD)

	fmt.Printf("ShopSpring decimal: %s\n", dec4.String())
	fmt.Printf("APD decimal: %s\n", dec5.String())

	// Comparação entre valores (mesmos valores, providers diferentes)
	fmt.Printf("dec4 == dec5 (valor): %t\n", dec4.Equals(dec5))

	// Demonstração de serialização JSON
	jsonData1, _ := dec4.MarshalJSON()
	fmt.Printf("JSON ShopSpring: %s\n", string(jsonData1))

	jsonData2, _ := dec5.MarshalJSON()
	fmt.Printf("JSON APD: %s\n", string(jsonData2))

	// Tratamento de erros
	_, err := dec.NewDecimal("not-a-number")
	if err != nil {
		fmt.Printf("Erro esperado ao criar decimal de string inválida: %v\n", err)
	}

	// Teste de divisão por zero
	num, _ := dec.NewDecimal("100")
	zero, _ := dec.NewDecimal("0")

	_, err = num.Div(zero)
	if err != nil {
		fmt.Printf("Erro esperado na divisão por zero: %v\n", err)
	}
}

func main() {
	fmt.Println("=== Executando exemplos da API ===")
	fmt.Println("===================================")

	RunExamples()
}
