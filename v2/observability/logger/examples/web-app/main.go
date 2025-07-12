// Package main demonstra o uso do logger em aplicações web
// Este exemplo mostra como integrar logging em aplicações HTTP,
// middleware de logging, e logging de requisições/respostas.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// WebAppLogger encapsula funcionalidades de logging para web apps
type WebAppLogger struct {
	logger interfaces.Logger
}

// RequestInfo representa informações de uma requisição HTTP
type RequestInfo struct {
	Method     string
	Path       string
	RemoteAddr string
	UserAgent  string
	RequestID  string
	UserID     string
	StartTime  time.Time
}

// ResponseInfo representa informações de uma resposta HTTP
type ResponseInfo struct {
	StatusCode    int
	ContentLength int64
	Duration      time.Duration
}

func NewWebAppLogger(factory *logger.Factory) (*WebAppLogger, error) {
	config := logger.DefaultConfig()
	config.ServiceName = "web-app"
	config.ServiceVersion = "v1.0.0"
	config.Format = interfaces.JSONFormat
	config.Level = interfaces.InfoLevel
	config.AddCaller = true

	// Configuração otimizada para aplicações web
	config.GlobalFields = map[string]interface{}{
		"application": "e-commerce-web",
		"component":   "http-server",
	}

	logger, err := factory.CreateLogger("web-app", config)
	if err != nil {
		return nil, err
	}

	return &WebAppLogger{logger: logger}, nil
}

// LoggingMiddleware middleware para logging de requisições HTTP
func (wal *WebAppLogger) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Gera request ID único
		requestID := generateRequestID()

		// Extrai user ID do header (simulado)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		// Cria contexto enriquecido
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		ctx = context.WithValue(ctx, "user_id", userID)
		r = r.WithContext(ctx)

		// Informações da requisição
		reqInfo := RequestInfo{
			Method:     r.Method,
			Path:       r.URL.Path,
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			RequestID:  requestID,
			UserID:     userID,
			StartTime:  start,
		}

		// Logger específico da requisição
		reqLogger := wal.logger.WithFields(
			interfaces.String("request_id", requestID),
			interfaces.String("user_id", userID),
			interfaces.String("method", reqInfo.Method),
			interfaces.String("path", reqInfo.Path),
			interfaces.String("remote_addr", reqInfo.RemoteAddr),
			interfaces.String("user_agent", reqInfo.UserAgent),
		)

		// Log de entrada da requisição
		reqLogger.Info(ctx, "Requisição recebida")

		// Wrapper da resposta para capturar informações
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default
		}

		// Executa o handler
		next.ServeHTTP(wrappedWriter, r)

		// Informações da resposta
		duration := time.Since(start)
		respInfo := ResponseInfo{
			StatusCode:    wrappedWriter.statusCode,
			ContentLength: wrappedWriter.contentLength,
			Duration:      duration,
		}

		// Log de saída da requisição
		wal.logResponse(ctx, reqLogger, reqInfo, respInfo)
	})
}

// logResponse faz log da resposta com base no status code
func (wal *WebAppLogger) logResponse(ctx context.Context, reqLogger interfaces.Logger, req RequestInfo, resp ResponseInfo) {
	fields := []interfaces.Field{
		interfaces.Int("status_code", resp.StatusCode),
		interfaces.Duration("response_time", resp.Duration),
		interfaces.Int64("content_length", resp.ContentLength),
	}

	message := "Requisição processada"

	// Log baseado no status code
	switch {
	case resp.StatusCode >= 500:
		reqLogger.Error(ctx, message, fields...)
	case resp.StatusCode >= 400:
		reqLogger.Warn(ctx, message, fields...)
	case resp.StatusCode >= 300:
		reqLogger.Info(ctx, message, fields...)
	default:
		reqLogger.Info(ctx, message, fields...)
	}

	// Log adicional para operações lentas
	if resp.Duration > 1*time.Second {
		reqLogger.Warn(ctx, "Operação lenta detectada",
			interfaces.Duration("threshold", 1*time.Second),
			interfaces.String("performance_impact", "high"),
		)
	}
}

