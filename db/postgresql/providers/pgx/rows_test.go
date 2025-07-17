//go:build unit

package pgx

import (
	"testing"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/stretchr/testify/assert"
)

func TestPGXRows(t *testing.T) {
	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXRows implements IRows
		var _ interfaces.IRows = (*PGXRows)(nil)
	})

	t.Run("Close should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		// Close should be safe to call multiple times
		assert.NotPanics(t, func() {
			rows.Close()
			rows.Close()
		})
	})

	t.Run("Next should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		assert.NotPanics(t, func() {
			hasNext := rows.Next()
			// hasNext will be false for empty rows, which is expected
			_ = hasNext
		})
	})

	t.Run("Err should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		assert.NotPanics(t, func() {
			err := rows.Err()
			// err might be nil or an actual error, both are valid
			_ = err
		})
	})

	t.Run("Scan should not panic on empty rows", func(t *testing.T) {
		rows := &PGXRows{}
		var dest interface{}

		assert.NotPanics(t, func() {
			err := rows.Scan(&dest)
			// err will likely be non-nil for empty rows, which is expected
			_ = err
		})
	})

	t.Run("CommandTag should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		assert.NotPanics(t, func() {
			tag := rows.CommandTag()
			// tag might be nil for empty rows
			_ = tag
		})
	})

	t.Run("FieldDescriptions should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		assert.NotPanics(t, func() {
			fields := rows.FieldDescriptions()
			// fields will likely be nil or empty for empty rows
			_ = fields
		})
	})

	t.Run("RawValues should not panic", func(t *testing.T) {
		rows := &PGXRows{}

		assert.NotPanics(t, func() {
			values := rows.RawValues()
			// values will likely be nil or empty for empty rows
			_ = values
		})
	})
}

func TestPGXRow(t *testing.T) {
	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXRow implements IRow
		var _ interfaces.IRow = (*PGXRow)(nil)
	})

	t.Run("Scan should not panic on empty row", func(t *testing.T) {
		row := &PGXRow{}
		var dest interface{}

		assert.NotPanics(t, func() {
			err := row.Scan(&dest)
			// err will likely be non-nil for empty row, which is expected
			_ = err
		})
	})

	t.Run("Scan with multiple destinations", func(t *testing.T) {
		row := &PGXRow{}
		var dest1, dest2, dest3 interface{}

		assert.NotPanics(t, func() {
			err := row.Scan(&dest1, &dest2, &dest3)
			// err will likely be non-nil for empty row, which is expected
			_ = err
		})
	})

	t.Run("Scan with nil destination", func(t *testing.T) {
		row := &PGXRow{}

		assert.NotPanics(t, func() {
			err := row.Scan(nil)
			// err will likely be non-nil, which is expected
			_ = err
		})
	})
}

func TestPGXCommandTag(t *testing.T) {
	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXCommandTag implements CommandTag
		var _ interfaces.CommandTag = (*PGXCommandTag)(nil)
	})

	t.Run("String should not panic", func(t *testing.T) {
		tag := &PGXCommandTag{}

		assert.NotPanics(t, func() {
			str := tag.String()
			// str might be empty for empty tag
			_ = str
		})
	})

	t.Run("RowsAffected should not panic", func(t *testing.T) {
		tag := &PGXCommandTag{}

		assert.NotPanics(t, func() {
			rows := tag.RowsAffected()
			// rows will likely be 0 for empty tag
			_ = rows
		})
	})

	t.Run("Command type checks should not panic", func(t *testing.T) {
		tag := &PGXCommandTag{}

		assert.NotPanics(t, func() {
			_ = tag.Insert()
			_ = tag.Update()
			_ = tag.Delete()
			_ = tag.Select()
		})
	})
}

func TestPGXFieldDescription(t *testing.T) {
	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXFieldDescription implements FieldDescription
		var _ interfaces.FieldDescription = (*PGXFieldDescription)(nil)
	})

	t.Run("Field methods should not panic", func(t *testing.T) {
		field := &PGXFieldDescription{}

		assert.NotPanics(t, func() {
			_ = field.Name()
			_ = field.TableOID()
			_ = field.TableAttributeNumber()
			_ = field.DataTypeOID()
			_ = field.DataTypeSize()
			_ = field.TypeModifier()
			_ = field.Format()
		})
	})
}

// Benchmark tests
func BenchmarkPGXRows_Next(b *testing.B) {
	rows := &PGXRows{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows.Next()
	}
}

func BenchmarkPGXRows_Scan(b *testing.B) {
	rows := &PGXRows{}
	var dest interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows.Scan(&dest)
	}
}

func BenchmarkPGXRow_Scan(b *testing.B) {
	row := &PGXRow{}
	var dest interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row.Scan(&dest)
	}
}
