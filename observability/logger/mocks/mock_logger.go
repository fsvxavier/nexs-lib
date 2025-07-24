package mocks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/observability/logger"
)

// MockLogger implementa a interface Logger para testes
type MockLogger struct {
	mu       sync.RWMutex
	logs     []LogEntry
	level    logger.Level
	fields   []logger.Field
	isClosed bool
}

// LogEntry representa uma entrada de log capturada
type LogEntry struct {
	Level   logger.Level
	Message string
	Fields  []logger.Field
	Context context.Context
	Code    string
}

// NewMockLogger cria um novo mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs:  make([]LogEntry, 0),
		level: logger.InfoLevel,
	}
}

// Debug implementa Logger
func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.DebugLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.DebugLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
		})
	}
}

// Info implementa Logger
func (m *MockLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.InfoLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.InfoLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
		})
	}
}

// Warn implementa Logger
func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.WarnLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.WarnLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
		})
	}
}

// Error implementa Logger
func (m *MockLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.ErrorLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.ErrorLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
		})
	}
}

// Fatal implementa Logger
func (m *MockLogger) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logs = append(m.logs, LogEntry{
		Level:   logger.FatalLevel,
		Message: msg,
		Fields:  append(m.fields, fields...),
		Context: ctx,
	})
}

// Panic implementa Logger
func (m *MockLogger) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logs = append(m.logs, LogEntry{
		Level:   logger.PanicLevel,
		Message: msg,
		Fields:  append(m.fields, fields...),
		Context: ctx,
	})
}

// Debugf implementa Logger
func (m *MockLogger) Debugf(ctx context.Context, format string, args ...any) {
	m.Debug(ctx, format, logger.Any("args", args))
}

// Infof implementa Logger
func (m *MockLogger) Infof(ctx context.Context, format string, args ...any) {
	m.Info(ctx, format, logger.Any("args", args))
}

// Warnf implementa Logger
func (m *MockLogger) Warnf(ctx context.Context, format string, args ...any) {
	m.Warn(ctx, format, logger.Any("args", args))
}

// Errorf implementa Logger
func (m *MockLogger) Errorf(ctx context.Context, format string, args ...any) {
	m.Error(ctx, format, logger.Any("args", args))
}

// Fatalf implementa Logger
func (m *MockLogger) Fatalf(ctx context.Context, format string, args ...any) {
	m.Fatal(ctx, format, logger.Any("args", args))
}

// Panicf implementa Logger
func (m *MockLogger) Panicf(ctx context.Context, format string, args ...any) {
	m.Panic(ctx, format, logger.Any("args", args))
}

// DebugWithCode implementa Logger
func (m *MockLogger) DebugWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.DebugLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.DebugLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
			Code:    code,
		})
	}
}

// InfoWithCode implementa Logger
func (m *MockLogger) InfoWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.InfoLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.InfoLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
			Code:    code,
		})
	}
}

// WarnWithCode implementa Logger
func (m *MockLogger) WarnWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.WarnLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.WarnLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
			Code:    code,
		})
	}
}

// ErrorWithCode implementa Logger
func (m *MockLogger) ErrorWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.level <= logger.ErrorLevel {
		m.logs = append(m.logs, LogEntry{
			Level:   logger.ErrorLevel,
			Message: msg,
			Fields:  append(m.fields, fields...),
			Context: ctx,
			Code:    code,
		})
	}
}

// WithFields implementa Logger
func (m *MockLogger) WithFields(fields ...logger.Field) logger.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newMock := &MockLogger{
		logs:     m.logs,
		level:    m.level,
		fields:   append(m.fields, fields...),
		isClosed: m.isClosed,
	}

	return newMock
}

// WithContext implementa Logger
func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	// Para o mock, simplesmente retorna o próprio logger
	return m
}

// SetLevel implementa Logger
func (m *MockLogger) SetLevel(level logger.Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// GetLevel implementa Logger
func (m *MockLogger) GetLevel() logger.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

// Clone implementa Logger
func (m *MockLogger) Clone() logger.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newMock := &MockLogger{
		logs:     make([]LogEntry, len(m.logs)),
		level:    m.level,
		fields:   make([]logger.Field, len(m.fields)),
		isClosed: m.isClosed,
	}

	copy(newMock.logs, m.logs)
	copy(newMock.fields, m.fields)

	return newMock
}

// Close implementa Logger
func (m *MockLogger) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isClosed = true
	return nil
}

// Métodos específicos do mock para testes

// GetLogs retorna todos os logs capturados
func (m *MockLogger) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]LogEntry, len(m.logs))
	copy(logs, m.logs)
	return logs
}

// GetLogCount retorna o número de logs capturados
func (m *MockLogger) GetLogCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.logs)
}

// GetLogsByLevel retorna logs filtrados por nível
func (m *MockLogger) GetLogsByLevel(level logger.Level) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []LogEntry
	for _, log := range m.logs {
		if log.Level == level {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// GetLogsByMessage retorna logs filtrados por mensagem
func (m *MockLogger) GetLogsByMessage(message string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []LogEntry
	for _, log := range m.logs {
		if log.Message == message {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// GetLogsByCode retorna logs filtrados por código
func (m *MockLogger) GetLogsByCode(code string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []LogEntry
	for _, log := range m.logs {
		if log.Code == code {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// Reset limpa todos os logs capturados
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = make([]LogEntry, 0)
}

// IsClosed retorna se o logger foi fechado
func (m *MockLogger) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isClosed
}

// LastLog retorna o último log capturado
func (m *MockLogger) LastLog() *LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.logs) == 0 {
		return nil
	}

	return &m.logs[len(m.logs)-1]
}

// HasField verifica se algum log contém um campo específico
func (m *MockLogger) HasField(key string, value any) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, log := range m.logs {
		for _, field := range log.Fields {
			if field.Key == key && field.Value == value {
				return true
			}
		}
	}
	return false
}

// Certifica que MockLogger implementa a interface Logger
var _ logger.Logger = (*MockLogger)(nil)
