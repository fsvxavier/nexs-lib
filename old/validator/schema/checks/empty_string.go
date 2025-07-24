package checks

type EmptyStringChecker struct{}

func (EmptyStringChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if asString == "" {
		return false
	}

	return true
}
