package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("📊 Domain Errors v2 - Observability Examples")
	fmt.Println("============================================")

	structuredLoggingExample()
	metricsCollectionExample()
	distributedTracingExample()
	healthChecksExample()
	alertingExample()
	performanceMonitoringExample()
	errorAggregationExample()
	dashboardExample()
}

// structuredLoggingExample demonstrates structured logging with error context
func structuredLoggingExample() {
	fmt.Println("\n📝 Structured Logging Example:")

	logger := NewStructuredLogger()

	// Configure different log levels and outputs
	logger.SetLevel(LogLevelInfo)
	logger.AddOutput("console", &ConsoleOutput{})
	logger.AddOutput("file", &FileOutput{Filename: "/tmp/app.log"})
	logger.AddOutput("elastic", &ElasticOutput{URL: "http://elasticsearch:9200"})

	// Test different error scenarios with structured logging
	testCases := []struct {
		name      string
		operation func() interfaces.DomainErrorInterface
		context   map[string]interface{}
	}{
		{
			"Database Connection Error",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("DB_CONNECTION_FAILED").
					WithMessage("Failed to connect to PostgreSQL database").
					WithType(string(types.ErrorTypeDatabase)).
					WithSeverity(interfaces.Severity(types.SeverityCritical)).
					WithDetail("database", "postgresql").
					WithDetail("host", "db.cluster.local").
					WithDetail("port", 5432).
					WithDetail("connection_timeout", "30s").
					WithTag("database").
					WithTag("critical").
					Build()
			},
			map[string]interface{}{
				"user_id":    "user123",
				"request_id": "req456",
				"service":    "user-service",
				"version":    "v1.2.3",
			},
		},
		{
			"API Rate Limit Exceeded",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("RATE_LIMIT_EXCEEDED").
					WithMessage("API rate limit exceeded for client").
					WithType(string(types.ErrorTypeRateLimit)).
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("client_id", "client789").
					WithDetail("limit", 1000).
					WithDetail("window", "1h").
					WithDetail("requests_made", 1001).
					WithTag("rate_limit").
					Build()
			},
			map[string]interface{}{
				"endpoint":   "/api/v1/users",
				"method":     "GET",
				"client_ip":  "192.168.1.100",
				"user_agent": "MyApp/1.0",
			},
		},
		{
			"External Service Timeout",
			func() interfaces.DomainErrorInterface {
				return factory.GetDefaultFactory().Builder().
					WithCode("EXTERNAL_SERVICE_TIMEOUT").
					WithMessage("Payment service request timed out").
					WithType(string(types.ErrorTypeTimeout)).
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("service_name", "payment-gateway").
					WithDetail("timeout", "10s").
					WithDetail("endpoint", "/api/v2/payments").
					WithDetail("retry_count", 3).
					WithTag("external_service").
					WithTag("timeout").
					Build()
			},
			map[string]interface{}{
				"payment_id":     "pay123",
				"amount":         "99.99",
				"currency":       "USD",
				"merchant_id":    "merchant456",
				"transaction_id": "tx789",
			},
		},
	}

	fmt.Println("  Structured Log Entries:")
	for _, tc := range testCases {
		fmt.Printf("    %s:\n", tc.name)

		err := tc.operation()
		logEntry := logger.LogError(err, tc.context)

		fmt.Printf("      Log Level: %s\n", logEntry.Level)
		fmt.Printf("      Timestamp: %s\n", logEntry.Timestamp)
		fmt.Printf("      Message: %s\n", logEntry.Message)
		fmt.Printf("      Fields: %s\n", formatJSON(logEntry.Fields))
		fmt.Printf("      Outputs: %v\n", logEntry.Outputs)
	}
}

// metricsCollectionExample demonstrates metrics collection for error tracking
func metricsCollectionExample() {
	fmt.Println("\n📈 Metrics Collection Example:")

	metrics := NewMetricsCollector()

	// Configure metric types
	metrics.RegisterCounter("errors_total", "Total number of errors", []string{"code", "type", "severity"})
	metrics.RegisterHistogram("error_duration", "Error processing duration", []string{"operation"})
	metrics.RegisterGauge("active_errors", "Currently active errors", []string{"service"})

	// Simulate error scenarios and collect metrics
	errorScenarios := []struct {
		service   string
		operation string
		errors    []interfaces.DomainErrorInterface
	}{
		{
			"user-service",
			"get_user",
			[]interfaces.DomainErrorInterface{
				createTestError("USER_NOT_FOUND", types.ErrorTypeNotFound, types.SeverityLow),
				createTestError("USER_NOT_FOUND", types.ErrorTypeNotFound, types.SeverityLow),
				createTestError("DB_CONNECTION_ERROR", types.ErrorTypeDatabase, types.SeverityCritical),
			},
		},
		{
			"payment-service",
			"process_payment",
			[]interfaces.DomainErrorInterface{
				createTestError("INSUFFICIENT_FUNDS", types.ErrorTypeBusinessRule, types.SeverityMedium),
				createTestError("PAYMENT_GATEWAY_ERROR", types.ErrorTypeExternalService, types.SeverityHigh),
				createTestError("INVALID_CARD", types.ErrorTypeValidation, types.SeverityMedium),
				createTestError("PAYMENT_GATEWAY_ERROR", types.ErrorTypeExternalService, types.SeverityHigh),
			},
		},
		{
			"notification-service",
			"send_email",
			[]interfaces.DomainErrorInterface{
				createTestError("SMTP_CONNECTION_ERROR", types.ErrorTypeExternalService, types.SeverityHigh),
				createTestError("INVALID_EMAIL", types.ErrorTypeValidation, types.SeverityLow),
			},
		},
	}

	fmt.Println("  Metrics Collection Results:")
	for _, scenario := range errorScenarios {
		fmt.Printf("    %s.%s:\n", scenario.service, scenario.operation)

		for _, err := range scenario.errors {
			start := time.Now()

			// Simulate error processing
			time.Sleep(time.Duration(10+len(err.Code())) * time.Millisecond)

			// Collect metrics
			metrics.IncrementCounter("errors_total", 1, map[string]string{
				"code":     err.Code(),
				"type":     err.Type(),
				"severity": err.Severity().String(),
			})

			metrics.RecordHistogram("error_duration", time.Since(start).Seconds(), map[string]string{
				"operation": scenario.operation,
			})

			metrics.SetGauge("active_errors", float64(len(scenario.errors)), map[string]string{
				"service": scenario.service,
			})
		}

		// Show metrics for this service
		serviceMetrics := metrics.GetServiceMetrics(scenario.service)
		fmt.Printf("      Total Errors: %d\n", serviceMetrics.TotalErrors)
		fmt.Printf("      Error Rate: %.2f errors/min\n", serviceMetrics.ErrorRate)
		fmt.Printf("      Average Duration: %v\n", serviceMetrics.AverageDuration)
		fmt.Printf("      Top Error Codes: %v\n", serviceMetrics.TopErrorCodes)
	}

	// Show global metrics summary
	fmt.Println("\n  Global Metrics Summary:")
	summary := metrics.GetGlobalSummary()
	fmt.Printf("    Total Errors Across All Services: %d\n", summary.TotalErrors)
	fmt.Printf("    Error Rate: %.2f errors/min\n", summary.ErrorRate)
	fmt.Printf("    Services with Errors: %d\n", summary.ServicesWithErrors)
	fmt.Printf("    Most Common Error Types: %v\n", summary.TopErrorTypes)
	fmt.Printf("    Critical Errors: %d\n", summary.CriticalErrors)
}

