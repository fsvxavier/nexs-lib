# PostgreSQL Database Provider

Um módulo Go genérico e extensível para conexões PostgreSQL que suporta múltiplos drivers e funcionalidades avançadas.

## Características Principais

### 🔧 Provider
- **PGX**: Driver nativo PostgreSQL com alta performance

### 🚀 Funcionalidades Avançadas
- **Connection Pooling**: Gerenciamento inteligente de pool de conexões
- **Transações**: Suporte completo a transações com isolamento
- **Operações em Batch**: Execução eficiente de múltiplas queries
- **Multi-tenancy**: Suporte a múltiplos inquilinos por schema/database
- **Read Replicas**: Load balancing automático para réplicas de leitura
- **Failover**: Recuperação automática de falhas de conexão
- **LISTEN/NOTIFY**: Suporte a notificações PostgreSQL
- **Copy Operations**: Operações de COPY FROM/TO para alta performance

### 🎯 Sistema de Hooks
Sistema completo de hooks para interceptação de operações:
- **Connection Hooks**: Before/After connection, acquire, release
- **Operation Hooks**: Before/After query, exec, transaction, batch
- **Error Hooks**: Tratamento personalizado de erros
- **Custom Hooks**: Hooks customizados para necessidades específicas

### 🔗 Sistema de Middlewares
Chain de middlewares com execução ordenada:
- **Logging**: Log detalhado de operações
- **Timing**: Medição de performance
- **Validation**: Validação de queries e parâmetros
- **Metrics**: Coleta de métricas operacionais
- **Cache**: Cache inteligente de resultados
- **Custom Middlewares**: Middlewares personalizados

### 🛡️ Segurança e Confiabilidade
- **Thread-safe**: Seguro para uso concorrente
- **Retry Logic**: Retry automático com backoff exponencial
- **Health Checks**: Verificação de saúde de conexões
- **SSL/TLS**: Suporte completo a criptografia
- **Context Support**: Cancelamento e timeout via context

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/db/postgresql
```

## Uso Básico

### Configuração Simples

```go
package main

