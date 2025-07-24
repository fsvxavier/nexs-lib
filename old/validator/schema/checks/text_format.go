package checks

import (
	"regexp"
)

type TextMatch struct{}

func (TextMatch) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z_ ]*$")

	return r.MatchString(asString)
}

type TextMatchWithNumber struct{}

func (TextMatchWithNumber) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z1-9_ ]*$")

	return r.MatchString(asString)
}

type TextMatchCustom struct {
	regex string
}

func NewTextMatchCustom(regex string) TextMatchCustom {
	return TextMatchCustom{regex: regex}
}

func (match TextMatchCustom) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	r, err := regexp.Compile(match.regex)
	if err != nil {
		return false
	}

	return r.MatchString(asString)
}
