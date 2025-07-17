package pgxprovider

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/monitoring"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool implementa IPool com recursos avançados
type Pool struct {
	pool          *pgxpool.Pool
	config        interfaces.IConfig
	bufferPool    interfaces.IBufferPool
	safetyMonitor interfaces.ISafetyMonitor
	hookManager   interfaces.IHookManager
	monitor       *monitoring.ConnectionMonitor

	// Recursos avançados
	healthChecker    *HealthChecker
	loadBalancer     *LoadBalancer
	connectionWarmer *ConnectionWarmer
	metrics          *PoolMetrics

	// Estado
	mu     sync.RWMutex
	closed int32
	warmed int32

	// Controle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// HealthChecker implementa health checks periódicos
type HealthChecker struct {
	pool     *pgxpool.Pool
	interval time.Duration
	timeout  time.Duration
	metrics  *PoolMetrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// LoadBalancer implementa load balancing round-robin
type LoadBalancer struct {
	counter int64
	nodes   []string
	mu      sync.RWMutex
}

// ConnectionWarmer implementa connection warming
type ConnectionWarmer struct {
	pool        *pgxpool.Pool
	targetConns int32
	warmedConns int32
	metrics     *PoolMetrics
}

// PoolMetrics coleta métricas do pool
type PoolMetrics struct {
	// Contadores
	connectionsActive    int64
	connectionsIdle      int64
	connectionsCreated   int64
	connectionsDestroyed int64

	// Latências
	connectionLatency time.Duration
	queryLatency      time.Duration

	// Erros
	connectionErrors int64
	queryErrors      int64

	// Health
	healthChecks       int64
	healthChecksFailed int64

	mu sync.RWMutex
}

// NewPool cria um pool avançado com todos os recursos
func NewPool(ctx context.Context, config interfaces.IConfig,
	bufferPool interfaces.IBufferPool,
	safetyMonitor interfaces.ISafetyMonitor,
	hookManager interfaces.IHookManager) (interfaces.IPool, error) {

	// Parse da string de conexão
	pgxConfig, err := pgxpool.ParseConfig(config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Aplicar configurações avançadas do pool
	poolConfig := config.GetPoolConfig()
	pgxConfig.MaxConns = poolConfig.MaxConns
	pgxConfig.MinConns = poolConfig.MinConns
	pgxConfig.MaxConnLifetime = poolConfig.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = poolConfig.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = poolConfig.HealthCheckPeriod

	// Configurações avançadas
	pgxConfig.MaxConnLifetimeJitter = time.Duration(float64(poolConfig.MaxConnLifetime) * 0.1) // 10% jitter

	// Criar pool PGX
	pgxPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Criar contexto com cancel para lifecycle
	poolCtx, cancel := context.WithCancel(ctx)

	// Inicializar métricas
	metrics := &PoolMetrics{}

	// Criar pool avançado
	pool := &Pool{
		pool:          pgxPool,
		config:        config,
		bufferPool:    bufferPool,
		safetyMonitor: safetyMonitor,
		hookManager:   hookManager,
		monitor:       monitoring.NewConnectionMonitor(),
		metrics:       metrics,
		ctx:           poolCtx,
		cancel:        cancel,
	}

	// Inicializar componentes avançados
	pool.healthChecker = &HealthChecker{
		pool:     pgxPool,
		interval: 30 * time.Second,
		timeout:  5 * time.Second,
		metrics:  metrics,
	}

	pool.loadBalancer = &LoadBalancer{
		nodes: []string{config.GetConnectionString()}, // Pode ser expandido para múltiplos nodes
	}

	pool.connectionWarmer = &ConnectionWarmer{
		pool:        pgxPool,
		targetConns: poolConfig.MinConns,
		metrics:     metrics,
	}

	// Iniciar componentes em background
	pool.startBackgroundTasks()

	// Connection warming
	if err := pool.warmConnections(poolCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to warm connections: %w", err)
	}

	return pool, nil
}

// startBackgroundTasks inicia todas as tarefas de background
func (p *Pool) startBackgroundTasks() {
	// Health checks
	p.wg.Add(1)
	go p.runHealthChecks()

	// Metrics collection
	p.wg.Add(1)
	go p.collectMetrics()

	// Connection recycling
	p.wg.Add(1)
	go p.recycleConnections()
}

// warmConnections implementa connection warming
func (p *Pool) warmConnections(ctx context.Context) error {
	if atomic.LoadInt32(&p.warmed) == 1 {
		return nil
	}

	start := time.Now()
	defer func() {
		p.metrics.mu.Lock()
		p.metrics.connectionLatency = time.Since(start)
		p.metrics.mu.Unlock()
	}()

	// Warm up connections até minConns
	targetConns := p.config.GetPoolConfig().MinConns
	for i := int32(0); i < targetConns; i++ {
		conn, err := p.pool.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("failed to warm connection %d: %w", i, err)
		}

		// Teste básico da conexão
		if err := conn.Ping(ctx); err != nil {
			conn.Release()
			return fmt.Errorf("failed to ping warm connection %d: %w", i, err)
		}

		conn.Release()
		atomic.AddInt32(&p.connectionWarmer.warmedConns, 1)
		atomic.AddInt64(&p.metrics.connectionsCreated, 1)
	}

	atomic.StoreInt32(&p.warmed, 1)
	return nil
}

// runHealthChecks executa health checks periódicos
func (p *Pool) runHealthChecks() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.healthChecker.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck executa um health check
func (p *Pool) performHealthCheck() {
	ctx, cancel := context.WithTimeout(p.ctx, p.healthChecker.timeout)
	defer cancel()

	atomic.AddInt64(&p.metrics.healthChecks, 1)

	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		atomic.AddInt64(&p.metrics.healthChecksFailed, 1)
		return
	}
	defer conn.Release()

	if err := conn.Ping(ctx); err != nil {
		atomic.AddInt64(&p.metrics.healthChecksFailed, 1)
		return
	}
}

// collectMetrics coleta métricas em tempo real
func (p *Pool) collectMetrics() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
		}
	}
}

