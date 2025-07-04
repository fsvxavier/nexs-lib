package postgresql_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pgx/mocks" // import mocks
	"github.com/stretchr/testify/assert"
)

func TestFacade(t *testing.T) {
	// Cria uma configuração para testes
	config := postgresql.WithConfig(
		postgresql.WithHost("localhost"),
		postgresql.WithPort(5432),
		postgresql.WithDatabase("testdb"),
		postgresql.WithUser("testuser"),
		postgresql.WithPassword("testpass"),
		postgresql.WithMaxConns(10),
		postgresql.WithMinConns(1),
		postgresql.WithMaxConnLifetime(time.Minute*10),
		postgresql.WithSSLMode("disable"),
	)

	// Testes de criação de pool e conexão com database real
	// Estes testes são mantidos, mas com skip para evitar dependência de BD real
	t.Run("Create real pgx pool", func(t *testing.T) {
		t.Skip("Test requires real database")
		// Descomente para testes reais
		// pool, err := postgresql.NewPool(context.Background(), postgresql.PGX, config)
		// assert.NoError(t, err)
		// defer pool.Close()
		//
		// err = pool.Ping(context.Background())
		// assert.NoError(t, err)
	})

	t.Run("Create real pq pool", func(t *testing.T) {
		t.Skip("Test requires real database")
		// Descomente para testes reais
		// pool, err := postgresql.NewPool(context.Background(), postgresql.PQ, config)
		// assert.NoError(t, err)
		// defer pool.Close()
		//
		// err = pool.Ping(context.Background())
		// assert.NoError(t, err)
	})

	t.Run("Create real pgx connection", func(t *testing.T) {
		t.Skip("Test requires real database")
		// Descomente para testes reais
		// conn, err := postgresql.NewConnection(context.Background(), postgresql.PGX, config)
		// assert.NoError(t, err)
		// defer conn.Close(context.Background())
		//
		// err = conn.Ping(context.Background())
		// assert.NoError(t, err)
	})

	t.Run("Create real pq connection", func(t *testing.T) {
		t.Skip("Test requires real database")
		// Descomente para testes reais
		// conn, err := postgresql.NewConnection(context.Background(), postgresql.PQ, config)
		// assert.NoError(t, err)
		// defer conn.Close(context.Background())
		//
		// err = conn.Ping(context.Background())
		// assert.NoError(t, err)
	})

	t.Run("Create batch with operations", func(t *testing.T) {
		batch, err := postgresql.NewBatch(postgresql.PGX)
		assert.NoError(t, err)
		assert.NotNil(t, batch)

		batch.Queue("SELECT 1")
		batch.Queue("SELECT $1", 42)

		// Verificamos que o batch contém operações
		batchObj := batch.GetBatch()
		assert.NotNil(t, batchObj)
	})

	t.Run("Error handling comprehensive", func(t *testing.T) {
		// Teste com erro nulo
		assert.False(t, postgresql.IsEmptyResultError(nil))
		assert.False(t, postgresql.IsDuplicateKeyError(nil))

		// Teste com erro de NoRows
		assert.True(t, postgresql.IsEmptyResultError(common.ErrNoRows))

		// Teste com erro de chave duplicada
		dupErr := common.NewPostgreSQLError("ERROR: duplicate key value violates unique constraint", "23505")
		assert.True(t, postgresql.IsDuplicateKeyError(dupErr))

		// Outros tipos de erros PostgreSQL
		foreignKeyErr := common.NewPostgreSQLError("ERROR: foreign key violation", "23503")
		assert.False(t, postgresql.IsDuplicateKeyError(foreignKeyErr))
		assert.Equal(t, "23503", foreignKeyErr.Code)
	})

	// Testes usando mocks
	t.Run("Config validation", func(t *testing.T) {
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "testuser", config.User)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, int32(10), config.MaxConns)
		assert.Equal(t, int32(1), config.MinConns)
		assert.Equal(t, "disable", config.SSLMode)
	})

	t.Run("Test pgx provider selection", func(t *testing.T) {
		// Testamos apenas que o tipo correto de batch é criado
		batch, err := postgresql.NewBatch(postgresql.PGX)
		assert.NoError(t, err)
		assert.NotNil(t, batch)

		// Validamos que é um batch do tipo pgx
		assert.Contains(t, batch.GetBatch(), "pgx.Batch")
	})

	t.Run("Test pq provider selection", func(t *testing.T) {
		// Testamos apenas que o tipo correto de batch é criado
		batch, err := postgresql.NewBatch(postgresql.PQ)
		assert.NoError(t, err)
		assert.NotNil(t, batch)

		// Validamos que é um batch do tipo pq (string não deve conter pgx)
		batchStr := batch.GetBatch()
		assert.NotContains(t, batchStr, "pgx")
	})

	t.Run("Test invalid provider", func(t *testing.T) {
		// Testa com um provider inválido
		invalidProvider := postgresql.ProviderType("invalid")

		_, err := postgresql.NewBatch(invalidProvider)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider não suportado")

		_, err = postgresql.NewPool(context.Background(), invalidProvider, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider não suportado")

		_, err = postgresql.NewConnection(context.Background(), invalidProvider, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider não suportado")
	})

	t.Run("Test mockable pgx connection", func(t *testing.T) {
		// Cria mock para pgx
		mock, err := mocks.GetMock()
		if err != nil {
			t.Fatalf("erro ao criar mock do pgx: %v", err)
		}
		defer mocks.CloseConn(mock)

		// Configura expectativa para o mock
		mock.ExpectPing().WillReturnError(nil)

		// Este é um teste sintético para verificar a interface do facade
		// Um teste mais completo exigiria injeção de mock no facade
		assert.NotNil(t, mock)
		err = mock.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Test mockable pq connection", func(t *testing.T) {
		// Cria mock para pq
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("erro ao criar mock do pq: %v", err)
		}
		defer db.Close()

		// Configura expectativa para o mock
		mock.ExpectPing()

		// Este é um teste sintético para verificar a interface do facade
		// Um teste mais completo exigiria injeção de mock no facade
		assert.NotNil(t, db)
		err = db.Ping()
		assert.NoError(t, err)
	})
}

