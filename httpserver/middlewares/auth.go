package middlewares

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// AuthMiddleware provides authentication and authorization functionality.
type AuthMiddleware struct {
	*BaseMiddleware

	// Configuration
	config AuthConfig

	// Providers
	providers map[string]AuthProvider

	// Metrics
	authAttempts     int64
	authSuccesses    int64
	authFailures     int64
	authErrors       int64
	tokenValidations int64

	// Internal state
	startTime time.Time
}

// AuthConfig defines configuration options for the authentication middleware.
type AuthConfig struct {
	// Authentication types
	EnableBasicAuth  bool
	EnableBearerAuth bool
	EnableAPIKeyAuth bool
	EnableJWTAuth    bool
	EnableCustomAuth bool

	// Basic Auth configuration
	BasicAuthRealm string
	BasicAuthUsers map[string]string // username -> password

	// Bearer/API Key configuration
	ValidTokens map[string]AuthUser // token -> user info

	// JWT configuration
	JWTSecret     string
	JWTAlgorithm  string
	JWTExpiration time.Duration

	// Request configuration
	AuthHeader     string
	APIKeyHeader   string
	AuthQueryParam string

	// Behavior configuration
	SkipPaths         []string
	RequireAuth       bool
	AllowMultipleAuth bool
	CaseSensitiveAuth bool

	// Security configuration
	MaxLoginAttempts    int
	LoginAttemptsWindow time.Duration
	LockoutDuration     time.Duration

	// Response configuration
	UnauthorizedMessage string
	ForbiddenMessage    string
	AuthRequiredMessage string
}

// AuthProvider defines the interface for authentication providers.
type AuthProvider interface {
	Authenticate(ctx context.Context, credentials interface{}) (*AuthUser, error)
	ValidateToken(ctx context.Context, token string) (*AuthUser, error)
	GetName() string
	IsEnabled() bool
}

// AuthUser represents an authenticated user.
type AuthUser struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt   time.Time              `json:"expires_at,omitempty"`
}

// AuthContext represents authentication context for a request.
type AuthContext struct {
	User          *AuthUser
	Token         string
	AuthType      string
	Provider      string
	Authenticated bool
	Timestamp     time.Time
}

// BasicAuthProvider implements basic HTTP authentication.
type BasicAuthProvider struct {
	name    string
	enabled bool
	users   map[string]string
	realm   string
}

// APIKeyAuthProvider implements API key authentication.
type APIKeyAuthProvider struct {
	name    string
	enabled bool
	tokens  map[string]AuthUser
	header  string
}

// NewAuthMiddleware creates a new authentication middleware with default configuration.
func NewAuthMiddleware(priority int) *AuthMiddleware {
	return &AuthMiddleware{
		BaseMiddleware: NewBaseMiddleware("auth", priority),
		config:         DefaultAuthConfig(),
		providers:      make(map[string]AuthProvider),
		startTime:      time.Now(),
	}
}

// NewAuthMiddlewareWithConfig creates a new authentication middleware with custom configuration.
func NewAuthMiddlewareWithConfig(priority int, config AuthConfig) *AuthMiddleware {
	middleware := &AuthMiddleware{
		BaseMiddleware: NewBaseMiddleware("auth", priority),
		config:         config,
		providers:      make(map[string]AuthProvider),
		startTime:      time.Now(),
	}

	// Initialize default providers based on configuration
	middleware.initializeProviders()

	return middleware
}

// DefaultAuthConfig returns a default authentication configuration.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		EnableBasicAuth:     true,
		EnableBearerAuth:    true,
		EnableAPIKeyAuth:    false,
		EnableJWTAuth:       false,
		EnableCustomAuth:    false,
		BasicAuthRealm:      "Restricted Area",
		BasicAuthUsers:      make(map[string]string),
		ValidTokens:         make(map[string]AuthUser),
		AuthHeader:          "Authorization",
		APIKeyHeader:        "X-API-Key",
		AuthQueryParam:      "token",
		SkipPaths:           []string{"/health", "/metrics", "/favicon.ico"},
		RequireAuth:         true,
		AllowMultipleAuth:   false,
		CaseSensitiveAuth:   true,
		MaxLoginAttempts:    5,
		LoginAttemptsWindow: time.Minute * 15,
		LockoutDuration:     time.Minute * 30,
		UnauthorizedMessage: "Authentication required",
		ForbiddenMessage:    "Access denied",
		AuthRequiredMessage: "Authentication credentials required",
	}
}

