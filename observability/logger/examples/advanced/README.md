# Exemplo Avançado de Logging

Este exemplo demonstra funcionalidades avançadas do sistema de logging multi-provider.

## Executando o Exemplo

```bash
cd examples/advanced
go run main.go
```

## Funcionalidades Demonstradas

### 1. Serviços com Logging Multi-Provider
- **UserService**: Serviço com logger configurável
- **Provider switching**: Troca dinâmica entre providers
- **Logging por operação**: Cada método tem logs específicos
- **Integração com domain errors**: Tratamento de erros estruturados

### 2. Context-Aware Service Layer
- **Trace ID**: Rastreamento distribuído
- **User ID**: Identificação do usuário
- **Request ID**: Identificação da requisição
- **Span ID**: Rastreamento de spans

### 3. Multi-Provider Demonstration
- **Zap**: Provider padrão de alta performance
- **Slog**: Provider da biblioteca padrão Go
- **Zerolog**: Provider otimizado para baixo consumo
- **Comparação**: Demonstração de diferenças entre providers

### 4. Structured Error Handling
- **Domain Errors**: Integração com sistema de erros
- **Error codes**: Códigos padronizados
- **Stack traces**: Rastreamento de origem
- **Contexto de erro**: Informações detalhadas

### 5. Performance Benchmarking
- **Provider benchmarks**: Teste de performance individual
- **Comparative analysis**: Análise comparativa
- **Memory usage**: Consumo de memória
- **Throughput**: Taxa de logs processados

### 6. Service Integration Patterns
- **Constructor injection**: Injeção de dependência
- **Context propagation**: Propagação de contexto
- **Logging middleware**: Padrão de middleware
- **Error boundaries**: Tratamento de erros

## Estrutura dos Logs

### Log de Criação de Usuário (Zap)
```json
{
  "level": "info",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "user_id": "user456",
  "msg": "Criando usuário",
  "nome": "João Silva",
  "email": "joao@email.com"
}
```

### Log de Erro (Zerolog)
```json
{
  "level": "error",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "message": "Erro ao criar usuário",
  "error_type": "ValidationError",
  "error_code": "USER_INVALID_EMAIL",
  "details": "Email já existe"
}
```

### Performance Benchmark Results
```
Provider: zap
Logs/second: ~240,000
Memory usage: Low
CPU usage: Low
Recommendation: High-performance applications

Provider: zerolog  
Logs/second: ~174,000
Memory usage: Very Low
CPU usage: Very Low
Recommendation: Memory-constrained applications

Provider: slog
Logs/second: ~132,000
Memory usage: Medium
CPU usage: Medium
Recommendation: Standard library compatibility
```

## Cenários de Uso

### 1. Aplicação Web com Trace ID
```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx = context.WithValue(ctx, "trace_id", generateTraceID())
    ctx = context.WithValue(ctx, "request_id", generateRequestID())
    
    service := &UserService{logger: logger.NewLogger()}
    user, err := service.GetUser(ctx, "user123")
    // Logs incluirão automaticamente trace_id e request_id
}
```

### 2. Microserviço com Provider Switching
```go
func NewUserService(providerName string) *UserService {
    logger := logger.NewLogger()
    if providerName != "" {
        logger.ConfigureProvider(providerName, nil)
    }
    return &UserService{logger: logger}
}
```

### 3. Benchmark de Performance
```go
func BenchmarkProviders() {
    providers := []string{"zap", "slog", "zerolog"}
    for _, provider := range providers {
        // Benchmark cada provider
        benchmarkProvider(provider)
    }
}
```

## Vantagens de Cada Provider

### Zap (Padrão)
- ✅ Mais rápido (~240k logs/sec)
- ✅ Baixo consumo de memória
- ✅ Amplamente usado em produção
- ✅ Configuração flexível

### Zerolog
- ✅ Menor consumo de memória
- ✅ Boa performance (~174k logs/sec)
- ✅ API simples e intuitiva
- ✅ Zero allocation em muitos cenários

### Slog
- ✅ Biblioteca padrão Go
- ✅ Compatibilidade garantida
- ✅ API estável e simples
- ✅ Boa integração com tooling Go
  "msg": "Usuário criado com sucesso",
  "service": "user-api",
  "version": "2.1.0",
  "environment": "development",
  "datacenter": "us-east-1",
  "instance": "web-01",
  "trace_id": "trace-1642248645123",
  "request_id": "req-1642248645123",
  "user_id": "authenticated-user-123",
  "code": "USER_CREATED",
  "user_id": "user-123",
  "email": "success@example.com",
  "name": "João Silva"
}
```

### Log de Middleware HTTP
```json
{
  "time": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Requisição processada",
  "service": "user-api",
  "version": "2.1.0",
  "trace_id": "trace-1642248645123",
  "request_id": "req-1642248645123",
  "user_id": "authenticated-user-123",
  "request_duration": "125ms",
  "status": "200"
}
```

## Configurações por Ambiente

### Development
- **Nível**: Debug
- **Formato**: Console (legível)
- **Source**: Habilitado
- **Stacktrace**: Desabilitado

### Staging
- **Nível**: Info
- **Formato**: JSON
- **Source**: Desabilitado
- **Stacktrace**: Habilitado

### Production
- **Nível**: Warn
- **Formato**: JSON
- **Source**: Desabilitado
- **Stacktrace**: Habilitado
- **Sampling**: 1000 inicial, depois 1 a cada 100

## Casos de Uso Demonstrados

1. **Aplicação Web**: Middleware HTTP com tracing
2. **Microserviços**: Serviços com logging estruturado
3. **Error Handling**: Tratamento de erros com contexto
4. **Performance**: Monitoramento de operações
5. **Debugging**: Logs detalhados para desenvolvimento
6. **Production**: Configuração otimizada para produção

## Padrões de Logging

### Início de Operação
```go
logger.Info(ctx, "Iniciando operação",
    logger.String("operation", "create_user"),
    logger.String("user_id", userID),
)
```

### Sucesso
```go
logger.InfoWithCode(ctx, "USER_CREATED", "Usuário criado com sucesso",
    logger.String("user_id", userID),
    logger.Duration("duration", elapsed),
)
```

### Erro
```go
logger.ErrorWithCode(ctx, "USER_INVALID_EMAIL", "Email é obrigatório",
    logger.String("user_id", userID),
    logger.ErrorField(err),
)
```

### Performance
```go
start := time.Now()
// ... operação ...
duration := time.Since(start)

logger.Info(ctx, "Operação completada",
    logger.Duration("duration", duration),
    logger.String("status", "success"),
)
```
