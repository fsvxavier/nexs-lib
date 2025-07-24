# Atreugo Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework Atreugo.

## Funcionalidades

- Simulação de uso com Atreugo framework
- Handlers com JSON helpers
- Roteamento baseado em tipo de IP
- Exemplo de código real comentado

## Como executar

```bash
cd pkg/ip/examples/atreugo
go run main.go
```

## O que este exemplo demonstra

1. **Simulação de contexto Atreugo**: Como a biblioteca funciona com requests HTTP
2. **JSON responses**: Uso dos helpers JSON do Atreugo
3. **Roteamento inteligente**: Decisões baseadas no tipo de IP
4. **Handlers estruturados**: Uso das informações de IP nos handlers

## Cenários simulados

- API REST de alta performance
- Endpoint GraphQL
- Conexão WebSocket

## Uso real com Atreugo

Para usar com o Atreugo real, descomente o código no final do arquivo e adicione a dependência:

```bash
go mod tidy
go get github.com/savsgio/atreugo/v11
```

### Exemplo de handler Atreugo

```go
func atreugoHandler(ctx *atreugo.RequestCtx) error {
    clientIP := ip.GetRealIP(ctx)
    ipInfo := ip.GetRealIPInfo(ctx)

    if ipInfo.IsPrivate {
        return ctx.JSONResponse(map[string]interface{}{
            "error": "Access denied",
            "clientIP": clientIP,
        }, atreugo.StatusForbidden)
    }

    return ctx.JSONResponse(map[string]interface{}{
        "clientIP": clientIP,
        "ipType": ipInfo.Type.String(),
        "isPublic": ipInfo.IsPublic,
        "framework": "atreugo",
    }, atreugo.StatusOK)
}
```

### Servidor Atreugo completo

```go
func main() {
    config := atreugo.Config{
        Host: "0.0.0.0",
        Port: 8080,
    }
    
    server := atreugo.New(config)

    // Middleware para todas as rotas
    server.UseBefore(func(ctx *atreugo.RequestCtx) error {
        clientIP := ip.GetRealIP(ctx)
        ipInfo := ip.GetRealIPInfo(ctx)
        
        ctx.SetUserValue("clientIP", clientIP)
        ctx.SetUserValue("ipInfo", ipInfo)
        
        return ctx.Next()
    })

    server.GET("/", atreugoHandler)
    server.GET("/health", healthHandler)

    log.Fatal(server.ListenAndServe())
}
```

## Benefícios

- ✅ Performance do FastHTTP com API amigável
- ✅ JSON helpers built-in
- ✅ Middleware system
- ✅ User values para context storage

## Recursos do Atreugo utilizados

- JSON responses (ctx.JSONResponse)
- User values (ctx.SetUserValue/ctx.UserValue)
- Middleware chain (UseBefore/UseAfter)
- HTTP status constants
- Request context access

## Características do Atreugo

O Atreugo combina:
- Performance do FastHTTP
- API mais amigável que FastHTTP puro
- Middleware system integrado
- JSON helpers automáticos
- Configuração simplificada

## Quando usar Atreugo

Ideal para:
- APIs REST de alta performance
- Aplicações que precisam de performance mas querem API simples
- Migração de FastHTTP puro para algo mais amigável
- Projetos que valorizam both performance e developer experience

## Próximos passos

- Implemente em seu projeto Atreugo
- Compare performance com [FastHTTP Example](../fasthttp/)
- Veja [Fiber Example](../fiber/) para alternativa similar
