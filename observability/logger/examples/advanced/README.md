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

## Código de Exemplo

```go
package main

import (
    "context"
    "time"
    
    "github.com/fsvxavier/nexs-lib/observability/logger"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

// UserService demonstra integração com serviços
type UserService struct {
    logger logger.Logger
}

// NewUserService cria um novo serviço com logger
func NewUserService(logger logger.Logger) *UserService {
    return &UserService{logger: logger}
}

// GetUser busca um usuário com logging estruturado
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    s.logger.Info(ctx, "Buscando usuário",
        logger.String("user_id", userID),
        logger.String("operation", "get_user"),
    )
    
    // Simulação de busca
    time.Sleep(time.Millisecond * 10)
    
    if userID == "invalid" {
        err := domainerrors.NewValidationError("USER_INVALID", "Usuário inválido")
        s.logger.Error(ctx, "Erro ao buscar usuário",
            logger.String("user_id", userID),
            logger.String("error", err.Error()),
        )
        return nil, err
    }
    
    user := &User{ID: userID, Name: "João Silva"}
    s.logger.Info(ctx, "Usuário encontrado",
        logger.String("user_id", userID),
        logger.String("user_name", user.Name),
    )
    
    return user, nil
}

func main() {
    // Testa todos os providers
    providers := []string{"zap", "slog", "zerolog"}
    
    for _, provider := range providers {
        fmt.Printf("\n=== Testando Provider: %s ===\n", provider)
        
        // Configura o provider
        logger.ConfigureProvider(provider, nil)
        
        // Cria contexto com trace info
        ctx := context.WithValue(context.Background(), "trace_id", "abc123")
        ctx = context.WithValue(ctx, "user_id", "user456")
        
        // Testa serviço
        service := NewUserService(logger.NewLogger())
        user, err := service.GetUser(ctx, "user123")
        
        if err != nil {
            logger.Error(ctx, "Erro no serviço", logger.String("error", err.Error()))
        } else {
            logger.Info(ctx, "Serviço executado com sucesso",
                logger.String("user_id", user.ID),
                logger.String("user_name", user.Name),
            )
        }
    }
}
```

## Estrutura dos Logs Detalhada

### Log de Sucesso (Zap)
```json
{
  "level": "info",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "user_id": "user456",
  "msg": "Usuário encontrado",
  "user_id": "user123",
  "user_name": "João Silva"
}
```

### Log de Erro (Zerolog)
```json
{
  "level": "error",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "message": "Erro ao buscar usuário",
  "user_id": "invalid",
  "error": "Usuário inválido"
}
```

### Log de Performance (Slog)
```json
{
  "time": "2025-07-18T10:30:45Z",
  "level": "INFO",
  "trace_id": "abc123",
  "user_id": "user456",
  "msg": "Operação completada",
  "duration": "125ms",
  "status": "success"
}
```

## Configuração por Ambiente

### Development
```go
config := map[string]interface{}{
    "level":      "debug",
    "format":     "console",
    "colorize":   true,
    "addSource":  true,
    "pretty":     true,
}
```

### Staging
```go
config := map[string]interface{}{
    "level":      "info",
    "format":     "json",
    "addSource":  false,
    "sampling":   false,
}
```

### Production
```go
config := map[string]interface{}{
    "level":      "warn",
    "format":     "json",
    "addSource":  false,
    "sampling":   true,
    "outputPath": "/var/log/app.log",
}
```

## Padrões de Uso

### 1. Início de Operação
```go
logger.Info(ctx, "Iniciando operação",
    logger.String("operation", "create_user"),
    logger.String("user_id", userID),
    logger.Time("start_time", time.Now()),
)
```

### 2. Sucesso
```go
logger.Info(ctx, "Operação concluída com sucesso",
    logger.String("operation", "create_user"),
    logger.String("user_id", userID),
    logger.Duration("duration", elapsed),
)
```

### 3. Erro
```go
logger.Error(ctx, "Erro na operação",
    logger.String("operation", "create_user"),
    logger.String("user_id", userID),
    logger.String("error", err.Error()),
    logger.String("error_type", reflect.TypeOf(err).String()),
)
```

### 4. Performance Tracking
```go
start := time.Now()
defer func() {
    duration := time.Since(start)
    logger.Info(ctx, "Performance metrics",
        logger.String("operation", "database_query"),
        logger.Duration("duration", duration),
        logger.Bool("slow_query", duration > time.Second),
    )
}()
```

## Integração com Middleware

### HTTP Middleware
```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Enriquece contexto
        ctx := r.Context()
        ctx = context.WithValue(ctx, "request_id", generateRequestID())
        ctx = context.WithValue(ctx, "trace_id", generateTraceID())
        
        // Log de início
        logger.Info(ctx, "Request iniciada",
            logger.String("method", r.Method),
            logger.String("path", r.URL.Path),
            logger.String("remote_addr", r.RemoteAddr),
        )
        
        // Executa próximo handler
        next.ServeHTTP(w, r.WithContext(ctx))
        
        // Log de fim
        duration := time.Since(start)
        logger.Info(ctx, "Request finalizada",
            logger.Duration("duration", duration),
            logger.Int("status", 200), // Assumindo sucesso
        )
    })
}
```

## Troubleshooting

### Logs não aparecem
```bash
# Verifica nível de log
current := logger.GetCurrentProviderName()
fmt.Printf("Provider atual: %s\n", current)

# Testa com nível debug
logger.ConfigureProvider("zap", map[string]interface{}{
    "level": "debug",
})
```

### Performance degradada
```bash
# Executa benchmark
go run examples/benchmark/main.go
```

### Formato inconsistente
```bash
# Verifica configuração
logger.Info(context.Background(), "Teste de formato",
    logger.String("provider", logger.GetCurrentProviderName()),
)
```

## Próximos Passos

1. **Benchmark**: Execute `examples/benchmark/` para análise de performance
2. **Multi-provider**: Veja `examples/multi-provider/` para comparação
3. **Básico**: Consulte `examples/basic/` para funcionalidades básicas
4. **Default**: Veja `examples/default-provider/` para uso simples
