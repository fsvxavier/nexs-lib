package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
	// Create JSON provider directly for advanced features
	provider := providers.NewJSONProvider()

	// Load translations for multiple languages
	languages := []string{"pt-BR", "en", "es"}
	for _, lang := range languages {
		filename := fmt.Sprintf("./translations/translations.%s.json", lang)
		err := provider.LoadTranslations(filename, "json")
		if err != nil {
			panic(fmt.Errorf("failed to load translations: %w", err))
		}
	}

	// Set supported languages
	err := provider.SetLanguages("pt-BR", "en", "es")
	if err != nil {
		panic(err)
	}

	// Demonstrate translations in different languages
	messages := []struct {
		lang string
		key  string
		vars map[string]interface{}
	}{
		{
			lang: "pt-BR",
			key:  "welcome_message",
			vars: map[string]interface{}{
				"Name": "João",
				"Time": "manhã",
			},
		},
		{
			lang: "en",
			key:  "welcome_message",
			vars: map[string]interface{}{
				"Name": "John",
				"Time": "morning",
			},
		},
		{
			lang: "es",
			key:  "welcome_message",
			vars: map[string]interface{}{
				"Name": "Juan",
				"Time": "mañana",
			},
		},
	}

	// Print translations
	for _, msg := range messages {
		provider.SetLanguages(msg.lang) // Set current language
		result, err := provider.Translate(msg.key, msg.vars)
		if err != nil {
			fmt.Printf("Error translating to %s: %v\n", msg.lang, err)
			continue
		}
		fmt.Printf("%s: %s\n", msg.lang, result)
	}

	// Demonstrate plural translations
	counts := []int{0, 1, 2, 5}
	for _, count := range counts {
		result, err := provider.TranslatePlural("items_count", count, map[string]interface{}{
			"Count": count,
		})
		if err != nil {
			fmt.Printf("Error with plural translation: %v\n", err)
			continue
		}
		fmt.Printf("Count %d: %s\n", count, result)
	}
}
