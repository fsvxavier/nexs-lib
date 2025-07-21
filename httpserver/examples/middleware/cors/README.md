# CORS Middleware Example

Este exemplo demonstra o uso completo do middleware CORS (Cross-Origin Resource Sharing) da nexs-lib.

## 🌐 Sobre CORS

CORS é um mecanismo de segurança que permite ou restringe recursos web a serem acessados de outro domínio diferente do que está servindo o recurso.

### Por que CORS é Importante?
- **Segurança**: Previne ataques de cross-site scripting
- **Controle**: Define quais origens podem acessar sua API
- **Flexibilidade**: Permite configuração granular de permissões

## 🚀 Executando o Exemplo

```bash
go run main.go
```

O servidor iniciará na porta `:8080` com diferentes políticas de CORS.

## 📍 Endpoints

| Endpoint | CORS Policy | Características |
|----------|-------------|----------------|
| `GET /health` | Sem CORS | Endpoint interno |
| `GET /api/test` | Restritivo | Origens específicas |
| `POST /api/test` | Restritivo | Com credenciais |
| `GET /public` | Aberto | Todas as origens |

## 🔧 Configurações Implementadas

### 1. CORS Restritivo (API Endpoints)
```go
corsConfig := cors.Config{
    Enabled: true,
    SkipPaths: []string{"/health"},
    AllowedOrigins: []string{
        "http://localhost:3000",
        "http://localhost:8000", 
        "https://mydomain.com",
    },
    AllowedMethods: []string{
        "GET", "POST", "PUT", "DELETE", "OPTIONS",
    },
    AllowedHeaders: []string{
        "Content-Type",
        "Authorization", 
        "X-Requested-With",
        "X-User-ID",
    },
    ExposedHeaders: []string{
        "X-Total-Count",
        "X-Rate-Limit-Remaining",
    },
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}
```

### 2. CORS Aberto (Public Endpoint)
```go
publicCorsConfig := cors.Config{
    Enabled:        true,
    AllowedOrigins: []string{"*"}, // Todas as origens
    AllowedMethods: []string{"GET", "POST"},
    AllowedHeaders: []string{"Content-Type"},
    AllowCredentials: false, // Não pode usar com wildcard
    MaxAge: 1 * time.Hour,
}
```

## 🧪 Testando

### Teste Básico (Origem Permitida)
```bash
# Teste com origem permitida
curl -H 'Origin: http://localhost:3000' \
     -H 'Content-Type: application/json' \
     http://localhost:8080/api/test
```

**Headers de Resposta Esperados:**
```
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Credentials: true
Access-Control-Expose-Headers: X-Total-Count, X-Rate-Limit-Remaining
Vary: Origin
```

### Teste com Origem Não Permitida
```bash
# Teste com origem não autorizada
curl -H 'Origin: https://malicious.com' \
     http://localhost:8080/api/test
```

**Resultado:** Request será bloqueado (sem headers CORS)

### Teste Preflight Request
```bash
# OPTIONS request (preflight)
curl -X OPTIONS \
     -H 'Origin: http://localhost:3000' \
     -H 'Access-Control-Request-Method: POST' \
     -H 'Access-Control-Request-Headers: Content-Type, Authorization' \
     -v http://localhost:8080/api/test
```

**Headers de Resposta Esperados:**
```
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With, X-User-ID
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 43200
```

### Teste Endpoint Público
```bash
# Qualquer origem é permitida
curl -H 'Origin: https://qualquer-site.com' \
     http://localhost:8080/public
```

**Headers de Resposta:**
```
Access-Control-Allow-Origin: *
```

## 🔍 Cenários de Teste

### 1. Aplicação Frontend Local
```bash
# Simula Next.js rodando em localhost:3000
curl -H 'Origin: http://localhost:3000' \
     -H 'Content-Type: application/json' \
     -X POST \
     -d '{"name": "test"}' \
     http://localhost:8080/api/test
```

### 2. Aplicação Mobile/Desktop
```bash
# Apps nativos geralmente não enviam Origin
curl -H 'Content-Type: application/json' \
     -X GET \
     http://localhost:8080/api/test
```

### 3. Teste de Credenciais
```bash
# Com cookie/auth header
curl -H 'Origin: http://localhost:3000' \
     -H 'Authorization: Bearer token123' \
     -H 'Cookie: session=abc123' \
     http://localhost:8080/api/test
```

## ⚙️ Configurações Detalhadas

### AllowedOrigins
```go
// Origens específicas
AllowedOrigins: []string{
    "https://app.exemplo.com",
    "https://admin.exemplo.com",
}

// Todas as origens (cuidado!)
AllowedOrigins: []string{"*"}

// Origens com porta específica
AllowedOrigins: []string{
    "http://localhost:3000",  // React dev
    "http://localhost:8080",  // Vue dev
}
```

