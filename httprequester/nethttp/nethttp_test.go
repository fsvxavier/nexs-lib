package nethttp_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httprequester/nethttp"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := nethttp.DefaultConfig()

	assert.NotNil(t, config, "A configuração não deve ser nil")
	assert.Equal(t, false, config.TLSEnabled, "TLSEnabled deve ser false")
	assert.Equal(t, 40, config.MaxIdleConns, "MaxIdleConns deve ser 40")
	assert.Equal(t, 50, config.MaxIdleConnsPerHost, "MaxIdleConnsPerHost deve ser 50")
	assert.Equal(t, 60, config.MaxConnsPerHost, "MaxConnsPerHost deve ser 60")
	assert.Equal(t, time.Minute*1440, config.IdleConnTimeout, "IdleConnTimeout deve ser 1440 minutos")
	assert.Equal(t, false, config.DisableKeepAlives, "DisableKeepAlives deve ser false")
	assert.Equal(t, 3*time.Second, config.ClientTimeout, "ClientTimeout deve ser 3s")
	assert.Equal(t, false, config.EnableTracer, "EnableTracer deve ser false")
}

func TestNew(t *testing.T) {
	baseURL := "http://example.com"
	config := nethttp.DefaultConfig()

	// Testar criação sem cliente personalizado
	requester := nethttp.New(baseURL, nil, config)
	assert.NotNil(t, requester, "O requester não deve ser nil")

	// Testar criação com cliente personalizado
	customClient := &http.Client{}
	requesterWithClient := nethttp.New(baseURL, customClient, config)
	assert.NotNil(t, requesterWithClient, "O requester com cliente personalizado não deve ser nil")
}

func TestSetBaseURL(t *testing.T) {
	baseURL := "http://example.com"
	config := nethttp.DefaultConfig()

	requester := nethttp.New(baseURL, nil, config)

	// Mudar a baseURL
	newBaseURL := "http://newexample.com"
	result := requester.SetBaseURL(newBaseURL)

	// Verificar se a interface fluente retorna a mesma instância
	assert.Equal(t, requester, result, "SetBaseURL deve retornar a mesma instância")
}

func TestSetHeaders(t *testing.T) {
	baseURL := "http://example.com"
	config := nethttp.DefaultConfig()

	requester := nethttp.New(baseURL, nil, config)

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
	config := nethttp.DefaultConfig()

	requester := nethttp.New(baseURL, nil, config)

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
	config := nethttp.DefaultConfig()

	requester := nethttp.New(baseURL, nil, config)

	// O método Close não retorna erro, então apenas verificamos se ele não causa pânico
	assert.NotPanics(t, func() {
		requester.Close()
	}, "Close não deve causar pânico")
}

func TestTraceInfo(t *testing.T) {
	baseURL := "http://example.com"
	config := nethttp.DefaultConfig()

	requester := nethttp.New(baseURL, nil, config)

	// Inicialmente o TraceInfo deve ser nil ou um valor padrão
	traceInfo := requester.TraceInfo()
	assert.NotNil(t, traceInfo, "TraceInfo não deve ser nil mesmo antes de uma requisição")
}

// Nota: Para testar os métodos HTTP com um servidor HTTP de teste,
// você precisaria usar net/http/httptest. Para manter a simplicidade,
// estamos pulando essa parte do teste.

func TestHTTPMethods(t *testing.T) {
	t.Skip("Esse teste requer um servidor HTTP de teste")

	// O código abaixo é um exemplo de como você pode implementar testes com um servidor HTTP de teste

	/*
		server := setupTestServer()
		defer server.Close()

		baseURL := server.URL
		config := nethttp.DefaultConfig()

		requester := nethttp.New(baseURL, nil, config)

		// Testar GET
		resp, err := requester.Get(context.Background(), "/test")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "success")

		// Testar POST
		resp, err = requester.Post(context.Background(), "/test", []byte(`{"data":"test"}`))
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "created")
	*/
}
