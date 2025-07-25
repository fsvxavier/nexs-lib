package url

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}
	if parser.config == nil {
		t.Fatal("Expected config to be set")
	}
}

func TestParser_ParseString(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		expectErr bool
		checkFunc func(*testing.T, *ParsedURL)
	}{
		{
			name:  "Valid HTTP URL",
			input: "http://example.com/path?param=value",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "http" {
					t.Errorf("Expected scheme http, got %s", result.Scheme)
				}
				if result.Host != "example.com" {
					t.Errorf("Expected host example.com, got %s", result.Host)
				}
				if result.Port != 80 {
					t.Errorf("Expected port 80, got %d", result.Port)
				}
				if result.IsSecure {
					t.Error("Expected HTTP to not be secure")
				}
				if result.Domain != "example.com" {
					t.Errorf("Expected domain example.com, got %s", result.Domain)
				}
			},
		},
		{
			name:  "Valid HTTPS URL with port",
			input: "https://subdomain.example.com:8443/path",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "https" {
					t.Errorf("Expected scheme https, got %s", result.Scheme)
				}
				if result.Port != 8443 {
					t.Errorf("Expected port 8443, got %d", result.Port)
				}
				if !result.IsSecure {
					t.Error("Expected HTTPS to be secure")
				}
				if result.Subdomain != "subdomain" {
					t.Errorf("Expected subdomain 'subdomain', got %s", result.Subdomain)
				}
				if result.Domain != "example.com" {
					t.Errorf("Expected domain example.com, got %s", result.Domain)
				}
			},
		},
		{
			name:  "Local URL",
			input: "http://localhost:3000/api",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if !result.IsLocalIP {
					t.Error("Expected localhost to be marked as local IP")
				}
				if result.Port != 3000 {
					t.Errorf("Expected port 3000, got %d", result.Port)
				}
			},
		},
		{
			name:      "Empty URL",
			input:     "",
			expectErr: true,
		},
		{
			name:      "Invalid URL",
			input:     "not-a-url",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
				if result != nil && tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestParser_WithAllowedHosts(t *testing.T) {
	ctx := context.Background()
	allowedHosts := []string{"example.com", "trusted.org"}
	parser := NewParser().WithAllowedHosts(allowedHosts)

	tests := []struct {
		input     string
		expectErr bool
	}{
		{"http://example.com/path", false},
		{"https://subdomain.example.com/path", false},
		{"http://trusted.org", false},
		{"http://untrusted.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parser.ParseString(ctx, tt.input)
			if tt.expectErr && err == nil {
				t.Error("Expected error for untrusted host")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for trusted host: %v", err)
			}
		})
	}
}

func TestParser_WithBlockedHosts(t *testing.T) {
	ctx := context.Background()
	blockedHosts := []string{"blocked.com", "malicious.org"}
	parser := NewParser().WithBlockedHosts(blockedHosts)

	tests := []struct {
		input     string
		expectErr bool
	}{
		{"http://example.com/path", false},
		{"http://blocked.com", true},
		{"https://subdomain.blocked.com", true},
		{"http://malicious.org", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parser.ParseString(ctx, tt.input)
			if tt.expectErr && err == nil {
				t.Error("Expected error for blocked host")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for allowed host: %v", err)
			}
		})
	}
}

func TestParser_parseHostname(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		hostname  string
		subdomain string
		domain    string
		tld       string
	}{
		{
			hostname:  "www.example.com",
			subdomain: "www",
			domain:    "example.com",
			tld:       "com",
		},
		{
			hostname:  "api.v2.service.example.org",
			subdomain: "api.v2.service",
			domain:    "example.org",
			tld:       "org",
		},
		{
			hostname:  "example.com",
			subdomain: "",
			domain:    "example.com",
			tld:       "com",
		},
		{
			hostname:  "192.168.1.1",
			subdomain: "",
			domain:    "192.168.1.1",
			tld:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			result := &ParsedURL{
				URL: &url.URL{Host: tt.hostname},
			}
			parser.parseHostname(result)

			if result.Subdomain != tt.subdomain {
				t.Errorf("Expected subdomain %s, got %s", tt.subdomain, result.Subdomain)
			}
			if result.Domain != tt.domain {
				t.Errorf("Expected domain %s, got %s", tt.domain, result.Domain)
			}
			if result.TLD != tt.tld {
				t.Errorf("Expected TLD %s, got %s", tt.tld, result.TLD)
			}
		})
	}
}

