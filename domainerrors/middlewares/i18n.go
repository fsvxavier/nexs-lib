package middlewares

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	i18nLib "github.com/fsvxavier/nexs-lib/i18n"
	i18nInterfaces "github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// I18nMiddleware gerencia middlewares de internacionalização usando nexs-lib/i18n
type I18nMiddleware struct {
	middlewares []interfaces.I18nMiddlewareFunc
	i18nClient  i18nInterfaces.I18n
	mu          sync.RWMutex
}

// NewI18nMiddleware cria um novo gerenciador de middlewares de i18n
func NewI18nMiddleware(i18nClient i18nInterfaces.I18n) *I18nMiddleware {
	return &I18nMiddleware{
		middlewares: make([]interfaces.I18nMiddlewareFunc, 0),
		i18nClient:  i18nClient,
	}
}

// NewI18nMiddlewareWithRegistry cria um novo middleware usando o registry do i18n
func NewI18nMiddlewareWithRegistry(registry *i18nLib.Registry, providerType string, config interface{}) (*I18nMiddleware, error) {
	provider, err := registry.CreateProvider(providerType, config)
	if err != nil {
		return nil, err
	}

	return &I18nMiddleware{
		middlewares: make([]interfaces.I18nMiddlewareFunc, 0),
		i18nClient:  provider,
	}, nil
}

// Register registra um middleware de i18n
func (m *I18nMiddleware) Register(middleware interfaces.I18nMiddlewareFunc) {
	if middleware == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = append(m.middlewares, middleware)
}

// Execute executa todos os middlewares de i18n registrados
func (m *I18nMiddleware) Execute(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := err

	// Executa os middlewares em sequência
	for _, middleware := range m.middlewares {
		result = middleware(ctx, result, locale, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})
	}

	return result
}

// Count retorna o número de middlewares registrados
func (m *I18nMiddleware) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.middlewares)
}

// Clear remove todos os middlewares registrados
func (m *I18nMiddleware) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = make([]interfaces.I18nMiddlewareFunc, 0)
}

// GetI18nClient retorna o cliente i18n utilizado
func (m *I18nMiddleware) GetI18nClient() i18nInterfaces.I18n {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.i18nClient
}

// SetI18nClient define o cliente i18n
func (m *I18nMiddleware) SetI18nClient(client i18nInterfaces.I18n) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.i18nClient = client
}

// Instância global para uso em toda a aplicação
var GlobalI18nMiddleware *I18nMiddleware

// InitializeGlobalI18nMiddleware inicializa o middleware global com um cliente i18n
func InitializeGlobalI18nMiddleware(i18nClient i18nInterfaces.I18n) {
	GlobalI18nMiddleware = NewI18nMiddleware(i18nClient)
}

// InitializeGlobalI18nMiddlewareWithRegistry inicializa o middleware global usando registry
func InitializeGlobalI18nMiddlewareWithRegistry(registry *i18nLib.Registry, providerType string, config interface{}) error {
	middleware, err := NewI18nMiddlewareWithRegistry(registry, providerType, config)
	if err != nil {
		return err
	}
	GlobalI18nMiddleware = middleware
	return nil
}

// RegisterI18nMiddleware registra um middleware de i18n globalmente
func RegisterI18nMiddleware(middleware interfaces.I18nMiddlewareFunc) {
	if GlobalI18nMiddleware != nil {
		GlobalI18nMiddleware.Register(middleware)
	}
}

// ExecuteI18nMiddlewares executa todos os middlewares de i18n globais
func ExecuteI18nMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	if GlobalI18nMiddleware != nil {
		return GlobalI18nMiddleware.Execute(ctx, err, locale)
	}
	return err
}

// GetI18nMiddlewareCount retorna o número de middlewares de i18n globais
func GetI18nMiddlewareCount() int {
	if GlobalI18nMiddleware != nil {
		return GlobalI18nMiddleware.Count()
	}
	return 0
}

// ClearI18nMiddlewares limpa todos os middlewares de i18n globais
func ClearI18nMiddlewares() {
	if GlobalI18nMiddleware != nil {
		GlobalI18nMiddleware.Clear()
	}
}

// TranslationMiddleware é um middleware que traduz mensagens de erro usando nexs-lib/i18n
func TranslationMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil || GlobalI18nMiddleware == nil || GlobalI18nMiddleware.i18nClient == nil {
		return next(err)
	}

	// Verifica se o locale é suportado
	supportedLanguages := GlobalI18nMiddleware.i18nClient.GetSupportedLanguages()
	isSupported := false
	for _, lang := range supportedLanguages {
		if lang == locale {
			isSupported = true
			break
		}
	}

	// Se não suportado, usa o idioma padrão
	targetLocale := locale
	if !isSupported {
		targetLocale = GlobalI18nMiddleware.i18nClient.GetDefaultLanguage()
	}

	// Enriquece o erro com informações de tradução
	enrichedErr := err.WithMetadata("middleware_translation_processed", true)
	enrichedErr = enrichedErr.WithMetadata("target_locale", targetLocale)
	enrichedErr = enrichedErr.WithMetadata("locale_supported", isSupported)

	// Tenta traduzir a mensagem
	translationKey := "error." + err.Code()
	translatedMessage, translateErr := GlobalI18nMiddleware.i18nClient.Translate(ctx, translationKey, targetLocale, nil)

	if translateErr == nil && translatedMessage != "" {
		enrichedErr = enrichedErr.WithMetadata("translated_message", translatedMessage)
		enrichedErr = enrichedErr.WithMetadata("translation_successful", true)
	} else {
		enrichedErr = enrichedErr.WithMetadata("translation_successful", false)
		if translateErr != nil {
			enrichedErr = enrichedErr.WithMetadata("translation_error", translateErr.Error())
		}
	}

	return next(enrichedErr)
}

