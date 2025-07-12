package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// asyncProcessor processa logs de forma assíncrona para alta performance
type asyncProcessor struct {
	logger     *CoreLogger
	config     *interfaces.AsyncConfig
	queue      chan *interfaces.Entry
	workers    []worker
	wg         sync.WaitGroup
	stopCh     chan struct{}
	flushTimer *time.Timer
	running    bool
	mu         sync.RWMutex
}

// worker representa um worker para processamento assíncrono
type worker struct {
	id     int
	queue  chan *interfaces.Entry
	stop   chan struct{}
	wg     *sync.WaitGroup
	logger *CoreLogger
}

// sampler implementa sampling para controle de volume de logs
type sampler struct {
	config   *interfaces.SamplingConfig
	counters map[interfaces.Level]*levelCounter
	mu       sync.RWMutex
	ticker   *time.Ticker
	stopCh   chan struct{}
}

// levelCounter contador por nível de log
type levelCounter struct {
	count     int
	threshold int
	mu        sync.Mutex
}

// CoreLogger implementação principal do sistema de logging
type CoreLogger struct {
	config       interfaces.Config
	provider     interfaces.Provider
	level        interfaces.Level
	globalFields []interfaces.Field
	hooks        []interfaces.Hook
	middlewares  []interfaces.Middleware
	metrics      interfaces.MetricsCollector

	// Pools para otimização de performance
	entryPool  sync.Pool
	bufferPool sync.Pool

	// Context fields
	contextFields map[string]interface{}
	mu            sync.RWMutex

	// Async processing
	async *asyncProcessor

	// Sampling
	sampler *sampler
}

// NewCoreLogger cria uma nova instância do CoreLogger
func NewCoreLogger(provider interfaces.Provider, config interfaces.Config) *CoreLogger {
	logger := &CoreLogger{
		config:        config,
		provider:      provider,
		level:         config.Level,
		globalFields:  convertGlobalFields(config.GlobalFields),
		contextFields: make(map[string]interface{}),
		hooks:         config.Hooks,
		middlewares:   config.Middlewares,
	}

	// Inicializa pools para performance
	logger.entryPool.New = func() interface{} {
		return &interfaces.Entry{
			Fields: make([]interfaces.Field, 0, 16),
		}
	}

	logger.bufferPool.New = func() interface{} {
		return make([]byte, 0, 1024)
	}

	// Configuração assíncrona se habilitada
	if config.Async != nil && config.Async.Enabled {
		logger.async = newAsyncProcessor(logger, config.Async)
	}

	// Configuração de sampling se habilitada
	if config.Sampling != nil && config.Sampling.Enabled {
		logger.sampler = newSampler(config.Sampling)
	}

	return logger
}

// Implementação da interface Logger

func (l *CoreLogger) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.TraceLevel) {
		return
	}
	l.log(ctx, interfaces.TraceLevel, msg, fields)
}

func (l *CoreLogger) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.DebugLevel) {
		return
	}
	l.log(ctx, interfaces.DebugLevel, msg, fields)
}

func (l *CoreLogger) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.InfoLevel) {
		return
	}
	l.log(ctx, interfaces.InfoLevel, msg, fields)
}

func (l *CoreLogger) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.WarnLevel) {
		return
	}
	l.log(ctx, interfaces.WarnLevel, msg, fields)
}

func (l *CoreLogger) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.ErrorLevel) {
		return
	}
	l.log(ctx, interfaces.ErrorLevel, msg, fields)
}

func (l *CoreLogger) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	l.log(ctx, interfaces.FatalLevel, msg, fields)
	l.provider.Fatal(ctx, msg, fields...)
}

func (l *CoreLogger) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	l.log(ctx, interfaces.PanicLevel, msg, fields)
	l.provider.Panic(ctx, msg, fields...)
}

// Métodos com formatação

func (l *CoreLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	if !l.IsLevelEnabled(interfaces.TraceLevel) {
		return
	}
	l.log(ctx, interfaces.TraceLevel, fmt.Sprintf(format, args...), nil)
}

func (l *CoreLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if !l.IsLevelEnabled(interfaces.DebugLevel) {
		return
	}
	l.log(ctx, interfaces.DebugLevel, fmt.Sprintf(format, args...), nil)
}

func (l *CoreLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if !l.IsLevelEnabled(interfaces.InfoLevel) {
		return
	}
	l.log(ctx, interfaces.InfoLevel, fmt.Sprintf(format, args...), nil)
}

func (l *CoreLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	if !l.IsLevelEnabled(interfaces.WarnLevel) {
		return
	}
	l.log(ctx, interfaces.WarnLevel, fmt.Sprintf(format, args...), nil)
}

