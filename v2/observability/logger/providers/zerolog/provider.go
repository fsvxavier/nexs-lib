// Package zerolog implementa o provider Zerolog para o sistema de logging v2
package zerolog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/rs/zerolog"
)

const ProviderName = "zerolog"

// Provider implementação do Zerolog seguindo a arquitetura hexagonal
type Provider struct {
	logger zerolog.Logger
	config interfaces.Config
	level  zerolog.Level
}

// NewProvider cria uma nova instância do provider Zerolog
func NewProvider() *Provider {
	return &Provider{
		level: zerolog.InfoLevel,
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config interfaces.Config) error {
	p.config = config

	// Converte o nível
	p.level = convertLevel(config.Level)
	zerolog.SetGlobalLevel(p.level)

	// Writer de saída
	writer := config.Output
	if writer == nil {
		writer = os.Stdout
	}

	// Configuração do logger baseado no formato
	switch config.Format {
	case interfaces.JSONFormat:
		p.logger = zerolog.New(writer)
	case interfaces.ConsoleFormat:
		p.logger = zerolog.New(zerolog.ConsoleWriter{Out: writer, TimeFormat: time.RFC3339})
	case interfaces.TextFormat:
		p.logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
			NoColor:    true,
		})
	default:
		p.logger = zerolog.New(writer)
	}

	// Adiciona timestamp
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
	}
	p.logger = p.logger.With().Timestamp().Logger()

	// Adiciona caller se configurado
	if config.AddCaller {
		p.logger = p.logger.With().Caller().Logger()
	}

	// Adiciona campos globais
	if len(config.GlobalFields) > 0 {
		ctx := p.logger.With()
		for key, value := range config.GlobalFields {
			ctx = ctx.Interface(key, value)
		}
		p.logger = ctx.Logger()
	}

	// Adiciona informações do serviço
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
	p.logger = ctx.Logger()

	// Define o nível do logger
	p.level = convertLevel(config.Level)
	zerolog.SetGlobalLevel(p.level)
	p.logger = p.logger.Level(p.level)

	return nil
}

// Name retorna o nome do provider
func (p *Provider) Name() string {
	return ProviderName
}

// Version retorna a versão do provider
func (p *Provider) Version() string {
	return "1.34.0" // Versão atual do zerolog
}

// HealthCheck verifica se o provider está funcionando
func (p *Provider) HealthCheck() error {
	// Zerolog sempre está pronto
	return nil
}

// Implementação da interface Logger

func (p *Provider) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.GetLevel() <= zerolog.TraceLevel {
		event := p.logger.Trace()
		event = addFieldsToEvent(event, fields)
		event = addContextToEvent(event, ctx)
		event.Msg(msg)
	}
}

func (p *Provider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.GetLevel() <= zerolog.DebugLevel {
		event := p.logger.Debug()
		event = addFieldsToEvent(event, fields)
		event = addContextToEvent(event, ctx)
		event.Msg(msg)
	}
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.GetLevel() <= zerolog.InfoLevel {
		event := p.logger.Info()
		event = addFieldsToEvent(event, fields)
		event = addContextToEvent(event, ctx)
		event.Msg(msg)
	}
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.GetLevel() <= zerolog.WarnLevel {
		event := p.logger.Warn()
		event = addFieldsToEvent(event, fields)
		event = addContextToEvent(event, ctx)
		event.Msg(msg)
	}
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.GetLevel() <= zerolog.ErrorLevel {
		event := p.logger.Error()
		event = addFieldsToEvent(event, fields)
		event = addContextToEvent(event, ctx)

		// Adiciona stack trace se configurado
		if p.config.AddStacktrace {
			event = event.Stack()
		}

		event.Msg(msg)
	}
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	event := p.logger.Fatal()
	event = addFieldsToEvent(event, fields)
	event = addContextToEvent(event, ctx)
	event = event.Stack()
	event.Msg(msg)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	event := p.logger.Panic()
	event = addFieldsToEvent(event, fields)
	event = addContextToEvent(event, ctx)
	event = event.Stack()
	event.Msg(msg)
}

// Métodos com formatação

func (p *Provider) Tracef(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Trace(ctx, msg)
}

