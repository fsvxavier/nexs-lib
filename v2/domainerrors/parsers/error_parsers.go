// Package parsers implementa parsers especializados para diferentes tipos de erro
// seguindo o padrão Strategy para diferentes estratégias de parsing.
package parsers

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"syscall"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// PostgreSQLErrorParser parser para erros do PostgreSQL.
type PostgreSQLErrorParser struct {
	sqlStateRegex *regexp.Regexp
	codeRegex     *regexp.Regexp
}

// NewPostgreSQLErrorParser cria um novo parser para PostgreSQL.
func NewPostgreSQLErrorParser() interfaces.ErrorParser {
	return &PostgreSQLErrorParser{
		sqlStateRegex: regexp.MustCompile(`SQLSTATE (\w{5})`),
		codeRegex:     regexp.MustCompile(`ERROR:\s*(.+?)\s*\(SQLSTATE\s+(\w{5})\)`),
	}
}

// CanParse verifica se pode processar o erro do PostgreSQL.
func (p *PostgreSQLErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "SQLSTATE") ||
		strings.Contains(errStr, "postgres") ||
		strings.Contains(errStr, "pq:")
}

// Parse processa erro do PostgreSQL.
func (p *PostgreSQLErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:      string(types.ErrorTypeDatabase),
		Severity:  interfaces.Severity(types.SeverityHigh),
		Category:  interfaces.Category(types.ErrorTypeDatabase.Category()),
		Details:   make(map[string]interface{}),
		Retryable: false,
		Temporary: false,
	}

	// Extrai SQLSTATE
	if matches := p.sqlStateRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		sqlState := matches[1]
		parsed.Code = "DB_" + sqlState
		parsed.Details["sqlstate"] = sqlState

		// Mapeia códigos SQLSTATE conhecidos
		parsed.Message = p.mapSQLStateToMessage(sqlState)
		parsed.Retryable = p.isSQLStateRetryable(sqlState)
		parsed.Temporary = parsed.Retryable
	} else {
		parsed.Code = "DB_UNKNOWN"
		parsed.Message = "Database error"
	}

	// Adiciona mensagem original como detalhe
	parsed.Details["original_message"] = errStr
	parsed.Details["database_type"] = "postgresql"

	return parsed
}

// mapSQLStateToMessage mapeia códigos SQLSTATE para mensagens legíveis.
func (p *PostgreSQLErrorParser) mapSQLStateToMessage(sqlState string) string {
	switch sqlState {
	case "23505":
		return "Unique constraint violation"
	case "23503":
		return "Foreign key constraint violation"
	case "23502":
		return "Not null constraint violation"
	case "23514":
		return "Check constraint violation"
	case "42P01":
		return "Table does not exist"
	case "42703":
		return "Column does not exist"
	case "08006":
		return "Connection failure"
	case "08001":
		return "Unable to connect to database"
	case "57P01":
		return "Admin shutdown"
	case "53300":
		return "Too many connections"
	default:
		return fmt.Sprintf("Database error (SQLSTATE %s)", sqlState)
	}
}

// isSQLStateRetryable verifica se o erro é recuperável.
func (p *PostgreSQLErrorParser) isSQLStateRetryable(sqlState string) bool {
	switch sqlState {
	case "08006", "08001", "57P01", "53300": // Connection issues
		return true
	case "40001", "40P01": // Serialization failures
		return true
	default:
		return false
	}
}

// MySQLErrorParser parser para erros do MySQL.
type MySQLErrorParser struct {
	errorRegex *regexp.Regexp
}

// NewMySQLErrorParser cria um novo parser para MySQL.
func NewMySQLErrorParser() interfaces.ErrorParser {
	return &MySQLErrorParser{
		errorRegex: regexp.MustCompile(`Error (\d+): (.+)`),
	}
}

// CanParse verifica se pode processar o erro do MySQL.
func (p *MySQLErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "mysql") ||
		strings.Contains(errStr, "Error 1") ||
		p.errorRegex.MatchString(errStr)
}

// Parse processa erro do MySQL.
func (p *MySQLErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:      string(types.ErrorTypeDatabase),
		Severity:  interfaces.Severity(types.SeverityHigh),
		Category:  interfaces.Category(types.ErrorTypeDatabase.Category()),
		Details:   make(map[string]interface{}),
		Retryable: false,
		Temporary: false,
	}

	// Extrai código de erro MySQL
	if matches := p.errorRegex.FindStringSubmatch(errStr); len(matches) > 2 {
		errorCode := matches[1]
		message := matches[2]

		parsed.Code = "MYSQL_" + errorCode
		parsed.Message = message
		parsed.Details["mysql_code"] = errorCode
		parsed.Retryable = p.isMySQLErrorRetryable(errorCode)
	} else {
		parsed.Code = "MYSQL_UNKNOWN"
		parsed.Message = "MySQL error"
	}

	parsed.Details["original_message"] = errStr
	parsed.Details["database_type"] = "mysql"
	parsed.Temporary = parsed.Retryable

	return parsed
}

