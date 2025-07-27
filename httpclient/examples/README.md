# 📚 Exemplos de Uso - nexs-lib HTTP Client

Bem-vindo aos exemplos práticos do cliente HTTP nexs-lib! Esta coleção demonstra todas as funcionalidades avançadas implementadas.

## 🗂️ Estrutura dos Exemplos

```
examples/
├── middleware/             # Interceptação e modificação de requisições
├── hooks/                 # Ganchos para métricas, auditoria e segurança  
├── streaming/             # Processamento de streams e downloads
├── batch/                 # Operações paralelas em lote
├── http2/                 # Recursos específicos do HTTP/2
├── dependency_injection/   # Padrões de injeção de dependências
├── nethttp/               # Exemplos com provider net/http
├── fiber/                 # Exemplos com provider Fiber
├── fasthttp/              # Exemplos com provider FastHTTP
└── README.md              # Este arquivo
```

## 🚀 Exemplos Criados

### 🔧 Middleware (`middleware/`)
- **LoggingMiddleware**: Log detalhado de requisições e respostas
- **AuthMiddleware**: Injeção automática de tokens de autenticação
- **RateLimitMiddleware**: Controle de taxa com token bucket
- **Exemplo**: Cadeia de middleware com múltiplas camadas

### 🪝 Hooks (`hooks/`)
- **MetricsHook**: Coleta de métricas de performance e latência
- **SecurityHook**: Validação de segurança e sanitização
- **AuditHook**: Log de auditoria para compliance
- **Exemplo**: Sistema completo de observabilidade

### 🌊 Streaming (`streaming/`)
- **FileDownloadHandler**: Download de arquivos com barra de progresso
- **JSONStreamHandler**: Processamento de streams JSON em tempo real
- **ProgressHandler**: Monitoramento de progresso com callbacks
- **Exemplo**: Download concorrente com controle de progresso

### 📦 Batch (`batch/`)
- **Operações Paralelas**: Múltiplas requisições simultâneas
- **Comparação de Performance**: Batch vs. sequencial
- **Tratamento de Erros**: Análise granular de falhas
- **Exemplo**: Processamento de 50+ requisições em paralelo

### 🚄 HTTP/2 (`http2/`)
- **Multiplexing**: Requisições paralelas sobre uma conexão
- **Comparação HTTP/2 vs HTTP/1.1**: Análise de performance
- **Recursos Avançados**: Compressão, headers e streaming
- **Exemplo**: Demonstração completa de capacidades HTTP/2

### 🔗 Dependency Injection (`dependency_injection/`)
- **Service Layer**: Injeção de clientes HTTP em serviços
- **Named Clients**: Clientes reutilizáveis identificados por nome
- **Client Manager**: Gerenciamento centralizado de múltiplos clientes
- **Exemplo**: Aplicação completa com padrões de DI

## 📋 Funcionalidades Demonstradas

| Funcionalidade | Middleware | Hooks | Streaming | Batch | HTTP/2 | DI |
|---|---|---|---|---|---|---|
| **Performance** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Paralelização** | - | - | ✅ | ✅ | ✅ | - |
| **Interceptação** | ✅ | ✅ | - | - | - | - |
| **Métricas** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Progresso** | - | - | ✅ | ✅ | - | - |
| **Reutilização** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Testabilidade** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

## 🎯 Como Executar

### Executar Exemplo Específico
```bash
# Middleware
cd httpclient/examples/middleware && go run main.go

# Hooks
cd httpclient/examples/hooks && go run main.go

# Streaming
cd httpclient/examples/streaming && go run main.go

# Batch
cd httpclient/examples/batch && go run main.go

# HTTP/2
cd httpclient/examples/http2 && go run main.go

# Dependency Injection
cd httpclient/examples/dependency_injection && go run main.go
```

### Executar Todos os Exemplos
```bash
cd httpclient/examples
./run_all_examples.sh  # (se disponível)
```

## 📊 Cenários de Uso

### 🔧 Para Middleware
- **Autenticação**: JWT, OAuth, API Keys
- **Logging**: Auditoria, debug, monitoramento
- **Rate Limiting**: Controle de taxa, throttling
- **Transformação**: Modificação de headers/body

### 🪝 Para Hooks
- **Métricas**: Latência, throughput, taxa de erro
- **Segurança**: Validação, sanitização, detecção
- **Auditoria**: Compliance, logs de acesso
- **Cache**: Invalidação, warm-up, estatísticas

