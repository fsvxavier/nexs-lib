# Graceful Operations Examples

Este diretório contém exemplos abrangentes de como usar as **operações graceful completas** da biblioteca nexs-lib httpserver, incluindo graceful shutdown, restart, health monitoring e multi-server management.

## ✨ Funcionalidades Demonstradas

### 🔄 Operações Graceful Completas
- **Graceful Shutdown**: Shutdown controlado com connection draining
- **Zero-Downtime Restart**: Restart sem perder conexões ativas  
- **Health Monitoring**: Monitoramento de saúde em tempo real
- **Connection Tracking**: Rastreamento de conexões ativas
- **Signal Handling**: Captura de sinais do sistema operacional

### 🏗️ Arquitetura Multi-Server
- **Manager Pattern**: Gerenciamento centralizado de múltiplos servidores
- **Hook System**: Hooks de pré e pós shutdown para cleanup
- **Health Checks**: Sistema de health checks customizáveis
- **Production Ready**: Configurações prontas para produção

## 🚀 Como Executar

```bash
# Execute o exemplo
go run main.go
```

O exemplo demonstra um **cenário real de produção** com:

### 🌐 Múltiplos Servidores
- **API Server** (Gin) na porta 8080 - API principal da aplicação
- **Admin Server** (Echo) na porta 8081 - Interface administrativa  
- **Health Server** (NetHTTP) na porta 9090 - Status geral da aplicação

### 📊 Endpoints Disponíveis
- `http://localhost:8080/api/health` - Health check da API
- `http://localhost:8080/api/data` - Dados da API com simulação de carga
- `http://localhost:8081/admin/health` - Health check administrativo
- `http://localhost:8081/admin/status` - Status detalhado do sistema
- `http://localhost:9090/health` - Status geral e agregado da aplicação

### 🔄 Teste de Graceful Operations

1. **Inicie o exemplo**: `go run main.go`
2. **Faça algumas requisições** para simular carga ativa
3. **Teste graceful shutdown**: Pressione `Ctrl+C` 
4. **Observe o comportamento**:
   - Hooks de pré-shutdown executados
   - Conexões drenadas graciosamente  
   - Hooks de pós-shutdown executados
   - Shutdown sem interrupção de requisições ativas

## 🏗️ Estrutura do Graceful Manager

### Configuração Básica
```go
// Criar manager centralizado
manager := graceful.NewManager()

// Registrar múltiplos servidores
manager.RegisterServer("api", apiServer)
manager.RegisterServer("admin", adminServer)  
manager.RegisterServer("health", healthServer)

// Configurar timeouts
manager.SetDrainTimeout(10 * time.Second)
manager.SetShutdownTimeout(30 * time.Second)
```

### Health Checks Customizados
```go
// Database health check
manager.AddHealthCheck("database", func() interfaces.HealthCheck {
    return interfaces.HealthCheck{
        Status:    "healthy",
        Message:   "Database connection OK",
        Duration:  time.Millisecond * 50,
        Timestamp: time.Now(),
    }
})

// Cache health check  
manager.AddHealthCheck("cache", func() interfaces.HealthCheck {
    return interfaces.HealthCheck{
        Status:    "warning", 
        Message:   "Cache response time high",
        Duration:  time.Millisecond * 150,
        Timestamp: time.Now(),
    }
})
```

### Hook System Avançado
```go
// Pre-shutdown hooks
manager.AddPreShutdownHook(func() error {
    log.Println("🔄 Preparing for shutdown...")
    log.Println("📊 Saving metrics to disk...")
    log.Println("🔌 Closing external connections...")
    return nil
})

// Post-shutdown hooks
manager.AddPostShutdownHook(func() error {
    log.Println("🧹 Cleanup completed")
    log.Println("💾 Final data saved")
    log.Println("✅ Graceful shutdown successful")
    return nil
})
```

### Signal Handling Production-Ready
```go
// Setup graceful shutdown
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

go func() {
    <-stop
    log.Println("🛑 Shutdown signal received...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    if err := manager.GracefulShutdown(ctx); err != nil {
        log.Printf("❌ Forced shutdown: %v", err)
        os.Exit(1)
    }
    
    log.Println("✅ Server stopped gracefully")
    os.Exit(0)
}()
```

// Configurar timeouts
manager.SetDrainTimeout(30 * time.Second)
manager.SetShutdownTimeout(60 * time.Second)

// Adicionar health checks
manager.AddHealthCheck("database", func() interfaces.HealthCheck {
    return interfaces.HealthCheck{
        Status: "healthy",
        Message: "Database OK",
        Duration: 5 * time.Millisecond,
        Timestamp: time.Now(),
    }
})

// Adicionar hooks
manager.AddPreShutdownHook(func() error {
    log.Println("Salvando estado...")
    return nil
})

manager.AddPostShutdownHook(func() error {
    log.Println("Limpando recursos...")
    return nil
})

// Registrar servidores
manager.RegisterServer("api", ginServer)
manager.RegisterServer("admin", echoServer)

// Executar graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

if err := manager.GracefulShutdown(ctx); err != nil {
    log.Printf("Erro no shutdown: %v", err)
}
```

## 🔧 Funcionalidades Avançadas

### Connection Monitoring
```go
// Monitorar conexões ativas
activeConns := server.GetConnectionsCount()
log.Printf("Conexões ativas: %d", activeConns)

// Aguardar fim das conexões
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := server.WaitForConnections(ctx)
if err != nil {
    log.Printf("Timeout aguardando conexões: %v", err)
}
```

### Health Status Agregado
```go
// Status geral do sistema
status := manager.GetHealthStatus()
fmt.Printf("Status: %s\n", status.Status)           // healthy/warning/unhealthy
fmt.Printf("Uptime: %s\n", status.Uptime)          // tempo desde o início
fmt.Printf("Conexões: %d\n", status.Connections)   // total de conexões ativas
fmt.Printf("Checks: %d\n", len(status.Checks))     // número de health checks

// Status individual por check
for name, check := range status.Checks {
    fmt.Printf("Check %s: %s - %s\n", name, check.Status, check.Message)
}
```

### Restart Zero-Downtime
```go
// Restart sem perder conexões
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

err := server.Restart(ctx)
if err != nil {
    log.Printf("Falha no restart: %v", err)
}
```

## 🚀 Production Benefits

- ✅ **Zero-Downtime Operations**: Restart e shutdown sem perder requisições
- ✅ **Connection Draining**: Espera inteligente pelo fim das conexões ativas  
- ✅ **Health Monitoring**: Monitoramento em tempo real da saúde da aplicação
- ✅ **Signal Handling**: Resposta adequada a sinais do sistema operacional
- ✅ **Hook System**: Cleanup e preparação customizáveis
- ✅ **Multi-Server**: Gerenciamento coordenado de múltiplos servidores
- ✅ **Timeout Configuration**: Timeouts configuráveis para diferentes cenários
- ✅ **Status Reporting**: Relatórios detalhados de status e saúde

Este exemplo demonstra um **padrão de produção real** para aplicações Go que precisam de alta disponibilidade e operações graceful robustas.
