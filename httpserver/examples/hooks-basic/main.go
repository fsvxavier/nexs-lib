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
	"github.com/gin-gonic/gin"
)

// Exemplo básico demonstrando como usar hooks para monitoramento

func main() {
	log.Println("🚀 Exemplo Básico - Sistema de Hooks")

	// ==============================
	// CONFIGURAÇÃO DE HOOKS
	// ==============================

	hookManager := hooks.NewHookManager()

	// Hook de ciclo de vida do servidor
	startHook := hooks.NewStartHook("server-lifecycle")
	startHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("start", startHook)

	stopHook := hooks.NewStopHook("server-lifecycle")
	stopHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("stop", stopHook)

	// Hook de rastreamento de requisições
	requestHook := hooks.NewRequestHook("request-tracker")
	requestHook.SetMetricsEnabled(true)
	hookManager.RegisterHook("request", requestHook)

	// Hook de rastreamento de erros
	errorHook := hooks.NewErrorHook("error-tracker")
	errorHook.SetMetricsEnabled(true)
	errorHook.SetErrorThreshold(3) // Alertar após 3 erros
	hookManager.RegisterHook("error", errorHook)

	log.Printf("✅ %d hooks registrados", len(hookManager.ListHooks()))

	// ==============================
	// CONFIGURAÇÃO DO SERVIDOR
	// ==============================

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware que integra os hooks
	router.Use(func(c *gin.Context) {
		ctx := c.Request.Context()

		// Hook de entrada da requisição
		requestHook.OnRequest(ctx, c.Request)

		c.Next()

		// Hook de resposta (simples)
		duration := time.Since(time.Now())
		requestHook.OnResponse(ctx, c.Request, c.Writer, duration)
	})

	// ==============================
	// ROTAS DE EXEMPLO
	// ==============================

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Bem-vindo ao exemplo básico de hooks!",
			"hooks":   len(hookManager.ListHooks()),
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	router.GET("/users", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond) // Simular processamento
		c.JSON(http.StatusOK, gin.H{
			"users": []map[string]interface{}{
				{"id": 1, "name": "João"},
				{"id": 2, "name": "Maria"},
			},
		})
	})

	router.GET("/error", func(c *gin.Context) {
		ctx := c.Request.Context()
		err := &CustomError{Msg: "Erro simulado para demonstração", Code: "DEMO_ERROR"}
		errorHook.OnError(ctx, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro interno do servidor",
		})
	})

	router.GET("/metrics", func(c *gin.Context) {
		metrics := map[string]interface{}{
			"hooks": map[string]interface{}{
				"registered":      len(hookManager.ListHooks()),
				"start_count":     startHook.GetStartCount(),
				"stop_count":      stopHook.GetStopCount(),
				"request_count":   requestHook.GetRequestCount(),
				"active_requests": requestHook.GetActiveRequestCount(),
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

	// Notificar início do servidor
	ctx := context.Background()
	startHook.OnStart(ctx, ":8080")

	go func() {
		log.Printf("🌟 Servidor iniciado na porta 8080")
		log.Printf("📊 Endpoints disponíveis:")
		log.Printf("   GET  /           - Página inicial")
		log.Printf("   GET  /health     - Health check")
		log.Printf("   GET  /users      - Lista de usuários")
		log.Printf("   GET  /error      - Simular erro")
		log.Printf("   GET  /metrics    - Métricas dos hooks")
		log.Printf("")
		log.Printf("🧪 Teste com: curl http://localhost:8080/")

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopHook.OnStop(shutdownCtx)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Erro durante shutdown: %v", err)
		errorHook.OnError(shutdownCtx, err)
	}

	log.Printf("✅ Servidor finalizado com sucesso")
	log.Printf("📊 Estatísticas Finais:")
	log.Printf("   Total de requisições: %d", requestHook.GetRequestCount())
	log.Printf("   Inicializações: %d", startHook.GetStartCount())
	log.Printf("   Paradas: %d", stopHook.GetStopCount())
}

// CustomError implementa uma estrutura de erro personalizada
type CustomError struct {
	Msg  string
	Code string
}

func (e *CustomError) Error() string {
	return e.Msg
}
