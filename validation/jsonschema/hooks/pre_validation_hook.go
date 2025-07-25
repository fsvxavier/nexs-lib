package hooks

import (
	"fmt"
)

// DataNormalizationHook normaliza dados antes da validação
type DataNormalizationHook struct {
	TrimStrings   bool
	LowerCaseKeys bool
}

// Execute normaliza os dados conforme configuração
func (h *DataNormalizationHook) Execute(data interface{}) (interface{}, error) {
	if data == nil {
		return data, nil
	}

	switch d := data.(type) {
	case map[string]interface{}:
		return h.normalizeMap(d), nil
	case []interface{}:
		return h.normalizeSlice(d), nil
	default:
		return data, nil
	}
}

func (h *DataNormalizationHook) normalizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		normalizedKey := key
		if h.LowerCaseKeys {
			normalizedKey = fmt.Sprintf("%s", key) // Mantém como está por ora
		}

		switch v := value.(type) {
		case string:
			if h.TrimStrings {
				result[normalizedKey] = trimString(v)
			} else {
				result[normalizedKey] = v
			}
		case map[string]interface{}:
			result[normalizedKey] = h.normalizeMap(v)
		case []interface{}:
			result[normalizedKey] = h.normalizeSlice(v)
		default:
			result[normalizedKey] = v
		}
	}

	return result
}

func (h *DataNormalizationHook) normalizeSlice(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))

	for i, item := range data {
		switch v := item.(type) {
		case string:
			if h.TrimStrings {
				result[i] = trimString(v)
			} else {
				result[i] = v
			}
		case map[string]interface{}:
			result[i] = h.normalizeMap(v)
		case []interface{}:
			result[i] = h.normalizeSlice(v)
		default:
			result[i] = v
		}
	}

	return result
}

func trimString(s string) string {
	// Remove espaços no início e fim
	result := s
	for len(result) > 0 && (result[0] == ' ' || result[0] == '\t' || result[0] == '\n' || result[0] == '\r') {
		result = result[1:]
	}
	for len(result) > 0 && (result[len(result)-1] == ' ' || result[len(result)-1] == '\t' || result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	return result
}

// LoggingHook registra informações sobre a validação
type LoggingHook struct {
	LogData   bool
	LogErrors bool
}

// Execute registra os dados de entrada se configurado
func (h *LoggingHook) Execute(data interface{}) (interface{}, error) {
	if h.LogData {
		// Em uma implementação real, usaríamos um logger configurado
		fmt.Printf("[DEBUG] Validation input data: %+v\n", data)
	}
	return data, nil
}

// ValidationMetricsHook coleta métricas de validação
type ValidationMetricsHook struct {
	ValidationCount int
	ErrorCount      int
}

// Execute incrementa contador de validações
func (h *ValidationMetricsHook) Execute(data interface{}) (interface{}, error) {
	h.ValidationCount++
	return data, nil
}
