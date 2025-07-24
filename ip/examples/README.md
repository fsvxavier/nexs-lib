# IP Library Examples

Esta pasta contém exemplos completos demonstrando como usar a biblioteca de identificação e manipulação de IPs em diferentes cenários e frameworks.

## 📁 Estrutura dos Exemplos

### [📂 basic/](./basic/)
**Uso básico da biblioteca**
- Extração de IP real do cliente
- Análise de informações detalhadas do IP
- Classificação de tipos de IP
- Demonstração de cadeia de proxy

### [📂 nethttp/](./nethttp/)
**Framework net/http padrão**
- Servidor HTTP completo
- Middleware para extração de IP
- Múltiplos endpoints demonstrativos
- Validação de IP e controle de acesso

### [📂 middleware/](./middleware/)
**Middlewares avançados**
- Cadeia de múltiplos middlewares
- Logging, segurança, geolocalização
- Rate limiting baseado em IP
- Headers informativos

## 🌐 Frameworks Web Suportados

### [📂 gin/](./gin/)
**Gin Web Framework**
- Middleware para contexto Gin
- Handlers com validação de IP
- Uso do contexto Gin para armazenamento

### [📂 fiber/](./fiber/)
**Fiber Framework**
- Middleware de alta performance
- Rate limiting inteligente
- Context storage do Fiber

### [📂 echo/](./echo/)
**Echo Framework**
- Middleware chain do Echo
- Geolocalização baseada em IP
- JSON responses estruturadas

### [📂 fasthttp/](./fasthttp/)
**FastHTTP Framework**
- Performance extrema
- Handlers otimizados
- Zero allocation paths

### [📂 atreugo/](./atreugo/)
**Atreugo Framework**
- FastHTTP com API amigável
- JSON helpers integrados
- User values para contexto

## 🚀 Como executar os exemplos

Cada exemplo pode ser executado independentemente:

```bash
# Exemplo básico
cd basic && go run main.go

# Servidor net/http
cd nethttp && go run main.go

# Middlewares avançados
cd middleware && go run main.go

# Framework específico (exemplo: Gin)
cd gin && go run main.go
```

## 🧪 Testando os exemplos

### Testes básicos
```bash
# Teste local
curl http://localhost:8080

# Teste com proxy headers
curl -H 'X-Forwarded-For: 8.8.8.8, 192.168.1.1' http://localhost:8080

# Teste com Cloudflare
curl -H 'CF-Connecting-IP: 203.0.113.100' http://localhost:8080
```

### Testes avançados
```bash
# Teste multiple proxies
curl -H 'X-Forwarded-For: 203.0.113.45, 198.51.100.10, 172.16.0.1' http://localhost:8080

# Teste rate limiting
for i in {1..12}; do curl http://localhost:8080/api/info; done

# Teste geolocalização
curl -H 'X-Real-IP: 8.8.8.8' http://localhost:8080
```

## 📋 O que cada exemplo demonstra

| Exemplo | Funcionalidades | Ideal para |
|---------|----------------|------------|
| **basic** | Uso fundamental da biblioteca | Aprender conceitos básicos |
| **nethttp** | Servidor HTTP completo | Integração com projetos Go padrão |
| **middleware** | Middlewares avançados | Sistemas complexos com múltiplas camadas |
| **gin** | Framework Gin | APIs REST rápidas e populares |
| **fiber** | Framework Fiber | Performance com API familiar |
| **echo** | Framework Echo | APIs robustas e middleware avançado |
| **fasthttp** | FastHTTP puro | Performance extrema |
| **atreugo** | Atreugo | FastHTTP com API amigável |

## 🎯 Casos de uso demonstrados

### Segurança
- Bloqueio de IPs maliciosos
- Whitelist para endpoints administrativos
- Validação de origem das requisições

### Performance
- Rate limiting por IP
- Cache baseado em tipo de IP
- Otimizações para IPs internos

### Analytics
- Logging detalhado de IPs
- Geolocalização de usuários
- Análise de cadeia de proxy

### Compliance
- Auditoria de acesso por IP
- Headers de transparência
- Rastreabilidade de requisições

## 🔧 Dependências

### Exemplos básicos (sem dependências externas)
- `basic/` - apenas biblioteca padrão
- `nethttp/` - apenas net/http
- `middleware/` - apenas net/http

### Exemplos de frameworks (precisam de dependências)
Para usar os exemplos de frameworks, adicione as dependências:

```bash
# Gin
go get github.com/gin-gonic/gin

# Fiber
go get github.com/gofiber/fiber/v2

# Echo
go get github.com/labstack/echo/v4

# FastHTTP
go get github.com/valyala/fasthttp

# Atreugo
go get github.com/savsgio/atreugo/v11
```

## 📚 Próximos passos

1. **Comece com basic/** - Para entender os conceitos fundamentais
2. **Explore nethttp/** - Para ver integração completa
3. **Teste middleware/** - Para casos de uso avançados
4. **Escolha seu framework** - Baseado nas suas necessidades
5. **Adapte para seu projeto** - Use como base para sua implementação

## 🆘 Suporte

- Veja o [README principal](../README.md) para documentação completa da API
- Consulte [next_steps.md](../next_steps.md) para roadmap e melhorias
- Cada exemplo tem seu próprio README com detalhes específicos

## Prerequisites

Make sure you have Go 1.19+ installed and the IP library available in your module path.
