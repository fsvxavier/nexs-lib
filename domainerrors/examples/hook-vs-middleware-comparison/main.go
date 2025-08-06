package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	domainInterfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/middlewares"
)

func main() {
	fmt.Println("=== Comparação: Hook vs Middleware para Tradução i18n ===")
	fmt.Println()

	// Configura diretório temporário para as traduções
	tempDir := setupTranslationFiles()
	defer os.RemoveAll(tempDir)

	fmt.Println("🪝 HOOK - Event-Driven (Side Effects)")
	fmt.Println("   • Reage a eventos específicos")
	fmt.Println("   • NÃO modifica o erro diretamente")
	fmt.Println("   • Usado para logging, auditoria")
	fmt.Println()

	fmt.Println("🔧 MIDDLEWARE - Processing Pipeline (Transformation)")
	fmt.Println("   • Transforma/enriquece erros")
	fmt.Println("   • MODIFICA o erro diretamente")
	fmt.Println("   • Chain of Responsibility pattern")
	fmt.Println()

	// Demonstra diferenças práticas
	demonstrateHookApproach(tempDir)
	fmt.Println()
	demonstrateMiddlewareApproach(tempDir)
	fmt.Println()

	// Comparação lado a lado
	compareBothApproaches(tempDir)
}

// demonstrateHookApproach mostra como funciona o hook
func demonstrateHookApproach(tempDir string) {
	fmt.Println("--- 🪝 ABORDAGEM HOOK ---")

	// Cria hook de tradução
	i18nHook, err := hooks.NewI18nTranslationHook(
		domainInterfaces.HookTypeAfterError,
		hooks.I18nTranslationConfig{
			TranslationsPath: tempDir,
			DefaultLanguage:  "en",
			FallbackLanguage: "en",
			SupportedLangs:   []string{"en", "pt"},
			FilePattern:      "{lang}.json",
		},
	)
	if err != nil {
		log.Fatalf("Erro ao criar hook: %v", err)
	}

	fmt.Printf("✅ Hook criado: %s\n", i18nHook.Name())

	// Simula uso do hook (evento-driven)
	ctx := context.WithValue(context.Background(), "language", "pt")

	originalError := &domainInterfaces.DomainError{
		Code:     "USER_NOT_FOUND",
		Message:  "User not found",
		Type:     "not_found",
		Metadata: make(map[string]interface{}),
	}

	fmt.Printf("📥 Erro original: [%s] %s\n", originalError.Code, originalError.Message)

	hookCtx := &domainInterfaces.HookContext{
		Context:   ctx,
		Error:     originalError,
		Operation: "user_lookup",
		Timestamp: time.Now(),
	}

	// Hook executa (side effect - logging/auditoria com tradução)
	fmt.Println("🔄 Hook executando...")
	if err := i18nHook.Execute(hookCtx); err != nil {
		fmt.Printf("❌ Erro no hook: %v\n", err)
		return
	}

	// Hook MODIFICOU o erro (para fins de demonstração)
	fmt.Printf("📤 Erro após hook: [%s] %s\n", hookCtx.Error.Code, hookCtx.Error.Message)
	fmt.Printf("📊 Metadados: %+v\n", hookCtx.Error.Metadata)
	fmt.Println("⚠️  Nota: Hook modificou o erro para demonstração, mas tipicamente seria usado para logging/auditoria")
}

// demonstrateMiddlewareApproach mostra como funciona o middleware
func demonstrateMiddlewareApproach(tempDir string) {
	fmt.Println("--- 🔧 ABORDAGEM MIDDLEWARE ---")

	// Cria middleware de tradução
	i18nMiddleware, err := middlewares.NewI18nTranslationMiddleware(
		middlewares.I18nTranslationConfig{
			TranslationsPath:  tempDir,
			DefaultLanguage:   "en",
			FallbackLanguage:  "en",
			SupportedLangs:    []string{"en", "pt"},
			FilePattern:       "{lang}.json",
			TranslateCodes:    true,
			TranslateMetadata: true,
		},
	)
	if err != nil {
		log.Fatalf("Erro ao criar middleware: %v", err)
	}

	fmt.Printf("✅ Middleware criado: %s\n", i18nMiddleware.Name())

	// Simula uso do middleware (processing pipeline)
	ctx := context.WithValue(context.Background(), "language", "pt")

	originalError := &domainInterfaces.DomainError{
		Code:    "USER_NOT_FOUND",
		Message: "User not found",
		Type:    "not_found",
		Metadata: map[string]interface{}{
			"validation_message": "Field validation failed",
		},
	}

	fmt.Printf("📥 Erro original: [%s] %s\n", originalError.Code, originalError.Message)

	middlewareCtx := &domainInterfaces.MiddlewareContext{
		Context:   ctx,
		Error:     originalError,
		Operation: "user_lookup",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"operation_description": "Processing user data",
		},
	}

	// Simula cadeia de middlewares
	nextMiddleware := func(ctx *domainInterfaces.MiddlewareContext) error {
		fmt.Println("📦 Próximo middleware na cadeia executou")
		return nil
	}

	// Middleware executa (transformação na cadeia)
	fmt.Println("🔄 Middleware executando na cadeia...")
	if err := i18nMiddleware.Handle(middlewareCtx, nextMiddleware); err != nil {
		fmt.Printf("❌ Erro no middleware: %v\n", err)
		return
	}

	// Middleware TRANSFORMOU o erro
	fmt.Printf("📤 Erro após middleware: [%s] %s\n", middlewareCtx.Error.Code, middlewareCtx.Error.Message)
	fmt.Printf("📊 Metadados erro: %+v\n", middlewareCtx.Error.Metadata)
	fmt.Printf("🌐 Metadados contexto: %+v\n", middlewareCtx.Metadata)
	fmt.Println("✅ Middleware transformou o erro como parte da cadeia de processamento")
}

