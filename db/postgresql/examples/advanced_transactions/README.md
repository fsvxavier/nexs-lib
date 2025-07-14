# Advanced Transactions Example

Este exemplo demonstra o uso avançado de transações com PostgreSQL, incluindo savepoints, rollbacks parciais e cenários complexos.

## Funcionalidades Demonstradas

- Transações com diferentes níveis de isolamento
- Savepoints e rollbacks parciais
- Transações aninhadas
- Tratamento de deadlocks
- Retry automático em falhas
- Transações longas com timeout

## Estrutura

- `main.go` - Exemplo principal com transações avançadas
- `scenarios.go` - Diferentes cenários de transação
- `banking.go` - Exemplo de sistema bancário
- `go.mod` - Dependências do exemplo

## Como Executar

```bash
go run .
```

## Requisitos

- PostgreSQL 12+
- Go 1.21+