// distributedTracingExample demonstrates distributed tracing integration
func distributedTracingExample() {
	fmt.Println("\n🔍 Distributed Tracing Example:")

	tracer := NewDistributedTracer()

	// Configure tracing
	tracer.SetSamplingRate(1.0) // 100% for demo
	tracer.AddExporter("jaeger", &JaegerExporter{Endpoint: "http://jaeger:14268"})
	tracer.AddExporter("zipkin", &ZipkinExporter{Endpoint: "http://zipkin:9411"})

	// Simulate distributed request with errors
	traceContext := tracer.StartTrace("order_processing", map[string]interface{}{
		"user_id":    "user123",
		"order_id":   "order456",
		"total":      199.99,
		"currency":   "USD",
		"ip_address": "192.168.1.100",
	})

	fmt.Printf("  Trace ID: %s\n", traceContext.TraceID)
	fmt.Println("  Distributed Request Flow:")

	// Service 1: User Service - Success
	span1 := tracer.StartSpan(traceContext, "user-service", "validate_user", map[string]interface{}{
		"user_id": "user123",
	})

	time.Sleep(50 * time.Millisecond) // Simulate processing
	span1.Finish(nil)
	fmt.Printf("    ✅ user-service.validate_user: %v\n", span1.Duration)

	// Service 2: Inventory Service - Success
	span2 := tracer.StartSpan(traceContext, "inventory-service", "check_availability", map[string]interface{}{
		"product_ids": []string{"prod1", "prod2"},
		"quantities":  []int{2, 1},
	})

	time.Sleep(75 * time.Millisecond)
	span2.Finish(nil)
	fmt.Printf("    ✅ inventory-service.check_availability: %v\n", span2.Duration)

	// Service 3: Payment Service - Error
	span3 := tracer.StartSpan(traceContext, "payment-service", "process_payment", map[string]interface{}{
		"amount":      199.99,
		"card_number": "****1234",
		"gateway":     "stripe",
	})

	time.Sleep(100 * time.Millisecond)
	paymentError := factory.GetDefaultFactory().Builder().
		WithCode("PAYMENT_DECLINED").
		WithMessage("Credit card payment was declined").
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("decline_code", "insufficient_funds").
		WithDetail("gateway_response", "Your card was declined").
		WithDetail("trace_id", traceContext.TraceID).
		WithDetail("span_id", span3.SpanID).
		WithTag("payment").
		WithTag("declined").
		Build()

	span3.Finish(paymentError)
	fmt.Printf("    ❌ payment-service.process_payment: %s\n", paymentError.Error())

	// Show distributed trace summary
	traceSummary := tracer.GetTraceSummary(traceContext.TraceID)
	fmt.Printf("\n  Trace Summary:\n")
	fmt.Printf("    Trace ID: %s\n", traceSummary.TraceID)
	fmt.Printf("    Total Duration: %v\n", traceSummary.TotalDuration)
	fmt.Printf("    Services Involved: %d\n", traceSummary.ServiceCount)
	fmt.Printf("    Total Spans: %d\n", traceSummary.SpanCount)
	fmt.Printf("    Errors: %d\n", traceSummary.ErrorCount)
	fmt.Printf("    Success Rate: %.2f%%\n", traceSummary.SuccessRate*100)
	fmt.Printf("    Critical Path: %v\n", traceSummary.CriticalPath)

	tracer.FinishTrace(traceContext)
}

// healthChecksExample demonstrates health check integration with error monitoring
func healthChecksExample() {
	fmt.Println("\n🏥 Health Checks Example:")

	healthChecker := NewHealthChecker()

	// Register health checks for different components
	healthChecker.RegisterCheck("database", &DatabaseHealthCheck{
		Host:    "db.cluster.local",
		Port:    5432,
		Timeout: 5 * time.Second,
	})

	healthChecker.RegisterCheck("redis", &RedisHealthCheck{
		Host:    "redis.cluster.local",
		Port:    6379,
		Timeout: 3 * time.Second,
	})

	healthChecker.RegisterCheck("payment_gateway", &HTTPHealthCheck{
		URL:     "https://api.stripe.com/v1/charges",
		Timeout: 10 * time.Second,
		Headers: map[string]string{"Authorization": "Bearer sk_test_xxx"},
	})

	healthChecker.RegisterCheck("notification_service", &ServiceHealthCheck{
		ServiceName: "notification-service",
		Endpoint:    "http://notification:8080/health",
		Timeout:     5 * time.Second,
	})

	// Perform health checks
	fmt.Println("  Health Check Results:")

	results := healthChecker.CheckAll()
	for checkName, result := range results {
		fmt.Printf("    %s:\n", checkName)

		if result.Error != nil {
			fmt.Printf("      ❌ Status: UNHEALTHY\n")
			fmt.Printf("      Error: %s\n", result.Error.Error())
			fmt.Printf("      Code: %s\n", result.Error.Code())
			fmt.Printf("      Duration: %v\n", result.Duration)
			fmt.Printf("      Last Success: %v\n", result.LastSuccess)
		} else {
			fmt.Printf("      ✅ Status: HEALTHY\n")
			fmt.Printf("      Duration: %v\n", result.Duration)
			fmt.Printf("      Details: %v\n", result.Details)
		}
	}

	// Show overall health status
	overallHealth := healthChecker.GetOverallHealth()
	fmt.Printf("\n  Overall Health Status:\n")
	fmt.Printf("    Status: %s\n", overallHealth.Status)
	fmt.Printf("    Healthy Checks: %d/%d\n", overallHealth.HealthyCount, overallHealth.TotalChecks)
	fmt.Printf("    Success Rate: %.2f%%\n", overallHealth.SuccessRate*100)
	fmt.Printf("    Last Check: %v\n", overallHealth.LastCheck)

	if len(overallHealth.FailedChecks) > 0 {
		fmt.Printf("    Failed Checks: %v\n", overallHealth.FailedChecks)
	}
}

