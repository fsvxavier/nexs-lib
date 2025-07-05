package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

// User representa um usuário do sistema
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserService simula um serviço de usuários
type UserService struct {
	logger logger.Logger
}

// NewUserService cria um novo serviço de usuários
func NewUserService() *UserService {
	// Logger específico do serviço com campos permanentes
	logger := logger.WithFields(
		logger.String("component", "user-service"),
		logger.String("version", "2.1.0"),
	)

	return &UserService{
		logger: logger,
	}
}

// CreateUser cria um novo usuário
func (s *UserService) CreateUser(ctx context.Context, user User) error {
	start := time.Now()

	// Log de início da operação
	s.logger.Info(ctx, "Iniciando criação de usuário",
		logger.String("user_email", user.Email),
		logger.String("operation", "create_user"),
	)

	// Simula validação
	if user.Email == "" {
		s.logger.WarnWithCode(ctx, "VALIDATION_ERROR", "Email é obrigatório",
			logger.String("user_id", user.ID),
			logger.String("field", "email"),
		)
		return fmt.Errorf("email é obrigatório")
	}

	// Simula operação no banco
	time.Sleep(50 * time.Millisecond) // Simula latência

	// Simula erro ocasional
	if user.Email == "error@test.com" {
		err := fmt.Errorf("falha na conexão com banco")
		s.logger.ErrorWithCode(ctx, "DATABASE_ERROR", "Erro ao criar usuário",
			logger.String("user_email", user.Email),
			logger.ErrorField(err),
			logger.Duration("duration", time.Since(start)),
		)
		return err
	}

	// Log de sucesso
	s.logger.Info(ctx, "Usuário criado com sucesso",
		logger.String("user_id", user.ID),
		logger.String("user_email", user.Email),
		logger.Duration("duration", time.Since(start)),
		logger.String("status", "success"),
	)

	return nil
}

// GetUser busca um usuário por ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Logger com contexto específico da operação
	opLogger := s.logger.WithFields(
		logger.String("operation", "get_user"),
		logger.String("user_id", userID),
	)

	start := time.Now()

	opLogger.Debug(ctx, "Buscando usuário no banco")

	// Simula busca no banco
	time.Sleep(20 * time.Millisecond)

	// Simula usuário não encontrado
	if userID == "not-found" {
		opLogger.WarnWithCode(ctx, "USER_NOT_FOUND", "Usuário não encontrado",
			logger.Duration("duration", time.Since(start)),
		)
		return nil, fmt.Errorf("usuário não encontrado")
	}

	user := &User{
		ID:    userID,
		Email: "user@example.com",
		Name:  "Test User",
	}

	opLogger.Info(ctx, "Usuário encontrado",
		logger.Duration("duration", time.Since(start)),
		logger.String("user_email", user.Email),
	)

	return user, nil
}

// HTTPMiddleware simula um middleware HTTP que adiciona dados ao contexto
func HTTPMiddleware(next func(ctx context.Context)) func(ctx context.Context) {
	return func(ctx context.Context) {
		// Simula dados de requisição HTTP
		requestID := "req-" + fmt.Sprintf("%d", time.Now().UnixNano())
		traceID := "trace-" + fmt.Sprintf("%d", time.Now().UnixNano())
		userID := "user-12345"

		// Adiciona dados ao contexto
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "trace_id", traceID)
		ctx = context.WithValue(ctx, "user_id", userID)

		// Log de início da requisição
		logger.WithContext(ctx).Info(ctx, "Requisição HTTP iniciada",
			logger.String("method", "POST"),
			logger.String("path", "/api/users"),
			logger.String("user_agent", "Mozilla/5.0"),
		)

		start := time.Now()

		// Executa handler
		next(ctx)

		// Log de fim da requisição
		logger.WithContext(ctx).Info(ctx, "Requisição HTTP concluída",
			logger.Duration("duration", time.Since(start)),
			logger.Int("status_code", 200),
		)
	}
}

