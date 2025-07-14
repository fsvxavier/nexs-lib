# Basic Operations Example

Este exemplo demonstra operações básicas de CRUD (Create, Read, Update, Delete) usando o módulo PostgreSQL.

## Funcionalidades Demonstradas

- Conexão com banco de dados
- Operações de inserção (INSERT)
- Consultas simples (SELECT)
- Atualizações (UPDATE)
- Remoções (DELETE)
- Tratamento de erros

## Estrutura

- `main.go` - Exemplo principal com operações CRUD
- `models.go` - Estruturas de dados
- `go.mod` - Dependências do exemplo

## Como Executar

1. Configure um banco PostgreSQL local ou use Docker:
```bash
docker run --name postgres-example -e POSTGRES_PASSWORD=password -e POSTGRES_DB=example -p 5432:5432 -d postgres:15
```

2. Execute o exemplo:
```bash
go run .
```

## Requisitos

- PostgreSQL 12+
- Go 1.21+
