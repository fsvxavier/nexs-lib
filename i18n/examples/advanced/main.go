// Package main demonstrates advanced usage of the i18n library with hooks.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/hooks"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

func main() {
	// Create a temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_advanced_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create example translation files
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal(err)
	}

	// Configure the i18n system
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 5*time.Minute).
		WithLoadTimeout(10 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     tempDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
	if err != nil {
		log.Fatal("Failed to create configuration:", err)
	}

	// Create the registry
	registry := i18n.NewRegistry()

	// Register provider factory
	jsonFactory := &json.Factory{}
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		log.Fatal("Failed to register provider:", err)
	}

	// Add hooks for logging and metrics
	fmt.Println("=== Registering Hooks ===")

	// Logging hook
	loggingHook, err := hooks.NewLoggingHook("logging", 1, hooks.LoggingHookConfig{
		LogLevel:        "info",
		LogTranslations: true,
		LogErrors:       true,
	}, nil)
	if err != nil {
		log.Fatal("Failed to create logging hook:", err)
	}
	if err := registry.AddHook(loggingHook); err != nil {
		log.Fatal("Failed to register logging hook:", err)
	}
	fmt.Println("✓ Logging hook registered")

	// Metrics hook
	metricsHook, err := hooks.NewMetricsHook("metrics", 2, hooks.MetricsHookConfig{
		CollectTranslationMetrics: true,
		CollectErrorMetrics:       true,
		CollectPerformanceMetrics: true,
		MetricsInterval:           time.Minute,
	})
	if err != nil {
		log.Fatal("Failed to create metrics hook:", err)
	}
	if err := registry.AddHook(metricsHook); err != nil {
		log.Fatal("Failed to register metrics hook:", err)
	}
	fmt.Println("✓ Metrics hook registered")

	// Create the provider
	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		log.Fatal("Failed to create provider:", err)
	}

	// Start the provider
	ctx := context.Background()
	if err := provider.Start(ctx); err != nil {
		log.Fatal("Failed to start provider:", err)
	}
	defer provider.Stop(ctx)

	fmt.Println("\n=== I18n Library Advanced Demo ===")

	// Demonstrate different translation scenarios
	runTranslationDemo(ctx, provider)

	// Show hook behavior with error scenarios
	runErrorScenarios(ctx, provider)

	// Health check
	fmt.Println("\n=== System Health ===")
	if err := provider.Health(ctx); err != nil {
		fmt.Printf("Provider is unhealthy: %v\n", err)
	} else {
		fmt.Println("✓ Provider is healthy")
	}

	// Show registry information
	showRegistryInfo(registry)
}

func runTranslationDemo(ctx context.Context, provider interfaces.I18n) {
	fmt.Println("\n=== Translation Scenarios ===")

	// Basic translations
	fmt.Println("\n1. Basic Translations:")
	showTranslation(ctx, provider, "hello", "en", nil)
	showTranslation(ctx, provider, "hello", "pt", nil)
	showTranslation(ctx, provider, "hello", "es", nil)

	// Translations with parameters
	fmt.Println("\n2. Parameterized Translations:")
	params := map[string]interface{}{
		"name":  "Maria",
		"count": 3,
	}
	showTranslation(ctx, provider, "notification", "en", params)
	showTranslation(ctx, provider, "notification", "pt", params)

	// Nested key translations
	fmt.Println("\n3. Nested Key Translations:")
	showTranslation(ctx, provider, "api.errors.not_found", "en", nil)
	showTranslation(ctx, provider, "api.errors.not_found", "pt", nil)
	showTranslation(ctx, provider, "api.success.created", "en", nil)
	showTranslation(ctx, provider, "api.success.created", "pt", nil)
}

func runErrorScenarios(ctx context.Context, provider interfaces.I18n) {
	fmt.Println("\n=== Error Handling Scenarios ===")

	// Non-existent key
	fmt.Println("\n1. Non-existent Keys:")
	showTranslation(ctx, provider, "non.existent.key", "en", nil)
	showTranslation(ctx, provider, "another.missing.key", "pt", nil)

	// Unsupported language (should fallback to default)
	fmt.Println("\n2. Unsupported Language (fallback behavior):")
	showTranslation(ctx, provider, "hello", "fr", nil) // French not supported
}

func showTranslation(ctx context.Context, provider interfaces.I18n, key, lang string, params map[string]interface{}) {
	result, err := provider.Translate(ctx, key, lang, params)
	if err != nil {
		fmt.Printf("  %s [%s]: ERROR - %v\n", key, lang, err)
	} else {
		fmt.Printf("  %s [%s]: %s\n", key, lang, result)
	}
}

func showRegistryInfo(registry *i18n.Registry) {
	fmt.Println("\n=== Registry Information ===")

	// Show registered providers
	providerNames := registry.GetProviderNames()
	fmt.Printf("Registered Providers: %v\n", providerNames)

	// Show registered hooks
	hooks := registry.GetHooks()
	fmt.Printf("Registered Hooks: %d\n", len(hooks))
	for _, hook := range hooks {
		fmt.Printf("  - %s (priority: %d)\n", hook.Name(), hook.Priority())
	}

	// Show registered middlewares
	middlewares := registry.GetMiddlewares()
	fmt.Printf("Registered Middlewares: %d\n", len(middlewares))
	for _, middleware := range middlewares {
		fmt.Printf("  - %s\n", middleware.Name())
	}

	// Show active instances
	fmt.Printf("Active Provider Instances: %d\n", registry.GetActiveInstances())
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "hello": "Hello",
  "goodbye": "Goodbye",
  "notification": "Hello {{name}}, you have {{count}} new messages!",
  "api": {
    "errors": {
      "not_found": "Resource not found",
      "unauthorized": "Access denied",
      "server_error": "Internal server error"
    },
    "success": {
      "created": "Resource created successfully",
      "updated": "Resource updated successfully",
      "deleted": "Resource deleted successfully"
    }
  }
}`

	// Portuguese translations
	ptContent := `{
  "hello": "Olá",
  "goodbye": "Tchau",
  "notification": "Olá {{name}}, você tem {{count}} novas mensagens!",
  "api": {
    "errors": {
      "not_found": "Recurso não encontrado",
      "unauthorized": "Acesso negado",
      "server_error": "Erro interno do servidor"
    },
    "success": {
      "created": "Recurso criado com sucesso",
      "updated": "Recurso atualizado com sucesso",
      "deleted": "Recurso excluído com sucesso"
    }
  }
}`

	// Spanish translations
	esContent := `{
  "hello": "Hola",
  "goodbye": "Adiós",
  "notification": "Hola {{name}}, tienes {{count}} mensajes nuevos!",
  "api": {
    "errors": {
      "not_found": "Recurso no encontrado",
      "unauthorized": "Acceso denegado",
      "server_error": "Error interno del servidor"
    },
    "success": {
      "created": "Recurso creado exitosamente",
      "updated": "Recurso actualizado exitosamente",
      "deleted": "Recurso eliminado exitosamente"
    }
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
