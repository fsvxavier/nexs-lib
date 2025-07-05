package rabbitmq

import (
	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// NewRabbitMQProducer cria um novo producer RabbitMQ (stub)
func NewRabbitMQProducer(provider *RabbitMQProvider, config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	return nil, domainerrors.New(
		"NOT_IMPLEMENTED",
		"RabbitMQ producer not yet implemented",
	).WithType(domainerrors.ErrorTypeRepository)
}

// NewRabbitMQConsumer cria um novo consumer RabbitMQ (stub)
func NewRabbitMQConsumer(provider *RabbitMQProvider, config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	return nil, domainerrors.New(
		"NOT_IMPLEMENTED",
		"RabbitMQ consumer not yet implemented",
	).WithType(domainerrors.ErrorTypeRepository)
}