// compareBothApproaches compara as duas abordagens lado a lado
func compareBothApproaches(tempDir string) {
	fmt.Println("--- 📊 COMPARAÇÃO LADO A LADO ---")

	// Prepara contexto comum
	ctx := context.WithValue(context.Background(), "language", "pt")

	// Erro original (mesmo para ambos)
	createOriginalError := func() *domainInterfaces.DomainError {
		return &domainInterfaces.DomainError{
			Code:    "VALIDATION_FAILED",
			Message: "Validation failed",
			Type:    "validation",
			Metadata: map[string]interface{}{
				"validation_message": "Field validation failed",
				"field_name":         "email",
			},
		}
	}

	fmt.Println("🆚 Processando mesmo erro com ambas abordagens:")
	fmt.Println()

	// HOOK
	fmt.Println("🪝 HOOK RESULT:")
	hookError := createOriginalError()
	hookCtx := &domainInterfaces.HookContext{
		Context:   ctx,
		Error:     hookError,
		Operation: "validation",
	}

	hook, _ := hooks.NewI18nTranslationHook(
		domainInterfaces.HookTypeAfterError,
		hooks.I18nTranslationConfig{
			TranslationsPath: tempDir,
			DefaultLanguage:  "en",
			SupportedLangs:   []string{"en", "pt"},
			FilePattern:      "{lang}.json",
		},
	)

	hook.Execute(hookCtx)
	fmt.Printf("   Mensagem: %s\n", hookCtx.Error.Message)
	fmt.Printf("   Código: %s\n", hookCtx.Error.Code)
	fmt.Printf("   Fonte: %v\n", hookCtx.Error.Metadata["translation_source"])

	fmt.Println()

	// MIDDLEWARE
	fmt.Println("🔧 MIDDLEWARE RESULT:")
	middlewareError := createOriginalError()
	middlewareCtx := &domainInterfaces.MiddlewareContext{
		Context:   ctx,
		Error:     middlewareError,
		Operation: "validation",
		Metadata: map[string]interface{}{
			"operation_description": "Processing user data",
		},
	}

	middleware, _ := middlewares.NewI18nTranslationMiddleware(
		middlewares.I18nTranslationConfig{
			TranslationsPath:  tempDir,
			DefaultLanguage:   "en",
			SupportedLangs:    []string{"en", "pt"},
			FilePattern:       "{lang}.json",
			TranslateCodes:    true,
			TranslateMetadata: true,
		},
	)

	middleware.Handle(middlewareCtx, nil)
	fmt.Printf("   Mensagem: %s\n", middlewareCtx.Error.Message)
	fmt.Printf("   Código: %s\n", middlewareCtx.Error.Code)
	fmt.Printf("   Fonte: %v\n", middlewareCtx.Error.Metadata["translation_source"])
	fmt.Printf("   Metadados traduzidos: %v\n", middlewareCtx.Error.Metadata["validation_message"])

	fmt.Println()
	fmt.Println("💡 RESUMO DAS DIFERENÇAS:")
	fmt.Println("   🪝 Hook: Melhor para side effects (logging, auditoria, notificações)")
	fmt.Println("   🔧 Middleware: Melhor para transformações (enriquecimento, tradução, validação)")
	fmt.Println("   🔗 Middleware permite composição em cadeias complexas")
	fmt.Println("   ⚡ Hook executa em eventos específicos do ciclo de vida")
}

// setupTranslationFiles cria arquivos de tradução temporários
func setupTranslationFiles() string {
	tempDir := os.TempDir() + "/nexs_i18n_comparison_test"
	os.MkdirAll(tempDir, 0755)

	enTranslations := `{
  "USER_NOT_FOUND": "User not found",
  "VALIDATION_FAILED": "Validation failed",
  "validation_message": "Field validation failed",
  "operation_description": "Processing user data",
  "code.user_not_found": "USER_NOT_FOUND",
  "code.validation_failed": "VALIDATION_FAILED"
}`

	ptTranslations := `{
  "USER_NOT_FOUND": "Usuário não encontrado",
  "VALIDATION_FAILED": "Falha na validação", 
  "validation_message": "Falha na validação do campo",
  "operation_description": "Processando dados do usuário",
  "code.user_not_found": "USUARIO_NAO_ENCONTRADO",
  "code.validation_failed": "FALHA_VALIDACAO"
}`

	writeFile(filepath.Join(tempDir, "en.json"), enTranslations)
	writeFile(filepath.Join(tempDir, "pt.json"), ptTranslations)

	return tempDir
}

// writeFile escreve conteúdo em um arquivo
func writeFile(filePath, content string) {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Printf("Erro ao escrever arquivo %s: %v", filePath, err)
	}
}
