package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// User represents a user in the system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UserService simulates a user service with error handling
type UserService struct {
	users map[int]*User
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[int]*User),
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id int) (*User, error) {
	if id <= 0 {
		return nil, domainerrors.NewValidationError("INVALID_USER_ID", "Invalid user ID", nil).
			WithField("id", "must be positive integer")
	}

	// Simulate database timeout
	if id == 999 {
		return nil, domainerrors.NewTimeoutError("DB_TIMEOUT", "database-query",
			"Database query timeout", errors.New("context deadline exceeded")).
			WithDuration(5*time.Second, 3*time.Second)
	}

	// Simulate external service error
	if id == 888 {
		return nil, domainerrors.NewExternalServiceError("USER_PROFILE_API_ERROR", "user-profile-api",
			"Failed to fetch user profile", errors.New("service unavailable")).
			WithEndpoint("/api/v1/profiles/888").
			WithResponse(503, "Service temporarily unavailable")
	}

	user, exists := s.users[id]
	if !exists {
		return nil, domainerrors.NewNotFoundError("USER_NOT_FOUND", "User not found", nil).
			WithResource("user", strconv.Itoa(id))
	}

	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(user *User) error {
	if user == nil {
		return domainerrors.NewValidationError("INVALID_USER_DATA", "User data is required", nil).
			WithField("user", "cannot be null")
	}

	// Validate user data
	validationErr := s.validateUser(user)
	if validationErr != nil {
		return validationErr
	}

	// Check business rules
	if user.Age < 18 {
		return domainerrors.NewBusinessError("UNDERAGE_USER", "User must be 18 or older").
			WithRule("Minimum age requirement: 18 years")
	}

	// Check if email already exists
	for _, existingUser := range s.users {
		if existingUser.Email == user.Email {
			return domainerrors.NewConflictError("EMAIL_ALREADY_EXISTS", "Email already exists").
				WithConflictingResource("user", "email address already in use")
		}
	}

	// Simulate database error
	if user.ID == 777 {
		return domainerrors.NewDatabaseError("DB_INSERT_FAILED", "Failed to insert user",
			errors.New("unique constraint violation")).
			WithOperation("INSERT", "users").
			WithQuery("INSERT INTO users (name, email, age) VALUES (?, ?, ?)")
	}

	s.users[user.ID] = user
	return nil
}

// validateUser validates user data
func (s *UserService) validateUser(user *User) error {
	validationErr := domainerrors.NewValidationError("USER_VALIDATION_FAILED", "User validation failed", nil)
	hasErrors := false

	if user.Name == "" {
		validationErr.WithField("name", "name is required")
		hasErrors = true
	} else if len(user.Name) < 2 {
		validationErr.WithField("name", "name must be at least 2 characters")
		hasErrors = true
	}

	if user.Email == "" {
		validationErr.WithField("email", "email is required")
		hasErrors = true
	} else if !isValidEmail(user.Email) {
		validationErr.WithField("email", "invalid email format")
		hasErrors = true
	}

	if user.Age < 0 {
		validationErr.WithField("age", "age must be non-negative")
		hasErrors = true
	} else if user.Age > 150 {
		validationErr.WithField("age", "age must be realistic (max 150)")
		hasErrors = true
	}

	if hasErrors {
		return validationErr
	}

	return nil
}

// isValidEmail simulates email validation
func isValidEmail(email string) bool {
	return len(email) > 5 && email[0] != '@' && email[len(email)-1] != '@'
}

// HTTP Handlers

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// handleError converts domain errors to HTTP responses
func handleError(w http.ResponseWriter, err error) {
	var response ErrorResponse
	var statusCode int

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		statusCode = domainErr.HTTPStatus()
		response = ErrorResponse{
			Error:     "domain_error",
			Code:      domainErr.Code,
			Message:   domainErr.Message,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// Add specific details based on error type
		details := make(map[string]interface{})

		switch domainErr.Type {
		case domainerrors.ErrorTypeValidation:
			if validationErr, ok := err.(*domainerrors.ValidationError); ok {
				details["fields"] = validationErr.Fields
			}
		case domainerrors.ErrorTypeBusinessRule:
			if businessErr, ok := err.(*domainerrors.BusinessError); ok {
				details["business_code"] = businessErr.BusinessCode
				details["rules"] = businessErr.Rules
			}
		case domainerrors.ErrorTypeNotFound:
			if notFoundErr, ok := err.(*domainerrors.NotFoundError); ok {
				details["resource_type"] = notFoundErr.ResourceType
				details["resource_id"] = notFoundErr.ResourceID
			}
		case domainerrors.ErrorTypeConflict:
			if conflictErr, ok := err.(*domainerrors.ConflictError); ok {
				details["resource"] = conflictErr.Resource
				details["conflict_reason"] = conflictErr.ConflictReason
			}
		case domainerrors.ErrorTypeTimeout:
			if timeoutErr, ok := err.(*domainerrors.TimeoutError); ok {
				details["operation"] = timeoutErr.Operation
				details["duration"] = timeoutErr.Duration.String()
				details["timeout"] = timeoutErr.Timeout.String()
			}
		case domainerrors.ErrorTypeExternalService:
			if extErr, ok := err.(*domainerrors.ExternalServiceError); ok {
				details["service"] = extErr.Service
				details["endpoint"] = extErr.Endpoint
				details["status_code"] = extErr.StatusCode
			}
		}

		if len(details) > 0 {
			response.Details = details
		}
	} else {
		// Handle standard errors
		statusCode = http.StatusInternalServerError
		response = ErrorResponse{
			Error:     "internal_error",
			Code:      "INTERNAL_SERVER_ERROR",
			Message:   "An internal server error occurred",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// getUserHandler handles GET /users/{id}
func (s *UserService) getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(w, domainerrors.NewValidationError("INVALID_USER_ID", "Invalid user ID format", err).
			WithField("id", "must be numeric"))
		return
	}

	user, err := s.GetUser(id)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// createUserHandler handles POST /users
func (s *UserService) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, domainerrors.NewValidationError("INVALID_JSON", "Invalid JSON format", err).
			WithField("body", "must be valid JSON"))
		return
	}

	if err := s.CreateUser(&user); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// simulateRateLimitHandler simulates rate limiting
func simulateRateLimitHandler(w http.ResponseWriter, r *http.Request) {
	rateLimitErr := domainerrors.NewRateLimitError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	rateLimitErr.WithRateLimit(100, 0, time.Now().Add(time.Hour).Format(time.RFC3339), "3600s")

	handleError(w, rateLimitErr)
}

// simulateAuthErrorHandler simulates authentication error
func simulateAuthErrorHandler(w http.ResponseWriter, r *http.Request) {
	authErr := domainerrors.NewAuthenticationError("INVALID_TOKEN", "Invalid authentication token",
		errors.New("token expired"))
	authErr.WithScheme("Bearer")

	handleError(w, authErr)
}

// simulateServerErrorHandler simulates server error
func simulateServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	serverErr := domainerrors.NewServerError("INTERNAL_SERVER_ERROR", "Internal server error",
		errors.New("database connection failed"))
	serverErr.WithRequestInfo("req-12345", "trace-67890")

	handleError(w, serverErr)
}

