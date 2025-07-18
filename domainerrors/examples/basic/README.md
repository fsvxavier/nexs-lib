# Exemplo Básico - DomainErrors

Este exemplo demonstra o uso básico do módulo domainerrors, incluindo:

## Funcionalidades Demonstradas

### 1. Criação de Erros Básicos
- `domainerrors.New()` - Erro básico com código e mensagem
- `domainerrors.NewWithError()` - Erro com causa subjacente
- `domainerrors.NewWithType()` - Erro com tipo específico

### 2. Tipos de Erro Específicos
- **ValidationError**: Erros de validação com campos específicos
- **NotFoundError**: Recursos não encontrados
- **BusinessError**: Violações de regras de negócio
- **DatabaseError**: Falhas de banco de dados
- **ExternalServiceError**: Falhas em serviços externos

### 3. Metadados e Contexto
- Adicionar metadados com `WithMetadata()`
- Verificar tipos de erro com `IsType()`
- Serialização JSON com `JSON()`

### 4. Empilhamento de Erros
- Encadear erros com `Wrap()`
- Formatação de cadeia de erros
- Identificação de causa raiz

### 5. Grupo de Erros
- Coletar múltiplos erros
- Filtrar por tipo
- Operações em lote

### 6. Utilitários
- Verificação de severidade
- Lógica de retry
- Mapeamento HTTP

## Como Executar

```bash
cd domainerrors/examples/basic
go run main.go
```

## Saída Esperada

O exemplo produzirá uma saída detalhada mostrando:
- Diferentes tipos de erro criados
- Metadados e contexto adicionados
- Serialização JSON
- Empilhamento e formatação de erros
- Operações com grupos de erros
- Mapeamento para códigos HTTP

## Conceitos Importantes

### Códigos de Erro
- Use códigos únicos e descritivos (ex: "USER_001", "DB_001")
- Siga um padrão consistente na aplicação

### Tipos de Erro
- Escolha o tipo apropriado para cada situação
- Tipos determinam o comportamento e mapeamento HTTP

### Metadados
- Adicione contexto relevante para debugging
- Evite informações sensíveis nos metadados

### Empilhamento
- Preserve a causa original do erro
- Use `Wrap()` para adicionar contexto sem perder informações

## Próximos Passos

Após dominar este exemplo básico, veja:
- `examples/advanced/` - Padrões avançados e integração
- `examples/global/` - Configuração global e middleware
