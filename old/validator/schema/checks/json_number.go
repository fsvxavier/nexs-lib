package checks

import "encoding/json"

type JsonNumber struct{}

func (JsonNumber) IsFormat(input interface{}) bool {
	_, ok := input.(json.Number)
	return ok
}
