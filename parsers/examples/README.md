# Examples

This directory contains comprehensive examples demonstrating the capabilities of the nexs-lib parsers, with a focus on datetime parsing and **100% legacy compatibility**.

## 📁 Example Categories

### 🔄 [Legacy Compatibility](legacy_compatibility/)
**Perfect drop-in replacement for old dateparse libraries**

Shows how to migrate from legacy dateparse libraries with zero code changes (except import statements). Demonstrates:

- ✅ 100% API compatibility (`ParseAny`, `ParseIn`, `ParseLocal`, `MustParseAny`)
- ✅ All legacy options (`PreferMonthFirst`, `RetryAmbiguousDateWithSwap`)
- ✅ Same function signatures and behavior
- ✅ Enhanced performance and error handling
- ✅ Real-world migration examples

**Run:** `go run legacy_compatibility/main.go`

### 🚀 [Advanced Parsing](advanced_parsing/)
**Modern features and advanced capabilities**

Demonstrates advanced features beyond basic compatibility:

- 🔍 Automatic format detection and caching
- ⚡ Performance optimization with intelligent caching
- 🌍 Enhanced Unix timestamp support (seconds, millis, micros, decimal)
- 🎯 Context support for cancellation and timeouts
- 📝 Advanced text parsing with natural language
- ⚙️ Custom parser configuration
- 🛠️ Comprehensive error handling with suggestions

**Run:** `go run advanced_parsing/main.go`

### 🔄 [Migration Guide](migration_guide/)
**Step-by-step migration from legacy libraries**

Complete migration guide with practical examples:

- 📋 Step-by-step migration checklist
- 📊 Compatibility matrix
- 🔍 Before/after code examples
- 🆕 New features overview
- 📈 Performance improvements
- ✅ Live demonstration of migrated code

**Run:** `go run migration_guide/main.go`

### 🌐 [Web Application](web_application/)
**Real-world web server implementation**

Production-ready example showing practical usage:

- 📅 Event management system
- 🎯 Flexible date input handling
- 🔍 Real-time format detection
- 🔄 Date range searching
- 📡 RESTful API endpoints
- 🛡️ Comprehensive error handling
- 📱 HTML forms with flexible date inputs

**Run:** `go run web_application/main.go`

## 🚀 Quick Start

### For Legacy Users (Migration)

If you're coming from an old dateparse library, start here:

```bash
# See exactly how to migrate your existing code
go run migration_guide/main.go

# Test the legacy compatibility
go run legacy_compatibility/main.go
```

### For New Users

If you're starting fresh, explore the modern capabilities:

```bash
# Learn advanced features
go run advanced_parsing/main.go

# See real-world usage
go run web_application/main.go
```

## 📋 What Each Example Shows

| Example | Legacy API | Modern API | Real-world Usage | Performance | Error Handling |
|---------|------------|------------|------------------|-------------|----------------|
| **Legacy Compatibility** | ✅✅✅ | ❌ | ⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Advanced Parsing** | ✅ | ✅✅✅ | ⭐⭐ | ✅✅✅ | ✅✅✅ |
| **Migration Guide** | ✅✅✅ | ⭐ | ⭐ | ⭐⭐ | ⭐ |
| **Web Application** | ✅ | ✅✅ | ✅✅✅ | ⭐⭐ | ✅✅ |

## 🔧 Common Use Cases

### 1. **Drop-in Replacement**
```bash
# Your existing code works immediately
go run legacy_compatibility/main.go
```

### 2. **Format Detection**
```go
format, err := datetime.ParseFormat("January 15, 2023 10:30 AM")
// Returns: "January 2, 2006 15:04 PM"
```

### 3. **Enhanced Unix Timestamps**
```go
// All these work automatically
datetime.ParseAny("1673778645")     // seconds
datetime.ParseAny("1673778645123")  // milliseconds  
datetime.ParseAny("1673778645.123") // decimal precision
```

### 4. **Smart Date Preferences**
```go
// European format preference with auto-retry
date, err := datetime.ParseAny("02/03/2023", 
    datetime.PreferMonthFirst(false),
    datetime.RetryAmbiguousDateWithSwap(true))
```

### 5. **Web Form Handling**
```go
// Handle user input from web forms
event, err := service.createEvent(title, description, 
    "15/03/2024 14:30",  // Flexible input
    "17/03/2024 17:00",  // Any format works
    location)
```

## 📊 Performance Highlights

- **🚀 Format Caching**: 2-10x faster on repeated parsing
- **🎯 Smart Ordering**: Most common formats tried first
- **💾 Memory Efficient**: Reduced allocations
- **⚡ Optimized**: Better than legacy libraries

## 🛠️ Error Handling

The library provides detailed error information:

```go
_, err := datetime.ParseAny("invalid-date")
if parseErr, ok := err.(*parsers.ParseError); ok {
    fmt.Printf("Type: %s\n", parseErr.Type)
    fmt.Printf("Message: %s\n", parseErr.Message) 
    fmt.Printf("Suggestions: %v\n", parseErr.Suggestions)
}
```

## 📚 Documentation

- **README.md**: Main library documentation
- **COMPARISON_REPORT.md**: Detailed compatibility analysis
- **IMPLEMENTATION_SUCCESS.md**: Implementation details
- **Each example/**: Individual README files with specific instructions

## 🤝 Contributing

Feel free to add more examples or improve existing ones:

1. Create a new directory under `examples/`
2. Add a descriptive `main.go` file
3. Include a README.md explaining the example
4. Add an entry to this main examples README

## 🎯 Next Steps

1. **Start with migration guide** if you're migrating
2. **Try legacy compatibility** to verify your use case
3. **Explore advanced parsing** for new features
4. **Check web application** for production patterns

All examples are self-contained and can be run independently!
