# Gin Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework Gin.

## Funcionalidades

- Simulação de uso com Gin framework
- Middleware para extração de IP
- Handlers com validação de IP
- Exemplo de código real comentado

## Como executar

```bash
cd pkg/ip/examples/gin
go run main.go
```

## O que este exemplo demonstra

1. **Simulação de contexto Gin**: Como a biblioteca funciona com requests HTTP
2. **Middleware de IP**: Extração e armazenamento no contexto
3. **Validação de segurança**: Bloqueio de IPs privados
4. **Handlers estruturados**: Uso das informações de IP nos handlers

## Cenários simulados

- Requisição com Cloudflare CDN
- Requisição através de Load Balancer
- Requisição de rede privada (bloqueada)

## Uso real com Gin

Para usar com o Gin real, descomente o código no final do arquivo e adicione a dependência:

```bash
go mod tidy
go get github.com/gin-gonic/gin
```

### Exemplo de middleware Gin

```go
func IPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := ip.GetRealIP(c.Request)
        ipInfo := ip.GetRealIPInfo(c.Request)
        
        // Armazenar no contexto Gin
        c.Set("clientIP", clientIP)
        c.Set("ipInfo", ipInfo)
        
        c.Next()
    }
}
```

### Exemplo de handler

```go
func apiHandler(c *gin.Context) {
    clientIP := c.GetString("clientIP")
    ipInfoInterface, _ := c.Get("ipInfo")
    ipInfo := ipInfoInterface.(*ip.IPInfo)
    
    if ipInfo.IsPrivate {
        c.JSON(403, gin.H{"error": "Access denied"})
        return
    }
    
    c.JSON(200, gin.H{
        "clientIP": clientIP,
        "data": []string{"item1", "item2"},
    })
}
```

## Benefícios

- ✅ Framework web rápido e popular
- ✅ Middleware fácil de implementar
- ✅ Contexto integrado para armazenar dados
- ✅ JSON helpers built-in

## Recursos do Gin utilizados

- Middleware chain
- Context storage (c.Set/c.Get)
- JSON responses (c.JSON)
- Request access (c.Request)

## Próximos passos

- Implemente o middleware em seu projeto Gin
- Veja [Net/HTTP Example](../nethttp/) para comparação
- Explore [Fiber Example](../fiber/) para alternativa similar
