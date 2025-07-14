package gorm

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/stretchr/testify/assert"
)

func TestConn_Interface(t *testing.T) {
	conn := &Conn{}
	// Verify that Conn implements IConn interface
	var _ postgresql.IConn = conn
}

func TestConn_QueryOne(t *testing.T) {
	conn := &Conn{released: true}
	ctx := context.Background()
	var result interface{}

	t.Run("query with released connection", func(t *testing.T) {
		err := conn.QueryOne(ctx, &result, "SELECT 1")
		assert.Error(t, err, "Should fail with released connection")
		assert.Contains(t, err.Error(), "released", "Error should mention connection is released")
	})
}

func TestConn_QueryAll(t *testing.T) {
	conn := &Conn{released: true}
	ctx := context.Background()
	var results []interface{}

	t.Run("query all with released connection", func(t *testing.T) {
		err := conn.QueryAll(ctx, &results, "SELECT 1")
		assert.Error(t, err, "Should fail with released connection")
		assert.Contains(t, err.Error(), "released", "Error should mention connection is released")
	})
}

func TestConn_QueryCount(t *testing.T) {
	conn := &Conn{released: true}
	ctx := context.Background()

	t.Run("query count with released connection", func(t *testing.T) {
		_, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM test")
		assert.Error(t, err, "Should fail with released connection")
		assert.Contains(t, err.Error(), "released", "Error should mention connection is released")
	})
}

func TestConn_Exec(t *testing.T) {
	conn := &Conn{released: true}
	ctx := context.Background()

	t.Run("exec with released connection", func(t *testing.T) {
		err := conn.Exec(ctx, "INSERT INTO test VALUES (1)")
		assert.Error(t, err, "Should fail with released connection")
		assert.Contains(t, err.Error(), "released", "Error should mention connection is released")
	})
}

func TestConn_Release(t *testing.T) {
	conn := &Conn{}
	ctx := context.Background()

	t.Run("release connection", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			conn.Release(ctx)
		}, "Release should not panic")

		// After release, operations should fail
		var result interface{}
		err := conn.QueryOne(ctx, &result, "SELECT 1")
		assert.Error(t, err, "Operations should fail after release")
	})
}

func TestConn_BeginTransaction(t *testing.T) {
	conn := &Conn{released: true}
	ctx := context.Background()

	t.Run("begin transaction with released connection", func(t *testing.T) {
		_, err := conn.BeginTransaction(ctx)
		assert.Error(t, err, "Should fail with released connection")
		assert.Contains(t, err.Error(), "released", "Error should mention connection is released")
	})
}
