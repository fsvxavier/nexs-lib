package pq_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	pqprovider "github.com/fsvxavier/nexs-lib/db/postgresql/pq"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pq/mocks"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// Testa a configuração e string de conexão
	config := common.DefaultConfig()
	config.Host = "testhost"
	config.Port = 5555
	config.Database = "testdb"
	config.User = "user"
	config.Password = "pass"
	config.SSLMode = "disable"

	// Verifica se os valores foram configurados corretamente
	assert.Equal(t, "testhost", config.Host)
	assert.Equal(t, 5555, config.Port)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "user", config.User)
	assert.Equal(t, "pass", config.Password)
	assert.Equal(t, "disable", config.SSLMode)

	// Verifica a string de conexão
	connStr := config.ConnectionString()
	assert.Contains(t, connStr, "host=testhost")
	assert.Contains(t, connStr, "port=5555")
	assert.Contains(t, connStr, "dbname=testdb")
	assert.Contains(t, connStr, "user=user")
	assert.Contains(t, connStr, "password=pass")
	assert.Contains(t, connStr, "sslmode=disable")
}

func TestErrorHandling(t *testing.T) {
	// Testes para manipulação de erros
	assert.False(t, common.IsEmptyResultError(nil))
	assert.True(t, common.IsEmptyResultError(common.ErrNoRows))

	// Teste para erro de chave duplicada
	err := common.NewPostgreSQLError("ERROR: duplicate key value violates unique constraint", "23505")
	assert.True(t, common.IsDuplicateKeyError(err))
	assert.True(t, err.IsUniqueViolation())
}

func TestQueries(t *testing.T) {
	// Cria um mock de banco de dados
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar mock do sql: %v", err)
	}
	defer db.Close()

	// Teste de QueryOne
	t.Run("QueryOne", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "John", 30)

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(1).
			WillReturnRows(rows)

		// Estrutura para receber os resultados
		type User struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
			Age  int    `db:"age"`
		}

		// Cria um pool usando o mock
		pool := pqprovider.NewPoolForTest(db)

		// Adquire uma conexão do pool
		conn, err := pool.Acquire(context.Background())
		assert.NoError(t, err)
		defer conn.Close(context.Background())

		// Executa a consulta
		var user User
		err = conn.QueryOne(context.Background(), &user, "SELECT id, name, age FROM users WHERE id = ?", 1)

		// Verifica os resultados
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "John", user.Name)
		assert.Equal(t, 30, user.Age)

		// Verifica se todas as expectativas foram atendidas
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas não foram atendidas: %s", err)
		}
	})

	// Teste de QueryAll
	t.Run("QueryAll", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "John", 30).
			AddRow(2, "Jane", 25).
			AddRow(3, "Bob", 40)

		mock.ExpectQuery("SELECT (.+) FROM users").
			WillReturnRows(rows)

		// Estrutura para receber os resultados
		type User struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
			Age  int    `db:"age"`
		}

		// Cria um pool usando o mock
		pool := pqprovider.NewPoolForTest(db)

		// Adquire uma conexão do pool
		conn, err := pool.Acquire(context.Background())
		assert.NoError(t, err)
		defer conn.Close(context.Background())

		// Executa a consulta
		var users []User
		err = conn.QueryAll(context.Background(), &users, "SELECT id, name, age FROM users")

		// Verifica os resultados
		assert.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, "John", users[0].Name)
		assert.Equal(t, "Jane", users[1].Name)
		assert.Equal(t, "Bob", users[2].Name)

		// Verifica se todas as expectativas foram atendidas
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas não foram atendidas: %s", err)
		}
	})

	// Teste de QueryCount
	t.Run("QueryCount", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).
			AddRow(3)

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
			WillReturnRows(rows)

		// Cria um pool usando o mock
		pool := pqprovider.NewPoolForTest(db)

		// Adquire uma conexão do pool
		conn, err := pool.Acquire(context.Background())
		assert.NoError(t, err)
		defer conn.Close(context.Background())

		// Executa a consulta
		count, err := conn.QueryCount(context.Background(), "SELECT COUNT(*) FROM users")

		// Verifica os resultados
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, 3, *count)

		// Verifica se todas as expectativas foram atendidas
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas não foram atendidas: %s", err)
		}
	})
}