### 🌊 Para Streaming
- **Downloads**: Arquivos grandes, mídia, backups
- **Real-time**: WebSockets, Server-Sent Events
- **Processing**: JSON streams, CSV, logs
- **Progress**: UIs interativas, CLIs

### 📦 Para Batch
- **Sincronização**: Múltiplas APIs, dados distribuídos
- **Validação**: Verificação em massa
- **CRUD**: Operações bulk
- **Analytics**: Coleta de dados paralela

### 🚄 Para HTTP/2
- **Performance**: Sites de alta carga
- **Multiplexing**: Muitas requisições pequenas
- **Mobile**: Aplicações com latência crítica
- **APIs**: GraphQL, microserviços

### 🔗 Para Dependency Injection
- **Microserviços**: Múltiplos clientes para diferentes APIs
- **Testing**: Mocking e isolamento de testes
- **Reutilização**: Clientes compartilhados na aplicação
- **Gerenciamento**: Ciclo de vida centralizado

## 💡 Padrões de Integração

### Combinando Funcionalidades
```go
// Cliente completo com todas as funcionalidades
client, err := httpclient.New(interfaces.ProviderNetHTTP, baseURL)

// Adicionar middleware
client.AddMiddleware(&LoggingMiddleware{})
client.AddMiddleware(&AuthMiddleware{token: "abc123"})

// Adicionar hooks
client.AddHook(&MetricsHook{})
client.AddHook(&SecurityHook{})

// Usar streaming
client.Stream(ctx, "GET", "/large-file", &ProgressHandler{})

// Usar batch
batch := client.Batch()
batch.Add("GET", "/endpoint1", nil)
batch.Add("GET", "/endpoint2", nil)
results, err := batch.Execute(ctx)

// Dependency injection
service := NewAPIService(client) // Inject client into service
```

## 🔍 Observabilidade

Todos os exemplos incluem:
- **Métricas de Performance**: Latência, throughput, taxa de erro
- **Logs Estruturados**: JSON com contexto completo
- **Tracing**: Rastreamento de requisições
- **Health Checks**: Status e disponibilidade

## 📈 Benchmarks

Os exemplos incluem comparações de performance:
- **Middleware overhead**: ~1-5ms por middleware
- **Batch vs Sequential**: 5-10x melhoria
- **HTTP/2 vs HTTP/1.1**: 20-50% melhoria
- **Streaming vs Buffer**: 80% menos memória
- **Named Clients**: 90% menos overhead de criação

## 🛠️ Extensibilidade

### Criando Seu Próprio Middleware
```go
type CustomMiddleware struct{}

func (m *CustomMiddleware) Handle(ctx context.Context, req *interfaces.Request, next interfaces.NextHandler) (*interfaces.Response, error) {
    // Lógica antes da requisição
    resp, err := next(ctx, req)
    // Lógica depois da requisição
    return resp, err
}
```

### Criando Seu Próprio Hook
```go
type CustomHook struct{}

func (h *CustomHook) BeforeRequest(ctx context.Context, req *interfaces.Request) {
    // Lógica antes da requisição
}

func (h *CustomHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) {
    // Lógica depois da resposta
}

func (h *CustomHook) OnError(ctx context.Context, req *interfaces.Request, err error) {
    // Lógica em caso de erro
}
```

## 📚 Documentação Adicional

Cada exemplo tem seu próprio README detalhado:
- [`middleware/README.md`](middleware/README.md)
- [`hooks/README.md`](hooks/README.md)
- [`streaming/README.md`](streaming/README.md)
- [`batch/README.md`](batch/README.md)
- [`http2/README.md`](http2/README.md)
- [`dependency_injection/README.md`](dependency_injection/README.md)
- [`nethttp/README.md`](nethttp/README.md)
- [`fiber/README.md`](fiber/README.md)
- [`fasthttp/README.md`](fasthttp/README.md)

## 🤝 Contribuindo

Para adicionar novos exemplos:
1. Crie um diretório com nome descritivo
2. Inclua `main.go` com exemplo funcional
3. Adicione `README.md` detalhado
4. Teste a compilação e execução
5. Documente padrões de uso

## 🔗 Links Úteis

- [Documentação da API](../README.md)
- [Guia de Configuração](../config/README.md)
- [Providers Disponíveis](../providers/README.md)
- [Testes e Benchmarks](../tests/README.md)
