# Decimal Providers

Este módulo fornece implementações de providers decimais usando as bibliotecas `github.com/shopspring/decimal` e `github.com/cockroachdb/apd`. Estas implementações seguem uma interface comum, permitindo trocar facilmente entre as bibliotecas de acordo com suas necessidades.

## Estrutura

```
dec/
  ├── api/          # Define a interface comum para todos os providers
  ├── shopspring/   # Implementação usando github.com/shopspring/decimal
  ├── apd/          # Implementação usando github.com/cockroachdb/apd
  └── factory/      # Factory para criar providers de forma simples
```

## Uso

### Interface comum

A interface `Provider` define as operações para criar novos decimais:

```go
type Provider interface {
	// NewFromString cria um novo decimal a partir de uma string
	NewFromString(numString string) (Decimal, error)
	
	// NewFromFloat cria um novo decimal a partir de um float64
	NewFromFloat(numFloat float64) (Decimal, error)
	
	// NewFromInt cria um novo decimal a partir de um int64
	NewFromInt(numInt int64) (Decimal, error)
}
```

### Decimal

A interface `Decimal` define as operações que podem ser realizadas em um decimal:

```go
type Decimal interface {
	// String retorna a representação em string do decimal
	String() string
	
	// Float64 retorna a representação em float64 do decimal
	Float64() (float64, error)
	
	// Int64 retorna a representação em int64 do decimal
	Int64() (int64, error)
	
	// IsZero retorna true se o decimal é zero
	IsZero() bool
	
	// IsNegative retorna true se o decimal é negativo
	IsNegative() bool
	
	// IsPositive retorna true se o decimal é positivo
	IsPositive() bool
	
	// Equals retorna true se o decimal é igual a outro
	Equals(d Decimal) bool
	
	// GreaterThan retorna true se o decimal é maior que outro
	GreaterThan(d Decimal) bool
	
	// LessThan retorna true se o decimal é menor que outro
	LessThan(d Decimal) bool
	
	// GreaterThanOrEqual retorna true se o decimal é maior ou igual a outro
	GreaterThanOrEqual(d Decimal) bool
	
	// LessThanOrEqual retorna true se o decimal é menor ou igual a outro
	LessThanOrEqual(d Decimal) bool
	
	// Add adiciona outro decimal ao decimal atual e retorna o resultado
	Add(d Decimal) Decimal
	
	// Sub subtrai outro decimal do decimal atual e retorna o resultado
	Sub(d Decimal) Decimal
	
	// Mul multiplica outro decimal pelo decimal atual e retorna o resultado
	Mul(d Decimal) Decimal
	
	// Div divide o decimal atual por outro decimal e retorna o resultado
	Div(d Decimal) (Decimal, error)
	
	// Abs retorna o valor absoluto do decimal
	Abs() Decimal
	
	// Round arredonda o decimal para o número especificado de casas decimais
	Round(places int32) Decimal
	
	// Truncate trunca o decimal para o número especificado de casas decimais
	Truncate(places int32) Decimal
	
	// MarshalJSON implementa a interface json.Marshaler
	MarshalJSON() ([]byte, error)
	
	// UnmarshalJSON implementa a interface json.Unmarshaler
	UnmarshalJSON(data []byte) error
}
```

### Factory

```go
import (
    "github.com/fsvxavier/nexs-lib/decimal/api"
    "github.com/fsvxavier/nexs-lib/decimal/factory"
)

func main() {
    // Criar provider ShopSpring
    shopSpringProvider := factory.NewProvider(factory.ShopSpring)
    
    // Criar provider APD (CockroachDB)
    apdProvider := factory.NewProvider(factory.APD)
    
    // Usar o provider através da interface comum
    useDecimalProvider(shopSpringProvider)
    useDecimalProvider(apdProvider)
}

func useDecimalProvider(provider api.Provider) {
    // Criar decimais
    dec1, _ := provider.NewFromString("123.456")
    dec2, _ := provider.NewFromFloat(789.123)
    
    // Operações
    sum := dec1.Add(dec2)
    fmt.Println("Sum:", sum.String())
}
```

## Escolhendo o Provider Adequado

- **ShopSpring**: Mais simples e fácil de usar, bom para a maioria dos casos de uso gerais.
- **APD (CockroachDB)**: Maior precisão e mais controle sobre o comportamento de arredondamento, melhor para cálculos financeiros e científicos que exigem alta precisão.

## Testes de Benchmark

Os dois providers incluem testes de benchmark para comparar o desempenho das operações. Execute os benchmarks usando:

```bash
go test ./dec/shopspring -bench=.
go test ./dec/apd -bench=.
```

## Testes de Race Condition

Os dois providers incluem testes de race condition para garantir a segurança em ambientes concorrentes. Execute os testes com a flag `-race`:

```bash
go test ./dec/shopspring -race
go test ./dec/apd -race
```
