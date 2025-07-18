package internal

import (
	"fmt"
	"runtime"
	"strings"
)

// StackTraceConfig define configurações para captura de stack trace
type StackTraceConfig struct {
	Enabled    bool
	MaxDepth   int
	SkipFrames int
}

// DefaultStackTraceConfig retorna configuração padrão
func DefaultStackTraceConfig() *StackTraceConfig {
	return &StackTraceConfig{
		Enabled:    true,
		MaxDepth:   10,
		SkipFrames: 2,
	}
}

// StackFrame representa um frame do stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
	PC       uintptr
}

// String retorna representação textual do frame
func (sf StackFrame) String() string {
	return fmt.Sprintf("%s:%d %s", sf.File, sf.Line, sf.Function)
}

// StackTrace captura e armazena stack trace
type StackTrace struct {
	Frames []StackFrame
	config *StackTraceConfig
}

// NewStackTrace cria um novo stack trace
func NewStackTrace(config *StackTraceConfig) *StackTrace {
	if config == nil {
		config = DefaultStackTraceConfig()
	}

	st := &StackTrace{
		config: config,
	}

	if config.Enabled {
		st.capture()
	}

	return st
}

// capture captura o stack trace atual
func (st *StackTrace) capture() {
	if !st.config.Enabled {
		return
	}

	// Captura program counters
	pcs := make([]uintptr, st.config.MaxDepth)
	n := runtime.Callers(st.config.SkipFrames, pcs)

	if n == 0 {
		return
	}

	// Obtém frames
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()

		st.Frames = append(st.Frames, StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
			PC:       frame.PC,
		})

		if !more {
			break
		}
	}
}

// String retorna representação textual do stack trace
func (st *StackTrace) String() string {
	if len(st.Frames) == 0 {
		return ""
	}

	var b strings.Builder

	for i, frame := range st.Frames {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("  %d. %s", i+1, frame.String()))
	}

	return b.String()
}

// Format formata o stack trace com opções
func (st *StackTrace) Format(options FormatOptions) string {
	if len(st.Frames) == 0 {
		return ""
	}

	var b strings.Builder

	if options.IncludeHeader {
		b.WriteString("Stack Trace:\n")
	}

	for i, frame := range st.Frames {
		if options.MaxFrames > 0 && i >= options.MaxFrames {
			break
		}

		if i > 0 {
			b.WriteString("\n")
		}

		if options.IncludeNumbers {
			b.WriteString(fmt.Sprintf("  %d. ", i+1))
		} else {
			b.WriteString("  ")
		}

		if options.ShortPaths {
			b.WriteString(st.formatShortPath(frame))
		} else {
			b.WriteString(frame.String())
		}
	}

	return b.String()
}

// formatShortPath formata o path de forma mais curta
func (st *StackTrace) formatShortPath(frame StackFrame) string {
	// Extrai apenas o nome do arquivo e diretório pai
	parts := strings.Split(frame.File, "/")
	if len(parts) > 2 {
		shortPath := strings.Join(parts[len(parts)-2:], "/")
		return fmt.Sprintf("%s:%d %s", shortPath, frame.Line, st.extractFunctionName(frame.Function))
	}

	return fmt.Sprintf("%s:%d %s", frame.File, frame.Line, st.extractFunctionName(frame.Function))
}

// extractFunctionName extrai o nome curto da função
func (st *StackTrace) extractFunctionName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// GetTopFrame retorna o frame do topo do stack
func (st *StackTrace) GetTopFrame() *StackFrame {
	if len(st.Frames) == 0 {
		return nil
	}
	return &st.Frames[0]
}

// GetFrames retorna todos os frames
func (st *StackTrace) GetFrames() []StackFrame {
	return st.Frames
}

// HasFrames retorna se há frames capturados
func (st *StackTrace) HasFrames() bool {
	return len(st.Frames) > 0
}

// FormatOptions define opções de formatação
type FormatOptions struct {
	IncludeHeader  bool
	IncludeNumbers bool
	ShortPaths     bool
	MaxFrames      int
}

// DefaultFormatOptions retorna opções padrão de formatação
func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		IncludeHeader:  true,
		IncludeNumbers: true,
		ShortPaths:     false,
		MaxFrames:      10,
	}
}

// CompactFormatOptions retorna opções de formatação compacta
func CompactFormatOptions() FormatOptions {
	return FormatOptions{
		IncludeHeader:  false,
		IncludeNumbers: false,
		ShortPaths:     true,
		MaxFrames:      5,
	}
}

// CaptureStackTrace captura um stack trace com configuração padrão
func CaptureStackTrace() string {
	st := NewStackTrace(DefaultStackTraceConfig())
	return st.String()
}

// CaptureStackTraceWithSkip captura um stack trace pulando frames
func CaptureStackTraceWithSkip(skip int) string {
	config := DefaultStackTraceConfig()
	config.SkipFrames = skip
	st := NewStackTrace(config)
	return st.String()
}

// CaptureCompactStackTrace captura um stack trace compacto
func CaptureCompactStackTrace() string {
	st := NewStackTrace(DefaultStackTraceConfig())
	return st.Format(CompactFormatOptions())
}

// StackTraceUtils fornece utilitários para stack trace
type StackTraceUtils struct{}

// NewStackTraceUtils cria uma nova instância de utilitários
func NewStackTraceUtils() *StackTraceUtils {
	return &StackTraceUtils{}
}

// ExtractCallerInfo extrai informações do caller
func (stu *StackTraceUtils) ExtractCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", file, line
	}

	return fn.Name(), file, line
}

// GetCurrentFunction retorna o nome da função atual
func (stu *StackTraceUtils) GetCurrentFunction() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return fn.Name()
}

// GetCallerFunction retorna o nome da função que chamou a atual
func (stu *StackTraceUtils) GetCallerFunction() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return fn.Name()
}

// FormatPanicStackTrace formata um stack trace de panic
func FormatPanicStackTrace(stackTrace []byte) string {
	lines := strings.Split(string(stackTrace), "\n")
	var formatted strings.Builder

	formatted.WriteString("Panic Stack Trace:\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if i%2 == 0 {
			// Linha da função
			formatted.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(line)))
		} else {
			// Linha do arquivo e linha
			formatted.WriteString(fmt.Sprintf("    %s\n", strings.TrimSpace(line)))
		}
	}

	return formatted.String()
}

// IsStackTraceEnabled verifica se stack trace está habilitado globalmente
var globalStackTraceEnabled = true

// SetGlobalStackTraceEnabled habilita/desabilita stack trace globalmente
func SetGlobalStackTraceEnabled(enabled bool) {
	globalStackTraceEnabled = enabled
}

// IsGlobalStackTraceEnabled retorna se stack trace está habilitado globalmente
func IsGlobalStackTraceEnabled() bool {
	return globalStackTraceEnabled
}
