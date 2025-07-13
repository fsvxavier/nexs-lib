package tracer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// PerformanceConfig configures performance optimizations
type PerformanceConfig struct {
	EnableSpanPooling    bool `json:"enable_span_pooling"`
	SpanPoolSize         int  `json:"span_pool_size"`
	EnableFastPaths      bool `json:"enable_fast_paths"`
	EnableZeroAlloc      bool `json:"enable_zero_alloc"`
	MaxAttributesPerSpan int  `json:"max_attributes_per_span"`
	MaxEventsPerSpan     int  `json:"max_events_per_span"`
	BatchSize            int  `json:"batch_size"`
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		EnableSpanPooling:    true,
		SpanPoolSize:         1000,
		EnableFastPaths:      true,
		EnableZeroAlloc:      true,
		MaxAttributesPerSpan: 128,
		MaxEventsPerSpan:     64,
		BatchSize:            512,
	}
}

// SpanPool manages a pool of reusable span objects to reduce allocations
type SpanPool struct {
	pool           sync.Pool
	config         PerformanceConfig
	spansCreated   int64
	spansReused    int64
	spansDestroyed int64
}

// NewSpanPool creates a new span pool
func NewSpanPool(config PerformanceConfig) *SpanPool {
	sp := &SpanPool{
		config: config,
	}

	sp.pool = sync.Pool{
		New: func() interface{} {
			return sp.newPooledSpan()
		},
	}

	return sp
}

// PooledSpan represents a reusable span with pre-allocated capacity
type PooledSpan struct {
	// Core span data
	traceID       string
	spanID        string
	parentID      string
	operationName string
	startTime     time.Time
	endTime       time.Time
	duration      time.Duration

	// Pre-allocated slices to avoid allocations
	attributes []Attribute
	events     []Event
	links      []SpanLink
	tags       []Tag

	// Status and state
	status      StatusCode
	statusMsg   string
	isRecording bool
	isFinished  bool

	// Performance tracking
	pool    *SpanPool
	context context.Context

	// Zero-allocation optimization flags
	fastPath  bool
	zeroAlloc bool

	// Synchronization for concurrent access
	mu sync.RWMutex
}

// Attribute represents a span attribute with zero-allocation design
type Attribute struct {
	Key   string
	Value AttributeValue
}

// AttributeValue represents different attribute value types
type AttributeValue struct {
	Type      AttributeType
	StringVal string
	IntVal    int64
	FloatVal  float64
	BoolVal   bool
}

// AttributeType represents the type of attribute value
type AttributeType int

const (
	AttributeTypeString AttributeType = iota
	AttributeTypeInt
	AttributeTypeFloat
	AttributeTypeBool
)

// Event represents a span event with optimized memory layout
type Event struct {
	Name       string
	Timestamp  time.Time
	Attributes []Attribute
}

// Tag represents a simple key-value tag for fast access
type Tag struct {
	Key   string
	Value string
}

// Get retrieves a span from the pool
func (sp *SpanPool) Get() *PooledSpan {
	if !sp.config.EnableSpanPooling {
		return sp.newPooledSpan()
	}

	span := sp.pool.Get().(*PooledSpan)
	span.reset()
	atomic.AddInt64(&sp.spansReused, 1)
	return span
}

// Put returns a span to the pool
func (sp *SpanPool) Put(span *PooledSpan) {
	if !sp.config.EnableSpanPooling {
		atomic.AddInt64(&sp.spansDestroyed, 1)
		return
	}

	if span != nil {
		span.cleanup()
		sp.pool.Put(span)
	}
}

// GetMetrics returns span pool metrics
func (sp *SpanPool) GetMetrics() SpanPoolMetrics {
	return SpanPoolMetrics{
		SpansCreated:   atomic.LoadInt64(&sp.spansCreated),
		SpansReused:    atomic.LoadInt64(&sp.spansReused),
		SpansDestroyed: atomic.LoadInt64(&sp.spansDestroyed),
	}
}

