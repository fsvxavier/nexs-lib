# Streaming Example

Este exemplo demonstra como usar streaming HTTP para processar grandes volumes de dados de forma eficiente.

## Funcionalidades Demonstradas

### 1. **FileDownloadHandler**
- Download de arquivos grandes com streaming
- Monitoramento de progresso em tempo real
- Cálculo de velocidade de download
- Gravação eficiente em disco

### 2. **JSONStreamHandler**
- Processamento de JSON em streaming
- Contagem de objetos em tempo real
- Acumulação de chunks para processamento posterior
- Memória eficiente para grandes responses

### 3. **ProgressHandler**
- Barra de progresso visual
- Cálculo de percentual de conclusão
- Estimativa de velocidade de transferência
- Updates em tempo real

## Como Executar

```bash
cd httpclient/examples/streaming
go run main.go
```

## Saída Esperada

```
🌊 Streaming Example
====================

1️⃣ Streaming JSON data from /json endpoint...
📄 JSON chunk received: 429 bytes (1 total objects so far)
✅ JSON stream completed: 429 bytes, 1 objects processed
📄 Received JSON data preview: {
  "slideshow": {
    "author": "Yours Truly",
    "date": "date of publication",
    "slides": [
      {
        "title": "Wake up to WonderWidgets!",
        "type": "all"
      },
...

2️⃣ Streaming large response with progress bar...
🔄 [██████████████████████████████████████████████████] 100.0% (10240/10240 bytes) 45.23 KB/s
✅ Transfer completed in 226ms (avg 45.23 KB/s)

3️⃣ Downloading file with streaming...
📥 Downloaded: 102400 bytes (67.45 KB/s)
📥 Downloaded: 204800 bytes (89.12 KB/s)
📥 Downloaded: 307200 bytes (102.56 KB/s)
✅ Download completed: 50000 bytes in 487ms (avg 102.56 KB/s)
📁 File saved: downloaded_data.bin (50000 bytes)

4️⃣ Multiple concurrent streaming downloads...
🚀 Starting stream 1...
🚀 Starting stream 2...
🚀 Starting stream 3...
📄 JSON chunk received: 429 bytes (1 total objects so far)
✅ Stream 1 completed (429 bytes)
📄 JSON chunk received: 429 bytes (1 total objects so far)
✅ Stream 2 completed (429 bytes)
📄 JSON chunk received: 429 bytes (1 total objects so far)
✅ Stream 3 completed (429 bytes)
All concurrent streams completed!

5️⃣ Streaming with timeout...
⏰ Streaming timed out as expected: context deadline exceeded

🧹 Cleaning up...
🗑️  Removed file: downloaded_data.bin

🎉 Streaming example completed!

💡 Key Features Demonstrated:
  • JSON data streaming
  • Progress bar for large downloads
  • File download with streaming
  • Concurrent streaming operations
  • Streaming with timeout handling
  • Memory-efficient large data processing
```

## Interface de StreamHandler

```go
type StreamHandler interface {
    OnData(data []byte) error
    OnError(err error)
    OnComplete()
}
```

## Vantagens do Streaming

### **Eficiência de Memória**
- Processa dados em chunks pequenos
- Não carrega toda a resposta na memória
- Ideal para arquivos grandes (GB+)

### **Processamento em Tempo Real**
- Processa dados conforme chegam
- Permite feedback de progresso
- Reduz latência percebida

### **Flexibilidade**
- Handlers customizados para diferentes tipos de dados
- Processamento específico por caso de uso
- Integração com diferentes sistemas de armazenamento

## Casos de Uso

### **Download de Arquivos Grandes**
```go
handler := NewFileDownloadHandler("large_file.zip")
client.Stream(ctx, "GET", "/large-file", handler)
```

### **Processamento de APIs de Streaming**
```go
handler := NewJSONStreamHandler()
client.Stream(ctx, "GET", "/stream/events", handler)
```

### **Backup e Sincronização**
```go
handler := NewProgressHandler(expectedSize)
client.Stream(ctx, "GET", "/backup/data", handler)
```

### **Log Processing**
```go
handler := NewLogProcessingHandler()
client.Stream(ctx, "GET", "/logs/stream", handler)
```

## Considerações de Performance

- **Chunk Size**: Otimizar tamanho do buffer baseado no caso de uso
- **Memory Management**: Evitar acumular todos os chunks na memória
- **Error Handling**: Implementar retry e recovery para streams longos
- **Timeout**: Configurar timeouts apropriados para streams longos
- **Concurrency**: Usar goroutines para processamento paralelo de chunks
