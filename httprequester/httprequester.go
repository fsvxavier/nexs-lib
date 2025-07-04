// Package httprequester fornece uma biblioteca unificada para requisições HTTP.
package httprequester

import (
	"net/http"

	"github.com/fsvxavier/nexs-lib/httprequester/fiber"
	"github.com/fsvxavier/nexs-lib/httprequester/nethttp"
	"github.com/fsvxavier/nexs-lib/httprequester/resty"
	"github.com/fsvxavier/nexs-lib/httprequester/types"
	gorest "github.com/go-resty/resty/v2"
	gofib "github.com/gofiber/fiber/v2"
)

// Reexporte os tipos do pacote types para compatibilidade
type (
	Response     = types.Response
	TraceInfo    = types.TraceInfo
	IRequester   = types.IRequester
	ErrorHandler = types.ErrorHandler
)

// Reexporte os erros do pacote types para compatibilidade
var (
	ErrInvalidClient      = types.ErrInvalidClient
	ErrInvalidResponse    = types.ErrInvalidResponse
	ErrTimeout            = types.ErrTimeout
	ErrServiceUnavailable = types.ErrServiceUnavailable
)

// ClientType define o tipo de cliente HTTP
type ClientType string

const (
	// ClientFiber é o cliente Fiber
	ClientFiber ClientType = "fiber"
	// ClientResty é o cliente Resty
	ClientResty ClientType = "resty"
	// ClientNetHttp é o cliente Net/HTTP
	ClientNetHttp ClientType = "nethttp"
)

// Factory para criação de clientes HTTP
type Factory struct{}

// NewFactory cria uma nova instância da factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create cria um novo cliente HTTP de acordo com o tipo especificado
func (f *Factory) Create(clientType ClientType, baseURL string) types.IRequester {
	switch clientType {
	case ClientFiber:
		config := fiber.DefaultConfig()
		return fiber.New(baseURL, config)
	case ClientResty:
		config := resty.DefaultConfig()
		return resty.New(baseURL, nil, config)
	case ClientNetHttp:
		config := nethttp.DefaultConfig()
		return nethttp.New(baseURL, nil, config)
	default:
		// Por padrão, utilizamos o cliente Net/HTTP
		config := nethttp.DefaultConfig()
		return nethttp.New(baseURL, nil, config)
	}
}

// CreateWithClient cria um novo cliente HTTP de acordo com o tipo especificado, utilizando o cliente fornecido
func (f *Factory) CreateWithClient(clientType ClientType, baseURL string, client interface{}) (types.IRequester, error) {
	switch clientType {
	case ClientFiber:
		config := fiber.DefaultConfig()
		if client == nil {
			return fiber.New(baseURL, config), nil
		}
		fiberClient, ok := client.(*gofib.Client)
		if !ok {
			return nil, types.ErrInvalidClient
		}
		// Aqui usamos o cliente Fiber fornecido
		requester := fiber.New(baseURL, config)
		requester.SetClient(fiberClient)
		return requester, nil
	case ClientResty:
		config := resty.DefaultConfig()
		if client == nil {
			return resty.New(baseURL, nil, config), nil
		}
		restyClient, ok := client.(*gorest.Client)
		if !ok {
			return nil, types.ErrInvalidClient
		}
		// Já passamos o cliente Resty diretamente para o construtor
		return resty.New(baseURL, restyClient, config), nil
	case ClientNetHttp:
		config := nethttp.DefaultConfig()
		if client == nil {
			return nethttp.New(baseURL, nil, config), nil
		}
		httpClient, ok := client.(*http.Client)
		if !ok {
			return nil, types.ErrInvalidClient
		}
		return nethttp.New(baseURL, httpClient, config), nil
	default:
		// Por padrão, utilizamos o cliente Net/HTTP
		config := nethttp.DefaultConfig()
		if client == nil {
			return nethttp.New(baseURL, nil, config), nil
		}
		httpClient, ok := client.(*http.Client)
		if !ok {
			return nil, types.ErrInvalidClient
		}
		return nethttp.New(baseURL, httpClient, config), nil
	}
}
