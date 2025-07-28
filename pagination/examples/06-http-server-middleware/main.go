package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/middleware"
)

// User represents a sample user entity
type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// Sample data
var users = []User{
	{1, "Alice Johnson", "alice@example.com", "2024-01-01T10:00:00Z"},
	{2, "Bob Smith", "bob@example.com", "2024-01-02T10:00:00Z"},
	{3, "Charlie Brown", "charlie@example.com", "2024-01-03T10:00:00Z"},
	{4, "Diana Wilson", "diana@example.com", "2024-01-04T10:00:00Z"},
	{5, "Edward Davis", "edward@example.com", "2024-01-05T10:00:00Z"},
	{6, "Fiona Garcia", "fiona@example.com", "2024-01-06T10:00:00Z"},
	{7, "George Miller", "george@example.com", "2024-01-07T10:00:00Z"},
	{8, "Helen Taylor", "helen@example.com", "2024-01-08T10:00:00Z"},
	{9, "Ivan Rodriguez", "ivan@example.com", "2024-01-09T10:00:00Z"},
	{10, "Julia Martinez", "julia@example.com", "2024-01-10T10:00:00Z"},
}

func main() {
	// Create pagination configuration with custom hooks
	paginationConfig := middleware.DefaultPaginationConfig()

	// Configure routes with specific sortable fields
	paginationConfig.ConfigureRoute("/api/users", []string{"id", "name", "email", "created_at"})
	paginationConfig.ConfigureRoute("/api/posts", []string{"id", "title", "created_at"})

	// Add custom hooks
	paginationConfig.WithHooks().
		PreValidation(NewLoggingHook("pre-validation")).
		PostValidation(NewLoggingHook("post-validation")).
		PreQuery(NewLoggingHook("pre-query")).
		PostQuery(NewLoggingHook("post-query")).
		Done()

	// Create HTTP server with pagination middleware
	mux := http.NewServeMux()

	// Add the pagination middleware
	paginatedMux := middleware.PaginationMiddleware(paginationConfig)(mux)

	// Add routes
	mux.HandleFunc("/api/users", usersHandler)
	mux.HandleFunc("/api/users/paginated", paginatedUsersHandler(paginationConfig.Service))
	mux.HandleFunc("/health", healthHandler)

	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("\n📖 Example URLs:")
	fmt.Println("   http://localhost:8080/api/users?page=1&limit=3")
	fmt.Println("   http://localhost:8080/api/users?page=2&limit=5&sort=name&order=ASC")
	fmt.Println("   http://localhost:8080/api/users/paginated?page=1&limit=4&sort=email&order=DESC")
	fmt.Println("   http://localhost:8080/health")

	log.Fatal(http.ListenAndServe(":8080", paginatedMux))
}

// usersHandler demonstrates manual pagination
func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters from middleware
	params := middleware.GetPaginationParams(r)
	if params == nil {
		http.Error(w, "Pagination parameters not found", http.StatusInternalServerError)
		return
	}

	// Apply pagination logic
	paginatedUsers := applyPagination(users, params)

	// Set total count header for middleware
	w.Header().Set("X-Total-Count", strconv.Itoa(len(users)))
	w.Header().Set("Content-Type", "application/json")

	// Return raw data (middleware will wrap it in pagination format)
	json.NewEncoder(w).Encode(paginatedUsers)
}

// paginatedUsersHandler demonstrates using pagination service directly
func paginatedUsersHandler(service *pagination.PaginationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters
		params, err := service.ParseRequest(r.URL.Query(), "id", "name", "email", "created_at")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Apply pagination
		paginatedUsers := applyPagination(users, params)

		// Create paginated response
		response := service.CreateResponse(paginatedUsers, params, len(users))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// healthHandler for health checks (skipped by pagination middleware)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// applyPagination applies pagination logic to the user slice
func applyPagination(allUsers []User, params *interfaces.PaginationParams) []User {
	// Apply sorting
	sortedUsers := make([]User, len(allUsers))
	copy(sortedUsers, allUsers)

	// Simple sorting implementation (in real app, use database)
	if params.SortField != "" {
		// For demo purposes, only implement basic sorting
		// In production, this would be done at the database level
	}

	// Apply pagination
	start := (params.Page - 1) * params.Limit
	end := start + params.Limit

	if start >= len(sortedUsers) {
		return []User{}
	}

	if end > len(sortedUsers) {
		end = len(sortedUsers)
	}

	return sortedUsers[start:end]
}

// NewLoggingHook creates a hook that logs pagination operations
func NewLoggingHook(stage string) interfaces.Hook {
	return &LoggingHook{stage: stage}
}

// LoggingHook implements the Hook interface for logging
type LoggingHook struct {
	stage string
}

func (h *LoggingHook) Execute(ctx context.Context, data interface{}) error {
	fmt.Printf("🔗 Hook executed: %s - Data: %T\n", h.stage, data)
	return nil
}
