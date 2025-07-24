package zap

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

	// Tenta fazer parse da entrada de log do Zap
	var zapEntry map[string]interface{}
	if err := json.Unmarshal(p, &zapEntry); err != nil {
		// Se não conseguir fazer parse, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	// Converte para LogEntry
	entry := bw.zapEntryToLogEntry(zapEntry)

	// Escreve no buffer
	if err := bw.provider.buffer.Write(entry); err != nil {
		// Se falhar no buffer, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	return len(p), nil
}

// zapEntryToLogEntry converte uma entrada do Zap para LogEntry
func (bw *bufferWriter) zapEntryToLogEntry(zapEntry map[string]interface{}) *interfaces.LogEntry {
	entry := &interfaces.LogEntry{
		Timestamp: time.Now(),
		Level:     interfaces.InfoLevel,
		Message:   "",
		Fields:    make(map[string]any),
	}

	// Extrai campos conhecidos
	if ts, ok := zapEntry["timestamp"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
	}

	if level, ok := zapEntry["level"].(string); ok {
		entry.Level = bw.stringToLevel(level)
	}

	if msg, ok := zapEntry["message"].(string); ok {
		entry.Message = msg
	}

	if code, ok := zapEntry["code"].(string); ok {
		entry.Code = code
	}

	// Copia outros campos
	for key, value := range zapEntry {
		if key != "timestamp" && key != "level" && key != "message" && key != "code" {
			entry.Fields[key] = value
		}
	}

	return entry
}

// stringToLevel converte string de nível para Level
func (bw *bufferWriter) stringToLevel(level string) interfaces.Level {
	switch level {
	case "debug", "DEBUG":
		return interfaces.DebugLevel
	case "info", "INFO":
		return interfaces.InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		return interfaces.WarnLevel
	case "error", "ERROR":
		return interfaces.ErrorLevel
	case "fatal", "FATAL":
		return interfaces.FatalLevel
	case "panic", "PANIC":
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}

// Provider implementa o provider de logging usando Zap
type Provider struct {
	config *logger.Config
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	writer io.Writer
	buffer *logger.CircularBuffer
}

// NewProvider cria uma nova instância do provider Zap
func NewProvider() *Provider {
	return &Provider{}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Mapeia os níveis
	var level zapcore.Level
	switch config.Level {
	case logger.DebugLevel:
		level = zapcore.DebugLevel
	case logger.InfoLevel:
		level = zapcore.InfoLevel
	case logger.WarnLevel:
		level = zapcore.WarnLevel
	case logger.ErrorLevel:
		level = zapcore.ErrorLevel
	case logger.FatalLevel:
		level = zapcore.FatalLevel
	case logger.PanicLevel:
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configura o writer
	if config.Output != nil {
		if w, ok := config.Output.(io.Writer); ok {
			p.writer = w
		} else {
			p.writer = os.Stdout
		}
	} else {
		p.writer = os.Stdout
	}

	// Configura o buffer se habilitado
	if config.BufferConfig != nil && config.BufferConfig.Enabled {
		p.buffer = logger.NewCircularBuffer(config.BufferConfig, p.writer)
	}

	// Configura o encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()

	// Configura o formato de timestamp
	if config.TimeFormat != "" {
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimeFormat)
	} else {
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	}

	// Configura o formato de saída
	switch config.Format {
	case logger.JSONFormat:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case logger.ConsoleFormat:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case logger.TextFormat:
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Escolhe o writer final (buffer ou direto)
	var finalWriter io.Writer
	if p.buffer != nil {
		// Usa um writer customizado que escreve através do buffer
		finalWriter = &bufferWriter{provider: p}
	} else {
		finalWriter = p.writer
	}

	// Configura o core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(finalWriter),
		level,
	)

	// Configura sampling se especificado
	if config.SamplingConfig != nil {
		samplingConfig := &zap.SamplingConfig{
			Initial:    config.SamplingConfig.Initial,
			Thereafter: config.SamplingConfig.Thereafter,
		}
		core = zapcore.NewSamplerWithOptions(core, time.Second, samplingConfig.Initial, samplingConfig.Thereafter)
	}

	// Cria o logger
	var opts []zap.Option

	// Adiciona caller se necessário
	if config.AddSource {
		opts = append(opts, zap.AddCaller())
	}

	// Adiciona stacktrace se necessário
	if config.AddStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	p.logger = zap.New(core, opts...)

	// Adiciona campos globais
	fields := make([]zap.Field, 0)
	if config.ServiceName != "" {
		fields = append(fields, zap.String("service", config.ServiceName))
	}
	if config.ServiceVersion != "" {
		fields = append(fields, zap.String("version", config.ServiceVersion))
	}
	if config.Environment != "" {
		fields = append(fields, zap.String("environment", config.Environment))
	}

	// Adiciona campos customizados
	for k, v := range config.Fields {
		fields = append(fields, zap.Any(k, v))
	}

	if len(fields) > 0 {
		p.logger = p.logger.With(fields...)
	}

	p.sugar = p.logger.Sugar()
	return nil
}

// extractContextFields extrai campos relevantes do contexto
func (p *Provider) extractContextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0)

	if traceID := ctx.Value(logger.TraceIDKey); traceID != nil {
		fields = append(fields, zap.Any(string(logger.TraceIDKey), traceID))
	}

	if spanID := ctx.Value(logger.SpanIDKey); spanID != nil {
		fields = append(fields, zap.Any(string(logger.SpanIDKey), spanID))
	}

	if userID := ctx.Value(logger.UserIDKey); userID != nil {
		fields = append(fields, zap.Any(string(logger.UserIDKey), userID))
	}

	if requestID := ctx.Value(logger.RequestIDKey); requestID != nil {
		fields = append(fields, zap.Any(string(logger.RequestIDKey), requestID))
	}

	return fields
}

