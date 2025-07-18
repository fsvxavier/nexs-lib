package replicas

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// LoadBalancer implementa estratégias de balanceamento de carga
type LoadBalancer struct {
	strategy        interfaces.LoadBalancingStrategy
	roundRobinIndex int
	mu              sync.RWMutex
}

// NewLoadBalancer cria um novo load balancer
func NewLoadBalancer(strategy interfaces.LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
	}
}

// SelectReplica seleciona uma réplica baseada na estratégia configurada
func (lb *LoadBalancer) SelectReplica(ctx context.Context, replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available")
	}

	// Filtrar apenas réplicas disponíveis
	available := make([]interfaces.IReplicaInfo, 0, len(replicas))
	for _, replica := range replicas {
		if replica.IsAvailable() {
			available = append(available, replica)
		}
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no healthy replicas available")
	}

	switch lb.strategy {
	case interfaces.LoadBalancingRoundRobin:
		return lb.selectRoundRobin(available)
	case interfaces.LoadBalancingRandom:
		return lb.selectRandom(available)
	case interfaces.LoadBalancingWeighted:
		return lb.selectWeighted(available)
	case interfaces.LoadBalancingLatency:
		return lb.selectLatency(available)
	default:
		return lb.selectRoundRobin(available)
	}
}

// selectRoundRobin implementa seleção round-robin
func (lb *LoadBalancer) selectRoundRobin(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for round-robin")
	}

	selected := replicas[lb.roundRobinIndex%len(replicas)]
	lb.roundRobinIndex++

	return selected, nil
}

// selectRandom implementa seleção aleatória
func (lb *LoadBalancer) selectRandom(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for random selection")
	}

	index := rand.Intn(len(replicas))
	return replicas[index], nil
}

// selectWeighted implementa seleção baseada em peso
func (lb *LoadBalancer) selectWeighted(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for weighted selection")
	}

	// Calcular peso total
	totalWeight := 0
	for _, replica := range replicas {
		weight := replica.GetWeight()
		if weight <= 0 {
			weight = 1 // Peso mínimo
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return lb.selectRoundRobin(replicas)
	}

	// Selecionar baseado no peso
	target := rand.Intn(totalWeight)
	current := 0

	for _, replica := range replicas {
		weight := replica.GetWeight()
		if weight <= 0 {
			weight = 1
		}
		current += weight
		if current > target {
			return replica, nil
		}
	}

	// Fallback para o primeiro se algo der errado
	return replicas[0], nil
}

// selectLatency implementa seleção baseada na latência
func (lb *LoadBalancer) selectLatency(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for latency-based selection")
	}

	// Ordenar por latência (menor primeiro)
	sorted := make([]interfaces.IReplicaInfo, len(replicas))
	copy(sorted, replicas)

	sort.Slice(sorted, func(i, j int) bool {
		latencyI := sorted[i].GetAvgLatency()
		latencyJ := sorted[j].GetAvgLatency()

		// Se uma das latências for zero, usar critério alternativo
		if latencyI == 0 && latencyJ == 0 {
			return sorted[i].GetSuccessRate() > sorted[j].GetSuccessRate()
		}
		if latencyI == 0 {
			return false
		}
		if latencyJ == 0 {
			return true
		}

		return latencyI < latencyJ
	})

	return sorted[0], nil
}

// GetStrategy retorna a estratégia atual
func (lb *LoadBalancer) GetStrategy() interfaces.LoadBalancingStrategy {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.strategy
}

// SetStrategy define a estratégia de balanceamento
func (lb *LoadBalancer) SetStrategy(strategy interfaces.LoadBalancingStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.strategy = strategy
}

// Reset reseta o estado interno do load balancer
func (lb *LoadBalancer) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.roundRobinIndex = 0
}

// WeightedReplica representa uma réplica com peso para algoritmos mais complexos
type WeightedReplica struct {
	Replica interfaces.IReplicaInfo
	Weight  int
	Current int
}

// selectWeightedSmooth implementa Weighted Round Robin suave (smooth)
func (lb *LoadBalancer) selectWeightedSmooth(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for weighted smooth selection")
	}

	// Criar estrutura de pesos
	weighted := make([]*WeightedReplica, len(replicas))
	totalWeight := 0

	for i, replica := range replicas {
		weight := replica.GetWeight()
		if weight <= 0 {
			weight = 1
		}
		weighted[i] = &WeightedReplica{
			Replica: replica,
			Weight:  weight,
			Current: 0,
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return replicas[0], nil
	}

	// Algoritmo Weighted Round Robin suave
	var best *WeightedReplica
	for _, wr := range weighted {
		wr.Current += wr.Weight
		if best == nil || wr.Current > best.Current {
			best = wr
		}
	}

	best.Current -= totalWeight
	return best.Replica, nil
}

// selectLeastConnections implementa seleção baseada em menor número de conexões
func (lb *LoadBalancer) selectLeastConnections(replicas []interfaces.IReplicaInfo) (interfaces.IReplicaInfo, error) {
	if len(replicas) == 0 {
		return nil, fmt.Errorf("no replicas available for least connections selection")
	}

	var best interfaces.IReplicaInfo
	minConnections := int(^uint(0) >> 1) // Max int

	for _, replica := range replicas {
		connections := replica.GetConnectionCount()
		if connections < minConnections {
			minConnections = connections
			best = replica
		}
	}

	if best == nil {
		return replicas[0], nil
	}

	return best, nil
}
