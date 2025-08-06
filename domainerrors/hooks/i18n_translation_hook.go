package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainInterfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n"
	i18nConfig "github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

// I18nTranslationHook hook para tradução de mensagens de erro usando i18n
type I18nTranslationHook struct {
	name           string
	hookType       domainInterfaces.HookType
	priority       int
	enabled        bool
	i18nRegistry   *i18n.Registry
	i18nProvider   interfaces.I18n
	defaultLang    string
	fallbackLang   string
	translateCodes bool // se deve traduzir codes além de messages
}

// I18nTranslationConfig configurações do hook de tradução
type I18nTranslationConfig struct {
	TranslationsPath   string                            // caminho para arquivos de tradução
	DefaultLanguage    string                            // idioma padrão
	FallbackLanguage   string                            // idioma de fallback
	SupportedLangs     []string                          // idiomas suportados
	FilePattern        string                            // padrão dos arquivos (ex: "{{.Lang}}.json")
	TranslateCodes     bool                              // se deve traduzir codes além de messages
	CustomTranslations map[string]map[string]interface{} // traduções customizadas
}

// NewI18nTranslationHook cria um novo hook de tradução i18n
func NewI18nTranslationHook(hookType domainInterfaces.HookType, config I18nTranslationConfig) (*I18nTranslationHook, error) {
	if config.DefaultLanguage == "" {
		config.DefaultLanguage = "en"
	}
	if config.FallbackLanguage == "" {
		config.FallbackLanguage = "en"
	}
	if config.FilePattern == "" {
		config.FilePattern = "{{.Lang}}.json"
	}
	if len(config.SupportedLangs) == 0 {
		config.SupportedLangs = []string{"en", "pt", "es"}
	}
	if config.TranslationsPath == "" {
		config.TranslationsPath = "./translations"
	}

	// Configuração do i18n
	i18nCfg := &i18nConfig.Config{
		DefaultLanguage:    config.DefaultLanguage,
		SupportedLanguages: config.SupportedLangs,
		LoadTimeout:        30 * time.Second,
		CacheEnabled:       true,
		CacheTTL:           1 * time.Hour,
		FallbackToDefault:  true,
		StrictMode:         false,
		ProviderConfig: &i18nConfig.JSONProviderConfig{
			FilePath:    config.TranslationsPath,
			FilePattern: config.FilePattern,
		},
	}

	// Cria o registry i18n
	registry := i18n.NewRegistry()

	// Registra o factory JSON
	jsonFactory := json.NewFactory()
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		return nil, fmt.Errorf("failed to register JSON provider: %w", err)
	}

	// Cria o provider i18n
	provider, err := registry.CreateProvider("json", i18nCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create i18n provider: %w", err)
	}

	hook := &I18nTranslationHook{
		name:           fmt.Sprintf("i18n_translation_hook_%s", hookType),
		hookType:       hookType,
		priority:       50, // alta prioridade para traduzir antes de outros hooks processarem
		enabled:        true,
		i18nRegistry:   registry,
		i18nProvider:   provider,
		defaultLang:    config.DefaultLanguage,
		fallbackLang:   config.FallbackLanguage,
		translateCodes: config.TranslateCodes,
	}

	return hook, nil
}

// Name retorna o nome do hook
func (h *I18nTranslationHook) Name() string {
	return h.name
}

