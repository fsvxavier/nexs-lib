package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middlewares"
	"github.com/gin-gonic/gin"
)

// Exemplo básico demonstrando como usar middlewares para autenticação e logging

func main() {
	log.Println("🚀 Exemplo Básico - Sistema de Middlewares")

	// ==============================
	// CONFIGURAÇÃO DE MIDDLEWARES
	// ==============================

	middlewareManager := middlewares.NewMiddlewareManager()

	// 1. Middleware de Logging
	loggingConfig := middlewares.LoggingConfig{
		LogRequests:      true,
		LogResponses:     true,
		LogHeaders:       true,
		LogBody:          false, // Desabilitado para este exemplo simples
		LogSensitiveData: false,
		SkipPaths:        []string{"/health"},
		SkipMethods:      []string{"OPTIONS"},
		MaxBodySize:      1024,
		TruncateBody:     true,
	}
	loggingMiddleware := middlewares.NewLoggingMiddlewareWithConfig(0, loggingConfig)
	middlewareManager.AddMiddleware(loggingMiddleware)

	// 2. Middleware de Autenticação
	authConfig := middlewares.AuthConfig{
		EnableBasicAuth: true,
		BasicAuthRealm:  "API Access",
		BasicAuthUsers: map[string]string{
			"admin": "secret123",
			"user":  "password456",
		},
		SkipPaths: []string{"/", "/health", "/public"},
	}
	authMiddleware := middlewares.NewAuthMiddlewareWithConfig(1, authConfig)
	middlewareManager.AddMiddleware(authMiddleware)

	log.Printf("✅ %d middlewares configurados", len(middlewareManager.ListMiddlewares()))

	// ==============================
	// CONFIGURAÇÃO DO SERVIDOR
	// ==============================

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware que integra nossos middlewares personalizados
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()

		// Processar middlewares
		_, err := middlewareManager.ProcessRequest(ctx, c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Acesso negado",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	})

	// ==============================
	// ROTAS PÚBLICAS
	// ==============================

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Bem-vindo ao exemplo básico de middlewares!",
			"middlewares": len(middlewareManager.ListMiddlewares()),
			"info":        "Use /api/* para rotas protegidas",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Esta é uma rota pública, sem autenticação",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// ==============================
	// ROTAS PROTEGIDAS
	// ==============================

	api := router.Group("/api")
	{
		api.GET("/users", func(c *gin.Context) {
			time.Sleep(50 * time.Millisecond) // Simular processamento
			c.JSON(http.StatusOK, gin.H{
				"users": []map[string]interface{}{
					{"id": 1, "name": "João", "role": "admin"},
					{"id": 2, "name": "Maria", "role": "user"},
					{"id": 3, "name": "Pedro", "role": "user"},
				},
				"authenticated": true,
			})
		})

		api.POST("/users", func(c *gin.Context) {
			var user map[string]interface{}
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
				return
			}

			// Simular criação de usuário
			user["id"] = 4
			user["created_at"] = time.Now().Format(time.RFC3339)

			c.JSON(http.StatusCreated, gin.H{
				"message": "Usuário criado com sucesso",
				"user":    user,
			})
		})

		api.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"profile": map[string]interface{}{
					"user_id":     "current_user",
					"permissions": []string{"read", "write"},
					"login_time":  time.Now().Format(time.RFC3339),
				},
			})
		})

		api.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Área administrativa",
				"data":    "Dados sensíveis aqui",
				"admin":   true,
			})
		})
	}

	// ==============================
	// ROTA DE INFORMAÇÕES
	// ==============================

	router.GET("/info", func(c *gin.Context) {
		info := map[string]interface{}{
			"middlewares": map[string]interface{}{
				"total": len(middlewareManager.ListMiddlewares()),
				"list":  middlewareManager.ListMiddlewares(),
			},
			"authentication": map[string]interface{}{
				"type":            "Basic Auth",
				"users":           []string{"admin", "user"},
				"protected_paths": []string{"/api/*"},
				"public_paths":    []string{"/", "/health", "/public", "/info"},
			},
			"logging": map[string]interface{}{
				"enabled":     true,
				"skip_paths":  []string{"/health"},
				"log_headers": true,
				"log_body":    false,
			},
		}
		c.JSON(http.StatusOK, info)
	})

	// ==============================
	// INICIALIZAÇÃO DO SERVIDOR
	// ==============================

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Printf("🌟 Servidor iniciado na porta 8080")
		log.Printf("📊 Endpoints disponíveis:")
		log.Printf("")
		log.Printf("   🌍 PÚBLICOS:")
		log.Printf("   GET  /           - Página inicial")
		log.Printf("   GET  /health     - Health check")
		log.Printf("   GET  /public     - Rota pública")
		log.Printf("   GET  /info       - Informações do sistema")
		log.Printf("")
		log.Printf("   🔒 PROTEGIDOS (Basic Auth):")
		log.Printf("   GET  /api/users     - Lista de usuários")
		log.Printf("   POST /api/users     - Criar usuário")
		log.Printf("   GET  /api/profile   - Perfil do usuário")
		log.Printf("   GET  /api/admin     - Área administrativa")
		log.Printf("")
		log.Printf("🔐 Credenciais:")
		log.Printf("   admin:secret123")
		log.Printf("   user:password456")
		log.Printf("")
		log.Printf("🧪 Exemplos de uso:")
		log.Printf("   curl http://localhost:8080/")
		log.Printf("   curl http://localhost:8080/public")
		log.Printf("   curl -u admin:secret123 http://localhost:8080/api/users")
		log.Printf("   curl -u user:password456 http://localhost:8080/api/profile")

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Erro durante shutdown: %v", err)
	}

	log.Printf("✅ Servidor finalizado com sucesso")
	log.Printf("📊 Estatísticas:")
	log.Printf("   Middlewares registrados: %d", len(middlewareManager.ListMiddlewares()))
}
