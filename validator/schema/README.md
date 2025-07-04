# Schema Validator - Nexs-Lib

Uma biblioteca moderna e extensível de validação de esquemas JSON para Go, integrada ao ecossistema nexs-lib e seguindo as melhores práticas de desenvolvimento.

## 🚀 Características

- **JSON Schema Validation**: Validação completa baseada no padrão JSON Schema
- **Formatos Customizados**: Suporte extensivo a formatos de validação personalizados
- **Interface Fluente**: APIs builders para criação intuitiva de regras de validação
- **Validação de Struct**: Suporte a tags de validação em structs Go
- **Context Support**: Suporte completo a context.Context para timeouts e cancelamentos
- **Thread-Safe**: Seguro para uso concorrente
- **Integração com Domain Errors**: Integrado com o sistema de erros do nexs-lib
- **Validadores de Formato Especializados**: Datas, decimais, strings, números JSON, etc.
- **Performance**: Otimizado para alta performance com cache e reutilização

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/validator/schema
```

## 🛠️ Uso Básico

### Validação JSON Schema

```go
package main

import (
    "context"
    "fmt"
    "github.com/fsvxavier/nexs-lib/validator/schema"
)

func main() {
    ctx := context.Background()
    
    // Criar um validador de JSON Schema
    validator := schema.NewJSONSchemaValidator()
    
    // Definir o schema
    schemaStr := `{
        "type": "object",
        "properties": {
            "name": {"type": "string", "minLength": 1},
            "email": {"type": "string", "format": "email"},
            "age": {"type": "integer", "minimum": 0}
        },
        "required": ["name", "email"]
    }`
    
    // Dados para validar
    data := map[string]interface{}{
        "name":  "João Silva",
        "email": "joao@example.com",
        "age":   30,
    }
    
    // Validar
    result := validator.ValidateSchema(ctx, data, schemaStr)
    if !result.Valid {
        fmt.Printf("Validação falhou: %s\n", result.String())
    } else {
        fmt.Println("Dados válidos!")
    }
}
```

### Validação de Struct com Tags

```go
type User struct {
    ID       string `validate:"required,uuid"`
    Name     string `validate:"required,min=2,max=100"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"min=18,max=120"`
    Website  string `validate:"url"`
}

func validateUser() {
    ctx := context.Background()
    v := schema.NewValidator()
    
    user := User{
        ID:      "550e8400-e29b-41d4-a716-446655440000",
        Name:    "John Doe",
        Email:   "john@example.com",
        Age:     30,
        Website: "https://johndoe.com",
    }
    
    result := v.ValidateStruct(ctx, user)
    if !result.Valid {
        for field, errors := range result.Errors {
            for _, err := range errors {
                fmt.Printf("%s: %s\n", field, err)
            }
        }
    }
}
```

### API Fluente com Builders

```go
func fluentValidation() {
    ctx := context.Background()
    
    // Validação de string complexa
    rule := schema.NewRuleBuilder().
        Required().
        String().
        MinLength(5).
        MaxLength(50).
        Email().
        Build()
    
    if err := rule.Validate(ctx, "test@example.com"); err != nil {
        fmt.Printf("Validação falhou: %s\n", err)
    }
    
    // Validação de número com range
    numberRule := schema.NewRuleBuilder().
        Required().
        Number().
        Range(18, 65).
        Integer().
        Build()
    
    if err := numberRule.Validate(ctx, 25); err != nil {
        fmt.Printf("Número inválido: %s\n", err)
    }
    
    // Validação de data com range temporal
    dateRule := schema.NewRuleBuilder().
        Required().
        DateTime().
        RFC3339().
        After("2020-01-01T00:00:00Z").
        Before("2030-01-01T00:00:00Z").
        Build()
    
    if err := dateRule.Validate(ctx, "2025-07-04T12:00:00Z"); err != nil {
        fmt.Printf("Data inválida: %s\n", err)
    }
}
```

### Validação com JSON Schema

```go
func jsonSchemaValidation() {
    ctx := context.Background()
    schemaValidator := schema.NewJSONSchemaValidator()
    
    schemaStr := `{
        "type": "object",
        "properties": {
            "name": {
                "type": "string",
                "minLength": 1,
                "maxLength": 100
            },
            "email": {
                "type": "string",
                "format": "email"
            },
            "age": {
                "type": "integer",
                "minimum": 18,
                "maximum": 120
            }
        },
        "required": ["name", "email"]
    }`
    
    data := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }
    
    result := schemaValidator.ValidateSchema(ctx, data, schemaStr)
    if !result.Valid {
        fmt.Printf("Validação falhou: %s\n", result.String())
    }
}
```

## 🎯 Formatos de Validação Disponíveis

### Formatos Padrão

A biblioteca vem com diversos formatos de validação pré-configurados:

- **`date_time`**: Formatos de data/hora diversos
- **`iso_8601_date`**: Data no formato ISO 8601
- **`text_match`**: Texto apenas com letras, underscore e espaços
- **`text_match_with_number`**: Texto com letras, números, underscore e espaços  
- **`strong_name`**: Nome forte (identificador válido)
- **`json_number`**: Número JSON válido
- **`decimal`**: Decimal genérico
- **`decimal_by_factor_of_8`**: Decimal com fator de 8
- **`empty_string`**: String vazia

### Validações de Formato em JSON Schema

```go
schemaStr := `{
    "type": "object",
    "properties": {
        "birthDate": {"type": "string", "format": "iso_8601_date"},
        "fullName": {"type": "string", "format": "strong_name"},
        "price": {"type": "string", "format": "decimal"},
        "quantity": {"type": "string", "format": "json_number"}
    }
}`

