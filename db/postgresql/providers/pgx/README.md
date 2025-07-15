# PGX Provider - Error Handling

Este provider oferece um sistema robusto de **wrapper de erros** que classifica e fornece contexto adicional para todos os tipos de erros retornados pelo driver PGX.

## 🚨 Tipos de Erro Suportados

### Erros de Conexão
- `ErrorTypeConnectionFailed` - Falha geral de conexão
- `ErrorTypeConnectionLost` - Conexão perdida durante operação
- `ErrorTypeConnectionTimeout` - Timeout de conexão
- `ErrorTypeConnectionRefused` - Conexão recusada pelo servidor
- `ErrorTypePoolExhausted` - Pool de conexões esgotado
- `ErrorTypeAuthenticationFail` - Falha de autenticação

### Erros de Query
- `ErrorTypeSyntaxError` - Erro de sintaxe SQL
- `ErrorTypeUndefinedTable` - Tabela não existe
- `ErrorTypeUndefinedColumn` - Coluna não existe
- `ErrorTypeUndefinedFunction` - Função não existe
- `ErrorTypeDataTypeMismatch` - Incompatibilidade de tipos
- `ErrorTypeDivisionByZero` - Divisão por zero

### Violações de Constraint
- `ErrorTypeUniqueViolation` - Violação de constraint única
- `ErrorTypeForeignKeyViolation` - Violação de chave estrangeira
- `ErrorTypeNotNullViolation` - Violação de NOT NULL
- `ErrorTypeCheckViolation` - Violação de CHECK constraint

### Erros de Transação
- `ErrorTypeTransactionRollback` - Rollback de transação
- `ErrorTypeSerializationFailure` - Falha de serialização
- `ErrorTypeDeadlockDetected` - Deadlock detectado
- `ErrorTypeTransactionAborted` - Transação abortada
- `ErrorTypeInvalidTransactionState` - Estado inválido de transação

### Erros de Sistema
- `ErrorTypeDiskFull` - Disco cheio
- `ErrorTypeInsufficientMemory` - Memória insuficiente
- `ErrorTypeSystemError` - Erro de sistema

## 🛠️ Uso Básico

### Wrapping de Erros

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"

// Qualquer erro retornado pelo driver é automaticamente envolvido
err := someOperation()
if err != nil {
    wrappedErr := pgx.WrapError(err)
    
    // Acesso às informações detalhadas
    if dbErr, ok := wrappedErr.(*pgx.DatabaseError); ok {
        fmt.Printf("Tipo: %s\n", dbErr.Type)
        fmt.Printf("Mensagem: %s\n", dbErr.Message)
        fmt.Printf("Detalhe: %s\n", dbErr.Detail)
        fmt.Printf("SQL State: %s\n", dbErr.SQLState)
    }
}
```

### Verificação de Tipo de Erro

```go
// Verificar se é erro de conexão
if pgx.IsConnectionError(err) {
    // Implementar reconexão
    fmt.Println("Erro de conexão detectado, tentando reconectar...")
}

// Verificar se é violação de constraint
if pgx.IsConstraintViolation(err) {
    // Tratar violação específica
    fmt.Println("Violação de constraint detectada")
}

// Verificar se é erro de transação
if pgx.IsTransactionError(err) {
    // Tratar erro de transação
    fmt.Println("Erro de transação detectado")
}

// Verificar se o erro é retry-able
if pgx.IsRetryable(err) {
    // Implementar retry logic
    fmt.Println("Erro pode ser retentado")
}
```

### Acesso a Informações Detalhadas

```go
if dbErr, ok := wrappedErr.(*pgx.DatabaseError); ok {
    // Informações específicas do PostgreSQL
    fmt.Printf("Schema: %s\n", dbErr.SchemaName)
    fmt.Printf("Tabela: %s\n", dbErr.TableName)
    fmt.Printf("Coluna: %s\n", dbErr.ColumnName)
    fmt.Printf("Constraint: %s\n", dbErr.ConstraintName)
    fmt.Printf("Posição do erro: %d\n", dbErr.Position)
    fmt.Printf("Arquivo fonte: %s:%d\n", dbErr.FileName, dbErr.LineNumber)
    fmt.Printf("Rotina: %s\n", dbErr.RoutineName)
}
```

## 🎯 Mapeamento de Códigos SQL State

O sistema mapeia automaticamente códigos de erro PostgreSQL para tipos de erro:

| SQL State | Tipo de Erro | Descrição |
|-----------|--------------|-----------|
| 23505 | `ErrorTypeUniqueViolation` | Violação de constraint única |
| 23503 | `ErrorTypeForeignKeyViolation` | Violação de chave estrangeira |
| 23502 | `ErrorTypeNotNullViolation` | Violação de NOT NULL |
| 23514 | `ErrorTypeCheckViolation` | Violação de CHECK constraint |
| 42601 | `ErrorTypeSyntaxError` | Erro de sintaxe SQL |
| 42P01 | `ErrorTypeUndefinedTable` | Tabela não existe |
| 42703 | `ErrorTypeUndefinedColumn` | Coluna não existe |
| 40001 | `ErrorTypeSerializationFailure` | Falha de serialização |
| 40P01 | `ErrorTypeDeadlockDetected` | Deadlock detectado |

## 🔄 Estratégias de Retry

```go
func executeWithRetry(operation func() error, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        wrappedErr := pgx.WrapError(err)
        if !pgx.IsRetryable(wrappedErr) {
            return wrappedErr
        }
        
        lastErr = wrappedErr
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return lastErr
}
```

## 📊 Cobertura de Testes

- **Cobertura:** 86.4%
- **Testes unitários:** 100% dos tipos de erro
- **Testes de integração:** Todos os casos de sucesso e falha
- **Benchmarks:** Performance otimizada

## 🔗 Compatibilidade

- **Driver:** jackc/pgx/v5
- **PostgreSQL:** 12+
- **Go:** 1.19+

## 📝 Observações

- Todos os erros originais são preservados via `Unwrap()`
- Implementa `errors.Is()` para comparação de tipos
- Thread-safe para uso concorrente
- Performance otimizada com minimal overhead
