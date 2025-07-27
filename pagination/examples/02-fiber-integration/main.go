package main

import (
	"log"

	"github.com/fsvxavier/nexs-lib/pagination/config"
	paginationFiber "github.com/fsvxavier/nexs-lib/pagination/providers/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Product representa um produto no e-commerce
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
}

// ProductRepository simula um repositório de produtos
type ProductRepository struct {
	products []Product
}

func NewProductRepository() *ProductRepository {
	// Simular dados de produtos
	products := []Product{
		{ID: 1, Name: "Smartphone Galaxy", Description: "Smartphone Android", Price: 899.99, Category: "Electronics", InStock: true},
		{ID: 2, Name: "Laptop Dell", Description: "Laptop para trabalho", Price: 1299.99, Category: "Electronics", InStock: true},
		{ID: 3, Name: "Tênis Nike", Description: "Tênis esportivo", Price: 199.99, Category: "Sports", InStock: true},
		{ID: 4, Name: "Livro Golang", Description: "Aprenda Go programming", Price: 59.99, Category: "Books", InStock: true},
		{ID: 5, Name: "Cafeteira Expresso", Description: "Máquina de café", Price: 299.99, Category: "Home", InStock: false},
		{ID: 6, Name: "Mouse Gamer", Description: "Mouse RGB para gaming", Price: 79.99, Category: "Electronics", InStock: true},
		{ID: 7, Name: "Teclado Mecânico", Description: "Teclado para programadores", Price: 149.99, Category: "Electronics", InStock: true},
		{ID: 8, Name: "Monitor 4K", Description: "Monitor ultra HD", Price: 499.99, Category: "Electronics", InStock: true},
		{ID: 9, Name: "Cadeira Gamer", Description: "Cadeira ergonômica", Price: 349.99, Category: "Furniture", InStock: true},
		{ID: 10, Name: "Webcam HD", Description: "Câmera para videoconferência", Price: 89.99, Category: "Electronics", InStock: true},
		{ID: 11, Name: "Headset Wireless", Description: "Fone sem fio", Price: 129.99, Category: "Electronics", InStock: true},
		{ID: 12, Name: "Tablet Android", Description: "Tablet para estudos", Price: 299.99, Category: "Electronics", InStock: false},
		{ID: 13, Name: "Smartwatch", Description: "Relógio inteligente", Price: 249.99, Category: "Electronics", InStock: true},
		{ID: 14, Name: "Power Bank", Description: "Bateria portátil", Price: 39.99, Category: "Electronics", InStock: true},
		{ID: 15, Name: "Cabo USB-C", Description: "Cabo para carregamento", Price: 19.99, Category: "Electronics", InStock: true},
	}

	return &ProductRepository{products: products}
}

func (r *ProductRepository) GetProducts(offset, limit int, sortField, sortOrder string) ([]Product, int) {
	// Simular ordenação (em produção seria feito no banco)
	total := len(r.products)

	// Calcular slice com offset e limit
	start := offset
	if start > total {
		start = total
	}

	end := start + limit
	if end > total {
		end = total
	}

	result := r.products[start:end]
	return result, total
}