// isMySQLErrorRetryable verifica se o erro MySQL é recuperável.
func (p *MySQLErrorParser) isMySQLErrorRetryable(errorCode string) bool {
	switch errorCode {
	case "1040", "1041": // Too many connections
		return true
	case "1205": // Lock wait timeout
		return true
	case "2003", "2006": // Connection errors
		return true
	default:
		return false
	}
}

// NetworkErrorParser parser para erros de rede.
type NetworkErrorParser struct{}

// NewNetworkErrorParser cria um novo parser para erros de rede.
func NewNetworkErrorParser() interfaces.ErrorParser {
	return &NetworkErrorParser{}
}

// CanParse verifica se pode processar erros de rede.
func (p *NetworkErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	// Verifica tipos de erro específicos de rede
	switch err.(type) {
	case *net.OpError, *net.DNSError, *net.AddrError:
		return true
	case *url.Error:
		return true
	}
	// Verifica syscall errors
	if netErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := netErr.Err.(syscall.Errno); ok {
			return sysErr == syscall.ECONNREFUSED ||
				sysErr == syscall.ETIMEDOUT ||
				sysErr == syscall.ECONNRESET
		}
	}

	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "dns")
}

// Parse processa erros de rede.
func (p *NetworkErrorParser) Parse(err error) interfaces.ParsedError {
	parsed := interfaces.ParsedError{
		Type:      string(types.ErrorTypeNetwork),
		Severity:  interfaces.Severity(types.SeverityHigh),
		Category:  interfaces.Category(types.ErrorTypeNetwork.Category()),
		Details:   make(map[string]interface{}),
		Retryable: true,
		Temporary: true,
	}

	switch netErr := err.(type) {
	case *net.OpError:
		parsed.Code = "NET_OP_ERROR"
		parsed.Message = fmt.Sprintf("Network operation failed: %s", netErr.Op)
		parsed.Details["operation"] = netErr.Op
		parsed.Details["network"] = netErr.Net
		if netErr.Addr != nil {
			parsed.Details["address"] = netErr.Addr.String()
		}

	case *net.DNSError:
		parsed.Code = "NET_DNS_ERROR"
		parsed.Message = fmt.Sprintf("DNS resolution failed: %s", netErr.Name)
		parsed.Details["name"] = netErr.Name
		parsed.Details["server"] = netErr.Server
		parsed.Retryable = !netErr.IsNotFound

	case *net.AddrError:
		parsed.Code = "NET_ADDR_ERROR"
		parsed.Message = fmt.Sprintf("Address error: %s", netErr.Err)
		parsed.Details["address"] = netErr.Addr
		parsed.Retryable = false

	case *url.Error:
		parsed.Code = "NET_URL_ERROR"
		parsed.Message = fmt.Sprintf("URL error: %s", netErr.Op)
		parsed.Details["url"] = netErr.URL
		parsed.Details["operation"] = netErr.Op

	default:
		parsed.Code = "NET_UNKNOWN"
		parsed.Message = "Network error"
	}

	parsed.Details["original_message"] = err.Error()
	return parsed
}

// TimeoutErrorParser parser para erros de timeout.
type TimeoutErrorParser struct{}

// NewTimeoutErrorParser cria um novo parser para timeouts.
func NewTimeoutErrorParser() interfaces.ErrorParser {
	return &TimeoutErrorParser{}
}

// CanParse verifica se pode processar erros de timeout.
func (p *TimeoutErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	// Verifica interface Timeout
	if timeoutErr, ok := err.(interface{ Timeout() bool }); ok {
		return timeoutErr.Timeout()
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "context deadline")
}

// Parse processa erros de timeout.
func (p *TimeoutErrorParser) Parse(err error) interfaces.ParsedError {
	parsed := interfaces.ParsedError{
		Code:      "TIMEOUT_ERROR",
		Message:   "Operation timeout",
		Type:      string(types.ErrorTypeTimeout),
		Severity:  interfaces.Severity(types.SeverityHigh),
		Category:  interfaces.Category(types.ErrorTypeTimeout.Category()),
		Details:   make(map[string]interface{}),
		Retryable: true,
		Temporary: true,
	}

	errStr := err.Error()
	parsed.Details["original_message"] = errStr

	// Detecta tipo específico de timeout
	if strings.Contains(errStr, "context deadline") {
		parsed.Details["timeout_type"] = "context_deadline"
		parsed.Message = "Context deadline exceeded"
	} else if strings.Contains(errStr, "connection timeout") {
		parsed.Details["timeout_type"] = "connection"
		parsed.Message = "Connection timeout"
	} else if strings.Contains(errStr, "read timeout") {
		parsed.Details["timeout_type"] = "read"
		parsed.Message = "Read timeout"
	} else if strings.Contains(errStr, "write timeout") {
		parsed.Details["timeout_type"] = "write"
		parsed.Message = "Write timeout"
	}

	return parsed
}

