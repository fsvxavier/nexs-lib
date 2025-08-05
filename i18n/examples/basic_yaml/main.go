// Package main demonstrates basic usage of the i18n library with YAML provider.
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
	"github.com/fsvxavier/nexs-lib/i18n/providers/yaml"
)

func main() {
	// Create a temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_yaml_example")
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
		WithProviderConfig(&config.YAMLProviderConfig{
			FilePath:     tempDir,
			FilePattern:  "{lang}.yaml",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateYAML: true,
		}).
		Build()
	if err != nil {
		log.Fatal("Failed to create configuration:", err)
	}

	// Create and register the YAML provider factory
	yamlFactory := &yaml.Factory{}
	registry := i18n.NewRegistry()
	if err := registry.RegisterProvider(yamlFactory); err != nil {
		log.Fatal("Failed to register provider:", err)
	}

	// Create the provider
	provider, err := registry.CreateProvider("yaml", cfg)
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
	fmt.Println("=== I18n Library Demo (YAML) ===")
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
		"name": "Maria",
		"age":  25,
	}
	showTranslation(ctx, provider, "greeting", "en", params)
	showTranslation(ctx, provider, "greeting", "pt", params)
	showTranslation(ctx, provider, "greeting", "es", params)
	fmt.Println()

	// Nested key translations
	fmt.Println("Nested Key Translations:")
	showTranslation(ctx, provider, "app.navigation.home", "en", nil)
	showTranslation(ctx, provider, "app.navigation.home", "pt", nil)
	showTranslation(ctx, provider, "app.navigation.home", "es", nil)
	fmt.Println()

	// Complex nested structures
	fmt.Println("Complex Nested Structures:")
	showTranslation(ctx, provider, "forms.validation.required", "en", nil)
	showTranslation(ctx, provider, "forms.validation.required", "pt", nil)
	showTranslation(ctx, provider, "forms.validation.required", "es", nil)
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
	enContent := `hello: Hello
goodbye: Goodbye
greeting: "Hello {{name}}, you are {{age}} years old!"

app:
  title: My Application
  navigation:
    home: Home
    about: About
    contact: Contact
    profile: Profile

forms:
  buttons:
    save: Save
    cancel: Cancel
    delete: Delete
  validation:
    required: This field is required
    email: Please enter a valid email address
    password: Password must be at least 8 characters

messages:
  success: Operation completed successfully
  error: An error occurred
  warning: Please check your input`

	// Portuguese translations
	ptContent := `hello: Olá
goodbye: Tchau
greeting: "Olá {{name}}, você tem {{age}} anos!"

app:
  title: Minha Aplicação
  navigation:
    home: Início
    about: Sobre
    contact: Contato
    profile: Perfil

forms:
  buttons:
    save: Salvar
    cancel: Cancelar
    delete: Excluir
  validation:
    required: Este campo é obrigatório
    email: Por favor, insira um endereço de email válido
    password: A senha deve ter pelo menos 8 caracteres

messages:
  success: Operação concluída com sucesso
  error: Ocorreu um erro
  warning: Por favor, verifique sua entrada`

	// Spanish translations
	esContent := `hello: Hola
goodbye: Adiós
greeting: "¡Hola {{name}}, tienes {{age}} años!"

app:
  title: Mi Aplicación
  navigation:
    home: Inicio
    about: Acerca de
    contact: Contacto
    profile: Perfil

forms:
  buttons:
    save: Guardar
    cancel: Cancelar
    delete: Eliminar
  validation:
    required: Este campo es obligatorio
    email: Por favor, ingresa una dirección de email válida
    password: La contraseña debe tener al menos 8 caracteres

messages:
  success: Operación completada exitosamente
  error: Ocurrió un error
  warning: Por favor, verifica tu entrada`

	files := map[string]string{
		"en.yaml": enContent,
		"pt.yaml": ptContent,
		"es.yaml": esContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
