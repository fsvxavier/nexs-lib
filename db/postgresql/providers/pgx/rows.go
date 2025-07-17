package pgx

import (
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PGXRows implements the IRows interface
type PGXRows struct {
	rows   pgx.Rows
	conn   *PGXConn
	query  string
	args   []interface{}
	closed bool
}

// Next implements IRows.Next
func (r *PGXRows) Next() bool {
	if r.closed || r.rows == nil {
		return false
	}
	return r.rows.Next()
}

// Scan implements IRows.Scan
func (r *PGXRows) Scan(dest ...interface{}) error {
	if r.closed {
		return pgx.ErrNoRows
	}
	if r.rows == nil {
		return pgx.ErrNoRows
	}
	return r.rows.Scan(dest...)
}

// Close implements IRows.Close
func (r *PGXRows) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.rows != nil {
		r.rows.Close()
	}
	return nil
}

// Err implements IRows.Err
func (r *PGXRows) Err() error {
	if r.rows == nil {
		return nil
	}
	return r.rows.Err()
}

// CommandTag implements IRows.CommandTag
func (r *PGXRows) CommandTag() interfaces.CommandTag {
	if r.closed || r.rows == nil {
		return &PGXCommandTag{tag: pgconn.CommandTag{}}
	}
	return &PGXCommandTag{tag: r.rows.CommandTag()}
}

// FieldDescriptions implements IRows.FieldDescriptions
func (r *PGXRows) FieldDescriptions() []interfaces.FieldDescription {
	if r.closed || r.rows == nil {
		return nil
	}

	if r.rows == nil {
		return nil
	}

	descriptions := r.rows.FieldDescriptions()
	result := make([]interfaces.FieldDescription, len(descriptions))

	for i, desc := range descriptions {
		result[i] = &PGXFieldDescription{desc: desc}
	}

	return result
}

// RawValues implements IRows.RawValues
func (r *PGXRows) RawValues() [][]byte {
	if r.closed || r.rows == nil {
		return nil
	}
	return r.rows.RawValues()
}

// PGXRow implements the IRow interface
type PGXRow struct {
	row   pgx.Row
	conn  *PGXConn
	query string
	args  []interface{}
}

// Scan implements IRow.Scan
func (r *PGXRow) Scan(dest ...interface{}) error {
	if r.row == nil {
		return pgx.ErrNoRows
	}
	return r.row.Scan(dest...)
}

// PGXCommandTag implements the CommandTag interface
type PGXCommandTag struct {
	tag pgconn.CommandTag
}

// String implements CommandTag.String
func (c *PGXCommandTag) String() string {
	return c.tag.String()
}

// RowsAffected implements CommandTag.RowsAffected
func (c *PGXCommandTag) RowsAffected() int64 {
	return c.tag.RowsAffected()
}

// Insert implements CommandTag.Insert
func (c *PGXCommandTag) Insert() bool {
	return c.tag.Insert()
}

// Update implements CommandTag.Update
func (c *PGXCommandTag) Update() bool {
	return c.tag.Update()
}

// Delete implements CommandTag.Delete
func (c *PGXCommandTag) Delete() bool {
	return c.tag.Delete()
}

// Select implements CommandTag.Select
func (c *PGXCommandTag) Select() bool {
	return c.tag.Select()
}

// PGXFieldDescription implements the FieldDescription interface
type PGXFieldDescription struct {
	desc pgconn.FieldDescription
}

// Name implements FieldDescription.Name
func (f *PGXFieldDescription) Name() string {
	return f.desc.Name
}

// TableOID implements FieldDescription.TableOID
func (f *PGXFieldDescription) TableOID() uint32 {
	return f.desc.TableOID
}

// TableAttributeNumber implements FieldDescription.TableAttributeNumber
func (f *PGXFieldDescription) TableAttributeNumber() uint16 {
	return f.desc.TableAttributeNumber
}

// DataTypeOID implements FieldDescription.DataTypeOID
func (f *PGXFieldDescription) DataTypeOID() uint32 {
	return f.desc.DataTypeOID
}

// DataTypeSize implements FieldDescription.DataTypeSize
func (f *PGXFieldDescription) DataTypeSize() int16 {
	return f.desc.DataTypeSize
}

// TypeModifier implements FieldDescription.TypeModifier
func (f *PGXFieldDescription) TypeModifier() int32 {
	return f.desc.TypeModifier
}

// Format implements FieldDescription.Format
func (f *PGXFieldDescription) Format() int16 {
	return f.desc.Format
}

// Helper function to create PGXRows
func newPGXRows(rows pgx.Rows, conn *PGXConn, query string, args []interface{}) *PGXRows {
	return &PGXRows{
		rows:  rows,
		conn:  conn,
		query: query,
		args:  args,
	}
}

// Helper function to create PGXRow
func newPGXRow(row pgx.Row, conn *PGXConn, query string, args []interface{}) *PGXRow {
	return &PGXRow{
		row:   row,
		conn:  conn,
		query: query,
		args:  args,
	}
}
