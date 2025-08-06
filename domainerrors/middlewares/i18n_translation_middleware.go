package middlewares

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

// I18nTranslationMiddleware middleware para tradução de mensagens de erro usando i18n
type I18nTranslationMiddleware struct {
	name           string
	middlewareType domainInterfaces.MiddlewareType
	priority       int
	enabled        bool
	i18nRegistry   *i18n.Registry
	i18nProvider   interfaces.I18n
	defaultLang    string
	fallbackLang   string
	translateCodes bool // se deve traduzir codes além de messages
	translateMeta  bool // se deve traduzir metadados
}

// I18nTranslationConfig configurações do middleware de tradução
type I18nTranslationConfig struct {
	TranslationsPath   string                            // caminho para arquivos de tradução
	DefaultLanguage    string                            // idioma padrão
	FallbackLanguage   string                            // idioma de fallback
	SupportedLangs     []string                          // idiomas suportados
	FilePattern        string                            // padrão dos arquivos (ex: "{lang}.json")
	TranslateCodes     bool                              // se deve traduzir codes além de messages
	TranslateMetadata  bool                              // se deve traduzir metadados
	CustomTranslations map[string]map[string]interface{} // traduções customizadas
}

// NewI18nTranslationMiddleware cria um novo middleware de tradução i18n
func NewI18nTranslationMiddleware(config I18nTranslationConfig) (*I18nTranslationMiddleware, error) {
	if config.DefaultLanguage == "" {
		config.DefaultLanguage = "en"
	}
	if config.FallbackLanguage == "" {
		config.FallbackLanguage = "en"
	}
	if config.FilePattern == "" {
		config.FilePattern = "{lang}.json"
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

	middleware := &I18nTranslationMiddleware{
		name:           "i18n_translation_middleware",
		middlewareType: domainInterfaces.MiddlewareTypeTransformation,
		priority:       10, // alta prioridade para traduzir cedo na cadeia
		enabled:        true,
		i18nRegistry:   registry,
		i18nProvider:   provider,
		defaultLang:    config.DefaultLanguage,
		fallbackLang:   config.FallbackLanguage,
		translateCodes: config.TranslateCodes,
		translateMeta:  config.TranslateMetadata,
	}

	return middleware, nil
}

// Name retorna o nome do middleware
func (m *I18nTranslationMiddleware) Name() string {
	return m.name
}

// Handle processa o middleware de tradução
func (m *I18nTranslationMiddleware) Handle(ctx *domainInterfaces.MiddlewareContext, next domainInterfaces.NextFunction) error {
	// Primeiro executa a tradução se há um erro no contexto
	if ctx.Error != nil {
		if err := m.translateError(ctx); err != nil {
			// Log do erro, mas não interrompe a cadeia
			if ctx.Metadata == nil {
				ctx.Metadata = make(map[string]interface{})
			}
			ctx.Metadata["i18n_translation_error"] = err.Error()
		}
	}

	// Chama o próximo middleware na cadeia
	if next != nil {
		return next(ctx)
	}

	return nil
}

// translateError realiza a tradução do erro no contexto
func (m *I18nTranslationMiddleware) translateError(ctx *domainInterfaces.MiddlewareContext) error {
	if ctx.Error == nil {
		return nil
	}

	// Detecta o idioma do contexto
	lang := m.detectLanguage(ctx.Context)

	// Traduz a mensagem principal
	translatedMessage := m.translateMessage(ctx.Context, ctx.Error.Message, ctx.Error.Code, lang)
	if translatedMessage != "" {
		// Salva a mensagem original como metadado
		if ctx.Error.Metadata == nil {
			ctx.Error.Metadata = make(map[string]interface{})
		}

		// Preserva informações de tradução
		ctx.Error.Metadata["original_message"] = ctx.Error.Message
		ctx.Error.Metadata["translated_language"] = lang
		ctx.Error.Metadata["translation_source"] = "i18n_middleware"
		ctx.Error.Metadata["translation_timestamp"] = ctx.Timestamp

		// Atualiza a mensagem com a tradução
		ctx.Error.Message = translatedMessage
	}

	// Traduz o código se configurado para tal
	if m.translateCodes && ctx.Error.Code != "" {
		translatedCode := m.translateCode(ctx.Context, ctx.Error.Code, lang)
		if translatedCode != "" {
			if ctx.Error.Metadata == nil {
				ctx.Error.Metadata = make(map[string]interface{})
			}
			ctx.Error.Metadata["original_code"] = ctx.Error.Code
			ctx.Error.Code = translatedCode
		}
	}

	// Traduz metadados específicos se configurado
	if m.translateMeta {
		m.translateMetadata(ctx.Context, ctx.Error.Metadata, lang)
	}

	// Também traduz metadados do contexto se existirem
	if m.translateMeta && ctx.Metadata != nil {
		m.translateContextMetadata(ctx.Context, ctx.Metadata, lang)
	}

	return nil
}

// detectLanguage detecta o idioma do contexto
func (m *I18nTranslationMiddleware) detectLanguage(ctx context.Context) string {
	// Tenta detectar idioma do contexto
	if ctx != nil {
		// Verifica por Accept-Language header
		if lang, ok := ctx.Value("Accept-Language").(string); ok && lang != "" {
			if parsed := m.parseAcceptLanguage(lang); parsed != "" {
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

		// Verifica por preferência do usuário
		if lang, ok := ctx.Value("user_language").(string); ok && lang != "" {
			return lang
		}
	}

	return m.defaultLang
}

// parseAcceptLanguage parseia string Accept-Language e retorna o primeiro idioma válido
func (m *I18nTranslationMiddleware) parseAcceptLanguage(acceptLang string) string {
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
func (m *I18nTranslationMiddleware) translateMessage(ctx context.Context, message, code, lang string) string {
	// Estratégia 1: Usar o código do erro como chave
	if code != "" {
		if translated := m.tryTranslate(ctx, code, lang); translated != "" {
			return translated
		}

		// Tenta com prefixos comuns
		prefixes := []string{"error.", "errors.", "messages.", "msg."}
		for _, prefix := range prefixes {
			key := fmt.Sprintf("%s%s", prefix, strings.ToLower(code))
			if translated := m.tryTranslate(ctx, key, lang); translated != "" {
				return translated
			}
		}
	}

	// Estratégia 2: Usar a mensagem como chave (normalizada)
	normalizedMessage := m.normalizeMessageKey(message)
	if translated := m.tryTranslate(ctx, normalizedMessage, lang); translated != "" {
		return translated
	}

	// Estratégia 3: Buscar por chaves comuns baseadas na mensagem
	commonKeys := m.generateCommonKeys(message, code)
	for _, key := range commonKeys {
		if translated := m.tryTranslate(ctx, key, lang); translated != "" {
			return translated
		}
	}

	return "" // não encontrou tradução
}

// translateCode traduz um código de erro
func (m *I18nTranslationMiddleware) translateCode(ctx context.Context, code, lang string) string {
	// Tenta traduzir o código diretamente
	prefixes := []string{"code.", "codes.", "error_code.", "error_codes."}
	for _, prefix := range prefixes {
		key := fmt.Sprintf("%s%s", prefix, strings.ToLower(code))
		if translated := m.tryTranslate(ctx, key, lang); translated != "" {
			return translated
		}
	}

	return "" // não encontrou tradução para o código
}

// translateMetadata traduz campos específicos dos metadados do erro
func (m *I18nTranslationMiddleware) translateMetadata(ctx context.Context, metadata map[string]interface{}, lang string) {
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
		"user_message",
		"description",
		"reason",
		"suggestion",
	}

	for _, field := range translatableFields {
		if value, ok := metadata[field].(string); ok && value != "" {
			if translated := m.tryTranslate(ctx, value, lang); translated != "" {
				metadata[field] = translated
				metadata[fmt.Sprintf("%s_original", field)] = value
			}
		}
	}

	// Traduz arrays de mensagens
	arrayFields := []string{"validation_errors", "field_errors", "business_rules", "constraints"}
	for _, field := range arrayFields {
		if messages, ok := metadata[field].([]string); ok {
			translatedMessages := make([]string, len(messages))
			hasTranslations := false

			for i, msg := range messages {
				if translated := m.tryTranslate(ctx, msg, lang); translated != "" {
					translatedMessages[i] = translated
					hasTranslations = true
				} else {
					translatedMessages[i] = msg
				}
			}

			if hasTranslations {
				metadata[fmt.Sprintf("%s_original", field)] = messages
				metadata[field] = translatedMessages
			}
		}
	}
}

// translateContextMetadata traduz campos dos metadados do contexto
func (m *I18nTranslationMiddleware) translateContextMetadata(ctx context.Context, metadata map[string]interface{}, lang string) {
	if metadata == nil {
		return
	}

	// Campos do contexto que podem ser traduzidos
	contextFields := []string{
		"operation_description",
		"step_description",
		"process_name",
		"action_description",
		"status_message",
	}

	for _, field := range contextFields {
		if value, ok := metadata[field].(string); ok && value != "" {
			if translated := m.tryTranslate(ctx, value, lang); translated != "" {
				metadata[field] = translated
				metadata[fmt.Sprintf("%s_original", field)] = value
			}
		}
	}
}

// tryTranslate tenta traduzir uma chave, retornando string vazia se não encontrar
func (m *I18nTranslationMiddleware) tryTranslate(ctx context.Context, key, lang string) string {
	if ctx == nil {
		ctx = context.Background()
	}

	translated, err := m.i18nProvider.Translate(ctx, key, lang, nil)
	if err != nil || translated == key {
		// Tenta com idioma de fallback se o idioma solicitado falhar
		if lang != m.fallbackLang {
			translated, err = m.i18nProvider.Translate(ctx, key, m.fallbackLang, nil)
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
func (m *I18nTranslationMiddleware) normalizeMessageKey(message string) string {
	// Remove espaços extras e converte para lowercase
	normalized := strings.TrimSpace(strings.ToLower(message))
	// Remove pontuação e substitui espaços por underscores
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, "!", "")
	normalized = strings.ReplaceAll(normalized, "?", "")
	normalized = strings.ReplaceAll(normalized, ",", "")
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, ";", "")
	return normalized
}

// generateCommonKeys gera chaves comuns baseadas na mensagem e código
func (m *I18nTranslationMiddleware) generateCommonKeys(message, code string) []string {
	var keys []string

	// Chaves baseadas no tipo de erro (detectado da mensagem)
	messageLower := strings.ToLower(message)

	// Erros comuns
	errorPatterns := map[string][]string{
		"not found":    {"error.not_found", "errors.not_found", "messages.not_found"},
		"unauthorized": {"error.unauthorized", "errors.access_denied", "messages.unauthorized"},
		"forbidden":    {"error.forbidden", "errors.forbidden", "messages.forbidden"},
		"validation":   {"error.validation", "errors.validation_failed", "messages.validation"},
		"invalid":      {"error.invalid", "errors.invalid_input", "messages.invalid"},
		"expired":      {"error.expired", "errors.expired", "messages.expired"},
		"timeout":      {"error.timeout", "errors.timeout", "messages.timeout"},
		"unavailable":  {"error.unavailable", "errors.service_unavailable", "messages.unavailable"},
		"internal":     {"error.internal", "errors.internal_server_error", "messages.internal"},
		"bad request":  {"error.bad_request", "errors.bad_request", "messages.bad_request"},
		"conflict":     {"error.conflict", "errors.conflict", "messages.conflict"},
	}

	for pattern, patternKeys := range errorPatterns {
		if strings.Contains(messageLower, pattern) {
			keys = append(keys, patternKeys...)
		}
	}

	// Adiciona variações do código se fornecido
	if code != "" {
		codeVariations := []string{
			strings.ToLower(code),
			fmt.Sprintf("errors.%s", strings.ToLower(code)),
			fmt.Sprintf("messages.%s", strings.ToLower(code)),
			fmt.Sprintf("codes.%s", strings.ToLower(code)),
		}
		keys = append(keys, codeVariations...)
	}

	return keys
}

// Type retorna o tipo do middleware
func (m *I18nTranslationMiddleware) Type() domainInterfaces.MiddlewareType {
	return m.middlewareType
}

// Priority retorna a prioridade de execução (menor número = maior prioridade)
func (m *I18nTranslationMiddleware) Priority() int {
	return m.priority
}

// Enabled indica se o middleware está habilitado
func (m *I18nTranslationMiddleware) Enabled() bool {
	return m.enabled
}

// SetEnabled habilita/desabilita o middleware
func (m *I18nTranslationMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// SetPriority define a prioridade do middleware
func (m *I18nTranslationMiddleware) SetPriority(priority int) {
	m.priority = priority
}

// GetSupportedLanguages retorna os idiomas suportados
func (m *I18nTranslationMiddleware) GetSupportedLanguages() []string {
	return m.i18nProvider.GetSupportedLanguages()
}

// SetName define um nome customizado para o middleware
func (m *I18nTranslationMiddleware) SetName(name string) {
	m.name = name
}

// GetTranslationStats retorna estatísticas básicas de tradução
func (m *I18nTranslationMiddleware) GetTranslationStats() map[string]interface{} {
	return map[string]interface{}{
		"name":                m.name,
		"supported_languages": m.GetSupportedLanguages(),
		"default_language":    m.defaultLang,
		"fallback_language":   m.fallbackLang,
		"translates_codes":    m.translateCodes,
		"translates_metadata": m.translateMeta,
		"enabled":             m.enabled,
		"priority":            m.priority,
		"type":                string(m.middlewareType),
	}
}
