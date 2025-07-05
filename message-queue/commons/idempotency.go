package commons

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// IdempotencyManager define a interface para gerenciamento de idempotência
type IdempotencyManager interface {
	// IsProcessed verifica se uma mensagem já foi processada
	IsProcessed(ctx context.Context, messageID string) (bool, error)

	// MarkAsProcessed marca uma mensagem como processada
	MarkAsProcessed(ctx context.Context, messageID string) error

	// MarkAsProcessedWithTTL marca uma mensagem como processada com TTL específico
	MarkAsProcessedWithTTL(ctx context.Context, messageID string, ttl time.Duration) error

	// Remove remove uma entrada de idempotência
	Remove(ctx context.Context, messageID string) error

	// Clear limpa todas as entradas de idempotência
	Clear(ctx context.Context) error

	// GetStats retorna estatísticas do cache de idempotência
	GetStats() *IdempotencyStats
}

// IdempotencyStorage define a interface para storage de idempotência
type IdempotencyStorage interface {
	// Set armazena uma chave com TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get recupera um valor por chave
	Get(ctx context.Context, key string) (interface{}, error)

	// Exists verifica se uma chave existe
	Exists(ctx context.Context, key string) (bool, error)

	// Delete remove uma chave
	Delete(ctx context.Context, key string) error

	// Clear remove todas as chaves
	Clear(ctx context.Context) error

	// Close fecha o storage
	Close() error
}

// IdempotencyStats representa estatísticas de idempotência
type IdempotencyStats struct {
	// Número total de verificações
	TotalChecks int64

	// Número de hits (mensagens já processadas)
	Hits int64

	// Número de misses (mensagens não processadas)
	Misses int64

	// Taxa de hit
	HitRate float64

	// Número total de entradas armazenadas
	TotalEntries int64

	// Última verificação
	LastCheck time.Time
}

// MemoryIdempotencyManager implementa IdempotencyManager usando memória
type MemoryIdempotencyManager struct {
	storage map[string]*idempotencyEntry
	mutex   sync.RWMutex
	ttl     time.Duration
	stats   *IdempotencyStats
}

// idempotencyEntry representa uma entrada no cache de idempotência
type idempotencyEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewMemoryIdempotencyManager cria um novo gerenciador de idempotência em memória
func NewMemoryIdempotencyManager(ttl time.Duration) IdempotencyManager {
	manager := &MemoryIdempotencyManager{
		storage: make(map[string]*idempotencyEntry),
		ttl:     ttl,
		stats: &IdempotencyStats{
			TotalChecks:  0,
			Hits:         0,
			Misses:       0,
			HitRate:      0.0,
			TotalEntries: 0,
			LastCheck:    time.Now(),
		},
	}

	// Inicia limpeza automática de entradas expiradas
	go manager.cleanupExpired()

	return manager
}

// IsProcessed verifica se uma mensagem já foi processada
func (m *MemoryIdempotencyManager) IsProcessed(ctx context.Context, messageID string) (bool, error) {
	if messageID == "" {
		return false, domainerrors.New(
			"INVALID_MESSAGE_ID",
			"message ID cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	key := m.generateKey(messageID)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Atualiza estatísticas
	m.stats.TotalChecks++
	m.stats.LastCheck = time.Now()

	entry, exists := m.storage[key]
	if !exists {
		m.stats.Misses++
		m.updateHitRate()
		return false, nil
	}

	// Verifica se a entrada expirou
	if time.Now().After(entry.ExpiresAt) {
		m.stats.Misses++
		m.updateHitRate()
		return false, nil
	}

	m.stats.Hits++
	m.updateHitRate()
	return true, nil
}

// MarkAsProcessed marca uma mensagem como processada
func (m *MemoryIdempotencyManager) MarkAsProcessed(ctx context.Context, messageID string) error {
	return m.MarkAsProcessedWithTTL(ctx, messageID, m.ttl)
}

// MarkAsProcessedWithTTL marca uma mensagem como processada com TTL específico
func (m *MemoryIdempotencyManager) MarkAsProcessedWithTTL(ctx context.Context, messageID string, ttl time.Duration) error {
	if messageID == "" {
		return domainerrors.New(
			"INVALID_MESSAGE_ID",
			"message ID cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	key := m.generateKey(messageID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.storage[key] = &idempotencyEntry{
		Value:     true,
		ExpiresAt: time.Now().Add(ttl),
	}

	m.stats.TotalEntries = int64(len(m.storage))

	return nil
}

// Remove remove uma entrada de idempotência
func (m *MemoryIdempotencyManager) Remove(ctx context.Context, messageID string) error {
	if messageID == "" {
		return domainerrors.New(
			"INVALID_MESSAGE_ID",
			"message ID cannot be empty",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	key := m.generateKey(messageID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.storage, key)
	m.stats.TotalEntries = int64(len(m.storage))

	return nil
}

// Clear limpa todas as entradas de idempotência
func (m *MemoryIdempotencyManager) Clear(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.storage = make(map[string]*idempotencyEntry)
	m.stats.TotalEntries = 0

	return nil
}

// GetStats retorna estatísticas do cache de idempotência
func (m *MemoryIdempotencyManager) GetStats() *IdempotencyStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Cria uma cópia das estatísticas para evitar race conditions
	return &IdempotencyStats{
		TotalChecks:  m.stats.TotalChecks,
		Hits:         m.stats.Hits,
		Misses:       m.stats.Misses,
		HitRate:      m.stats.HitRate,
		TotalEntries: m.stats.TotalEntries,
		LastCheck:    m.stats.LastCheck,
	}
}

// generateKey gera uma chave hash para o messageID
func (m *MemoryIdempotencyManager) generateKey(messageID string) string {
	hash := sha256.Sum256([]byte(messageID))
	return hex.EncodeToString(hash[:])
}

// updateHitRate atualiza a taxa de hit do cache
func (m *MemoryIdempotencyManager) updateHitRate() {
	if m.stats.TotalChecks > 0 {
		m.stats.HitRate = float64(m.stats.Hits) / float64(m.stats.TotalChecks)
	}
}

// cleanupExpired remove entradas expiradas do cache
func (m *MemoryIdempotencyManager) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // Limpeza a cada 5 minutos
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()

		for key, entry := range m.storage {
			if now.After(entry.ExpiresAt) {
				delete(m.storage, key)
			}
		}

		m.stats.TotalEntries = int64(len(m.storage))
		m.mutex.Unlock()
	}
}

// GenerateMessageKey gera uma chave única para uma mensagem
func GenerateMessageKey(message *interfaces.Message) string {
	// Se a mensagem já tem um ID, usa ele
	if message.ID != "" {
		return message.ID
	}

	// Gera uma chave baseada no conteúdo da mensagem
	content := fmt.Sprintf("%s:%s:%d",
		string(message.Body),
		message.Source,
		message.Timestamp.UnixNano(),
	)

	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// IsIdempotentMessage verifica se uma mensagem deve ser tratada como idempotente
func IsIdempotentMessage(message *interfaces.Message) bool {
	// Verifica se tem headers específicos para idempotência
	if message.Headers != nil {
		if idempotent, exists := message.Headers["idempotent"]; exists {
			if val, ok := idempotent.(bool); ok {
				return val
			}
			if val, ok := idempotent.(string); ok {
				return val == "true" || val == "1"
			}
		}

		// Verifica se tem messageID nos headers
		if messageID, exists := message.Headers["messageId"]; exists {
			return messageID != nil && messageID != ""
		}
	}

	// Por padrão, considera como idempotente se tem um ID
	return message.ID != ""
}
