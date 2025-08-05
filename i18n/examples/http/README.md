# Exemplo HTTP

Este exemplo demonstra a integração do pacote i18n com um servidor HTTP usando middleware:
- Detecção automática de idioma via cabeçalhos HTTP
- Middleware de internacionalização
- Traduções em diferentes endpoints

## Estrutura

```
http/
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
    "welcome": "Bem-vindo ao exemplo HTTP!",
    "greeting": "Olá, {{.Name}}!",
    "items": {
        "one": "{{.Count}} item",
        "other": "{{.Count}} itens"
    }
}
```

## Endpoints

1. `GET /`
   - Retorna uma mensagem de boas-vindas traduzida
   ```bash
   curl -H "Accept-Language: pt-BR" http://localhost:3000/
   ```

2. `GET /with-vars`
   - Retorna uma saudação personalizada com variáveis
   ```bash
   curl -H "Accept-Language: pt-BR" http://localhost:3000/with-vars
   ```

3. `GET /plural`
   - Demonstra pluralização
   ```bash
   curl -H "Accept-Language: pt-BR" http://localhost:3000/plural
   ```

## Funcionalidades Demonstradas

1. Configuração do middleware HTTP
2. Detecção automática de idioma
3. Tradução em handlers HTTP
4. Uso de variáveis em traduções
5. Pluralização em respostas HTTP

## Código

O exemplo demonstra:
- Como configurar o middleware i18n
- Como integrar com servidor HTTP padrão
- Como usar traduções em handlers
- Como detectar o idioma do usuário
- Como responder com conteúdo traduzido

## Testes

Para testar os endpoints:

```bash
# Mensagem de boas-vindas
curl -H "Accept-Language: pt-BR" http://localhost:3000/

# Saudação com nome
curl -H "Accept-Language: pt-BR" http://localhost:3000/with-vars

# Pluralização
curl -H "Accept-Language: pt-BR" http://localhost:3000/plural
```
