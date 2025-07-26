# Exemplo Básico - Decimal Module

Este exemplo demonstra o uso básico da biblioteca decimal com diferentes providers e configurações.

## Como Executar

```bash
cd examples/basic
go run main.go
```

## O que este exemplo demonstra

### 1. Uso Básico com Provider Padrão (Cockroach)
- Criação de decimais usando `NewFromString` e `NewFromFloat`
- Operações aritméticas básicas (soma, multiplicação)
- Comparações entre decimais

### 2. Configuração Customizada
- Como criar uma configuração personalizada
- Mudança de provider em tempo de execução
- Configuração de precisão e expoentes

### 3. Operações em Lote
- Soma de múltiplos decimais usando `Sum(...)`
- Cálculo de média com `Average(...)`
- Encontrar máximo e mínimo com `Max(...)` e `Min(...)`

### 4. Sistema de Hooks
- Habilitação de hooks para logging e validação
- Captura de erros (como divisão por zero)
- Demonstração de métodos de comparação

## Saída Esperada

```
=== Exemplo Básico - Provider Cockroach (Padrão) ===
a = 10.50
b = 3.25
a + b = 13.75
a * b = 34.125
10.50 é maior que 3.25

=== Exemplo com Configuração Customizada ===
Provider: shopspring
Valor: 123.4567
Divisor: 7
Resultado (alta precisão): 17.6366714285714286

Mudando para provider cockroach:
Provider atual: cockroach
Resultado com cockroach: 17.6366714285714286

=== Exemplo de Operações Batch ===
Valores originais:
  10.50
  25.75
  8.90
  12.30
  45.60

Soma total: 103.05
Média: 20.61
Máximo: 45.60
Mínimo: 8.90

=== Exemplo com Hooks ===
Executando operações com hooks habilitados:
Resultado final: 125.50

Testando divisão por zero:
Erro esperado capturado: division by zero

Exemplos de comparação:
100.00 > 25.50
100.00 == 100.00
25.50 != 0
```

## Conceitos Importantes

### Provider Switching
O exemplo mostra como alternar entre providers (cockroach e shopspring) em tempo de execução, permitindo flexibilidade baseada em necessidades específicas.

### Tratamento de Erros
Todas as operações que podem falhar retornam erros que devem ser tratados adequadamente.

### Operações Batch
Para melhor performance ao trabalhar com múltiplos decimais, use as operações batch do manager.

### Hooks e Observabilidade
O sistema de hooks permite adicionar logging, validação e métricas de forma transparente.

## Próximos Passos

- Veja o exemplo avançado para operações mais complexas
- Consulte o exemplo de providers para comparações detalhadas
- Explore o exemplo de hooks para customizações avançadas
