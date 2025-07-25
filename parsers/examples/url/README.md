# URL Parser Examples

This directory contains examples demonstrating the usage of the URL parser from the `parsers/url` module.

## Overview

The URL parser provides functionality to parse and manipulate URLs with advanced features like validation, component extraction, and query parameter handling.

## Features Demonstrated

### 1. Basic URL Parsing
- Parse standard URLs
- Extract URL components (scheme, host, path, etc.)
- Handle different URL formats

### 2. Query Parameter Handling
- Parse query parameters
- Extract specific parameters
- Handle multiple values for same parameter
- URL encoding/decoding

### 3. URL Validation
- Validate URL format
- Check for required components
- Domain validation
- Protocol validation

### 4. URL Manipulation
- Build URLs from components
- Modify existing URLs
- Add/remove query parameters
- Path manipulation

### 5. Advanced Features
- Handle internationalized domain names (IDN)
- Parse fragment identifiers
- Handle relative URLs
- URL normalization

## Files

### `basic_usage.go`
Comprehensive examples showing all URL parser features with practical use cases.

## Running the Examples

```bash
cd parsers/examples/url
go run basic_usage.go
```

## Key Concepts

### URL Components
- **Scheme**: Protocol (http, https, ftp, etc.)
- **Host**: Domain name or IP address
- **Port**: Network port number
- **Path**: Resource path
- **Query**: Query parameters
- **Fragment**: Fragment identifier

### Query Parameters
- Single values: `?name=value`
- Multiple values: `?name=value1&name=value2`
- Empty values: `?name=`
- No values: `?name`

### URL Encoding
- Special characters must be encoded
- Spaces become `%20` or `+`
- Reserved characters use percent encoding

## Common Use Cases

1. **Web API URLs**: Parse REST API endpoints
2. **Configuration URLs**: Database connection strings
3. **Redirect URLs**: Handle URL redirections
4. **Form Processing**: Parse form submission URLs
5. **Link Validation**: Validate user-provided URLs

## Best Practices

1. Always validate URLs before processing
2. Handle encoding/decoding properly
3. Use context for timeouts and cancellation
4. Check for required URL components
5. Normalize URLs for comparison
6. Handle edge cases (empty strings, malformed URLs)

## Error Handling

The URL parser provides detailed error messages for:
- Malformed URLs
- Invalid characters
- Missing required components
- Encoding/decoding errors

## Integration

This parser integrates well with:
- HTTP clients (setting request URLs)
- Web frameworks (routing and middleware)
- Configuration systems (connection strings)
- Validation systems (input validation)
