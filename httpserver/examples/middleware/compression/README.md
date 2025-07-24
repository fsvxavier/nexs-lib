# Compression Middleware Example

Este exemplo demonstra o uso do middleware de compressão de respostas HTTP da nexs-lib.

## 📦 Sobre Compressão HTTP

A compressão HTTP reduz significativamente o tamanho das respostas, resultando em:
- **Menor uso de banda** - Até 90% de redução no tamanho
- **Carregamento mais rápido** - Menos dados para transferir
- **Melhor experiência** - Páginas e APIs mais responsivas
- **Economia de custos** - Menor transferência de dados

## 🚀 Executando o Exemplo

```bash
go run main.go
```

O servidor iniciará na porta `:8080` com compressão habilitada.

## 📍 Endpoints

| Endpoint | Conteúdo | Tamanho | Compressível |
|----------|----------|---------|--------------|
| `GET /health` | JSON pequeno | ~100 bytes | ❌ (muito pequeno) |
| `GET /api/small` | JSON pequeno | ~200 bytes | ❌ (abaixo do mínimo) |
| `GET /api/medium` | JSON médio | ~2KB | ✅ |
| `GET /api/large` | JSON grande | ~50KB | ✅ |
| `GET /api/text` | Texto simples | ~5KB | ✅ |
| `GET /static/file.css` | CSS | ~10KB | ✅ |

## 🔧 Configuração

```go
compressionConfig := compression.Config{
    Enabled: true,
    Level:   6,        // Compression level (1-9)
    MinSize: 1024,     // Minimum size to compress (1KB)
    Types: []string{
        "text/html",
        "text/css", 
        "text/javascript",
        "text/plain",
        "application/json",
        "application/javascript",
        "application/xml",
        "application/rss+xml",
        "application/atom+xml",
        "image/svg+xml",
    },
}
```

## 🧪 Testando

### Teste Básico com Compressão
```bash
# Request com Accept-Encoding
curl -H 'Accept-Encoding: gzip' \
     http://localhost:8080/api/large
```

**Headers de Resposta:**
```
Content-Encoding: gzip
Content-Type: application/json
Vary: Accept-Encoding
```

### Teste sem Compressão
```bash
# Request sem Accept-Encoding
curl http://localhost:8080/api/large
```

**Headers de Resposta:**
```
Content-Type: application/json
# Sem Content-Encoding
```

### Comparação de Tamanhos
```bash
# Sem compressão
echo "Tamanho sem compressão:"
curl -s http://localhost:8080/api/large | wc -c

# Com compressão gzip
echo "Tamanho com gzip:"
curl -s -H 'Accept-Encoding: gzip' \
     http://localhost:8080/api/large | wc -c

# Com compressão deflate  
echo "Tamanho com deflate:"
curl -s -H 'Accept-Encoding: deflate' \
     http://localhost:8080/api/large | wc -c
```

### Teste de Performance
```bash
# Tempo sem compressão
time curl -s http://localhost:8080/api/large > /dev/null

# Tempo com compressão
time curl -s -H 'Accept-Encoding: gzip' \
     http://localhost:8080/api/large > /dev/null
```

## 📊 Algoritmos Suportados

### 1. Gzip (Recomendado)
```bash
curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large
```
- **Compressão:** Excelente (~80-90% redução)
- **CPU:** Balanceado
- **Suporte:** Universal

### 2. Deflate
```bash
curl -H 'Accept-Encoding: deflate' http://localhost:8080/api/large
```
- **Compressão:** Boa (~75-85% redução)
- **CPU:** Menor uso
- **Suporte:** Amplo

### 3. Múltiplos Formatos
```bash
# Cliente aceita ambos (servidor escolhe melhor)
curl -H 'Accept-Encoding: gzip, deflate' \
     http://localhost:8080/api/large
```

## ⚙️ Configurações Avançadas

### Níveis de Compressão
```go
// Configurações por nível
configs := map[string]compression.Config{
    "fast": {
        Level: 1,  // Compressão rápida, menor taxa
        MinSize: 512,
    },
    "balanced": {
        Level: 6,  // Padrão - bom equilíbrio
        MinSize: 1024,
    },
    "best": {
        Level: 9,  // Máxima compressão, mais CPU
        MinSize: 2048,
    },
}
```

### Tipos de Conteúdo Específicos
```go
// Apenas JSON e texto
textOnlyConfig := compression.Config{
    Types: []string{
        "application/json",
        "text/plain",
        "text/html",
    },
}

// Incluir imagens SVG
withSVGConfig := compression.Config{
    Types: []string{
        "application/json",
        "image/svg+xml",
        "application/xml",
    },
}
```

### Exclusão de Paths
```go
compressionConfig := compression.Config{
    SkipPaths: []string{
        "/health",        // Health checks pequenos
        "/metrics",       // Prometheus metrics
        "/api/binary",    // Dados binários
        "/images/",       // Imagens já comprimidas
    },
}
```

## 📈 Métricas e Performance

### Eficiência de Compressão por Tipo

