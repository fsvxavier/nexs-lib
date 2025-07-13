package nethttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type IClient interface {
	GetClient() *http.Client
}

type Client struct {
	client  *http.Client
	Timeout time.Duration
}

func New(options ...NetHttpClientConfig) IClient {

	cfg := &netHttpClientConfig{}

	defaultsClient(cfg)
	for _, opt := range options {
		opt(cfg)
	}

	transport := &http.Transport{
		Dial: (&net.Dialer{
			DualStack:     false,
			FallbackDelay: 0,
			Timeout:       cfg.clientTimeout,
			KeepAlive:     0, // Set to 0 to enable keep-alives if supported by the protocol and operating system.
		}).Dial,
		MaxIdleConns:        cfg.maxIdleConns,
		MaxIdleConnsPerHost: cfg.maxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.maxConnsPerHost,
		IdleConnTimeout:     cfg.idleConnTimeout,
		DisableKeepAlives:   cfg.disableKeepAlives,
	}

	if !cfg.tlsEnabled {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := &Client{
		Timeout: cfg.clientTimeout,
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.clientTimeout,
		},
	}

	return client
}

// GetClient returns the underlying HTTP client.
func (c *Client) GetClient() *http.Client {
	return c.client
}
