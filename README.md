# NEXS-LIB 🚀

[![Go Version](https://img.shields.io/badge/go-1.23.3+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-lib)](https://goreportcard.com/report/github.com/fsvxavier/nexs-lib)
[![codecov](https://codecov.io/gh/fsvxavier/nexs-lib/branch/main/graph/badge.svg)](https://codecov.io/gh/fsvxavier/nexs-lib)

**NEXS-LIB** é uma biblioteca Go moderna e abrangente que fornece implementações unificadas e abstrações para ferramentas comuns de desenvolvimento. Ela oferece interfaces consistentes para diferentes providers e frameworks, permitindo que você troque implementações facilmente sem alterar sua lógica de negócio.

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
- DynamoDB
- Redis
- Valkey

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

### Exemplo: Validação com Schema JSON

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/validator/schema"
    "github.com/fsvxavier/nexs-lib/json"
)

func validateUser(userData map[string]interface{}) error {
    // Schema JSON para validação
    userSchema := `{
        "type": "object",
        "required": ["name", "email", "age"],
        "properties": {
            "name": {"type": "string", "minLength": 2},
            "email": {"type": "string", "format": "email"},
            "age": {"type": "integer", "minimum": 0, "maximum": 120}
        }
    }`
    
    // Criar validator
    validator := schema.NewJSONSchemaValidator()
    
    // Validar dados
    result := validator.ValidateSchema(context.Background(), userData, userSchema)
    
    if !result.Valid {
        return schema.NewSchemaValidationError(result)
    }
    
    return nil
}
```

## 🧪 Testes

Execute todos os testes:

```bash
make test
```

Execute testes com coverage:

```bash
make cover
```

Execute testes com HTML coverage:

```bash
make cover-html-open
```

## 🏗️ Arquitetura

```
nexs-lib/
├── decimal/         # Providers para números decimais
├── db/             # Abstrações para bancos de dados
├── domainerrors/   # Sistema estruturado de erros
├── httprequester/  # Clientes HTTP unificados
├── httpservers/    # Servidores HTTP abstraídos
├── json/           # Providers JSON de alta performance
├── paginate/       # Sistema completo de paginação
├── parsers/        # Parsers modernos com compatibilidade legada
├── strutl/         # Utilitários avançados de string
├── uid/            # Geração de identificadores únicos
└── validator/      # Sistema robusto de validação
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

- [ ] **v2.0.0**: Suporte a Go Generics
- [ ] **Cache Module**: Abstrações para Redis, Memcached, etc.
- [ ] **Message Queue**: Suporte a RabbitMQ, Kafka, etc.
- [ ] **Metrics**: Integração com Prometheus, DataDog
- [ ] **Tracing**: Suporte a OpenTelemetry
- [ ] **Config**: Sistema unificado de configuração

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Agradecimentos

Agradecimentos especiais a todas as bibliotecas open source que inspiraram e fundamentaram este projeto:

- [shopspring/decimal](https://github.com/shopspring/decimal)
- [cockroachdb/apd](https://github.com/cockroachdb/apd)
- [gofiber/fiber](https://github.com/gofiber/fiber)
- [labstack/echo](https://github.com/labstack/echo)
- [gin-gonic/gin](https://github.com/gin-gonic/gin)
- [json-iterator/go](https://github.com/json-iterator/go)
- [goccy/go-json](https://github.com/goccy/go-json)

---

**NEXS-LIB** - Construindo aplicações Go modernas com abstrações sólidas e interfaces unificadas.

Para mais detalhes, consulte a documentação específica de cada módulo nos respectivos diretórios.