// Execute executa o hook de tradução
func (h *I18nTranslationHook) Execute(ctx *domainInterfaces.HookContext) error {
	if ctx.Error == nil {
		return nil
	}

	// Detecta o idioma do contexto
	lang := h.detectLanguage(ctx.Context)

	// Traduz a mensagem principal
	translatedMessage := h.translateMessage(ctx.Context, ctx.Error.Message, ctx.Error.Code, lang)
	if translatedMessage != "" {
		// Salva a mensagem original como metadado
		if ctx.Error.Metadata == nil {
			ctx.Error.Metadata = make(map[string]interface{})
		}
		ctx.Error.Metadata["original_message"] = ctx.Error.Message
		ctx.Error.Metadata["translated_language"] = lang
		ctx.Error.Metadata["translation_source"] = "i18n_hook"

		// Atualiza a mensagem com a tradução
		ctx.Error.Message = translatedMessage
	}

	// Traduz o código se configurado para tal
	if h.translateCodes && ctx.Error.Code != "" {
		translatedCode := h.translateCode(ctx.Context, ctx.Error.Code, lang)
		if translatedCode != "" {
			if ctx.Error.Metadata == nil {
				ctx.Error.Metadata = make(map[string]interface{})
			}
			ctx.Error.Metadata["original_code"] = ctx.Error.Code
			ctx.Error.Code = translatedCode
		}
	}

	// Traduz metadados específicos se existirem
	h.translateMetadata(ctx.Context, ctx.Error.Metadata, lang)

	return nil
}

// detectLanguage detecta o idioma do contexto
func (h *I18nTranslationHook) detectLanguage(ctx context.Context) string {
	// Tenta detectar idioma do contexto
	if ctx != nil {
		// Verifica por Accept-Language header
		if lang, ok := ctx.Value("Accept-Language").(string); ok && lang != "" {
			// Extrai o primeiro idioma da string Accept-Language
			if parsed := h.parseAcceptLanguage(lang); parsed != "" {
				return parsed
			}
		}

		// Verifica por idioma definido no contexto
		if lang, ok := ctx.Value("language").(string); ok && lang != "" {
			return lang
		}

		// Verifica por locale do usuário
		if locale, ok := ctx.Value("user_locale").(string); ok && locale != "" {
			return strings.Split(locale, "_")[0] // extrai idioma do locale (ex: pt_BR -> pt)
		}
	}

	return h.defaultLang
}

// parseAcceptLanguage parseia string Accept-Language e retorna o primeiro idioma válido
func (h *I18nTranslationHook) parseAcceptLanguage(acceptLang string) string {
	// Exemplo: "pt-BR,pt;q=0.9,en;q=0.8" -> "pt"
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		firstLang := strings.TrimSpace(parts[0])
		langCode := strings.Split(firstLang, "-")[0]
		return strings.ToLower(langCode)
	}
	return ""
}

// translateMessage traduz uma mensagem usando diferentes estratégias
func (h *I18nTranslationHook) translateMessage(ctx context.Context, message, code, lang string) string {
	// Estratégia 1: Usar o código do erro como chave
	if code != "" {
		if translated := h.tryTranslate(ctx, code, lang); translated != "" {
			return translated
		}

		// Tenta com prefixo
		key := fmt.Sprintf("error.%s", code)
		if translated := h.tryTranslate(ctx, key, lang); translated != "" {
			return translated
		}
	}

	// Estratégia 2: Usar a mensagem como chave (normalizada)
	normalizedMessage := h.normalizeMessageKey(message)
	if translated := h.tryTranslate(ctx, normalizedMessage, lang); translated != "" {
		return translated
	}

	// Estratégia 3: Buscar por chaves comuns baseadas na mensagem
	commonKeys := h.generateCommonKeys(message, code)
	for _, key := range commonKeys {
		if translated := h.tryTranslate(ctx, key, lang); translated != "" {
			return translated
		}
	}

	return "" // não encontrou tradução
}

// translateCode traduz um código de erro
func (h *I18nTranslationHook) translateCode(ctx context.Context, code, lang string) string {
	// Tenta traduzir o código diretamente
	key := fmt.Sprintf("code.%s", code)
	return h.tryTranslate(ctx, key, lang)
}

