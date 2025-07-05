package slog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
)

const ProviderName = "slog"

// Provider implementação do Slog para o sistema de logging
type Provider struct {
	logger        *slog.Logger
	config        *logger.Config
	level         slog.Level
	mu            sync.RWMutex
	contextFields map[string]any
	handler       slog.Handler
}

// NewProvider cria uma nova instância do provider Slog
func NewProvider() *Provider {
	return &Provider{
		level:         slog.LevelInfo,
		contextFields: make(map[string]any),
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Configura o nível
	slogLevel := p.convertLevel(config.Level)
	p.level = slogLevel

	// Configura o writer
	var writer io.Writer = config.Output
	if writer == nil {
		writer = os.Stdout
	}

	// Configura as opções do handler
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: config.AddSource,
	}

	// Configura o formato de tempo se especificado
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

	// Cria o handler baseado no formato
	var handler slog.Handler
	switch config.Format {
	case logger.JSONFormat:
		handler = slog.NewJSONHandler(writer, opts)
	case logger.ConsoleFormat, logger.TextFormat:
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	p.handler = handler

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
		// Converte []slog.Attr para []any
		args := make([]any, len(globalAttrs))
		for i, attr := range globalAttrs {
			args[i] = attr
		}
		p.logger = p.logger.With(args...)
	}

	// Define como logger padrão do slog
	slog.SetDefault(p.logger)

	return nil
}

// convertLevel converte o nível interno para o nível do Slog
func (p *Provider) convertLevel(level logger.Level) slog.Level {
	switch level {
	case logger.DebugLevel:
		return slog.LevelDebug
	case logger.InfoLevel:
		return slog.LevelInfo
	case logger.WarnLevel:
		return slog.LevelWarn
	case logger.ErrorLevel:
		return slog.LevelError
	case logger.FatalLevel:
		return slog.LevelError + 4 // Slog não tem Fatal, usa Error+4
	case logger.PanicLevel:
		return slog.LevelError + 8 // Slog não tem Panic, usa Error+8
	default:
		return slog.LevelInfo
	}
}

// convertFields converte os campos internos para atributos do Slog
func (p *Provider) convertFields(fields []logger.Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, field := range fields {
		attrs[i] = slog.Any(field.Key, field.Value)
	}
	return attrs
}

// extractContext extrai informações relevantes do contexto
func (p *Provider) extractContextFields(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	// Adiciona trace ID se disponível
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			attrs = append(attrs, slog.String("trace_id", id))
		}
	}

	// Adiciona span ID se disponível
	if spanID := ctx.Value("span_id"); spanID != nil {
		if id, ok := spanID.(string); ok && id != "" {
			attrs = append(attrs, slog.String("span_id", id))
		}
	}

	// Adiciona user ID se disponível
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			attrs = append(attrs, slog.String("user_id", id))
		}
	}

	// Adiciona request ID se disponível
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			attrs = append(attrs, slog.String("request_id", id))
		}
	}

	return attrs
}

// logWithContext helper para adicionar stack trace se necessário
func (p *Provider) logWithContext(ctx context.Context, level slog.Level, msg string, fields []logger.Field) {
	if !p.logger.Enabled(ctx, level) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	fieldAttrs := p.convertFields(fields)
	allAttrs := append(contextAttrs, fieldAttrs...)

	// Adiciona stack trace se necessário
	if p.config != nil && p.config.AddStacktrace && level >= slog.LevelError {
		allAttrs = append(allAttrs, slog.String("stack_trace", p.getStackTrace(3)))
	}

	p.logger.LogAttrs(ctx, level, msg, allAttrs...)
}

// getStackTrace captura o stack trace atual
func (p *Provider) getStackTrace(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack []string
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return strings.Join(stack, "\n")
}

// Implementação da interface Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	p.logWithContext(ctx, slog.LevelDebug, msg, fields)
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	p.logWithContext(ctx, slog.LevelInfo, msg, fields)
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	p.logWithContext(ctx, slog.LevelWarn, msg, fields)
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	p.logWithContext(ctx, slog.LevelError, msg, fields)
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	fatalLevel := slog.LevelError + 4
	p.logWithContext(ctx, fatalLevel, msg, fields)
	os.Exit(1)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	panicLevel := slog.LevelError + 8
	p.logWithContext(ctx, panicLevel, msg, fields)
	panic(msg)
}

func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelDebug, msg, contextAttrs...)
}

