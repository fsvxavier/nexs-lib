package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// HookManager implementa a interface interfaces.HookManager
type HookManager struct {
	startHooks []interfaces.StartHookFunc
	stopHooks  []interfaces.StopHookFunc
	errorHooks []interfaces.ErrorHookFunc
	i18nHooks  []interfaces.I18nHookFunc
	mu         sync.RWMutex
}

// NewHookManager cria um novo gerenciador de hooks completo
func NewHookManager() *HookManager {
	return &HookManager{
		startHooks: make([]interfaces.StartHookFunc, 0),
		stopHooks:  make([]interfaces.StopHookFunc, 0),
		errorHooks: make([]interfaces.ErrorHookFunc, 0),
		i18nHooks:  make([]interfaces.I18nHookFunc, 0),
	}
}

// RegisterStartHook registra um hook de start
func (m *HookManager) RegisterStartHook(hook interfaces.StartHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.startHooks = append(m.startHooks, hook)
}

// RegisterStopHook registra um hook de stop
func (m *HookManager) RegisterStopHook(hook interfaces.StopHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopHooks = append(m.stopHooks, hook)
}

// RegisterErrorHook registra um hook de erro
func (m *HookManager) RegisterErrorHook(hook interfaces.ErrorHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorHooks = append(m.errorHooks, hook)
}

// RegisterI18nHook registra um hook de i18n
func (m *HookManager) RegisterI18nHook(hook interfaces.I18nHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.i18nHooks = append(m.i18nHooks, hook)
}

// ExecuteStartHooks executa todos os hooks de start registrados
func (m *HookManager) ExecuteStartHooks(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.startHooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteStopHooks executa todos os hooks de stop registrados
func (m *HookManager) ExecuteStopHooks(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.stopHooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteErrorHooks executa todos os hooks de erro registrados
func (m *HookManager) ExecuteErrorHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.errorHooks {
		if hookErr := hook(ctx, err); hookErr != nil {
			return hookErr
		}
	}

	return nil
}

// ExecuteI18nHooks executa todos os hooks de i18n registrados
func (m *HookManager) ExecuteI18nHooks(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.i18nHooks {
		if hookErr := hook(ctx, err, locale); hookErr != nil {
			return hookErr
		}
	}

	return nil
}

// Clear limpa todos os hooks registrados
func (m *HookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.startHooks = make([]interfaces.StartHookFunc, 0)
	m.stopHooks = make([]interfaces.StopHookFunc, 0)
	m.errorHooks = make([]interfaces.ErrorHookFunc, 0)
	m.i18nHooks = make([]interfaces.I18nHookFunc, 0)
}

// GetCounts retorna o número de hooks registrados por tipo
func (m *HookManager) GetCounts() (start, stop, error, i18n int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.startHooks), len(m.stopHooks), len(m.errorHooks), len(m.i18nHooks)
}

// Instância global do HookManager
var GlobalHookManager = NewHookManager()

// Funções globais de conveniência
func RegisterGlobalStartHook(hook interfaces.StartHookFunc) {
	GlobalHookManager.RegisterStartHook(hook)
}

func RegisterGlobalStopHook(hook interfaces.StopHookFunc) {
	GlobalHookManager.RegisterStopHook(hook)
}

func RegisterGlobalErrorHook(hook interfaces.ErrorHookFunc) {
	GlobalHookManager.RegisterErrorHook(hook)
}

func RegisterGlobalI18nHook(hook interfaces.I18nHookFunc) {
	GlobalHookManager.RegisterI18nHook(hook)
}

func ExecuteGlobalStartHooks(ctx context.Context) error {
	return GlobalHookManager.ExecuteStartHooks(ctx)
}

func ExecuteGlobalStopHooks(ctx context.Context) error {
	return GlobalHookManager.ExecuteStopHooks(ctx)
}

func ExecuteGlobalErrorHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	return GlobalHookManager.ExecuteErrorHooks(ctx, err)
}

func ExecuteGlobalI18nHooks(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	return GlobalHookManager.ExecuteI18nHooks(ctx, err, locale)
}

func GetGlobalHookCounts() (start, stop, error, i18n int) {
	return GlobalHookManager.GetCounts()
}
