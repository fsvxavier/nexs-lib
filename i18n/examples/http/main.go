package main

import (
	"net/http"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/middleware"
	"github.com/fsvxavier/nexs-lib/i18n/providers"
	"golang.org/x/text/language"
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

	// Configure middleware
	i18nMiddleware := middleware.New(middleware.Config{
		Provider:        provider,
		QueryParam:      "lang",
		DefaultLanguage: language.MustParse("pt-BR"),
	})

	// Create handlers
	http.Handle("/", i18nMiddleware(http.HandlerFunc(handler)))
	http.Handle("/with-vars", i18nMiddleware(http.HandlerFunc(handlerWithVars)))
	http.Handle("/plural", i18nMiddleware(http.HandlerFunc(handlerPlural)))

	// Start server
	println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	provider, ok := middleware.GetProvider(r.Context())
	if !ok {
		http.Error(w, "i18n provider not found", http.StatusInternalServerError)
		return
	}

	result, err := provider.Translate("hello_world", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}

func handlerWithVars(w http.ResponseWriter, r *http.Request) {
	provider, ok := middleware.GetProvider(r.Context())
	if !ok {
		http.Error(w, "i18n provider not found", http.StatusInternalServerError)
		return
	}

	result, err := provider.Translate("hello_name", map[string]interface{}{
		"Name": "John",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}

func handlerPlural(w http.ResponseWriter, r *http.Request) {
	provider, ok := middleware.GetProvider(r.Context())
	if !ok {
		http.Error(w, "i18n provider not found", http.StatusInternalServerError)
		return
	}

	result, err := provider.TranslatePlural("users_count", 2, map[string]interface{}{
		"Count": 2,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}
