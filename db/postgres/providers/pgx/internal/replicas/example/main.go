package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/replicas"
)

func main() {
	fmt.Println("=== Sistema de Read Replicas PostgreSQL ===")

	// Criar contexto
	ctx := context.Background()

	// Configurar replica manager
	config := replicas.ReplicaManagerConfig{
		LoadBalancingStrategy: interfaces.LoadBalancingRoundRobin,
		ReadPreference:        interfaces.ReadPreferenceSecondaryPreferred,
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    5 * time.Second,
	}

	// Criar replica manager
	replicaManager := replicas.NewReplicaManager(config)

	// Adicionar réplicas
	fmt.Println("\n1. Adicionando réplicas...")

	err := replicaManager.AddReplica(ctx, "replica-1", "postgres://replica1:5432/mydb", 10)
	if err != nil {
		log.Printf("Erro ao adicionar replica-1: %v", err)
	} else {
		fmt.Println("✓ Réplica replica-1 adicionada com sucesso")
	}

	err = replicaManager.AddReplica(ctx, "replica-2", "postgres://replica2:5432/mydb", 20)
	if err != nil {
		log.Printf("Erro ao adicionar replica-2: %v", err)
	} else {
		fmt.Println("✓ Réplica replica-2 adicionada com sucesso")
	}

	err = replicaManager.AddReplica(ctx, "replica-3", "postgres://replica3:5432/mydb", 30)
	if err != nil {
		log.Printf("Erro ao adicionar replica-3: %v", err)
	} else {
		fmt.Println("✓ Réplica replica-3 adicionada com sucesso")
	}

	// Listar réplicas
	fmt.Println("\n2. Listando réplicas...")
	replicas_list := replicaManager.ListReplicas()
	for _, replica := range replicas_list {
		fmt.Printf("   - ID: %s, DSN: %s, Peso: %d, Status: %s\n",
			replica.GetID(), replica.GetDSN(), replica.GetWeight(), replica.GetStatus())
	}

	// Iniciar replica manager
	fmt.Println("\n3. Iniciando replica manager...")
	err = replicaManager.Start(ctx)
	if err != nil {
		log.Printf("Erro ao iniciar replica manager: %v", err)
	} else {
		fmt.Println("✓ Replica manager iniciado com sucesso")
	}

	// Testar diferentes estratégias de balanceamento
	fmt.Println("\n4. Testando estratégias de balanceamento...")

	strategies := []interfaces.LoadBalancingStrategy{
		interfaces.LoadBalancingRoundRobin,
		interfaces.LoadBalancingRandom,
		interfaces.LoadBalancingWeighted,
		interfaces.LoadBalancingLatency,
	}

	for _, strategy := range strategies {
		fmt.Printf("\n   Estratégia: %s\n", strategy)

		// Testar seleção de réplicas
		for i := 0; i < 3; i++ {
			replica, err := replicaManager.SelectReplicaWithStrategy(ctx, strategy)
			if err != nil {
				fmt.Printf("   Erro ao selecionar réplica: %v\n", err)
				continue
			}

			fmt.Printf("   Tentativa %d: Réplica selecionada: %s (peso: %d)\n",
				i+1, replica.GetID(), replica.GetWeight())
		}
	}

	// Testar preferências de leitura
	fmt.Println("\n5. Testando preferências de leitura...")

	preferences := []interfaces.ReadPreference{
		interfaces.ReadPreferenceSecondary,
		interfaces.ReadPreferenceSecondaryPreferred,
		interfaces.ReadPreferenceNearest,
	}

	for _, preference := range preferences {
		fmt.Printf("\n   Preferência: %s\n", preference)

		replica, err := replicaManager.SelectReplica(ctx, preference)
		if err != nil {
			fmt.Printf("   Erro: %v\n", err)
			continue
		}

		fmt.Printf("   Réplica selecionada: %s\n", replica.GetID())
	}

	// Simular mudanças de status
	fmt.Println("\n6. Simulando mudanças de status...")

	// Configurar callbacks
	replicaManager.OnReplicaHealthChange(func(replica interfaces.IReplicaInfo, oldStatus, newStatus interfaces.ReplicaStatus) {
		fmt.Printf("   🔄 Réplica %s mudou de %s para %s\n", replica.GetID(), oldStatus, newStatus)
	})

	// Marcar uma réplica como não saudável
	replica1, err := replicaManager.GetReplica("replica-1")
	if err == nil {
		replica1.MarkUnhealthy()
		fmt.Printf("   ❌ Réplica %s marcada como não saudável\n", replica1.GetID())
	}

	// Marcar uma réplica como em manutenção
	err = replicaManager.SetReplicaMaintenance("replica-2", true)
	if err == nil {
		fmt.Printf("   🔧 Réplica replica-2 em modo de manutenção\n")
	}

	// Verificar réplicas saudáveis
	fmt.Println("\n7. Verificando réplicas saudáveis...")
	healthy := replicaManager.GetHealthyReplicas()
	fmt.Printf("   Réplicas saudáveis: %d\n", len(healthy))
	for _, replica := range healthy {
		fmt.Printf("   - %s (status: %s)\n", replica.GetID(), replica.GetStatus())
	}

	// Verificar réplicas não saudáveis
	unhealthy := replicaManager.GetUnhealthyReplicas()
	fmt.Printf("   Réplicas não saudáveis: %d\n", len(unhealthy))
	for _, replica := range unhealthy {
		fmt.Printf("   - %s (status: %s)\n", replica.GetID(), replica.GetStatus())
	}

	// Mostrar estatísticas
	fmt.Println("\n8. Estatísticas do sistema...")
	stats := replicaManager.GetStats()
	fmt.Printf("   Total de réplicas: %d\n", stats.GetTotalReplicas())
	fmt.Printf("   Réplicas saudáveis: %d\n", stats.GetHealthyReplicas())
	fmt.Printf("   Réplicas não saudáveis: %d\n", stats.GetUnhealthyReplicas())
	fmt.Printf("   Réplicas em manutenção: %d\n", stats.GetMaintenanceReplicas())
	fmt.Printf("   Total de queries: %d\n", stats.GetTotalQueries())
	fmt.Printf("   Uptime: %v\n", stats.GetUptime())

	// Recuperar réplica
	fmt.Println("\n9. Recuperando réplica...")
	if replica1 != nil {
		replica1.MarkHealthy()
		fmt.Printf("   ✅ Réplica %s recuperada\n", replica1.GetID())
	}

	// Verificar novamente
	healthy = replicaManager.GetHealthyReplicas()
	fmt.Printf("   Réplicas saudáveis após recuperação: %d\n", len(healthy))

	// Parar replica manager
	fmt.Println("\n10. Parando replica manager...")
	err = replicaManager.Stop(ctx)
	if err != nil {
		log.Printf("Erro ao parar replica manager: %v", err)
	} else {
		fmt.Println("✓ Replica manager parado com sucesso")
	}

	fmt.Println("\n=== Demonstração concluída ===")
}
