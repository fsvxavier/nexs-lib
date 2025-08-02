# Exemplos - Nexs Lib HTTP Server

Esta pasta contém exemplos práticos demonstrando como usar a biblioteca `nexs-lib/httpserver` com diferentes frameworks e funcionalidades.

## 📚 Índice de Exemplos

### 🎯 Exemplos Principais

| Exemplo | Descrição | Nível |
|---------|-----------|-------|
| [**hooks-basic**](./hooks-basic/) | Sistema básico de hooks para monitoramento | Iniciante |
| [**middlewares-basic**](./middlewares-basic/) | Sistema básico de middlewares (auth + logging) | Iniciante |
| [**complete**](./complete/) | Exemplo completo com hooks + middlewares | Avançado |

### � Exemplos por Framework

| Framework | Exemplo | Status |
|-----------|---------|--------|
| [**gin**](./gin/) | Integração com Gin Framework | ✅ Disponível |
| [**echo**](./echo/) | Integração com Echo Framework | ✅ Disponível |
| [**fasthttp**](./fasthttp/) | Integração com FastHTTP | ✅ Disponível |
| [**atreugo**](./atreugo/) | Integração com Atreugo | ✅ Disponível |
| [**basic**](./basic/) | Servidor HTTP nativo Go | ✅ Disponível |
| [**advanced**](./advanced/) | Configurações avançadas | ✅ Disponível |

## 🚀 Início Rápido

### 1. Exemplo Básico com Hooks

Para começar com monitoramento básico:

```bash
cd hooks-basic
go run main.go
```

**O que você vai aprender:**
- Como registrar e usar hooks
- Monitoramento de requisições e respostas
- Rastreamento de erros
- Métricas básicas de servidor

### 2. Exemplo Básico com Middlewares

Para adicionar autenticação e logging:

```bash
cd middlewares-basic
go run main.go
```

**O que você vai aprender:**
- Configuração de middlewares
- Basic Auth para proteção de rotas
- Logging estruturado de requisições
- Gerenciamento de rotas públicas vs protegidas

### 3. Exemplo Completo

Para ver todos os recursos em ação:

```bash
cd complete
go run main.go
```

**O que você vai aprender:**
- Integração completa hooks + middlewares
- Sistema de monitoramento avançado
- API com múltiplos níveis de autenticação
- Métricas detalhadas em tempo real
```bash
curl http://localhost:8080/
```

### 📁 advanced/
Exemplo avançado demonstrando:
- Configuração detalhada com Builder pattern
- Observer customizado para logging
- Middleware de logging
- Múltiplas rotas com parâmetros
- Endpoint de health check
- Graceful shutdown com sinais

**Como executar:**
```bash
cd examples/advanced
go run main.go
```

**Teste:**
```bash
# Rota básica
curl http://localhost:8080/hello

# Rota com parâmetro
curl http://localhost:8080/hello/world

# Health check
curl http://localhost:8080/health
```

## Conceitos Demonstrados

### 1. **Provider Independence**
Ambos os exemplos usam o provider Fiber, mas podem facilmente ser trocados para outros providers (quando implementados) apenas mudando o nome do provider na criação do servidor.

### 2. **Observer Pattern**
O exemplo avançado mostra como implementar um observer customizado para logging de eventos do ciclo de vida do servidor.

### 3. **Middleware**
Demonstra como registrar middleware customizado que funciona nativamente com o provider escolhido.

### 4. **Configuration Builder**
Mostra o uso do padrão Builder para configuração flexível e validada.

### 5. **Type Safety**
Os handlers são tipados especificamente para o provider (fiber.Ctx), garantindo type safety em tempo de compilação.

### 6. **Statistics**
O endpoint /health demonstra como acessar estatísticas em tempo real do servidor.

## Executando os Exemplos

### Pré-requisitos
```bash
go mod tidy
```

### Executar Exemplo Básico
```bash
cd examples/basic
go run main.go
```

### Executar Exemplo Avançado
```bash
cd examples/advanced
go run main.go
```

## Próximos Exemplos (Planejados)

- [ ] **nethttp/** - Exemplo usando provider net/http
- [ ] **gin/** - Exemplo usando provider Gin (quando implementado)
- [ ] **echo/** - Exemplo usando provider Echo (quando implementado)
- [ ] **middleware/** - Exemplo focado em middleware customizado
- [ ] **tls/** - Exemplo com HTTPS/TLS
- [ ] **metrics/** - Exemplo com métricas Prometheus
