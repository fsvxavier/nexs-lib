# HTTP/2 Examples

Este diretório contém exemplos práticos demonstrando o uso de HTTP/2 com o cliente HTTP nexs-lib.

## 📋 Exemplos Disponíveis

### 1. Requisição HTTP/2 Básica
Demonstra como fazer uma requisição básica utilizando HTTP/2 e verificar características da resposta.

### 2. Multiplexing com HTTP/2
Mostra como o HTTP/2 permite múltiplas requisições simultâneas sobre a mesma conexão TCP.

### 3. Comparação de Performance
Compara a performance entre HTTP/2 e HTTP/1.1 usando requisições em lote.

### 4. Configuração TLS
Demonstra HTTP/2 funcionando sobre TLS (h2) com verificação de cabeçalhos de segurança.

### 5. Processamento de Stream
Testa capacidades de streaming do HTTP/2 com diferentes tipos de conteúdo.

### 6. Recursos Avançados
Explora recursos avançados do HTTP/2 como headers customizados e compressão.

## 🚀 Como Executar

```bash
cd httpclient/examples/http2
go run main.go
```

## 🔧 Características do HTTP/2

### Multiplexing
- **O que é**: Múltiplas requisições simultâneas sobre uma única conexão TCP
- **Benefício**: Reduz latência e melhora performance
- **Exemplo**: Requisições paralelas para diferentes endpoints

### Server Push (Simulado)
- **O que é**: Servidor pode enviar recursos antes da requisição
- **Benefício**: Reduce round trips para recursos críticos
- **Limitação**: Simulado nos exemplos (depende do servidor)

### Compressão de Headers
- **O que é**: Headers HTTP são comprimidos usando HPACK
- **Benefício**: Reduz overhead de headers repetitivos
- **Exemplo**: Headers de autenticação e cookies

### Priorização de Stream
- **O que é**: Requisições podem ter prioridades diferentes
- **Benefício**: Recursos críticos são carregados primeiro
- **Uso**: CSS/JS crítico vs. imagens/analytics

## 📊 Performance

### Vantagens do HTTP/2:
- ✅ Multiplexing elimina bloqueio de requisições
- ✅ Compressão de headers reduz overhead
- ✅ Única conexão TCP reduz handshakes
- ✅ Priorização melhora experiência do usuário

### Considerações:
- ⚠️ Requer HTTPS para browsers
- ⚠️ Performance depende da implementação do servidor
- ⚠️ Benefícios são mais notáveis com múltiplas requisições

## 🔗 Endpoints de Teste

Os exemplos utilizam [httpbin.org](https://httpbin.org) que suporta HTTP/2:

- `/get` - Requisição GET básica
- `/headers` - Inspeção de headers
- `/stream/n` - Stream de n objetos JSON
- `/drip` - Stream com delay controlado
- `/range/n` - Requisição de range de n bytes

## 🏗️ Configuração do Cliente

```go
// Cliente básico com HTTP/2
client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
if err != nil {
    log.Fatal(err)
}

// Configurar timeout
client = client.SetTimeout(10 * time.Second)

// Fazer requisição
response, err := client.Get(ctx, "/get")
```

## 📈 Métricas e Monitoramento

O exemplo inclui medição de:
- **Latência**: Tempo total de requisição
- **Throughput**: Bytes por segundo
- **Taxa de Sucesso**: Requisições bem-sucedidas vs. falhas
- **Tempo Médio**: Performance por requisição

## 🔍 Debugging

Para debugar conexões HTTP/2:
1. Verifique se o servidor suporta HTTP/2
2. Confirme que HTTPS está sendo usado
3. Monitor logs de conexão TCP
4. Analise headers de resposta

## 📚 Recursos Adicionais

- [RFC 7540 - HTTP/2](https://tools.ietf.org/html/rfc7540)
- [HTTP/2 explained](https://daniel.haxx.se/http2/)
- [Can I use HTTP/2](https://caniuse.com/http2)
- [HTTP/2 vs HTTP/1.1](https://developers.google.com/web/fundamentals/performance/http2)

## 🤝 Integração

Este exemplo pode ser integrado com:
- **Middleware**: Logging, autenticação, rate limiting
- **Hooks**: Métricas, auditoria, cache
- **Streaming**: Download de arquivos grandes
- **Batch**: Operações paralelas otimizadas