// responseWriter wrapper para capturar informações da resposta
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int64
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.contentLength += int64(n)
	return n, err
}

// Handlers de exemplo
func (wal *WebAppLogger) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id").(string)
	userID := ctx.Value("user_id").(string)

	handlerLogger := wal.logger.WithFields(
		interfaces.String("request_id", requestID),
		interfaces.String("user_id", userID),
		interfaces.String("handler", "users"),
	)

	switch r.Method {
	case "GET":
		wal.handleGetUsers(ctx, handlerLogger, w, r)
	case "POST":
		wal.handleCreateUser(ctx, handlerLogger, w, r)
	default:
		handlerLogger.Warn(ctx, "Método HTTP não suportado",
			interfaces.String("method", r.Method),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (wal *WebAppLogger) handleGetUsers(ctx context.Context, logger interfaces.Logger, w http.ResponseWriter, r *http.Request) {
	// Simulação de parâmetros de consulta
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	logger.Debug(ctx, "Parâmetros de consulta processados",
		interfaces.Int("page", page),
		interfaces.Int("limit", limit),
	)

	// Simulação de consulta ao banco de dados
	start := time.Now()
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Simula latência do DB
	dbDuration := time.Since(start)

	logger.Info(ctx, "Consulta ao banco de dados executada",
		interfaces.String("operation", "SELECT users"),
		interfaces.Duration("db_duration", dbDuration),
		interfaces.Int("rows_returned", limit),
	)

	// Simulação de resposta
	users := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		users[i] = map[string]interface{}{
			"id":    i + ((page - 1) * limit),
			"name":  fmt.Sprintf("User %d", i+1),
			"email": fmt.Sprintf("user%d@example.com", i+1),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"page":  page,
		"limit": limit,
	})

	logger.Info(ctx, "Lista de usuários retornada",
		interfaces.Int("user_count", len(users)),
	)
}

func (wal *WebAppLogger) handleCreateUser(ctx context.Context, logger interfaces.Logger, w http.ResponseWriter, r *http.Request) {
	// Parsing do body
	var userData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		logger.Error(ctx, "Erro ao fazer parse do JSON",
			interfaces.ErrorNamed("parse_error", err),
			interfaces.String("content_type", r.Header.Get("Content-Type")),
		)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logger.Debug(ctx, "Dados do usuário recebidos",
		interfaces.String("email", fmt.Sprintf("%v", userData["email"])),
		interfaces.Bool("has_name", userData["name"] != nil),
	)

	// Validação simulada
	if userData["email"] == nil {
		logger.Warn(ctx, "Validação falhou",
			interfaces.String("field", "email"),
			interfaces.String("error", "required_field_missing"),
		)
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Simulação de criação no banco
	start := time.Now()
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond) // Simula latência
	dbDuration := time.Since(start)

	newUserID := rand.Intn(10000) + 1000

	logger.Info(ctx, "Usuário criado no banco de dados",
		interfaces.String("operation", "INSERT user"),
		interfaces.Duration("db_duration", dbDuration),
		interfaces.Int("new_user_id", newUserID),
	)

	// Resposta de sucesso
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      newUserID,
		"email":   userData["email"],
		"name":    userData["name"],
		"created": time.Now().Format(time.RFC3339),
	})

	logger.Info(ctx, "Usuário criado com sucesso",
		interfaces.Int("user_id", newUserID),
		interfaces.String("email", fmt.Sprintf("%v", userData["email"])),
	)
}

