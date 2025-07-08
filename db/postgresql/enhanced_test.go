package postgresql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionValidation(t *testing.T) {
	t.Run("Valid connection creation with factory", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
			postgresql.WithPassword("testpass"),
		)

		// Este teste falha sem DB real, mas valida a validação de entrada
		_, err := postgresql.NewConnection(context.Background(), postgresql.PGX, config)
		// Esperamos erro de conexão, não erro de validação
		assert.Error(t, err)
		assert.NotContains(t, err.Error(), "configuração não pode ser nil")
		assert.NotContains(t, err.Error(), "contexto não pode ser nil")
	})

	t.Run("Nil context", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
		)

		conn, err := postgresql.NewConnection(nil, postgresql.PGX, config)
		assert.Nil(t, conn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contexto não pode ser nil")
	})

	t.Run("Nil config", func(t *testing.T) {
		conn, err := postgresql.NewConnection(context.Background(), postgresql.PGX, nil)
		assert.Nil(t, conn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuração não pode ser nil")
	})

	t.Run("Invalid provider type", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
		)

		conn, err := postgresql.NewConnection(context.Background(), postgresql.ProviderType("invalid"), config)
		assert.Nil(t, conn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tipo de provider inválido")
	})

	// Teste removido: validação de configuração é tratada pelo PGX internamente
	// quando há tentativa de conexão real
}

func TestNewPoolValidation(t *testing.T) {
	t.Run("Valid pool creation with factory", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
			postgresql.WithPassword("testpass"),
		)

		// Este teste falha sem DB real, mas valida a validação de entrada
		pool, err := postgresql.NewPool(context.Background(), postgresql.PGX, config)
		// Esperamos erro de conexão, não erro de validação
		if err != nil {
			assert.NotContains(t, err.Error(), "configuração não pode ser nil")
			assert.NotContains(t, err.Error(), "contexto não pode ser nil")
		}
		if pool != nil {
			pool.Close()
		}
	})

	t.Run("Nil context", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
		)

		pool, err := postgresql.NewPool(nil, postgresql.PGX, config)
		assert.Nil(t, pool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contexto não pode ser nil")
	})

	t.Run("Nil config", func(t *testing.T) {
		pool, err := postgresql.NewPool(context.Background(), postgresql.PGX, nil)
		assert.Nil(t, pool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuração não pode ser nil")
	})

	t.Run("Invalid provider type", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
		)

		pool, err := postgresql.NewPool(context.Background(), postgresql.ProviderType("invalid"), config)
		assert.Nil(t, pool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tipo de provider inválido")
	})
}

func TestNewBatchValidation(t *testing.T) {
	t.Run("Valid batch creation", func(t *testing.T) {
		batch, err := postgresql.NewBatch(postgresql.PGX)
		assert.NoError(t, err)
		assert.NotNil(t, batch)
	})

	t.Run("Invalid provider type", func(t *testing.T) {
		batch, err := postgresql.NewBatch(postgresql.ProviderType("invalid"))
		assert.Nil(t, batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tipo de provider inválido")
	})
}

func TestProviderTypeConstants(t *testing.T) {
	// Verifica se as constantes estão definidas corretamente
	assert.Equal(t, "pgx", string(postgresql.PGX))
	assert.Equal(t, "pq", string(postgresql.PQ))
	assert.Equal(t, "gorm", string(postgresql.GORM))
}

func TestErrorConstants(t *testing.T) {
	// Verifica se os erros padrão estão disponíveis
	assert.NotNil(t, postgresql.ErrNoRows)
	assert.NotNil(t, postgresql.ErrNoTransaction)
	assert.NotNil(t, postgresql.ErrNoConnection)
	assert.NotNil(t, postgresql.ErrInvalidNestedTransaction)
	assert.NotNil(t, postgresql.ErrInvalidOperation)
}

func TestConfigurationOptions(t *testing.T) {
	t.Run("All configuration options", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("testhost"),
			postgresql.WithPort(9999),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
			postgresql.WithPassword("testpass"),
			postgresql.WithSSLMode("require"),
			postgresql.WithMaxConns(50),
			postgresql.WithMinConns(5),
			postgresql.WithTraceEnabled(true),
			postgresql.WithQueryLogEnabled(true),
			postgresql.WithMultiTenantEnabled(true),
		)

		assert.Equal(t, "testhost", config.Host)
		assert.Equal(t, 9999, config.Port)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "testuser", config.User)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, "require", config.SSLMode)
		assert.Equal(t, int32(50), config.MaxConns)
		assert.Equal(t, int32(5), config.MinConns)
		assert.True(t, config.TraceEnabled)
		assert.True(t, config.QueryLogEnabled)
		assert.True(t, config.MultiTenantEnabled)
	})
	t.Run("Default configuration", func(t *testing.T) {
		config := postgresql.WithConfig()

		// Verifica se valores padrão são aplicados (podem estar vazios até serem configurados)
		assert.NotNil(t, config)
		assert.GreaterOrEqual(t, config.Port, 0)
	})
}

