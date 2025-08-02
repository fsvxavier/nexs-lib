# Exemplo Básico - Sistema de Middlewares

Este exemplo demonstra o uso básico do sistema de middlewares da biblioteca `nexs-lib/httpserver`.

## 📋 O que este exemplo demonstra

- **Middleware de Logging**: Registra todas as requisições e respostas HTTP
- **Middleware de Autenticação**: Implementa Basic Auth para proteger rotas
- **Middleware Manager**: Gerencia a cadeia de middlewares com prioridades
- **Rotas Públicas e Protegidas**: Diferentes níveis de acesso

## 🚀 Como executar

```bash
cd httpserver/examples/middlewares-basic
go run main.go
```

O servidor iniciará na porta 8080.

## 🧪 Testando os endpoints

### Rotas Públicas (sem autenticação)

```bash
# Página inicial
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Rota pública
curl http://localhost:8080/public

# Informações do sistema
curl http://localhost:8080/info
```

### Rotas Protegidas (com Basic Auth)

```bash
# Lista de usuários (admin)
curl -u admin:secret123 http://localhost:8080/api/users

# Lista de usuários (user)
curl -u user:password456 http://localhost:8080/api/users

# Perfil do usuário
curl -u admin:secret123 http://localhost:8080/api/profile

# Área administrativa
curl -u admin:secret123 http://localhost:8080/api/admin

# Criar novo usuário (POST)
curl -u admin:secret123 -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Novo Usuario","email":"novo@example.com"}' \
  http://localhost:8080/api/users
```

### Testando acesso negado

```bash
# Tentar acessar rota protegida sem autenticação
curl http://localhost:8080/api/users

# Tentar com credenciais inválidas
curl -u admin:wrongpassword http://localhost:8080/api/users
```

## 📊 Funcionalidades demonstradas

### 1. Middleware de Logging
- **Configuração**:
  - Registra requisições e respostas
  - Registra headers (mas não dados sensíveis)
  - Pula rotas de health check
  - Limita tamanho do body logado
- **Funcionalidades**:
  - Log detalhado de cada requisição
  - Formatação estruturada
  - Filtros por path e método HTTP

### 2. Middleware de Autenticação
- **Configuração**:
  - Basic Auth habilitado
  - Dois usuários: `admin` e `user`
  - Rotas públicas configuradas
- **Funcionalidades**:
  - Validação de credenciais
  - Proteção de rotas específicas
  - Mensagens de erro customizadas

### 3. Middleware Manager
- **Características**:
  - Execução em ordem de prioridade
  - Processamento em cadeia
  - Gerenciamento centralizado
- **Benefícios**:
  - Fácil adição/remoção de middlewares
  - Controle fino sobre execução
  - Isolamento de responsabilidades

## 🔐 Credenciais de teste

| Usuário | Senha | Acesso |
|---------|-------|--------|
| admin   | secret123 | Todas as rotas protegidas |
| user    | password456 | Todas as rotas protegidas |

## 📁 Estrutura de rotas

```
/                    - Página inicial (pública)
/health              - Health check (pública)
/public              - Exemplo de rota pública
/info                - Informações do sistema (pública)
/api/
  ├── users          - CRUD de usuários (protegida)
  ├── profile        - Perfil do usuário (protegida)
  └── admin          - Área administrativa (protegida)
```

## 🔍 Logs produzidos

Durante a execução, você verá logs detalhados mostrando:
- Configuração dos middlewares
- Cada requisição HTTP interceptada
- Tentativas de autenticação
- Processamento das rotas
- Headers e metadados das requisições

## 📖 Conceitos importantes

### Middleware Manager
O `MiddlewareManager` é responsável por:
- Registrar middlewares em ordem de prioridade
- Executar a cadeia de middlewares
- Gerenciar falhas e propagação de erros
- Fornecer interface unificada

### Configuração de Middlewares
Cada middleware pode ser configurado com:
- **Prioridade**: Define ordem de execução
- **Filtros**: Quais rotas/métodos processar
- **Comportamento**: Como processar requisições
- **Segurança**: Configurações específicas

### Thread Safety
Todos os middlewares são thread-safe e podem processar requisições concorrentes.

## 🎯 Próximos passos

Após entender este exemplo básico, você pode explorar:
- [Exemplo com Hooks](../hooks-basic/) - Adiciona monitoramento e métricas
- [Exemplo Completo](../complete/) - Combina hooks e middlewares
- [Exemplos específicos por framework](../) - Gin, Echo, FastHTTP, etc.

## 🛠️ Personalizações

Você pode facilmente:
- Adicionar novos middlewares (CORS, Rate Limiting, etc.)
- Modificar configurações de autenticação
- Customizar formato de logs
- Implementar diferentes tipos de autenticação (JWT, API Keys, etc.)
