package resty

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/fsvxavier/nexs-lib/httprequester/types"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Config contém as configurações do cliente Resty
type Config struct {
	EnableTrace     bool
	EnableTraceLogs bool
	Timeout         time.Duration
	RetryCount      int
	RetryWaitTime   time.Duration
	MaxRetryWait    time.Duration
}

// DefaultConfig retorna a configuração padrão para o cliente Resty
func DefaultConfig() *Config {
	return &Config{
		EnableTrace:     true,
		EnableTraceLogs: false,
		Timeout:         30 * time.Second,
		RetryCount:      3,
		RetryWaitTime:   100 * time.Millisecond,
		MaxRetryWait:    2 * time.Second,
	}
}

// Requester implementa a interface httprequester.IRequester
type Requester struct {
	client          *resty.Client
	request         *resty.Request
	headers         map[string]string
	baseURL         string
	structUnmarshal interface{}
	errHandler      types.ErrorHandler
	lastTraceInfo   *types.TraceInfo
	config          *Config
}

// NewClient cria uma nova instância do cliente Resty
func NewClient(config *Config) *resty.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := resty.New()
	client.JSONMarshal = json.Marshal
	client.JSONUnmarshal = json.Unmarshal
	client.SetTimeout(config.Timeout)
	client.SetRetryCount(config.RetryCount)
	client.SetRetryWaitTime(config.RetryWaitTime)
	client.SetRetryMaxWaitTime(config.MaxRetryWait)

	if config.EnableTrace {
		client.EnableTrace()
	} else {
		client.DisableTrace()
	}

	return client
}

// New cria um novo requester com a URL base e cliente especificados
func New(baseURL string, client *resty.Client, config *Config) *Requester {
	if client == nil {
		client = NewClient(config)
	}

	if config == nil {
		config = DefaultConfig()
	}

	return &Requester{
		client:  client,
		baseURL: baseURL,
		headers: make(map[string]string),
		config:  config,
		request: client.R(),
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
	return r.Execute(ctx, http.MethodPost, endpoint, body)
}

// Put realiza uma requisição HTTP PUT
func (r *Requester) Put(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, http.MethodPut, endpoint, body)
}

// Delete realiza uma requisição HTTP DELETE
func (r *Requester) Delete(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, http.MethodDelete, endpoint, nil)
}

// Patch realiza uma requisição HTTP PATCH
func (r *Requester) Patch(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, http.MethodPatch, endpoint, body)
}

// Head realiza uma requisição HTTP HEAD
func (r *Requester) Head(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, http.MethodHead, endpoint, nil)
}

// TraceInfo retorna informações de trace da última requisição
func (r *Requester) TraceInfo() *types.TraceInfo {
	return r.lastTraceInfo
}

// Close limpa recursos do cliente
func (r *Requester) Close() error {
	// Resty não precisa de fechamento explícito
	return nil
}

// Execute realiza a requisição HTTP com o método, endpoint e corpo fornecidos
func (r *Requester) Execute(ctx context.Context, method, endpoint string, body []byte) (*types.Response, error) {
	// Iniciamos o span do tracer se o contexto possuir um
	span, ctx := tracer.StartSpanFromContext(ctx, "http.request")
	defer span.Finish()

	// Criamos uma nova requisição para cada chamada
	req := r.client.R()

	// Definimos o corpo se houver
	if body != nil {
		req.SetBody(body)
	}

	// Configuramos a URL base e os cabeçalhos
	r.client.SetBaseURL(r.baseURL)
	r.client.SetHeaders(r.headers)

	// Configuramos o contexto da requisição
	req.SetContext(ctx)

	// Injetamos o span do Datadog nos cabeçalhos
	if span != nil {
		err := tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(r.client.Header))
		if err != nil {
			return nil, err
		}
	}

	// Se houver uma estrutura para desserializar, configuramos
	if r.structUnmarshal != nil {
		req.SetResult(r.structUnmarshal)
	}

	// Executamos a requisição
	uriRequest := r.baseURL + endpoint
	res, err := req.Execute(method, uriRequest)
	if err != nil {
		return nil, err
	}

	// Se o trace estiver habilitado e os logs também, exibimos informações de trace
	if r.config.EnableTrace && r.config.EnableTraceLogs {
		ti := res.Request.TraceInfo()
		fmt.Println("Request Info:")
		fmt.Println("  URI       :", uriRequest)
		fmt.Println("Response Info:")
		fmt.Println("  Status Code:", res.StatusCode())
		fmt.Println("  Time       :", res.Time())
		fmt.Println("Trace Info:")
		fmt.Println("  DNSLookup     :", ti.DNSLookup)
		fmt.Println("  ConnTime      :", ti.ConnTime)
		fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
		fmt.Println("  ServerTime    :", ti.ServerTime)
		fmt.Println("  ResponseTime  :", ti.ResponseTime)
		fmt.Println("  TotalTime     :", ti.TotalTime)
		fmt.Println("  IsConnReused  :", ti.IsConnReused)
		fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
		fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	}

	// Convertemos as informações de trace para o nosso formato
	if r.config.EnableTrace {
		ti := res.Request.TraceInfo()
		r.lastTraceInfo = &types.TraceInfo{
			DNSLookup:      ti.DNSLookup,
			ConnTime:       ti.ConnTime,
			TLSHandshake:   ti.TLSHandshake,
			ServerTime:     ti.ServerTime,
			ResponseTime:   ti.ResponseTime,
			TotalTime:      ti.TotalTime,
			IsConnReused:   ti.IsConnReused,
			IsConnWasIdle:  ti.IsConnWasIdle,
			ConnIdleTime:   ti.ConnIdleTime,
			RequestAttempt: ti.RequestAttempt,
		}
	}

	// Criamos a resposta padronizada
	response := &types.Response{
		Body:       res.Body(),
		StatusCode: res.StatusCode(),
		IsError:    res.IsError(),
	}

	// Tratamos erros se houver um tratador configurado e a resposta indicar erro
	if res.IsError() && r.errHandler != nil {
		err = r.errHandler(response)
		if err != nil {
			return response, err
		}
	} else if res.StatusCode() < 200 || res.StatusCode() >= 300 {
		// Se não houver um tratador de erro, mas o código de status indicar erro
		return response, fmt.Errorf("HTTP error: %d - %s", res.StatusCode(), string(res.Body()))
	}

	return response, nil
}

// SetErrorHandler configura o tratador de erros
func (r *Requester) SetErrorHandler(h types.ErrorHandler) *Requester {
	r.errHandler = h
	return r
}

// DefaultErrorHandler é um tratador de erros padrão
func DefaultErrorHandler(res *types.Response) error {
	if res.IsError {
		return fmt.Errorf("HTTP error: %d - %s", res.StatusCode, string(res.Body))
	}
	return nil
}
