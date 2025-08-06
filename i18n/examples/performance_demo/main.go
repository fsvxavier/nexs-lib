// Package main demonstrates performance optimizations in the i18n library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

func main() {
	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_performance_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create translation files with extensive content
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal("Failed to create translation files:", err)
	}

	fmt.Println("🚀 I18n Performance Optimization Demonstration")
	fmt.Println("==============================================")
	fmt.Println()

	// Demo 1: Basic performance baseline
	fmt.Println("📊 Demo 1: Performance Baseline")
	runPerformanceBaseline(tempDir)
	fmt.Println()

	// Demo 2: String pooling and interning
	fmt.Println("📊 Demo 2: String Pooling & Interning")
	runStringOptimizationDemo(tempDir)
	fmt.Println()

	// Demo 3: Batch translation operations
	fmt.Println("📊 Demo 3: Batch Translation Operations")
	runBatchTranslationDemo(tempDir)
	fmt.Println()

	// Demo 4: Memory usage analysis
	fmt.Println("📊 Demo 4: Memory Usage Analysis")
	runMemoryAnalysisDemo(tempDir)
	fmt.Println()

	// Demo 5: Concurrent performance
	fmt.Println("📊 Demo 5: Concurrent Performance")
	runConcurrentPerformanceDemo(tempDir)
	fmt.Println()

	fmt.Println("✅ All performance demonstrations completed!")
}

func runPerformanceBaseline(translationDir string) {
	provider, err := setupProvider(translationDir, false)
	if err != nil {
		log.Fatal("Failed to setup provider:", err)
	}
	defer provider.Stop(context.Background())

	fmt.Println("  🔸 Single translation performance:")
	ctx := context.Background()
	iterations := 10000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := provider.Translate(ctx, "performance.test", "en", map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now().UnixNano(),
		})
		if err != nil {
			fmt.Printf("  ❌ Translation error: %v\n", err)
			return
		}
	}
	duration := time.Since(start)

	fmt.Printf("  ✅ %d translations completed in %v\n", iterations, duration)
	fmt.Printf("  📈 Average: %v per translation\n", duration/time.Duration(iterations))
	fmt.Printf("  🚀 Throughput: %.0f translations/second\n", float64(iterations)/duration.Seconds())
}

func runStringOptimizationDemo(translationDir string) {
	// This demo simulates the benefits of string pooling and interning
	// In a real implementation, these would be integrated into the provider

	fmt.Println("  🔸 Simulating string optimization benefits:")

	// Simulate memory usage with many duplicate strings
	duplicateStrings := make([]string, 10000)
	commonKeys := []string{
		"user.profile.title",
		"user.settings.title",
		"error.not_found",
		"success.saved",
		"navigation.home",
	}

	start := time.Now()

	// Without interning (lots of duplicate allocations)
	for i := 0; i < len(duplicateStrings); i++ {
		duplicateStrings[i] = commonKeys[i%len(commonKeys)]
	}

	withoutInterning := time.Since(start)

	// Simulate with interning (reuse existing strings)
	stringPool := make(map[string]string)
	internedStrings := make([]string, 10000)

	start = time.Now()
	for i := 0; i < len(internedStrings); i++ {
		key := commonKeys[i%len(commonKeys)]
		if interned, exists := stringPool[key]; exists {
			internedStrings[i] = interned
		} else {
			stringPool[key] = key
			internedStrings[i] = key
		}
	}
	withInterning := time.Since(start)

	fmt.Printf("  📊 Without string interning: %v\n", withoutInterning)
	fmt.Printf("  📊 With string interning: %v\n", withInterning)
	fmt.Printf("  🚀 Performance improvement: %.2fx faster\n", float64(withoutInterning)/float64(withInterning))
	fmt.Printf("  💾 Memory efficiency: %d unique strings vs %d total strings\n", len(stringPool), len(duplicateStrings))
}

