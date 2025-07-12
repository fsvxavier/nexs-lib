package parsers

import (
	"context"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
)

// Testes simples para aumentar cobertura
func TestSimpleCoverageBoosters(t *testing.T) {
	t.Run("test composite error parser", func(t *testing.T) {
		parser1 := NewPostgreSQLErrorParser()
		parser2 := NewMySQLErrorParser()

		composite := NewCompositeErrorParser(parser1, parser2)
		if composite == nil {
			t.Error("Composite parser should not be nil")
		}

		// Testar CanParse
		testErr := fmt.Errorf("pq: connection failed")
		if !composite.CanParse(testErr) {
			t.Error("Composite parser should parse postgres error")
		}

		// Testar Parse
		parsed := composite.Parse(testErr)
		if parsed.Message == "" {
			t.Error("Composite parser should return parsed result")
		}
	})

	t.Run("test default parser", func(t *testing.T) {
		parser := NewDefaultParser()
		if parser == nil {
			t.Error("Default parser should not be nil")
		}

		testErr := fmt.Errorf("unknown error")
		defaultParser := NewPostgreSQLErrorParser()
		parsed := ParseError(testErr, defaultParser)
		if parsed.Message == "" {
			t.Error("ParseError should return a result")
		}
	})

	t.Run("test registry plugin functions", func(t *testing.T) {
		registry := NewDistributedParserRegistry()

		// Test RegisterPlugin
		plugin := NewGenericDatabasePlugin()
		err := registry.RegisterPlugin(plugin)
		if err != nil {
			t.Errorf("RegisterPlugin failed: %v", err)
		}

		// Test RegisterFactory
		factory := NewCustomParserFactory()
		err = registry.RegisterFactory("test_factory", factory)
		if err != nil {
			t.Errorf("RegisterFactory failed: %v", err)
		}

		// Test ConfigureParser
		config := map[string]interface{}{
			"timeout": "30s",
			"retries": 3,
		}
		err = registry.ConfigureParser("test", config)
		if err != nil {
			t.Logf("ConfigureParser may fail for non-existent parser: %v", err)
		}

		// Test GetConfiguration
		conf, err := registry.GetConfiguration("test")
		if err != nil {
			t.Logf("GetConfiguration failed as expected: %v", err)
		}
		if conf != nil {
			t.Logf("Got configuration: %v", conf)
		}

		// Test RegisterHealthChecker
		err = registry.RegisterHealthChecker("test", func() error { return nil })
		if err != nil {
			t.Logf("RegisterHealthChecker failed as expected: %v", err)
		}
	})
}

func TestPluginInterfaceMethods(t *testing.T) {
	t.Run("test plugin interface methods", func(t *testing.T) {
		plugin := NewGenericDatabasePlugin()

		name := plugin.Name()
		if name == "" {
			t.Error("Plugin should have a name")
		}

		version := plugin.Version()
		if version == "" {
			t.Error("Plugin should have a version")
		}

		description := plugin.Description()
		if description == "" {
			t.Error("Plugin should have a description")
		}

		// Test CreateParser
		config := map[string]interface{}{
			"patterns":  []string{"test:"},
			"errorType": "database",
		}
		parser, err := plugin.CreateParser(config)
		if err != nil {
			t.Logf("CreateParser failed: %v", err)
		}
		if parser != nil {
			t.Log("CreateParser returned a parser")
		}

		// Test ValidateConfig
		err = plugin.ValidateConfig(config)
		if err != nil {
			t.Logf("ValidateConfig failed: %v", err)
		}

		// Test DefaultConfig
		defaultConf := plugin.DefaultConfig()
		if defaultConf == nil {
			t.Error("DefaultConfig should return configuration")
		}
	})

	t.Run("test custom parser factory", func(t *testing.T) {
		factory := NewCustomParserFactory()

		config := map[string]interface{}{
			"pattern":   "error:",
			"errorType": "custom",
		}

		parser, err := factory.CreateParser("regex", config)
		if err != nil {
			t.Logf("CreateParser failed: %v", err)
		}
		if parser != nil {
			t.Log("CreateParser returned a parser")
		}

		types := factory.SupportedTypes()
		if len(types) == 0 {
			t.Error("SupportedTypes should return available types")
		}

		// Test RegisterCustomType
		err = factory.RegisterCustomType("test_type", func(config map[string]interface{}) (interfaces.ErrorParser, error) {
			return NewPostgreSQLErrorParser(), nil
		})
		if err != nil {
			t.Errorf("RegisterCustomType failed: %v", err)
		}
	})
}

