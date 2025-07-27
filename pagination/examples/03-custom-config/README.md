# Exemplo 3: Configuração Personalizada

Este exemplo demonstra como personalizar completamente o comportamento do módulo de paginação através de:

- Configurações personalizadas para diferentes contextos
- Providers customizados com lógica específica
- Validações de negócio personalizadas
- Múltiplos perfis de configuração
- Validação automática de configurações

## Como executar

```bash
cd examples/03-custom-config
go run main.go
```

## O que o exemplo demonstra

### 1. 🎛️ Configuração Básica Personalizada
- Alteração de limites padrão e máximo
- Campo de ordenação personalizado
- Ordens de classificação customizadas
- Modo rigoroso vs. flexível

### 2. 🔧 Providers Personalizados

#### CustomRequestParser
- Aceita parâmetros alternativos: `p` (page), `size` (limit)
- Usa `order_by` e `direction` para ordenação
- Mantém compatibilidade com validação padrão

#### CustomValidator  
- Adiciona regras de negócio específicas
- Limita páginas muito altas (proteção de performance)
- Regras específicas por campo de ordenação
- Integra com validação padrão

### 3. 🧪 Regras de Validação Personalizadas
- Limite de página máxima (1000) para performance
- Restrições específicas por campo (ex: preço max 20 registros)
- Validação em cascata com regras padrão

### 4. 📱 Configurações por Contexto

#### Mobile
```go
{
    DefaultLimit: 10,
    MaxLimit: 25,
    DefaultSortOrder: "desc",
    AllowedSortOrders: []string{"desc"}, // Apenas decrescente
    StrictMode: true,
}
```

#### Web
```go
{
    DefaultLimit: 50,
    MaxLimit: 200,
    AllowedSortOrders: []string{"asc", "desc"},
    StrictMode: false,
}
```

#### API Interna
```go
{
    DefaultLimit: 100,
    MaxLimit: 1000,
    AllowedSortOrders: []string{"asc", "desc", "random"},
    ValidationEnabled: false, // Sem validação
}
```

### 5. ✅ Validação Automática de Configuração
- Correção automática de valores inválidos
- Validação de consistência entre limites
- Garantia de funcionamento mesmo com configuração problemática

## Saída do Exemplo

```
🎛️  Exemplos de Configuração Personalizada - Módulo de Paginação
================================================================

=== 1. Demonstração de Configuração Personalizada ===
📋 Configuração personalizada aplicada:
  Limit padrão: 25
  Campo de ordenação padrão: created_at
  Ordem padrão: desc
  Limit máximo: 200

=== 2. Demonstração de Providers Personalizados ===
🔧 Parâmetros de entrada personalizados: p=3&size=15&order_by=name&direction=desc
🔍 Parsing página personalizada: 3
🔍 Parsing tamanho personalizado: 15
🔍 Parsing campo de ordenação personalizado: name
🔍 Parsing direção personalizada: desc
✅ Validando parâmetros personalizados
✅ Validação personalizada passou
📊 Resultado do parsing personalizado:
  Página: 3
  Limite: 15
  Campo de ordenação: name
  Ordem: desc

=== 3. Demonstração de Regras de Validação Personalizadas ===
🧪 Teste 1: Página muito alta (1001)
❌ Erro esperado: [CUSTOM_PAGE_LIMIT] página não pode exceder 1000

🧪 Teste 2: Limite alto (25) com ordenação por preço
❌ Erro esperado: [CUSTOM_PRICE_LIMIT] quando ordenando por preço, limite máximo é 20

🧪 Teste 3: Parâmetros válidos
✅ Validação passou: página 5, limite 15, ordenação name
```

## Casos de Uso Práticos

### 1. 📱 Aplicação Mobile
- Limites menores para economizar dados
- Apenas ordenação decrescente (conteúdo mais recente)
- Validação rigorosa para UX consistente

### 2. 🌐 Interface Web
- Limites maiores para melhor UX
- Flexibilidade na ordenação
- Modo menos rigoroso para admin users

### 3. 🔌 API Interna
- Limites muito altos para processamento em lote
- Sem validação para performance máxima
- Suporte a ordenação aleatória

### 4. 🏢 Sistema Empresarial
- Validações de negócio específicas
- Controle de acesso por perfil
- Auditoria e logs detalhados

## Conceitos Demonstrados

- ✅ **Configuração Multi-Contexto** - Diferentes perfis para diferentes necessidades
- ✅ **Providers Pluggáveis** - Substituição de componentes específicos
- ✅ **Validação Customizada** - Regras de negócio personalizadas
- ✅ **Parsing Flexível** - Suporte a diferentes formatos de parâmetros
- ✅ **Auto-Correção** - Configurações inválidas são corrigidas automaticamente
- ✅ **Compatibilidade** - Integração com providers padrão
- ✅ **Performance** - Otimizações específicas por contexto

## Arquitetura de Providers

```
PaginationService
├── RequestParser (Customizável)
│   ├── StandardRequestParser
│   └── CustomRequestParser ✨
├── Validator (Customizável)  
│   ├── StandardValidator
│   └── CustomValidator ✨
├── QueryBuilder (Padrão)
└── PaginationCalculator (Padrão)
```

## Próximos Passos

Após entender configurações personalizadas, veja:
- `04-error-handling` - Tratamento avançado de erros
- `05-database-integration` - Integração com banco real
- `06-performance-optimization` - Otimizações de performance
