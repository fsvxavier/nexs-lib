package pagination

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

// PooledQueryBuilder implements a poolable query builder
type PooledQueryBuilder struct {
	mu            sync.RWMutex
	baseQuery     string
	orderByClause string
	limitClause   string
	offsetClause  string
}

// NewPooledQueryBuilder creates a new poolable query builder
func NewPooledQueryBuilder() *PooledQueryBuilder {
	return &PooledQueryBuilder{}
}

// BuildQuery constructs the final query with ORDER BY and LIMIT/OFFSET
func (b *PooledQueryBuilder) BuildQuery(baseQuery string, params *interfaces.PaginationParams) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.baseQuery = baseQuery
	b.buildOrderByClause(params)
	b.buildLimitOffsetClause(params)

	return b.assembleQuery()
}

// BuildCountQuery constructs a count query for total records
func (b *PooledQueryBuilder) BuildCountQuery(baseQuery string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Extract the main query without ORDER BY
	query := strings.TrimSpace(baseQuery)

	// Handle subqueries and complex queries
	if strings.Contains(strings.ToUpper(query), "ORDER BY") {
		orderByIndex := strings.LastIndex(strings.ToUpper(query), "ORDER BY")
		query = strings.TrimSpace(query[:orderByIndex])
	}

	// Wrap in count query
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_query", query)
}

// Reset resets the builder state for reuse
func (b *PooledQueryBuilder) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.baseQuery = ""
	b.orderByClause = ""
	b.limitClause = ""
	b.offsetClause = ""
}

// Clone creates a copy of the builder
func (b *PooledQueryBuilder) Clone() interfaces.PoolableQueryBuilder {
	b.mu.RLock()
	defer b.mu.RUnlock()

	clone := &PooledQueryBuilder{
		baseQuery:     b.baseQuery,
		orderByClause: b.orderByClause,
		limitClause:   b.limitClause,
		offsetClause:  b.offsetClause,
	}

	return clone
}

// buildOrderByClause builds the ORDER BY clause
func (b *PooledQueryBuilder) buildOrderByClause(params *interfaces.PaginationParams) {
	if params.SortField == "" {
		b.orderByClause = ""
		return
	}

	// Sanitize sort field (basic protection)
	sortField := strings.ReplaceAll(params.SortField, ";", "")
	sortField = strings.ReplaceAll(sortField, "--", "")

	sortOrder := "ASC"
	if strings.ToUpper(params.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	b.orderByClause = fmt.Sprintf("ORDER BY %s %s", sortField, sortOrder)
}

// buildLimitOffsetClause builds the LIMIT and OFFSET clauses
func (b *PooledQueryBuilder) buildLimitOffsetClause(params *interfaces.PaginationParams) {
	offset := (params.Page - 1) * params.Limit
	b.limitClause = fmt.Sprintf("LIMIT %d", params.Limit)
	b.offsetClause = fmt.Sprintf("OFFSET %d", offset)
}

// assembleQuery assembles the final query
func (b *PooledQueryBuilder) assembleQuery() string {
	var parts []string

	parts = append(parts, b.baseQuery)

	if b.orderByClause != "" {
		parts = append(parts, b.orderByClause)
	}

	if b.limitClause != "" {
		parts = append(parts, b.limitClause)
	}

	if b.offsetClause != "" {
		parts = append(parts, b.offsetClause)
	}

	return strings.Join(parts, " ")
}

// DefaultQueryBuilderPool implements a simple query builder pool
type DefaultQueryBuilderPool struct {
	pool sync.Pool
	size int
	mu   sync.RWMutex
}

// NewDefaultQueryBuilderPool creates a new query builder pool
func NewDefaultQueryBuilderPool(initialSize int) *DefaultQueryBuilderPool {
	pool := &DefaultQueryBuilderPool{
		size: 0,
	}

	pool.pool = sync.Pool{
		New: func() interface{} {
			return NewPooledQueryBuilder()
		},
	}

	// Pre-populate pool
	for i := 0; i < initialSize; i++ {
		pool.Put(NewPooledQueryBuilder())
	}

	return pool
}

// Get retrieves a builder from the pool
func (p *DefaultQueryBuilderPool) Get() interfaces.PoolableQueryBuilder {
	builder := p.pool.Get().(*PooledQueryBuilder)
	builder.Reset() // Ensure clean state

	p.mu.Lock()
	if p.size > 0 {
		p.size--
	}
	p.mu.Unlock()

	return builder
}

// Put returns a builder to the pool
func (p *DefaultQueryBuilderPool) Put(builder interfaces.PoolableQueryBuilder) {
	if builder == nil {
		return
	}

	// Reset builder before returning to pool
	builder.Reset()

	p.mu.Lock()
	p.size++
	p.mu.Unlock()

	p.pool.Put(builder)
}

// Size returns the current pool size
func (p *DefaultQueryBuilderPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.size
}
