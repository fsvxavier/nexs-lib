package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middleware/bodyvalidator"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/contenttype"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/errorhandler"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/tenantid"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/traceid"
)

// User representa um usuário no sistema
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
}

// CreateUserRequest representa uma requisição de criação de usuário
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// APIResponse representa uma resposta padrão da API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
	Tenant  string      `json:"tenant,omitempty"`
}

func main() {
	// Configurar rotas
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/users", usersHandler)
	mux.HandleFunc("/users/create", createUserHandler)
	mux.HandleFunc("/panic", panicHandler) // Para demonstrar error handler

	// Aplicar middlewares (ordem manual para demonstração)
	handler := setupMiddlewares(mux)

	// Configurar servidor
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("🚀 Servidor iniciado em http://localhost:8080")
	fmt.Println("📝 Endpoints disponíveis:")
	fmt.Println("  GET  /health          - Health check")
	fmt.Println("  GET  /users           - Listar usuários")
	fmt.Println("  POST /users/create    - Criar usuário (requer JSON válido)")
	fmt.Println("  GET  /panic           - Demonstrar error handler")
	fmt.Println()
	fmt.Println("🔧 Headers sugeridos para teste:")
	fmt.Println("  X-Tenant-ID: tenant123")
	fmt.Println("  X-Trace-ID: trace-456")
	fmt.Println("  Content-Type: application/json")
	fmt.Println()
	fmt.Println("📘 Exemplo de teste com curl:")
	fmt.Println(`  curl -X POST http://localhost:8080/users/create \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: tenant123" \
    -H "X-Trace-ID: trace-456" \
    -d '{"name":"João Silva","email":"joao@example.com"}'`)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Erro ao iniciar servidor:", err)
	}
}

func setupMiddlewares(handler http.Handler) http.Handler {
	// 5. Body Validator (aplicado primeiro na cadeia de wrapping)
	bodyConfig := bodyvalidator.DefaultConfig()
	bodyConfig.MaxBodySize = 1024 * 1024                 // 1MB máximo
	bodyConfig.SkipPaths = []string{"/health", "/users"} // GET endpoints não precisam de validação
	bodyConfig.SkipMethods = []string{"GET", "HEAD", "OPTIONS"}
	handler = bodyvalidator.NewMiddleware(bodyConfig).Wrap(handler)

	// 4. Content Type (aplicado antes do body validator)
	contentTypeMiddleware := contenttype.CreateJSONOnly("POST", "PUT", "PATCH")
	handler = contentTypeMiddleware.Wrap(handler)

	// 3. Tenant ID
	tenantConfig := tenantid.DefaultConfig()
	tenantConfig.HeaderName = "X-Tenant-ID"
	tenantConfig.ContextKey = "tenant_id"
	tenantConfig.DefaultTenant = "default"
	tenantConfig.Required = false // Permite acesso sem tenant para alguns endpoints
	tenantConfig.CaseSensitive = false
	tenantConfig.SkipPaths = []string{"/health"} // Health check não precisa de tenant
	handler = tenantid.NewMiddleware(tenantConfig).Wrap(handler)

	// 2. Trace ID
	traceConfig := traceid.DefaultConfig()
	traceConfig.HeaderName = "X-Trace-ID"
	traceConfig.AlternativeHeaders = []string{"X-Request-ID", "Request-ID"}
	traceConfig.ContextKey = "trace_id"
	handler = traceid.NewMiddleware(traceConfig).Wrap(handler)

	// 1. Error Handler (aplicado por último para capturar tudo)
	errorConfig := errorhandler.DefaultConfig()
	errorConfig.IncludeStackTrace = true // Para desenvolvimento
	errorConfig.CustomErrorFormatter = func(err error, statusCode int, traceID string) interface{} {
		return map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"status":    statusCode,
			"trace_id":  traceID,
			"timestamp": time.Now().Format(time.RFC3339),
		}
	}
	handler = errorhandler.NewMiddleware(errorConfig).Wrap(handler)

	return handler
}

// healthHandler - Health check simples
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// usersHandler - Listar usuários (demonstração)
func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Extrair dados do contexto
	traceID := traceid.GetTraceIDFromContext(r.Context(), "trace_id")
	tenantID := tenantid.GetTenantIDFromContext(r.Context(), "tenant_id")

	// Dados mockados
	users := []User{
		{ID: "1", Name: "Alice Santos", Email: "alice@example.com", TenantID: tenantID},
		{ID: "2", Name: "Bob Silva", Email: "bob@example.com", TenantID: tenantID},
	}

	response := APIResponse{
		Success: true,
		Data:    users,
		TraceID: traceID,
		Tenant:  tenantID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// createUserHandler - Criar usuário (demonstra validação de body)
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extrair dados do contexto
	traceID := traceid.GetTraceIDFromContext(r.Context(), "trace_id")
	tenantID := tenantid.GetTenantIDFromContext(r.Context(), "tenant_id")

	// Fazer parse do JSON (já validado pelo bodyvalidator)
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validações de negócio
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Criar usuário (simulado)
	user := User{
		ID:       fmt.Sprintf("user_%d", time.Now().Unix()),
		Name:     req.Name,
		Email:    req.Email,
		TenantID: tenantID,
	}

	response := APIResponse{
		Success: true,
		Data:    user,
		TraceID: traceID,
		Tenant:  tenantID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// panicHandler - Demonstra o error handler capturando panics
func panicHandler(w http.ResponseWriter, r *http.Request) {
	// Simular um panic para demonstrar o error handler
	panic("Ops! Algo deu muito errado aqui!")
}
