# Parsers Library

A comprehensive, modern parsing library for Go applications that provides robust parsing capabilities for datetime, duration, and environment variables. 

## 🚀 **100% Legacy Compatibility**

This library provides **complete API compatibility** with legacy datetime parsing libraries while adding modern features and improvements. You can drop it in as a replacement without changing your existing code!

```go
// Your existing code continues to work exactly the same
date, err := datetime.ParseAny("02/03/2023", datetime.PreferMonthFirst(false))
date, err := datetime.ParseIn("2023-01-15 10:30", loc)
date, err := datetime.ParseLocal("15/01/2023")
date := datetime.MustParseAny("2023-01-15T10:30:45Z")
```

## Key Features

### 🕰️ DateTime Parser
- **🔄 100% Legacy API Compatibility**: Drop-in replacement for old dateparse libraries
- **⚡ Performance Optimized**: Smart caching and format ordering for better performance  
- **🌍 Complete Unix Timestamp Support**: Seconds, milliseconds, microseconds, and decimal precision
- **📅 Smart Date Handling**: `PreferMonthFirst`, `RetryAmbiguousDateWithSwap`, and intelligent format detection
- **🎯 Context-Aware**: Full cancellation and timeout support
- **📍 Timezone Intelligence**: Parse with custom locations, convert to UTC, handle local time
- **🤖 Natural Language**: Support for "today", "yesterday", "2 days ago", relative parsing
- **🔍 Format Detection**: Automatically detect and return the format used for any date string
- **💡 Advanced Text Parsing**: "January 15, 2023 10:30 AM", "15th of March 2023", etc.
- **📊 Comprehensive Errors**: Detailed error reporting with helpful suggestions

### ⏱️ Duration Parser
- **Extended units**: Support for days, weeks, months, years beyond standard Go durations
- **Multiple formats**: Parse "1h30m", "2 hours 30 minutes", "half an hour", etc.
- **Custom units**: Add your own custom time units
- **Flexible syntax**: Support for verbose and compact formats
- **Backwards compatible**: Falls back to standard `time.ParseDuration`

### 🌍 Environment Parser
- **Type-safe parsing**: Parse environment variables to specific types
- **Default values**: Built-in support for default values
- **Validation**: Required field validation with detailed error reporting
- **Prefix support**: Namespace environment variables with prefixes
- **Complex types**: Parse slices, maps, and custom types
- **Pointer methods**: Optional value parsing with nil returns

## Installation

```bash
go get github.com/fsvxavier/nexs-lib/parsers
```

## 🚀 Quick Start

### Legacy API - Drop-in Replacement

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
    // EXACTLY the same API as old dateparse libraries
    
    // ParseAny - most common legacy function
    date, err := datetime.ParseAny("02/03/2023", datetime.PreferMonthFirst(false))
    if err == nil {
        fmt.Println("European format:", date.Format("2006-01-02")) // 2023-03-02
    }
    
    // ParseIn - with timezone
    loc, _ := time.LoadLocation("America/New_York")
    date, err = datetime.ParseIn("2023-01-15 10:30:00", loc)
    
    // ParseLocal - with local timezone  
    date, err = datetime.ParseLocal("15/01/2023")
    
    // MustParseAny - panic on error (for testing)
    date = datetime.MustParseAny("2023-01-15T10:30:45Z")
    
    // ParseStrict - strict mode
    date, err = datetime.ParseStrict("2023-01-15", datetime.PreferMonthFirst(true))
}
```

### Modern API - New Features

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/fsvxavier/nexs-lib/parsers/datetime"
)

func main() {
    ctx := context.Background()
    parser := datetime.NewParser()
    
    // Parse various formats with enhanced capabilities
    formats := []string{
        "2023-01-15T10:30:45Z",
        "January 15, 2023 10:30 AM",
        "15th of March 2023",
        "15/01/2023",
        "01/15/2023",
        "1673778645",        // Unix timestamp
        "1673778645.123",    // Unix timestamp with decimals
        "today",
        "yesterday",
        "2 days ago",
        "next Monday",
    }
    
    for _, format := range formats {
        if date, err := parser.Parse(ctx, format); err == nil {
            fmt.Printf("%-25s -> %s\n", format, date.Format(time.RFC3339))
        }
    }
    
    // Format detection - NEW feature
    format, err := datetime.ParseFormat("January 15, 2023 10:30 AM")
    if err == nil {
        fmt.Printf("Detected format: %s\n", format) // "January 2, 2006 15:04 PM"
    }
    
    // Parse with specific format - NEW feature
    date, err := datetime.ParseWithFormat("15/01/2023", "02/01/2006")
    if err == nil {
        fmt.Printf("Specific format result: %s\n", date.Format("2006-01-02"))
    }
}
}
```

