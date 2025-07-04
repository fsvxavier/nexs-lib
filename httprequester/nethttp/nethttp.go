package nethttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/fsvxavier/nexs-lib/httprequester/types"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Config contém as configurações do cliente Net/HTTP
type Config struct {
	TLSEnabled          bool
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	DisableKeepAlives   bool
	ClientTimeout       time.Duration
	EnableTracer        bool
}

// DefaultConfig retorna a configuração padrão para o cliente Net/HTTP
func DefaultConfig() *Config {
	return &Config{
		TLSEnabled:          false,
		MaxIdleConns:        40,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     60,
		IdleConnTimeout:     time.Minute * 1440, // 24 horas
		DisableKeepAlives:   false,
		ClientTimeout:       3 * time.Second,
		EnableTracer:        false,
	}
}

// Requester implementa a interface types.IRequester
type Requester struct {
	client          *http.Client
	headers         map[string]string
	baseURL         string
	structUnmarshal interface{}
	errorUnmarshal  interface{}
	enableTracer    bool

	// Campos para trace
	getConn              time.Time
	dnsStart             time.Time
	dnsDone              time.Time
	connectDone          time.Time
	tlsHandshakeStart    time.Time
	tlsHandshakeDone     time.Time
	gotConn              time.Time
	gotFirstResponseByte time.Time
	endTime              time.Time
	gotConnInfo          httptrace.GotConnInfo
}

// NewClient cria um novo cliente HTTP
func NewClient(config *Config) *http.Client {
	if config == nil {
		config = DefaultConfig()
	}

	// Criamos o transporte com as configurações
	transport := &http.Transport{
		Dial: (&net.Dialer{
			DualStack:     false,
			FallbackDelay: 0,
			Timeout:       config.ClientTimeout,
			KeepAlive:     0, // Habilitado se suportado pelo protocolo e sistema operacional
		}).Dial,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
	}

	// Se TLS não estiver habilitado, ignoramos verificação de certificados
	if !config.TLSEnabled {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Criamos o cliente com o transporte configurado
	client := &http.Client{
		Transport: transport,
		Timeout:   config.ClientTimeout,
	}

	return client
}

// New cria um novo requester com a URL base e cliente especificados
func New(baseURL string, client *http.Client, config *Config) *Requester {
	if config == nil {
		config = DefaultConfig()
	}

	if client == nil {
		client = NewClient(config)
	}

	return &Requester{
		client:       client,
		baseURL:      baseURL,
		headers:      make(map[string]string),
		enableTracer: config.EnableTracer,
	}
}

// SetHeaders configura os cabeçalhos para todas as requisições
func (r *Requester) SetHeaders(headers map[string]string) types.IRequester {
	r.headers = headers
	return r
}

// SetBaseURL configura a URL base para todas as requisições
func (r *Requester) SetBaseURL(baseURL string) types.IRequester {
	r.baseURL = baseURL
	return r
}

// Unmarshal configura a estrutura para desserialização da resposta
func (r *Requester) Unmarshal(v interface{}) types.IRequester {
	r.structUnmarshal = v
	return r
}

// Get realiza uma requisição HTTP GET
func (r *Requester) Get(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, http.MethodGet, endpoint, nil)
}

// Post realiza uma requisição HTTP POST
func (r *Requester) Post(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
}

// Put realiza uma requisição HTTP PUT
func (r *Requester) Put(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, http.MethodPut, endpoint, bytes.NewBuffer(body))
}

// Delete realiza uma requisição HTTP DELETE
func (r *Requester) Delete(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, http.MethodDelete, endpoint, nil)
}

// Patch realiza uma requisição HTTP PATCH
func (r *Requester) Patch(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, http.MethodPatch, endpoint, bytes.NewBuffer(body))
}

// Head realiza uma requisição HTTP HEAD
func (r *Requester) Head(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, http.MethodHead, endpoint, nil)
}

// SetErrorUnmarshal configura a estrutura para desserialização de erros
func (r *Requester) SetErrorUnmarshal(v interface{}) *Requester {
	r.errorUnmarshal = v
	return r
}

// TraceInfo retorna informações de trace da última requisição
func (r *Requester) TraceInfo() *types.TraceInfo {
	traceInfo := &types.TraceInfo{
		DNSLookup:     r.dnsDone.Sub(r.dnsStart),
		TLSHandshake:  r.tlsHandshakeDone.Sub(r.tlsHandshakeStart),
		ServerTime:    r.gotFirstResponseByte.Sub(r.gotConn),
		IsConnReused:  r.gotConnInfo.Reused,
		IsConnWasIdle: r.gotConnInfo.WasIdle,
		ConnIdleTime:  r.gotConnInfo.IdleTime,
	}

	// Calculamos o tempo total de acordo com a reutilização da conexão
	if r.gotConnInfo.Reused {
		traceInfo.TotalTime = r.endTime.Sub(r.getConn)
	} else {
		traceInfo.TotalTime = r.endTime.Sub(r.dnsStart)
	}

	// Calculamos apenas para conexões bem-sucedidas
	if !r.connectDone.IsZero() {
		traceInfo.TCPConnTime = r.connectDone.Sub(r.dnsDone)
	}

	// Calculamos apenas para conexões bem-sucedidas
	if !r.gotConn.IsZero() {
		traceInfo.ConnTime = r.gotConn.Sub(r.getConn)
	}

	// Calculamos apenas para conexões bem-sucedidas
	if !r.gotFirstResponseByte.IsZero() {
		traceInfo.ResponseTime = r.endTime.Sub(r.gotFirstResponseByte)
	}

	return traceInfo
}

