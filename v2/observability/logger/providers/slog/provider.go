// Package slog implementa o provider Slog para o sistema de logging v2
package slog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

const ProviderName = "slog"

// Provider implementação do Slog seguindo a arquitetura hexagonal
type Provider struct {
	logger *slog.Logger
	config interfaces.Config
	level  slog.Level
}

// NewProvider cria uma nova instância do provider Slog
func NewProvider() *Provider {
	return &Provider{
		level: slog.LevelInfo,
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config interfaces.Config) error {
	p.config = config

	// Converte o nível
	p.level = convertLevel(config.Level)

	// Configuração do handler baseado no formato
	var handler slog.Handler

	// Opções do handler
	opts := &slog.HandlerOptions{
		Level:     p.level,
		AddSource: config.AddCaller,
	}

	// Writer de saída
	writer := config.Output
	if writer == nil {
		writer = os.Stdout
	}

	// Cria handler baseado no formato
	switch config.Format {
	case interfaces.JSONFormat:
		handler = slog.NewJSONHandler(writer, opts)
	case interfaces.TextFormat, interfaces.ConsoleFormat:
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	// Cria o logger base
	p.logger = slog.New(handler)

	// Adiciona campos globais se existirem
	if len(config.GlobalFields) > 0 {
		args := make([]any, 0, len(config.GlobalFields)*2)
		for key, value := range config.GlobalFields {
			args = append(args, key, value)
		}
		p.logger = p.logger.With(args...)
	}

	// Adiciona informações do serviço
	if config.ServiceName != "" {
		p.logger = p.logger.With(slog.String("service", config.ServiceName))
	}
	if config.ServiceVersion != "" {
		p.logger = p.logger.With(slog.String("version", config.ServiceVersion))
	}
	if config.Environment != "" {
		p.logger = p.logger.With(slog.String("environment", config.Environment))
	}

	return nil
}

// Name retorna o nome do provider
func (p *Provider) Name() string {
	return ProviderName
}

// Version retorna a versão do provider
func (p *Provider) Version() string {
	return "1.21.0" // Versão mínima do Go que suporta slog
}

// HealthCheck verifica se o provider está funcionando
func (p *Provider) HealthCheck() error {
	if p.logger == nil {
		return fmt.Errorf("slog logger not initialized")
	}
	return nil
}

// Implementação da interface Logger

func (p *Provider) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.Enabled(ctx, slog.LevelDebug-1) { // Trace como debug-1
		attrs := convertFields(fields)
		attrs = append(attrs, extractContextAttrs(ctx)...)
		p.logger.LogAttrs(ctx, slog.LevelDebug-1, msg, attrs...)
	}
}

func (p *Provider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.Enabled(ctx, slog.LevelDebug) {
		attrs := convertFields(fields)
		attrs = append(attrs, extractContextAttrs(ctx)...)
		p.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	}
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.Enabled(ctx, slog.LevelInfo) {
		attrs := convertFields(fields)
		attrs = append(attrs, extractContextAttrs(ctx)...)
		p.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	}
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.Enabled(ctx, slog.LevelWarn) {
		attrs := convertFields(fields)
		attrs = append(attrs, extractContextAttrs(ctx)...)
		p.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	}
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.logger.Enabled(ctx, slog.LevelError) {
		attrs := convertFields(fields)
		attrs = append(attrs, extractContextAttrs(ctx)...)

		// Adiciona stack trace se configurado
		if p.config.AddStacktrace {
			attrs = append(attrs, slog.String("stacktrace", getStackTrace()))
		}

		p.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	}
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	attrs := convertFields(fields)
	attrs = append(attrs, extractContextAttrs(ctx)...)
	attrs = append(attrs, slog.String("stacktrace", getStackTrace()))
	p.logger.LogAttrs(ctx, slog.LevelError+1, msg, attrs...)
	os.Exit(1)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	attrs := convertFields(fields)
	attrs = append(attrs, extractContextAttrs(ctx)...)
	attrs = append(attrs, slog.String("stacktrace", getStackTrace()))
	p.logger.LogAttrs(ctx, slog.LevelError+2, msg, attrs...)
	panic(msg)
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
	args := convertFieldsToArgs(fields)
	return &Provider{
		logger: p.logger.With(args...),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithContext(ctx context.Context) interfaces.Logger {
	contextArgs := extractContextArgs(ctx)
	if len(contextArgs) == 0 {
		return p
	}

	return &Provider{
		logger: p.logger.With(contextArgs...),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithError(err error) interfaces.Logger {
	if err == nil {
		return p
	}

	return &Provider{
		logger: p.logger.With(slog.Any("error", err)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithTraceID(traceID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With(slog.String("trace_id", traceID)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithSpanID(spanID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With(slog.String("span_id", spanID)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) SetLevel(level interfaces.Level) {
	p.level = convertLevel(level)
	// Note: slog não permite mudança dinâmica de nível, requer reconfiguração
}

func (p *Provider) GetLevel() interfaces.Level {
	return convertLevelFromSlog(p.level)
}

func (p *Provider) IsLevelEnabled(level interfaces.Level) bool {
	slogLevel := convertLevel(level)
	return p.logger.Enabled(context.Background(), slogLevel)
}

func (p *Provider) Clone() interfaces.Logger {
	return &Provider{
		logger: p.logger,
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) Flush() error {
	// slog não tem método de flush explícito
	return nil
}

func (p *Provider) Close() error {
	// slog não requer fechamento explícito
	return nil
}

// Funções auxiliares

func convertLevel(level interfaces.Level) slog.Level {
	switch level {
	case interfaces.TraceLevel:
		return slog.LevelDebug - 1
	case interfaces.DebugLevel:
		return slog.LevelDebug
	case interfaces.InfoLevel:
		return slog.LevelInfo
	case interfaces.WarnLevel:
		return slog.LevelWarn
	case interfaces.ErrorLevel:
		return slog.LevelError
	case interfaces.FatalLevel:
		return slog.LevelError + 1
	case interfaces.PanicLevel:
		return slog.LevelError + 2
	default:
		return slog.LevelInfo
	}
}

func convertLevelFromSlog(level slog.Level) interfaces.Level {
	switch {
	case level < slog.LevelDebug:
		return interfaces.TraceLevel
	case level == slog.LevelDebug:
		return interfaces.DebugLevel
	case level == slog.LevelInfo:
		return interfaces.InfoLevel
	case level == slog.LevelWarn:
		return interfaces.WarnLevel
	case level == slog.LevelError:
		return interfaces.ErrorLevel
	case level > slog.LevelError:
		return interfaces.FatalLevel
	default:
		return interfaces.InfoLevel
	}
}

func convertFields(fields []interfaces.Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, field := range fields {
		attrs[i] = convertField(field)
	}
	return attrs
}

func convertField(field interfaces.Field) slog.Attr {
	switch field.Type {
	case interfaces.StringType:
		return slog.String(field.Key, field.Value.(string))
	case interfaces.IntType:
		return slog.Int(field.Key, field.Value.(int))
	case interfaces.Int64Type:
		return slog.Int64(field.Key, field.Value.(int64))
	case interfaces.Float64Type:
		return slog.Float64(field.Key, field.Value.(float64))
	case interfaces.BoolType:
		return slog.Bool(field.Key, field.Value.(bool))
	case interfaces.TimeType:
		return slog.Time(field.Key, field.Value.(time.Time))
	case interfaces.DurationType:
		return slog.Duration(field.Key, field.Value.(time.Duration))
	case interfaces.ErrorType:
		if field.Value == nil {
			return slog.Any(field.Key, nil)
		}
		if err, ok := field.Value.(error); ok {
			return slog.Any(field.Key, err)
		}
		if str, ok := field.Value.(string); ok {
			return slog.String(field.Key, str)
		}
		return slog.Any(field.Key, field.Value)
	default:
		return slog.Any(field.Key, field.Value)
	}
}

func convertFieldsToArgs(fields []interfaces.Field) []any {
	args := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		args = append(args, field.Key, convertFieldValue(field))
	}
	return args
}

func convertFieldValue(field interfaces.Field) any {
	switch field.Type {
	case interfaces.ErrorType:
		if field.Value == nil {
			return nil
		}
		if err, ok := field.Value.(error); ok {
			return err
		}
		return field.Value
	default:
		return field.Value
	}
}

func extractContextAttrs(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}

	var attrs []slog.Attr

	// Extrai trace ID
	if traceID := extractTraceID(ctx); traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}

	// Extrai span ID
	if spanID := extractSpanID(ctx); spanID != "" {
		attrs = append(attrs, slog.String("span_id", spanID))
	}

	// Extrai user ID
	if userID := extractUserID(ctx); userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	// Extrai request ID
	if requestID := extractRequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	return attrs
}

func extractContextArgs(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	var args []any

	// Extrai trace ID
	if traceID := extractTraceID(ctx); traceID != "" {
		args = append(args, "trace_id", traceID)
	}

	// Extrai span ID
	if spanID := extractSpanID(ctx); spanID != "" {
		args = append(args, "span_id", spanID)
	}

	// Extrai user ID
	if userID := extractUserID(ctx); userID != "" {
		args = append(args, "user_id", userID)
	}

	// Extrai request ID
	if requestID := extractRequestID(ctx); requestID != "" {
		args = append(args, "request_id", requestID)
	}

	return args
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

func getStackTrace() string {
	// Implementação simples de stack trace
	// Em produção, pode usar bibliotecas mais sofisticadas
	return "stack trace not implemented for slog"
}
