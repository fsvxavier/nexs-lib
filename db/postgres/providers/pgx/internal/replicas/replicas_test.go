package replicas

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// MockPool implementa IPool para testes
type MockPool struct {
	dsn      string
	closed   bool
	acquirer func(ctx context.Context) (interfaces.IConn, error)
}

func NewMockPool(dsn string) *MockPool {
	return &MockPool{
		dsn: dsn,
		acquirer: func(ctx context.Context) (interfaces.IConn, error) {
			return &MockConn{}, nil
		},
	}
}

func (m *MockPool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	if m.closed {
		return nil, fmt.Errorf("pool is closed")
	}
	return m.acquirer(ctx)
}

func (m *MockPool) Release(conn interfaces.IConn) {}

func (m *MockPool) Close() {
	m.closed = true
}

func (m *MockPool) AcquireFunc(ctx context.Context, f func(interfaces.IConn) error) error {
	conn, err := m.acquirer(ctx)
	if err != nil {
		return err
	}
	return f(conn)
}

func (m *MockPool) Reset() {}

func (m *MockPool) Config() interfaces.PoolConfig {
	return interfaces.PoolConfig{}
}

func (m *MockPool) Ping(ctx context.Context) error {
	return nil
}

func (m *MockPool) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockPool) GetHookManager() interfaces.IHookManager {
	return nil
}

func (m *MockPool) GetBufferPool() interfaces.IBufferPool {
	return nil
}

func (m *MockPool) GetSafetyMonitor() interfaces.ISafetyMonitor {
	return nil
}

func (m *MockPool) Stats() interfaces.PoolStats {
	return interfaces.PoolStats{}
}

// MockConn implementa IConn para testes
type MockConn struct {
	closed bool
}

func (m *MockConn) Ping(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("connection is closed")
	}
	return nil
}

func (m *MockConn) Close(ctx context.Context) error {
	m.closed = true
	return nil
}

func (m *MockConn) Release() {}

func (m *MockConn) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockConn) BeginTx(ctx context.Context, opts interfaces.TxOptions) (interfaces.ITransaction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockConn) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.ICommandTag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockConn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockConn) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	return nil
}

func (m *MockConn) QueryAll(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) CopyTo(ctx context.Context, w interfaces.ICopyToWriter, query string, args ...interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.ICopyFromSource) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockConn) IsClosed() bool {
	return m.closed
}

func (m *MockConn) Prepare(ctx context.Context, name, query string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) Deallocate(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) Unlisten(ctx context.Context, channel string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockConn) SetTenant(ctx context.Context, tenantID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) GetTenant(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (m *MockConn) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	return nil
}

func (m *MockConn) GetHookManager() interfaces.IHookManager {
	return nil
}

func (m *MockConn) HealthCheck(ctx context.Context) error {
	return m.Ping(ctx)
}

func (m *MockConn) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockConn) QueryOne(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) Batch() interfaces.IBatch {
	return nil
}

func (m *MockConn) Listen(ctx context.Context, channel string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) Notify(ctx context.Context, channel string, payload string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockConn) Stats() interfaces.ConnectionStats {
	return interfaces.ConnectionStats{}
}

