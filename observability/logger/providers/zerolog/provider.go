package zerolog

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// bufferWriterZerolog implementa io.Writer para integrar com o buffer
type bufferWriterZerolog struct {
	provider *Provider
}

// Write implementa io.Writer escrevendo através do buffer
func (bw *bufferWriterZerolog) Write(p []byte) (n int, err error) {
	if bw.provider.buffer == nil {
		return bw.provider.writer.Write(p)
	}

	// Tenta fazer parse da entrada de log do Zerolog
	var zerologEntry map[string]interface{}
	if err := json.Unmarshal(p, &zerologEntry); err != nil {
		// Se não conseguir fazer parse, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	// Converte para LogEntry
	entry := bw.zerologEntryToLogEntry(zerologEntry)

	// Escreve no buffer
	if err := bw.provider.buffer.Write(entry); err != nil {
		// Se falhar no buffer, escreve diretamente
		return bw.provider.writer.Write(p)
	}

	return len(p), nil
}

// zerologEntryToLogEntry converte uma entrada do Zerolog para LogEntry
func (bw *bufferWriterZerolog) zerologEntryToLogEntry(zerologEntry map[string]interface{}) *interfaces.LogEntry {
	entry := &interfaces.LogEntry{
		Timestamp: time.Now(),
		Level:     interfaces.InfoLevel,
		Message:   "",
		Fields:    make(map[string]any),
	}

	// Extrai campos conhecidos
	if ts, ok := zerologEntry["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			entry.Timestamp = parsed
		}
	}

	if level, ok := zerologEntry["level"].(string); ok {
		entry.Level = bw.stringToLevel(level)
	}

	if msg, ok := zerologEntry["message"].(string); ok {
		entry.Message = msg
	}

	if code, ok := zerologEntry["code"].(string); ok {
		entry.Code = code
	}

	// Copia outros campos
	for key, value := range zerologEntry {
		if key != "time" && key != "level" && key != "message" && key != "code" {
			entry.Fields[key] = value
		}
	}

	return entry
}

// stringToLevel converte string de nível para Level
func (bw *bufferWriterZerolog) stringToLevel(level string) interfaces.Level {
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

// Provider implementa o provider de logging usando Zerolog
type Provider struct {
	config *logger.Config
	logger *zerolog.Logger
	level  zerolog.Level
	writer io.Writer
	buffer *logger.CircularBuffer
}

// NewProvider cria uma nova instância do provider Zerolog
func NewProvider() *Provider {
	return &Provider{}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Mapeia os níveis
	switch config.Level {
	case logger.DebugLevel:
		p.level = zerolog.DebugLevel
	case logger.InfoLevel:
		p.level = zerolog.InfoLevel
	case logger.WarnLevel:
		p.level = zerolog.WarnLevel
	case logger.ErrorLevel:
		p.level = zerolog.ErrorLevel
	case logger.FatalLevel:
		p.level = zerolog.FatalLevel
	case logger.PanicLevel:
		p.level = zerolog.PanicLevel
	default:
		p.level = zerolog.InfoLevel
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

	// Configura o writer final (com ou sem buffer)
	var finalWriter io.Writer
	if p.buffer != nil {
		finalWriter = &bufferWriterZerolog{provider: p}
	} else {
		finalWriter = p.writer
	}

	// Configura o format
	switch config.Format {
	case logger.JSONFormat:
		// JSON é o formato padrão do zerolog
	case logger.ConsoleFormat:
		finalWriter = zerolog.ConsoleWriter{
			Out:        finalWriter,
			TimeFormat: config.TimeFormat,
		}
	case logger.TextFormat:
		finalWriter = zerolog.ConsoleWriter{
			Out:        finalWriter,
			TimeFormat: config.TimeFormat,
			NoColor:    true,
		}
	}

	// Configura o logger base
	zerologLogger := zerolog.New(finalWriter).Level(p.level)

	// Configura timestamp
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
		zerologLogger = zerologLogger.With().Timestamp().Logger()
	} else {
		zerologLogger = zerologLogger.With().Timestamp().Logger()
	}

	// Configura caller se necessário
	if config.AddSource {
		zerologLogger = zerologLogger.With().Caller().Logger()
	}

	// Cria contexto base
	ctx := zerologLogger.With()

	// Adiciona campos globais se especificados
	if config.ServiceName != "" {
		ctx = ctx.Str("service", config.ServiceName)
	}
	if config.ServiceVersion != "" {
		ctx = ctx.Str("version", config.ServiceVersion)
	}
	if config.Environment != "" {
		ctx = ctx.Str("environment", config.Environment)
	}

	// Adiciona campos customizados
	for k, v := range config.Fields {
		ctx = ctx.Interface(k, v)
	}

	ctxLogger := ctx.Logger()
	p.logger = &ctxLogger

	// Configura sampling se especificado
	if config.SamplingConfig != nil {
		sampler := &zerolog.BasicSampler{N: uint32(config.SamplingConfig.Thereafter)}
		loggerSample := p.logger.Sample(sampler)
		p.logger = &loggerSample
	}

	return nil
}

// extractContextFields extrai campos relevantes do contexto
func (p *Provider) extractContextFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})

	if traceID := ctx.Value(logger.TraceIDKey); traceID != nil {
		fields[string(logger.TraceIDKey)] = traceID
	}

	if spanID := ctx.Value(logger.SpanIDKey); spanID != nil {
		fields[string(logger.SpanIDKey)] = spanID
	}

	if userID := ctx.Value(logger.UserIDKey); userID != nil {
		fields[string(logger.UserIDKey)] = userID
	}

	if requestID := ctx.Value(logger.RequestIDKey); requestID != nil {
		fields[string(logger.RequestIDKey)] = requestID
	}

	return fields
}

