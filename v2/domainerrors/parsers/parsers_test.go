package parsers

import (
	"fmt"
	"testing"
)

// TestBasicParserOperations testa operações básicas dos parsers
func TestBasicParserOperations(t *testing.T) {
	// Testa PostgreSQL parser
	pgParser := NewPostgreSQLErrorParser()
	if pgParser == nil {
		t.Error("PostgreSQL parser should not be nil")
	}

	// Testa se consegue detectar erros PostgreSQL
	pgErr := fmt.Errorf("pq: duplicate key value violates unique constraint")
	if !pgParser.CanParse(pgErr) {
		t.Error("Should be able to parse PostgreSQL error")
	}

	// Testa parse
	parsed := pgParser.Parse(pgErr)
	_ = parsed // avoid unused variable

	// Testa MySQL parser
	mysqlParser := NewMySQLErrorParser()
	if mysqlParser == nil {
		t.Error("MySQL parser should not be nil")
	}

	// Testa Network parser
	networkParser := NewNetworkErrorParser()
	if networkParser == nil {
		t.Error("Network parser should not be nil")
	}

	// Testa Timeout parser
	timeoutParser := NewTimeoutErrorParser()
	if timeoutParser == nil {
		t.Error("Timeout parser should not be nil")
	}

	// Testa SQL parser
	sqlParser := NewSQLErrorParser()
	if sqlParser == nil {
		t.Error("SQL parser should not be nil")
	}
}

// TestDistributedRegistry testa o registry distribuído
func TestDistributedRegistry(t *testing.T) {
	registry := NewDistributedParserRegistry()
	if registry == nil {
		t.Error("Distributed registry should not be nil")
	}

	// Testa global registry
	global := GetGlobalParserRegistry()
	if global == nil {
		t.Error("Global parser registry should not be nil")
	}
}

// TestEnhancedParser testa o parser aprimorado
func TestEnhancedParser(t *testing.T) {
	parser := NewEnhancedPostgreSQLErrorParser()
	if parser == nil {
		t.Error("Enhanced parser should not be nil")
	}
}
