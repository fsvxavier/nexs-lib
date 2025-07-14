package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configurar captura de sinais
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("🔌 Demonstração de Gerenciamento de Pool de Conexões")
	fmt.Println("====================================================")

	// Criar provedor PGX
	provider := pgx.NewProvider()
	defer func() {
		if err := provider.Close(); err != nil {
			log.Printf("Erro ao fechar provedor: %v", err)
		}
	}()

	// Configurar pool com diferentes cenários
	configs := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "Pool Pequeno (2-5 conexões)",
			cfg: config.NewConfig(
				config.WithHost(getEnv("DB_HOST", "localhost")),
				config.WithPort(getEnvInt("DB_PORT", 5432)),
				config.WithDatabase(getEnv("DB_NAME", "example")),
				config.WithUsername(getEnv("DB_USER", "postgres")),
				config.WithPassword(getEnv("DB_PASSWORD", "password")),
				config.WithMaxConns(5),
				config.WithMinConns(2),
				config.WithConnectTimeout(10*time.Second),
				config.WithQueryTimeout(5*time.Second),
				config.WithMaxConnLifetime(30*time.Minute),
				config.WithMaxConnIdleTime(5*time.Minute),
			),
		},
		{
			name: "Pool Médio (10-20 conexões)",
			cfg: config.NewConfig(
				config.WithHost(getEnv("DB_HOST", "localhost")),
				config.WithPort(getEnvInt("DB_PORT", 5432)),
				config.WithDatabase(getEnv("DB_NAME", "example")),
				config.WithUsername(getEnv("DB_USER", "postgres")),
				config.WithPassword(getEnv("DB_PASSWORD", "password")),
				config.WithMaxConns(20),
				config.WithMinConns(10),
				config.WithConnectTimeout(10*time.Second),
				config.WithQueryTimeout(5*time.Second),
				config.WithMaxConnLifetime(1*time.Hour),
				config.WithMaxConnIdleTime(10*time.Minute),
			),
		},
		{
			name: "Pool Grande (50-100 conexões)",
			cfg: config.NewConfig(
				config.WithHost(getEnv("DB_HOST", "localhost")),
				config.WithPort(getEnvInt("DB_PORT", 5432)),
				config.WithDatabase(getEnv("DB_NAME", "example")),
				config.WithUsername(getEnv("DB_USER", "postgres")),
				config.WithPassword(getEnv("DB_PASSWORD", "password")),
				config.WithMaxConns(100),
				config.WithMinConns(50),
				config.WithConnectTimeout(15*time.Second),
				config.WithQueryTimeout(10*time.Second),
				config.WithMaxConnLifetime(2*time.Hour),
				config.WithMaxConnIdleTime(15*time.Minute),
			),
		},
	}

	// Demonstrar cada configuração
	for i, configData := range configs {
		fmt.Printf("\n🎯 Testando: %s\n", configData.name)
		fmt.Println("─────────────────────────────────────────")

		if err := demonstratePoolConfiguration(ctx, provider, configData.cfg, i+1); err != nil {
			log.Printf("❌ Erro na demonstração %d: %v", i+1, err)
			continue
		}

		// Aguardar entre testes
		if i < len(configs)-1 {
			fmt.Printf("\n⏳ Aguardando 5 segundos antes do próximo teste...\n")
			time.Sleep(5 * time.Second)
		}
	}

	// Demonstração final com pool otimizado
	fmt.Printf("\n🚀 Demonstração Final: Pool Otimizado com Monitoramento\n")
	fmt.Println("═══════════════════════════════════════════════════════")

	if err := demonstrateAdvancedPoolManagement(ctx, provider, sigChan); err != nil {
		log.Fatalf("❌ Erro na demonstração avançada: %v", err)
	}

	fmt.Println("\n🎉 Demonstração de gerenciamento de pool concluída!")
}

// demonstratePoolConfiguration demonstra uma configuração específica de pool
func demonstratePoolConfiguration(ctx context.Context, provider *pgx.Provider, cfg *config.Config, testNumber int) error {
	// Criar pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}
	defer pool.Close()

	// Testar conexão inicial
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("erro no ping inicial: %w", err)
	}
	fmt.Printf("✅ Pool criado e conectado com sucesso\n")

	// Exibir configuração inicial
	stats := pool.Stats()
	fmt.Printf("📊 Configuração inicial:\n")
	fmt.Printf("   Max: %d | Min: %d | Atual: %d\n",
		stats.MaxConns, stats.MinConns, stats.TotalConns)

	// Criar monitoramento
	monitor := NewMonitoringManager(pool)

	// Iniciar monitoramento
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()

	monitor.StartMonitoring(monitorCtx)
	defer monitor.StopMonitoring()

	// Teste de carga básico
	workerManager := NewWorkerManager(pool)

	// Executar workers por tempo limitado
	fmt.Printf("\n🏭 Executando %d workers por 10 segundos...\n", int(stats.MaxConns/2))
	workerManager.StartWorkers(int(stats.MaxConns/2), 10*time.Second)

	// Exibir estatísticas finais
	finalStats := workerManager.GetAggregatedStats()
	fmt.Printf("\n📈 Resultados do teste %d:\n", testNumber)
	for key, value := range finalStats {
		fmt.Printf("   %s: %v\n", key, value)
	}

	return nil
}

