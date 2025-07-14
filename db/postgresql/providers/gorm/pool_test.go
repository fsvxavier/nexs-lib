package gorm

import (
	"testing"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/stretchr/testify/assert"
)

func TestPool_Interface(t *testing.T) {
	pool := &Pool{}
	// Verify that Pool implements IPool interface
	var _ interfaces.IPool = pool
}

func TestPool_Close(t *testing.T) {
	pool := &Pool{}

	t.Run("close pool", func(t *testing.T) {
		// Test that Close method exists and can be called
		assert.NotPanics(t, func() {
			pool.Close()
		}, "Close should not panic")
	})
}
