package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
	// Create configuration
	cfg := config.DefaultConfig().
		WithTranslationsPath("./translations").
		WithTranslationsFormat("json")

	// Create provider
	provider, err := providers.CreateProvider(cfg.TranslationsFormat)
	if err != nil {
		panic(err)
	}

	// Load translations
	err = provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), cfg.TranslationsFormat)
	if err != nil {
		panic(err)
	}

	// Set languages
	err = provider.SetLanguages("pt-BR", "en")
	if err != nil {
		panic(err)
	}

	// Simple translation
	result, err := provider.Translate("hello_world", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Simple translation:", result)

	// Translation with variables
	result, err = provider.Translate("hello_name", map[string]interface{}{
		"Name": "John",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("With variables:", result)

	// Plural translation
	result, err = provider.TranslatePlural("users_count", 2, map[string]interface{}{
		"Count": 2,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Plural form:", result)
}
