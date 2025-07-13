package nethttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"testing"
	"time"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock types and helpers ---

type mockClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

type mockIClient struct {
	client *mockClient
}

func (m *mockIClient) GetClient() *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(m.client.doFunc),
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// --- Tests ---

func TestSetAndGetHeaders(t *testing.T) {
	r := &Requester{}
	headers := map[string]string{"X-Test": "value"}
	r.SetHeaders(headers)
	got := r.GetHeaders()
	if got["X-Test"] != "value" {
		t.Errorf("expected header X-Test to be 'value', got '%s'", got["X-Test"])
	}
}

func TestSetAndGetBaseURL(t *testing.T) {
	r := &Requester{}
	baseURL := "http://example.com"
	r.SetBaseURL(baseURL)
	if r.GetBaseURL() != baseURL {
		t.Errorf("expected baseURL '%s', got '%s'", baseURL, r.GetBaseURL())
	}
}

func TestUnmarshalAndGetStructUnmarshal(t *testing.T) {
	r := &Requester{}
	var v struct{}
	r.Unmarshal(&v)
	if r.GetStructUnmarshal() != &v {
		t.Errorf("expected structUnmarshal to be set")
	}
}

func TestErrorUnmarshalAndGetErrorUnmarshal(t *testing.T) {
	r := &Requester{}
	var v struct{}
	r.ErrorUnmarshal(&v)
	if r.GetErrorUnmarshal() != &v {
		t.Errorf("expected errorUnmarshal to be set")
	}
}

