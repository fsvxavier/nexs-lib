// Package main demonstrates a RESTful API with i18n using Echo framework.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Product represents a product model
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Error    *APIError   `json:"error,omitempty"`
	Metadata *Metadata   `json:"metadata,omitempty"`
}

// APIError represents error information
type APIError struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Metadata represents response metadata
type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language"`
	Version   string    `json:"version"`
}

// I18nAPI holds the i18n provider and manages translations
type I18nAPI struct {
	provider interfaces.I18n
	echo     *echo.Echo
}

// Sample product data
var products = []Product{
	{ID: 1, Name: "Smartphone", Description: "High-performance smartphone", Price: 899.99, Category: "electronics"},
	{ID: 2, Name: "Laptop", Description: "Gaming laptop with powerful GPU", Price: 1299.99, Category: "electronics"},
	{ID: 3, Name: "Coffee Mug", Description: "Ceramic coffee mug", Price: 12.99, Category: "home"},
	{ID: 4, Name: "Book", Description: "Programming best practices", Price: 39.99, Category: "books"},
}

func main() {
	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_echo_api")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create translation files
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal("Failed to create translation files:", err)
	}

	// Setup i18n API
	api, err := setupI18nAPI(tempDir)
	if err != nil {
		log.Fatal("Failed to setup i18n API:", err)
	}

	// Start server
	fmt.Println("🚀 RESTful API with i18n running on http://localhost:8080")
	fmt.Println("📖 Available endpoints:")
	fmt.Println("  GET    /api/v1/products         - List all products")
	fmt.Println("  GET    /api/v1/products/:id     - Get product by ID")
	fmt.Println("  POST   /api/v1/products         - Create new product")
	fmt.Println("  PUT    /api/v1/products/:id     - Update product")
	fmt.Println("  DELETE /api/v1/products/:id     - Delete product")
	fmt.Println("  GET    /api/v1/categories       - List categories")
	fmt.Println("  GET    /health                  - Health check")
	fmt.Println("  GET    /metrics                 - API metrics")
	fmt.Println("")
	fmt.Println("💡 Use Accept-Language header or ?lang=pt query parameter")
	fmt.Println("   Supported languages: en, pt, es")

	log.Fatal(api.echo.Start(":8080"))
}

func setupI18nAPI(translationDir string) (*I18nAPI, error) {
	// Configure i18n
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 15*time.Minute).
		WithLoadTimeout(20 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     translationDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	// Create registry and register provider
	registry := i18n.NewRegistry()
	jsonFactory := &json.Factory{}
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Create provider
	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Start provider
	ctx := context.Background()
	if err := provider.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start provider: %w", err)
	}

	// Setup Echo
	e := echo.New()
	e.HideBanner = true

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	api := &I18nAPI{
		provider: provider,
		echo:     e,
	}

	// Add i18n middleware
	e.Use(api.i18nMiddleware)

	// Setup routes
	api.setupRoutes()

	return api, nil
}

func (api *I18nAPI) i18nMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Determine language from query param, header, or default
		lang := c.QueryParam("lang")
		if lang == "" {
			lang = c.Request().Header.Get("Accept-Language")
			if lang == "" {
				lang = "en"
			}
		}

		// Store language in context
		c.Set("lang", lang)
		return next(c)
	}
}

func (api *I18nAPI) setupRoutes() {
	// API v1 routes
	v1 := api.echo.Group("/api/v1")
	{
		// Product routes
		v1.GET("/products", api.listProducts)
		v1.GET("/products/:id", api.getProduct)
		v1.POST("/products", api.createProduct)
		v1.PUT("/products/:id", api.updateProduct)
		v1.DELETE("/products/:id", api.deleteProduct)

		// Category routes
		v1.GET("/categories", api.listCategories)
	}

	// System routes
	api.echo.GET("/health", api.healthCheck)
	api.echo.GET("/metrics", api.getMetrics)
}

// Helper function to translate text
func (api *I18nAPI) translate(c echo.Context, key string, params map[string]interface{}) string {
	lang := c.Get("lang").(string)

	result, err := api.provider.Translate(c.Request().Context(), key, lang, params)
	if err != nil {
		log.Printf("Translation error for key '%s' in language '%s': %v", key, lang, err)
		return key // Return key as fallback
	}
	return result
}

// Helper function to create API response
func (api *I18nAPI) response(c echo.Context, success bool, messageKey string, data interface{}, params map[string]interface{}) *APIResponse {
	message := api.translate(c, messageKey, params)
	lang := c.Get("lang").(string)

	response := &APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Metadata: &Metadata{
			Timestamp: time.Now().UTC(),
			Language:  lang,
			Version:   "1.0.0",
		},
	}

	if !success {
		response.Error = &APIError{
			Code: messageKey,
		}
	}

	return response
}

