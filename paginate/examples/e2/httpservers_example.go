package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpservers"
	"github.com/fsvxavier/nexs-lib/httpservers/common"
	httpserversfiber "github.com/fsvxavier/nexs-lib/httpservers/fiber"
	page "github.com/fsvxavier/nexs-lib/paginate"
	pagefiber "github.com/fsvxavier/nexs-lib/paginate/fiber"
	"github.com/gofiber/fiber/v2"
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
	// Total de recursos
	totalResources := len(s.resources)

	// Calcular índices para paginação
	startIndex := (metadata.Page.CurrentPage - 1) * metadata.Page.RecordsPerPage
	if startIndex >= totalResources {
		startIndex = 0
	}

	endIndex := startIndex + metadata.Page.RecordsPerPage
	if endIndex > totalResources {
		endIndex = totalResources
	}

	// Aplicar paginação
	var paginatedResources []Resource
	if startIndex < totalResources {
		paginatedResources = s.resources[startIndex:endIndex]
	} else {
		paginatedResources = []Resource{}
	}

	// Criar saída paginada
	return page.NewOutputWithTotal(ctx, paginatedResources, totalResources, metadata)
}

func HttpserversExample() {
	// Criar serviço de recursos
	resourceService := NewResourceService()

	// Criar servidor Fiber através da biblioteca httpservers
	server, err := httpservers.NewServer(
		httpservers.ServerTypeFiber,
		common.WithPort("8080"),
		common.WithHost("0.0.0.0"),
		common.WithSwagger(true),
		common.WithMetrics(true),
	)
	if err != nil {
		log.Fatalf("Erro ao criar servidor: %v", err)
	}

	// Acessar a instância específica do Fiber
	fiberServer, ok := server.(*httpserversfiber.FiberServer)
	if !ok {
		log.Fatalf("Erro ao obter instância do Fiber")
	}

	app := fiberServer.App()

	// Configurar rota para listar recursos
	app.Get("/api/resources", func(c *fiber.Ctx) error {
		ctx := context.Background()

		// Extrair parâmetros de paginação usando adaptador Fiber
		metadata, err := pagefiber.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Usar serviço comum
		result, err := resourceService.ListResources(ctx, metadata)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Retornar resultado paginado
		return c.JSON(result)
	})

	fmt.Println("Servidor Fiber iniciado através da biblioteca httpservers")
	fmt.Println("Acesse http://localhost:8080/api/resources?page=1&limit=3&sort=id&order=desc")

	// Iniciar servidor
	if err := server.Start(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func main() {
	HttpserversExample()
}
