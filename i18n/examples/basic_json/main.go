// Package main demonstrates basic usage of the i18n library with JSON provider.
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
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

func main() {
	// Create a temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_example")
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

	// Create and register the JSON provider factory
	jsonFactory := &json.Factory{}
	registry := i18n.NewRegistry()
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		log.Fatal("Failed to register provider:", err)
	}

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

	// Demonstrate various translation features
	fmt.Println("=== I18n Library Demo ===")
	fmt.Println()

	// Basic translations
	fmt.Println("Basic Translations:")
	showTranslation(ctx, provider, "hello", "en", nil)
	showTranslation(ctx, provider, "hello", "pt", nil)
	showTranslation(ctx, provider, "hello", "es", nil)
	fmt.Println()

	// Translations with parameters
	fmt.Println("Translations with Parameters:")
	params := map[string]interface{}{
		"name": "João",
		"age":  30,
	}
	showTranslation(ctx, provider, "greeting", "en", params)
	showTranslation(ctx, provider, "greeting", "pt", params)
	showTranslation(ctx, provider, "greeting", "es", params)
	fmt.Println()

	// Nested key translations
	fmt.Println("Nested Key Translations:")
	showTranslation(ctx, provider, "user.profile.title", "en", nil)
	showTranslation(ctx, provider, "user.profile.title", "pt", nil)
	showTranslation(ctx, provider, "user.profile.title", "es", nil)
	fmt.Println()

	// Fallback behavior
	fmt.Println("Fallback Behavior (key exists only in English):")
	showTranslation(ctx, provider, "error.notfound", "pt", nil) // Falls back to English
	showTranslation(ctx, provider, "error.notfound", "es", nil) // Falls back to English
	fmt.Println()

	// Missing key behavior (non-strict mode)
	fmt.Println("Missing Key Behavior (non-strict mode):")
	showTranslation(ctx, provider, "nonexistent.key", "en", nil) // Returns key itself
	fmt.Println()

	// Provider information
	fmt.Println("Provider Information:")
	fmt.Printf("Supported Languages: %v\n", provider.GetSupportedLanguages())
	fmt.Printf("Default Language: %s\n", provider.GetDefaultLanguage())
	fmt.Printf("Has 'hello' in 'en': %t\n", provider.HasTranslation("hello", "en"))
	fmt.Printf("Has 'hello' in 'fr': %t\n", provider.HasTranslation("hello", "fr"))
	fmt.Println()

	// Health check
	fmt.Println("Health Check:")
	if err := provider.Health(ctx); err != nil {
		fmt.Printf("Provider is unhealthy: %v\n", err)
	} else {
		fmt.Println("Provider is healthy")
	}
}

func showTranslation(ctx context.Context, provider interfaces.I18n, key, lang string, params map[string]interface{}) {
	result, err := provider.Translate(ctx, key, lang, params)
	if err != nil {
		fmt.Printf("  %s [%s]: ERROR - %v\n", key, lang, err)
	} else {
		fmt.Printf("  %s [%s]: %s\n", key, lang, result)
	}
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "hello": "Hello",
  "goodbye": "Goodbye",
  "greeting": "Hello {{name}}, you are {{age}} years old!",
  "user": {
    "profile": {
      "title": "User Profile",
      "edit": "Edit Profile"
    },
    "settings": {
      "title": "Settings",
      "language": "Language"
    }
  },
  "error": {
    "notfound": "Not found",
    "unauthorized": "Unauthorized access"
  }
}`

	// Portuguese translations
	ptContent := `{
  "hello": "Olá",
  "goodbye": "Tchau",
  "greeting": "Olá {{name}}, você tem {{age}} anos!",
  "user": {
    "profile": {
      "title": "Perfil do Usuário",
      "edit": "Editar Perfil"
    },
    "settings": {
      "title": "Configurações",
      "language": "Idioma"
    }
  }
}`

	// Spanish translations
	esContent := `{
  "hello": "Hola",
  "goodbye": "Adiós",
  "greeting": "¡Hola {{name}}, tienes {{age}} años!",
  "user": {
    "profile": {
      "title": "Perfil de Usuario",
      "edit": "Editar Perfil"
    },
    "settings": {
      "title": "Configuración",
      "language": "Idioma"
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
