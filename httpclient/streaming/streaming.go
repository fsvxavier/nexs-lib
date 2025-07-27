// Package streaming provides streaming support for HTTP requests and responses.
package streaming

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// StreamProcessor handles streaming HTTP operations.
type StreamProcessor struct {
	client       interfaces.Client
	chunkSize    int
	bufferSize   int
	readTimeout  time.Duration
	writeTimeout time.Duration
	errorHandler func(error)
}

// NewStreamProcessor creates a new stream processor.
func NewStreamProcessor(client interfaces.Client) *StreamProcessor {
	return &StreamProcessor{
		client:       client,
		chunkSize:    8192,  // 8KB chunks
		bufferSize:   32768, // 32KB buffer
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		errorHandler: func(error) {}, // Default no-op error handler
	}
}

// SetChunkSize sets the chunk size for streaming operations.
func (s *StreamProcessor) SetChunkSize(size int) *StreamProcessor {
	s.chunkSize = size
	return s
}

// SetBufferSize sets the buffer size for streaming operations.
func (s *StreamProcessor) SetBufferSize(size int) *StreamProcessor {
	s.bufferSize = size
	return s
}

// SetReadTimeout sets the read timeout for streaming operations.
func (s *StreamProcessor) SetReadTimeout(timeout time.Duration) *StreamProcessor {
	s.readTimeout = timeout
	return s
}

// SetWriteTimeout sets the write timeout for streaming operations.
func (s *StreamProcessor) SetWriteTimeout(timeout time.Duration) *StreamProcessor {
	s.writeTimeout = timeout
	return s
}

// SetErrorHandler sets the error handler for streaming operations.
func (s *StreamProcessor) SetErrorHandler(handler func(error)) *StreamProcessor {
	s.errorHandler = handler
	return s
}

// StreamDownload downloads content using streaming with a custom handler.
func (s *StreamProcessor) StreamDownload(ctx context.Context, method, url string, handler interfaces.StreamHandler) error {
	req := &interfaces.Request{
		Method:  method,
		URL:     url,
		Context: ctx,
	}

	provider := s.client.GetProvider()
	if provider == nil {
		return fmt.Errorf("no provider available")
	}

	resp, err := provider.DoRequest(ctx, req)
	if err != nil {
		handler.OnError(err)
		return err
	}

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("HTTP error: %d", resp.StatusCode)
		handler.OnError(err)
		return err
	}

	// Process response body in chunks
	err = s.processStreamingResponse(resp.Body, handler)
	if err != nil {
		handler.OnError(err)
		return err
	}

	handler.OnComplete()
	return nil
}

// StreamUpload uploads content using streaming.
func (s *StreamProcessor) StreamUpload(ctx context.Context, method, url string, reader io.Reader, contentType string) (*interfaces.Response, error) {
	// Create a pipe for streaming upload
	pipeReader, pipeWriter := io.Pipe()

	// Start copying data in a goroutine
	go func() {
		defer pipeWriter.Close()
		_, err := io.Copy(pipeWriter, reader)
		if err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	req := &interfaces.Request{
		Method:      method,
		URL:         url,
		Body:        pipeReader,
		ContentType: contentType,
		Context:     ctx,
	}

	provider := s.client.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("no provider available")
	}

	return provider.DoRequest(ctx, req)
}