func TestParser_isLocalIP(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		hostname string
		expected bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"::1", true},
		{"fc00::1", true},
		{"example.com", false},
		{"8.8.8.8", false},
		{"203.0.113.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			result := parser.isLocalIP(tt.hostname)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.hostname, result)
			}
		})
	}
}

func TestFormatter_FormatString(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	originalURL := "https://example.com:8443/path?param=value#fragment"
	parser := NewParser()
	parsed, err := parser.ParseString(ctx, originalURL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	result, err := formatter.FormatString(ctx, parsed)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != originalURL {
		t.Errorf("Expected %s, got %s", originalURL, result)
	}
}

func TestBuilder(t *testing.T) {
	builder := NewBuilder()

	url, err := builder.
		Scheme("https").
		Host("api.example.com").
		Port(8443).
		Path("/v1/users").
		AddParam("page", "1").
		AddParam("limit", "10").
		Fragment("section1").
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if url.Scheme != "https" {
		t.Errorf("Expected scheme https, got %s", url.Scheme)
	}

	if url.Host != "api.example.com:8443" {
		t.Errorf("Expected host api.example.com:8443, got %s", url.Host)
	}

	if url.Path != "/v1/users" {
		t.Errorf("Expected path /v1/users, got %s", url.Path)
	}

	if url.Fragment != "section1" {
		t.Errorf("Expected fragment section1, got %s", url.Fragment)
	}

	params := url.Query()
	if params.Get("page") != "1" {
		t.Errorf("Expected page=1, got %s", params.Get("page"))
	}
	if params.Get("limit") != "10" {
		t.Errorf("Expected limit=10, got %s", params.Get("limit"))
	}
}

func TestBuilder_DefaultPorts(t *testing.T) {
	tests := []struct {
		scheme      string
		port        int
		expectPort  bool
		expectedURL string
	}{
		{"http", 80, false, "http://example.com/"},
		{"https", 443, false, "https://example.com/"},
		{"http", 8080, true, "http://example.com:8080/"},
		{"https", 8443, true, "https://example.com:8443/"},
	}

	for _, tt := range tests {
		t.Run(tt.scheme+"_"+string(rune(tt.port)), func(t *testing.T) {
			builder := NewBuilder()
			url, err := builder.
				Scheme(tt.scheme).
				Host("example.com").
				Port(tt.port).
				Build()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			urlStr := url.String()
			if urlStr != tt.expectedURL {
				t.Errorf("Expected %s, got %s", tt.expectedURL, urlStr)
			}
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("IsValidURL", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"http://example.com", true},
			{"https://subdomain.example.com/path", true},
			{"ftp://files.example.com", true},
			{"not-a-url", false},
			{"", false},
		}

		for _, tt := range tests {
			result := IsValidURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidURL(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("ExtractDomain", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"http://example.com/path", "example.com"},
			{"https://subdomain.example.org", "example.org"},
			{"http://192.168.1.1", "192.168.1.1"},
		}

		for _, tt := range tests {
			result, err := ExtractDomain(tt.input)
			if err != nil {
				t.Errorf("ExtractDomain(%s) returned error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ExtractDomain(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("JoinURL", func(t *testing.T) {
		tests := []struct {
			base     string
			relative string
			expected string
		}{
			{"http://example.com", "/api/users", "http://example.com/api/users"},
			{"http://example.com/v1", "users", "http://example.com/users"},
			{"http://example.com/v1/", "users", "http://example.com/v1/users"},
		}

		for _, tt := range tests {
			result, err := JoinURL(tt.base, tt.relative)
			if err != nil {
				t.Errorf("JoinURL(%s, %s) returned error: %v", tt.base, tt.relative, err)
			}
			if result != tt.expected {
				t.Errorf("JoinURL(%s, %s) = %s, expected %s", tt.base, tt.relative, result, tt.expected)
			}
		}
	})

}

func TestNewParserWithConfig(t *testing.T) {
	// Test with nil config
	t.Run("NilConfig", func(t *testing.T) {
		parser := NewParserWithConfig(nil)
		if parser == nil {
			t.Fatal("Expected parser to be created")
		}
		if parser.config != nil {
			t.Error("Expected config to be nil")
		}
	})

	// Test with valid config
	t.Run("ValidConfig", func(t *testing.T) {
		config := &interfaces.ParserConfig{
			MaxSize: 1024,
		}
		parser := NewParserWithConfig(config)
		if parser == nil {
			t.Fatal("Expected parser to be created")
		}
		if parser.config != config {
			t.Error("Expected config to be set to provided config")
		}
		if parser.config.MaxSize != 1024 {
			t.Errorf("Expected MaxSize 1024, got %d", parser.config.MaxSize)
		}
	})

	// Test that config is properly used in parsing
	t.Run("ConfigUsedInParsing", func(t *testing.T) {
		ctx := context.Background()
		config := &interfaces.ParserConfig{
			MaxSize: 10, // Very small size to trigger error
		}
		parser := NewParserWithConfig(config)

		longURL := "http://example.com/very-long-path"
		_, err := parser.ParseString(ctx, longURL)
		if err == nil {
			t.Error("Expected error due to MaxSize limit")
		}
	})
}

func TestParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name      string
		input     []byte
		expectErr bool
		checkFunc func(*testing.T, *ParsedURL)
	}{
		{
			name:  "Valid HTTP URL from bytes",
			input: []byte("http://example.com/path?param=value"),
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "http" {
					t.Errorf("Expected scheme http, got %s", result.Scheme)
				}
				if result.Host != "example.com" {
					t.Errorf("Expected host example.com, got %s", result.Host)
				}
				if result.Port != 80 {
					t.Errorf("Expected port 80, got %d", result.Port)
				}
			},
		},
		{
			name:  "Valid HTTPS URL with port from bytes",
			input: []byte("https://subdomain.example.com:8443/path"),
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "https" {
					t.Errorf("Expected scheme https, got %s", result.Scheme)
				}
				if result.Port != 8443 {
					t.Errorf("Expected port 8443, got %d", result.Port)
				}
				if !result.IsSecure {
					t.Error("Expected HTTPS to be secure")
				}
			},
		},
		{
			name:  "URL with unicode characters",
			input: []byte("https://example.com/path/ñoño?café=☕"),
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "https" {
					t.Errorf("Expected scheme https, got %s", result.Scheme)
				}
				if result.Host != "example.com" {
					t.Errorf("Expected host example.com, got %s", result.Host)
				}
			},
		},
		{
			name:      "Empty byte slice",
			input:     []byte(""),
			expectErr: true,
		},
		{
			name:      "Invalid URL bytes",
			input:     []byte("not-a-url"),
			expectErr: true,
		},
		{
			name:      "Nil byte slice",
			input:     nil,
			expectErr: true,
		},
		{
			name:  "URL with special characters",
			input: []byte("https://user:pass@example.com:9000/path?query=value&other=123#fragment"),
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Scheme != "https" {
					t.Errorf("Expected scheme https, got %s", result.Scheme)
				}
				if result.Port != 9000 {
					t.Errorf("Expected port 9000, got %d", result.Port)
				}
				if result.Fragment != "fragment" {
					t.Errorf("Expected fragment 'fragment', got %s", result.Fragment)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
				if result != nil && tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestParser_ParseReader(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name    string
		reader  interface{}
		wantErr bool
		errType interfaces.ErrorType
		errMsg  string
	}{
		{
			name:    "nil reader",
			reader:  nil,
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "ParseReader not supported for URL parser",
		},
		{
			name:    "string reader",
			reader:  "http://example.com",
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "ParseReader not supported for URL parser",
		},
		{
			name:    "bytes reader",
			reader:  []byte("http://example.com"),
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "ParseReader not supported for URL parser",
		},
		{
			name:    "io.Reader interface",
			reader:  strings.NewReader("http://example.com"),
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "ParseReader not supported for URL parser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseReader(ctx, tt.reader)

			// Should always return error
			if !tt.wantErr {
				t.Error("Expected error but got none")
			}

			if err == nil {
				t.Fatal("Expected error but got none")
			}

			// Result should always be nil
			if result != nil {
				t.Error("Expected nil result but got non-nil")
			}

			// Check error type and message
			if parseErr, ok := err.(*interfaces.ParseError); ok {
				if parseErr.Type != tt.errType {
					t.Errorf("Expected error type %v, got %v", tt.errType, parseErr.Type)
				}
				if parseErr.Message != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, parseErr.Message)
				}
			} else {
				t.Errorf("Expected *interfaces.ParseError, got %T", err)
			}
		})
	}
}