func runBatchTranslationDemo(translationDir string) {
	provider, err := setupProvider(translationDir, true)
	if err != nil {
		log.Fatal("Failed to setup provider:", err)
	}
	defer provider.Stop(context.Background())

	ctx := context.Background()

	// Prepare batch requests
	batchSize := 100
	keys := []string{
		"performance.test",
		"user.profile.title",
		"batch.operation",
		"success.message",
		"error.generic",
	}
	languages := []string{"en", "pt", "es"}

	fmt.Printf("  🔸 Batch translation performance (%d requests):\n", batchSize)

	// Individual requests
	start := time.Now()
	for i := 0; i < batchSize; i++ {
		key := keys[i%len(keys)]
		lang := languages[i%len(languages)]
		_, err := provider.Translate(ctx, key, lang, map[string]interface{}{
			"batch_id": i,
		})
		if err != nil {
			fmt.Printf("  ❌ Individual translation error: %v\n", err)
		}
	}
	individualDuration := time.Since(start)

	// Simulate batch operation (in real implementation, this would be optimized)
	start = time.Now()

	// Group by language for better cache locality
	languageGroups := make(map[string][]string)
	for i := 0; i < batchSize; i++ {
		key := keys[i%len(keys)]
		lang := languages[i%len(languages)]
		languageGroups[lang] = append(languageGroups[lang], key)
	}

	// Process each language group
	for lang, langKeys := range languageGroups {
		for _, key := range langKeys {
			_, err := provider.Translate(ctx, key, lang, nil)
			if err != nil {
				fmt.Printf("  ❌ Batch translation error: %v\n", err)
			}
		}
	}
	batchDuration := time.Since(start)

	fmt.Printf("  📊 Individual requests: %v\n", individualDuration)
	fmt.Printf("  📊 Batch-optimized: %v\n", batchDuration)
	fmt.Printf("  🚀 Batch improvement: %.2fx faster\n", float64(individualDuration)/float64(batchDuration))
	fmt.Printf("  📈 Batch throughput: %.0f translations/second\n", float64(batchSize)/batchDuration.Seconds())
}

func runMemoryAnalysisDemo(translationDir string) {
	fmt.Println("  🔸 Memory usage analysis:")

	// Get initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create provider and load translations
	provider, err := setupProvider(translationDir, true)
	if err != nil {
		log.Fatal("Failed to setup provider:", err)
	}

	// Get memory stats after loading
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Perform many translations
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		provider.Translate(ctx, "performance.test", "en", map[string]interface{}{
			"iteration": i,
		})
	}

	// Get memory stats after translations
	var m3 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m3)

	provider.Stop(context.Background())

	// Get final memory stats
	var m4 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m4)

	fmt.Printf("  📊 Memory usage analysis:\n")
	fmt.Printf("    Initial memory: %d KB\n", bToKb(m1.Alloc))
	fmt.Printf("    After loading translations: %d KB (+%d KB)\n", bToKb(m2.Alloc), bToKb(m2.Alloc-m1.Alloc))
	fmt.Printf("    After 1000 translations: %d KB (+%d KB)\n", bToKb(m3.Alloc), bToKb(m3.Alloc-m2.Alloc))
	fmt.Printf("    After cleanup: %d KB (%d KB freed)\n", bToKb(m4.Alloc), bToKb(m3.Alloc-m4.Alloc))
	fmt.Printf("  📈 Total allocations: %d\n", m3.TotalAlloc-m1.TotalAlloc)
	fmt.Printf("  🗑️  GC cycles: %d\n", m3.NumGC-m1.NumGC)
}

