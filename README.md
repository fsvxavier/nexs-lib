# NEXS-LIB 🚀

[![Go Version](https://img.shields.io/badge/go-1.23.3+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-lib)](https://goreportcard.com/report/github.com/fsvxavier/nexs-lib)
[![codecov](https://codecov.io/gh/fsvxavier/nexs-lib/branch/main/graph/badge.svg)](https://codecov.io/gh/fsvxavier/nexs-lib)
[![Documentation](https://img.shields.io/badge/docs-pkg.go.dev-blue)](https://pkg.go.dev/github.com/fsvxavier/nexs-lib)
[![Release](https://img.shields.io/github/release/fsvxavier/nexs-lib.svg)](https://github.com/fsvxavier/nexs-lib/releases)
[![Go.Dev Reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/fsvxavier/nexs-lib)

**NEXS-LIB** é uma biblioteca Go moderna e abrangente que fornece implementações unificadas e abstrações para ferramentas comuns de desenvolvimento. Ela oferece interfaces consistentes para diferentes providers e frameworks, permitindo que você troque implementações facilmente sem alterar sua lógica de negócio.

## 🚀 Performance & Compatibilidade

- **Go Version**: 1.23.3+ (suporte completo a generics)
- **Zero Allocations**: Otimizado para aplicações de alta performance
- **Thread-Safe**: Todas as implementações são seguras para uso concorrente
- **Memory Efficient**: Baixo footprint de memória
- **Production Ready**: Usado em produção por aplicações críticas

## 🎯 Filosofia

Esta biblioteca foi projetada seguindo os princípios de:

- **Interface Segregation**: Interfaces pequenas e específicas
- **Dependency Inversion**: Dependa de abstrações, não de implementações concretas
- **Provider Pattern**: Múltiplas implementações através de uma interface comum
- **Factory Pattern**: Criação simplificada de instâncias
- **Domain-Driven Design**: Separação clara entre domínio e infraestrutura

## 📦 Módulos Disponíveis

### 🔢 Decimal
Trabalhe com números decimais de alta precisão usando diferentes providers.

**Providers Suportados:**
- `github.com/shopspring/decimal` (padrão)
- `github.com/cockroachdb/apd/v3`

```go
import "github.com/fsvxavier/nexs-lib/decimal"

// Usando ShopSpring (padrão)
provider := decimal.NewProvider(decimal.ShopSpring)
dec := provider.FromString("123.456")

// Usando APD para alta precisão
provider := decimal.NewProvider(decimal.APD)
dec := provider.FromString("999999999999.123456789")
```

### 🌐 HTTP Servers
Abstrações unificadas para diferentes frameworks web Go.

**Frameworks Suportados:**
- Fiber
- Echo
- Gin
- net/http
- FastHTTP
- Atreugo

```go
import "github.com/fsvxavier/nexs-lib/httpservers"

// Interface comum para todos os frameworks
server := httpservers.NewFiberServer(config)
server.Start(":8080")
```

### 📡 HTTP Requester
Cliente HTTP unificado com suporte a diferentes implementações.

**Clientes Suportados:**
- Resty
- Fiber Client
- net/http

```go
import "github.com/fsvxavier/nexs-lib/httprequester"

client := httprequester.NewRestyClient()
response, err := client.Get("https://api.example.com/data")
```

### 📊 JSON
Interface unificada para diferentes bibliotecas JSON com foco em performance.

**Providers Suportados:**
- `encoding/json` (stdlib)
- `github.com/json-iterator/go`
- `github.com/goccy/go-json`
- `github.com/buger/jsonparser`

```go
import "github.com/fsvxavier/nexs-lib/json"

// Troque entre providers facilmente
provider := json.NewProvider(json.GoCCY) // Alta performance
data, err := provider.Marshal(object)
```

### 📄 Paginação
Sistema completo de paginação para APIs REST e consultas de banco.

**Frameworks Suportados:**
- Fiber
- Echo
- Gin
- net/http
- Atreugo
- FastHTTP

```go
import "github.com/fsvxavier/nexs-lib/paginate"

// Parse automático de parâmetros HTTP
page := paginate.ParseFiberRequest(c)

// Paginação de slice em memória
result := paginate.PaginateSlice(data, page)

// Integração com banco de dados
query := paginate.BuildSQLQuery(baseQuery, page)
```

### 🔍 Parsers
Biblioteca de parsing moderna com compatibilidade 100% com bibliotecas legadas.

**Parsers Disponíveis:**
- DateTime (compatível com dateparse)
- Duration
- Environment Variables

```go
import "github.com/fsvxavier/nexs-lib/parsers/datetime"

// 100% compatível com bibliotecas legadas
date, err := datetime.ParseAny("02/03/2023")
date, err := datetime.ParseIn("2023-01-15 10:30", loc)
date := datetime.MustParseAny("2023-01-15T10:30:45Z")
```

### 🧵 String Utilities (strutl)
Utilitários avançados para manipulação de strings com suporte completo a Unicode.

```go
import "github.com/fsvxavier/nexs-lib/strutl"

// Conversão de casos
camelCase := strutl.ToCamelCase("hello_world")
snakeCase := strutl.ToSnakeCase("HelloWorld")

// Alinhamento e padding
aligned := strutl.PadCenter("texto", 20, " ")

// WordWrap e formatação
wrapped := strutl.WordWrap("texto longo", 50)
```

### 🆔 UID
Geração e manipulação de identificadores únicos (UUID/ULID).

```go
import "github.com/fsvxavier/nexs-lib/uid"

// ULID - Lexicographically sortable
ulid := uid.NewULID()

// UUID - Standard UUID
uuid := uid.NewUUID()

// Conversões entre formatos
converted := uid.ULIDToUUID(ulid)
```

### ✅ Validator
Sistema robusto de validação com suporte a JSON Schema.

```go
import "github.com/fsvxavier/nexs-lib/validator/schema"

validator := schema.NewJSONSchemaValidator()
result := validator.ValidateSchema(ctx, data, jsonSchema)

if !result.Valid {
    // Tratar erros de validação
    for field, errors := range result.Errors {
        log.Printf("Campo %s: %v", field, errors)
    }
}
```

### 🗄️ Database
Abstrações para diferentes bancos de dados e ORMs.

**Suporte a:**
- PostgreSQL (pgx, pq, GORM)
- MongoDB
- Redis
- Valkey

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql"

// Diferentes providers PostgreSQL
db := postgresql.NewPgxConnection(config)
db := postgresql.NewGormConnection(config)
db := postgresql.NewPqConnection(config)

// Interface unificada
rows, err := db.Query(ctx, "SELECT * FROM users WHERE active = $1", true)
```

### 📨 Message Queue
Sistema completo de filas de mensagem com múltiplos providers e retry avançado.

**Providers Suportados:**
- RabbitMQ
- Apache Kafka
- AWS SQS
- Apache ActiveMQ (STOMP)

```go
import "github.com/fsvxavier/nexs-lib/message-queue"

// Producer - enviar mensagens
producer := messagequeue.NewProducer(messagequeue.RabbitMQ, config)
producer.Publish(ctx, topic, message)

// Consumer - processar mensagens
consumer := messagequeue.NewConsumer(messagequeue.Kafka, config)
consumer.Subscribe(ctx, topic, handler)
```

### 📊 Observability
Sistema completo de observabilidade com logging e tracing distribuído.

#### Logger
Sistema de logging estruturado com múltiplos providers.

**Providers Suportados:**
- Zap (uber-go/zap)
- Zerolog (rs/zerolog)
- Slog (Go stdlib)

```go
import "github.com/fsvxavier/nexs-lib/observability/logger"

// Configuração flexível
logger := logger.NewLogger(logger.ZapProvider, config)
logger.Info(ctx, "Operação realizada", 
    logger.String("user_id", "123"),
    logger.Int("count", 42))
```

#### Tracer
Sistema de tracing distribuído seguindo padrões OpenTelemetry.

**Providers Suportados:**
- Datadog APM
- New Relic APM
- Prometheus/Grafana

```go
import "github.com/fsvxavier/nexs-lib/observability/tracer"

// Configurar tracer
tracer := tracer.NewTracer(tracer.DatadogProvider, config)

// Criar spans
span := tracer.StartSpan(ctx, "operacao-importante")
defer span.Finish()

// Adicionar atributos
span.SetTag("user.id", "123")
span.SetTag("operation.type", "payment")
```

### ❌ Domain Errors
Sistema estruturado de tratamento de erros seguindo DDD.

```go
import "github.com/fsvxavier/nexs-lib/domainerrors"

// Diferentes tipos de erros de domínio
err := domainerrors.NewValidationError("Campo obrigatório")
err := domainerrors.NewBusinessRuleError("Regra de negócio violada")
err := domainerrors.NewNotFoundError("Recurso não encontrado")

// Mapeamento automático para códigos HTTP
httpCode := err.HTTPStatusCode()
```

## 🚀 Instalação

```bash
go get github.com/fsvxavier/nexs-lib
```

Ou instale módulos específicos:

```bash
# Apenas decimal
go get github.com/fsvxavier/nexs-lib/decimal

# Apenas HTTP servers
go get github.com/fsvxavier/nexs-lib/httpservers

# Apenas parsers
go get github.com/fsvxavier/nexs-lib/parsers

# Apenas message queue
go get github.com/fsvxavier/nexs-lib/message-queue

# Apenas observability
go get github.com/fsvxavier/nexs-lib/observability/logger
go get github.com/fsvxavier/nexs-lib/observability/tracer
```

## 📚 Exemplos de Uso

### Exemplo Completo: API REST com Paginação

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/paginate"
    "github.com/fsvxavier/nexs-lib/json"
    "github.com/gofiber/fiber/v2"
)

func main() {
    // Configurar servidor
    app := fiber.New()
    
    // Configurar JSON provider de alta performance
    jsonProvider := json.NewProvider(json.GoCCY)
    
    app.Get("/users", func(c *fiber.Ctx) error {
        // Parse automático de parâmetros de paginação
        page := paginate.ParseFiberRequest(c)
        
        // Simular dados
        users := getUsersFromDB()
        
        // Aplicar paginação
        result := paginate.PaginateSlice(users, page)
        
        // Serializar resposta
        response, _ := jsonProvider.Marshal(result)
        
        c.Set("Content-Type", "application/json")
        return c.Send(response)
    })
    
    app.Listen(":3000")
}
```

### Exemplo: Message Queue com Retry

```go
package main

import (
    "context"
    "time"
    "github.com/fsvxavier/nexs-lib/message-queue"
    "github.com/fsvxavier/nexs-lib/observability/logger"
)

func main() {
    // Configurar logger
    log := logger.NewLogger(logger.ZapProvider, &logger.Config{
        Level: "info",
        Format: "json",
    })
    
    // Configurar producer
    producer := messagequeue.NewProducer(messagequeue.RabbitMQ, &messagequeue.Config{
        URL: "amqp://localhost:5672",
        Exchange: "events",
    })
    
    // Enviar mensagem com retry
    message := &messagequeue.Message{
        ID: "msg-001",
        Body: []byte(`{"event": "user_created", "user_id": "123"}`),
        Headers: map[string]string{
            "content-type": "application/json",
            "source": "user-service",
        },
        RetryPolicy: &messagequeue.RetryPolicy{
            MaxRetries: 3,
            BackoffType: messagequeue.ExponentialBackoff,
            InitialInterval: time.Second,
        },
    }
    
    err := producer.Publish(context.Background(), "user.events", message)
    if err != nil {
        log.Error(context.Background(), "Falha ao enviar mensagem", 
            logger.Error(err),
            logger.String("message_id", message.ID))
    }
}
```

### Exemplo: Observabilidade Completa

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/logger"
    "github.com/fsvxavier/nexs-lib/observability/tracer"
    "github.com/fsvxavier/nexs-lib/httprequester"
)

func processPayment(ctx context.Context, userID string, amount float64) error {
    // Configurar tracer
    tracer := tracer.NewTracer(tracer.DatadogProvider, &tracer.Config{
        ServiceName: "payment-service",
        Environment: "production",
    })
    
    // Configurar logger
    log := logger.NewLogger(logger.ZapProvider, &logger.Config{
        Level: "info",
        Format: "json",
    })
    
    // Criar span para a operação
    span := tracer.StartSpan(ctx, "process-payment")
    defer span.Finish()
    
    // Adicionar tags ao span
    span.SetTag("user.id", userID)
    span.SetTag("payment.amount", amount)
    span.SetTag("payment.currency", "BRL")
    
    // Log estruturado com contexto
    log.Info(ctx, "Iniciando processamento de pagamento",
        logger.String("user_id", userID),
        logger.Float64("amount", amount),
        logger.String("trace_id", span.TraceID()),
        logger.String("span_id", span.SpanID()))
    
    // Fazer chamada HTTP com tracing
    client := httprequester.NewRestyClient()
    response, err := client.Get("https://api.payment.com/validate")
    
    if err != nil {
        span.SetTag("error", true)
        span.SetTag("error.message", err.Error())
        log.Error(ctx, "Falha na validação do pagamento",
            logger.Error(err),
            logger.String("user_id", userID))
        return err
    }
    
    log.Info(ctx, "Pagamento processado com sucesso",
        logger.String("user_id", userID),
        logger.Int("status_code", response.StatusCode()))
    
    return nil
}
```

## 🧪 Testes

Execute todos os testes:

```bash
go test ./...
```

Execute testes com coverage:

```bash
go test -cover ./...
```

Execute testes específicos por módulo:

```bash
# Testar apenas decimal
go test ./decimal/...

# Testar apenas message-queue
go test ./message-queue/...

# Testar apenas observability
go test ./observability/...
```

Execute testes com race detection:

```bash
go test -race ./...
```

Execute benchmarks:

```bash
go test -bench=. ./...
```

## 🏗️ Arquitetura

```
nexs-lib/
├── decimal/            # Providers para números decimais
├── db/                # Abstrações para bancos de dados
├── domainerrors/      # Sistema estruturado de erros
├── httprequester/     # Clientes HTTP unificados
├── httpservers/       # Servidores HTTP abstraídos
├── json/              # Providers JSON de alta performance
├── message-queue/     # Sistema de filas de mensagem
├── observability/     # Sistema de observabilidade
│   ├── logger/        # Logging estruturado
│   └── tracer/        # Tracing distribuído
├── paginate/          # Sistema completo de paginação
├── parsers/           # Parsers modernos com compatibilidade legada
├── strutl/            # Utilitários avançados de string
├── uid/               # Geração de identificadores únicos
└── validator/         # Sistema robusto de validação
```

Cada módulo segue o padrão:
- `interfaces/` - Definições de interfaces
- `providers/` - Implementações específicas
- `examples/` - Exemplos de uso
- `*_test.go` - Testes unitários
- `README.md` - Documentação específica

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor:

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Padrões de Desenvolvimento

- **Testes**: Toda funcionalidade deve ter testes unitários
- **Documentação**: Funções públicas devem ter documentação GoDoc
- **Interfaces**: Prefira interfaces pequenas e específicas
- **Erros**: Use o sistema de domain errors da biblioteca
- **Compatibilidade**: Mantenha compatibilidade com versões anteriores

## 📋 Roadmap

### 🚀 Em Desenvolvimento
- [x] **Message Queue**: Sistema completo de filas (RabbitMQ, Kafka, SQS, ActiveMQ)
- [x] **Observability**: Logging estruturado e tracing distribuído
- [x] **Database**: Abstrações para PostgreSQL com múltiplos drivers

### 🎯 Próximas Versões
- [ ] **v2.0.0**: Suporte a Go Generics
- [ ] **Cache Module**: Abstrações para Redis, Memcached, etc.
- [ ] **Config**: Sistema unificado de configuração
- [ ] **Metrics**: Integração com Prometheus, DataDog
- [ ] **NoSQL Database**: Suporte a MongoDB, DynamoDB
- [ ] **Event Sourcing**: Padrões de Event Sourcing e CQRS

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Agradecimentos

Agradecimentos especiais a todas as bibliotecas open source que inspiraram e fundamentaram este projeto:

### Decimal & Numeric
- [shopspring/decimal](https://github.com/shopspring/decimal)
- [cockroachdb/apd](https://github.com/cockroachdb/apd)

### HTTP Frameworks
- [gofiber/fiber](https://github.com/gofiber/fiber)
- [labstack/echo](https://github.com/labstack/echo)
- [gin-gonic/gin](https://github.com/gin-gonic/gin)
- [valyala/fasthttp](https://github.com/valyala/fasthttp)
- [savsgio/atreugo](https://github.com/savsgio/atreugo)

### JSON Libraries
- [json-iterator/go](https://github.com/json-iterator/go)
- [goccy/go-json](https://github.com/goccy/go-json)
- [buger/jsonparser](https://github.com/buger/jsonparser)

### Message Queue
- [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
- [IBM/sarama](https://github.com/IBM/sarama) (Kafka)
- [go-stomp/stomp](https://github.com/go-stomp/stomp) (ActiveMQ)

### Database
- [jackc/pgx](https://github.com/jackc/pgx)
- [lib/pq](https://github.com/lib/pq)
- [gorm.io/gorm](https://gorm.io/)

### Observability
- [uber-go/zap](https://github.com/uber-go/zap)
- [rs/zerolog](https://github.com/rs/zerolog)
- [DataDog/dd-trace-go](https://github.com/DataDog/dd-trace-go)
- [newrelic/go-agent](https://github.com/newrelic/go-agent)

### Utilities
- [google/uuid](https://github.com/google/uuid)
- [oklog/ulid](https://github.com/oklog/ulid)
- [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema)

---

**NEXS-LIB** - Construindo aplicações Go modernas com abstrações sólidas e interfaces unificadas.

Para mais detalhes, consulte a documentação específica de cada módulo nos respectivos diretórios.
