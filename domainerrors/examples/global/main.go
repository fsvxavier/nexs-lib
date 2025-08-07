package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/middlewares"
)

func init() {
	// Registrar hooks globais
	hooks.RegisterGlobalStartHook(func(ctx context.Context) error {
		fmt.Println("🚀 Global Start Hook: Sistema iniciando...")
		return nil
	})

	hooks.RegisterGlobalStopHook(func(ctx context.Context) error {
		fmt.Println("🛑 Global Stop Hook: Sistema parando...")
		return nil
	})

	hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
		fmt.Printf("❌ Global Error Hook: %s [%s]\n", err.Error(), err.Code())
		return nil
	})

	hooks.RegisterGlobalI18nHook(func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
		fmt.Printf("🌐 Global I18n Hook: Locale %s - %s\n", locale, err.Error())
		return nil
	})

	// Registrar middlewares globais
	middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("🔧 Global Middleware: Processando erro %s\n", err.Code())

		// Adicionar timestamp aos metadados
		processed := err.WithMetadata("processed_at", "2024-12-14T20:30:00Z")

		return next(processed)
	})

	middlewares.RegisterGlobalI18nMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("🌍 Global I18n Middleware: Traduzindo para %s\n", locale)

		// Simular tradução baseada no locale
		var translatedMessage string
		switch locale {
		case "pt-BR":
			translatedMessage = "Erro traduzido para português"
		case "es-ES":
			translatedMessage = "Error traducido al español"
		default:
			translatedMessage = err.Error()
		}

		// Criar um novo erro com mensagem traduzida
		translated := domainerrors.NewWithMetadata(err.Type(), err.Code(), translatedMessage, err.Metadata())

		return next(translated)
	})
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Exemplo de Hooks e Middlewares Globais ===\n")

	// 1. Executar hooks de início
	fmt.Println("1. Executando hooks de início:")
	if err := hooks.ExecuteGlobalStartHooks(ctx); err != nil {
		log.Printf("Erro ao executar hooks de início: %v", err)
	}
	fmt.Print("\n")

	// 2. Criar um erro para demonstrar os middlewares
	fmt.Println("2. Criando erro para demonstrar middlewares:")
	originalError := domainerrors.NewValidationError(
		"VALIDATION_ERROR",
		"Campo obrigatório não informado",
	).WithMetadata("field", "email").WithMetadata("value", "")

	fmt.Printf("Erro original: %s\n", originalError.Error())

	// 3. Executar middlewares globais
	fmt.Println("\n3. Executando middlewares globais:")
	processedError := middlewares.ExecuteGlobalMiddlewares(ctx, originalError)

	// Mostrar erro processado
	fmt.Println("\nErro após middlewares:")
	errorJSON, _ := processedError.ToJSON()
	fmt.Println(string(errorJSON))

	// 4. Executar hooks de erro
	fmt.Println("\n4. Executando hooks de erro:")
	if err := hooks.ExecuteGlobalErrorHooks(ctx, processedError); err != nil {
		log.Printf("Erro ao executar hooks de erro: %v", err)
	}

	// 5. Demonstrar middleware de i18n
	fmt.Println("\n5. Demonstrando middleware de i18n:")
	locales := []string{"pt-BR", "es-ES", "en-US"}

	for _, locale := range locales {
		fmt.Printf("\nLocale: %s\n", locale)
		translatedError := middlewares.ExecuteGlobalI18nMiddlewares(ctx, originalError, locale)

		// Executar hook de i18n
		if err := hooks.ExecuteGlobalI18nHooks(ctx, translatedError, locale); err != nil {
			log.Printf("Erro ao executar hooks de i18n: %v", err)
		}

		fmt.Printf("Mensagem traduzida: %s\n", translatedError.Error())
	}

	// 6. Mostrar estatísticas dos hooks e middlewares globais
	fmt.Println("\n6. Estatísticas dos hooks e middlewares globais:")
	startHooks, stopHooks, errorHooks, i18nHooks := hooks.GetGlobalHookCounts()
	generalMiddlewares, i18nMiddlewares := middlewares.GetGlobalMiddlewareCounts()

	fmt.Printf("Hooks registrados - Start: %d, Stop: %d, Error: %d, I18n: %d\n",
		startHooks, stopHooks, errorHooks, i18nHooks)
	fmt.Printf("Middlewares registrados - Geral: %d, I18n: %d\n",
		generalMiddlewares, i18nMiddlewares)

	// 7. Executar hooks de parada
	fmt.Println("\n7. Executando hooks de parada:")
	if err := hooks.ExecuteGlobalStopHooks(ctx); err != nil {
		log.Printf("Erro ao executar hooks de parada: %v", err)
	}

	fmt.Println("\n=== Fim do exemplo ===")
}
