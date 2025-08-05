package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

func main() {
	// JSON Example
	jsonExample()

	fmt.Print("\n---\n")

	// YAML Example
	yamlExample()
}

func jsonExample() {
	fmt.Println("JSON Example:")
	cfg := config.DefaultConfig().
		WithTranslationsPath("./translations").
		WithTranslationsFormat("json")

	provider, _ := providers.CreateProvider("json")
	provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), "json")
	provider.SetLanguages("pt-BR")

	result, _ := provider.Translate("hello_world", nil)
	fmt.Println("Simple:", result)
}

func yamlExample() {
	fmt.Println("YAML Example:")
	cfg := config.DefaultConfig().
		WithTranslationsPath("./translations").
		WithTranslationsFormat("yaml")

	provider, _ := providers.CreateProvider("yaml")
	provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), "yaml")
	provider.SetLanguages("pt-BR")

	result, _ := provider.Translate("hello_world", nil)
	fmt.Println("Simple:", result)
}
