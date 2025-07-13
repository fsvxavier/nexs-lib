package redis

import (
	"crypto/tls"
	"time"
)

const (
	CONNECT_TIMEOUT_SEC = 5
)

type RedigoConfig func(*redigoConfig)

type redigoConfig struct {
	tlsConfig        *tls.Config
	password         string
	traceServiceName string
	clientName       string
	addresses        []string
	maxConnLifetime  time.Duration
	idleTimeout      time.Duration
	maxIdle          int
	poolSize         int
	database         int
	maxActive        int
	skipVerify       bool
	usageTLS         bool
	executePing      bool
}

func GetConfig() *redigoConfig {
	return &redigoConfig{
		tlsConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		},
		password:         "",
		traceServiceName: "redis.db",
		clientName:       "",
		addresses:        []string{},
		maxConnLifetime:  30 * time.Second,
		idleTimeout:      20 * time.Second,
		maxIdle:          20,
		poolSize:         40,
		database:         0,
		maxActive:        20,
		skipVerify:       true,
		usageTLS:         false,
		executePing:      false,
	}
}

func SetPassword(password *string) RedigoConfig {
	return func(do *redigoConfig) {
		if password != nil {
			do.password = *password
		}
	}
}

func SetTraceServiceName(traceServiceName *string) RedigoConfig {
	return func(do *redigoConfig) {
		if traceServiceName != nil {
			do.traceServiceName = *traceServiceName
		}
	}
}

func SetClientName(clientName *string) RedigoConfig {
	return func(do *redigoConfig) {
		if clientName != nil {
			do.clientName = *clientName
		}
	}
}

func SetAddresses(addresses *[]string) RedigoConfig {
	return func(do *redigoConfig) {
		if addresses != nil {
			do.addresses = *addresses
		}
	}
}

func SetIdleTimeout(numIdleTimeout *time.Duration) RedigoConfig {
	return func(do *redigoConfig) {
		if numIdleTimeout != nil {
			do.idleTimeout = *numIdleTimeout
		}
	}
}

func SetMaxConnLifetime(numMaxConnLifetime *time.Duration) RedigoConfig {
	return func(do *redigoConfig) {
		if numMaxConnLifetime != nil {
			do.maxConnLifetime = *numMaxConnLifetime
		}
	}
}

func SetMaxIdle(numMaxIdle *int) RedigoConfig {
	return func(do *redigoConfig) {
		if numMaxIdle != nil {
			do.maxIdle = *numMaxIdle
		}
	}
}

func SetPoolSize(numPoolSize *int) RedigoConfig {
	return func(do *redigoConfig) {
		if numPoolSize != nil {
			do.poolSize = *numPoolSize
		}
	}
}

func SetDatabase(numDatabase *int) RedigoConfig {
	return func(do *redigoConfig) {
		if numDatabase != nil {
			do.database = *numDatabase
		}
	}
}

func SetMaxActive(numMaxActive *int) RedigoConfig {
	return func(do *redigoConfig) {
		if numMaxActive != nil {
			do.maxActive = *numMaxActive
		}
	}
}

func SetSkipVerify(enabled *bool) RedigoConfig {
	return func(do *redigoConfig) {
		if enabled != nil {
			do.skipVerify = *enabled
		}
	}
}

func SetUsageTLS(enabled *bool) RedigoConfig {
	return func(do *redigoConfig) {
		if enabled != nil {
			do.usageTLS = *enabled
			do.tlsConfig = &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			}
		}
	}
}

func SetExecutePing(enabled *bool) RedigoConfig {
	return func(do *redigoConfig) {
		if enabled != nil {
			do.executePing = *enabled
		}
	}
}