### Adicionando Formatos Customizados

```go
// Usando regex diretamente
schema.AddCustomFormatByRegex("phone", `^\+\d{1,3}\d{10}$`)

// Usando função customizada
schemaValidator := schema.NewJSONSchemaValidator()
schemaValidator.RegisterFormatValidator("credit-card", func(input interface{}) bool {
    if str, ok := input.(string); ok {
        // Lógica de validação do cartão de crédito
        return isValidCreditCard(str)
    }
    return false
})

// Usando um FormatValidator customizado
type CPFValidator struct{}

func (CPFValidator) IsFormat(input interface{}) bool {
    cpf, ok := input.(string)
    if !ok {
        return false
    }
    return isValidCPF(cpf)
}

func (CPFValidator) FormatName() string {
    return "cpf"
}

schemaValidator.RegisterFormatValidator("cpf", &CPFValidator{})
```

## 🏗️ Validação Customizada

### Validadores Customizados de Interface

```go
// Implementar a interface FormatValidator
type EmailDomainValidator struct {
    allowedDomains []string
}

func (v *EmailDomainValidator) IsFormat(input interface{}) bool {
    email, ok := input.(string)
    if !ok {
        return false
    }
    
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return false
    }
    
    domain := parts[1]
    for _, allowed := range v.allowedDomains {
        if domain == allowed {
            return true
        }
    }
    return false
}

func (v *EmailDomainValidator) FormatName() string {
    return "corporate_email"
}

// Usar o validador
validator := schema.NewJSONSchemaValidator()
validator.RegisterFormatValidator("corporate_email", &EmailDomainValidator{
    allowedDomains: []string{"company.com", "corp.com"},
})
```

### Validator Personalizado com Rules

```go
// Criar um validator customizado com múltiplas regras
customValidator := schema.NewValidator().
    AddRule(schema.NewRequiredRule()).
    AddFieldRule("name", schema.NewMinLengthRule(2)).
    AddFieldRule("email", schema.NewEmailRule())

result := customValidator.ValidateStruct(ctx, user)
```

## 📊 Tratamento de Erros

### ValidationResult

```go
type ValidationResult struct {
    Valid        bool
    Errors       map[string][]string
    GlobalErrors []string
    Warnings     map[string][]string
}

// Métodos úteis
result.Valid              // bool - se a validação passou
result.HasErrors()        // bool - se há erros
result.ErrorCount()       // int - total de erros
result.FirstError()       // string - primeiro erro encontrado
result.AllErrors()        // []string - todos os erros
result.FieldErrors(field) // []string - erros de um campo específico
result.String()           // string - representação formatada de todos os erros
```