// Process implements the Middleware interface for authentication.
func (am *AuthMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !am.IsEnabled() {
		return next(ctx, req)
	}

	atomic.AddInt64(&am.authAttempts, 1)

	// Extract request information
	reqInfo := am.extractRequestInfo(req)

	// Check if path should skip authentication
	if am.shouldSkipAuth(reqInfo.Path) {
		return next(ctx, req)
	}

	// Perform authentication
	authCtx, err := am.authenticate(ctx, reqInfo)
	if err != nil {
		atomic.AddInt64(&am.authErrors, 1)
		am.GetLogger().Error("Authentication error: %v", err)
		return nil, am.createAuthError(err)
	}

	if !authCtx.Authenticated && am.config.RequireAuth {
		atomic.AddInt64(&am.authFailures, 1)
		am.GetLogger().Warn("Authentication failed for path: %s", reqInfo.Path)
		return nil, am.createUnauthorizedError()
	}

	if authCtx.Authenticated {
		atomic.AddInt64(&am.authSuccesses, 1)
		am.GetLogger().Debug("Authentication successful for user: %s", authCtx.User.Username)
	}

	// Add authentication context to request context
	ctxWithAuth := context.WithValue(ctx, "auth", authCtx)
	ctxWithAuth = context.WithValue(ctxWithAuth, "user", authCtx.User)

	return next(ctxWithAuth, req)
}

// AddProvider adds an authentication provider.
func (am *AuthMiddleware) AddProvider(provider AuthProvider) error {
	if provider == nil {
		return errors.New("provider cannot be nil")
	}

	am.providers[provider.GetName()] = provider
	am.GetLogger().Info("Authentication provider '%s' added", provider.GetName())
	return nil
}

// RemoveProvider removes an authentication provider.
func (am *AuthMiddleware) RemoveProvider(name string) error {
	if _, exists := am.providers[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	delete(am.providers, name)
	am.GetLogger().Info("Authentication provider '%s' removed", name)
	return nil
}

// GetProvider retrieves an authentication provider by name.
func (am *AuthMiddleware) GetProvider(name string) (AuthProvider, error) {
	provider, exists := am.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return provider, nil
}

// GetConfig returns the current authentication configuration.
func (am *AuthMiddleware) GetConfig() AuthConfig {
	return am.config
}

// SetConfig updates the authentication configuration.
func (am *AuthMiddleware) SetConfig(config AuthConfig) {
	am.config = config
	am.initializeProviders()
	am.GetLogger().Info("Authentication middleware configuration updated")
}

// GetMetrics returns authentication metrics.
func (am *AuthMiddleware) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"auth_attempts":     atomic.LoadInt64(&am.authAttempts),
		"auth_successes":    atomic.LoadInt64(&am.authSuccesses),
		"auth_failures":     atomic.LoadInt64(&am.authFailures),
		"auth_errors":       atomic.LoadInt64(&am.authErrors),
		"token_validations": atomic.LoadInt64(&am.tokenValidations),
		"success_rate":      am.calculateSuccessRate(),
		"uptime":            time.Since(am.startTime),
		"provider_count":    len(am.providers),
	}
}

// initializeProviders initializes authentication providers based on configuration.
func (am *AuthMiddleware) initializeProviders() {
	am.providers = make(map[string]AuthProvider)

	if am.config.EnableBasicAuth {
		provider := NewBasicAuthProvider("basic", am.config.BasicAuthUsers, am.config.BasicAuthRealm)
		am.providers[provider.GetName()] = provider
	}

	if am.config.EnableAPIKeyAuth {
		provider := NewAPIKeyAuthProvider("apikey", am.config.ValidTokens, am.config.APIKeyHeader)
		am.providers[provider.GetName()] = provider
	}
}

