# Decimal Module

[![Go Reference](https://pkg.go.dev/badge/github.com/fsvxavier/nexs-lib/decimal.svg)](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/decimal)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-lib/decimal)](https://goreportcard.com/report/github.com/fsvxavier/nexs-lib/decimal)

Módulo decimal modular e extensível com suporte a múltiplos providers, configuração flexível, sistema de hooks e operações batch. Projetado para alta performance e precision em cálculos financeiros e matemáticos.

## 🎯 Características Principais

- **Múltiplos Providers**: Suporte a CockroachDB APD (padrão) e Shopspring Decimal
- **Arquitetura Modular**: Interfaces bem definidas e inversão de dependências
- **Sistema de Hooks**: Pre, post e error hooks para extensibilidade
- **Operações Batch**: Sum, Average, Max, Min com otimizações
- **Configuração Flexível**: Precision, rounding, timeouts configuráveis
- **Thread-Safe**: Operações concorrentes seguras
- **Cobertura de Testes**: >98% de cobertura com testes unitários, integração e benchmarks

## 🚀 Quick Start

### Instalação

```bash
go get github.com/fsvxavier/nexs-lib/decimal
```

### Uso Básico

```go
package main

import (
    "fmt"
    "github.com/fsvxavier/nexs-lib/decimal"
)

func main() {
    // Usando funções de conveniência (provider padrão: cockroach)
    a, _ := decimal.NewFromString("123.45")
    b, _ := decimal.NewFromString("67.89")
    
    sum, _ := a.Add(b)
    fmt.Println(sum.String()) // "191.34"
    
    // Operações batch
    nums := []interfaces.Decimal{a, b}
    total, _ := decimal.Sum(nums...)
    avg, _ := decimal.Average(nums...)
    
    fmt.Printf("Total: %s, Average: %s\n", total.String(), avg.String())
}
```

### Configuração com Providers

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/decimal"
    "github.com/fsvxavier/nexs-lib/decimal/config"
)

func main() {
    // Configuração customizada com provider Shopspring
    cfg := config.NewConfig(
        config.WithProvider("shopspring"),
        config.WithMaxPrecision(50),
        config.WithRounding("RoundHalfUp"),
        config.WithHooksEnabled(true),
    )
    
    manager := decimal.NewManager(cfg)
    
    dec, _ := manager.NewFromString("999.999")
    rounded := dec.Round(2)
    fmt.Println(rounded.String()) // "1000.00"
}
```

## 📖 Providers

### CockroachDB APD (Padrão)
- **Vantagens**: Alta precisão, performance superior, suporte nativo a contextos
- **Uso**: Aplicações financeiras, cálculos de precisão crítica
- **Configuração**: `config.WithProvider("cockroach")`

### Shopspring Decimal
- **Vantagens**: API familiar, ampla adoção, funcionalidades extras
- **Uso**: Aplicações gerais, migração de projetos existentes
- **Configuração**: `config.WithProvider("shopspring")`

## ⚙️ Configuração Avançada

```go
cfg := config.NewConfig(
    config.WithMaxPrecision(28),    // Precisão máxima
    config.WithMaxExponent(13),     // Expoente máximo
    config.WithMinExponent(-8),     // Expoente mínimo  
    config.WithRounding("RoundHalfEven"), // Modo de arredondamento
    config.WithProvider("cockroach"), // Provider
    config.WithHooksEnabled(true),   // Ativar hooks
    config.WithTimeout(30),          // Timeout em segundos
    config.WithProviderConfig("custom_key", "custom_value"), // Config específica
)

// Validação automática
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

### Modos de Arredondamento Suportados

- `RoundDown`: Trunca em direção ao zero
- `RoundUp`: Arredonda para longe do zero
- `RoundHalfUp`: Arredonda 0.5 para cima
- `RoundHalfDown`: Arredonda 0.5 para baixo
- `RoundHalfEven`: Banker's rounding (padrão IEEE)
- `RoundCeiling`: Arredonda em direção ao infinito positivo
- `RoundFloor`: Arredonda em direção ao infinito negativo
- `Round05Up`: Arredonda para longe do zero se dígito for 0 ou 5

## 🪝 Sistema de Hooks

```go
// Hook de logging
loggingHook := hooks.NewBasicLoggingHook(func(msg string) {
    log.Println(msg)
})

// Hook de validação
validationHook := hooks.NewValidationHook()
validationHook.ValidateString = func(s string) error {
    if len(s) > 50 {
        return errors.New("string too long")
    }
    return nil
}

// Hook de métricas
metricsHook := hooks.NewMetricsHook()

// Registrar hooks
manager.GetHookManager().RegisterPreHook(loggingHook)
manager.GetHookManager().RegisterPreHook(validationHook)
manager.GetHookManager().RegisterPostHook(metricsHook)

// Uso normal - hooks executados automaticamente
dec, err := manager.NewFromString("123.45")

// Verificar métricas
metrics := metricsHook.GetMetrics()
fmt.Printf("Operações realizadas: %v\n", metrics)
```

## 📊 Operações Batch

```go
// Criar decimais
decimals := make([]interfaces.Decimal, 100)
for i := 0; i < 100; i++ {
    decimals[i], _ = decimal.NewFromInt(int64(i))
}

// Operações otimizadas
sum, _ := decimal.Sum(decimals...)       // Soma de todos
avg, _ := decimal.Average(decimals...)   // Média
max, _ := decimal.Max(decimals...)       // Máximo
min, _ := decimal.Min(decimals...)       // Mínimo

fmt.Printf("Sum: %s, Avg: %s, Max: %s, Min: %s\n", 
    sum.String(), avg.String(), max.String(), min.String())
```

## 🔄 Parsing Flexível

```go
// Parse de diferentes tipos
values := []interface{}{
    "123.45",           // string
    123.45,             // float64
    float32(123.45),    // float32
    123,                // int
    int32(123),         // int32
    int64(123),         // int64
}

for _, val := range values {
    dec, err := decimal.Parse(val)
    if err != nil {
        log.Printf("Error parsing %v: %v", val, err)
        continue
    }
    fmt.Printf("Parsed %v -> %s\n", val, dec.String())
}
```

## 🏗️ Arquitetura

```
decimal/
├── interfaces/          # Contratos e interfaces
├── config/             # Configuração e opções
├── providers/          # Implementações dos providers
│   ├── cockroach/     # Provider CockroachDB APD
│   └── shopspring/    # Provider Shopspring
├── hooks/             # Sistema de hooks
└── examples/          # Exemplos práticos
```

## 🧪 Testes e Benchmarks

```bash
# Testes unitários com cobertura
go test -race -timeout 30s -coverprofile=coverage.out ./decimal/...

# Visualizar cobertura
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. -benchmem ./decimal/...

# Testes de integração
go test -tags=integration -v ./decimal/...
```

## 📈 Performance

### Benchmarks Típicos (Go 1.21+)

```
BenchmarkNewFromString-8         5000000    250 ns/op    64 B/op    2 allocs/op
BenchmarkArithmetic/Add-8       10000000    120 ns/op    48 B/op    1 allocs/op
BenchmarkArithmetic/Mul-8        8000000    180 ns/op    64 B/op    2 allocs/op
BenchmarkBatch/Sum-8              100000  15000 ns/op   800 B/op   50 allocs/op
```

## 🛠️ Desenvolvimento

### Requisitos
- Go 1.21+
- Dependências: `cockroachdb/apd/v3`, `shopspring/decimal`, `stretchr/testify`

### Estrutura de Testes
- **Unitários**: `*_test.go` - Cobertura >98%
- **Benchmarks**: `*_benchmark_test.go` - Performance
- **Integração**: `*_integration_test.go` - Fluxos completos

### Linting e Formatação
```bash
golangci-lint run ./decimal/...
gofmt -w ./decimal/
go vet ./decimal/...
```

## 📝 Contribuição

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Guidelines
- Mantenha cobertura de testes >98%
- Siga os padrões de código Go
- Documente APIs públicas
- Inclua benchmarks para funcionalidades críticas

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🤝 Acknowledgments

- [CockroachDB APD](https://github.com/cockroachdb/apd) - Biblioteca de decimais de alta precisão
- [Shopspring Decimal](https://github.com/shopspring/decimal) - Biblioteca decimal amplamente utilizada
- Comunidade Go pela inspiração e melhores práticas

---

**Nota**: Este módulo faz parte da `nexs-lib`, uma coleção de utilitários Go para desenvolvimento enterprise.