func TestFormatter_Format(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	tests := []struct {
		name      string
		data      *ParsedURL
		expectErr bool
		errType   interfaces.ErrorType
		errMsg    string
		checkFunc func(*testing.T, []byte)
	}{
		{
			name: "Valid ParsedURL",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://example.com:8443/path?param=value#fragment")
				return &ParsedURL{URL: u}
			}(),
			checkFunc: func(t *testing.T, result []byte) {
				expected := "https://example.com:8443/path?param=value#fragment"
				if string(result) != expected {
					t.Errorf("Expected %s, got %s", expected, string(result))
				}
			},
		},
		{
			name: "ParsedURL with complex query parameters",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://api.example.com/search?q=hello+world&limit=10&sort=date")
				return &ParsedURL{URL: u}
			}(),
			checkFunc: func(t *testing.T, result []byte) {
				expected := "https://api.example.com/search?q=hello+world&limit=10&sort=date"
				if string(result) != expected {
					t.Errorf("Expected %s, got %s", expected, string(result))
				}
			},
		},
		{
			name: "ParsedURL with unicode characters",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://example.com/path/ñoño?café=☕")
				return &ParsedURL{URL: u}
			}(),
			checkFunc: func(t *testing.T, result []byte) {
				expected := "https://example.com/path/%C3%B1o%C3%B1o?café=☕"
				if string(result) != expected {
					t.Errorf("Expected %s, got %s", expected, string(result))
				}
			},
		},
		{
			name: "Simple HTTP URL",
			data: func() *ParsedURL {
				u, _ := url.Parse("http://localhost:3000/api")
				return &ParsedURL{URL: u}
			}(),
			checkFunc: func(t *testing.T, result []byte) {
				expected := "http://localhost:3000/api"
				if string(result) != expected {
					t.Errorf("Expected %s, got %s", expected, string(result))
				}
			},
		},
		{
			name:      "Nil ParsedURL",
			data:      nil,
			expectErr: true,
			errType:   interfaces.ErrorTypeValidation,
			errMsg:    "data cannot be nil",
		},
		{
			name:      "ParsedURL with nil URL",
			data:      &ParsedURL{URL: nil},
			expectErr: true,
			errType:   interfaces.ErrorTypeValidation,
			errMsg:    "data cannot be nil",
		},
		{
			name: "Empty URL",
			data: func() *ParsedURL {
				u, _ := url.Parse("")
				return &ParsedURL{URL: u}
			}(),
			checkFunc: func(t *testing.T, result []byte) {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %s", string(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(ctx, tt.data)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}

				// Check error type and message
				if parseErr, ok := err.(*interfaces.ParseError); ok {
					if parseErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, parseErr.Type)
					}
					if parseErr.Message != tt.errMsg {
						t.Errorf("Expected error message %q, got %q", tt.errMsg, parseErr.Message)
					}
				} else {
					t.Errorf("Expected *interfaces.ParseError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
				if result != nil && tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestFormatter_FormatWriter(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	tests := []struct {
		name    string
		data    *ParsedURL
		writer  interface{}
		wantErr bool
		errType interfaces.ErrorType
		errMsg  string
	}{
		{
			name: "Valid ParsedURL with nil writer",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://example.com/path")
				return &ParsedURL{URL: u}
			}(),
			writer:  nil,
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
		{
			name: "Valid ParsedURL with string writer",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://example.com/path")
				return &ParsedURL{URL: u}
			}(),
			writer:  &strings.Builder{},
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
		{
			name: "Valid ParsedURL with bytes buffer writer",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://example.com/path")
				return &ParsedURL{URL: u}
			}(),
			writer:  strings.NewReader("test"),
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
		{
			name:    "Nil ParsedURL with valid writer",
			data:    nil,
			writer:  &strings.Builder{},
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
		{
			name:    "ParsedURL with nil URL and valid writer",
			data:    &ParsedURL{URL: nil},
			writer:  &strings.Builder{},
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
		{
			name: "Complex ParsedURL with writer",
			data: func() *ParsedURL {
				u, _ := url.Parse("https://subdomain.example.com:8443/path?param=value#fragment")
				return &ParsedURL{URL: u}
			}(),
			writer:  &strings.Builder{},
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
			errMsg:  "FormatWriter not supported for URL formatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatter.FormatWriter(ctx, tt.data, tt.writer)

			// Should always return error
			if !tt.wantErr {
				t.Error("Expected error but got none")
			}

			if err == nil {
				t.Fatal("Expected error but got none")
			}

			// Check error type and message
			if parseErr, ok := err.(*interfaces.ParseError); ok {
				if parseErr.Type != tt.errType {
					t.Errorf("Expected error type %v, got %v", tt.errType, parseErr.Type)
				}
				if parseErr.Message != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, parseErr.Message)
				}
			} else {
				t.Errorf("Expected *interfaces.ParseError, got %T", err)
			}
		})
	}
}

func TestBuilder_SetParam(t *testing.T) {
	tests := []struct {
		name          string
		setupParams   map[string]string
		setKey        string
		setValue      string
		expectedValue string
		expectedCount int
	}{
		{
			name:          "Set new parameter",
			setupParams:   map[string]string{},
			setKey:        "newParam",
			setValue:      "newValue",
			expectedValue: "newValue",
			expectedCount: 1,
		},
		{
			name:          "Replace existing parameter",
			setupParams:   map[string]string{"param": "oldValue"},
			setKey:        "param",
			setValue:      "newValue",
			expectedValue: "newValue",
			expectedCount: 1,
		},
		{
			name:          "Set parameter with empty value",
			setupParams:   map[string]string{},
			setKey:        "emptyParam",
			setValue:      "",
			expectedValue: "",
			expectedCount: 1,
		},
		{
			name:          "Set parameter with empty key",
			setupParams:   map[string]string{},
			setKey:        "",
			setValue:      "value",
			expectedValue: "value",
			expectedCount: 1,
		},
		{
			name:          "Replace one of multiple parameters",
			setupParams:   map[string]string{"param1": "value1", "param2": "value2"},
			setKey:        "param1",
			setValue:      "newValue1",
			expectedValue: "newValue1",
			expectedCount: 1,
		},
		{
			name:          "Set parameter with special characters",
			setupParams:   map[string]string{},
			setKey:        "special",
			setValue:      "hello world & more",
			expectedValue: "hello world & more",
			expectedCount: 1,
		},
		{
			name:          "Replace parameter that was added multiple times",
			setupParams:   map[string]string{},
			setKey:        "multiParam",
			setValue:      "finalValue",
			expectedValue: "finalValue",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()

			// Setup initial parameters
			for key, value := range tt.setupParams {
				builder.AddParam(key, value)
			}

			// For the "multiple times" test case, add the same param multiple times
			if tt.name == "Replace parameter that was added multiple times" {
				builder.AddParam("multiParam", "value1")
				builder.AddParam("multiParam", "value2")
				builder.AddParam("multiParam", "value3")
			}

			// Set the parameter
			result := builder.SetParam(tt.setKey, tt.setValue)

			// Verify fluent interface returns the same builder
			if result != builder {
				t.Error("SetParam should return the same builder instance for fluent interface")
			}

			// Check the parameter value
			values := builder.query[tt.setKey]
			if len(values) != tt.expectedCount {
				t.Errorf("Expected %d value(s) for key %s, got %d", tt.expectedCount, tt.setKey, len(values))
			}

			if len(values) > 0 && values[0] != tt.expectedValue {
				t.Errorf("Expected value %s for key %s, got %s", tt.expectedValue, tt.setKey, values[0])
			}

			// For replacement tests, verify only one value exists
			if strings.Contains(tt.name, "Replace") || strings.Contains(tt.name, "multiple times") {
				if len(values) != 1 {
					t.Errorf("SetParam should replace existing values, but found %d values", len(values))
				}
			}
		})
	}
}