### Integration com Domain Errors

```go
// O validator pode retornar erros de domínio específicos
if jsv, ok := schemaValidator.(*schema.JSONSchemaValidator); ok {
    if err := jsv.ValidateWithDomainError(ctx, data, schemaStr); err != nil {
        // Retorna um errordomain.InvalidSchemaError
        return err
    }
}
```

## ⚡ Performance e Context

### Context Support

```go
// Timeout para validação
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result := schemaValidator.ValidateSchema(ctx, data, schemaStr)

// Cancelamento manual
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(time.Second)
    cancel() // Cancela a validação se demorar muito
}()

result := validator.Validate(ctx, data)
```

### Reutilização de Validators

```go
// ✅ Bom - reutilize validators (thread-safe)
var (
    userSchemaValidator    = schema.NewJSONSchemaValidator()
    productSchemaValidator = schema.NewJSONSchemaValidator()
)

func init() {
    // Configurar validators uma vez
    userSchemaValidator.RegisterFormatValidator("cpf", &CPFValidator{})
    productSchemaValidator.RegisterFormatValidator("sku", &SKUValidator{})
}

// ❌ Evite - criar validator a cada uso
func validateUser(user User) {
    v := schema.NewJSONSchemaValidator() // Custoso
    // ...
}
```

## 🧪 Testes

Execute os testes:

```bash
cd validator/schema
go test -v ./...
```

Execute os benchmarks:

```bash
go test -bench=. -benchmem ./...
```

Execute os testes com coverage:

```bash
go test -cover ./...
```

## 📈 Migração e Integração

### Usando com HTTP Servers

```go
// Exemplo com Gin
func createUserHandler(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    validator := schema.NewJSONSchemaValidator()
    result := validator.ValidateSchema(c.Request.Context(), user, userSchema)
    
    if !result.Valid {
        c.JSON(400, gin.H{
            "error": "validation failed",
            "details": result.Errors,
        })
        return
    }
    
    // Processar usuário válido...
    c.JSON(201, user)
}
```

### Integração com Outras Bibliotecas

```go
// Com go-playground/validator (migração)
type User struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

// Migração para nexs-lib/validator/schema
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

const userSchema = `{
    "type": "object",
    "properties": {
        "name": {"type": "string", "minLength": 2},
        "email": {"type": "string", "format": "email"}
    },
    "required": ["name", "email"]
}`

```

## 🎯 Melhores Práticas

### 1. Use Context para Timeouts

```go
// ✅ Sempre use context com timeout para validações
ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
defer cancel()

result := validator.ValidateSchema(ctx, data, schema)
```

### 2. Reutilize Validators (Thread-Safe)

```go
// ✅ Bom - validators são thread-safe, reutilize-os
var userValidator = schema.NewJSONSchemaValidator()

func init() {
    userValidator.RegisterFormatValidator("cpf", &CPFValidator{})
}

func validateUser(user User) *ValidationResult {
    return userValidator.ValidateSchema(context.Background(), user, userSchema)
}

// ❌ Evite - criar validator a cada validação é custoso
func validateUser(user User) *ValidationResult {
    v := schema.NewJSONSchemaValidator() // Custoso!
    return v.ValidateSchema(context.Background(), user, userSchema)
}
```

### 3. Use Schemas JSON Bem Estruturados

```go
// ✅ Bom - schema bem estruturado e documentado
const userSchema = `{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "User",
    "description": "Esquema de validação para usuários do sistema",
    "type": "object",
    "properties": {
        "name": {
            "type": "string",
            "minLength": 2,
            "maxLength": 100,
            "description": "Nome completo do usuário"
        },
        "email": {
            "type": "string",
            "format": "email",
            "description": "Email único do usuário"
        }
    },
    "required": ["name", "email"],
    "additionalProperties": false
}`
```

### 4. Trate Erros de Forma Granular

```go
result := validator.ValidateSchema(ctx, data, schema)
if !result.Valid {
    // Log estruturado de erros
    for field, errors := range result.Errors {
        for _, err := range errors {
            log.Printf("Campo %s: %s", field, err)
        }
    }
    
    // Retorno estruturado para APIs
    return &ValidationResponse{
        Success: false,
        Errors:  result.Errors,
        Message: "Dados de entrada inválidos",
    }
}
```

