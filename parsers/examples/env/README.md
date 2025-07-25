# Environment Parser Examples

Este diretório contém exemplos práticos de uso do parser Environment do nexs-lib.

## Funcionalidades Demonstradas

- **Parsing de variáveis**: Leitura e conversão de variáveis de ambiente
- **Tipos suportados**: string, int, float, bool, slice, map
- **Valores padrão**: Definição de valores padrão quando variável não existe
- **Validação**: Validação de formatos e valores
- **Conversão automática**: Conversão automática de tipos

## Arquivos de Exemplo

- `basic_usage.go` - Exemplos básicos de parsing de variáveis de ambiente

## Como Executar

```bash
cd parsers/examples/env
# Defina algumas variáveis de ambiente
export APP_NAME="MyApp"
export APP_PORT="8080"
export APP_DEBUG="true"
export APP_TAGS="web,api,service"

go run basic_usage.go
```

## Principais Funcionalidades

### Parsing Básico
```go
parser := env.NewParser()
value, err := parser.ParseString(ctx, "APP_NAME")
```

### Tipos Suportados
- `string` - texto
- `int`, `int64` - números inteiros
- `float64` - números decimais
- `bool` - verdadeiro/falso
- `[]string` - lista de strings
- `map[string]string` - mapa chave-valor

### Conversão Automática
```go
// Converte automaticamente para o tipo desejado
port, err := env.ParseInt("APP_PORT")
debug, err := env.ParseBool("APP_DEBUG")
tags, err := env.ParseStringSlice("APP_TAGS")
```
