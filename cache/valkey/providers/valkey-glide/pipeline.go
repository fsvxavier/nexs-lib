package valkeyglide

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// pipeline implementa interfaces.IPipeline para valkey-glide.
type pipeline struct {
	commands []interfaces.ICommand
}

// newPipeline cria um novo pipeline.
func newPipeline(client interface{}) interfaces.IPipeline {
	return &pipeline{
		commands: make([]interfaces.ICommand, 0),
	}
}

// Get adiciona comando GET ao pipeline.
func (p *pipeline) Get(key string) interfaces.ICommand {
	cmd := &command{
		result: nil,
		err:    fmt.Errorf("comando GET não executado ainda"),
	}
	p.commands = append(p.commands, cmd)
	return cmd
}

// Set adiciona comando SET ao pipeline.
func (p *pipeline) Set(key string, value interface{}, expiration time.Duration) interfaces.ICommand {
	cmd := &command{
		result: nil,
		err:    fmt.Errorf("comando SET não executado ainda"),
	}
	p.commands = append(p.commands, cmd)
	return cmd
}

// Del adiciona comando DEL ao pipeline.
func (p *pipeline) Del(keys ...string) interfaces.ICommand {
	cmd := &command{
		result: nil,
		err:    fmt.Errorf("comando DEL não executado ainda"),
	}
	p.commands = append(p.commands, cmd)
	return cmd
}

// HGet adiciona comando HGET ao pipeline.
func (p *pipeline) HGet(key, field string) interfaces.ICommand {
	cmd := &command{
		result: nil,
		err:    fmt.Errorf("comando HGET não executado ainda"),
	}
	p.commands = append(p.commands, cmd)
	return cmd
}

// HSet adiciona comando HSET ao pipeline.
func (p *pipeline) HSet(key string, values ...interface{}) interfaces.ICommand {
	cmd := &command{
		result: nil,
		err:    fmt.Errorf("comando HSET não executado ainda"),
	}
	p.commands = append(p.commands, cmd)
	return cmd
}

// Exec executa todos os comandos do pipeline.
func (p *pipeline) Exec(ctx context.Context) ([]interface{}, error) {
	// Por enquanto, retorna erro indicando que pipeline não está implementado
	// TODO: Implementar execução real dos comandos usando valkey-glide
	return nil, fmt.Errorf("pipeline execution not implemented yet")
}

// Discard descarta o pipeline.
func (p *pipeline) Discard() error {
	p.commands = p.commands[:0]
	return nil
}
