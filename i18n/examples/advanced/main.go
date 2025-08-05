package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n"
)

func main() {
	// Create advanced configuration
	cfg := i18n.ProviderConfig{
		Type:               i18n.ProviderTypeBasic,
		TranslationsPath:   "./translations",
		TranslationsFormat: "json",
	}

	// Create provider using factory
	provider, err := i18n.NewProvider(cfg)
	if err != nil {
		panic(err)
	}

	// Load translations for multiple languages
	languages := []string{"pt-BR", "en", "es"}
	for _, lang := range languages {
		err = provider.LoadTranslations(cfg.TranslationsPath+"/"+lang, cfg.TranslationsFormat)
		if err != nil {
			panic(err)
		}
	}

	// Set supported languages
	err = provider.SetLanguages("pt-BR", "en", "es")
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
		result, err := provider.Translate("items_count", map[string]interface{}{
			"Count": count,
		})
		if err != nil {
			fmt.Printf("Error with plural translation: %v\n", err)
			continue
		}
		fmt.Printf("Count %d: %s\n", count, result)
	}
}
