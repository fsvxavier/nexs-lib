# Exemplos de Uso dos Providers Decimal

Este diretório contém exemplos de uso dos providers decimal implementados na biblioteca.

## Estrutura dos Exemplos

- `example.go`: Exemplos básicos de uso direto dos providers (ShopSpring e APD)
- `apiexample/main.go`: Exemplos de uso da API com factory e helpers
- `comprehensive_example.go`: Exemplos abrangentes cobrindo todos os aspectos dos providers decimal

## Como executar os exemplos

Para executar os exemplos, use os seguintes comandos a partir da raiz do projeto:

```bash
# Exemplo básico
go run dec/examples/example.go

# Exemplo da API com factory
go run dec/examples/apiexample/main.go

# Exemplo abrangente
go run dec/examples/comprehensive_example.go
```

## Características demonstradas nos exemplos

1. **Criação de decimais**
   - A partir de string
   - A partir de float64
   - A partir de int64

2. **Operações aritméticas**
   - Adição
   - Subtração
   - Multiplicação
   - Divisão

3. **Operações de comparação**
   - Igualdade
   - Maior que
   - Menor que
   - Maior ou igual a
   - Menor ou igual a

4. **Formatação e arredondamento**
   - Arredondamento para um número específico de casas decimais
   - Truncamento para um número específico de casas decimais

5. **Uso da interface genérica**
   - Criação de código independente da implementação

6. **Uso da factory**
   - Criação de providers a partir de um tipo
   - Uso intercambiável de diferentes providers

7. **Serialização JSON**
   - Conversão de decimais para JSON
   - Conversão de JSON para decimais

8. **Tratamento de erros**
   - Tratamento de erros de conversão
   - Tratamento de divisão por zero
