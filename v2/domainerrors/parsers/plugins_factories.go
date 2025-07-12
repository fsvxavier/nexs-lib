// Package parsers contém exemplos de plugins e factories customizadas
package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// GenericDatabasePlugin implementa um plugin genérico para bancos de dados.
type GenericDatabasePlugin struct {
	name        string
	version     string
	description string
}

// NewGenericDatabasePlugin cria um novo plugin genérico para DB.
func NewGenericDatabasePlugin() ParserPlugin {
	return &GenericDatabasePlugin{
		name:        "generic_database",
		version:     "1.0.0",
		description: "Generic database error parser with configurable patterns",
	}
}

// Name retorna o nome do plugin.
func (p *GenericDatabasePlugin) Name() string {
	return p.name
}

// Version retorna a versão do plugin.
func (p *GenericDatabasePlugin) Version() string {
	return p.version
}

// Description retorna a descrição do plugin.
func (p *GenericDatabasePlugin) Description() string {
	return p.description
}

// CreateParser cria uma nova instância do parser.
func (p *GenericDatabasePlugin) CreateParser(config map[string]interface{}) (interfaces.ErrorParser, error) {
	patterns, ok := config["patterns"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("patterns configuration is required and must be an array")
	}

	codes, ok := config["error_codes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error_codes configuration is required and must be a map")
	}

	return &GenericDatabaseParser{
		patterns:   patterns,
		errorCodes: codes,
	}, nil
}

// ValidateConfig valida a configuração do parser.
func (p *GenericDatabasePlugin) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	patterns, exists := config["patterns"]
	if !exists {
		return fmt.Errorf("patterns configuration is required")
	}

	if _, ok := patterns.([]interface{}); !ok {
		return fmt.Errorf("patterns must be an array")
	}

	codes, exists := config["error_codes"]
	if !exists {
		return fmt.Errorf("error_codes configuration is required")
	}

	if _, ok := codes.(map[string]interface{}); !ok {
		return fmt.Errorf("error_codes must be a map")
	}

	return nil
}

// DefaultConfig retorna a configuração padrão.
func (p *GenericDatabasePlugin) DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"patterns": []interface{}{
			"database",
			"sql",
			"connection",
			"query",
		},
		"error_codes": map[string]interface{}{
			"timeout":    "DB_TIMEOUT",
			"connection": "DB_CONNECTION_ERROR",
			"constraint": "DB_CONSTRAINT_VIOLATION",
			"syntax":     "DB_SYNTAX_ERROR",
		},
	}
}

// GenericDatabaseParser implementa um parser genérico configurável.
type GenericDatabaseParser struct {
	patterns   []interface{}
	errorCodes map[string]interface{}
}

// CanParse verifica se pode processar o erro.
func (p *GenericDatabaseParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	for _, pattern := range p.patterns {
		if patternStr, ok := pattern.(string); ok {
			if strings.Contains(errStr, strings.ToLower(patternStr)) {
				return true
			}
		}
	}

	return false
}

// Parse processa o erro genérico.
func (p *GenericDatabaseParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeDatabase),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	// Tenta mapear para códigos conhecidos
	for keyword, code := range p.errorCodes {
		if strings.Contains(errLower, strings.ToLower(keyword)) {
			if codeStr, ok := code.(string); ok {
				parsed.Code = codeStr
				parsed.Details["matched_keyword"] = keyword
				break
			}
		}
	}

	if parsed.Code == "" {
		parsed.Code = "GENERIC_DB_ERROR"
	}

	// Define propriedades baseadas no código
	parsed.Severity = p.mapSeverity(parsed.Code)
	parsed.Retryable = p.isRetryable(parsed.Code)
	parsed.Temporary = p.isTemporary(parsed.Code)

	return parsed
}

// mapSeverity mapeia códigos para severidade.
func (p *GenericDatabaseParser) mapSeverity(code string) interfaces.Severity {
	switch {
	case strings.Contains(code, "TIMEOUT"):
		return interfaces.SeverityHigh
	case strings.Contains(code, "CONNECTION"):
		return interfaces.SeverityHigh
	case strings.Contains(code, "CONSTRAINT"):
		return interfaces.SeverityMedium
	case strings.Contains(code, "SYNTAX"):
		return interfaces.SeverityMedium
	default:
		return interfaces.SeverityMedium
	}
}

// isRetryable verifica se é retryable.
func (p *GenericDatabaseParser) isRetryable(code string) bool {
	return strings.Contains(code, "TIMEOUT") || strings.Contains(code, "CONNECTION")
}

// isTemporary verifica se é temporário.
func (p *GenericDatabaseParser) isTemporary(code string) bool {
	return p.isRetryable(code)
}

// CustomParserFactory implementa uma factory para parsers customizados.
type CustomParserFactory struct {
	customTypes map[string]func(config map[string]interface{}) (interfaces.ErrorParser, error)
}

// NewCustomParserFactory cria uma nova factory customizada.
func NewCustomParserFactory() ParserFactory {
	return &CustomParserFactory{
		customTypes: make(map[string]func(config map[string]interface{}) (interfaces.ErrorParser, error)),
	}
}

