# Fiber Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework Fiber.

## Funcionalidades

- Simulação de uso com Fiber framework
- Middleware para extração de IP
- Rate limiting baseado em tipo de IP
- Exemplo de código real comentado

## Como executar

```bash
cd pkg/ip/examples/fiber
go run main.go
```

## O que este exemplo demonstra

1. **Simulação de contexto Fiber**: Como a biblioteca funciona com requests HTTP
2. **Middleware de IP**: Extração e armazenamento no contexto
3. **Rate limiting inteligente**: Aplicado diferentemente para IPs públicos/privados
4. **Handlers otimizados**: Uso das informações de IP nos handlers

## Cenários simulados

- Requisição com Cloudflare CDN
- Requisição através de Load Balancer
- Conexão direta (sem proxy)

## Uso real com Fiber

Para usar com o Fiber real, descomente o código no final do arquivo e adicione a dependência:

```bash
go mod tidy
go get github.com/gofiber/fiber/v2
```

### Exemplo de middleware Fiber

```go
func IPMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        clientIP := ip.GetRealIP(c.Context())
        ipInfo := ip.GetRealIPInfo(c.Context())
        
        // Armazenar no contexto Fiber
        c.Locals("clientIP", clientIP)
        c.Locals("ipInfo", ipInfo)
        
        return c.Next()
    }
}
```

### Exemplo de handler

```go
func apiHandler(c *fiber.Ctx) error {
    clientIP := c.Locals("clientIP").(string)
    ipInfo := c.Locals("ipInfo").(*ip.IPInfo)
    
    if ipInfo.IsPrivate {
        return c.Status(403).JSON(fiber.Map{
            "error": "Access denied",
            "clientIP": clientIP,
        })
    }
    
    return c.JSON(fiber.Map{
        "clientIP": clientIP,
        "data": []string{"item1", "item2"},
    })
}
```

## Benefícios

- ✅ Framework ultra-rápido baseado em FastHTTP
- ✅ API similar ao Express.js
- ✅ Context storage (c.Locals)
- ✅ Built-in JSON helpers

## Recursos do Fiber utilizados

- Middleware handler
- Context storage (c.Locals)
- JSON responses (c.JSON)
- HTTP context access (c.Context())
- Status codes (c.Status)

## Performance

O Fiber é uma excelente escolha para aplicações que precisam de:
- Alta performance
- Baixo uso de memória
- API familiar (Express-like)
- Middleware avançado

## Próximos passos

- Implemente o middleware em seu projeto Fiber
- Veja [FastHTTP Example](../fasthttp/) para comparação de performance
- Explore [Echo Example](../echo/) para alternativa similar
