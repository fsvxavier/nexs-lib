package mocks

// MockRow implements interfaces.IRow for testing
type MockRow struct {
	ScanFunc func(dest ...interface{}) error
}

// NewMockRow creates a new mock row
func NewMockRow() *MockRow {
	return &MockRow{}
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}
