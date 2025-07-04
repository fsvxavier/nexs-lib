package main

import (
	"context"
	"fmt"
	"log"
	"time"

	page "github.com/fsvxavier/nexs-lib/paginate"
)

// Este exemplo demonstra como usar a função ApplyPaginationToSlice para paginar dados em memória
// em serviços que lidam com slices em vez de consultas ao banco de dados

// Item representa um exemplo de entidade para paginação
type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ItemService representa um serviço que gerencia itens em memória
type ItemService struct {
	items []Item // Simula um repositório de dados em memória
}

// NewItemService cria um novo serviço de itens com dados de exemplo
func NewItemService() *ItemService {
	// Criar dados de exemplo
	items := make([]Item, 0, 50)
	for i := 1; i <= 50; i++ {
		items = append(items, Item{
			ID:        i,
			Name:      fmt.Sprintf("Item %d", i),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	return &ItemService{
		items: items,
	}
}

// GetPaginatedItems retorna uma lista paginada de itens usando a paginação centralizada
func (s *ItemService) GetPaginatedItems(ctx context.Context, pageNum, limit int, sortField, sortOrder string) (*page.Output, error) {
	// Criar metadados com as opções de paginação
	metadata := page.NewMetadata(
		page.WithPage(pageNum),
		page.WithLimit(limit),
		page.WithSort(sortField),
		page.WithOrder(sortOrder),
	)

	// Aplicar a paginação à slice de itens
	return page.ApplyPaginationToSlice(ctx, s.items, metadata)
}

// Função principal para demonstrar o uso
func ExamplePaginateSlice() {
	ctx := context.Background()

	// Criar serviço de exemplo
	service := NewItemService()

	// Obter primeira página (5 itens por página)
	output1, err := service.GetPaginatedItems(ctx, 1, 5, "id", "asc")
	if err != nil {
		log.Fatalf("Erro ao paginar itens: %v", err)
	}

	// Exibir informações da primeira página
	fmt.Printf("Página 1: %d itens, Total: %d, Total de Páginas: %d\n",
		len(output1.Content.([]interface{})),
		output1.Metadata.TotalData,
		output1.Metadata.Page.TotalPages)
	fmt.Println("Próxima página:", output1.Metadata.Page.Next)

	// Obter segunda página
	output2, err := service.GetPaginatedItems(ctx, 2, 5, "id", "asc")
	if err != nil {
		log.Fatalf("Erro ao paginar itens: %v", err)
	}

	// Exibir informações da segunda página
	fmt.Printf("Página 2: %d itens, Total: %d, Total de Páginas: %d\n",
		len(output2.Content.([]interface{})),
		output2.Metadata.TotalData,
		output2.Metadata.Page.TotalPages)
	fmt.Println("Página anterior:", output2.Metadata.Page.Previous)
	fmt.Println("Próxima página:", output2.Metadata.Page.Next)

	// Tentar obter uma página além do limite (deve gerar um erro)
	_, err = service.GetPaginatedItems(ctx, 20, 5, "id", "asc")
	if err != nil {
		fmt.Println("Erro esperado ao solicitar página inválida:", err)
	}
}

func main() {
	ExamplePaginateSlice()
}
