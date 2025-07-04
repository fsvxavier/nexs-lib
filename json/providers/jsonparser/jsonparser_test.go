package jsonparser

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type testStruct struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
	Active  bool   `json:"active"`
}

var (
	testJSON     = `{"name":"Test User","age":30,"address":"123 Test St","active":true}`
	invalidJSON  = `{"name":"Test User","age":30,"address":"123 Test St",active:true}`
	expectedData = testStruct{
		Name:    "Test User",
		Age:     30,
		Address: "123 Test St",
		Active:  true,
	}
)

func TestNew(t *testing.T) {
	provider := New()
	if provider == nil {
		t.Error("New() returned nil")
	}
}

func TestProviderMarshal(t *testing.T) {
	provider := New()
	data := testStruct{
		Name:    "Test User",
		Age:     30,
		Address: "123 Test St",
		Active:  true,
	}

	got, err := provider.Marshal(data)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(got, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
		return
	}

	if result["name"] != data.Name {
		t.Errorf("Marshal() name = %v, want %v", result["name"], data.Name)
	}
	if int(result["age"].(float64)) != data.Age {
		t.Errorf("Marshal() age = %v, want %v", result["age"], data.Age)
	}
	if result["address"] != data.Address {
		t.Errorf("Marshal() address = %v, want %v", result["address"], data.Address)
	}
	if result["active"] != data.Active {
		t.Errorf("Marshal() active = %v, want %v", result["active"], data.Active)
	}
}

func TestProviderUnmarshal(t *testing.T) {
	provider := New()
	var got testStruct
	err := provider.Unmarshal([]byte(testJSON), &got)
	if err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}

	if got.Name != expectedData.Name {
		t.Errorf("Unmarshal() name = %v, want %v", got.Name, expectedData.Name)
	}
	if got.Age != expectedData.Age {
		t.Errorf("Unmarshal() age = %v, want %v", got.Age, expectedData.Age)
	}
	if got.Address != expectedData.Address {
		t.Errorf("Unmarshal() address = %v, want %v", got.Address, expectedData.Address)
	}
	if got.Active != expectedData.Active {
		t.Errorf("Unmarshal() active = %v, want %v", got.Active, expectedData.Active)
	}
}

func TestProviderNewDecoder(t *testing.T) {
	provider := New()
	reader := strings.NewReader(testJSON)
	decoder := provider.NewDecoder(reader)
	if decoder == nil {
		t.Errorf("NewDecoder() returned nil")
		return
	}

	var got testStruct
	err := decoder.Decode(&got)
	if err != nil {
		t.Errorf("Decoder.Decode() error = %v", err)
		return
	}

	if got.Name != expectedData.Name {
		t.Errorf("Decoder.Decode() name = %v, want %v", got.Name, expectedData.Name)
	}
	if got.Age != expectedData.Age {
		t.Errorf("Decoder.Decode() age = %v, want %v", got.Age, expectedData.Age)
	}
	if got.Address != expectedData.Address {
		t.Errorf("Decoder.Decode() address = %v, want %v", got.Address, expectedData.Address)
	}
	if got.Active != expectedData.Active {
		t.Errorf("Decoder.Decode() active = %v, want %v", got.Active, expectedData.Active)
	}
}

func TestProviderNewEncoder(t *testing.T) {
	provider := New()
	buf := new(bytes.Buffer)
	encoder := provider.NewEncoder(buf)
	if encoder == nil {
		t.Errorf("NewEncoder() returned nil")
		return
	}

	err := encoder.Encode(expectedData)
	if err != nil {
		t.Errorf("Encoder.Encode() error = %v", err)
		return
	}

	var got testStruct
	err = json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		t.Errorf("Failed to unmarshal encoded data: %v", err)
		return
	}

	if got.Name != expectedData.Name {
		t.Errorf("Encoder.Encode() name = %v, want %v", got.Name, expectedData.Name)
	}
	if got.Age != expectedData.Age {
		t.Errorf("Encoder.Encode() age = %v, want %v", got.Age, expectedData.Age)
	}
	if got.Address != expectedData.Address {
		t.Errorf("Encoder.Encode() address = %v, want %v", got.Address, expectedData.Address)
	}
	if got.Active != expectedData.Active {
		t.Errorf("Encoder.Encode() active = %v, want %v", got.Active, expectedData.Active)
	}
}

