package json

import (
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-lib/json/interfaces"
)

type benchStruct struct {
	Name        string                 `json:"name"`
	Age         int                    `json:"age"`
	Address     string                 `json:"address"`
	Active      bool                   `json:"active"`
	PhoneNumber string                 `json:"phone_number"`
	Email       string                 `json:"email"`
	Tags        []string               `json:"tags"`
	Scores      []int                  `json:"scores"`
	Metadata    map[string]interface{} `json:"metadata"`
}

var (
	benchmarkValue = benchStruct{
		Name:        "Jane Doe",
		Age:         28,
		Address:     "456 Oak Street, Apt 2B, New York, NY 10001",
		Active:      true,
		PhoneNumber: "+1-555-123-4567",
		Email:       "jane.doe@example.com",
		Tags:        []string{"developer", "gopher", "json", "testing"},
		Scores:      []int{95, 87, 92, 78, 85},
		Metadata: map[string]interface{}{
			"created_at": "2023-06-01T10:30:00Z",
			"updated_at": "2023-06-15T14:45:30Z",
			"verified":   true,
			"preferences": map[string]interface{}{
				"theme":      "dark",
				"newsletter": true,
				"timezone":   "UTC-5",
			},
		},
	}

	smallJSON = `{"name":"Test User","age":30,"address":"123 Test St","active":true}`
	largeJSON = `{
		"name": "Jane Doe",
		"age": 28,
		"address": "456 Oak Street, Apt 2B, New York, NY 10001",
		"active": true,
		"phone_number": "+1-555-123-4567",
		"email": "jane.doe@example.com",
		"tags": ["developer", "gopher", "json", "testing"],
		"scores": [95, 87, 92, 78, 85],
		"metadata": {
			"created_at": "2023-06-01T10:30:00Z",
			"updated_at": "2023-06-15T14:45:30Z",
			"verified": true,
			"preferences": {
				"theme": "dark",
				"newsletter": true,
				"timezone": "UTC-5"
			}
		}
	}`
)

// Benchmark comparisons for all providers

func BenchmarkMarshalSmall(b *testing.B) {
	data := testStruct{
		Name:    "Test User",
		Age:     30,
		Address: "123 Test St",
		Active:  true,
	}

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = provider.Marshal(data)
			}
		})
	}
}

func BenchmarkUnmarshalSmall(b *testing.B) {
	jsonData := []byte(smallJSON)

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result testStruct
				_ = provider.Unmarshal(jsonData, &result)
			}
		})
	}
}

func BenchmarkMarshalLarge(b *testing.B) {
	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = provider.Marshal(benchmarkValue)
			}
		})
	}
}

func BenchmarkUnmarshalLarge(b *testing.B) {
	jsonData := []byte(largeJSON)

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result benchStruct
				_ = provider.Unmarshal(jsonData, &result)
			}
		})
	}
}

func BenchmarkDecodeReaderSmall(b *testing.B) {
	jsonStr := smallJSON

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(jsonStr)
				var result testStruct
				_ = provider.DecodeReader(reader, &result)
			}
		})
	}
}

func BenchmarkDecodeReaderLarge(b *testing.B) {
	jsonStr := largeJSON

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(jsonStr)
				var result benchStruct
				_ = provider.DecodeReader(reader, &result)
			}
		})
	}
}

func BenchmarkValidSmall(b *testing.B) {
	jsonData := []byte(smallJSON)

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = provider.Valid(jsonData)
			}
		})
	}
}

func BenchmarkValidLarge(b *testing.B) {
	jsonData := []byte(largeJSON)

	providers := map[string]interfaces.Provider{
		"stdlib":     New(Stdlib),
		"jsoniter":   New(JSONIter),
		"goccy":      New(GoccyJSON),
		"jsonparser": New(JSONParser),
	}

	for name, provider := range providers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = provider.Valid(jsonData)
			}
		})
	}
}
