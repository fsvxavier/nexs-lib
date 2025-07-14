package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

// PoolMonitor monitora o pool de conexões
type PoolMonitor struct {
	pool postgresql.IPool
	stop chan bool
}

// NewPoolMonitor cria um novo monitor de pool
func NewPoolMonitor(pool postgresql.IPool) *PoolMonitor {
	return &PoolMonitor{
		pool: pool,
		stop: make(chan bool),
	}
}

// Start inicia o monitoramento
func (pm *PoolMonitor) Start(interval time.Duration) {
	fmt.Printf("📊 Iniciando monitoramento do pool (intervalo: %v)\n", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stop:
			fmt.Println("📊 Monitoramento parado")
			return
		case <-ticker.C:
			pm.printStats()
		}
	}
}

// Stop para o monitoramento
func (pm *PoolMonitor) Stop() {
	close(pm.stop)
}

// printStats imprime estatísticas do pool
func (pm *PoolMonitor) printStats() {
	stats := pm.pool.Stats()

	fmt.Printf("\n📊 Estatísticas do Pool (%s):\n", time.Now().Format("15:04:05"))
	fmt.Printf("┌─────────────────────────────────────────┐\n")
	fmt.Printf("│ Conexões Totais: %-22d │\n", stats.TotalConns)
	fmt.Printf("│ Conexões Ativas: %-22d │\n", stats.AcquiredConns)
	fmt.Printf("│ Conexões Ociosas: %-21d │\n", stats.IdleConns)
	fmt.Printf("│ Conexões Construindo: %-17d │\n", stats.ConstructingConns)
	fmt.Printf("│ Max Conexões: %-25d │\n", stats.MaxConns)
	fmt.Printf("│ Min Conexões: %-25d │\n", stats.MinConns)
	fmt.Printf("├─────────────────────────────────────────┤\n")
	fmt.Printf("│ Total de Aquisições: %-19d │\n", stats.AcquireCount)
	fmt.Printf("│ Aquisições Canceladas: %-17d │\n", stats.CanceledAcquireCount)
	fmt.Printf("│ Aquisições Vazias: %-21d │\n", stats.EmptyAcquireCount)
	fmt.Printf("│ Novas Conexões: %-24d │\n", stats.NewConnsCount)
	fmt.Printf("├─────────────────────────────────────────┤\n")
	fmt.Printf("│ Tempo Médio Aquisição: %-17v │\n", stats.AcquireDuration)
	fmt.Printf("│ Destruídas (Lifetime): %-17d │\n", stats.MaxLifetimeDestroyCount)
	fmt.Printf("│ Destruídas (Idle): %-23d │\n", stats.MaxIdleDestroyCount)
	fmt.Printf("└─────────────────────────────────────────┘\n")
}

// HealthChecker verifica saúde do pool
type HealthChecker struct {
	pool postgresql.IPool
}

// NewHealthChecker cria um novo verificador de saúde
func NewHealthChecker(pool postgresql.IPool) *HealthChecker {
	return &HealthChecker{pool: pool}
}

// CheckHealth verifica a saúde do pool
func (hc *HealthChecker) CheckHealth(ctx context.Context) error {
	// Verificar ping básico
	if err := hc.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping falhou: %w", err)
	}

	// Verificar aquisição de conexão
	conn, err := hc.pool.AcquireWithTimeout(ctx, 5*time.Second)
	if err != nil {
		return fmt.Errorf("falha ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	// Verificar query simples
	var result int
	row, _ := conn.QueryRow(ctx, "SELECT 1")
	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("falha na query de teste: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("resultado inesperado da query de teste: %d", result)
	}

	return nil
}

// PerformHealthCheck executa verificação de saúde com relatório
func (hc *HealthChecker) PerformHealthCheck(ctx context.Context) {
	fmt.Printf("🏥 Executando verificação de saúde...\n")

	start := time.Now()
	err := hc.CheckHealth(ctx)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Verificação de saúde falhou (%.2fms): %v\n",
			float64(duration.Nanoseconds())/1e6, err)
		return
	}

	fmt.Printf("✅ Pool saudável (%.2fms)\n", float64(duration.Nanoseconds())/1e6)
}

// LoadTest executa teste de carga no pool
func (hc *HealthChecker) LoadTest(ctx context.Context, connections int, duration time.Duration) {
	fmt.Printf("🔥 Iniciando teste de carga: %d conexões por %v\n", connections, duration)

	start := time.Now()

	// Canal para controlar a duração do teste
	done := make(chan bool, 1)
	go func() {
		time.Sleep(duration)
		done <- true
	}()

	// Estatísticas do teste
	var (
		totalQueries  int
		totalErrors   int
		totalDuration time.Duration
	)

	// Executar queries em paralelo
	queryChan := make(chan bool, connections)

	for {
		select {
		case <-done:
			close(queryChan)

			// Aguardar todas as queries terminarem
			for len(queryChan) > 0 {
				time.Sleep(10 * time.Millisecond)
			}

			elapsed := time.Since(start)
			avgResponseTime := time.Duration(0)
			if totalQueries > 0 {
				avgResponseTime = totalDuration / time.Duration(totalQueries)
			}

			fmt.Printf("\n🏁 Teste de carga concluído:\n")
			fmt.Printf("   Tempo total: %v\n", elapsed)
			fmt.Printf("   Queries executadas: %d\n", totalQueries)
			fmt.Printf("   Erros: %d (%.2f%%)\n", totalErrors, float64(totalErrors)/float64(totalQueries)*100)
			fmt.Printf("   Tempo médio de resposta: %v\n", avgResponseTime)
			fmt.Printf("   Queries por segundo: %.2f\n", float64(totalQueries)/elapsed.Seconds())

			return

		default:
			// Enviar nova query se houver espaço
			select {
			case queryChan <- true:
				go func() {
					defer func() { <-queryChan }()

					queryStart := time.Now()

					// Adquirir conexão e executar query
					conn, err := hc.pool.AcquireWithTimeout(ctx, 1*time.Second)
					if err != nil {
						totalErrors++
						return
					}
					defer conn.Release(ctx)

					var result int
					row, _ := conn.QueryRow(ctx, "SELECT $1::int", time.Now().Unix()%1000)
					if err := row.Scan(&result); err != nil {
						totalErrors++
						return
					}

					queryDuration := time.Since(queryStart)
					totalQueries++
					totalDuration += queryDuration
				}()
			default:
				// Pool cheio, aguardar um pouco
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

// MonitoringManager gerencia todos os aspectos de monitoramento
type MonitoringManager struct {
	poolMonitor   *PoolMonitor
	healthChecker *HealthChecker
}

// NewMonitoringManager cria um novo gerenciador de monitoramento
func NewMonitoringManager(pool postgresql.IPool) *MonitoringManager {
	return &MonitoringManager{
		poolMonitor:   NewPoolMonitor(pool),
		healthChecker: NewHealthChecker(pool),
	}
}

// StartMonitoring inicia o monitoramento completo
func (mm *MonitoringManager) StartMonitoring(ctx context.Context) {
	fmt.Println("🚀 Iniciando monitoramento completo...")

	// Iniciar monitoramento de estatísticas
	go mm.poolMonitor.Start(3 * time.Second)

	// Verificações de saúde periódicas
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mm.healthChecker.PerformHealthCheck(ctx)
			}
		}
	}()
}

// StopMonitoring para o monitoramento
func (mm *MonitoringManager) StopMonitoring() {
	fmt.Println("🛑 Parando monitoramento...")
	mm.poolMonitor.Stop()
}
