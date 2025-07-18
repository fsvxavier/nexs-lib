package slog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
)

// Provider implementa o provider de logging usando slog
type Provider struct {
	config *logger.Config
	logger *slog.Logger
	level  slog.Level
}

// NewProvider cria uma nova instância do provider slog
func NewProvider() *Provider {
	return &Provider{}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Mapeia os níveis
	switch config.Level {
	case logger.DebugLevel:
		p.level = slog.LevelDebug
	case logger.InfoLevel:
		p.level = slog.LevelInfo
	case logger.WarnLevel:
		p.level = slog.LevelWarn
	case logger.ErrorLevel:
		p.level = slog.LevelError
	case logger.FatalLevel:
		p.level = slog.LevelError + 4
	case logger.PanicLevel:
		p.level = slog.LevelError + 8
	default:
		p.level = slog.LevelInfo
	}

	// Configura o handler
	var handler slog.Handler
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

	opts := &slog.HandlerOptions{
		Level:     p.level,
		AddSource: config.AddSource,
	}

	if config.TimeFormat != "" {
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format(config.TimeFormat))
				}
			}
			return a
		}
	}

	switch config.Format {
	case logger.JSONFormat:
		handler = slog.NewJSONHandler(writer, opts)
	case logger.ConsoleFormat, logger.TextFormat:
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	// Cria o logger base
	p.logger = slog.New(handler)

	// Adiciona campos globais se especificados
	var globalAttrs []slog.Attr

	if config.ServiceName != "" {
		globalAttrs = append(globalAttrs, slog.String("service", config.ServiceName))
	}
	if config.ServiceVersion != "" {
		globalAttrs = append(globalAttrs, slog.String("version", config.ServiceVersion))
	}
	if config.Environment != "" {
		globalAttrs = append(globalAttrs, slog.String("environment", config.Environment))
	}

	// Adiciona campos customizados
	for k, v := range config.Fields {
		globalAttrs = append(globalAttrs, slog.Any(k, v))
	}

	if len(globalAttrs) > 0 {
		// Converte Attr para any
		globalFields := make([]any, len(globalAttrs))
		for i, attr := range globalAttrs {
			globalFields[i] = attr
		}
		p.logger = p.logger.With(globalFields...)
	}

	return nil
}

// extractContextFields extrai campos relevantes do contexto
func (p *Provider) extractContextFields(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	if traceID := ctx.Value(logger.TraceIDKey); traceID != nil {
		attrs = append(attrs, slog.Any(string(logger.TraceIDKey), traceID))
	}

	if spanID := ctx.Value(logger.SpanIDKey); spanID != nil {
		attrs = append(attrs, slog.Any(string(logger.SpanIDKey), spanID))
	}

	if userID := ctx.Value(logger.UserIDKey); userID != nil {
		attrs = append(attrs, slog.Any(string(logger.UserIDKey), userID))
	}

	if requestID := ctx.Value(logger.RequestIDKey); requestID != nil {
		attrs = append(attrs, slog.Any(string(logger.RequestIDKey), requestID))
	}

	return attrs
}

// fieldsToAttrs converte fields para slog.Attr
func (p *Provider) fieldsToAttrs(fields []logger.Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, field := range fields {
		attrs[i] = slog.Any(field.Key, field.Value)
	}
	return attrs
}

// Debug implementa Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Info implementa Logger
func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn implementa Logger
func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error implementa Logger
func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelError) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)

	if p.config != nil && p.config.AddStacktrace {
		attrs = append(attrs, slog.String("stacktrace", "stack_trace"))
	}

	p.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Fatal implementa Logger
func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	attrs = append(attrs, slog.String("stacktrace", "stack_trace"))

	p.logger.LogAttrs(ctx, slog.LevelError+4, msg, attrs...)
	os.Exit(1)
}

// Panic implementa Logger
func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	attrs = append(attrs, slog.String("stacktrace", "stack_trace"))

	p.logger.LogAttrs(ctx, slog.LevelError+8, msg, attrs...)
	panic(msg)
}