func (p *Provider) Debugf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Debug(ctx, msg)
}

func (p *Provider) Infof(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Info(ctx, msg)
}

func (p *Provider) Warnf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Warn(ctx, msg)
}

func (p *Provider) Errorf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Error(ctx, msg)
}

func (p *Provider) Fatalf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Fatal(ctx, msg)
}

func (p *Provider) Panicf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	p.Panic(ctx, msg)
}

// Métodos com código

func (p *Provider) TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.String("code", code))
	p.Trace(ctx, msg, allFields...)
}

func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.String("code", code))
	p.Debug(ctx, msg, allFields...)
}

func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.String("code", code))
	p.Info(ctx, msg, allFields...)
}

func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.String("code", code))
	p.Warn(ctx, msg, allFields...)
}

func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	allFields := append(fields, interfaces.String("code", code))
	p.Error(ctx, msg, allFields...)
}

// Métodos utilitários

func (p *Provider) WithFields(fields ...interfaces.Field) interfaces.Logger {
	ctx := p.logger.With()
	for _, field := range fields {
		ctx = addFieldToContext(ctx, field)
	}

	return &Provider{
		logger: ctx.Logger(),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithContext(ctx context.Context) interfaces.Logger {
	loggerCtx := p.logger.With()

	// Extrai trace ID
	if traceID := extractTraceID(ctx); traceID != "" {
		loggerCtx = loggerCtx.Str("trace_id", traceID)
	}

	// Extrai span ID
	if spanID := extractSpanID(ctx); spanID != "" {
		loggerCtx = loggerCtx.Str("span_id", spanID)
	}

	// Extrai user ID
	if userID := extractUserID(ctx); userID != "" {
		loggerCtx = loggerCtx.Str("user_id", userID)
	}

	// Extrai request ID
	if requestID := extractRequestID(ctx); requestID != "" {
		loggerCtx = loggerCtx.Str("request_id", requestID)
	}

	return &Provider{
		logger: loggerCtx.Logger(),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithError(err error) interfaces.Logger {
	if err == nil {
		return p
	}

	return &Provider{
		logger: p.logger.With().Err(err).Logger(),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithTraceID(traceID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With().Str("trace_id", traceID).Logger(),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithSpanID(spanID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With().Str("span_id", spanID).Logger(),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) SetLevel(level interfaces.Level) {
	p.level = convertLevel(level)
	zerolog.SetGlobalLevel(p.level)
}

func (p *Provider) GetLevel() interfaces.Level {
	return convertLevelFromZerolog(p.level)
}

func (p *Provider) IsLevelEnabled(level interfaces.Level) bool {
	zerologLevel := convertLevel(level)
	return p.logger.GetLevel() <= zerologLevel
}

func (p *Provider) Clone() interfaces.Logger {
	return &Provider{
		logger: p.logger,
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) Flush() error {
	// Zerolog não tem método de flush explícito
	return nil
}

func (p *Provider) Close() error {
	// Zerolog não requer fechamento explícito
	return nil
}

// Funções auxiliares

func convertLevel(level interfaces.Level) zerolog.Level {
	switch level {
	case interfaces.TraceLevel:
		return zerolog.TraceLevel
	case interfaces.DebugLevel:
		return zerolog.DebugLevel
	case interfaces.InfoLevel:
		return zerolog.InfoLevel
	case interfaces.WarnLevel:
		return zerolog.WarnLevel
	case interfaces.ErrorLevel:
		return zerolog.ErrorLevel
	case interfaces.FatalLevel:
		return zerolog.FatalLevel
	case interfaces.PanicLevel:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func convertLevelFromZerolog(level zerolog.Level) interfaces.Level {
	switch level {
	case zerolog.TraceLevel:
		return interfaces.TraceLevel
	case zerolog.DebugLevel:
		return interfaces.DebugLevel
	case zerolog.InfoLevel:
		return interfaces.InfoLevel
	case zerolog.WarnLevel:
		return interfaces.WarnLevel
	case zerolog.ErrorLevel:
		return interfaces.ErrorLevel
	case zerolog.FatalLevel:
		return interfaces.FatalLevel
	case zerolog.PanicLevel:
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}

func addFieldsToEvent(event *zerolog.Event, fields []interfaces.Field) *zerolog.Event {
	for _, field := range fields {
		event = addFieldToEvent(event, field)
	}
	return event
}

func addFieldToEvent(event *zerolog.Event, field interfaces.Field) *zerolog.Event {
	switch field.Type {
	case interfaces.StringType:
		return event.Str(field.Key, field.Value.(string))
	case interfaces.IntType:
		return event.Int(field.Key, field.Value.(int))
	case interfaces.Int64Type:
		return event.Int64(field.Key, field.Value.(int64))
	case interfaces.Float64Type:
		return event.Float64(field.Key, field.Value.(float64))
	case interfaces.BoolType:
		return event.Bool(field.Key, field.Value.(bool))
	case interfaces.TimeType:
		return event.Time(field.Key, field.Value.(time.Time))
	case interfaces.DurationType:
		return event.Dur(field.Key, field.Value.(time.Duration))
	case interfaces.ErrorType:
		if field.Value == nil {
			return event.Interface(field.Key, nil)
		}
		if err, ok := field.Value.(error); ok {
			if field.Key == "error" {
				return event.Err(err)
			}
			return event.AnErr(field.Key, err)
		}
		if str, ok := field.Value.(string); ok {
			return event.Str(field.Key, str)
		}
		return event.Interface(field.Key, field.Value)
	default:
		return event.Interface(field.Key, field.Value)
	}
}

func addFieldToContext(ctx zerolog.Context, field interfaces.Field) zerolog.Context {
	switch field.Type {
	case interfaces.StringType:
		return ctx.Str(field.Key, field.Value.(string))
	case interfaces.IntType:
		return ctx.Int(field.Key, field.Value.(int))
	case interfaces.Int64Type:
		return ctx.Int64(field.Key, field.Value.(int64))
	case interfaces.Float64Type:
		return ctx.Float64(field.Key, field.Value.(float64))
	case interfaces.BoolType:
		return ctx.Bool(field.Key, field.Value.(bool))
	case interfaces.TimeType:
		return ctx.Time(field.Key, field.Value.(time.Time))
	case interfaces.DurationType:
		return ctx.Dur(field.Key, field.Value.(time.Duration))
	case interfaces.ErrorType:
		if field.Value == nil {
			return ctx.Interface(field.Key, nil)
		}
		if err, ok := field.Value.(error); ok {
			if field.Key == "error" {
				return ctx.Err(err)
			}
			return ctx.AnErr(field.Key, err)
		}
		if str, ok := field.Value.(string); ok {
			return ctx.Str(field.Key, str)
		}
		return ctx.Interface(field.Key, field.Value)
	default:
		return ctx.Interface(field.Key, field.Value)
	}
}

func addContextToEvent(event *zerolog.Event, ctx context.Context) *zerolog.Event {
	if ctx == nil {
		return event
	}

	// Extrai trace ID
	if traceID := extractTraceID(ctx); traceID != "" {
		event = event.Str("trace_id", traceID)
	}

	// Extrai span ID
	if spanID := extractSpanID(ctx); spanID != "" {
		event = event.Str("span_id", spanID)
	}

	// Extrai user ID
	if userID := extractUserID(ctx); userID != "" {
		event = event.Str("user_id", userID)
	}

	// Extrai request ID
	if requestID := extractRequestID(ctx); requestID != "" {
		event = event.Str("request_id", requestID)
	}

	return event
}

// Context extractors - mesmas implementações do core
func extractTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value("trace_id"); value != nil {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}

	if value := ctx.Value("traceId"); value != nil {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}

	return ""
}

func extractSpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value("span_id"); value != nil {
		if spanID, ok := value.(string); ok {
			return spanID
		}
	}

	if value := ctx.Value("spanId"); value != nil {
		if spanID, ok := value.(string); ok {
			return spanID
		}
	}

	return ""
}

func extractUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value("user_id"); value != nil {
		if userID, ok := value.(string); ok {
			return userID
		}
	}

	return ""
}

func extractRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value("request_id"); value != nil {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}

	return ""
}