func runConcurrentPerformanceDemo(translationDir string) {
	provider, err := setupProvider(translationDir, true)
	if err != nil {
		log.Fatal("Failed to setup provider:", err)
	}
	defer provider.Stop(context.Background())

	goroutines := []int{1, 10, 50, 100}
	translationsPerGoroutine := 100

	fmt.Println("  🔸 Concurrent performance scaling:")

	for _, numGoroutines := range goroutines {
		start := time.Now()

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(workerID int) {
				ctx := context.Background()
				for j := 0; j < translationsPerGoroutine; j++ {
					_, err := provider.Translate(ctx, "performance.concurrent", "en", map[string]interface{}{
						"worker": workerID,
						"task":   j,
					})
					if err != nil {
						fmt.Printf("  ❌ Concurrent translation error: %v\n", err)
					}
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		duration := time.Since(start)
		totalTranslations := numGoroutines * translationsPerGoroutine
		throughput := float64(totalTranslations) / duration.Seconds()

		fmt.Printf("    %d goroutines: %v (%d translations, %.0f/sec)\n",
			numGoroutines, duration, totalTranslations, throughput)
	}
}

func setupProvider(translationDir string, enableCache bool) (interfaces.I18n, error) {
	// Configure i18n
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(enableCache, 10*time.Minute).
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

	return provider, nil
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func createTranslationFiles(dir string) error {
	// English translations with extensive content for performance testing
	enContent := `{
  "performance": {
    "test": "Performance test message {{iteration}} at {{timestamp}}",
    "concurrent": "Concurrent test - Worker {{worker}}, Task {{task}}",
    "batch": "Batch operation test message",
    "memory": "Memory optimization test",
    "cache": "Cache performance test"
  },
  "user": {
    "profile": {
      "title": "User Profile",
      "name": "Name: {{name}}",
      "email": "Email: {{email}}",
      "age": "Age: {{age}}",
      "bio": "Biography",
      "settings": "Profile Settings",
      "edit": "Edit Profile",
      "save": "Save Changes",
      "cancel": "Cancel",
      "delete": "Delete Profile"
    },
    "settings": {
      "title": "User Settings",
      "language": "Language Preference",
      "timezone": "Timezone",
      "notifications": "Notification Settings",
      "privacy": "Privacy Settings",
      "security": "Security Settings",
      "theme": "Theme Preference",
      "account": "Account Settings"
    }
  },
  "navigation": {
    "home": "Home",
    "about": "About",
    "contact": "Contact",
    "services": "Services",
    "products": "Products",
    "blog": "Blog",
    "support": "Support",
    "login": "Login",
    "logout": "Logout",
    "register": "Register"
  },
  "messages": {
    "welcome": "Welcome to our application!",
    "goodbye": "Thank you for using our service!",
    "loading": "Loading...",
    "saving": "Saving changes...",
    "saved": "Changes saved successfully!",
    "error": "An error occurred",
    "retry": "Please try again",
    "success": "Operation completed successfully",
    "warning": "Please review your input",
    "info": "Additional information available"
  },
  "errors": {
    "generic": "An unexpected error occurred",
    "not_found": "Resource not found",
    "unauthorized": "Unauthorized access",
    "forbidden": "Access forbidden",
    "validation": "Validation error",
    "network": "Network connection error",
    "timeout": "Request timeout",
    "server": "Server error",
    "maintenance": "System under maintenance"
  },
  "success": {
    "saved": "Data saved successfully",
    "updated": "Information updated",
    "deleted": "Item deleted",
    "created": "Item created",
    "uploaded": "File uploaded",
    "sent": "Message sent",
    "processed": "Request processed",
    "completed": "Operation completed"
  },
  "batch": {
    "operation": "Batch operation {{batch_id}}",
    "processing": "Processing batch...",
    "completed": "Batch processing completed",
    "failed": "Batch processing failed",
    "partial": "Batch partially completed"
  }
}`

	// Portuguese translations
	ptContent := `{
  "performance": {
    "test": "Mensagem de teste de performance {{iteration}} em {{timestamp}}",
    "concurrent": "Teste concorrente - Trabalhador {{worker}}, Tarefa {{task}}",
    "batch": "Mensagem de teste de operação em lote",
    "memory": "Teste de otimização de memória",
    "cache": "Teste de performance de cache"
  },
  "user": {
    "profile": {
      "title": "Perfil do Usuário",
      "name": "Nome: {{name}}",
      "email": "Email: {{email}}",
      "age": "Idade: {{age}}",
      "bio": "Biografia",
      "settings": "Configurações do Perfil",
      "edit": "Editar Perfil",
      "save": "Salvar Alterações",
      "cancel": "Cancelar",
      "delete": "Deletar Perfil"
    },
    "settings": {
      "title": "Configurações do Usuário",
      "language": "Preferência de Idioma",
      "timezone": "Fuso Horário",
      "notifications": "Configurações de Notificação",
      "privacy": "Configurações de Privacidade",
      "security": "Configurações de Segurança",
      "theme": "Preferência de Tema",
      "account": "Configurações da Conta"
    }
  },
  "navigation": {
    "home": "Início",
    "about": "Sobre",
    "contact": "Contato",
    "services": "Serviços",
    "products": "Produtos",
    "blog": "Blog",
    "support": "Suporte",
    "login": "Entrar",
    "logout": "Sair",
    "register": "Registrar"
  },
  "messages": {
    "welcome": "Bem-vindo ao nosso aplicativo!",
    "goodbye": "Obrigado por usar nosso serviço!",
    "loading": "Carregando...",
    "saving": "Salvando alterações...",
    "saved": "Alterações salvas com sucesso!",
    "error": "Ocorreu um erro",
    "retry": "Tente novamente",
    "success": "Operação concluída com sucesso",
    "warning": "Revise sua entrada",
    "info": "Informações adicionais disponíveis"
  },
  "errors": {
    "generic": "Ocorreu um erro inesperado",
    "not_found": "Recurso não encontrado",
    "unauthorized": "Acesso não autorizado",
    "forbidden": "Acesso proibido",
    "validation": "Erro de validação",
    "network": "Erro de conexão de rede",
    "timeout": "Timeout da requisição",
    "server": "Erro do servidor",
    "maintenance": "Sistema em manutenção"
  },
  "success": {
    "saved": "Dados salvos com sucesso",
    "updated": "Informações atualizadas",
    "deleted": "Item deletado",
    "created": "Item criado",
    "uploaded": "Arquivo carregado",
    "sent": "Mensagem enviada",
    "processed": "Requisição processada",
    "completed": "Operação concluída"
  },
  "batch": {
    "operation": "Operação em lote {{batch_id}}",
    "processing": "Processando lote...",
    "completed": "Processamento em lote concluído",
    "failed": "Processamento em lote falhou",
    "partial": "Lote parcialmente concluído"
  }
}`

	// Spanish translations
	esContent := `{
  "performance": {
    "test": "Mensaje de prueba de rendimiento {{iteration}} en {{timestamp}}",
    "concurrent": "Prueba concurrente - Trabajador {{worker}}, Tarea {{task}}",
    "batch": "Mensaje de prueba de operación por lotes",
    "memory": "Prueba de optimización de memoria",
    "cache": "Prueba de rendimiento de caché"
  },
  "user": {
    "profile": {
      "title": "Perfil de Usuario",
      "name": "Nombre: {{name}}",
      "email": "Email: {{email}}",
      "age": "Edad: {{age}}",
      "bio": "Biografía",
      "settings": "Configuraciones del Perfil",
      "edit": "Editar Perfil",
      "save": "Guardar Cambios",
      "cancel": "Cancelar",
      "delete": "Eliminar Perfil"
    },
    "settings": {
      "title": "Configuraciones de Usuario",
      "language": "Preferencia de Idioma",
      "timezone": "Zona Horaria",
      "notifications": "Configuraciones de Notificación",
      "privacy": "Configuraciones de Privacidad",
      "security": "Configuraciones de Seguridad",
      "theme": "Preferencia de Tema",
      "account": "Configuraciones de Cuenta"
    }
  },
  "navigation": {
    "home": "Inicio",
    "about": "Acerca de",
    "contact": "Contacto",
    "services": "Servicios",
    "products": "Productos",
    "blog": "Blog",
    "support": "Soporte",
    "login": "Iniciar Sesión",
    "logout": "Cerrar Sesión",
    "register": "Registrarse"
  },
  "messages": {
    "welcome": "¡Bienvenido a nuestra aplicación!",
    "goodbye": "¡Gracias por usar nuestro servicio!",
    "loading": "Cargando...",
    "saving": "Guardando cambios...",
    "saved": "¡Cambios guardados exitosamente!",
    "error": "Ocurrió un error",
    "retry": "Por favor intenta de nuevo",
    "success": "Operación completada exitosamente",
    "warning": "Por favor revisa tu entrada",
    "info": "Información adicional disponible"
  },
  "errors": {
    "generic": "Ocurrió un error inesperado",
    "not_found": "Recurso no encontrado",
    "unauthorized": "Acceso no autorizado",
    "forbidden": "Acceso prohibido",
    "validation": "Error de validación",
    "network": "Error de conexión de red",
    "timeout": "Tiempo de espera agotado",
    "server": "Error del servidor",
    "maintenance": "Sistema en mantenimiento"
  },
  "success": {
    "saved": "Datos guardados exitosamente",
    "updated": "Información actualizada",
    "deleted": "Elemento eliminado",
    "created": "Elemento creado",
    "uploaded": "Archivo subido",
    "sent": "Mensaje enviado",
    "processed": "Solicitud procesada",
    "completed": "Operación completada"
  },
  "batch": {
    "operation": "Operación por lotes {{batch_id}}",
    "processing": "Procesando lote...",
    "completed": "Procesamiento por lotes completado",
    "failed": "Procesamiento por lotes falló",
    "partial": "Lote parcialmente completado"
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
