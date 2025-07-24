// Package main demonstrates comprehensive usage of the httpserver library
package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

func main() {
	fmt.Println("🚀 Demonstração do uso customizado do httpserver")
	fmt.Println("Este exemplo demonstra hooks customizados e middleware")

	// Executar exemplos
	fmt.Println("\n1. Demonstrando Custom Hooks...")
	demonstrateCustomHooks()

	fmt.Println("\n2. Demonstrando Custom Middleware...")
	demonstrateCustomMiddleware()

	fmt.Println("\n3. Demonstrando Middleware Chain...")
	demonstrateMiddlewareChain()

	fmt.Println("\n✅ Todos os exemplos executados com sucesso!")
}

// demonstrateCustomHooks shows how to create and use custom hooks
func demonstrateCustomHooks() {
	// Create a custom hook factory
	hookFactory := hooks.NewCustomHookFactory()

	// Create sample hooks using factory methods
	logHook := hookFactory.NewSimpleHook(
		"request-logger",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart, interfaces.HookEventRequestEnd},
		100,
		func(ctx *interfaces.HookContext) error {
			fmt.Printf("    [LOG] Hook request-logger executado\n")
			return nil
		},
	)

	metricHook := hookFactory.NewSimpleHook(
		"metrics-collector",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		200,
		func(ctx *interfaces.HookContext) error {
			fmt.Printf("    [METRICS] Hook metrics-collector coletando métricas\n")
			return nil
		},
	)

	// Execute hooks for demonstration
	fmt.Printf("  ✓ Criados 2 hooks customizados com factory\n")

	// Simulate hook execution
	fmt.Printf("  ✓ Executando hook: %s\n", logHook.Name())
	logHook.Execute(nil) // nil context for demo

	fmt.Printf("  ✓ Executando hook: %s\n", metricHook.Name())
	metricHook.Execute(nil) // nil context for demo
} // demonstrateCustomMiddleware shows custom middleware creation
func demonstrateCustomMiddleware() {
	// Create custom middleware
	customMiddleware := &CustomMiddleware{name: "demo-middleware"}

	fmt.Printf("  ✓ Middleware criado: %s (prioridade: %d)\n",
		customMiddleware.Name(), customMiddleware.Priority())

	// Create a sample handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from wrapped handler"))
	})

	// Wrap handler with custom middleware
	wrappedHandler := customMiddleware.Wrap(handler)
	fmt.Printf("  ✓ Handler envolvido com middleware customizado\n")
	_ = wrappedHandler // Avoid unused variable warning
}

// demonstrateMiddlewareChain shows middleware chain usage
func demonstrateMiddlewareChain() {
	// Create middleware chain
	chain := middleware.NewChain()

	// Add multiple middlewares
	chain.Add(&CustomMiddleware{name: "auth-middleware"})
	chain.Add(&CustomMiddleware{name: "logging-middleware"})
	chain.Add(&CustomMiddleware{name: "compression-middleware"})

	fmt.Printf("  ✓ Chain criada com %d middlewares\n", 3)

	// Create final handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final handler response"))
	})

	// Apply chain
	chainedHandler := chain.Then(finalHandler)
	fmt.Printf("  ✓ Chain aplicada ao handler final\n")
	_ = chainedHandler // Avoid unused variable warning
}

// CustomMiddleware is a sample middleware implementation
type CustomMiddleware struct {
	name string
}

func (m *CustomMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pre-processing
		start := time.Now()
		fmt.Printf("    [%s] Requisição iniciada: %s %s\n",
			strings.ToUpper(m.name), r.Method, r.URL.Path)

		// Call next handler
		next.ServeHTTP(w, r)

		// Post-processing
		duration := time.Since(start)
		fmt.Printf("    [%s] Requisição finalizada em %v\n",
			strings.ToUpper(m.name), duration)
	})
}

func (m *CustomMiddleware) Name() string {
	return m.name
}

func (m *CustomMiddleware) Priority() int {
	// Return different priorities based on name for demonstration
	switch m.name {
	case "auth-middleware":
		return 100
	case "logging-middleware":
		return 200
	case "compression-middleware":
		return 300
	default:
		return 500
	}
}
