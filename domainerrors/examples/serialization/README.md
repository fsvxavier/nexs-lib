# Serialization Example

This example demonstrates how to serialize and deserialize domain errors to and from various formats (JSON, XML) while preserving all error details and type information.

## Features

- **JSON Serialization**: Convert domain errors to JSON format
- **XML Serialization**: Convert domain errors to XML format
- **Error Reconstruction**: Recreate domain errors from serialized data
- **Type Preservation**: Maintain error type and all specific details
- **Error Collections**: Handle multiple errors in a single response
- **Compact Format**: Space-efficient serialization for high-volume scenarios

## Running the Example

```bash
cd /home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/domainerrors/examples/serialization
go run main.go
```

## Key Components

### 1. Error Envelope
```go
type ErrorEnvelope struct {
    Error     ErrorInfo `json:"error" xml:"error"`
    Timestamp string    `json:"timestamp" xml:"timestamp"`
    RequestID string    `json:"request_id,omitempty" xml:"request_id,omitempty"`
    TraceID   string    `json:"trace_id,omitempty" xml:"trace_id,omitempty"`
}
```

### 2. Error Information
```go
type ErrorInfo struct {
    Code       string                 `json:"code" xml:"code"`
    Type       string                 `json:"type" xml:"type"`
    Message    string                 `json:"message" xml:"message"`
    Details    map[string]interface{} `json:"details,omitempty" xml:"details,omitempty"`
    HTTPStatus int                    `json:"http_status" xml:"http_status"`
    Cause      string                 `json:"cause,omitempty" xml:"cause,omitempty"`
}
```

## Serialization Examples

### JSON Format
```json
{
  "error": {
    "code": "INVALID_EMAIL",
    "type": "validation",
    "message": "Invalid email format",
    "details": {
      "fields": {
        "email": "must be a valid email address",
        "format": "example@domain.com"
      }
    },
    "http_status": 400
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-001",
  "trace_id": "trace-001"
}
```

### XML Format
```xml
<ErrorEnvelope>
  <error>
    <code>EMAIL_ALREADY_EXISTS</code>
    <type>conflict</type>
    <message>Email address already exists</message>
    <details>
      <resource>user</resource>
      <conflict_reason>email already registered</conflict_reason>
    </details>
    <http_status>409</http_status>
  </error>
  <timestamp>2024-01-15T10:30:00Z</timestamp>
  <request_id>req-xml-001</request_id>
  <trace_id>trace-xml-001</trace_id>
</ErrorEnvelope>
```

## Error Type Mapping

The serialization process preserves all error-specific details:

### Validation Errors
- Fields with validation messages
- Field-specific error details

### Business Rule Errors
- Business rule codes
- Rule descriptions

### Not Found Errors
- Resource type and ID
- Context information

### Conflict Errors
- Resource information
- Conflict reason

### Timeout Errors
- Operation details
- Duration and timeout values

### Rate Limit Errors
- Current limit and remaining count
- Reset time and window information

### External Service Errors
- Service name and endpoint
- HTTP status code and response

### Database Errors
- Operation and table information
- SQL query details

### Authentication Errors
- Authentication scheme
- Token information (if applicable)

### Authorization Errors
- Required permissions
- Resource access information

### Server Errors
- Request ID and correlation ID
- Component information

## Usage Patterns

### 1. Single Error Serialization
```go
// Serialize
envelope := SerializeDomainError(err, "req-123", "trace-456")
jsonData, _ := json.Marshal(envelope)

// Deserialize
var envelope ErrorEnvelope
json.Unmarshal(jsonData, &envelope)
recreatedErr := DeserializeDomainError(&envelope)
```

### 2. Error Collection
```go
type ErrorCollection struct {
    Errors    []ErrorEnvelope `json:"errors"`
    Count     int             `json:"count"`
    Timestamp string          `json:"timestamp"`
}
```

### 3. Compact Format
```go
type CompactError struct {
    C string `json:"c"`           // Code
    M string `json:"m"`           // Message
    T string `json:"t"`           // Type
    S int    `json:"s"`           // Status
    D string `json:"d,omitempty"` // Details (JSON string)
}
```

## Benefits

### 1. Type Safety
- Full error type preservation
- Detailed information retention
- Proper error reconstruction

### 2. Interoperability
- JSON and XML support
- Standard HTTP status codes
- Cross-service communication

### 3. Debugging Support
- Request/trace ID tracking
- Timestamp information
- Detailed error context

### 4. Performance
- Compact format options
- Efficient serialization
- Minimal overhead

## Integration Scenarios

### 1. API Responses
- HTTP error responses
- Structured error information
- Client-friendly format

### 2. Logging Systems
- Structured error logging
- Searchable error data
- Correlation tracking

### 3. Message Queues
- Error message serialization
- Cross-service error propagation
- Retry mechanisms

### 4. Monitoring Systems
- Error metrics collection
- Alert rule configuration
- Performance tracking

## Best Practices

1. **Include Context**: Always include request ID and trace ID
2. **Preserve Details**: Maintain all error-specific information
3. **Standard Format**: Use consistent envelope structure
4. **Compact When Needed**: Use compact format for high-volume scenarios
5. **Validation**: Validate deserialized errors before use
6. **Error Handling**: Handle serialization/deserialization errors gracefully

This example demonstrates production-ready error serialization that maintains full fidelity across service boundaries while providing multiple format options for different use cases.