func TestProviderValid(t *testing.T) {
	provider := New()

	tests := []struct {
		name     string
		jsonData string
		want     bool
	}{
		{
			name:     "Valid JSON",
			jsonData: testJSON,
			want:     true,
		},
		{
			name:     "Invalid JSON",
			jsonData: invalidJSON,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.Valid([]byte(tt.jsonData)); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderDecodeReader(t *testing.T) {
	provider := New()
	reader := strings.NewReader(testJSON)
	var got testStruct
	err := provider.DecodeReader(reader, &got)
	if err != nil {
		t.Errorf("DecodeReader() error = %v", err)
		return
	}

	if got.Name != expectedData.Name {
		t.Errorf("DecodeReader() name = %v, want %v", got.Name, expectedData.Name)
	}
	if got.Age != expectedData.Age {
		t.Errorf("DecodeReader() age = %v, want %v", got.Age, expectedData.Age)
	}
	if got.Address != expectedData.Address {
		t.Errorf("DecodeReader() address = %v, want %v", got.Address, expectedData.Address)
	}
	if got.Active != expectedData.Active {
		t.Errorf("DecodeReader() active = %v, want %v", got.Active, expectedData.Active)
	}
}

func TestProviderEncode(t *testing.T) {
	provider := New()
	got, err := provider.Encode(expectedData)
	if err != nil {
		t.Errorf("Encode() error = %v", err)
		return
	}

	var result testStruct
	err = json.Unmarshal(got, &result)
	if err != nil {
		t.Errorf("Failed to unmarshal encoded data: %v", err)
		return
	}

	if result.Name != expectedData.Name {
		t.Errorf("Encode() name = %v, want %v", result.Name, expectedData.Name)
	}
	if result.Age != expectedData.Age {
		t.Errorf("Encode() age = %v, want %v", result.Age, expectedData.Age)
	}
	if result.Address != expectedData.Address {
		t.Errorf("Encode() address = %v, want %v", result.Address, expectedData.Address)
	}
	if result.Active != expectedData.Active {
		t.Errorf("Encode() active = %v, want %v", result.Active, expectedData.Active)
	}
}

// Jsonparser may have limited support for Decoder/Encoder features
// but we test what it does support
func TestDecoderMethods(t *testing.T) {
	provider := New()
	reader := strings.NewReader(testJSON)
	decoder := provider.NewDecoder(reader)

	// Test UseNumber
	decoder = decoder.UseNumber()
	if decoder == nil {
		t.Errorf("UseNumber() returned nil")
	}

	// Test DisallowUnknownFields
	decoder = decoder.DisallowUnknownFields()
	if decoder == nil {
		t.Errorf("DisallowUnknownFields() returned nil")
	}

	// Test Decode
	var got testStruct
	err := decoder.Decode(&got)
	if err != nil {
		t.Errorf("Decode() after method chaining error = %v", err)
	}

	// Test Buffered
	buffered := decoder.Buffered()
	if buffered == nil {
		t.Errorf("Buffered() returned nil")
	}
}

func TestEncoderMethods(t *testing.T) {
	provider := New()
	buf := new(bytes.Buffer)
	encoder := provider.NewEncoder(buf)

	// Test SetIndent
	encoder = encoder.SetIndent("", "  ")
	if encoder == nil {
		t.Errorf("SetIndent() returned nil")
	}

	// Test SetEscapeHTML
	encoder = encoder.SetEscapeHTML(false)
	if encoder == nil {
		t.Errorf("SetEscapeHTML() returned nil")
	}

	// Test Encode after method chaining
	err := encoder.Encode(expectedData)
	if err != nil {
		t.Errorf("Encode() after method chaining error = %v", err)
	}
}

// Benchmarks
func BenchmarkProviderMarshal(b *testing.B) {
	provider := New()
	data := testStruct{
		Name:    "Test User",
		Age:     30,
		Address: "123 Test St",
		Active:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.Marshal(data)
	}
}

func BenchmarkProviderUnmarshal(b *testing.B) {
	provider := New()
	jsonData := []byte(testJSON)
	var result testStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Unmarshal(jsonData, &result)
	}
}

func BenchmarkProviderValid(b *testing.B) {
	provider := New()
	jsonData := []byte(testJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Valid(jsonData)
	}
}

func BenchmarkProviderDecodeReader(b *testing.B) {
	provider := New()
	jsonStr := testJSON
	var result testStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(jsonStr)
		_ = provider.DecodeReader(reader, &result)
	}
}

func BenchmarkProviderEncode(b *testing.B) {
	provider := New()
	data := testStruct{
		Name:    "Test User",
		Age:     30,
		Address: "123 Test St",
		Active:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.Encode(data)
	}
}
