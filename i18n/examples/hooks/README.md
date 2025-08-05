# Exemplo de Hooks

Este exemplo demonstra o uso do sistema de hooks do pacote i18n para reagir a eventos de tradução:
- Hook OnTranslationLoaded
- Hook OnMissingTranslation
- Hook OnTranslationRequested

## Estrutura

```
hooks/
├── main.go
├── translations/
│   └── translations.pt-BR.json
└── README.md
```

## Execução

```bash
go run main.go
```

## Arquivo de Traduções

O arquivo `translations.pt-BR.json` contém:

```json
{
    "welcome": "Bem-vindo ao exemplo de hooks!",
    "greeting": "Olá, {{.Name}}!",
    "items": {
        "one": "{{.Count}} item",
        "other": "{{.Count}} itens"
    }
}
```

## Hooks Demonstrados

1. `OnTranslationLoaded`
   - Executado quando um arquivo de tradução é carregado
   - Útil para validação e log

2. `OnMissingTranslation`
   - Executado quando uma chave de tradução não é encontrada
   - Útil para logging e fallback

3. `OnTranslationRequested`
   - Executado antes de cada tradução
   - Útil para métricas e debug

## Funcionalidades Demonstradas

1. Configuração de hooks
2. Registro de eventos de tradução
3. Tratamento de traduções ausentes
4. Logging de atividades de tradução
5. Métricas de uso

## Código

O exemplo demonstra:
- Como registrar hooks
- Como tratar eventos de tradução
- Como implementar fallbacks
- Como coletar métricas
- Como fazer logging de eventos
