package middlewares

import (
	"context"
	"encoding/base64"
	"testing"
	"time"
)

func TestNewAuthMiddleware(t *testing.T) {
	middleware := NewAuthMiddleware(1)

	if middleware == nil {
		t.Fatal("Expected auth middleware to be created")
	}

	if middleware.Name() != "auth" {
		t.Errorf("Expected name to be 'auth', got %s", middleware.Name())
	}

	if middleware.Priority() != 1 {
		t.Errorf("Expected priority to be 1, got %d", middleware.Priority())
	}

	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled by default")
	}

	config := middleware.GetConfig()
	if !config.EnableBasicAuth {
		t.Error("Expected EnableBasicAuth to be true by default")
	}

	if !config.EnableBearerAuth {
		t.Error("Expected EnableBearerAuth to be true by default")
	}

	if !config.RequireAuth {
		t.Error("Expected RequireAuth to be true by default")
	}
}

func TestNewAuthMiddlewareWithConfig(t *testing.T) {
	customConfig := AuthConfig{
		EnableBasicAuth:     false,
		EnableBearerAuth:    true,
		EnableAPIKeyAuth:    true,
		RequireAuth:         false,
		BasicAuthRealm:      "Test Realm",
		UnauthorizedMessage: "Custom unauthorized message",
	}

	middleware := NewAuthMiddlewareWithConfig(2, customConfig)

	if middleware == nil {
		t.Fatal("Expected auth middleware to be created")
	}

	config := middleware.GetConfig()
	if config.EnableBasicAuth {
		t.Error("Expected EnableBasicAuth to be false")
	}

	if !config.EnableBearerAuth {
		t.Error("Expected EnableBearerAuth to be true")
	}

	if !config.EnableAPIKeyAuth {
		t.Error("Expected EnableAPIKeyAuth to be true")
	}

	if config.RequireAuth {
		t.Error("Expected RequireAuth to be false")
	}

	if config.BasicAuthRealm != "Test Realm" {
		t.Errorf("Expected BasicAuthRealm to be 'Test Realm', got %s", config.BasicAuthRealm)
	}

	if config.UnauthorizedMessage != "Custom unauthorized message" {
		t.Errorf("Expected custom unauthorized message, got %s", config.UnauthorizedMessage)
	}
}

func TestDefaultAuthConfig(t *testing.T) {
	config := DefaultAuthConfig()

	if !config.EnableBasicAuth {
		t.Error("Expected EnableBasicAuth to be true by default")
	}

	if !config.EnableBearerAuth {
		t.Error("Expected EnableBearerAuth to be true by default")
	}

	if config.EnableAPIKeyAuth {
		t.Error("Expected EnableAPIKeyAuth to be false by default")
	}

	if config.EnableJWTAuth {
		t.Error("Expected EnableJWTAuth to be false by default")
	}

	if !config.RequireAuth {
		t.Error("Expected RequireAuth to be true by default")
	}

	if config.BasicAuthRealm != "Restricted Area" {
		t.Errorf("Expected BasicAuthRealm to be 'Restricted Area', got %s", config.BasicAuthRealm)
	}

	if config.AuthHeader != "Authorization" {
		t.Errorf("Expected AuthHeader to be 'Authorization', got %s", config.AuthHeader)
	}

	expectedSkipPaths := []string{"/health", "/metrics", "/favicon.ico"}
	if len(config.SkipPaths) != len(expectedSkipPaths) {
		t.Errorf("Expected %d skip paths, got %d", len(expectedSkipPaths), len(config.SkipPaths))
	}

	if config.MaxLoginAttempts != 5 {
		t.Errorf("Expected MaxLoginAttempts to be 5, got %d", config.MaxLoginAttempts)
	}

	if config.UnauthorizedMessage != "Authentication required" {
		t.Errorf("Expected default unauthorized message, got %s", config.UnauthorizedMessage)
	}
}

