package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/config"
	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

var (
	// Tracer global para toda a aplicação
	globalTracer trace.Tracer
)

func main() {
	fmt.Println("🌍 Exemplo Global OpenTelemetry")
	fmt.Println("===============================")

	// Inicializar tracing global
	shutdown := initGlobalTracing()
	defer shutdown()

	// Obter tracer global
	globalTracer = otel.Tracer("global-example")

	fmt.Println("🚀 Iniciando aplicação web com tracing global...")

	// Simular aplicação web com múltiplos componentes
	runWebApplication()

	fmt.Println("✅ Aplicação concluída!")
}

func initGlobalTracing() func() {
	fmt.Println("⚙️ Configurando tracing global...")

	// Configuração baseada em ambiente
	cfg := config.NewConfigFromEnv()

	// Fallback para desenvolvimento se não configurado
	if cfg.ServiceName == "" {
		cfg = interfaces.Config{
			ServiceName:   "global-web-app",
			Environment:   "development",
			ExporterType:  "opentelemetry",
			Endpoint:      "http://otel-collector:4318/v1/traces",
			SamplingRatio: 1.0,
			Version:       "1.0.0",
			Propagators:   []string{"tracecontext", "b3", "baggage"},
			Insecure:      true,
			Attributes: map[string]string{
				"service.type":   "web-application",
				"deployment.env": "development",
				"team":           "platform",
			},
		}
	}

	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("❌ Erro na configuração: %v", err)
	}

	// Inicializar TracerManager
	tracerManager := tracer.NewTracerManager()
	ctx := context.Background()

	fmt.Printf("📡 Inicializando %s tracer...\n", cfg.ExporterType)
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar tracer: %v", err)
	}

	// ⭐ CONFIGURAR COMO TRACER GLOBAL ⭐
	// Isso permite que qualquer código da aplicação use otel.Tracer()
	// sem precisar passar o TracerProvider explicitamente
	otel.SetTracerProvider(tracerProvider)

	fmt.Println("✅ TracerProvider configurado globalmente")
	fmt.Printf("🔧 Configuração: %s (%s)\n", cfg.ServiceName, cfg.ExporterType)

	// Retornar função de shutdown
	return func() {
		fmt.Println("🔄 Fazendo shutdown do tracer global...")
		if err := tracerManager.Shutdown(ctx); err != nil {
			log.Printf("⚠️ Erro no shutdown: %v", err)
		}
		fmt.Println("✅ Tracer global finalizado")
	}
}

func runWebApplication() {
	ctx := context.Background()

	// Simular multiple requests HTTP
	for i := 1; i <= 3; i++ {
		handleHTTPRequest(ctx, fmt.Sprintf("request-%d", i))
		time.Sleep(100 * time.Millisecond)
	}
}

func handleHTTPRequest(ctx context.Context, requestID string) {
	// Como o TracerProvider foi configurado globalmente,
	// podemos obter um tracer em qualquer lugar da aplicação
	tracer := otel.Tracer("http-handler")

	ctx, span := tracer.Start(ctx, "http-request")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.route", "/api/v1/users"),
		attribute.String("request.id", requestID),
		attribute.String("user.agent", "Go-Example/1.0"),
	)

	fmt.Printf("🌐 Processando request: %s\n", requestID)

	// Middleware de autenticação
	if !authenticationMiddleware(ctx) {
		span.SetStatus(codes.Error, "Authentication failed")
		span.SetAttributes(attribute.Int("http.status_code", 401))
		return
	}

	// Business logic
	if !processBusinessLogic(ctx, requestID) {
		span.SetStatus(codes.Error, "Business logic failed")
		span.SetAttributes(attribute.Int("http.status_code", 500))
		return
	}

	// Response
	span.SetAttributes(
		attribute.Int("http.status_code", 200),
		attribute.String("response.status", "success"),
	)
	span.SetStatus(codes.Ok, "Request processed successfully")

	fmt.Printf("✅ Request %s processado com sucesso\n", requestID)
}

func authenticationMiddleware(ctx context.Context) bool {
	// Cada função pode criar seu próprio tracer usando o provider global
	tracer := otel.Tracer("auth-middleware")

	ctx, span := tracer.Start(ctx, "authentication")
	defer span.End()

	span.SetAttributes(
		attribute.String("auth.method", "bearer-token"),
		attribute.String("auth.provider", "jwt"),
	)

	// Simular validação de token
	if !validateJWTToken(ctx) {
		span.SetStatus(codes.Error, "Invalid token")
		return false
	}

	// Simular verificação de permissões
	if !checkPermissions(ctx) {
		span.SetStatus(codes.Error, "Insufficient permissions")
		return false
	}

	span.SetStatus(codes.Ok, "Authentication successful")
	fmt.Println("🔐 Autenticação realizada")
	return true
}

func validateJWTToken(ctx context.Context) bool {
	tracer := otel.Tracer("jwt-validator")

	ctx, span := tracer.Start(ctx, "validate-jwt")
	defer span.End()

	span.SetAttributes(
		attribute.String("token.type", "JWT"),
		attribute.String("token.algorithm", "RS256"),
		attribute.Bool("token.expired", false),
	)

	// Simular validação
	time.Sleep(20 * time.Millisecond)

	span.SetStatus(codes.Ok, "Token valid")
	return true
}

func checkPermissions(ctx context.Context) bool {
	tracer := otel.Tracer("permission-checker")

	ctx, span := tracer.Start(ctx, "check-permissions")
	defer span.End()

	span.SetAttributes(
		attribute.String("permission.resource", "users"),
		attribute.String("permission.action", "read"),
		attribute.Bool("permission.granted", true),
	)

	// Simular verificação
	time.Sleep(15 * time.Millisecond)

	span.SetStatus(codes.Ok, "Permission granted")
	return true
}

