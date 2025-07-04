package types

import "errors"

var (
	// ErrInvalidClient é retornado quando o cliente fornecido não é do tipo esperado
	ErrInvalidClient = errors.New("cliente HTTP inválido")

	// ErrInvalidResponse é retornado quando a resposta não pode ser processada
	ErrInvalidResponse = errors.New("resposta HTTP inválida")

	// ErrTimeout é retornado quando a requisição excede o tempo limite
	ErrTimeout = errors.New("timeout na requisição HTTP")

	// ErrServiceUnavailable é retornado quando o serviço não está disponível
	ErrServiceUnavailable = errors.New("serviço não disponível")
)