func (l *CoreLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if !l.IsLevelEnabled(interfaces.ErrorLevel) {
		return
	}
	l.log(ctx, interfaces.ErrorLevel, fmt.Sprintf(format, args...), nil)
}

func (l *CoreLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(ctx, interfaces.FatalLevel, msg, nil)
	l.provider.Fatalf(ctx, format, args...)
}

func (l *CoreLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(ctx, interfaces.PanicLevel, msg, nil)
	l.provider.Panicf(ctx, format, args...)
}

// Métodos com código

func (l *CoreLogger) TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.TraceLevel) {
		return
	}
	allFields := append(fields, interfaces.String("code", code))
	l.log(ctx, interfaces.TraceLevel, msg, allFields)
}

func (l *CoreLogger) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.DebugLevel) {
		return
	}
	allFields := append(fields, interfaces.String("code", code))
	l.log(ctx, interfaces.DebugLevel, msg, allFields)
}

func (l *CoreLogger) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.InfoLevel) {
		return
	}
	allFields := append(fields, interfaces.String("code", code))
	l.log(ctx, interfaces.InfoLevel, msg, allFields)
}

func (l *CoreLogger) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.WarnLevel) {
		return
	}
	allFields := append(fields, interfaces.String("code", code))
	l.log(ctx, interfaces.WarnLevel, msg, allFields)
}

func (l *CoreLogger) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if !l.IsLevelEnabled(interfaces.ErrorLevel) {
		return
	}
	allFields := append(fields, interfaces.String("code", code))
	l.log(ctx, interfaces.ErrorLevel, msg, allFields)
}

// Métodos utilitários

func (l *CoreLogger) WithFields(fields ...interfaces.Field) interfaces.Logger {
	clone := l.Clone().(*CoreLogger)
	clone.globalFields = append(clone.globalFields, fields...)
	return clone
}

func (l *CoreLogger) WithContext(ctx context.Context) interfaces.Logger {
	clone := l.Clone().(*CoreLogger)

	// Extrai informações do contexto
	if traceID := extractTraceID(ctx); traceID != "" {
		clone.globalFields = append(clone.globalFields, interfaces.TraceID(traceID))
	}

	if spanID := extractSpanID(ctx); spanID != "" {
		clone.globalFields = append(clone.globalFields, interfaces.SpanID(spanID))
	}

	if userID := extractUserID(ctx); userID != "" {
		clone.globalFields = append(clone.globalFields, interfaces.UserID(userID))
	}

	if requestID := extractRequestID(ctx); requestID != "" {
		clone.globalFields = append(clone.globalFields, interfaces.RequestID(requestID))
	}

	return clone
}

func (l *CoreLogger) WithError(err error) interfaces.Logger {
	if err == nil {
		return l
	}
	return l.WithFields(interfaces.Error(err))
}

func (l *CoreLogger) WithTraceID(traceID string) interfaces.Logger {
	return l.WithFields(interfaces.TraceID(traceID))
}

func (l *CoreLogger) WithSpanID(spanID string) interfaces.Logger {
	return l.WithFields(interfaces.SpanID(spanID))
}

func (l *CoreLogger) SetLevel(level interfaces.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	l.provider.SetLevel(level) // Propaga a mudança para o provider
}

func (l *CoreLogger) GetLevel() interfaces.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

func (l *CoreLogger) IsLevelEnabled(level interfaces.Level) bool {
	return level >= l.GetLevel()
}

func (l *CoreLogger) Clone() interfaces.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	clone := &CoreLogger{
		config:        l.config,
		provider:      l.provider,
		level:         l.level,
		globalFields:  make([]interfaces.Field, len(l.globalFields)),
		hooks:         l.hooks,
		middlewares:   l.middlewares,
		metrics:       l.metrics,
		contextFields: make(map[string]interface{}),
		async:         l.async,
		sampler:       l.sampler,
	}

	copy(clone.globalFields, l.globalFields)

	for k, v := range l.contextFields {
		clone.contextFields[k] = v
	}

	// Compartilha pools entre clones (por referência)
	clone.entryPool.New = l.entryPool.New
	clone.bufferPool.New = l.bufferPool.New

	return clone
}

func (l *CoreLogger) Flush() error {
	if l.async != nil {
		l.async.flush()
	}
	return l.provider.Flush()
}

func (l *CoreLogger) Close() error {
	if l.async != nil {
		l.async.close()
	}
	return l.provider.Close()
}

// Método principal de logging

