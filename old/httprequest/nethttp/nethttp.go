package nethttp

import (
	"bytes"
	"context"
	"crypto/tls"
	errs "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/dock-tech/isis-golang-lib/domainerrors"
	jsoniter "github.com/json-iterator/go"
)

type Requester struct {
	client               *http.Client
	headers              map[string]string
	BaseURL              string
	structUnmarshal      any
	errorUnmarshal       any
	getConn              time.Time
	dnsStart             time.Time
	dnsDone              time.Time
	connectDone          time.Time
	tlsHandshakeStart    time.Time
	tlsHandshakeDone     time.Time
	gotConn              time.Time
	gotFirstResponseByte time.Time
	endTime              time.Time
	gotConnInfo          httptrace.GotConnInfo
	clientTracerEnabled  bool
}

type Response struct {
	Request    Requester
	Body       []byte
	StatusCode int
	IsError    bool
}

type IHttpRequester interface {
	Execute(ctx context.Context, method, url string, body io.Reader) (*Response, error)
	Patch(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Head(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Post(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Put(ctx context.Context, endpoint string, body []byte) (*Response, error)
	Delete(ctx context.Context, endpoint string) (*Response, error)
	Get(ctx context.Context, endpoint string) (*Response, error)
	SetHeaders(headers map[string]string) *Requester
	GetHeaders() map[string]string
	SetBaseURL(baseURL string) *Requester
	GetBaseURL() string
	Unmarshal(v any) *Requester
	GetStructUnmarshal() any
	ErrorUnmarshal(v any) *Requester
	GetErrorUnmarshal() any
	TraceInfo() TraceInfo
	Close(response *http.Response) error
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func NewRequester(client IClient, options ...NetHttpClientConfig) IHttpRequester {

	cfg := &netHttpClientConfig{}

	defaultsClient(cfg)
	for _, opt := range options {
		opt(cfg)
	}

	return &Requester{
		client:              client.GetClient(),
		clientTracerEnabled: cfg.clientTracerEnabled,
	}
}

// SetHeaders method sets multiple headers field and its values at one go in the client instance.
// These headers will be applied to all requests raised from this client instance. Also it can be
// overridden at request level headers options.
// For Example: To set `Content-Type` and `Accept` as `application/json`
//
//	request.SetHeaders(map[string]string{
//			"Content-Type": "application/json",
//			"Accept": "application/json",
//		})
func (r *Requester) SetHeaders(headers map[string]string) *Requester {
	r.headers = headers
	return r
}

func (r *Requester) GetHeaders() map[string]string {
	return r.headers
}

// SetErrorHandler method is to register the response `ErrorHandler` for current `Request`.
func (r *Requester) SetBaseURL(baseURL string) *Requester {
	r.BaseURL = baseURL
	return r
}

func (r *Requester) GetBaseURL() string {
	return r.BaseURL
}

// Head method performs the HTTP HEAD request for current `Request`.
func (r *Requester) Head(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodHead, endpoint, bytes.NewBuffer(body))
}

// Patch method performs the HTTP PATCH request for current `Request`.
func (r *Requester) Patch(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodPatch, endpoint, bytes.NewBuffer(body))
}

// Post method performs the HTTP POST request for current `Request`.
func (r *Requester) Post(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
}

// Get method performs the HTTP GET request for current `Request`.
func (r *Requester) Get(ctx context.Context, endpoint string) (*Response, error) {
	return r.Execute(ctx, http.MethodGet, endpoint, nil)
}

// Put method performs the HTTP PUT request for current `Request`.
func (r *Requester) Put(ctx context.Context, endpoint string, body []byte) (*Response, error) {
	return r.Execute(ctx, http.MethodPut, endpoint, bytes.NewBuffer(body))
}

// Delete method performs the HTTP DELETE request for current `Request`.
func (r *Requester) Delete(ctx context.Context, endpoint string) (*Response, error) {
	return r.Execute(ctx, http.MethodDelete, endpoint, nil)
}

// Unmarshal method unmarshal the HTTP response body to given struct.
func (r *Requester) Unmarshal(v any) *Requester {
	r.structUnmarshal = v
	return r
}

func (r *Requester) GetStructUnmarshal() any {
	return r.structUnmarshal
}

// ErrorUnmarshal method unmarshals the HTTP response when the status code do not match body to given struct.
func (r *Requester) ErrorUnmarshal(v any) *Requester {
	r.errorUnmarshal = v
	return r
}

func (r *Requester) GetErrorUnmarshal() any {
	return r.errorUnmarshal
}

// Close method closes the HTTP client connection.
func (r *Requester) Close(response *http.Response) error {
	if response != nil {
		return response.Body.Close()
	}
	return nil
}

func (r *Requester) Execute(ctx context.Context, method, url string, body io.Reader) (response *Response, err error) {
	var reqs *http.Request

	ddSpan, ok := tracer.SpanFromContext(ctx)
	defer ddSpan.Finish()

	if ok {
		err = tracer.Inject(ddSpan.Context(), tracer.TextMapCarrier(r.headers))
		if err != nil {
			return nil, err
		}
	}

	uriREquest := r.BaseURL + url

	if r.clientTracerEnabled {
		clientTracer := &httptrace.ClientTrace{
			DNSStart: func(dnsstartInfo httptrace.DNSStartInfo) {
				r.dnsStart = time.Now()
			},
			DNSDone: func(dnsinfo httptrace.DNSDoneInfo) {
				r.dnsDone = time.Now()
			},
			ConnectStart: func(network, addr string) {
				if r.dnsDone.IsZero() {
					r.dnsDone = time.Now()
				}
				if r.dnsStart.IsZero() {
					r.dnsStart = r.dnsDone
				}
			},
			ConnectDone: func(net, addr string, err error) {
				r.connectDone = time.Now()
			},
			GetConn: func(hostPort string) {
				r.getConn = time.Now()
			},
			GotConn: func(ci httptrace.GotConnInfo) {
				r.gotConn = time.Now()
				r.gotConnInfo = ci
			},
			GotFirstResponseByte: func() {
				r.gotFirstResponseByte = time.Now()
			},
			TLSHandshakeStart: func() {
				r.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(tlscon tls.ConnectionState, errr error) {
				r.tlsHandshakeDone = time.Now()
			},
		}

		reqs, err = http.NewRequestWithContext(httptrace.WithClientTrace(ctx, clientTracer), method, uriREquest, body)
		if err != nil {
			return nil, err
		}

	} else {

		reqs, err = http.NewRequest(method, uriREquest, body)
		if err != nil {
			return nil, err
		}
	}

	if r.headers != nil {
		for k, v := range r.headers {
			reqs.Header.Set(k, v)
		}
	}

	reqs.Header.Set("Content-Type", "application/json")

	isErrors := false
	resp, err := r.client.Do(reqs)
	if err != nil {
		isErrors = true
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		isErrors = true
		return nil, err
	}
	// read response body

	if r.structUnmarshal != nil {
		err := json.Unmarshal(respBody, r.structUnmarshal)
		if err != nil {
			isErrors = true
			return nil, err
		}
	}

	response = &Response{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		IsError:    isErrors,
	}

	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, err
	}

	if r.clientTracerEnabled {
		ti := r.TraceInfo()

		jsonTracer := `{"DNSLookup":"%v","URI":"%s","RemoteAddr":"%v","LocalAddr":"%v","ConnTime":"%v", "TCPConnTime":"%v",` +
			`"TLSHandshake":"%v","ServerTime":"%v","ResponseTime":"%v","TotalTime":"%v","IsConnReused":"%v","IsConnWasIdle":"%v",` +
			`"ConnIdleTime":"%v"}`

		fmt.Println(fmt.Sprintf(jsonTracer, ti.DNSLookup, uriREquest, ti.RemoteAddr, ti.LocalAddr, ti.ConnTime, ti.TCPConnTime, ti.TLSHandshake, ti.ServerTime, ti.ResponseTime, ti.TotalTime, ti.IsConnReused, ti.IsConnWasIdle, ti.ConnIdleTime))
	}
	return response, err
}

func WrapErrors(err error) error {
	if err == nil {
		return nil
	}

	v, ok := err.(*url.Error)
	if !ok {
		return &domainerrors.ServerError{InternalError: err}
	}

	if v.Timeout() {
		return &domainerrors.TimeoutError{}
	}

	var netErr *net.OpError
	if errs.As(v, &netErr) {
		return &domainerrors.ErrTargetServiceUnavailable{InternalError: err}
	}

	return err
}

// ‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// TraceInfo struct
// _______________________________________________________________________

// TraceInfo struct is used provide request trace info such as DNS lookup
// duration, Connection obtain duration, Server processing duration, etc.
//
// Since v2.0.0.
type TraceInfo struct {
	// DNSLookup is a duration that transport took to perform
	// DNS lookup.
	DNSLookup time.Duration

	// ConnTime is a duration that took to obtain a successful connection.
	ConnTime time.Duration

	// TCPConnTime is a duration that took to obtain the TCP connection.
	TCPConnTime time.Duration

	// TLSHandshake is a duration that TLS handshake took place.
	TLSHandshake time.Duration

	// ServerTime is a duration that server took to respond first byte.
	ServerTime time.Duration

	// ResponseTime is a duration since first response byte from server to
	// request completion.
	ResponseTime time.Duration

	// TotalTime is a duration that total request took end-to-end.
	TotalTime time.Duration

	// IsConnReused is whether this connection has been previously
	// used for another HTTP request.
	IsConnReused bool

	// IsConnWasIdle is whether this connection was obtained from an
	// idle pool.
	IsConnWasIdle bool

	// ConnIdleTime is a duration how long the connection was previously
	// idle, if IsConnWasIdle is true.
	ConnIdleTime time.Duration

	// RequestAttempt is to represent the request attempt made during a Resty
	// request execution flow, including retry count.
	RequestAttempt int

	// RemoteAddr returns the remote network address.
	RemoteAddr net.Addr

	// LocalAddr returns the local network address.
	LocalAddr net.Addr
}

func (r *Requester) TraceInfo() TraceInfo {
	if r == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		DNSLookup:      r.dnsDone.Sub(r.dnsStart),
		TLSHandshake:   r.tlsHandshakeDone.Sub(r.tlsHandshakeStart),
		ServerTime:     r.gotFirstResponseByte.Sub(r.gotConn),
		IsConnReused:   r.gotConnInfo.Reused,
		IsConnWasIdle:  r.gotConnInfo.WasIdle,
		ConnIdleTime:   r.gotConnInfo.IdleTime,
		RemoteAddr:     r.gotConnInfo.Conn.RemoteAddr(),
		LocalAddr:      r.gotConnInfo.Conn.RemoteAddr(),
		RequestAttempt: 0,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if r.gotConnInfo.Reused {
		ti.TotalTime = r.endTime.Sub(r.getConn)
	} else {
		ti.TotalTime = r.endTime.Sub(r.dnsStart)
	}

	// Only calculate on successful connections
	if !r.connectDone.IsZero() {
		ti.TCPConnTime = r.connectDone.Sub(r.dnsDone)
	}

	// Only calculate on successful connections
	if !r.gotConn.IsZero() {
		ti.ConnTime = r.gotConn.Sub(r.getConn)
	}

	// Only calculate on successful connections
	if !r.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = r.endTime.Sub(r.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if r.gotConnInfo.Conn != nil {
		ti.RemoteAddr = r.gotConnInfo.Conn.RemoteAddr()
		ti.LocalAddr = r.gotConnInfo.Conn.LocalAddr()
	}

	return ti
}
