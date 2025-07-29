# Exemplo Avançado PostgreSQL

Este exemplo demonstra funcionalidades avançadas do PostgreSQL usando a biblioteca nexs-lib.

## Funcionalidades Demonstradas

### 1. Pool Management
- Criação e gerenciamento de pool de conexões
- Aquisição e liberação de conexões
- Configuração avançada de pool

### 2. Transactions
- Transações com commit
- Transações com rollback
- Tratamento de erros em transações

### 3. Batch Operations
- Inserção em lote
- Operações múltiplas em transação
- Otimização de performance

### 4. Concurrent Operations
- Operações concorrentes com múltiplos workers
- Uso seguro de pool de conexões
- Sincronização com sync.WaitGroup

### 5. Connection Pooling
- Simulação de carga de trabalho
- Reutilização eficiente de conexões
- Teste de performance com múltiplos workers

### 6. Error Handling
- Tratamento de erros de sintaxe SQL
- Erros de tabela inexistente
- Erros de constraint
- Estratégias de recuperação

### 7. Multi-tenancy
- Uso de schemas para separação de dados
- Configuração de search_path
- Isolamento de dados por tenant

### 8. LISTEN/NOTIFY
- Configuração de listener
- Envio de notificações
- Comunicação assíncrona

### 9. Performance Testing
- Testes de performance de INSERT
- Testes de performance de SELECT
- Medição de throughput

## Como Executar

### Usando Docker (Recomendado)
```bash
# Iniciar infraestrutura
./infraestructure/manage.sh start

# Executar exemplo avançado
./infraestructure/manage.sh example advanced

# Parar infraestrutura
./infraestructure/manage.sh stop
```

### Execução Direta
```bash
# Configurar variáveis de ambiente
export NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"

# Executar
go run main.go
```

## Configuração

O exemplo usa as seguintes configurações avançadas:

```go
cfg := postgres.NewConfigWithOptions(
    dsn,
    postgres.WithMaxConns(50),
    postgres.WithMinConns(10),
    postgres.WithMaxConnLifetime(time.Hour),
    postgres.WithMaxConnIdleTime(30*time.Minute),
)
```

## Pré-requisitos

- PostgreSQL em execução
- Schema `nexs_testdb` criado
- Tabelas necessárias criadas (ver infraestructure/database/init/)

## Saída Esperada

```
=== Exemplo Avançado PostgreSQL ===

Pool Management
===============
  Criando pool de conexões...
  Adquirindo múltiplas conexões...
  Conexão 1 adquirida
  Conexão 2 adquirida
  ...
  Liberando conexões...
✓ Pool Management concluído com sucesso

Transactions
============
  Demonstrando transações...
  Executando transação com commit...
  Transação confirmada com sucesso
  Executando transação com rollback...
  Simulando erro e fazendo rollback...
  Rollback executado com sucesso
  Total de registros após transações: 2
✓ Transactions concluído com sucesso

[... outros exemplos ...]

=== Exemplos avançados concluídos ===
```

## Características Avançadas

### Pool de Conexões
- Configuração dinâmica de min/max conexões
- Controle de lifetime das conexões
- Monitoramento de idle time

### Transações
- Suporte a nested transactions
- Rollback automático em caso de erro
- Isolation levels configuráveis

### Concorrência
- Thread-safe operations
- Pool sharing entre goroutines
- Sincronização adequada

### Performance
- Batch operations otimizadas
- Prepared statements
- Connection reuse

### Observabilidade
- Logging detalhado
- Métricas de performance
- Monitoramento de erros

## Próximos Passos

1. Adicionar exemplos de prepared statements
2. Implementar connection health checks
3. Adicionar métricas de monitoramento
4. Implementar retry policies
5. Adicionar exemplos de streaming
