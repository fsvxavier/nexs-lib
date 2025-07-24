package internal

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// StackFrame representa informações de um frame do stack trace
type StackFrame struct {
	Function string    `json:"function"`
	File     string    `json:"file"`
	Line     int       `json:"line"`
	Message  string    `json:"message,omitempty"`
	Time     time.Time `json:"time"`
}

// CaptureStackTrace captura o stack trace atual
func CaptureStackTrace(skip int, message string) []StackFrame {
	var frames []StackFrame

	// Captura até 50 frames do stack trace
	for i := skip; i < skip+50; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		fnName := "unknown"
		if fn != nil {
			fnName = fn.Name()
		}

		frame := StackFrame{
			Function: fnName,
			File:     file,
			Line:     line,
			Message:  message,
			Time:     time.Now(),
		}

		frames = append(frames, frame)
	}

	return frames
}

// FormatStackTrace formata o stack trace para exibição
// Baseado nas linhas 252-269 do arquivo de referência
func FormatStackTrace(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Error Stack Trace:\n")

	for i, st := range frames {
		b.WriteString(fmt.Sprintf("%d: [%s] in %s (%s:%d)\n",
			i+1, st.Message, st.Function, st.File, st.Line))
	}

	return b.String()
}

// GetCallerInfo retorna informações sobre o chamador
func GetCallerInfo(skip int) (function, file string, line int, ok bool) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0, false
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown", file, line, true
	}

	return fn.Name(), file, line, true
}

// TrimStackTrace remove frames desnecessários do stack trace
func TrimStackTrace(frames []StackFrame, maxFrames int) []StackFrame {
	if len(frames) <= maxFrames {
		return frames
	}

	// Mantém os primeiros frames (mais relevantes)
	return frames[:maxFrames]
}

// FilterStackTrace filtra frames do stack trace baseado em critérios
func FilterStackTrace(frames []StackFrame, filter func(StackFrame) bool) []StackFrame {
	var filtered []StackFrame

	for _, frame := range frames {
		if filter(frame) {
			filtered = append(filtered, frame)
		}
	}

	return filtered
}

// IsSystemFrame verifica se o frame é do sistema (runtime, etc.)
func IsSystemFrame(frame StackFrame) bool {
	systemPrefixes := []string{
		"runtime.",
		"testing.",
		"reflect.",
		"sync.",
		"os.",
		"syscall.",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(frame.Function, prefix) {
			return true
		}
	}

	return false
}

// IsTestFrame verifica se o frame é de teste
func IsTestFrame(frame StackFrame) bool {
	return strings.Contains(frame.Function, "Test") ||
		strings.Contains(frame.File, "_test.go") ||
		strings.Contains(frame.Function, "Benchmark")
}

// GetUserFrames retorna apenas frames do código do usuário
func GetUserFrames(frames []StackFrame) []StackFrame {
	return FilterStackTrace(frames, func(frame StackFrame) bool {
		return !IsSystemFrame(frame) && !IsTestFrame(frame)
	})
}
