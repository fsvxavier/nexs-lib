# Exemplo de Formatos

Este exemplo demonstra o uso de diferentes formatos de arquivo de tradução suportados pelo pacote i18n:
- JSON
- YAML

## Estrutura

```
formats/
├── main.go
├── translations/
│   ├── translations.pt-BR.json
│   └── translations.pt-BR.yaml
└── README.md
```

## Execução

```bash
go run main.go
```

## Arquivos de Tradução

### JSON (`translations.pt-BR.json`)

```json
{
    "welcome": "Bem-vindo ao exemplo de formatos!",
    "greeting": "Olá, {{.Name}}!",
    "items": {
        "one": "{{.Count}} item",
        "other": "{{.Count}} itens"
    }
}
```

### YAML (`translations.pt-BR.yaml`)

```yaml
welcome: "Bem-vindo ao exemplo de formatos!"
greeting: "Olá, {{.Name}}!"
items:
  one: "{{.Count}} item"
  other: "{{.Count}} itens"
```

## Funcionalidades Demonstradas

1. Configuração para diferentes formatos
2. Criação de providers específicos
3. Carregamento de traduções em JSON
4. Carregamento de traduções em YAML
5. Comparação entre formatos

## Código

O exemplo demonstra:
- Como configurar o provider para diferentes formatos
- Como carregar arquivos JSON e YAML
- Como usar o mesmo conjunto de traduções em diferentes formatos
- Como escolher o formato mais adequado para seu caso de uso
