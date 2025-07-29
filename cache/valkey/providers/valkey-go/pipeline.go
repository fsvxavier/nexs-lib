package valkeygo

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
	"github.com/valkey-io/valkey-go"
)

// Pipeline implementa interfaces.IPipeline.
type Pipeline struct {
	client valkey.Client
	cmds   []valkey.Completed
}

// Get implementa interfaces.IPipeline.Get.
func (p *Pipeline) Get(key string) interfaces.ICommand {
	cmd := p.client.B().Get().Key(key).Build()
	p.cmds = append(p.cmds, cmd)
	return &Command{index: len(p.cmds) - 1}
}

// Set implementa interfaces.IPipeline.Set.
func (p *Pipeline) Set(key string, value interface{}, expiration time.Duration) interfaces.ICommand {
	cmd := p.client.B().Set().Key(key).Value(fmt.Sprintf("%v", value))

	var builtCmd valkey.Completed
	if expiration > 0 {
		builtCmd = cmd.Ex(expiration).Build()
	} else {
		builtCmd = cmd.Build()
	}

	p.cmds = append(p.cmds, builtCmd)
	return &Command{index: len(p.cmds) - 1}
}

// Del implementa interfaces.IPipeline.Del.
func (p *Pipeline) Del(keys ...string) interfaces.ICommand {
	cmd := p.client.B().Del().Key(keys...).Build()
	p.cmds = append(p.cmds, cmd)
	return &Command{index: len(p.cmds) - 1}
}

// HGet implementa interfaces.IPipeline.HGet.
func (p *Pipeline) HGet(key, field string) interfaces.ICommand {
	cmd := p.client.B().Hget().Key(key).Field(field).Build()
	p.cmds = append(p.cmds, cmd)
	return &Command{index: len(p.cmds) - 1}
}

// HSet implementa interfaces.IPipeline.HSet.
func (p *Pipeline) HSet(key string, values ...interface{}) interfaces.ICommand {
	cmd := p.client.B().Hset().Key(key).FieldValue()
	for i := 0; i < len(values); i += 2 {
		field := fmt.Sprintf("%v", values[i])
		value := fmt.Sprintf("%v", values[i+1])
		cmd = cmd.FieldValue(field, value)
	}
	p.cmds = append(p.cmds, cmd.Build())
	return &Command{index: len(p.cmds) - 1}
}

// Exec implementa interfaces.IPipeline.Exec.
func (p *Pipeline) Exec(ctx context.Context) ([]interface{}, error) {
	if len(p.cmds) == 0 {
		return []interface{}{}, nil
	}

	results := p.client.DoMulti(ctx, p.cmds...)

	responses := make([]interface{}, len(results))
	for i, result := range results {
		if result.Error() != nil {
			responses[i] = result.Error()
		} else {
			// Tentar converter para string como padrão
			if str, err := result.ToString(); err == nil {
				responses[i] = str
			} else {
				responses[i] = result
			}
		}
	}

	return responses, nil
}

// Discard implementa interfaces.IPipeline.Discard.
func (p *Pipeline) Discard() error {
	p.cmds = p.cmds[:0]
	return nil
}