### AllowedMethods
```go
// REST API completa
AllowedMethods: []string{
    "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"
}

// Somente leitura
AllowedMethods: []string{"GET", "HEAD", "OPTIONS"}
```

### AllowedHeaders
```go
// Headers comuns
AllowedHeaders: []string{
    "Content-Type",
    "Authorization",
    "X-Requested-With",
    "Accept",
    "X-CSRF-Token",
}

// Headers customizados
AllowedHeaders: []string{
    "X-API-Key",
    "X-Client-Version",
    "X-Request-ID",
}
```

### ExposedHeaders
```go
// Headers que o cliente pode acessar
ExposedHeaders: []string{
    "X-Total-Count",      // Paginação
    "X-Rate-Limit-Remaining", // Rate limiting
    "X-Request-ID",       // Tracking
    "Link",               // Paginação (RFC 5988)
}
```

## 🛡️ Segurança

### Boas Práticas

#### ✅ Recomendado
```go
// Origens específicas em produção
AllowedOrigins: []string{
    "https://meuapp.com",
    "https://app.meuapp.com",
}

// Headers mínimos necessários
AllowedHeaders: []string{
    "Content-Type",
    "Authorization",
}

// Credenciais apenas quando necessário
AllowCredentials: true // Apenas com origens específicas
```

#### ❌ Evitar
```go
// Wildcard com credenciais (impossível)
AllowedOrigins: []string{"*"}
AllowCredentials: true // ERRO!

// Headers muito permissivos
AllowedHeaders: []string{"*"} // Perigoso

// Cache muito longo em desenvolvimento
MaxAge: 24 * time.Hour // Use valores menores em dev
```

### Cenários de Segurança

#### Ambiente de Desenvolvimento
```go
corsConfig := cors.Config{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"*"},
    AllowedHeaders: []string{"*"},
    AllowCredentials: false,
    MaxAge: 1 * time.Hour, // Cache curto
}
```

#### Ambiente de Produção
```go
corsConfig := cors.Config{
    AllowedOrigins: []string{
        "https://app.example.com",
        "https://admin.example.com",
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{
        "Content-Type",
        "Authorization",
    },
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}
```

## 🔧 Troubleshooting

### Problemas Comuns

#### CORS Error no Browser
```
Access to fetch at 'http://localhost:8080/api/test' from origin 
'http://localhost:3000' has been blocked by CORS policy
```

**Soluções:**
1. Adicionar origem à `AllowedOrigins`
2. Verificar se método está em `AllowedMethods`
3. Verificar headers em `AllowedHeaders`

#### Preflight Request Falha
```bash
# Debug preflight
curl -X OPTIONS \
     -H 'Origin: http://localhost:3000' \
     -H 'Access-Control-Request-Method: POST' \
     -v http://localhost:8080/api/test
```

**Verificar:**
- Status 200 para OPTIONS
- Headers `Access-Control-Allow-*` presentes
- `MaxAge` configurado corretamente

#### Credentials Not Allowed
```
The value of the 'Access-Control-Allow-Credentials' header in the response 
is '' which must be 'true' when the request's credentials mode is 'include'
```

**Solução:**
```go
AllowCredentials: true
// E origens específicas (não "*")
```

### Debug Headers

Use este comando para ver todos os headers CORS:
```bash
curl -X OPTIONS \
     -H 'Origin: http://localhost:3000' \
     -H 'Access-Control-Request-Method: POST' \
     -H 'Access-Control-Request-Headers: Content-Type' \
     -v http://localhost:8080/api/test 2>&1 | grep -i "access-control\|origin"
```

## 📱 Integração com Frameworks

### React/Next.js
```javascript
// fetch com credenciais
fetch('http://localhost:8080/api/test', {
    method: 'POST',
    credentials: 'include', // Envia cookies
    headers: {
        'Content-Type': 'application/json',
        'Origin': 'http://localhost:3000'
    },
    body: JSON.stringify({data: 'test'})
})
```

### Vue.js
```javascript
// axios com CORS
axios.defaults.withCredentials = true;
axios.post('http://localhost:8080/api/test', data, {
    headers: {
        'Content-Type': 'application/json'
    }
});
```

### Angular
```typescript
// HTTP client com CORS
this.http.post('http://localhost:8080/api/test', data, {
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json'
    }
});
```

## 📚 Referências

- [MDN CORS Documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [W3C CORS Specification](https://www.w3.org/TR/cors/)
- [CORS Best Practices](https://web.dev/cross-origin-resource-sharing/)
