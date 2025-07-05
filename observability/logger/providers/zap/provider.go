package zap

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ProviderName = "zap"

// Provider implementação do Zap para o sistema de logging
type Provider struct {
	logger        *zap.Logger
	sugar         *zap.SugaredLogger
	config        *logger.Config
	level         zap.AtomicLevel
	mu            sync.RWMutex
	fields        []zap.Field
	contextFields map[string]any
}

// NewProvider cria uma nova instância do provider Zap
func NewProvider() *Provider {
	return &Provider{
		level:         zap.NewAtomicLevel(),
		contextFields: make(map[string]any),
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config *logger.Config) error {
	p.config = config

	// Configura o nível
	zapLevel := p.convertLevel(config.Level)
	p.level.SetLevel(zapLevel)

	// Configura o encoder
	encoderConfig := p.buildEncoderConfig()
	var encoder zapcore.Encoder

	switch config.Format {
	case logger.JSONFormat:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case logger.ConsoleFormat, logger.TextFormat:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configura o writer
	writer := zapcore.AddSync(config.Output)
	if writer == nil {
		writer = zapcore.AddSync(os.Stdout)
	}

	// Cria o core
	core := zapcore.NewCore(encoder, writer, p.level)

	// Configura sampling se especificado
	if config.SamplingConfig != nil {
		samplingConfig := &zap.SamplingConfig{
			Initial:    config.SamplingConfig.Initial,
			Thereafter: config.SamplingConfig.Thereafter,
			Hook:       nil,
		}
		core = zapcore.NewSamplerWithOptions(core, config.SamplingConfig.Tick, samplingConfig.Initial, samplingConfig.Thereafter)
	}

	// Opções adicionais
	opts := []zap.Option{}

	if config.AddSource {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	if config.AddStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Adiciona campos globais
	var globalFields []zap.Field
	if config.ServiceName != "" {
		globalFields = append(globalFields, zap.String("service", config.ServiceName))
	}
	if config.ServiceVersion != "" {
		globalFields = append(globalFields, zap.String("version", config.ServiceVersion))
	}
	if config.Environment != "" {
		globalFields = append(globalFields, zap.String("environment", config.Environment))
	}

	// Adiciona campos customizados
	for k, v := range config.Fields {
		globalFields = append(globalFields, zap.Any(k, v))
	}

	// Cria o logger
	p.logger = zap.New(core, opts...).With(globalFields...)
	p.sugar = p.logger.Sugar()
	p.fields = globalFields

	return nil
}

// buildEncoderConfig constrói a configuração do encoder
func (p *Provider) buildEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()

	if p.config.TimeFormat != "" {
		config.TimeKey = "timestamp"
		config.EncodeTime = zapcore.TimeEncoderOfLayout(p.config.TimeFormat)
	}

	config.LevelKey = "level"
	config.MessageKey = "message"
	config.CallerKey = "caller"
	config.StacktraceKey = "stacktrace"
	config.EncodeLevel = zapcore.LowercaseLevelEncoder

	if p.config.Format == logger.ConsoleFormat {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return config
}

// convertLevel converte o nível interno para o nível do Zap
func (p *Provider) convertLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	case logger.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertFields converte os campos internos para campos do Zap
func (p *Provider) convertFields(fields []logger.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// extractContext extrai informações relevantes do contexto
func (p *Provider) extractContext(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// Adiciona trace ID se disponível
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			fields = append(fields, zap.String("trace_id", id))
		}
	}

	// Adiciona span ID se disponível
	if spanID := ctx.Value("span_id"); spanID != nil {
		if id, ok := spanID.(string); ok && id != "" {
			fields = append(fields, zap.String("span_id", id))
		}
	}

	// Adiciona user ID se disponível
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			fields = append(fields, zap.String("user_id", id))
		}
	}

	// Adiciona request ID se disponível
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			fields = append(fields, zap.String("request_id", id))
		}
	}

	return fields
}

// Implementação da interface Logger
func (p *Provider) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Debug(msg, allFields...)
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Info(msg, allFields...)
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Warn(msg, allFields...)
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Error(msg, allFields...)
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Fatal(msg, allFields...)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...logger.Field) {
	contextFields := p.extractContext(ctx)
	convertedFields := p.convertFields(fields)

	p.mu.RLock()
	allFields := append(contextFields, convertedFields...)
	allFields = append(allFields, p.fields...)
	p.mu.RUnlock()

	p.logger.Panic(msg, allFields...)
}

