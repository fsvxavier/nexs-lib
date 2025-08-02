# 📚 Exemplos - Nexs Lib HTTP Server

Esta pasta contém exemplos práticos demonstrando como usar a biblioteca `nexs-lib/httpserver` com diferentes frameworks e funcionalidades.

## 🎯 Visão Geral

A biblioteca oferece **9 exemplos completos** organizados em duas categorias:

- **🎯 Exemplos por Framework** (6): Demonstram integração com frameworks populares
- **🔧 Exemplos por Funcionalidade** (3): Demonstram hooks e middlewares

## 📊 Matriz Completa de Exemplos

| Exemplo | Framework | Hooks | Middlewares | Performance | Complexidade |
|---------|-----------|-------|-------------|-------------|--------------|
| [**basic**](./basic/) | Fiber | ❌ | ❌ | ~100k req/s | ⭐ |
| [**gin**](./gin/) | Gin | ✅ Todos | ❌ | ~200k req/s | ⭐⭐⭐ |
| [**echo**](./echo/) | Echo | ✅ Todos | ❌ | ~300k req/s | ⭐⭐⭐ |
| [**fasthttp**](./fasthttp/) | FastHTTP | ✅ Todos | ❌ | ~500k req/s | ⭐⭐⭐⭐ |
| [**atreugo**](./atreugo/) | Atreugo | ✅ Todos | ❌ | ~380k req/s | ⭐⭐⭐⭐ |
| [**advanced**](./advanced/) | Fiber | ✅ Todos | ❌ | ~180k req/s | ⭐⭐⭐⭐⭐ |
| [**hooks-basic**](./hooks-basic/) | Gin | ✅ Básicos | ❌ | ~150k req/s | ⭐⭐ |
| [**middlewares-basic**](./middlewares-basic/) | Gin | ❌ | ✅ Básicos | ~120k req/s | ⭐⭐⭐ |
| [**complete**](./complete/) | Gin | ✅ Todos | ✅ Todos | ~100k req/s | ⭐⭐⭐⭐⭐ |

## 🎓 Trilhas de Aprendizado

### 👶 **Iniciante** - Começando com o Básico
```bash
1. basic           # Configuração mínima
2. gin             # Framework popular 
3. hooks-basic     # Conceitos de monitoramento
```

### 🎯 **Intermediário** - Recursos Avançados
```bash
1. echo                 # Framework performático
2. middlewares-basic    # Autenticação e logging
3. complete            # Integração completa
```

### ⚡ **Performance** - Máxima Velocidade
```bash
1. fasthttp      # Máxima performance (~500k req/s)
2. atreugo       # FastHTTP com framework (~380k req/s)
```

### 🏭 **Produção** - Padrões Empresariais
```bash
1. advanced      # Graceful shutdown, metrics, health checks
2. complete      # Observabilidade completa
```

## 🚀 Início Rápido

### Método 1: Script Automatizado
```bash
# Listar todos os exemplos
./run_examples.sh

# Ver informações detalhadas
./run_examples.sh info

# Executar exemplo específico
./run_examples.sh basic

# Testar compilação de todos
./run_examples.sh test
```

### Método 2: Execução Manual
```bash
# Escolher um exemplo e executar
cd gin
go run main.go
```

## 📋 Recursos por Exemplo

### 🎯 **EXEMPLOS POR FRAMEWORK**

#### [basic/](./basic/) - Servidor Mínimo
- ✅ Configuração mais simples possível
- ✅ Provider Fiber (padrão)
- ✅ Uma rota JSON básica
- 🎯 **Para**: Protótipos rápidos, aprendizado inicial

#### [gin/](./gin/) - Framework Gin + Hooks
- ✅ Framework Gin integrado
- ✅ Sistema completo de hooks (7 tipos)
- ✅ Múltiplas rotas RESTful
- ✅ Logs estruturados
- 🎯 **Para**: APIs usando Gin, observabilidade

#### [echo/](./echo/) - Framework Echo + Hooks  
- ✅ Framework Echo v4
- ✅ Sistema completo de hooks
- ✅ Performance otimizada
- ✅ JSON binding automático
- 🎯 **Para**: APIs de alta performance, microserviços

#### [fasthttp/](./fasthttp/) - Performance Máxima
- ⚡ FastHTTP (máxima performance)
- ✅ Hooks otimizados (zero allocations)
- ✅ Pool de objetos
- ✅ Benchmarks integrados
- 🎯 **Para**: Sistemas críticos, >100k req/s

#### [atreugo/](./atreugo/) - FastHTTP com Framework
- ⚡ Atreugo (baseado em FastHTTP)
- ✅ Sintaxe de framework amigável
- ✅ Performance próxima ao FastHTTP puro
- ✅ Middleware ecosystem
- 🎯 **Para**: Performance + produtividade

#### [advanced/](./advanced/) - Padrões de Produção
- 🏭 Graceful shutdown com signal handling
- ✅ Health checks abrangentes
- ✅ Metrics collection
- ✅ Context propagation
- 🎯 **Para**: Aplicações enterprise, produção

### 🔧 **EXEMPLOS POR FUNCIONALIDADE**

