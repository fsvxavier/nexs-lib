package mocks

import (
	rgo "github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/mock"
)

type RedisConnMock struct {
	mock.Mock
	rgo.Conn
}

func (conn *RedisConnMock) Do(commandName string, arguments ...interface{}) (reply interface{}, err error) {
	args := conn.Called(append([]interface{}{commandName}, arguments...)...)
	return args.Get(0), args.Error(1)
}