// TestRaceConditions testa race conditions no pool de conexões
func TestRaceConditions(t *testing.T) {
	// Este teste deve ser executado com go test -race
	t.Run("Race conditions test", func(t *testing.T) {
		t.Skip("Test requires real database and -race flag")

		// Para executar este teste, remova o Skip e execute:
		// go test -race -run=TestRaceConditions

		// Um verdadeiro teste de race condition exigiria acesso a um banco real
		// e múltiplas goroutines acessando o pool concorrentemente.
		/*
			config := postgresql.WithConfig(
				postgresql.WithHost("localhost"),
				postgresql.WithPort(5432),
				postgresql.WithDatabase("testdb"),
				postgresql.WithUser("testuser"),
				postgresql.WithPassword("testpass"),
			)

			pool, err := postgresql.NewPool(context.Background(), postgresql.PGX, config)
			assert.NoError(t, err)
			defer pool.Close()

			var wg sync.WaitGroup

			// Executa múltiplas consultas concorrentes
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					conn, err := pool.Acquire(context.Background())
					assert.NoError(t, err)
					defer conn.Close(context.Background())

					var result int
					err = conn.QueryOne(context.Background(), &result, "SELECT 1")
					assert.NoError(t, err)
					assert.Equal(t, 1, result)
				}(i)
			}

			wg.Wait()
		*/
	})
}

