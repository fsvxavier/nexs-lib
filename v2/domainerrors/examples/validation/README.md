# Validation Examples

Este exemplo demonstra validação avançada e estruturada usando Domain Errors v2.

## 🎯 Objetivo

Mostrar estratégias completas de validação:
- Validação básica de campos
- Validação estruturada com regras complexas
- Regras de negócio
- Validação de estruturas aninhadas
- Validadores customizados
- Chaining de validações
- Validação contextual

## 🚀 Como Executar

```bash
go run main.go
```

## 📝 Funcionalidades Demonstradas

### 1. Validação Básica de Campos
```go
validationErr := domainerrors.NewValidationError("User registration failed", map[string][]string{
    "email":    {"invalid format", "domain not allowed"},
    "age":      {"must be 18 or older"},
    "password": {"too short", "missing special characters"},
})

// Adicionar erros dinamicamente
validationErr.AddField("username", "already taken")
validationErr.AddField("phone", "invalid country code")
```

### 2. Validação Estruturada
- Validator pattern com regex
- Validação de email, username, telefone
- Validação de senha com múltiplos critérios
- Validação de idade com ranges

### 3. Regras de Negócio
```go
func validateBusinessRules(user User) interfaces.DomainErrorInterface {
    err := domainerrors.NewBuilder().
        WithCode("BIZ001").
        WithMessage("Business rule validation failed").
        WithType(string(types.ErrorTypeBusinessRule))
    
    // Verificar domínios de email de concorrentes
    if strings.Contains(user.Email, "@competitor.com") {
        violations = append(violations, "competitor email domains not allowed")
    }
    
    // Verificar usernames reservados
    if isReservedUsername(user.Username) {
        violations = append(violations, "username is reserved")
    }
}
```

### 4. Validação Aninhada
- Validação de estruturas Profile
- Campos opcionais com validação condicional
- Prefixos de campo para nested structures

### 5. Validadores Customizados
- Email Domain Validator
- Password Strength Validator
- Age Range Validator
- Reserved Username Validator

### 6. Chaining de Validações
```go
validators := []func(User) error{
    validateEmailDomain,
    validatePasswordStrength,
    validateAgeRange,
    validateReservedUsername,
}

// Executar todos e combinar erros
var allErrors []error
for _, validator := range validators {
    if err := validator(user); err != nil {
        allErrors = append(allErrors, err)
    }
}
```

### 7. Validação Contextual
- Validação baseada em role (admin, user)
- Validação baseada em tipo de conta (premium, basic)
- Context-aware business rules

## 🔧 Estrutura do Código

### Modelos
- `User` - Modelo principal de usuário
- `Profile` - Perfil aninhado do usuário

### Validadores
- `UserValidator` - Validador principal com regex
- Custom validators por campo
- Business rule validators
- Context-aware validators

### Padrões Implementados
- **Strategy Pattern** - Diferentes estratégias de validação
- **Composite Pattern** - Combinação de múltiplos validadores
- **Chain of Responsibility** - Chain de validações
- **Factory Pattern** - Criação de erros específicos

## 📊 Tipos de Validação

### Validação Técnica
- Formato de email
- Força de senha
- Formato de telefone
- Range de idade

### Validação de Negócio
- Domínios bloqueados
- Usernames reservados
- Países restritos
- Regras de compliance

### Validação Contextual
- Permissões por role
- Restrições por tipo de conta
- Validação condicional

## ⚡ Performance

- Regex compilado uma vez
- Validação lazy quando possível
- Combinação eficiente de erros
- Memory-efficient error creation

## 📋 Próximos Passos

Veja outros exemplos:
- [Factory Usage](../factory-usage/) - Factories especializadas
- [Registry System](../registry-system/) - Códigos centralizados
- [Web Integration](../web-integration/) - Validação em APIs