// translateMetadata traduz campos específicos dos metadados
func (h *I18nTranslationHook) translateMetadata(ctx context.Context, metadata map[string]interface{}, lang string) {
	if metadata == nil {
		return
	}

	// Campos que podem conter mensagens traduzíveis
	translatableFields := []string{
		"validation_message",
		"business_rule_message",
		"constraint_message",
		"field_error",
		"detail_message",
	}

	for _, field := range translatableFields {
		if value, ok := metadata[field].(string); ok && value != "" {
			if translated := h.tryTranslate(ctx, value, lang); translated != "" {
				metadata[field] = translated
				metadata[fmt.Sprintf("%s_original", field)] = value
			}
		}
	}

	// Traduz arrays de mensagens
	if messages, ok := metadata["validation_errors"].([]string); ok {
		translatedMessages := make([]string, len(messages))
		hasTranslations := false

		for i, msg := range messages {
			if translated := h.tryTranslate(ctx, msg, lang); translated != "" {
				translatedMessages[i] = translated
				hasTranslations = true
			} else {
				translatedMessages[i] = msg
			}
		}

		if hasTranslations {
			metadata["validation_errors_original"] = messages
			metadata["validation_errors"] = translatedMessages
		}
	}
}

// tryTranslate tenta traduzir uma chave, retornando string vazia se não encontrar
func (h *I18nTranslationHook) tryTranslate(ctx context.Context, key, lang string) string {
	if ctx == nil {
		ctx = context.Background()
	}

	translated, err := h.i18nProvider.Translate(ctx, key, lang, nil)
	if err != nil || translated == key {
		// Tenta com idioma de fallback se o idioma solicitado falhar
		if lang != h.fallbackLang {
			translated, err = h.i18nProvider.Translate(ctx, key, h.fallbackLang, nil)
			if err != nil || translated == key {
				return ""
			}
		} else {
			return ""
		}
	}

	return translated
}

// normalizeMessageKey normaliza uma mensagem para ser usada como chave
func (h *I18nTranslationHook) normalizeMessageKey(message string) string {
	// Remove espaços extras e converte para lowercase
	normalized := strings.TrimSpace(strings.ToLower(message))
	// Remove pontuação e substitui espaços por underscores
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, "!", "")
	normalized = strings.ReplaceAll(normalized, "?", "")
	return normalized
}

// generateCommonKeys gera chaves comuns baseadas na mensagem e código
func (h *I18nTranslationHook) generateCommonKeys(message, code string) []string {
	var keys []string

	// Chaves baseadas no tipo de erro (detectado da mensagem)
	messageLower := strings.ToLower(message)
	if strings.Contains(messageLower, "not found") {
		keys = append(keys, "error.not_found", "errors.not_found")
	}
	if strings.Contains(messageLower, "validation") {
		keys = append(keys, "error.validation", "errors.validation_failed")
	}
	if strings.Contains(messageLower, "unauthorized") {
		keys = append(keys, "error.unauthorized", "errors.access_denied")
	}
	if strings.Contains(messageLower, "forbidden") {
		keys = append(keys, "error.forbidden", "errors.forbidden")
	}
	if strings.Contains(messageLower, "internal") {
		keys = append(keys, "error.internal", "errors.internal_server_error")
	}

	// Adiciona variações do código se fornecido
	if code != "" {
		keys = append(keys,
			fmt.Sprintf("errors.%s", strings.ToLower(code)),
			fmt.Sprintf("messages.%s", strings.ToLower(code)),
		)
	}

	return keys
}

// Type retorna o tipo do hook
func (h *I18nTranslationHook) Type() domainInterfaces.HookType {
	return h.hookType
}

// Priority retorna a prioridade de execução (menor número = maior prioridade)
func (h *I18nTranslationHook) Priority() int {
	return h.priority
}

// Enabled indica se o hook está habilitado
func (h *I18nTranslationHook) Enabled() bool {
	return h.enabled
}

// SetEnabled habilita/desabilita o hook
func (h *I18nTranslationHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// SetPriority define a prioridade do hook
func (h *I18nTranslationHook) SetPriority(priority int) {
	h.priority = priority
}

// GetSupportedLanguages retorna os idiomas suportados
func (h *I18nTranslationHook) GetSupportedLanguages() []string {
	return h.i18nProvider.GetSupportedLanguages()
}

// UpdateTranslation atualiza uma tradução específica
func (h *I18nTranslationHook) UpdateTranslation(lang, key, value string) error {
	// Para uso em runtime, poderia usar um provider de memória
	// Esta é uma implementação básica - em produção poderia ser mais sofisticada
	return fmt.Errorf("runtime translation updates not implemented - use configuration")
}