// alertingExample demonstrates alerting integration with error thresholds
func alertingExample() {
	fmt.Println("\n🚨 Alerting Example:")

	alertManager := NewAlertManager()

	// Configure alert rules
	alertManager.AddRule("high_error_rate", AlertRule{
		Name:        "High Error Rate",
		Description: "Error rate exceeded threshold",
		Condition:   "error_rate > 0.05", // 5% error rate
		Severity:    "warning",
		Duration:    "5m",
		Labels: map[string]string{
			"team":    "backend",
			"service": "{{ .service }}",
		},
		Annotations: map[string]string{
			"summary":     "High error rate detected in {{ .service }}",
			"description": "Error rate is {{ .error_rate }}% which exceeds the threshold of 5%",
		},
	})

	alertManager.AddRule("critical_errors", AlertRule{
		Name:        "Critical Errors",
		Description: "Critical errors detected",
		Condition:   "critical_errors > 0",
		Severity:    "critical",
		Duration:    "1m",
		Labels: map[string]string{
			"team":     "sre",
			"priority": "high",
		},
		Annotations: map[string]string{
			"summary":     "Critical errors detected",
			"description": "{{ .critical_errors }} critical errors in the last minute",
			"runbook":     "https://wiki.company.com/runbooks/critical-errors",
		},
	})

	alertManager.AddRule("service_down", AlertRule{
		Name:        "Service Down",
		Description: "Service is not responding",
		Condition:   "up == 0",
		Severity:    "critical",
		Duration:    "2m",
		Labels: map[string]string{
			"team":    "sre",
			"oncall":  "true",
			"service": "{{ .service }}",
		},
		Annotations: map[string]string{
			"summary":     "Service {{ .service }} is down",
			"description": "Service has been down for more than 2 minutes",
			"action":      "Check service logs and restart if necessary",
		},
	})

	// Simulate alert scenarios
	fmt.Println("  Alert Scenarios:")

	scenarios := []struct {
		name    string
		metrics map[string]float64
		context map[string]string
	}{
		{
			"High Error Rate in Payment Service",
			map[string]float64{
				"error_rate":      0.08, // 8%
				"critical_errors": 0,
				"up":              1,
			},
			map[string]string{
				"service": "payment-service",
			},
		},
		{
			"Critical Database Errors",
			map[string]float64{
				"error_rate":      0.02, // 2%
				"critical_errors": 3,
				"up":              1,
			},
			map[string]string{
				"service": "user-service",
			},
		},
		{
			"Notification Service Down",
			map[string]float64{
				"error_rate":      0,
				"critical_errors": 0,
				"up":              0,
			},
			map[string]string{
				"service": "notification-service",
			},
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("    %s:\n", scenario.name)

		alerts := alertManager.EvaluateRules(scenario.metrics, scenario.context)

		if len(alerts) == 0 {
			fmt.Printf("      ✅ No alerts triggered\n")
		} else {
			for _, alert := range alerts {
				fmt.Printf("      🚨 ALERT: %s\n", alert.Name)
				fmt.Printf("         Severity: %s\n", alert.Severity)
				fmt.Printf("         Summary: %s\n", alert.Annotations["summary"])
				fmt.Printf("         Labels: %v\n", alert.Labels)
				fmt.Printf("         State: %s\n", alert.State)
				fmt.Printf("         Started: %v\n", alert.StartedAt)
			}
		}
	}

	// Show alert summary
	fmt.Println("\n  Alert Manager Summary:")
	summary := alertManager.GetSummary()
	fmt.Printf("    Total Rules: %d\n", summary.TotalRules)
	fmt.Printf("    Active Alerts: %d\n", summary.ActiveAlerts)
	fmt.Printf("    Firing Alerts: %d\n", summary.FiringAlerts)
	fmt.Printf("    Pending Alerts: %d\n", summary.PendingAlerts)
	fmt.Printf("    Silenced Alerts: %d\n", summary.SilencedAlerts)
}

// performanceMonitoringExample demonstrates performance monitoring for error handling
func performanceMonitoringExample() {
	fmt.Println("\n⚡ Performance Monitoring Example:")

	perfMonitor := NewPerformanceMonitor()

	// Configure performance thresholds
	perfMonitor.SetThreshold("error_creation", 1*time.Millisecond)
	perfMonitor.SetThreshold("error_serialization", 5*time.Millisecond)
	perfMonitor.SetThreshold("error_logging", 10*time.Millisecond)
	perfMonitor.SetThreshold("error_transmission", 100*time.Millisecond)

	// Test performance scenarios
	scenarios := []struct {
		name      string
		operation func() (time.Duration, interfaces.DomainErrorInterface)
	}{
		{
			"Simple Error Creation",
			func() (time.Duration, interfaces.DomainErrorInterface) {
				start := time.Now()
				err := factory.GetDefaultFactory().Builder().
					WithCode("SIMPLE_ERROR").
					WithMessage("Simple error message").
					Build()
				return time.Since(start), err
			},
		},
		{
			"Complex Error with Details",
			func() (time.Duration, interfaces.DomainErrorInterface) {
				start := time.Now()
				err := factory.GetDefaultFactory().Builder().
					WithCode("COMPLEX_ERROR").
					WithMessage("Complex error with extensive metadata").
					WithType(string(types.ErrorTypeBusinessRule)).
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("user_id", "user123").
					WithDetail("transaction_id", "tx456").
					WithDetail("amount", 199.99).
					WithDetail("metadata", map[string]interface{}{
						"ip_address": "192.168.1.100",
						"user_agent": "MyApp/1.0",
						"session_id": "sess789",
					}).
					WithTag("transaction").
					WithTag("payment").
					WithTag("high_value").
					Build()
				return time.Since(start), err
			},
		},
		{
			"Error Chain Creation",
			func() (time.Duration, interfaces.DomainErrorInterface) {
				start := time.Now()

				// Create root error
				rootErr := factory.GetDefaultFactory().Builder().
					WithCode("DB_ERROR").
					WithMessage("Database connection failed").
					Build()

				// Wrap error
				wrappedErr := factory.GetDefaultFactory().Builder().
					WithCode("SERVICE_ERROR").
					WithMessage("Service operation failed").
					Build()

				// Chain errors (simplified for demo)
				_ = rootErr
				_ = wrappedErr

				return time.Since(start), wrappedErr
			},
		},
	}

	fmt.Println("  Performance Test Results:")
	for _, scenario := range scenarios {
		fmt.Printf("    %s:\n", scenario.name)

		// Run multiple iterations for accurate measurement
		var totalDuration time.Duration
		var lastError interfaces.DomainErrorInterface
		iterations := 1000

		for i := 0; i < iterations; i++ {
			duration, err := scenario.operation()
			totalDuration += duration
			lastError = err
		}

		avgDuration := totalDuration / time.Duration(iterations)
		threshold := perfMonitor.GetThreshold("error_creation")

		fmt.Printf("      Average Duration: %v\n", avgDuration)
		fmt.Printf("      Threshold: %v\n", threshold)
		fmt.Printf("      Performance: %s\n", getPerformanceIcon(avgDuration <= threshold))
		fmt.Printf("      Error Code: %s\n", lastError.Code())

		// Record performance metrics
		perfMonitor.RecordMetric("error_creation", avgDuration)
	}

	// Test serialization performance
	fmt.Println("\n  Serialization Performance:")
	testError := factory.GetDefaultFactory().Builder().
		WithCode("SERIALIZATION_TEST").
		WithMessage("Error for serialization testing").
		WithType(string(types.ErrorTypeValidation)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("large_data", generateLargeData()).
		Build()

	// JSON serialization
	start := time.Now()
	jsonData, _ := json.Marshal(testError)
	jsonDuration := time.Since(start)

	fmt.Printf("    JSON Serialization:\n")
	fmt.Printf("      Duration: %v\n", jsonDuration)
	fmt.Printf("      Size: %d bytes\n", len(jsonData))
	fmt.Printf("      Threshold: %v\n", perfMonitor.GetThreshold("error_serialization"))

	// Show performance summary
	fmt.Println("\n  Performance Summary:")
	summary := perfMonitor.GetSummary()
	fmt.Printf("    Total Metrics Recorded: %d\n", summary.TotalMetrics)
	fmt.Printf("    Average Error Creation: %v\n", summary.AverageErrorCreation)
	fmt.Printf("    95th Percentile: %v\n", summary.P95Duration)
	fmt.Printf("    99th Percentile: %v\n", summary.P99Duration)
	fmt.Printf("    Threshold Violations: %d\n", summary.ThresholdViolations)
}

// errorAggregationExample demonstrates error aggregation and analysis
func errorAggregationExample() {
	fmt.Println("\n📊 Error Aggregation Example:")

	aggregator := NewErrorAggregator()

	// Configure aggregation windows
	aggregator.SetWindow("1m", time.Minute)
	aggregator.SetWindow("5m", 5*time.Minute)
	aggregator.SetWindow("1h", time.Hour)

	// Simulate error stream
	fmt.Println("  Simulating Error Stream:")

	errorStream := []interfaces.DomainErrorInterface{
		createTestError("USER_NOT_FOUND", types.ErrorTypeNotFound, types.SeverityLow),
		createTestError("USER_NOT_FOUND", types.ErrorTypeNotFound, types.SeverityLow),
		createTestError("DB_CONNECTION_ERROR", types.ErrorTypeDatabase, types.SeverityCritical),
		createTestError("RATE_LIMIT_EXCEEDED", types.ErrorTypeRateLimit, types.SeverityMedium),
		createTestError("USER_NOT_FOUND", types.ErrorTypeNotFound, types.SeverityLow),
		createTestError("PAYMENT_FAILED", types.ErrorTypeExternalService, types.SeverityHigh),
		createTestError("INVALID_INPUT", types.ErrorTypeValidation, types.SeverityLow),
		createTestError("DB_CONNECTION_ERROR", types.ErrorTypeDatabase, types.SeverityCritical),
		createTestError("PAYMENT_FAILED", types.ErrorTypeExternalService, types.SeverityHigh),
		createTestError("TIMEOUT", types.ErrorTypeTimeout, types.SeverityMedium),
	}

	for i, err := range errorStream {
		fmt.Printf("    Error %d: %s (%s)\n", i+1, err.Code(), err.Type())
		aggregator.AddError(err)
		time.Sleep(100 * time.Millisecond) // Simulate time between errors
	}

	// Show aggregation results
	fmt.Println("\n  Aggregation Results:")

	windows := []string{"1m", "5m", "1h"}
	for _, window := range windows {
		fmt.Printf("    %s Window:\n", window)

		stats := aggregator.GetWindowStats(window)
		fmt.Printf("      Total Errors: %d\n", stats.TotalErrors)
		fmt.Printf("      Error Rate: %.2f errors/min\n", stats.ErrorRate)
		fmt.Printf("      Unique Error Codes: %d\n", stats.UniqueErrorCodes)
		fmt.Printf("      Most Common Code: %s (%d occurrences)\n",
			stats.MostCommonCode, stats.MostCommonCount)

		fmt.Printf("      Error Distribution:\n")
		for code, count := range stats.ErrorCodeDistribution {
			percentage := float64(count) / float64(stats.TotalErrors) * 100
			fmt.Printf("        %s: %d (%.1f%%)\n", code, count, percentage)
		}

		fmt.Printf("      Severity Distribution:\n")
		for severity, count := range stats.SeverityDistribution {
			percentage := float64(count) / float64(stats.TotalErrors) * 100
			fmt.Printf("        %s: %d (%.1f%%)\n", severity, count, percentage)
		}
	}

	// Show trending analysis
	fmt.Println("\n  Trending Analysis:")
	trends := aggregator.GetTrends()
	fmt.Printf("    Error Rate Trend: %s\n", trends.ErrorRateTrend)
	fmt.Printf("    Critical Error Trend: %s\n", trends.CriticalErrorTrend)
	fmt.Printf("    Top Growing Error: %s (+%.1f%%)\n", trends.TopGrowingError, trends.GrowthPercentage)
	fmt.Printf("    Top Declining Error: %s (-%.1f%%)\n", trends.TopDecliningError, trends.DeclinePercentage)
}

// dashboardExample demonstrates dashboard data preparation
func dashboardExample() {
	fmt.Println("\n📈 Dashboard Example:")

	dashboard := NewDashboardDataProvider()

	// Generate dashboard data for the last hour
	fmt.Println("  Generating Dashboard Data:")

	dashboardData := dashboard.GenerateDashboardData(time.Hour)

	// Overview metrics
	fmt.Printf("    Overview Metrics:\n")
	fmt.Printf("      Total Errors: %d\n", dashboardData.Overview.TotalErrors)
	fmt.Printf("      Error Rate: %.2f errors/min\n", dashboardData.Overview.ErrorRate)
	fmt.Printf("      Critical Errors: %d\n", dashboardData.Overview.CriticalErrors)
	fmt.Printf("      Services Affected: %d\n", dashboardData.Overview.ServicesAffected)
	fmt.Printf("      Mean Time to Recovery: %v\n", dashboardData.Overview.MTTR)

	// Service breakdown
	fmt.Printf("\n    Service Breakdown:\n")
	for _, service := range dashboardData.Services {
		fmt.Printf("      %s:\n", service.Name)
		fmt.Printf("        Errors: %d\n", service.ErrorCount)
		fmt.Printf("        Error Rate: %.2f%%\n", service.ErrorRate*100)
		fmt.Printf("        Status: %s\n", service.Status)
		fmt.Printf("        Last Error: %v\n", service.LastError)
	}

	// Error trends
	fmt.Printf("\n    Error Trends (Hourly):\n")
	for _, trend := range dashboardData.HourlyTrends {
		fmt.Printf("      %s: %d errors\n", trend.Hour, trend.ErrorCount)
	}

	// Top errors
	fmt.Printf("\n    Top Errors:\n")
	for i, topError := range dashboardData.TopErrors {
		fmt.Printf("      %d. %s: %d occurrences\n", i+1, topError.Code, topError.Count)
	}

	// SLA metrics
	fmt.Printf("\n    SLA Metrics:\n")
	fmt.Printf("      Availability: %.3f%%\n", dashboardData.SLA.Availability*100)
	fmt.Printf("      Error Budget Remaining: %.1f%%\n", dashboardData.SLA.ErrorBudgetRemaining*100)
	fmt.Printf("      Monthly Error Budget: %d\n", dashboardData.SLA.MonthlyErrorBudget)
	fmt.Printf("      Errors This Month: %d\n", dashboardData.SLA.ErrorsThisMonth)

	// Alerts summary
	fmt.Printf("\n    Active Alerts:\n")
	for _, alert := range dashboardData.ActiveAlerts {
		fmt.Printf("      🚨 %s: %s (Since: %v)\n", alert.Severity, alert.Summary, alert.StartedAt)
	}
}

// Implementation types and structs would go here...
// For brevity, I'll include key structs but not all implementation details

// Structured Logger Implementation
type StructuredLogger struct {
	level   LogLevel
	outputs map[string]LogOutput
	factory interfaces.ErrorFactory
	mu      sync.RWMutex
}

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
	Outputs   []string               `json:"outputs"`
}

type LogOutput interface {
	Write(entry *LogEntry) error
}

type ConsoleOutput struct{}
type FileOutput struct{ Filename string }
type ElasticOutput struct{ URL string }

func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		level:   LogLevelInfo,
		outputs: make(map[string]LogOutput),
		factory: factory.GetDefaultFactory(),
	}
}

