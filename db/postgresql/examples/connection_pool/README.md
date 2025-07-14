# Connection Pool Management Example

Este exemplo demonstra o gerenciamento avançado de pools de conexão com monitoramento e configurações otimizadas.

## Funcionalidades Demonstradas

- Configuração avançada de pool de conexões
- Monitoramento de estatísticas do pool
- Health checks automáticos
- Gerenciamento de conexões sob alta carga
- Timeouts e recuperação de falhas

## Estrutura

- `main.go` - Exemplo principal com gerenciamento de pool
- `monitor.go` - Monitoramento e métricas
- `worker.go` - Simulação de carga de trabalho
- `go.mod` - Dependências do exemplo

## Como Executar

1. Configure PostgreSQL e execute o exemplo:
```bash
go run .
```

## Requisitos

- PostgreSQL 12+
- Go 1.21+