// TestReplicaInfo testa a estrutura ReplicaInfo
func TestReplicaInfo(t *testing.T) {
	replica := NewReplicaInfo("test-replica", "postgres://localhost:5432/test", 10)

	// Test basic getters
	if replica.GetID() != "test-replica" {
		t.Errorf("Expected ID 'test-replica', got %s", replica.GetID())
	}

	if replica.GetWeight() != 10 {
		t.Errorf("Expected weight 10, got %d", replica.GetWeight())
	}

	if replica.GetStatus() != interfaces.ReplicaStatusHealthy {
		t.Errorf("Expected status healthy, got %s", replica.GetStatus())
	}

	// Test status changes
	replica.MarkUnhealthy()
	if replica.GetStatus() != interfaces.ReplicaStatusUnhealthy {
		t.Errorf("Expected status unhealthy, got %s", replica.GetStatus())
	}

	replica.MarkRecovering()
	if replica.GetStatus() != interfaces.ReplicaStatusRecovering {
		t.Errorf("Expected status recovering, got %s", replica.GetStatus())
	}

	replica.MarkMaintenance()
	if replica.GetStatus() != interfaces.ReplicaStatusMaintenance {
		t.Errorf("Expected status maintenance, got %s", replica.GetStatus())
	}

	replica.MarkHealthy()
	if replica.GetStatus() != interfaces.ReplicaStatusHealthy {
		t.Errorf("Expected status healthy, got %s", replica.GetStatus())
	}

	// Test availability
	if !replica.IsAvailable() {
		t.Error("Expected replica to be available")
	}

	replica.MarkUnhealthy()
	if replica.IsAvailable() {
		t.Error("Expected replica to be unavailable")
	}

	// Test connection counting
	replica.IncrementConnections()
	if replica.GetConnectionCount() != 1 {
		t.Errorf("Expected connection count 1, got %d", replica.GetConnectionCount())
	}

	replica.DecrementConnections()
	if replica.GetConnectionCount() != 0 {
		t.Errorf("Expected connection count 0, got %d", replica.GetConnectionCount())
	}

	// Test query recording
	replica.RecordQuery(true, 10*time.Millisecond)
	if replica.GetTotalQueries() != 1 {
		t.Errorf("Expected total queries 1, got %d", replica.GetTotalQueries())
	}

	if replica.GetFailedQueries() != 0 {
		t.Errorf("Expected failed queries 0, got %d", replica.GetFailedQueries())
	}

	replica.RecordQuery(false, 20*time.Millisecond)
	if replica.GetTotalQueries() != 2 {
		t.Errorf("Expected total queries 2, got %d", replica.GetTotalQueries())
	}

	if replica.GetFailedQueries() != 1 {
		t.Errorf("Expected failed queries 1, got %d", replica.GetFailedQueries())
	}

	// Test rates
	successRate := replica.GetSuccessRate()
	if successRate != 50.0 {
		t.Errorf("Expected success rate 50.0, got %f", successRate)
	}

	errorRate := replica.GetErrorRate()
	if errorRate != 50.0 {
		t.Errorf("Expected error rate 50.0, got %f", errorRate)
	}

	// Test tags
	tags := map[string]string{
		"region": "us-east-1",
		"zone":   "1a",
	}
	replica.SetTags(tags)

	returnedTags := replica.GetTags()
	if returnedTags["region"] != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %s", returnedTags["region"])
	}

	// Test pool
	pool := NewMockPool("postgres://localhost:5432/test")
	replica.SetPool(pool)

	if replica.GetPool() == nil {
		t.Error("Expected pool to be set correctly")
	} // Test close
	err := replica.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

// TestLoadBalancer testa o load balancer
func TestLoadBalancer(t *testing.T) {
	ctx := context.Background()

	// Criar réplicas mock
	replicas := []interfaces.IReplicaInfo{
		NewReplicaInfo("replica1", "postgres://host1:5432/db", 10),
		NewReplicaInfo("replica2", "postgres://host2:5432/db", 20),
		NewReplicaInfo("replica3", "postgres://host3:5432/db", 30),
	}

	// Test round-robin
	lb := NewLoadBalancer(interfaces.LoadBalancingRoundRobin)

	selected1, err := lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	selected2, err := lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	selected3, err := lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should cycle through replicas
	if selected1.GetID() == selected2.GetID() {
		t.Error("Expected different replicas in round-robin")
	}

	if selected2.GetID() == selected3.GetID() {
		t.Error("Expected different replicas in round-robin")
	}

	// Test random
	lb.SetStrategy(interfaces.LoadBalancingRandom)
	selected, err := lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if selected == nil {
		t.Error("Expected replica to be selected")
	}

	// Test weighted
	lb.SetStrategy(interfaces.LoadBalancingWeighted)
	selected, err = lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if selected == nil {
		t.Error("Expected replica to be selected")
	}

	// Test latency-based
	// Primeiro definir algumas latências
	replicas[0].(*ReplicaInfo).UpdateLatency(100 * time.Millisecond)
	replicas[1].(*ReplicaInfo).UpdateLatency(50 * time.Millisecond)
	replicas[2].(*ReplicaInfo).UpdateLatency(200 * time.Millisecond)

	lb.SetStrategy(interfaces.LoadBalancingLatency)
	selected, err = lb.SelectReplica(ctx, replicas)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should select replica with lowest latency (replica2)
	if selected.GetID() != "replica2" {
		t.Errorf("Expected replica2 (lowest latency), got %s", selected.GetID())
	}

	// Test with no replicas
	_, err = lb.SelectReplica(ctx, []interfaces.IReplicaInfo{})
	if err == nil {
		t.Error("Expected error with no replicas")
	}

	// Test with unhealthy replicas
	replicas[0].MarkUnhealthy()
	replicas[1].MarkUnhealthy()
	replicas[2].MarkUnhealthy()

	_, err = lb.SelectReplica(ctx, replicas)
	if err == nil {
		t.Error("Expected error with no healthy replicas")
	}
}

