package httprequest

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

var ctx context.Context

type Request struct {
	ctx        context.Context
	client     *resty.Client
	restyReq   *resty.Request
	restyRes   *resty.Response
	errHandler ErrorHandler
	baseURL    string
}

type Response struct {
	Body       []byte
	StatusCode int
	IsError    bool
}

type IHttpRequest interface {
	Get(endpoint string) (*Response, error)
	Post(endpoint string, body interface{}) (*Response, error)
	Put(endpoint string, body interface{}) (*Response, error)
	Delete(endpoint string) (*Response, error)
}

type ErrorHandler func(*Response) error

// New method creates a new httprequest client.
func New(url string) *Request {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	client := resty.New()
	client.JSONMarshal = json.Marshal
	client.JSONUnmarshal = json.Unmarshal
	client.DisableTrace()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	client.SetBaseURL(url)

	HttpRequest := &Request{ctx, client, client.R(), nil, nil, url}

	return HttpRequest
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
func (req *Request) SetHeaders(headers map[string]string) *Request {
	req.client.SetHeaders(headers)

	return req
}

// SetErrorHandler method is to register the response `ErrorHandler` for current `Request`.
func (req *Request) SetErrorHandler(h ErrorHandler) *Request {
	req.errHandler = h

	return req
}

func (req *Request) SetContext(ctx context.Context) *Request {
	req.ctx = ctx
	return req
}

// Post method performs the HTTP POST request for current `Request`.
func (req *Request) Post(endpoint string, body interface{}) (*Response, error) {
	return req.Execute("POST", endpoint, body)
}

// Get method performs the HTTP GET request for current `Request`.
func (req *Request) Get(endpoint string) (*Response, error) {
	return req.Execute("GET", endpoint, nil)
}

// Put method performs the HTTP PUT request for current `Request`.
func (req *Request) Put(endpoint string, body interface{}) (*Response, error) {
	return req.Execute("PUT", endpoint, body)
}

// Delete method performs the HTTP DELETE request for current `Request`.
func (req *Request) Delete(endpoint string) (*Response, error) {
	return req.Execute("DELETE", endpoint, nil)
}

// Unmarshal method unmarshals the HTTP response body to given struct.
func (req *Request) Unmarshal(v any) *Request {
	req.restyReq.SetResult(v)

	return req
}

// Execute method performs the HTTP request with given HTTP method, Endpoint and Body for current `Request`.
//
//	resp, err := httprequest.New("http://httpbin.org").Execute("GET", "/get", nil)
func (req *Request) Execute(method, endpoint string, body interface{}) (*Response, error) {
	span, ctxs := tracer.StartSpanFromContext(req.ctx, "post.process")
	defer span.Finish()

	method = strings.ToUpper(method)
	rreq := req.restyReq

	if body != nil {
		rreq.SetBody(body)
	}

	rreq.SetContext(ctxs)
	// Inject the span Context in the Request headers
	err := tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.client.Header))
	if err != nil {
		return nil, err
	}
	rres, err := rreq.Execute(method, endpoint)
	if err != nil {
		return nil, err
	}

	res := parseResponse(rres)

	var respError error

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		respError = fmt.Errorf("%d-%s", res.StatusCode, string(res.Body))
		if req.errHandler != nil {
			respError = req.errHandler(res)
		}
	}

	return res, respError
}

// Execute method performs the raw Resty HTTP request with given HTTP method, URL, Headers, Query and Payload for current `Request`.
func RawExecute(method, url, payload string, headers, query map[string]string) (int, string, map[string]string, error) {
	var res *resty.Response
	var err error
	client := resty.New()

	req := client.R().
		SetHeaders(headers).
		SetQueryParams(query).
		SetBody(payload)

	res, err = req.Execute(strings.ToUpper(method), url)

	responseHeaders := make(map[string]string)
	for k, v := range res.Header() {
		responseHeaders[k] = strings.Join(v, ", ")
	}

	return res.StatusCode(), string(res.Body()), responseHeaders, err
}

func parseResponse(res *resty.Response) *Response {
	return &Response{res.Body(), res.StatusCode(), res.IsError()}
}