// processStreamingResponse processes a streaming response body.
func (s *StreamProcessor) processStreamingResponse(body []byte, handler interfaces.StreamHandler) error {
	reader := strings.NewReader(string(body))
	buffer := make([]byte, s.chunkSize)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])

			if handlerErr := handler.OnData(chunk); handlerErr != nil {
				return handlerErr
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// ServerSentEventsHandler handles Server-Sent Events streams.
type ServerSentEventsHandler struct {
	eventCallback func(event *SSEEvent)
	errorCallback func(error)
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	ID    string
	Event string
	Data  string
	Retry int
}

// NewServerSentEventsHandler creates a new SSE handler.
func NewServerSentEventsHandler(eventCallback func(*SSEEvent), errorCallback func(error)) *ServerSentEventsHandler {
	return &ServerSentEventsHandler{
		eventCallback: eventCallback,
		errorCallback: errorCallback,
	}
}

// OnData implements the StreamHandler interface.
func (h *ServerSentEventsHandler) OnData(data []byte) error {
	events := h.parseSSEData(string(data))
	for _, event := range events {
		if h.eventCallback != nil {
			h.eventCallback(event)
		}
	}
	return nil
}

// OnError implements the StreamHandler interface.
func (h *ServerSentEventsHandler) OnError(err error) {
	if h.errorCallback != nil {
		h.errorCallback(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *ServerSentEventsHandler) OnComplete() {
	// SSE streams typically don't complete, but we can handle it if needed
}

// parseSSEData parses Server-Sent Events data.
func (h *ServerSentEventsHandler) parseSSEData(data string) []*SSEEvent {
	events := make([]*SSEEvent, 0)
	lines := strings.Split(data, "\n")

	var currentEvent *SSEEvent

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			// Empty line indicates end of event
			if currentEvent != nil {
				events = append(events, currentEvent)
				currentEvent = nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment line, ignore
			continue
		}

		if currentEvent == nil {
			currentEvent = &SSEEvent{}
		}

		// Parse field: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch field {
		case "id":
			currentEvent.ID = value
		case "event":
			currentEvent.Event = value
		case "data":
			if currentEvent.Data != "" {
				currentEvent.Data += "\n"
			}
			currentEvent.Data += value
		case "retry":
			// Parse retry time (not implemented in this example)
		}
	}

	// Add final event if exists
	if currentEvent != nil {
		events = append(events, currentEvent)
	}

	return events
}

// ChunkedHandler handles chunked transfer encoding.
type ChunkedHandler struct {
	chunkCallback func(chunk []byte)
	errorCallback func(error)
}

// NewChunkedHandler creates a new chunked handler.
func NewChunkedHandler(chunkCallback func([]byte), errorCallback func(error)) *ChunkedHandler {
	return &ChunkedHandler{
		chunkCallback: chunkCallback,
		errorCallback: errorCallback,
	}
}

// OnData implements the StreamHandler interface.
func (h *ChunkedHandler) OnData(data []byte) error {
	if h.chunkCallback != nil {
		h.chunkCallback(data)
	}
	return nil
}

// OnError implements the StreamHandler interface.
func (h *ChunkedHandler) OnError(err error) {
	if h.errorCallback != nil {
		h.errorCallback(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *ChunkedHandler) OnComplete() {
	// Chunk transfer complete
}

// ProgressHandler tracks download/upload progress.
type ProgressHandler struct {
	totalSize        int64
	processedSize    int64
	progressCallback func(processed, total int64, percentage float64)
	errorCallback    func(error)
}

// NewProgressHandler creates a new progress handler.
func NewProgressHandler(totalSize int64, progressCallback func(int64, int64, float64), errorCallback func(error)) *ProgressHandler {
	return &ProgressHandler{
		totalSize:        totalSize,
		progressCallback: progressCallback,
		errorCallback:    errorCallback,
	}
}

// OnData implements the StreamHandler interface.
func (h *ProgressHandler) OnData(data []byte) error {
	h.processedSize += int64(len(data))

	var percentage float64
	if h.totalSize > 0 {
		percentage = float64(h.processedSize) / float64(h.totalSize) * 100
	}

	if h.progressCallback != nil {
		h.progressCallback(h.processedSize, h.totalSize, percentage)
	}

	return nil
}

// OnError implements the StreamHandler interface.
func (h *ProgressHandler) OnError(err error) {
	if h.errorCallback != nil {
		h.errorCallback(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *ProgressHandler) OnComplete() {
	// Download/upload complete
	if h.progressCallback != nil {
		h.progressCallback(h.processedSize, h.totalSize, 100.0)
	}
}

// FileHandler saves streamed data to a file.
type FileHandler struct {
	writer        io.Writer
	errorCallback func(error)
}

// NewFileHandler creates a new file handler.
func NewFileHandler(writer io.Writer, errorCallback func(error)) *FileHandler {
	return &FileHandler{
		writer:        writer,
		errorCallback: errorCallback,
	}
}

// OnData implements the StreamHandler interface.
func (h *FileHandler) OnData(data []byte) error {
	_, err := h.writer.Write(data)
	if err != nil {
		h.OnError(err)
		return err
	}
	return nil
}

// OnError implements the StreamHandler interface.
func (h *FileHandler) OnError(err error) {
	if h.errorCallback != nil {
		h.errorCallback(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *FileHandler) OnComplete() {
	// File write complete
}

// CompositeHandler combines multiple handlers.
type CompositeHandler struct {
	handlers []interfaces.StreamHandler
}

// NewCompositeHandler creates a new composite handler.
func NewCompositeHandler(handlers ...interfaces.StreamHandler) *CompositeHandler {
	return &CompositeHandler{
		handlers: handlers,
	}
}

// OnData implements the StreamHandler interface.
func (h *CompositeHandler) OnData(data []byte) error {
	for _, handler := range h.handlers {
		if err := handler.OnData(data); err != nil {
			return err
		}
	}
	return nil
}

// OnError implements the StreamHandler interface.
func (h *CompositeHandler) OnError(err error) {
	for _, handler := range h.handlers {
		handler.OnError(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *CompositeHandler) OnComplete() {
	for _, handler := range h.handlers {
		handler.OnComplete()
	}
}

// BufferedHandler buffers data before processing.
type BufferedHandler struct {
	buffer        []byte
	bufferSize    int
	handler       interfaces.StreamHandler
	errorCallback func(error)
}

// NewBufferedHandler creates a new buffered handler.
func NewBufferedHandler(bufferSize int, handler interfaces.StreamHandler, errorCallback func(error)) *BufferedHandler {
	return &BufferedHandler{
		buffer:        make([]byte, 0, bufferSize),
		bufferSize:    bufferSize,
		handler:       handler,
		errorCallback: errorCallback,
	}
}

// OnData implements the StreamHandler interface.
func (h *BufferedHandler) OnData(data []byte) error {
	h.buffer = append(h.buffer, data...)

	// Process buffer when it reaches the desired size
	if len(h.buffer) >= h.bufferSize {
		return h.flushBuffer()
	}

	return nil
}

// OnError implements the StreamHandler interface.
func (h *BufferedHandler) OnError(err error) {
	if h.errorCallback != nil {
		h.errorCallback(err)
	}
	if h.handler != nil {
		h.handler.OnError(err)
	}
}

// OnComplete implements the StreamHandler interface.
func (h *BufferedHandler) OnComplete() {
	// Flush any remaining data
	if len(h.buffer) > 0 {
		h.flushBuffer()
	}
	if h.handler != nil {
		h.handler.OnComplete()
	}
}

// flushBuffer processes the buffered data.
func (h *BufferedHandler) flushBuffer() error {
	if h.handler != nil && len(h.buffer) > 0 {
		err := h.handler.OnData(h.buffer)
		if err != nil {
			return err
		}
	}
	h.buffer = h.buffer[:0] // Reset buffer
	return nil
}
