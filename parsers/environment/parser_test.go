package environment

import (
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_GetString(t *testing.T) {
	parser := NewParser()

	t.Run("existing value", func(t *testing.T) {
		os.Setenv("TEST_STRING", "hello world")
		defer os.Unsetenv("TEST_STRING")

		result := parser.GetString("TEST_STRING")
		assert.Equal(t, "hello world", result)
	})

	t.Run("with default", func(t *testing.T) {
		result := parser.GetString("NON_EXISTENT", "default value")
		assert.Equal(t, "default value", result)
	})

	t.Run("empty string", func(t *testing.T) {
		os.Setenv("EMPTY_STRING", "")
		defer os.Unsetenv("EMPTY_STRING")

		result := parser.GetString("EMPTY_STRING", "default")
		assert.Equal(t, "default", result)
	})
}

func TestParser_GetInt(t *testing.T) {
	parser := NewParser()

	t.Run("valid integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		result := parser.GetInt("TEST_INT")
		assert.Equal(t, 42, result)
	})

	t.Run("invalid integer with default", func(t *testing.T) {
		os.Setenv("INVALID_INT", "not-a-number")
		defer os.Unsetenv("INVALID_INT")

		result := parser.GetInt("INVALID_INT", 100)
		assert.Equal(t, 100, result)
	})

	t.Run("missing with default", func(t *testing.T) {
		result := parser.GetInt("MISSING_INT", 200)
		assert.Equal(t, 200, result)
	})

	t.Run("missing without default", func(t *testing.T) {
		result := parser.GetInt("MISSING_INT")
		assert.Equal(t, 0, result)
	})
}

func TestParser_GetInt32(t *testing.T) {
	parser := NewParser()

	t.Run("valid int32", func(t *testing.T) {
		os.Setenv("TEST_INT32", "2147483647")
		defer os.Unsetenv("TEST_INT32")

		result := parser.GetInt32("TEST_INT32")
		assert.Equal(t, int32(2147483647), result)
	})

	t.Run("with default", func(t *testing.T) {
		result := parser.GetInt32("MISSING_INT32", 42)
		assert.Equal(t, int32(42), result)
	})
}

func TestParser_GetInt64(t *testing.T) {
	parser := NewParser()

	t.Run("valid int64", func(t *testing.T) {
		os.Setenv("TEST_INT64", "9223372036854775807")
		defer os.Unsetenv("TEST_INT64")

		result := parser.GetInt64("TEST_INT64")
		assert.Equal(t, int64(9223372036854775807), result)
	})

	t.Run("with default", func(t *testing.T) {
		result := parser.GetInt64("MISSING_INT64", 42)
		assert.Equal(t, int64(42), result)
	})
}

func TestParser_GetFloat64(t *testing.T) {
	parser := NewParser()

	t.Run("valid float", func(t *testing.T) {
		os.Setenv("TEST_FLOAT", "3.14159")
		defer os.Unsetenv("TEST_FLOAT")

		result := parser.GetFloat64("TEST_FLOAT")
		assert.Equal(t, 3.14159, result)
	})

	t.Run("with default", func(t *testing.T) {
		result := parser.GetFloat64("MISSING_FLOAT", 2.71828)
		assert.Equal(t, 2.71828, result)
	})
}

func TestParser_GetBool(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"on", "on", true},
		{"enabled", "enabled", true},
		{"TRUE", "TRUE", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"off", "off", false},
		{"disabled", "disabled", false},
		{"FALSE", "FALSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL_" + tt.name
			os.Setenv(key, tt.value)
			defer os.Unsetenv(key)

			result := parser.GetBool(key)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("with default", func(t *testing.T) {
		result := parser.GetBool("MISSING_BOOL", true)
		assert.Equal(t, true, result)
	})

	t.Run("invalid value with default", func(t *testing.T) {
		os.Setenv("INVALID_BOOL", "maybe")
		defer os.Unsetenv("INVALID_BOOL")

		result := parser.GetBool("INVALID_BOOL", true)
		assert.Equal(t, true, result)
	})
}

func TestParser_GetDuration(t *testing.T) {
	parser := NewParser()

	t.Run("valid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "1h30m")
		defer os.Unsetenv("TEST_DURATION")

		result := parser.GetDuration("TEST_DURATION")
		expected := time.Hour + 30*time.Minute
		assert.Equal(t, expected, result)
	})

	t.Run("with default", func(t *testing.T) {
		defaultDuration := 5 * time.Minute
		result := parser.GetDuration("MISSING_DURATION", defaultDuration)
		assert.Equal(t, defaultDuration, result)
	})

	t.Run("invalid duration with default", func(t *testing.T) {
		os.Setenv("INVALID_DURATION", "not-a-duration")
		defer os.Unsetenv("INVALID_DURATION")

		defaultDuration := 10 * time.Second
		result := parser.GetDuration("INVALID_DURATION", defaultDuration)
		assert.Equal(t, defaultDuration, result)
	})
}

