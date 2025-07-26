package checks

import "time"

// Constantes para formatos de data e hora
const (
	// RFC3339TimeOnlyFormat representa o formato de hora RFC3339 apenas
	RFC3339TimeOnlyFormat = "15:04:05Z07:00"

	// ISO8601DateTimeFormat representa o formato de data/hora ISO8601
	ISO8601DateTimeFormat = "2006-01-02T15:04:05-07:00"

	// ISO8601DateFormat representa o formato de data ISO8601
	ISO8601DateFormat = "2006-01-02"

	// ISO8601TimeFormat representa o formato de hora ISO8601
	ISO8601TimeFormat = "15:04:05"
)

// Formatos de data/hora comumente utilizados
var CommonDateTimeFormats = []string{
	time.TimeOnly,         // 15:04:05
	RFC3339TimeOnlyFormat, // 15:04:05Z07:00
	time.DateOnly,         // 2006-01-02
	time.RFC3339,          // 2006-01-02T15:04:05Z07:00
	time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
	ISO8601DateTimeFormat, // 2006-01-02T15:04:05-07:00
}
