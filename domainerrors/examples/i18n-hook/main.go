package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	domainInterfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

func main() {
	fmt.Println("=== Exemplo de Hook i18n Translation ===")

	// Configura diretório temporário para as traduções
	tempDir := setupTranslationFiles()
	defer os.RemoveAll(tempDir)

	// Cria hook de tradução
	i18nHook, err := hooks.NewI18nTranslationHook(
		domainInterfaces.HookTypeAfterError,
		hooks.I18nTranslationConfig{
			TranslationsPath: tempDir,
			DefaultLanguage:  "en",
			FallbackLanguage: "en",
			SupportedLangs:   []string{"en", "pt", "es"},
			FilePattern:      "{lang}.json",
			TranslateCodes:   false,
		},
	)
	if err != nil {
		log.Fatalf("Erro ao criar hook i18n: %v", err)
	}

	fmt.Printf("Hook criado: %s\n", i18nHook.Name())
	fmt.Printf("Idiomas suportados: %v\n", i18nHook.GetSupportedLanguages())

	// Exemplos de tradução com diferentes idiomas
	testTranslations(i18nHook, "en")
	testTranslations(i18nHook, "pt")
	testTranslations(i18nHook, "es")
	testTranslations(i18nHook, "fr") // idioma não suportado, deve usar fallback
}

// setupTranslationFiles cria arquivos de tradução temporários
func setupTranslationFiles() string {
	tempDir := os.TempDir() + "/nexs_i18n_test"
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
  "errors.password_weak": "Password must be at least 8 characters long"
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
  "errors.password_weak": "A senha deve ter pelo menos 8 caracteres"
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
  "errors.password_weak": "La contraseña debe tener al menos 8 caracteres"
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

// testTranslations testa traduções para um idioma específico
func testTranslations(i18nHook *hooks.I18nTranslationHook, lang string) {
	fmt.Printf("\n--- Testando traduções para idioma: %s ---\n", lang)

	// Cria contexto com idioma
	ctx := context.WithValue(context.Background(), "language", lang)

	// Testa diferentes tipos de erros
	testCases := []struct {
		code    string
		message string
		desc    string
	}{
		{"USER_NOT_FOUND", "User not found", "Erro com código traduzível"},
		{"VALIDATION_FAILED", "Validation failed", "Erro de validação"},
		{"UNKNOWN_CODE", "Some unknown error", "Código não traduzível"},
		{"", "not found", "Sem código, mensagem com palavras-chave"},
		{"", "validation error occurred", "Erro de validação genérico"},
	}

	for _, tc := range testCases {
		// Cria erro de domínio
		domainErr := &domainInterfaces.DomainError{
			Code:    tc.code,
			Message: tc.message,
			Type:    "validation",
		}

		// Cria contexto do hook
		hookCtx := &domainInterfaces.HookContext{
			Context:   ctx,
			Error:     domainErr,
			Operation: "test_translation",
		}

		// Executa o hook
		fmt.Printf("  Caso: %s\n", tc.desc)
		fmt.Printf("    Original: [%s] %s\n", tc.code, tc.message)

		if err := i18nHook.Execute(hookCtx); err != nil {
			fmt.Printf("    Erro na tradução: %v\n", err)
			continue
		}

		fmt.Printf("    Traduzido: [%s] %s\n", hookCtx.Error.Code, hookCtx.Error.Message)

		// Mostra metadados de tradução
		if hookCtx.Error.Metadata != nil {
			if origMsg, ok := hookCtx.Error.Metadata["original_message"]; ok {
				fmt.Printf("    Metadados: original='%s', lang='%s'\n",
					origMsg, hookCtx.Error.Metadata["translated_language"])
			}
		}
		fmt.Println()
	}
}
