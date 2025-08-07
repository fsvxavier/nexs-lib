package performance

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// LazyStackTrace implementa lazy evaluation de stack traces para performance
type LazyStackTrace struct {
	frames          []interfaces.StackFrame
	captured        bool
	skipLevel       int
	mu              sync.RWMutex
	programCounters []uintptr
}

// NewLazyStackTrace cria nova instância de lazy stack trace
func NewLazyStackTrace(skipLevel int) *LazyStackTrace {
	lst := &LazyStackTrace{
		skipLevel: skipLevel + 1, // +1 para pular esta própria função
		captured:  false,
	}

	// Captura apenas program counters inicialmente (muito mais rápido)
	lst.captureProgramCounters()

	return lst
}

// captureProgramCounters captura apenas os program counters (muito rápido)
func (lst *LazyStackTrace) captureProgramCounters() {
	// Pré-aloca slice com capacidade suficiente para maioria dos casos
	pcs := make([]uintptr, 32)
	n := runtime.Callers(lst.skipLevel, pcs)
	lst.programCounters = pcs[:n]
}

// GetFrames retorna frames do stack trace, capturando detalhes se necessário
func (lst *LazyStackTrace) GetFrames() []interfaces.StackFrame {
	lst.mu.Lock()
	defer lst.mu.Unlock()

	if !lst.captured && len(lst.programCounters) > 0 {
		lst.frames = lst.captureStackTrace()
		lst.captured = true
	}

	return lst.frames
}

// captureStackTrace converte program counters em stack frames detalhados
func (lst *LazyStackTrace) captureStackTrace() []interfaces.StackFrame {
	if len(lst.programCounters) == 0 {
		return nil
	}

	frames := make([]interfaces.StackFrame, 0, len(lst.programCounters))
	runtimeFrames := runtime.CallersFrames(lst.programCounters)

	for {
		frame, more := runtimeFrames.Next()

		frames = append(frames, interfaces.StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
			Message:  "", // Será preenchido se necessário
			Time:     "", // Pode ser adicionado se necessário
		})

		if !more {
			break
		}
	}

	return frames
}

// String implementa formatação eficiente de stack trace
func (lst *LazyStackTrace) String() string {
	frames := lst.GetFrames()
	if len(frames) == 0 {
		return ""
	}

	// Use strings.Builder para concatenação eficiente
	var builder strings.Builder
	builder.Grow(len(frames) * 80) // Pré-aloca espaço estimado

	for i, frame := range frames {
		if i > 0 {
			builder.WriteByte('\n')
		}

		builder.WriteString(frame.Function)
		builder.WriteString(" at ")
		builder.WriteString(frame.File)
		builder.WriteByte(':')
		builder.WriteString(strconv.Itoa(frame.Line))
	}

	return builder.String()
}

// HasFrames verifica se há frames disponíveis sem capturar detalhes
func (lst *LazyStackTrace) HasFrames() bool {
	lst.mu.RLock()
	defer lst.mu.RUnlock()
	return len(lst.programCounters) > 0
}

// FrameCount retorna número de frames sem capturar detalhes
func (lst *LazyStackTrace) FrameCount() int {
	lst.mu.RLock()
	defer lst.mu.RUnlock()
	return len(lst.programCounters)
}

// IsCaptured verifica se detalhes já foram capturados
func (lst *LazyStackTrace) IsCaptured() bool {
	lst.mu.RLock()
	defer lst.mu.RUnlock()
	return lst.captured
}

// Reset reseta o lazy stack trace para reutilização
func (lst *LazyStackTrace) Reset(skipLevel int) {
	lst.mu.Lock()
	defer lst.mu.Unlock()

	lst.frames = lst.frames[:0]
	lst.programCounters = lst.programCounters[:0]
	lst.captured = false
	lst.skipLevel = skipLevel + 1

	// Recaptura program counters
	lst.captureProgramCounters()
}

// StackTracePool pool para reutilizar instâncias de LazyStackTrace
type StackTracePool struct {
	pool sync.Pool
}

// NewStackTracePool cria novo pool de stack traces
func NewStackTracePool() *StackTracePool {
	return &StackTracePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &LazyStackTrace{
					frames:          make([]interfaces.StackFrame, 0, 16),
					programCounters: make([]uintptr, 0, 32),
				}
			},
		},
	}
}

// Get obtém LazyStackTrace do pool
func (stp *StackTracePool) Get(skipLevel int) *LazyStackTrace {
	lst := stp.pool.Get().(*LazyStackTrace)
	lst.Reset(skipLevel + 1) // +1 para pular esta função
	return lst
}

// Put retorna LazyStackTrace ao pool
func (stp *StackTracePool) Put(lst *LazyStackTrace) {
	if lst != nil && cap(lst.frames) <= 64 && cap(lst.programCounters) <= 128 {
		// Só retorna ao pool se não cresceu muito
		stp.pool.Put(lst)
	}
}

// OptimizedStackCapture captura stack trace otimizada
type OptimizedStackCapture struct {
	pool         *StackTracePool
	enabledDepth int  // Profundidade máxima de captura
	enabled      bool // Se captura está habilitada globalmente
	mu           sync.RWMutex
}

