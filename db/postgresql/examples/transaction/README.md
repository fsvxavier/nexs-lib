# Transaction PostgreSQL Provider Example

Este exemplo demonstra o gerenciamento avançado de transações PostgreSQL, incluindo diferentes níveis de isolamento, savepoints, rollbacks e operações em lote.

## 📋 Funcionalidades Demonstradas

- ✅ **Transações Básicas**: Begin, Commit, Rollback
- ✅ **Níveis de Isolamento**: READ UNCOMMITTED, READ COMMITTED, REPEATABLE READ, SERIALIZABLE
- ✅ **Transações Aninhadas**: Savepoints e rollbacks parciais
- ✅ **Cenários de Rollback**: Automático e manual baseado em lógica de negócio
- ✅ **Timeouts**: Transações com controle de tempo
- ✅ **Operações em Lote**: Bulk operations dentro de transações
- ✅ **Prepared Statements**: Otimização para operações repetitivas

## 🚀 Pré-requisitos

1. **PostgreSQL Database**:
   ```bash
   # Usando Docker
   docker run --name postgres-tx \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=testdb \
     -p 5432:5432 -d postgres:15
   ```

2. **Dependências Go**:
   ```bash
   go mod tidy
   ```

## ⚙️ Configuração

1. **Atualize a string de conexão** no arquivo `main.go`:
   ```go
   cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
   ```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/transaction

# Executar o exemplo
go run main.go
```

## 📊 Saída Esperada

```
=== Basic Transaction Example ===
✅ Committing transaction
📊 Account balances:
  Alice: $1000.00
  Bob: $500.00

=== Isolation Levels Example ===
🔒 Testing isolation level: READ UNCOMMITTED
  ✅ Successfully queried with READ UNCOMMITTED: 2 records
🔒 Testing isolation level: READ COMMITTED
  ✅ Successfully queried with READ COMMITTED: 2 records
🔒 Testing isolation level: REPEATABLE READ
  ✅ Successfully queried with REPEATABLE READ: 2 records
🔒 Testing isolation level: SERIALIZABLE
  ✅ Successfully queried with SERIALIZABLE: 2 records

=== Nested Transactions (Savepoints) Example ===
💾 Creating savepoint 'sp1'
✅ Updated Alice's balance (+$100)
💾 Creating savepoint 'sp2'
⚠️  Updated Bob's balance (-$999999) - this will be rolled back
🔄 Rolling back to savepoint 'sp2'
🗑️  Releasing savepoint 'sp2'
✅ Updated Bob's balance (+$50)
📊 Final balances after savepoint operations:
  Alice: $1100.00
  Bob: $550.00
✅ Committing main transaction

=== Rollback Scenarios Example ===
🔄 Scenario 1: Automatic rollback on constraint violation
  ✅ Expected error occurred: constraint violation

🔄 Scenario 2: Manual rollback based on business logic
  ⚠️  Insufficient funds: trying to transfer $1200.00, but only $1100.00 available
  🔄 Rolling back due to business rule violation

=== Transaction Timeout Example ===
⏰ Starting transaction with 2-second timeout...
💤 Simulating long operation (3 seconds)...
  ✅ Transaction timed out as expected: context deadline exceeded
✅ Transaction timeout handled correctly: context deadline exceeded

=== Bulk Operations in Transaction Example ===
📝 Using prepared statements for bulk inserts...
📝 Performing bulk insert operations...
  ✅ Inserted transaction 1: Alice $100.00 (deposit)
  ✅ Inserted transaction 2: Bob $50.00 (deposit)
  ✅ Inserted transaction 3: Alice $-25.00 (withdrawal)
  ✅ Inserted transaction 4: Bob $-10.00 (withdrawal)
  ✅ Inserted transaction 5: Alice $75.00 (deposit)
📊 Transaction summary:
  Alice: $150.00
  Bob: $40.00
✅ Committing bulk transaction
Transaction examples completed!
```

## 📝 Conceitos Demonstrados

### 1. Transação Básica com Defer Pattern
```go
tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
    IsoLevel:   interfaces.TxIsoLevelReadCommitted,
    AccessMode: interfaces.TxAccessModeReadWrite,
})

