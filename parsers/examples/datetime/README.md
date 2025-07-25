# DateTime Parser Examples

Este diretório contém exemplos práticos de uso do parser DateTime do nexs-lib.

## Funcionalidades Demonstradas

- **Auto-detecção de formato**: Detecção automática de formatos de data/hora
- **Parsing flexível**: Suporte a múltiplos formatos de entrada
- **Precisão configurável**: Controle de precisão (ano, mês, dia, hora, etc.)
- **Fuso horário**: Suporte a UTC e fusos horários locais
- **Formatação**: Conversão para diferentes formatos de saída
- **Validação**: Validação robusta de datas e horários
- **Compatibilidade**: Funções de compatibilidade com módulo antigo

## Arquivos de Exemplo

- `basic_usage.go` - Exemplos básicos de parsing de data/hora
- `format_detection.go` - Demonstração de auto-detecção de formatos
- `timezone_handling.go` - Manipulação de fusos horários
- `precision_control.go` - Controle de precisão temporal
- `validation.go` - Validação e tratamento de erros
- `compatibility.go` - Migração do módulo antigo

## Como Executar

```bash
cd parsers/examples/datetime
go run basic_usage.go
go run format_detection.go
go run timezone_handling.go
go run precision_control.go
go run validation.go
go run compatibility.go
```

## Principais Funcionalidades

### Parsing Básico
```go
parser := datetime.NewParser()
result, err := parser.ParseString(ctx, "2025-01-15 14:30:00")
```

### Auto-detecção de Formato
```go
// Detecta automaticamente diferentes formatos
formats := []string{
    "2025-01-15 14:30:00",
    "15/01/2025 14:30",
    "Jan 15, 2025 2:30 PM",
    "2025-01-15T14:30:00Z",
}
```

### Precisão Configurável
```go
// Define precisão desejada
result, err := parser.ParseWithPrecision(ctx, input, "second")
```

### Fuso Horário
```go
// Parsing com fuso horário específico
result, err := parser.ParseInTimezone(ctx, input, "America/Sao_Paulo")
```

### Formatação
```go
// Converte para formato específico
formatted := result.Format("2006-01-02 15:04:05")
```

## Formatos Suportados

- **ISO 8601**: `2025-01-15T14:30:00Z`
- **RFC 3339**: `2025-01-15T14:30:00-03:00`
- **Brasileiro**: `15/01/2025 14:30:00`
- **Americano**: `01/15/2025 2:30:00 PM`
- **Europeu**: `15.01.2025 14:30:00`
- **Unix Timestamp**: `1737033000`
- **Formato customizado**: Qualquer formato válido do Go

## Precisão Temporal

- `year` - Apenas ano
- `month` - Até mês
- `day` - Até dia
- `hour` - Até hora
- `minute` - Até minuto
- `second` - Até segundo
- `nanosecond` - Precisão máxima

## Fusos Horários

- Suporte completo a fusos horários IANA
- Conversão automática para UTC
- Preservação do fuso original
- Detecção de horário de verão

## Validação

- Validação de datas impossíveis
- Verificação de formatos
- Detecção de ambiguidades
- Tratamento de erros detalhado

## Compatibilidade

Mantém compatibilidade total com o módulo `_old/parse/datetime`:

```go
// Função de compatibilidade
result, err := datetime.ParseDateTimeCompat(input)

// Aliases para migração
result, err := datetime.Parse(input)
result, err := datetime.ParseString(input)
```
