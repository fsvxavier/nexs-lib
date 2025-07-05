package zerolog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/rs/zerolog"
)

const ProviderName = "zerolog"

// Provider implementação do Zerolog para o sistema de logging
type Provider struct {
	logger        *zerolog.Logger
	config        *logger.Config
	level         zerolog.Level
	mu            sync.RWMutex
	contextFields map[string]any
}

// NewProvider cria uma nova instância do provider Zerolog
func NewProvider() *Provider {
	return &Provider{
		level:         zerolog.InfoLevel,
		contextFields: make(map[string]any),
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Configura o nível
	zerologLevel := p.convertLevel(config.Level)
	p.level = zerologLevel
	zerolog.SetGlobalLevel(zerologLevel)

	// Configura o writer
	var writer io.Writer = config.Output
	if writer == nil {
		writer = os.Stdout
	}

	// Configura o formato
	switch config.Format {
	case logger.ConsoleFormat:
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: config.TimeFormat,
			NoColor:    false,
		}
	case logger.TextFormat:
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: config.TimeFormat,
			NoColor:    true,
		}
	case logger.JSONFormat:
		// JSON é o formato padrão do zerolog
	}

	// Cria o logger base
	newWriter := zerolog.New(writer)
	p.logger = &newWriter

	// Configura timestamp
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
		timeZerolog := p.logger.With().Timestamp().Logger()
		p.logger = &timeZerolog
	}

	// Adiciona caller se solicitado
	if config.AddSource {
		sourceZerolog := p.logger.With().Caller().Logger()
		p.logger = &sourceZerolog
	}

	// Configura campos globais
	ctx := p.logger.With()

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

// convertLevel converte o nível interno para o nível do Zerolog
func (p *Provider) convertLevel(level logger.Level) zerolog.Level {
	switch level {
	case logger.DebugLevel:
		return zerolog.DebugLevel
	case logger.InfoLevel:
		return zerolog.InfoLevel
	case logger.WarnLevel:
		return zerolog.WarnLevel
	case logger.ErrorLevel:
		return zerolog.ErrorLevel
	case logger.FatalLevel:
		return zerolog.FatalLevel
	case logger.PanicLevel:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// addFields adiciona campos ao evento de log
func (p *Provider) addFields(event *zerolog.Event, fields []logger.Field) *zerolog.Event {
	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			event = event.Str(field.Key, v)
		case int:
			event = event.Int(field.Key, v)
		case int32:
			event = event.Int32(field.Key, v)
		case int64:
			event = event.Int64(field.Key, v)
		case float32:
			event = event.Float32(field.Key, v)
		case float64:
			event = event.Float64(field.Key, v)
		case bool:
			event = event.Bool(field.Key, v)
		case time.Duration:
			event = event.Dur(field.Key, v)
		case time.Time:
			event = event.Time(field.Key, v)
		case error:
			event = event.Err(v)
		case []byte:
			event = event.Bytes(field.Key, v)
		default:
			event = event.Interface(field.Key, v)
		}
	}
	return event
}

// extractContext extrai informações relevantes do contexto
func (p *Provider) extractContextFields(ctx context.Context) []logger.Field {
	var fields []logger.Field

	// Adiciona trace ID se disponível
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			fields = append(fields, logger.String("trace_id", id))
		}
	}

	// Adiciona span ID se disponível
	if spanID := ctx.Value("span_id"); spanID != nil {
		if id, ok := spanID.(string); ok && id != "" {
			fields = append(fields, logger.String("span_id", id))
		}
	}

	// Adiciona user ID se disponível
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			fields = append(fields, logger.String("user_id", id))
		}
	}

	// Adiciona request ID se disponível
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			fields = append(fields, logger.String("request_id", id))
		}
	}

	return fields
}

