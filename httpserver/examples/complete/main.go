package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/middlewares"
	"github.com/gin-gonic/gin"
)

// Exemplo completo combinando hooks e middlewares para monitoramento avançado

func main() {
	log.Println("🚀 Exemplo Completo - Hooks + Middlewares")

	// ==============================
	// CONFIGURAÇÃO DE HOOKS
	// ==============================

	hookManager := hooks.NewHookManager()

	// Hooks de ciclo de vida
	startHook := hooks.NewStartHook("server-lifecycle")
	startHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("start", startHook)

	stopHook := hooks.NewStopHook("server-lifecycle")
	stopHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("stop", stopHook)

	// Hooks de monitoramento
	requestHook := hooks.NewRequestHook("request-monitor")
	requestHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("request", requestHook)

	responseHook := hooks.NewResponseHook("response-monitor")
	responseHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("response", responseHook)

	errorHook := hooks.NewErrorHook("error-monitor")
	errorHook.SetMetricsEnabled(true)
	errorHook.SetErrorThreshold(5)
	hookManager.RegisterHook("error", errorHook)

	// Hooks de rota
	routeInHook := hooks.NewRouteInHook("route-in-monitor")
	routeInHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("route-in", routeInHook)

	routeOutHook := hooks.NewRouteOutHook("route-out-monitor")
	routeOutHook.SetMetricsEnabled(true)
	routeOutHook.SetSlowThreshold(500 * time.Millisecond) // Alertar requisições > 500ms
	hookManager.RegisterHook("route-out", routeOutHook)

	log.Printf("✅ %d hooks registrados", len(hookManager.ListHooks()))

	// ==============================
	// CONFIGURAÇÃO DE MIDDLEWARES
	// ==============================

	middlewareManager := middlewares.NewMiddlewareManager()

	// Middleware de logging avançado
	loggingConfig := middlewares.LoggingConfig{
		LogRequests:      true,
		LogResponses:     true,
		LogHeaders:       true,
		LogBody:          true,
		LogSensitiveData: false,
		SkipPaths:        []string{"/health", "/favicon.ico"},
		SkipMethods:      []string{"OPTIONS"},
		MaxBodySize:      2048,
		TruncateBody:     true,
	}
	loggingMiddleware := middlewares.NewLoggingMiddlewareWithConfig(0, loggingConfig)
	middlewareManager.AddMiddleware(loggingMiddleware)

	// Middleware de autenticação com múltiplas opções
	authConfig := middlewares.AuthConfig{
		EnableBasicAuth:  true,
		EnableAPIKeyAuth: true,
		BasicAuthRealm:   "Nexs API",
		BasicAuthUsers: map[string]string{
			"admin":     "admin123",
			"user":      "user123",
			"developer": "dev123",
		},
		ValidTokens: map[string]middlewares.AuthUser{
			"api-key-123": {
				ID:    "api-user-1",
				Roles: []string{"api", "read"},
			},
			"admin-key-456": {
				ID:    "admin-user-1",
				Roles: []string{"admin", "read", "write"},
			},
		},
		SkipPaths: []string{"/", "/health", "/public", "/docs"},
	}
	authMiddleware := middlewares.NewAuthMiddlewareWithConfig(1, authConfig)
	middlewareManager.AddMiddleware(authMiddleware)

	log.Printf("✅ %d middlewares configurados", len(middlewareManager.ListMiddlewares()))

	// ==============================
	// CONFIGURAÇÃO DO SERVIDOR
	// ==============================

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware principal que integra hooks e middlewares
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		startTime := time.Now()

		// Hooks de entrada
		requestHook.OnRequest(ctx, c.Request)
		routeInHook.OnRouteEnter(ctx, c.Request.Method, c.FullPath(), c.Request)

		// Processar middlewares
		_, err := middlewareManager.ProcessRequest(ctx, c.Request)
		if err != nil {
			errorHook.OnError(ctx, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Acesso negado",
				"message":   err.Error(),
				"timestamp": time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()

		// Hooks de saída
		duration := time.Since(startTime)
		responseHook.OnResponse(ctx, c.Request, c.Writer, duration)
		routeOutHook.OnRouteExit(ctx, c.Request.Method, c.FullPath(), c.Writer, duration)
	})

	// ==============================
	// ROTAS PÚBLICAS
	// ==============================

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "🎯 Nexs Lib - Exemplo Completo",
			"version":     "1.0.0",
			"features":    []string{"hooks", "middlewares", "monitoring", "auth"},
			"hooks":       len(hookManager.ListHooks()),
			"middlewares": len(middlewareManager.ListMiddlewares()),
			"docs":        "/docs",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(time.Now()).String(),
			"server":    "nexs-example",
		})
	})

	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "📢 Esta é uma área pública",
			"info":    "Nenhuma autenticação necessária",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	router.GET("/docs", func(c *gin.Context) {
		docs := map[string]interface{}{
			"title": "Nexs Lib API Documentation",
			"endpoints": map[string]interface{}{
				"public":    []string{"/", "/health", "/public", "/docs"},
				"protected": []string{"/api/*", "/admin/*", "/metrics"},
			},
			"authentication": map[string]interface{}{
				"basic_auth": map[string]string{
					"admin":     "admin123",
					"user":      "user123",
					"developer": "dev123",
				},
				"api_keys": []string{
					"api-key-123 (read access)",
					"admin-key-456 (full access)",
				},
			},
		}
		c.JSON(http.StatusOK, docs)
	})

	// ==============================
	// API PROTEGIDA
	// ==============================

	api := router.Group("/api")
	{
		api.GET("/users", func(c *gin.Context) {
			time.Sleep(100 * time.Millisecond) // Simular DB query
			c.JSON(http.StatusOK, gin.H{
				"users": []map[string]interface{}{
					{"id": 1, "name": "João Silva", "role": "admin", "active": true},
					{"id": 2, "name": "Maria Santos", "role": "user", "active": true},
					{"id": 3, "name": "Pedro Costa", "role": "developer", "active": false},
				},
				"total": 3,
				"page":  1,
			})
		})

		api.POST("/users", func(c *gin.Context) {
			var user map[string]interface{}
			if err := c.ShouldBindJSON(&user); err != nil {
				ctx := c.Request.Context()
				errorHook.OnError(ctx, err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Dados inválidos",
					"details": err.Error(),
				})
				return
			}

			time.Sleep(200 * time.Millisecond) // Simular criação
			user["id"] = 4
			user["created_at"] = time.Now().Format(time.RFC3339)
			user["active"] = true

			c.JSON(http.StatusCreated, gin.H{
				"message": "Usuário criado com sucesso",
				"user":    user,
			})
		})

		api.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"profile": map[string]interface{}{
					"user_id":     "current_user",
					"name":        "Usuário Atual",
					"permissions": []string{"read", "write", "admin"},
					"login_time":  time.Now().Format(time.RFC3339),
					"session_id":  "sess_" + time.Now().Format("20060102150405"),
				},
			})
		})

		api.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"requests": map[string]interface{}{
					"total":          requestHook.GetRequestCount(),
					"active":         requestHook.GetActiveRequestCount(),
					"max_concurrent": requestHook.GetMaxActiveRequestCount(),
					"avg_size":       requestHook.GetAverageRequestSize(),
				},
				"routes": map[string]interface{}{
					"count": routeInHook.GetMetricsCount(),
				},
				"server": map[string]interface{}{
					"starts":  startHook.GetStartCount(),
					"stops":   stopHook.GetStopCount(),
					"running": startHook.IsServerRunning(),
				},
			})
		})

		api.GET("/slow", func(c *gin.Context) {
			// Rota intencionalmente lenta para testar alertas
			time.Sleep(1 * time.Second)
			c.JSON(http.StatusOK, gin.H{
				"message":  "⏱️ Processamento lento completado",
				"duration": "1 segundo",
				"warning":  "Esta rota é intencionalmente lenta",
			})
		})

		api.GET("/error", func(c *gin.Context) {
			ctx := c.Request.Context()
			err := &APIError{
				Message: "Erro simulado para demonstração",
				Code:    "DEMO_ERROR",
				Status:  500,
			}
			errorHook.OnError(ctx, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Message,
				"code":  err.Code,
			})
		})
	}

	// ==============================
	// ÁREA ADMINISTRATIVA
	// ==============================

	admin := router.Group("/admin")
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"dashboard": "🎛️ Painel Administrativo",
				"metrics": map[string]interface{}{
					"total_requests":    requestHook.GetRequestCount(),
					"active_requests":   requestHook.GetActiveRequestCount(),
					"hooks_count":       len(hookManager.ListHooks()),
					"middlewares_count": len(middlewareManager.ListMiddlewares()),
				},
				"system": map[string]interface{}{
					"uptime": time.Since(time.Now()).String(),
					"memory": "N/A", // Poderia incluir métricas de memória
					"status": "operational",
				},
			})
		})

		admin.GET("/logs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"logs": "📋 Logs do sistema disponíveis via hooks",
				"note": "Em produção, isso consultaria logs reais",
			})
		})
	}

	// ==============================
	// MÉTRICAS COMPLETAS
	// ==============================

	router.GET("/metrics", func(c *gin.Context) {
		metrics := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"hooks": map[string]interface{}{
				"registered": len(hookManager.ListHooks()),
				"list":       hookManager.ListHooks(),
			},
			"middlewares": map[string]interface{}{
				"registered": len(middlewareManager.ListMiddlewares()),
				"list":       middlewareManager.ListMiddlewares(),
			},
			"requests": map[string]interface{}{
				"total":          requestHook.GetRequestCount(),
				"active":         requestHook.GetActiveRequestCount(),
				"max_concurrent": requestHook.GetMaxActiveRequestCount(),
				"total_size":     requestHook.GetTotalRequestSize(),
				"average_size":   requestHook.GetAverageRequestSize(),
			},
			"server": map[string]interface{}{
				"start_count": startHook.GetStartCount(),
				"stop_count":  stopHook.GetStopCount(),
				"is_running":  startHook.IsServerRunning(),
			},
			"routes": map[string]interface{}{
				"metrics_count": routeInHook.GetMetricsCount(),
			},
		}
		c.JSON(http.StatusOK, metrics)
	})

	// ==============================
	// INICIALIZAÇÃO DO SERVIDOR
	// ==============================

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	ctx := context.Background()
	startHook.OnStart(ctx, ":8080")

	go func() {
		log.Printf("🌟 Servidor completo iniciado na porta 8080")
		log.Printf("")
		log.Printf("📊 ENDPOINTS PÚBLICOS:")
		log.Printf("   GET  /           - Página inicial")
		log.Printf("   GET  /health     - Health check")
		log.Printf("   GET  /public     - Área pública")
		log.Printf("   GET  /docs       - Documentação")
		log.Printf("")
		log.Printf("🔒 ENDPOINTS PROTEGIDOS:")
		log.Printf("   GET  /api/users     - Lista usuários")
		log.Printf("   POST /api/users     - Criar usuário")
		log.Printf("   GET  /api/profile   - Perfil atual")
		log.Printf("   GET  /api/stats     - Estatísticas")
		log.Printf("   GET  /api/slow      - Teste de latência")
		log.Printf("   GET  /api/error     - Teste de erro")
		log.Printf("   GET  /admin/*       - Área administrativa")
		log.Printf("   GET  /metrics       - Métricas completas")
		log.Printf("")
		log.Printf("🔐 AUTENTICAÇÃO:")
		log.Printf("   Basic Auth:")
		log.Printf("     admin:admin123 | user:user123 | developer:dev123")
		log.Printf("   API Keys:")
		log.Printf("     X-API-Key: api-key-123 | X-API-Key: admin-key-456")
		log.Printf("")
		log.Printf("🧪 EXEMPLOS:")
		log.Printf("   curl http://localhost:8080/")
		log.Printf("   curl -u admin:admin123 http://localhost:8080/api/users")
		log.Printf("   curl -H 'X-API-Key: api-key-123' http://localhost:8080/metrics")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
		}
	}()

	// ==============================
	// GRACEFUL SHUTDOWN
	// ==============================

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("🛑 Iniciando shutdown graceful...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopHook.OnStop(shutdownCtx)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Erro durante shutdown: %v", err)
		errorHook.OnError(shutdownCtx, err)
	}

	log.Printf("✅ Servidor finalizado com sucesso")
	log.Printf("📊 ESTATÍSTICAS FINAIS:")
	log.Printf("   Hooks registrados: %d", len(hookManager.ListHooks()))
	log.Printf("   Middlewares registrados: %d", len(middlewareManager.ListMiddlewares()))
	log.Printf("   Total de requisições: %d", requestHook.GetRequestCount())
	log.Printf("   Pico de requisições concorrentes: %d", requestHook.GetMaxActiveRequestCount())
	log.Printf("   Inicializações do servidor: %d", startHook.GetStartCount())
	log.Printf("   Paradas do servidor: %d", stopHook.GetStopCount())
}

// APIError implementa uma estrutura de erro personalizada para API
type APIError struct {
	Message string
	Code    string
	Status  int
}

func (e *APIError) Error() string {
	return e.Message
}
