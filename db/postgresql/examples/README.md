# Exemplos do Módulo PostgreSQL

Este diretório contém exemplos práticos de uso do módulo PostgreSQL da nexs-lib.

## Estrutura dos Exemplos

- `basic/` - Exemplos básicos de conexão e operações simples
- `advanced/` - Exemplos avançados com transações, batch operations e pools
- `patterns/` - Exemplos de padrões de design aplicados
- `testing/` - Exemplos de como testar código que usa este módulo

## Como Executar

1. Configure um banco PostgreSQL local:
```bash
docker run --name postgres-test \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15
```

2. Execute os exemplos:
```bash
cd examples/basic
go run main.go
```

## Configuração

Todos os exemplos usam as seguintes variáveis de ambiente opcionais:
- `DB_HOST` (default: localhost)
- `DB_PORT` (default: 5432)  
- `DB_NAME` (default: testdb)
- `DB_USER` (default: postgres)
- `DB_PASSWORD` (default: postgres)
- `DB_SSLMODE` (default: disable)
