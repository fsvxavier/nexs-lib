# JSON Package

This package provides a unified interface for working with multiple JSON libraries in Go. It allows you to easily switch between different JSON implementations without changing your application code.

## Features

- **Unified Interface**: Common API for all JSON providers
- **Multiple Provider Support**:
  - `encoding/json` (Go standard library)
  - `github.com/json-iterator/go` (high-performance drop-in replacement)
  - `github.com/goccy/go-json` (high-performance JSON encoder/decoder)
  - `github.com/buger/jsonparser` (fast JSON parser without reflection)
- **Factory Pattern**: Easily create and switch between providers
- **Consistent API**: Same methods regardless of the underlying implementation
- **Performance Benchmarking**: Built-in benchmarks to compare provider performance

## Installation

```bash
go get github.com/fsvxavier/nexs-lib
```

## Quick Start

```go
import "github.com/fsvxavier/nexs-lib/json"

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Address string `json:"address"`
}

// Using default provider (stdlib)
jsonData := `{"name":"John Doe","age":30,"address":"123 Main St"}`
var person Person
err := json.Unmarshal([]byte(jsonData), &person)

// Marshal back to JSON
bytes, err := json.Marshal(person)
```

## Available Providers

| Provider    | Package                     | Best For                                           | Performance |
|-------------|-----------------------------|----------------------------------------------------|-------------|
| `Stdlib`    | encoding/json               | Standard compatibility, built-in to Go            | Baseline    |
| `JSONIter`  | github.com/json-iterator/go | High performance with full compatibility          | High        |
| `GoccyJSON` | github.com/goccy/go-json    | Maximum performance for encoding/decoding         | Highest     |
| `JSONParser`| github.com/buger/jsonparser | Specific field extraction without full parsing    | Specialized |

### Provider Selection Guide

- **Stdlib**: Default choice for most applications, maximum compatibility
- **JSONIter**: Drop-in replacement for stdlib with better performance
- **GoccyJSON**: Best choice for performance-critical applications
- **JSONParser**: Ideal for extracting specific values from large JSON documents

## Basic Usage

### Using the Default Provider

```go
import "github.com/fsvxavier/nexs-lib/json"

// All package-level functions use the default provider (stdlib)
err := json.Unmarshal(data, &result)
bytes, err := json.Marshal(object)
valid := json.Valid(data)
```

### Using a Specific Provider

```go
import "github.com/fsvxavier/nexs-lib/json"

// Create a specific provider
provider := json.New(json.JSONIter)

// Use provider methods
err := provider.Unmarshal(data, &result)
bytes, err := provider.Marshal(object)
valid := provider.Valid(data)
```

## Working with Streams

### Decoding from Reader

```go
import (
    "strings"
    "github.com/fsvxavier/nexs-lib/json"
)

reader := strings.NewReader(`{"name":"John","age":30}`)
var person Person
err := json.DecodeReader(reader, &person)
```

### Using Decoder and Encoder

```go
import (
    "os"
    "strings" 
    "github.com/fsvxavier/nexs-lib/json"
)

// Decoder with customization
reader := strings.NewReader(jsonData)
decoder := json.NewDecoder(reader)
decoder = decoder.UseNumber().DisallowUnknownFields()

var result MyStruct
err := decoder.Decode(&result)

// Encoder with pretty printing
encoder := json.NewEncoder(os.Stdout)
encoder = encoder.SetIndent("", "  ").SetEscapeHTML(false)
err = encoder.Encode(data)
```

## Examples

### Basic Example

```go
package main

import (
	"fmt"
	"github.com/fsvxavier/nexs-lib/json"
)

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

func main() {
	personJSON := `{"name":"John Doe","age":30,"address":"123 Main St"}`

	// Using default provider
	var person Person
	if err := json.Unmarshal([]byte(personJSON), &person); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Default provider: %+v\n", person)

	// Using specific providers
	providers := map[string]json.ProviderType{
		"JSONIter":   json.JSONIter,
		"GoccyJSON":  json.GoccyJSON,
		"JSONParser": json.JSONParser,
	}

	for name, providerType := range providers {
		provider := json.New(providerType)
		person = Person{} // Reset
		
		if err := provider.Unmarshal([]byte(personJSON), &person); err != nil {
			fmt.Printf("Error with %s: %v\n", name, err)
			continue
		}
		fmt.Printf("%s provider: %+v\n", name, person)
	}
}
```

### Stream Processing Example

```go
package main

import (
	"fmt"
	"strings"
	"github.com/fsvxavier/nexs-lib/json"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	// Processing JSON stream
	jsonData := `[
		{"id":1,"username":"user1","email":"user1@example.com"},
		{"id":2,"username":"user2","email":"user2@example.com"}
	]`

	reader := strings.NewReader(jsonData)
	decoder := json.NewDecoder(reader)

	var users []User
	if err := decoder.Decode(&users); err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}

	fmt.Println("Decoded users:")
	for i, user := range users {
		fmt.Printf("  %d: %+v\n", i+1, user)
	}
}
```

## Advanced Features

### Custom Decoder Configuration

