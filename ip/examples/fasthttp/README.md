# FastHTTP Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework FastHTTP.

## Funcionalidades

- Simulação de uso com FastHTTP framework
- Handlers de alta performance
- Otimizações baseadas em tipo de IP
- Exemplo de código real comentado

## Como executar

```bash
cd pkg/ip/examples/fasthttp
go run main.go
```

## O que este exemplo demonstra

1. **Simulação de contexto FastHTTP**: Como a biblioteca funciona com requests HTTP
2. **Handlers otimizados**: Extração direta de IP sem overhead
3. **Performance**: Otimizações baseadas no tipo de IP (CDN cache)
4. **Resposta direta**: Uso das informações de IP nas respostas

## Cenários simulados

- Requisição de API de alta performance
- Comunicação entre microsserviços
- Requisição WebSocket upgrade

## Uso real com FastHTTP

Para usar com o FastHTTP real, descomente o código no final do arquivo e adicione a dependência:

```bash
go mod tidy
go get github.com/valyala/fasthttp
```

### Exemplo de handler FastHTTP

```go
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
    clientIP := ip.GetRealIP(ctx)
    ipInfo := ip.GetRealIPInfo(ctx)

    // Resposta otimizada
    ctx.SetContentType("application/json")
    fmt.Fprintf(ctx, `{
        "clientIP": "%s",
        "ipType": "%s",
        "isPublic": %t,
        "framework": "fasthttp"
    }`, clientIP, ipInfo.Type.String(), ipInfo.IsPublic)
}
```

### Servidor FastHTTP completo

```go
func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        clientIP := ip.GetRealIP(ctx)
        ipInfo := ip.GetRealIPInfo(ctx)

        switch string(ctx.Path()) {
        case "/":
            // Handler principal
        case "/health":
            // Health check
        }
    }

    log.Fatal(fasthttp.ListenAndServe(":8080", handler))
}
```

## Benefícios

- ✅ Performance extrema (10x mais rápido que net/http)
- ✅ Baixo uso de memória
- ✅ Zero allocations em hot paths
- ✅ Ideal para alta concorrência

## Características do FastHTTP

O FastHTTP é ideal para:
- APIs de alta performance
- Microsserviços
- Proxies e load balancers
- Aplicações com alta carga
- WebSocket servers

## Optimizações demonstradas

- Cache CDN baseado em tipo de IP
- Bypass de verificações para IPs internos
- Respostas otimizadas sem middleware overhead

## Comparação de Performance

FastHTTP vs net/http:
- ~10x mais rápido em throughput
- ~10x menos allocações de memória
- Melhor para alta concorrência
- API mais low-level

## Próximos passos

- Implemente em seu projeto FastHTTP de alta performance
- Veja [Atreugo Example](../atreugo/) para wrapper mais amigável
- Compare com [Net/HTTP Example](../nethttp/) para diferenças
