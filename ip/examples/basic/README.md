# Basic IP Library Example

Este exemplo demonstra o uso básico da biblioteca de identificação de IPs.

## Funcionalidades

- Extração de IP real do cliente
- Análise de informações detalhadas do IP
- Classificação de tipos de IP (público, privado, loopback, etc.)
- Demonstração de cadeia de proxy

## Como executar

```bash
cd pkg/ip/examples/basic
go run main.go
```

## O que este exemplo demonstra

1. **Extração básica de IP**: Como obter o IP real do cliente usando `ip.GetRealIP()`
2. **Informações detalhadas**: Como obter informações completas usando `ip.GetRealIPInfo()`
3. **Cadeia de proxy**: Como visualizar toda a cadeia de IPs usando `ip.GetIPChain()`
4. **Classificação de IP**: Como determinar se um IP é público, privado, IPv4, IPv6, etc.

## Cenários testados

- Conexão direta
- Proxy com X-Forwarded-For
- Cloudflare CDN
- Load balancer AWS
- Múltiplos proxies

## Saída esperada

O exemplo mostra informações detalhadas para cada cenário, incluindo:
- IP real extraído
- Tipo do IP
- Se é público/privado
- Versão (IPv4/IPv6)
- Fonte da informação
- Cadeia completa de IPs

## Próximos passos

Após entender este exemplo básico, explore:
- [Middleware Example](../middleware/) - Para uso em servidores HTTP
- [Net/HTTP Example](../nethttp/) - Para integração com net/http
- [Framework Examples](../) - Para outros frameworks web
