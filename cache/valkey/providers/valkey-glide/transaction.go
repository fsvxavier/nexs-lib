package valkeyglide

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// transaction implementa interfaces.ITransaction para valkey-glide.
type transaction struct {
	commands []interfaces.ICommand
}

// newTransaction cria uma nova transação.
func newTransaction(client interface{}) interfaces.ITransaction {
	return &transaction{
		commands: make([]interfaces.ICommand, 0),
	}
}

// Get adiciona comando GET à transação.
func (t *transaction) Get(key string) interfaces.ICommand {
	cmd := newCommand(nil, fmt.Errorf("comando GET não executado ainda"))
	t.commands = append(t.commands, cmd)
	return cmd
}

// Set adiciona comando SET à transação.
func (t *transaction) Set(key string, value interface{}, expiration time.Duration) interfaces.ICommand {
	cmd := newCommand(nil, fmt.Errorf("comando SET não executado ainda"))
	t.commands = append(t.commands, cmd)
	return cmd
}

// Del adiciona comando DEL à transação.
func (t *transaction) Del(keys ...string) interfaces.ICommand {
	cmd := newCommand(nil, fmt.Errorf("comando DEL não executado ainda"))
	t.commands = append(t.commands, cmd)
	return cmd
}

// HGet adiciona comando HGET à transação.
func (t *transaction) HGet(key, field string) interfaces.ICommand {
	cmd := newCommand(nil, fmt.Errorf("comando HGET não executado ainda"))
	t.commands = append(t.commands, cmd)
	return cmd
}

// HSet adiciona comando HSET à transação.
func (t *transaction) HSet(key string, values ...interface{}) interfaces.ICommand {
	cmd := newCommand(nil, fmt.Errorf("comando HSET não executado ainda"))
	t.commands = append(t.commands, cmd)
	return cmd
}

// Watch adiciona chaves para watch.
func (t *transaction) Watch(ctx context.Context, keys ...string) error {
	// TODO: Implementar watch real usando valkey-glide
	return fmt.Errorf("watch not implemented yet")
}

// Unwatch remove todas as chaves do watch.
func (t *transaction) Unwatch(ctx context.Context) error {
	// TODO: Implementar unwatch real usando valkey-glide
	return fmt.Errorf("unwatch not implemented yet")
}

// Exec executa todos os comandos da transação.
func (t *transaction) Exec(ctx context.Context) ([]interface{}, error) {
	// Por enquanto, retorna erro indicando que transação não está implementada
	// TODO: Implementar execução real da transação usando valkey-glide
	results := make([]interface{}, len(t.commands))
	for i := range t.commands {
		results[i] = fmt.Errorf("transaction execution not implemented yet")
	}
	return results, fmt.Errorf("transaction execution not implemented yet")
}

// Discard descarta a transação.
func (t *transaction) Discard() error {
	t.commands = t.commands[:0]
	return nil
}