### 5. Valide Entradas de API Consistentemente

```go
type APIValidator struct {
    validator schema.SchemaValidator
    schemas   map[string]string
}

func (av *APIValidator) ValidateRequest(ctx context.Context, endpoint string, data interface{}) error {
    schemaStr, exists := av.schemas[endpoint]
    if !exists {
        return fmt.Errorf("schema não encontrado para endpoint: %s", endpoint)
    }
    
    result := av.validator.ValidateSchema(ctx, data, schemaStr)
    if !result.Valid {
        return &ValidationError{
            Message: "Dados inválidos",
            Details: result.Errors,
        }
    }
    
    return nil
}
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Diretrizes de Contribuição

- Mantenha a cobertura de testes acima de 80%
- Implemente testes para novos formatos de validação
- Documente todas as interfaces públicas
- Siga as convenções de código Go
- Adicione exemplos para novas funcionalidades

## 📜 Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## � Arquitetura e Design

### Interfaces Principais

```go
// SchemaValidator - Interface principal para validação de schemas
type SchemaValidator interface {
    ValidateSchema(ctx context.Context, data interface{}, schema string) *ValidationResult
    RegisterFormatValidator(name string, validator FormatValidator) 
    ValidateWithDomainError(ctx context.Context, data interface{}, schema string) error
}

// FormatValidator - Interface para validadores de formato customizados
type FormatValidator interface {
    IsFormat(input interface{}) bool
    FormatName() string
}

// ValidationResult - Resultado detalhado da validação
type ValidationResult struct {
    Valid        bool
    Errors       map[string][]string
    GlobalErrors []string
    Warnings     map[string][]string
}
```

### Componentes

- **JSONSchemaValidator**: Implementação principal usando gojsonschema
- **Format Validators**: Validadores especializados para formatos específicos
- **Domain Error Integration**: Integração com sistema de erros do nexs-lib
- **Context Support**: Suporte completo a context.Context

## 📚 Exemplos Adicionais

### Validação em Batch

```go
func validateMultipleUsers(users []User) map[int]*ValidationResult {
    ctx := context.Background()
    validator := schema.NewJSONSchemaValidator()
    results := make(map[int]*ValidationResult)
    
    for i, user := range users {
        result := validator.ValidateSchema(ctx, user, userSchema)
        if !result.Valid {
            results[i] = result
        }
    }
    
    return results
}
```

### Validação Condicional

```go
const userSchema = `{
    "type": "object",
    "properties": {
        "type": {"type": "string", "enum": ["admin", "user"]},
        "permissions": {"type": "array"}
    },
    "if": {
        "properties": {"type": {"const": "admin"}}
    },
    "then": {
        "required": ["permissions"]
    }
}`
```

Veja mais exemplos no diretório [examples/](examples/) que inclui:

- **Exemplo Principal** (`main.go`): Demonstração completa de todas as funcionalidades
- **Validação de Formatos**: Todos os validadores de formato disponíveis
- **Schemas Avançados**: Validação condicional, objetos aninhados, arrays
- **Context e Performance**: Uso de timeouts, cancelamentos e medição de performance
- **Integração com Domain Errors**: Tratamento avançado de erros
- **Validação em Batch**: Processamento em lote para alta performance
- **Casos de Uso Complexos**: Exemplos práticos para aplicações reais

### Performance Benchmarks

O exemplo inclui benchmarks que demonstram:
- **Validação Individual**: ~50µs por validação
- **Validação em Batch**: ~10+ validações/ms
- **Reutilização de Validators**: Significativamente mais eficiente
- **Context Timeout**: Suporte robusto a cancelamento

### Executando os Exemplos

```bash
cd validator/schema/examples
go run main.go
```

Isso executará todos os 10 exemplos demonstrando:
1. Validações básicas
2. Validação de structs
3. API fluente com builders
4. Validação JSON Schema
5. Regras customizadas
6. Todos os validadores de formato
7. Schemas avançados
8. Context e performance
9. Integração com domain errors
10. Validação em batch
