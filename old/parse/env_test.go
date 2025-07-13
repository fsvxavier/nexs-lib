package parse

import (
	"os"
	"testing"
	"time"
)

func TestParseEnvInt32(t *testing.T) {
	os.Setenv("TEST_INT32", "123")
	defer os.Unsetenv("TEST_INT32")

	result := ParseEnvInt32("TEST_INT32")
	if result == nil || *result != 123 {
		t.Errorf("Expected 123, got %v", result)
	}

	os.Setenv("TEST_INT32", "invalid")
	result = ParseEnvInt32("TEST_INT32")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "456")
	defer os.Unsetenv("TEST_INT")

	result := ParseEnvInt("TEST_INT")
	if result == nil || *result != 456 {
		t.Errorf("Expected 456, got %v", result)
	}

	os.Setenv("TEST_INT", "invalid")
	result = ParseEnvInt("TEST_INT")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseEnvDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "1h30m")
	defer os.Unsetenv("TEST_DURATION")

	expected, _ := time.ParseDuration("1h30m")
	result := ParseEnvDuration("TEST_DURATION")
	if result == nil || *result != expected {
		t.Errorf("Expected 1h30m, got %v", result)
	}

	os.Setenv("TEST_DURATION", "invalid")
	result = ParseEnvDuration("TEST_DURATION")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseEnvBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := ParseEnvBool("TEST_BOOL")
	if result == nil || *result != true {
		t.Errorf("Expected true, got %v", result)
	}

	os.Setenv("TEST_BOOL", "invalid")
	result = ParseEnvBool("TEST_BOOL")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseEnvSliceString(t *testing.T) {
	os.Setenv("TEST_SLICE", "a,b,c")
	defer os.Unsetenv("TEST_SLICE")

	expected := []string{"a", "b", "c"}
	result := ParseEnvSliceString("TEST_SLICE", ",")
	if result == nil || len(*result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	for i, v := range expected {
		if (*result)[i] != v {
			t.Errorf("Expected %v, got %v", v, (*result)[i])
		}
	}

	os.Setenv("TEST_SLICE", "")
	result = ParseEnvSliceString("TEST_SLICE", ",")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseEnvString(t *testing.T) {
	os.Setenv("TEST_STRING", "hello")
	defer os.Unsetenv("TEST_STRING")

	result := ParseEnvString("TEST_STRING")
	if result == nil || *result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}

	os.Setenv("TEST_STRING", "")
	result = ParseEnvString("TEST_STRING")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}
