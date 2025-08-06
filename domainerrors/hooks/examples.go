package hooks

import (
	"fmt"
	"log"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// LoggingHook hook básico de logging
type LoggingHook struct {
	name     string
	hookType interfaces.HookType
	priority int
	enabled  bool
}

// NewLoggingHook cria um novo hook de logging
func NewLoggingHook(hookType interfaces.HookType) *LoggingHook {
	return &LoggingHook{
		name:     fmt.Sprintf("logging_hook_%s", hookType),
		hookType: hookType,
		priority: 100,
		enabled:  true,
	}
}

// Name retorna o nome do hook
func (h *LoggingHook) Name() string {
	return h.name
}

// Execute executa o hook
func (h *LoggingHook) Execute(ctx *interfaces.HookContext) error {
	log.Printf("[HOOK] %s: %s (Operation: %s)",
		h.hookType, ctx.Error.Message, ctx.Operation)
	return nil
}

// Type retorna o tipo do hook
func (h *LoggingHook) Type() interfaces.HookType {
	return h.hookType
}

// Priority retorna a prioridade de execução
func (h *LoggingHook) Priority() int {
	return h.priority
}

// Enabled indica se o hook está habilitado
func (h *LoggingHook) Enabled() bool {
	return h.enabled
}

// SetEnabled habilita/desabilita o hook
func (h *LoggingHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// MetricsHook hook básico de métricas
type MetricsHook struct {
	name       string
	hookType   interfaces.HookType
	priority   int
	enabled    bool
	executions int
	lastExecAt time.Time
}

// NewMetricsHook cria um novo hook de métricas
func NewMetricsHook(hookType interfaces.HookType) *MetricsHook {
	return &MetricsHook{
		name:     fmt.Sprintf("metrics_hook_%s", hookType),
		hookType: hookType,
		priority: 200,
		enabled:  true,
	}
}

// Name retorna o nome do hook
func (h *MetricsHook) Name() string {
	return h.name
}

// Execute executa o hook
func (h *MetricsHook) Execute(ctx *interfaces.HookContext) error {
	h.executions++
	h.lastExecAt = time.Now()

	// Adiciona métricas aos metadados do erro
	if ctx.Error.Metadata == nil {
		ctx.Error.Metadata = make(map[string]interface{})
	}

	ctx.Error.Metadata[fmt.Sprintf("hook_%s_executions", h.hookType)] = h.executions
	ctx.Error.Metadata[fmt.Sprintf("hook_%s_last_exec", h.hookType)] = h.lastExecAt

	return nil
}

// Type retorna o tipo do hook
func (h *MetricsHook) Type() interfaces.HookType {
	return h.hookType
}

// Priority retorna a prioridade de execução
func (h *MetricsHook) Priority() int {
	return h.priority
}

// Enabled indica se o hook está habilitado
func (h *MetricsHook) Enabled() bool {
	return h.enabled
}

// SetEnabled habilita/desabilita o hook
func (h *MetricsHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// GetStats retorna estatísticas do hook
func (h *MetricsHook) GetStats() (int, time.Time) {
	return h.executions, h.lastExecAt
}

// ValidationHook hook para validação de erros
type ValidationHook struct {
	name     string
	hookType interfaces.HookType
	priority int
	enabled  bool
}

// NewValidationHook cria um novo hook de validação
func NewValidationHook() *ValidationHook {
	return &ValidationHook{
		name:     "validation_hook",
		hookType: interfaces.HookTypeBeforeError,
		priority: 10, // Alta prioridade
		enabled:  true,
	}
}

// Name retorna o nome do hook
func (h *ValidationHook) Name() string {
	return h.name
}

// Execute executa o hook
func (h *ValidationHook) Execute(ctx *interfaces.HookContext) error {
	if ctx.Error == nil {
		return fmt.Errorf("error cannot be nil")
	}

	if ctx.Error.Code == "" {
		return fmt.Errorf("error code cannot be empty")
	}

	if ctx.Error.Message == "" {
		return fmt.Errorf("error message cannot be empty")
	}

	log.Printf("[VALIDATION] Error validated successfully: %s", ctx.Error.Code)
	return nil
}

// Type retorna o tipo do hook
func (h *ValidationHook) Type() interfaces.HookType {
	return h.hookType
}

// Priority retorna a prioridade de execução
func (h *ValidationHook) Priority() int {
	return h.priority
}

// Enabled indica se o hook está habilitado
func (h *ValidationHook) Enabled() bool {
	return h.enabled
}

// SetEnabled habilita/desabilita o hook
func (h *ValidationHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// AuditHook hook para auditoria de erros
type AuditHook struct {
	name     string
	hookType interfaces.HookType
	priority int
	enabled  bool
	auditLog []AuditEntry
}

// AuditEntry entrada de auditoria
type AuditEntry struct {
	Timestamp time.Time
	ErrorCode string
	ErrorType interfaces.ErrorType
	Operation string
	UserID    string
	TraceID   string
}

// NewAuditHook cria um novo hook de auditoria
func NewAuditHook() *AuditHook {
	return &AuditHook{
		name:     "audit_hook",
		hookType: interfaces.HookTypeAfterError,
		priority: 300,
		enabled:  true,
		auditLog: make([]AuditEntry, 0),
	}
}

// Name retorna o nome do hook
func (h *AuditHook) Name() string {
	return h.name
}

// Execute executa o hook
func (h *AuditHook) Execute(ctx *interfaces.HookContext) error {
	entry := AuditEntry{
		Timestamp: time.Now(),
		ErrorCode: ctx.Error.Code,
		ErrorType: ctx.Error.Type,
		Operation: ctx.Operation,
		UserID:    ctx.UserID,
		TraceID:   ctx.TraceID,
	}

	h.auditLog = append(h.auditLog, entry)

	log.Printf("[AUDIT] Error logged: %s (User: %s, Trace: %s)",
		ctx.Error.Code, ctx.UserID, ctx.TraceID)

	return nil
}

// Type retorna o tipo do hook
func (h *AuditHook) Type() interfaces.HookType {
	return h.hookType
}

// Priority retorna a prioridade de execução
func (h *AuditHook) Priority() int {
	return h.priority
}

// Enabled indica se o hook está habilitado
func (h *AuditHook) Enabled() bool {
	return h.enabled
}

// SetEnabled habilita/desabilita o hook
func (h *AuditHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// GetAuditLog retorna o log de auditoria
func (h *AuditHook) GetAuditLog() []AuditEntry {
	log := make([]AuditEntry, len(h.auditLog))
	copy(log, h.auditLog)
	return log
}

// PrintAuditLog imprime o log de auditoria
func (h *AuditHook) PrintAuditLog() {
	fmt.Println("=== Audit Log ===")
	for _, entry := range h.auditLog {
		fmt.Printf("[%s] %s (%s) - User: %s, Trace: %s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.ErrorCode,
			entry.ErrorType,
			entry.UserID,
			entry.TraceID)
	}
	fmt.Println("==================")
}
