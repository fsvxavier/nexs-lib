package advanced

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// ConditionalHook representa um hook que executa sob condições específicas
type ConditionalHook struct {
	condition func(interfaces.DomainErrorInterface) bool
	hook      interfaces.ErrorHookFunc
	name      string
	priority  int
}

// ConditionalHookManager gerencia hooks condicionais
type ConditionalHookManager struct {
	hooks []ConditionalHook
	mu    sync.RWMutex
}

// NewConditionalHookManager cria um novo gerenciador de hooks condicionais
func NewConditionalHookManager() *ConditionalHookManager {
	return &ConditionalHookManager{
		hooks: make([]ConditionalHook, 0),
	}
}

// RegisterConditionalErrorHook registra um hook condicional
func (chm *ConditionalHookManager) RegisterConditionalErrorHook(
	name string,
	priority int,
	condition func(interfaces.DomainErrorInterface) bool,
	hook interfaces.ErrorHookFunc,
) {
	chm.mu.Lock()
	defer chm.mu.Unlock()

	conditionalHook := ConditionalHook{
		condition: condition,
		hook:      hook,
		name:      name,
		priority:  priority,
	}

	// Inserir mantendo ordem de prioridade (maior prioridade primeiro)
	inserted := false
	for i, h := range chm.hooks {
		if priority > h.priority {
			chm.hooks = append(chm.hooks[:i], append([]ConditionalHook{conditionalHook}, chm.hooks[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		chm.hooks = append(chm.hooks, conditionalHook)
	}
}

// ExecuteConditionalHooks executa todos os hooks que satisfazem as condições
func (chm *ConditionalHookManager) ExecuteConditionalHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	chm.mu.RLock()
	defer chm.mu.RUnlock()

	for _, hook := range chm.hooks {
		if hook.condition(err) {
			if hookErr := hook.hook(ctx, err); hookErr != nil {
				// Log o erro do hook mas continua executando outros hooks
				// Em um ambiente real, isso seria logado apropriadamente
				continue
			}
		}
	}

	return nil
}

// RemoveConditionalHook remove um hook por nome
func (chm *ConditionalHookManager) RemoveConditionalHook(name string) bool {
	chm.mu.Lock()
	defer chm.mu.Unlock()

	for i, hook := range chm.hooks {
		if hook.name == name {
			chm.hooks = append(chm.hooks[:i], chm.hooks[i+1:]...)
			return true
		}
	}

	return false
}

// GetRegisteredHooks retorna lista de hooks registrados
func (chm *ConditionalHookManager) GetRegisteredHooks() []string {
	chm.mu.RLock()
	defer chm.mu.RUnlock()

	names := make([]string, len(chm.hooks))
	for i, hook := range chm.hooks {
		names[i] = hook.name
	}

	return names
}

// Instância global do gerenciador
var globalConditionalHookManager = NewConditionalHookManager()

// RegisterConditionalErrorHook registra um hook condicional globalmente
func RegisterConditionalErrorHook(
	name string,
	priority int,
	condition func(interfaces.DomainErrorInterface) bool,
	hook interfaces.ErrorHookFunc,
) {
	globalConditionalHookManager.RegisterConditionalErrorHook(name, priority, condition, hook)

	// Também registra no sistema de hooks global para integração
	hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
		return globalConditionalHookManager.ExecuteConditionalHooks(ctx, err)
	})
}

// Funções utilitárias para condições comuns

// ErrorTypeCondition cria condição baseada no tipo de erro
func ErrorTypeCondition(errorType interfaces.ErrorType) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		return err.Type() == errorType
	}
}

// ErrorCodeCondition cria condição baseada no código de erro
func ErrorCodeCondition(code string) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		return err.Code() == code
	}
}

// HTTPStatusCondition cria condição baseada no status HTTP
func HTTPStatusCondition(status int) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		return err.HTTPStatus() == status
	}
}

// MetadataCondition cria condição baseada em metadados
func MetadataCondition(key string, value interface{}) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		if metadata := err.Metadata(); metadata != nil {
			if val, exists := metadata[key]; exists {
				return val == value
			}
		}
		return false
	}
}

// CombinedCondition combina múltiplas condições com AND lógico
func CombinedCondition(conditions ...func(interfaces.DomainErrorInterface) bool) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		for _, condition := range conditions {
			if !condition(err) {
				return false
			}
		}
		return true
	}
}

// OrCondition combina múltiplas condições com OR lógico
func OrCondition(conditions ...func(interfaces.DomainErrorInterface) bool) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		for _, condition := range conditions {
			if condition(err) {
				return true
			}
		}
		return false
	}
}

// NotCondition inverte uma condição
func NotCondition(condition func(interfaces.DomainErrorInterface) bool) func(interfaces.DomainErrorInterface) bool {
	return func(err interfaces.DomainErrorInterface) bool {
		return !condition(err)
	}
}

// Exemplos de hooks pré-configurados

// SecurityErrorHook hook para erros de segurança
func SecurityErrorHook(ctx context.Context, err interfaces.DomainErrorInterface) error {
	// Implementação para notificar equipe de segurança
	// Em um ambiente real, isso enviaria alertas, logs especiais, etc.
	return nil
}

// CriticalErrorHook hook para erros críticos
func CriticalErrorHook(ctx context.Context, err interfaces.DomainErrorInterface) error {
	// Implementação para erros críticos
	// Poderia disparar pagers, notificações imediatas, etc.
	return nil
}

// BusinessErrorHook hook para erros de negócio
func BusinessErrorHook(ctx context.Context, err interfaces.DomainErrorInterface) error {
	// Implementação para erros de negócio
	// Poderia registrar em analytics, dashboards de negócio, etc.
	return nil
}

// HighVolumeErrorHook hook para erros de alto volume
func HighVolumeErrorHook(ctx context.Context, err interfaces.DomainErrorInterface) error {
	// Implementação para erros frequentes
	// Poderia implementar throttling, rate limiting, etc.
	return nil
}
