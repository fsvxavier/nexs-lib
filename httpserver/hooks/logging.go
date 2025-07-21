// Package hooks provides observer implementations for HTTP server lifecycle events.
package hooks

import (
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// LoggingObserver implements ServerObserver for logging server events.
type LoggingObserver struct {
	logger *log.Logger
}

// NewLoggingObserver creates a new logging observer.
func NewLoggingObserver(logger *log.Logger) *LoggingObserver {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingObserver{
		logger: logger,
	}
}

// OnStart logs when the server starts.
func (l *LoggingObserver) OnStart(name string) {
	l.logger.Printf("Server %s started", name)
}

// OnStop logs when the server stops.
func (l *LoggingObserver) OnStop(name string) {
	l.logger.Printf("Server %s stopped", name)
}

// OnRequest logs each HTTP request.
func (l *LoggingObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
	l.logger.Printf("Server %s - %s %s - Status: %d - Duration: %v",
		name, req.Method, req.URL.Path, status, duration)
}

// OnBeforeRequest logs before processing a request.
func (l *LoggingObserver) OnBeforeRequest(name string, req *http.Request) {
	l.logger.Printf("Server %s - Processing %s %s", name, req.Method, req.URL.Path)
}

// OnAfterRequest logs after processing a request.
func (l *LoggingObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
	l.logger.Printf("Server %s - Completed %s %s - Status: %d - Duration: %v",
		name, req.Method, req.URL.Path, status, duration)
}

// Ensure LoggingObserver implements ServerObserver.
var _ interfaces.ServerObserver = (*LoggingObserver)(nil)
