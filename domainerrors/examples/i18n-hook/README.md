# Hook de Tradução i18n para Domain Errors

Este hook implementa tradução automática de mensagens de erro usando o módulo i18n da nexs-lib.

## Funcionalidades

- **Tradução Automática**: Traduz mensagens de erro baseando-se no idioma detectado do contexto
- **Múltiplas Estratégias**: Usa diferentes estratégias para encontrar traduções (código, mensagem normalizada, palavras-chave)
- **Fallback Language**: Suporte a idioma de fallback quando a tradução não está disponível
- **Detecção de Idioma**: Detecta idioma do contexto através de headers Accept-Language, contexto ou locale do usuário
- **Metadados de Tradução**: Preserva informações originais e de tradução nos metadados do erro
- **Tradução de Códigos**: Opcionalmente traduz códigos de erro além das mensagens

## Uso Básico

```go
// Configuração do hook
config := hooks.I18nTranslationConfig{
    TranslationsPath: "./translations",
    DefaultLanguage:  "en",
    FallbackLanguage: "en", 
    SupportedLangs:   []string{"en", "pt", "es"},
    FilePattern:      "{lang}.json",
    TranslateCodes:   false,
}

// Criação do hook
i18nHook, err := hooks.NewI18nTranslationHook(
    interfaces.HookTypeAfterError,
    config,
)
if err != nil {
    log.Fatal(err)
}

// Registro do hook no sistema de domain errors
// (depende da implementação do sistema de hooks)
```

## Configuração

### I18nTranslationConfig

| Campo | Tipo | Descrição | Padrão |
|-------|------|-----------|--------|
| `TranslationsPath` | `string` | Caminho para arquivos de tradução | `./translations` |
| `DefaultLanguage` | `string` | Idioma padrão | `en` |
| `FallbackLanguage` | `string` | Idioma de fallback | `en` |
| `SupportedLangs` | `[]string` | Lista de idiomas suportados | `["en", "pt", "es"]` |
| `FilePattern` | `string` | Padrão dos arquivos de tradução | `{lang}.json` |
| `TranslateCodes` | `bool` | Se deve traduzir códigos de erro | `false` |
| `CustomTranslations` | `map[string]map[string]interface{}` | Traduções customizadas em memória | `nil` |

## Estrutura dos Arquivos de Tradução

### Exemplo: en.json
```json
{
  "USER_NOT_FOUND": "User not found",
  "VALIDATION_FAILED": "Validation failed", 
  "error.not_found": "Resource not found",
  "error.validation": "Validation error",
  "errors.user_not_found": "The requested user could not be found",
  "errors.email_invalid": "Please provide a valid email address"
}
```

### Exemplo: pt.json
```json
{
  "USER_NOT_FOUND": "Usuário não encontrado",
  "VALIDATION_FAILED": "Falha na validação",
  "error.not_found": "Recurso não encontrado", 
  "error.validation": "Erro de validação",
  "errors.user_not_found": "O usuário solicitado não pôde ser encontrado",
  "errors.email_invalid": "Por favor, forneça um endereço de email válido"
}
```

## Estratégias de Tradução

O hook utiliza múltiplas estratégias para encontrar traduções, na seguinte ordem:

1. **Código Direto**: Usa o código do erro como chave de tradução
2. **Código com Prefixo**: Tenta `error.{code}` como chave
3. **Mensagem Normalizada**: Normaliza a mensagem removendo pontuação e espaços
4. **Palavras-chave Comuns**: Detecta padrões comuns como "not found", "validation", etc.

## Detecção de Idioma

O hook detecta o idioma do contexto na seguinte ordem de prioridade:

1. Header `Accept-Language` do contexto
2. Chave `language` do contexto
3. Chave `user_locale` do contexto (extrai idioma do locale)
4. Idioma padrão configurado

## Metadados Adicionados

Quando uma tradução é realizada, o hook adiciona os seguintes metadados ao erro:

- `original_message`: Mensagem original antes da tradução
- `translated_language`: Idioma para o qual foi traduzido
- `translation_source`: Fonte da tradução (`i18n_hook`)
- `original_code`: Código original (se `TranslateCodes` estiver habilitado)

## Exemplo de Execução

Execute o exemplo:

```bash
cd domainerrors/examples/i18n-hook
go run main.go
```

### Saída Esperada

```
=== Exemplo de Hook i18n Translation ===
Hook criado: i18n_translation_hook_after_error
Idiomas suportados: [en pt es]

--- Testando traduções para idioma: en ---
  Caso: Erro com código traduzível
    Original: [USER_NOT_FOUND] User not found
    Traduzido: [USER_NOT_FOUND] User not found
    Metadados: original='User not found', lang='en'

--- Testando traduções para idioma: pt ---
  Caso: Erro com código traduzível
    Original: [USER_NOT_FOUND] User not found  
    Traduzido: [USER_NOT_FOUND] Usuário não encontrado
    Metadados: original='User not found', lang='pt'

--- Testando traduções para idioma: es ---
  Caso: Erro com código traduzível
    Original: [USER_NOT_FOUND] User not found
    Traduzido: [USER_NOT_FOUND] Usuario no encontrado
    Metadados: original='User not found', lang='es'
```

## Integração com Middlewares

O hook pode ser combinado com middlewares para criar pipelines completos de processamento de erros:

```go
// Registrar hook primeiro (tradução)
registry.RegisterHook(i18nHook)

// Depois registrar middlewares (logging, métricas, etc.)
registry.RegisterMiddleware(loggingMiddleware)
registry.RegisterMiddleware(metricsMiddleware)
```

## Considerações de Performance

- As traduções são carregadas uma vez no início
- Suporte a cache interno do módulo i18n
- Estratégias de tradução executam em ordem de eficiência
- Fallback rápido para idioma padrão em caso de erro

## Limitações

- Atualizações de tradução em runtime não são suportadas
- Depende da estrutura de arquivos configurada
- Requer que o contexto contenha informações de idioma para detecção automática