func main() {
	fmt.Println("=== Domain Errors - HTTP Integration Example ===")
	fmt.Println()

	userService := NewUserService()

	// Add some sample users
	userService.users[1] = &User{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30}
	userService.users[2] = &User{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 25}

	// Setup HTTP routes
	http.HandleFunc("/users/", userService.getUserHandler)
	http.HandleFunc("/users", userService.createUserHandler)
	http.HandleFunc("/rate-limit", simulateRateLimitHandler)
	http.HandleFunc("/auth-error", simulateAuthErrorHandler)
	http.HandleFunc("/server-error", simulateServerErrorHandler)

	// Root handler with instructions
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		instructions := `
Domain Errors HTTP Integration Example

Available endpoints:
- GET /users/1          - Get existing user
- GET /users/999        - Simulate timeout error
- GET /users/888        - Simulate external service error
- GET /users/404        - Simulate not found error
- POST /users           - Create user (try with invalid data)
- GET /rate-limit       - Simulate rate limit error
- GET /auth-error       - Simulate authentication error
- GET /server-error     - Simulate server error

Example requests:
curl http://localhost:8080/users/1
curl http://localhost:8080/users/999
curl -X POST http://localhost:8080/users -d '{"name":"","email":"invalid","age":-1}'
curl http://localhost:8080/rate-limit

All errors are properly mapped to HTTP status codes and include detailed information.
`
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(instructions))
	})

	fmt.Println("Server starting on :8080")
	fmt.Println("Visit http://localhost:8080 for usage instructions")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
