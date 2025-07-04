# Strutl - Utilitários de String para Go

`strutl` é um pacote Go que fornece funções avançadas para manipulação de strings, otimizado para performance e com suporte completo a caracteres Unicode.

## Características

- Manipulação adequada de caracteres UTF-8/Unicode
- Alta performance em operações com strings
- Funções para manipulação de casos (camelCase, snake_case, kebab-case)
- Funções para extrair substrings
- Funções para normalização de strings (remoção de acentos, slugify)
- Funções para alinhamento de texto (esquerda, direita, centro)
- Funções para preenchimento de strings (pad left, pad right, pad center)
- Funções para manipulação de palavras
- Utilitários para verificação de strings (IsASCII)
- Suporte a reverter strings preservando caracteres UTF-8
- Funções para quebra de linhas (WordWrap)
- Funções para desenhar caixas de texto (DrawBox)
- Funções para expandir tabs (ExpandTabs)
- Funções para indentação de texto (Indent)
- Funções para geração de strings aleatórias (Random)
- Funções para criar resumos de texto (Summary)
- Detecção automática de quebra de linha por sistema operacional (OSNewLine)
- 100% coberto por testes unitários e benchmarks

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/strutl
```

## Uso

```go
import "github.com/fsvxavier/nexs-lib/strutl"
```

### Funções Básicas

```go
// Obter o tamanho correto de strings (considera caracteres Unicode)
length := strutl.Len("こんにちは")  // = 5, não 15

// Reverter uma string
reversed := strutl.Reverse("Hello, 世界")  // = "界世 ,olleH"
```

### Manipulação de Caso

```go
// Converter para camelCase
camel := strutl.ToCamel("hello_world")  // = "HelloWorld"
lowerCamel := strutl.ToLowerCamel("hello_world")  // = "helloWorld"

// Converter para snake_case
snake := strutl.ToSnake("HelloWorld")  // = "hello_world"
screamingSnake := strutl.ToScreamingSnake("HelloWorld")  // = "HELLO_WORLD"

// Converter para kebab-case
kebab := strutl.ToKebab("HelloWorld")  // = "hello-world"
screamingKebab := strutl.ToScreamingKebab("HelloWorld")  // = "HELLO-WORLD"
```

### Substrings

```go
// Extrair substrings (com suporte a Unicode)
sub := strutl.Substring("Hello, 世界!", 0, 5)  // = "Hello"
subUtf := strutl.Substring("Hello, 世界!", 7, 9)  // = "世界"

// Extrair antes/depois de um separador
after := strutl.SubstringAfter("hello-world", "-")  // = "world"
before := strutl.SubstringBefore("hello-world", "-")  // = "hello"
afterLast := strutl.SubstringAfterLast("hello-world-again", "-")  // = "again"
beforeLast := strutl.SubstringBeforeLast("hello-world-again", "-")  // = "hello-world"
```

### Normalização

```go
// Remover acentos
noAccents := strutl.RemoveAccents("olá mundo")  // = "ola mundo"

// Criar slugs para URLs
slug := strutl.Slugify("Hello, World!")  // = "hello-world"
```

### Alinhamento e Preenchimento

```go
// Alinhar texto
center := strutl.AlignCenter("hello", 11)  // = "   hello   "
right := strutl.AlignRight("hello", 10)    // = "     hello"
left := strutl.AlignLeft("  hello  ")      // = "hello  "

// Preencher strings
padLeft := strutl.PadLeft("123", 5, "0")   // = "00123"
padRight := strutl.PadRight("123", 5, "0") // = "12300"
padCenter := strutl.Pad("text", 8, "<", ">") // = "<<text>>"
```

### Manipulação de Palavras

```go
// Extrair palavras
words := strutl.Words("Hello, world!")     // = ["Hello", "world"]
count := strutl.CountWords("Hello, world!") // = 2

// Verificar se uma string é ASCII
isASCII := strutl.IsASCII("hello")         // = true
isASCII = strutl.IsASCII("olá")            // = false
```

### Indentação e Formatação de Texto

```go
// Indentar texto
indented := strutl.Indent("linha1\nlinha2", "  ")  // = "  linha1\n  linha2"

// Expandir tabs
expanded := strutl.ExpandTabs("\tHello\tWorld", 4)  // = "    Hello    World"

// Quebrar texto em linhas
wrapped := strutl.WordWrap("This is a long text", 10, false)  // = "This is a\nlong text"

// Desenhar caixas de texto
box, _ := strutl.DrawBox("Hello World", 20, strutl.Center)
// ┌──────────────────┐
// │   Hello World    │
// └──────────────────┘

// Usar caixas de texto personalizadas
customBox, _ := strutl.DrawCustomBox("Hello", 10, strutl.Left, &strutl.SimpleBox9Slice(), "\n")
// +--------+
// |Hello   |
// +--------+

// Detectar quebra de linha do sistema
newLine := strutl.OSNewLine()  // "\r\n" no Windows, "\n" em outros sistemas
```

### Strings Aleatórias

```go
// Gerar string aleatória a partir de um conjunto de caracteres
random, _ := strutl.Random("abcdef123456", 8)  // Exemplo: "a3d5ef2b"
```

### Resumos de Texto

```go
// Criar resumos de texto
summary := strutl.Summary("Lorem ipsum dolor sit amet", 12, "...")  // = "Lorem ipsum..."
```

### Substituição

```go
// Substituir substrings
replaced := strutl.ReplaceAll("hello world", "world", "golang")  // = "hello golang"
replacedFirst := strutl.ReplaceFirst("hello hello", "hello", "hi")  // = "hi hello"
replacedLast := strutl.ReplaceLast("hello hello", "hello", "hi")   // = "hello hi"
```

### Configuração de Acrônimos

```go
// Configurar acrônimos para conversão de casos
strutl.ConfigureAcronym("API", "api")
camel := strutl.ToCamel("API")  // = "Api" em vez de "Api"
```

## Performance

O pacote `strutl` foi desenvolvido com foco em performance. Todas as funções foram benchmarked para garantir operações eficientes, mesmo com strings grandes.

## Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou enviar pull requests.

## Licença

Este projeto está licenciado sob os termos da licença MIT.
