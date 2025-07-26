# Enhanced Features Demo

Este exemplo demonstra as melhorias implementadas no módulo decimal, incluindo otimizações de performance, melhor tratamento de edge cases e documentação GoDoc aprimorada.

## Funcionalidades Demonstradas

### 1. Operações Básicas
- Criação de decimais a partir de strings e floats
- Operações aritméticas básicas (multiplicação, adição)
- Formatação de saída

### 2. Edge Cases
- Números muito pequenos (0.001, 0.002)
- Números grandes (123456789.123456)
- Notação científica (1.5e3)

### 3. Operações Batch Otimizadas
- **Abordagem Tradicional**: Operações separadas (Sum, Average, Max, Min)
- **Abordagem Otimizada**: Uma única passagem pelos dados com `ProcessBatchSlice`

#### Benefícios de Performance
- **Operações Separadas**: ~16249 ns/op, 9696 B/op, 202 allocs/op
- **BatchProcessor**: ~9998 ns/op, 5024 B/op, 104 allocs/op
- **Melhoria**: ~38% mais rápido com ~48% menos alocações

### 4. Análise Financeira Real
- Processamento de dados de receita mensal
- Cálculo de estatísticas anuais (total, média, máximo, mínimo)
- Exemplo prático de uso em cenários financeiros

### 5. Gerenciamento de Providers
- Alternância entre providers para diferentes casos de uso:
  - **CockroachDB APD**: Alta precisão para cálculos críticos
  - **Shopspring**: Melhor performance para operações em massa

## Como Executar

```bash
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/decimal/examples/enhanced_features
go run main.go
```

## Saída Esperada

O programa exibirá:
1. Demonstração de operações básicas com cálculo de imposto
2. Exemplos de edge cases com números pequenos, grandes e notação científica
3. Comparação entre abordagens tradicionais e otimizadas para operações batch
4. Análise completa de dados financeiros anuais
5. Demonstração de troca de providers

## Melhorias Implementadas

### ✅ Enhanced Edge Case Handling
- Tratamento robusto de números muito pequenos e grandes
- Suporte aprimorado para notação científica
- Validação melhorada de entrada

### ✅ Performance Optimized Batch Operations
- Novas funções `*Slice()` que evitam alocação de varargs
- `BatchProcessor` para operações estatísticas em uma única passagem
- Redução significativa de alocações de memória

### ✅ Comprehensive GoDoc Documentation
- Documentação detalhada para todos os métodos públicos
- Exemplos práticos de uso
- Comparações de performance documentadas

### ✅ Real-world Usage Examples
- Cenários financeiros realistas
- Demonstração de melhores práticas
- Casos de uso para diferentes providers

## Estrutura do Código

- **Operações Básicas**: Demonstra uso fundamental da biblioteca
- **Edge Cases**: Testa limites e casos especiais
- **Batch Operations**: Compara abordagens de performance
- **Financial Analysis**: Mostra aplicação prática
- **Provider Management**: Demonstra flexibilidade da arquitetura

Este exemplo serve como um guia abrangente para utilizar as funcionalidades avançadas do módulo decimal de forma eficiente e robusta.
