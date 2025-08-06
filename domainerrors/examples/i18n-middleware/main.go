package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	domainInterfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/middlewares"
)

func main() {
	fmt.Println("=== Exemplo de Middleware i18n Translation ===")

	// Configura diretório temporário para as traduções
	tempDir := setupTranslationFiles()
	defer os.RemoveAll(tempDir)

	// Cria middleware de tradução
	i18nMiddleware, err := middlewares.NewI18nTranslationMiddleware(
		middlewares.I18nTranslationConfig{
			TranslationsPath:  tempDir,
			DefaultLanguage:   "en",
			FallbackLanguage:  "en",
			SupportedLangs:    []string{"en", "pt", "es"},
			FilePattern:       "{lang}.json",
			TranslateCodes:    true,
			TranslateMetadata: true,
		},
	)
	if err != nil {
		log.Fatalf("Erro ao criar middleware i18n: %v", err)
	}

	fmt.Printf("Middleware criado: %s\n", i18nMiddleware.Name())
	fmt.Printf("Idiomas suportados: %v\n", i18nMiddleware.GetSupportedLanguages())
	fmt.Printf("Estatísticas: %+v\n", i18nMiddleware.GetTranslationStats())

	// Testa o middleware com diferentes cenários
	testMiddlewareChain(i18nMiddleware, "en")
	testMiddlewareChain(i18nMiddleware, "pt")
	testMiddlewareChain(i18nMiddleware, "es")
	testMiddlewareChain(i18nMiddleware, "fr") // idioma não suportado
}

// setupTranslationFiles cria arquivos de tradução temporários
func setupTranslationFiles() string {
	tempDir := os.TempDir() + "/nexs_i18n_middleware_test"
	os.MkdirAll(tempDir, 0755)

	// Arquivo de traduções em inglês
	enTranslations := `{
  "USER_NOT_FOUND": "User not found",
  "VALIDATION_FAILED": "Validation failed",
  "UNAUTHORIZED_ACCESS": "Unauthorized access",
  "INTERNAL_ERROR": "Internal server error",
  "error.not_found": "Resource not found",
  "error.validation": "Validation error",
  "error.unauthorized": "Access denied",
  "error.internal": "Internal error occurred",
  "errors.user_not_found": "The requested user could not be found",
  "errors.email_invalid": "Please provide a valid email address",
  "errors.password_weak": "Password must be at least 8 characters long",
  "code.usr_404": "USER_NOT_FOUND",
  "code.val_400": "VALIDATION_FAILED",
  "validation_message": "Field validation failed",
  "business_rule_message": "Business rule violation",
  "operation_description": "Processing user data"
}`

	// Arquivo de traduções em português
	ptTranslations := `{
  "USER_NOT_FOUND": "Usuário não encontrado",
  "VALIDATION_FAILED": "Falha na validação",
  "UNAUTHORIZED_ACCESS": "Acesso não autorizado",
  "INTERNAL_ERROR": "Erro interno do servidor",
  "error.not_found": "Recurso não encontrado",
  "error.validation": "Erro de validação",
  "error.unauthorized": "Acesso negado",
  "error.internal": "Erro interno ocorreu",
  "errors.user_not_found": "O usuário solicitado não pôde ser encontrado",
  "errors.email_invalid": "Por favor, forneça um endereço de email válido",
  "errors.password_weak": "A senha deve ter pelo menos 8 caracteres",
  "code.usr_404": "USUARIO_NAO_ENCONTRADO",
  "code.val_400": "FALHA_VALIDACAO",
  "validation_message": "Falha na validação do campo",
  "business_rule_message": "Violação de regra de negócio",
  "operation_description": "Processando dados do usuário"
}`

	// Arquivo de traduções em espanhol
	esTranslations := `{
  "USER_NOT_FOUND": "Usuario no encontrado",
  "VALIDATION_FAILED": "Validación fallida",
  "UNAUTHORIZED_ACCESS": "Acceso no autorizado",
  "INTERNAL_ERROR": "Error interno del servidor",
  "error.not_found": "Recurso no encontrado",
  "error.validation": "Error de validación",
  "error.unauthorized": "Acceso denegado",
  "error.internal": "Se produjo un error interno",
  "errors.user_not_found": "El usuario solicitado no pudo ser encontrado",
  "errors.email_invalid": "Por favor, proporcione una dirección de correo válida",
  "errors.password_weak": "La contraseña debe tener al menos 8 caracteres",
  "code.usr_404": "USUARIO_NO_ENCONTRADO",
  "code.val_400": "VALIDACION_FALLIDA",
  "validation_message": "Falló la validación del campo",
  "business_rule_message": "Violación de regla de negocio",
  "operation_description": "Procesando datos del usuario"
}`

	// Escreve os arquivos
	writeFile(filepath.Join(tempDir, "en.json"), enTranslations)
	writeFile(filepath.Join(tempDir, "pt.json"), ptTranslations)
	writeFile(filepath.Join(tempDir, "es.json"), esTranslations)

	return tempDir
}