// NewOptimizedStackCapture cria novo capturador otimizado
func NewOptimizedStackCapture() *OptimizedStackCapture {
	return &OptimizedStackCapture{
		pool:         NewStackTracePool(),
		enabledDepth: 32, // Padrão: 32 frames
		enabled:      true,
	}
}

// CaptureStackTrace captura stack trace otimizado
func (osc *OptimizedStackCapture) CaptureStackTrace(skip int) *LazyStackTrace {
	osc.mu.RLock()
	enabled := osc.enabled
	osc.mu.RUnlock()

	if !enabled {
		return &LazyStackTrace{} // Retorna instância vazia
	}

	return osc.pool.Get(skip + 1) // +1 para pular esta função
}

// ReleaseStackTrace libera stack trace de volta ao pool
func (osc *OptimizedStackCapture) ReleaseStackTrace(lst *LazyStackTrace) {
	osc.pool.Put(lst)
}

// SetEnabled habilita/desabilita captura de stack trace
func (osc *OptimizedStackCapture) SetEnabled(enabled bool) {
	osc.mu.Lock()
	defer osc.mu.Unlock()
	osc.enabled = enabled
}

// SetMaxDepth define profundidade máxima de captura
func (osc *OptimizedStackCapture) SetMaxDepth(depth int) {
	osc.mu.Lock()
	defer osc.mu.Unlock()
	osc.enabledDepth = depth
}

// IsEnabled verifica se captura está habilitada
func (osc *OptimizedStackCapture) IsEnabled() bool {
	osc.mu.RLock()
	defer osc.mu.RUnlock()
	return osc.enabled
}

// GetMaxDepth retorna profundidade máxima
func (osc *OptimizedStackCapture) GetMaxDepth() int {
	osc.mu.RLock()
	defer osc.mu.RUnlock()
	return osc.enabledDepth
}

// ConditionalStackCapture captura stack trace baseado em condições
type ConditionalStackCapture struct {
	*OptimizedStackCapture
	conditions []func() bool
	mu         sync.RWMutex
}

// NewConditionalStackCapture cria capturador condicional
func NewConditionalStackCapture() *ConditionalStackCapture {
	return &ConditionalStackCapture{
		OptimizedStackCapture: NewOptimizedStackCapture(),
		conditions:            make([]func() bool, 0),
	}
}

// AddCondition adiciona condição para captura
func (csc *ConditionalStackCapture) AddCondition(condition func() bool) {
	csc.mu.Lock()
	defer csc.mu.Unlock()
	csc.conditions = append(csc.conditions, condition)
}

// ClearConditions limpa todas as condições
func (csc *ConditionalStackCapture) ClearConditions() {
	csc.mu.Lock()
	defer csc.mu.Unlock()
	csc.conditions = csc.conditions[:0]
}

// ShouldCapture verifica se deve capturar baseado nas condições
func (csc *ConditionalStackCapture) ShouldCapture() bool {
	csc.mu.RLock()
	defer csc.mu.RUnlock()

	if !csc.enabled {
		return false
	}

	// Se não há condições, sempre captura (quando habilitado)
	if len(csc.conditions) == 0 {
		return true
	}

	// Verifica se alguma condição é verdadeira
	for _, condition := range csc.conditions {
		if condition() {
			return true
		}
	}

	return false
}

// CaptureConditionalStackTrace captura apenas se condições são atendidas
func (csc *ConditionalStackCapture) CaptureConditionalStackTrace(skip int) *LazyStackTrace {
	if !csc.ShouldCapture() {
		return &LazyStackTrace{} // Retorna instância vazia
	}

	return csc.OptimizedStackCapture.CaptureStackTrace(skip + 1)
}

// Instâncias globais
var (
	GlobalStackTracePool     = NewStackTracePool()
	GlobalOptimizedCapture   = NewOptimizedStackCapture()
	GlobalConditionalCapture = NewConditionalStackCapture()
)

// CaptureStackTrace função global para captura otimizada
func CaptureStackTrace(skip int) *LazyStackTrace {
	return GlobalOptimizedCapture.CaptureStackTrace(skip + 1)
}

// CaptureConditionalStackTrace função global para captura condicional
func CaptureConditionalStackTrace(skip int) *LazyStackTrace {
	return GlobalConditionalCapture.CaptureConditionalStackTrace(skip + 1)
}

// ReleaseStackTrace função global para liberar stack trace
func ReleaseStackTrace(lst *LazyStackTrace) {
	GlobalOptimizedCapture.ReleaseStackTrace(lst)
}

// SetStackTraceEnabled habilita/desabilita captura global
func SetStackTraceEnabled(enabled bool) {
	GlobalOptimizedCapture.SetEnabled(enabled)
	GlobalConditionalCapture.SetEnabled(enabled)
}

// AddGlobalStackCondition adiciona condição global para captura
func AddGlobalStackCondition(condition func() bool) {
	GlobalConditionalCapture.AddCondition(condition)
}
