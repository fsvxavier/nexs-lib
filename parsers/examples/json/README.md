# JSON Parser Examples

Este diretório contém exemplos práticos de uso do parser JSON do nexs-lib.

## Funcionalidades Demonstradas

- **Parsing básico**: Conversão de JSON string/bytes para Go types
- **Parsing avançado**: Suporte a comentários, vírgulas finais
- **Formatação**: Compactação e pretty-printing
- **Formatos especiais**: JSON Lines (JSONL), NDJSON, JSON5
- **Utilitários**: Merge, extração de caminhos, validação
- **Streaming**: Parsing de grandes volumes de dados
- **Compatibilidade**: Funções de compatibilidade com módulo antigo

## Arquivos de Exemplo

- `basic_usage/main.go` - Exemplos básicos de parsing e formatação
- `advanced_features/main.go` - Funcionalidades avançadas (comentários, vírgulas finais)
- `special_formats/main.go` - JSONL, NDJSON, JSON5
- `utilities/main.go` - Merge, extração de caminhos, validação
- `streaming/main.go` - Parsing de streaming para grandes datasets
- `compatibility/main.go` - Exemplos de migração do módulo antigo

## Como Executar

```bash
cd parsers/examples/json
go run basic_usage/main.go
go run advanced_features/main.go
go run special_formats/main.go
go run utilities/main.go
go run streaming/main.go
go run compatibility/main.go
```

## Principais Funcionalidades

### Parsing Básico
```go
// String para interface{}
result, err := json.ParseJSONString(`{"name": "Alice", "age": 30}`)

// Bytes para interface{}
result, err := json.ParseJSONBytes([]byte(`[1, 2, 3]`))

// Type-safe parsing
user, err := json.ParseJSONToType[User](data)
```

### Formatação
```go
// Compactação
compact, err := json.CompactJSON(jsonString)

// Pretty printing
pretty, err := json.PrettyJSON(jsonString, "  ")

// Formatação customizada
formatter := json.NewFormatterWithIndent("\t")
formatted, err := formatter.FormatString(ctx, data)
```

### Funcionalidades Avançadas
```go
parser := json.NewAdvancedParser().
    WithComments(true).
    WithTrailingCommas(true)

result, err := parser.ParseAdvanced(ctx, jsonWithComments)
```

### Formatos Especiais
```go
// JSON Lines
results, err := json.ParseJSONL(jsonlString)

// JSON5 (básico)
result, err := json.ParseJSON5(json5String)
```

### Utilitários
```go
// Merge de objetos
merged, err := json.MergeJSON(obj1, obj2, obj3)

// Extração de caminhos
value, err := json.ExtractPath(data, "user.profile.name")

// Validação
err := json.ValidateJSONString(jsonString)
```

### Streaming
```go
parser := json.NewStreamParser(reader)
for parser.HasMore() {
    var obj MyType
    err := parser.ParseNext(&obj)
    // processa obj
}
```

## Compatibilidade

O módulo mantém total compatibilidade com o módulo `_old/parse`:

```go
// Função de compatibilidade direta
result, err := json.ParseJSONToTypeCompat[User](data)

// Aliases para migração fácil
result, err := json.Parse(data)
result, err := json.ParseString(jsonString)
result, err := json.ParseBytes(jsonBytes)
err := json.Validate(data)
```