// demonstrateAdvancedPoolManagement demonstra gerenciamento avançado
func demonstrateAdvancedPoolManagement(ctx context.Context, provider *pgx.Provider, sigChan chan os.Signal) error {
	// Configuração otimizada para demonstração
	cfg := config.NewConfig(
		config.WithHost(getEnv("DB_HOST", "localhost")),
		config.WithPort(getEnvInt("DB_PORT", 5432)),
		config.WithDatabase(getEnv("DB_NAME", "example")),
		config.WithUsername(getEnv("DB_USER", "postgres")),
		config.WithPassword(getEnv("DB_PASSWORD", "password")),
		config.WithMaxConns(30),
		config.WithMinConns(15),
		config.WithConnectTimeout(10*time.Second),
		config.WithQueryTimeout(8*time.Second),
		config.WithMaxConnLifetime(1*time.Hour),
		config.WithMaxConnIdleTime(10*time.Minute),
	)

	// Criar pool
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}
	defer pool.Close()

	// Verificar saúde inicial
	fmt.Printf("🏥 Verificação de saúde inicial...\n")
	healthChecker := NewHealthChecker(pool)
	healthChecker.PerformHealthCheck(ctx)

	// Iniciar monitoramento completo
	monitor := NewMonitoringManager(pool)
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()

	monitor.StartMonitoring(monitorCtx)
	defer monitor.StopMonitoring()

	// Cenários de teste diferentes
	scenarios := []struct {
		name        string
		workers     int
		duration    time.Duration
		description string
	}{
		{
			name:        "Carga Baixa",
			workers:     5,
			duration:    15 * time.Second,
			description: "Simulando uso normal do sistema",
		},
		{
			name:        "Carga Média",
			workers:     15,
			duration:    20 * time.Second,
			description: "Simulando horário de pico moderado",
		},
		{
			name:        "Carga Alta",
			workers:     25,
			duration:    15 * time.Second,
			description: "Simulando horário de pico intenso",
		},
	}

	// Executar cenários
	for i, scenario := range scenarios {
		fmt.Printf("\n🎬 Cenário %d: %s\n", i+1, scenario.name)
		fmt.Printf("📝 %s\n", scenario.description)
		fmt.Printf("⚙️ Workers: %d | Duração: %v\n", scenario.workers, scenario.duration)

		// Verificar se deve continuar
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Recebido sinal de interrupção, finalizando...")
			return nil
		default:
		}

		// Executar cenário
		workerManager := NewWorkerManager(pool)
		workerManager.StartWorkers(scenario.workers, scenario.duration)

		// Exibir resultados
		results := workerManager.GetAggregatedStats()
		fmt.Printf("📊 Resultados do cenário:\n")
		for key, value := range results {
			fmt.Printf("   %s: %v\n", key, value)
		}

		// Teste de carga para este cenário
		fmt.Printf("\n🔥 Teste de carga para cenário: %s\n", scenario.name)
		healthChecker.LoadTest(ctx, scenario.workers*2, 8*time.Second)

		// Verificação de saúde pós-teste
		fmt.Printf("\n🏥 Verificação de saúde pós-teste...\n")
		healthChecker.PerformHealthCheck(ctx)

		// Aguardar entre cenários
		if i < len(scenarios)-1 {
			fmt.Printf("\n⏳ Pausa de 10 segundos entre cenários...\n")

			select {
			case <-sigChan:
				fmt.Println("🛑 Recebido sinal de interrupção, finalizando...")
				return nil
			case <-time.After(10 * time.Second):
				// Continuar
			}
		}
	}

	fmt.Printf("\n🎯 Demonstração avançada concluída com sucesso!\n")
	fmt.Printf("💡 Dica: Use Ctrl+C para interromper a qualquer momento\n")

	return nil
}

// Funções utilitárias
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d"); err == nil && intValue == 1 {
			var result int
			fmt.Sscanf(value, "%d", &result)
			return result
		}
	}
	return defaultValue
}
