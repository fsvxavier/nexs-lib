package middlewares

import (
	"errors"
	"sort"
	"sync"

	interfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

var (
	// ErrMiddlewareAlreadyExists é retornado quando tentamos adicionar um middleware que já existe
	ErrMiddlewareAlreadyExists = errors.New("middleware already exists")
	// ErrMiddlewareNotFound é retornado quando um middleware não é encontrado
	ErrMiddlewareNotFound = errors.New("middleware not found")
	// ErrInvalidMiddleware é retornado quando um middleware é inválido
	ErrInvalidMiddleware = errors.New("invalid middleware")
	// ErrMiddlewareChainEmpty é retornado quando a cadeia está vazia
	ErrMiddlewareChainEmpty = errors.New("middleware chain is empty")
)

// DefaultMiddlewareChain implementação padrão da cadeia de middlewares
type DefaultMiddlewareChain struct {
	middlewares []interfaces.Middleware
	mutex       sync.RWMutex
}

// NewDefaultMiddlewareChain cria uma nova instância da cadeia padrão
func NewDefaultMiddlewareChain() *DefaultMiddlewareChain {
	return &DefaultMiddlewareChain{
		middlewares: make([]interfaces.Middleware, 0),
	}
}

// Use adiciona um middleware à cadeia
func (c *DefaultMiddlewareChain) Use(middleware interfaces.Middleware) interfaces.MiddlewareChain {
	if middleware == nil {
		return c
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Verifica se o middleware já existe
	for _, existing := range c.middlewares {
		if existing.Name() == middleware.Name() {
			return c // Ignora silenciosamente se já existe
		}
	}

	// Adiciona o middleware
	c.middlewares = append(c.middlewares, middleware)

	// Ordena por prioridade (0 = maior prioridade)
	sort.Slice(c.middlewares, func(i, j int) bool {
		return c.middlewares[i].Priority() < c.middlewares[j].Priority()
	})

	return c
}

// Execute executa toda a cadeia de middlewares
func (c *DefaultMiddlewareChain) Execute(ctx *interfaces.MiddlewareContext) error {
	c.mutex.RLock()
	middlewares := c.getEnabledMiddlewares()
	c.mutex.RUnlock()

	if len(middlewares) == 0 {
		return nil // Não há middlewares para executar
	}

	// Cria a cadeia de execução
	index := 0
	var next interfaces.NextFunction
	next = func(ctx *interfaces.MiddlewareContext) error {
		if index >= len(middlewares) {
			return nil // Fim da cadeia
		}

		middleware := middlewares[index]
		index++

		return middleware.Handle(ctx, next)
	}

	return next(ctx)
}

// Size retorna o número de middlewares na cadeia
func (c *DefaultMiddlewareChain) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.middlewares)
}

// Clear limpa a cadeia
func (c *DefaultMiddlewareChain) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.middlewares = make([]interfaces.Middleware, 0)
}

// Remove remove um middleware pelo nome
func (c *DefaultMiddlewareChain) Remove(name string) bool {
	if name == "" {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, middleware := range c.middlewares {
		if middleware.Name() == name {
			// Remove o middleware
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			return true
		}
	}

	return false
}

// GetMiddlewares retorna todos os middlewares ordenados por prioridade
func (c *DefaultMiddlewareChain) GetMiddlewares() []interfaces.Middleware {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make([]interfaces.Middleware, len(c.middlewares))
	copy(result, c.middlewares)
	return result
}

// getEnabledMiddlewares retorna apenas middlewares habilitados (método interno)
func (c *DefaultMiddlewareChain) getEnabledMiddlewares() []interfaces.Middleware {
	enabledMiddlewares := make([]interfaces.Middleware, 0, len(c.middlewares))
	for _, middleware := range c.middlewares {
		if middleware.Enabled() {
			enabledMiddlewares = append(enabledMiddlewares, middleware)
		}
	}
	return enabledMiddlewares
}

// GlobalMiddlewareChain instância global da cadeia de middlewares
var GlobalMiddlewareChain interfaces.MiddlewareChain = NewDefaultMiddlewareChain()

// UseMiddleware adiciona um middleware à cadeia global
func UseMiddleware(middleware interfaces.Middleware) interfaces.MiddlewareChain {
	return GlobalMiddlewareChain.Use(middleware)
}

// ExecuteMiddlewares executa toda a cadeia global de middlewares
func ExecuteMiddlewares(ctx *interfaces.MiddlewareContext) error {
	return GlobalMiddlewareChain.Execute(ctx)
}

// ClearMiddlewares limpa todos os middlewares da cadeia global
func ClearMiddlewares() {
	GlobalMiddlewareChain.Clear()
}

// RemoveMiddleware remove um middleware da cadeia global
func RemoveMiddleware(name string) bool {
	return GlobalMiddlewareChain.Remove(name)
}

// GetMiddlewares retorna todos os middlewares da cadeia global
func GetMiddlewares() []interfaces.Middleware {
	return GlobalMiddlewareChain.GetMiddlewares()
}

// SizeMiddlewares retorna o número de middlewares na cadeia global
func SizeMiddlewares() int {
	return GlobalMiddlewareChain.Size()
}