// updateMetrics atualiza as métricas atuais
func (p *Pool) updateMetrics() {
	stats := p.pool.Stat()

	atomic.StoreInt64(&p.metrics.connectionsActive, int64(stats.AcquiredConns()))
	atomic.StoreInt64(&p.metrics.connectionsIdle, int64(stats.IdleConns()))
	atomic.StoreInt64(&p.metrics.connectionsCreated, int64(stats.NewConnsCount()))
}

// recycleConnections implementa connection recycling
func (p *Pool) recycleConnections() {
	defer p.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Força limpeza de conexões idle antigas
			// O pgxpool já faz isso automaticamente, mas podemos forçar
			p.pool.Reset()
		}
	}
}

// Acquire implementa IPool.Acquire
func (p *Pool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrPoolClosed
	}

	start := time.Now()

	// Executar hook de acquire
	if p.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "acquire",
			StartTime: start,
		}
		if err := p.hookManager.ExecuteHooks(interfaces.BeforeAcquireHook, execCtx); err != nil {
			return nil, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			p.hookManager.ExecuteHooks(interfaces.AfterAcquireHook, execCtx)
		}()
	}

	// Usar load balancer para distribuir conexões
	node := p.loadBalancer.getNextNode()
	_ = node // Por enquanto só temos um node

	// Adquirir conexão do pool
	pgxConn, err := p.pool.Acquire(ctx)
	if err != nil {
		atomic.AddInt64(&p.metrics.connectionErrors, 1)
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	// Atualizar métricas
	p.metrics.mu.Lock()
	p.metrics.connectionLatency = time.Since(start)
	p.metrics.mu.Unlock()

	// Criar wrapper da conexão
	conn := &Conn{
		conn:        pgxConn,
		config:      p.config,
		bufferPool:  p.bufferPool,
		hookManager: p.hookManager,
		monitor:     p.monitor,
		acquired:    true,
		fromPool:    true,
	}

	return conn, nil
}

// AcquireFunc implementa IPool.AcquireFunc
func (p *Pool) AcquireFunc(ctx context.Context, f func(interfaces.IConn) error) error {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return f(conn)
}

// Close implementa graceful shutdown
func (p *Pool) Close() {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return // Já fechado
	}

	// Cancelar contexto para parar background tasks
	p.cancel()

	// Aguardar todas as goroutines terminarem
	p.wg.Wait()

	// Fechar pool
	p.pool.Close()
}

// Reset implementa IPool.Reset
func (p *Pool) Reset() {
	p.pool.Reset()
	// Resetar métricas
	p.monitor = monitoring.NewConnectionMonitor()
}