func TestParser_GetSlice(t *testing.T) {
	parser := NewParser()

	t.Run("comma separated", func(t *testing.T) {
		os.Setenv("TEST_SLICE", "apple,banana,cherry")
		defer os.Unsetenv("TEST_SLICE")

		result := parser.GetSlice("TEST_SLICE", ",")
		expected := []string{"apple", "banana", "cherry"}
		assert.Equal(t, expected, result)
	})

	t.Run("custom separator", func(t *testing.T) {
		os.Setenv("TEST_SLICE_PIPE", "one|two|three")
		defer os.Unsetenv("TEST_SLICE_PIPE")

		result := parser.GetSlice("TEST_SLICE_PIPE", "|")
		expected := []string{"one", "two", "three"}
		assert.Equal(t, expected, result)
	})

	t.Run("with spaces", func(t *testing.T) {
		os.Setenv("TEST_SLICE_SPACES", "  first  ,  second  ,  third  ")
		defer os.Unsetenv("TEST_SLICE_SPACES")

		result := parser.GetSlice("TEST_SLICE_SPACES", ",")
		expected := []string{"first", "second", "third"}
		assert.Equal(t, expected, result)
	})

	t.Run("with default", func(t *testing.T) {
		defaultSlice := []string{"default1", "default2"}
		result := parser.GetSlice("MISSING_SLICE", ",", defaultSlice)
		assert.Equal(t, defaultSlice, result)
	})

	t.Run("empty value with default", func(t *testing.T) {
		os.Setenv("EMPTY_SLICE", "")
		defer os.Unsetenv("EMPTY_SLICE")

		defaultSlice := []string{"default"}
		result := parser.GetSlice("EMPTY_SLICE", ",", defaultSlice)
		assert.Equal(t, defaultSlice, result)
	})
}

