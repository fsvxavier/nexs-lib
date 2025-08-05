package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/hooks"
	"github.com/fsvxavier/nexs-lib/i18n/providers"
)

// LogHook implementa a interface hooks.Observer para logging
type LogHook struct{}

func (h *LogHook) OnMissingTranslation(key string, language string) {
	log.Printf("[HOOK] Tradução não encontrada: chave='%s', idioma='%s'", key, language)
}

func (h *LogHook) OnTranslationLoaded(language string, format string) {
	log.Printf("[HOOK] Traduções carregadas: idioma='%s', formato='%s'", language, format)
}

func main() {
	fmt.Println("=== Exemplo de Hooks ===")
	fmt.Println("Este exemplo demonstra o uso de hooks para monitorar eventos de tradução")

	// Criar configuração
	cfg := config.DefaultConfig().
		WithTranslationsPath("./translations").
		WithTranslationsFormat("json")

	// Criar provider
	provider, err := providers.CreateProvider(cfg.TranslationsFormat)
	if err != nil {
		log.Fatal(err)
	}

	// Registrar hook personalizado
	hooks.RegisterHook(&LogHook{})

	fmt.Println("\n1. Carregando traduções (observe o hook OnTranslationLoaded):")
	err = provider.LoadTranslations(cfg.GetTranslationFilePath("pt-BR"), cfg.TranslationsFormat)
	if err != nil {
		log.Fatal(err)
	}

	err = provider.SetLanguages("pt-BR")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n2. Testando tradução existente:")
	result, err := provider.Translate("hello_world", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Tradução: %s\n", result)

	fmt.Println("\n3. Testando tradução inexistente (observe o hook OnMissingTranslation):")
	_, err = provider.Translate("chave_inexistente", nil)
	if err != nil {
		fmt.Printf("✓ Erro esperado: %v\n", err)
	}

	fmt.Println("\n4. Testando tradução com variáveis:")
	result, err = provider.Translate("hello_name", map[string]interface{}{
		"Name": "Maria",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Tradução: %s\n", result)
}
