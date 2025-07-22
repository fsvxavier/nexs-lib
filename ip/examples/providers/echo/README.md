# Echo Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework Echo.

## Funcionalidades

- Simulação de uso com Echo framework
- Middleware para extração de IP
- Geolocalização baseada em IP
- Exemplo de código real comentado

## Como executar

```bash
cd pkg/ip/examples/echo
go run main.go
```

## O que este exemplo demonstra

1. **Simulação de contexto Echo**: Como a biblioteca funciona com requests HTTP
2. **Middleware de IP**: Extração e armazenamento no contexto
3. **Geolocalização**: Decisões baseadas no tipo de IP
4. **Handlers estruturados**: Uso das informações de IP nos handlers

## Cenários simulados

- Requisição AWS ALB
- Requisição Nginx Proxy
- Requisição de serviço interno

## Uso real com Echo

Para usar com o Echo real, descomente o código no final do arquivo e adicione a dependência:

```bash
go mod tidy
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
```

### Exemplo de middleware Echo

```go
func IPMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            clientIP := ip.GetRealIP(c.Request())
            ipInfo := ip.GetRealIPInfo(c.Request())
            
            // Armazenar no contexto Echo
            c.Set("clientIP", clientIP)
            c.Set("ipInfo", ipInfo)
            
            return next(c)
        }
    }
}
```

### Exemplo de handler

```go
func apiHandler(c echo.Context) error {
    clientIP := c.Get("clientIP").(string)
    ipInfo := c.Get("ipInfo").(*ip.IPInfo)
    
    if ipInfo.IsPrivate {
        return c.JSON(http.StatusForbidden, map[string]interface{}{
            "error": "Access denied",
            "clientIP": clientIP,
        })
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "clientIP": clientIP,
        "data": []string{"item1", "item2"},
    })
}
```

## Benefícios

- ✅ Framework minimalista e eficiente
- ✅ Middleware chain poderoso
- ✅ Context storage integrado
- ✅ HTTP status codes estruturados

## Recursos do Echo utilizados

- Middleware functions
- Context storage (c.Set/c.Get)
- JSON responses (c.JSON)
- Request access (c.Request())
- HTTP status constants

## Características do Echo

O Echo é ideal para:
- APIs REST robustas
- Middleware complexo
- Logging e recover automático
- Binding e validação
- Static file serving

## Próximos passos

- Implemente o middleware em seu projeto Echo
- Veja [Net/HTTP Example](../nethttp/) para comparação
- Explore [Gin Example](../gin/) para alternativa popular