// Close libera recursos do cliente
func (r *Requester) Close() error {
	// Net/HTTP gerencia conexões automaticamente, mas podemos fechar o transporte
	// para liberar todas as conexões ativas quando o cliente não for mais usado
	transport, ok := r.client.Transport.(*http.Transport)
	if ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// Execute realiza a requisição HTTP com o método, endpoint e corpo fornecidos
func (r *Requester) Execute(ctx context.Context, method, endpoint string, body io.Reader) (*types.Response, error) {
	var req *http.Request
	var err error

	// Obtemos o span do tracer Datadog se estiver presente no contexto
	ddSpan, ok := tracer.SpanFromContext(ctx)
	if ok {
		// Injetamos informações de trace nos cabeçalhos
		err = tracer.Inject(ddSpan.Context(), tracer.TextMapCarrier(r.headers))
		if err != nil {
			return nil, err
		}
	}

	uriRequest := r.baseURL + endpoint

	// Se o tracer estiver habilitado, adicionamos callbacks para coletar métricas
	if r.enableTracer {
		clientTrace := &httptrace.ClientTrace{
			DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
				r.dnsStart = time.Now()
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				r.dnsDone = time.Now()
			},
			ConnectStart: func(network, addr string) {
				if r.dnsDone.IsZero() {
					r.dnsDone = time.Now()
				}
				if r.dnsStart.IsZero() {
					r.dnsStart = r.dnsDone
				}
			},
			ConnectDone: func(network, addr string, err error) {
				r.connectDone = time.Now()
			},
			GetConn: func(hostPort string) {
				r.getConn = time.Now()
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				r.gotConn = time.Now()
				r.gotConnInfo = connInfo
			},
			GotFirstResponseByte: func() {
				r.gotFirstResponseByte = time.Now()
			},
			TLSHandshakeStart: func() {
				r.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(state tls.ConnectionState, err error) {
				r.tlsHandshakeDone = time.Now()
			},
		}

		// Criamos a requisição com o contexto de trace
		req, err = http.NewRequestWithContext(httptrace.WithClientTrace(ctx, clientTrace), method, uriRequest, body)
	} else {
		// Criamos a requisição sem o contexto de trace
		req, err = http.NewRequestWithContext(ctx, method, uriRequest, body)
	}

	if err != nil {
		return nil, err
	}

	// Adicionamos os cabeçalhos à requisição
	if r.headers != nil {
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}
	}

	// Garantimos que o cabeçalho Content-Type esteja definido
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Executamos a requisição
	isError := false
	resp, err := r.client.Do(req)
	if err != nil {
		isError = true
		return nil, err
	}
	defer resp.Body.Close()

	// Registramos o tempo de conclusão
	r.endTime = time.Now()

	// Lemos o corpo da resposta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		isError = true
		return nil, err
	}

	// Desserializamos a resposta se uma estrutura foi configurada
	if r.structUnmarshal != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err = json.Unmarshal(respBody, r.structUnmarshal)
		if err != nil {
			isError = true
			return nil, err
		}
	} else if r.errorUnmarshal != nil && resp.StatusCode >= 400 {
		// Se temos uma estrutura para erros e o código de status é de erro
		err = json.Unmarshal(respBody, r.errorUnmarshal)
		// Ignoramos erros de desserialização aqui, pois o corpo pode não estar no formato esperado
	}

	// Criamos a resposta padronizada
	response := &types.Response{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		IsError:    isError || resp.StatusCode >= 400,
	}

	// Descartar qualquer conteúdo restante
	_, _ = io.Copy(io.Discard, resp.Body)

	return response, nil
}

// SetErrorHandler configura o tratador de erros
func (r *Requester) SetErrorHandler(h types.ErrorHandler) *Requester {
	// Esta função está incluída para compatibilidade com os outros clientes,
	// mas a implementação atual de Execute não a utiliza diretamente
	return r
}

// DefaultErrorHandler é um tratador de erros padrão
func DefaultErrorHandler(res *types.Response) error {
	if res.IsError {
		return fmt.Errorf("HTTP error: %d - %s", res.StatusCode, string(res.Body))
	}
	return nil
}
