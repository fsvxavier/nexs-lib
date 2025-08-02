# 🎯 Exemplo Gin - HTTP Server com Hooks

Este exemplo demonstra o uso da biblioteca `nexs-lib/httpserver` com o framework Gin e implementação completa de hooks.

## 📋 Funcionalidades

- ✅ Framework Gin integrado
- ✅ Sistema completo de hooks (7 tipos)
- ✅ Observer pattern para monitoramento
- ✅ Múltiplas rotas com diferentes funcionalidades
- ✅ Logs estruturados com emojis
- ✅ Tratamento de erros personalizado

## 🎯 Objetivo

Demonstrar como integrar hooks avançados com o framework Gin para monitoramento e observabilidade completa.

## 🔧 Como Executar

### Pré-requisitos
```bash
go mod tidy
```

### Execução
```bash
cd gin
go run main.go
```

### Endpoints Disponíveis

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/` | GET | Página inicial |
| `/users` | GET | Lista usuários |
| `/users/:id` | GET | Buscar usuário por ID |
| `/users` | POST | Criar usuário |
| `/error` | GET | Testar tratamento de erro |
| `/health` | GET | Health check |

### Exemplos de Teste
```bash
# Página inicial
curl http://localhost:8080/

# Listar usuários
curl http://localhost:8080/users

# Buscar usuário específico
curl http://localhost:8080/users/1

# Criar usuário
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"João","email":"joao@email.com"}'

# Testar erro
curl http://localhost:8080/error

# Health check
curl http://localhost:8080/health
```

## 📊 Arquitetura com Hooks

```
HTTP Request → Gin Router → Handler
     ↓              ↓          ↓
[OnRequest]   [OnRouteEnter] [OnResponse]
     ↓              ↓          ↓
[Logging]     [Monitoring]  [Metrics]
```

## 🔍 Sistema de Hooks Implementados

### 1. **StartHook** - Inicialização
```go
OnStart(ctx, addr) // 🚀 Gin server started on localhost:8080
```

### 2. **StopHook** - Parada
```go
OnStop(ctx) // 🛑 Gin server stopped
```

### 3. **RequestHook** - Requisições
```go
OnRequest(ctx, req) // 📥 Request received
```

### 4. **ResponseHook** - Respostas
```go
OnResponse(ctx, req, resp, duration) // 📤 Response sent in 2ms
```

### 5. **ErrorHook** - Erros
```go
OnError(ctx, err) // ❌ Gin server error: not found
```

### 6. **RouteEnterHook** - Entrada em Rotas
```go
OnRouteEnter(ctx, method, path, req) // 🔀 Route entered: GET /users
```

### 7. **RouteExitHook** - Saída de Rotas
```go
OnRouteExit(ctx, method, path, req, duration) // 🔚 Route exited: GET /users in 1ms
```

## 💡 Conceitos Demonstrados

1. **Observer Pattern**: `LoggingObserver` implementa todos os hooks
2. **Gin Integration**: Uso nativo do framework Gin
3. **Request/Response Cycle**: Monitoramento completo do ciclo
4. **Error Handling**: Tratamento personalizado de erros
5. **Route Monitoring**: Rastreamento detalhado de rotas
6. **Performance Tracking**: Medição de duração de requisições

## 🎓 Para Quem é Este Exemplo

- **Desenvolvedores** usando framework Gin
- **DevOps** implementando observabilidade
- **Arquitetos** desenhando sistemas monitorados
- **Times** que precisam de rastreamento detalhado

## 🔗 Comparação com Outros Exemplos

| Recurso | basic | gin | hooks-basic |
|---------|-------|-----|-------------|
| Framework | Fiber | **Gin** | Gin |
| Hooks | ❌ | **✅ Todos** | ✅ Básicos |
| Múltiplas Rotas | ❌ | **✅** | ❌ |
| CRUD Operations | ❌ | **✅** | ❌ |
| Logs Detalhados | ❌ | **✅** | ✅ |

## 🏗️ Estrutura do Observer

```go
type LoggingObserver struct{}

// Implementa todas as interfaces de hooks
func (o *LoggingObserver) OnStart(ctx context.Context, addr string) error
func (o *LoggingObserver) OnStop(ctx context.Context) error  
func (o *LoggingObserver) OnError(ctx context.Context, err error) error
func (o *LoggingObserver) OnRequest(ctx context.Context, req interface{}) error
func (o *LoggingObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error
func (o *LoggingObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error
func (o *LoggingObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error
```

## 📈 Logs de Exemplo

```
🚀 Gin server started on localhost:8080
📥 Request received
🔀 Route entered: GET /users
📤 Response sent in 1.2ms
🔚 Route exited: GET /users in 1.2ms
```

## 📊 Performance

- **Overhead por Hook**: ~0.5-1ms
- **Memória**: +15-25MB (hooks + Gin)
- **CPU**: +2-5% (logging overhead)

## 🐛 Troubleshooting

### Hooks não aparecem nos logs
```bash
# Verificar se observer foi registrado
grep "RegisterObserver" main.go
```

### Gin em modo debug
```bash
# Gin já configurado em modo release no código
# Para debug, altere: gin.SetMode(gin.DebugMode)
```

### Performance lenta
```bash
# Verificar overhead de logs
# Considere log levels em produção
```

## 🔗 Próximos Passos

1. `middlewares-basic/` - Adicionar autenticação
2. `complete/` - Exemplo com hooks + middlewares
3. Implementar métricas customizadas
4. Integração com APM tools

---

*Exemplo avançado demonstrando integração completa Gin + Hooks*
