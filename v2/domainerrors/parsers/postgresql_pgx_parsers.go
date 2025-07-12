// Package parsers implementa parsers para sistemas de banco de dados PostgreSQL
package parsers

import (
	"regexp"
	"strings"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// PGXErrorParser parser especializado para erros do PGX (driver PostgreSQL moderno).
type PGXErrorParser struct {
	sqlStateRegex   *regexp.Regexp
	constraintRegex *regexp.Regexp
	connectionRegex *regexp.Regexp
}

// NewPGXErrorParser cria um novo parser para PGX.
func NewPGXErrorParser() interfaces.ErrorParser {
	return &PGXErrorParser{
		sqlStateRegex:   regexp.MustCompile(`SQLSTATE (\w{5})`),
		constraintRegex: regexp.MustCompile(`constraint "([^"]+)"`),
		connectionRegex: regexp.MustCompile(`(connection|dial|timeout|refused)`),
	}
}

// CanParse verifica se pode processar o erro PGX.
func (p *PGXErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "pgx") ||
		strings.Contains(errStr, "pgxpool") ||
		strings.Contains(errStr, "conn closed") ||
		strings.Contains(errStr, "canceling statement") ||
		strings.Contains(errStr, "too many clients") ||
		p.sqlStateRegex.MatchString(err.Error())
}

// Parse processa erro PGX.
func (p *PGXErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeDatabase),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	errLower := strings.ToLower(errStr)

	// Extrai SQLSTATE se disponível
	if matches := p.sqlStateRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		sqlState := matches[1]
		parsed.Details["sql_state"] = sqlState
		parsed.Code = "PGX_" + sqlState

		// Mapeia estados SQL conhecidos
		parsed.Severity = p.mapSQLStateSeverity(sqlState)
		parsed.Retryable = p.isSQLStateRetryable(sqlState)
		parsed.Temporary = p.isSQLStateTemporary(sqlState)
	} else {
		// Classifica por padrões na mensagem
		switch {
		case strings.Contains(errLower, "too many clients"):
			parsed.Code = "PGX_TOO_MANY_CLIENTS"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "connection_limit"

		case strings.Contains(errLower, "conn closed") || strings.Contains(errLower, "connection"):
			parsed.Code = "PGX_CONNECTION_ERROR"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "connection"

		case strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline"):
			parsed.Code = "PGX_TIMEOUT"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "timeout"

		case strings.Contains(errLower, "canceling statement"):
			parsed.Code = "PGX_QUERY_CANCELED"
			parsed.Severity = interfaces.SeverityMedium
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "cancellation"

		case strings.Contains(errLower, "pool"):
			parsed.Code = "PGX_POOL_ERROR"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "pool"

		default:
			parsed.Code = "PGX_UNKNOWN"
			parsed.Severity = interfaces.SeverityMedium
		}
	}

	// Extrai constraint se disponível
	if matches := p.constraintRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["constraint_name"] = matches[1]
	}

	return parsed
}

// mapSQLStateSeverity mapeia SQLSTATE para severidade.
func (p *PGXErrorParser) mapSQLStateSeverity(sqlState string) interfaces.Severity {
	switch sqlState[:2] {
	case "08": // Connection Exception
		return interfaces.SeverityHigh
	case "53": // Insufficient Resources
		return interfaces.SeverityHigh
	case "57": // Operator Intervention
		return interfaces.SeverityCritical
	case "XX": // Internal Error
		return interfaces.SeverityCritical
	case "23": // Integrity Constraint Violation
		return interfaces.SeverityMedium
	case "42": // Syntax Error or Access Rule Violation
		return interfaces.SeverityMedium
	default:
		return interfaces.SeverityMedium
	}
}

// isSQLStateRetryable verifica se o SQLSTATE é retryable.
func (p *PGXErrorParser) isSQLStateRetryable(sqlState string) bool {
	switch sqlState[:2] {
	case "08": // Connection Exception
		return true
	case "53": // Insufficient Resources
		return true
	case "40": // Transaction Rollback
		return true
	default:
		return false
	}
}

