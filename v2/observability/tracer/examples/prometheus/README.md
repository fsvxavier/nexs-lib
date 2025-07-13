# Prometheus Provider Example

This example demonstrates how to use the Prometheus provider for metrics collection and monitoring with a comprehensive HTTP service that generates various types of metrics.

## Prerequisites

1. **Prometheus Server**: Prometheus server for scraping metrics
2. **Go 1.21+**: Required for running the example
3. **Grafana** (Optional): For advanced dashboards and visualization

## Setup

### 1. Run the Example

```bash
cd examples/prometheus
go run main.go
```

The server will start on `http://localhost:8080`.

### 2. Setup Prometheus Server

Create a `prometheus.yml` configuration file:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'nexs-lib-example'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

Run Prometheus:
```bash
# Using Docker
docker run -d --name=prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Or download and run locally
./prometheus --config.file=prometheus.yml
```

Access Prometheus UI at `http://localhost:9090`

### 3. Setup Grafana (Optional)

```bash
# Using Docker
docker run -d --name=grafana \
  -p 3000:3000 \
  grafana/grafana
```

Access Grafana at `http://localhost:3000` (admin/admin)

## Testing the Endpoints

### Home Page
```bash
curl http://localhost:8080/
```
Generates basic HTTP request metrics.

### Products API
```bash
# Without API key (will fail)
curl http://localhost:8080/api/products

# With API key (will succeed)
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/products
```
Demonstrates authentication metrics and database operation tracking.

### Inventory API
```bash
curl "http://localhost:8080/api/inventory?product_id=laptop"
curl "http://localhost:8080/api/inventory?product_id=mouse"
```
Shows complex business metrics with inventory levels and warehouse data.

### Slow Endpoint
```bash
curl http://localhost:8080/api/slow
```
Demonstrates performance metrics for slow operations and SLA monitoring.

### Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```
View the raw Prometheus metrics format.

## Generated Metrics

### HTTP Request Metrics
```prometheus
# Request count by endpoint and status
myapp_api_requests_total{method="GET", endpoint="home", status="200"}

# Request duration histograms
myapp_api_request_duration_seconds{endpoint="products", quantile="0.5"}
myapp_api_request_duration_seconds{endpoint="products", quantile="0.95"}
myapp_api_request_duration_seconds{endpoint="products", quantile="0.99"}

# Request size distribution
myapp_api_request_size_bytes{endpoint="products"}
myapp_api_response_size_bytes{endpoint="products"}
```

### Business Metrics
```prometheus
# Inventory levels
myapp_api_inventory_level{product_id="laptop"} 25
myapp_api_inventory_level{product_id="mouse"} 42

# Authentication metrics
myapp_api_auth_attempts_total{method="api_key", status="success"}
myapp_api_auth_attempts_total{method="api_key", status="failure"}

# Database operations
myapp_api_database_queries_total{operation="SELECT", table="products"}
myapp_api_database_query_duration_seconds{operation="SELECT", table="products"}
```

### Error Metrics
```prometheus
# Error counts by type
myapp_api_errors_total{endpoint="products", error_type="database_error"}
myapp_api_errors_total{endpoint="products", error_type="authentication_failed"}

# SLA violations
myapp_api_sla_violations_total{endpoint="slow", threshold="3s"}
```

### Performance Metrics
```prometheus
# Operation performance
myapp_api_operation_duration_seconds{operation="heavy_computation"}
myapp_api_operation_duration_seconds{operation="dataset_processing"}
myapp_api_operation_duration_seconds{operation="report_generation"}

# Resource usage
myapp_api_memory_usage_bytes{operation="dataset_processing"}
myapp_api_cpu_usage_percent{operation="heavy_computation"}
```

## What You'll See in Prometheus

### Queries to Try

#### Request Rate
```promql
# Requests per second
rate(myapp_api_requests_total[5m])

# Requests per second by endpoint
sum(rate(myapp_api_requests_total[5m])) by (endpoint)
```

#### Response Time
```promql
# 95th percentile response time
histogram_quantile(0.95, rate(myapp_api_request_duration_seconds_bucket[5m]))

# Average response time by endpoint
rate(myapp_api_request_duration_seconds_sum[5m]) / rate(myapp_api_request_duration_seconds_count[5m])
```

#### Error Rate
```promql
# Error rate percentage
sum(rate(myapp_api_errors_total[5m])) / sum(rate(myapp_api_requests_total[5m])) * 100

# Error rate by type
sum(rate(myapp_api_errors_total[5m])) by (error_type)
```

#### Business Metrics
```promql
# Current inventory levels
myapp_api_inventory_level

# Low stock alerts (inventory < 10)
myapp_api_inventory_level < 10

# Authentication success rate
sum(rate(myapp_api_auth_attempts_total{status="success"}[5m])) / sum(rate(myapp_api_auth_attempts_total[5m])) * 100
```

## Configuration Options

### Production Configuration
```go
config := &prometheus.Config{
    ServiceName:           "my-production-service",
    ServiceVersion:        "2.1.0",
    Environment:           "production",
    Namespace:             "mycompany",
    Subsystem:             "api",
    EnableDetailedMetrics: false, // Reduce cardinality in production
    CustomLabels: map[string]string{
        "datacenter": "us-east-1",
        "team":       "backend",
        "tier":       "production",
    },
    BucketBoundaries: []float64{
        0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0,
    },
    MaxCardinality:     500,  // Limit metric cardinality
    CollectionInterval: 60 * time.Second,
    BatchSize:          50,
}
```

