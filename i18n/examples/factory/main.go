package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n"
)

func main() {
	// Create configuration
	cfg := i18n.ProviderConfig{
		Type:               i18n.ProviderTypeBasic,
		TranslationsPath:   "./translations",
		TranslationsFormat: "json",
		Languages:          []string{"pt-BR", "en"},
	}

	// Create provider using factory
	provider, err := i18n.NewProvider(cfg)
	if err != nil {
		panic(err)
	}

	// Load translations for each language
	err = provider.LoadTranslations(cfg.TranslationsPath, cfg.TranslationsFormat)
	if err != nil {
		panic(err)
	}

	// Set supported languages
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

	// Change language
	err = provider.SetLanguages("en")
	if err != nil {
		panic(err)
	}

	// Translation in different language
	result, err = provider.Translate("hello_world", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("English translation:", result)
}