func (l *CoreLogger) log(ctx context.Context, level interfaces.Level, msg string, fields []interfaces.Field) {
	// Verifica sampling
	if l.sampler != nil && l.sampler.shouldSample(level) {
		return
	}

	// Obtém entry do pool
	entry := l.entryPool.Get().(*interfaces.Entry)
	defer l.entryPool.Put(entry)

	// Reseta entry
	entry.Level = level
	entry.Message = msg
	entry.Time = time.Now()
	entry.Context = ctx
	entry.Fields = entry.Fields[:0]
	entry.TraceID = extractTraceID(ctx)
	entry.SpanID = extractSpanID(ctx)

	// Adiciona campos globais
	entry.Fields = append(entry.Fields, l.globalFields...)

	// Adiciona campos da chamada
	if fields != nil {
		entry.Fields = append(entry.Fields, fields...)
	}

	// Adiciona caller info se configurado
	if l.config.AddCaller {
		if caller := getCaller(3); caller != nil {
			entry.Caller = caller
		}
	}

	// Adiciona stack trace se configurado e nível apropriado
	if l.config.AddStacktrace && level >= interfaces.ErrorLevel {
		entry.StackTrace = getStackTrace(3)
	}

	// Aplica middlewares
	for _, middleware := range l.middlewares {
		entry = middleware.Process(entry)
		if entry == nil {
			return // Middleware cancelou o log
		}
	}

	// Aplica hooks
	for _, hook := range l.hooks {
		if containsLevel(hook.Levels(), level) {
			if err := hook.Fire(entry); err != nil {
				// Log hook error mas não falha o log principal
				fmt.Printf("Hook error: %v\n", err)
			}
		}
	}

	// Coleta métricas
	if l.metrics != nil {
		l.metrics.IncrementCounter("logs_total", map[string]string{
			"level":   level.String(),
			"service": l.config.ServiceName,
		})
	}

	// Processa de forma assíncrona ou síncrona
	if l.async != nil {
		l.async.process(entry)
	} else {
		l.processEntry(entry)
	}
}

func (l *CoreLogger) processEntry(entry *interfaces.Entry) {
	// Delega para o provider específico
	switch entry.Level {
	case interfaces.TraceLevel:
		l.provider.Trace(entry.Context, entry.Message, entry.Fields...)
	case interfaces.DebugLevel:
		l.provider.Debug(entry.Context, entry.Message, entry.Fields...)
	case interfaces.InfoLevel:
		l.provider.Info(entry.Context, entry.Message, entry.Fields...)
	case interfaces.WarnLevel:
		l.provider.Warn(entry.Context, entry.Message, entry.Fields...)
	case interfaces.ErrorLevel:
		l.provider.Error(entry.Context, entry.Message, entry.Fields...)
	case interfaces.FatalLevel:
		l.provider.Fatal(entry.Context, entry.Message, entry.Fields...)
	case interfaces.PanicLevel:
		l.provider.Panic(entry.Context, entry.Message, entry.Fields...)
	}
}

// Funções auxiliares

func convertGlobalFields(fields map[string]interface{}) []interfaces.Field {
	result := make([]interfaces.Field, 0, len(fields))
	for k, v := range fields {
		result = append(result, interfaces.Object(k, v))
	}
	return result
}

func getCaller(skip int) *interfaces.Caller {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	function := runtime.FuncForPC(pc)
	var funcName string
	if function != nil {
		funcName = function.Name()
		// Remove package path, keep only function name
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
	}

	// Remove full path, keep only filename
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}

	return &interfaces.Caller{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

func getStackTrace(skip int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func containsLevel(levels []interfaces.Level, target interfaces.Level) bool {
	for _, level := range levels {
		if level == target {
			return true
		}
	}
	return false
}

// Context extractors - implementações podem variar baseado no sistema de tracing usado

func extractTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Tenta extrair de diferentes sources comuns
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

// Implementações dos tipos auxiliares

// newAsyncProcessor cria um novo processador assíncrono
func newAsyncProcessor(logger *CoreLogger, config *interfaces.AsyncConfig) *asyncProcessor {
	processor := &asyncProcessor{
		logger:  logger,
		config:  config,
		queue:   make(chan *interfaces.Entry, config.BufferSize),
		workers: make([]worker, config.Workers),
		stopCh:  make(chan struct{}),
		running: true,
	}

	// Inicia workers
	for i := 0; i < config.Workers; i++ {
		processor.workers[i] = worker{
			id:     i,
			queue:  processor.queue,
			stop:   processor.stopCh,
			wg:     &processor.wg,
			logger: logger,
		}
		processor.wg.Add(1)
		go processor.workers[i].run()
	}

	// Inicia timer de flush se configurado
	if config.FlushInterval > 0 {
		processor.flushTimer = time.NewTimer(config.FlushInterval)
		go processor.flushRoutine()
	}

	return processor
}

// process envia uma entry para processamento assíncrono
func (p *asyncProcessor) process(entry *interfaces.Entry) {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		// Se não está rodando, processa sincronamente
		p.logger.processEntry(entry)
		return
	}
	p.mu.RUnlock()

	// Clona a entry para evitar race conditions
	clonedEntry := p.cloneEntry(entry)

	select {
	case p.queue <- clonedEntry:
		// Entry enviada com sucesso
	default:
		// Queue está cheia
		if p.config.DropOnFull {
			// Simplesmente descarta o log
			return
		} else {
			// Processa sincronamente como fallback
			p.logger.processEntry(entry)
		}
	}
}

// flush força o flush de todos os logs pendentes
func (p *asyncProcessor) flush() {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return
	}
	p.mu.RUnlock()

	// Aguarda queue esvaziar
	for len(p.queue) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

// close para o processador assíncrono
func (p *asyncProcessor) close() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	// Para o timer de flush
	if p.flushTimer != nil {
		p.flushTimer.Stop()
	}

	// Sinaliza parada para todos os workers
	close(p.stopCh)

	// Aguarda todos os workers terminarem
	p.wg.Wait()

	// Processa items restantes na queue sincronamente
	close(p.queue)
	for entry := range p.queue {
		p.logger.processEntry(entry)
	}
}