// SQLErrorParser parser genérico para erros SQL.
type SQLErrorParser struct{}

// NewSQLErrorParser cria um novo parser genérico para SQL.
func NewSQLErrorParser() interfaces.ErrorParser {
	return &SQLErrorParser{}
}

// CanParse verifica se pode processar erros SQL genéricos.
func (p *SQLErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	// Verifica tipos SQL padrão
	switch err {
	case sql.ErrNoRows, sql.ErrTxDone, sql.ErrConnDone:
		return true
	}

	// Verifica interface driver.Valuer
	if _, ok := err.(driver.Valuer); ok {
		return true
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "sql") ||
		strings.Contains(errStr, "database") ||
		strings.Contains(errStr, "query") ||
		strings.Contains(errStr, "transaction")
}

// Parse processa erros SQL genéricos.
func (p *SQLErrorParser) Parse(err error) interfaces.ParsedError {
	parsed := interfaces.ParsedError{
		Type:      string(types.ErrorTypeDatabase),
		Severity:  interfaces.Severity(types.SeverityMedium),
		Category:  interfaces.Category(types.ErrorTypeDatabase.Category()),
		Details:   make(map[string]interface{}),
		Retryable: false,
		Temporary: false,
	}

	switch err {
	case sql.ErrNoRows:
		parsed.Code = "SQL_NO_ROWS"
		parsed.Message = "No rows found"
		parsed.Severity = interfaces.Severity(types.SeverityLow)

	case sql.ErrTxDone:
		parsed.Code = "SQL_TX_DONE"
		parsed.Message = "Transaction already committed or rolled back"

	case sql.ErrConnDone:
		parsed.Code = "SQL_CONN_DONE"
		parsed.Message = "Database connection is closed"
		parsed.Retryable = true
		parsed.Temporary = true

	default:
		parsed.Code = "SQL_UNKNOWN"
		parsed.Message = "SQL error"
	}

	parsed.Details["original_message"] = err.Error()
	return parsed
}

// CompositeErrorParser parser composto que usa múltiplos parsers.
type CompositeErrorParser struct {
	parsers []interfaces.ErrorParser
}

// NewCompositeErrorParser cria um parser composto.
func NewCompositeErrorParser(parsers ...interfaces.ErrorParser) interfaces.ErrorParser {
	return &CompositeErrorParser{
		parsers: parsers,
	}
}

// CanParse verifica se algum parser filho pode processar o erro.
func (p *CompositeErrorParser) CanParse(err error) bool {
	for _, parser := range p.parsers {
		if parser.CanParse(err) {
			return true
		}
	}
	return false
}

// Parse usa o primeiro parser capaz de processar o erro.
func (p *CompositeErrorParser) Parse(err error) interfaces.ParsedError {
	for _, parser := range p.parsers {
		if parser.CanParse(err) {
			return parser.Parse(err)
		}
	}

	// Fallback para erro genérico
	return interfaces.ParsedError{
		Code:      "UNKNOWN_ERROR",
		Message:   err.Error(),
		Type:      string(types.ErrorTypeInternal),
		Severity:  interfaces.Severity(types.SeverityMedium),
		Category:  interfaces.Category(types.ErrorTypeInternal.Category()),
		Details:   map[string]interface{}{"original_message": err.Error()},
		Retryable: false,
		Temporary: false,
	}
}

// NewDefaultParser cria um parser padrão com todos os parsers comuns.
func NewDefaultParser() interfaces.ErrorParser {
	return NewCompositeErrorParser(
		NewTimeoutErrorParser(),
		NewNetworkErrorParser(),
		NewPostgreSQLErrorParser(),
		NewMySQLErrorParser(),
		NewSQLErrorParser(),
	)
}

// ParseError é uma função utilitária para parsing de erros.
func ParseError(err error, parser interfaces.ErrorParser) interfaces.ParsedError {
	if parser == nil {
		parser = NewDefaultParser()
	}

	if !parser.CanParse(err) {
		return interfaces.ParsedError{
			Code:      "UNPARSEABLE_ERROR",
			Message:   err.Error(),
			Type:      string(types.ErrorTypeInternal),
			Severity:  interfaces.Severity(types.SeverityMedium),
			Category:  interfaces.Category(types.ErrorTypeInternal.Category()),
			Details:   map[string]interface{}{"original_message": err.Error()},
			Retryable: false,
			Temporary: false,
		}
	}

	return parser.Parse(err)
}