func main() {
	slogExec()
	fmt.Println()
	fmt.Println()
	zapExec()
	fmt.Println()
	fmt.Println()
	zerologExec()
	fmt.Println()
	fmt.Println()

	fmt.Println("\n=== Providers Disponíveis ===")
	providers := logger.ListProviders()
	for _, provider := range providers {
		fmt.Printf("- %s\n", provider)
	}

	fmt.Println("\n=== Exemplo Concluído ===")
}

func slogExec() {
	fmt.Println("=== Exemplo Avançado de Logging SLOG ===\n")

	// Configuração para desenvolvimento
	devConfig := &logger.Config{
		Level:          logger.DebugLevel,
		Format:         logger.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
		AddSource:      false,
		AddStacktrace:  false,
		Fields: map[string]any{
			"datacenter": "local",
			"instance":   "dev-01",
		},
	}

	err := logger.SetProvider("slog", devConfig)
	if err != nil {
		panic(err)
	}

	// Cria serviço
	userService := NewUserService()

	// Simula requisição HTTP com middleware
	httpHandler := func(ctx context.Context) {
		// Caso de sucesso
		user := User{
			ID:    "user-123",
			Email: "success@example.com",
			Name:  "João Silva",
		}

		err := userService.CreateUser(ctx, user)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao criar usuário", logger.ErrorField(err))
			return
		}

		// Busca o usuário criado
		foundUser, err := userService.GetUser(ctx, user.ID)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao buscar usuário", logger.ErrorField(err))
			return
		}

		logger.Info(ctx, "Operação completada com sucesso",
			logger.String("created_user", foundUser.Email),
		)
	}

	// Executa com middleware
	middlewareHandler := HTTPMiddleware(httpHandler)
	middlewareHandler(context.Background())

	fmt.Println("\n=== Testando Casos de Erro ===\n")

	// Testa caso de erro de validação
	errorHandler1 := func(ctx context.Context) {
		invalidUser := User{
			ID:   "user-456",
			Name: "Usuário Inválido",
			// Email vazio - causará erro de validação
		}

		err := userService.CreateUser(ctx, invalidUser)
		if err != nil {
			logger.Warn(ctx, "Operação falhou por validação",
				logger.ErrorField(err),
				logger.String("reason", "validation"),
			)
		}
	}

	HTTPMiddleware(errorHandler1)(context.Background())

	fmt.Println()

	// Testa caso de erro de banco
	errorHandler2 := func(ctx context.Context) {
		errorUser := User{
			ID:    "user-789",
			Email: "error@test.com", // Email que simula erro de banco
			Name:  "Error User",
		}

		err := userService.CreateUser(ctx, errorUser)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Operação falhou por erro interno",
				logger.ErrorField(err),
				logger.String("reason", "database"),
			)
		}
	}

	HTTPMiddleware(errorHandler2)(context.Background())
}