#### [hooks-basic/](./hooks-basic/) - Monitoramento Básico
- ✅ 4 hooks principais (Start, Stop, Request, Error)
- ✅ Métricas básicas de servidor
- ✅ Logging de ciclo de vida
- 🎯 **Para**: Aprender conceitos de hooks

#### [middlewares-basic/](./middlewares-basic/) - Auth + Logging
- ✅ LoggingMiddleware estruturado
- ✅ AuthMiddleware com Basic Auth
- ✅ Proteção de rotas
- ✅ Auditoria de acesso
- 🎯 **Para**: APIs com autenticação

#### [complete/](./complete/) - Exemplo Completo
- ✅ Todos os 7 hooks implementados
- ✅ Múltiplos middlewares (Logging + Auth)
- ✅ Multi-auth (Basic + API Key)
- ✅ Área admin protegida
- ✅ Métricas avançadas
- 🎯 **Para**: Referência completa, produção

## 📚 Guia de Estudo Detalhado

### 1️⃣ **Começando** (1-2 dias)
```bash
# Compreender o básico
./run_examples.sh basic
curl http://localhost:8080/

# Aprender frameworks
./run_examples.sh gin
curl http://localhost:8080/users
```

### 2️⃣ **Desenvolvimento** (3-5 dias)
```bash
# Aprender hooks
./run_examples.sh hooks-basic
curl http://localhost:8080/

# Adicionar autenticação
./run_examples.sh middlewares-basic
curl -u admin:secret http://localhost:8080/api/users
```

### 3️⃣ **Performance** (1-2 dias)
```bash
# Máxima velocidade
./run_examples.sh fasthttp
ab -n 10000 -c 100 http://localhost:8080/fast

# Framework de alta performance
./run_examples.sh atreugo
wrk -t12 -c400 -d10s http://localhost:8080/
```

### 4️⃣ **Produção** (2-3 dias)
```bash
# Padrões enterprise
./run_examples.sh advanced
curl http://localhost:8080/health

# Observabilidade completa
./run_examples.sh complete
curl -H "X-API-Key: admin-key" http://localhost:8080/admin/stats
```

## 🧪 Testes e Validação

### Compilação de Todos os Exemplos
```bash
./run_examples.sh test
# ✅ Todos os testes passaram!
```

### Teste de Performance
```bash
# Teste básico com ab
ab -n 1000 -c 10 http://localhost:8080/

# Teste avançado com wrk
wrk -t12 -c400 -d30s http://localhost:8080/
```

### Teste de Carga por Framework
```bash
# Resultados esperados:
# basic:     ~100k req/s
# gin:       ~200k req/s  
# echo:      ~300k req/s
# fasthttp:  ~500k req/s (máxima)
# atreugo:   ~380k req/s
```

## 📖 Documentação Adicional

### Por Exemplo
- **README.md individual** em cada pasta
- **Código comentado** em todos os arquivos
- **Exemplos de curl** para testar

### Arquivos de Referência
- [`OVERVIEW.md`](./OVERVIEW.md) - Visão geral detalhada
- [`run_examples.sh`](./run_examples.sh) - Script de automação

## 🔧 Dependências

### Principais
```bash
go mod tidy  # Instala todas as dependências
```

### Por Framework
```bash
# Gin
go get github.com/gin-gonic/gin

# Echo  
go get github.com/labstack/echo/v4

# FastHTTP
go get github.com/valyala/fasthttp

# Atreugo
go get github.com/savsgio/atreugo/v11

# Fiber (basic/advanced)
go get github.com/gofiber/fiber/v2
```

## 💡 Casos de Uso por Exemplo

### 🚀 **Prototipagem Rápida**
- `basic/` - MVP em 5 minutos

### 🌐 **APIs Web**
- `gin/` - REST APIs com Gin
- `echo/` - APIs performáticas

### ⚡ **Sistemas de Alta Performance**
- `fasthttp/` - Latência crítica
- `atreugo/` - Performance + produtividade

### 🔒 **APIs com Segurança**
- `middlewares-basic/` - Auth básico
- `complete/` - Multi-auth enterprise

### 📊 **Observabilidade**
- `hooks-basic/` - Monitoring simples
- `advanced/` - Métricas completas
- `complete/` - APM completo

### 🏭 **Produção Enterprise**
- `advanced/` - Graceful shutdown
- `complete/` - Observabilidade total

## 🚨 Troubleshooting

### Compilação
```bash
# Verificar versão Go
go version  # Requer Go 1.19+

# Limpar módulos
go clean -modcache
go mod tidy
```

### Execução
```bash
# Porta em uso
lsof -i :8080
kill -9 <PID>

# Dependências
./run_examples.sh test
```

### Performance
```bash
# Verificar recursos
htop
netstat -an | grep 8080

# Profile
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 🔗 Próximos Passos

Após dominar estes exemplos:

1. **Implementar seu próprio server**
2. **Adicionar middlewares customizados**
3. **Integrar com bancos de dados**
4. **Implementar distributed tracing**
5. **Adicionar métricas Prometheus**
6. **Deploy em Kubernetes**

---

*Documentação atualizada: Agosto 2025*
*Total de exemplos: 9 (6 frameworks + 3 recursos)*
*Status: ✅ Todos funcionais e documentados*