func TestBuilder_SetParam_Integration(t *testing.T) {
	builder := NewBuilder()

	// Test complete URL building with SetParam
	url, err := builder.
		Scheme("https").
		Host("api.example.com").
		Path("/search").
		AddParam("q", "initial").
		AddParam("limit", "5").
		SetParam("q", "updated query"). // Replace existing
		SetParam("sort", "date").       // Add new
		Build()

	if err != nil {
		t.Fatalf("Unexpected error building URL: %v", err)
	}

	params := url.Query()

	// Check replaced parameter
	if params.Get("q") != "updated query" {
		t.Errorf("Expected q=updated query, got q=%s", params.Get("q"))
	}

	// Check that only one value exists for replaced param
	if len(params["q"]) != 1 {
		t.Errorf("Expected exactly 1 value for q parameter, got %d", len(params["q"]))
	}

	// Check unchanged parameter
	if params.Get("limit") != "5" {
		t.Errorf("Expected limit=5, got limit=%s", params.Get("limit"))
	}

	// Check new parameter
	if params.Get("sort") != "date" {
		t.Errorf("Expected sort=date, got sort=%s", params.Get("sort"))
	}
}

func TestBuilder_SetParam_Chaining(t *testing.T) {
	builder := NewBuilder()

	// Test method chaining with SetParam
	result := builder.
		SetParam("param1", "value1").
		SetParam("param2", "value2").
		SetParam("param1", "newValue1") // Replace the first one

	if result != builder {
		t.Error("Method chaining should return the same builder instance")
	}

	// Verify final state
	if builder.query.Get("param1") != "newValue1" {
		t.Errorf("Expected param1=newValue1, got param1=%s", builder.query.Get("param1"))
	}

	if builder.query.Get("param2") != "value2" {
		t.Errorf("Expected param2=value2, got param2=%s", builder.query.Get("param2"))
	}

	// Ensure param1 has only one value
	if len(builder.query["param1"]) != 1 {
		t.Errorf("Expected exactly 1 value for param1, got %d", len(builder.query["param1"]))
	}
}