func TestParser_GetMap(t *testing.T) {
	parser := NewParser()

	t.Run("key=value pairs", func(t *testing.T) {
		os.Setenv("TEST_MAP", "key1=value1,key2=value2,key3=value3")
		defer os.Unsetenv("TEST_MAP")

		result := parser.GetMap("TEST_MAP", ",", "=")
		expected := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("custom separators", func(t *testing.T) {
		os.Setenv("TEST_MAP_CUSTOM", "a:1|b:2|c:3")
		defer os.Unsetenv("TEST_MAP_CUSTOM")

		result := parser.GetMap("TEST_MAP_CUSTOM", "|", ":")
		expected := map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("with spaces", func(t *testing.T) {
		os.Setenv("TEST_MAP_SPACES", "  name = John  ,  age = 30  ")
		defer os.Unsetenv("TEST_MAP_SPACES")

		result := parser.GetMap("TEST_MAP_SPACES", ",", "=")
		expected := map[string]string{
			"name": "John",
			"age":  "30",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("empty value", func(t *testing.T) {
		result := parser.GetMap("MISSING_MAP", ",", "=")
		assert.Empty(t, result)
	})
}

func TestParser_WithPrefix(t *testing.T) {
	parser := NewParser(WithPrefix("APP"))

	t.Run("gets prefixed value", func(t *testing.T) {
		os.Setenv("APP_PORT", "8080")
		defer os.Unsetenv("APP_PORT")

		result := parser.GetInt("PORT")
		assert.Equal(t, 8080, result)
	})

	t.Run("nested prefix", func(t *testing.T) {
		nestedParser := parser.WithPrefix("DB")

		os.Setenv("APP_DB_HOST", "localhost")
		defer os.Unsetenv("APP_DB_HOST")

		result := nestedParser.GetString("HOST")
		assert.Equal(t, "localhost", result)
	})
}

func TestParser_WithDefaults(t *testing.T) {
	defaults := map[string]string{
		"PORT":    "3000",
		"HOST":    "localhost",
		"DEBUG":   "false",
		"TIMEOUT": "30s",
	}

	parser := NewParser(WithDefaults(defaults))

	t.Run("uses default when not set", func(t *testing.T) {
		assert.Equal(t, 3000, parser.GetInt("PORT"))
		assert.Equal(t, "localhost", parser.GetString("HOST"))
		assert.Equal(t, false, parser.GetBool("DEBUG"))
		assert.Equal(t, 30*time.Second, parser.GetDuration("TIMEOUT"))
	})

	t.Run("overrides default when set", func(t *testing.T) {
		// Create a new parser to avoid cache issues
		newParser := NewParser(WithDefaults(defaults))

		os.Setenv("PORT", "8080")
		defer os.Unsetenv("PORT")

		assert.Equal(t, 8080, newParser.GetInt("PORT"))
	})
}

func TestParser_Validation(t *testing.T) {
	t.Run("required fields present", func(t *testing.T) {
		os.Setenv("REQUIRED_VAR", "value")
		defer os.Unsetenv("REQUIRED_VAR")

		parser := NewParser(WithRequired("REQUIRED_VAR"))

		err := parser.Validate()
		assert.NoError(t, err)
	})

	t.Run("required fields missing", func(t *testing.T) {
		parser := NewParser(WithRequired("MISSING_VAR1", "MISSING_VAR2"))

		err := parser.Validate()
		assert.Error(t, err)

		var multiErr *parsers.MultiError
		assert.ErrorAs(t, err, &multiErr)
		assert.Len(t, multiErr.Errors, 2)
	})
}

func TestParser_IsSet(t *testing.T) {
	parser := NewParser()

	t.Run("variable is set", func(t *testing.T) {
		os.Setenv("SET_VAR", "value")
		defer os.Unsetenv("SET_VAR")

		assert.True(t, parser.IsSet("SET_VAR"))
	})

	t.Run("variable is not set", func(t *testing.T) {
		assert.False(t, parser.IsSet("UNSET_VAR"))
	})

	t.Run("variable is empty", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		assert.False(t, parser.IsSet("EMPTY_VAR"))
	})
}

func TestParser_PointerMethods(t *testing.T) {
	parser := NewParser()

	t.Run("GetStringPtr", func(t *testing.T) {
		os.Setenv("STRING_PTR", "value")
		defer os.Unsetenv("STRING_PTR")

		result := parser.GetStringPtr("STRING_PTR")
		require.NotNil(t, result)
		assert.Equal(t, "value", *result)

		nilResult := parser.GetStringPtr("MISSING_STRING_PTR")
		assert.Nil(t, nilResult)
	})

	t.Run("GetIntPtr", func(t *testing.T) {
		os.Setenv("INT_PTR", "42")
		defer os.Unsetenv("INT_PTR")

		result := parser.GetIntPtr("INT_PTR")
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)

		nilResult := parser.GetIntPtr("MISSING_INT_PTR")
		assert.Nil(t, nilResult)
	})

	t.Run("GetBoolPtr", func(t *testing.T) {
		os.Setenv("BOOL_PTR", "true")
		defer os.Unsetenv("BOOL_PTR")

		result := parser.GetBoolPtr("BOOL_PTR")
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		nilResult := parser.GetBoolPtr("MISSING_BOOL_PTR")
		assert.Nil(t, nilResult)
	})

	t.Run("GetDurationPtr", func(t *testing.T) {
		os.Setenv("DURATION_PTR", "1h")
		defer os.Unsetenv("DURATION_PTR")

		result := parser.GetDurationPtr("DURATION_PTR")
		require.NotNil(t, result)
		assert.Equal(t, time.Hour, *result)

		nilResult := parser.GetDurationPtr("MISSING_DURATION_PTR")
		assert.Nil(t, nilResult)
	})
}

func TestParser_Caching(t *testing.T) {
	parser := NewParser()

	t.Run("caches values", func(t *testing.T) {
		os.Setenv("CACHED_VAR", "original")

		// First call
		result1 := parser.GetString("CACHED_VAR")
		assert.Equal(t, "original", result1)

		// Change environment variable
		os.Setenv("CACHED_VAR", "changed")

		// Second call should return cached value
		result2 := parser.GetString("CACHED_VAR")
		assert.Equal(t, "original", result2) // Still cached

		os.Unsetenv("CACHED_VAR")
	})
}

// Package-level function tests

func TestPackageFunctions(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		os.Setenv("PKG_STRING", "package value")
		defer os.Unsetenv("PKG_STRING")

		result := GetString("PKG_STRING")
		assert.Equal(t, "package value", result)
	})

	t.Run("GetInt", func(t *testing.T) {
		os.Setenv("PKG_INT", "123")
		defer os.Unsetenv("PKG_INT")

		result := GetInt("PKG_INT")
		assert.Equal(t, 123, result)
	})

	t.Run("GetBool", func(t *testing.T) {
		os.Setenv("PKG_BOOL", "true")
		defer os.Unsetenv("PKG_BOOL")

		result := GetBool("PKG_BOOL")
		assert.Equal(t, true, result)
	})

	t.Run("IsSet", func(t *testing.T) {
		os.Setenv("PKG_SET", "value")
		defer os.Unsetenv("PKG_SET")

		assert.True(t, IsSet("PKG_SET"))
		assert.False(t, IsSet("PKG_NOT_SET"))
	})
}

// Benchmark tests

func BenchmarkParser_GetString(b *testing.B) {
	parser := NewParser()
	os.Setenv("BENCH_STRING", "benchmark value")
	defer os.Unsetenv("BENCH_STRING")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.GetString("BENCH_STRING")
	}
}

func BenchmarkParser_GetInt(b *testing.B) {
	parser := NewParser()
	os.Setenv("BENCH_INT", "42")
	defer os.Unsetenv("BENCH_INT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.GetInt("BENCH_INT")
	}
}

func BenchmarkParser_GetSlice(b *testing.B) {
	parser := NewParser()
	os.Setenv("BENCH_SLICE", "one,two,three,four,five")
	defer os.Unsetenv("BENCH_SLICE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.GetSlice("BENCH_SLICE", ",")
	}
}
