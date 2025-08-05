# Exemplo Básico

Este exemplo demonstra o uso básico do pacote i18n, incluindo:
- Configuração inicial
- Carregamento de traduções
- Tradução de textos simples
- Uso de variáveis em traduções
- Pluralização

## Estrutura

```
basic/
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

O arquivo `translations.pt-BR.json` contém as seguintes traduções:

```json
{
    "welcome": "Bem-vindo ao exemplo básico!",
    "greeting": "Olá, {{.Name}}!",
    "items": {
        "one": "{{.Count}} item",
        "other": "{{.Count}} itens"
    }
}
```

## Funcionalidades Demonstradas

1. Configuração básica do pacote
2. Criação e configuração do provider
3. Carregamento de traduções
4. Tradução de texto simples
5. Tradução com variáveis
6. Tradução com pluralização

## Código

O exemplo demonstra:
- Como criar uma configuração básica
- Como criar e configurar um provider
- Como carregar arquivos de tradução
- Como realizar traduções simples e com variáveis
- Como usar pluralização
