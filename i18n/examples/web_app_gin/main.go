// Package main demonstrates a complete web application using i18n with Gin framework.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
	"github.com/gin-gonic/gin"
)

// User represents a user model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ResponseData represents API response structure
type ResponseData struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// I18nService holds the i18n provider for dependency injection
type I18nService struct {
	provider interfaces.I18n
}

func main() {
	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_gin_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create translation files
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal("Failed to create translation files:", err)
	}

	// Setup i18n service
	i18nService, err := setupI18nService(tempDir)
	if err != nil {
		log.Fatal("Failed to setup i18n service:", err)
	}

	// Setup Gin router
	router := setupRouter(i18nService)

	// Start server
	fmt.Println("🌍 Gin Web App with i18n running on http://localhost:8080")
	fmt.Println("📖 Available endpoints:")
	fmt.Println("  GET  /               - Home page")
	fmt.Println("  GET  /api/users      - List users")
	fmt.Println("  POST /api/users      - Create user")
	fmt.Println("  GET  /api/users/:id  - Get user by ID")
	fmt.Println("  GET  /health         - Health check")
	fmt.Println("  GET  /lang/:lang     - Change language")
	fmt.Println("")
	fmt.Println("💡 Use Accept-Language header or ?lang=pt query parameter")
	fmt.Println("   Supported languages: en, pt, es")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupI18nService(translationDir string) (*I18nService, error) {
	// Configure i18n
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 10*time.Minute).
		WithLoadTimeout(15 * time.Second).
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

	return &I18nService{provider: provider}, nil
}

func setupRouter(i18nService *I18nService) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Add i18n middleware
	router.Use(i18nMiddleware(i18nService))

	// Home page
	router.GET("/", homeHandler)

	// API routes
	api := router.Group("/api")
	{
		api.GET("/users", listUsersHandler)
		api.POST("/users", createUserHandler)
		api.GET("/users/:id", getUserHandler)
	}

	// Health check
	router.GET("/health", healthHandler)

	// Language switcher
	router.GET("/lang/:lang", languageHandler)

	return router
}

// i18nMiddleware adds i18n context to Gin
func i18nMiddleware(i18nService *I18nService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine language from query param, header, or default
		lang := c.Query("lang")
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
			if lang == "" {
				lang = "en"
			}
		}

		// Store i18n service and language in context
		c.Set("i18n", i18nService)
		c.Set("lang", lang)
		c.Next()
	}
}

// Helper function to translate text
func translate(c *gin.Context, key string, params map[string]interface{}) string {
	i18nService := c.MustGet("i18n").(*I18nService)
	lang := c.MustGet("lang").(string)

	result, err := i18nService.provider.Translate(c.Request.Context(), key, lang, params)
	if err != nil {
		log.Printf("Translation error for key '%s' in language '%s': %v", key, lang, err)
		return key // Return key as fallback
	}
	return result
}

// Handlers
func homeHandler(c *gin.Context) {
	message := translate(c, "welcome.title", nil)
	description := translate(c, "welcome.description", nil)

	c.JSON(http.StatusOK, ResponseData{
		Success: true,
		Message: message,
		Data: map[string]string{
			"description": description,
			"language":    c.MustGet("lang").(string),
		},
	})
}

func listUsersHandler(c *gin.Context) {
	users := []User{
		{ID: 1, Name: "João Silva", Email: "joao@example.com"},
		{ID: 2, Name: "Maria Santos", Email: "maria@example.com"},
		{ID: 3, Name: "Carlos Oliveira", Email: "carlos@example.com"},
	}

	message := translate(c, "api.users.list_success", map[string]interface{}{
		"count": len(users),
	})

	c.JSON(http.StatusOK, ResponseData{
		Success: true,
		Message: message,
		Data:    users,
	})
}

func createUserHandler(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		message := translate(c, "api.errors.invalid_data", nil)
		c.JSON(http.StatusBadRequest, ResponseData{
			Success: false,
			Message: message,
		})
		return
	}

	// Simulate user creation
	user.ID = 4

	message := translate(c, "api.users.create_success", map[string]interface{}{
		"name": user.Name,
	})

	c.JSON(http.StatusCreated, ResponseData{
		Success: true,
		Message: message,
		Data:    user,
	})
}