// writeFile escreve conteúdo em um arquivo
func writeFile(filePath, content string) {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Printf("Erro ao escrever arquivo %s: %v", filePath, err)
	}
}

// testMiddlewareChain testa o middleware para um idioma específico
func testMiddlewareChain(middleware *middlewares.I18nTranslationMiddleware, lang string) {
	fmt.Printf("\n--- Testando middleware para idioma: %s ---\n", lang)

	// Cria contexto com idioma
	ctx := context.WithValue(context.Background(), "language", lang)

	// Casos de teste
	testCases := []struct {
		code        string
		message     string
		metadata    map[string]interface{}
		description string
	}{
		{
			code:    "USER_NOT_FOUND",
			message: "User not found",
			metadata: map[string]interface{}{
				"validation_message": "Field validation failed",
				"user_id":            12345,
			},
			description: "Erro com código e metadados traduzíveis",
		},
		{
			code:    "VAL_400",
			message: "Validation failed",
			metadata: map[string]interface{}{
				"business_rule_message": "Business rule violation",
				"validation_errors":     []string{"validation_message", "Field validation failed"},
			},
			description: "Erro de validação com array de mensagens",
		},
		{
			code:    "UNKNOWN_ERROR",
			message: "Some unknown error occurred",
			metadata: map[string]interface{}{
				"operation_description": "Processing user data",
				"step":                  "validation",
			},
			description: "Erro desconhecido com metadados contextuais",
		},
		{
			code:    "",
			message: "not found error",
			metadata: map[string]interface{}{
				"resource_type": "user",
				"resource_id":   "123",
			},
			description: "Erro sem código, detectado por palavras-chave",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("  %d. %s\n", i+1, tc.description)
		fmt.Printf("     Original: [%s] %s\n", tc.code, tc.message)

		// Cria erro de domínio
		domainErr := &domainInterfaces.DomainError{
			Code:      tc.code,
			Message:   tc.message,
			Type:      "validation",
			Metadata:  tc.metadata,
			Timestamp: time.Now(),
		}

		// Cria contexto do middleware
		middlewareCtx := &domainInterfaces.MiddlewareContext{
			Context:   ctx,
			Error:     domainErr,
			Operation: "test_translation",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"operation_description": "Processing user data",
				"request_id":            "req-123",
			},
		}

		// Simula próximo middleware na cadeia (dummy)
		nextFunc := func(ctx *domainInterfaces.MiddlewareContext) error {
			fmt.Printf("     [Next Middleware] Processando erro traduzido...\n")
			return nil
		}

		// Executa o middleware
		if err := middleware.Handle(middlewareCtx, nextFunc); err != nil {
			fmt.Printf("     Erro no middleware: %v\n", err)
			continue
		}

		fmt.Printf("     Traduzido: [%s] %s\n", middlewareCtx.Error.Code, middlewareCtx.Error.Message)

		// Mostra metadados de tradução
		if middlewareCtx.Error.Metadata != nil {
			if origMsg, ok := middlewareCtx.Error.Metadata["original_message"]; ok {
				fmt.Printf("     Metadados erro: original='%s', lang='%s'\n",
					origMsg, middlewareCtx.Error.Metadata["translated_language"])
			}

			// Mostra outros metadados traduzidos
			for key, value := range middlewareCtx.Error.Metadata {
				if strings.HasSuffix(key, "_original") {
					originalKey := strings.TrimSuffix(key, "_original")
					if translatedValue, exists := middlewareCtx.Error.Metadata[originalKey]; exists {
						fmt.Printf("     Metadado traduzido: %s='%v' (original='%v')\n",
							originalKey, translatedValue, value)
					}
				}
			}
		}

		// Mostra metadados do contexto traduzidos
		if middlewareCtx.Metadata != nil {
			for key, value := range middlewareCtx.Metadata {
				if strings.HasSuffix(key, "_original") {
					originalKey := strings.TrimSuffix(key, "_original")
					if translatedValue, exists := middlewareCtx.Metadata[originalKey]; exists {
						fmt.Printf("     Contexto traduzido: %s='%v' (original='%v')\n",
							originalKey, translatedValue, value)
					}
				}
			}
		}

		fmt.Println()
	}
}
