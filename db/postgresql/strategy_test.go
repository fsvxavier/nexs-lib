package postgresql

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/stretchr/testify/assert"
)

func TestPGXStrategy_ValidateConfig(t *testing.T) {
	strategy := NewPGXStrategy()

	t.Run("Valid Config", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			MaxConns: 10,
			MinConns: 2,
		}

		err := strategy.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Missing Host", func(t *testing.T) {
		config := &common.Config{
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host é obrigatório")
	})

	t.Run("Missing Database", func(t *testing.T) {
		config := &common.Config{
			Host: "localhost",
			Port: 5432,
			User: "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database é obrigatório")
	})

	t.Run("Missing User", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user é obrigatório")
	})

	t.Run("Invalid Port", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     0,
			Database: "testdb",
			User:     "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port deve estar entre 1 e 65535")
	})

	t.Run("Port Too High", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     65536,
			Database: "testdb",
			User:     "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port deve estar entre 1 e 65535")
	})

	t.Run("Negative MaxConns", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			MaxConns: -1,
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maxConns deve ser >= 0")
	})

	t.Run("Negative MinConns", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			MinConns: -1,
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minConns deve ser >= 0")
	})

	t.Run("MinConns Greater Than MaxConns", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			MaxConns: 5,
			MinConns: 10,
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minConns (10) não pode ser maior que maxConns (5)")
	})

	t.Run("Zero MaxConns With MinConns", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			MaxConns: 0,
			MinConns: 2,
		}

		err := strategy.ValidateConfig(config)
		assert.NoError(t, err) // MaxConns = 0 significa sem limite, então MinConns > 0 é válido
	})
}

func TestPQStrategy_ValidateConfig(t *testing.T) {
	strategy := NewPQStrategy()

	t.Run("Valid Config", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			MaxConns: 10,
			MinConns: 2,
		}

		err := strategy.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Missing Host", func(t *testing.T) {
		config := &common.Config{
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host é obrigatório para PQ")
	})
}

func TestGORMStrategy_ValidateConfig(t *testing.T) {
	strategy := NewGORMStrategy()

	t.Run("Valid Config", func(t *testing.T) {
		config := &common.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			MaxConns: 10,
			MinConns: 2,
		}

		err := strategy.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Missing Host", func(t *testing.T) {
		config := &common.Config{
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
		}

		err := strategy.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host é obrigatório para GORM")
	})
}

func TestStrategyCreation(t *testing.T) {
	t.Run("PGX Strategy Creation", func(t *testing.T) {
		strategy := NewPGXStrategy()
		assert.NotNil(t, strategy)
		assert.IsType(t, &PGXStrategy{}, strategy)
	})

	t.Run("PQ Strategy Creation", func(t *testing.T) {
		strategy := NewPQStrategy()
		assert.NotNil(t, strategy)
		assert.IsType(t, &PQStrategy{}, strategy)
	})

	t.Run("GORM Strategy Creation", func(t *testing.T) {
		strategy := NewGORMStrategy()
		assert.NotNil(t, strategy)
		assert.IsType(t, &GORMStrategy{}, strategy)
	})
}
