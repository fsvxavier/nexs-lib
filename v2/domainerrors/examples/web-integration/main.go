package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🌐 Domain Errors v2 - Web Integration Examples")
	fmt.Println("============================================")

	httpErrorHandlingExample()
	restAPIErrorsExample()
	middlewareErrorExample()
	clientErrorHandlingExample()
	graphQLErrorExample()
	webhookErrorExample()
	rateLimitErrorExample()
	corsErrorExample()
}

// httpErrorHandlingExample demonstrates HTTP error handling with proper status codes
func httpErrorHandlingExample() {
	fmt.Println("\n🔌 HTTP Error Handling Example:")

	handler := NewHTTPErrorHandler()

	// Test different error scenarios
	testCases := []struct {
		errorType    string
		domainError  interfaces.DomainErrorInterface
		expectedCode int
	}{
		{
			"Validation Error",
			factory.GetDefaultFactory().Builder().
				WithCode("INVALID_EMAIL").
				WithMessage("Email format is invalid").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("field", "email").
				WithDetail("value", "invalid-email").
				Build(),
			400,
		},
		{
			"Not Found Error",
			factory.GetDefaultFactory().Builder().
				WithCode("USER_NOT_FOUND").
				WithMessage("User with ID 123 not found").
				WithType(string(types.ErrorTypeNotFound)).
				WithSeverity(interfaces.Severity(types.SeverityLow)).
				WithDetail("user_id", "123").
				Build(),
			404,
		},
		{
			"Authentication Error",
			factory.GetDefaultFactory().Builder().
				WithCode("INVALID_TOKEN").
				WithMessage("Authentication token is invalid or expired").
				WithType(string(types.ErrorTypeAuthentication)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("token_type", "bearer").
				Build(),
			401,
		},
		{
			"Authorization Error",
			factory.GetDefaultFactory().Builder().
				WithCode("INSUFFICIENT_PERMISSIONS").
				WithMessage("User does not have permission to access this resource").
				WithType(string(types.ErrorTypeAuthorization)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("required_role", "admin").
				WithDetail("user_role", "user").
				Build(),
			403,
		},
		{
			"Rate Limit Error",
			factory.GetDefaultFactory().Builder().
				WithCode("RATE_LIMIT_EXCEEDED").
				WithMessage("Too many requests, please try again later").
				WithType(string(types.ErrorTypeRateLimit)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("limit", "100").
				WithDetail("window", "1h").
				WithDetail("retry_after", "3600").
				Build(),
			429,
		},
		{
			"Internal Server Error",
			factory.GetDefaultFactory().Builder().
				WithCode("DATABASE_CONNECTION_ERROR").
				WithMessage("Unable to connect to database").
				WithType(string(types.ErrorTypeDatabase)).
				WithSeverity(interfaces.Severity(types.SeverityCritical)).
				WithDetail("database", "postgresql").
				WithDetail("host", "db.cluster.local").
				Build(),
			500,
		},
	}

	fmt.Println("  HTTP Status Code Mapping:")
	for _, tc := range testCases {
		statusCode := handler.GetHTTPStatusCode(tc.domainError)
		response := handler.CreateErrorResponse(tc.domainError)

		fmt.Printf("    %s:\n", tc.errorType)
		fmt.Printf("      Status Code: %d (expected: %d) %s\n",
			statusCode, tc.expectedCode, getStatusIcon(statusCode == tc.expectedCode))
		fmt.Printf("      Response: %s\n", formatJSON(response))
	}
}

// restAPIErrorsExample demonstrates REST API error standards
func restAPIErrorsExample() {
	fmt.Println("\n🔗 REST API Error Standards Example:")

	api := NewRESTAPI()

	// Simulate various API endpoints
	endpoints := []struct {
		method   string
		path     string
		scenario string
		handler  func() *APIResponse
	}{
		{
			"GET", "/users/123", "User Not Found",
			func() *APIResponse {
				return api.GetUser("123")
			},
		},
		{
			"POST", "/users", "Validation Error",
			func() *APIResponse {
				return api.CreateUser(map[string]interface{}{
					"email": "invalid-email",
					"age":   -5,
				})
			},
		},
		{
			"PUT", "/users/456", "Conflict Error",
			func() *APIResponse {
				return api.UpdateUser("456", map[string]interface{}{
					"email": "existing@email.com",
				})
			},
		},
		{
			"DELETE", "/users/789", "Authorization Error",
			func() *APIResponse {
				return api.DeleteUser("789", "user")
			},
		},
		{
			"GET", "/users", "Database Error",
			func() *APIResponse {
				return api.ListUsers()
			},
		},
	}

	fmt.Println("  REST API Error Responses:")
	for _, endpoint := range endpoints {
		fmt.Printf("    %s %s (%s):\n", endpoint.method, endpoint.path, endpoint.scenario)
		response := endpoint.handler()

		fmt.Printf("      Status: %d\n", response.StatusCode)
		fmt.Printf("      Response: %s\n", formatJSON(response))
	}
}

// middlewareErrorExample demonstrates error handling in middleware
func middlewareErrorExample() {
	fmt.Println("\n🔧 Middleware Error Handling Example:")

	middleware := NewErrorMiddleware()

	// Simulate middleware chain
	middlewares := []Middleware{
		middleware.AuthMiddleware,
		middleware.RateLimitMiddleware,
		middleware.ValidationMiddleware,
		middleware.LoggingMiddleware,
	}

	// Test requests
	requests := []struct {
		name    string
		headers map[string]string
		body    string
	}{
		{
			"Valid Request",
			map[string]string{
				"Authorization": "Bearer valid-token",
				"Content-Type":  "application/json",
			},
			`{"name": "John Doe", "email": "john@example.com"}`,
		},
		{
			"Missing Authorization",
			map[string]string{
				"Content-Type": "application/json",
			},
			`{"name": "Jane Doe", "email": "jane@example.com"}`,
		},
		{
			"Rate Limited",
			map[string]string{
				"Authorization":  "Bearer valid-token",
				"Content-Type":   "application/json",
				"X-Rate-Limited": "true",
			},
			`{"name": "Bob Smith", "email": "bob@example.com"}`,
		},
		{
			"Invalid JSON",
			map[string]string{
				"Authorization": "Bearer valid-token",
				"Content-Type":  "application/json",
			},
			`{"name": "Invalid", "email":}`,
		},
	}

	fmt.Println("  Middleware Chain Processing:")
	for _, req := range requests {
		fmt.Printf("    %s:\n", req.name)

		request := &HTTPRequest{
			Headers: req.headers,
			Body:    req.body,
		}

		response := middleware.ProcessRequest(request, middlewares)

		if response.Error != nil {
			fmt.Printf("      ❌ Error at %s middleware: %s\n",
				response.FailedMiddleware, response.Error.Error())
			fmt.Printf("      Status Code: %d\n", response.StatusCode)
			fmt.Printf("      Error Details: %v\n", response.Error.Details())
		} else {
			fmt.Printf("      ✅ Request processed successfully\n")
			fmt.Printf("      Response: %s\n", response.Body)
		}
	}
}

// clientErrorHandlingExample demonstrates HTTP client error handling
func clientErrorHandlingExample() {
	fmt.Println("\n📱 HTTP Client Error Handling Example:")

	client := NewHTTPClient()

	// Test different client scenarios
	requests := []struct {
		name     string
		url      string
		method   string
		expected string
	}{
		{"Successful Request", "https://api.example.com/users", "GET", "success"},
		{"Timeout Error", "https://slow-api.example.com/data", "GET", "timeout"},
		{"Connection Error", "https://nonexistent.api.com/data", "GET", "connection"},
		{"Server Error", "https://api.example.com/error", "GET", "server_error"},
		{"Invalid Response", "https://api.example.com/invalid", "GET", "invalid_response"},
	}

	fmt.Println("  HTTP Client Error Scenarios:")
	for _, req := range requests {
		fmt.Printf("    %s (%s %s):\n", req.name, req.method, req.url)

		response, err := client.MakeRequest(req.method, req.url, nil)

		if err != nil {
			fmt.Printf("      ❌ Error: %s\n", err.Error())
			fmt.Printf("      Code: %s\n", err.Code())
			fmt.Printf("      Type: %s\n", err.Type())
			fmt.Printf("      Retryable: %v\n", err.IsRetryable())
			fmt.Printf("      Details: %v\n", err.Details())
		} else {
			fmt.Printf("      ✅ Success: %s\n", response)
		}
	}
}

// graphQLErrorExample demonstrates GraphQL error handling
func graphQLErrorExample() {
	fmt.Println("\n📊 GraphQL Error Handling Example:")

	gql := NewGraphQLHandler()

	// Test GraphQL queries
	queries := []struct {
		name  string
		query string
	}{
		{
			"Valid Query",
			`query GetUser($id: ID!) { user(id: $id) { name email } }`,
		},
		{
			"Syntax Error",
			`query GetUser($id: ID!) { user(id: $id) { name email }`,
		},
		{
			"Field Error",
			`query GetUser($id: ID!) { user(id: $id) { nonexistentField } }`,
		},
		{
			"Validation Error",
			`query GetUser { user { name email } }`,
		},
		{
			"Execution Error",
			`query GetUser($id: ID!) { user(id: $id) { sensitiveData } }`,
		},
	}

	fmt.Println("  GraphQL Error Handling:")
	for _, q := range queries {
		fmt.Printf("    %s:\n", q.name)

		response := gql.ExecuteQuery(q.query, map[string]interface{}{"id": "123"})

		if response.Errors != nil && len(response.Errors) > 0 {
			fmt.Printf("      ❌ GraphQL Errors:\n")
			for i, err := range response.Errors {
				fmt.Printf("        %d. %s\n", i+1, err.Error())
				fmt.Printf("           Path: %v\n", err.Details()["path"])
				fmt.Printf("           Extensions: %v\n", err.Details()["extensions"])
			}
		} else {
			fmt.Printf("      ✅ Query executed successfully\n")
			fmt.Printf("      Data: %v\n", response.Data)
		}
	}
}

// webhookErrorExample demonstrates webhook error handling
func webhookErrorExample() {
	fmt.Println("\n🪝 Webhook Error Handling Example:")

	webhook := NewWebhookHandler()

	// Test webhook scenarios
	webhooks := []struct {
		name     string
		url      string
		payload  map[string]interface{}
		expected string
	}{
		{
			"Successful Webhook",
			"https://api.partner.com/webhook",
			map[string]interface{}{"event": "user.created", "data": map[string]string{"id": "123"}},
			"success",
		},
		{
			"Webhook Timeout",
			"https://slow-webhook.com/endpoint",
			map[string]interface{}{"event": "order.completed", "data": map[string]string{"id": "456"}},
			"timeout",
		},
		{
			"Webhook Unauthorized",
			"https://secure-webhook.com/endpoint",
			map[string]interface{}{"event": "payment.failed", "data": map[string]string{"id": "789"}},
			"unauthorized",
		},
		{
			"Webhook Server Error",
			"https://unstable-webhook.com/endpoint",
			map[string]interface{}{"event": "user.deleted", "data": map[string]string{"id": "101"}},
			"server_error",
		},
	}

	fmt.Println("  Webhook Delivery Results:")
	for _, wh := range webhooks {
		fmt.Printf("    %s:\n", wh.name)

		result := webhook.DeliverWebhook(wh.url, wh.payload)

		if result.Error != nil {
			fmt.Printf("      ❌ Delivery failed: %s\n", result.Error.Error())
			fmt.Printf("      Attempts: %d\n", result.Attempts)
			fmt.Printf("      Next Retry: %v\n", result.NextRetry)
			fmt.Printf("      Will Retry: %v\n", result.Error.IsRetryable())
		} else {
			fmt.Printf("      ✅ Delivered successfully\n")
			fmt.Printf("      Response Code: %d\n", result.StatusCode)
			fmt.Printf("      Delivery Time: %v\n", result.DeliveryTime)
		}
	}
}

// rateLimitErrorExample demonstrates rate limiting error handling
func rateLimitErrorExample() {
	fmt.Println("\n⚡ Rate Limiting Error Example:")

	limiter := NewRateLimiter()

	// Configure rate limits
	limiter.SetLimit("api", 5, time.Minute)     // 5 requests per minute
	limiter.SetLimit("auth", 3, time.Minute*10) // 3 requests per 10 minutes
	limiter.SetLimit("upload", 1, time.Hour)    // 1 request per hour

	// Test rate limiting
	clients := []string{"client1", "client2", "client3"}
	endpoints := []string{"api", "auth", "upload"}

	fmt.Println("  Rate Limiting Test:")
	for _, client := range clients {
		fmt.Printf("    Client: %s\n", client)

		for _, endpoint := range endpoints {
			// Make multiple requests to trigger rate limiting
			limit := limiter.GetLimit(endpoint)
			for i := 0; i < limit+2; i++ {
				allowed, err := limiter.AllowRequest(client, endpoint)

				if err != nil {
					fmt.Printf("      %s[%d]: ❌ Rate limited - %s\n", endpoint, i+1, err.Error())
					fmt.Printf("        Retry After: %v\n", err.Details()["retry_after"])
					fmt.Printf("        Requests Left: %v\n", err.Details()["requests_left"])
					break
				} else if allowed {
					fmt.Printf("      %s[%d]: ✅ Request allowed\n", endpoint, i+1)
				}
			}
		}
	}

	// Show rate limiter statistics
	fmt.Println("\n  Rate Limiter Statistics:")
	stats := limiter.GetStatistics()
	for endpoint, stat := range stats {
		fmt.Printf("    %s:\n", endpoint)
		fmt.Printf("      Total Requests: %d\n", stat.TotalRequests)
		fmt.Printf("      Allowed Requests: %d\n", stat.AllowedRequests)
		fmt.Printf("      Blocked Requests: %d\n", stat.BlockedRequests)
		fmt.Printf("      Success Rate: %.2f%%\n", stat.SuccessRate*100)
	}
}

// corsErrorExample demonstrates CORS error handling
func corsErrorExample() {
	fmt.Println("\n🔗 CORS Error Handling Example:")

	cors := NewCORSHandler()

	// Configure CORS policy
	cors.SetPolicy(CORSPolicy{
		AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"X-Total-Count", "X-Page-Info"},
		AllowCredentials: true,
		MaxAge:           3600,
	})

	// Test CORS scenarios
	requests := []struct {
		name     string
		origin   string
		method   string
		headers  []string
		expected string
	}{
		{
			"Valid CORS Request",
			"https://app.example.com",
			"GET",
			[]string{"Content-Type"},
			"allowed",
		},
		{
			"Invalid Origin",
			"https://malicious.com",
			"GET",
			[]string{"Content-Type"},
			"blocked",
		},
		{
			"Invalid Method",
			"https://app.example.com",
			"PATCH",
			[]string{"Content-Type"},
			"blocked",
		},
		{
			"Invalid Headers",
			"https://app.example.com",
			"POST",
			[]string{"X-Custom-Header"},
			"blocked",
		},
		{
			"Preflight Request",
			"https://admin.example.com",
			"OPTIONS",
			[]string{"Authorization", "Content-Type"},
			"preflight",
		},
	}

	fmt.Println("  CORS Validation Results:")
	for _, req := range requests {
		fmt.Printf("    %s:\n", req.name)

		result := cors.ValidateRequest(CORSRequest{
			Origin:  req.origin,
			Method:  req.method,
			Headers: req.headers,
		})

		if result.Error != nil {
			fmt.Printf("      ❌ CORS Error: %s\n", result.Error.Error())
			fmt.Printf("      Code: %s\n", result.Error.Code())
			fmt.Printf("      Allowed Origins: %v\n", result.Error.Details()["allowed_origins"])
			fmt.Printf("      Allowed Methods: %v\n", result.Error.Details()["allowed_methods"])
		} else {
			fmt.Printf("      ✅ CORS validation passed\n")
			fmt.Printf("      Headers: %v\n", result.Headers)
		}
	}
}

// HTTP Error Handler Implementation
type HTTPErrorHandler struct {
	factory interfaces.ErrorFactory
}

func NewHTTPErrorHandler() *HTTPErrorHandler {
	return &HTTPErrorHandler{
		factory: factory.GetDefaultFactory(),
	}
}

func (h *HTTPErrorHandler) GetHTTPStatusCode(err interfaces.DomainErrorInterface) int {
	switch err.Type() {
	case string(types.ErrorTypeValidation):
		return 400
	case string(types.ErrorTypeAuthentication):
		return 401
	case string(types.ErrorTypeAuthorization):
		return 403
	case string(types.ErrorTypeNotFound):
		return 404
	case string(types.ErrorTypeConflict):
		return 409
	case string(types.ErrorTypeRateLimit):
		return 429
	case string(types.ErrorTypeExternalService):
		return 502
	case string(types.ErrorTypeTimeout):
		return 504
	case string(types.ErrorTypeDatabase):
		return 500
	default:
		return 500
	}
}

func (h *HTTPErrorHandler) CreateErrorResponse(err interfaces.DomainErrorInterface) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":      err.Code(),
			"message":   err.Error(),
			"type":      err.Type(),
			"severity":  err.Severity(),
			"details":   err.Details(),
			"timestamp": time.Now().Format(time.RFC3339),
			"trace_id":  generateTraceID(),
		},
	}
}

