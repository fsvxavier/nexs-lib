# Net/HTTP Framework Example

Este exemplo demonstra como usar a biblioteca de IP com o framework net/http padrão do Go.

## Funcionalidades

- Servidor HTTP completo usando net/http
- Middleware para extração de IP
- Múltiplos endpoints demonstrativos
- Validação de IP e controle de acesso
- Logging e monitoramento

## Como executar

```bash
cd pkg/ip/examples/nethttp
go run main.go
```

O servidor iniciará em `http://localhost:8080`

## Endpoints disponíveis

- `GET /` - Informações básicas do IP do cliente
- `GET /api/data` - API com validação de IP (bloqueia IPs privados)
- `GET /health` - Health check com informações do IP
- `GET /admin` - Endpoint administrativo com whitelist de IPs

## Como testar

```bash
# Teste básico
curl http://localhost:8080

# Teste com headers de proxy
curl -H 'X-Forwarded-For: 8.8.8.8' http://localhost:8080

# Teste com Cloudflare
curl -H 'CF-Connecting-IP: 203.0.113.100' http://localhost:8080/api/data

# Teste que será bloqueado (IP privado)
curl -H 'X-Real-IP: 192.168.1.1' http://localhost:8080/api/data
```

## Código principal

O exemplo demonstra:

1. **Middleware de IP**: Extrai e loga informações do IP
2. **Validação de segurança**: Bloqueia IPs privados em endpoints públicos
3. **Controle de acesso**: Whitelist de IPs para endpoints administrativos
4. **Respostas JSON estruturadas**: Com informações completas do IP

## Estrutura do middleware

```go
func IPMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientIP := ip.GetRealIP(r)
        ipInfo := ip.GetRealIPInfo(r)
        // ... logging e processamento
        next(w, r)
    }
}
```

## Benefícios

- ✅ Framework padrão do Go - sem dependências externas
- ✅ Compatível com qualquer router/mux
- ✅ Middleware reutilizável
- ✅ Fácil integração em projetos existentes

## Próximos passos

- Veja [Middleware Example](../middleware/) para exemplos avançados de middleware
- Explore outros frameworks: [Gin](../gin/), [Fiber](../fiber/), [Echo](../echo/)
