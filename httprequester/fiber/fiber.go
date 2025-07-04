package fiber

import (
	"context"
	"errors"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/fsvxavier/nexs-lib/httprequester/types"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var (
	headerContentTypeJson = []byte("application/json")
	json                  = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Config contém as configurações do cliente Fiber
type Config struct {
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	MaxIdleConnDuration time.Duration
	MaxConnDuration     time.Duration
	MaxConnWaitTimeout  time.Duration
	MaxConns            int
}

// DefaultConfig retorna a configuração padrão para o cliente Fiber
func DefaultConfig() *Config {
	return &Config{
		ReadTimeout:         500 * time.Millisecond,
		WriteTimeout:        500 * time.Millisecond,
		MaxIdleConnDuration: 30 * time.Minute,
		MaxConnDuration:     30 * time.Second,
		MaxConnWaitTimeout:  3 * time.Second,
		MaxConns:            2000,
	}
}

// Requester implementa a interface types.IRequester
type Requester struct {
	client          *fiber.Client
	agent           *fiber.Agent
	headers         map[string]string
	baseURL         string
	structUnmarshal interface{}
	errHandler      types.ErrorHandler
	config          *Config
}

// NewClient cria uma nova instância do cliente Fiber
func NewClient(config *Config) *fiber.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := fiber.AcquireClient()
	client.JSONEncoder = json.Marshal
	client.JSONDecoder = json.Unmarshal

	return client
}

// New cria um novo requester com a URL base especificada
func New(baseURL string, config *Config) *Requester {
	if config == nil {
		config = DefaultConfig()
	}

	client := fiber.AcquireClient()
	client.JSONEncoder = json.Marshal
	client.JSONDecoder = json.Unmarshal

	return &Requester{
		baseURL: baseURL,
		client:  client,
		config:  config,
		headers: make(map[string]string),
	}
}

// SetHeaders configura os cabeçalhos para todas as requisições
func (r *Requester) SetHeaders(headers map[string]string) types.IRequester {
	r.headers = headers
	return r
}

// SetClient define o cliente Fiber a ser usado
func (r *Requester) SetClient(client *fiber.Client) *Requester {
	if r.client != nil {
		fiber.ReleaseClient(r.client)
	}
	r.client = client
	return r
}

// SetBaseURL configura a URL base para todas as requisições
func (r *Requester) SetBaseURL(baseURL string) types.IRequester {
	r.baseURL = baseURL
	return r
}

// SetErrorHandler configura o tratador de erros
func (r *Requester) SetErrorHandler(h types.ErrorHandler) *Requester {
	r.errHandler = h
	return r
}

// Unmarshal configura a estrutura para desserialização da resposta
func (r *Requester) Unmarshal(v interface{}) types.IRequester {
	r.structUnmarshal = v
	return r
}

// Get realiza uma requisição HTTP GET
func (r *Requester) Get(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodGet, endpoint, nil)
}

// Post realiza uma requisição HTTP POST
func (r *Requester) Post(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodPost, endpoint, body)
}

// Put realiza uma requisição HTTP PUT
func (r *Requester) Put(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodPut, endpoint, body)
}

// Delete realiza uma requisição HTTP DELETE
func (r *Requester) Delete(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodDelete, endpoint, nil)
}

// Patch realiza uma requisição HTTP PATCH
func (r *Requester) Patch(ctx context.Context, endpoint string, body []byte) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodPatch, endpoint, body)
}

// Head realiza uma requisição HTTP HEAD
func (r *Requester) Head(ctx context.Context, endpoint string) (*types.Response, error) {
	return r.Execute(ctx, fiber.MethodHead, endpoint, nil)
}

// TraceInfo retorna informações de trace da última requisição
func (r *Requester) TraceInfo() *types.TraceInfo {
	// O Fiber não tem informações detalhadas de trace como net/http,
	// então retornamos uma estrutura básica
	return &types.TraceInfo{}
}

// Close libera os recursos do cliente
func (r *Requester) Close() error {
	if r.agent != nil {
		fiber.ReleaseAgent(r.agent)
	}
	if r.client != nil {
		fiber.ReleaseClient(r.client)
	}
	return nil
}

// Execute realiza a requisição HTTP com o método, endpoint e corpo fornecidos
func (r *Requester) Execute(ctx context.Context, method, endpoint string, body []byte) (*types.Response, error) {
	// Se houver um span do Datadog no contexto, injetamos os cabeçalhos de trace
	ddSpan, ok := tracer.SpanFromContext(ctx)
	if ok {
		err := tracer.Inject(ddSpan.Context(), tracer.TextMapCarrier(r.headers))
		if err != nil {
			return nil, err
		}
	}

	// Configuramos o agent para reutilização de conexões
	agent := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(agent)

	// Configuração do HostClient
	agent.HostClient = &fasthttp.HostClient{
		ReadTimeout:              r.config.ReadTimeout,
		WriteTimeout:             r.config.WriteTimeout,
		MaxIdleConnDuration:      r.config.MaxIdleConnDuration,
		MaxConnDuration:          r.config.MaxConnDuration,
		MaxConnWaitTimeout:       r.config.MaxConnWaitTimeout,
		MaxConns:                 r.config.MaxConns,
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	// Configurando a requisição
	agent.Request().Header.SetMethod(method)
	agent.Request().SetRequestURI(r.baseURL + endpoint)
	agent.InsecureSkipVerify()
	agent.Reuse()
	agent.Request().Header.SetContentTypeBytes(headerContentTypeJson)

	// Adicionando o corpo se houver
	if body != nil {
		agent.Request().SetBody(body)
	}

	// Adicionando os cabeçalhos
	for k, v := range r.headers {
		agent.Request().Header.Set(k, v)
	}

	// Fazendo o parse da requisição
	err := agent.Parse()
	if err != nil {
		return nil, err
	}

	// Executando a requisição
	isErrors := false
	respStatusCode, respBody, respErrs := agent.Bytes()
	if len(respErrs) > 0 {
		isErrors = true
	}

	// Desserializando o resultado se necessário
	if r.structUnmarshal != nil {
		err = json.Unmarshal(respBody, r.structUnmarshal)
		if err != nil {
			return nil, err
		}
	}

	// Criando a resposta
	response := &types.Response{
		Body:       respBody,
		StatusCode: respStatusCode,
		IsError:    isErrors,
	}

	// Tratando erros se necessário
	if isErrors && r.errHandler != nil {
		err = r.errHandler(response)
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

// ErrorHandler é um tratador de erros padrão
func DefaultErrorHandler(res *types.Response) error {
	if res.IsError {
		return errors.New(string(res.Body))
	}
	return nil
}