// Product handlers
func (api *I18nAPI) listProducts(c echo.Context) error {
	category := c.QueryParam("category")
	var filteredProducts []Product

	if category != "" {
		for _, product := range products {
			if product.Category == category {
				filteredProducts = append(filteredProducts, product)
			}
		}
	} else {
		filteredProducts = products
	}

	params := map[string]interface{}{
		"count": len(filteredProducts),
	}

	response := api.response(c, true, "api.products.list_success", filteredProducts, params)
	return c.JSON(http.StatusOK, response)
}

func (api *I18nAPI) getProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := api.response(c, false, "api.errors.invalid_id", nil, map[string]interface{}{"id": idStr})
		return c.JSON(http.StatusBadRequest, response)
	}

	for _, product := range products {
		if product.ID == id {
			response := api.response(c, true, "api.products.get_success", product, nil)
			return c.JSON(http.StatusOK, response)
		}
	}

	params := map[string]interface{}{"id": id}
	response := api.response(c, false, "api.errors.product_not_found", nil, params)
	return c.JSON(http.StatusNotFound, response)
}

func (api *I18nAPI) createProduct(c echo.Context) error {
	var product Product
	if err := c.Bind(&product); err != nil {
		response := api.response(c, false, "api.errors.invalid_data", nil, nil)
		return c.JSON(http.StatusBadRequest, response)
	}

	// Validate required fields
	if product.Name == "" || product.Price <= 0 {
		response := api.response(c, false, "api.errors.missing_required_fields", nil, nil)
		return c.JSON(http.StatusBadRequest, response)
	}

	// Assign new ID (in real app, this would be from database)
	product.ID = len(products) + 1
	products = append(products, product)

	params := map[string]interface{}{
		"name": product.Name,
		"id":   product.ID,
	}

	response := api.response(c, true, "api.products.create_success", product, params)
	return c.JSON(http.StatusCreated, response)
}

func (api *I18nAPI) updateProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := api.response(c, false, "api.errors.invalid_id", nil, map[string]interface{}{"id": idStr})
		return c.JSON(http.StatusBadRequest, response)
	}

	var updatedProduct Product
	if err := c.Bind(&updatedProduct); err != nil {
		response := api.response(c, false, "api.errors.invalid_data", nil, nil)
		return c.JSON(http.StatusBadRequest, response)
	}

	for i, product := range products {
		if product.ID == id {
			updatedProduct.ID = id
			products[i] = updatedProduct

			params := map[string]interface{}{
				"name": updatedProduct.Name,
				"id":   id,
			}

			response := api.response(c, true, "api.products.update_success", updatedProduct, params)
			return c.JSON(http.StatusOK, response)
		}
	}

	params := map[string]interface{}{"id": id}
	response := api.response(c, false, "api.errors.product_not_found", nil, params)
	return c.JSON(http.StatusNotFound, response)
}

func (api *I18nAPI) deleteProduct(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := api.response(c, false, "api.errors.invalid_id", nil, map[string]interface{}{"id": idStr})
		return c.JSON(http.StatusBadRequest, response)
	}

	for i, product := range products {
		if product.ID == id {
			// Remove product from slice
			products = append(products[:i], products[i+1:]...)

			params := map[string]interface{}{
				"name": product.Name,
				"id":   id,
			}

			response := api.response(c, true, "api.products.delete_success", nil, params)
			return c.JSON(http.StatusOK, response)
		}
	}

	params := map[string]interface{}{"id": id}
	response := api.response(c, false, "api.errors.product_not_found", nil, params)
	return c.JSON(http.StatusNotFound, response)
}

func (api *I18nAPI) listCategories(c echo.Context) error {
	categoryMap := make(map[string]int)
	for _, product := range products {
		categoryMap[product.Category]++
	}

	var categories []map[string]interface{}
	for category, count := range categoryMap {
		categories = append(categories, map[string]interface{}{
			"name":          category,
			"product_count": count,
		})
	}

	params := map[string]interface{}{
		"count": len(categories),
	}

	response := api.response(c, true, "api.categories.list_success", categories, params)
	return c.JSON(http.StatusOK, response)
}

