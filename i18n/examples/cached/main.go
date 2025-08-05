package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n"
)

func main() {
	// Create configuration with cache
	cfg := i18n.ProviderConfig{
		Type:               i18n.ProviderTypeCached,
		TranslationsPath:   "./translations",
		TranslationsFormat: "json",
		Languages:          []string{"pt-BR", "en"},
		CacheTTL:           24,
	}

	// Create cached provider using factory
	provider, err := i18n.NewProvider(cfg)
	if err != nil {
		panic(err)
	}

	// First translation (will be cached)
	result, err := provider.Translate("hello_world", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("First translation (not cached):", result)

	// Second translation (will use cache)
	result, err = provider.Translate("hello_world", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Second translation (from cache):", result)

	// Translation with variables (will be cached with variables)
	result, err = provider.Translate("hello_name", map[string]interface{}{
		"Name": "John",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("With variables:", result)

	// Change language (will clear relevant cache entries)
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