// extractRequestInfo extracts relevant information from the request.
func (am *AuthMiddleware) extractRequestInfo(req interface{}) *RequestInfo {
	info := &RequestInfo{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	if httpReq, ok := req.(map[string]interface{}); ok {
		if method, exists := httpReq["method"]; exists {
			if m, ok := method.(string); ok {
				info.Method = m
			}
		}

		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				info.Path = p
			}
		}

		if headers, exists := httpReq["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				info.Headers = h
			}
		}

		if query, exists := httpReq["query"]; exists {
			if q, ok := query.(map[string]string); ok {
				info.Query = q
			}
		}
	}

	return info
}

// RequestInfo holds request information for authentication.
type RequestInfo struct {
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
}

// shouldSkipAuth determines if authentication should be skipped for a path.
func (am *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range am.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// authenticate performs authentication using available providers.
func (am *AuthMiddleware) authenticate(ctx context.Context, reqInfo *RequestInfo) (*AuthContext, error) {
	authCtx := &AuthContext{
		Authenticated: false,
		Timestamp:     time.Now(),
	}

	// Try different authentication methods
	if am.config.EnableBasicAuth {
		if authHeader, exists := reqInfo.Headers[am.config.AuthHeader]; exists {
			if strings.HasPrefix(authHeader, "Basic ") {
				user, err := am.authenticateBasic(ctx, authHeader)
				if err == nil && user != nil {
					authCtx.User = user
					authCtx.AuthType = "basic"
					authCtx.Provider = "basic"
					authCtx.Authenticated = true
					return authCtx, nil
				}
			}
		}
	}

	if am.config.EnableBearerAuth || am.config.EnableAPIKeyAuth {
		if authHeader, exists := reqInfo.Headers[am.config.AuthHeader]; exists {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				user, err := am.authenticateToken(ctx, token)
				if err == nil && user != nil {
					authCtx.User = user
					authCtx.Token = token
					authCtx.AuthType = "bearer"
					authCtx.Provider = "token"
					authCtx.Authenticated = true
					return authCtx, nil
				}
			}
		}

		// Check API key header
		if apiKey, exists := reqInfo.Headers[am.config.APIKeyHeader]; exists {
			user, err := am.authenticateToken(ctx, apiKey)
			if err == nil && user != nil {
				authCtx.User = user
				authCtx.Token = apiKey
				authCtx.AuthType = "apikey"
				authCtx.Provider = "apikey"
				authCtx.Authenticated = true
				return authCtx, nil
			}
		}

		// Check query parameter
		if token, exists := reqInfo.Query[am.config.AuthQueryParam]; exists {
			user, err := am.authenticateToken(ctx, token)
			if err == nil && user != nil {
				authCtx.User = user
				authCtx.Token = token
				authCtx.AuthType = "query"
				authCtx.Provider = "token"
				authCtx.Authenticated = true
				return authCtx, nil
			}
		}
	}

	return authCtx, nil
}

// authenticateBasic performs basic HTTP authentication.
func (am *AuthMiddleware) authenticateBasic(ctx context.Context, authHeader string) (*AuthUser, error) {
	// Decode base64 credentials
	payload := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, errors.New("invalid basic auth encoding")
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, errors.New("invalid basic auth format")
	}

	username, password := credentials[0], credentials[1]

	// Use basic auth provider if available
	if provider, exists := am.providers["basic"]; exists && provider.IsEnabled() {
		return provider.Authenticate(ctx, map[string]string{
			"username": username,
			"password": password,
		})
	}

	// Fallback to config-based authentication
	if expectedPassword, exists := am.config.BasicAuthUsers[username]; exists {
		if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1 {
			return &AuthUser{
				ID:       username,
				Username: username,
				Roles:    []string{"user"},
			}, nil
		}
	}

	return nil, errors.New("invalid credentials")
}

// authenticateToken performs token-based authentication.
func (am *AuthMiddleware) authenticateToken(ctx context.Context, token string) (*AuthUser, error) {
	atomic.AddInt64(&am.tokenValidations, 1)

	// Try each available provider
	for _, provider := range am.providers {
		if !provider.IsEnabled() {
			continue
		}

		user, err := provider.ValidateToken(ctx, token)
		if err == nil && user != nil {
			return user, nil
		}
	}

	// Fallback to config-based validation
	if user, exists := am.config.ValidTokens[token]; exists {
		if !user.ExpiresAt.IsZero() && time.Now().After(user.ExpiresAt) {
			return nil, errors.New("token expired")
		}
		return &user, nil
	}

	return nil, errors.New("invalid token")
}

