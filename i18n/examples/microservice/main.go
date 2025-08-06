// Package main demonstrates i18n in a microservice architecture.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	jsonProvider "github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

// ServiceConfig represents microservice configuration
type ServiceConfig struct {
	ServiceName     string   `json:"service_name"`
	Port            string   `json:"port"`
	Environment     string   `json:"environment"`
	DefaultLanguage string   `json:"default_language"`
	Languages       []string `json:"languages"`
}

// MicroserviceI18n represents the i18n microservice
type MicroserviceI18n struct {
	config   ServiceConfig
	provider interfaces.I18n
	server   *http.Server
}

// ServiceStatus represents the service health status
type ServiceStatus struct {
	Service      string         `json:"service"`
	Status       string         `json:"status"`
	Version      string         `json:"version"`
	Environment  string         `json:"environment"`
	Languages    []string       `json:"languages"`
	Uptime       string         `json:"uptime"`
	Translations map[string]int `json:"translations"`
	Timestamp    time.Time      `json:"timestamp"`
}

// TranslationRequest represents a translation request
type TranslationRequest struct {
	Key        string                 `json:"key"`
	Language   string                 `json:"language"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// TranslationResponse represents a translation response
type TranslationResponse struct {
	Success     bool      `json:"success"`
	Translation string    `json:"translation,omitempty"`
	Error       string    `json:"error,omitempty"`
	Language    string    `json:"language"`
	Key         string    `json:"key"`
	Timestamp   time.Time `json:"timestamp"`
}

// BatchTranslationRequest represents a batch translation request
type BatchTranslationRequest struct {
	Requests []TranslationRequest `json:"requests"`
}

// BatchTranslationResponse represents a batch translation response
type BatchTranslationResponse struct {
	Success   bool                  `json:"success"`
	Results   []TranslationResponse `json:"results"`
	Total     int                   `json:"total"`
	Processed int                   `json:"processed"`
	Errors    int                   `json:"errors"`
	Duration  string                `json:"duration"`
	Timestamp time.Time             `json:"timestamp"`
}

var startTime = time.Now()

func main() {
	// Load service configuration
	serviceConfig := ServiceConfig{
		ServiceName:     "i18n-microservice",
		Port:            "8080",
		Environment:     "development",
		DefaultLanguage: "en",
		Languages:       []string{"en", "pt", "es", "fr", "de"},
	}

	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_microservice")
	if err != nil {
		log.Fatal("Failed to create temp directory:", err)
	}
	defer os.RemoveAll(tempDir)

	// Create translation files
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal("Failed to create translation files:", err)
	}

	// Initialize microservice
	service, err := NewMicroserviceI18n(serviceConfig, tempDir)
	if err != nil {
		log.Fatal("Failed to initialize microservice:", err)
	}

	// Start the microservice
	fmt.Printf("🚀 Starting %s microservice on port %s\n", serviceConfig.ServiceName, serviceConfig.Port)
	fmt.Printf("🌍 Environment: %s\n", serviceConfig.Environment)
	fmt.Printf("📖 Supported languages: %v\n", serviceConfig.Languages)
	fmt.Printf("🔗 Health check: http://localhost:%s/health\n", serviceConfig.Port)
	fmt.Printf("🔗 API docs: http://localhost:%s/api/docs\n", serviceConfig.Port)
	fmt.Println()

	if err := service.Start(); err != nil {
		log.Fatal("Failed to start microservice:", err)
	}
}

func NewMicroserviceI18n(config ServiceConfig, translationDir string) (*MicroserviceI18n, error) {
	// Configure i18n provider
	cfg, err := buildI18nConfig(config, translationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create i18n configuration: %w", err)
	}

	// Create registry and register provider
	registry := i18n.NewRegistry()
	jsonFactory := &jsonProvider.Factory{}
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Create provider
	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Start provider
	ctx := context.Background()
	if err := provider.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start provider: %w", err)
	}

	service := &MicroserviceI18n{
		config:   config,
		provider: provider,
	}

	return service, nil
}

func buildI18nConfig(serviceConfig ServiceConfig, translationDir string) (interface{}, error) {
	return config.NewConfigBuilder().
		WithSupportedLanguages(serviceConfig.Languages...).
		WithDefaultLanguage(serviceConfig.DefaultLanguage).
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 15*time.Minute).
		WithLoadTimeout(30 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     translationDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
}

func config_builder(serviceConfig ServiceConfig, translationDir string) (interface{}, error) {
	return config.NewConfigBuilder().
		WithSupportedLanguages(serviceConfig.Languages...).
		WithDefaultLanguage(serviceConfig.DefaultLanguage).
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 15*time.Minute).
		WithLoadTimeout(30 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     translationDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
}

func (s *MicroserviceI18n) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Translation endpoints
	mux.HandleFunc("/api/v1/translate", s.translateHandler)
	mux.HandleFunc("/api/v1/translate/batch", s.batchTranslateHandler)

	// Service information endpoints
	mux.HandleFunc("/api/v1/languages", s.languagesHandler)
	mux.HandleFunc("/api/v1/status", s.statusHandler)

	// Documentation endpoint
	mux.HandleFunc("/api/docs", s.docsHandler)

	// Metrics endpoint
	mux.HandleFunc("/metrics", s.metricsHandler)

	// Add CORS middleware
	handler := corsMiddleware(mux)

	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

func (s *MicroserviceI18n) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.provider != nil {
		return s.provider.Stop(ctx)
	}

	return nil
}

// HTTP Handlers

func (s *MicroserviceI18n) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := ServiceStatus{
		Service:      s.config.ServiceName,
		Status:       "healthy",
		Version:      "1.0.0",
		Environment:  s.config.Environment,
		Languages:    s.provider.GetSupportedLanguages(),
		Uptime:       time.Since(startTime).String(),
		Translations: make(map[string]int),
		Timestamp:    time.Now().UTC(),
	}

	// Check provider health
	if err := s.provider.Health(ctx); err != nil {
		status.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Get translation counts by language
	for _, lang := range s.provider.GetSupportedLanguages() {
		status.Translations[lang] = s.provider.GetTranslationCountByLanguage(lang)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *MicroserviceI18n) translateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := TranslationResponse{
			Success:   false,
			Error:     "Invalid JSON payload",
			Timestamp: time.Now().UTC(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if req.Key == "" || req.Language == "" {
		response := TranslationResponse{
			Success:   false,
			Error:     "Key and language are required",
			Timestamp: time.Now().UTC(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Perform translation
	ctx := r.Context()
	translation, err := s.provider.Translate(ctx, req.Key, req.Language, req.Parameters)

	response := TranslationResponse{
		Success:   err == nil,
		Key:       req.Key,
		Language:  req.Language,
		Timestamp: time.Now().UTC(),
	}

	if err != nil {
		response.Error = err.Error()
		w.WriteHeader(http.StatusNotFound)
	} else {
		response.Translation = translation
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *MicroserviceI18n) batchTranslateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var batchReq BatchTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	start := time.Now()
	ctx := r.Context()

	var results []TranslationResponse
	errorCount := 0

	for _, req := range batchReq.Requests {
		translation, err := s.provider.Translate(ctx, req.Key, req.Language, req.Parameters)

		result := TranslationResponse{
			Success:   err == nil,
			Key:       req.Key,
			Language:  req.Language,
			Timestamp: time.Now().UTC(),
		}

		if err != nil {
			result.Error = err.Error()
			errorCount++
		} else {
			result.Translation = translation
		}

		results = append(results, result)
	}

	response := BatchTranslationResponse{
		Success:   errorCount == 0,
		Results:   results,
		Total:     len(batchReq.Requests),
		Processed: len(results),
		Errors:    errorCount,
		Duration:  time.Since(start).String(),
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *MicroserviceI18n) languagesHandler(w http.ResponseWriter, r *http.Request) {
	languages := map[string]interface{}{
		"supported_languages": s.provider.GetSupportedLanguages(),
		"default_language":    s.provider.GetDefaultLanguage(),
		"loaded_languages":    s.provider.GetLoadedLanguages(),
		"total_translations":  s.provider.GetTranslationCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(languages)
}

func (s *MicroserviceI18n) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"service":             s.config.ServiceName,
		"version":             "1.0.0",
		"environment":         s.config.Environment,
		"uptime":              time.Since(startTime).String(),
		"supported_languages": s.provider.GetSupportedLanguages(),
		"default_language":    s.provider.GetDefaultLanguage(),
		"translation_stats": map[string]interface{}{
			"total_translations": s.provider.GetTranslationCount(),
			"loaded_languages":   s.provider.GetLoadedLanguages(),
		},
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *MicroserviceI18n) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// In a real microservice, this would return Prometheus metrics
	metrics := map[string]interface{}{
		"http_requests_total":      0, // Would track actual metrics
		"translation_requests":     0,
		"translation_errors":       0,
		"translation_cache_hits":   0,
		"translation_cache_misses": 0,
		"uptime_seconds":           time.Since(startTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (s *MicroserviceI18n) docsHandler(w http.ResponseWriter, r *http.Request) {
	docs := `
<!DOCTYPE html>
<html>
<head>
    <title>I18n Microservice API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1, h2 { color: #333; }
        pre { background: #f4f4f4; padding: 10px; border-radius: 5px; }
        .endpoint { margin: 20px 0; }
    </style>
</head>
<body>
    <h1>I18n Microservice API Documentation</h1>
    
    <div class="endpoint">
        <h2>Health Check</h2>
        <p><strong>GET</strong> /health</p>
        <p>Returns service health status and statistics.</p>
    </div>

    <div class="endpoint">
        <h2>Single Translation</h2>
        <p><strong>POST</strong> /api/v1/translate</p>
        <pre>{
  "key": "user.profile.title",
  "language": "pt",
  "parameters": {
    "name": "João"
  }
}</pre>
    </div>

    <div class="endpoint">
        <h2>Batch Translation</h2>
        <p><strong>POST</strong> /api/v1/translate/batch</p>
        <pre>{
  "requests": [
    {
      "key": "hello",
      "language": "en"
    },
    {
      "key": "goodbye", 
      "language": "pt"
    }
  ]
}</pre>
    </div>

    <div class="endpoint">
        <h2>Languages</h2>
        <p><strong>GET</strong> /api/v1/languages</p>
        <p>Returns supported languages and statistics.</p>
    </div>

    <div class="endpoint">
        <h2>Service Status</h2>
        <p><strong>GET</strong> /api/v1/status</p>
        <p>Returns detailed service status information.</p>
    </div>

    <div class="endpoint">
        <h2>Metrics</h2>
        <p><strong>GET</strong> /metrics</p>
        <p>Returns service metrics for monitoring.</p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(docs))
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "microservice": {
    "welcome": "Welcome to the I18n Microservice",
    "description": "This microservice provides internationalization capabilities",
    "version": "Version {{version}} running in {{environment}} mode"
  },
  "api": {
    "translation": {
      "success": "Translation completed successfully",
      "error": "Translation failed",
      "not_found": "Translation key '{{key}}' not found for language '{{language}}'"
    },
    "batch": {
      "processing": "Processing batch of {{count}} translations",
      "completed": "Batch processing completed: {{processed}} of {{total}}",
      "partial_success": "Batch partially completed with {{errors}} errors"
    }
  },
  "user": {
    "profile": {
      "title": "User Profile",
      "welcome": "Welcome, {{name}}!",
      "settings": "Profile Settings"
    }
  },
  "common": {
    "hello": "Hello",
    "goodbye": "Goodbye",
    "yes": "Yes",
    "no": "No",
    "save": "Save",
    "cancel": "Cancel",
    "loading": "Loading...",
    "error": "Error",
    "success": "Success"
  }
}`

	// Portuguese translations
	ptContent := `{
  "microservice": {
    "welcome": "Bem-vindo ao Microserviço I18n",
    "description": "Este microserviço fornece capacidades de internacionalização",
    "version": "Versão {{version}} executando em modo {{environment}}"
  },
  "api": {
    "translation": {
      "success": "Tradução concluída com sucesso",
      "error": "Tradução falhou",
      "not_found": "Chave de tradução '{{key}}' não encontrada para idioma '{{language}}'"
    },
    "batch": {
      "processing": "Processando lote de {{count}} traduções",
      "completed": "Processamento em lote concluído: {{processed}} de {{total}}",
      "partial_success": "Lote parcialmente concluído com {{errors}} erros"
    }
  },
  "user": {
    "profile": {
      "title": "Perfil do Usuário",
      "welcome": "Bem-vindo, {{name}}!",
      "settings": "Configurações do Perfil"
    }
  },
  "common": {
    "hello": "Olá",
    "goodbye": "Tchau",
    "yes": "Sim",
    "no": "Não",
    "save": "Salvar",
    "cancel": "Cancelar",
    "loading": "Carregando...",
    "error": "Erro",
    "success": "Sucesso"
  }
}`

	// Spanish translations
	esContent := `{
  "microservice": {
    "welcome": "Bienvenido al Microservicio I18n",
    "description": "Este microservicio proporciona capacidades de internacionalización",
    "version": "Versión {{version}} ejecutándose en modo {{environment}}"
  },
  "api": {
    "translation": {
      "success": "Traducción completada exitosamente",
      "error": "Traducción falló",
      "not_found": "Clave de traducción '{{key}}' no encontrada para idioma '{{language}}'"
    },
    "batch": {
      "processing": "Procesando lote de {{count}} traducciones",
      "completed": "Procesamiento por lotes completado: {{processed}} de {{total}}",
      "partial_success": "Lote parcialmente completado con {{errors}} errores"
    }
  },
  "user": {
    "profile": {
      "title": "Perfil de Usuario",
      "welcome": "¡Bienvenido, {{name}}!",
      "settings": "Configuraciones del Perfil"
    }
  },
  "common": {
    "hello": "Hola",
    "goodbye": "Adiós",
    "yes": "Sí",
    "no": "No",
    "save": "Guardar",
    "cancel": "Cancelar",
    "loading": "Cargando...",
    "error": "Error",
    "success": "Éxito"
  }
}`

	// French translations
	frContent := `{
  "microservice": {
    "welcome": "Bienvenue au Microservice I18n",
    "description": "Ce microservice fournit des capacités d'internationalisation",
    "version": "Version {{version}} fonctionnant en mode {{environment}}"
  },
  "api": {
    "translation": {
      "success": "Traduction terminée avec succès",
      "error": "Échec de la traduction",
      "not_found": "Clé de traduction '{{key}}' non trouvée pour la langue '{{language}}'"
    },
    "batch": {
      "processing": "Traitement d'un lot de {{count}} traductions",
      "completed": "Traitement par lot terminé: {{processed}} sur {{total}}",
      "partial_success": "Lot partiellement terminé avec {{errors}} erreurs"
    }
  },
  "user": {
    "profile": {
      "title": "Profil Utilisateur",
      "welcome": "Bienvenue, {{name}}!",
      "settings": "Paramètres du Profil"
    }
  },
  "common": {
    "hello": "Bonjour",
    "goodbye": "Au revoir",
    "yes": "Oui",
    "no": "Non",
    "save": "Enregistrer",
    "cancel": "Annuler",
    "loading": "Chargement...",
    "error": "Erreur",
    "success": "Succès"
  }
}`

	// German translations
	deContent := `{
  "microservice": {
    "welcome": "Willkommen zum I18n Microservice",
    "description": "Dieser Microservice bietet Internationalisierungsfähigkeiten",
    "version": "Version {{version}} läuft im {{environment}}-Modus"
  },
  "api": {
    "translation": {
      "success": "Übersetzung erfolgreich abgeschlossen",
      "error": "Übersetzung fehlgeschlagen",
      "not_found": "Übersetzungsschlüssel '{{key}}' nicht gefunden für Sprache '{{language}}'"
    },
    "batch": {
      "processing": "Verarbeitung eines Stapels von {{count}} Übersetzungen",
      "completed": "Stapelverarbeitung abgeschlossen: {{processed}} von {{total}}",
      "partial_success": "Stapel teilweise abgeschlossen mit {{errors}} Fehlern"
    }
  },
  "user": {
    "profile": {
      "title": "Benutzerprofil",
      "welcome": "Willkommen, {{name}}!",
      "settings": "Profil-Einstellungen"
    }
  },
  "common": {
    "hello": "Hallo",
    "goodbye": "Auf Wiedersehen",
    "yes": "Ja",
    "no": "Nein",
    "save": "Speichern",
    "cancel": "Abbrechen",
    "loading": "Laden...",
    "error": "Fehler",
    "success": "Erfolg"
  }
}`

	files := map[string]string{
		"en.json": enContent,
		"pt.json": ptContent,
		"es.json": esContent,
		"fr.json": frContent,
		"de.json": deContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