func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *StructuredLogger) AddOutput(name string, output LogOutput) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.outputs[name] = output
}

func (l *StructuredLogger) LogError(err interfaces.DomainErrorInterface, context map[string]interface{}) *LogEntry {
	fields := make(map[string]interface{})

	// Add error fields
	fields["error_code"] = err.Code()
	fields["error_type"] = err.Type()
	fields["error_severity"] = err.Severity()
	fields["error_details"] = err.Details()

	// Add context fields
	for k, v := range context {
		fields[k] = v
	}

	entry := &LogEntry{
		Level:     "error",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   err.Error(),
		Fields:    fields,
		Outputs:   make([]string, 0, len(l.outputs)),
	}

	// Write to all outputs
	for name, output := range l.outputs {
		if err := output.Write(entry); err == nil {
			entry.Outputs = append(entry.Outputs, name)
		}
	}

	return entry
}

func (o *ConsoleOutput) Write(entry *LogEntry) error {
	// Console output implementation
	return nil
}

func (o *FileOutput) Write(entry *LogEntry) error {
	// File output implementation
	return nil
}

func (o *ElasticOutput) Write(entry *LogEntry) error {
	// Elasticsearch output implementation
	return nil
}

// Metrics Collector Implementation
type MetricsCollector struct {
	counters   map[string]*Counter
	histograms map[string]*Histogram
	gauges     map[string]*Gauge
	mu         sync.RWMutex
}

