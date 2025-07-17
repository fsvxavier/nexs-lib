# PostgreSQL Provider Mocks

Esta pasta contém mocks organizados para testes do provedor PostgreSQL. Cada tipo de mock está em seu próprio arquivo para facilitar manutenção e uso.

## Estrutura dos Arquivos

### `connection.go`
- **MockConnection**: Mock principal para interfaces.IConn
- Implementa todos os métodos de conexão de banco de dados
- Inclui contadores de chamadas para verificação em testes
- Permite customização através de funções de callback

### `transaction.go`
- **MockTransaction**: Mock para interfaces.ITransaction
- Herda de MockConnection para operações dentro de transação
- Rastreia estado de commit/rollback
- Métodos auxiliares para verificação de estado

### `row.go`
- **MockRow**: Mock simples para interfaces.IRow
- Implementa método Scan com callback customizável

### `rows.go`
- **MockRows**: Mock para interfaces.IRows
- Implementa iteração sobre múltiplas linhas
- Métodos para simular resultados de consultas

### `command_tag.go`
- **MockCommandTag**: Mock para interfaces.CommandTag
- Simula metadados de execução de comandos
- Permite customização de resultados

### `batch_results.go`
- **MockBatchResults**: Mock para interfaces.IBatchResults
- Simula resultados de execução em lote
- Combina funcionalidades de Row, Rows e CommandTag

### `hook_manager.go`
- **MockHookManager**: Mock para interfaces.HookManager
- Simula sistema de hooks do PostgreSQL provider
- Permite teste de funcionalidades de interceptação

### `provider.go`
- **MockProvider**: Mock de alto nível para o provider
- **MockPool**: Mock para pool de conexões
- Utilitários e construtores convenientes

## Uso Básico

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx/mocks"

// Criar conexão mock
conn := mocks.NewMockConnection()

// Customizar comportamento
conn.QueryFunc = func(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
    // Lógica personalizada para testes
    return mocks.NewMockRows(), nil
}

// Verificar chamadas
assert.Equal(t, 1, conn.GetCallCount("Query"))
```

## Recursos dos Mocks

### Contadores de Chamadas
Todos os mocks principais mantêm contadores de quantas vezes cada método foi chamado:

```go
conn := mocks.NewMockConnection()
conn.Query(ctx, "SELECT 1")
assert.Equal(t, 1, conn.GetCallCount("Query"))
```

### Callbacks Customizáveis
Cada método pode ter seu comportamento customizado:

```go
conn := mocks.NewMockConnection()
conn.QueryFunc = func(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
    return nil, errors.New("erro simulado")
}
```

### Estado de Transação
MockTransaction rastreia seu estado:

```go
tx := mocks.NewMockTransaction()
tx.Commit(ctx)
assert.True(t, tx.IsCommitted())
assert.False(t, tx.IsRolledBack())
```

### Reset para Reutilização
```go
conn.ResetCallCounts()
tx.Reset()
```

## Compatibilidade

Todos os mocks implementam as interfaces definidas em `db/postgresql/interface/interfaces.go` e são compatíveis com os testes existentes.