// isSQLStateTemporary verifica se o SQLSTATE é temporário.
func (p *PGXErrorParser) isSQLStateTemporary(sqlState string) bool {
	switch sqlState {
	case "08003", "08006": // Connection does not exist, Connection failure
		return true
	case "53000", "53100", "53200", "53300": // Insufficient resources
		return true
	default:
		return false
	}
}

// EnhancedPostgreSQLErrorParser parser aprimorado para PostgreSQL com mais detalhes.
type EnhancedPostgreSQLErrorParser struct {
	*PostgreSQLErrorParser
	severityRegex *regexp.Regexp
	positionRegex *regexp.Regexp
	schemaRegex   *regexp.Regexp
	tableRegex    *regexp.Regexp
}

// NewEnhancedPostgreSQLErrorParser cria um parser PostgreSQL aprimorado.
func NewEnhancedPostgreSQLErrorParser() interfaces.ErrorParser {
	return &EnhancedPostgreSQLErrorParser{
		PostgreSQLErrorParser: NewPostgreSQLErrorParser().(*PostgreSQLErrorParser),
		severityRegex:         regexp.MustCompile(`SEVERITY:\s*(\w+)`),
		positionRegex:         regexp.MustCompile(`POSITION:\s*(\d+)`),
		schemaRegex:           regexp.MustCompile(`SCHEMA:\s*(\w+)`),
		tableRegex:            regexp.MustCompile(`TABLE:\s*(\w+)`),
	}
}

// CanParse verifica se pode processar o erro PostgreSQL avançado.
func (p *EnhancedPostgreSQLErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	// Primeiro verifica com o parser base
	if !p.PostgreSQLErrorParser.CanParse(err) {
		return false
	}

	errStr := err.Error()
	// Verifica se tem informações extras que justifiquem usar o parser aprimorado
	return strings.Contains(errStr, "DETAIL:") ||
		strings.Contains(errStr, "HINT:") ||
		strings.Contains(errStr, "CONTEXT:") ||
		p.severityRegex.MatchString(errStr) ||
		p.positionRegex.MatchString(errStr)
}

// Parse processa erro PostgreSQL com detalhes extras.
func (p *EnhancedPostgreSQLErrorParser) Parse(err error) interfaces.ParsedError {
	// Começa com o parse básico
	parsed := p.PostgreSQLErrorParser.Parse(err)
	parsed.Code = "POSTGRES_" + strings.TrimPrefix(parsed.Code, "POSTGRES_")

	errStr := err.Error()

	// Adiciona informações extras específicas do PostgreSQL
	if matches := p.severityRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["postgres_severity"] = matches[1]
	}

	if matches := p.positionRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["error_position"] = matches[1]
	}

	if matches := p.schemaRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["schema_name"] = matches[1]
	}

	if matches := p.tableRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["table_name"] = matches[1]
	}

	// Extrai DETAIL se disponível
	if detailIdx := strings.Index(errStr, "DETAIL:"); detailIdx != -1 {
		detailEnd := len(errStr)
		if hintIdx := strings.Index(errStr[detailIdx:], "HINT:"); hintIdx != -1 {
			detailEnd = detailIdx + hintIdx
		}
		if contextIdx := strings.Index(errStr[detailIdx:], "CONTEXT:"); contextIdx != -1 {
			if contextIdx < detailEnd-detailIdx {
				detailEnd = detailIdx + contextIdx
			}
		}
		detail := strings.TrimSpace(errStr[detailIdx+7 : detailEnd])
		if detail != "" {
			parsed.Details["detail"] = detail
		}
	}

	// Extrai HINT se disponível
	if hintIdx := strings.Index(errStr, "HINT:"); hintIdx != -1 {
		hintEnd := len(errStr)
		if contextIdx := strings.Index(errStr[hintIdx:], "CONTEXT:"); contextIdx != -1 {
			hintEnd = hintIdx + contextIdx
		}
		hint := strings.TrimSpace(errStr[hintIdx+5 : hintEnd])
		if hint != "" {
			parsed.Details["hint"] = hint
		}
	}

	// Extrai CONTEXT se disponível
	if contextIdx := strings.Index(errStr, "CONTEXT:"); contextIdx != -1 {
		context := strings.TrimSpace(errStr[contextIdx+8:])
		if context != "" {
			parsed.Details["context"] = context
		}
	}

	return parsed
}
