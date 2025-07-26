package strutil

import (
	"fmt"
	"runtime"
	"strings"
)

// MapLines applies a function to each line of the input string
func MapLines(str string, fn func(string) string) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = fn(line)
	}
	return strings.Join(lines, "\n")
}

// SplitAndMap splits a string by a separator and applies a function to each part
func SplitAndMap(str, separator string, fn func(string) string) []string {
	parts := strings.Split(str, separator)
	for i, part := range parts {
		parts[i] = fn(part)
	}
	return parts
}

// OSNewLine returns the newline character(s) for the current operating system
func OSNewLine() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	default:
		return "\n"
	}
}

// ReplaceAllToOne replaces every string in the "from" slice with the string "to"
func ReplaceAllToOne(str string, from []string, to string) string {
	arr := make([]string, len(from)*2)
	for i, s := range from {
		arr[i*2] = s
		arr[i*2+1] = to
	}
	r := strings.NewReplacer(arr...)
	return r.Replace(str)
}

// Splice inserts a new string in place of the string between start and end indexes.
// It is based on runes so start and end indexes are rune based indexes.
// It can be used to remove a part of string by giving newStr as empty string.
func Splice(str, newStr string, start, end int) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	size := len(runes)
	if start < 0 || start > size {
		panic(fmt.Sprintf("start (%d) is out of range (%d)", start, size))
	}
	if end < start || end > size {
		panic(fmt.Sprintf("end (%d) is out of range (%d)", end, size))
	}
	return string(runes[:start]) + newStr + string(runes[end:])
}

// MustSubstring gets a part of the string between start and end. If end is 0,
// end is taken as the length of the string.
//
// It is UTF8 safe version of using slice notations in strings. It panics
// when the indexes are out of range. String length can be get with
// Len function before using Substring. You can use "Substring" if
// you prefer errors to panics.
func MustSubstring(str string, start, end int) string {
	res, err := Substring(str, start, end)
	if err != nil {
		panic(err)
	}
	return res
}

// Substring gets a part of the string between start and end. If end is 0,
// end is taken as the length of the string.
//
// MustSubstring can be used for the cases where the boundaries are well known and/or panics are
// acceptable
//
// It is UTF8 safe version of using slice notations in strings.
func Substring(str string, start, end int) (string, error) {
	runes := []rune(str)
	size := len(runes)

	if start < 0 || start >= size {
		return "", fmt.Errorf("start (%d) is out of range", start)
	}
	if end == 0 {
		end = size
	}
	if end <= start {
		return "", fmt.Errorf("end (%d) cannot be equal to or smaller than start (%d)", end, start)
	}
	if end > size {
		return "", fmt.Errorf("end (%d) is out of range", end)
	}

	return string(runes[start:end]), nil
}
