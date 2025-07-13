package nethttp

import (
	"net/http"
	"testing"
	"time"
)

func TestNew_DefaultConfig(t *testing.T) {
	client := New()
	httpClient := client.GetClient()
	if httpClient == nil {
		t.Fatal("expected non-nil http.Client")
	}
}

func TestNew_CustomTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	client := New(func(cfg *netHttpClientConfig) {
		cfg.clientTimeout = customTimeout
	})
	c, ok := client.(*Client)
	if !ok {
		t.Fatal("expected *Client type")
	}
	if c.Timeout != customTimeout {
		t.Errorf("expected Timeout %v, got %v", customTimeout, c.Timeout)
	}
	if c.client.Timeout != customTimeout {
		t.Errorf("expected http.Client.Timeout %v, got %v", customTimeout, c.client.Timeout)
	}
}

func TestNew_TLSDisabled(t *testing.T) {
	client := New(func(cfg *netHttpClientConfig) {
		cfg.tlsEnabled = false
	})
	c, ok := client.(*Client)
	if !ok {
		t.Fatal("expected *Client type")
	}
	transport, ok := c.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.TLSClientConfig == nil {
		t.Error("expected TLSClientConfig to be set when tlsEnabled is false")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true when tlsEnabled is false")
	}
}

func TestGetClient_ReturnsSameInstance(t *testing.T) {
	client := New()
	c, ok := client.(*Client)
	if !ok {
		t.Fatal("expected *Client type")
	}
	if c.GetClient() != c.client {
		t.Error("GetClient should return the underlying http.Client instance")
	}
}