// SpanPoolMetrics contains metrics about span pool performance
type SpanPoolMetrics struct {
	SpansCreated   int64 `json:"spans_created"`
	SpansReused    int64 `json:"spans_reused"`
	SpansDestroyed int64 `json:"spans_destroyed"`
}

// newPooledSpan creates a new pooled span with pre-allocated capacity
func (sp *SpanPool) newPooledSpan() *PooledSpan {
	atomic.AddInt64(&sp.spansCreated, 1)

	return &PooledSpan{
		attributes:  make([]Attribute, 0, sp.config.MaxAttributesPerSpan),
		events:      make([]Event, 0, sp.config.MaxEventsPerSpan),
		links:       make([]SpanLink, 0, 8), // reasonable default
		tags:        make([]Tag, 0, 32),     // reasonable default
		pool:        sp,
		fastPath:    sp.config.EnableFastPaths,
		zeroAlloc:   sp.config.EnableZeroAlloc,
		isRecording: true,
	}
}

// reset prepares the span for reuse
func (ps *PooledSpan) reset() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.traceID = ""
	ps.spanID = ""
	ps.parentID = ""
	ps.operationName = ""
	ps.startTime = time.Time{}
	ps.endTime = time.Time{}
	ps.duration = 0
	ps.status = StatusCodeUnset
	ps.statusMsg = ""
	ps.isRecording = true
	ps.isFinished = false
	ps.context = nil

	// Reset slices without reallocating
	ps.attributes = ps.attributes[:0]
	ps.events = ps.events[:0]
	ps.links = ps.links[:0]
	ps.tags = ps.tags[:0]
}

// cleanup prepares the span for return to pool
func (ps *PooledSpan) cleanup() {
	// Clear references to help GC
	ps.context = nil

	// Clear string references
	ps.traceID = ""
	ps.spanID = ""
	ps.parentID = ""
	ps.operationName = ""
	ps.statusMsg = ""

	// Clear slices content
	for i := range ps.attributes {
		ps.attributes[i] = Attribute{}
	}
	for i := range ps.events {
		ps.events[i] = Event{}
	}
	for i := range ps.links {
		ps.links[i] = SpanLink{}
	}
	for i := range ps.tags {
		ps.tags[i] = Tag{}
	}
}

// SetOperationName sets the operation name using fast path if enabled
func (ps *PooledSpan) SetOperationName(name string) {
	if ps.fastPath && ps.isRecording {
		ps.operationName = name
		return
	}

	// Fallback to regular path
	ps.operationName = name
}

// SetAttributeFast sets an attribute using zero-allocation fast path
func (ps *PooledSpan) SetAttributeFast(key string, value interface{}) {
	if !ps.zeroAlloc || !ps.isRecording {
		ps.SetAttributeRegular(key, value)
		return
	}

	ps.mu.Lock()

	// Check capacity
	if len(ps.attributes) >= cap(ps.attributes) {
		ps.mu.Unlock() // Unlock before calling regular method
		ps.SetAttributeRegular(key, value)
		return
	}

	// Fast path: add directly to pre-allocated slice
	attr := Attribute{Key: key}

	switch v := value.(type) {
	case string:
		attr.Value = AttributeValue{Type: AttributeTypeString, StringVal: v}
	case int:
		attr.Value = AttributeValue{Type: AttributeTypeInt, IntVal: int64(v)}
	case int64:
		attr.Value = AttributeValue{Type: AttributeTypeInt, IntVal: v}
	case float64:
		attr.Value = AttributeValue{Type: AttributeTypeFloat, FloatVal: v}
	case bool:
		attr.Value = AttributeValue{Type: AttributeTypeBool, BoolVal: v}
	default:
		// Fallback for complex types
		ps.mu.Unlock() // Unlock before calling regular method
		ps.SetAttributeRegular(key, value)
		return
	}

	ps.attributes = append(ps.attributes, attr)
	ps.mu.Unlock() // Unlock at the end of normal path
}

