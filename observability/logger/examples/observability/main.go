package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
	"github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
)

func main() {
	fmt.Println("=== Demonstração: Observabilidade Avançada - Fase 7 ===")

	// 1. Cria provider base (Slog)
	provider := slog.NewProvider()
	err := provider.Configure(&interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "observability-demo",
	})
	if err != nil {
		log.Fatal("Erro configurando provider:", err)
	}

	// 2. Cria logger observável
	observableLogger := logger.ConfigureObservableLogger(provider, nil)

	fmt.Println("\n1. MÉTRICAS DE LOGGING")
	fmt.Println("======================")

	// Demonstra coleta de métricas
	ctx := context.Background()

	// Gera alguns logs para coletar métricas
	observableLogger.Info(ctx, "Aplicação iniciada",
		logger.String("version", "2.0.0"),
		logger.String("environment", "development"))

	observableLogger.Warn(ctx, "Configuração deprecated encontrada",
		logger.String("config", "old_format"))

	observableLogger.Error(ctx, "Falha na conexão com banco de dados",
		logger.String("database", "postgres"),
		logger.Int("retry_attempt", 1))

	// Exibe métricas coletadas
	metrics := observableLogger.GetMetrics()
	fmt.Printf("Total de logs: %d\n", metrics.GetTotalLogCount())
	fmt.Printf("Logs INFO: %d\n", metrics.GetLogCount(interfaces.InfoLevel))
	fmt.Printf("Logs WARN: %d\n", metrics.GetLogCount(interfaces.WarnLevel))
	fmt.Printf("Logs ERROR: %d\n", metrics.GetLogCount(interfaces.ErrorLevel))
	fmt.Printf("Taxa de erro: %.2f%%\n", metrics.GetErrorRate()*100)
	fmt.Printf("Tempo médio de processamento: %v\n", metrics.GetAverageProcessingTime())

	fmt.Println("\n2. HOOKS CUSTOMIZADOS")
	fmt.Println("=====================")

	// Registra hook de validação
	validationHook := logger.NewValidationHook(
		func(entry *interfaces.LogEntry) error {
			if entry.Level >= interfaces.ErrorLevel && entry.Message == "" {
				return fmt.Errorf("mensagens de erro não podem estar vazias")
			}
			return nil
		},
	)

	observableLogger.RegisterHook(interfaces.BeforeHook, validationHook)
	fmt.Println("✓ Hook de validação registrado")

	// Registra hook de transformação
	transformHook := logger.NewTransformHook(
		func(entry *interfaces.LogEntry) error {
			if entry.Fields == nil {
				entry.Fields = make(map[string]any)
			}
			entry.Fields["processed_at"] = time.Now().Format(time.RFC3339)
			entry.Fields["service"] = "observability-demo"
			return nil
		},
	)

	observableLogger.RegisterHook(interfaces.AfterHook, transformHook)
	fmt.Println("✓ Hook de transformação registrado")

	// Registra hook de filtro
	filterHook := logger.NewFilterHook(
		func(entry *interfaces.LogEntry) bool {
			// Filtra logs de debug em produção
			if entry.Level == interfaces.DebugLevel {
				return false // Filtra debug logs
			}
			return true
		},
	)

	observableLogger.RegisterHook(interfaces.BeforeHook, filterHook)
	fmt.Println("✓ Hook de filtro registrado")

	// Testa hooks em ação
	fmt.Println("\nTestando hooks:")

	observableLogger.Info(ctx, "Log processado com hooks",
		logger.String("data", "exemplo"))

	// Este log será filtrado (debug level)
	observableLogger.Debug(ctx, "Este log será filtrado")

	observableLogger.Error(ctx, "Erro processado com validação e transformação",
		logger.String("error_code", "DB_CONNECTION_FAILED"))

	fmt.Println("\n3. ESTATÍSTICAS DOS HOOKS")
	fmt.Println("=========================")

	hookManager := observableLogger.GetHookManager()
	beforeHooks := hookManager.ListHooks(interfaces.BeforeHook)
	afterHooks := hookManager.ListHooks(interfaces.AfterHook)

	fmt.Printf("Hooks 'before': %d\n", len(beforeHooks))
	for _, hook := range beforeHooks {
		fmt.Printf("  - %s (ativo: %t)\n", hook.GetName(), hook.IsEnabled())
	}

	fmt.Printf("Hooks 'after': %d\n", len(afterHooks))
	for _, hook := range afterHooks {
		fmt.Printf("  - %s (ativo: %t)\n", hook.GetName(), hook.IsEnabled())
	}

	fmt.Println("\n4. MÉTRICAS FINAIS")
	fmt.Println("==================")

	// Atualiza métricas após processamento com hooks
	finalMetrics := observableLogger.GetMetrics()
	fmt.Printf("Total final de logs: %d\n", finalMetrics.GetTotalLogCount())

	// Exporta métricas detalhadas
	exportedMetrics := finalMetrics.Export()
	fmt.Println("\nMétricas exportadas:")
	fmt.Printf("  Start time: %v\n", exportedMetrics["start_time"])
	fmt.Printf("  Performance: %v\n", exportedMetrics["performance"])

	fmt.Println("\n5. DEMONSTRAÇÃO DE SAMPLING")
	fmt.Println("============================")

	// Configura sampling para demonstração
	samplingConfig := &interfaces.SamplingConfig{
		Initial:    5, // Primeiros 5 logs sempre passam
		Thereafter: 2, // Depois, 1 a cada 2 logs
		Tick:       time.Second,
	}

	configWithSampling := &interfaces.Config{
		Level:          interfaces.DebugLevel,
		Format:         interfaces.JSONFormat,
		ServiceName:    "sampling-demo",
		SamplingConfig: samplingConfig,
	}

	samplingProvider := slog.NewProvider()
	err = samplingProvider.Configure(configWithSampling)
	if err != nil {
		log.Fatal("Erro configurando provider com sampling:", err)
	}

	samplingLogger := logger.ConfigureObservableLogger(samplingProvider, configWithSampling)

	fmt.Println("Gerando 10 logs com sampling (inicial: 5, depois: 1 a cada 2):")
	for i := 1; i <= 10; i++ {
		samplingLogger.Info(ctx, fmt.Sprintf("Log de exemplo #%d", i),
			logger.Int("sequence", i))
	}

	samplingMetrics := samplingLogger.GetMetrics()
	fmt.Printf("Taxa de sampling: %.2f%%\n", samplingMetrics.GetSamplingRate()*100)

	fmt.Println("\n6. GESTÃO DE HOOKS EM RUNTIME")
	fmt.Println("==============================")

	// Demonstra habilitação/desabilitação de hooks
	fmt.Println("Desabilitando todos os hooks...")
	hookManager.DisableAllHooks()

	observableLogger.Info(ctx, "Log sem hooks (hooks desabilitados)")

	fmt.Println("Reabilitando hooks...")
	hookManager.EnableAllHooks()

	observableLogger.Info(ctx, "Log com hooks (hooks reabilitados)")

	// Remove hook específico
	fmt.Println("Removendo hook de filtro...")
	err = hookManager.UnregisterHook(interfaces.BeforeHook, "log_filter")
	if err != nil {
		fmt.Printf("Erro removendo hook: %v\n", err)
	} else {
		fmt.Println("✓ Hook de filtro removido")
	}

	// Agora debug logs não serão mais filtrados
	observableLogger.Debug(ctx, "Este debug log agora passa (filtro removido)")

	fmt.Println("\n7. MÉTRICAS DETALHADAS POR NÍVEL")
	fmt.Println("=================================")

	for level := interfaces.DebugLevel; level <= interfaces.ErrorLevel; level++ {
		count := finalMetrics.GetLogCount(level)
		avgTime := finalMetrics.GetProcessingTimeByLevel(level)
		fmt.Printf("%s: %d logs, tempo médio: %v\n",
			level.String(), count, avgTime)
	}

	fmt.Println("\n✅ Demonstração da Fase 7 - Observabilidade Avançada concluída!")
	fmt.Println("\nFuncionalidades implementadas:")
	fmt.Println("  ✓ Métricas de logging (contadores, tempo, taxa de erro)")
	fmt.Println("  ✓ Hooks customizados (validação, transformação, filtro)")
	fmt.Println("  ✓ Gestão de hooks em runtime")
	fmt.Println("  ✓ Coleta automática de métricas")
	fmt.Println("  ✓ Exportação de métricas para sistemas externos")
	fmt.Println("  ✓ Sampling configurável com métricas")
	fmt.Println("  ✓ Thread-safety completa")
}
