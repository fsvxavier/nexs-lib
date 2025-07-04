package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/httprequester"
)

func main() {
	// Criar uma instância da factory
	factory := httprequester.NewFactory()

	// URL de exemplo
	baseURL := "https://jsonplaceholder.typicode.com"

	// Demonstração com diferentes clientes
	demoFiber(factory, baseURL)
	demoResty(factory, baseURL)
	demoNetHttp(factory, baseURL)
}

func demoFiber(factory *httprequester.Factory, baseURL string) {
	fmt.Println("=== Cliente Fiber ===")

	// Criar cliente Fiber
	client := factory.Create(httprequester.ClientFiber, baseURL)
	defer client.Close()

	// Configurar cabeçalhos
	client.SetHeaders(map[string]string{
		"Accept": "application/json",
	})

	// Realizar uma requisição GET
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.Get(ctx, "/posts/1")
	if err != nil {
		fmt.Printf("Erro na requisição: %v\n", err)
		return
	}

	// Exibir a resposta
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Body: %s\n\n", response.Body)
}

func demoResty(factory *httprequester.Factory, baseURL string) {
	fmt.Println("=== Cliente Resty ===")

	// Criar cliente Resty
	client := factory.Create(httprequester.ClientResty, baseURL)
	defer client.Close()

	// Configurar cabeçalhos
	client.SetHeaders(map[string]string{
		"Accept": "application/json",
	})

	// Criar uma estrutura para receber a resposta
	type Post struct {
		ID     int    `json:"id"`
		UserID int    `json:"userId"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	var post Post
	client.Unmarshal(&post)

	// Realizar uma requisição GET
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.Get(ctx, "/posts/2")
	if err != nil {
		fmt.Printf("Erro na requisição: %v\n", err)
		return
	}

	// Exibir a resposta
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Post ID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Body: %s\n\n", post.Body)

	// Exibir informações de trace
	traceInfo := client.TraceInfo()
	fmt.Printf("Tempo total: %v\n", traceInfo.TotalTime)
	fmt.Printf("Reusou conexão: %v\n\n", traceInfo.IsConnReused)
}

func demoNetHttp(factory *httprequester.Factory, baseURL string) {
	fmt.Println("=== Cliente Net/HTTP ===")

	// Criar cliente Net/HTTP
	client := factory.Create(httprequester.ClientNetHttp, baseURL)
	defer client.Close()

	// Configurar cabeçalhos
	client.SetHeaders(map[string]string{
		"Accept": "application/json",
	})

	// Criar uma estrutura para receber a resposta
	type Post struct {
		ID     int    `json:"id"`
		UserID int    `json:"userId"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	var post Post
	client.Unmarshal(&post)

	// Realizar uma requisição GET
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.Get(ctx, "/posts/3")
	if err != nil {
		fmt.Printf("Erro na requisição: %v\n", err)
		return
	}

	// Exibir a resposta
	fmt.Printf("Status Code: %d\n", response.StatusCode)
	fmt.Printf("Post ID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Body: %s\n\n", post.Body)
}
