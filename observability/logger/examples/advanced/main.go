package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/observability/logger"

	// Importa todos os providers para auto-registração
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

// User representa um usuário do sistema
type User struct {
	ID    string
	Email string
	Name  string
}

// UserService simula um serviço de usuários
type UserService struct {
	logger logger.Logger
}

// NewUserService cria um novo serviço de usuários
func NewUserService() *UserService {
	return &UserService{
		logger: logger.WithFields(
			logger.String("service", "user-service"),
			logger.String("version", "2.1.0"),
		),
	}
}

// CreateUser simula a criação de um usuário
func (s *UserService) CreateUser(ctx context.Context, user User) error {
	s.logger.Info(ctx, "Iniciando criação de usuário",
		logger.String("user_id", user.ID),
		logger.String("email", user.Email),
	)

	// Simula validação
	if user.Email == "" {
		s.logger.ErrorWithCode(ctx, "USER_INVALID_EMAIL", "Email é obrigatório",
			logger.String("user_id", user.ID),
		)
		return fmt.Errorf("email é obrigatório")
	}

	// Simula processamento
	time.Sleep(50 * time.Millisecond)

	s.logger.InfoWithCode(ctx, "USER_CREATED", "Usuário criado com sucesso",
		logger.String("user_id", user.ID),
		logger.String("email", user.Email),
		logger.String("name", user.Name),
	)

	return nil
}

// GetUser simula a busca de um usuário
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	s.logger.Debug(ctx, "Buscando usuário",
		logger.String("user_id", userID),
	)

	// Simula busca no banco
	start := time.Now()
	time.Sleep(25 * time.Millisecond)
	duration := time.Since(start)

	if userID == "not-found" {
		s.logger.WarnWithCode(ctx, "USER_NOT_FOUND", "Usuário não encontrado",
			logger.String("user_id", userID),
			logger.Duration("query_duration", duration),
		)
		return nil, fmt.Errorf("usuário não encontrado")
	}

	user := &User{
		ID:    userID,
		Email: "user@example.com",
		Name:  "John Doe",
	}

	s.logger.Info(ctx, "Usuário encontrado",
		logger.String("user_id", userID),
		logger.Duration("query_duration", duration),
	)

	return user, nil
}

// HTTPMiddleware simula um middleware HTTP que adiciona dados ao contexto
func HTTPMiddleware(next func(ctx context.Context)) func(ctx context.Context) {
	return func(ctx context.Context) {
		// Simula dados de requisição HTTP
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())

		ctx = context.WithValue(ctx, logger.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.TraceIDKey, traceID)

		// Log de início da requisição
		logger.Info(ctx, "Iniciando processamento da requisição",
			logger.String("method", "POST"),
			logger.String("path", "/api/users"),
			logger.String("user_agent", "Go-http-client/1.1"),
		)

		start := time.Now()

		// Executa o handler
		next(ctx)

		// Log de fim da requisição
		duration := time.Since(start)
		logger.Info(ctx, "Requisição processada",
			logger.Duration("request_duration", duration),
			logger.String("status", "200"),
		)
	}
}

