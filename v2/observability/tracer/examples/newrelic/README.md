# New Relic Provider Example

This example demonstrates how to use the New Relic provider for APM monitoring and distributed tracing with a comprehensive HTTP service.

## Prerequisites

1. **New Relic Account**: Active New Relic account with APM enabled
2. **License Key**: Your 40-character New Relic license key
3. **Go 1.21+**: Required for running the example

## Setup

### 1. Get Your New Relic License Key

1. Log into your New Relic account
2. Go to "API Keys" in account settings
3. Copy your License Key (40-character string)

### 2. Configure the Example

Replace the license key in `main.go`:
```go
config := &newrelic.Config{
    AppName:    "example-newrelic-service",
    LicenseKey: "your-40-character-license-key-here", // Replace this!
    // ... other config
}
```

Or set it via environment variable:
```bash
export NEW_RELIC_LICENSE_KEY=your-40-character-license-key
```

### 3. Run the Example

```bash
cd examples/newrelic
go run main.go
```

The server will start on `http://localhost:8080`.

## Testing the Endpoints

### Home Page
```bash
curl http://localhost:8080/
```
Creates a web transaction with custom attributes and page view events.

### Orders Endpoint
```bash
# Without authentication (will fail)
curl http://localhost:8080/orders

# With authentication (will succeed)
curl -H "Authorization: Bearer token123" http://localhost:8080/orders
```
Demonstrates business transaction tracking with user context and revenue metrics.

### Metrics Dashboard
```bash
curl http://localhost:8080/metrics
```
Shows custom metrics collection and business intelligence data.

### Error Simulation
```bash
# Generic error
curl http://localhost:8080/error

# Specific error types
curl "http://localhost:8080/error?type=validation"
curl "http://localhost:8080/error?type=database"
curl "http://localhost:8080/error?type=external"
curl "http://localhost:8080/error?type=timeout"
```
Demonstrates comprehensive error tracking and classification.

## What You'll See in New Relic

### APM Overview
- **Application Name**: `example-newrelic-service`
- **Environment**: `development`
- **Version**: `1.0.0`
- **Throughput**: Requests per minute
- **Response Time**: Average response times
- **Error Rate**: Percentage of failed requests

### Transactions
Each endpoint creates detailed transaction traces:

#### Web Transactions
- `web.transaction` - Home page requests
- `orders.list` - Order listing with business context
- `metrics.dashboard` - Metrics collection
- `error.simulation` - Error scenarios

#### Database Operations
- `db.orders.select` - Order queries with performance metrics
- Connection pool monitoring
- Query performance analysis

#### External Services
- `external.payment.status` - Payment gateway health checks
- Service dependency mapping
- Response time tracking

### Custom Attributes
All transactions include business context:
```json
{
  "transaction.type": "business",
  "user.id": "user_12345",
  "user.tier": "premium",
  "orders.count": 3,
  "orders.total_value": 299.97,
  "business.revenue_impact": 299.97,
  "user.authenticated": true
}
```

### Custom Events
Business intelligence events:
```json
{
  "eventType": "orders.retrieved",
  "user_id": "user_12345",
  "order_count": 3,
  "total_value": 299.97,
  "timestamp": 1678901234
}
```

### Error Tracking
Comprehensive error classification:
- **Validation Errors**: Input validation failures
- **Database Errors**: Connection and query issues
- **External Service Errors**: API dependency failures
- **Timeout Errors**: Performance-related issues

### Business Metrics
Custom metrics for business insights:
- `active_users`: Current active user count
- `orders_today`: Daily order volume
- `revenue_today`: Daily revenue tracking
- `avg_response_time`: Performance monitoring
- `error_rate`: Quality metrics

## Configuration Options

### Production Configuration
```go
config := &newrelic.Config{
    AppName:               "my-production-app",
    LicenseKey:            os.Getenv("NEW_RELIC_LICENSE_KEY"),
    Environment:           "production",
    DistributedTracer:     true,
    HighSecurity:          true,  // Enable high security mode
    CodeLevelMetrics:      true,
    LogLevel:              "warn",
    MaxSamplesStored:      20000,
    AttributesExclude: []string{
        "request.headers.authorization",
        "request.headers.cookie",
        "user.password",
        "payment.card_number",
    },
    Labels: map[string]string{
        "team":        "backend",
        "environment": "production",
        "datacenter":  "us-east-1",
    },
}
```