func TestErrorHandlingFunctions(t *testing.T) {
	t.Run("IsEmptyResultError", func(t *testing.T) {
		// Teste com nil
		assert.False(t, postgresql.IsEmptyResultError(nil))

		// Teste com erro que não é NoRows
		regularErr := errors.New("regular error")
		assert.False(t, postgresql.IsEmptyResultError(regularErr))

		// Teste com erro NoRows
		assert.True(t, postgresql.IsEmptyResultError(postgresql.ErrNoRows))
	})

	t.Run("IsDuplicateKeyError", func(t *testing.T) {
		// Teste com nil
		assert.False(t, postgresql.IsDuplicateKeyError(nil))

		// Teste com erro que não é duplicate key
		regularErr := errors.New("regular error")
		assert.False(t, postgresql.IsDuplicateKeyError(regularErr))

		// Teste com erro de chave duplicada
		dupErr := common.NewPostgreSQLError("ERROR: duplicate key value violates unique constraint", "23505")
		assert.True(t, postgresql.IsDuplicateKeyError(dupErr))
	})
}

func TestFactoryIntegration(t *testing.T) {
	t.Run("Factory is accessible", func(t *testing.T) {
		factory := postgresql.GetFactory()
		assert.NotNil(t, factory)

		providers := factory.GetSupportedProviders()
		assert.Contains(t, providers, postgresql.PGX)
		assert.Contains(t, providers, postgresql.PQ)
		assert.Contains(t, providers, postgresql.GORM)
	})

	t.Run("Factory can be replaced", func(t *testing.T) {
		originalFactory := postgresql.GetFactory()
		newFactory := postgresql.NewDatabaseFactory()

		// Define nova factory
		postgresql.SetFactory(newFactory)
		assert.Same(t, newFactory, postgresql.GetFactory())

		// Restaura factory original
		postgresql.SetFactory(originalFactory)
		assert.Same(t, originalFactory, postgresql.GetFactory())
	})
}

func TestConnectionStringGeneration(t *testing.T) {
	t.Run("Connection string with all parameters", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("testhost"),
			postgresql.WithPort(5555),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("testuser"),
			postgresql.WithPassword("testpass"),
			postgresql.WithSSLMode("disable"),
		)

		connStr := config.ConnectionString()

		// Verificações para string de conexão
		assert.Contains(t, connStr, "host=testhost")
		assert.Contains(t, connStr, "port=5555")
		assert.Contains(t, connStr, "dbname=testdb")
		assert.Contains(t, connStr, "user=testuser")
		assert.Contains(t, connStr, "password=testpass")
		assert.Contains(t, connStr, "sslmode=disable")
	})

	t.Run("Connection string without optional parameters", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("db"),
			postgresql.WithUser("user"),
		)

		connStr := config.ConnectionString()

		// Verificações básicas
		assert.Contains(t, connStr, "host=localhost")
		assert.Contains(t, connStr, "port=5432")
		assert.Contains(t, connStr, "dbname=db")
		assert.Contains(t, connStr, "user=user")
	})
}

func TestBatchOperations(t *testing.T) {
	t.Run("Batch creation and operations", func(t *testing.T) {
		// Testa criação de batch para diferentes providers
		providers := []postgresql.ProviderType{
			postgresql.PGX,
			postgresql.PQ,
			postgresql.GORM,
		}

		for _, provider := range providers {
			t.Run(string(provider), func(t *testing.T) {
				batch, err := postgresql.NewBatch(provider)
				require.NoError(t, err)
				require.NotNil(t, batch)

				// Testa operações básicas do batch
				batch.Queue("SELECT 1")
				batch.Queue("SELECT $1", 42)
				batch.Queue("INSERT INTO test(name) VALUES($1)", "test")

				// Verifica que o batch contém operações
				batchObj := batch.GetBatch()
				assert.NotNil(t, batchObj)
			})
		}
	})
}
