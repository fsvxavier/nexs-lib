# Cache Valkey Module

Um módulo Go genérico e desacoplado para **Valkey** com suporte a múltiplos drivers, implementando padrão Factory para intercambiabilidade completa entre providers.

## 🚀 Características Principais

- ✅ **Provider Genérico**: Totalmente desacoplado dos drivers específicos
- ✅ **Multiple Drivers**: Suporte a `valkey-go` e `valkey-glide` (planejado)
- ✅ **Thread-Safe**: 100% seguro para uso concorrente
- ✅ **Arquitetura Orientada a Interfaces**: Máxima flexibilidade e testabilidade
- ✅ **Factory Pattern**: Troca de drivers sem alteração no código cliente
- ✅ **Retry & Circuit Breaker**: Resilência automática com backoff exponencial
- ✅ **Hooks Extensíveis**: Sistema de interceptação pré/pós execução
- ✅ **Context Support**: Suporte completo a `context.Context`
- ✅ **Configuração Flexível**: Via struct, environment variables ou builder pattern
- ✅ **Multi-Mode**: Standalone, Cluster e Sentinel
- ✅ **Performance**: Pool de conexões otimizado e reutilização de buffers

## 📦 Drivers Suportados

| Driver | Status | Versão |
|--------|--------|--------|
| [valkey-go](https://github.com/valkey-io/valkey-go) | ✅ Implementado | v1.0.63 |
| [valkey-glide](https://github.com/valkey-io/valkey-glide/tree/main/go) | 🚧 Planejado | - |

## 🏗️ Arquitetura

```
cache/valkey/
├── interfaces/          # Interfaces principais (IClient, IPipeline, etc.)
├── config/             # Configuração com suporte a env vars
├── hooks/              # Sistema de hooks extensível
├── providers/          # Implementações específicas por driver
│   └── valkey-go/     # Provider para valkey-go
├── valkey.go          # Client principal e Manager
└── retry_circuit_breaker.go  # Políticas de resilência
```

## 🚀 Uso Rápido

### Uso Genérico (Recomendado)

```go
package main

import (
    "context"
    "log"

    "github.com/fsvxavier/nexs-lib/cache/valkey"
    "github.com/fsvxavier/nexs-lib/cache/valkey/config"
    valkeygo "github.com/fsvxavier/nexs-lib/cache/valkey/providers/valkey-go"
)

func main() {
    // Registrar provider
    manager := valkey.NewManager()
    manager.RegisterProvider("valkey-go", valkeygo.NewProvider())
    
    // Configuração
    cfg := &config.Config{
        Provider: "valkey-go",
        Host:     "localhost",
        Port:     6379,
    }
    
    // Criar cliente
    client, err := manager.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    ctx := context.Background()
    
    // Operações básicas
    err = client.Set(ctx, "key", "value", 0)
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := client.Get(ctx, "key")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Value: %s", value)
}
```

### Configuração via Environment Variables

```bash
# Configuração básica
export VALKEY_HOST=localhost
export VALKEY_PORT=6379
export VALKEY_PASSWORD=mypassword
export VALKEY_PROVIDER=valkey-go

# Pool de conexões
export VALKEY_POOL_SIZE=10
export VALKEY_MIN_IDLE_CONNS=2
export VALKEY_MAX_IDLE_CONNS=5

# Timeouts
export VALKEY_DIAL_TIMEOUT=5s
export VALKEY_READ_TIMEOUT=3s
export VALKEY_WRITE_TIMEOUT=3s

# Retry e Circuit Breaker
export VALKEY_MAX_RETRIES=3
export VALKEY_MIN_RETRY_BACKOFF=8ms
export VALKEY_MAX_RETRY_BACKOFF=512ms
export VALKEY_CIRCUIT_BREAKER_THRESHOLD=5

# Cluster mode
export VALKEY_CLUSTER_MODE=true
export VALKEY_ADDRS=node1:6379,node2:6379,node3:6379

# TLS
export VALKEY_TLS_ENABLED=true
export VALKEY_TLS_CERT_FILE=/path/to/cert.pem
export VALKEY_TLS_KEY_FILE=/path/to/key.pem
```

```go
// Carregar configuração do ambiente
cfg := config.LoadFromEnv()
client, err := manager.NewClient(cfg)
```

### Builder Pattern

```go
cfg := config.DefaultConfig().
    WithProvider("valkey-go").
    WithHost("localhost").
    WithPort(6379).
    WithPoolSize(20).
    WithMaxRetries(5)

client, err := manager.NewClient(cfg)
```

## 🔧 Operações Suportadas

### Comandos String
```go
// SET/GET
client.Set(ctx, "key", "value", time.Minute)
value, err := client.Get(ctx, "key")

// DEL/EXISTS/TTL/EXPIRE
count, err := client.Del(ctx, "key1", "key2")
exists, err := client.Exists(ctx, "key1", "key2")
ttl, err := client.TTL(ctx, "key")
err = client.Expire(ctx, "key", time.Hour)
```

### Comandos Hash
```go
// HSET/HGET
err = client.HSet(ctx, "hash", "field1", "value1", "field2", "value2")
value, err := client.HGet(ctx, "hash", "field1")

// HDEL/HEXISTS/HGETALL
count, err := client.HDel(ctx, "hash", "field1", "field2")
exists, err := client.HExists(ctx, "hash", "field1")
values, err := client.HGetAll(ctx, "hash")
```

### Pipeline
```go
pipe := client.Pipeline()
pipe.Set("key1", "value1", 0)
pipe.Set("key2", "value2", 0)
cmd1 := pipe.Get("key1")
cmd2 := pipe.Get("key2")

results, err := pipe.Exec(ctx)
value1, err := cmd1.String()
value2, err := cmd2.String()
```

### Transações
```go
tx := client.TxPipeline()
tx.Set("key1", "value1", 0)
tx.Set("key2", "value2", 0)

results, err := tx.Exec(ctx)
```

### Pub/Sub
```go
pubsub, err := client.Subscribe(ctx, "channel1", "channel2")
defer pubsub.Close()

// Publicar
count, err := client.Publish(ctx, "channel1", "message")

// Receber mensagens
ch := pubsub.Channel()
for msg := range ch {
    fmt.Printf("Received: %s on %s\n", msg.Payload, msg.Channel)
}
```

### Streams
```go
// XADD
streamID, err := client.XAdd(ctx, "stream", map[string]interface{}{
    "field1": "value1",
    "field2": "value2",
})

// XREAD
messages, err := client.XRead(ctx, map[string]string{
    "stream": "0",
})
```

## 🎯 Configuração Avançada

### Cluster Mode
```go
cfg := &config.Config{
    Provider:    "valkey-go",
    ClusterMode: true,
    Addrs:       []string{"node1:6379", "node2:6379", "node3:6379"},
    PoolSize:    20,
}
```

### Sentinel Mode
```go
cfg := &config.Config{
    Provider:           "valkey-go",
    SentinelMode:       true,
    SentinelAddrs:      []string{"sentinel1:26379", "sentinel2:26379"},
    SentinelMasterName: "mymaster",
    SentinelPassword:   "sentinelpass",
}
```

### TLS Configuration
```go
cfg := &config.Config{
    Provider:               "valkey-go",
    Host:                   "secure.valkey.com",
    Port:                   6380,
    TLSEnabled:             true,
    TLSCertFile:           "/path/to/cert.pem",
    TLSKeyFile:            "/path/to/key.pem",
    TLSCACertFile:         "/path/to/ca.pem",
    TLSInsecureSkipVerify: false,
}
```

## 🔌 Hooks e Extensibilidade

### Logging Hook
```go
import "github.com/fsvxavier/nexs-lib/cache/valkey/hooks"

loggingHook := hooks.NewLoggingHook()
cfg.Hooks = []interfaces.IHook{loggingHook}
```

### Metrics Hook
```go
metricsHook := hooks.NewMetricsHook()
cfg.Hooks = []interfaces.IHook{metricsHook}
```

### Custom Hook
```go
type CustomHook struct{}

func (h *CustomHook) PreExecution(ctx context.Context, command string, args []interface{}) context.Context {
    // Lógica antes da execução
    return ctx
}

func (h *CustomHook) PostExecution(ctx context.Context, command string, args []interface{}, result interface{}, err error) {
    // Lógica após a execução
}

func (h *CustomHook) PreConnection(ctx context.Context, config interface{}) context.Context {
    return ctx
}

func (h *CustomHook) PostConnection(ctx context.Context, config interface{}, err error) {
    // Lógica após conexão
}

cfg.Hooks = []interfaces.IHook{&CustomHook{}}
```

## 🏃‍♂️ Performance e Otimizações

- **Pool de Conexões**: Configurável e otimizado para alta concorrência
- **Pipeline Automático**: Agrupa comandos automaticamente para melhor throughput
- **Buffer Reuse**: Reutilização de buffers para reduzir alocações
- **Connection Pooling**: Gestão inteligente de conexões idle
- **Circuit Breaker**: Evita cascata de falhas em cenários de alta carga

## 🧪 Testes

```bash
# Testes unitários
go test -tags=unit -timeout 30s -race ./...

# Testes com cobertura
go test -tags=unit -timeout 30s -race -cover ./...

# Benchmarks
go test -bench=. -benchmem ./...
```

## 📊 Monitoramento e Observabilidade

### Health Check
```go
healthy := client.IsHealthy(ctx)
if !healthy {
    log.Println("Client is not healthy")
}

// Health check com timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := client.Ping(ctx)
```

### Metrics (via Hook)
O módulo suporta coleta automática de métricas via hooks:

- Latência de comandos
- Taxa de sucesso/erro
- Pool de conexões (ativas/idle)
- Circuit breaker status
- Retry attempts

## 🔐 Segurança

- **TLS Support**: Suporte completo a TLS/SSL
- **Authentication**: Suporte a username/password
- **Connection Security**: Validação de certificados
- **Timeout Protection**: Prevenção de operações hanging

## 🐛 Troubleshooting

### Logs de Debugging
```go
// Habilitar logs detalhados
cfg.LogLevel = "debug"
```

### Common Issues

1. **Connection Timeout**: Ajustar `DialTimeout`
2. **Pool Exhaustion**: Aumentar `PoolSize` ou reduzir `ConnMaxAge`
3. **Circuit Breaker Open**: Verificar conectividade e ajustar threshold
4. **Memory Usage**: Ajustar `MaxIdleConns` e `IdleTimeout`

## 📚 Exemplos Completos

Confira o diretório `examples/` para exemplos detalhados:

- `global/`: Uso genérico básico
- `advanced/`: Configurações avançadas
- `valkey-go/`: Específico para valkey-go driver

## 🤝 Contribuição

1. Todas as mudanças devem manter compatibilidade com interfaces
2. Cobertura de testes mínima: 98%
3. Benchmark tests para mudanças de performance
4. Documentação atualizada

## 📄 Licença

Este módulo faz parte da nexs-lib e segue a mesma licença do projeto principal.

---

**Desenvolvido com foco em produção, performance e confiabilidade** 🚀