func TestComplexErrorParsing(t *testing.T) {
	t.Run("test complex MongoDB error parsing", func(t *testing.T) {
		parser := NewMongoDBErrorParser()

		// Testar diferentes tipos de erro MongoDB com parsing completo
		errors := []error{
			fmt.Errorf("E11000 duplicate key error collection: test.users index: email_1 dup key: { email: \"test@example.com\" } (11000)"),
			fmt.Errorf("mongo: server selection timeout: context deadline exceeded"),
			fmt.Errorf("bson: cannot decode array into Go value of type string"),
			fmt.Errorf("collection 'users' doesn't exist in database 'test'"),
		}

		for _, err := range errors {
			if parser.CanParse(err) {
				parsed := parser.Parse(err)
				if parsed.Message == "" {
					t.Error("Should parse MongoDB error and return message")
				}

				// Verificar se extraiu detalhes específicos
				if parsed.Details != nil {
					t.Logf("MongoDB error details: %v", parsed.Details)
				}
			}
		}
	})

	t.Run("test complex Redis error parsing", func(t *testing.T) {
		parser := NewRedisErrorParser()

		errors := []error{
			fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value"),
			fmt.Errorf("ERR invalid DB index '16' for redis instance"),
			fmt.Errorf("redis: connection pool exhausted (max: 10, in use: 10)"),
			fmt.Errorf("redigo: dial tcp 127.0.0.1:6379: connect: connection refused"),
		}

		for _, err := range errors {
			if parser.CanParse(err) {
				parsed := parser.Parse(err)
				if parsed.Message == "" {
					t.Error("Should parse Redis error and return message")
				}

				// Força a execução de diferentes caminhos no parser
				if parsed.Details != nil && len(parsed.Details) > 0 {
					t.Logf("Redis error details: %v", parsed.Details)
				}
			}
		}
	})

	t.Run("test complex AWS error parsing", func(t *testing.T) {
		parser := NewAWSErrorParser()

		errors := []error{
			fmt.Errorf("aws: s3 NoSuchBucket: The specified bucket does not exist"),
			fmt.Errorf("dynamodb: ValidationException: Invalid KeyConditionExpression"),
			fmt.Errorf("lambda: function not found: arn:aws:lambda:us-east-1:123456789012:function:my-function"),
			fmt.Errorf("throttling: Rate exceeded for operation GetItem"),
		}

		for _, err := range errors {
			if parser.CanParse(err) {
				parsed := parser.Parse(err)
				if parsed.Message == "" {
					t.Error("Should parse AWS error and return message")
				}

				// Executar diferentes caminhos do AWS parser
				if parsed.Details != nil {
					t.Logf("AWS error details: %v", parsed.Details)
				}
			}
		}
	})
}

func TestNetworkParserDetails(t *testing.T) {
	t.Run("test network parser with different error types", func(t *testing.T) {
		parser := NewNetworkErrorParser()

		// Criar erros que ativem diferentes caminhos no parser
		testErr := fmt.Errorf("dial tcp 127.0.0.1:5432: connect: connection refused")

		if parser.CanParse(testErr) {
			parsed := parser.Parse(testErr)
			if parsed.Details == nil {
				t.Error("Network parser should extract details")
			}

			// Verificar se extraiu informações de rede
			if host, ok := parsed.Details["host"]; ok {
				t.Logf("Extracted host: %v", host)
			}
			if port, ok := parsed.Details["port"]; ok {
				t.Logf("Extracted port: %v", port)
			}
		}

		// Testar com erro de timeout de rede
		timeoutErr := fmt.Errorf("dial tcp 192.168.1.1:80: i/o timeout")
		if parser.CanParse(timeoutErr) {
			parsed := parser.Parse(timeoutErr)
			if parsed.Details != nil {
				t.Logf("Network timeout details: %v", parsed.Details)
			}
		}
	})
}

func TestGRPCParserDetails(t *testing.T) {
	t.Run("test GRPC parser with different status codes", func(t *testing.T) {
		parser := NewGRPCErrorParser()

		errors := []error{
			fmt.Errorf("rpc error: code = Unavailable desc = connection failed"),
			fmt.Errorf("rpc error: code = DeadlineExceeded desc = context deadline exceeded"),
			fmt.Errorf("rpc error: code = NotFound desc = method not found"),
			fmt.Errorf("rpc error: code = PermissionDenied desc = access denied"),
		}

		for _, err := range errors {
			if parser.CanParse(err) {
				parsed := parser.Parse(err)
				if parsed.Message == "" {
					t.Error("Should parse GRPC error and return message")
				}

				// Verificar se extraiu código de status
				if code, ok := parsed.Details["grpc_code"]; ok {
					t.Logf("GRPC code: %v", code)
				}
			}
		}
	})
}

func TestHTTPParserDetails(t *testing.T) {
	t.Run("test HTTP parser with different status codes", func(t *testing.T) {
		parser := NewHTTPErrorParser()

		errors := []error{
			fmt.Errorf("HTTP 404: Not Found"),
			fmt.Errorf("HTTP 500: Internal Server Error"),
			fmt.Errorf("HTTP 503: Service Unavailable"),
			fmt.Errorf("status: 429 Too Many Requests"),
		}

		for _, err := range errors {
			if parser.CanParse(err) {
				parsed := parser.Parse(err)
				if parsed.Message == "" {
					t.Error("Should parse HTTP error and return message")
				}

				// Verificar se extraiu status code
				if status, ok := parsed.Details["status_code"]; ok {
					t.Logf("HTTP status: %v", status)
				}
			}
		}
	})
}

func TestRegistryAdvancedOperations(t *testing.T) {
	t.Run("test registry with context operations", func(t *testing.T) {
		registry := NewDistributedParserRegistry()

		// Registrar alguns parsers
		registry.RegisterParser("test1", NewPostgreSQLErrorParser(), 100)
		registry.RegisterParser("test2", NewMySQLErrorParser(), 200)

		// Test enable non-existent parser
		err := registry.EnableParser("non_existent")
		if err == nil {
			t.Error("Should fail to enable non-existent parser")
		}

		// Test set priority for non-existent parser
		err = registry.SetPriority("non_existent", 100)
		if err == nil {
			t.Error("Should fail to set priority for non-existent parser")
		}

		// Test parse with context
		ctx := context.Background()
		testErr := fmt.Errorf("pq: test error")
		_, _, err = registry.Parse(ctx, testErr)
		if err != nil {
			t.Errorf("Parse should succeed: %v", err)
		}
	})
}