// Debugf implementa Logger
func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelDebug, msg, contextAttrs...)
}

// Infof implementa Logger
func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelInfo, msg, contextAttrs...)
}

// Warnf implementa Logger
func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelWarn, msg, contextAttrs...)
}

// Errorf implementa Logger
func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelError) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)

	if p.config != nil && p.config.AddStacktrace {
		contextAttrs = append(contextAttrs, slog.String("stacktrace", "stack_trace"))
	}

	p.logger.LogAttrs(ctx, slog.LevelError, msg, contextAttrs...)
}

// Fatalf implementa Logger
func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	contextAttrs := p.extractContextFields(ctx)
	contextAttrs = append(contextAttrs, slog.String("stacktrace", "stack_trace"))
	msg := fmt.Sprintf(format, args...)

	p.logger.LogAttrs(ctx, slog.LevelError+4, msg, contextAttrs...)
	os.Exit(1)
}

// Panicf implementa Logger
func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	contextAttrs := p.extractContextFields(ctx)
	contextAttrs = append(contextAttrs, slog.String("stacktrace", "stack_trace"))
	msg := fmt.Sprintf(format, args...)

	p.logger.LogAttrs(ctx, slog.LevelError+8, msg, contextAttrs...)
	panic(msg)
}

// DebugWithCode implementa Logger
func (p *Provider) DebugWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, slog.String("code", code))
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// InfoWithCode implementa Logger
func (p *Provider) InfoWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, slog.String("code", code))
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// WarnWithCode implementa Logger
func (p *Provider) WarnWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, slog.String("code", code))
	attrs = append(attrs, p.fieldsToAttrs(fields)...)
	p.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// ErrorWithCode implementa Logger
func (p *Provider) ErrorWithCode(ctx context.Context, code, msg string, fields ...logger.Field) {
	if !p.logger.Enabled(ctx, slog.LevelError) {
		return
	}

	attrs := p.extractContextFields(ctx)
	attrs = append(attrs, slog.String("code", code))
	attrs = append(attrs, p.fieldsToAttrs(fields)...)

	if p.config != nil && p.config.AddStacktrace {
		attrs = append(attrs, slog.String("stacktrace", "stack_trace"))
	}

	p.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// WithFields implementa Logger
func (p *Provider) WithFields(fields ...logger.Field) logger.Logger {
	attrs := p.fieldsToAttrs(fields)
	// Converte Attr para any
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	newLogger := p.logger.With(anyAttrs...)
	return &Provider{
		config: p.config,
		logger: newLogger,
		level:  p.level,
	}
}

// WithContext implementa Logger
func (p *Provider) WithContext(ctx context.Context) logger.Logger {
	attrs := p.extractContextFields(ctx)
	if len(attrs) == 0 {
		return p
	}

	// Converte Attr para any
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	newLogger := p.logger.With(anyAttrs...)
	return &Provider{
		config: p.config,
		logger: newLogger,
		level:  p.level,
	}
}

// SetLevel implementa Logger
func (p *Provider) SetLevel(level logger.Level) {
	switch level {
	case logger.DebugLevel:
		p.level = slog.LevelDebug
	case logger.InfoLevel:
		p.level = slog.LevelInfo
	case logger.WarnLevel:
		p.level = slog.LevelWarn
	case logger.ErrorLevel:
		p.level = slog.LevelError
	case logger.FatalLevel:
		p.level = slog.LevelError + 4
	case logger.PanicLevel:
		p.level = slog.LevelError + 8
	default:
		p.level = slog.LevelInfo
	}
}

// GetLevel implementa Logger
func (p *Provider) GetLevel() logger.Level {
	switch p.level {
	case slog.LevelDebug:
		return logger.DebugLevel
	case slog.LevelInfo:
		return logger.InfoLevel
	case slog.LevelWarn:
		return logger.WarnLevel
	case slog.LevelError:
		return logger.ErrorLevel
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