func (wal *WebAppLogger) handleOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctx.Value("request_id").(string)
	userID := ctx.Value("user_id").(string)

	handlerLogger := wal.logger.WithFields(
		interfaces.String("request_id", requestID),
		interfaces.String("user_id", userID),
		interfaces.String("handler", "orders"),
	)

	// Simulação de operação mais complexa
	handlerLogger.Info(ctx, "Processando pedidos")

	// Simulação de múltiplas operações
	operations := []string{"validate_user", "check_inventory", "calculate_price", "create_order"}

	for _, operation := range operations {
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		duration := time.Since(start)

		handlerLogger.Debug(ctx, "Operação executada",
			interfaces.String("operation", operation),
			interfaces.Duration("duration", duration),
		)
	}

	// Simulação de falha ocasional
	if rand.Float32() < 0.1 { // 10% de chance de falha
		handlerLogger.Error(ctx, "Falha no processamento do pedido",
			interfaces.String("error_code", "INVENTORY_UNAVAILABLE"),
			interfaces.String("error_type", "business_logic"),
		)
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": "ord_" + generateRequestID(),
		"status":   "created",
		"total":    299.99,
	})

	handlerLogger.Info(ctx, "Pedido criado com sucesso")
}

func (wal *WebAppLogger) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Health check não deve gerar muito log
	wal.logger.Debug(ctx, "Health check executado")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "v1.0.0",
	})
}

func main() {
	fmt.Println("=== Logger v2 - Web Application ===")

	// Inicialização do logger
	factory := logger.NewFactory()
	factory.RegisterDefaultProviders()

	webLogger, err := NewWebAppLogger(factory)
	if err != nil {
		log.Fatalf("Erro ao criar web logger: %v", err)
	}

	// Configuração das rotas
	mux := http.NewServeMux()

	// Rotas da aplicação
	mux.HandleFunc("/users", webLogger.handleUsers)
	mux.HandleFunc("/orders", webLogger.handleOrders)
	mux.HandleFunc("/health", webLogger.handleHealth)

	// Aplicação do middleware de logging
	handler := webLogger.LoggingMiddleware(mux)

	// Log de inicialização
	ctx := context.Background()
	webLogger.logger.Info(ctx, "Servidor HTTP iniciando",
		interfaces.String("address", ":8080"),
		interfaces.String("environment", "development"),
	)

	// Simulação de algumas requisições para demonstração
	go simulateRequests(webLogger)

	// Inicia o servidor
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Simula execução por alguns segundos
	go func() {
		time.Sleep(10 * time.Second)
		webLogger.logger.Info(ctx, "Parando servidor para demonstração")
		server.Close()
	}()

	webLogger.logger.Info(ctx, "Servidor HTTP iniciado")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		webLogger.logger.Error(ctx, "Erro no servidor HTTP",
			interfaces.ErrorNamed("server_error", err),
		)
	}

	// Cleanup
	webLogger.logger.Info(ctx, "Servidor HTTP finalizado")
	webLogger.logger.Flush()
	webLogger.logger.Close()

	fmt.Println("\n=== Web Application Concluído ===")
}

// simulateRequests simula algumas requisições para demonstração
func simulateRequests(webLogger *WebAppLogger) {
	time.Sleep(1 * time.Second) // Aguarda servidor iniciar

	client := &http.Client{Timeout: 5 * time.Second}

	requests := []struct {
		method string
		url    string
		userID string
	}{
		{"GET", "http://localhost:8080/health", ""},
		{"GET", "http://localhost:8080/users?page=1&limit=5", "user123"},
		{"POST", "http://localhost:8080/users", "admin456"},
		{"GET", "http://localhost:8080/orders", "user789"},
		{"GET", "http://localhost:8080/users?page=2&limit=10", "user999"},
	}

	for i, req := range requests {
		time.Sleep(500 * time.Millisecond)

		httpReq, _ := http.NewRequest(req.method, req.url, nil)
		if req.userID != "" {
			httpReq.Header.Set("X-User-ID", req.userID)
		}

		if req.method == "POST" && req.url == "http://localhost:8080/users" {
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Body = http.NoBody // Simplificado para demonstração
		}

		fmt.Printf("Simulando requisição %d: %s %s\n", i+1, req.method, req.url)
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Printf("Erro na requisição %d: %v\n", i+1, err)
			continue
		}
		resp.Body.Close()
	}
}

// generateRequestID gera um ID único para requisições
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%04x", time.Now().Unix(), rand.Intn(0xFFFF))
}
