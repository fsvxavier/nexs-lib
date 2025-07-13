package checks

func IsString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}