### Duration Parsing

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/parsers/duration"
)

func main() {
    // Parse various duration formats
    formats := []string{
        "1h30m45s",      // Standard Go format
        "2d3h",          // Extended with days
        "1w2d",          // Weeks and days
        "2 hours 30 minutes", // Verbose format
        "half an hour",  // Natural language
    }
    
    for _, format := range formats {
        if d, err := duration.Parse(format); err == nil {
            fmt.Printf("%s -> %v\n", format, d)
        }
    }
}
```

### Environment Variable Parsing

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/parsers/environment"
)

func main() {
    // Create parser with configuration
    env := environment.NewParser(
        environment.WithPrefix("APP"),
        environment.WithDefaults(map[string]string{
            "PORT": "8080",
            "HOST": "localhost",
            "DEBUG": "false",
            "TIMEOUT": "30s",
        }),
        environment.WithRequired("DATABASE_URL"),
    )
    
    // Parse various types
    port := env.GetInt("PORT")                    // APP_PORT or default 8080
    host := env.GetString("HOST")                 // APP_HOST or default "localhost"
    debug := env.GetBool("DEBUG")                 // APP_DEBUG or default false
    timeout := env.GetDuration("TIMEOUT")         // APP_TIMEOUT or default 30s
    
    // Parse complex types
    tags := env.GetSlice("TAGS", ",")             // "tag1,tag2,tag3" -> []string{"tag1", "tag2", "tag3"}
    config := env.GetMap("CONFIG", ",", "=")      // "key1=val1,key2=val2" -> map[string]string
    
    // Optional values
    dbPool := env.GetIntPtr("DB_POOL_SIZE")       // Returns *int or nil
    
    // Validation
    if err := env.Validate(); err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    }
    
    fmt.Printf("Server: %s:%d (debug=%v, timeout=%v)\n", host, port, debug, timeout)
}
```

## 📚 Complete Examples

The `examples/` directory contains comprehensive, working examples:

### 📖 [Legacy Compatibility](examples/legacy_compatibility/)
Perfect drop-in replacement examples showing 100% API compatibility with old dateparse libraries.
**Run:** `go run parsers/examples/legacy_compatibility/main.go`

### 🚀 [Advanced Parsing](examples/advanced_parsing/)  
Advanced datetime parsing with format detection, Unix timestamps, and modern features.
**Run:** `go run parsers/examples/advanced_parsing/main.go`

### 🔄 [Migration Guide](examples/migration_guide/)
Step-by-step migration guide with before/after code, compatibility matrix, and performance comparisons.
**Run:** `go run parsers/examples/migration_guide/main.go`

### 🌐 [Web Application](examples/web_application/)
Real-world web server example showing practical usage in HTTP handlers, form parsing, and API endpoints.
**Run:** `go run parsers/examples/web_application/main.go`

**See [examples/README.md](examples/README.md) for detailed documentation.**

## 🔧 Advanced Configuration

### Legacy Options (100% Compatible)

