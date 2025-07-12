# Basic Usage Examples

Este exemplo demonstra o uso básico do sistema Domain Errors v2.

## 🎯 Objetivo

Mostrar as funcionalidades fundamentais:
- Criação simples de erros
- Builder pattern para construção fluente
- Erros de validação especializados
- Serialização JSON
- Diferentes tipos de erro

## 🚀 Como Executar

```bash
go run main.go
```

## 📝 Funcionalidades Demonstradas

### 1. Criação Básica de Erros
```go
err := domainerrors.New("E001", "User not found")
```

### 2. Builder Pattern
```go
err := domainerrors.NewBuilder().
    WithCode("USR001").
    WithMessage("User validation failed").
    WithType(string(types.ErrorTypeValidation)).
    WithDetail("user_id", "12345").
    WithTag("validation").
    Build()
```

### 3. Erros de Validação
```go
fields := map[string][]string{
    "email": {"invalid format", "required"},
    "age":   {"must be positive"},
}
validationErr := domainerrors.NewValidationError("Validation failed", fields)
```

### 4. Serialização JSON
Todos os erros suportam serialização/deserialização JSON automaticamente.

### 5. Tipos de Erro
- NotFound
- Validation
- BusinessRule
- Authentication
- Authorization

## 🔧 Estrutura do Código

- `basicErrorCreation()` - Criação simples
- `builderPatternExample()` - Builder pattern
- `validationErrorExample()` - Erros de validação
- `jsonSerializationExample()` - JSON serialization
- `errorTypesExample()` - Diferentes tipos

## 📊 Próximos Passos

Veja outros exemplos mais avançados:
- [Builder Pattern](../builder-pattern/) - Construção avançada
- [Error Stacking](../error-stacking/) - Empilhamento de erros
- [Validation](../validation/) - Validação avançada
