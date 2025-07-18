# Exemplos - DomainErrors

Esta pasta contém exemplos práticos de uso do módulo domainerrors em diferentes cenários e níveis de complexidade.

## Estrutura dos Exemplos

### 📁 [basic/](basic/)
**Exemplo Básico** - Introdução ao módulo
- ✅ Criação de erros básicos
- ✅ Tipos de erro específicos
- ✅ Metadados e contexto
- ✅ Empilhamento de erros
- ✅ Grupo de erros
- ✅ Utilitários e mapeamento HTTP

**Ideal para**: Primeiro contato com o módulo

### 📁 [advanced/](advanced/)
**Exemplo Avançado** - Padrões de produção
- ✅ Cenários complexos de erro
- ✅ Middleware de error handling
- ✅ Padrões de recovery (Circuit Breaker, Retry)
- ✅ Observabilidade e monitoramento
- ✅ Validação contextual
- ✅ Integração com serviços

**Ideal para**: Aplicações em produção

### 📁 [global/](global/)
**Configuração Global** - Setup de aplicação
- ✅ Configuração global do módulo
- ✅ Handler centralizado
- ✅ Integração com contexto
- ✅ Panic recovery global
- ✅ Logging e métricas
- ✅ Customização de tipos

**Ideal para**: Configuração de aplicações

## Guia de Uso

### 1. Iniciantes
Comece com o exemplo básico:
```bash
cd basic/
go run main.go
```

### 2. Desenvolvedores Experientes
Explore padrões avançados:
```bash
cd advanced/
go run main.go
```

### 3. Configuração de Aplicações
Configure seu ambiente:
```bash
cd global/
go run main.go
```

## Dependências

Todos os exemplos usam:
- Go 1.19+
- Módulo domainerrors
- Biblioteca padrão do Go

## Execução

### Executar Todos os Exemplos
```bash
# Executar script de exemplo
./run_all_examples.sh
```

### Executar Individualmente
```bash
# Exemplo básico
cd basic && go run main.go

# Exemplo avançado
cd advanced && go run main.go

# Configuração global
cd global && go run main.go
```

## Conceitos Principais

### 1. Tipos de Erro
- **Validation**: Erros de validação de dados
- **NotFound**: Recursos não encontrados
- **Business**: Violações de regras de negócio
- **Database**: Falhas de banco de dados
- **ExternalService**: Falhas em APIs externas
- **Infrastructure**: Problemas de infraestrutura

### 2. Funcionalidades
- **Stack Trace**: Captura automática de stack traces
- **Contexto**: Enriquecimento com informações contextuais
- **Metadados**: Dados adicionais para debugging
- **Serialização**: Conversão para JSON
- **HTTP Mapping**: Mapeamento para códigos HTTP

### 3. Padrões
- **Error Wrapping**: Encadeamento de erros
- **Recovery**: Captura de panics
- **Retry**: Tentativas com backoff
- **Circuit Breaker**: Proteção contra falhas
- **Middleware**: Interceptação centralizada

## Casos de Uso

### APIs REST
```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.userService.Create(r.Context(), request)
    if err != nil {
        response := h.errorHandler.HandleError(err)
        writeJSONResponse(w, response)
        return
    }
    writeJSONResponse(w, user)
}
```

### gRPC Services
```go
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    user, err := s.userRepo.Create(ctx, req)
    if err != nil {
        return nil, domainerrors.Wrap("Failed to create user", err)
    }
    return user, nil
}
```

### Background Jobs
```go
func (w *Worker) ProcessMessage(ctx context.Context, msg *Message) error {
    defer func() {
        if r := recover(); r != nil {
            if err := domainerrors.RecoverWithStackTrace(); err != nil {
                w.errorHandler.HandleCriticalError(err)
            }
        }
    }()
    
    return w.processor.Process(ctx, msg)
}
```

## Integração com Frameworks

### Gin
```go
func ErrorMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                response := errorHandler.HandleError(err.(error))
                c.JSON(response["http_status"].(int), response)
            }
        }()
        c.Next()
    }
}
```

### Echo
```go
func ErrorHandler(err error, c echo.Context) {
    response := errorHandler.HandleError(err)
    c.JSON(response["http_status"].(int), response)
}
```

## Observabilidade

### Métricas
```go
// Prometheus
errorCounter := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "errors_total",
        Help: "Total number of errors",
    },
    []string{"type", "service"},
)

// Incrementar métrica
errorCounter.WithLabelValues(errorType, serviceName).Inc()
```

### Tracing
```go
// OpenTelemetry
span := trace.SpanFromContext(ctx)
span.SetStatus(codes.Error, err.Error())
span.RecordError(err)
```

### Logging
```go
// Structured logging
logger.Error("Operation failed",
    zap.String("error_code", domainErr.Code),
    zap.String("error_type", domainErr.ErrorType),
    zap.Any("metadata", domainErr.Metadata()),
)
```

## Contribuindo

Para adicionar novos exemplos:
1. Crie pasta com nome descritivo
2. Adicione `main.go` com exemplo
3. Adicione `README.md` com documentação
4. Atualize este README
5. Adicione ao script `run_all_examples.sh`

## Recursos Adicionais

- [Documentação do Módulo](../README.md)
- [API Reference](../docs/api.md)
- [Guia de Migração](../docs/migration.md)
- [Melhores Práticas](../docs/best-practices.md)
