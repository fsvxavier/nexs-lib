package streaming

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Mock client for streaming tests
type mockStreamClient struct {
	provider mockStreamProvider
}

func (c *mockStreamClient) GetProvider() interfaces.Provider {
	return &c.provider
}

func (c *mockStreamClient) Get(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Post(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Put(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Delete(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Patch(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Head(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Options(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) Execute(ctx context.Context, method, endpoint string, body interface{}) (*interfaces.Response, error) {
	return nil, nil
}
func (c *mockStreamClient) SetHeaders(headers map[string]string) interfaces.Client { return c }
func (c *mockStreamClient) SetTimeout(timeout time.Duration) interfaces.Client     { return c }
func (c *mockStreamClient) SetErrorHandler(handler interfaces.ErrorHandler) interfaces.Client {
	return c
}
func (c *mockStreamClient) SetRetryConfig(config *interfaces.RetryConfig) interfaces.Client { return c }
func (c *mockStreamClient) Unmarshal(v interface{}) interfaces.Client                       { return c }
func (c *mockStreamClient) UnmarshalResponse(resp *interfaces.Response, v interface{}) error {
	return nil
}
func (c *mockStreamClient) AddMiddleware(middleware interfaces.Middleware) interfaces.Client {
	return c
}
func (c *mockStreamClient) RemoveMiddleware(middleware interfaces.Middleware) interfaces.Client {
	return c
}
func (c *mockStreamClient) AddHook(hook interfaces.Hook) interfaces.Client    { return c }
func (c *mockStreamClient) RemoveHook(hook interfaces.Hook) interfaces.Client { return c }
func (c *mockStreamClient) Batch() interfaces.BatchRequestBuilder             { return nil }
func (c *mockStreamClient) Stream(ctx context.Context, method, endpoint string, handler interfaces.StreamHandler) error {
	return nil
}
func (c *mockStreamClient) GetConfig() *interfaces.Config { return &interfaces.Config{} }
func (c *mockStreamClient) GetID() string                 { return "mock-stream-client" }
func (c *mockStreamClient) IsHealthy() bool               { return true }
func (c *mockStreamClient) GetMetrics() *interfaces.ProviderMetrics {
	return &interfaces.ProviderMetrics{}
}

// Mock provider for streaming tests
type mockStreamProvider struct {
	response *interfaces.Response
	err      error
}

func (p *mockStreamProvider) Name() string                              { return "mock-stream" }
func (p *mockStreamProvider) Version() string                           { return "1.0.0" }
func (p *mockStreamProvider) Configure(config *interfaces.Config) error { return nil }
func (p *mockStreamProvider) SetDefaults()                              {}
func (p *mockStreamProvider) IsHealthy() bool                           { return true }
func (p *mockStreamProvider) GetMetrics() *interfaces.ProviderMetrics {
	return &interfaces.ProviderMetrics{}
}

func (p *mockStreamProvider) DoRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.response != nil {
		return p.response, nil
	}
	return &interfaces.Response{
		StatusCode: 200,
		Body:       []byte("chunk1chunk2chunk3"),
	}, nil
}

// Mock stream handler for testing
type mockStreamHandler struct {
	receivedData [][]byte
	errors       []error
	completed    bool
}

func (h *mockStreamHandler) OnData(data []byte) error {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	h.receivedData = append(h.receivedData, dataCopy)
	return nil
}

func (h *mockStreamHandler) OnError(err error) {
	h.errors = append(h.errors, err)
}

func (h *mockStreamHandler) OnComplete() {
	h.completed = true
}

func TestStreamProcessor_Configuration(t *testing.T) {
	client := &mockStreamClient{}
	processor := NewStreamProcessor(client)

	processor.SetChunkSize(4096).
		SetBufferSize(16384).
		SetReadTimeout(10 * time.Second).
		SetWriteTimeout(15 * time.Second)

	if processor.chunkSize != 4096 {
		t.Errorf("Expected chunk size 4096, got %d", processor.chunkSize)
	}

	if processor.bufferSize != 16384 {
		t.Errorf("Expected buffer size 16384, got %d", processor.bufferSize)
	}

	if processor.readTimeout != 10*time.Second {
		t.Errorf("Expected read timeout 10s, got %v", processor.readTimeout)
	}

	if processor.writeTimeout != 15*time.Second {
		t.Errorf("Expected write timeout 15s, got %v", processor.writeTimeout)
	}
}

func TestStreamProcessor_StreamDownload(t *testing.T) {
	client := &mockStreamClient{}
	client.provider.response = &interfaces.Response{
		StatusCode: 200,
		Body:       []byte("Hello, streaming world! This is a test."),
	}

	processor := NewStreamProcessor(client).SetChunkSize(10)
	handler := &mockStreamHandler{}

	ctx := context.Background()
	err := processor.StreamDownload(ctx, "GET", "http://test.com/stream", handler)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !handler.completed {
		t.Error("Expected handler to be completed")
	}

	if len(handler.errors) > 0 {
		t.Errorf("Expected no errors, got %v", handler.errors)
	}

	// Verify data was received
	if len(handler.receivedData) == 0 {
		t.Error("Expected to receive data chunks")
	}

	// Concatenate all received chunks
	var allData []byte
	for _, chunk := range handler.receivedData {
		allData = append(allData, chunk...)
	}

	expected := "Hello, streaming world! This is a test."
	if string(allData) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(allData))
	}
}

func TestStreamProcessor_StreamDownloadError(t *testing.T) {
	client := &mockStreamClient{}
	client.provider.err = errors.New("network error")

	processor := NewStreamProcessor(client)
	handler := &mockStreamHandler{}

	ctx := context.Background()
	err := processor.StreamDownload(ctx, "GET", "http://test.com/stream", handler)

	if err == nil {
		t.Error("Expected error from provider")
	}

	if len(handler.errors) == 0 {
		t.Error("Expected handler to receive error")
	}
}

func TestStreamProcessor_StreamDownloadHTTPError(t *testing.T) {
	client := &mockStreamClient{}
	client.provider.response = &interfaces.Response{
		StatusCode: 500,
		Body:       []byte("Internal Server Error"),
	}

	processor := NewStreamProcessor(client)
	handler := &mockStreamHandler{}

	ctx := context.Background()
	err := processor.StreamDownload(ctx, "GET", "http://test.com/stream", handler)

	if err == nil {
		t.Error("Expected HTTP error")
	}

	if len(handler.errors) == 0 {
		t.Error("Expected handler to receive error")
	}
}

func TestStreamProcessor_StreamUpload(t *testing.T) {
	client := &mockStreamClient{}
	client.provider.response = &interfaces.Response{
		StatusCode: 201,
		Body:       []byte("Upload successful"),
	}

	processor := NewStreamProcessor(client)
	reader := strings.NewReader("This is test upload data")

	ctx := context.Background()
	resp, err := processor.StreamUpload(ctx, "POST", "http://test.com/upload", reader, "text/plain")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestServerSentEventsHandler(t *testing.T) {
	events := []*SSEEvent{}
	errors := []error{}

	eventCallback := func(event *SSEEvent) {
		events = append(events, event)
	}

	errorCallback := func(err error) {
		errors = append(errors, err)
	}

	handler := NewServerSentEventsHandler(eventCallback, errorCallback)

	sseData := `id: 1
event: message
data: Hello, World!

id: 2
event: notification
data: This is a test
data: multiline message

`

	err := handler.OnData([]byte(sseData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].ID != "1" {
		t.Errorf("Expected event ID '1', got '%s'", events[0].ID)
	}

	if events[0].Event != "message" {
		t.Errorf("Expected event type 'message', got '%s'", events[0].Event)
	}

	if events[0].Data != "Hello, World!" {
		t.Errorf("Expected data 'Hello, World!', got '%s'", events[0].Data)
	}

	if events[1].Data != "This is a test\nmultiline message" {
		t.Errorf("Expected multiline data, got '%s'", events[1].Data)
	}
}

func TestChunkedHandler(t *testing.T) {
	chunks := [][]byte{}
	errors := []error{}

	chunkCallback := func(chunk []byte) {
		chunkCopy := make([]byte, len(chunk))
		copy(chunkCopy, chunk)
		chunks = append(chunks, chunkCopy)
	}

	errorCallback := func(err error) {
		errors = append(errors, err)
	}

	handler := NewChunkedHandler(chunkCallback, errorCallback)

	testData := [][]byte{
		[]byte("chunk1"),
		[]byte("chunk2"),
		[]byte("chunk3"),
	}

	for _, chunk := range testData {
		err := handler.OnData(chunk)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	handler.OnComplete()

	if len(chunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(chunks))
	}

	for i, chunk := range chunks {
		expected := testData[i]
		if string(chunk) != string(expected) {
			t.Errorf("Expected chunk %d to be '%s', got '%s'", i, string(expected), string(chunk))
		}
	}
}

func TestProgressHandler(t *testing.T) {
	totalSize := int64(100)
	progressUpdates := []struct {
		processed  int64
		total      int64
		percentage float64
	}{}

	progressCallback := func(processed, total int64, percentage float64) {
		progressUpdates = append(progressUpdates, struct {
			processed  int64
			total      int64
			percentage float64
		}{processed, total, percentage})
	}

	handler := NewProgressHandler(totalSize, progressCallback, nil)

	// Simulate receiving data in chunks
	chunks := [][]byte{
		make([]byte, 30), // 30%
		make([]byte, 50), // 80%
		make([]byte, 20), // 100%
	}

	for _, chunk := range chunks {
		err := handler.OnData(chunk)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	handler.OnComplete()

	if len(progressUpdates) != 4 { // 3 chunks + completion
		t.Errorf("Expected 4 progress updates, got %d", len(progressUpdates))
	}

	// Check final progress
	final := progressUpdates[len(progressUpdates)-1]
	if final.processed != totalSize {
		t.Errorf("Expected final processed size %d, got %d", totalSize, final.processed)
	}

	if final.percentage != 100.0 {
		t.Errorf("Expected final percentage 100.0, got %f", final.percentage)
	}
}

func TestFileHandler(t *testing.T) {
	buffer := &bytes.Buffer{}
	errors := []error{}

	errorCallback := func(err error) {
		errors = append(errors, err)
	}

	handler := NewFileHandler(buffer, errorCallback)

	testData := [][]byte{
		[]byte("Hello, "),
		[]byte("World!"),
		[]byte(" This is a test."),
	}

	for _, chunk := range testData {
		err := handler.OnData(chunk)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	handler.OnComplete()

	expected := "Hello, World! This is a test."
	if buffer.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buffer.String())
	}

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}
}

func TestCompositeHandler(t *testing.T) {
	handler1 := &mockStreamHandler{}
	handler2 := &mockStreamHandler{}

	composite := NewCompositeHandler(handler1, handler2)

	testData := []byte("test data")
	err := composite.OnData(testData)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	composite.OnComplete()

	// Both handlers should receive the data
	if len(handler1.receivedData) != 1 {
		t.Errorf("Expected handler1 to receive 1 chunk, got %d", len(handler1.receivedData))
	}

	if len(handler2.receivedData) != 1 {
		t.Errorf("Expected handler2 to receive 1 chunk, got %d", len(handler2.receivedData))
	}

	if !handler1.completed {
		t.Error("Expected handler1 to be completed")
	}

	if !handler2.completed {
		t.Error("Expected handler2 to be completed")
	}
}

func TestBufferedHandler(t *testing.T) {
	targetHandler := &mockStreamHandler{}
	bufferSize := 10

	buffered := NewBufferedHandler(bufferSize, targetHandler, nil)

	// Send data smaller than buffer size
	smallData := []byte("small")
	err := buffered.OnData(smallData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Target handler should not receive data yet
	if len(targetHandler.receivedData) != 0 {
		t.Error("Expected target handler to not receive data yet")
	}

	// Send more data to exceed buffer size
	moreData := []byte("more data")
	err = buffered.OnData(moreData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Target handler should now receive buffered data
	if len(targetHandler.receivedData) != 1 {
		t.Errorf("Expected target handler to receive 1 chunk, got %d", len(targetHandler.receivedData))
	}

	// Complete and flush remaining data
	buffered.OnComplete()

	if !targetHandler.completed {
		t.Error("Expected target handler to be completed")
	}
}
