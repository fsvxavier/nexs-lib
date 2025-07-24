# Basic Domain Errors Examples

Este exemplo demonstra o uso básico da biblioteca de erros de domínio.

## Executar o Exemplo

```bash
cd examples/basic
go run main.go
```

## Exemplos Incluídos

### 1. Criação Básica de Erro
- Como criar um erro de domínio simples
- Propriedades básicas do erro (code, message, type)

### 2. Erro com Causa
- Criando erro com uma causa subjacente
- Uso do método `Unwrap()` para acessar a causa raiz

### 3. Erro com Metadados
- Adicionando metadados personalizados aos erros
- Útil para informações de contexto (request_id, user_id, etc.)

### 4. Encadeamento de Erros
- Criando uma cadeia de erros com contexto
- Empilhamento de erros através de diferentes camadas
- Visualização do stack trace

### 5. Mapeamento de Status HTTP
- Como diferentes tipos de erro mapeiam para códigos HTTP
- Uso prático em APIs REST

### 6. Verificação de Tipos
- Verificando se um erro é de um tipo específico
- Usando `IsType()` para lógica condicional

### 7. Compatibilidade com Interface Padrão
- Demonstrando compatibilidade com `errors.Is` e `errors.As`
- Integração com o pacote `errors` padrão do Go

## Conceitos Importantes

### Tipos de Erro
- **Validation**: Erros de validação → HTTP 400
- **NotFound**: Recurso não encontrado → HTTP 404
- **Authentication**: Falha de autenticação → HTTP 401
- **Authorization**: Falha de autorização → HTTP 403
- **RateLimit**: Limite de taxa excedido → HTTP 429
- **Server**: Erro interno do servidor → HTTP 500

### Stack Trace
O stack trace é capturado automaticamente quando um erro é criado, fornecendo informações detalhadas sobre onde o erro ocorreu.

### Metadados
Use metadados para adicionar informações contextuais que podem ser úteis para debugging e logging.

## Próximos Passos

- Veja os exemplos avançados em `../advanced/`
- Explore os diferentes tipos de erro em `../types/`
- Aprenda sobre integração com APIs em `../api/`