// SetAttributeRegular sets an attribute using regular allocation path
func (ps *PooledSpan) SetAttributeRegular(key string, value interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.isRecording {
		return
	}

	// Find existing attribute or append new one
	for i := range ps.attributes {
		if ps.attributes[i].Key == key {
			ps.updateAttributeValue(&ps.attributes[i], value)
			return
		}
	}

	// Add new attribute (check capacity or use default limit)
	maxAttrs := 128 // default limit
	if ps.pool != nil {
		maxAttrs = ps.pool.config.MaxAttributesPerSpan
	}

	if len(ps.attributes) < maxAttrs {
		attr := Attribute{Key: key}
		ps.updateAttributeValue(&attr, value)
		ps.attributes = append(ps.attributes, attr)
	}
}

// updateAttributeValue updates an attribute value
func (ps *PooledSpan) updateAttributeValue(attr *Attribute, value interface{}) {
	switch v := value.(type) {
	case string:
		attr.Value = AttributeValue{Type: AttributeTypeString, StringVal: v}
	case int:
		attr.Value = AttributeValue{Type: AttributeTypeInt, IntVal: int64(v)}
	case int64:
		attr.Value = AttributeValue{Type: AttributeTypeInt, IntVal: v}
	case float64:
		attr.Value = AttributeValue{Type: AttributeTypeFloat, FloatVal: v}
	case bool:
		attr.Value = AttributeValue{Type: AttributeTypeBool, BoolVal: v}
	default:
		// Convert unknown types to string
		attr.Value = AttributeValue{Type: AttributeTypeString, StringVal: stringValueOf(value)}
	}
}

// AddEventFast adds an event using fast path
func (ps *PooledSpan) AddEventFast(name string, attrs map[string]interface{}) {
	ps.mu.Lock()

	if !ps.isRecording {
		ps.mu.Unlock()
		return
	}

	if ps.fastPath && len(ps.events) < cap(ps.events) {
		// Fast path: pre-allocated event
		event := Event{
			Name:      name,
			Timestamp: time.Now(),
		}

		// Fast path for attributes if small number
		if len(attrs) <= 4 && ps.zeroAlloc {
			event.Attributes = make([]Attribute, 0, len(attrs))
			for k, v := range attrs {
				attr := Attribute{Key: k}
				ps.updateAttributeValue(&attr, v)
				event.Attributes = append(event.Attributes, attr)
			}
		} else {
			// Regular path for complex attributes
			event.Attributes = ps.convertMapToAttributes(attrs)
		}

		ps.events = append(ps.events, event)
		ps.mu.Unlock()
		return
	}

	// Fallback to regular path
	ps.mu.Unlock() // Unlock before calling regular method
	ps.AddEventRegular(name, attrs)
}

// AddEventRegular adds an event using regular allocation path
func (ps *PooledSpan) AddEventRegular(name string, attrs map[string]interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.isRecording {
		return
	}

	// Check event capacity (use default limit if pool is nil)
	maxEvents := 64 // default limit
	if ps.pool != nil {
		maxEvents = ps.pool.config.MaxEventsPerSpan
	}

	if len(ps.events) >= maxEvents {
		return
	}

	event := Event{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: ps.convertMapToAttributes(attrs),
	}

	ps.events = append(ps.events, event)
}

// convertMapToAttributes converts a map to attributes slice
func (ps *PooledSpan) convertMapToAttributes(attrs map[string]interface{}) []Attribute {
	if len(attrs) == 0 {
		return nil
	}

	result := make([]Attribute, 0, len(attrs))
	for k, v := range attrs {
		attr := Attribute{Key: k}
		ps.updateAttributeValue(&attr, v)
		result = append(result, attr)
	}

	return result
}

// Start begins span timing
func (ps *PooledSpan) Start() {
	ps.startTime = time.Now()
	ps.isRecording = true
}

// End finishes the span
func (ps *PooledSpan) End() {
	if ps.isFinished {
		return
	}

	ps.endTime = time.Now()
	ps.duration = ps.endTime.Sub(ps.startTime)
	ps.isRecording = false
	ps.isFinished = true

	// Return to pool
	if ps.pool != nil {
		ps.pool.Put(ps)
	}
}

