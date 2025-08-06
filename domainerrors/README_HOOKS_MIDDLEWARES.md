# Domain Errors - Sistema de Hooks e Middlewares

Uma biblioteca Go avançada para gerenciamento robusto de erros de domínio com suporte completo a hooks e middlewares.

## ✨ Características Principais

- ✅ **29 tipos específicos de erro** (validação, autenticação, timeout, circuit breaker, etc.)
- ✅ **Stack trace detalhado** com captura automática
- ✅ **Mapeamento automático para códigos HTTP**
- ✅ **Encadeamento de erros** com suporte nativo ao `errors.Unwrap()`
- ✅ **Metadados flexíveis** para contexto adicional
- ✅ **Sistema de hooks** para interceptação de eventos
- ✅ **Cadeia de middlewares** para processamento customizado
- ✅ **Thread-safe** com suporte a contexto
- ✅ **Cobertura de testes 98%+**

## 🚀 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/domainerrors
```

## 📖 Uso Básico

### Criação de Erros

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
    // Erro básico
    err := domainerrors.New("USER_001", "Usuário não encontrado")
    
    // Erro com tipo específico
    validationErr := domainerrors.NewWithType(
        "VALIDATION_001", 
        "Campo obrigatório", 
        domainerrors.ErrorTypeValidation,
    )
    
    // Erro com causa
    dbErr := domainerrors.NewWithCause(
        "DB_001", 
        "Falha na consulta", 
        originalErr,
    )
}
```

### Metadados e Contexto

```go
err := domainerrors.New("BUSINESS_001", "Saldo insuficiente")
err.WithMetadata("account_id", "12345")
err.WithMetadata("requested_amount", 1500.00)
err.WithContext(ctx, "Tentativa de transferência")
```

## 🎣 Sistema de Hooks

Os hooks permitem interceptar eventos específicos no ciclo de vida dos erros.

### Registrando Hooks

```go
// Hook para logging
domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
    log.Printf("Erro criado: %s [%s]", err.Code, err.Type)
    return nil
})

// Hook para auditoria
domainerrors.RegisterHook("before_metadata", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
    // Lógica de auditoria antes de adicionar metadados
    return nil
})
```

### Tipos de Hooks Disponíveis

- `before_error` - Antes da criação do erro
- `after_error` - Após criação do erro  
- `before_metadata` - Antes de adicionar metadados
- `after_metadata` - Após adicionar metadados
- `before_stack_trace` - Antes de capturar stack trace
- `after_stack_trace` - Após capturar stack trace

## 🔧 Sistema de Middlewares

Middlewares permitem processar erros em uma cadeia de transformações.

### Registrando Middlewares

```go
// Middleware de enriquecimento
domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
    if err.Metadata == nil {
        err.Metadata = make(map[string]interface{})
    }
    
    err.Metadata["service"] = "user-service"
    err.Metadata["version"] = "1.2.0"
    err.Metadata["processed_at"] = time.Now()
    
    return next(err)
})

// Middleware de validação
domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
    if err.Code == "" {
        log.Warn("Erro sem código detectado")
    }
    
    return next(err)
})
```

## 📊 Tipos de Erro Suportados

| Tipo | Código HTTP | Descrição |
|------|-------------|-----------|
| `ErrorTypeValidation` | 400 | Erros de validação |
| `ErrorTypeAuthentication` | 401 | Falhas de autenticação |
| `ErrorTypeAuthorization` | 403 | Falhas de autorização |
| `ErrorTypeNotFound` | 404 | Recurso não encontrado |
| `ErrorTypeTimeout` | 408 | Timeout de operação |
| `ErrorTypeConflict` | 409 | Conflito de recursos |
| `ErrorTypeBusinessRule` | 422 | Violação de regra de negócio |
| `ErrorTypeRateLimit` | 429 | Limite de taxa excedido |
| `ErrorTypeServer` | 500 | Erro interno do servidor |
| `ErrorTypeServiceUnavailable` | 503 | Serviço indisponível |
| E mais 19 tipos específicos... | | |

