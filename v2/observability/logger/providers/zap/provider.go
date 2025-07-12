// Package zap implementa o provider Zap para o sistema de logging v2
package zap

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ProviderName = "zap"

// Provider implementação do Zap seguindo a arquitetura hexagonal
type Provider struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	config interfaces.Config
	level  zap.AtomicLevel
}

// NewProvider cria uma nova instância do provider Zap
func NewProvider() *Provider {
	return &Provider{
		level: zap.NewAtomicLevel(),
	}
}

// Configure configura o provider com as opções fornecidas
func (p *Provider) Configure(config interfaces.Config) error {
	p.config = config

	// Converte o nível
	zapLevel := convertLevel(config.Level)
	p.level.SetLevel(zapLevel)

	// Configuração do encoder
	encoderConfig := buildEncoderConfig(config)

	var encoder zapcore.Encoder
	switch config.Format {
	case interfaces.JSONFormat:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case interfaces.ConsoleFormat:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case interfaces.TextFormat:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configuração do writer
	var writer zapcore.WriteSyncer
	if config.Output != nil {
		writer = zapcore.AddSync(config.Output)
	} else {
		writer = zapcore.AddSync(os.Stdout)
	}

	// Core do zap
	core := zapcore.NewCore(encoder, writer, p.level)

	// Opções do logger
	var options []zap.Option

	if config.AddCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(2)) // Skip core logger calls
	}

	if config.AddStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Campos globais
	if len(config.GlobalFields) > 0 {
		fields := make([]zap.Field, 0, len(config.GlobalFields))
		for key, value := range config.GlobalFields {
			fields = append(fields, zap.Any(key, value))
		}
		options = append(options, zap.Fields(fields...))
	}

	// Cria o logger
	p.logger = zap.New(core, options...)
	p.sugar = p.logger.Sugar()

	return nil
}

// Name retorna o nome do provider
func (p *Provider) Name() string {
	return ProviderName
}

// Version retorna a versão do provider
func (p *Provider) Version() string {
	return "1.27.0" // Versão do Zap
}

// HealthCheck verifica se o provider está funcionando
func (p *Provider) HealthCheck() error {
	if p.logger == nil {
		return fmt.Errorf("zap logger not initialized")
	}
	return nil
}

// Implementação da interface Logger

func (p *Provider) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Debug(msg, zapFields...) // Zap não tem trace nativo, usa debug
}

func (p *Provider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Debug(msg, zapFields...)
}

func (p *Provider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Info(msg, zapFields...)
}

func (p *Provider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Warn(msg, zapFields...)
}

func (p *Provider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Error(msg, zapFields...)
}

func (p *Provider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Fatal(msg, zapFields...)
}

func (p *Provider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	p.logger.Panic(msg, zapFields...)
}

// Métodos com formatação

func (p *Provider) Tracef(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Debugf(format, args...) // Zap não tem trace nativo
}

func (p *Provider) Debugf(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Debugf(format, args...)
}

func (p *Provider) Infof(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Infof(format, args...)
}

func (p *Provider) Warnf(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Warnf(format, args...)
}

func (p *Provider) Errorf(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Errorf(format, args...)
}

func (p *Provider) Fatalf(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Fatalf(format, args...)
}

