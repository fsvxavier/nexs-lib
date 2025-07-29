package valkeygo

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
	"github.com/valkey-io/valkey-go"
)

// Transaction implementa interfaces.ITransaction.
type Transaction struct {
	client valkey.Client
	cmds   []valkey.Completed
}

// Get implementa interfaces.ITransaction.Get.
func (t *Transaction) Get(key string) interfaces.ICommand {
	cmd := t.client.B().Get().Key(key).Build()
	t.cmds = append(t.cmds, cmd)
	return &Command{index: len(t.cmds) - 1}
}

// Set implementa interfaces.ITransaction.Set.
func (t *Transaction) Set(key string, value interface{}, expiration time.Duration) interfaces.ICommand {
	cmd := t.client.B().Set().Key(key).Value(fmt.Sprintf("%v", value))

	var builtCmd valkey.Completed
	if expiration > 0 {
		builtCmd = cmd.Ex(expiration).Build()
	} else {
		builtCmd = cmd.Build()
	}

	t.cmds = append(t.cmds, builtCmd)
	return &Command{index: len(t.cmds) - 1}
}

// Del implementa interfaces.ITransaction.Del.
func (t *Transaction) Del(keys ...string) interfaces.ICommand {
	cmd := t.client.B().Del().Key(keys...).Build()
	t.cmds = append(t.cmds, cmd)
	return &Command{index: len(t.cmds) - 1}
}

// HGet implementa interfaces.ITransaction.HGet.
func (t *Transaction) HGet(key, field string) interfaces.ICommand {
	cmd := t.client.B().Hget().Key(key).Field(field).Build()
	t.cmds = append(t.cmds, cmd)
	return &Command{index: len(t.cmds) - 1}
}

// HSet implementa interfaces.ITransaction.HSet.
func (t *Transaction) HSet(key string, values ...interface{}) interfaces.ICommand {
	cmd := t.client.B().Hset().Key(key).FieldValue()
	for i := 0; i < len(values); i += 2 {
		field := fmt.Sprintf("%v", values[i])
		value := fmt.Sprintf("%v", values[i+1])
		cmd = cmd.FieldValue(field, value)
	}
	t.cmds = append(t.cmds, cmd.Build())
	return &Command{index: len(t.cmds) - 1}
}

// Watch implementa interfaces.ITransaction.Watch.
func (t *Transaction) Watch(ctx context.Context, keys ...string) error {
	cmd := t.client.B().Watch().Key(keys...).Build()
	result := t.client.Do(ctx, cmd)
	return result.Error()
}

// Unwatch implementa interfaces.ITransaction.Unwatch.
func (t *Transaction) Unwatch(ctx context.Context) error {
	cmd := t.client.B().Unwatch().Build()
	result := t.client.Do(ctx, cmd)
	return result.Error()
}

// Exec implementa interfaces.ITransaction.Exec.
func (t *Transaction) Exec(ctx context.Context) ([]interface{}, error) {
	if len(t.cmds) == 0 {
		return []interface{}{}, nil
	}

	// Começar transação
	multiCmd := t.client.B().Multi().Build()
	if err := t.client.Do(ctx, multiCmd).Error(); err != nil {
		return nil, err
	}

	// Adicionar comandos
	for _, cmd := range t.cmds {
		if err := t.client.Do(ctx, cmd).Error(); err != nil {
			// Descartar transação em caso de erro
			discardCmd := t.client.B().Discard().Build()
			t.client.Do(ctx, discardCmd)
			return nil, err
		}
	}

	// Executar transação
	execCmd := t.client.B().Exec().Build()
	result := t.client.Do(ctx, execCmd)

	if result.Error() != nil {
		return nil, result.Error()
	}

	// Implementação básica - seria melhorada para parsear resultados
	return []interface{}{}, nil
}

// Discard implementa interfaces.ITransaction.Discard.
func (t *Transaction) Discard() error {
	t.cmds = t.cmds[:0]
	return nil
}