## 🔄 Exemplo Completo com Hooks e Middlewares

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
    // Registra hooks
    domainerrors.RegisterHook("after_error", func(ctx context.Context, err *domainerrors.DomainError, operation string) error {
        log.Printf("[AUDIT] Erro: %s - Tipo: %s", err.Code, err.Type)
        return nil
    })

    // Registra middlewares
    domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
        // Enriquece com dados do serviço
        if err.Metadata == nil {
            err.Metadata = make(map[string]interface{})
        }
        err.Metadata["service"] = "payment-service"
        err.Metadata["environment"] = "production"
        
        return next(err)
    })

    // Cria erro que será processado por hooks e middlewares
    paymentErr := domainerrors.NewWithType(
        "PAYMENT_001",
        "Pagamento rejeitado",
        domainerrors.ErrorTypeBusinessRule,
    )
    
    paymentErr.WithMetadata("payment_id", "pay_123456")
    paymentErr.WithMetadata("amount", 250.50)
    
    log.Printf("Erro final: %s", paymentErr.Error())
    log.Printf("Status HTTP: %d", paymentErr.HTTPStatus())
    log.Printf("Metadados: %+v", paymentErr.Metadata)
}
```

## 🎯 Arquitetura de Interfaces

O módulo possui uma arquitetura bem estruturada com interfaces centralizadas:

- **`/interfaces`** - Todas as interfaces de hooks e middlewares
- **`/hooks`** - Implementação do sistema de hooks e registry
- **`/middleware`** - Implementação do sistema de middlewares e cadeia
- **`/examples`** - Exemplos práticos de uso

## 🧪 Testabilidade

Cobertura de testes superior a 98% com testes para:

- ✅ Criação e manipulação de erros
- ✅ Sistema completo de hooks
- ✅ Cadeia de middlewares  
- ✅ Encadeamento de erros
- ✅ Mapeamento HTTP
- ✅ Thread safety
- ✅ Performance benchmarks

```bash
go test -v -timeout 30s ./...
```

## 📈 Performance

A biblioteca é otimizada para alta performance com:

- Registries thread-safe com sync.RWMutex
- Execução eficiente de cadeias de middlewares
- Captura otimizada de stack traces
- Baixo overhead na criação de erros

## 🔍 Debugging e Observabilidade

Recursos avançados para debugging:

```go
// Stack trace detalhado
fmt.Println(err.StackTrace())

// Informações do erro
fmt.Printf("Código: %s\n", err.Code)
fmt.Printf("Tipo: %s\n", err.Type)
fmt.Printf("HTTP Status: %d\n", err.HTTPStatus())
fmt.Printf("Timestamp: %s\n", err.Timestamp)
```

## ⚡ Hooks vs Middlewares

### Quando usar Hooks:
- **Event-driven**: Para reagir a eventos específicos
- **Side effects**: Logging, auditoria, notificações
- **Não modificam o erro**: Apenas observam/reagem
- **Múltiplos hooks** podem ser registrados para o mesmo evento

### Quando usar Middlewares:
- **Processing pipeline**: Para transformar/enriquecer erros
- **Chain of responsibility**: Processamento em sequência
- **Modificam o erro**: Adicionam metadados, contexto
- **Ordem importa**: Executados em ordem de registro

## 🔄 Pipeline de Execução

```
Erro Criado
    ↓
Hooks "before_*"
    ↓
Middleware Chain
    ↓
Processamento Interno
    ↓
Hooks "after_*"
    ↓
Erro Final
```

## 🎯 Próximos Passos

Consulte `NEXT_STEPS.md` para:

- Integrações planejadas
- Melhorias de performance
- Novos tipos de middleware
- Recursos de observabilidade

## 📄 Licença

MIT License - consulte LICENSE para detalhes.

---

**Desenvolvido como parte do nexs-lib - Uma biblioteca Go de componentes empresariais**
