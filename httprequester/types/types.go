package types

import (
	"context"
	"time"
)

// Response representa a resposta HTTP padronizada
type Response struct {
	Body       []byte
	StatusCode int
	IsError    bool
}

// TraceInfo estrutura para informações de trace da requisição
type TraceInfo struct {
	DNSLookup      time.Duration
	ConnTime       time.Duration
	TCPConnTime    time.Duration
	TLSHandshake   time.Duration
	ServerTime     time.Duration
	ResponseTime   time.Duration
	TotalTime      time.Duration
	IsConnReused   bool
	IsConnWasIdle  bool
	ConnIdleTime   time.Duration
	RequestAttempt int
}

// IRequester é a interface comum para todos os clientes HTTP
type IRequester interface {
	// Métodos HTTP
	Get(ctx context.Context, endpoint string) (*Response, error)
	Post(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Put(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Delete(ctx context.Context, endpoint string) (*Response, error)
	Patch(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Head(ctx context.Context, endpoint string) (*Response, error)

	// Configuração
	SetHeaders(headers map[string]string) IRequester
	SetBaseURL(baseURL string) IRequester

	// Unmarshaling
	Unmarshal(v interface{}) IRequester

	// Informações de trace
	TraceInfo() *TraceInfo

	// Gerenciamento de conexão
	Close() error
}

// ErrorHandler é um tipo de função para tratamento de erros da resposta
type ErrorHandler func(*Response) error