func TestTransaction(t *testing.T) {
	// Cria um mock de banco de dados
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("erro ao criar mock do sql: %v", err)
	}
	defer db.Close()

	// Teste de transação com commit
	t.Run("Transaction Commit", func(t *testing.T) {
		// Expectativas para a transação
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO users").
			WithArgs("Alice", 35).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE users").
			WithArgs("Alice Updated", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Cria um pool usando o mock
		pool := pqprovider.NewPoolForTest(db)

		// Adquire uma conexão do pool
		conn, err := pool.Acquire(context.Background())
		assert.NoError(t, err)
		defer conn.Close(context.Background())

		// Inicia uma transação
		tx, err := conn.BeginTransaction(context.Background())
		assert.NoError(t, err)

		// Executa comandos
		err = tx.Exec(context.Background(), "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 35)
		assert.NoError(t, err)

		err = tx.Exec(context.Background(), "UPDATE users SET name = ? WHERE id = ?", "Alice Updated", 1)
		assert.NoError(t, err)

		// Confirma a transação
		err = tx.Commit(context.Background())
		assert.NoError(t, err)

		// Verifica se todas as expectativas foram atendidas
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas não foram atendidas: %s", err)
		}
	})

	// Teste de transação com rollback
	t.Run("Transaction Rollback", func(t *testing.T) {
		// Expectativas para a transação
		mock.ExpectBegin()
		mock.ExpectRollback()

		// Cria um pool usando o mock
		pool := pqprovider.NewPoolForTest(db)

		// Adquire uma conexão do pool
		conn, err := pool.Acquire(context.Background())
		assert.NoError(t, err)
		defer conn.Close(context.Background())

		// Inicia uma transação
		tx, err := conn.BeginTransaction(context.Background())
		assert.NoError(t, err)

		// Cancela a transação
		err = tx.Rollback(context.Background())
		assert.NoError(t, err)

		// Verifica se todas as expectativas foram atendidas
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas não foram atendidas: %s", err)
		}
	})
}

// TestPool testa a implementação do pool de conexões
func TestPool(t *testing.T) {
	t.Run("Pool Configuration", func(t *testing.T) {
		// Configuração para o teste
		config := common.DefaultConfig()
		config.Host = "localhost"
		config.Port = 5432
		config.Database = "testdb"
		config.User = "testuser"
		config.Password = "testpass"
		config.MaxConns = 10
		config.MinConns = 2

		// Verificamos apenas que a configuração é aplicada corretamente
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "testuser", config.User)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, int32(10), config.MaxConns)
		assert.Equal(t, int32(2), config.MinConns)

		// Um teste completo exigiria um banco de dados real
	})
}

// TestCustomMocks testa o uso dos mocks personalizados
func TestCustomMocks(t *testing.T) {
	t.Run("MockConn", func(t *testing.T) {
		// Cria um mock personalizado
		mockConn := &mocks.MockConn{
			PingFunc: func(ctx context.Context) error {
				return nil
			},
			QueryContextFunc: func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
				// Simularia retorno de linhas
				return nil, nil
			},
		}

		// Testa o método Ping
		err := mockConn.Ping(context.Background())
		assert.NoError(t, err)
	})
}

// TestRaceConditions executa testes de race condition no pool
func TestRaceConditions(t *testing.T) {
	// Este teste deve ser executado com a flag -race
	t.Skip("Race condition test requires real database")
}

// BenchmarkQueryOne executa benchmark para QueryOne
func BenchmarkQueryOne(b *testing.B) {
	// Implementação do benchmark
	b.Skip("Benchmark requires real database")
}

// BenchmarkQueryAll executa benchmark para QueryAll
func BenchmarkQueryAll(b *testing.B) {
	// Implementação do benchmark
	b.Skip("Benchmark requires real database")
}

// BenchmarkPoolAcquireRelease executa benchmark para acquire/release no pool
func BenchmarkPoolAcquireRelease(b *testing.B) {
	// Implementação do benchmark
	b.Skip("Benchmark requires real database")
}
