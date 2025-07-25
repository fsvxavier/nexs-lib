# JSON Schema Validation Library

Uma biblioteca modular e extensível para validação JSON Schema em Go, com suporte a múltiplos engines de validação, hooks customizados e checks adicionais.

## 🚀 Características

- **Múltiplos Engines**: Suporte aos principais libraries JSON Schema do Go
  - `kaptinlin/jsonschema` (principal)
  - `xeipuuv/gojsonschema` (retrocompatibilidade)
  - `santhosh-tekuri/jsonschema` (v6)
- **Arquitetura Modular**: Hooks e checks customizáveis
- **Retrocompatibilidade**: Compatível com `_old/validator`
- **Alta Performance**: Otimizado para uso em produção
- **Configuração Flexível**: Provider agnóstico com injeção de dependências

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/validation/jsonschema
```

## 🔧 Uso Básico

### Validação Simples

```go
package main

import (
    "fmt"
    "github.com/fsvxavier/nexs-lib/validation/jsonschema"
    "github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

func main() {
    // Criar validador com configuração padrão
    validator, err := jsonschema.NewValidator(nil)
    if err != nil {
        panic(err)
    }

    // Schema JSON
    schema := []byte(`{
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "number"}
        },
        "required": ["name"]
    }`)

    // Dados para validar
    data := map[string]interface{}{
        "name": "John",
        "age":  30,
    }

    // Executar validação
    errors, err := validator.ValidateFromBytes(schema, data)
    if err != nil {
        panic(err)
    }

    if len(errors) > 0 {
        fmt.Printf("Validation failed with %d errors\n", len(errors))
        for _, validationError := range errors {
            fmt.Printf("- %s: %s\n", validationError.Field, validationError.Message)
        }
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### Configuração com Provider Específico

```go
// Usar gojsonschema para compatibilidade
cfg := config.NewConfig().WithProvider(config.GoJSONSchemaProvider)
validator, err := jsonschema.NewValidator(cfg)

// Usar kaptinlin para performance
cfg := config.NewConfig().WithProvider(config.JSONSchemaProvider)
validator, err := jsonschema.NewValidator(cfg)
```

### Validação com Schema de Arquivo

```go
errors, err := validator.ValidateFromFile("schema.json", data)
```

### Registrar Schemas para Reutilização

```go
cfg := config.NewConfig()
cfg.RegisterSchema("user-schema", userSchema)

validator, _ := jsonschema.NewValidator(cfg)
errors, err := validator.ValidateFromStruct("user-schema", userData)
```

## 🔧 Configuração Avançada

### Hooks de Pré-Validação

```go
cfg := config.NewConfig()

// Normalização de dados
normalizationHook := &hooks.DataNormalizationHook{
    TrimStrings:   true,
    LowerCaseKeys: false,
}
cfg.AddPreValidationHook(normalizationHook)

// Logging
loggingHook := &hooks.LoggingHook{LogData: true}
cfg.AddPreValidationHook(loggingHook)
```

### Hooks de Pós-Validação

```go
// Enriquecimento de erros
enrichmentHook := &hooks.ErrorEnrichmentHook{
    AddContext:     true,
    AddSuggestions: true,
}
cfg.AddPostValidationHook(enrichmentHook)

// Resumo de validação
summaryHook := &hooks.ValidationSummaryHook{LogSummary: true}
cfg.AddPostValidationHook(summaryHook)
```

### Hooks de Erro

```go
// Filtrar erros
filterHook := &hooks.ErrorFilterHook{
    IgnoreFields: []string{"debug_info"},
    MaxErrors:    10,
}
cfg.AddErrorHook(filterHook)

// Notificações
notificationHook := &hooks.ErrorNotificationHook{
    NotifyOnError: true,
    WebhookURL:    "https://api.company.com/alerts",
}
cfg.AddErrorHook(notificationHook)
```

### Checks Adicionais

```go
// Validação de campos obrigatórios
requiredCheck := &checks.RequiredFieldsCheck{
    RequiredFields: []string{"email", "password"},
}
cfg.AddCheck(requiredCheck)

// Validação de enums
enumCheck := &checks.EnumConstraintsCheck{
    Constraints: map[string][]interface{}{
        "status": {"active", "inactive", "pending"},
        "role":   {"admin", "user", "guest"},
    },
}
cfg.AddCheck(enumCheck)

// Validação de data
dateCheck := &checks.DateValidationCheck{
    DateFields:  []string{"birth_date", "created_at"},
    AllowFuture: false,
    AllowPast:   true,
}
cfg.AddCheck(dateCheck)
```

### Formatos Customizados

```go
// Registrar formato customizado
cfg.AddCustomFormat("cpf", &CPFFormatChecker{})

// Ou usando função
cfg.AddCustomFormat("phone", func(input interface{}) bool {
    if str, ok := input.(string); ok {
        return len(str) >= 10 && len(str) <= 15
    }
    return false
})
```

## 🔄 Retrocompatibilidade

A biblioteca mantém total compatibilidade com o código existente:

```go
// Função legacy - continua funcionando
err := jsonschema.Validate(data, schemaString)

// Formatos customizados legacy
jsonschema.AddCustomFormat("custom-format", "^[A-Z]+$")
```

## 🏗️ Estrutura do Projeto

```
validation/jsonschema/
├── interfaces/          # Contratos e interfaces
├── config/             # Configuração do sistema
├── hooks/              # Hooks de pré/pós-validação e erro
├── checks/             # Validações customizadas adicionais
├── providers/          # Implementações dos engines
│   ├── gojsonschema/   # Provider xeipuuv/gojsonschema
│   ├── kaptinlin/      # Provider kaptinlin/jsonschema
│   └── santhosh/       # Provider santhosh-tekuri/jsonschema
├── examples/           # Exemplos de uso
├── json_schema.go      # API principal
└── README.md
```

## 🧪 Execução de Testes

```bash
# Todos os testes
go test -race -timeout 30s ./validation/jsonschema/...

# Com cobertura
go test -race -timeout 30s -coverprofile=coverage.out ./validation/jsonschema/...

# Visualizar cobertura
go tool cover -html=coverage.out
```

## 📈 Performance

A biblioteca foi otimizada para uso em produção:

- **Cache de schemas**: Schemas compilados são reutilizados
- **Execução paralela**: Hooks e checks podem ser executados em paralelo
- **Minimal allocations**: Reduz alocações desnecessárias
- **Provider otimizado**: Engine principal (kaptinlin) oferece melhor performance

## 🔧 Configuração de Mapeamento de Erros

```go
customMapping := map[string]string{
    "required":     "CAMPO_OBRIGATORIO",
    "invalid_type": "TIPO_INVALIDO", 
    "format":       "FORMATO_INVALIDO",
}
cfg.SetErrorMapping(customMapping)
```

## 🚀 Próximos Passos

Veja [NEXT_STEPS.md](NEXT_STEPS.md) para roadmap e melhorias planejadas.

## 📄 Licença

Este projeto está licenciado sob a Licença MIT.