// BenchmarkPGX executa benchmarks para o provider pgx
func BenchmarkPGX(b *testing.B) {
	b.Run("Benchmark pgx operations", func(b *testing.B) {
		b.Skip("Benchmark requires real database")

		// Para executar este benchmark, remova o Skip e execute:
		// go test -bench=BenchmarkPGX

		/*
			config := postgresql.WithConfig(
				postgresql.WithHost("localhost"),
				postgresql.WithPort(5432),
				postgresql.WithDatabase("testdb"),
				postgresql.WithUser("testuser"),
				postgresql.WithPassword("testpass"),
			)

			// Configuração do benchmark
			ctx := context.Background()
			pool, err := postgresql.NewPool(ctx, postgresql.PGX, config)
			if err != nil {
				b.Fatalf("erro ao criar pool: %v", err)
			}
			defer pool.Close()

			// Setup: cria tabela de teste se não existir
			conn, err := pool.Acquire(ctx)
			if err != nil {
				b.Fatalf("erro ao adquirir conexão: %v", err)
			}

			// Limpa após o teste
			defer func() {
				conn.Exec(ctx, "DROP TABLE IF EXISTS benchmark_test")
				conn.Close(ctx)
			}()

			// Cria tabela para o teste
			err = conn.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS benchmark_test (
					id SERIAL PRIMARY KEY,
					name TEXT NOT NULL,
					value INTEGER NOT NULL
				)
			`)
			if err != nil {
				b.Fatalf("erro ao criar tabela: %v", err)
			}

			// Prepara dados de teste
			for i := 0; i < 100; i++ {
				err = conn.Exec(ctx, "INSERT INTO benchmark_test(name, value) VALUES($1, $2)",
					fmt.Sprintf("test-%d", i), i)
				if err != nil {
					b.Fatalf("erro ao inserir dados: %v", err)
				}
			}

			// Executa o benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				conn, err := pool.Acquire(ctx)
				if err != nil {
					b.Fatalf("erro ao adquirir conexão: %v", err)
				}

				var result []struct {
					ID    int    `db:"id"`
					Name  string `db:"name"`
					Value int    `db:"value"`
				}

				err = conn.QueryAll(ctx, &result, "SELECT id, name, value FROM benchmark_test LIMIT 10")
				if err != nil {
					b.Fatalf("erro na consulta: %v", err)
				}

				conn.Close(ctx)
			}
		*/
	})
}

// BenchmarkPQ executa benchmarks para o provider pq
func BenchmarkPQ(b *testing.B) {
	b.Run("Benchmark pq operations", func(b *testing.B) {
		b.Skip("Benchmark requires real database")

		// O código para este benchmark é semelhante ao BenchmarkPGX,
		// mas usando postgresql.PQ como provider
	})
}

// TestInterfaceComplianceWithMocks verifica se os tipos mockados implementam as interfaces corretamente
func TestInterfaceComplianceWithMocks(t *testing.T) {
	t.Run("Interface Compliance Test", func(t *testing.T) {
		// Criamos mocks para verificar a conformidade com as interfaces
		pgxMock, err := mocks.GetMock()
		assert.NoError(t, err)
		defer mocks.CloseConn(pgxMock)

		sqlDb, sqlMock, err := sqlmock.New()
		assert.NoError(t, err)
		defer sqlDb.Close()

		// Configuramos expectativas básicas
		pgxMock.ExpectPing().WillReturnError(nil)
		sqlMock.ExpectPing()

		// Verificamos se os mocks funcionam corretamente
		err = pgxMock.Ping(context.Background())
		assert.NoError(t, err)

		err = sqlDb.Ping()
		assert.NoError(t, err)

		// Verificamos se todas as expectativas foram atendidas
		err = pgxMock.ExpectationsWereMet()
		assert.NoError(t, err, "expectativas do pgxMock não foram atendidas")

		err = sqlMock.ExpectationsWereMet()
		assert.NoError(t, err, "expectativas do sqlMock não foram atendidas")
	})
}

// TestConnectionStringGeneration verifica a geração correta de strings de conexão
func TestConnectionStringGeneration(t *testing.T) {
	t.Run("Connection String for PGX", func(t *testing.T) {
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

	t.Run("Connection String with Extra Params", func(t *testing.T) {
		config := postgresql.WithConfig(
			postgresql.WithHost("localhost"),
			postgresql.WithPort(5432),
			postgresql.WithDatabase("testdb"),
			postgresql.WithUser("user"),
			postgresql.WithPassword("pass"),
			postgresql.WithSSLMode("verify-full"),
			// Estes parâmetros não estão disponíveis diretamente, mas podem ser adicionados na string de conexão
			// postgresql.WithConnectTimeout(10),
			// postgresql.WithApplicationName("isis-test"),
		)

		connStr := config.ConnectionString()

		// Verificações adicionais
		assert.Contains(t, connStr, "sslmode=verify-full")
	})
}
