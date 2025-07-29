package valkeyglide

import (
	"fmt"
	"strconv"
)

// command implementa interfaces.ICommand para comandos individuais.
type command struct {
	result interface{}
	err    error
}

// newCommand cria um novo comando com resultado e erro.
func newCommand(result interface{}, err error) *command {
	return &command{
		result: result,
		err:    err,
	}
}

// Result retorna o resultado do comando.
func (c *command) Result() (interface{}, error) {
	return c.result, c.err
}

// Err retorna o erro do comando.
func (c *command) Err() error {
	return c.err
}

// String converte o resultado para string.
func (c *command) String() (string, error) {
	if c.err != nil {
		return "", c.err
	}

	if c.result == nil {
		return "", nil
	}

	switch v := c.result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		if v {
			return "1", nil
		}
		return "0", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// Int64 converte o resultado para int64.
func (c *command) Int64() (int64, error) {
	if c.err != nil {
		return 0, c.err
	}

	if c.result == nil {
		return 0, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("não é possível converter %T para int64", v)
	}
}

// Bool converte o resultado para bool.
func (c *command) Bool() (bool, error) {
	if c.err != nil {
		return false, c.err
	}

	if c.result == nil {
		return false, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v) != "0", nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v) != "0", nil
	case string:
		return v != "" && v != "0", nil
	case []byte:
		return len(v) > 0 && string(v) != "0", nil
	default:
		return false, fmt.Errorf("não é possível converter %T para bool", v)
	}
}

// Float64 converte o resultado para float64.
func (c *command) Float64() (float64, error) {
	if c.err != nil {
		return 0, c.err
	}

	if c.result == nil {
		return 0, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int8, int16, int32, int64:
		return float64(fmt.Sprintf("%d", v)[0]), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(fmt.Sprintf("%d", v)[0]), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	default:
		return 0, fmt.Errorf("não é possível converter %T para float64", v)
	}
}

// Slice converte o resultado para []interface{}.
func (c *command) Slice() ([]interface{}, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.result == nil {
		return nil, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	case []int:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []int64:
		result := make([]interface{}, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	default:
		return []interface{}{v}, nil
	}
}

// StringSlice converte o resultado para []string.
func (c *command) StringSlice() ([]string, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.result == nil {
		return nil, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result, nil
	case string:
		return []string{v}, nil
	default:
		return nil, fmt.Errorf("não é possível converter %T para []string", v)
	}
}

// StringMap converte o resultado para map[string]string.
func (c *command) StringMap() (map[string]string, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.result == nil {
		return nil, fmt.Errorf("resultado é nil")
	}

	switch v := c.result.(type) {
	case map[string]string:
		return v, nil
	case map[string]interface{}:
		result := make(map[string]string)
		for key, value := range v {
			result[key] = fmt.Sprintf("%v", value)
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make(map[string]string)
		for key, value := range v {
			result[fmt.Sprintf("%v", key)] = fmt.Sprintf("%v", value)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("não é possível converter %T para map[string]string", v)
	}
}
