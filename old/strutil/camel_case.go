package strutil

import (
	"strings"
)

// Converts a string to CamelCase.
func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	a, hasAcronym := uppercaseAcronym.Load(s)
	if hasAcronym {
		s = a.(string)
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := initCase
	prevIsCap := false
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		switch {
		case capNext:
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		case i == 0:
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		case prevIsCap && vIsCap && !hasAcronym:
			v += 'a'
			v -= 'A'
		}

		prevIsCap = vIsCap

		if vIsCap || vIsLow {
			n.WriteByte(v)
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}

// ToCamel converts a string to CamelCase.
func ToCamel(s string) string {
	return toCamelInitCase(s, true)
}

// ToLowerCamel converts a string to lowerCamelCase.
func ToLowerCamel(s string) string {
	return toCamelInitCase(s, false)
}