// Stats implementa IPool.Stats
func (p *Pool) Stats() interfaces.PoolStats {
	if atomic.LoadInt32(&p.closed) == 1 {
		return interfaces.PoolStats{}
	}

	stats := p.pool.Stat()

	return interfaces.PoolStats{
		AcquiredConns:           int32(stats.AcquiredConns()),
		IdleConns:               int32(stats.IdleConns()),
		TotalConns:              int32(stats.TotalConns()),
		NewConnsCount:           int64(stats.NewConnsCount()),
		MaxConns:                int32(stats.MaxConns()),
		AcquireCount:            atomic.LoadInt64(&p.metrics.connectionsCreated),
		AcquireDuration:         p.metrics.connectionLatency,
		CanceledAcquireCount:    atomic.LoadInt64(&p.metrics.connectionErrors),
		ConstructingConns:       0, // PGX não expõe essa métrica
		EmptyAcquireCount:       0, // PGX não expõe essa métrica
		MaxLifetimeDestroyCount: 0, // PGX não expõe essa métrica
		MaxIdleDestroyCount:     0, // PGX não expõe essa métrica
	}
}

// GetStats retorna estatísticas detalhadas do pool
func (p *Pool) GetStats() map[string]interface{} {
	stats := p.pool.Stat()

	return map[string]interface{}{
		"acquired_conns":       stats.AcquiredConns(),
		"idle_conns":           stats.IdleConns(),
		"total_conns":          stats.TotalConns(),
		"new_conns_count":      stats.NewConnsCount(),
		"max_conns":            stats.MaxConns(),
		"connections_active":   atomic.LoadInt64(&p.metrics.connectionsActive),
		"connections_idle":     atomic.LoadInt64(&p.metrics.connectionsIdle),
		"connections_created":  atomic.LoadInt64(&p.metrics.connectionsCreated),
		"connection_errors":    atomic.LoadInt64(&p.metrics.connectionErrors),
		"health_checks":        atomic.LoadInt64(&p.metrics.healthChecks),
		"health_checks_failed": atomic.LoadInt64(&p.metrics.healthChecksFailed),
		"warmed_connections":   atomic.LoadInt32(&p.connectionWarmer.warmedConns),
		"is_warmed":            atomic.LoadInt32(&p.warmed) == 1,
	}
}

// Config implementa IPool.Config
func (p *Pool) Config() interfaces.PoolConfig {
	return p.config.GetPoolConfig()
}

// Ping implementa IPool.Ping
func (p *Pool) Ping(ctx context.Context) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	return p.pool.Ping(ctx)
}

// HealthCheck implementa IPool.HealthCheck
func (p *Pool) HealthCheck(ctx context.Context) error {
	if !p.IsHealthy() {
		return fmt.Errorf("pool is not healthy")
	}

	if err := p.Ping(ctx); err != nil {
		return err
	}

	// Verificar se safety monitor está saudável
	if p.safetyMonitor != nil && !p.safetyMonitor.IsHealthy() {
		return ErrUnhealthyState
	}

	return nil
}

// GetHookManager implementa IPool.GetHookManager
func (p *Pool) GetHookManager() interfaces.IHookManager {
	return p.hookManager
}

// GetBufferPool implementa IPool.GetBufferPool
func (p *Pool) GetBufferPool() interfaces.IBufferPool {
	return p.bufferPool
}

// GetSafetyMonitor implementa IPool.GetSafetyMonitor
func (p *Pool) GetSafetyMonitor() interfaces.ISafetyMonitor {
	return p.safetyMonitor
}

// IsHealthy verifica se o pool está saudável
func (p *Pool) IsHealthy() bool {
	if atomic.LoadInt32(&p.closed) == 1 {
		return false
	}

	// Verificar se há conexões disponíveis
	stats := p.pool.Stat()
	return stats.TotalConns() > 0
}

// IsClosed verifica se o pool está fechado
func (p *Pool) IsClosed() bool {
	return atomic.LoadInt32(&p.closed) == 1
}

// LoadBalancer methods
func (lb *LoadBalancer) getNextNode() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.nodes) == 0 {
		return ""
	}

	if len(lb.nodes) == 1 {
		return lb.nodes[0]
	}

	// Round-robin
	next := atomic.AddInt64(&lb.counter, 1)
	return lb.nodes[next%int64(len(lb.nodes))]
}

func (lb *LoadBalancer) addNode(node string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.nodes = append(lb.nodes, node)
}

func (lb *LoadBalancer) removeNode(node string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, n := range lb.nodes {
		if n == node {
			lb.nodes = append(lb.nodes[:i], lb.nodes[i+1:]...)
			break
		}
	}
}