// addFields adiciona campos ao evento do zerolog
func (p *Provider) addFields(event *zerolog.Event, fields map[string]interface{}) *zerolog.Event {
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return event
}

// fieldsToMap converte logger.Field para map
func (p *Provider) fieldsToMap(fields []logger.Field) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}

// Debug implementa Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	event := p.logger.Debug()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// Info implementa Logger
func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.InfoLevel {
		return
	}

	event := p.logger.Info()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// Warn implementa Logger
func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.WarnLevel {
		return
	}

	event := p.logger.Warn()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// Error implementa Logger
func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.ErrorLevel {
		return
	}

	event := p.logger.Error()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	if p.config != nil && p.config.AddStacktrace {
		event = event.Stack()
	}

	event.Msg(msg)
}

// Fatal implementa Logger
func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	event := p.logger.Fatal().Stack()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// Panic implementa Logger
func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	event := p.logger.Panic().Stack()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// Debugf implementa Logger
func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	if p.logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	event := p.logger.Debug()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

// Infof implementa Logger
func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	if p.logger.GetLevel() > zerolog.InfoLevel {
		return
	}

	event := p.logger.Info()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

// Warnf implementa Logger
func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	if p.logger.GetLevel() > zerolog.WarnLevel {
		return
	}

	event := p.logger.Warn()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

// Errorf implementa Logger
func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	if p.logger.GetLevel() > zerolog.ErrorLevel {
		return
	}

	event := p.logger.Error()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	if p.config != nil && p.config.AddStacktrace {
		event = event.Stack()
	}

	event.Msgf(format, args...)
}

// Fatalf implementa Logger
func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	event := p.logger.Fatal().Stack()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

// Panicf implementa Logger
func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	event := p.logger.Panic().Stack()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

// DebugWithCode implementa Logger
func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	event := p.logger.Debug()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event = event.Str("code", code)

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// InfoWithCode implementa Logger
func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.InfoLevel {
		return
	}

	event := p.logger.Info()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event = event.Str("code", code)

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// WarnWithCode implementa Logger
func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.WarnLevel {
		return
	}

	event := p.logger.Warn()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event = event.Str("code", code)

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	event.Msg(msg)
}

// ErrorWithCode implementa Logger
func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.ErrorLevel {
		return
	}

	event := p.logger.Error()
	contextFields := p.extractContextFields(ctx)
	loggerFields := p.fieldsToMap(fields)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event = event.Str("code", code)

	if len(loggerFields) > 0 {
		event = p.addFields(event, loggerFields)
	}

	if p.config != nil && p.config.AddStacktrace {
		event = event.Stack()
	}

	event.Msg(msg)
}

// WithFields implementa Logger
func (p *Provider) WithFields(fields ...logger.Field) logger.Logger {
	ctx := p.logger.With()
	for _, field := range fields {
		ctx = ctx.Interface(field.Key, field.Value)
	}

	ctxLogger := ctx.Logger()
	return &Provider{
		config: p.config,
		logger: &ctxLogger,
		level:  p.level,
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

	loggerCtx := p.logger.With()
	for k, v := range contextFields {
		loggerCtx = loggerCtx.Interface(k, v)
	}

	ctxLogger := loggerCtx.Logger()
	return &Provider{
		config: p.config,
		logger: &ctxLogger,
		level:  p.level,
		writer: p.writer,
		buffer: p.buffer,
	}
}

// SetLevel implementa Logger
func (p *Provider) SetLevel(level logger.Level) {
	switch level {
	case logger.DebugLevel:
		p.level = zerolog.DebugLevel
	case logger.InfoLevel:
		p.level = zerolog.InfoLevel
	case logger.WarnLevel:
		p.level = zerolog.WarnLevel
	case logger.ErrorLevel:
		p.level = zerolog.ErrorLevel
	case logger.FatalLevel:
		p.level = zerolog.FatalLevel
	case logger.PanicLevel:
		p.level = zerolog.PanicLevel
	default:
		p.level = zerolog.InfoLevel
	}

	// Atualiza o nível do logger
	*p.logger = p.logger.Level(p.level)
}

// GetLevel implementa Logger
func (p *Provider) GetLevel() logger.Level {
	switch p.level {
	case zerolog.DebugLevel:
		return logger.DebugLevel
	case zerolog.InfoLevel:
		return logger.InfoLevel
	case zerolog.WarnLevel:
		return logger.WarnLevel
	case zerolog.ErrorLevel:
		return logger.ErrorLevel
	case zerolog.FatalLevel:
		return logger.FatalLevel
	case zerolog.PanicLevel:
		return logger.PanicLevel
	default:
		return logger.InfoLevel
	}
}

// Clone implementa Logger
func (p *Provider) Clone() logger.Logger {
	return &Provider{
		config: p.config,
		logger: p.logger,
		level:  p.level,
		writer: p.writer,
		buffer: p.buffer,
	}
}

// Close implementa Logger
func (p *Provider) Close() error {
	if p.buffer != nil {
		return p.buffer.Close()
	}
	return nil
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
