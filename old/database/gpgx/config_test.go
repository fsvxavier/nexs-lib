package gpgx

import (
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	cfg := GetConfig()

	if cfg.maxConns != 40 {
		t.Errorf("Expected maxConns to be 40, got %d", cfg.maxConns)
	}
	if cfg.minConns != 2 {
		t.Errorf("Expected minConns to be 2, got %d", cfg.minConns)
	}
	if cfg.maxConnLifetime != time.Second*9 {
		t.Errorf("Expected maxConnLifetime to be 9s, got %v", cfg.maxConnLifetime)
	}
	if cfg.maxConnIdletime != time.Second*3 {
		t.Errorf("Expected maxConnIdletime to be 3s, got %v", cfg.maxConnIdletime)
	}
}

func TestSetters(t *testing.T) {
	cfg := GetConfig()

	// Test boolean setters
	SetMultiTenantEnabled(true)(cfg)
	if !cfg.multiTenantEnabled {
		t.Error("SetMultiTenantEnabled failed")
	}

	SetDatadogEnabled(true)(cfg)
	if !cfg.datadogEnabled {
		t.Error("SetDatadogEnabled failed")
	}

	SetRlsEnabled(true)(cfg)
	if !cfg.rlsEnabled {
		t.Error("SetRlsEnabled failed")
	}

	SetQueryTracerEnabled(true)(cfg)
	if !cfg.queryTracerEnabled {
		t.Error("SetQueryTracerEnabled failed")
	}

	// Test numeric setters
	maxConns := int32(100)
	SetMaxConns(&maxConns)(cfg)
	if cfg.maxConns != maxConns {
		t.Errorf("SetMaxConns failed, expected %d got %d", maxConns, cfg.maxConns)
	}

	minConns := int32(10)
	SetMinConns(&minConns)(cfg)
	if cfg.minConns != minConns {
		t.Errorf("SetMinConns failed, expected %d got %d", minConns, cfg.minConns)
	}

	// Test duration setters
	lifetime := time.Second * 30
	SetMaxConnLifetime(&lifetime)(cfg)
	if cfg.maxConnLifetime != lifetime {
		t.Errorf("SetMaxConnLifetime failed, expected %v got %v", lifetime, cfg.maxConnLifetime)
	}

	idletime := time.Second * 15
	SetMaxConnIdleTime(&idletime)(cfg)
	if cfg.maxConnIdletime != idletime {
		t.Errorf("SetMaxConnIdleTime failed, expected %v got %v", idletime, cfg.maxConnIdletime)
	}

	// Test connection string setter
	connString := "postgresql://localhost:5432/testdb"
	SetConnectionStrings(cfg, connString)(cfg)
	if cfg.connString != connString {
		t.Errorf("SetConnectionStrings failed, expected %s got %s", connString, cfg.connString)
	}
}