var success bool
defer func() {
    if success {
        tx.Commit(ctx)
    } else {
        tx.Rollback(ctx)
    }
}()
```

### 2. Níveis de Isolamento
```go
// READ UNCOMMITTED - Permite dirty reads
interfaces.TxIsoLevelReadUncommitted

// READ COMMITTED - Previne dirty reads (padrão)
interfaces.TxIsoLevelReadCommitted

// REPEATABLE READ - Previne dirty e non-repeatable reads
interfaces.TxIsoLevelRepeatableRead

// SERIALIZABLE - Mais alto nível, previne phantom reads
interfaces.TxIsoLevelSerializable
```

### 3. Savepoints (Transações Aninhadas)
```go
// Criar savepoint
_, err = tx.Exec(ctx, "SAVEPOINT sp1")

// Rollback para savepoint específico
_, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp1")

// Liberar savepoint
_, err = tx.Exec(ctx, "RELEASE SAVEPOINT sp1")
```

### 4. Timeout em Transações
```go
timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()

tx, err := conn.BeginTx(timeoutCtx, interfaces.TxOptions{...})
```

### 5. Operações em Lote
```go
preparedSQL := "INSERT INTO table (col1, col2) VALUES ($1, $2)"
for _, data := range bulkData {
    _, err = tx.Exec(ctx, preparedSQL, data.col1, data.col2)
}
```

## 🔒 Níveis de Isolamento Explicados

| Nível | Dirty Read | Non-Repeatable Read | Phantom Read | Performance |
|-------|------------|-------------------|--------------|-------------|
| READ UNCOMMITTED | ✅ Permite | ✅ Permite | ✅ Permite | 🚀 Máxima |
| READ COMMITTED | ❌ Previne | ✅ Permite | ✅ Permite | 🏃 Alta |
| REPEATABLE READ | ❌ Previne | ❌ Previne | ✅ Permite | 🚶 Média |
| SERIALIZABLE | ❌ Previne | ❌ Previne | ❌ Previne | 🐌 Mínima |

## 🔧 Padrões de Uso

### 1. Transação Simples (Recomendado)
```go
err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
    tx, err := conn.BeginTx(ctx, interfaces.TxOptions{})
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // Sempre rollback no defer
    
    // Suas operações aqui
    
    return tx.Commit(ctx) // Commit só se chegou até aqui
})
```

### 2. Transação com Lógica de Negócio
```go
var success bool
defer func() {
    if success {
        tx.Commit(ctx)
    } else {
        tx.Rollback(ctx)
    }
}()

// Lógica de negócio que define success = true/false
```

### 3. Transação com Savepoints
```go
tx.Exec(ctx, "SAVEPOINT before_risky_operation")
// Operação arriscada
if riskyOperationFailed {
    tx.Exec(ctx, "ROLLBACK TO SAVEPOINT before_risky_operation")
} else {
    tx.Exec(ctx, "RELEASE SAVEPOINT before_risky_operation")
}
```

## 🐛 Troubleshooting

### Deadlock Detectado
```
deadlock detected
```
**Solução**: Use timeouts e retry logic, ordene aquisição de locks.

### Serialization Failure
```
could not serialize access
```
**Solução**: Implemente retry para transações SERIALIZABLE.

### Transaction Timeout
```
context deadline exceeded
```
**Solução**: Aumente timeout ou otimize operações.

### Lock Wait Timeout
```
lock wait timeout exceeded
```
**Solução**: Reduza duração das transações, use índices apropriados.

## 📚 Próximos Passos

Após dominar transações, explore:

1. **[Batch Example](../batch/)**: Operações em lote otimizadas
2. **[Advanced Example](../advanced/)**: Hooks e middleware avançados
3. **[Performance Example](../performance/)**: Otimização de performance

## 🔍 Debugging de Transações

Para debug detalhado:
```bash
export LOG_LEVEL=debug
export TX_DEBUG=true
export SHOW_SQL=true
```

Logs incluirão:
- Início/fim de transações
- Comandos SQL executados
- Savepoints criados/liberados
- Tempos de execução
- Locks adquiridos
