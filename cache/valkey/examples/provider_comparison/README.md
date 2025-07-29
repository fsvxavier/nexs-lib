# Exemplo: Comparação entre Providers

Este exemplo demonstra o uso e comparação entre os providers `valkey-go` e `valkey-glide`, mostrando sua compatibilidade e diferenças de performance.

## Requisitos

1. **Servidor Valkey** rodando em `localhost:6379`
2. **Go 1.21+**
3. **Sem autenticação** no servidor (ou configure credenciais no código)

## Executando o Exemplo

```bash
# A partir do diretório raiz do projeto
cd cache/valkey/examples/provider_comparison

# Executar
go run main.go
```

## Preparando o Ambiente

### Opção 1: Docker (Recomendado)
```bash
# Rodar Valkey sem autenticação
docker run -d -p 6379:6379 valkey/valkey:7.2-alpine

# Verificar se está rodando
docker ps
```

### Opção 2: Instalação Local
```bash
# No Ubuntu/Debian
sudo apt update && sudo apt install valkey-server
sudo systemctl start valkey-server

# No macOS
brew install valkey
brew services start valkey
```

## Saída Esperada

```
=== Demonstração de Compatibilidade entre Providers ===

--- Testando Provider: valkey-go ---
✅ Conectado com sucesso usando valkey-go
  📝 Testando operações básicas (valkey-go)...
    ✅ TTL definido: 9.999s
    ✅ Operações básicas OK (1 chaves deletadas)
  🗂️ Testando operações de Hash (valkey-go)...
    ✅ Operações de Hash OK (2 campos)
  📝 Testando operações de Lista (valkey-go)...
    ✅ Operações de Lista OK (último item: item2)
  🎲 Testando operações de Set (valkey-go)...
    ✅ Operações de Set OK (3 membros adicionados)
  🏆 Testando operações de Sorted Set (valkey-go)...
    ✅ Operações de Sorted Set OK (3 membros adicionados)
  ⚡ Medindo performance básica (valkey-go)...
    📊 Performance (100 ops):
       SET: 2500 ops/sec (40.00ms total)
       GET: 3333 ops/sec (30.00ms total)

--- Testando Provider: valkey-glide ---
✅ Conectado com sucesso usando valkey-glide
  📝 Testando operações básicas (valkey-glide)...
    ✅ TTL definido: 9.999s
    ✅ Operações básicas OK (1 chaves deletadas)
  🗂️ Testando operações de Hash (valkey-glide)...
    ✅ Operações de Hash OK (2 campos)
  📝 Testando operações de Lista (valkey-glide)...
    ✅ Operações de Lista OK (último item: item2)
  🎲 Testando operações de Set (valkey-glide)...
    ✅ Operações de Set OK (3 membros adicionados)
  🏆 Testando operações de Sorted Set (valkey-glide)...
    ✅ Operações de Sorted Set OK (3 membros adicionados)
  ⚡ Medindo performance básica (valkey-glide)...
    📊 Performance (100 ops):
       SET: 2800 ops/sec (35.71ms total)
       GET: 3571 ops/sec (28.00ms total)

=== Demonstração Concluída ===
```

## O que Este Exemplo Demonstra

### 1. **Compatibilidade Total**
- Ambos os providers implementam a mesma interface `IClient`
- Código cliente é idêntico independente do provider
- Operações retornam resultados consistentes

### 2. **Operações Testadas**
- **Básicas**: SET, GET, DEL, EXPIRE, TTL
- **Hash**: HSET, HGET, HGETALL
- **Lista**: LPUSH, LLEN, LPOP
- **Set**: SADD, SISMEMBER, SMEMBERS
- **Sorted Set**: ZADD, ZSCORE, ZRANGE

### 3. **Performance Comparativa**
- Medição simples de operações por segundo
- Comparação de latência entre providers
- Demonstração de overhead de cada driver

### 4. **Error Handling**
- Tratamento de erros de conexão
- Graceful degradation quando servidor não disponível
- Mensagens informativas para troubleshooting

## Casos de Erro Comuns

### Servidor Não Disponível
```
❌ Erro ao conectar com valkey-go: context deadline exceeded
   (Certifique-se de que o Valkey está rodando em localhost:6379)
```
**Solução**: Verificar se o servidor está rodando na porta correta.

### Autenticação Requerida
```
❌ Erro ao criar cliente valkey-go: NOAUTH Authentication required
```
**Solução**: Configurar password na struct Config ou desabilitar auth no servidor.

### Permissões Insuficientes
```
❌ HSET falhou: READONLY You can't write against a read only replica
```
**Solução**: Conectar ao servidor primário ou usar um servidor com write permissions.

## Estendendo o Exemplo

Para adicionar mais testes:

```go
// Teste customizado
func testCustomOperation(ctx context.Context, client interfaces.IClient, providerName string) {
    fmt.Printf("  🔧 Testando operação customizada (%s)...\n", providerName)
    
    // Sua lógica aqui
    
    fmt.Printf("    ✅ Operação customizada OK\n")
}

// Adicionar no main()
testCustomOperation(ctx, client, p.Name)
```

## Próximos Passos

Após executar este exemplo com sucesso:

1. **Executar testes de compatibilidade**: `go test -run TestProviderCompatibility`
2. **Rodar benchmarks**: `go test -bench=BenchmarkProviders`
3. **Experimentar com configurações**: Modificar timeouts, pool sizes, etc.
4. **Testar em produção**: Avaliar qual provider melhor atende seu caso de uso

## Resultados da Fase 3

Este exemplo demonstra o **sucesso da Fase 3** do roadmap:

- ✅ **Provider valkey-glide implementado** e funcional
- ✅ **Compatibilidade total** entre providers 
- ✅ **Performance comparável** entre implementações
- ✅ **API unificada** simplifica migração entre drivers
- ✅ **Documentação completa** com exemplos práticos
