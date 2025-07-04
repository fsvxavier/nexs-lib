package types_test

import (
	"errors"
	"testing"

	"github.com/fsvxavier/nexs-lib/httprequester/types"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	// Verificar se os erros são únicos e não nil
	assert.NotNil(t, types.ErrInvalidClient, "ErrInvalidClient não deve ser nil")
	assert.NotNil(t, types.ErrInvalidResponse, "ErrInvalidResponse não deve ser nil")
	assert.NotNil(t, types.ErrTimeout, "ErrTimeout não deve ser nil")
	assert.NotNil(t, types.ErrServiceUnavailable, "ErrServiceUnavailable não deve ser nil")

	// Verificar se as mensagens dos erros são diferentes
	clientErrMsg := types.ErrInvalidClient.Error()
	responseErrMsg := types.ErrInvalidResponse.Error()
	timeoutErrMsg := types.ErrTimeout.Error()
	unavailableErrMsg := types.ErrServiceUnavailable.Error()

	assert.NotEqual(t, clientErrMsg, responseErrMsg, "As mensagens de erro devem ser diferentes")
	assert.NotEqual(t, clientErrMsg, timeoutErrMsg, "As mensagens de erro devem ser diferentes")
	assert.NotEqual(t, clientErrMsg, unavailableErrMsg, "As mensagens de erro devem ser diferentes")
	assert.NotEqual(t, responseErrMsg, timeoutErrMsg, "As mensagens de erro devem ser diferentes")
	assert.NotEqual(t, responseErrMsg, unavailableErrMsg, "As mensagens de erro devem ser diferentes")
	assert.NotEqual(t, timeoutErrMsg, unavailableErrMsg, "As mensagens de erro devem ser diferentes")
}

func TestResponse(t *testing.T) {
	// Testar criação de uma resposta
	response := types.Response{
		Body:       []byte("teste"),
		StatusCode: 200,
		IsError:    false,
	}

	assert.Equal(t, []byte("teste"), response.Body, "Body deve ser igual ao valor definido")
	assert.Equal(t, 200, response.StatusCode, "StatusCode deve ser igual ao valor definido")
	assert.Equal(t, false, response.IsError, "IsError deve ser igual ao valor definido")
}

func TestTraceInfo(t *testing.T) {
	// Testar criação de uma informação de trace
	traceInfo := types.TraceInfo{
		IsConnReused:   true,
		IsConnWasIdle:  true,
		RequestAttempt: 1,
	}

	assert.Equal(t, true, traceInfo.IsConnReused, "IsConnReused deve ser igual ao valor definido")
	assert.Equal(t, true, traceInfo.IsConnWasIdle, "IsConnWasIdle deve ser igual ao valor definido")
	assert.Equal(t, 1, traceInfo.RequestAttempt, "RequestAttempt deve ser igual ao valor definido")
}

func TestErrorHandlerFunc(t *testing.T) {
	// Criar um ErrorHandler de exemplo
	customErr := errors.New("erro personalizado")

	// Handler que sempre retorna o erro personalizado
	handler := func(err error) error {
		return customErr
	}

	// Testar o handler
	inputErr := errors.New("erro de entrada")
	result := handler(inputErr)

	assert.Equal(t, customErr, result, "O handler deve retornar o erro personalizado")
}