type Counter struct {
	Value  float64
	Labels map[string]string
}

type Histogram struct {
	Values []float64
	Labels map[string]string
}

type Gauge struct {
	Value  float64
	Labels map[string]string
}

type ServiceMetrics struct {
	TotalErrors     int
	ErrorRate       float64
	AverageDuration time.Duration
	TopErrorCodes   []string
}

type GlobalMetricsSummary struct {
	TotalErrors        int
	ErrorRate          float64
	ServicesWithErrors int
	TopErrorTypes      []string
	CriticalErrors     int
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]*Counter),
		histograms: make(map[string]*Histogram),
		gauges:     make(map[string]*Gauge),
	}
}

func (m *MetricsCollector) RegisterCounter(name, help string, labels []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Implementation
}

func (m *MetricsCollector) RegisterHistogram(name, help string, labels []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Implementation
}

func (m *MetricsCollector) RegisterGauge(name, help string, labels []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Implementation
}

func (m *MetricsCollector) IncrementCounter(name string, value float64, labels map[string]string) {
	// Implementation
}

func (m *MetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	// Implementation
}

func (m *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	// Implementation
}

func (m *MetricsCollector) GetServiceMetrics(serviceName string) *ServiceMetrics {
	// Implementation
	return &ServiceMetrics{
		TotalErrors:     10,
		ErrorRate:       2.5,
		AverageDuration: 150 * time.Millisecond,
		TopErrorCodes:   []string{"USER_NOT_FOUND", "DB_ERROR"},
	}
}

