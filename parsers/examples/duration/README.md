# Duration Parser Examples

Este diretório contém exemplos práticos de uso do parser Duration do nexs-lib.

## Funcionalidades Demonstradas

- **Parsing flexível**: Suporte a múltiplas unidades de tempo
- **Unidades estendidas**: Suporte a dias (d) e semanas (w) além das unidades padrão do Go
- **Formatos múltiplos**: Parse de diferentes formatos de duração
- **Validação**: Validação de durações e detecção de erros
- **Conversão**: Conversão entre diferentes unidades de tempo
- **Formatação**: Formatação legível de durações

## Arquivos de Exemplo

- `basic_usage.go` - Exemplos básicos de parsing de duração
- `extended_units.go` - Demonstração de unidades estendidas (dias, semanas)
- `formatting.go` - Formatação e conversão de durações
- `validation.go` - Validação de durações

## Como Executar

```bash
cd parsers/examples/duration
go run basic_usage.go
go run extended_units.go
go run formatting.go
go run validation.go
```

## Principais Funcionalidades

### Parsing Básico
```go
parser := duration.NewParser()
result, err := parser.ParseString(ctx, "1h30m")
```

### Unidades Suportadas
- `ns` - nanosegundos
- `us`, `µs`, `μs` - microssegundos  
- `ms` - milissegundos
- `s` - segundos
- `m` - minutos
- `h` - horas
- `d` - dias (24 horas)
- `w` - semanas (7 dias)

### Formatos Aceitos
```go
inputs := []string{
    "1h30m",           // 1 hora e 30 minutos
    "2h45m30s",        // 2 horas, 45 minutos e 30 segundos
    "1d12h",           // 1 dia e 12 horas
    "2w3d",            // 2 semanas e 3 dias
    "500ms",           // 500 milissegundos
    "1.5h",            // 1.5 horas
    "90m",             // 90 minutos
}
```

### Conversão
```go
// Converter para diferentes unidades
hours := result.Duration.Hours()
minutes := result.Duration.Minutes()
seconds := result.Duration.Seconds()
```

### Formatação
```go
// Formatação legível
formatted := result.Duration.String()
```

## Validação

- Detecção de formatos inválidos
- Validação de valores numéricos
- Verificação de unidades suportadas
- Tratamento de overflow

## Conversões Úteis

- Segundos para horas/minutos/segundos
- Milissegundos para segundos
- Dias para horas
- Semanas para dias
- E vice-versa para todas as unidades