// LocaleValidationMiddleware valida se o locale é suportado usando nexs-lib/i18n
func LocaleValidationMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil || GlobalI18nMiddleware == nil || GlobalI18nMiddleware.i18nClient == nil {
		return next(err)
	}

	// Obter idiomas suportados do cliente i18n
	supportedLanguages := GlobalI18nMiddleware.i18nClient.GetSupportedLanguages()

	// Verifica se o locale é suportado
	isSupported := false
	for _, supported := range supportedLanguages {
		if supported == locale {
			isSupported = true
			break
		}
	}

	processedErr := err.WithMetadata("locale_validation_processed", true)
	processedErr = processedErr.WithMetadata("locale_supported", isSupported)
	processedErr = processedErr.WithMetadata("supported_locales", supportedLanguages)

	if !isSupported {
		fallbackLocale := GlobalI18nMiddleware.i18nClient.GetDefaultLanguage()
		processedErr = processedErr.WithMetadata("fallback_locale", fallbackLocale)
		processedErr = processedErr.WithMetadata("locale_validation_failed", true)
	}

	return next(processedErr)
}

// EnrichmentI18nMiddleware adiciona informações contextuais de i18n ao erro
func EnrichmentI18nMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil || GlobalI18nMiddleware == nil || GlobalI18nMiddleware.i18nClient == nil {
		return next(err)
	}

	// Adiciona informações contextuais do sistema i18n
	enrichedErr := err.WithMetadata("i18n_middleware_enriched", true)
	enrichedErr = enrichedErr.WithMetadata("i18n_provider_type", "nexs-lib/i18n")
	enrichedErr = enrichedErr.WithMetadata("default_language", GlobalI18nMiddleware.i18nClient.GetDefaultLanguage())
	enrichedErr = enrichedErr.WithMetadata("total_translations", GlobalI18nMiddleware.i18nClient.GetTranslationCount())
	enrichedErr = enrichedErr.WithMetadata("loaded_languages", GlobalI18nMiddleware.i18nClient.GetLoadedLanguages())
	enrichedErr = enrichedErr.WithMetadata("requested_locale", locale)

	// Verifica se existe tradução para este erro específico
	translationKey := "error." + err.Code()
	hasTranslation := GlobalI18nMiddleware.i18nClient.HasTranslation(translationKey, locale)
	enrichedErr = enrichedErr.WithMetadata("has_translation_for_locale", hasTranslation)

	if !hasTranslation {
		// Verifica se existe no idioma padrão
		defaultLang := GlobalI18nMiddleware.i18nClient.GetDefaultLanguage()
		hasDefaultTranslation := GlobalI18nMiddleware.i18nClient.HasTranslation(translationKey, defaultLang)
		enrichedErr = enrichedErr.WithMetadata("has_default_translation", hasDefaultTranslation)
	}

	return next(enrichedErr)
}

// FallbackLanguageMiddleware aplica um idioma de fallback se necessário
func FallbackLanguageMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil || GlobalI18nMiddleware == nil || GlobalI18nMiddleware.i18nClient == nil {
		return next(err)
	}

	// Verifica se o locale é suportado
	supportedLanguages := GlobalI18nMiddleware.i18nClient.GetSupportedLanguages()
	isSupported := false
	for _, lang := range supportedLanguages {
		if lang == locale {
			isSupported = true
			break
		}
	}

	processedErr := err.WithMetadata("fallback_middleware_processed", true)

	if !isSupported {
		fallbackLanguage := GlobalI18nMiddleware.i18nClient.GetDefaultLanguage()
		processedErr = processedErr.WithMetadata("fallback_language", fallbackLanguage)
		processedErr = processedErr.WithMetadata("original_requested_locale", locale)
		processedErr = processedErr.WithMetadata("fallback_applied", true)

		// Tenta traduzir usando o idioma de fallback
		translationKey := "error." + err.Code()
		fallbackMessage, translateErr := GlobalI18nMiddleware.i18nClient.Translate(ctx, translationKey, fallbackLanguage, nil)

		if translateErr == nil && fallbackMessage != "" {
			processedErr = processedErr.WithMetadata("fallback_translation", fallbackMessage)
			processedErr = processedErr.WithMetadata("fallback_translation_successful", true)
		}
	} else {
		processedErr = processedErr.WithMetadata("fallback_applied", false)
	}

	return next(processedErr)
}

// CreateCustomI18nMiddleware cria um middleware customizado com parâmetros específicos
func CreateCustomI18nMiddleware(translationPrefix string, includeMetadata bool) interfaces.I18nMiddlewareFunc {
	return func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		if err == nil || GlobalI18nMiddleware == nil || GlobalI18nMiddleware.i18nClient == nil {
			return next(err)
		}

		processedErr := err

		if includeMetadata {
			processedErr = processedErr.WithMetadata("custom_i18n_middleware", true)
			processedErr = processedErr.WithMetadata("translation_prefix", translationPrefix)
		}

		// Usa o prefixo customizado para a chave de tradução
		translationKey := translationPrefix + "." + err.Code()
		translatedMessage, translateErr := GlobalI18nMiddleware.i18nClient.Translate(ctx, translationKey, locale, nil)

		if translateErr == nil && translatedMessage != "" {
			processedErr = processedErr.WithMetadata("custom_translated_message", translatedMessage)
		}

		return next(processedErr)
	}
}
