package valkeygo

import (
	"github.com/valkey-io/valkey-go"
)

// Command implementa interfaces.ICommand.
type Command struct {
	index  int
	result valkey.ValkeyResult
	err    error
}

// Result implementa interfaces.ICommand.Result.
func (c *Command) Result() (interface{}, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.result, nil
}

// Err implementa interfaces.ICommand.Err.
func (c *Command) Err() error {
	return c.err
}

// String implementa interfaces.ICommand.String.
func (c *Command) String() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return c.result.ToString()
}

// Int64 implementa interfaces.ICommand.Int64.
func (c *Command) Int64() (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return c.result.AsInt64()
}

// Bool implementa interfaces.ICommand.Bool.
func (c *Command) Bool() (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	val, err := c.result.AsInt64()
	return val == 1, err
}

// Float64 implementa interfaces.ICommand.Float64.
func (c *Command) Float64() (float64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return c.result.AsFloat64()
}

// Slice implementa interfaces.ICommand.Slice.
func (c *Command) Slice() ([]interface{}, error) {
	if c.err != nil {
		return nil, c.err
	}
	// Implementação básica - pode ser melhorada
	strs, err := c.result.AsStrSlice()
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result, nil
}

// StringSlice implementa interfaces.ICommand.StringSlice.
func (c *Command) StringSlice() ([]string, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.result.AsStrSlice()
}

// StringMap implementa interfaces.ICommand.StringMap.
func (c *Command) StringMap() (map[string]string, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.result.AsStrMap()
}
