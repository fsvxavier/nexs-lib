# Datadog Provider Example

This example demonstrates how to use the Datadog provider for distributed tracing with a simple HTTP server.

## Prerequisites

1. **Datadog Agent**: You need a Datadog Agent running locally or specify a remote agent
2. **Datadog Account**: Valid Datadog account with API access
3. **Go 1.21+**: Required for running the example

## Setup

### 1. Start Datadog Agent (Local Development)

Using Docker:
```bash
docker run -d --name datadog-agent \
    -e DD_API_KEY=your_api_key_here \
    -e DD_SITE=datadoghq.com \
    -e DD_APM_ENABLED=true \
    -e DD_APM_NON_LOCAL_TRAFFIC=true \
    -p 8126:8126 \
    -p 8125:8125/udp \
    datadog/agent:latest
```

### 2. Configure Environment

Set your Datadog API key:
```bash
export DD_API_KEY=your_datadog_api_key
```

Or modify the config in `main.go`:
```go
config := &datadog.Config{
    ServiceName: "example-datadog-service",
    AgentHost:   "your-dd-agent-host",
    AgentPort:   8126,
    // ... other config
}
```

### 3. Run the Example

```bash
cd examples/datadog
go run main.go
```

The server will start on `http://localhost:8080`.

## Testing the Endpoints

### Basic Request
```bash
curl http://localhost:8080/
```
This creates a simple trace with basic span attributes.

### Users Endpoint (with Authentication)
```bash
# Without authentication (will fail)
curl http://localhost:8080/users

# With authentication (will succeed)
curl -H "Authorization: Bearer token123" http://localhost:8080/users
```
This demonstrates error handling and conditional tracing.

### Health Check
```bash
curl http://localhost:8080/health
```
This shows health check tracing with multiple component checks.

## What You'll See in Datadog

### Service Map
- Service: `example-datadog-service`
- Operations: `http.request`, `get_users`, `health_check`
- Dependencies: Database, Cache, External Service connections

### Traces
Each request creates a trace with:
- **Root Span**: HTTP request span with method, URL, status code
- **Child Spans**: Authentication, database queries, business logic
- **Attributes**: HTTP headers, user IDs, database info, error details
- **Events**: Request lifecycle events (started, completed)

### Metrics
- Request throughput
- Response times
- Error rates
- Database query performance

### Tags
All traces include:
- `env:development`
- `service:example-datadog-service`
- `version:1.0.0`
- `team:backend`
- `region:us-east-1`

## Features Demonstrated

### 1. Span Hierarchy
```
http.request
├── authenticate_user
├── process_request
│   ├── validate_input
│   └── business_logic
└── db.query
```

### 2. Error Handling
- Authentication failures
- Database timeouts
- Proper error attributes and status codes

### 3. Span Attributes
- HTTP method, URL, status codes
- Database operations and row counts
- User identification
- Performance metrics

### 4. Events
- Request started/completed timestamps
- Custom application events

### 5. Resource Management
- Proper provider shutdown
- Context propagation
- Timeout handling

## Configuration Options

### Production Settings
```go
config := &datadog.Config{
    ServiceName:        "my-production-service",
    Environment:        "production",
    SampleRate:         0.1,  // 10% sampling
    EnableProfiling:    true,
    RuntimeMetrics:     true,
    AnalyticsEnabled:   true,
    MaxTracesPerSecond: 1000,
    ObfuscationEnabled: true,
    ObfuscatedTags:     []string{"password", "token", "api_key"},
}
```

### Development Settings
```go
config := &datadog.Config{
    ServiceName:     "my-dev-service",
    Environment:     "development",
    SampleRate:      1.0,  // 100% sampling
    Debug:           true,
    RuntimeMetrics:  false,
}
```

## Troubleshooting

### Agent Connection Issues
- Verify agent is running: `curl http://localhost:8126/v0.4/traces` should return 404
- Check firewall settings
- Verify DD_APM_ENABLED=true on agent

### Missing Traces
- Check sampling rate (set to 1.0 for development)
- Verify service name configuration
- Check agent logs: `docker logs datadog-agent`

### Performance Issues
- Reduce sampling rate in production
- Adjust MaxTracesPerSecond
- Enable span filtering

## Advanced Usage

### Custom Metrics
```go
span.SetAttribute("custom.metric", 42)
span.SetAttribute("business.value", 1000.50)
```

### Custom Events
```go
span.AddEvent("payment.processed", map[string]interface{}{
    "amount":    100.00,
    "currency":  "USD",
    "method":    "credit_card",
})
```

### Error Tracking
```go
if err != nil {
    span.SetStatus(tracer.StatusCodeError, err.Error())
    span.SetAttribute("error", true)
    span.SetAttribute("error.type", "validation_error")
    span.SetAttribute("error.message", err.Error())
}
```

## Next Steps

1. **Integration**: Integrate with your existing HTTP framework (Gin, Echo, Fiber)
2. **Database**: Add database auto-instrumentation
3. **gRPC**: Add gRPC interceptors for microservices
4. **Custom Instrumentation**: Add business-specific spans and metrics
5. **Alerting**: Set up Datadog monitors and alerts based on trace data
