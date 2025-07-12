package registry

import (
	"testing"
)

func TestSimpleRegistry(t *testing.T) {
	registry := NewErrorCodeRegistry()
	if registry == nil {
		t.Error("Registry should not be nil")
	}

	// Test basic operations
	list := registry.List()
	_ = list

	exists := registry.Exists("NON_EXISTENT")
	_ = exists

	global := GetGlobalRegistry()
	_ = global
}
