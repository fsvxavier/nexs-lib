# Parsers Library

[![Go Reference](https://pkg.go.dev/badge/github.com/fsvxavier/nexs-lib/parsers.svg)](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/parsers)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-lib/parsers)](https://goreportcard.com/report/github.com/fsvxavier/nexs-lib/parsers)

A comprehensive and high-performance parsing library for Go that provides robust parsing capabilities for JSON, CSV, URLs, and more. Built with production-ready features including validation, error handling, streaming support, and extensive configuration options.

## 🚀 Features

### Core Parsers
- **JSON Parser**: High-performance JSON parsing with validation and error context
- **CSV Parser**: Flexible CSV parsing with type conversion and streaming support  
- **URL Parser**: Advanced URL parsing with domain extraction and validation
- **Extensible**: Easy to add new parser types

### Advanced Capabilities
- **Type Safety**: Generic parsers with compile-time type safety
- **Streaming Support**: Memory-efficient parsing for large datasets
- **Validation**: Built-in validation with detailed error reporting
- **Configuration**: Extensive configuration options for all parsers
- **Metadata**: Rich parsing metadata including timing and statistics
- **Error Handling**: Comprehensive error types with context information

### Production Features
- **Performance**: Optimized for high-throughput scenarios
- **Memory Efficient**: Streaming parsers and memory pooling
- **Context Support**: Full context.Context integration
- **Timeout Handling**: Configurable timeouts for parsing operations
- **Size Limits**: Configurable input size limits for security

## 📦 Installation

```bash
go get github.com/fsvxavier/nexs-lib/parsers
```

## 🔧 Quick Start

### Basic JSON Parsing

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fsvxavier/nexs-lib/parsers"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func main() {
    ctx := context.Background()
    
    // Parse JSON with convenience function
    jsonData := `{"name":"John","email":"john@example.com","age":30}`
    result, err := parsers.ParseJSONString(ctx, jsonData)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Parsed in %v\n", result.Metadata.Duration)
    fmt.Printf("Data: %+v\n", result.Data)
}
```

### Typed JSON Parsing

```go
func main() {
    ctx := context.Background()
    manager := parsers.NewManager()
    
    jsonData := []byte(`{"name":"John","email":"john@example.com","age":30}`)
    result, err := parsers.ParseJSONTyped[User](ctx, manager, jsonData)
    if err != nil {
        log.Fatal(err)
    }
    
    user := result.Data
    fmt.Printf("User: %s (%d) - %s\n", user.Name, user.Age, user.Email)
}
```

### CSV Parsing

```go
type Employee struct {
    Name     string `csv:"name"`
    Position string `csv:"position"`
    Salary   int    `csv:"salary"`
}

func main() {
    ctx := context.Background()
    factory := parsers.NewFactory()
    parser := parsers.CSVTyped[Employee](factory)
    
    csvData := []byte("name,position,salary\nJohn Doe,Developer,75000\nJane Smith,Manager,85000")
    
    // Parse all records
    employees, err := parser.ParseAll(ctx, csvData)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, emp := range employees {
        fmt.Printf("%s: %s ($%d)\n", emp.Name, emp.Position, emp.Salary)
    }
}
```

### URL Parsing

```go
func main() {
    ctx := context.Background()
    parser := parsers.NewFactory().URL()
    
    urlStr := "https://api.example.com:8443/v1/users?page=1&limit=10"
    result, err := parser.ParseString(ctx, urlStr)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Scheme: %s\n", result.Scheme)
    fmt.Printf("Host: %s\n", result.Host)
    fmt.Printf("Port: %d\n", result.Port)
    fmt.Printf("Path: %s\n", result.Path)
    fmt.Printf("Domain: %s\n", result.Domain)
    fmt.Printf("Is Secure: %v\n", result.IsSecure)
}
```

## 🔧 Advanced Usage

### Custom Configuration

```go
config := &interfaces.ParserConfig{
    Timeout:       5 * time.Second,
    MaxSize:       1024 * 1024, // 1MB limit
    StrictMode:    true,
    AllowComments: false,
    Encoding:      "utf-8",
}

factory := parsers.NewFactoryWithConfig(config)
parser := factory.JSON()
```

### Streaming CSV Parser

```go
func main() {
    ctx := context.Background()
    parser := parsers.CSVTyped[Employee](parsers.NewFactory())
    
    file, err := os.Open("large_file.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    count := 0
    err = parser.ParseStream(ctx, file, func(emp *Employee) error {
        count++
        fmt.Printf("Processed employee: %s\n", emp.Name)
        return nil
    })
    
    fmt.Printf("Processed %d employees\n", count)
}
```

### URL Builder

```go
url, err := parsers.NewBuilder().
    Scheme("https").
    Host("api.example.com").
    Port(8443).
    Path("/v1/users").
    AddParam("page", "1").
    AddParam("limit", "10").
    Fragment("results").
    Build()

if err != nil {
    log.Fatal(err)
}

fmt.Println(url.String())
// Output: https://api.example.com:8443/v1/users?page=1&limit=10#results
```

### URL Validation and Filtering

```go
parser := parsers.NewFactory().URL().
    WithAllowedHosts([]string{"example.com", "api.example.com"}).
    WithBlockedHosts([]string{"malicious.com"})

result, err := parser.ParseString(ctx, "https://api.example.com/data")
// Will succeed

result, err = parser.ParseString(ctx, "https://malicious.com/data")
// Will fail with validation error
```

## 🛠️ Error Handling

The library provides comprehensive error handling with detailed context:

```go
result, err := parser.Parse(ctx, invalidData)
if err != nil {
    var parseErr *interfaces.ParseError
    if errors.As(err, &parseErr) {
        fmt.Printf("Error Type: %s\n", parseErr.Type.String())
        fmt.Printf("Message: %s\n", parseErr.Message)
        fmt.Printf("Line: %d, Column: %d\n", parseErr.Line, parseErr.Column)
        fmt.Printf("Context: %s\n", parseErr.Context)
    }
}
```

## 📊 Performance

The library is optimized for high-performance scenarios:

- **Memory Pooling**: Reuses objects to reduce GC pressure
- **Streaming**: Memory-efficient processing of large files
- **Zero-Copy**: Minimizes memory allocations where possible
- **Concurrent Safe**: All parsers are safe for concurrent use

### Benchmarks

```
BenchmarkJSON_Parse-8        1000000    1250 ns/op    256 B/op    4 allocs/op
BenchmarkCSV_Parse-8          500000    2500 ns/op    512 B/op    8 allocs/op
BenchmarkURL_Parse-8         2000000     850 ns/op    128 B/op    2 allocs/op
```

## 🔧 Validation

### JSON Validation

```go
// Validate without parsing
err := parsers.ValidateJSONStr(`{"name":"test","age":30}`)
if err != nil {
    log.Printf("Invalid JSON: %v", err)
}
```

### URL Validation

```go
// Quick validation
if parsers.IsValidURL("https://example.com") {
    fmt.Println("Valid URL")
}
```

## 🔄 Transformation

### JSON Transformation

```go
transformer := parsers.NewTransformer()

// Compact JSON
compact, err := transformer.CompactJSON([]byte(`{
    "name": "test",
    "value": 123
}`))

// Pretty print JSON
pretty, err := transformer.PrettyJSON(compact, "  ")
```

### URL Transformation

```go
transformer := parsers.NewTransformer()

// Normalize URL (remove default ports, etc.)
normalized, err := transformer.NormalizeURL("https://example.com:443/path")
// Result: "https://example.com/path"

// Join URLs
joined, err := transformer.JoinURL("https://api.example.com/v1", "users/123")
// Result: "https://api.example.com/v1/users/123"

// Extract domain
domain, err := transformer.ExtractDomain("https://subdomain.example.com/path")
// Result: "example.com"
```

## 🎯 Type Safety

The library leverages Go generics for compile-time type safety:

```go
// Type-safe JSON parsing
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

parser := parsers.JSONTyped[Config](factory)
config, err := parser.Parse(ctx, jsonData)
// config is *Config, not interface{}
```

## 📋 Best Practices

### 1. Use Typed Parsers
```go
// Preferred: Type-safe
parser := parsers.JSONTyped[User](factory)

// Avoid: Requires type assertion
parser := factory.JSON()
```

### 2. Configure Limits
```go
config := &interfaces.ParserConfig{
    MaxSize: 10 * 1024 * 1024, // 10MB limit
    Timeout: 30 * time.Second,
}
```

### 3. Handle Errors Properly
```go
if err != nil {
    var parseErr *interfaces.ParseError
    if errors.As(err, &parseErr) {
        // Handle specific parse error
        switch parseErr.Type {
        case interfaces.ErrorTypeSyntax:
            // Handle syntax errors
        case interfaces.ErrorTypeValidation:
            // Handle validation errors
        }
    }
}
```

### 4. Use Streaming for Large Files
```go
// For large CSV files
err = parser.ParseStream(ctx, reader, func(record *Record) error {
    // Process each record individually
    return nil
})
```

## 🏗️ Architecture

The library follows clean architecture principles:

```
parsers/
├── interfaces/          # Core interfaces and types
├── json/               # JSON parser implementation
├── csv/                # CSV parser implementation  
├── url/                # URL parser implementation
├── parsers.go          # High-level API and convenience functions
└── examples/           # Usage examples
```

### Key Components

- **Parser Interface**: Generic interface for all parsers
- **StreamParser Interface**: For streaming operations
- **Formatter Interface**: For data formatting/serialization
- **Validator Interface**: For data validation
- **Factory**: Centralized parser creation
- **Manager**: High-level operations with metadata

## 🧪 Testing

The library has comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run with race detection
go test -race -timeout 30s ./...
```

## 🔧 Configuration Options

### Parser Configuration

```go
type ParserConfig struct {
    Timeout       time.Duration // Operation timeout
    MaxSize       int64         // Maximum input size
    StrictMode    bool          // Enable strict parsing
    AllowComments bool          // Allow comments (where applicable)
    Encoding      string        // Character encoding
}
```

### CSV-Specific Options

```go
parser := csvparser.NewParser[User]().
    WithDelimiter(';').                    // Custom delimiter
    WithHeaders([]string{"name", "age"})   // Explicit headers
```

### URL-Specific Options

```go
parser := urlparser.NewParser().
    WithAllowedHosts([]string{"example.com"}).  // Whitelist hosts
    WithBlockedHosts([]string{"blocked.com"})   // Blacklist hosts
```

## 📈 Monitoring and Metrics

Access detailed parsing metrics:

```go
result, err := manager.ParseJSON(ctx, data)
if err == nil {
    metrics := result.Metadata
    fmt.Printf("Parse Duration: %v\n", metrics.Duration)
    fmt.Printf("Bytes Processed: %d\n", metrics.BytesProcessed)
    fmt.Printf("Lines Processed: %d\n", metrics.LinesProcessed)
    fmt.Printf("Parser Version: %s\n", metrics.Version)
}
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Run linting (`golangci-lint run`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built following Go best practices and idioms
- Inspired by the need for high-performance, type-safe parsing
- Designed for production environments with comprehensive error handling

---

For more examples and detailed documentation, visit the [examples directory](examples/) or check the [GoDoc](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/parsers).
