package zerolog

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"

	"github.com/fsvxavier/nexs-lib/observability/logger"
)

// Provider implementa o provider de logging usando Zerolog
type Provider struct {
	config *logger.Config
	logger *zerolog.Logger
	level  zerolog.Level
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
	var writer io.Writer
	if config.Output != nil {
		if w, ok := config.Output.(io.Writer); ok {
			writer = w
		} else {
			writer = os.Stdout
		}
	} else {
		writer = os.Stdout
	}

	// Configura o format
	switch config.Format {
	case logger.JSONFormat:
		// JSON é o formato padrão do zerolog
	case logger.ConsoleFormat:
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: config.TimeFormat,
		}
	case logger.TextFormat:
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: config.TimeFormat,
			NoColor:    true,
		}
	}

	// Configura o logger base
	zerologLogger := zerolog.New(writer).Level(p.level)

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
	}
}

// Close implementa Logger
func (p *Provider) Close() error {
	return nil
}

// Certifica que Provider implementa as interfaces
var (
	_ logger.Logger   = (*Provider)(nil)
	_ logger.Provider = (*Provider)(nil)
)
