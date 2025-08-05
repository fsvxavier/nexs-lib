# I18n Module

This module provides internationalization (i18n) support for Go applications with the following features:

- Multiple translation file formats support (JSON, YAML, PO, MO)
- Template variables support using Go's text/template syntax
- Plural forms support
- Language fallback support
- HTTP middleware for automatic language detection
- Hooks system for logging and metrics
- Thread-safe operations

## Installation

```bash
go get github.com/fsvxavier/nexs-lib/i18n
```

## Usage

### Basic Usage

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
    // Create configuration
    cfg := config.DefaultConfig().
        WithTranslationsPath("./translations").
        WithTranslationsFormat("json")

    // Create provider
    provider, err := providers.CreateProvider(cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Load translations
    err = provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Set languages
    err = provider.SetLanguages("pt-BR", "en")
    if err != nil {
        panic(err)
    }

    // Translate
    result, err := provider.Translate("hello_world", nil)
    if err != nil {
        panic(err)
    }
    println(result)

    // Translate with variables
    result, err = provider.Translate("hello_name", map[string]interface{}{
        "Name": "John",
    })
    if err != nil {
        panic(err)
    }
    println(result)

    // Translate plural
    result, err = provider.TranslatePlural("users_count", 2, map[string]interface{}{
        "Count": 2,
    })
    if err != nil {
        panic(err)
    }
    println(result)
}
```

### HTTP Middleware Usage

```go
package main

import (
    "net/http"

    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/middleware"
    "github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
    // Create provider
    provider, _ := providers.CreateProvider("json")

    // Configure middleware
    i18nMiddleware := middleware.New(middleware.Config{
        Provider:        provider,
        QueryParam:      "lang",
        DefaultLanguage: language.MustParse("pt-BR"),
    })

    // Use middleware
    http.Handle("/", i18nMiddleware(http.HandlerFunc(handler)))
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Get provider from context
    provider, _ := middleware.GetProvider(r.Context())

    // Use provider
    result, _ := provider.Translate("hello_world", nil)
    w.Write([]byte(result))
}
```

### Translation File Formats

#### JSON
```json
{
    "hello_world": "Olá mundo!",
    "hello_name": "Olá {{.Name}}!",
    "users_count": {
        "one": "{{.Count}} usuário",
        "other": "{{.Count}} usuários"
    }
}
```

#### YAML
```yaml
hello_world: "Olá mundo!"
hello_name: "Olá {{.Name}}!"
users_count:
  one: "{{.Count}} usuário"
  other: "{{.Count}} usuários"
```

#### PO
```po
msgid "hello_world"
msgstr "Olá mundo!"

msgid "hello_name"
msgstr "Olá {{.Name}}!"

msgid "users_count"
msgid_plural "users_count_plural"
msgstr[0] "{{.Count}} usuário"
msgstr[1] "{{.Count}} usuários"
```

## Features

- [x] Multiple translation file formats
- [x] Template variables
- [x] Plural forms
- [x] Language fallback
- [x] HTTP middleware
- [x] Hooks system
- [x] Thread-safe operations

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