func (p *Provider) Panicf(ctx context.Context, format string, args ...interface{}) {
	p.sugar.Panicf(format, args...)
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
	zapFields := convertFields(fields)
	sugarArgs := convertFieldsToSugarArgs(fields)

	// Retorna uma nova instância com campos adicionais
	return &Provider{
		logger: p.logger.With(zapFields...),
		sugar:  p.sugar.With(sugarArgs...),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithContext(ctx context.Context) interfaces.Logger {
	contextFields := extractContextFields(ctx)
	if len(contextFields) == 0 {
		return p
	}

	contextArgs := convertZapFieldsToSugarArgs(contextFields)

	return &Provider{
		logger: p.logger.With(contextFields...),
		sugar:  p.sugar.With(contextArgs...),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithError(err error) interfaces.Logger {
	if err == nil {
		return p
	}

	return &Provider{
		logger: p.logger.With(zap.Error(err)),
		sugar:  p.sugar.With(zap.Error(err)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithTraceID(traceID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With(zap.String("trace_id", traceID)),
		sugar:  p.sugar.With(zap.String("trace_id", traceID)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) WithSpanID(spanID string) interfaces.Logger {
	return &Provider{
		logger: p.logger.With(zap.String("span_id", spanID)),
		sugar:  p.sugar.With(zap.String("span_id", spanID)),
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) SetLevel(level interfaces.Level) {
	zapLevel := convertLevel(level)
	p.level.SetLevel(zapLevel)
}

func (p *Provider) GetLevel() interfaces.Level {
	return convertLevelFromZap(p.level.Level())
}

func (p *Provider) IsLevelEnabled(level interfaces.Level) bool {
	zapLevel := convertLevel(level)
	return p.level.Enabled(zapLevel)
}

func (p *Provider) Clone() interfaces.Logger {
	return &Provider{
		logger: p.logger,
		sugar:  p.sugar,
		config: p.config,
		level:  p.level,
	}
}

func (p *Provider) Flush() error {
	return p.logger.Sync()
}

func (p *Provider) Close() error {
	return p.logger.Sync()
}

// Funções de conversão para Sugar logger
func convertFieldsToSugarArgs(fields []interfaces.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, field := range fields {
		args = append(args, field.Key, field.Value)
	}
	return args
}

func convertZapFieldsToSugarArgs(zapFields []zap.Field) []interface{} {
	args := make([]interface{}, 0, len(zapFields)*2)
	for _, field := range zapFields {
		args = append(args, field.Key, getZapFieldValue(field))
	}
	return args
}

func getZapFieldValue(field zap.Field) interface{} {
	switch field.Type {
	case zapcore.StringType:
		return field.String
	case zapcore.Int64Type:
		return field.Integer
	case zapcore.Float64Type:
		return math.Float64frombits(uint64(field.Integer))
	case zapcore.BoolType:
		return field.Integer == 1
	case zapcore.TimeType:
		return time.Unix(0, field.Integer)
	case zapcore.DurationType:
		return time.Duration(field.Integer)
	case zapcore.ErrorType:
		return field.Interface
	default:
		return field.Interface
	}
}

// Funções auxiliares

func convertLevel(level interfaces.Level) zapcore.Level {
	switch level {
	case interfaces.TraceLevel:
		return zapcore.DebugLevel // Zap não tem trace
	case interfaces.DebugLevel:
		return zapcore.DebugLevel
	case interfaces.InfoLevel:
		return zapcore.InfoLevel
	case interfaces.WarnLevel:
		return zapcore.WarnLevel
	case interfaces.ErrorLevel:
		return zapcore.ErrorLevel
	case interfaces.FatalLevel:
		return zapcore.FatalLevel
	case interfaces.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func convertLevelFromZap(level zapcore.Level) interfaces.Level {
	switch level {
	case zapcore.DebugLevel:
		return interfaces.DebugLevel
	case zapcore.InfoLevel:
		return interfaces.InfoLevel
	case zapcore.WarnLevel:
		return interfaces.WarnLevel
	case zapcore.ErrorLevel:
		return interfaces.ErrorLevel
	case zapcore.FatalLevel:
		return interfaces.FatalLevel
	case zapcore.PanicLevel:
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}

func convertFields(fields []interfaces.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = convertField(field)
	}
	return zapFields
}

func convertField(field interfaces.Field) zap.Field {
	switch field.Type {
	case interfaces.StringType:
		return zap.String(field.Key, field.Value.(string))
	case interfaces.IntType:
		return zap.Int(field.Key, field.Value.(int))
	case interfaces.Int64Type:
		return zap.Int64(field.Key, field.Value.(int64))
	case interfaces.Float64Type:
		return zap.Float64(field.Key, field.Value.(float64))
	case interfaces.BoolType:
		return zap.Bool(field.Key, field.Value.(bool))
	case interfaces.TimeType:
		return zap.Time(field.Key, field.Value.(time.Time))
	case interfaces.DurationType:
		return zap.Duration(field.Key, field.Value.(time.Duration))
	case interfaces.ErrorType:
		if err, ok := field.Value.(error); ok {
			return zap.Error(err)
		}
		if str, ok := field.Value.(string); ok {
			return zap.String(field.Key, str)
		}
		return zap.Any(field.Key, field.Value)
	default:
		return zap.Any(field.Key, field.Value)
	}
}

func buildEncoderConfig(config interfaces.Config) zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()

	// Configuração de tempo
	if config.TimeFormat != "" {
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimeFormat)
	}

	// Configuração de nível
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// Configuração de mensagem
	encoderConfig.MessageKey = "message"

	// Configuração de caller
	if config.AddCaller {
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Configuração de stacktrace
	if config.AddStacktrace {
		encoderConfig.StacktraceKey = "stacktrace"
	}

	// Configuração específica para console
	if config.Format == interfaces.ConsoleFormat {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	}

	return encoderConfig
}

func extractContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	var fields []zap.Field

	// Extrai trace ID
	if traceID := extractTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// Extrai span ID
	if spanID := extractSpanID(ctx); spanID != "" {
		fields = append(fields, zap.String("span_id", spanID))
	}

	// Extrai user ID
	if userID := extractUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// Extrai request ID
	if requestID := extractRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	return fields
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
