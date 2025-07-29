package hooks

import (
	"context"
	"log"
	"time"
)

// LoggingHook implementa hooks para logging de operações.
type LoggingHook struct {
	logExecution  bool
	logConnection bool
	logPipeline   bool
	logRetry      bool
}

// NewLoggingHook cria um novo LoggingHook.
func NewLoggingHook() *LoggingHook {
	return &LoggingHook{
		logExecution:  true,
		logConnection: true,
		logPipeline:   true,
		logRetry:      true,
	}
}

// WithExecutionLogging habilita/desabilita logging de execução.
func (l *LoggingHook) WithExecutionLogging(enabled bool) *LoggingHook {
	l.logExecution = enabled
	return l
}

// WithConnectionLogging habilita/desabilita logging de conexão.
func (l *LoggingHook) WithConnectionLogging(enabled bool) *LoggingHook {
	l.logConnection = enabled
	return l
}

// WithPipelineLogging habilita/desabilita logging de pipeline.
func (l *LoggingHook) WithPipelineLogging(enabled bool) *LoggingHook {
	l.logPipeline = enabled
	return l
}

// WithRetryLogging habilita/desabilita logging de retry.
func (l *LoggingHook) WithRetryLogging(enabled bool) *LoggingHook {
	l.logRetry = enabled
	return l
}

// BeforeExecution implementa ExecutionHook.
func (l *LoggingHook) BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context {
	if l.logExecution {
		log.Printf("[VALKEY] Executing command: %s with %d args", cmd, len(args))
	}
	return context.WithValue(ctx, "start_time", time.Now())
}

// AfterExecution implementa ExecutionHook.
func (l *LoggingHook) AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration) {
	if l.logExecution {
		if err != nil {
			log.Printf("[VALKEY] Command %s failed after %v: %v", cmd, duration, err)
		} else {
			log.Printf("[VALKEY] Command %s completed successfully in %v", cmd, duration)
		}
	}
}

// BeforeConnect implementa ConnectionHook.
func (l *LoggingHook) BeforeConnect(ctx context.Context, network, addr string) context.Context {
	if l.logConnection {
		log.Printf("[VALKEY] Connecting to %s://%s", network, addr)
	}
	return ctx
}

// AfterConnect implementa ConnectionHook.
func (l *LoggingHook) AfterConnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	if l.logConnection {
		if err != nil {
			log.Printf("[VALKEY] Failed to connect to %s://%s after %v: %v", network, addr, duration, err)
		} else {
			log.Printf("[VALKEY] Connected to %s://%s in %v", network, addr, duration)
		}
	}
}

// BeforeDisconnect implementa ConnectionHook.
func (l *LoggingHook) BeforeDisconnect(ctx context.Context, network, addr string) context.Context {
	if l.logConnection {
		log.Printf("[VALKEY] Disconnecting from %s://%s", network, addr)
	}
	return ctx
}

// AfterDisconnect implementa ConnectionHook.
func (l *LoggingHook) AfterDisconnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	if l.logConnection {
		if err != nil {
			log.Printf("[VALKEY] Failed to disconnect from %s://%s after %v: %v", network, addr, duration, err)
		} else {
			log.Printf("[VALKEY] Disconnected from %s://%s in %v", network, addr, duration)
		}
	}
}

// BeforePipelineExecution implementa PipelineHook.
func (l *LoggingHook) BeforePipelineExecution(ctx context.Context, commands []string) context.Context {
	if l.logPipeline {
		log.Printf("[VALKEY] Executing pipeline with %d commands", len(commands))
	}
	return ctx
}

// AfterPipelineExecution implementa PipelineHook.
func (l *LoggingHook) AfterPipelineExecution(ctx context.Context, commands []string, results []interface{}, err error, duration time.Duration) {
	if l.logPipeline {
		if err != nil {
			log.Printf("[VALKEY] Pipeline with %d commands failed after %v: %v", len(commands), duration, err)
		} else {
			log.Printf("[VALKEY] Pipeline with %d commands completed successfully in %v", len(commands), duration)
		}
	}
}

// BeforeRetry implementa RetryHook.
func (l *LoggingHook) BeforeRetry(ctx context.Context, attempt int, err error) context.Context {
	if l.logRetry {
		log.Printf("[VALKEY] Retrying operation (attempt %d) after error: %v", attempt, err)
	}
	return ctx
}

// AfterRetry implementa RetryHook.
func (l *LoggingHook) AfterRetry(ctx context.Context, attempt int, success bool, err error) {
	if l.logRetry {
		if success {
			log.Printf("[VALKEY] Retry attempt %d succeeded", attempt)
		} else {
			log.Printf("[VALKEY] Retry attempt %d failed: %v", attempt, err)
		}
	}
}
