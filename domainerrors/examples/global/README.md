# Exemplo Global - DomainErrors

Este exemplo demonstra como configurar e usar o módulo domainerrors de forma global em uma aplicação.

## Funcionalidades Demonstradas

### 1. Configuração Global
- **Stack Trace**: Configuração global para captura de stack traces
- **Logging**: Setup centralizado de logging
- **Panic Recovery**: Captura global de panics
- **Variáveis Globais**: Configuração em nível de aplicação

### 2. Handler Centralizado
- **ErrorHandler**: Classe centralizada para tratamento de erros
- **Logging Estruturado**: Logs organizados por severidade
- **Métricas**: Coleta automática de métricas de erro
- **Tracing**: Integração com sistemas de tracing

### 3. Integração com Contexto
- **Context Enrichment**: Enriquecimento de erros com contexto
- **Request ID**: Rastreamento de requisições
- **User ID**: Identificação de usuários
- **Trace ID**: Correlação de traces

### 4. Customização de Tipos
- **Mapeamento HTTP**: Customização de códigos HTTP
- **Severidade**: Configuração de níveis de severidade
- **Categorização**: Organização de tipos de erro

### 5. Configuração de Stack Trace
- **Profundidade**: Controle de profundidade do stack trace
- **Frames**: Configuração de frames ignorados
- **Habilitação**: Controle global de captura

## Configuração Global

### Variáveis Globais
```go
domainerrors.GlobalStackTraceEnabled = true
domainerrors.GlobalMaxStackDepth = 15
domainerrors.GlobalSkipFrames = 3
```

### Handler Centralizado
```go
type ErrorHandler struct {
    config *AppConfig
    logger *log.Logger
}
```

### Panic Recovery
```go
func setupGlobalPanicRecovery() {
    defer func() {
        if r := recover(); r != nil {
            if err := domainerrors.RecoverWithStackTrace(); err != nil {
                handler.HandleCriticalError(err)
            }
        }
    }()
}
```

## Como Executar

```bash
cd domainerrors/examples/global
go run main.go
```

## Saída Esperada

O exemplo produzirá saída demonstrando:
- Configuração global aplicada
- Handler centralizado processando erros
- Enriquecimento com contexto
- Diferentes configurações de stack trace
- Logging estruturado com níveis

## Integração em Aplicações

### 1. Inicialização
```go
func main() {
    // Configurar domainerrors globalmente
    domainerrors.GlobalStackTraceEnabled = true
    domainerrors.GlobalMaxStackDepth = 10
    
    // Criar handler centralizado
    errorHandler := NewErrorHandler(config)
    
    // Configurar recovery global
    setupGlobalPanicRecovery(errorHandler)
}
```

### 2. Middleware HTTP
```go
func ErrorMiddleware(handler ErrorHandler) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    response := handler.HandleError(err.(error))
                    writeJSONResponse(w, response)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### 3. gRPC Interceptor
```go
func ErrorInterceptor(handler ErrorHandler) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        if err != nil {
            enrichedErr := handler.EnrichWithContext(ctx, err)
            return resp, enrichedErr
        }
        return resp, nil
    }
}
```

## Benefícios

1. **Consistência**: Tratamento uniforme em toda aplicação
2. **Centralização**: Configuração e lógica em um local
3. **Observabilidade**: Logging, métricas e tracing automáticos
4. **Manutenibilidade**: Fácil modificação de comportamento
5. **Debugging**: Informações ricas para diagnóstico

## Configuração por Ambiente

### Desenvolvimento
```go
config := &AppConfig{
    Environment:     "development",
    LogLevel:        "debug",
    EnableMetrics:   true,
    EnableTracing:   true,
}
```

### Produção
```go
config := &AppConfig{
    Environment:     "production",
    LogLevel:        "info",
    EnableMetrics:   true,
    EnableTracing:   true,
}
```

## Melhores Práticas

1. **Configuração Centralizada**: Use struct de configuração
2. **Logging Estruturado**: Inclua contexto relevante
3. **Métricas**: Colete dados para monitoramento
4. **Alertas**: Configure alertas para erros críticos
5. **Contexto**: Enriqueça erros com informações da requisição

## Próximos Passos

Após implementar configuração global:
- Integre com framework web (Gin, Echo)
- Configure sistema de monitoramento
- Implemente alertas automáticos
- Adicione dashboards de métricas