// flushRoutine rotina de flush periódico
func (p *asyncProcessor) flushRoutine() {
	for {
		select {
		case <-p.flushTimer.C:
			p.flush()
			p.flushTimer.Reset(p.config.FlushInterval)
		case <-p.stopCh:
			return
		}
	}
}

// cloneEntry clona uma entry para evitar race conditions
func (p *asyncProcessor) cloneEntry(entry *interfaces.Entry) *interfaces.Entry {
	cloned := &interfaces.Entry{
		Level:      entry.Level,
		Message:    entry.Message,
		Time:       entry.Time,
		Context:    entry.Context,
		TraceID:    entry.TraceID,
		SpanID:     entry.SpanID,
		StackTrace: entry.StackTrace,
		Fields:     make([]interfaces.Field, len(entry.Fields)),
	}

	copy(cloned.Fields, entry.Fields)

	if entry.Caller != nil {
		cloned.Caller = &interfaces.Caller{
			File:     entry.Caller.File,
			Line:     entry.Caller.Line,
			Function: entry.Caller.Function,
		}
	}

	return cloned
}

// run executa o loop principal do worker
func (w *worker) run() {
	defer w.wg.Done()

	for {
		select {
		case entry := <-w.queue:
			if entry != nil {
				w.logger.processEntry(entry)
			}
		case <-w.stop:
			// Processa items restantes antes de parar
			for {
				select {
				case entry := <-w.queue:
					if entry != nil {
						w.logger.processEntry(entry)
					}
				default:
					return
				}
			}
		}
	}
}

// newSampler cria um novo sampler
func newSampler(config *interfaces.SamplingConfig) *sampler {
	s := &sampler{
		config:   config,
		counters: make(map[interfaces.Level]*levelCounter),
		stopCh:   make(chan struct{}),
	}

	// Inicializa contadores para os níveis configurados
	for _, level := range config.Levels {
		s.counters[level] = &levelCounter{
			threshold: config.Initial,
		}
	}

	// Inicia ticker para reset periódico
	if config.Tick > 0 {
		s.ticker = time.NewTicker(config.Tick)
		go s.resetCounters()
	}

	return s
}

// shouldSample determina se um log deve ser amostrado (descartado)
func (s *sampler) shouldSample(level interfaces.Level) bool {
	s.mu.RLock()
	counter, exists := s.counters[level]
	s.mu.RUnlock()

	if !exists {
		// Nível não está sendo amostrado
		return false
	}

	counter.mu.Lock()
	defer counter.mu.Unlock()

	counter.count++

	if counter.count <= s.config.Initial {
		// Permite os primeiros logs
		return false
	}

	// Depois do initial, só permite 1 a cada 'thereafter'
	if (counter.count-s.config.Initial)%s.config.Thereafter == 0 {
		return false
	}

	// Amostra (descarta) este log
	return true
}

// resetCounters reseta os contadores periodicamente
func (s *sampler) resetCounters() {
	for {
		select {
		case <-s.ticker.C:
			s.mu.RLock()
			for _, counter := range s.counters {
				counter.mu.Lock()
				counter.count = 0
				counter.mu.Unlock()
			}
			s.mu.RUnlock()
		case <-s.stopCh:
			return
		}
	}
}

// close para o sampler
func (s *sampler) close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)
}
