package interfaces

import (
	"testing"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected string
	}{
		{"Low", SeverityLow, "low"},
		{"Medium", SeverityMedium, "medium"},
		{"High", SeverityHigh, "high"},
		{"Critical", SeverityCritical, "critical"},
		{"Unknown", Severity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.severity.String()
			if result != tt.expected {
				t.Errorf("Severity.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		expected string
	}{
		{"Business", CategoryBusiness, "business"},
		{"Technical", CategoryTechnical, "technical"},
		{"Infrastructure", CategoryInfrastructure, "infrastructure"},
		{"Security", CategorySecurity, "security"},
		{"Performance", CategoryPerformance, "performance"},
		{"Integration", CategoryIntegration, "integration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.category)
			if result != tt.expected {
				t.Errorf("Category string = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsedError_Creation(t *testing.T) {
	parsedError := ParsedError{
		Code:      "TEST_001",
		Message:   "Test error message",
		Type:      "validation",
		Details:   map[string]interface{}{"field": "email"},
		Severity:  SeverityMedium,
		Category:  CategoryTechnical,
		Retryable: false,
		Temporary: false,
	}

	if parsedError.Code != "TEST_001" {
		t.Errorf("ParsedError.Code = %v, want %v", parsedError.Code, "TEST_001")
	}
	if parsedError.Message != "Test error message" {
		t.Errorf("ParsedError.Message = %v, want %v", parsedError.Message, "Test error message")
	}
	if parsedError.Type != "validation" {
		t.Errorf("ParsedError.Type = %v, want %v", parsedError.Type, "validation")
	}
	if parsedError.Severity != SeverityMedium {
		t.Errorf("ParsedError.Severity = %v, want %v", parsedError.Severity, SeverityMedium)
	}
	if parsedError.Category != CategoryTechnical {
		t.Errorf("ParsedError.Category = %v, want %v", parsedError.Category, CategoryTechnical)
	}
	if parsedError.Retryable != false {
		t.Errorf("ParsedError.Retryable = %v, want %v", parsedError.Retryable, false)
	}
	if parsedError.Temporary != false {
		t.Errorf("ParsedError.Temporary = %v, want %v", parsedError.Temporary, false)
	}
}

func TestErrorCodeInfo_Creation(t *testing.T) {
	errorInfo := ErrorCodeInfo{
		Code:        "ERR_001",
		Message:     "Test error",
		Type:        "validation",
		StatusCode:  400,
		Severity:    SeverityLow,
		Retryable:   false,
		Temporary:   false,
		Tags:        []string{"test", "validation"},
		Description: "Test error description",
		Examples:    []string{"Invalid email format"},
	}

	if errorInfo.Code != "ERR_001" {
		t.Errorf("ErrorCodeInfo.Code = %v, want %v", errorInfo.Code, "ERR_001")
	}
	if errorInfo.Message != "Test error" {
		t.Errorf("ErrorCodeInfo.Message = %v, want %v", errorInfo.Message, "Test error")
	}
	if errorInfo.Type != "validation" {
		t.Errorf("ErrorCodeInfo.Type = %v, want %v", errorInfo.Type, "validation")
	}
	if errorInfo.StatusCode != 400 {
		t.Errorf("ErrorCodeInfo.StatusCode = %v, want %v", errorInfo.StatusCode, 400)
	}
	if errorInfo.Severity != SeverityLow {
		t.Errorf("ErrorCodeInfo.Severity = %v, want %v", errorInfo.Severity, SeverityLow)
	}
	if len(errorInfo.Tags) != 2 {
		t.Errorf("ErrorCodeInfo.Tags length = %v, want %v", len(errorInfo.Tags), 2)
	}
	if len(errorInfo.Examples) != 1 {
		t.Errorf("ErrorCodeInfo.Examples length = %v, want %v", len(errorInfo.Examples), 1)
	}
}

func TestErrorResponse_Creation(t *testing.T) {
	response := ErrorResponse{
		Error:      "Test error",
		StatusCode: 400,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Metadata:   map[string]interface{}{"request_id": "123"},
	}

	if response.Error != "Test error" {
		t.Errorf("ErrorResponse.Error = %v, want %v", response.Error, "Test error")
	}
	if response.StatusCode != 400 {
		t.Errorf("ErrorResponse.StatusCode = %v, want %v", response.StatusCode, 400)
	}
	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("ErrorResponse.Headers[Content-Type] = %v, want %v", response.Headers["Content-Type"], "application/json")
	}
	if response.Metadata["request_id"] != "123" {
		t.Errorf("ErrorResponse.Metadata[request_id] = %v, want %v", response.Metadata["request_id"], "123")
	}
}

func TestHTTPErrorResponse_Creation(t *testing.T) {
	httpResponse := HTTPErrorResponse{
		Code:      "VAL_001",
		Message:   "Validation failed",
		Type:      "validation",
		Details:   map[string]interface{}{"field": "email"},
		Timestamp: "2025-07-12T10:30:00Z",
		Path:      "/api/users",
		RequestID: "req_123",
		TraceID:   "trace_456",
	}

	if httpResponse.Code != "VAL_001" {
		t.Errorf("HTTPErrorResponse.Code = %v, want %v", httpResponse.Code, "VAL_001")
	}
	if httpResponse.Message != "Validation failed" {
		t.Errorf("HTTPErrorResponse.Message = %v, want %v", httpResponse.Message, "Validation failed")
	}
	if httpResponse.Type != "validation" {
		t.Errorf("HTTPErrorResponse.Type = %v, want %v", httpResponse.Type, "validation")
	}
	if httpResponse.Path != "/api/users" {
		t.Errorf("HTTPErrorResponse.Path = %v, want %v", httpResponse.Path, "/api/users")
	}
	if httpResponse.RequestID != "req_123" {
		t.Errorf("HTTPErrorResponse.RequestID = %v, want %v", httpResponse.RequestID, "req_123")
	}
	if httpResponse.TraceID != "trace_456" {
		t.Errorf("HTTPErrorResponse.TraceID = %v, want %v", httpResponse.TraceID, "trace_456")
	}
}

// Testes de casos de borda
func TestSeverity_EdgeCases(t *testing.T) {
	t.Run("Negative Severity", func(t *testing.T) {
		severity := Severity(-1)
		result := severity.String()
		if result != "unknown" {
			t.Errorf("Negative Severity.String() = %v, want 'unknown'", result)
		}
	})

	t.Run("Large Severity Value", func(t *testing.T) {
		severity := Severity(1000)
		result := severity.String()
		if result != "unknown" {
			t.Errorf("Large Severity.String() = %v, want 'unknown'", result)
		}
	})
}

func TestParsedError_WithNilDetails(t *testing.T) {
	parsedError := ParsedError{
		Code:      "TEST_001",
		Message:   "Test error",
		Type:      "validation",
		Details:   nil, // Teste com nil
		Severity:  SeverityLow,
		Category:  CategoryTechnical,
		Retryable: false,
		Temporary: false,
	}

	if parsedError.Details != nil {
		t.Errorf("ParsedError.Details should be nil, got %v", parsedError.Details)
	}
}

func TestErrorCodeInfo_WithEmptyFields(t *testing.T) {
	errorInfo := ErrorCodeInfo{
		Code:        "",
		Message:     "",
		Type:        "",
		StatusCode:  0,
		Severity:    SeverityLow,
		Retryable:   false,
		Temporary:   false,
		Tags:        []string{},
		Description: "",
		Examples:    []string{},
	}

	if errorInfo.Code != "" {
		t.Errorf("Expected empty Code, got %v", errorInfo.Code)
	}
	if len(errorInfo.Tags) != 0 {
		t.Errorf("Expected empty Tags slice, got length %v", len(errorInfo.Tags))
	}
	if len(errorInfo.Examples) != 0 {
		t.Errorf("Expected empty Examples slice, got length %v", len(errorInfo.Examples))
	}
}

func TestErrorResponse_WithNilMaps(t *testing.T) {
	response := ErrorResponse{
		Error:      "Test error",
		StatusCode: 500,
		Headers:    nil,
		Metadata:   nil,
	}

	if response.Headers != nil {
		t.Errorf("Expected nil Headers, got %v", response.Headers)
	}
	if response.Metadata != nil {
		t.Errorf("Expected nil Metadata, got %v", response.Metadata)
	}
}

// Benchmarks
func BenchmarkSeverity_String(b *testing.B) {
	severity := SeverityMedium
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = severity.String()
	}
}

func BenchmarkParsedError_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParsedError{
			Code:      "TEST_001",
			Message:   "Test error",
			Type:      "validation",
			Details:   map[string]interface{}{"field": "email"},
			Severity:  SeverityMedium,
			Category:  CategoryTechnical,
			Retryable: false,
			Temporary: false,
		}
	}
}

func BenchmarkErrorCodeInfo_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrorCodeInfo{
			Code:        "ERR_001",
			Message:     "Test error",
			Type:        "validation",
			StatusCode:  400,
			Severity:    SeverityLow,
			Retryable:   false,
			Temporary:   false,
			Tags:        []string{"test"},
			Description: "Test description",
			Examples:    []string{"example"},
		}
	}
}
