package httprequester_test

import (
	"net/http"
	"testing"

	"github.com/fsvxavier/nexs-lib/httprequester"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	factory := httprequester.NewFactory()
	assert.NotNil(t, factory, "A factory não deve ser nil")
}

func TestFactoryCreate(t *testing.T) {
	factory := httprequester.NewFactory()

	tests := []struct {
		name       string
		clientType httprequester.ClientType
		baseURL    string
		expectNil  bool
	}{
		{
			name:       "Criação de cliente Fiber",
			clientType: httprequester.ClientFiber,
			baseURL:    "http://example.com",
			expectNil:  false,
		},
		{
			name:       "Criação de cliente Resty",
			clientType: httprequester.ClientResty,
			baseURL:    "http://example.com",
			expectNil:  false,
		},
		{
			name:       "Criação de cliente NetHttp",
			clientType: httprequester.ClientNetHttp,
			baseURL:    "http://example.com",
			expectNil:  false,
		},
		{
			name:       "Tipo de cliente desconhecido (usa padrão NetHttp)",
			clientType: "desconhecido",
			baseURL:    "http://example.com",
			expectNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := factory.Create(tt.clientType, tt.baseURL)
			if tt.expectNil {
				assert.Nil(t, client, "Cliente não deveria ser criado")
			} else {
				assert.NotNil(t, client, "Cliente deveria ser criado")
			}
		})
	}
}

func TestFactoryCreateWithClient(t *testing.T) {
	factory := httprequester.NewFactory()

	// Criar clientes para passar para a factory
	fiberClient := fiber.AcquireClient()
	restyClient := resty.New()
	httpClient := &http.Client{}

	tests := []struct {
		name       string
		clientType httprequester.ClientType
		baseURL    string
		client     interface{}
		expectErr  bool
	}{
		{
			name:       "Criação de cliente Fiber com cliente personalizado",
			clientType: httprequester.ClientFiber,
			baseURL:    "http://example.com",
			client:     fiberClient,
			expectErr:  false,
		},
		{
			name:       "Criação de cliente Resty com cliente personalizado",
			clientType: httprequester.ClientResty,
			baseURL:    "http://example.com",
			client:     restyClient,
			expectErr:  false,
		},
		{
			name:       "Criação de cliente NetHttp com cliente personalizado",
			clientType: httprequester.ClientNetHttp,
			baseURL:    "http://example.com",
			client:     httpClient,
			expectErr:  false,
		},
		{
			name:       "Criação de cliente Fiber com cliente inválido",
			clientType: httprequester.ClientFiber,
			baseURL:    "http://example.com",
			client:     "cliente inválido",
			expectErr:  true,
		},
		{
			name:       "Criação de cliente Resty com cliente inválido",
			clientType: httprequester.ClientResty,
			baseURL:    "http://example.com",
			client:     "cliente inválido",
			expectErr:  true,
		},
		{
			name:       "Criação de cliente NetHttp com cliente inválido",
			clientType: httprequester.ClientNetHttp,
			baseURL:    "http://example.com",
			client:     "cliente inválido",
			expectErr:  true,
		},
		{
			name:       "Tipo de cliente desconhecido com cliente válido",
			clientType: "desconhecido",
			baseURL:    "http://example.com",
			client:     httpClient,
			expectErr:  false,
		},
		{
			name:       "Clientes nil são permitidos",
			clientType: httprequester.ClientFiber,
			baseURL:    "http://example.com",
			client:     nil,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateWithClient(tt.clientType, tt.baseURL, tt.client)

			if tt.expectErr {
				assert.Error(t, err, "Deveria retornar um erro")
				assert.Nil(t, client, "Cliente deveria ser nil quando há erro")
			} else {
				assert.NoError(t, err, "Não deveria retornar erro")
				assert.NotNil(t, client, "Cliente não deveria ser nil")
			}
		})
	}
}