// Implementação da interface Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.DebugLevel {
		return
	}

	event := p.logger.Debug()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	if p.config != nil && p.config.AddStacktrace {
		event = event.Stack()
	}

	p.addFields(event, allFields).Msg(msg)
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.InfoLevel {
		return
	}

	event := p.logger.Info()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	p.addFields(event, allFields).Msg(msg)
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.WarnLevel {
		return
	}

	event := p.logger.Warn()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	p.addFields(event, allFields).Msg(msg)
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	if p.logger.GetLevel() > zerolog.ErrorLevel {
		return
	}

	event := p.logger.Error()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	if p.config != nil && p.config.AddStacktrace {
		event = event.Stack()
	}

	p.addFields(event, allFields).Msg(msg)
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	event := p.logger.Fatal()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	event = event.Stack()
	p.addFields(event, allFields).Msg(msg)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	event := p.logger.Panic()
	contextFields := p.extractContextFields(ctx)
	allFields := append(contextFields, fields...)

	event = event.Stack()
	p.addFields(event, allFields).Msg(msg)
}

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

func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	event := p.logger.Fatal().Stack()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	event := p.logger.Panic().Stack()
	contextFields := p.extractContextFields(ctx)

	if len(contextFields) > 0 {
		event = p.addFields(event, contextFields)
	}

	event.Msgf(format, args...)
}

func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	allFields := append(fields, logger.String("code", code))
	p.Debug(ctx, msg, allFields...)
}

func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	allFields := append(fields, logger.String("code", code))
	p.Info(ctx, msg, allFields...)
}

func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	allFields := append(fields, logger.String("code", code))
	p.Warn(ctx, msg, allFields...)
}

func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	allFields := append(fields, logger.String("code", code))
	p.Error(ctx, msg, allFields...)
}

func (p *Provider) WithFields(fields ...logger.Field) logger.Logger {
	ctx := p.logger.With()

	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			ctx = ctx.Str(field.Key, v)
		case int:
			ctx = ctx.Int(field.Key, v)
		case int64:
			ctx = ctx.Int64(field.Key, v)
		case float64:
			ctx = ctx.Float64(field.Key, v)
		case bool:
			ctx = ctx.Bool(field.Key, v)
		case time.Duration:
			ctx = ctx.Dur(field.Key, v)
		case time.Time:
			ctx = ctx.Time(field.Key, v)
		case error:
			ctx = ctx.AnErr(field.Key, v)
		default:
			ctx = ctx.Interface(field.Key, v)
		}
	}

	ctxLogger := ctx.Logger()
	newProvider := &Provider{
		logger:        &ctxLogger,
		config:        p.config,
		level:         p.level,
		contextFields: make(map[string]any),
	}

	p.mu.RLock()
	// Copia campos existentes de forma thread-safe
	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}
	p.mu.RUnlock()

	// Adiciona novos campos
	for _, field := range fields {
		newProvider.contextFields[field.Key] = field.Value
	}

	return newProvider
}

func (p *Provider) WithContext(ctx context.Context) logger.Logger {
	contextFields := p.extractContextFields(ctx)
	if len(contextFields) == 0 {
		return p
	}

	return p.WithFields(contextFields...)
}

func (p *Provider) SetLevel(level logger.Level) {
	zerologLevel := p.convertLevel(level)
	p.level = zerologLevel
	zerolog.SetGlobalLevel(zerologLevel)
}

func (p *Provider) GetLevel() logger.Level {
	switch p.logger.GetLevel() {
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

func (p *Provider) Clone() logger.Logger {
	newProvider := &Provider{
		logger:        p.logger,
		config:        p.config,
		level:         p.level,
		contextFields: make(map[string]any),
	}

	p.mu.RLock()
	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}
	p.mu.RUnlock()

	return newProvider
}

func (p *Provider) Close() error {
	// Zerolog não tem método de close explícito
	return nil
}

// GetZerologLogger retorna o logger Zerolog subjacente para uso avançado
func (p *Provider) GetZerologLogger() zerolog.Logger {
	return *p.logger
}

// StackTrace captura e retorna o stack trace atual
func StackTrace(skip int) logger.Field {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack []string
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return logger.String("stack_trace", strings.Join(stack, "\n"))
}

// Hook personalizado para adicionar funcionalidades extras
type Hook struct{}

func (h Hook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Adiciona hostname se disponível
	if hostname, err := os.Hostname(); err == nil {
		e.Str("hostname", hostname)
	}
}

// init registra automaticamente o provider
func init() {
	logger.RegisterProvider(ProviderName, NewProvider())
}
