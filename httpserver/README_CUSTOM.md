# Custom Hooks and Middleware

Este documento explica como usar o sistema de hooks e middleware customizados na biblioteca nexs-lib httpserver.

## Índice

- [Custom Hooks](#custom-hooks)
  - [Criação Básica](#criação-básica-de-hooks)
  - [Hooks Condicionais](#hooks-condicionais)
  - [Hooks Filtrados](#hooks-filtrados)
  - [Hooks Assíncronos](#hooks-assíncronos)
  - [Builder Pattern](#usando-o-builder-pattern-para-hooks)
- [Custom Middleware](#custom-middleware)
  - [Criação Básica](#criação-básica-de-middleware)
  - [Middleware Condicional](#middleware-condicional)
  - [Middleware com Before/After](#middleware-com-beforeafter)
  - [Builder Pattern](#usando-o-builder-pattern-para-middleware)
- [Integração](#integração)
- [Exemplos Completos](#exemplos-completos)

## Custom Hooks

Os custom hooks permitem que você execute código personalizado em resposta a eventos específicos do ciclo de vida das requisições HTTP.

### Criação Básica de Hooks

```go
package main

import (
    "log"
    "github.com/fsvxavier/nexs-lib/httpserver/hooks"
    "github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func main() {
    // Criar factory de hooks
    hookFactory := hooks.NewCustomHookFactory()

    // Hook simples para logging de requisições
    loggingHook := hookFactory.NewSimpleHook(
        "request-logger",
        []interfaces.HookEvent{
            interfaces.HookEventRequestStart,
            interfaces.HookEventRequestEnd,
        },
        100, // prioridade (menor número = maior prioridade)
        func(ctx *interfaces.HookContext) error {
            if ctx.Event == interfaces.HookEventRequestStart {
                log.Printf("Iniciando requisição: %s %s", 
                    ctx.Request.Method, ctx.Request.URL.Path)
            } else {
                log.Printf("Requisição finalizada: %s %s (Status: %d, Duração: %v)",
                    ctx.Request.Method, ctx.Request.URL.Path, 
                    ctx.StatusCode, ctx.Duration)
            }
            return nil
        },
    )

    log.Printf("Hook criado: %s", loggingHook.Name())
}
```

### Hooks Condicionais

Hooks que executam apenas quando certas condições são atendidas:

```go
// Hook que executa apenas para endpoints de API
apiOnlyHook := hookFactory.NewConditionalHook(
    "api-monitor",
    []interfaces.HookEvent{interfaces.HookEventRequestStart},
    200,
    func(ctx *interfaces.HookContext) bool {
        // Condição: apenas paths que começam com /api/
        return strings.HasPrefix(ctx.Request.URL.Path, "/api/")
    },
    func(ctx *interfaces.HookContext) error {
        log.Printf("Monitorando API: %s", ctx.Request.URL.Path)
        return nil
    },
)
```

### Hooks Filtrados

Hooks com filtros específicos por path, método HTTP ou headers:

```go
// Hook para operações POST/PUT em recursos de usuários
userOperationsHook := hookFactory.NewFilteredHook(
    "user-operations",
    []interfaces.HookEvent{interfaces.HookEventRequestStart},
    300,
    func(path string) bool { 
        return strings.Contains(path, "/users/") 
    }, // filtro por path
    func(method string) bool { 
        return method == "POST" || method == "PUT" 
    }, // filtro por método
    func(ctx *interfaces.HookContext) error {
        log.Printf("Operação em usuário: %s %s", 
            ctx.Request.Method, ctx.Request.URL.Path)
        return nil
    },
)
```

### Hooks Assíncronos

Hooks que executam de forma assíncrona sem bloquear a requisição:

```go
// Hook assíncrono para analytics
analyticsHook := hookFactory.NewAsyncHook(
    "analytics",
    []interfaces.HookEvent{interfaces.HookEventRequestEnd},
    400,
    10,               // buffer size
    5*time.Second,    // timeout
    func(ctx *interfaces.HookContext) error {
        // Processamento pesado de analytics
        time.Sleep(100 * time.Millisecond)
        log.Printf("Analytics processado para: %s", ctx.Request.URL.Path)
        return nil
    },
)
```

### Usando o Builder Pattern para Hooks

Para hooks mais complexos, use o builder pattern:

```go
complexHook, err := hooks.NewCustomHookBuilder().
    WithName("security-monitor").
    WithEvents(interfaces.HookEventRequestStart, interfaces.HookEventRequestError).
    WithPriority(50).
    WithPathFilter(func(path string) bool {
        return !strings.HasPrefix(path, "/health")
    }).
    WithMethodFilter(func(method string) bool {
        return method != "OPTIONS"
    }).
    WithCondition(func(ctx *interfaces.HookContext) bool {
        return ctx.Request.Header.Get("X-Monitor") == "true"
    }).
    WithAsyncExecution(5, 3*time.Second).
    WithExecuteFunc(func(ctx *interfaces.HookContext) error {
        log.Printf("Monitoramento de segurança: %s %s",
            ctx.Request.Method, ctx.Request.URL.Path)
        return nil
    }).
    Build()

if err != nil {
    log.Fatal("Erro criando hook:", err)
}
```

## Custom Middleware

O custom middleware permite modificar requisições e respostas HTTP de forma flexível.

### Criação Básica de Middleware

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    "github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

func main() {
    // Criar factory de middleware
    middlewareFactory := middleware.NewCustomMiddlewareFactory()

    // Middleware simples para adicionar Request ID
    requestIDMiddleware := middlewareFactory.NewSimpleMiddleware(
        "request-id",
        100, // prioridade
        func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
                w.Header().Set("X-Request-ID", requestID)
                r.Header.Set("X-Request-ID", requestID)
                next.ServeHTTP(w, r)
            })
        },
    )

    fmt.Printf("Middleware criado: %s", requestIDMiddleware.Name())
}
```

### Middleware Condicional

Middleware que se aplica apenas a certos paths:

```go
// Middleware de autenticação apenas para rotas admin
adminAuthMiddleware := middlewareFactory.NewConditionalMiddleware(
    "admin-auth",
    200,
    func(path string) bool { 
        return !strings.HasPrefix(path, "/admin/") // skip non-admin paths
    },
    func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Header.Get("X-Admin-Token") == "" {
                http.Error(w, "Admin access required", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    },
)
```

### Middleware com Before/After

Middleware que executa código antes e depois da requisição:

```go
monitoringMiddleware, err := middleware.NewCustomMiddlewareBuilder().
    WithName("request-monitor").
    WithPriority(50).
    WithBeforeFunc(func(w http.ResponseWriter, r *http.Request) {
        // Código executado antes da requisição
        w.Header().Set("X-Start-Time", fmt.Sprintf("%d", time.Now().UnixNano()))
        log.Printf("Iniciando: %s %s", r.Method, r.URL.Path)
    }).
    WithAfterFunc(func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration) {
        // Código executado depois da requisição
        log.Printf("Finalizado: %s %s - %d (%v)", 
            r.Method, r.URL.Path, statusCode, duration)
    }).
    WithWrapFunc(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            next.ServeHTTP(w, r)
        })
    }).
    Build()
```

### Usando o Builder Pattern para Middleware

Para middleware complexo com múltiplas configurações:

```go
securityMiddleware, err := middleware.NewCustomMiddlewareBuilder().
    WithName("security").
    WithPriority(25).
    WithSkipPaths("/health", "/metrics").
    WithSkipFunc(func(path string) bool {
        return strings.HasSuffix(path, ".css") || 
               strings.HasSuffix(path, ".js")
    }).
    WithBeforeFunc(func(w http.ResponseWriter, r *http.Request) {
        // Headers de segurança
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
    }).
    WithAfterFunc(func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration) {
        // Log de eventos de segurança
        if statusCode >= 400 {
            log.Printf("Alerta de segurança: %s %s retornou %d", 
                r.Method, r.URL.Path, statusCode)
        }
    }).
    WithWrapFunc(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Rate limiting, validações, etc.
            next.ServeHTTP(w, r)
        })
    }).
    Build()
```

## Integração

### Usando Hooks e Middleware Juntos

```go
func setupIntegration() {
    hookFactory := hooks.NewCustomHookFactory()
    middlewareFactory := middleware.NewCustomMiddlewareFactory()

    // Middleware que adiciona metadata
    metadataMiddleware := middlewareFactory.NewSimpleMiddleware(
        "metadata-enricher",
        50,
        func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                r.Header.Set("X-Start-Time", fmt.Sprintf("%d", time.Now().UnixNano()))
                r.Header.Set("X-User-Type", detectUserType(r.UserAgent()))
                next.ServeHTTP(w, r)
            })
        },
    )

    // Hook que usa a metadata
    analyticsHook := hookFactory.NewConditionalHook(
        "enriched-analytics",
        []interfaces.HookEvent{interfaces.HookEventRequestEnd},
        600,
        func(ctx *interfaces.HookContext) bool {
            // Processa apenas traffic não-bot
            return ctx.Request.Header.Get("X-User-Type") != "bot"
        },
        func(ctx *interfaces.HookContext) error {
            userType := ctx.Request.Header.Get("X-User-Type")
            log.Printf("Analytics: %s de user agent %s (Duração: %v)",
                ctx.Request.URL.Path, userType, ctx.Duration)
            return nil
        },
    )
}
```

## Exemplos Completos

### Sistema de Rate Limiting Customizado

```go
func createRateLimitingSystem() {
    middlewareFactory := middleware.NewCustomMiddlewareFactory()
    hookFactory := hooks.NewCustomHookFactory()

    // Middleware de rate limiting
    rateLimitMiddleware := middlewareFactory.NewSimpleMiddleware(
        "rate-limiter",
        10,
        func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                clientIP := getClientIP(r)
                if isRateLimited(clientIP) {
                    w.Header().Set("X-Rate-Limited", "true")
                    http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                    return
                }
                next.ServeHTTP(w, r)
            })
        },
    )

    // Hook para monitorar rate limiting
    rateLimitHook := hookFactory.NewSimpleHook(
        "rate-limit-monitor",
        []interfaces.HookEvent{interfaces.HookEventRequestEnd},
        500,
        func(ctx *interfaces.HookContext) error {
            if ctx.StatusCode == http.StatusTooManyRequests {
                log.Printf("Rate limit hit: IP %s, Path: %s", 
                    getClientIP(ctx.Request), ctx.Request.URL.Path)
                // Poderia enviar alerta, incrementar métricas, etc.
            }
            return nil
        },
    )
}
```

### Sistema de Logging Estruturado

```go
func createStructuredLogging() {
    middlewareFactory := middleware.NewCustomMiddlewareFactory()

    loggingMiddleware := middlewareFactory.NewLoggingMiddleware(
        "structured-logger",
        75,
        func(method, path string, statusCode int, duration time.Duration) {
            log.Printf(`{"method":"%s","path":"%s","status":%d,"duration_ms":%d}`,
                method, path, statusCode, duration.Milliseconds())
        },
    )

    timingMiddleware := middlewareFactory.NewTimingMiddleware(
        "performance-timer",
        50,
        func(duration time.Duration, path string) {
            if duration > 1*time.Second {
                log.Printf(`{"alert":"slow_request","path":"%s","duration_ms":%d}`,
                    path, duration.Milliseconds())
            }
        },
    )
}
```

## Eventos Disponíveis

### Eventos de Servidor
- `HookEventServerStart`: Servidor iniciando
- `HookEventServerStop`: Servidor parando
- `HookEventServerReady`: Servidor pronto
- `HookEventServerError`: Erro no servidor

### Eventos de Requisição
- `HookEventRequestStart`: Início da requisição
- `HookEventRequestEnd`: Fim da requisição
- `HookEventRequestError`: Erro na requisição
- `HookEventRequestTimeout`: Timeout da requisição
- `HookEventRequestPanic`: Panic durante requisição

### Eventos de Middleware
- `HookEventMiddlewareStart`: Início do middleware
- `HookEventMiddlewareEnd`: Fim do middleware
- `HookEventMiddlewareError`: Erro no middleware

### Eventos Personalizados
- `HookEventCustom`: Para eventos customizados

## Melhores Práticas

1. **Prioridades**: Use prioridades baixas (0-50) para middleware crítico, médias (50-200) para funcionalidades, altas (200+) para logging/analytics.

2. **Performance**: Use hooks assíncronos para operações pesadas que não afetam a resposta.

3. **Filtros**: Use filtros para aplicar hooks/middleware apenas onde necessário.

4. **Error Handling**: Sempre trate erros nos hooks para não interromper o fluxo da aplicação.

5. **Testabilidade**: Crie factories customizadas para facilitar testes.

Para mais exemplos, veja o arquivo `examples/custom_usage.go`.
