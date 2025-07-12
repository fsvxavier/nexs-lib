// Package parsers implementa parsers para sistemas de banco de dados NoSQL
package parsers

import (
	"regexp"
	"strings"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// RedisErrorParser parser para erros Redis.
type RedisErrorParser struct {
	errorCodeRegex *regexp.Regexp
	timeoutRegex   *regexp.Regexp
}

// NewRedisErrorParser cria um novo parser para Redis.
func NewRedisErrorParser() interfaces.ErrorParser {
	return &RedisErrorParser{
		errorCodeRegex: regexp.MustCompile(`(ERR|WRONGTYPE|NOSCRIPT|BUSY|READONLY|NOAUTH|LOADING|MASTERDOWN|MISCONF)`),
		timeoutRegex:   regexp.MustCompile(`(timeout|deadline exceeded|connection refused)`),
	}
}

// CanParse verifica se pode processar o erro Redis.
func (p *RedisErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "redis") ||
		strings.Contains(errStr, "dial tcp") ||
		p.errorCodeRegex.MatchString(err.Error()) ||
		strings.Contains(errStr, "connection pool") ||
		strings.Contains(errStr, "redigo")
}

// Parse processa erro Redis.
func (p *RedisErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeDatabase),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	// Identifica tipos específicos de erro Redis
	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline"):
		parsed.Code = "REDIS_TIMEOUT"
		parsed.Severity = interfaces.SeverityHigh
		parsed.Retryable = true
		parsed.Temporary = true
		parsed.Details["error_type"] = "timeout"

	case strings.Contains(errLower, "connection refused") || strings.Contains(errLower, "connection reset"):
		parsed.Code = "REDIS_CONNECTION_ERROR"
		parsed.Severity = interfaces.SeverityHigh
		parsed.Retryable = true
		parsed.Temporary = true
		parsed.Details["error_type"] = "connection"

	case strings.Contains(errStr, "WRONGTYPE"):
		parsed.Code = "REDIS_WRONG_TYPE"
		parsed.Severity = interfaces.SeverityMedium
		parsed.Retryable = false
		parsed.Temporary = false
		parsed.Details["error_type"] = "data_type"

	case strings.Contains(errStr, "NOAUTH"):
		parsed.Code = "REDIS_AUTH_ERROR"
		parsed.Severity = interfaces.SeverityHigh
		parsed.Retryable = false
		parsed.Temporary = false
		parsed.Details["error_type"] = "authentication"

	case strings.Contains(errStr, "READONLY"):
		parsed.Code = "REDIS_READONLY"
		parsed.Severity = interfaces.SeverityMedium
		parsed.Retryable = true
		parsed.Temporary = true
		parsed.Details["error_type"] = "readonly"

	case strings.Contains(errStr, "BUSY"):
		parsed.Code = "REDIS_BUSY"
		parsed.Severity = interfaces.SeverityMedium
		parsed.Retryable = true
		parsed.Temporary = true
		parsed.Details["error_type"] = "busy"

	default:
		parsed.Code = "REDIS_UNKNOWN"
		parsed.Severity = interfaces.SeverityMedium
		parsed.Details["error_type"] = "unknown"
	}

	// Adiciona detalhes específicos se encontrados
	if matches := p.errorCodeRegex.FindStringSubmatch(errStr); len(matches) > 0 {
		parsed.Details["redis_error_code"] = matches[0]
	}

	return parsed
}

// MongoDBErrorParser parser para erros MongoDB.
type MongoDBErrorParser struct {
	errorCodeRegex *regexp.Regexp
	oplogRegex     *regexp.Regexp
}

// NewMongoDBErrorParser cria um novo parser para MongoDB.
func NewMongoDBErrorParser() interfaces.ErrorParser {
	return &MongoDBErrorParser{
		errorCodeRegex: regexp.MustCompile(`\((\d+)\)`),
		oplogRegex:     regexp.MustCompile(`oplog|replica set`),
	}
}

// CanParse verifica se pode processar o erro MongoDB.
func (p *MongoDBErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "mongo") ||
		strings.Contains(errStr, "bson") ||
		strings.Contains(errStr, "replica set") ||
		strings.Contains(errStr, "oplog") ||
		strings.Contains(errStr, "collection") ||
		p.errorCodeRegex.MatchString(err.Error())
}