// TestReplicaStats testa as estatísticas
func TestReplicaStats(t *testing.T) {
	stats := NewReplicaStats()

	// Test initial values
	if stats.GetTotalReplicas() != 0 {
		t.Errorf("Expected total replicas 0, got %d", stats.GetTotalReplicas())
	}

	if stats.GetTotalQueries() != 0 {
		t.Errorf("Expected total queries 0, got %d", stats.GetTotalQueries())
	}

	// Test replica count update
	stats.UpdateReplicaCount(5, 3, 1, 1, 0)
	if stats.GetTotalReplicas() != 5 {
		t.Errorf("Expected total replicas 5, got %d", stats.GetTotalReplicas())
	}

	if stats.GetHealthyReplicas() != 3 {
		t.Errorf("Expected healthy replicas 3, got %d", stats.GetHealthyReplicas())
	}

	if stats.GetUnhealthyReplicas() != 1 {
		t.Errorf("Expected unhealthy replicas 1, got %d", stats.GetUnhealthyReplicas())
	}

	// Test query recording
	stats.RecordQuery("replica1", true, 10*time.Millisecond)
	if stats.GetTotalQueries() != 1 {
		t.Errorf("Expected total queries 1, got %d", stats.GetTotalQueries())
	}

	if stats.GetSuccessfulQueries() != 1 {
		t.Errorf("Expected successful queries 1, got %d", stats.GetSuccessfulQueries())
	}

	stats.RecordQuery("replica2", false, 20*time.Millisecond)
	if stats.GetTotalQueries() != 2 {
		t.Errorf("Expected total queries 2, got %d", stats.GetTotalQueries())
	}

	if stats.GetFailedQueries() != 1 {
		t.Errorf("Expected failed queries 1, got %d", stats.GetFailedQueries())
	}

	// Test distributions
	queryDist := stats.GetQueryDistribution()
	if queryDist["replica1"] != 1 {
		t.Errorf("Expected replica1 queries 1, got %d", queryDist["replica1"])
	}

	if queryDist["replica2"] != 1 {
		t.Errorf("Expected replica2 queries 1, got %d", queryDist["replica2"])
	}

	errorDist := stats.GetErrorDistribution()
	if errorDist["replica1"] != 0 {
		t.Errorf("Expected replica1 errors 0, got %d", errorDist["replica1"])
	}

	if errorDist["replica2"] != 1 {
		t.Errorf("Expected replica2 errors 1, got %d", errorDist["replica2"])
	}

	// Test failover recording
	stats.RecordFailover()
	if stats.GetFailoverCount() != 1 {
		t.Errorf("Expected failover count 1, got %d", stats.GetFailoverCount())
	}

	// Test JSON export
	jsonData, err := stats.ToJSON()
	if err != nil {
		t.Errorf("Expected no error on JSON export, got %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected JSON data to be non-empty")
	}

	// Test map export
	mapData := stats.ToMap()
	if mapData["total_replicas"] != 5 {
		t.Errorf("Expected total_replicas 5, got %v", mapData["total_replicas"])
	}

	// Test reset
	stats.Reset()
	if stats.GetTotalReplicas() != 0 {
		t.Errorf("Expected total replicas 0 after reset, got %d", stats.GetTotalReplicas())
	}

	if stats.GetTotalQueries() != 0 {
		t.Errorf("Expected total queries 0 after reset, got %d", stats.GetTotalQueries())
	}
}
