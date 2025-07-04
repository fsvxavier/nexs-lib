package fiber_test

import (
	"testing"
	"time"

	fiberclient "github.com/fsvxavier/nexs-lib/httprequester/fiber"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := fiberclient.DefaultConfig()

	assert.NotNil(t, config, "A configuração não deve ser nil")
	assert.Equal(t, 500*time.Millisecond, config.ReadTimeout, "ReadTimeout deve ser 500ms")
	assert.Equal(t, 500*time.Millisecond, config.WriteTimeout, "WriteTimeout deve ser 500ms")
	assert.Equal(t, 30*time.Minute, config.MaxIdleConnDuration, "MaxIdleConnDuration deve ser 30min")
	assert.Equal(t, 30*time.Second, config.MaxConnDuration, "MaxConnDuration deve ser 30s")
	assert.Equal(t, 3*time.Second, config.MaxConnWaitTimeout, "MaxConnWaitTimeout deve ser 3s")
	assert.Equal(t, 2000, config.MaxConns, "MaxConns deve ser 2000")
}

func TestNew(t *testing.T) {
	baseURL := "http://example.com"
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

	assert.NotNil(t, requester, "O requester não deve ser nil")
}

func TestSetClient(t *testing.T) {
	baseURL := "http://example.com"
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

	// Verificando que o método não causa pânico quando chamado com nil
	assert.NotPanics(t, func() {
		requester.SetClient(nil)
	}, "SetClient não deve causar pânico quando chamado com nil")
}

func TestSetBaseURL(t *testing.T) {
	baseURL := "http://example.com"
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

	// Mudar a baseURL
	newBaseURL := "http://newexample.com"
	result := requester.SetBaseURL(newBaseURL)

	// Verificar se a interface fluente retorna a mesma instância
	assert.Equal(t, requester, result, "SetBaseURL deve retornar a mesma instância")
}

func TestSetHeaders(t *testing.T) {
	baseURL := "http://example.com"
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

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
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

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
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

	// O método Close não retorna erro, então apenas verificamos se ele não causa pânico
	assert.NotPanics(t, func() {
		requester.Close()
	}, "Close não deve causar pânico")
}

func TestTraceInfo(t *testing.T) {
	baseURL := "http://example.com"
	config := fiberclient.DefaultConfig()

	requester := fiberclient.New(baseURL, config)

	// Inicialmente o TraceInfo deve ser nil ou um valor padrão
	traceInfo := requester.TraceInfo()
	assert.NotNil(t, traceInfo, "TraceInfo não deve ser nil mesmo antes de uma requisição")
}

func TestHTTPMethods(t *testing.T) {
	t.Skip("Esse teste requer um servidor HTTP mock apropriado")

	// Aqui você pode implementar testes para os métodos HTTP (GET, POST, etc.)
	// usando um servidor mock adequado para o cliente Fiber
}