func (m *MetricsCollector) GetGlobalSummary() *GlobalMetricsSummary {
	// Implementation
	return &GlobalMetricsSummary{
		TotalErrors:        50,
		ErrorRate:          5.2,
		ServicesWithErrors: 3,
		TopErrorTypes:      []string{"not_found", "database", "validation"},
		CriticalErrors:     5,
	}
}

// Distributed Tracer Implementation
type DistributedTracer struct {
	samplingRate float64
	exporters    map[string]TraceExporter
	traces       map[string]*TraceContext
	mu           sync.RWMutex
}

type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Baggage   map[string]interface{}
	StartTime time.Time
}

type SpanContext struct {
	TraceID   string
	SpanID    string
	Service   string
	Operation string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Error     interfaces.DomainErrorInterface
	Tags      map[string]interface{}
}

type TraceSummary struct {
	TraceID       string
	TotalDuration time.Duration
	ServiceCount  int
	SpanCount     int
	ErrorCount    int
	SuccessRate   float64
	CriticalPath  []string
}

type TraceExporter interface {
	Export(span *SpanContext) error
}

type JaegerExporter struct{ Endpoint string }
type ZipkinExporter struct{ Endpoint string }

func NewDistributedTracer() *DistributedTracer {
	return &DistributedTracer{
		samplingRate: 0.1, // 10% by default
		exporters:    make(map[string]TraceExporter),
		traces:       make(map[string]*TraceContext),
	}
}

func (t *DistributedTracer) SetSamplingRate(rate float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samplingRate = rate
}

func (t *DistributedTracer) AddExporter(name string, exporter TraceExporter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.exporters[name] = exporter
}

func (t *DistributedTracer) StartTrace(operation string, baggage map[string]interface{}) *TraceContext {
	traceID := fmt.Sprintf("trace_%d", time.Now().UnixNano())

	context := &TraceContext{
		TraceID:   traceID,
		Baggage:   baggage,
		StartTime: time.Now(),
	}

	t.mu.Lock()
	t.traces[traceID] = context
	t.mu.Unlock()

	return context
}

func (t *DistributedTracer) StartSpan(traceCtx *TraceContext, service, operation string, tags map[string]interface{}) *SpanContext {
	span := &SpanContext{
		TraceID:   traceCtx.TraceID,
		SpanID:    fmt.Sprintf("span_%d", time.Now().UnixNano()),
		Service:   service,
		Operation: operation,
		StartTime: time.Now(),
		Tags:      tags,
	}

	return span
}

func (s *SpanContext) Finish(err interfaces.DomainErrorInterface) {
	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
	s.Error = err
}

func (t *DistributedTracer) GetTraceSummary(traceID string) *TraceSummary {
	// Implementation
	return &TraceSummary{
		TraceID:       traceID,
		TotalDuration: 225 * time.Millisecond,
		ServiceCount:  3,
		SpanCount:     3,
		ErrorCount:    1,
		SuccessRate:   0.67,
		CriticalPath:  []string{"user-service", "inventory-service", "payment-service"},
	}
}

func (t *DistributedTracer) FinishTrace(traceCtx *TraceContext) {
	// Implementation
}

func (e *JaegerExporter) Export(span *SpanContext) error {
	// Jaeger export implementation
	return nil
}

func (e *ZipkinExporter) Export(span *SpanContext) error {
	// Zipkin export implementation
	return nil
}

// Health Checker Implementation
type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

type HealthCheck interface {
	Check() *HealthResult
	Name() string
}

type HealthResult struct {
	Error       interfaces.DomainErrorInterface
	Duration    time.Duration
	LastSuccess time.Time
	Details     map[string]interface{}
}

type OverallHealth struct {
	Status       string
	HealthyCount int
	TotalChecks  int
	SuccessRate  float64
	LastCheck    time.Time
	FailedChecks []string
}

type DatabaseHealthCheck struct {
	Host    string
	Port    int
	Timeout time.Duration
}

type RedisHealthCheck struct {
	Host    string
	Port    int
	Timeout time.Duration
}

type HTTPHealthCheck struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
}

type ServiceHealthCheck struct {
	ServiceName string
	Endpoint    string
	Timeout     time.Duration
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

func (h *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

func (h *HealthChecker) CheckAll() map[string]*HealthResult {
	results := make(map[string]*HealthResult)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for name, check := range h.checks {
		results[name] = check.Check()
	}

	return results
}

func (h *HealthChecker) GetOverallHealth() *OverallHealth {
	// Implementation
	return &OverallHealth{
		Status:       "HEALTHY",
		HealthyCount: 3,
		TotalChecks:  4,
		SuccessRate:  0.75,
		LastCheck:    time.Now(),
		FailedChecks: []string{"payment_gateway"},
	}
}

func (d *DatabaseHealthCheck) Check() *HealthResult {
	start := time.Now()
	// Simulate database check
	time.Sleep(50 * time.Millisecond)

	return &HealthResult{
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"host":       d.Host,
			"port":       d.Port,
			"connection": "ok",
		},
	}
}

func (d *DatabaseHealthCheck) Name() string { return "database" }

func (r *RedisHealthCheck) Check() *HealthResult {
	start := time.Now()
	time.Sleep(30 * time.Millisecond)

	return &HealthResult{
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"host":   r.Host,
			"port":   r.Port,
			"memory": "512MB",
		},
	}
}