// SetStatus sets the span status
func (ps *PooledSpan) SetStatus(code StatusCode, message string) {
	if ps.isRecording {
		ps.status = code
		ps.statusMsg = message
	}
}

// IsRecording returns true if the span is recording
func (ps *PooledSpan) IsRecording() bool {
	return ps.isRecording
}

// GetDuration returns the span duration
func (ps *PooledSpan) GetDuration() time.Duration {
	if ps.isFinished {
		return ps.duration
	}
	return time.Since(ps.startTime)
}

// GetAttributes returns span attributes (zero-copy if possible)
func (ps *PooledSpan) GetAttributes() []Attribute {
	if ps.zeroAlloc {
		// Return slice without copying
		return ps.attributes
	}

	// Copy for safety
	result := make([]Attribute, len(ps.attributes))
	copy(result, ps.attributes)
	return result
}

// GetEvents returns span events (zero-copy if possible)
func (ps *PooledSpan) GetEvents() []Event {
	if ps.zeroAlloc {
		return ps.events
	}

	result := make([]Event, len(ps.events))
	copy(result, ps.events)
	return result
}

// FastTracer implements optimized tracer with performance improvements
type FastTracer struct {
	spanPool    *SpanPool
	config      PerformanceConfig
	activeSpans sync.Map
	metrics     *FastTracerMetrics
	serializer  *FastSerializer
}

// FastTracerMetrics contains performance metrics
type FastTracerMetrics struct {
	SpansCreated     int64         `json:"spans_created"`
	SpansPooled      int64         `json:"spans_pooled"`
	ZeroAllocHits    int64         `json:"zero_alloc_hits"`
	FastPathHits     int64         `json:"fast_path_hits"`
	AvgSpanDuration  time.Duration `json:"avg_span_duration"`
	MemoryAllocated  int64         `json:"memory_allocated"`
	SerializationOps int64         `json:"serialization_ops"`
}

// NewFastTracer creates a new performance-optimized tracer
func NewFastTracer(config PerformanceConfig) *FastTracer {
	return &FastTracer{
		spanPool:   NewSpanPool(config),
		config:     config,
		metrics:    &FastTracerMetrics{},
		serializer: NewFastSerializer(config),
	}
}

// StartSpanFast creates a new span using optimized paths
func (ft *FastTracer) StartSpanFast(ctx context.Context, operationName string) (context.Context, *PooledSpan) {
	span := ft.spanPool.Get()
	span.SetOperationName(operationName)
	span.Start()

	atomic.AddInt64(&ft.metrics.SpansCreated, 1)

	// Store span in context using fast path if enabled
	if ft.config.EnableFastPaths {
		spanKey := ft.generateSpanKey(span)
		ft.activeSpans.Store(spanKey, span)
		ctx = context.WithValue(ctx, spanContextKey, spanKey)
		atomic.AddInt64(&ft.metrics.FastPathHits, 1)
	}

	return ctx, span
}

// spanContextKey is the key for storing span context
type spanContextKeyType struct{}

var spanContextKey = spanContextKeyType{}

// generateSpanKey generates a unique key for span context storage
func (ft *FastTracer) generateSpanKey(span *PooledSpan) string {
	// Use memory address as unique key for performance
	return unsafeStringFromPointer(unsafe.Pointer(span))
}

// GetMetrics returns tracer performance metrics
func (ft *FastTracer) GetMetrics() FastTracerMetrics {
	poolMetrics := ft.spanPool.GetMetrics()

	return FastTracerMetrics{
		SpansCreated:     atomic.LoadInt64(&ft.metrics.SpansCreated),
		SpansPooled:      poolMetrics.SpansReused,
		ZeroAllocHits:    atomic.LoadInt64(&ft.metrics.ZeroAllocHits),
		FastPathHits:     atomic.LoadInt64(&ft.metrics.FastPathHits),
		SerializationOps: atomic.LoadInt64(&ft.metrics.SerializationOps),
	}
}

