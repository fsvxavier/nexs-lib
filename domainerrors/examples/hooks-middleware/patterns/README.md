# Hooks and Middleware Pattern Examples

This directory contains specific examples of common patterns using hooks and middleware.

## Available Patterns

### 1. **enrichment_pattern.go**
Demonstrates how to enrich errors with contextual information using middleware chains.

**Run with:**
```bash
go run enrichment_pattern.go
```

**Key Features:**
- Request context enrichment
- Service information addition  
- Timing information injection
- Multi-layer metadata structure

### 2. **circuit_breaker_pattern.go** 
Shows implementation of circuit breaker pattern using middleware.

**Run with:**
```bash  
go run circuit_breaker_pattern.go
```

**Key Features:**
- Failure count tracking
- Circuit state management
- Automatic circuit opening/closing
- Service degradation handling

### 3. **security_audit_pattern.go**
Demonstrates security-focused error handling with hooks and middleware.

**Run with:**
```bash
go run security_audit_pattern.go  
```

**Key Features:**
- Security event detection
- PII sanitization
- Audit logging
- Alert generation

### 4. **metrics_collection_pattern.go**
Shows how to collect metrics and monitoring data from errors.

**Run with:**
```bash
go run metrics_collection_pattern.go
```

**Key Features:**
- Error frequency tracking
- Response time metrics
- Error type distribution
- Performance monitoring

## Usage Instructions

Each pattern example is self-contained and can be run independently:

```bash
cd patterns/
go run <pattern_name>.go
```

## Pattern Categories

### **Observability Patterns**
- Logging and structured output
- Metrics collection
- Tracing integration
- Monitoring hooks

### **Resilience Patterns** 
- Circuit breaker implementation
- Retry logic with backoff
- Timeout handling
- Graceful degradation

### **Security Patterns**
- Audit logging
- PII sanitization  
- Access control
- Threat detection

### **Integration Patterns**
- Service mesh integration
- External system communication
- Event bus publishing
- Notification systems

Each pattern demonstrates best practices and production-ready implementations that can be adapted for your specific use cases.