```go
// All these options work exactly like the old libraries
date, err := datetime.ParseAny("02/03/2023",
    datetime.PreferMonthFirst(false),              // European format preference  
    datetime.RetryAmbiguousDateWithSwap(true))     // Auto-retry ambiguous dates

// Parse with timezone
loc, _ := time.LoadLocation("Europe/London")
date, err := datetime.ParseIn("15/01/2023 14:30", loc)

// Strict parsing mode
date, err := datetime.ParseStrict("2023-01-15", datetime.PreferMonthFirst(true))
```

### Modern Configuration (New Features)

```go
parser := datetime.NewParser(
    parsers.WithLocation(time.UTC),
    parsers.WithCustomFormats("02/01/2006 15:04", "2006.01.02"),
    parsers.WithDateOrder(parsers.DateOrderDMY),
    parsers.WithStrictMode(false),
    parsers.WithCaching(true), // Performance optimization
)

// Parse with per-request options
date, err := parser.ParseWithOptions(ctx, "15/03/2023", 
    parsers.WithLocation(time.Local),
    parsers.WithStrictMode(true),
)
```
```

### Custom Duration Units

```go
customUnits := map[string]time.Duration{
    "fortnight": 14 * duration.Day,
    "jiffy":     10 * time.Millisecond,
}

parser := duration.NewParser(
    parsers.WithCustomUnits(customUnits),
)

d, err := parser.Parse(ctx, "2fortnight5jiffy") // 28 days + 50ms
```

### Environment Parser with Validation

```go
env := environment.NewParser(
    environment.WithPrefix("MYAPP"),
    environment.WithRequired("DATABASE_URL", "SECRET_KEY"),
    environment.WithDefaults(map[string]string{
        "LOG_LEVEL": "info",
        "PORT": "8080",
    }),
)

// Validate all required fields are present
if err := env.Validate(); err != nil {
    log.Fatal("Environment validation failed:", err)
}

// Use hierarchical prefixes
dbEnv := env.WithPrefix("DB")
dbHost := dbEnv.GetString("HOST")     // Reads MYAPP_DB_HOST
dbPort := dbEnv.GetInt("PORT", 5432)  // Reads MYAPP_DB_PORT with default
```

## Error Handling

The library provides detailed error information with suggestions:

```go
_, err := datetime.Parse("invalid-date")
if err != nil {
    var parseErr *parsers.ParseError
    if errors.As(err, &parseErr) {
        fmt.Printf("Type: %s\n", parseErr.Type)
        fmt.Printf("Input: %s\n", parseErr.Input)
        fmt.Printf("Message: %s\n", parseErr.Message)
        fmt.Printf("Suggestions: %v\n", parseErr.Suggestions)
    }
}
```

## Performance

The library is designed for high performance:

- **Caching**: Successful format patterns are cached for repeated parsing
- **Context support**: Proper cancellation prevents resource waste
- **Minimal allocations**: Optimized for reduced garbage collection pressure
- **Benchmarked**: Comprehensive benchmarks ensure performance regression detection

## Architecture

The library follows modern Go patterns:

- **Interfaces**: Well-defined interfaces for all parsers
- **Options pattern**: Flexible configuration via functional options
- **Context support**: Proper context handling throughout
- **Error types**: Structured error types with detailed information
- **Generics**: Type-safe APIs where appropriate
- **Testing**: Comprehensive test coverage with edge cases

## Compatibility

- **Go version**: Requires Go 1.18+ (for generics support)
- **Standard library**: Compatible with all standard time and environment functions
- **Dependencies**: Minimal external dependencies

## Examples

See the `examples/` directory for complete working examples:

- **Legacy compatibility**: Drop-in replacement examples for old dateparse libraries
- **Advanced parsing**: Complex datetime parsing with custom formats and options
- **Web application**: Real-world web server configuration parsing
- **CLI tools**: Command-line applications with environment parsing
- **Performance**: Benchmarking and performance optimization examples
- **Migration guide**: Step-by-step migration from old parsing libraries

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This library is part of the nexs-lib project and follows the same license terms.