// Parse processa erro MongoDB.
func (p *MongoDBErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeDatabase),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	errLower := strings.ToLower(errStr)

	// Mapeia códigos de erro específicos do MongoDB
	if matches := p.errorCodeRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["mongo_error_code"] = matches[1]

		// Mapeia códigos conhecidos
		switch matches[1] {
		case "11000": // Duplicate key error
			parsed.Code = "MONGO_DUPLICATE_KEY"
			parsed.Severity = interfaces.SeverityMedium
			parsed.Retryable = false
			parsed.Details["error_type"] = "duplicate_key"

		case "50": // Exceeded time limit
			parsed.Code = "MONGO_TIMEOUT"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true
			parsed.Details["error_type"] = "timeout"

		case "13": // Unauthorized
			parsed.Code = "MONGO_UNAUTHORIZED"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = false
			parsed.Details["error_type"] = "authorization"

		case "18": // Authentication failed
			parsed.Code = "MONGO_AUTH_FAILED"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = false
			parsed.Details["error_type"] = "authentication"

		default:
			parsed.Code = "MONGO_ERROR_" + matches[1]
			parsed.Severity = interfaces.SeverityMedium
		}
	} else {
		// Classifica por conteúdo da mensagem
		switch {
		case strings.Contains(errLower, "timeout") || strings.Contains(errLower, "deadline"):
			parsed.Code = "MONGO_TIMEOUT"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true

		case strings.Contains(errLower, "connection") && strings.Contains(errLower, "refused"):
			parsed.Code = "MONGO_CONNECTION_ERROR"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true

		case strings.Contains(errLower, "authentication") || strings.Contains(errLower, "auth"):
			parsed.Code = "MONGO_AUTH_ERROR"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = false

		case strings.Contains(errLower, "replica set"):
			parsed.Code = "MONGO_REPLICA_SET_ERROR"
			parsed.Severity = interfaces.SeverityHigh
			parsed.Retryable = true
			parsed.Temporary = true

		default:
			parsed.Code = "MONGO_UNKNOWN"
			parsed.Severity = interfaces.SeverityMedium
		}
	}

	return parsed
}

// AWSErrorParser parser para erros AWS/Cloud.
type AWSErrorParser struct {
	errorCodeRegex  *regexp.Regexp
	requestIdRegex  *regexp.Regexp
	throttlingRegex *regexp.Regexp
}

// NewAWSErrorParser cria um novo parser para AWS.
func NewAWSErrorParser() interfaces.ErrorParser {
	return &AWSErrorParser{
		errorCodeRegex:  regexp.MustCompile(`([A-Z][a-zA-Z]+Exception|[A-Z][a-zA-Z]+Error)`),
		requestIdRegex:  regexp.MustCompile(`RequestId:\s*([a-f0-9\-]+)`),
		throttlingRegex: regexp.MustCompile(`(Throttling|TooManyRequests|RequestLimitExceeded)`),
	}
}

// CanParse verifica se pode processar o erro AWS.
func (p *AWSErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "aws") ||
		strings.Contains(errStr, "s3") ||
		strings.Contains(errStr, "dynamodb") ||
		strings.Contains(errStr, "lambda") ||
		strings.Contains(errStr, "requestid") ||
		p.errorCodeRegex.MatchString(err.Error()) ||
		strings.Contains(errStr, "throttling")
}

// Parse processa erro AWS.
func (p *AWSErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeCloud),
		Category: interfaces.CategoryInfrastructure,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	// Extrai código de erro AWS
	if matches := p.errorCodeRegex.FindStringSubmatch(errStr); len(matches) > 0 {
		awsErrorCode := matches[0]
		parsed.Code = "AWS_" + strings.ToUpper(awsErrorCode)
		parsed.Details["aws_error_code"] = awsErrorCode

		// Mapeia severidade e propriedades baseadas no tipo de erro
		parsed.Severity = p.mapAWSSeverity(awsErrorCode)
		parsed.Retryable = p.isAWSRetryable(awsErrorCode)
		parsed.Temporary = p.isAWSTemporary(awsErrorCode)
	} else {
		parsed.Code = "AWS_UNKNOWN"
		parsed.Severity = interfaces.SeverityMedium
	}

	// Extrai Request ID se disponível
	if matches := p.requestIdRegex.FindStringSubmatch(errStr); len(matches) > 1 {
		parsed.Details["request_id"] = matches[1]
	}

	// Verifica throttling
	if p.throttlingRegex.MatchString(errStr) {
		parsed.Details["throttled"] = true
		parsed.Retryable = true
		parsed.Temporary = true
		parsed.Severity = interfaces.SeverityHigh
	}

	return parsed
}

// mapAWSSeverity mapeia códigos AWS para severidade.
func (p *AWSErrorParser) mapAWSSeverity(errorCode string) interfaces.Severity {
	errorLower := strings.ToLower(errorCode)

	switch {
	case strings.Contains(errorLower, "internal") || strings.Contains(errorLower, "service"):
		return interfaces.SeverityCritical
	case strings.Contains(errorLower, "throttl") || strings.Contains(errorLower, "limit"):
		return interfaces.SeverityHigh
	case strings.Contains(errorLower, "access") || strings.Contains(errorLower, "permission"):
		return interfaces.SeverityMedium
	default:
		return interfaces.SeverityMedium
	}
}

// isAWSRetryable verifica se o erro AWS é retryable.
func (p *AWSErrorParser) isAWSRetryable(errorCode string) bool {
	errorLower := strings.ToLower(errorCode)

	return strings.Contains(errorLower, "throttl") ||
		strings.Contains(errorLower, "internal") ||
		strings.Contains(errorLower, "service") ||
		strings.Contains(errorLower, "timeout")
}

// isAWSTemporary verifica se o erro AWS é temporário.
func (p *AWSErrorParser) isAWSTemporary(errorCode string) bool {
	errorLower := strings.ToLower(errorCode)

	return strings.Contains(errorLower, "throttl") ||
		strings.Contains(errorLower, "timeout") ||
		strings.Contains(errorLower, "busy")
}