func processBusinessLogic(ctx context.Context, requestID string) bool {
	tracer := otel.Tracer("business-logic")

	ctx, span := tracer.Start(ctx, "process-business-logic")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("operation.type", "user-management"),
	)

	// Múltiplas operações de negócio
	if !fetchUserData(ctx, "user-123") {
		return false
	}

	if !enrichUserProfile(ctx, "user-123") {
		return false
	}

	if !auditUserAccess(ctx, requestID, "user-123") {
		return false
	}

	span.SetStatus(codes.Ok, "Business logic completed")
	fmt.Printf("💼 Lógica de negócio processada para %s\n", requestID)
	return true
}

func fetchUserData(ctx context.Context, userID string) bool {
	tracer := otel.Tracer("user-service")

	ctx, span := tracer.Start(ctx, "fetch-user-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("data.source", "database"),
	)

	// Simular consulta ao banco
	if !queryUserDatabase(ctx, userID) {
		span.SetStatus(codes.Error, "Database query failed")
		return false
	}

	// Simular cache
	cacheUserData(ctx, userID)

	span.SetStatus(codes.Ok, "User data fetched")
	fmt.Printf("👤 Dados do usuário %s obtidos\n", userID)
	return true
}

func queryUserDatabase(ctx context.Context, userID string) bool {
	tracer := otel.Tracer("database")

	ctx, span := tracer.Start(ctx, "query-user")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "users_db"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
		attribute.String("db.user", userID),
		attribute.Int("db.rows_affected", 1),
	)

	// Simular consulta
	time.Sleep(50 * time.Millisecond)

	span.SetStatus(codes.Ok, "Query successful")
	return true
}

func cacheUserData(ctx context.Context, userID string) {
	tracer := otel.Tracer("cache")

	ctx, span := tracer.Start(ctx, "cache-user-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.system", "redis"),
		attribute.String("cache.key", "user:"+userID),
		attribute.Int("cache.ttl", 3600),
	)

	// Simular cache
	time.Sleep(10 * time.Millisecond)

	span.SetStatus(codes.Ok, "Data cached")
	fmt.Printf("🗄️ Dados do usuário %s em cache\n", userID)
}

func enrichUserProfile(ctx context.Context, userID string) bool {
	tracer := otel.Tracer("profile-enricher")

	ctx, span := tracer.Start(ctx, "enrich-user-profile")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("enrichment.type", "profile-data"),
	)

	// Múltiplas fontes de dados
	fetchUserPreferences(ctx, userID)
	fetchUserHistory(ctx, userID)
	fetchUserRecommendations(ctx, userID)

	span.SetStatus(codes.Ok, "Profile enriched")
	fmt.Printf("✨ Perfil do usuário %s enriquecido\n", userID)
	return true
}

func fetchUserPreferences(ctx context.Context, userID string) {
	tracer := otel.Tracer("preferences-service")

	ctx, span := tracer.Start(ctx, "fetch-preferences")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("preferences.categories", "theme,language,notifications"),
	)

	time.Sleep(30 * time.Millisecond)
	span.SetStatus(codes.Ok, "Preferences fetched")
}

func fetchUserHistory(ctx context.Context, userID string) {
	tracer := otel.Tracer("history-service")

	ctx, span := tracer.Start(ctx, "fetch-history")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.Int("history.items", 25),
		attribute.String("history.period", "30d"),
	)

	time.Sleep(40 * time.Millisecond)
	span.SetStatus(codes.Ok, "History fetched")
}

func fetchUserRecommendations(ctx context.Context, userID string) {
	tracer := otel.Tracer("recommendation-service")

	ctx, span := tracer.Start(ctx, "fetch-recommendations")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("recommendation.algorithm", "collaborative-filtering"),
		attribute.Int("recommendation.count", 10),
	)

	time.Sleep(60 * time.Millisecond)
	span.SetStatus(codes.Ok, "Recommendations fetched")
}

func auditUserAccess(ctx context.Context, requestID, userID string) bool {
	tracer := otel.Tracer("audit-service")

	ctx, span := tracer.Start(ctx, "audit-user-access")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("user.id", userID),
		attribute.String("audit.action", "user-data-access"),
		attribute.String("audit.result", "success"),
	)

	// Simular logging de auditoria
	time.Sleep(25 * time.Millisecond)

	span.SetStatus(codes.Ok, "Access audited")
	fmt.Printf("📋 Acesso auditado para usuário %s\n", userID)
	return true
}

// Exemplo de como usar em handlers HTTP reais
func setupHTTPServer() *http.ServeMux {
	mux := http.NewServeMux()

	// Handler que usa o tracer global automaticamente
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Obter tracer global
		tracer := otel.Tracer("health-check")

		_, span := tracer.Start(r.Context(), "health-check")
		defer span.End()

		span.SetAttributes(
			attribute.String("health.component", "api"),
			attribute.String("health.status", "healthy"),
		)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))

		span.SetStatus(codes.Ok, "Health check passed")
	})

	return mux
}

// Exemplo de middleware que propaga context automaticamente
func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("http-middleware")

		ctx, span := tracer.Start(r.Context(), "http-request")
		defer span.End()

		// Extrair e propagar context
		r = r.WithContext(ctx)

		// Adicionar atributos HTTP
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
		)

		// Chamar próximo handler
		next.ServeHTTP(w, r)
	})
}
