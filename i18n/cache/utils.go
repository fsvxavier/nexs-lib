package cache

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

// hashMap creates a deterministic hash of a map for cache key generation
func hashMap(m map[string]interface{}) string {
	if m == nil {
		return ""
	}

	// Sort keys for deterministic order
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build string representation
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(fmt.Sprintf("%v", m[k]))
		sb.WriteString(";")
	}

	// Hash the string
	h := fnv.New32a()
	h.Write([]byte(sb.String()))
	return strconv.FormatUint(uint64(h.Sum32()), 16)
}

// getPluralKey returns the plural key for a given count
func getPluralKey(count int) string {
	if count == 1 {
		return "one"
	}
	return "other"
}
