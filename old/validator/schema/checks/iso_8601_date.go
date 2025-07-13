package checks

import "time"

type Iso8601Date struct{}

func (Iso8601Date) IsFormat(input interface{}) bool {
	return IsISO8601Date(input)
}

func IsISO8601Date(v interface{}) bool {
	if !IsString(v) {
		return false
	}
	_, err := time.Parse(ISO8601DateTimeFormat, v.(string))
	return err == nil
}
