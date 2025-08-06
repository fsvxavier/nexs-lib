# Middleware de Tradução i18n para Domain Errors

Este middleware implementa tradução automática de mensagens de erro e metadados usando o módulo i18n da nexs-lib, funcionando como parte de uma cadeia de middlewares.

## Diferenças entre Hook e Middleware

### 🪝 Hook (Event-Driven)
- **Reage** a eventos específicos
- Executa **side effects** (logging, auditoria, notificações)
- **NÃO modifica** o erro diretamente
- Múltiplos hooks podem ser registrados para o mesmo evento

### 🔧 Middleware (Processing Pipeline)
- **Transforma/enriquece** erros em uma cadeia de processamento
- Segue o padrão **Chain of Responsibility**
- **MODIFICA** o erro: adiciona metadados, contexto, traduz mensagens
- **Ordem importa**: executados em sequência de prioridade

## Pipeline de Execução

```
1. Erro criado
2. Middleware chain executa (transforma erro) ← MIDDLEWARE i18n
3. Hooks before_* executam (side effects)
4. Processamento interno
5. Hooks after_* executam (side effects) 
6. Erro final retornado
```

## Funcionalidades do Middleware

- **Tradução de Mensagens**: Traduz mensagens de erro baseando-se no idioma detectado
- **Tradução de Códigos**: Opcionalmente traduz códigos de erro  
- **Tradução de Metadados**: Traduz campos específicos nos metadados do erro
- **Tradução de Contexto**: Traduz metadados do contexto do middleware
- **Múltiplas Estratégias**: Usa diferentes estratégias para encontrar traduções
- **Detecção de Idioma**: Detecta idioma do contexto através de headers, contexto ou locale
- **Fallback Language**: Suporte a idioma de fallback
- **Chain Integration**: Integra-se perfeitamente com cadeia de middlewares
- **Preservação de Originais**: Mantém valores originais nos metadados

## Uso Básico

```go
// Configuração do middleware
config := middlewares.I18nTranslationConfig{
    TranslationsPath:  "./translations",
    DefaultLanguage:   "en",
    FallbackLanguage:  "en",
    SupportedLangs:    []string{"en", "pt", "es"},
    FilePattern:       "{lang}.json",
    TranslateCodes:    true,
    TranslateMetadata: true,
}

// Criação do middleware
i18nMiddleware, err := middlewares.NewI18nTranslationMiddleware(config)
if err != nil {
    log.Fatal(err)
}

// Uso em uma cadeia de middlewares
chain := NewMiddlewareChain()
chain.Use(i18nMiddleware)           // Tradução (alta prioridade)
chain.Use(enrichmentMiddleware)     // Enriquecimento 
chain.Use(loggingMiddleware)        // Logging (baixa prioridade)

// Execução da cadeia
err := chain.Execute(middlewareContext)
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
| `TranslateMetadata` | `bool` | Se deve traduzir metadados | `false` |
| `CustomTranslations` | `map[string]map[string]interface{}` | Traduções customizadas | `nil` |

## Estrutura dos Arquivos de Tradução

### Exemplo: pt.json
```json
{
  "USER_NOT_FOUND": "Usuário não encontrado",
  "VALIDATION_FAILED": "Falha na validação",
  "error.not_found": "Recurso não encontrado",
  "error.validation": "Erro de validação",
  "code.usr_404": "USUARIO_NAO_ENCONTRADO",
  "validation_message": "Falha na validação do campo",
  "business_rule_message": "Violação de regra de negócio",
  "operation_description": "Processando dados do usuário"
}
```

## Estratégias de Tradução

### Mensagens de Erro
1. **Código Direto**: `USER_NOT_FOUND` → tradução
2. **Código com Prefixos**: `error.USER_NOT_FOUND`, `errors.USER_NOT_FOUND`, etc.
3. **Mensagem Normalizada**: `"user not found"` → `"user_not_found"`
4. **Palavras-chave Comuns**: detecta padrões como "not found", "validation", etc.

### Códigos de Erro
- **Prefixos**: `code.{CODE}`, `codes.{CODE}`, `error_code.{CODE}`

### Metadados Traduzidos

#### Metadados do Erro
- `validation_message`
- `business_rule_message`
- `constraint_message`
- `field_error`
- `detail_message`
- `user_message`
- `description`
- `reason`
- `suggestion`
- Arrays: `validation_errors`, `field_errors`, `business_rules`, `constraints`

#### Metadados do Contexto
- `operation_description`
- `step_description`
- `process_name`
- `action_description`
- `status_message`

## Detecção de Idioma

Ordem de prioridade:
1. Header `Accept-Language` do contexto
2. Chave `language` do contexto
3. Chave `user_locale` do contexto
4. Chave `user_language` do contexto
5. Idioma padrão configurado

## Metadados Adicionados

O middleware preserva informações originais:

### No Erro
- `original_message`: Mensagem original antes da tradução
- `original_code`: Código original (se `TranslateCodes` habilitado)
- `translated_language`: Idioma para o qual foi traduzido
- `translation_source`: Fonte da tradução (`i18n_middleware`)
- `translation_timestamp`: Timestamp da tradução
- `{campo}_original`: Valor original de cada campo traduzido

### No Contexto
- `{campo}_original`: Valores originais dos campos traduzidos do contexto

## Exemplo de Execução

Execute o exemplo:

```bash
cd domainerrors/examples/i18n-middleware
go run main.go
```

### Saída Esperada

```
=== Exemplo de Middleware i18n Translation ===
Middleware criado: i18n_translation_middleware
Idiomas suportados: [en pt es]