func (p *Provider) Debugf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Debugf(format, args...)
	} else {
		p.sugar.Debugf(format, args...)
	}
}

func (p *Provider) Infof(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Infof(format, args...)
	} else {
		p.sugar.Infof(format, args...)
	}
}

func (p *Provider) Warnf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Warnf(format, args...)
	} else {
		p.sugar.Warnf(format, args...)
	}
}

func (p *Provider) Errorf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Errorf(format, args...)
	} else {
		p.sugar.Errorf(format, args...)
	}
}

func (p *Provider) Fatalf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Fatalf(format, args...)
	} else {
		p.sugar.Fatalf(format, args...)
	}
}

func (p *Provider) Panicf(ctx context.Context, format string, args ...any) {
	contextFields := p.extractContext(ctx)
	if len(contextFields) > 0 {
		p.logger.With(contextFields...).Sugar().Panicf(format, args...)
	} else {
		p.sugar.Panicf(format, args...)
	}
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
	newProvider := &Provider{
		logger:        p.logger,
		sugar:         p.sugar,
		config:        p.config,
		level:         p.level,
		contextFields: make(map[string]any),
	}

	p.mu.RLock()
	// Copia campos existentes de forma thread-safe
	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}

	// Copia fields existentes
	existingFields := make([]zap.Field, len(p.fields))
	copy(existingFields, p.fields)
	p.mu.RUnlock()

	// Adiciona novos campos
	newFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		newFields[i] = zap.Any(field.Key, field.Value)
		newProvider.contextFields[field.Key] = field.Value
	}

	newProvider.fields = append(existingFields, newFields...)
	newProvider.logger = p.logger.With(newFields...)
	newProvider.sugar = newProvider.logger.Sugar()

	return newProvider
}

func (p *Provider) WithContext(ctx context.Context) logger.Logger {
	contextFields := p.extractContext(ctx)
	if len(contextFields) == 0 {
		return p
	}

	newProvider := &Provider{
		logger:        p.logger.With(contextFields...),
		config:        p.config,
		level:         p.level,
		fields:        append(p.fields, contextFields...),
		contextFields: make(map[string]any),
	}

	// Copia campos existentes
	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}

	newProvider.sugar = newProvider.logger.Sugar()
	return newProvider
}

func (p *Provider) SetLevel(level logger.Level) {
	zapLevel := p.convertLevel(level)
	p.level.SetLevel(zapLevel)
}

func (p *Provider) GetLevel() logger.Level {
	switch p.level.Level() {
	case zapcore.DebugLevel:
		return logger.DebugLevel
	case zapcore.InfoLevel:
		return logger.InfoLevel
	case zapcore.WarnLevel:
		return logger.WarnLevel
	case zapcore.ErrorLevel:
		return logger.ErrorLevel
	case zapcore.FatalLevel:
		return logger.FatalLevel
	case zapcore.PanicLevel:
		return logger.PanicLevel
	default:
		return logger.InfoLevel
	}
}

func (p *Provider) Clone() logger.Logger {
	newProvider := &Provider{
		logger:        p.logger,
		sugar:         p.sugar,
		config:        p.config,
		level:         p.level,
		contextFields: make(map[string]any),
	}

	p.mu.RLock()
	// Copia campos de forma thread-safe
	newProvider.fields = make([]zap.Field, len(p.fields))
	copy(newProvider.fields, p.fields)

	for k, v := range p.contextFields {
		newProvider.contextFields[k] = v
	}
	p.mu.RUnlock()

	return newProvider
}

func (p *Provider) Close() error {
	if p.logger != nil {
		return p.logger.Sync()
	}
	return nil
}

// GetZapLogger retorna o logger Zap subjacente para uso avançado
func (p *Provider) GetZapLogger() *zap.Logger {
	return p.logger
}

// GetSugaredLogger retorna o logger Zap sugared para uso avançado
func (p *Provider) GetSugaredLogger() *zap.SugaredLogger {
	return p.sugar
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

// init registra automaticamente o provider
func init() {
	logger.RegisterProvider(ProviderName, NewProvider())
}