// createAuthError creates an authentication error response.
func (am *AuthMiddleware) createAuthError(err error) error {
	return fmt.Errorf("authentication error: %w", err)
}

// createUnauthorizedError creates an unauthorized error response.
func (am *AuthMiddleware) createUnauthorizedError() error {
	return errors.New(am.config.UnauthorizedMessage)
}

// calculateSuccessRate calculates the authentication success rate.
func (am *AuthMiddleware) calculateSuccessRate() float64 {
	attempts := atomic.LoadInt64(&am.authAttempts)
	if attempts == 0 {
		return 0.0
	}
	successes := atomic.LoadInt64(&am.authSuccesses)
	return float64(successes) / float64(attempts) * 100.0
}

// Reset resets all metrics.
func (am *AuthMiddleware) Reset() {
	atomic.StoreInt64(&am.authAttempts, 0)
	atomic.StoreInt64(&am.authSuccesses, 0)
	atomic.StoreInt64(&am.authFailures, 0)
	atomic.StoreInt64(&am.authErrors, 0)
	atomic.StoreInt64(&am.tokenValidations, 0)
	am.startTime = time.Now()
	am.GetLogger().Info("Authentication middleware metrics reset")
}

// NewBasicAuthProvider creates a new basic authentication provider.
func NewBasicAuthProvider(name string, users map[string]string, realm string) *BasicAuthProvider {
	return &BasicAuthProvider{
		name:    name,
		enabled: true,
		users:   users,
		realm:   realm,
	}
}

// Authenticate implements AuthProvider interface for basic auth.
func (bap *BasicAuthProvider) Authenticate(ctx context.Context, credentials interface{}) (*AuthUser, error) {
	creds, ok := credentials.(map[string]string)
	if !ok {
		return nil, errors.New("invalid credentials format")
	}

	username, exists := creds["username"]
	if !exists {
		return nil, errors.New("username required")
	}

	password, exists := creds["password"]
	if !exists {
		return nil, errors.New("password required")
	}

	expectedPassword, userExists := bap.users[username]
	if !userExists {
		return nil, errors.New("user not found")
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) != 1 {
		return nil, errors.New("invalid password")
	}

	return &AuthUser{
		ID:       username,
		Username: username,
		Roles:    []string{"user"},
	}, nil
}

// ValidateToken implements AuthProvider interface for basic auth (not applicable).
func (bap *BasicAuthProvider) ValidateToken(ctx context.Context, token string) (*AuthUser, error) {
	return nil, errors.New("token validation not supported for basic auth")
}

// GetName returns the provider name.
func (bap *BasicAuthProvider) GetName() string {
	return bap.name
}

// IsEnabled returns whether the provider is enabled.
func (bap *BasicAuthProvider) IsEnabled() bool {
	return bap.enabled
}

// NewAPIKeyAuthProvider creates a new API key authentication provider.
func NewAPIKeyAuthProvider(name string, tokens map[string]AuthUser, header string) *APIKeyAuthProvider {
	return &APIKeyAuthProvider{
		name:    name,
		enabled: true,
		tokens:  tokens,
		header:  header,
	}
}

// Authenticate implements AuthProvider interface for API key auth.
func (akap *APIKeyAuthProvider) Authenticate(ctx context.Context, credentials interface{}) (*AuthUser, error) {
	return nil, errors.New("credential authentication not supported for API key auth")
}

// ValidateToken implements AuthProvider interface for API key auth.
func (akap *APIKeyAuthProvider) ValidateToken(ctx context.Context, token string) (*AuthUser, error) {
	user, exists := akap.tokens[token]
	if !exists {
		return nil, errors.New("invalid API key")
	}

	if !user.ExpiresAt.IsZero() && time.Now().After(user.ExpiresAt) {
		return nil, errors.New("API key expired")
	}

	return &user, nil
}

// GetName returns the provider name.
func (akap *APIKeyAuthProvider) GetName() string {
	return akap.name
}

// IsEnabled returns whether the provider is enabled.
func (akap *APIKeyAuthProvider) IsEnabled() bool {
	return akap.enabled
}
