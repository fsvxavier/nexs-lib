# Exemplo Básico de Domain Errors

Este exemplo demonstra o uso básico do módulo `domainerrors`, incluindo:

## Funcionalidades Demonstradas

### 1. Criação de Erros Básicos
- Erro de validação (`ValidationError`)
- Erro de não encontrado (`NotFoundError`)  
- Erro de negócio (`BusinessError`)

### 2. Metadados
- Adição de metadados contextuais aos erros
- Recuperação e manipulação de metadados

### 3. Encapsulamento de Erros
- Wrapping de erros existentes
- Preservação da causa raiz

### 4. Verificação de Tipos
- Verificação de tipos de erro usando `IsType`
- Compatibilidade com interface padrão Go

### 5. Análise de Cadeia de Erros
- Formatação de cadeia de erros
- Recuperação da causa raiz
- Navegação em hierarquia de erros

### 6. Serialização JSON
- Conversão de erros para JSON
- Preservação de metadados na serialização

### 7. Context Integration
- Integração com `context.Context`
- Preservação de informações contextuais

### 8. Factory Personalizada
- Uso de factory para criação de erros
- Personalização de comportamento

## Como Executar

```bash
cd examples/basic
go run main.go
```

## Saída Esperada

O exemplo produzirá uma saída detalhada mostrando:
- Criação e manipulação de diferentes tipos de erro
- Códigos HTTP correspondentes
- Formatação de cadeias de erro
- Serialização JSON com metadados
- Timestamps e informações contextuais

## Principais Conceitos

### Tipos de Erro Suportados
- `ValidationError` → HTTP 400
- `NotFoundError` → HTTP 404  
- `BusinessError` → HTTP 422
- `DatabaseError` → HTTP 500
- E muitos outros...

### Metadados Contextuais
Os erros podem carregar informações adicionais como:
- Campo que falhou na validação
- Valor que causou o erro
- Regra de negócio violada
- IDs de request/usuário

### Encapsulamento Seguro
- Preserva a causa original
- Adiciona contexto de domínio
- Mantém rastreabilidade

Este exemplo serve como ponto de partida para entender as capacidades básicas do módulo de tratamento de erros de domínio.
