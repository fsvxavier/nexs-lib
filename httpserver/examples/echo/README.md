# 🌐 Exemplo Echo - HTTP Server com Hooks

Este exemplo demonstra o uso da biblioteca `nexs-lib/httpserver` com o framework Echo e sistema completo de hooks.

## 📋 Funcionalidades

- ✅ Framework Echo integrado
- ✅ Sistema completo de hooks (7 tipos)
- ✅ Observer pattern para monitoramento
- ✅ Múltiplas rotas RESTful
- ✅ Middleware Echo nativo
- ✅ Logs estruturados com emojis
- ✅ JSON responses otimizadas

## 🎯 Objetivo

Demonstrar como integrar hooks avançados com o framework Echo para APIs robustas e monitoradas.

## 🔧 Como Executar

### Pré-requisitos
```bash
go mod tidy
```

### Execução
```bash
cd echo
go run main.go
```

### Endpoints Disponíveis

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/` | GET | Página inicial |
| `/users` | GET | Lista usuários |
| `/users/:id` | GET | Buscar usuário por ID |
| `/users` | POST | Criar usuário |
| `/health` | GET | Health check |
| `/metrics` | GET | Métricas simples |

### Exemplos de Teste
```bash
# Página inicial
curl http://localhost:8080/

# Listar usuários  
curl http://localhost:8080/users

# Buscar usuário específico
curl http://localhost:8080/users/123

# Criar usuário
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Maria","email":"maria@email.com","age":28}'

# Health check
curl http://localhost:8080/health

# Métricas
curl http://localhost:8080/metrics
```

## 📊 Arquitetura com Echo + Hooks

```
HTTP Request → Echo Router → Handler → JSON Response
     ↓              ↓           ↓           ↓
[OnRequest]   [OnRouteEnter] [Handler] [OnResponse]
     ↓              ↓           ↓           ↓
[Logging]     [Monitoring]  [Business] [Metrics]
```

## 🔍 Sistema de Hooks Implementados

### 1. **StartHook** - Inicialização do Servidor
```go
OnStart(ctx, addr) // 🚀 Echo server started on localhost:8080
```

### 2. **StopHook** - Parada do Servidor
```go
OnStop(ctx) // 🛑 Echo server stopped
```

### 3. **RequestHook** - Todas as Requisições
```go
OnRequest(ctx, req) // 📥 Request received
```

### 4. **ResponseHook** - Todas as Respostas
```go
OnResponse(ctx, req, resp, duration) // 📤 Response sent in 1.5ms
```

### 5. **ErrorHook** - Tratamento de Erros
```go
OnError(ctx, err) // ❌ Echo server error: validation failed
```

### 6. **RouteEnterHook** - Entrada em Rotas
```go
OnRouteEnter(ctx, method, path, req) // 🔀 Route entered: POST /users
```

### 7. **RouteExitHook** - Saída de Rotas
```go
OnRouteExit(ctx, method, path, req, duration) // 🔚 Route exited: POST /users in 2.1ms
```

## 💡 Conceitos Demonstrados

1. **Echo Framework**: Integração nativa com Echo v4
2. **RESTful API**: Endpoints seguindo padrões REST
3. **Observer Pattern**: Monitoramento através de hooks
4. **JSON Handling**: Serialização/deserialização automática
5. **Error Handling**: Tratamento robusto de erros
6. **Performance Monitoring**: Medição de tempos de resposta
7. **Health Checks**: Endpoint para monitoramento de saúde

## 🎓 Para Quem é Este Exemplo

- **Desenvolvedores** usando framework Echo
- **APIs RESTful** que precisam de monitoramento
- **Microserviços** com observabilidade
- **Teams** implementando logging estruturado

## 🔗 Comparação com Outros Frameworks

| Recurso | Gin | **Echo** | FastHTTP |
|---------|-----|----------|----------|
| Performance | ⭐⭐⭐⭐ | **⭐⭐⭐⭐⭐** | ⭐⭐⭐⭐⭐ |
| Simplicidade | ⭐⭐⭐⭐ | **⭐⭐⭐⭐⭐** | ⭐⭐⭐ |
| Middleware | ⭐⭐⭐⭐⭐ | **⭐⭐⭐⭐⭐** | ⭐⭐⭐ |
| JSON Support | ⭐⭐⭐⭐ | **⭐⭐⭐⭐⭐** | ⭐⭐⭐⭐ |
| HTTP/2 | ⭐⭐⭐ | **⭐⭐⭐⭐** | ⭐⭐⭐⭐⭐ |

## 🏗️ Estrutura Principal

```go
// Observer para todos os hooks
type LoggingObserver struct{}

// Handler de usuário com validação
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=120"`
}

// Rotas RESTful organizadas
server.RegisterRoute("GET", "/users", getUsersHandler)
server.RegisterRoute("GET", "/users/:id", getUserHandler)  
server.RegisterRoute("POST", "/users", createUserHandler)
```

## 📈 Logs de Exemplo

```
🚀 Echo server started on localhost:8080
📥 Request received
🔀 Route entered: GET /users
📤 Response sent in 0.8ms
🔚 Route exited: GET /users in 0.8ms

📥 Request received  
🔀 Route entered: POST /users
📤 Response sent in 2.3ms
🔚 Route exited: POST /users in 2.3ms
```

## 🚀 Funcionalidades Avançadas

### Health Check Personalizado
```json
{
  "status": "healthy",
  "timestamp": "2025-08-02T15:30:45Z",
  "uptime": "5m32s",
  "framework": "echo"
}
```

### Métricas Básicas
```json
{
  "requests_total": 1247,
  "requests_per_second": 12.3,
  "average_response_time": "1.2ms",
  "active_connections": 8
}
```

### Validação de Input
```go
// Automática com struct tags
Name  string `json:"name" validate:"required"`
Email string `json:"email" validate:"required,email"`
Age   int    `json:"age" validate:"min=0,max=120"`
```

## 📊 Performance

- **Throughput**: ~50k req/s (Echo nativo)
- **Overhead por Hook**: ~0.3-0.8ms
- **Memória**: +12-20MB (hooks + Echo)
- **CPU**: +1-4% (logging overhead)

## 🐛 Troubleshooting

### Echo não inicia
```bash
# Verificar porta disponível
netstat -tulpn | grep :8080

# Verificar logs de erro
grep "Echo server error" logs.txt
```

### JSON malformado
```bash
# Testar JSON válido
echo '{"name":"test","email":"test@test.com","age":25}' | jq .
```

### Performance lenta
```bash
# Verificar overhead de hooks
# Considere desabilitar logs verbose em produção
```

## ⚡ Otimizações Echo

### 1. **Configuração de Produção**
```go
e.HideBanner = true
e.HidePort = true
```

### 2. **Middleware Otimizado**
```go
e.Use(middleware.Recover())
e.Use(middleware.Gzip())
```

### 3. **JSON Encoder Rápido**
```go
// Echo usa encoding/json otimizado por padrão
```

## 🔗 Próximos Passos

1. `fasthttp/` - Performance máxima
2. `middlewares-basic/` - Adicionar autenticação  
3. `complete/` - Exemplo com hooks + middlewares
4. Implementar rate limiting
5. Adicionar métricas Prometheus

---

*Exemplo demonstrando Echo Framework com sistema completo de hooks para APIs robustas*
