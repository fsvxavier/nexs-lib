package logrus

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// bufferWriter implementa io.Writer para integrar com o buffer
type bufferWriter struct {
	provider *Provider
}

// Write implementa io.Writer escrevendo através do buffer
func (bw *bufferWriter) Write(p []byte) (n int, err error) {
	if bw.provider.buffer == nil {
		return bw.provider.writer.Write(p)
	}

	// Tenta fazer parse da entrada de log do Logrus
	var logrusEntry map[string]interface{}
	if err := json.Unmarshal(p, &logrusEntry); err != nil {
		// Se não conseguir fazer parse, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	// Converte para LogEntry
	entry := bw.logrusEntryToLogEntry(logrusEntry)

	// Escreve no buffer
	if err := bw.provider.buffer.Write(entry); err != nil {
		// Se falhar no buffer, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	return len(p), nil
}

// logrusEntryToLogEntry converte uma entrada do Logrus para LogEntry
func (bw *bufferWriter) logrusEntryToLogEntry(logrusEntry map[string]interface{}) *interfaces.LogEntry {
	entry := &interfaces.LogEntry{
		Timestamp: time.Now(),
		Level:     interfaces.InfoLevel,
		Message:   "",
		Fields:    make(map[string]any),
	}

	// Extrai campos conhecidos
	if ts, ok := logrusEntry["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
	}

	if level, ok := logrusEntry["level"].(string); ok {
		entry.Level = bw.stringToLevel(level)
	}

	if msg, ok := logrusEntry["msg"].(string); ok {
		entry.Message = msg
	}

	if code, ok := logrusEntry["code"].(string); ok {
		entry.Code = code
	}

	// Copia outros campos
	for key, value := range logrusEntry {
		if key != "time" && key != "level" && key != "msg" && key != "code" {
			entry.Fields[key] = value
		}
	}

	return entry
}

// stringToLevel converte string de nível para Level
func (bw *bufferWriter) stringToLevel(level string) interfaces.Level {
	switch level {
	case "debug":
		return interfaces.DebugLevel
	case "info":
		return interfaces.InfoLevel
	case "warn", "warning":
		return interfaces.WarnLevel
	case "error":
		return interfaces.ErrorLevel
	case "fatal":
		return interfaces.FatalLevel
	case "panic":
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}

// Provider implementa interfaces.Provider usando logrus
type Provider struct {
	logger      *logrus.Logger
	writer      io.Writer
	buffer      interfaces.Buffer
	bufferStats interfaces.BufferStats
	fields      map[string]any
	level       interfaces.Level
}

// NewProvider cria uma nova instância do provider Logrus
func NewProvider() *Provider {
	logrusLogger := logrus.New()

	provider := &Provider{
		logger: logrusLogger,
		writer: os.Stdout,
		fields: make(map[string]any),
		level:  interfaces.InfoLevel,
	}

	// Configura o buffer writer
	bufWriter := &bufferWriter{provider: provider}
	logrusLogger.SetOutput(bufWriter)
	logrusLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logrusLogger.SetLevel(logrus.InfoLevel)

	return provider
}

// NewProviderWithLogger cria um provider com um logger Logrus existente
func NewProviderWithLogger(logrusLogger *logrus.Logger) *Provider {
	provider := &Provider{
		logger: logrusLogger,
		writer: os.Stdout,
		fields: make(map[string]any),
		level:  interfaces.InfoLevel,
	}

	// Configura o buffer writer se não estiver configurado
	if logrusLogger.Out == os.Stdout || logrusLogger.Out == os.Stderr {
		bufWriter := &bufferWriter{provider: provider}
		logrusLogger.SetOutput(bufWriter)
	}

	return provider
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *interfaces.Config) error {
	if config == nil {
		return nil
	}

	// Configura nível
	p.SetLevel(config.Level)

	// Configura output
	if config.Output != nil {
		if writer, ok := config.Output.(io.Writer); ok {
			p.writer = writer
		}
	}

	// Configura formato
	switch config.Format {
	case interfaces.JSONFormat:
		p.logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.TimeFormat,
		})
	case interfaces.TextFormat, interfaces.ConsoleFormat:
		p.logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: config.TimeFormat,
			FullTimestamp:   true,
		})
	}

	// Configura campos globais
	if config.Fields != nil {
		for k, v := range config.Fields {
			p.fields[k] = v
		}
	}

	// Adiciona campos de contexto de serviço
	if config.ServiceName != "" {
		p.fields["service"] = config.ServiceName
	}
	if config.ServiceVersion != "" {
		p.fields["version"] = config.ServiceVersion
	}
	if config.Environment != "" {
		p.fields["env"] = config.Environment
	}

	// Configura buffer se especificado
	if config.BufferConfig != nil {
		if buffer := logger.NewCircularBuffer(config.BufferConfig, p.writer); buffer != nil {
			p.SetBuffer(buffer)
		}
	}

	return nil
}

