package parse

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func ParseEnvInt32(env string) *int32 {
	if valueStr := os.Getenv(env); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
			val := int32(value)
			return &val
		}
	}
	return nil
}

func ParseEnvInt(env string) *int {
	if valueStr := os.Getenv(env); valueStr != "" {
		if value, err := strconv.ParseInt(valueStr, 10, 0); err == nil {
			val := int(value)
			return &val
		}
	}
	return nil
}

func ParseEnvDuration(env string) *time.Duration {
	if valueStr := os.Getenv(env); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return &value
		}
	}
	return nil
}

func ParseEnvBool(env string) *bool {
	if valueStr := os.Getenv(env); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return &value
		}
	}
	return nil
}

func ParseEnvSliceString(env, separation string) *[]string {
	if valueStr := os.Getenv(env); valueStr != "" {
		if value := strings.Split(valueStr, separation); value != nil {
			return &value
		}
	}
	return nil
}

func ParseEnvString(env string) *string {
	if valueStr := os.Getenv(env); valueStr != "" {
		return &valueStr
	}
	return nil
}