| Tipo de Conteúdo | Tamanho Original | Com Gzip | Redução |
|-------------------|------------------|----------|---------|
| JSON repetitivo | 100KB | 15KB | 85% |
| HTML com texto | 50KB | 8KB | 84% |
| CSS | 25KB | 5KB | 80% |
| JavaScript | 100KB | 30KB | 70% |
| JSON compacto | 10KB | 4KB | 60% |

### CPU vs Compressão

| Nível | CPU Usage | Compressão | Cenário |
|-------|-----------|------------|---------|
| 1 | Baixo | ~60% | APIs alta frequência |
| 6 | Médio | ~80% | **Recomendado geral** |
| 9 | Alto | ~85% | Arquivos estáticos |

## 🔍 Detalhes de Implementação

### Headers HTTP Envolvidos

#### Request Headers
```
Accept-Encoding: gzip, deflate
```

#### Response Headers (Comprimido)
```
Content-Encoding: gzip
Content-Type: application/json
Vary: Accept-Encoding
# Content-Length removido (tamanho comprimido)
```

#### Response Headers (Não Comprimido)
```
Content-Type: application/json
Content-Length: 51234
Vary: Accept-Encoding
```

### Quando Compressão NÃO é Aplicada

1. **Cliente não suporta**
   ```bash
   # Sem Accept-Encoding
   curl http://localhost:8080/api/large
   ```

2. **Conteúdo muito pequeno**
   ```bash
   # Resposta < MinSize
   curl -H 'Accept-Encoding: gzip' http://localhost:8080/health
   ```

3. **Tipo não suportado**
   ```bash
   # Content-Type não está na lista
   curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/binary
   ```

4. **Path excluído**
   ```bash
   # Path em SkipPaths
   curl -H 'Accept-Encoding: gzip' http://localhost:8080/metrics
   ```

## 🛠️ Casos de Uso

### API REST JSON
```go
apiConfig := compression.Config{
    Enabled: true,
    Level:   6,
    MinSize: 1024,
    Types:   []string{"application/json"},
}
```

### Website Completo
```go
webConfig := compression.Config{
    Enabled: true,
    Level:   6,
    MinSize: 512,
    Types: []string{
        "text/html",
        "text/css",
        "text/javascript",
        "application/javascript",
        "application/json",
        "image/svg+xml",
    },
}
```

### Microserviço Alto Volume
```go
highVolumeConfig := compression.Config{
    Enabled: true,
    Level:   1,     // Compressão rápida
    MinSize: 2048,  // Apenas respostas grandes
    Types:   []string{"application/json"},
}
```

## 🔧 Troubleshooting

### Problemas Comuns

#### Compressão não funciona
```bash
# Debug: verificar headers
curl -v -H 'Accept-Encoding: gzip' http://localhost:8080/api/large
```

**Verificar:**
- Cliente envia `Accept-Encoding`
- Servidor responde com `Content-Encoding`
- Tamanho > `MinSize`
- Content-Type na lista `Types`

#### Performance degradada
```bash
# Medir tempo de resposta
time curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large
```

**Soluções:**
- Reduzir nível de compressão
- Aumentar `MinSize`
- Excluir endpoints críticos

#### Conteúdo corrompido
```bash
# Verificar integridade
curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large | gunzip | jq .
```

**Causas possíveis:**
- Compressão dupla
- Proxy intermediário
- Headers incorretos

### Debug Detalhado

#### Verificar Suporte do Cliente
```bash
# Testar diferentes encodings
curl -H 'Accept-Encoding: gzip' -v http://localhost:8080/api/large
curl -H 'Accept-Encoding: deflate' -v http://localhost:8080/api/large
curl -H 'Accept-Encoding: br' -v http://localhost:8080/api/large
```

#### Medir Eficiência
```bash
#!/bin/bash
echo "=== Teste de Compressão ==="

echo "Sem compressão:"
SIZE_UNCOMPRESSED=$(curl -s http://localhost:8080/api/large | wc -c)
echo "Tamanho: $SIZE_UNCOMPRESSED bytes"

echo -e "\nCom gzip:"
SIZE_GZIP=$(curl -s -H 'Accept-Encoding: gzip' http://localhost:8080/api/large | wc -c)
echo "Tamanho: $SIZE_GZIP bytes"

RATIO=$((100 - (SIZE_GZIP * 100 / SIZE_UNCOMPRESSED)))
echo "Redução: $RATIO%"
```

## 🚀 Otimizações

### Configuração por Ambiente

#### Desenvolvimento
```go
devConfig := compression.Config{
    Enabled: true,
    Level:   1,     // Rápido para debug
    MinSize: 512,   // Comprimir mais coisas
}
```

#### Produção
```go
prodConfig := compression.Config{
    Enabled: true,
    Level:   6,     // Balanceado
    MinSize: 1024,  // Otimizado
}
```

#### High Performance
```go
hpConfig := compression.Config{
    Enabled: true,
    Level:   1,     // Mínimo CPU
    MinSize: 4096,  // Apenas grandes
}
```

## 📚 Referências

- [HTTP Compression (RFC 7231)](https://tools.ietf.org/html/rfc7231#section-3.1.2.1)
- [Gzip Format Specification](https://tools.ietf.org/html/rfc1952)
- [Web Performance Best Practices](https://web.dev/fast/)