func TestAuthMiddleware_ProcessWithoutAuth(t *testing.T) {
	config := DefaultAuthConfig()
	config.RequireAuth = false
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/test",
		"headers": map[string]string{
			"User-Agent": "test-agent",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		return map[string]interface{}{"status": "ok"}, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Error("Expected response to not be nil")
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	metrics := middleware.GetMetrics()
	if metrics["auth_attempts"].(int64) != 1 {
		t.Errorf("Expected 1 auth attempt, got %v", metrics["auth_attempts"])
	}
}

func TestAuthMiddleware_ProcessWithSkippedPath(t *testing.T) {
	middleware := NewAuthMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method":  "GET",
		"path":    "/health",
		"headers": map[string]string{},
	}

	nextCalled := false
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		return map[string]interface{}{"status": "ok"}, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Error("Expected response to not be nil")
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}
}

func TestAuthMiddleware_ProcessBasicAuthSuccess(t *testing.T) {
	config := DefaultAuthConfig()
	config.BasicAuthUsers = map[string]string{
		"testuser": "testpass",
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	// Create basic auth header
	credentials := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
	authHeader := "Basic " + credentials

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/protected",
		"headers": map[string]string{
			"Authorization": authHeader,
		},
	}

	nextCalled := false
	var authCtx *AuthContext
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		if auth := ctx.Value("auth"); auth != nil {
			authCtx = auth.(*AuthContext)
		}
		return map[string]interface{}{"status": "ok"}, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Error("Expected response to not be nil")
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	if authCtx == nil {
		t.Error("Expected auth context to be set")
	} else {
		if !authCtx.Authenticated {
			t.Error("Expected user to be authenticated")
		}

		if authCtx.User.Username != "testuser" {
			t.Errorf("Expected username to be 'testuser', got %s", authCtx.User.Username)
		}

		if authCtx.AuthType != "basic" {
			t.Errorf("Expected auth type to be 'basic', got %s", authCtx.AuthType)
		}
	}

	metrics := middleware.GetMetrics()
	if metrics["auth_successes"].(int64) != 1 {
		t.Errorf("Expected 1 auth success, got %v", metrics["auth_successes"])
	}
}