func main() {
	// Configurar Fiber
	app := fiber.New(fiber.Config{
		AppName: "Pagination Example with Fiber",
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New())

	// Configurar paginação
	paginationConfig := config.NewDefaultConfig()
	paginationConfig.DefaultLimit = 5
	paginationConfig.MaxLimit = 50
	paginationConfig.DefaultSortField = "id"
	paginationConfig.DefaultSortOrder = "asc"

	paginationService := paginationFiber.NewFiberPaginationService(paginationConfig)

	// Repositório de produtos
	productRepo := NewProductRepository()

	// Rota para listar produtos com paginação
	app.Get("/api/products", func(c *fiber.Ctx) error {
		// Campos que podem ser usados para ordenação
		sortableFields := []string{"id", "name", "price", "category"}

		// Parse dos parâmetros de paginação do contexto Fiber
		paginationParams, err := paginationService.ParseFromFiber(c, sortableFields...)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid pagination parameters",
				"details": err.Error(),
			})
		}

		// Calcular offset
		offset := (paginationParams.Page - 1) * paginationParams.Limit

		// Buscar produtos do repositório
		products, total := productRepo.GetProducts(
			offset,
			paginationParams.Limit,
			paginationParams.SortField,
			paginationParams.SortOrder,
		)

		// Criar resposta paginada
		response := paginationService.CreateResponse(products, paginationParams, total)

		return c.JSON(response)
	})

	// Rota para demonstrar diferentes filtros
	app.Get("/api/products/in-stock", func(c *fiber.Ctx) error {
		sortableFields := []string{"id", "name", "price"}

		paginationParams, err := paginationService.ParseFromFiber(c, sortableFields...)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid pagination parameters",
				"details": err.Error(),
			})
		}

		// Filtrar apenas produtos em estoque
		var inStockProducts []Product
		for _, product := range productRepo.products {
			if product.InStock {
				inStockProducts = append(inStockProducts, product)
			}
		}

		// Aplicar paginação manualmente para demonstração
		total := len(inStockProducts)
		offset := (paginationParams.Page - 1) * paginationParams.Limit

		start := offset
		if start > total {
			start = total
		}

		end := start + paginationParams.Limit
		if end > total {
			end = total
		}

		result := inStockProducts[start:end]
		response := paginationService.CreateResponse(result, paginationParams, total)

		return c.JSON(response)
	})

	// Rota para demonstrar tratamento de erros
	app.Get("/api/products/error-demo", func(c *fiber.Ctx) error {
		// Campos de ordenação muito restritivos para demonstrar erro
		restrictedFields := []string{"id"} // Apenas ID permitido

		_, err := paginationService.ParseFromFiber(c, restrictedFields...)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":               "Validation failed",
				"details":             err.Error(),
				"allowed_sort_fields": restrictedFields,
			})
		}

		return c.JSON(fiber.Map{"message": "This should not be reached if invalid sort field is provided"})
	})

	// Rota de informações da API
	app.Get("/api/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Pagination Example API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"GET /api/products": fiber.Map{
					"description": "Lista produtos com paginação",
					"parameters": fiber.Map{
						"page":  "Número da página (padrão: 1)",
						"limit": "Registros por página (padrão: 5, máx: 50)",
						"sort":  "Campo de ordenação (id, name, price, category)",
						"order": "Ordem de classificação (asc, desc)",
					},
					"example": "/api/products?page=2&limit=3&sort=price&order=desc",
				},
				"GET /api/products/in-stock": fiber.Map{
					"description": "Lista apenas produtos em estoque",
					"parameters":  "Mesmos parâmetros de /api/products",
				},
				"GET /api/products/error-demo": fiber.Map{
					"description": "Demonstra tratamento de erros de validação",
					"note":        "Tente usar sort=name para ver o erro",
				},
			},
		})
	})

	// Servir página de exemplo simples
	app.Get("/", func(c *fiber.Ctx) error {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Pagination API Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .example { background: #e8f4f8; padding: 5px; margin: 5px 0; font-family: monospace; }
    </style>
</head>
<body>
    <h1>🚀 Pagination API Example</h1>
    
    <h2>Endpoints Disponíveis</h2>
    
    <div class="endpoint">
        <h3>📦 GET /api/products</h3>
        <p>Lista produtos com paginação completa</p>
        <div class="example">
            <a href="/api/products">/api/products</a><br>
            <a href="/api/products?page=2&limit=3">/api/products?page=2&limit=3</a><br>
            <a href="/api/products?sort=price&order=desc">/api/products?sort=price&order=desc</a>
        </div>
    </div>
    
    <div class="endpoint">
        <h3>✅ GET /api/products/in-stock</h3>
        <p>Lista apenas produtos em estoque</p>
        <div class="example">
            <a href="/api/products/in-stock">/api/products/in-stock</a><br>
            <a href="/api/products/in-stock?limit=2">/api/products/in-stock?limit=2</a>
        </div>
    </div>
    
    <div class="endpoint">
        <h3>❌ GET /api/products/error-demo</h3>
        <p>Demonstra tratamento de erros</p>
        <div class="example">
            <a href="/api/products/error-demo?sort=name">/api/products/error-demo?sort=name</a> (deve dar erro)
        </div>
    </div>
    
    <div class="endpoint">
        <h3>ℹ️ GET /api/info</h3>
        <p>Informações da API</p>
        <div class="example">
            <a href="/api/info">/api/info</a>
        </div>
    </div>
    
    <h2>Parâmetros de Query Suportados</h2>
    <ul>
        <li><strong>page</strong>: Número da página (padrão: 1)</li>
        <li><strong>limit</strong>: Registros por página (padrão: 5, máximo: 50)</li>
        <li><strong>sort</strong>: Campo de ordenação (id, name, price, category)</li>
        <li><strong>order</strong>: Ordem (asc, desc)</li>
    </ul>
</body>
</html>`
		return c.Type("html").SendString(html)
	})

	// Iniciar servidor
	log.Println("🚀 Servidor iniciando na porta :3000")
	log.Println("📖 Acesse http://localhost:3000 para ver a documentação")
	log.Println("🔗 Teste a API em http://localhost:3000/api/products")

	log.Fatal(app.Listen(":3000"))
}
