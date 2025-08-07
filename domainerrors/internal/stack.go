package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// StackTraceCapture implementa a captura de stack trace
type StackTraceCapture struct {
	withStackTrace bool
}

// NewStackTraceCapture cria uma nova instância de StackTraceCapture
func NewStackTraceCapture(withStackTrace bool) *StackTraceCapture {
	return &StackTraceCapture{
		withStackTrace: withStackTrace,
	}
}

// CaptureStackTrace captura o stack trace atual
func (s *StackTraceCapture) CaptureStackTrace(skip int) []interfaces.StackFrame {
	if !s.withStackTrace {
		return nil
	}

	var frames []interfaces.StackFrame
	pc := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pc)

	if n == 0 {
		return frames
	}

	pc = pc[:n]
	callersFrames := runtime.CallersFrames(pc)

	timestamp := time.Now().Format(time.RFC3339)

	for {
		frame, more := callersFrames.Next()

		// Filtra frames internos do Go runtime
		if !s.shouldIncludeFrame(frame.Function) {
			if !more {
				break
			}
			continue
		}

		stackFrame := interfaces.StackFrame{
			Function: frame.Function,
			File:     filepath.Base(frame.File),
			Line:     frame.Line,
			Time:     timestamp,
		}

		frames = append(frames, stackFrame)

		if !more {
			break
		}
	}

	return frames
}

// FormatStackTrace formata o stack trace para exibição
func (s *StackTraceCapture) FormatStackTrace(frames []interfaces.StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Stack trace:\n")

	for i, frame := range frames {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.Function))
		builder.WriteString(fmt.Sprintf("     %s:%d\n", frame.File, frame.Line))
		if frame.Message != "" {
			builder.WriteString(fmt.Sprintf("     Message: %s\n", frame.Message))
		}
		builder.WriteString(fmt.Sprintf("     Time: %s\n", frame.Time))
		if i < len(frames)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// shouldIncludeFrame determina se um frame deve ser incluído no stack trace
func (s *StackTraceCapture) shouldIncludeFrame(function string) bool {
	// Filtra frames internos do runtime Go
	excludedPrefixes := []string{
		"runtime.",
		"testing.",
		"reflect.",
		"syscall.",
		"os.",
	}

	for _, prefix := range excludedPrefixes {
		if strings.HasPrefix(function, prefix) {
			return false
		}
	}

	return true
}

// DefaultStackTraceCapture cria uma instância padrão com stack trace habilitado
func DefaultStackTraceCapture() *StackTraceCapture {
	return NewStackTraceCapture(true)
}

// NoStackTraceCapture cria uma instância sem captura de stack trace
func NoStackTraceCapture() *StackTraceCapture {
	return NewStackTraceCapture(false)
}