func TestAuthMiddleware_ProcessBasicAuthFailure(t *testing.T) {
	config := DefaultAuthConfig()
	config.BasicAuthUsers = map[string]string{
		"testuser": "testpass",
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	// Create basic auth header with wrong password
	credentials := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass"))
	authHeader := "Basic " + credentials

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/protected",
		"headers": map[string]string{
			"Authorization": authHeader,
		},
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Error("Expected next function to not be called")
		return nil, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err == nil {
		t.Error("Expected authentication error")
	}

	if err.Error() != "Authentication required" {
		t.Errorf("Expected 'Authentication required' error, got %v", err)
	}

	metrics := middleware.GetMetrics()
	if metrics["auth_failures"].(int64) != 1 {
		t.Errorf("Expected 1 auth failure, got %v", metrics["auth_failures"])
	}
}

func TestAuthMiddleware_ProcessBearerTokenSuccess(t *testing.T) {
	config := DefaultAuthConfig()
	config.ValidTokens = map[string]AuthUser{
		"valid-token": {
			ID:       "user1",
			Username: "tokenuser",
			Roles:    []string{"admin"},
		},
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/protected",
		"headers": map[string]string{
			"Authorization": "Bearer valid-token",
		},
	}

	nextCalled := false
	var authCtx *AuthContext
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		if auth := ctx.Value("auth"); auth != nil {
			authCtx = auth.(*AuthContext)
		}
		return map[string]interface{}{"status": "ok"}, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Error("Expected response to not be nil")
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	if authCtx == nil {
		t.Error("Expected auth context to be set")
	} else {
		if !authCtx.Authenticated {
			t.Error("Expected user to be authenticated")
		}

		if authCtx.User.Username != "tokenuser" {
			t.Errorf("Expected username to be 'tokenuser', got %s", authCtx.User.Username)
		}

		if authCtx.AuthType != "bearer" {
			t.Errorf("Expected auth type to be 'bearer', got %s", authCtx.AuthType)
		}

		if authCtx.Token != "valid-token" {
			t.Errorf("Expected token to be 'valid-token', got %s", authCtx.Token)
		}
	}

	metrics := middleware.GetMetrics()
	if metrics["token_validations"].(int64) != 1 {
		t.Errorf("Expected 1 token validation, got %v", metrics["token_validations"])
	}
}

func TestAuthMiddleware_ProcessExpiredToken(t *testing.T) {
	config := DefaultAuthConfig()
	config.ValidTokens = map[string]AuthUser{
		"expired-token": {
			ID:        "user1",
			Username:  "tokenuser",
			ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
		},
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/protected",
		"headers": map[string]string{
			"Authorization": "Bearer expired-token",
		},
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Error("Expected next function to not be called")
		return nil, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err == nil {
		t.Error("Expected authentication error for expired token")
	}
}

func TestAuthMiddleware_ProcessAPIKeyAuth(t *testing.T) {
	config := DefaultAuthConfig()
	config.EnableAPIKeyAuth = true
	config.ValidTokens = map[string]AuthUser{
		"api-key-123": {
			ID:       "api1",
			Username: "apiuser",
			Roles:    []string{"api"},
		},
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/api/data",
		"headers": map[string]string{
			"X-API-Key": "api-key-123",
		},
	}

	nextCalled := false
	var authCtx *AuthContext
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		if auth := ctx.Value("auth"); auth != nil {
			authCtx = auth.(*AuthContext)
		}
		return map[string]interface{}{"status": "ok"}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	if authCtx == nil {
		t.Error("Expected auth context to be set")
	} else {
		if !authCtx.Authenticated {
			t.Error("Expected user to be authenticated")
		}

		if authCtx.AuthType != "apikey" {
			t.Errorf("Expected auth type to be 'apikey', got %s", authCtx.AuthType)
		}
	}
}

func TestAuthMiddleware_ProcessQueryTokenAuth(t *testing.T) {
	config := DefaultAuthConfig()
	config.ValidTokens = map[string]AuthUser{
		"query-token": {
			ID:       "query1",
			Username: "queryuser",
		},
	}
	middleware := NewAuthMiddlewareWithConfig(1, config)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := map[string]interface{}{
		"method":  "GET",
		"path":    "/api/data",
		"headers": map[string]string{},
		"query": map[string]string{
			"token": "query-token",
		},
	}

	nextCalled := false
	var authCtx *AuthContext
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		if auth := ctx.Value("auth"); auth != nil {
			authCtx = auth.(*AuthContext)
		}
		return map[string]interface{}{"status": "ok"}, nil
	}

	_, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	if authCtx == nil {
		t.Error("Expected auth context to be set")
	} else {
		if !authCtx.Authenticated {
			t.Error("Expected user to be authenticated")
		}

		if authCtx.AuthType != "query" {
			t.Errorf("Expected auth type to be 'query', got %s", authCtx.AuthType)
		}
	}
}

func TestAuthMiddleware_AddProvider(t *testing.T) {
	middleware := NewAuthMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	provider := NewBasicAuthProvider("test-provider", map[string]string{
		"user": "pass",
	}, "Test Realm")

	err := middleware.AddProvider(provider)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test adding nil provider
	err = middleware.AddProvider(nil)
	if err == nil {
		t.Error("Expected error for nil provider")
	}

	// Test getting provider
	retrievedProvider, err := middleware.GetProvider("test-provider")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedProvider != provider {
		t.Error("Expected to get the same provider instance")
	}
}

func TestAuthMiddleware_RemoveProvider(t *testing.T) {
	middleware := NewAuthMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	provider := NewBasicAuthProvider("test-provider", map[string]string{}, "Test")
	middleware.AddProvider(provider)

	// Test removing existing provider
	err := middleware.RemoveProvider("test-provider")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test removing non-existent provider
	err = middleware.RemoveProvider("non-existent")
	if err == nil {
		t.Error("Expected error for removing non-existent provider")
	}
}

func TestAuthMiddleware_SetConfig(t *testing.T) {
	middleware := NewAuthMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	newConfig := AuthConfig{
		EnableBasicAuth:     false,
		EnableBearerAuth:    false,
		EnableAPIKeyAuth:    true,
		RequireAuth:         false,
		UnauthorizedMessage: "Custom message",
	}

	middleware.SetConfig(newConfig)

	config := middleware.GetConfig()
	if config.EnableBasicAuth {
		t.Error("Expected EnableBasicAuth to be false after config update")
	}

	if config.EnableAPIKeyAuth != true {
		t.Error("Expected EnableAPIKeyAuth to be true after config update")
	}

	if config.UnauthorizedMessage != "Custom message" {
		t.Errorf("Expected custom message, got %s", config.UnauthorizedMessage)
	}

	// Check that a log message was written about config update
	if len(testLogger.InfoMessages) == 0 {
		t.Error("Expected info message about config update")
	}
}

func TestAuthMiddleware_Reset(t *testing.T) {
	middleware := NewAuthMiddleware(1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	// Simulate some activity
	ctx := context.Background()
	req := map[string]interface{}{
		"method":  "GET",
		"path":    "/test",
		"headers": map[string]string{},
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"status": "ok"}, nil
	}

	middleware.Process(ctx, req, next)

	// Check that metrics are not zero
	metrics := middleware.GetMetrics()
	if metrics["auth_attempts"].(int64) == 0 {
		t.Error("Expected auth_attempts to be greater than 0 before reset")
	}

	// Reset metrics
	middleware.Reset()

	// Check that metrics are reset
	metrics = middleware.GetMetrics()
	if metrics["auth_attempts"].(int64) != 0 {
		t.Errorf("Expected auth_attempts to be 0 after reset, got %v", metrics["auth_attempts"])
	}

	if metrics["auth_successes"].(int64) != 0 {
		t.Errorf("Expected auth_successes to be 0 after reset, got %v", metrics["auth_successes"])
	}
}

func TestBasicAuthProvider_Authenticate(t *testing.T) {
	users := map[string]string{
		"testuser": "testpass",
		"admin":    "secret",
	}
	provider := NewBasicAuthProvider("basic", users, "Test Realm")

	// Test successful authentication
	user, err := provider.Authenticate(context.Background(), map[string]string{
		"username": "testuser",
		"password": "testpass",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to not be nil")
	} else {
		if user.Username != "testuser" {
			t.Errorf("Expected username to be 'testuser', got %s", user.Username)
		}
	}

	// Test failed authentication
	_, err = provider.Authenticate(context.Background(), map[string]string{
		"username": "testuser",
		"password": "wrongpass",
	})

	if err == nil {
		t.Error("Expected error for wrong password")
	}

	// Test invalid credentials format
	_, err = provider.Authenticate(context.Background(), "invalid")
	if err == nil {
		t.Error("Expected error for invalid credentials format")
	}
}

func TestAPIKeyAuthProvider_ValidateToken(t *testing.T) {
	tokens := map[string]AuthUser{
		"valid-key": {
			ID:       "user1",
			Username: "apiuser",
		},
		"expired-key": {
			ID:        "user2",
			Username:  "expireduser",
			ExpiresAt: time.Now().Add(-time.Hour),
		},
	}
	provider := NewAPIKeyAuthProvider("apikey", tokens, "X-API-Key")

	// Test valid token
	user, err := provider.ValidateToken(context.Background(), "valid-key")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to not be nil")
	} else {
		if user.Username != "apiuser" {
			t.Errorf("Expected username to be 'apiuser', got %s", user.Username)
		}
	}

	// Test invalid token
	_, err = provider.ValidateToken(context.Background(), "invalid-key")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Test expired token
	_, err = provider.ValidateToken(context.Background(), "expired-key")
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestAuthMiddleware_CalculateSuccessRate(t *testing.T) {
	middleware := NewAuthMiddleware(1)

	// Test with no attempts
	rate := middleware.calculateSuccessRate()
	if rate != 0.0 {
		t.Errorf("Expected success rate to be 0.0 with no attempts, got %f", rate)
	}

	// Simulate some attempts and successes
	middleware.authAttempts = 10
	middleware.authSuccesses = 7

	rate = middleware.calculateSuccessRate()
	expectedRate := 70.0
	if rate != expectedRate {
		t.Errorf("Expected success rate to be %f, got %f", expectedRate, rate)
	}
}