func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelInfo, msg, contextAttrs...)
}

func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, slog.LevelWarn, msg, contextAttrs...)
}

func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	if !p.logger.Enabled(ctx, slog.LevelError) {
		return
	}

	contextAttrs := p.extractContextFields(ctx)
	msg := fmt.Sprintf(format, args...)

	// Adiciona stack trace se necessário
	if p.config != nil && p.config.AddStacktrace {
		contextAttrs = append(contextAttrs, slog.String("stack_trace", p.getStackTrace(2)))
	}

	p.logger.LogAttrs(ctx, slog.LevelError, msg, contextAttrs...)
}

func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	fatalLevel := slog.LevelError + 4
	contextAttrs := p.extractContextFields(ctx)
	contextAttrs = append(contextAttrs, slog.String("stack_trace", p.getStackTrace(2)))
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, fatalLevel, msg, contextAttrs...)
	os.Exit(1)
}

func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	panicLevel := slog.LevelError + 8
	contextAttrs := p.extractContextFields(ctx)
	contextAttrs = append(contextAttrs, slog.String("stack_trace", p.getStackTrace(2)))
	msg := fmt.Sprintf(format, args...)
	p.logger.LogAttrs(ctx, panicLevel, msg, contextAttrs...)
	panic(msg)
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
	attrs := p.convertFields(fields)

	// Converte []slog.Attr para []any
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	newProvider := &Provider{
		logger:        p.logger.With(args...),
		config:        p.config,
		level:         p.level,
		handler:       p.handler,
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
	contextAttrs := p.extractContextFields(ctx)
	if len(contextAttrs) == 0 {
		return p
	}

	// Converte []slog.Attr para []any
	args := make([]any, len(contextAttrs))
	for i, attr := range contextAttrs {
		args[i] = attr
	}

	newProvider := &Provider{
		logger:        p.logger.With(args...),
		config:        p.config,
		level:         p.level,
		handler:       p.handler,
		contextFields: make(map[string]any),
	}

	p.mu.RLock()
	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}
	p.mu.RUnlock()

	return newProvider
}

func (p *Provider) SetLevel(level logger.Level) {
	slogLevel := p.convertLevel(level)
	p.level = slogLevel

	// Cria um novo handler com o nível atualizado
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: p.config != nil && p.config.AddSource,
	}

	var writer io.Writer = os.Stdout
	if p.config != nil && p.config.Output != nil {
		writer = p.config.Output
	}

	var handler slog.Handler
	if p.config != nil {
		switch p.config.Format {
		case logger.JSONFormat:
			handler = slog.NewJSONHandler(writer, opts)
		case logger.ConsoleFormat, logger.TextFormat:
			handler = slog.NewTextHandler(writer, opts)
		default:
			handler = slog.NewJSONHandler(writer, opts)
		}
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	p.handler = handler
	p.logger = slog.New(handler)
}

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
		if p.level >= slog.LevelError+4 {
			return logger.FatalLevel
		}
		return logger.InfoLevel
	}
}

func (p *Provider) Clone() logger.Logger {
	newProvider := &Provider{
		logger:        p.logger,
		config:        p.config,
		level:         p.level,
		handler:       p.handler,
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
	// Slog não tem método de close explícito
	return nil
}

// GetSlogLogger retorna o logger Slog subjacente para uso avançado
func (p *Provider) GetSlogLogger() *slog.Logger {
	return p.logger
}

// GetHandler retorna o handler Slog subjacente para uso avançado
func (p *Provider) GetHandler() slog.Handler {
	return p.handler
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

// CustomHandler permite criar handlers personalizados
type CustomHandler struct {
	slog.Handler
	attrs []slog.Attr
}

func NewCustomHandler(handler slog.Handler) *CustomHandler {
	return &CustomHandler{Handler: handler}
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// Adiciona hostname se disponível
	if hostname, err := os.Hostname(); err == nil {
		r.AddAttrs(slog.String("hostname", hostname))
	}

	// Adiciona PID
	r.AddAttrs(slog.Int("pid", os.Getpid()))

	return h.Handler.Handle(ctx, r)
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		Handler: h.Handler.WithAttrs(attrs),
		attrs:   append(h.attrs, attrs...),
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		Handler: h.Handler.WithGroup(name),
		attrs:   h.attrs,
	}
}

// init registra automaticamente o provider
func init() {
	logger.RegisterProvider(ProviderName, NewProvider())
}
