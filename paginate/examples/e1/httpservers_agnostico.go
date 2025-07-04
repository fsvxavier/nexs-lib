package main

import (
	"context"
	"fmt"
	"log"
	"time"

	page "github.com/fsvxavier/nexs-lib/paginate"
)

// Resource representa um recurso genérico
type Resource struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ResourceService gerencia recursos
type ResourceService struct {
	resources []Resource
}

// NewResourceService cria um novo serviço com dados simulados
func NewResourceService() *ResourceService {
	// Dados simulados
	resources := []Resource{
		{ID: 1, Name: "Recurso 1", Description: "Descrição do recurso 1", CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		{ID: 2, Name: "Recurso 2", Description: "Descrição do recurso 2", CreatedAt: time.Now().Add(-9 * 24 * time.Hour)},
		{ID: 3, Name: "Recurso 3", Description: "Descrição do recurso 3", CreatedAt: time.Now().Add(-8 * 24 * time.Hour)},
		{ID: 4, Name: "Recurso 4", Description: "Descrição do recurso 4", CreatedAt: time.Now().Add(-7 * 24 * time.Hour)},
		{ID: 5, Name: "Recurso 5", Description: "Descrição do recurso 5", CreatedAt: time.Now().Add(-6 * 24 * time.Hour)},
		{ID: 6, Name: "Recurso 6", Description: "Descrição do recurso 6", CreatedAt: time.Now().Add(-5 * 24 * time.Hour)},
		{ID: 7, Name: "Recurso 7", Description: "Descrição do recurso 7", CreatedAt: time.Now().Add(-4 * 24 * time.Hour)},
		{ID: 8, Name: "Recurso 8", Description: "Descrição do recurso 8", CreatedAt: time.Now().Add(-3 * 24 * time.Hour)},
		{ID: 9, Name: "Recurso 9", Description: "Descrição do recurso 9", CreatedAt: time.Now().Add(-2 * 24 * time.Hour)},
		{ID: 10, Name: "Recurso 10", Description: "Descrição do recurso 10", CreatedAt: time.Now().Add(-1 * 24 * time.Hour)},
	}
	return &ResourceService{resources: resources}
}

// ListResources lista recursos com paginação
func (s *ResourceService) ListResources(ctx context.Context, metadata *page.Metadata) (*page.Output, error) {
	// Aplicar paginação usando a função centralizada
	// Esta função cuida de todos os cálculos de índices e aplica a paginação automaticamente
	return page.ApplyPaginationToSlice(ctx, s.resources, metadata)
}

func main() {
	// Criar serviço de recursos
	resourceService := NewResourceService()

	// Este exemplo demonstra como usar a função ApplyPaginationToSlice
	// para simplificar a paginação em serviços
	ctx := context.Background()

	// Criar metadados de paginação
	metadata := page.NewMetadata(
		page.WithPage(1),
		page.WithLimit(3),
		page.WithSort("id"),
		page.WithOrder("desc"),
	)

	// Processar a solicitação usando o serviço com a nova função centralizada
	result, err := resourceService.ListResources(ctx, metadata)
	if err != nil {
		log.Fatalf("Erro ao processar recursos: %v", err)
	}

	// Exibir resultado
	fmt.Printf("Página: %d\n", result.Metadata.Page.CurrentPage)
	fmt.Printf("Total de dados: %d\n", result.Metadata.TotalData)
	fmt.Printf("Total de páginas: %d\n", result.Metadata.Page.TotalPages)
	fmt.Printf("Registros por página: %d\n", result.Metadata.Page.RecordsPerPage)
	fmt.Printf("Página anterior: %d\n", result.Metadata.Page.Previous)
	fmt.Printf("Próxima página: %d\n", result.Metadata.Page.Next)

	// Exibir conteúdo
	content := result.Content.([]interface{})
	fmt.Printf("Itens retornados: %d\n", len(content))

	fmt.Println()
	fmt.Println("Este exemplo mostra como usar a função ApplyPaginationToSlice")
	fmt.Println("para simplificar a paginação em camadas de serviço, eliminando")
	fmt.Println("a necessidade de implementar manualmente os cálculos de índices.")
}
