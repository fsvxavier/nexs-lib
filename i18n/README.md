# Pacote i18n

Este pacote fornece suporte a internacionalização (i18n) para aplicações Go com as seguintes funcionalidades:

- Suporte a múltiplos formatos de arquivo de tradução (JSON, YAML)
- Suporte a variáveis em templates usando a sintaxe do pacote `text/template`
- Suporte a pluralização
- Middleware HTTP para detecção automática de idioma
- Sistema de hooks para logging e métricas
- Operações thread-safe

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/i18n
```

## Uso

### Uso Básico

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
    // Criar configuração
    cfg := config.DefaultConfig().
        WithTranslationsPath("./translations").
        WithTranslationsFormat("json")

    // Criar provider
    provider, err := providers.CreateProvider(cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Carregar traduções
    err = provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Traduzir texto simples
    result, err := provider.Translate("welcome", nil)
    if err != nil {
        panic(err)
    }
    println(result)

    // Traduzir com variáveis
    result, err = provider.Translate("greeting", map[string]interface{}{
        "Name": "John",
    })
    if err != nil {
        panic(err)
    }
    println(result)

    // Traduzir com pluralização
    result, err = provider.TranslatePlural("items", 2, map[string]interface{}{
        "Count": 2,
    })
    if err != nil {
        panic(err)
    }
    println(result)
}
```

### Uso com Middleware HTTP

```go
package main

import (
    "net/http"
    "log"

    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/middleware"
    "github.com/fsvxavier/nexs-lib/i18n/providers"
    "golang.org/x/text/language"
)

func main() {
    // Criar configuração
    cfg := config.DefaultConfig().
        WithTranslationsPath("./translations").
        WithTranslationsFormat("json")

    // Criar provider
    provider, err := providers.CreateProvider(cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Carregar traduções
    err = provider.LoadTranslations("translations.pt-BR.json", cfg.TranslationsFormat)
    if err != nil {
        panic(err)
    }

    // Configurar middleware
    i18nMiddleware := middleware.New(middleware.Config{
        Provider:        provider,
        QueryParam:      "lang",
        DefaultLanguage: language.MustParse("pt-BR"),
    })

    // Criar handlers
    mux := http.NewServeMux()
    mux.Handle("/", i18nMiddleware(http.HandlerFunc(handler)))

    // Iniciar servidor
    log.Println("Servidor rodando em http://localhost:8080")
    http.ListenAndServe(":8080", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Obter provider do contexto
    provider, ok := middleware.GetProvider(r.Context())
    if !ok {
        http.Error(w, "i18n provider not found", http.StatusInternalServerError)
        return
    }

    // Usar o provider
    result, err := provider.Translate("welcome", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "` + result + `"}`))
}
```

### Formatos de Arquivo de Tradução

#### JSON
```json
{
    "welcome": "Bem-vindo!",
    "greeting": "Olá, {{.Name}}!",
    "items": {
        "one": "{{.Count}} item",
        "other": "{{.Count}} itens"
    }
}
```

#### YAML
```yaml
welcome: "Bem-vindo!"
greeting: "Olá, {{.Name}}!"
items:
  one: "{{.Count}} item"
  other: "{{.Count}} itens"
```

## Funcionalidades

- [x] Múltiplos formatos de arquivo de tradução (JSON, YAML)
- [x] Variáveis em templates
- [x] Pluralização
- [x] Middleware HTTP
- [x] Sistema de hooks
- [x] Operações thread-safe

## Exemplos

O pacote inclui diversos exemplos na pasta `examples/`:

- `basic/`: Exemplo básico de uso do pacote
- `formats/`: Demonstração de diferentes formatos de arquivo
- `hooks/`: Uso do sistema de hooks
- `http/`: Integração com servidor HTTP usando middleware

Para mais detalhes, consulte o README.md em cada pasta de exemplo.

## Contribuindo

1. Fork do repositório
2. Crie sua branch de feature (`git checkout -b feature/nova-funcionalidade`)
3. Faça commit das suas alterações (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Crie um novo Pull Request

Para mais informações sobre as próximas funcionalidades planejadas, consulte o arquivo [NEXT_STEPS.md](./NEXT_STEPS.md).
