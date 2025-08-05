package main

import (
	"log"
	"net/http"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/middleware"
	"github.com/fsvxavier/nexs-lib/i18n/providers"
	"golang.org/x/text/language"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling request to / with method %s", r.Method)
	provider, ok := middleware.GetProvider(r.Context())
	if !ok {
		println("Provider not found in context")
		http.Error(w, "i18n provider not found", http.StatusInternalServerError)
		return
	}

	result, err := provider.Translate("welcome", nil)
	if err != nil {
		log.Printf("Error translating: %v\n\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "` + result + `"}`))
}

func handlerWithVars(w http.ResponseWriter, r *http.Request) {
	provider, ok := middleware.GetProvider(r.Context())
	if !ok {
		http.Error(w, "i18n provider not found", http.StatusInternalServerError)
		return
	}

	result, err := provider.Translate("greeting", map[string]interface{}{
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

	result, err := provider.TranslatePlural("items", 2, map[string]interface{}{
		"Count": 2,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(result))
}

func main() {
	log.Println("Starting server setup...")

	// Create configuration
	cfg := config.DefaultConfig().
		WithTranslationsPath("/home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/i18n/examples/http/translations").
		WithTranslationsFormat("json")

	// Create provider
	provider, err := providers.CreateProvider(cfg.TranslationsFormat)
	if err != nil {
		panic(err)
	}

	// Load translations
	err = provider.LoadTranslations("/home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/i18n/examples/http/translations/translations.pt-BR.json", cfg.TranslationsFormat)
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
	mux := http.NewServeMux()
	mux.Handle("/", i18nMiddleware(http.HandlerFunc(handler)))
	mux.Handle("/with-vars", i18nMiddleware(http.HandlerFunc(handlerWithVars)))
	mux.Handle("/plural", i18nMiddleware(http.HandlerFunc(handlerPlural)))

	// Start server
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	log.Println("Server running on http://localhost:3000")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