### Development Configuration
```go
config := &newrelic.Config{
    AppName:           "my-dev-app",
    LicenseKey:        "your-license-key",
    Environment:       "development",
    LogLevel:          "debug",
    DistributedTracer: true,
    CustomInsightsEvents: true,
}
```

## New Relic Features Demonstrated

### 1. Application Performance Monitoring (APM)
- Request throughput and response times
- Database query performance
- External service dependencies
- Error rate tracking

### 2. Distributed Tracing
- Cross-service request tracking
- Service dependency mapping
- Performance bottleneck identification

### 3. Custom Insights
- Business metrics collection
- Custom event tracking
- User behavior analytics

### 4. Infrastructure Monitoring
- Resource usage tracking (CPU, memory)
- Database connection monitoring
- Cache performance metrics

### 5. Error Analytics
- Error classification and grouping
- Error rate trends
- Impact analysis

## Advanced Features

### Custom Metrics
```go
span.SetAttribute("business.conversion_rate", 0.15)
span.SetAttribute("user.lifetime_value", 1250.00)
span.SetAttribute("feature.flag_active", true)
```

### Business Events
```go
span.AddEvent("purchase.completed", map[string]interface{}{
    "amount":      100.00,
    "currency":    "USD",
    "product_id":  "prod_123",
    "user_id":     "user_456",
})
```

### Error Context
```go
span.SetAttribute("error.class", "PaymentError")
span.SetAttribute("error.fingerprint", "payment_gateway_timeout")
span.SetAttribute("error.user_id", userID)
span.SetAttribute("error.transaction_amount", 99.99)
```

## Dashboards and Alerts

### Recommended Dashboards
1. **Business Metrics**
   - Revenue tracking
   - User engagement
   - Conversion rates

2. **Performance Monitoring**
   - Response time percentiles
   - Database performance
   - External service health

3. **Error Tracking**
   - Error rates by type
   - Error impact analysis
   - Recovery time tracking

### Alert Policies
```bash
# High error rate
Error rate > 5% for 5 minutes

# Slow response time
Response time > 2 seconds for 10 minutes

# Database issues
Database query time > 1 second for 5 minutes

# External service failures
External service error rate > 10% for 3 minutes
```

## Troubleshooting

### License Key Issues
- Verify the license key is exactly 40 characters
- Check that the key hasn't expired
- Ensure proper environment variable setting

### Missing Data
- Verify application name is unique in your account
- Check network connectivity to New Relic collectors
- Review log output for connection errors

### Performance Impact
- Adjust sampling rates for high-traffic applications
- Configure attribute exclusions for sensitive data
- Monitor agent overhead in production

## Integration Examples

### With Popular Frameworks

#### Gin Framework
```go
func setupGinWithNewRelic(tr tracer.Tracer) *gin.Engine {
    r := gin.Default()
    
    r.Use(func(c *gin.Context) {
        ctx, span := tr.StartSpan(c.Request.Context(), c.FullPath())
        defer span.End()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        
        span.SetAttribute("http.status_code", c.Writer.Status())
    })
    
    return r
}
```

#### Database Integration
```go
func queryWithTracing(ctx context.Context, tr tracer.Tracer, query string) error {
    _, span := tr.StartSpan(ctx, "db.query",
        tracer.WithSpanKind(tracer.SpanKindClient),
        tracer.WithSpanAttributes(map[string]interface{}{
            "db.statement": query,
        }),
    )
    defer span.End()
    
    // Execute query
    return nil
}
```

## Next Steps

1. **Custom Dashboards**: Create business-specific dashboards in New Relic
2. **Alert Setup**: Configure alerts for critical business metrics
3. **Integration**: Add to your existing applications and frameworks
4. **Business Intelligence**: Leverage custom events for analytics
5. **Performance Optimization**: Use insights to optimize application performance