// fieldsToZapFields converte logger.Field para zap.Field
func (p *Provider) fieldsToZapFields(fields []logger.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// Debug implementa Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Debug(msg, allFields...)
}

// Info implementa Logger
func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Info(msg, allFields...)
}

// Warn implementa Logger
func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Warn(msg, allFields...)
}

// Error implementa Logger
func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Error(msg, allFields...)
}

// Fatal implementa Logger
func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Fatal(msg, allFields...)
}

// Panic implementa Logger
func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields))
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zapFields...)

	p.logger.Panic(msg, allFields...)
}

// Debugf implementa Logger
func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Debugf(format, args...)
	} else {
		p.sugar.Debugf(format, args...)
	}
}

// Infof implementa Logger
func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Infof(format, args...)
	} else {
		p.sugar.Infof(format, args...)
	}
}

// Warnf implementa Logger
func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Warnf(format, args...)
	} else {
		p.sugar.Warnf(format, args...)
	}
}

// Errorf implementa Logger
func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Errorf(format, args...)
	} else {
		p.sugar.Errorf(format, args...)
	}
}

// Fatalf implementa Logger
func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Fatalf(format, args...)
	} else {
		p.sugar.Fatalf(format, args...)
	}
}

// Panicf implementa Logger
func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Panicf(format, args...)
	} else {
		p.sugar.Panicf(format, args...)
	}
}

// DebugWithCode implementa Logger
func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields)+1)
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zap.String("code", code))
	allFields = append(allFields, zapFields...)

	p.logger.Debug(msg, allFields...)
}

// InfoWithCode implementa Logger
func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields)+1)
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zap.String("code", code))
	allFields = append(allFields, zapFields...)

	p.logger.Info(msg, allFields...)
}

// WarnWithCode implementa Logger
func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields)+1)
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zap.String("code", code))
	allFields = append(allFields, zapFields...)

	p.logger.Warn(msg, allFields...)
}

// ErrorWithCode implementa Logger
func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	contextFields := p.extractContextFields(ctx)
	zapFields := p.fieldsToZapFields(fields)

	allFields := make([]zap.Field, 0, len(contextFields)+len(zapFields)+1)
	allFields = append(allFields, contextFields...)
	allFields = append(allFields, zap.String("code", code))
	allFields = append(allFields, zapFields...)

	p.logger.Error(msg, allFields...)
}

// WithFields implementa Logger
func (p *Provider) WithFields(fields ...logger.Field) logger.Logger {
	zapFields := p.fieldsToZapFields(fields)
	newLogger := p.logger.With(zapFields...)

	return &Provider{
		config: p.config,
		logger: newLogger,
		sugar:  newLogger.Sugar(),
		writer: p.writer,
		buffer: p.buffer,
	}
}

// WithContext implementa Logger
func (p *Provider) WithContext(ctx context.Context) logger.Logger {
	contextFields := p.extractContextFields(ctx)
	if len(contextFields) == 0 {
		return p
	}

	newLogger := p.logger.With(contextFields...)
	return &Provider{
		config: p.config,
		logger: newLogger,
		sugar:  newLogger.Sugar(),
		writer: p.writer,
		buffer: p.buffer,
	}
}

// SetLevel implementa Logger
func (p *Provider) SetLevel(level logger.Level) {
	// Zap não suporta mudança de nível durante runtime
}

// GetLevel implementa Logger
func (p *Provider) GetLevel() logger.Level {
	if p.logger.Core().Enabled(zapcore.DebugLevel) {
		return logger.DebugLevel
	}
	if p.logger.Core().Enabled(zapcore.InfoLevel) {
		return logger.InfoLevel
	}
	if p.logger.Core().Enabled(zapcore.WarnLevel) {
		return logger.WarnLevel
	}
	if p.logger.Core().Enabled(zapcore.ErrorLevel) {
		return logger.ErrorLevel
	}
	return logger.InfoLevel
}

// Clone implementa Logger
func (p *Provider) Clone() logger.Logger {
	return &Provider{
		config: p.config,
		logger: p.logger,
		sugar:  p.sugar,
		writer: p.writer,
		buffer: p.buffer,
	}
}

// Close implementa Logger
func (p *Provider) Close() error {
	if p.buffer != nil {
		if err := p.buffer.Close(); err != nil {
			return err
		}
	}
	return p.logger.Sync()
}

// GetBuffer retorna o buffer atual
func (p *Provider) GetBuffer() interfaces.Buffer {
	return p.buffer
}

// SetBuffer define um novo buffer
func (p *Provider) SetBuffer(buffer interfaces.Buffer) error {
	// Flush do buffer anterior se existir
	if p.buffer != nil {
		if err := p.buffer.Flush(); err != nil {
			return err
		}
		if err := p.buffer.Close(); err != nil {
			return err
		}
	}

	p.buffer = buffer.(*logger.CircularBuffer)
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

// Certifica que Provider implementa as interfaces
var (
	_ logger.Logger   = (*Provider)(nil)
	_ logger.Provider = (*Provider)(nil)
)