func getUserHandler(c *gin.Context) {
	id := c.Param("id")

	// Simulate user lookup
	if id == "1" {
		user := User{ID: 1, Name: "João Silva", Email: "joao@example.com"}
		message := translate(c, "api.users.get_success", nil)

		c.JSON(http.StatusOK, ResponseData{
			Success: true,
			Message: message,
			Data:    user,
		})
		return
	}

	message := translate(c, "api.errors.user_not_found", map[string]interface{}{
		"id": id,
	})

	c.JSON(http.StatusNotFound, ResponseData{
		Success: false,
		Message: message,
	})
}

func healthHandler(c *gin.Context) {
	i18nService := c.MustGet("i18n").(*I18nService)

	// Check i18n provider health
	if err := i18nService.provider.Health(c.Request.Context()); err != nil {
		message := translate(c, "health.unhealthy", nil)
		c.JSON(http.StatusServiceUnavailable, ResponseData{
			Success: false,
			Message: message,
		})
		return
	}

	message := translate(c, "health.healthy", nil)
	c.JSON(http.StatusOK, ResponseData{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"timestamp":           time.Now().UTC(),
			"supported_languages": i18nService.provider.GetSupportedLanguages(),
			"default_language":    i18nService.provider.GetDefaultLanguage(),
		},
	})
}

func languageHandler(c *gin.Context) {
	lang := c.Param("lang")
	i18nService := c.MustGet("i18n").(*I18nService)

	// Check if language is supported
	supported := false
	for _, supportedLang := range i18nService.provider.GetSupportedLanguages() {
		if supportedLang == lang {
			supported = true
			break
		}
	}

	if !supported {
		message := translate(c, "api.errors.unsupported_language", map[string]interface{}{
			"language": lang,
		})
		c.JSON(http.StatusBadRequest, ResponseData{
			Success: false,
			Message: message,
		})
		return
	}

	// Set language for current request
	c.Set("lang", lang)
	message := translate(c, "api.language.changed", map[string]interface{}{
		"language": lang,
	})

	c.JSON(http.StatusOK, ResponseData{
		Success: true,
		Message: message,
		Data: map[string]string{
			"current_language": lang,
		},
	})
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "welcome": {
    "title": "Welcome to our i18n Web App!",
    "description": "This is a complete web application demonstrating internationalization with Gin framework."
  },
  "api": {
    "users": {
      "list_success": "Found {{count}} users",
      "create_success": "User '{{name}}' created successfully",
      "get_success": "User retrieved successfully"
    },
    "language": {
      "changed": "Language changed to {{language}}"
    },
    "errors": {
      "invalid_data": "Invalid data provided",
      "user_not_found": "User with ID {{id}} not found",
      "unsupported_language": "Language '{{language}}' is not supported"
    }
  },
  "health": {
    "healthy": "Service is healthy",
    "unhealthy": "Service is unhealthy"
  }
}`

	// Portuguese translations
	ptContent := `{
  "welcome": {
    "title": "Bem-vindo ao nosso Web App i18n!",
    "description": "Esta é uma aplicação web completa demonstrando internacionalização com framework Gin."
  },
  "api": {
    "users": {
      "list_success": "Encontrados {{count}} usuários",
      "create_success": "Usuário '{{name}}' criado com sucesso",
      "get_success": "Usuário recuperado com sucesso"
    },
    "language": {
      "changed": "Idioma alterado para {{language}}"
    },
    "errors": {
      "invalid_data": "Dados inválidos fornecidos",
      "user_not_found": "Usuário com ID {{id}} não encontrado",
      "unsupported_language": "Idioma '{{language}}' não é suportado"
    }
  },
  "health": {
    "healthy": "Serviço está saudável",
    "unhealthy": "Serviço não está saudável"
  }
}`

	// Spanish translations
	esContent := `{
  "welcome": {
    "title": "¡Bienvenido a nuestra Web App i18n!",
    "description": "Esta es una aplicación web completa que demuestra la internacionalización con el framework Gin."
  },
  "api": {
    "users": {
      "list_success": "Se encontraron {{count}} usuarios",
      "create_success": "Usuario '{{name}}' creado exitosamente",
      "get_success": "Usuario recuperado exitosamente"
    },
    "language": {
      "changed": "Idioma cambiado a {{language}}"
    },
    "errors": {
      "invalid_data": "Datos inválidos proporcionados",
      "user_not_found": "Usuario con ID {{id}} no encontrado",
      "unsupported_language": "El idioma '{{language}}' no es compatible"
    }
  },
  "health": {
    "healthy": "El servicio está saludable",
    "unhealthy": "El servicio no está saludable"
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
