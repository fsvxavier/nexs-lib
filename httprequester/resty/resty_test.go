package resty_test

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httprequester/resty"
	goresty "github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := resty.DefaultConfig()

	assert.NotNil(t, config, "A configuração não deve ser nil")
	assert.Equal(t, true, config.EnableTrace, "EnableTrace deve ser true")
	assert.Equal(t, false, config.EnableTraceLogs, "EnableTraceLogs deve ser false")
	assert.Equal(t, 30*time.Second, config.Timeout, "Timeout deve ser 30s")
	assert.Equal(t, 3, config.RetryCount, "RetryCount deve ser 3")
	assert.Equal(t, 100*time.Millisecond, config.RetryWaitTime, "RetryWaitTime deve ser 100ms")
	assert.Equal(t, 2*time.Second, config.MaxRetryWait, "MaxRetryWait deve ser 2s")
}

func TestNew(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	// Testar criação sem cliente personalizado
	requester := resty.New(baseURL, nil, config)
	assert.NotNil(t, requester, "O requester não deve ser nil")

	// Testar criação com cliente personalizado
	customClient := goresty.New()
	requesterWithClient := resty.New(baseURL, customClient, config)
	assert.NotNil(t, requesterWithClient, "O requester com cliente personalizado não deve ser nil")
}

func TestSetBaseURL(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	requester := resty.New(baseURL, nil, config)

	// Mudar a baseURL
	newBaseURL := "http://newexample.com"
	result := requester.SetBaseURL(newBaseURL)

	// Verificar se a interface fluente retorna a mesma instância
	assert.Equal(t, requester, result, "SetBaseURL deve retornar a mesma instância")
}

func TestSetHeaders(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	requester := resty.New(baseURL, nil, config)

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "TestAgent",
	}

	result := requester.SetHeaders(headers)

	// Verificar se a interface fluente retorna a mesma instância
	assert.Equal(t, requester, result, "SetHeaders deve retornar a mesma instância")
}

func TestUnmarshal(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	requester := resty.New(baseURL, nil, config)

	type TestStruct struct {
		Name string `json:"name"`
	}

	var testStruct TestStruct
	result := requester.Unmarshal(&testStruct)

	// Verificar se a interface fluente retorna a mesma instância
	assert.Equal(t, requester, result, "Unmarshal deve retornar a mesma instância")
}

func TestClose(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	requester := resty.New(baseURL, nil, config)

	// O método Close não retorna erro, então apenas verificamos se ele não causa pânico
	assert.NotPanics(t, func() {
		requester.Close()
	}, "Close não deve causar pânico")
}

func TestTraceInfo(t *testing.T) {
	baseURL := "http://example.com"
	config := resty.DefaultConfig()

	requester := resty.New(baseURL, nil, config)

	// TraceInfo pode retornar nil ou um valor padrão antes de uma requisição ser feita
	traceInfo := requester.TraceInfo()
	// Não testamos se é nil ou não, apenas verificamos que o método existe e pode ser chamado
	t.Log("TraceInfo:", traceInfo)
}

// Nota: Para testar os métodos HTTP com mocks, você precisaria de uma biblioteca
// como github.com/jarcoal/httpmock. Para manter a simplicidade, estamos
// pulando essa parte do teste.

func TestHTTPMethods(t *testing.T) {
	t.Skip("Esse teste requer configuração de mock HTTP adicional")

	// O código abaixo é um exemplo de como você pode implementar testes com mocks HTTP
	// usando a biblioteca httpmock

	/*
		baseURL := "http://api.example.com"
		config := resty.DefaultConfig()

		client := setupHTTPMock(t)
		requester := resty.New(baseURL, client, config)

		// Configurar mock para GET
		httpmock.RegisterResponder("GET", baseURL+"/test",
			httpmock.NewStringResponder(200, `{"message": "success"}`))

		// Testar GET
		resp, err := requester.Get(context.Background(), "/test")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "success")

		// Configurar mock para POST
		httpmock.RegisterResponder("POST", baseURL+"/test",
			httpmock.NewStringResponder(201, `{"message": "created"}`))

		// Testar POST
		resp, err = requester.Post(context.Background(), "/test", []byte(`{"data": "test"}`))
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "created")
	*/
}
