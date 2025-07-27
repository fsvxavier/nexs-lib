// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of the Provider interface.
type MockProvider struct {
	mock.Mock
}

// Name returns the provider name.
func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

// Version returns the provider version.
func (m *MockProvider) Version() string {
	args := m.Called()
	return args.String(0)
}

// DoRequest performs a mock HTTP request.
func (m *MockProvider) DoRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.Response), args.Error(1)
}

// Configure configures the mock provider.
func (m *MockProvider) Configure(config *interfaces.Config) error {
	args := m.Called(config)
	return args.Error(0)
}

// SetDefaults sets default values.
func (m *MockProvider) SetDefaults() {
	m.Called()
}

// IsHealthy returns the health status.
func (m *MockProvider) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetMetrics returns provider metrics.
func (m *MockProvider) GetMetrics() *interfaces.ProviderMetrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*interfaces.ProviderMetrics)
}

// NewMockProvider creates a new mock provider with default expectations.
func NewMockProvider() *MockProvider {
	mockProvider := &MockProvider{}

	// Set default expectations
	mockProvider.On("Name").Return("mock-nethttp")
	mockProvider.On("Version").Return("1.0.0-mock")
	mockProvider.On("IsHealthy").Return(true)
	mockProvider.On("SetDefaults").Return()
	mockProvider.On("Configure", mock.AnythingOfType("*interfaces.Config")).Return(nil)
	mockProvider.On("GetMetrics").Return(&interfaces.ProviderMetrics{
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		AverageLatency:     0,
		LastRequestTime:    time.Time{},
	})

	return mockProvider
}

// ExpectSuccessfulRequest sets up expectation for a successful HTTP request.
func (m *MockProvider) ExpectSuccessfulRequest(method, url string, responseBody []byte, statusCode int) *mock.Call {
	response := &interfaces.Response{
		StatusCode: statusCode,
		Body:       responseBody,
		Headers:    make(map[string]string),
		IsError:    statusCode < 200 || statusCode >= 300,
		Latency:    10 * time.Millisecond,
	}

	return m.On("DoRequest", mock.Anything, mock.MatchedBy(func(req *interfaces.Request) bool {
		return req.Method == method && req.URL == url
	})).Return(response, nil)
}

// ExpectFailedRequest sets up expectation for a failed HTTP request.
func (m *MockProvider) ExpectFailedRequest(method, url string, err error) *mock.Call {
	return m.On("DoRequest", mock.Anything, mock.MatchedBy(func(req *interfaces.Request) bool {
		return req.Method == method && req.URL == url
	})).Return(nil, err)
}

// ExpectAnyRequest sets up expectation for any HTTP request.
func (m *MockProvider) ExpectAnyRequest(response *interfaces.Response, err error) *mock.Call {
	return m.On("DoRequest", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*interfaces.Request")).Return(response, err)
}