func (r *RedisHealthCheck) Name() string { return "redis" }

func (h *HTTPHealthCheck) Check() *HealthResult {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)

	// Simulate failed check
	return &HealthResult{
		Error: factory.GetDefaultFactory().Builder().
			WithCode("HTTP_HEALTH_CHECK_FAILED").
			WithMessage("Health check endpoint returned 503").
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("url", h.URL).
			WithDetail("status_code", 503).
			Build(),
		Duration:    time.Since(start),
		LastSuccess: time.Now().Add(-5 * time.Minute),
	}
}

func (h *HTTPHealthCheck) Name() string { return "http_service" }

func (s *ServiceHealthCheck) Check() *HealthResult {
	start := time.Now()
	time.Sleep(40 * time.Millisecond)

	return &HealthResult{
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"service":  s.ServiceName,
			"endpoint": s.Endpoint,
			"version":  "v1.2.3",
		},
	}
}

func (s *ServiceHealthCheck) Name() string { return "service" }

// Alert Manager Implementation
type AlertManager struct {
	rules  map[string]AlertRule
	alerts map[string]*Alert
	mu     sync.RWMutex
}

type AlertRule struct {
	Name        string
	Description string
	Condition   string
	Severity    string
	Duration    string
	Labels      map[string]string
	Annotations map[string]string
}

type Alert struct {
	Name        string
	Severity    string
	State       string
	StartedAt   time.Time
	Labels      map[string]string
	Annotations map[string]string
}

type AlertSummary struct {
	TotalRules     int
	ActiveAlerts   int
	FiringAlerts   int
	PendingAlerts  int
	SilencedAlerts int
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:  make(map[string]AlertRule),
		alerts: make(map[string]*Alert),
	}
}

func (a *AlertManager) AddRule(name string, rule AlertRule) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rules[name] = rule
}

func (a *AlertManager) EvaluateRules(metrics map[string]float64, context map[string]string) []*Alert {
	var alerts []*Alert

	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, rule := range a.rules {
		if a.evaluateCondition(rule.Condition, metrics) {
			alert := &Alert{
				Name:        rule.Name,
				Severity:    rule.Severity,
				State:       "firing",
				StartedAt:   time.Now(),
				Labels:      a.interpolateLabels(rule.Labels, context),
				Annotations: a.interpolateAnnotations(rule.Annotations, context, metrics),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

func (a *AlertManager) evaluateCondition(condition string, metrics map[string]float64) bool {
	// Simple condition evaluation - in real implementation would use a proper expression parser
	switch condition {
	case "error_rate > 0.05":
		return metrics["error_rate"] > 0.05
	case "critical_errors > 0":
		return metrics["critical_errors"] > 0
	case "up == 0":
		return metrics["up"] == 0
	default:
		return false
	}
}

func (a *AlertManager) interpolateLabels(labels map[string]string, context map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		if strings.Contains(v, "{{ .service }}") && context["service"] != "" {
			result[k] = strings.ReplaceAll(v, "{{ .service }}", context["service"])
		} else {
			result[k] = v
		}
	}
	return result
}

func (a *AlertManager) interpolateAnnotations(annotations map[string]string, context map[string]string, metrics map[string]float64) map[string]string {
	result := make(map[string]string)
	for k, v := range annotations {
		// Simple template interpolation
		if strings.Contains(v, "{{ .service }}") && context["service"] != "" {
			v = strings.ReplaceAll(v, "{{ .service }}", context["service"])
		}
		if strings.Contains(v, "{{ .error_rate }}") {
			v = strings.ReplaceAll(v, "{{ .error_rate }}", fmt.Sprintf("%.1f", metrics["error_rate"]*100))
		}
		if strings.Contains(v, "{{ .critical_errors }}") {
			v = strings.ReplaceAll(v, "{{ .critical_errors }}", fmt.Sprintf("%.0f", metrics["critical_errors"]))
		}
		result[k] = v
	}
	return result
}

func (a *AlertManager) GetSummary() *AlertSummary {
	return &AlertSummary{
		TotalRules:     len(a.rules),
		ActiveAlerts:   2,
		FiringAlerts:   1,
		PendingAlerts:  1,
		SilencedAlerts: 0,
	}
}

// Performance Monitor Implementation
type PerformanceMonitor struct {
	thresholds map[string]time.Duration
	metrics    map[string][]time.Duration
	mu         sync.RWMutex
}

type PerformanceSummary struct {
	TotalMetrics         int
	AverageErrorCreation time.Duration
	P95Duration          time.Duration
	P99Duration          time.Duration
	ThresholdViolations  int
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		thresholds: make(map[string]time.Duration),
		metrics:    make(map[string][]time.Duration),
	}
}

func (p *PerformanceMonitor) SetThreshold(operation string, threshold time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.thresholds[operation] = threshold
}

func (p *PerformanceMonitor) GetThreshold(operation string) time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.thresholds[operation]
}

func (p *PerformanceMonitor) RecordMetric(operation string, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[operation] = append(p.metrics[operation], duration)
}

func (p *PerformanceMonitor) GetSummary() *PerformanceSummary {
	return &PerformanceSummary{
		TotalMetrics:         100,
		AverageErrorCreation: 800 * time.Nanosecond,
		P95Duration:          1200 * time.Nanosecond,
		P99Duration:          1500 * time.Nanosecond,
		ThresholdViolations:  2,
	}
}

// Error Aggregator Implementation
type ErrorAggregator struct {
	windows map[string]time.Duration
	data    map[string]*WindowData
	mu      sync.RWMutex
}

type WindowData struct {
	Errors    []interfaces.DomainErrorInterface
	StartTime time.Time
}

type WindowStats struct {
	TotalErrors           int
	ErrorRate             float64
	UniqueErrorCodes      int
	MostCommonCode        string
	MostCommonCount       int
	ErrorCodeDistribution map[string]int
	SeverityDistribution  map[string]int
}

type TrendAnalysis struct {
	ErrorRateTrend     string
	CriticalErrorTrend string
	TopGrowingError    string
	GrowthPercentage   float64
	TopDecliningError  string
	DeclinePercentage  float64
}

func NewErrorAggregator() *ErrorAggregator {
	return &ErrorAggregator{
		windows: make(map[string]time.Duration),
		data:    make(map[string]*WindowData),
	}
}

func (e *ErrorAggregator) SetWindow(name string, duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.windows[name] = duration
}

func (e *ErrorAggregator) AddError(err interfaces.DomainErrorInterface) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for window := range e.windows {
		if e.data[window] == nil {
			e.data[window] = &WindowData{
				Errors:    make([]interfaces.DomainErrorInterface, 0),
				StartTime: time.Now(),
			}
		}
		e.data[window].Errors = append(e.data[window].Errors, err)
	}
}