// System handlers
func (api *I18nAPI) healthCheck(c echo.Context) error {
	// Check i18n provider health
	if err := api.provider.Health(c.Request().Context()); err != nil {
		response := api.response(c, false, "health.unhealthy", nil, nil)
		return c.JSON(http.StatusServiceUnavailable, response)
	}

	healthData := map[string]interface{}{
		"timestamp":           time.Now().UTC(),
		"supported_languages": api.provider.GetSupportedLanguages(),
		"default_language":    api.provider.GetDefaultLanguage(),
		"loaded_languages":    api.provider.GetLoadedLanguages(),
		"translation_count":   api.provider.GetTranslationCount(),
	}

	response := api.response(c, true, "health.healthy", healthData, nil)
	return c.JSON(http.StatusOK, response)
}

func (api *I18nAPI) getMetrics(c echo.Context) error {
	metrics := map[string]interface{}{
		"api_version":         "1.0.0",
		"total_products":      len(products),
		"supported_languages": api.provider.GetSupportedLanguages(),
		"default_language":    api.provider.GetDefaultLanguage(),
		"translation_stats": map[string]interface{}{
			"total_translations": api.provider.GetTranslationCount(),
			"loaded_languages":   api.provider.GetLoadedLanguages(),
		},
		"uptime": time.Since(time.Now().Add(-time.Hour)), // Mock uptime
	}

	response := api.response(c, true, "api.metrics.success", metrics, nil)
	return c.JSON(http.StatusOK, response)
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "api": {
    "products": {
      "list_success": "Found {{count}} products",
      "get_success": "Product retrieved successfully",
      "create_success": "Product '{{name}}' created successfully with ID {{id}}",
      "update_success": "Product '{{name}}' (ID: {{id}}) updated successfully",
      "delete_success": "Product '{{name}}' (ID: {{id}}) deleted successfully"
    },
    "categories": {
      "list_success": "Found {{count}} categories"
    },
    "metrics": {
      "success": "API metrics retrieved successfully"
    },
    "errors": {
      "invalid_id": "Invalid ID format: {{id}}",
      "invalid_data": "Invalid or malformed data provided",
      "missing_required_fields": "Missing required fields: name and price must be provided",
      "product_not_found": "Product with ID {{id}} not found"
    }
  },
  "health": {
    "healthy": "API is healthy and operational",
    "unhealthy": "API is experiencing issues"
  }
}`

	// Portuguese translations
	ptContent := `{
  "api": {
    "products": {
      "list_success": "Encontrados {{count}} produtos",
      "get_success": "Produto recuperado com sucesso",
      "create_success": "Produto '{{name}}' criado com sucesso com ID {{id}}",
      "update_success": "Produto '{{name}}' (ID: {{id}}) atualizado com sucesso",
      "delete_success": "Produto '{{name}}' (ID: {{id}}) deletado com sucesso"
    },
    "categories": {
      "list_success": "Encontradas {{count}} categorias"
    },
    "metrics": {
      "success": "Métricas da API recuperadas com sucesso"
    },
    "errors": {
      "invalid_id": "Formato de ID inválido: {{id}}",
      "invalid_data": "Dados inválidos ou mal formados fornecidos",
      "missing_required_fields": "Campos obrigatórios ausentes: nome e preço devem ser fornecidos",
      "product_not_found": "Produto com ID {{id}} não encontrado"
    }
  },
  "health": {
    "healthy": "API está saudável e operacional",
    "unhealthy": "API está enfrentando problemas"
  }
}`

	// Spanish translations
	esContent := `{
  "api": {
    "products": {
      "list_success": "Se encontraron {{count}} productos",
      "get_success": "Producto recuperado exitosamente",
      "create_success": "Producto '{{name}}' creado exitosamente con ID {{id}}",
      "update_success": "Producto '{{name}}' (ID: {{id}}) actualizado exitosamente",
      "delete_success": "Producto '{{name}}' (ID: {{id}}) eliminado exitosamente"
    },
    "categories": {
      "list_success": "Se encontraron {{count}} categorías"
    },
    "metrics": {
      "success": "Métricas de API recuperadas exitosamente"
    },
    "errors": {
      "invalid_id": "Formato de ID inválido: {{id}}",
      "invalid_data": "Datos inválidos o mal formados proporcionados",
      "missing_required_fields": "Campos requeridos faltantes: nombre y precio deben ser proporcionados",
      "product_not_found": "Producto con ID {{id}} no encontrado"
    }
  },
  "health": {
    "healthy": "API está saludable y operacional",
    "unhealthy": "API está experimentando problemas"
  }
}`

	files := map[string]string{
		"en.json": enContent,
		"pt.json": ptContent,
		"es.json": esContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
