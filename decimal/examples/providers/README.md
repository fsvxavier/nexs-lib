# Exemplo de Comparação entre Providers

Este exemplo demonstra as diferenças entre os providers Cockroach e Shopspring, ajudando você a escolher o melhor para seu caso de uso.

## Como Executar

```bash
cd examples/providers
go run main.go
```

## Providers Disponíveis

### 1. Cockroach Provider (Padrão)
- **Baseado em**: github.com/cockroachdb/apd/v3
- **Características**:
  - Alta precisão (até 28 dígitos)
  - Controle rigoroso de expoentes
  - Ideal para cálculos financeiros críticos
  - Mais lento, mas mais preciso

### 2. Shopspring Provider
- **Baseado em**: github.com/shopspring/decimal
- **Características**:
  - Boa precisão (configurável)
  - Performance otimizada
  - Ideal para aplicações de alto volume
  - Mais rápido, precisão adequada

## Quando Usar Cada Provider

### Use Cockroach Quando:
- ✅ Precisão é crítica (cálculos financeiros)
- ✅ Lidando com dinheiro ou moedas
- ✅ Auditoria e compliance são importantes
- ✅ Números muito grandes ou muito pequenos
- ✅ Operações matemáticas complexas

### Use Shopspring Quando:
- ✅ Performance é prioridade
- ✅ Alto volume de operações
- ✅ Cálculos de inventário/estoque
- ✅ Métricas e estatísticas
- ✅ Aplicações web com muitas requisições

## Exemplos de Saída

### Comparação de Precisão
```
--- Precisão Decimal ---
Cockroach: 1.0000000000000001 + 2.0000000000000002 = 3.0000000000000003
Shopspring: 1.0000000000000001 + 2.0000000000000002 = 3.0000000000000003
Cockroach: 1.0000000000000001 / 2.0000000000000002 = 0.4999999999999999
Shopspring: 1.0000000000000001 / 2.0000000000000002 = 0.5
```

### Benchmark Simulado
```
Executando 10000 operações com Cockroach...
Executando 10000 operações com Shopspring...

Resultados (aproximados):
Cockroach: 4 por operação
Shopspring: 2.67 por operação
Shopspring foi 1.50x mais rápido
```

## Configurações Recomendadas

### Para Cálculos Financeiros (Cockroach)
```go
cfg := &config.Config{
    ProviderName:    "cockroach",
    MaxPrecision:    28,
    MaxExponent:     15,
    MinExponent:     -15,
    DefaultRounding: "RoundDown", // Conservador
    HooksEnabled:    true,        // Para auditoria
    Timeout:         60,
}
```

### Para Alto Volume (Shopspring)
```go
cfg := &config.Config{
    ProviderName:    "shopspring",
    MaxPrecision:    15,
    MaxExponent:     10,
    MinExponent:     -6,
    DefaultRounding: "RoundHalfUp", // Padrão matemático
    HooksEnabled:    false,         // Para performance
    Timeout:         30,
}
```

## Casos de Uso Demonstrados

### 1. Sistema Bancário
```go
// Use Cockroach para garantir precisão em transações
manager := decimal.NewManager(&config.Config{ProviderName: "cockroach"})
balance, _ := manager.NewFromString("1000.00")
debit, _ := manager.NewFromString("49.99")
newBalance, _ := balance.Sub(debit)
```

### 2. E-commerce
```go
// Use Shopspring para cálculos de carrinho de compras
manager := decimal.NewManager(&config.Config{ProviderName: "shopspring"})
price, _ := manager.NewFromString("29.99")
quantity, _ := manager.NewFromString("3")
total, _ := price.Mul(quantity)
```

### 3. Analytics/Métricas
```go
// Use Shopspring para processamento de grandes volumes
manager := decimal.NewManager(&config.Config{ProviderName: "shopspring"})
values := []interfaces.Decimal{...} // Milhares de valores
average, _ := manager.Average(values...)
```

## Switching de Providers

```go
manager := decimal.NewManager(&config.Config{ProviderName: "shopspring"})

// Processamento em lote rápido
for _, item := range items {
    // ... operações rápidas
}

// Mudança para cálculo final preciso
manager.SwitchProvider("cockroach")
finalResult, _ := manager.Sum(allResults...)
```

## Dicas de Performance

1. **Use Shopspring** para loops intensivos
2. **Use Cockroach** para cálculos finais importantes
3. **Evite switching** frequente de providers
4. **Configure precisão** adequada para seu caso de uso
5. **Desabilite hooks** quando performance for crítica

## Próximos Passos

- Veja o exemplo de hooks para observabilidade
- Consulte o benchmark real em `benchmark_test.go`
- Explore configurações avançadas no README principal
