// Package main demonstrates a CLI tool using i18n library.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

// CLIConfig represents CLI configuration
type CLIConfig struct {
	Language       string
	TranslationDir string
	Command        string
	Interactive    bool
	Verbose        bool
}

// CLITool represents the i18n CLI tool
type CLITool struct {
	config   CLIConfig
	provider interfaces.I18n
}

func main() {
	// Parse command line arguments
	config := parseFlags()

	// Show usage if no command provided
	if config.Command == "" && !config.Interactive {
		showUsage()
		return
	}

	// Create CLI tool
	tool, err := NewCLITool(config)
	if err != nil {
		log.Fatal("Failed to initialize CLI tool:", err)
	}
	defer tool.provider.Stop(context.Background())

	// Run the tool
	if config.Interactive {
		tool.runInteractive()
	} else {
		tool.runCommand(config.Command)
	}
}

func parseFlags() CLIConfig {
	var config CLIConfig

	flag.StringVar(&config.Language, "lang", "en", "Language for translations (en, pt, es)")
	flag.StringVar(&config.TranslationDir, "dir", "", "Directory containing translation files")
	flag.StringVar(&config.Command, "cmd", "", "Command to run (translate, list-keys, validate)")
	flag.BoolVar(&config.Interactive, "interactive", false, "Run in interactive mode")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")

	flag.Parse()

	return config
}

func showUsage() {
	fmt.Println("🌍 I18n CLI Tool")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -lang string     Language for translations (default: en)")
	fmt.Println("  -dir string      Directory containing translation files")
	fmt.Println("  -cmd string      Command to run (translate, list-keys, validate)")
	fmt.Println("  -interactive     Run in interactive mode")
	fmt.Println("  -verbose         Enable verbose output")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  translate        Translate a specific key")
	fmt.Println("  list-keys        List all available translation keys")
	fmt.Println("  validate         Validate translation files")
	fmt.Println("  stats            Show translation statistics")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go -interactive")
	fmt.Println("  go run main.go -cmd translate -lang pt")
	fmt.Println("  go run main.go -cmd list-keys -dir ./translations")
	fmt.Println("  go run main.go -cmd validate -verbose")
}

func NewCLITool(cliConfig CLIConfig) (*CLITool, error) {
	// Determine translation directory
	translationDir := cliConfig.TranslationDir
	if translationDir == "" {
		// Create temporary directory with sample translations
		tempDir, err := os.MkdirTemp("", "i18n_cli_tool")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		if err := createSampleTranslations(tempDir); err != nil {
			return nil, fmt.Errorf("failed to create sample translations: %w", err)
		}

		translationDir = tempDir
		if cliConfig.Verbose {
			fmt.Printf("📁 Using temporary translation directory: %s\n", translationDir)
		}
	}

	// Configure i18n
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage(cliConfig.Language).
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(true, 5*time.Minute).
		WithLoadTimeout(10 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     translationDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	// Create registry and register provider
	registry := i18n.NewRegistry()
	jsonFactory := &json.Factory{}
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Create provider
	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Start provider
	ctx := context.Background()
	if err := provider.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start provider: %w", err)
	}

	tool := &CLITool{
		config:   cliConfig,
		provider: provider,
	}

	return tool, nil
}

func (tool *CLITool) runCommand(command string) {
	ctx := context.Background()

	switch command {
	case "translate":
		tool.runTranslateCommand(ctx)
	case "list-keys":
		tool.runListKeysCommand(ctx)
	case "validate":
		tool.runValidateCommand(ctx)
	case "stats":
		tool.runStatsCommand(ctx)
	default:
		fmt.Printf("❌ Unknown command '%s'. Use -h for help.\n", command)
		os.Exit(1)
	}
}

func (tool *CLITool) runInteractive() {
	fmt.Println("🌍 I18n CLI Tool - Interactive Mode")
	fmt.Println("===================================")
	fmt.Printf("🔤 Current language: %s\n", tool.config.Language)
	fmt.Printf("📁 Translation directory: %s\n", tool.config.TranslationDir)
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  1. translate <key> [params]  - Translate a key")
	fmt.Println("  2. list-keys                 - List all keys")
	fmt.Println("  3. validate                  - Validate translations")
	fmt.Println("  4. stats                     - Show statistics")
	fmt.Println("  5. change-lang <lang>        - Change language")
	fmt.Println("  6. help                      - Show this help")
	fmt.Println("  7. exit                      - Exit the tool")
	fmt.Println()

	ctx := context.Background()

	for {
		fmt.Print("i18n> ")

		var input string
		fmt.Scanln(&input)

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "translate", "t":
			tool.handleInteractiveTranslate(ctx, args)
		case "list-keys", "list", "ls":
			tool.runListKeysCommand(ctx)
		case "validate", "val":
			tool.runValidateCommand(ctx)
		case "stats", "st":
			tool.runStatsCommand(ctx)
		case "change-lang", "lang":
			tool.handleChangeLang(args)
		case "help", "h":
			tool.showInteractiveHelp()
		case "exit", "quit", "q":
			fmt.Println("👋 Goodbye!")
			return
		default:
			fmt.Printf("❌ Unknown command '%s'. Type 'help' for available commands.\n", command)
		}
	}
}