func TestRequester_Get(t *testing.T) {
	body := `{"foo":"bar"}`
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		},
	}
	r := &Requester{client: &http.Client{Transport: roundTripperFunc(client.doFunc)}}
	resp, err := r.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Body) != body {
		t.Errorf("expected body '%s', got '%s'", body, string(resp.Body))
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequester_Post(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(req.Body)
			if string(b) != `{"foo":"bar"}` {
				t.Errorf("expected body to be sent")
			}
			return &http.Response{
				StatusCode: 201,
				Body:       io.NopCloser(bytes.NewBufferString("ok")),
				Header:     make(http.Header),
			}, nil
		},
	}
	r := &Requester{client: &http.Client{Transport: roundTripperFunc(client.doFunc)}}
	resp, err := r.Post(context.Background(), "/test", []byte(`{"foo":"bar"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestRequester_Execute_Error(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("fail")
		},
	}
	r := &Requester{client: &http.Client{Transport: roundTripperFunc(client.doFunc)}}
	_, err := r.Execute(context.Background(), http.MethodGet, "/fail", nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestRequester_Execute_Unmarshal(t *testing.T) {
	type foo struct{ Foo string }
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"Foo":"bar"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}
	r := &Requester{client: &http.Client{Transport: roundTripperFunc(client.doFunc)}}
	var f foo
	r.Unmarshal(&f)
	_, err := r.Execute(context.Background(), http.MethodGet, "/foo", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Foo != "bar" {
		t.Errorf("expected Foo to be 'bar', got '%s'", f.Foo)
	}
}

func TestRequester_Execute_UnmarshalError(t *testing.T) {
	type foo struct{ Bar int }
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`notjson`)),
				Header:     make(http.Header),
			}, nil
		},
	}
	r := &Requester{client: &http.Client{Transport: roundTripperFunc(client.doFunc)}}
	var f foo
	r.Unmarshal(&f)
	_, err := r.Execute(context.Background(), http.MethodGet, "/foo", nil)
	if err == nil {
		t.Errorf("expected unmarshal error, got nil")
	}
}

// func TestTraceInfo_Defaults(t *testing.T) {
// 	r := &Requester{}
// 	ti := r.TraceInfo()
// 	if ti.DNSLookup != 0 {
// 		t.Errorf("expected DNSLookup 0, got %v", ti.DNSLookup)
// 	}
// }

func TestTraceInfo_NilReceiver(t *testing.T) {
	var r *Requester
	ti := r.TraceInfo()
	if ti.DNSLookup != 0 {
		t.Errorf("expected DNSLookup 0, got %v", ti.DNSLookup)
	}
}

func TestRequester_TraceInfo_Fields(t *testing.T) {
	now := time.Now()
	r := &Requester{
		dnsStart:             now,
		dnsDone:              now.Add(10 * time.Millisecond),
		tlsHandshakeStart:    now.Add(11 * time.Millisecond),
		tlsHandshakeDone:     now.Add(21 * time.Millisecond),
		gotConn:              now.Add(22 * time.Millisecond),
		gotFirstResponseByte: now.Add(32 * time.Millisecond),
		endTime:              now.Add(42 * time.Millisecond),
		getConn:              now.Add(5 * time.Millisecond),
		connectDone:          now.Add(15 * time.Millisecond),
		gotConnInfo: httptrace.GotConnInfo{
			Reused:   true,
			WasIdle:  true,
			IdleTime: 123 * time.Millisecond,
			Conn:     &net.TCPConn{},
		},
	}
	ti := r.TraceInfo()
	if ti.DNSLookup != 10*time.Millisecond {
		t.Errorf("expected DNSLookup 10ms, got %v", ti.DNSLookup)
	}
	if !ti.IsConnReused || !ti.IsConnWasIdle {
		t.Errorf("expected connection reused and idle")
	}
}

type mockResponse struct {
	Message string `json:"message"`
}

func setupServer(t *testing.T, port string) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadBufferSize:        4096,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "get success"})
	})

	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "post success"})
	})

	app.Put("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "put success"})
	})

	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "delete success"})
	})

	app.Patch("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "patch success"})
	})

	app.Head("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{
			"error": "internal server error",
		})
	})

	app.Get("/timeout", func(c *fiber.Ctx) error {
		time.Sleep(2 * time.Second)
		return c.JSON(mockResponse{Message: "timeout"})
	})

	go func() {
		if err := app.Listen(":" + port); err != nil {
			t.Logf("Error starting server: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return app
}

func TestRequesterHTTPMethods(t *testing.T) {
	app := setupServer(t, "3000")
	defer app.Shutdown()

	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           []byte
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "GET Request",
			method:         http.MethodGet,
			endpoint:       "/test",
			expectedStatus: 200,
			expectedMsg:    "get success",
		},
		{
			name:           "POST Request",
			method:         http.MethodPost,
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedMsg:    "post success",
		},
		{
			name:           "PUT Request",
			method:         http.MethodPut,
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedMsg:    "put success",
		},
		{
			name:           "DELETE Request",
			method:         http.MethodDelete,
			endpoint:       "/test",
			expectedStatus: 200,
			expectedMsg:    "delete success",
		},
		{
			name:           "PATCH Request",
			method:         http.MethodPatch,
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedMsg:    "patch success",
		},
		{
			name:           "HEAD Request",
			method:         http.MethodHead,
			endpoint:       "/test",
			expectedStatus: 200,
		},
	}

	client := New(func(cfg *netHttpClientConfig) {
		cfg.clientTimeout = 5 * time.Second
	})

	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3000")

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *Response
			var err error

			switch tt.method {
			case http.MethodGet:
				resp, err = req.Get(ctx, tt.endpoint)
			case http.MethodPost:
				resp, err = req.Post(ctx, tt.endpoint, tt.body)
			case http.MethodPut:
				resp, err = req.Put(ctx, tt.endpoint, tt.body)
			case http.MethodDelete:
				resp, err = req.Delete(ctx, tt.endpoint)
			case http.MethodPatch:
				resp, err = req.Patch(ctx, tt.endpoint, tt.body)
			case http.MethodHead:
				resp, err = req.Head(ctx, tt.endpoint, tt.body)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedMsg != "" {
				var result mockResponse
				err = json.Unmarshal(resp.Body, &result)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMsg, result.Message)
			}
		})
	}
}

func TestRequesterConfiguration(t *testing.T) {
	client := New()
	req := NewRequester(client)

	t.Run("SetHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token",
			"Custom-Header": "test-value",
		}
		req.SetHeaders(headers)
		assert.Equal(t, headers, req.GetHeaders())
	})

	t.Run("SetBaseURL", func(t *testing.T) {
		baseURL := "http://localhost:3000"
		req.SetBaseURL(baseURL)
		assert.Equal(t, baseURL, req.GetBaseURL())
	})

	t.Run("Unmarshal", func(t *testing.T) {
		var data mockResponse
		req.Unmarshal(&data)
		assert.NotNil(t, req.GetStructUnmarshal())
	})

	t.Run("ErrorUnmarshal", func(t *testing.T) {
		var errData mockResponse
		req.ErrorUnmarshal(&errData)
		assert.NotNil(t, req.GetErrorUnmarshal())
	})
}

func TestTracing(t *testing.T) {
	app := setupServer(t, "3001")
	defer app.Shutdown()

	client := New(func(cfg *netHttpClientConfig) {
		cfg.clientTimeout = 5 * time.Second
	})

	req := NewRequester(client, func(cfg *netHttpClientConfig) {
		cfg.clientTracerEnabled = true
	})
	req.SetBaseURL("http://localhost:3001")

	resp, err := req.Get(context.Background(), "/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	traceInfo := req.TraceInfo()
	assert.NotZero(t, traceInfo.ServerTime)
}

func TestTimeout(t *testing.T) {
	app := setupServer(t, "3002")
	defer app.Shutdown()

	client := New(func(cfg *netHttpClientConfig) {
		cfg.clientTimeout = 1 * time.Second
	})

	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3002")

	_, err := req.Get(context.Background(), "/timeout")
	assert.Error(t, err)
}

func TestErrorResponse(t *testing.T) {
	app := setupServer(t, "3003")
	defer app.Shutdown()

	client := New(func(cfg *netHttpClientConfig) {
		cfg.clientTimeout = 5 * time.Second
	})

	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3003")

	resp, err := req.Get(context.Background(), "/error")
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var errResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(resp.Body, &errResp)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", errResp.Error)
}

func TestRequester_Close(t *testing.T) {
	r := &Requester{}

	t.Run("Close with nil response returns nil", func(t *testing.T) {
		err := r.Close(nil)
		assert.NoError(t, err)
	})

	t.Run("Close returns error from Body.Close", func(t *testing.T) {
		resp := &http.Response{
			Body: &errorCloser{},
		}
		err := r.Close(resp)
		assert.Error(t, err)
		assert.EqualError(t, err, "close error")
	})
}

func TestWrapErrors_ReturnsNilOnNilError(t *testing.T) {
	err := WrapErrors(nil)
	assert.Nil(t, err)
}

type dummyError struct{}

func (dummyError) Error() string { return "dummy" }

func TestWrapErrors_ReturnsServerErrorOnNonURLError(t *testing.T) {
	err := WrapErrors(dummyError{})
	serverErr, ok := err.(*domainerrors.ServerError)
	assert.True(t, ok)
	assert.Equal(t, "dummy", serverErr.InternalError.Error())
}

func TestWrapErrors_ReturnsTimeoutErrorOnTimeoutURLError(t *testing.T) {
	urlErr := &url.Error{
		Err: timeoutMockError{},
	}
	err := WrapErrors(urlErr)
	_, ok := err.(*domainerrors.TimeoutError)
	assert.True(t, ok)
}

func TestWrapErrors_ReturnsErrTargetServiceUnavailableOnNetOpError(t *testing.T) {
	netOpErr := &net.OpError{}
	urlErr := &url.Error{
		Err: netOpErr,
	}
	err := WrapErrors(urlErr)
	_, ok := err.(*domainerrors.ErrTargetServiceUnavailable)
	assert.True(t, ok)
}

func TestWrapErrors_ReturnsOriginalURLErrorIfNoSpecialCase(t *testing.T) {
	urlErr := &url.Error{
		Err: dummyError{},
	}
	err := WrapErrors(urlErr)
	assert.Equal(t, urlErr, err)
}

type closeTracker struct {
	onClose func()
}

func (c *closeTracker) Read(p []byte) (int, error) { return 0, io.EOF }
func (c *closeTracker) Close() error {
	if c.onClose != nil {
		c.onClose()
	}
	return nil
}

type errorCloser struct{}

func (e *errorCloser) Read(p []byte) (int, error) { return 0, io.EOF }
func (e *errorCloser) Close() error               { return errors.New("close error") }

// timeoutMockError implements Timeout() bool for testing
type timeoutMockError struct{}

func (timeoutMockError) Error() string   { return "timeout" }
func (timeoutMockError) Timeout() bool   { return true }
func (timeoutMockError) Temporary() bool { return false }
