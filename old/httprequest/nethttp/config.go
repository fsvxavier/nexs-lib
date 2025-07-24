package nethttp

import (
	"time"
)

type NetHttpClientConfig func(*netHttpClientConfig)

type netHttpClientConfig struct {
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	idleConnTimeout     time.Duration
	tlsEnabled          bool
	disableKeepAlives   bool
	clientTracerEnabled bool
	clientTimeout       time.Duration
}

func GetClientConfig() netHttpClientConfig {
	return netHttpClientConfig{}
}

func defaultsClient(cfg *netHttpClientConfig) {

	cfg.maxConnsPerHost = 60
	cfg.maxIdleConns = 40
	cfg.maxIdleConnsPerHost = 50
	cfg.idleConnTimeout = time.Minute * 1440
	cfg.clientTimeout = 3 * time.Second
	cfg.tlsEnabled = false
	cfg.disableKeepAlives = false
	cfg.clientTracerEnabled = false

}

func SetTlsEnabled(enabled bool) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.tlsEnabled = enabled
	}
}

func SetMaxIdleConnsPerHost(maxIdleConnsPerHost int) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func SetMaxIdleConns(maxIdleConns int) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.maxIdleConns = maxIdleConns
	}
}

func SetMaxConns(maxConns int) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.maxConnsPerHost = maxConns
	}
}

func SetDisableKeepAlives(disable bool) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.disableKeepAlives = disable
	}
}

func SetIdleConnTimeout(idleConnTimeout time.Duration) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.idleConnTimeout = idleConnTimeout
	}
}

func SetTracerEnabled(enabled bool) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.clientTracerEnabled = enabled
	}
}

func SetClientTimeout(timeout time.Duration) NetHttpClientConfig {
	return func(do *netHttpClientConfig) {
		do.clientTimeout = timeout
	}
}