### Development Configuration
```go
config := &prometheus.Config{
    ServiceName:           "my-dev-service",
    Environment:           "development",
    Namespace:             "dev",
    Subsystem:             "api",
    EnableDetailedMetrics: true,  // More detailed metrics for debugging
    CollectionInterval:    15 * time.Second,
}
```

## Grafana Dashboard Setup

### 1. Add Prometheus Data Source
1. Go to Configuration → Data Sources
2. Add Prometheus data source
3. URL: `http://localhost:9090`
4. Save & Test

### 2. Import Dashboard JSON
Create a dashboard with these panels:

#### Request Rate Panel
```json
{
  "title": "Request Rate",
  "targets": [{
    "expr": "sum(rate(myapp_api_requests_total[5m])) by (endpoint)",
    "legendFormat": "{{endpoint}}"
  }],
  "type": "graph"
}
```

#### Response Time Panel
```json
{
  "title": "Response Time",
  "targets": [{
    "expr": "histogram_quantile(0.95, rate(myapp_api_request_duration_seconds_bucket[5m]))",
    "legendFormat": "95th percentile"
  }],
  "type": "graph"
}
```

#### Error Rate Panel
```json
{
  "title": "Error Rate",
  "targets": [{
    "expr": "sum(rate(myapp_api_errors_total[5m])) / sum(rate(myapp_api_requests_total[5m])) * 100",
    "legendFormat": "Error Rate %"
  }],
  "type": "singlestat"
}
```

## Alerting Rules

Create `alerts.yml`:
```yaml
groups:
  - name: api_alerts
    rules:
      - alert: HighErrorRate
        expr: sum(rate(myapp_api_errors_total[5m])) / sum(rate(myapp_api_requests_total[5m])) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      - alert: SlowResponseTime
        expr: histogram_quantile(0.95, rate(myapp_api_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow response time detected"
          description: "95th percentile response time is {{ $value }}s"

      - alert: LowInventory
        expr: myapp_api_inventory_level < 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Low inventory alert"
          description: "Product {{ $labels.product_id }} has only {{ $value }} items in stock"

      - alert: SLAViolation
        expr: increase(myapp_api_sla_violations_total[1h]) > 5
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SLA violations detected"
          description: "{{ $value }} SLA violations in the last hour"
```

## Advanced Features

### Custom Histogram Buckets
```go
config := &prometheus.Config{
    BucketBoundaries: []float64{
        0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
    },
}
```

### Dynamic Labels
```go
span.SetAttribute("user_tier", "premium")
span.SetAttribute("feature_flag", "new_ui_enabled")
span.SetAttribute("ab_test_group", "variant_a")
```

### Business Metrics
```go
span.SetAttribute("revenue_impact", 99.99)
span.SetAttribute("conversion_event", true)
span.SetAttribute("user_action", "purchase")
```

## Monitoring Best Practices

### 1. Metric Naming
- Use consistent prefixes: `myapp_api_*`
- Include units in names: `_seconds`, `_bytes`, `_total`
- Use descriptive names: `request_duration` not `latency`

### 2. Label Management
- Keep cardinality low (< 1000 unique combinations)
- Avoid high-cardinality labels (user IDs, timestamps)
- Use meaningful label values

### 3. Alert Design
- **RED Method**: Rate, Errors, Duration
- **USE Method**: Utilization, Saturation, Errors
- Set appropriate thresholds and time windows

### 4. Performance Optimization
- Use histogram buckets appropriate for your use case
- Batch metric updates when possible
- Consider sampling for high-frequency events

## Integration Examples

### With HTTP Middleware
```go
func prometheusMiddleware(tr tracer.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        ctx, span := tr.StartSpan(c.Request.Context(), c.FullPath())
        defer func() {
            duration := time.Since(start)
            span.SetAttribute("duration_seconds", duration.Seconds())
            span.SetAttribute("http_status_code", c.Writer.Status())
            span.End()
        }()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

### With Database Monitoring
```go
func queryWithMetrics(ctx context.Context, tr tracer.Tracer, query string) error {
    _, span := tr.StartSpan(ctx, "database_query",
        tracer.WithSpanAttributes(map[string]interface{}{
            "db_operation": "SELECT",
            "db_table":     "users",
        }),
    )
    defer span.End()
    
    start := time.Now()
    // Execute query
    duration := time.Since(start)
    
    span.SetAttribute("query_duration_seconds", duration.Seconds())
    span.SetAttribute("rows_affected", 10)
    
    return nil
}
```

## Troubleshooting

### High Cardinality Issues
- Monitor metric cardinality: `/api/v1/label/__name__/values`
- Remove high-cardinality labels
- Use recording rules for complex queries

### Missing Metrics
- Check metric registration
- Verify label consistency
- Ensure proper span completion

### Performance Issues
- Reduce metric collection frequency
- Use sampling for high-volume metrics
- Optimize label usage

## Next Steps

1. **Custom Dashboards**: Create business-specific Grafana dashboards
2. **Alert Rules**: Implement comprehensive alerting based on SLOs
3. **Integration**: Add to your existing applications and services
4. **Recording Rules**: Create recording rules for complex queries
5. **Federation**: Set up Prometheus federation for multi-cluster monitoring