func main() {
	fmt.Println("=== Exemplo Avançado de Logging ===")

	// Configuração avançada
	config := &logger.Config{
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
			"datacenter": "us-east-1",
			"instance":   "web-01",
		},
		SamplingConfig: &logger.SamplingConfig{
			Initial:    100,
			Thereafter: 10,
			Tick:       time.Second,
		},
	}

	err := logger.SetProvider("slog", config)
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
			logger.Error(ctx, "Erro ao criar usuário",
				logger.ErrorField(err),
				logger.String("user_id", user.ID),
			)
		}

		// Caso de erro de validação
		invalidUser := User{
			ID:    "user-456",
			Email: "", // Email vazio para gerar erro
			Name:  "Maria Santos",
		}

		err = userService.CreateUser(ctx, invalidUser)
		if err != nil {
			logger.Error(ctx, "Erro de validação",
				logger.ErrorField(err),
				logger.String("user_id", invalidUser.ID),
			)
		}

		// Teste de busca
		foundUser, err := userService.GetUser(ctx, "user-789")
		if err != nil {
			logger.Error(ctx, "Erro ao buscar usuário",
				logger.ErrorField(err),
				logger.String("user_id", "user-789"),
			)
		} else {
			logger.Info(ctx, "Usuário retornado para cliente",
				logger.String("user_id", foundUser.ID),
				logger.String("email", foundUser.Email),
			)
		}

		// Teste de usuário não encontrado
		_, err = userService.GetUser(ctx, "not-found")
		if err != nil {
			logger.Warn(ctx, "Usuário não encontrado para requisição",
				logger.ErrorField(err),
				logger.String("user_id", "not-found"),
			)
		}
	}

	// Executa com middleware
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.UserIDKey, "authenticated-user-123")

	wrappedHandler := HTTPMiddleware(httpHandler)
	wrappedHandler(ctx)

	fmt.Println("\n=== Exemplo de Logging com Diferentes Formatos ===")

	// Teste com formato Console
	fmt.Println("\n--- Formato Console ---")
	consoleConfig := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.ConsoleFormat,
		Output:         os.Stdout,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
	}

	err = logger.SetProvider("slog", consoleConfig)
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "Exemplo de formato console",
		logger.String("format", "console"),
		logger.Bool("readable", true),
	)

	// Teste com formato Text
	fmt.Println("\n--- Formato Text ---")
	textConfig := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.TextFormat,
		Output:         os.Stdout,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
	}

	err = logger.SetProvider("slog", textConfig)
	if err != nil {
		panic(err)
	}

	logger.Info(ctx, "Exemplo de formato text",
		logger.String("format", "text"),
		logger.Bool("structured", true),
	)

	fmt.Println("\n=== Exemplo de Configuração por Ambiente ===")

	// Simula diferentes ambientes
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		fmt.Printf("\n--- Ambiente: %s ---\n", env)

		os.Setenv("ENVIRONMENT", env)
		os.Setenv("SERVICE_NAME", "user-api")
		os.Setenv("SERVICE_VERSION", "2.1.0")
		os.Setenv("LOG_LEVEL", "info")

		envConfig := logger.EnvironmentConfig()

		// Ajusta configuração baseada no ambiente
		switch env {
		case "development":
			envConfig.Level = logger.DebugLevel
			envConfig.Format = logger.ConsoleFormat
			envConfig.AddSource = true
		case "staging":
			envConfig.Level = logger.InfoLevel
			envConfig.Format = logger.JSONFormat
			envConfig.AddStacktrace = true
		case "production":
			envConfig.Level = logger.WarnLevel
			envConfig.Format = logger.JSONFormat
			envConfig.AddStacktrace = true
			envConfig.SamplingConfig = &logger.SamplingConfig{
				Initial:    1000,
				Thereafter: 100,
				Tick:       time.Second,
			}
		}

		err = logger.SetProvider("slog", envConfig)
		if err != nil {
			panic(err)
		}

		logger.Info(ctx, "Configuração aplicada",
			logger.String("environment", env),
			logger.String("level", envConfig.Level.String()),
			logger.String("format", string(envConfig.Format)),
		)
	}

	fmt.Println("\n=== Demonstração de Todos os Providers ===")

	// Configuração base para todos os providers
	baseConfig := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         os.Stdout,
		ServiceName:    "user-api",
		ServiceVersion: "2.1.0",
		Environment:    "development",
	}

	providers := []string{"slog", "zap", "zerolog"}

	for _, provider := range providers {
		fmt.Printf("\n--- Provider: %s ---\n", provider)

		// Configura o provider
		err = logger.SetProvider(provider, baseConfig)
		if err != nil {
			panic(err)
		}

		// Exemplo de uso do UserService com diferentes providers
		userService := NewUserService()

		// Busca um usuário que existe
		foundUser, err := userService.GetUser(ctx, "123")
		if err != nil {
			logger.Error(ctx, "Erro ao buscar usuário",
				logger.ErrorField(err),
				logger.String("user_id", "123"),
				logger.String("provider", provider),
			)
		} else {
			logger.Info(ctx, "Usuário encontrado",
				logger.String("user_id", foundUser.ID),
				logger.String("email", foundUser.Email),
				logger.String("provider", provider),
			)
		}

		// Simula um erro de banco de dados
		dbError := domainerrors.NewDatabaseError("DB_CONNECTION_ERROR", "Conexão com banco de dados perdida", nil)

		logger.Error(ctx, "Erro de infraestrutura",
			logger.ErrorField(dbError),
			logger.String("provider", provider),
			logger.String("component", "database"),
		)
	}

	fmt.Println("\n=== Exemplo de Performance e Benchmark ===")

	// Benchmark comparativo entre providers
	benchmarkConfig := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: os.Stdout,
	}

	iterations := 1000

	for _, provider := range providers {
		fmt.Printf("\n--- Benchmark Provider: %s ---\n", provider)

		err = logger.SetProvider(provider, benchmarkConfig)
		if err != nil {
			panic(err)
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			logger.Info(ctx, "Benchmark log message",
				logger.String("provider", provider),
				logger.String("iteration", fmt.Sprintf("%d", i)),
				logger.Int("number", i),
				logger.Bool("benchmark", true),
			)
		}
		duration := time.Since(start)

		logger.Info(ctx, "Benchmark completado",
			logger.String("provider", provider),
			logger.Int("iterations", iterations),
			logger.Duration("total_duration", duration),
			logger.Duration("avg_per_log", duration/time.Duration(iterations)),
		)
	}

	fmt.Println("\n=== Exemplo Concluído ===")
}
