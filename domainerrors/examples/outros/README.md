# Outros Casos de Uso - Domain Errors

Este exemplo demonstra casos de uso práticos e integração do sistema de domainerrors em cenários reais de aplicação.

## Funcionalidades Demonstradas

### 1. Validação de Formulário Complexo
Demonstra validação abrangente de dados de entrada com múltiplas regras:
- **Campos obrigatórios**: Nome não pode estar vazio
- **Formato de email**: Validação básica de email
- **Regras de idade**: Idade mínima de 18 anos
- **Segurança de senha**: Senha com pelo menos 8 caracteres
- **Autorização de roles**: Roles permitidas limitadas

```go
type ValidationResult struct {
    Valid  bool                   `json:"valid"`
    Errors []ValidationErrorItem  `json:"errors,omitempty"`
    Fields map[string]interface{} `json:"fields,omitempty"`
}
```

### 2. Processamento de Transação Bancária
Sistema completo de validação de transações financeiras:
- **Verificação de contas**: Origem e destino devem existir
- **Status de conta**: Contas devem estar ativas
- **Validação de saldo**: Saldo suficiente para transação
- **Limites diários**: Controle de limite por transação
- **Rastreabilidade**: IDs únicos para auditoria

### 3. API REST com Tratamento de Erros
Simulação de endpoints REST com diferentes cenários:
- **GET /users/{id}**: Busca de usuário (sucesso/não encontrado)
- **POST /users**: Criação com validação
- **PUT /users/{id}**: Atualização com autorização

Resposta padronizada de erro:
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "Usuário não encontrado",
    "type": "not_found_error",
    "details": {"user_id": "999"}
  },
  "request_id": "req-abc123",
  "timestamp": "2024-12-14T20:30:00Z"
}
```

### 4. Sistema de Autenticação
Múltiplas formas de autenticação e seus erros:
- **Credenciais**: Usuário/senha
- **Token JWT**: Validação e expiração
- **Status do usuário**: Bloqueios e permissões

Cenários de erro:
- Credenciais inválidas
- Token expirado
- Token inválido
- Usuário bloqueado
- Campos obrigatórios ausentes

### 5. Integração com Serviços Externos
Simulação de falhas em serviços externos:
- **Payment Gateway**: Timeout de conexão
- **Email Service**: Rate limiting
- **User Service**: Serviço indisponível
- **Cache Service**: Operação bem-sucedida

Inclui sugestões de recuperação baseadas no tipo de erro:
- Timeout → Retry com backoff
- Rate limit → Aguardar período
- Serviço indisponível → Circuit breaker

### 6. Sistema de Cache com Fallback
Estratégia de cache com recuperação automática:
- **Cache Hit**: Dados encontrados no cache
- **Cache Miss**: Busca na fonte original
- **Cache Expired**: Dados expirados, refazer busca
- **Fallback**: Recuperação quando cache falha

## Estruturas de Dados

### User (Usuário)
```go
type User struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Age    int    `json:"age"`
    Role   string `json:"role"`
    Status string `json:"status"`
}
```

### BankAccount (Conta Bancária)
```go
type BankAccount struct {
    ID      string  `json:"id"`
    UserID  string  `json:"user_id"`
    Balance float64 `json:"balance"`
    Status  string  `json:"status"`
    Type    string  `json:"type"`
}
```

### Transaction (Transação)
```go
type Transaction struct {
    ID          string    `json:"id"`
    FromAccount string    `json:"from_account"`
    ToAccount   string    `json:"to_account"`
    Amount      float64   `json:"amount"`
    Type        string    `json:"type"`
    Status      string    `json:"status"`
    Timestamp   time.Time `json:"timestamp"`
}
```

## Como Executar

```bash
cd examples/outros
go run main.go
```

Ou compile primeiro:

```bash
go build -o outros-example main.go
./outros-example
```

## Cenários Testados

### 1. Validação de Formulário
- Nome vazio (REQUIRED_FIELD)
- Email inválido (INVALID_EMAIL)
- Idade menor que 18 (INVALID_AGE)
- Senha fraca (WEAK_PASSWORD)
- Role não permitida (INVALID_ROLE)

### 2. Transações Bancárias
- ✅ Transação válida (R$ 200,00)
- ❌ Saldo insuficiente (R$ 1.500,00)
- ❌ Conta congelada
- ❌ Conta de destino não existe

### 3. API REST
- ✅ GET /users/123 → 200 OK
- ❌ GET /users/999 → 404 Not Found
- ❌ POST /users → 400 Bad Request (validação)
- ❌ PUT /users/123 → 403 Forbidden (autorização)

### 4. Autenticação
- ✅ admin/admin123 → Sucesso
- ❌ user/wrongpass → Credenciais inválidas
- ❌ invalid-jwt-token → Token inválido
- ❌ blocked_user → Usuário bloqueado
- ❌ expired-jwt-token → Token expirado

### 5. Serviços Externos
- ❌ Payment Gateway → Timeout
- ❌ Email Service → Rate limit
- ❌ User Service → Indisponível
- ✅ Cache Service → Sucesso

### 6. Sistema de Cache
- ✅ user:123 → Cache hit
- ❌ user:456 → Cache miss, fallback sucesso
- ❌ user:999 → Cache miss, fallback falhou
- ✅ config:app → Cache hit
- ❌ temp:session → Expirado, fallback sucesso

## Patterns e Práticas

### Error Mapping HTTP
Cada tipo de domain error é mapeado para status HTTP apropriado:
- `ValidationError` → 400 Bad Request
- `NotFoundError` → 404 Not Found
- `AuthenticationError` → 401 Unauthorized
- `AuthorizationError` → 403 Forbidden
- `BusinessError` → 422 Unprocessable Entity

### Context Enrichment
Erros são enriquecidos com contexto relevante:
- IDs de recursos
- Valores atuais vs esperados
- Metadados de rastreabilidade
- Timestamps de operação

### Fallback Strategies
Implementação de estratégias de recuperação:
- Cache miss → Database lookup
- Service timeout → Retry with backoff
- Rate limit → Queue for later
- Service unavailable → Circuit breaker

### Security Considerations
- Senhas são mascaradas em logs
- Informações sensíveis não vazam em erros
- Tokens são validados adequadamente
- Usuários bloqueados são identificados

## Casos de Uso Reais

Este exemplo pode ser adaptado para:
- **E-commerce**: Validação de pedidos e pagamentos
- **Banking**: Processamento de transações
- **Healthcare**: Validação de dados médicos
- **Education**: Sistema de notas e matrícula
- **SaaS**: Autenticação e autorização
- **IoT**: Processamento de dados de sensores