func (tool *CLITool) runTranslateCommand(ctx context.Context) {
	fmt.Print("🔤 Enter translation key: ")
	var key string
	fmt.Scanln(&key)

	if key == "" {
		fmt.Println("❌ Translation key is required")
		return
	}

	// Check for parameters
	fmt.Print("📝 Enter parameters (JSON format, optional): ")
	var paramsInput string
	fmt.Scanln(&paramsInput)

	var params map[string]interface{}
	if paramsInput != "" {
		// Simple parameter parsing for demo (in real tool would use proper JSON parsing)
		params = make(map[string]interface{})
		fmt.Printf("⚠️  Parameter parsing not implemented in this demo, using empty params\n")
	}

	result, err := tool.provider.Translate(ctx, key, tool.config.Language, params)
	if err != nil {
		fmt.Printf("❌ Translation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Translation: %s\n", result)
}

func (tool *CLITool) runListKeysCommand(ctx context.Context) {
	fmt.Printf("📋 Available translation keys for language '%s':\n", tool.config.Language)
	fmt.Println("================================================")

	// Sample keys based on our translation files
	keys := []string{
		"cli.welcome",
		"cli.goodbye",
		"cli.help.usage",
		"commands.translate",
		"commands.validate",
		"commands.stats",
		"errors.not_found",
		"errors.invalid_params",
		"success.operation_completed",
		"user.profile.title",
		"user.settings.language",
	}

	for i, key := range keys {
		translation, err := tool.provider.Translate(ctx, key, tool.config.Language, nil)
		if err != nil {
			fmt.Printf("  %d. %s (❌ %v)\n", i+1, key, err)
		} else {
			fmt.Printf("  %d. %s = %s\n", i+1, key, translation)
		}
	}

	fmt.Printf("\n📊 Total keys checked: %d\n", len(keys))
}

func (tool *CLITool) runValidateCommand(ctx context.Context) {
	fmt.Println("🔍 Validating translation files...")
	fmt.Println("==================================")

	supportedLangs := tool.provider.GetSupportedLanguages()

	for _, lang := range supportedLangs {
		fmt.Printf("🔤 Validating %s translations...\n", lang)

		count := tool.provider.GetTranslationCountByLanguage(lang)
		if count > 0 {
			fmt.Printf("  ✅ Found %d translations\n", count)
		} else {
			fmt.Printf("  ⚠️  No translations found\n")
		}
	}

	// Check provider health
	if err := tool.provider.Health(ctx); err != nil {
		fmt.Printf("❌ Provider health check failed: %v\n", err)
	} else {
		fmt.Println("✅ Provider health check passed")
	}

	fmt.Println("\n🏁 Validation completed")
}

func (tool *CLITool) runStatsCommand(ctx context.Context) {
	fmt.Println("📊 Translation Statistics")
	fmt.Println("========================")

	fmt.Printf("🔤 Current language: %s\n", tool.config.Language)
	fmt.Printf("🌍 Supported languages: %v\n", tool.provider.GetSupportedLanguages())
	fmt.Printf("📁 Default language: %s\n", tool.provider.GetDefaultLanguage())
	fmt.Printf("💾 Loaded languages: %v\n", tool.provider.GetLoadedLanguages())
	fmt.Printf("📈 Total translations: %d\n", tool.provider.GetTranslationCount())

	fmt.Println("\n📋 Translations by language:")
	for _, lang := range tool.provider.GetSupportedLanguages() {
		count := tool.provider.GetTranslationCountByLanguage(lang)
		fmt.Printf("  %s: %d translations\n", lang, count)
	}

	// Health check
	if err := tool.provider.Health(ctx); err != nil {
		fmt.Printf("\n❌ Provider status: Unhealthy (%v)\n", err)
	} else {
		fmt.Println("\n✅ Provider status: Healthy")
	}
}

func (tool *CLITool) handleInteractiveTranslate(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: translate <key> [params]")
		return
	}

	key := args[0]
	var params map[string]interface{}

	// Simple parameter handling for demo
	if len(args) > 1 {
		params = make(map[string]interface{})
		for i := 1; i < len(args); i++ {
			params[fmt.Sprintf("param%d", i)] = args[i]
		}
	}

	result, err := tool.provider.Translate(ctx, key, tool.config.Language, params)
	if err != nil {
		fmt.Printf("❌ Translation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ %s\n", result)
}

func (tool *CLITool) handleChangeLang(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: change-lang <language>")
		fmt.Printf("Available languages: %v\n", tool.provider.GetSupportedLanguages())
		return
	}

	newLang := args[0]
	supportedLangs := tool.provider.GetSupportedLanguages()

	// Check if language is supported
	found := false
	for _, lang := range supportedLangs {
		if lang == newLang {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ Language '%s' is not supported\n", newLang)
		fmt.Printf("Available languages: %v\n", supportedLangs)
		return
	}

	tool.config.Language = newLang
	fmt.Printf("✅ Language changed to: %s\n", newLang)
}

func (tool *CLITool) showInteractiveHelp() {
	fmt.Println("\n🆘 Interactive Mode Help")
	fmt.Println("======================")
	fmt.Println("Commands:")
	fmt.Println("  translate <key> [params]  - Translate a key with optional parameters")
	fmt.Println("  list-keys (ls)           - List all available translation keys")
	fmt.Println("  validate (val)           - Validate translation files")
	fmt.Println("  stats (st)               - Show translation statistics")
	fmt.Println("  change-lang <lang>       - Change current language")
	fmt.Println("  help (h)                 - Show this help message")
	fmt.Println("  exit (quit, q)           - Exit the interactive mode")
	fmt.Println("\nExamples:")
	fmt.Println("  translate cli.welcome")
	fmt.Println("  translate user.greeting John 25")
	fmt.Println("  change-lang pt")
	fmt.Println("  stats")
	fmt.Println()
}

func createSampleTranslations(dir string) error {
	// English translations
	enContent := `{
  "cli": {
    "welcome": "Welcome to the I18n CLI Tool!",
    "goodbye": "Thank you for using the CLI tool!",
    "help": {
      "usage": "Use -h flag for help and available commands"
    }
  },
  "commands": {
    "translate": "Translate a key to the target language",
    "validate": "Validate translation files for errors",
    "stats": "Show translation statistics and information"
  },
  "errors": {
    "not_found": "Translation key '{{key}}' not found",
    "invalid_params": "Invalid parameters provided: {{error}}",
    "file_not_found": "Translation file not found: {{file}}"
  },
  "success": {
    "operation_completed": "Operation completed successfully",
    "translation_loaded": "Translations loaded for {{language}}",
    "validation_passed": "All validation checks passed"
  },
  "user": {
    "profile": {
      "title": "User Profile",
      "greeting": "Hello {{name}}, you are {{age}} years old!"
    },
    "settings": {
      "language": "Language Settings",
      "preferences": "User Preferences"
    }
  }
}`

	// Portuguese translations
	ptContent := `{
  "cli": {
    "welcome": "Bem-vindo à Ferramenta CLI I18n!",
    "goodbye": "Obrigado por usar a ferramenta CLI!",
    "help": {
      "usage": "Use a flag -h para ajuda e comandos disponíveis"
    }
  },
  "commands": {
    "translate": "Traduzir uma chave para o idioma de destino",
    "validate": "Validar arquivos de tradução para erros",
    "stats": "Mostrar estatísticas e informações de tradução"
  },
  "errors": {
    "not_found": "Chave de tradução '{{key}}' não encontrada",
    "invalid_params": "Parâmetros inválidos fornecidos: {{error}}",
    "file_not_found": "Arquivo de tradução não encontrado: {{file}}"
  },
  "success": {
    "operation_completed": "Operação concluída com sucesso",
    "translation_loaded": "Traduções carregadas para {{language}}",
    "validation_passed": "Todas as verificações de validação passaram"
  },
  "user": {
    "profile": {
      "title": "Perfil do Usuário",
      "greeting": "Olá {{name}}, você tem {{age}} anos!"
    },
    "settings": {
      "language": "Configurações de Idioma",
      "preferences": "Preferências do Usuário"
    }
  }
}`

	// Spanish translations
	esContent := `{
  "cli": {
    "welcome": "¡Bienvenido a la Herramienta CLI I18n!",
    "goodbye": "¡Gracias por usar la herramienta CLI!",
    "help": {
      "usage": "Usa la bandera -h para ayuda y comandos disponibles"
    }
  },
  "commands": {
    "translate": "Traducir una clave al idioma objetivo",
    "validate": "Validar archivos de traducción para errores",
    "stats": "Mostrar estadísticas e información de traducción"
  },
  "errors": {
    "not_found": "Clave de traducción '{{key}}' no encontrada",
    "invalid_params": "Parámetros inválidos proporcionados: {{error}}",
    "file_not_found": "Archivo de traducción no encontrado: {{file}}"
  },
  "success": {
    "operation_completed": "Operación completada exitosamente",
    "translation_loaded": "Traducciones cargadas para {{language}}",
    "validation_passed": "Todas las verificaciones de validación pasaron"
  },
  "user": {
    "profile": {
      "title": "Perfil de Usuario",
      "greeting": "¡Hola {{name}}, tienes {{age}} años!"
    },
    "settings": {
      "language": "Configuraciones de Idioma",
      "preferences": "Preferencias del Usuario"
    }
  }
}`

	files := map[string]string{
		"en.json": enContent,
		"pt.json": ptContent,
		"es.json": esContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