// CreateParser cria um parser baseado no tipo.
func (f *CustomParserFactory) CreateParser(parserType string, config map[string]interface{}) (interfaces.ErrorParser, error) {
	if factory, exists := f.customTypes[parserType]; exists {
		return factory(config)
	}

	// Parsers built-in
	switch parserType {
	case "regex_matcher":
		return f.createRegexMatcher(config)
	case "keyword_matcher":
		return f.createKeywordMatcher(config)
	case "json_error_parser":
		return f.createJSONErrorParser(config)
	default:
		return nil, fmt.Errorf("unknown parser type: %s", parserType)
	}
}

// SupportedTypes retorna os tipos suportados.
func (f *CustomParserFactory) SupportedTypes() []string {
	types := []string{"regex_matcher", "keyword_matcher", "json_error_parser"}

	for customType := range f.customTypes {
		types = append(types, customType)
	}

	return types
}

// RegisterCustomType registra um tipo customizado.
func (f *CustomParserFactory) RegisterCustomType(typeName string, factory func(config map[string]interface{}) (interfaces.ErrorParser, error)) error {
	if typeName == "" {
		return fmt.Errorf("type name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	f.customTypes[typeName] = factory
	return nil
}

// createRegexMatcher cria um parser baseado em regex.
func (f *CustomParserFactory) createRegexMatcher(config map[string]interface{}) (interfaces.ErrorParser, error) {
	pattern, ok := config["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern is required for regex_matcher")
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	errorCode, _ := config["error_code"].(string)
	if errorCode == "" {
		errorCode = "REGEX_MATCH"
	}

	return &RegexMatcherParser{
		regex:     regex,
		errorCode: errorCode,
	}, nil
}

// createKeywordMatcher cria um parser baseado em palavras-chave.
func (f *CustomParserFactory) createKeywordMatcher(config map[string]interface{}) (interfaces.ErrorParser, error) {
	keywords, ok := config["keywords"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("keywords are required for keyword_matcher")
	}

	var keywordStrs []string
	for _, kw := range keywords {
		if kwStr, ok := kw.(string); ok {
			keywordStrs = append(keywordStrs, kwStr)
		}
	}

	if len(keywordStrs) == 0 {
		return nil, fmt.Errorf("at least one keyword is required")
	}

	errorCode, _ := config["error_code"].(string)
	if errorCode == "" {
		errorCode = "KEYWORD_MATCH"
	}

	return &KeywordMatcherParser{
		keywords:  keywordStrs,
		errorCode: errorCode,
	}, nil
}

// createJSONErrorParser cria um parser para erros em formato JSON.
func (f *CustomParserFactory) createJSONErrorParser(config map[string]interface{}) (interfaces.ErrorParser, error) {
	return &JSONErrorParser{}, nil
}

// RegexMatcherParser parser baseado em regex.
type RegexMatcherParser struct {
	regex     *regexp.Regexp
	errorCode string
}

// CanParse verifica se o regex faz match.
func (p *RegexMatcherParser) CanParse(err error) bool {
	return err != nil && p.regex.MatchString(err.Error())
}

// Parse processa com regex.
func (p *RegexMatcherParser) Parse(err error) interfaces.ParsedError {
	matches := p.regex.FindStringSubmatch(err.Error())

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeInternal),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  err.Error(),
		Code:     p.errorCode,
		Severity: interfaces.SeverityMedium,
	}

	if len(matches) > 1 {
		parsed.Details["regex_matches"] = matches[1:]
	}

	return parsed
}

// KeywordMatcherParser parser baseado em palavras-chave.
type KeywordMatcherParser struct {
	keywords  []string
	errorCode string
}

// CanParse verifica se alguma palavra-chave faz match.
func (p *KeywordMatcherParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	for _, keyword := range p.keywords {
		if strings.Contains(errStr, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// Parse processa com keywords.
func (p *KeywordMatcherParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()
	errLower := strings.ToLower(errStr)

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeInternal),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
		Code:     p.errorCode,
		Severity: interfaces.SeverityMedium,
	}

	var matchedKeywords []string
	for _, keyword := range p.keywords {
		if strings.Contains(errLower, strings.ToLower(keyword)) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	parsed.Details["matched_keywords"] = matchedKeywords

	return parsed
}

// JSONErrorParser parser para erros em formato JSON.
type JSONErrorParser struct{}

// CanParse verifica se o erro parece ser JSON.
func (p *JSONErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.TrimSpace(err.Error())
	return strings.HasPrefix(errStr, "{") && strings.HasSuffix(errStr, "}")
}

// Parse processa erro JSON.
func (p *JSONErrorParser) Parse(err error) interfaces.ParsedError {
	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeInternal),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  err.Error(),
		Code:     "JSON_ERROR",
		Severity: interfaces.SeverityMedium,
	}

	parsed.Details["raw_json"] = err.Error()

	return parsed
}
