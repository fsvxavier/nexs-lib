// Package hooks fornece interfaces e implementações para hooks de execução
// no módulo Valkey. Os hooks permitem interceptar operações para logging,
// métricas, autenticação, e outras funcionalidades transversais.
package hooks

import (
	"context"
	"time"
)

// ExecutionHook define um hook para interceptar execução de comandos.
type ExecutionHook interface {
	// BeforeExecution é chamado antes da execução de um comando.
	BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context

	// AfterExecution é chamado após a execução de um comando.
	AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration)
}

// ConnectionHook define um hook para interceptar eventos de conexão.
type ConnectionHook interface {
	// BeforeConnect é chamado antes de estabelecer uma conexão.
	BeforeConnect(ctx context.Context, network, addr string) context.Context

	// AfterConnect é chamado após estabelecer uma conexão.
	AfterConnect(ctx context.Context, network, addr string, err error, duration time.Duration)

	// BeforeDisconnect é chamado antes de fechar uma conexão.
	BeforeDisconnect(ctx context.Context, network, addr string) context.Context

	// AfterDisconnect é chamado após fechar uma conexão.
	AfterDisconnect(ctx context.Context, network, addr string, err error, duration time.Duration)
}

// PipelineHook define um hook para interceptar operações de pipeline.
type PipelineHook interface {
	// BeforePipelineExecution é chamado antes da execução de um pipeline.
	BeforePipelineExecution(ctx context.Context, commands []string) context.Context

	// AfterPipelineExecution é chamado após a execução de um pipeline.
	AfterPipelineExecution(ctx context.Context, commands []string, results []interface{}, err error, duration time.Duration)
}

// RetryHook define um hook para interceptar tentativas de retry.
type RetryHook interface {
	// BeforeRetry é chamado antes de uma tentativa de retry.
	BeforeRetry(ctx context.Context, attempt int, err error) context.Context

	// AfterRetry é chamado após uma tentativa de retry.
	AfterRetry(ctx context.Context, attempt int, success bool, err error)
}

// CompositeHook permite combinar múltiplos hooks em um único hook.
type CompositeHook struct {
	executionHooks  []ExecutionHook
	connectionHooks []ConnectionHook
	pipelineHooks   []PipelineHook
	retryHooks      []RetryHook
}

// NewCompositeHook cria um novo CompositeHook.
func NewCompositeHook() *CompositeHook {
	return &CompositeHook{
		executionHooks:  make([]ExecutionHook, 0),
		connectionHooks: make([]ConnectionHook, 0),
		pipelineHooks:   make([]PipelineHook, 0),
		retryHooks:      make([]RetryHook, 0),
	}
}

// AddExecutionHook adiciona um hook de execução.
func (c *CompositeHook) AddExecutionHook(hook ExecutionHook) {
	c.executionHooks = append(c.executionHooks, hook)
}

// AddConnectionHook adiciona um hook de conexão.
func (c *CompositeHook) AddConnectionHook(hook ConnectionHook) {
	c.connectionHooks = append(c.connectionHooks, hook)
}

// AddPipelineHook adiciona um hook de pipeline.
func (c *CompositeHook) AddPipelineHook(hook PipelineHook) {
	c.pipelineHooks = append(c.pipelineHooks, hook)
}

// AddRetryHook adiciona um hook de retry.
func (c *CompositeHook) AddRetryHook(hook RetryHook) {
	c.retryHooks = append(c.retryHooks, hook)
}

// BeforeExecution implementa ExecutionHook.
func (c *CompositeHook) BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context {
	for _, hook := range c.executionHooks {
		ctx = hook.BeforeExecution(ctx, cmd, args)
	}
	return ctx
}

// AfterExecution implementa ExecutionHook.
func (c *CompositeHook) AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration) {
	for _, hook := range c.executionHooks {
		hook.AfterExecution(ctx, cmd, args, result, err, duration)
	}
}

// BeforeConnect implementa ConnectionHook.
func (c *CompositeHook) BeforeConnect(ctx context.Context, network, addr string) context.Context {
	for _, hook := range c.connectionHooks {
		ctx = hook.BeforeConnect(ctx, network, addr)
	}
	return ctx
}

// AfterConnect implementa ConnectionHook.
func (c *CompositeHook) AfterConnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	for _, hook := range c.connectionHooks {
		hook.AfterConnect(ctx, network, addr, err, duration)
	}
}

// BeforeDisconnect implementa ConnectionHook.
func (c *CompositeHook) BeforeDisconnect(ctx context.Context, network, addr string) context.Context {
	for _, hook := range c.connectionHooks {
		ctx = hook.BeforeDisconnect(ctx, network, addr)
	}
	return ctx
}

// AfterDisconnect implementa ConnectionHook.
func (c *CompositeHook) AfterDisconnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	for _, hook := range c.connectionHooks {
		hook.AfterDisconnect(ctx, network, addr, err, duration)
	}
}

// BeforePipelineExecution implementa PipelineHook.
func (c *CompositeHook) BeforePipelineExecution(ctx context.Context, commands []string) context.Context {
	for _, hook := range c.pipelineHooks {
		ctx = hook.BeforePipelineExecution(ctx, commands)
	}
	return ctx
}

// AfterPipelineExecution implementa PipelineHook.
func (c *CompositeHook) AfterPipelineExecution(ctx context.Context, commands []string, results []interface{}, err error, duration time.Duration) {
	for _, hook := range c.pipelineHooks {
		hook.AfterPipelineExecution(ctx, commands, results, err, duration)
	}
}

// BeforeRetry implementa RetryHook.
func (c *CompositeHook) BeforeRetry(ctx context.Context, attempt int, err error) context.Context {
	for _, hook := range c.retryHooks {
		ctx = hook.BeforeRetry(ctx, attempt, err)
	}
	return ctx
}

// AfterRetry implementa RetryHook.
func (c *CompositeHook) AfterRetry(ctx context.Context, attempt int, success bool, err error) {
	for _, hook := range c.retryHooks {
		hook.AfterRetry(ctx, attempt, success, err)
	}
}
