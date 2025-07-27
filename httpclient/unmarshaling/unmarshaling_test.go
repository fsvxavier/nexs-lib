package unmarshaling

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func TestUnmarshaler_UnmarshalJSON(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalJSON)

	resp := &interfaces.Response{
		Body:        []byte(`{"name": "test", "value": 123}`),
		ContentType: "application/json",
	}

	var result map[string]interface{}
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name=test, got %v", result["name"])
	}

	if result["value"] != float64(123) {
		t.Errorf("Expected value=123, got %v", result["value"])
	}
}

func TestUnmarshaler_UnmarshalXML(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalXML)

	resp := &interfaces.Response{
		Body:        []byte(`<person><name>test</name><age>30</age></person>`),
		ContentType: "application/xml",
	}

	type Person struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	var result Person
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected name=test, got %v", result.Name)
	}

	if result.Age != 30 {
		t.Errorf("Expected age=30, got %v", result.Age)
	}
}

func TestUnmarshaler_AutoDetectJSON(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalAuto)

	resp := &interfaces.Response{
		Body:        []byte(`{"test": true}`),
		ContentType: "application/json",
	}

	var result map[string]interface{}
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["test"] != true {
		t.Errorf("Expected test=true, got %v", result["test"])
	}
}

func TestUnmarshaler_AutoDetectXML(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalAuto)

	resp := &interfaces.Response{
		Body:        []byte(`<root><test>value</test></root>`),
		ContentType: "text/xml",
	}

	type Root struct {
		Test string `xml:"test"`
	}

	var result Root
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Test != "value" {
		t.Errorf("Expected test=value, got %v", result.Test)
	}
}

func TestUnmarshaler_RawData(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalNone)

	resp := &interfaces.Response{
		Body:        []byte("raw text data"),
		ContentType: "text/plain",
	}

	var result string
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "raw text data" {
		t.Errorf("Expected 'raw text data', got %v", result)
	}
}

func TestUnmarshaler_EmptyResponse(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalJSON)

	resp := &interfaces.Response{
		Body: []byte{},
	}

	var result map[string]interface{}
	err := unmarshaler.UnmarshalResponse(resp, &result)

	if err == nil {
		t.Error("Expected error for empty response")
	}
}

func TestUnmarshaler_NilResponse(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalJSON)

	var result map[string]interface{}
	err := unmarshaler.UnmarshalResponse(nil, &result)

	if err == nil {
		t.Error("Expected error for nil response")
	}
}

func TestDetermineStrategy(t *testing.T) {
	unmarshaler := NewUnmarshaler(interfaces.UnmarshalAuto)

	tests := []struct {
		name        string
		contentType string
		expected    interfaces.UnmarshalStrategy
	}{
		{"JSON", "application/json", interfaces.UnmarshalJSON},
		{"JSON with charset", "application/json; charset=utf-8", interfaces.UnmarshalJSON},
		{"XML", "application/xml", interfaces.UnmarshalXML},
		{"Text XML", "text/xml", interfaces.UnmarshalXML},
		{"Plain text", "text/plain", interfaces.UnmarshalNone},
		{"Octet stream", "application/octet-stream", interfaces.UnmarshalNone},
		{"Unknown", "application/unknown", interfaces.UnmarshalJSON},
		{"Empty", "", interfaces.UnmarshalJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &interfaces.Response{
				ContentType: tt.contentType,
			}

			strategy := unmarshaler.determineStrategy(resp)
			if strategy != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, strategy)
			}
		})
	}
}
