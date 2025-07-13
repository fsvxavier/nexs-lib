package gpgx

import (
	"time"
)

type PgxConfig func(*pgxConfig)

type pgxConfig struct {
	maxConns           int32
	minConns           int32
	maxConnLifetime    time.Duration
	maxConnIdletime    time.Duration
	datadogEnabled     bool
	rlsEnabled         bool
	queryTracerEnabled bool
	multiTenantEnabled bool
	connString         string
}

func GetConfig() *pgxConfig {
	return &pgxConfig{
		maxConns:           40,
		minConns:           2,
		maxConnLifetime:    time.Second * 9,
		maxConnIdletime:    time.Second * 3,
		rlsEnabled:         false,
		queryTracerEnabled: false,
		datadogEnabled:     false,
		multiTenantEnabled: false,
	}
}

func SetMultiTenantEnabled(enabled bool) PgxConfig {
	return func(do *pgxConfig) {
		do.multiTenantEnabled = enabled
	}
}

func SetDatadogEnabled(enabled bool) PgxConfig {
	return func(do *pgxConfig) {
		do.datadogEnabled = enabled
	}
}

func SetRlsEnabled(enabled bool) PgxConfig {
	return func(do *pgxConfig) {
		do.rlsEnabled = enabled
	}
}

func SetQueryTracerEnabled(enabled bool) PgxConfig {
	return func(do *pgxConfig) {
		do.queryTracerEnabled = enabled
	}
}

func SetMaxConns(maxConns *int32) PgxConfig {
	return func(do *pgxConfig) {
		if maxConns != nil {
			do.maxConns = *maxConns
		}
	}
}

func SetMinConns(minConns *int32) PgxConfig {
	return func(do *pgxConfig) {
		if minConns != nil {
			do.minConns = *minConns
		}
	}
}

func SetMaxConnLifetime(maxConnLifetime *time.Duration) PgxConfig {
	return func(do *pgxConfig) {
		if maxConnLifetime != nil {
			do.maxConnLifetime = *maxConnLifetime
		}
	}
}

func SetMaxConnIdleTime(maxConnIdletime *time.Duration) PgxConfig {
	return func(do *pgxConfig) {
		if maxConnIdletime != nil {
			do.maxConnIdletime = *maxConnIdletime
		}
	}
}

func SetConnectionStrings(cfg *pgxConfig, connString string) PgxConfig {
	return func(do *pgxConfig) {
		if connString != "" {
			do.connString = connString
		}
	}
}