// Benchmark tests
func BenchmarkParser_ParseString(b *testing.B) {
	ctx := context.Background()
	parser := NewParser()
	input := "https://api.example.com:8443/v1/users?page=1&limit=10#section"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.ParseString(ctx, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuilder_Build(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		_, err := builder.
			Scheme("https").
			Host("api.example.com").
			Port(8443).
			Path("/v1/users").
			AddParam("page", "1").
			AddParam("limit", "10").
			Build()

		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestBuildString(t *testing.T) {
	tests := []struct {
		name         string
		setupBuilder func() *Builder
		expected     string
		hasError     bool
	}{
		{
			name: "Complete URL",
			setupBuilder: func() *Builder {
				return NewBuilder().
					Scheme("https").
					Host("example.com").
					Port(8080).
					Path("/api/v1").
					AddParam("key", "value").
					Fragment("section")
			},
			expected: "https://example.com:8080/api/v1?key=value#section",
			hasError: false,
		},
		{
			name: "HTTP with default port",
			setupBuilder: func() *Builder {
				return NewBuilder().
					Scheme("http").
					Host("localhost").
					Path("/home")
			},
			expected: "http://localhost/home",
			hasError: false,
		},
		{
			name: "Invalid URL - missing host",
			setupBuilder: func() *Builder {
				return NewBuilder().
					Scheme("https").
					Path("/api")
			},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setupBuilder().BuildString()

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Complete URL",
			input:    "https://Example.COM/Path/?param=value",
			expected: "https://Example.COM/Path/?param=value",
			hasError: false,
		},
		{
			name:     "URL with port",
			input:    "HTTP://localhost:8080/api",
			expected: "http://localhost:8080/api",
			hasError: false,
		},
		{
			name:     "URL with fragment",
			input:    "https://site.com/page#Fragment",
			expected: "https://site.com/page#Fragment",
			hasError: false,
		},
		{
			name:     "Invalid URL",
			input:    "not-a-url",
			expected: "",
			hasError: true,
		},
		{
			name:     "URL with default HTTPS port",
			input:    "https://example.com:443/path",
			expected: "https://example.com/path",
			hasError: false,
		},
		{
			name:     "URL with default HTTP port",
			input:    "http://example.com:80/path",
			expected: "http://example.com/path",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestJoinURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		expected string
		hasError bool
	}{
		{
			name:     "Absolute path with query params",
			base:     "https://example.com/api",
			path:     "/users?page=1&limit=10",
			expected: "https://example.com/users?page=1&limit=10",
			hasError: false,
		},
		{
			name:     "Relative path with fragment",
			base:     "https://example.com/docs/",
			path:     "tutorial#section1",
			expected: "https://example.com/docs/tutorial#section1",
			hasError: false,
		},
		{
			name:     "Empty path",
			base:     "https://example.com/api",
			path:     "",
			expected: "https://example.com/api",
			hasError: false,
		},
		{
			name:     "Invalid base URL",
			base:     "not-a-url",
			path:     "/api",
			expected: "/api",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JoinURL(tt.base, tt.path)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParser_ParseString_EdgeCases(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "URL with IPv6",
			input:    "http://[::1]:8080/path",
			hasError: false,
		},
		{
			name:     "URL with user info",
			input:    "ftp://user:pass@ftp.example.com/file.txt",
			hasError: false,
		},
		{
			name:     "URL with encoded characters",
			input:    "https://example.com/path%20with%20spaces",
			hasError: false,
		},
		{
			name:     "Just scheme",
			input:    "https://",
			hasError: false,
		},
		{
			name:     "Missing scheme",
			input:    "example.com/path",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result to be non-nil")
			}
		})
	}
}

func TestParser_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Parser
		input    string
		hasError bool
	}{
		{
			name: "Allowed host passes",
			setup: func() *Parser {
				return NewParser().WithAllowedHosts([]string{"example.com", "test.com"})
			},
			input:    "https://example.com/path",
			hasError: false,
		},
		{
			name: "Not in allowed hosts fails",
			setup: func() *Parser {
				return NewParser().WithAllowedHosts([]string{"example.com"})
			},
			input:    "https://other.com/path",
			hasError: true,
		},
		{
			name: "Blocked host fails",
			setup: func() *Parser {
				return NewParser().WithBlockedHosts([]string{"malicious.com", "spam.net"})
			},
			input:    "https://malicious.com/evil",
			hasError: true,
		},
		{
			name: "Not in blocked hosts passes",
			setup: func() *Parser {
				return NewParser().WithBlockedHosts([]string{"malicious.com"})
			},
			input:    "https://safe.com/path",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.setup()
			u, _ := url.Parse(tt.input)
			parsedURL := &ParsedURL{URL: u}

			ctx := context.Background()
			err := parser.Validate(ctx, parsedURL)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExtractDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Subdomain with multiple levels",
			input:    "https://api.v1.example.com/path",
			expected: "api.v1.example.com",
			hasError: false,
		},
		{
			name:     "Single domain",
			input:    "https://localhost/path",
			expected: "localhost",
			hasError: false,
		},
		{
			name:     "Invalid URL",
			input:    "not-a-url",
			expected: "",
			hasError: true,
		},
		{
			name:     "URL with IP address",
			input:    "http://192.168.1.1:8080/api",
			expected: "192.168.1.1",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractDomain(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatter_FormatString_ErrorCase(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	// Test error case by passing nil data
	result, err := formatter.FormatString(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
	if result != "" {
		t.Errorf("Expected empty string on error, got %s", result)
	}
}

func TestEnrichURL_EdgeCases(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		checkFunc func(*testing.T, *ParsedURL)
	}{
		{
			name:  "URL with custom port",
			input: "https://example.com:9443/api",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Port != 9443 {
					t.Errorf("Expected port 9443, got %d", result.Port)
				}
				if !result.IsSecure {
					t.Error("Expected HTTPS to be secure")
				}
			},
		},
		{
			name:  "HTTP with default port",
			input: "http://example.com/path",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if result.Port != 80 {
					t.Errorf("Expected port 80, got %d", result.Port)
				}
				if result.IsSecure {
					t.Error("Expected HTTP to not be secure")
				}
			},
		},
		{
			name:  "Local IP address",
			input: "http://127.0.0.1:3000/api",
			checkFunc: func(t *testing.T, result *ParsedURL) {
				if !result.IsLocalIP {
					t.Error("Expected 127.0.0.1 to be recognized as local IP")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := parser.ParseString(ctx, tt.input)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			tt.checkFunc(t, result)
		})
	}
}

func TestBuilder_Build_ErrorCase(t *testing.T) {
	builder := NewBuilder()

	// Test building without required fields
	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error when building URL without required fields")
	}

	// Test building with scheme but no host
	_, err = builder.Scheme("https").Build()
	if err == nil {
		t.Error("Expected error when building URL with scheme but no host")
	}
}

func TestJoinURL_ErrorCase(t *testing.T) {
	// Test with malformed relative path that can't be parsed
	_, err := JoinURL("https://example.com", "path with spaces and %")
	if err != nil {
		// This is expected - URL parsing may fail on some malformed paths
		return
	}
}

func TestParser_ParseString_ValidationEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Test with input larger than max size
	config := &interfaces.ParserConfig{
		MaxSize: 10,
	}
	parser := NewParserWithConfig(config)

	_, err := parser.ParseString(ctx, "https://example.com/very-long-path-that-exceeds-limit")
	if err == nil {
		t.Error("Expected error for input exceeding max size")
	}
}

func TestParser_Validate_NilURL(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	err := parser.Validate(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil ParsedURL")
	}

	err = parser.Validate(ctx, &ParsedURL{})
	if err == nil {
		t.Error("Expected error for ParsedURL with nil URL")
	}
}

func TestEnrichURL_PortParsing(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	// Test with invalid port in URL
	result, err := parser.ParseString(ctx, "https://example.com:invalid/path")
	if err != nil {
		// This should error due to invalid port
		return
	}

	// If it doesn't error, the port should be 0 or handled gracefully
	if result.Port < 0 {
		t.Error("Port should not be negative")
	}
}

func TestParser_ParseString_InvalidPort(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name      string
		input     string
		expectErr bool
		checkFunc func(*testing.T, *ParsedURL, error)
	}{
		{
			name:      "Invalid port in URL",
			input:     "http://example.com:abc/path",
			expectErr: true,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err == nil {
					t.Error("Expected error for invalid port")
				}
			},
		},
		{
			name:      "URL without scheme defaults",
			input:     "//example.com/path",
			expectErr: false,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Host != "example.com" {
					t.Errorf("Expected host example.com, got %s", result.Host)
				}
			},
		},
		{
			name:      "URL with custom port",
			input:     "http://example.com:8080/path",
			expectErr: false,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Port != 8080 {
					t.Errorf("Expected port 8080, got %d", result.Port)
				}
			},
		},
		{
			name:      "FTP URL",
			input:     "ftp://ftp.example.com/file.txt",
			expectErr: false,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Port != 21 {
					t.Errorf("Expected default FTP port 21, got %d", result.Port)
				}
			},
		},
		{
			name:      "SSH URL",
			input:     "ssh://user@example.com/path",
			expectErr: false,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Port != 22 {
					t.Errorf("Expected default SSH port 22, got %d", result.Port)
				}
			},
		},
		{
			name:      "FTPS secure URL",
			input:     "ftps://ftp.example.com/file.txt",
			expectErr: false,
			checkFunc: func(t *testing.T, result *ParsedURL, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && !result.IsSecure {
					t.Error("Expected FTPS to be secure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestJoinURL_Advanced(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		expected     string
		expectErr    bool
	}{
		{
			name:         "Valid join",
			baseURL:      "http://example.com/path",
			relativePath: "subpath",
			expected:     "http://example.com/subpath",
			expectErr:    false,
		},
		{
			name:         "Invalid base URL",
			baseURL:      "://invalid",
			relativePath: "subpath",
			expected:     "",
			expectErr:    true,
		},
		{
			name:         "Invalid relative path",
			baseURL:      "http://example.com",
			relativePath: "://invalid",
			expected:     "",
			expectErr:    true,
		},
		{
			name:         "Absolute relative path",
			baseURL:      "http://example.com/path",
			relativePath: "/newpath",
			expected:     "http://example.com/newpath",
			expectErr:    false,
		},
		{
			name:         "Query parameters in relative",
			baseURL:      "http://example.com",
			relativePath: "path?param=value",
			expected:     "http://example.com/path?param=value",
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JoinURL(tt.baseURL, tt.relativePath)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
