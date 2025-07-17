package mocks

// MockCommandTag implements interfaces.CommandTag for testing
type MockCommandTag struct {
	StringFunc       func() string
	RowsAffectedFunc func() int64
	InsertFunc       func() bool
	UpdateFunc       func() bool
	DeleteFunc       func() bool
	SelectFunc       func() bool
}

// NewMockCommandTag creates a new mock command tag
func NewMockCommandTag() *MockCommandTag {
	return &MockCommandTag{}
}

func (m *MockCommandTag) String() string {
	if m.StringFunc != nil {
		return m.StringFunc()
	}
	return ""
}

func (m *MockCommandTag) RowsAffected() int64 {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0
}

func (m *MockCommandTag) Insert() bool {
	if m.InsertFunc != nil {
		return m.InsertFunc()
	}
	return false
}

func (m *MockCommandTag) Update() bool {
	if m.UpdateFunc != nil {
		return m.UpdateFunc()
	}
	return false
}

func (m *MockCommandTag) Delete() bool {
	if m.DeleteFunc != nil {
		return m.DeleteFunc()
	}
	return false
}

func (m *MockCommandTag) Select() bool {
	if m.SelectFunc != nil {
		return m.SelectFunc()
	}
	return false
}