--- Testando middleware para idioma: pt ---
  1. Erro com código e metadados traduzíveis
     Original: [USER_NOT_FOUND] User not found
     [Next Middleware] Processando erro traduzido...
     Traduzido: [USUARIO_NAO_ENCONTRADO] Usuário não encontrado
     Metadados erro: original='User not found', lang='pt'
     Metadado traduzido: validation_message='Falha na validação do campo' (original='Field validation failed')
     Contexto traduzido: operation_description='Processando dados do usuário' (original='Processing user data')
```

## Comparação Hook vs Middleware

| Aspecto | Hook | Middleware |
|---------|------|------------|
| **Propósito** | Side effects | Transformação |
| **Modifica Erro** | ❌ Não | ✅ Sim |
| **Ordem de Execução** | Por tipo de evento | Por prioridade na cadeia |
| **Chain Pattern** | ❌ Não | ✅ Sim |
| **Uso Principal** | Logging, auditoria | Enriquecimento, tradução |
| **Função Next** | ❌ Não tem | ✅ Chama próximo |

## Vantagens do Middleware

1. **Chain of Responsibility**: Fácil composição com outros middlewares
2. **Controle de Fluxo**: Pode interromper a cadeia se necessário
3. **Transformação**: Modifica diretamente o erro e contexto
4. **Ordem Determinística**: Execução baseada em prioridade
5. **Reutilização**: Pode ser usado em diferentes cadeias
6. **Separação de Responsabilidades**: Cada middleware tem uma função específica

## Integração com Sistema de Domain Errors

```go
// Setup da cadeia de middlewares
middlewareChain := domainerrors.NewMiddlewareChain()

// Adição dos middlewares em ordem de prioridade
middlewareChain.Use(i18nMiddleware)        // Prio 10 - Tradução
middlewareChain.Use(enrichmentMiddleware)  // Prio 20 - Enriquecimento  
middlewareChain.Use(validationMiddleware)  // Prio 30 - Validação
middlewareChain.Use(loggingMiddleware)     // Prio 40 - Logging

// Execução automática na criação de erros
domainError := domainerrors.NewValidationError("USER_NOT_FOUND", "User not found")
// ↑ Automaticamente executa a cadeia de middlewares
```

## Performance e Considerações

- **Cache de Traduções**: Utiliza cache interno do módulo i18n
- **Execução Sequencial**: Middlewares executam em ordem de prioridade
- **Fallback Rápido**: Mudança para idioma padrão em caso de erro
- **Lazy Loading**: Traduções carregadas sob demanda
- **Memory Efficient**: Reutiliza instâncias de provider i18n
