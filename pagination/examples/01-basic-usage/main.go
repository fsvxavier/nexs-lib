package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
)

// User representa um usuário no sistema
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func main() {
	// 1. Configurar o serviço de paginação
	cfg := config.NewDefaultConfig()
	cfg.DefaultLimit = 10
	cfg.MaxLimit = 100
	cfg.DefaultSortField = "id"
	cfg.DefaultSortOrder = "asc"

	paginationService := pagination.NewPaginationService(cfg)

	// 2. Simular parâmetros de query vindos de uma requisição HTTP
	// Ex: /api/users?page=2&limit=5&sort=name&order=desc
	queryParams := url.Values{
		"page":  []string{"2"},
		"limit": []string{"5"},
		"sort":  []string{"name"},
		"order": []string{"desc"},
	}

	// 3. Campos que podem ser usados para ordenação
	sortableFields := []string{"id", "name", "email", "created_at"}

	// 4. Parse dos parâmetros de paginação
	paginationParams, err := paginationService.ParseRequest(queryParams, sortableFields...)
	if err != nil {
		log.Fatalf("Erro ao fazer parse dos parâmetros: %v", err)
	}

	fmt.Printf("Parâmetros de paginação:\n")
	fmt.Printf("  Página: %d\n", paginationParams.Page)
	fmt.Printf("  Limite: %d\n", paginationParams.Limit)
	fmt.Printf("  Campo de ordenação: %s\n", paginationParams.SortField)
	fmt.Printf("  Ordem: %s\n\n", paginationParams.SortOrder)

	// 5. Construir query SQL com paginação
	baseQuery := "SELECT id, name, email, active FROM users WHERE active = true"
	finalQuery := paginationService.BuildQuery(baseQuery, paginationParams)
	countQuery := paginationService.BuildCountQuery(baseQuery)

	fmt.Printf("Query principal:\n%s\n\n", finalQuery)
	fmt.Printf("Query de contagem:\n%s\n\n", countQuery)

	// 6. Simular dados de usuários (normalmente viriam do banco)
	users := []User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Active: true},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Active: true},
		{ID: 3, Name: "Carol Brown", Email: "carol@example.com", Active: true},
		{ID: 4, Name: "David Wilson", Email: "david@example.com", Active: true},
		{ID: 5, Name: "Eve Davis", Email: "eve@example.com", Active: true},
	}

	// Simular total de registros no banco
	totalRecords := 23

	// 7. Criar resposta paginada
	response := paginationService.CreateResponse(users, paginationParams, totalRecords)

	// 8. Exibir resultado
	fmt.Printf("Resposta paginada:\n")
	jsonResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(jsonResponse))

	// 9. Exemplo de navegação
	fmt.Printf("\nInformações de navegação:\n")
	fmt.Printf("  Página atual: %d\n", response.Metadata.CurrentPage)
	fmt.Printf("  Total de páginas: %d\n", response.Metadata.TotalPages)
	fmt.Printf("  Total de registros: %d\n", response.Metadata.TotalRecords)

	if response.Metadata.Previous != nil {
		fmt.Printf("  Página anterior: %d\n", *response.Metadata.Previous)
	} else {
		fmt.Printf("  Página anterior: não disponível\n")
	}

	if response.Metadata.Next != nil {
		fmt.Printf("  Próxima página: %d\n", *response.Metadata.Next)
	} else {
		fmt.Printf("  Próxima página: não disponível\n")
	}
}