func (e *ErrorAggregator) GetWindowStats(window string) *WindowStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data := e.data[window]
	if data == nil {
		return &WindowStats{}
	}

	stats := &WindowStats{
		TotalErrors:           len(data.Errors),
		ErrorCodeDistribution: make(map[string]int),
		SeverityDistribution:  make(map[string]int),
	}

	// Calculate distributions
	for _, err := range data.Errors {
		stats.ErrorCodeDistribution[err.Code()]++
		stats.SeverityDistribution[err.Severity().String()]++
	}

	// Find most common error
	maxCount := 0
	for code, count := range stats.ErrorCodeDistribution {
		if count > maxCount {
			maxCount = count
			stats.MostCommonCode = code
			stats.MostCommonCount = count
		}
	}

	stats.UniqueErrorCodes = len(stats.ErrorCodeDistribution)

	// Calculate error rate (errors per minute)
	duration := time.Since(data.StartTime)
	if duration > 0 {
		stats.ErrorRate = float64(stats.TotalErrors) / duration.Minutes()
	}

	return stats
}

func (e *ErrorAggregator) GetTrends() *TrendAnalysis {
	return &TrendAnalysis{
		ErrorRateTrend:     "increasing",
		CriticalErrorTrend: "stable",
		TopGrowingError:    "DB_CONNECTION_ERROR",
		GrowthPercentage:   25.5,
		TopDecliningError:  "USER_NOT_FOUND",
		DeclinePercentage:  10.2,
	}
}

// Dashboard Data Provider Implementation
type DashboardDataProvider struct {
	mu sync.RWMutex
}

type DashboardData struct {
	Overview     *OverviewMetrics
	Services     []*ServiceDashboard
	HourlyTrends []*HourlyTrend
	TopErrors    []*TopError
	SLA          *SLAMetrics
	ActiveAlerts []*DashboardAlert
}

type OverviewMetrics struct {
	TotalErrors      int
	ErrorRate        float64
	CriticalErrors   int
	ServicesAffected int
	MTTR             time.Duration
}

type ServiceDashboard struct {
	Name       string
	ErrorCount int
	ErrorRate  float64
	Status     string
	LastError  time.Time
}

type HourlyTrend struct {
	Hour       string
	ErrorCount int
}

type TopError struct {
	Code  string
	Count int
}

type SLAMetrics struct {
	Availability         float64
	ErrorBudgetRemaining float64
	MonthlyErrorBudget   int
	ErrorsThisMonth      int
}

type DashboardAlert struct {
	Severity  string
	Summary   string
	StartedAt time.Time
}

func NewDashboardDataProvider() *DashboardDataProvider {
	return &DashboardDataProvider{}
}

func (d *DashboardDataProvider) GenerateDashboardData(timeWindow time.Duration) *DashboardData {
	return &DashboardData{
		Overview: &OverviewMetrics{
			TotalErrors:      150,
			ErrorRate:        2.5,
			CriticalErrors:   8,
			ServicesAffected: 3,
			MTTR:             15 * time.Minute,
		},
		Services: []*ServiceDashboard{
			{
				Name:       "user-service",
				ErrorCount: 45,
				ErrorRate:  0.02,
				Status:     "healthy",
				LastError:  time.Now().Add(-10 * time.Minute),
			},
			{
				Name:       "payment-service",
				ErrorCount: 78,
				ErrorRate:  0.05,
				Status:     "degraded",
				LastError:  time.Now().Add(-2 * time.Minute),
			},
			{
				Name:       "notification-service",
				ErrorCount: 27,
				ErrorRate:  0.01,
				Status:     "healthy",
				LastError:  time.Now().Add(-30 * time.Minute),
			},
		},
		HourlyTrends: []*HourlyTrend{
			{"14:00", 15}, {"15:00", 22}, {"16:00", 31}, {"17:00", 45}, {"18:00", 37},
		},
		TopErrors: []*TopError{
			{"USER_NOT_FOUND", 35},
			{"PAYMENT_FAILED", 28},
			{"DB_CONNECTION_ERROR", 12},
			{"TIMEOUT", 18},
			{"VALIDATION_ERROR", 22},
		},
		SLA: &SLAMetrics{
			Availability:         0.9985,
			ErrorBudgetRemaining: 0.73,
			MonthlyErrorBudget:   1000,
			ErrorsThisMonth:      267,
		},
		ActiveAlerts: []*DashboardAlert{
			{
				Severity:  "warning",
				Summary:   "High error rate in payment-service",
				StartedAt: time.Now().Add(-15 * time.Minute),
			},
		},
	}
}

// Utility functions
func createTestError(code string, errorType types.ErrorType, severity types.ErrorSeverity) interfaces.DomainErrorInterface {
	return factory.GetDefaultFactory().Builder().
		WithCode(code).
		WithMessage(fmt.Sprintf("Test error: %s", code)).
		WithType(string(errorType)).
		WithSeverity(interfaces.Severity(severity)).
		Build()
}

func formatJSON(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}

func getPerformanceIcon(withinThreshold bool) string {
	if withinThreshold {
		return "✅ GOOD"
	}
	return "⚠️ SLOW"
}

func generateLargeData() map[string]interface{} {
	data := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		data[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d_with_some_long_content_to_make_it_larger", i)
	}
	return data
}