// levelToLogrus converte Level para logrus.Level
func (p *Provider) levelToLogrus(level interfaces.Level) logrus.Level {
	switch level {
	case interfaces.DebugLevel:
		return logrus.DebugLevel
	case interfaces.InfoLevel:
		return logrus.InfoLevel
	case interfaces.WarnLevel:
		return logrus.WarnLevel
	case interfaces.ErrorLevel:
		return logrus.ErrorLevel
	case interfaces.FatalLevel:
		return logrus.FatalLevel
	case interfaces.PanicLevel:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// logrusToLevel converte logrus.Level para Level
func (p *Provider) logrusToLevel(level logrus.Level) interfaces.Level {
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

// prepareFields prepara os campos para log, incluindo campos globais
func (p *Provider) prepareFields(fields []interfaces.Field) logrus.Fields {
	logrusFields := make(logrus.Fields)

	// Adiciona campos globais
	for k, v := range p.fields {
		logrusFields[k] = v
	}

	// Adiciona campos específicos da chamada
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}

	return logrusFields
}

// Debug registra uma mensagem de debug
func (p *Provider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Debug(msg)
}

// Info registra uma mensagem informativa
func (p *Provider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Info(msg)
}

// Warn registra uma mensagem de warning
func (p *Provider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Warn(msg)
}

// Error registra uma mensagem de erro
func (p *Provider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Error(msg)
}

// Fatal registra uma mensagem fatal
func (p *Provider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Fatal(msg)
}

// Panic registra uma mensagem de panic
func (p *Provider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	logrusFields := p.prepareFields(fields)
	p.logger.WithFields(logrusFields).Panic(msg)
}

// Debugf registra uma mensagem de debug formatada
func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	p.logger.Debugf(format, args...)
}

// Infof registra uma mensagem informativa formatada
func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	p.logger.Infof(format, args...)
}

// Warnf registra uma mensagem de warning formatada
func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	p.logger.Warnf(format, args...)
}

// Errorf registra uma mensagem de erro formatada
func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	p.logger.Errorf(format, args...)
}

// Fatalf registra uma mensagem fatal formatada
func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	p.logger.Fatalf(format, args...)
}

// Panicf registra uma mensagem de panic formatada
func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	p.logger.Panicf(format, args...)
}

// DebugWithCode registra uma mensagem de debug com código
func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.Field{Key: "code", Value: code})
	p.Debug(ctx, msg, allFields...)
}

// InfoWithCode registra uma mensagem informativa com código
func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.Field{Key: "code", Value: code})
	p.Info(ctx, msg, allFields...)
}

// WarnWithCode registra uma mensagem de warning com código
func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.Field{Key: "code", Value: code})
	p.Warn(ctx, msg, allFields...)
}

// ErrorWithCode registra uma mensagem de erro com código
func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.Field{Key: "code", Value: code})
	p.Error(ctx, msg, allFields...)
}

// WithFields retorna um novo logger com campos adicionais
func (p *Provider) WithFields(fields ...interfaces.Field) interfaces.Logger {
	clone := p.Clone().(*Provider)
	for _, field := range fields {
		clone.fields[field.Key] = field.Value
	}
	return clone
}

// WithContext retorna um novo logger com contexto
func (p *Provider) WithContext(ctx context.Context) interfaces.Logger {
	// Logrus não possui suporte nativo para contexto desta forma
	// Mantemos a mesma instância
	return p
}

// SetLevel define o nível de log
func (p *Provider) SetLevel(level interfaces.Level) {
	p.level = level
	p.logger.SetLevel(p.levelToLogrus(level))
}

// GetLevel retorna o nível atual de log
func (p *Provider) GetLevel() interfaces.Level {
	return p.logrusToLevel(p.logger.GetLevel())
}

// Clone cria uma cópia do provider
func (p *Provider) Clone() interfaces.Logger {
	newFields := make(map[string]any)
	for k, v := range p.fields {
		newFields[k] = v
	}

	// Cria um novo logger Logrus baseado no atual
	newLogrusLogger := logrus.New()
	newLogrusLogger.SetLevel(p.logger.GetLevel())
	newLogrusLogger.SetFormatter(p.logger.Formatter)

	clone := &Provider{
		logger: newLogrusLogger,
		writer: p.writer,
		buffer: p.buffer,
		fields: newFields,
		level:  p.level,
	}

	// Configura o buffer writer para o clone
	bufWriter := &bufferWriter{provider: clone}
	newLogrusLogger.SetOutput(bufWriter)

	return clone
}

// Close fecha o provider e seus recursos
func (p *Provider) Close() error {
	if p.buffer != nil {
		return p.buffer.Flush()
	}
	return nil
}

// GetBuffer retorna o buffer atual
func (p *Provider) GetBuffer() interfaces.Buffer {
	return p.buffer
}

// SetBuffer define o buffer para o provider
func (p *Provider) SetBuffer(buffer interfaces.Buffer) error {
	p.buffer = buffer
	return nil
}

// FlushBuffer força o flush do buffer
func (p *Provider) FlushBuffer() error {
	if p.buffer != nil {
		return p.buffer.Flush()
	}
	return nil
}

// GetBufferStats retorna estatísticas do buffer
func (p *Provider) GetBufferStats() interfaces.BufferStats {
	if p.buffer != nil {
		return p.buffer.Stats()
	}
	return interfaces.BufferStats{}
}

// GetLogrusLogger retorna o logger Logrus subjacente para acesso direto
func (p *Provider) GetLogrusLogger() *logrus.Logger {
	return p.logger
}

// AddHook adiciona um hook do Logrus
func (p *Provider) AddHook(hook logrus.Hook) {
	p.logger.AddHook(hook)
}

// ReplaceHooks substitui todos os hooks
func (p *Provider) ReplaceHooks(hooks logrus.LevelHooks) {
	p.logger.ReplaceHooks(hooks)
}