// REST API Implementation
type RESTAPI struct {
	factory interfaces.ErrorFactory
}

type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Data       interface{}            `json:"data,omitempty"`
	Error      interface{}            `json:"error,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

func NewRESTAPI() *RESTAPI {
	return &RESTAPI{
		factory: factory.GetDefaultFactory(),
	}
}

func (api *RESTAPI) GetUser(userID string) *APIResponse {
	if userID == "123" {
		return &APIResponse{
			StatusCode: 404,
			Error: map[string]interface{}{
				"code":    "USER_NOT_FOUND",
				"message": fmt.Sprintf("User with ID '%s' not found", userID),
				"type":    "not_found",
			},
		}
	}

	return &APIResponse{
		StatusCode: 200,
		Data: map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}
}

func (api *RESTAPI) CreateUser(data map[string]interface{}) *APIResponse {
	var errors []map[string]interface{}

	if email, ok := data["email"].(string); !ok || !strings.Contains(email, "@") {
		errors = append(errors, map[string]interface{}{
			"field":   "email",
			"code":    "INVALID_EMAIL",
			"message": "Email format is invalid",
		})
	}

	if age, ok := data["age"].(int); ok && age < 0 {
		errors = append(errors, map[string]interface{}{
			"field":   "age",
			"code":    "INVALID_AGE",
			"message": "Age must be a positive number",
		})
	}

	if len(errors) > 0 {
		return &APIResponse{
			StatusCode: 400,
			Error: map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": "Request validation failed",
				"type":    "validation",
				"errors":  errors,
			},
		}
	}

	return &APIResponse{
		StatusCode: 201,
		Data: map[string]interface{}{
			"id":      "new-user-id",
			"message": "User created successfully",
		},
	}
}

func (api *RESTAPI) UpdateUser(userID string, data map[string]interface{}) *APIResponse {
	if email, ok := data["email"].(string); ok && email == "existing@email.com" {
		return &APIResponse{
			StatusCode: 409,
			Error: map[string]interface{}{
				"code":    "EMAIL_ALREADY_EXISTS",
				"message": "A user with this email already exists",
				"type":    "conflict",
				"details": map[string]interface{}{
					"conflicting_field": "email",
					"existing_user_id":  "existing-user-123",
				},
			},
		}
	}

	return &APIResponse{
		StatusCode: 200,
		Data: map[string]interface{}{
			"id":      userID,
			"message": "User updated successfully",
		},
	}
}

func (api *RESTAPI) DeleteUser(userID, userRole string) *APIResponse {
	if userRole != "admin" {
		return &APIResponse{
			StatusCode: 403,
			Error: map[string]interface{}{
				"code":    "INSUFFICIENT_PERMISSIONS",
				"message": "Admin role required to delete users",
				"type":    "authorization",
				"details": map[string]interface{}{
					"required_role": "admin",
					"current_role":  userRole,
					"resource":      fmt.Sprintf("user:%s", userID),
				},
			},
		}
	}

	return &APIResponse{
		StatusCode: 204,
		Data: map[string]interface{}{
			"message": "User deleted successfully",
		},
	}
}

func (api *RESTAPI) ListUsers() *APIResponse {
	return &APIResponse{
		StatusCode: 500,
		Error: map[string]interface{}{
			"code":    "DATABASE_CONNECTION_ERROR",
			"message": "Unable to connect to database",
			"type":    "database",
			"details": map[string]interface{}{
				"database": "postgresql",
				"host":     "db.cluster.local",
				"error":    "connection timeout",
			},
		},
	}
}

// Middleware Implementation
type ErrorMiddleware struct {
	factory interfaces.ErrorFactory
}

type Middleware func(*HTTPRequest) interfaces.DomainErrorInterface

type HTTPRequest struct {
	Headers map[string]string
	Body    string
}

type MiddlewareResponse struct {
	Error            interfaces.DomainErrorInterface
	StatusCode       int
	Body             string
	FailedMiddleware string
}

func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{
		factory: factory.GetDefaultFactory(),
	}
}

func (m *ErrorMiddleware) AuthMiddleware(req *HTTPRequest) interfaces.DomainErrorInterface {
	auth, exists := req.Headers["Authorization"]
	if !exists {
		return m.factory.Builder().
			WithCode("MISSING_AUTHORIZATION").
			WithMessage("Authorization header is required").
			WithType(string(types.ErrorTypeAuthentication)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("header", "Authorization").
			Build()
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return m.factory.Builder().
			WithCode("INVALID_AUTH_FORMAT").
			WithMessage("Authorization header must use Bearer token format").
			WithType(string(types.ErrorTypeAuthentication)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("format", "Bearer <token>").
			Build()
	}

	return nil
}

func (m *ErrorMiddleware) RateLimitMiddleware(req *HTTPRequest) interfaces.DomainErrorInterface {
	if _, exists := req.Headers["X-Rate-Limited"]; exists {
		return m.factory.Builder().
			WithCode("RATE_LIMIT_EXCEEDED").
			WithMessage("Too many requests").
			WithType(string(types.ErrorTypeRateLimit)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("retry_after", "60").
			Build()
	}

	return nil
}

func (m *ErrorMiddleware) ValidationMiddleware(req *HTTPRequest) interfaces.DomainErrorInterface {
	if req.Body != "" {
		var data interface{}
		if err := json.Unmarshal([]byte(req.Body), &data); err != nil {
			return m.factory.Builder().
				WithCode("INVALID_JSON").
				WithMessage("Request body contains invalid JSON").
				WithType(string(types.ErrorTypeValidation)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("parse_error", err.Error()).
				Build()
		}
	}

	return nil
}

func (m *ErrorMiddleware) LoggingMiddleware(req *HTTPRequest) interfaces.DomainErrorInterface {
	// Logging middleware doesn't typically generate errors
	// This is just for demonstration
	return nil
}

func (m *ErrorMiddleware) ProcessRequest(req *HTTPRequest, middlewares []Middleware) *MiddlewareResponse {
	for i, middleware := range middlewares {
		if err := middleware(req); err != nil {
			middlewareNames := []string{"Auth", "RateLimit", "Validation", "Logging"}
			return &MiddlewareResponse{
				Error:            err,
				StatusCode:       getHTTPStatusFromError(err),
				FailedMiddleware: middlewareNames[i],
			}
		}
	}

	return &MiddlewareResponse{
		StatusCode: 200,
		Body:       `{"message": "Request processed successfully"}`,
	}
}

// HTTP Client Implementation
type HTTPClient struct {
	factory interfaces.ErrorFactory
	timeout time.Duration
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		factory: factory.GetDefaultFactory(),
		timeout: 5 * time.Second,
	}
}

func (c *HTTPClient) MakeRequest(method, url string, body io.Reader) (string, interfaces.DomainErrorInterface) {
	// Simulate different error scenarios based on URL
	switch {
	case strings.Contains(url, "slow-api"):
		time.Sleep(c.timeout + time.Second) // Simulate timeout
		return "", c.factory.Builder().
			WithCode("HTTP_TIMEOUT").
			WithMessage(fmt.Sprintf("Request to %s timed out", url)).
			WithType(string(types.ErrorTypeTimeout)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("url", url).
			WithDetail("timeout", c.timeout).
			WithTag("retryable").
			Build()

	case strings.Contains(url, "nonexistent"):
		return "", c.factory.Builder().
			WithCode("CONNECTION_ERROR").
			WithMessage(fmt.Sprintf("Failed to connect to %s", url)).
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("url", url).
			WithDetail("error", "no such host").
			WithTag("retryable").
			Build()

	case strings.Contains(url, "error"):
		return "", c.factory.Builder().
			WithCode("HTTP_SERVER_ERROR").
			WithMessage("Server returned 500 Internal Server Error").
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("url", url).
			WithDetail("status_code", 500).
			WithDetail("response_body", "Internal Server Error").
			WithTag("retryable").
			Build()

	case strings.Contains(url, "invalid"):
		return "", c.factory.Builder().
			WithCode("INVALID_RESPONSE").
			WithMessage("Server returned invalid response format").
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("url", url).
			WithDetail("content_type", "text/html").
			WithDetail("expected", "application/json").
			Build()

	default:
		return `{"message": "Request successful", "data": {"id": 123}}`, nil
	}
}

// GraphQL Handler Implementation
type GraphQLHandler struct {
	factory interfaces.ErrorFactory
}

type GraphQLResponse struct {
	Data   interface{}                       `json:"data,omitempty"`
	Errors []interfaces.DomainErrorInterface `json:"errors,omitempty"`
}

func NewGraphQLHandler() *GraphQLHandler {
	return &GraphQLHandler{
		factory: factory.GetDefaultFactory(),
	}
}

func (gql *GraphQLHandler) ExecuteQuery(query string, variables map[string]interface{}) *GraphQLResponse {
	// Simulate different GraphQL error scenarios
	switch {
	case strings.Contains(query, "nonexistentField"):
		return &GraphQLResponse{
			Errors: []interfaces.DomainErrorInterface{
				gql.factory.Builder().
					WithCode("FIELD_NOT_FOUND").
					WithMessage("Field 'nonexistentField' does not exist on type 'User'").
					WithType("graphql_field_error").
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("path", []string{"user", "nonexistentField"}).
					WithDetail("extensions", map[string]interface{}{
						"code": "FIELD_NOT_FOUND",
						"type": "User",
					}).
					Build(),
			},
		}

	case !strings.HasSuffix(query, "}"):
		return &GraphQLResponse{
			Errors: []interfaces.DomainErrorInterface{
				gql.factory.Builder().
					WithCode("SYNTAX_ERROR").
					WithMessage("Syntax error: Expected '}' but reached end of query").
					WithType("graphql_syntax_error").
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("path", []string{}).
					WithDetail("extensions", map[string]interface{}{
						"code":     "SYNTAX_ERROR",
						"line":     1,
						"column":   len(query),
						"expected": "}",
					}).
					Build(),
			},
		}

	case !strings.Contains(query, "$id") && strings.Contains(query, "user(id:"):
		return &GraphQLResponse{
			Errors: []interfaces.DomainErrorInterface{
				gql.factory.Builder().
					WithCode("VALIDATION_ERROR").
					WithMessage("Variable '$id' is required but not provided").
					WithType("graphql_validation_error").
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("path", []string{"user"}).
					WithDetail("extensions", map[string]interface{}{
						"code":     "VALIDATION_ERROR",
						"variable": "$id",
						"type":     "ID!",
					}).
					Build(),
			},
		}

	case strings.Contains(query, "sensitiveData"):
		return &GraphQLResponse{
			Errors: []interfaces.DomainErrorInterface{
				gql.factory.Builder().
					WithCode("AUTHORIZATION_ERROR").
					WithMessage("Access denied: insufficient permissions to access sensitiveData").
					WithType("graphql_auth_error").
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("path", []string{"user", "sensitiveData"}).
					WithDetail("extensions", map[string]interface{}{
						"code":            "AUTHORIZATION_ERROR",
						"required_role":   "admin",
						"current_role":    "user",
						"sensitive_field": "sensitiveData",
					}).
					Build(),
			},
		}

	default:
		return &GraphQLResponse{
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
		}
	}
}

// Webhook Handler Implementation
type WebhookHandler struct {
	factory interfaces.ErrorFactory
}

type WebhookResult struct {
	Error        interfaces.DomainErrorInterface
	StatusCode   int
	Attempts     int
	NextRetry    time.Time
	DeliveryTime time.Duration
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		factory: factory.GetDefaultFactory(),
	}
}

func (wh *WebhookHandler) DeliverWebhook(url string, payload map[string]interface{}) *WebhookResult {
	start := time.Now()

	// Simulate different webhook scenarios
	switch {
	case strings.Contains(url, "slow-webhook"):
		return &WebhookResult{
			Error: wh.factory.Builder().
				WithCode("WEBHOOK_TIMEOUT").
				WithMessage(fmt.Sprintf("Webhook delivery to %s timed out", url)).
				WithType(string(types.ErrorTypeTimeout)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("url", url).
				WithDetail("timeout", "30s").
				WithTag("retryable").
				Build(),
			Attempts:  1,
			NextRetry: time.Now().Add(5 * time.Minute),
		}

	case strings.Contains(url, "secure-webhook"):
		return &WebhookResult{
			Error: wh.factory.Builder().
				WithCode("WEBHOOK_UNAUTHORIZED").
				WithMessage("Webhook delivery failed: 401 Unauthorized").
				WithType(string(types.ErrorTypeAuthentication)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("url", url).
				WithDetail("status_code", 401).
				WithDetail("response", "Invalid webhook signature").
				Build(),
			Attempts: 3,
		}

	case strings.Contains(url, "unstable-webhook"):
		return &WebhookResult{
			Error: wh.factory.Builder().
				WithCode("WEBHOOK_SERVER_ERROR").
				WithMessage("Webhook delivery failed: 500 Internal Server Error").
				WithType(string(types.ErrorTypeExternalService)).
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("url", url).
				WithDetail("status_code", 500).
				WithTag("retryable").
				Build(),
			Attempts:  2,
			NextRetry: time.Now().Add(15 * time.Minute),
		}

	default:
		return &WebhookResult{
			StatusCode:   200,
			Attempts:     1,
			DeliveryTime: time.Since(start),
		}
	}
}

// Rate Limiter Implementation
type RateLimiter struct {
	limits     map[string]RateLimit
	counters   map[string]map[string]*RateCounter
	statistics map[string]*RateLimitStats
	factory    interfaces.ErrorFactory
	mu         sync.RWMutex
}

type RateLimit struct {
	Requests int
	Window   time.Duration
}

type RateCounter struct {
	Count     int
	ResetTime time.Time
}

type RateLimitStats struct {
	TotalRequests   int
	AllowedRequests int
	BlockedRequests int
	SuccessRate     float64
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits:     make(map[string]RateLimit),
		counters:   make(map[string]map[string]*RateCounter),
		statistics: make(map[string]*RateLimitStats),
		factory:    factory.GetDefaultFactory(),
	}
}

func (rl *RateLimiter) SetLimit(endpoint string, requests int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limits[endpoint] = RateLimit{
		Requests: requests,
		Window:   window,
	}
	rl.statistics[endpoint] = &RateLimitStats{}
}

func (rl *RateLimiter) GetLimit(endpoint string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if limit, exists := rl.limits[endpoint]; exists {
		return limit.Requests
	}
	return 0
}

func (rl *RateLimiter) AllowRequest(clientID, endpoint string) (bool, interfaces.DomainErrorInterface) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[endpoint]
	if !exists {
		return true, nil
	}

	// Initialize counters for client if not exists
	if rl.counters[clientID] == nil {
		rl.counters[clientID] = make(map[string]*RateCounter)
	}

	counter := rl.counters[clientID][endpoint]
	now := time.Now()

	// Initialize or reset counter if window expired
	if counter == nil || now.After(counter.ResetTime) {
		rl.counters[clientID][endpoint] = &RateCounter{
			Count:     0,
			ResetTime: now.Add(limit.Window),
		}
		counter = rl.counters[clientID][endpoint]
	}

	// Update statistics
	stats := rl.statistics[endpoint]
	stats.TotalRequests++

	// Check if limit exceeded
	if counter.Count >= limit.Requests {
		stats.BlockedRequests++
		stats.SuccessRate = float64(stats.AllowedRequests) / float64(stats.TotalRequests)

		retryAfter := counter.ResetTime.Sub(now)
		return false, rl.factory.Builder().
			WithCode("RATE_LIMIT_EXCEEDED").
			WithMessage(fmt.Sprintf("Rate limit exceeded for endpoint '%s'", endpoint)).
			WithType(string(types.ErrorTypeRateLimit)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("endpoint", endpoint).
			WithDetail("limit", limit.Requests).
			WithDetail("window", limit.Window).
			WithDetail("retry_after", retryAfter).
			WithDetail("requests_left", 0).
			Build()
	}

	counter.Count++
	stats.AllowedRequests++
	stats.SuccessRate = float64(stats.AllowedRequests) / float64(stats.TotalRequests)

	return true, nil
}

func (rl *RateLimiter) GetStatistics() map[string]*RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make(map[string]*RateLimitStats)
	for k, v := range rl.statistics {
		result[k] = &RateLimitStats{
			TotalRequests:   v.TotalRequests,
			AllowedRequests: v.AllowedRequests,
			BlockedRequests: v.BlockedRequests,
			SuccessRate:     v.SuccessRate,
		}
	}

	return result
}

// CORS Handler Implementation
type CORSHandler struct {
	policy  CORSPolicy
	factory interfaces.ErrorFactory
}

type CORSPolicy struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type CORSRequest struct {
	Origin  string
	Method  string
	Headers []string
}

type CORSResult struct {
	Error   interfaces.DomainErrorInterface
	Headers map[string]string
}

func NewCORSHandler() *CORSHandler {
	return &CORSHandler{
		factory: factory.GetDefaultFactory(),
	}
}

func (c *CORSHandler) SetPolicy(policy CORSPolicy) {
	c.policy = policy
}

func (c *CORSHandler) ValidateRequest(req CORSRequest) *CORSResult {
	// Check origin
	if !c.isOriginAllowed(req.Origin) {
		return &CORSResult{
			Error: c.factory.Builder().
				WithCode("CORS_ORIGIN_NOT_ALLOWED").
				WithMessage(fmt.Sprintf("Origin '%s' is not allowed by CORS policy", req.Origin)).
				WithType("cors_error").
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("origin", req.Origin).
				WithDetail("allowed_origins", c.policy.AllowedOrigins).
				Build(),
		}
	}

	// Check method
	if !c.isMethodAllowed(req.Method) {
		return &CORSResult{
			Error: c.factory.Builder().
				WithCode("CORS_METHOD_NOT_ALLOWED").
				WithMessage(fmt.Sprintf("Method '%s' is not allowed by CORS policy", req.Method)).
				WithType("cors_error").
				WithSeverity(interfaces.Severity(types.SeverityMedium)).
				WithDetail("method", req.Method).
				WithDetail("allowed_methods", c.policy.AllowedMethods).
				Build(),
		}
	}

	// Check headers
	for _, header := range req.Headers {
		if !c.isHeaderAllowed(header) {
			return &CORSResult{
				Error: c.factory.Builder().
					WithCode("CORS_HEADER_NOT_ALLOWED").
					WithMessage(fmt.Sprintf("Header '%s' is not allowed by CORS policy", header)).
					WithType("cors_error").
					WithSeverity(interfaces.Severity(types.SeverityMedium)).
					WithDetail("header", header).
					WithDetail("allowed_headers", c.policy.AllowedHeaders).
					Build(),
			}
		}
	}

	// Generate CORS headers
	headers := make(map[string]string)
	headers["Access-Control-Allow-Origin"] = req.Origin
	headers["Access-Control-Allow-Methods"] = strings.Join(c.policy.AllowedMethods, ", ")
	headers["Access-Control-Allow-Headers"] = strings.Join(c.policy.AllowedHeaders, ", ")

	if len(c.policy.ExposedHeaders) > 0 {
		headers["Access-Control-Expose-Headers"] = strings.Join(c.policy.ExposedHeaders, ", ")
	}

	if c.policy.AllowCredentials {
		headers["Access-Control-Allow-Credentials"] = "true"
	}

	if c.policy.MaxAge > 0 {
		headers["Access-Control-Max-Age"] = strconv.Itoa(c.policy.MaxAge)
	}

	return &CORSResult{
		Headers: headers,
	}
}

func (c *CORSHandler) isOriginAllowed(origin string) bool {
	for _, allowed := range c.policy.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (c *CORSHandler) isMethodAllowed(method string) bool {
	for _, allowed := range c.policy.AllowedMethods {
		if allowed == method {
			return true
		}
	}
	return false
}

func (c *CORSHandler) isHeaderAllowed(header string) bool {
	for _, allowed := range c.policy.AllowedHeaders {
		if strings.EqualFold(allowed, header) {
			return true
		}
	}
	return false
}

// Utility functions
func getStatusIcon(success bool) string {
	if success {
		return "✅"
	}
	return "❌"
}

func formatJSON(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}

func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func getHTTPStatusFromError(err interfaces.DomainErrorInterface) int {
	handler := NewHTTPErrorHandler()
	return handler.GetHTTPStatusCode(err)
}
