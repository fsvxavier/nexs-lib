package logrus

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// DefaultConfig retorna a configuração padrão para o provider Logrus
func DefaultConfig() *interfaces.Config {
	return &interfaces.Config{
		Level:          interfaces.InfoLevel,
		Format:         interfaces.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "",
		ServiceVersion: "",
		Environment:    "",
		AddSource:      false,
		AddStacktrace:  false,
		Fields:         make(map[string]any),
		SamplingConfig: nil,
		BufferConfig:   nil,
	}
}

// NewWithConfig cria um provider Logrus com configuração personalizada
func NewWithConfig(config *interfaces.Config) (*Provider, error) {
	provider := NewProvider()

	if config != nil {
		if err := provider.Configure(config); err != nil {
			return nil, err
		}
	}

	return provider, nil
}

// NewWithWriter cria um provider Logrus com um writer específico
func NewWithWriter(writer io.Writer) *Provider {
	config := DefaultConfig()
	config.Output = writer

	provider := NewProvider()
	provider.Configure(config)

	return provider
}

// NewWithLogrusLogger cria um provider a partir de um logger Logrus existente
func NewWithLogrusLogger(logrusLogger *logrus.Logger) *Provider {
	return NewProviderWithLogger(logrusLogger)
}

// NewTextProvider cria um provider configurado para saída em texto
func NewTextProvider() *Provider {
	config := DefaultConfig()
	config.Format = interfaces.TextFormat

	provider := NewProvider()
	provider.Configure(config)

	return provider
}

// NewJSONProvider cria um provider configurado para saída em JSON
func NewJSONProvider() *Provider {
	config := DefaultConfig()
	config.Format = interfaces.JSONFormat

	provider := NewProvider()
	provider.Configure(config)

	return provider
}

// NewConsoleProvider cria um provider otimizado para saída no console
func NewConsoleProvider() *Provider {
	config := DefaultConfig()
	config.Format = interfaces.ConsoleFormat
	config.Output = os.Stdout

	provider := NewProvider()
	provider.Configure(config)

	return provider
}

// NewBufferedProvider cria um provider com buffer configurado
func NewBufferedProvider(bufferSize int, flushTimeout time.Duration) *Provider {
	config := DefaultConfig()
	config.BufferConfig = &interfaces.BufferConfig{
		Size:         bufferSize,
		FlushTimeout: flushTimeout,
		BatchSize:    bufferSize / 10,
		MemoryLimit:  64 * 1024 * 1024, // 64MB
	}

	provider := NewProvider()
	provider.Configure(config)

	return provider
}

// CreateLogrusHookAdapter adapta hooks customizados para o Logrus
type LogrusHookAdapter struct {
	beforeHook interfaces.Hook
	afterHook  interfaces.Hook
}

// NewLogrusHookAdapter cria um adaptador para hooks personalizados
func NewLogrusHookAdapter(beforeHook, afterHook interfaces.Hook) *LogrusHookAdapter {
	return &LogrusHookAdapter{
		beforeHook: beforeHook,
		afterHook:  afterHook,
	}
}

// Levels retorna os níveis suportados pelo hook
func (h *LogrusHookAdapter) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire executa o hook
func (h *LogrusHookAdapter) Fire(entry *logrus.Entry) error {
	ctx := context.Background() // Contexto padrão para hooks do Logrus

	// Converte entrada do Logrus para nossa interface
	logEntry := &interfaces.LogEntry{
		Timestamp: entry.Time,
		Level:     logrusLevelToInterface(entry.Level),
		Message:   entry.Message,
		Fields:    make(map[string]any),
	}

	// Copia os campos
	for k, v := range entry.Data {
		logEntry.Fields[k] = v
	}

	// Executa hook before se disponível
	if h.beforeHook != nil {
		if err := h.beforeHook.Execute(ctx, logEntry); err != nil {
			return err
		}
	}

	// Executa hook after se disponível
	if h.afterHook != nil {
		if err := h.afterHook.Execute(ctx, logEntry); err != nil {
			return err
		}
	}

	// Atualiza a entrada original com possíveis modificações
	entry.Message = logEntry.Message
	for k, v := range logEntry.Fields {
		entry.Data[k] = v
	}

	return nil
}

// logrusLevelToInterface converte nível do Logrus para nossa interface
func logrusLevelToInterface(level logrus.Level) interfaces.Level {
	switch level {
	case logrus.DebugLevel:
		return interfaces.DebugLevel
	case logrus.InfoLevel:
		return interfaces.InfoLevel
	case logrus.WarnLevel:
		return interfaces.WarnLevel
	case logrus.ErrorLevel:
		return interfaces.ErrorLevel
	case logrus.FatalLevel:
		return interfaces.FatalLevel
	case logrus.PanicLevel:
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}