import (
    "context"
    "log"
    
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
    // Criar provider PGX
    provider, err := postgresql.NewPGXProvider()
    if err != nil {
        log.Fatal(err)
    }

    // Configuração básica
    config := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/dbname")

    // Criar pool de conexões
    ctx := context.Background()
    pool, err := provider.NewPool(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Usar conexão
    err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
        var result string
        return conn.QueryOne(ctx, &result, "SELECT 'Hello, World!'")
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### Configuração Avançada

```go
// Configuração com opções avançadas
config := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/dbname")

err := config.(*config.DefaultConfig).ApplyOptions(
    postgresql.WithMaxConns(50),
    postgresql.WithMinConns(5),
    postgresql.WithMaxConnLifetime(time.Hour),
    postgresql.WithLogging(true),
    postgresql.WithTiming(true),
    postgresql.WithMetrics(true),
    postgresql.WithMultiTenant(true),
    postgresql.WithReadReplicas([]string{
        "postgres://user:password@replica1:5432/dbname",
        "postgres://user:password@replica2:5432/dbname",
    }, interfaces.LoadBalanceModeRoundRobin),
    postgresql.WithFailover([]string{
        "postgres://user:password@backup:5432/dbname",
    }, 3),
)
```

### Hooks Personalizados

```go
// Registrar hook personalizado
hookManager := pool.GetHookManager()

// Hook de log personalizado
logHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
    log.Printf("Executing: %s", ctx.Operation)
    return &interfaces.HookResult{Continue: true}
}

err = hookManager.RegisterHook(interfaces.BeforeQueryHook, logHook)

// Hook customizado
customHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
    // Lógica personalizada
    return &interfaces.HookResult{Continue: true}
}

err = hookManager.RegisterCustomHook(
    interfaces.CustomHookBase+1, 
    "my_custom_hook", 
    customHook,
)
```

### Middlewares

```go
// Adicionar middleware personalizado
middlewareManager := pool.GetMiddlewareManager()

// Middleware de cache personalizado
cacheMiddleware := middlewares.NewCacheMiddleware(time.Minute * 5)
err = middlewareManager.AddMiddleware(cacheMiddleware)

// Middleware de log personalizado
logMiddleware := middlewares.NewLoggingMiddleware("APP")
err = middlewareManager.AddMiddleware(logMiddleware)
```

### Multi-tenancy

```go
// Configurar contexto com tenant
tenantCtx := context.WithValue(ctx, "tenant_id", "tenant_123")

// Usar conexão com tenant
err = pool.AcquireFunc(tenantCtx, func(conn interfaces.IConn) error {
    // Queries executadas automaticamente no contexto do tenant
    return conn.QueryAll(tenantCtx, &results, "SELECT * FROM data")
})
```

### Transações

```go
err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
    tx, err := conn.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Operações na transação
    _, err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - 100")
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
})
```

### Operações em Batch

```go
err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
    batch := &pgx.PGXBatch{}
    
    // Adicionar queries ao batch
    batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 1")
    batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 2")
    batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 3")

    // Executar batch
    batchResults := conn.SendBatch(ctx, batch)
    defer batchResults.Close()

    // Processar resultados
    for i := 0; i < 3; i++ {
        _, err := batchResults.Exec()
        if err != nil {
            return err
        }
    }

    return nil
})
```

### LISTEN/NOTIFY

```go
// Criar conexão específica para LISTEN
listenConn, err := provider.NewListenConn(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer listenConn.Close(ctx)

// Configurar listener
err = listenConn.Listen(ctx, "notifications")
if err != nil {
    log.Fatal(err)
}

// Aguardar notificações
for {
    notification, err := listenConn.WaitForNotification(ctx, time.Second*30)
    if err != nil {
        log.Printf("Error waiting for notification: %v", err)
        continue
    }
    
    log.Printf("Received notification: %s = %s", 
        notification.Channel, notification.Payload)
}
```

## Estrutura do Projeto

```
db/postgresql/
├── interface/           # Interfaces principais
├── config/             # Configuração
├── hooks/              # Sistema de hooks
├── middlewares/        # Sistema de middlewares
├── providers/          # Implementações específicas
│   └── pgx/           # Provider PGX
├── examples/          # Exemplos de uso
├── postgresql.go      # API principal
├── postgresql_test.go # Testes principais
├── README.md          # Este arquivo
└── NEXT_STEPS.md      # Próximos passos
```

## Providers Suportados

### PGX (Recomendado)
- Driver nativo PostgreSQL
- Alta performance
- Suporte completo a funcionalidades PostgreSQL
- Connection pooling nativo
- LISTEN/NOTIFY
- Copy operations
- SSL/TLS
## Configuração

### Opções de Pool
```go
WithMaxConns(50)                    // Máximo de conexões
WithMinConns(5)                     // Mínimo de conexões
WithMaxConnLifetime(time.Hour)      // Tempo de vida da conexão
WithMaxConnIdleTime(time.Minute*5)  // Tempo idle máximo
WithHealthCheckPeriod(time.Minute)  // Período de health check
WithConnectTimeout(time.Second*30)  // Timeout de conexão
```

### Opções de Middleware
```go
WithLogging(true)       // Habilitar logging
WithTiming(true)        // Habilitar timing
WithValidation(true)    // Habilitar validação
WithMetrics(true)       // Habilitar métricas
WithCache(true)         // Habilitar cache
```

### Opções de Segurança
```go
WithTLS(true, &tls.Config{...})     // Configurar TLS
WithTLSFiles("cert.pem", "key.pem", "ca.pem")  // TLS com arquivos
```

### Opções de Alta Disponibilidade
```go
WithReadReplicas([]string{...}, LoadBalanceModeRoundRobin)  // Read replicas
WithFailover([]string{...}, 3)                             // Failover nodes
WithMaxRetries(5)                                           // Máximo de retries
```

## Testes

O módulo possui cobertura de testes superior a 98%:

```bash
# Executar todos os testes
go test -tags=unit -timeout 30s -race ./...

# Executar com cobertura
go test -tags=unit -timeout 30s -race -coverprofile=coverage.out ./...

# Ver cobertura
go tool cover -html=coverage.out
```

## Performance

### Benchmarks
```bash
# Executar benchmarks
go test -bench=. -benchmem ./...
```

### Métricas
O sistema coleta automaticamente:
- Tempo de execução de queries
- Número de operações por tipo
- Taxa de erro
- Estatísticas de pool
- Métricas de cache hit/miss

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Próximos Passos

Consulte [NEXT_STEPS.md](NEXT_STEPS.md) para roadmap e melhorias planejadas.