func zapExec() {
	fmt.Println("=== Exemplo Avançado de Logging ZAP ===\n")

	// Configuração para desenvolvimento
	devConfig := &logger.Config{
		Level:          logger.DebugLevel,
		Format:         logger.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
		AddSource:      false,
		AddStacktrace:  false,
		Fields: map[string]any{
			"datacenter": "local",
			"instance":   "dev-01",
		},
	}

	err := logger.SetProvider("zap", devConfig)
	if err != nil {
		panic(err)
	}

	// Cria serviço
	userService := NewUserService()

	// Simula requisição HTTP com middleware
	httpHandler := func(ctx context.Context) {
		// Caso de sucesso
		user := User{
			ID:    "user-123",
			Email: "success@example.com",
			Name:  "João Silva",
		}

		err := userService.CreateUser(ctx, user)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao criar usuário", logger.ErrorField(err))
			return
		}

		// Busca o usuário criado
		foundUser, err := userService.GetUser(ctx, user.ID)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao buscar usuário", logger.ErrorField(err))
			return
		}

		logger.Info(ctx, "Operação completada com sucesso",
			logger.String("created_user", foundUser.Email),
		)
	}

	// Executa com middleware
	middlewareHandler := HTTPMiddleware(httpHandler)
	middlewareHandler(context.Background())

	fmt.Println("\n=== Testando Casos de Erro ===\n")

	// Testa caso de erro de validação
	errorHandler1 := func(ctx context.Context) {
		invalidUser := User{
			ID:   "user-456",
			Name: "Usuário Inválido",
			// Email vazio - causará erro de validação
		}

		err := userService.CreateUser(ctx, invalidUser)
		if err != nil {
			logger.Warn(ctx, "Operação falhou por validação",
				logger.ErrorField(err),
				logger.String("reason", "validation"),
			)
		}
	}

	HTTPMiddleware(errorHandler1)(context.Background())

	fmt.Println()

	// Testa caso de erro de banco
	errorHandler2 := func(ctx context.Context) {
		errorUser := User{
			ID:    "user-789",
			Email: "error@test.com", // Email que simula erro de banco
			Name:  "Error User",
		}

		err := userService.CreateUser(ctx, errorUser)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Operação falhou por erro interno",
				logger.ErrorField(err),
				logger.String("reason", "database"),
			)
		}
	}

	HTTPMiddleware(errorHandler2)(context.Background())
}

func zerologExec() {
	fmt.Println("=== Exemplo Avançado de Logging ZEROLOG ===\n")

	// Configuração para desenvolvimento
	devConfig := &logger.Config{
		Level:          logger.DebugLevel,
		Format:         logger.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
		AddSource:      false,
		AddStacktrace:  false,
		Fields: map[string]any{
			"datacenter": "local",
			"instance":   "dev-01",
		},
	}

	err := logger.SetProvider("zerolog", devConfig)
	if err != nil {
		panic(err)
	}

	// Cria serviço
	userService := NewUserService()

	// Simula requisição HTTP com middleware
	httpHandler := func(ctx context.Context) {
		// Caso de sucesso
		user := User{
			ID:    "user-123",
			Email: "success@example.com",
			Name:  "João Silva",
		}

		err := userService.CreateUser(ctx, user)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao criar usuário", logger.ErrorField(err))
			return
		}

		// Busca o usuário criado
		foundUser, err := userService.GetUser(ctx, user.ID)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Falha ao buscar usuário", logger.ErrorField(err))
			return
		}

		logger.Info(ctx, "Operação completada com sucesso",
			logger.String("created_user", foundUser.Email),
		)
	}

	// Executa com middleware
	middlewareHandler := HTTPMiddleware(httpHandler)
	middlewareHandler(context.Background())

	fmt.Println("\n=== Testando Casos de Erro ===\n")

	// Testa caso de erro de validação
	errorHandler1 := func(ctx context.Context) {
		invalidUser := User{
			ID:   "user-456",
			Name: "Usuário Inválido",
			// Email vazio - causará erro de validação
		}

		err := userService.CreateUser(ctx, invalidUser)
		if err != nil {
			logger.Warn(ctx, "Operação falhou por validação",
				logger.ErrorField(err),
				logger.String("reason", "validation"),
			)
		}
	}

	HTTPMiddleware(errorHandler1)(context.Background())

	fmt.Println()

	// Testa caso de erro de banco
	errorHandler2 := func(ctx context.Context) {
		errorUser := User{
			ID:    "user-789",
			Email: "error@test.com", // Email que simula erro de banco
			Name:  "Error User",
		}

		err := userService.CreateUser(ctx, errorUser)
		if err != nil {
			logger.WithContext(ctx).Error(ctx, "Operação falhou por erro interno",
				logger.ErrorField(err),
				logger.String("reason", "database"),
			)
		}
	}

	HTTPMiddleware(errorHandler2)(context.Background())
}