// FastSerializer provides optimized serialization for span data
type FastSerializer struct {
	config     PerformanceConfig
	bufferPool sync.Pool
}

// NewFastSerializer creates a new fast serializer
func NewFastSerializer(config PerformanceConfig) *FastSerializer {
	fs := &FastSerializer{
		config: config,
	}

	fs.bufferPool = sync.Pool{
		New: func() interface{} {
			// Pre-allocate buffer with reasonable size
			buf := make([]byte, 0, 4096)
			return &buf
		},
	}

	return fs
}

// SerializeSpanFast serializes a span using optimized methods
func (fs *FastSerializer) SerializeSpanFast(span *PooledSpan) []byte {
	if span == nil {
		return nil
	}

	buf := fs.bufferPool.Get().(*[]byte)
	defer fs.bufferPool.Put(buf)

	// Reset buffer
	*buf = (*buf)[:0]

	// Use optimized serialization format
	*buf = fs.appendString(*buf, span.operationName)
	*buf = fs.appendInt64(*buf, span.startTime.UnixNano())
	*buf = fs.appendInt64(*buf, span.endTime.UnixNano())
	*buf = fs.appendInt(*buf, int(span.status))

	// Serialize attributes using fast path
	*buf = fs.appendInt(*buf, len(span.attributes))
	for _, attr := range span.attributes {
		*buf = fs.serializeAttribute(*buf, attr)
	}

	// Serialize events
	*buf = fs.appendInt(*buf, len(span.events))
	for _, event := range span.events {
		*buf = fs.serializeEvent(*buf, event)
	}

	// Return copy of buffer
	result := make([]byte, len(*buf))
	copy(result, *buf)

	return result
}

// appendString appends a string to buffer
func (fs *FastSerializer) appendString(buf []byte, s string) []byte {
	buf = fs.appendInt(buf, len(s))
	return append(buf, s...)
}

// appendInt appends an integer to buffer
func (fs *FastSerializer) appendInt(buf []byte, i int) []byte {
	return fs.appendInt64(buf, int64(i))
}

// appendInt64 appends an int64 to buffer
func (fs *FastSerializer) appendInt64(buf []byte, i int64) []byte {
	// Simple big-endian encoding
	return append(buf,
		byte(i>>56), byte(i>>48), byte(i>>40), byte(i>>32),
		byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

// serializeAttribute serializes an attribute
func (fs *FastSerializer) serializeAttribute(buf []byte, attr Attribute) []byte {
	buf = fs.appendString(buf, attr.Key)
	buf = fs.appendInt(buf, int(attr.Value.Type))

	switch attr.Value.Type {
	case AttributeTypeString:
		buf = fs.appendString(buf, attr.Value.StringVal)
	case AttributeTypeInt:
		buf = fs.appendInt64(buf, attr.Value.IntVal)
	case AttributeTypeFloat:
		buf = fs.appendInt64(buf, int64(attr.Value.FloatVal))
	case AttributeTypeBool:
		if attr.Value.BoolVal {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}
	}

	return buf
}

// serializeEvent serializes an event
func (fs *FastSerializer) serializeEvent(buf []byte, event Event) []byte {
	buf = fs.appendString(buf, event.Name)
	buf = fs.appendInt64(buf, event.Timestamp.UnixNano())
	buf = fs.appendInt(buf, len(event.Attributes))

	for _, attr := range event.Attributes {
		buf = fs.serializeAttribute(buf, attr)
	}

	return buf
}

// Helper functions for zero-allocation optimizations

// stringValueOf converts interface{} to string without reflection
func stringValueOf(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return unsafeStringFromBytes(v)
	default:
		// This will allocate, but it's a fallback
		return ""
	}
}

// unsafeStringFromBytes converts []byte to string without allocation
func unsafeStringFromBytes(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// unsafeStringFromPointer converts pointer to string representation
func unsafeStringFromPointer(ptr unsafe.Pointer) string {
	return string(rune(uintptr(ptr)))
}