```go
import (
    "strings"
    "github.com/fsvxavier/nexs-lib/json"
)

reader := strings.NewReader(`{"id": 1, "name": "John", "unknown": "value"}`)
decoder := json.NewDecoder(reader)

// Chain configuration methods
decoder = decoder.
    UseNumber().                    // Use json.Number for numbers
    DisallowUnknownFields()        // Reject unknown fields

var data struct {
    ID   json.Number `json:"id"`
    Name string      `json:"name"`
}

err := decoder.Decode(&data)
```

### Pretty Printing and Formatting

```go
import (
    "bytes"
    "github.com/fsvxavier/nexs-lib/json"
)

data := Person{Name: "John", Age: 30, Address: "123 Main St"}
buf := new(bytes.Buffer)

encoder := json.NewEncoder(buf)
encoder = encoder.
    SetIndent("", "    ").         // 4-space indentation
    SetEscapeHTML(false)           // Don't escape HTML characters

if err := encoder.Encode(data); err != nil {
    // Handle error
}

prettyJSON := buf.String()
fmt.Println(prettyJSON)
```

### Validation

```go
import "github.com/fsvxavier/nexs-lib/json"

jsonData := []byte(`{"name":"John","age":30}`)

// Validate JSON syntax
if !json.Valid(jsonData) {
    fmt.Println("Invalid JSON")
    return
}

// You can also validate with specific providers
provider := json.New(json.GoccyJSON)
if !provider.Valid(jsonData) {
    fmt.Println("Invalid JSON according to GoccyJSON")
}
```

## Performance and Benchmarking

This package includes comprehensive benchmarks to help you choose the right provider for your use case. Run benchmarks with:

```bash
go test -bench=. -benchmem
```

### Typical Performance Characteristics

- **Stdlib**: Reliable baseline performance, full feature support
- **JSONIter**: 2-3x faster than stdlib for most operations
- **GoccyJSON**: 3-5x faster than stdlib, especially for large objects
- **JSONParser**: Excellent for selective field extraction, not suitable for full object marshaling

## Testing

The package includes comprehensive test coverage for all providers:

```bash
# Run all tests
go test -v ./...

# Run tests for a specific provider
go test -v ./providers/stdlib
go test -v ./providers/jsoniter
go test -v ./providers/goccy
go test -v ./providers/jsonparser

# Run benchmarks
go test -bench=BenchmarkMarshal -benchmem
go test -bench=BenchmarkUnmarshal -benchmem
```

## Error Handling

All providers implement consistent error handling:

```go
import "github.com/fsvxavier/nexs-lib/json"

// Unmarshal errors
var person Person
if err := json.Unmarshal(invalidJSON, &person); err != nil {
    fmt.Printf("Unmarshal error: %v\n", err)
}

// Marshal errors
invalidData := make(chan int) // channels can't be marshaled
if _, err := json.Marshal(invalidData); err != nil {
    fmt.Printf("Marshal error: %v\n", err)
}

// Validation
if !json.Valid([]byte("invalid json")) {
    fmt.Println("JSON validation failed")
}
```

## Extending the Package

To implement a custom provider, implement the `Provider` interface:

```go
package myprovider

import (
    "io"
    "github.com/fsvxavier/nexs-lib/json/interfaces"
)

type MyProvider struct {
    // Your implementation fields
}

func New() interfaces.Provider {
    return &MyProvider{}
}

func (p *MyProvider) Marshal(v interface{}) ([]byte, error) {
    // Your marshal implementation
}

func (p *MyProvider) Unmarshal(data []byte, v interface{}) error {
    // Your unmarshal implementation
}

// Implement other required methods...
```

Then register and use your provider:

```go
// In your factory or main code
func NewMyProvider() interfaces.Provider {
    return myprovider.New()
}

// Usage
provider := NewMyProvider()
data, err := provider.Marshal(myObject)
```

## API Reference

### Package-Level Functions (Default Provider)

```go
func Marshal(v interface{}) ([]byte, error)
func Unmarshal(data []byte, v interface{}) error
func Valid(data []byte) bool
func Encode(v interface{}) ([]byte, error)
func DecodeReader(r io.Reader, v interface{}) error
func NewDecoder(r io.Reader) interfaces.Decoder
func NewEncoder(w io.Writer) interfaces.Encoder
```

### Provider Interface

```go
type Provider interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
    NewDecoder(r io.Reader) Decoder
    NewEncoder(w io.Writer) Encoder
    Valid(data []byte) bool
    DecodeReader(r io.Reader, v interface{}) error
    Encode(v interface{}) ([]byte, error)
}
```

### Decoder Interface

```go
type Decoder interface {
    Decode(v interface{}) error
    UseNumber() Decoder
    DisallowUnknownFields() Decoder
    Buffered() io.Reader
    Token() (json.Token, error)
    More() bool
}
```

### Encoder Interface

```go
type Encoder interface {
    Encode(v interface{}) error
    SetIndent(prefix, indent string) Encoder
    SetEscapeHTML(on bool) Encoder
}
```

## Contributing

When contributing to this package:

1. Add tests for any new functionality
2. Include benchmarks for performance-sensitive changes
3. Update documentation as needed
4. Ensure all providers maintain interface compatibility

## License

This package is part of the nexs-lib library. See the main repository for license information.
