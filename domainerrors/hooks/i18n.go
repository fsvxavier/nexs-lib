package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	i18nLib "github.com/fsvxavier/nexs-lib/i18n"
	i18nInterfaces "github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// I18nHookManager gerencia hooks de internacionalização usando nexs-lib/i18n
type I18nHookManager struct {
	hooks      []interfaces.I18nHookFunc
	i18nClient i18nInterfaces.I18n
	mu         sync.RWMutex
}

// NewI18nHookManager cria um novo gerenciador de hooks de i18n
func NewI18nHookManager(i18nClient i18nInterfaces.I18n) *I18nHookManager {
	return &I18nHookManager{
		hooks:      make([]interfaces.I18nHookFunc, 0),
		i18nClient: i18nClient,
	}
}

// NewI18nHookManagerWithRegistry cria um novo gerenciador usando o registry do i18n
func NewI18nHookManagerWithRegistry(registry *i18nLib.Registry, providerType string, config interface{}) (*I18nHookManager, error) {
	provider, err := registry.CreateProvider(providerType, config)
	if err != nil {
		return nil, err
	}

	return &I18nHookManager{
		hooks:      make([]interfaces.I18nHookFunc, 0),
		i18nClient: provider,
	}, nil
}

// Register registra um hook de i18n
func (m *I18nHookManager) Register(hook interfaces.I18nHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hook)
}

// Execute executa todos os hooks de i18n registrados
func (m *I18nHookManager) Execute(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.hooks {
		if hookErr := hook(ctx, err, locale); hookErr != nil {
			return hookErr
		}
	}

	return nil
}

// Count retorna o número de hooks registrados
func (m *I18nHookManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.hooks)
}

// Clear remove todos os hooks registrados
func (m *I18nHookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = make([]interfaces.I18nHookFunc, 0)
}

// GetI18nClient retorna o cliente i18n utilizado
func (m *I18nHookManager) GetI18nClient() i18nInterfaces.I18n {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.i18nClient
}

// SetI18nClient define o cliente i18n
func (m *I18nHookManager) SetI18nClient(client i18nInterfaces.I18n) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.i18nClient = client
}

// Instância global para uso em toda a aplicação
var GlobalI18nHookManager *I18nHookManager

// InitializeGlobalI18nHookManager inicializa o gerenciador global com um cliente i18n
func InitializeGlobalI18nHookManager(i18nClient i18nInterfaces.I18n) {
	GlobalI18nHookManager = NewI18nHookManager(i18nClient)
}

// InitializeGlobalI18nHookManagerWithRegistry inicializa o gerenciador global usando registry
func InitializeGlobalI18nHookManagerWithRegistry(registry *i18nLib.Registry, providerType string, config interface{}) error {
	manager, err := NewI18nHookManagerWithRegistry(registry, providerType, config)
	if err != nil {
		return err
	}
	GlobalI18nHookManager = manager
	return nil
}

// RegisterI18nHook registra um hook de i18n globalmente
func RegisterI18nHook(hook interfaces.I18nHookFunc) {
	if GlobalI18nHookManager != nil {
		GlobalI18nHookManager.Register(hook)
	}
}

// ExecuteI18nHooks executa todos os hooks de i18n globais
func ExecuteI18nHooks(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if GlobalI18nHookManager != nil {
		return GlobalI18nHookManager.Execute(ctx, err, locale)
	}
	return nil
}

// GetI18nHookCount retorna o número de hooks de i18n globais
func GetI18nHookCount() int {
	if GlobalI18nHookManager != nil {
		return GlobalI18nHookManager.Count()
	}
	return 0
}

// ClearI18nHooks limpa todos os hooks de i18n globais
func ClearI18nHooks() {
	if GlobalI18nHookManager != nil {
		GlobalI18nHookManager.Clear()
	}
}

// TranslateErrorMessageHook é um hook que traduz mensagens de erro usando nexs-lib/i18n
func TranslateErrorMessageHook(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if GlobalI18nHookManager == nil || GlobalI18nHookManager.i18nClient == nil {
		return nil
	}

	// Verifica se o locale é suportado
	supportedLanguages := GlobalI18nHookManager.i18nClient.GetSupportedLanguages()
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
		targetLocale = GlobalI18nHookManager.i18nClient.GetDefaultLanguage()
	}

	// Tenta traduzir a mensagem de erro usando uma chave baseada no código do erro
	translationKey := "error." + err.Code()
	translatedMessage, translateErr := GlobalI18nHookManager.i18nClient.Translate(ctx, translationKey, targetLocale, nil)

	if translateErr == nil && translatedMessage != "" {
		// Se conseguiu traduzir, adiciona como metadata
		err.WithMetadata("translated_message", translatedMessage)
		err.WithMetadata("translation_locale", targetLocale)
		err.WithMetadata("original_message", err.Error())
	}

	// Adiciona informações de processamento i18n
	err.WithMetadata("i18n_processed", true)
	err.WithMetadata("requested_locale", locale)
	err.WithMetadata("locale_supported", isSupported)

	return nil
}

// LoggingI18nHook é um hook de exemplo que registra informações de i18n usando nexs-lib/i18n
func LoggingI18nHook(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if GlobalI18nHookManager == nil {
		return nil
	}

	// Adiciona informações de logging específicas para i18n
	err.WithMetadata("i18n_hook_executed", true)
	err.WithMetadata("i18n_provider_available", GlobalI18nHookManager.i18nClient != nil)

	if GlobalI18nHookManager.i18nClient != nil {
		err.WithMetadata("supported_languages", GlobalI18nHookManager.i18nClient.GetSupportedLanguages())
		err.WithMetadata("default_language", GlobalI18nHookManager.i18nClient.GetDefaultLanguage())
		err.WithMetadata("translation_count", GlobalI18nHookManager.i18nClient.GetTranslationCount())
	}

	return nil
}

// FallbackLanguageHook define um idioma de fallback se o solicitado não for suportado
func FallbackLanguageHook(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if GlobalI18nHookManager == nil || GlobalI18nHookManager.i18nClient == nil {
		return nil
	}

	// Verifica se o locale é suportado
	supportedLanguages := GlobalI18nHookManager.i18nClient.GetSupportedLanguages()
	isSupported := false
	for _, lang := range supportedLanguages {
		if lang == locale {
			isSupported = true
			break
		}
	}

	if !isSupported {
		fallbackLanguage := GlobalI18nHookManager.i18nClient.GetDefaultLanguage()
		err.WithMetadata("fallback_language", fallbackLanguage)
		err.WithMetadata("original_requested_locale", locale)
		err.WithMetadata("locale_fallback_applied", true)
	}

	return nil
}
