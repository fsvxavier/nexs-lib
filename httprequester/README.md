# HTTP Requester

`httprequester` é uma biblioteca Go para realizar requisições HTTP com suporte a diferentes implementações de clientes. A biblioteca oferece uma interface comum para trabalhar com diferentes bibliotecas de clientes HTTP, facilitando a mudança entre implementações sem alterar a lógica de negócio.

## Características

- Interface unificada para diferentes clientes HTTP
- Suporte para Fiber, Resty e net/http
- Reutilização de conexões para melhor performance
- Configuração flexível
- Suporte a rastreamento de requisições
- Tratamento padronizado de erros
- Factory para criação simplificada de clientes

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/httprequester
```

## Uso Básico

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/fsvxavier/nexs-lib/httprequester"
)

func main() {
	// Criar uma factory de clientes
	factory := httprequester.NewFactory()
	
	// Criar um cliente HTTP (net/http, resty ou fiber)
	client := factory.Create(httprequester.ClientNetHttp, "https://api.example.com")
	defer client.Close()
	
	// Configurar cabeçalhos
	client.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Authorization": "Bearer YOUR_TOKEN",
	})
	
	// Fazer uma requisição GET
	ctx := context.Background()
	response, err := client.Get(ctx, "/users")
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}
	
	// Processar a resposta
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Body: %s\n", response.Body)
}
```

## Clientes Suportados

### Net/HTTP

```go
// Criar um cliente Net/HTTP
client := factory.Create(httprequester.ClientNetHttp, "https://api.example.com")

// Com configuração personalizada
config := &nethttp.Config{
    TLSEnabled: true,
    ClientTimeout: 30 * time.Second,
}
httpClient := nethttp.NewClient(config)
client, err := factory.CreateWithClient(httprequester.ClientNetHttp, "https://api.example.com", httpClient)
```

### Resty

```go
// Criar um cliente Resty
client := factory.Create(httprequester.ClientResty, "https://api.example.com")

// Com configuração personalizada
config := &resty.Config{
    EnableTrace: true,
    Timeout: 30 * time.Second,
}
restyClient := resty.NewClient(config)
client, err := factory.CreateWithClient(httprequester.ClientResty, "https://api.example.com", restyClient)
```

### Fiber

```go
// Criar um cliente Fiber
client := factory.Create(httprequester.ClientFiber, "https://api.example.com")

// Com configuração personalizada
config := &fiber.Config{
    ReadTimeout: 10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
fiberClient := fiber.NewClient(config)
client, err := factory.CreateWithClient(httprequester.ClientFiber, "https://api.example.com", fiberClient)
```

## Desserializar respostas

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var user User
client.Unmarshal(&user)

response, err := client.Get(ctx, "/users/1")
if err != nil {
    return err
}

// user agora contém os dados desserializados
fmt.Printf("User ID: %d, Name: %s\n", user.ID, user.Name)
```

## Rastreamento de Requisições

```go
// Fazer a requisição
response, err := client.Get(ctx, "/users")

// Obter informações de trace
traceInfo := client.TraceInfo()
fmt.Printf("DNS Lookup: %v\n", traceInfo.DNSLookup)
fmt.Printf("Tempo total: %v\n", traceInfo.TotalTime)
fmt.Printf("Conexão reutilizada: %v\n", traceInfo.IsConnReused)
```
