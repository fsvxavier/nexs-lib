package checks

import (
	"reflect"
	"regexp"
)

type StrongNameFormat struct{}

var strongNameRegex *regexp.Regexp

func init() {
	strongNameRegex = regexp.MustCompile("^[A-Z_]*$")
}

func (StrongNameFormat) IsFormat(input any) bool {
	switch reflect.TypeOf(input).String() {
	case "string":
		if input.(string) == "" {
			return false
		}
		return strongNameRegex.MatchString(input.(string))
	default:
		return false
	}
}
