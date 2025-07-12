// Package parsers implementa parsers especializados para diferentes tipos de sistemas
// seguindo o padrão Strategy para diferentes estratégias de parsing.
package parsers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCErrorParser parser para erros gRPC.
type GRPCErrorParser struct {
	codeMapping map[codes.Code]string
}

// NewGRPCErrorParser cria um novo parser para gRPC.
func NewGRPCErrorParser() interfaces.ErrorParser {
	return &GRPCErrorParser{
		codeMapping: map[codes.Code]string{
			codes.OK:                 "GRPC_OK",
			codes.Canceled:           "GRPC_CANCELED",
			codes.Unknown:            "GRPC_UNKNOWN",
			codes.InvalidArgument:    "GRPC_INVALID_ARGUMENT",
			codes.DeadlineExceeded:   "GRPC_DEADLINE_EXCEEDED",
			codes.NotFound:           "GRPC_NOT_FOUND",
			codes.AlreadyExists:      "GRPC_ALREADY_EXISTS",
			codes.PermissionDenied:   "GRPC_PERMISSION_DENIED",
			codes.ResourceExhausted:  "GRPC_RESOURCE_EXHAUSTED",
			codes.FailedPrecondition: "GRPC_FAILED_PRECONDITION",
			codes.Aborted:            "GRPC_ABORTED",
			codes.OutOfRange:         "GRPC_OUT_OF_RANGE",
			codes.Unimplemented:      "GRPC_UNIMPLEMENTED",
			codes.Internal:           "GRPC_INTERNAL",
			codes.Unavailable:        "GRPC_UNAVAILABLE",
			codes.DataLoss:           "GRPC_DATA_LOSS",
			codes.Unauthenticated:    "GRPC_UNAUTHENTICATED",
		},
	}
}

// CanParse verifica se pode processar o erro gRPC.
func (p *GRPCErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	// Verifica se é um erro de status gRPC
	if _, ok := status.FromError(err); ok {
		return true
	}

	// Verifica na string do erro
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "grpc") ||
		strings.Contains(errStr, "rpc error") ||
		strings.Contains(errStr, "status code")
}

// Parse processa erro gRPC.
func (p *GRPCErrorParser) Parse(err error) interfaces.ParsedError {
	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeNetwork),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
	}

	// Tenta extrair status gRPC
	if st, ok := status.FromError(err); ok {
		code := st.Code()
		parsed.Code = p.codeMapping[code]
		parsed.Message = st.Message()
		parsed.Details["grpc_code"] = code.String()
		parsed.Details["grpc_code_number"] = int(code)

		// Mapeia severidade baseada no código
		parsed.Severity = p.mapSeverity(code)
		parsed.Retryable = p.isRetryable(code)
		parsed.Temporary = p.isTemporary(code)

		// Adiciona detalhes dos metadados se disponíveis
		if details := st.Details(); len(details) > 0 {
			parsed.Details["grpc_details"] = details
		}
	} else {
		// Fallback para parsing de string
		parsed.Code = "GRPC_UNKNOWN"
		parsed.Message = err.Error()
		parsed.Severity = interfaces.SeverityMedium
		parsed.Details["raw_error"] = err.Error()
	}

	return parsed
}

// mapSeverity mapeia códigos gRPC para severidade.
func (p *GRPCErrorParser) mapSeverity(code codes.Code) interfaces.Severity {
	switch code {
	case codes.Internal, codes.DataLoss:
		return interfaces.SeverityCritical
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return interfaces.SeverityHigh
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return interfaces.SeverityMedium
	default:
		return interfaces.SeverityLow
	}
}

// isRetryable verifica se o erro é retryable.
func (p *GRPCErrorParser) isRetryable(code codes.Code) bool {
	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}

// isTemporary verifica se o erro é temporário.
func (p *GRPCErrorParser) isTemporary(code codes.Code) bool {
	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// HTTPErrorParser parser para erros HTTP.
type HTTPErrorParser struct {
	statusCodeRegex *regexp.Regexp
}

// NewHTTPErrorParser cria um novo parser para HTTP.
func NewHTTPErrorParser() interfaces.ErrorParser {
	return &HTTPErrorParser{
		statusCodeRegex: regexp.MustCompile(`(\d{3})\s+(.+)`),
	}
}

// CanParse verifica se pode processar o erro HTTP.
func (p *HTTPErrorParser) CanParse(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "http") ||
		strings.Contains(errStr, "status") ||
		p.statusCodeRegex.MatchString(err.Error()) ||
		strings.Contains(errStr, "client") ||
		strings.Contains(errStr, "server")
}

// Parse processa erro HTTP.
func (p *HTTPErrorParser) Parse(err error) interfaces.ParsedError {
	errStr := err.Error()

	parsed := interfaces.ParsedError{
		Type:     string(types.ErrorTypeNetwork),
		Category: interfaces.CategoryTechnical,
		Details:  make(map[string]interface{}),
		Message:  errStr,
	}

	// Tenta extrair status code
	if matches := p.statusCodeRegex.FindStringSubmatch(errStr); len(matches) > 2 {
		if statusCode, err := strconv.Atoi(matches[1]); err == nil {
			parsed.Code = fmt.Sprintf("HTTP_%d", statusCode)
			parsed.Details["status_code"] = statusCode
			parsed.Details["status_text"] = matches[2]

			// Mapeia propriedades baseadas no status code
			parsed.Severity = p.mapSeverityByStatus(statusCode)
			parsed.Retryable = p.isRetryableByStatus(statusCode)
			parsed.Temporary = p.isTemporaryByStatus(statusCode)
		}
	} else {
		// Fallback
		parsed.Code = "HTTP_UNKNOWN"
		parsed.Severity = interfaces.SeverityMedium
		parsed.Details["raw_error"] = errStr
	}

	return parsed
}

// mapSeverityByStatus mapeia status HTTP para severidade.
func (p *HTTPErrorParser) mapSeverityByStatus(statusCode int) interfaces.Severity {
	switch {
	case statusCode >= 500:
		return interfaces.SeverityHigh
	case statusCode >= 400:
		return interfaces.SeverityMedium
	case statusCode >= 300:
		return interfaces.SeverityLow
	default:
		return interfaces.SeverityLow
	}
}

// isRetryableByStatus verifica se o status HTTP é retryable.
func (p *HTTPErrorParser) isRetryableByStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// isTemporaryByStatus verifica se o status HTTP é temporário.
func (p *HTTPErrorParser) isTemporaryByStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
