package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo de Sistema de Hooks ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Printf("💡 Exemplo de hooks seria executado com banco real: %v", err)
		demonstrateHooksConceptually()
		return
	}
	defer conn.Close(ctx)

	// 2. Obter gerenciador de hooks
	fmt.Println("2. Configurando sistema de hooks...")
	hookManager := conn.GetHookManager()
	if hookManager == nil {
		log.Printf("❌ Hook manager não disponível")
		demonstrateHooksConceptually()
		return
	}

	// 3. Exemplo: Hooks básicos
	fmt.Println("\n3. Exemplo: Hooks básicos...")
	if err := demonstrateBasicHooks(ctx, conn, hookManager); err != nil {
		log.Printf("Erro no exemplo básico: %v", err)
	}

	// 4. Exemplo: Hooks de performance
	fmt.Println("\n4. Exemplo: Hooks de performance...")
	if err := demonstratePerformanceHooks(ctx, conn, hookManager); err != nil {
		log.Printf("Erro no exemplo de performance: %v", err)
	}

	// 5. Exemplo: Hooks de auditoria
	fmt.Println("\n5. Exemplo: Hooks de auditoria...")
	if err := demonstrateAuditHooks(ctx, conn, hookManager); err != nil {
		log.Printf("Erro no exemplo de auditoria: %v", err)
	}

	// 6. Exemplo: Hooks de tratamento de erros
	fmt.Println("\n6. Exemplo: Hooks de tratamento de erros...")
	if err := demonstrateErrorHandlingHooks(ctx, conn, hookManager); err != nil {
		log.Printf("Erro no exemplo de tratamento de erros: %v", err)
	}

	// 7. Exemplo: Hooks customizados
	fmt.Println("\n7. Exemplo: Hooks customizados...")
	if err := demonstrateCustomHooks(ctx, conn, hookManager); err != nil {
		log.Printf("Erro no exemplo de hooks customizados: %v", err)
	}

	fmt.Println("\n=== Exemplo de Sistema de Hooks - CONCLUÍDO ===")
}

func demonstrateHooksConceptually() {
	fmt.Println("\n🎯 Demonstração Conceitual do Sistema de Hooks")
	fmt.Println("=============================================")

	fmt.Println("\n💡 Conceitos fundamentais:")
	fmt.Println("  - Hooks são funções executadas em pontos específicos do ciclo de vida")
	fmt.Println("  - Permitem interceptar e modificar comportamentos sem alterar o código core")
	fmt.Println("  - Suportam múltiplos hooks por evento")
	fmt.Println("  - Podem interromper a execução ou modificar dados")

	fmt.Println("\n🔄 Tipos de hooks disponíveis:")
	fmt.Println("  - BeforeQuery / AfterQuery: Interceptam operações de consulta")
	fmt.Println("  - BeforeExec / AfterExec: Interceptam operações de modificação")
	fmt.Println("  - BeforeTransaction / AfterTransaction: Interceptam transações")
	fmt.Println("  - BeforeConnection / AfterConnection: Interceptam conexões")
	fmt.Println("  - OnError: Interceptam erros para tratamento personalizado")

	fmt.Println("\n🛠️ Casos de uso comuns:")
	fmt.Println("  - Logging e auditoria de queries")
	fmt.Println("  - Monitoramento de performance")
	fmt.Println("  - Validação de segurança")
	fmt.Println("  - Tratamento customizado de erros")
	fmt.Println("  - Métricas e observabilidade")
	fmt.Println("  - Cache de resultados")

	fmt.Println("\n⚡ Vantagens:")
	fmt.Println("  - 🔍 Observabilidade completa")
	fmt.Println("  - 🛡️ Segurança avançada")
	fmt.Println("  - 📊 Métricas detalhadas")
	fmt.Println("  - 🎯 Customização sem modificar código core")
}

func demonstrateBasicHooks(ctx context.Context, conn postgres.IConn, hookManager postgres.IHookManager) error {
	fmt.Println("=== Hooks Básicos ===")

	// Hook para logging de queries
	loggingHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		fmt.Printf("   🔍 [LOG] Executando %s: %s\n", ctx.Operation, truncateQuery(ctx.Query, 50))
		return &postgres.HookResult{Continue: true}
	}

	// Hook para timing
	timingHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if ctx.Duration > 0 {
			fmt.Printf("   ⏱️  [TIMING] %s levou %v\n", ctx.Operation, ctx.Duration)
		}
		return &postgres.HookResult{Continue: true}
	}

	// Registrar hooks
	fmt.Println("   Registrando hooks básicos...")
	if err := hookManager.RegisterHook(postgres.BeforeQueryHook, loggingHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de logging: %w", err)
	}

	if err := hookManager.RegisterHook(postgres.AfterQueryHook, timingHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de timing: %w", err)
	}

	fmt.Println("   ✅ Hooks registrados com sucesso")

	// Testar hooks com algumas queries
	fmt.Println("   Testando hooks com queries...")

	queries := []string{
		"SELECT 1 as test",
		"SELECT version()",
		"SELECT current_timestamp",
		"SELECT COUNT(*) FROM information_schema.tables",
	}

	for i, query := range queries {
		fmt.Printf("   Executando query %d...\n", i+1)
		_, err := conn.Query(ctx, query)
		if err != nil {
			fmt.Printf("   ❌ Erro na query %d: %v\n", i+1, err)
		} else {
			fmt.Printf("   ✅ Query %d executada com sucesso\n", i+1)
		}
		time.Sleep(50 * time.Millisecond) // Pequeno delay para demonstrar timing
	}

	return nil
}

func demonstratePerformanceHooks(ctx context.Context, conn postgres.IConn, hookManager postgres.IHookManager) error {
	fmt.Println("=== Hooks de Performance ===")

	// Estrutura para coletar métricas
	metrics := &PerformanceMetrics{
		QueryCount:    0,
		TotalTime:     0,
		SlowQueries:   0,
		AverageTime:   0,
		SlowThreshold: 100 * time.Millisecond,
	}

	// Hook para coleta de métricas
	metricsHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if ctx.Duration > 0 {
			metrics.QueryCount++
			metrics.TotalTime += ctx.Duration
			metrics.AverageTime = metrics.TotalTime / time.Duration(metrics.QueryCount)

			if ctx.Duration > metrics.SlowThreshold {
				metrics.SlowQueries++
				fmt.Printf("   🐌 [SLOW] Query lenta detectada: %v (threshold: %v)\n",
					ctx.Duration, metrics.SlowThreshold)
			}
		}
		return &postgres.HookResult{Continue: true}
	}

	// Hook para detecção de queries suspeitas
	suspiciousHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if containsSuspiciousPattern(ctx.Query) {
			fmt.Printf("   ⚠️  [SUSPICIOUS] Query suspeita detectada: %s\n",
				truncateQuery(ctx.Query, 100))
		}
		return &postgres.HookResult{Continue: true}
	}

	// Registrar hooks de performance
	fmt.Println("   Registrando hooks de performance...")
	if err := hookManager.RegisterHook(postgres.AfterQueryHook, metricsHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de métricas: %w", err)
	}

	if err := hookManager.RegisterHook(postgres.BeforeQueryHook, suspiciousHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de suspeitas: %w", err)
	}

	// Testar com queries de diferentes complexidades
	fmt.Println("   Testando hooks de performance...")

	performanceQueries := []struct {
		name  string
		query string
	}{
		{"Query rápida", "SELECT 1"},
		{"Query média", "SELECT COUNT(*) FROM information_schema.tables"},
		{"Query lenta simulada", "SELECT pg_sleep(0.15), 'slow query'"},
		{"Query suspeita", "SELECT * FROM information_schema.tables WHERE table_name LIKE '%'"},
	}

	for _, pq := range performanceQueries {
		fmt.Printf("   Executando %s...\n", pq.name)
		startTime := time.Now()
		_, err := conn.Query(ctx, pq.query)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("   ❌ Erro em %s: %v\n", pq.name, err)
		} else {
			fmt.Printf("   ✅ %s executada em %v\n", pq.name, duration)
		}
	}

	// Mostrar métricas coletadas
	fmt.Println("\n   📊 Métricas de Performance:")
	fmt.Printf("   - Total de queries: %d\n", metrics.QueryCount)
	fmt.Printf("   - Tempo total: %v\n", metrics.TotalTime)
	fmt.Printf("   - Tempo médio: %v\n", metrics.AverageTime)
	fmt.Printf("   - Queries lentas: %d\n", metrics.SlowQueries)
	fmt.Printf("   - Threshold: %v\n", metrics.SlowThreshold)

	return nil
}

func demonstrateAuditHooks(ctx context.Context, conn postgres.IConn, hookManager postgres.IHookManager) error {
	fmt.Println("=== Hooks de Auditoria ===")

	// Estrutura para auditoria
	auditLog := &AuditLog{
		Entries: make([]AuditEntry, 0),
	}

	// Hook de auditoria
	auditHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		entry := AuditEntry{
			Timestamp: time.Now(),
			Operation: ctx.Operation,
			Query:     ctx.Query,
			Args:      ctx.Args,
			Duration:  ctx.Duration,
			Success:   ctx.Error == nil,
		}

		if ctx.Error != nil {
			entry.Error = ctx.Error.Error()
		}

		auditLog.Entries = append(auditLog.Entries, entry)

		// Log crítico para operações sensíveis
		if containsSensitiveOperation(ctx.Query) {
			fmt.Printf("   🔐 [AUDIT] Operação sensível: %s\n", truncateQuery(ctx.Query, 80))
		}

		return &postgres.HookResult{Continue: true}
	}

	// Hook de validação de segurança
	securityHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if containsSecurityRisk(ctx.Query) {
			fmt.Printf("   ⚠️  [SECURITY] Possível risco de segurança detectado\n")
			// Em produção, você poderia bloquear a query aqui
			// return &postgres.HookResult{Continue: false, Error: errors.New("query blocked by security policy")}
		}
		return &postgres.HookResult{Continue: true}
	}

	// Registrar hooks de auditoria
	fmt.Println("   Registrando hooks de auditoria...")
	if err := hookManager.RegisterHook(postgres.AfterQueryHook, auditHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de auditoria: %w", err)
	}

	if err := hookManager.RegisterHook(postgres.BeforeQueryHook, securityHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de segurança: %w", err)
	}

	// Testar com queries que simulam operações sensíveis
	fmt.Println("   Testando hooks de auditoria...")

	auditQueries := []struct {
		name  string
		query string
	}{
		{"Operação normal", "SELECT current_user"},
		{"Operação sensível", "SELECT * FROM information_schema.tables"},
		{"Operação com risco", "SELECT * FROM pg_stat_activity"},
		{"Operação de sistema", "SELECT version()"},
	}

	for _, aq := range auditQueries {
		fmt.Printf("   Executando %s...\n", aq.name)
		_, err := conn.Query(ctx, aq.query)
		if err != nil {
			fmt.Printf("   ❌ Erro em %s: %v\n", aq.name, err)
		} else {
			fmt.Printf("   ✅ %s executada com sucesso\n", aq.name)
		}
	}

	// Mostrar log de auditoria
	fmt.Println("\n   📋 Log de Auditoria:")
	for i, entry := range auditLog.Entries {
		status := "✅ SUCESSO"
		if !entry.Success {
			status = "❌ ERRO"
		}
		fmt.Printf("   %d. [%s] %s %s - %v\n",
			i+1, entry.Timestamp.Format("15:04:05"),
			status, entry.Operation, entry.Duration)
	}

	return nil
}

func demonstrateErrorHandlingHooks(ctx context.Context, conn postgres.IConn, hookManager postgres.IHookManager) error {
	fmt.Println("=== Hooks de Tratamento de Erros ===")

	// Contadores de erro
	errorStats := &ErrorStats{
		TotalErrors:      0,
		ConnectionErrors: 0,
		QueryErrors:      0,
		TimeoutErrors:    0,
		OtherErrors:      0,
	}

	// Hook para tratamento de erros
	errorHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if ctx.Error != nil {
			errorStats.TotalErrors++

			// Categorizar o erro
			switch categorizeError(ctx.Error) {
			case ErrorCategoryConnection:
				errorStats.ConnectionErrors++
				fmt.Printf("   🔌 [ERROR] Erro de conexão: %v\n", ctx.Error)
			case ErrorCategoryQuery:
				errorStats.QueryErrors++
				fmt.Printf("   📝 [ERROR] Erro de query: %v\n", ctx.Error)
			case ErrorCategoryTimeout:
				errorStats.TimeoutErrors++
				fmt.Printf("   ⏰ [ERROR] Erro de timeout: %v\n", ctx.Error)
			default:
				errorStats.OtherErrors++
				fmt.Printf("   ❓ [ERROR] Erro desconhecido: %v\n", ctx.Error)
			}
		}
		return &postgres.HookResult{Continue: true}
	}

	// Hook para retry automático (conceitual)
	retryHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if ctx.Error != nil && isRetryableError(ctx.Error) {
			fmt.Printf("   🔄 [RETRY] Erro retriável detectado, seria tentado novamente\n")
			// Em implementação real, você implementaria a lógica de retry aqui
		}
		return &postgres.HookResult{Continue: true}
	}

	// Registrar hooks de erro
	fmt.Println("   Registrando hooks de tratamento de erros...")
	if err := hookManager.RegisterHook(postgres.OnErrorHook, errorHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de erro: %w", err)
	}

	if err := hookManager.RegisterHook(postgres.OnErrorHook, retryHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de retry: %w", err)
	}

	// Testar com queries que geram erros
	fmt.Println("   Testando hooks com queries que geram erros...")

	errorQueries := []struct {
		name  string
		query string
	}{
		{"Query válida", "SELECT 1"},
		{"Query inválida", "SELECT FROM invalid_table"},
		{"Syntax error", "SELEC 1"},
		{"Tabela inexistente", "SELECT * FROM tabela_que_nao_existe"},
	}

	for _, eq := range errorQueries {
		fmt.Printf("   Executando %s...\n", eq.name)
		_, err := conn.Query(ctx, eq.query)
		if err != nil {
			fmt.Printf("   ❌ Erro esperado em %s: %v\n", eq.name, err)
		} else {
			fmt.Printf("   ✅ %s executada com sucesso\n", eq.name)
		}
	}

	// Mostrar estatísticas de erro
	fmt.Println("\n   📊 Estatísticas de Erros:")
	fmt.Printf("   - Total de erros: %d\n", errorStats.TotalErrors)
	fmt.Printf("   - Erros de conexão: %d\n", errorStats.ConnectionErrors)
	fmt.Printf("   - Erros de query: %d\n", errorStats.QueryErrors)
	fmt.Printf("   - Erros de timeout: %d\n", errorStats.TimeoutErrors)
	fmt.Printf("   - Outros erros: %d\n", errorStats.OtherErrors)

	return nil
}

func demonstrateCustomHooks(ctx context.Context, conn postgres.IConn, hookManager postgres.IHookManager) error {
	fmt.Println("=== Hooks Customizados ===")

	// Hook customizado para cache (conceitual)
	cacheHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if isSelectQuery(ctx.Query) {
			fmt.Printf("   💾 [CACHE] Query SELECT detectada, verificando cache...\n")
			// Em implementação real, você verificaria um cache aqui
		}
		return &postgres.HookResult{Continue: true}
	}

	// Hook customizado para rate limiting (conceitual)
	rateLimitHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		// Simulação de rate limiting
		fmt.Printf("   ⏳ [RATE_LIMIT] Verificando rate limit...\n")
		// Em implementação real, você implementaria lógica de rate limiting aqui
		return &postgres.HookResult{Continue: true}
	}

	// Hook customizado para métricas específicas
	customMetricsHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
		if ctx.Duration > 0 {
			fmt.Printf("   📊 [METRICS] Coletando métricas customizadas: %v\n", ctx.Duration)
		}
		return &postgres.HookResult{Continue: true}
	}

	// Registrar hooks customizados
	fmt.Println("   Registrando hooks customizados...")

	// Usar RegisterCustomHook para hooks com nomes específicos
	customHookType := postgres.CustomHookBase + 1
	if err := hookManager.RegisterCustomHook(customHookType, "cache_hook", cacheHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de cache: %w", err)
	}

	customHookType = postgres.CustomHookBase + 2
	if err := hookManager.RegisterCustomHook(customHookType, "rate_limit_hook", rateLimitHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de rate limit: %w", err)
	}

	if err := hookManager.RegisterHook(postgres.AfterQueryHook, customMetricsHook); err != nil {
		return fmt.Errorf("erro ao registrar hook de métricas customizadas: %w", err)
	}

	// Testar hooks customizados
	fmt.Println("   Testando hooks customizados...")

	customQueries := []string{
		"SELECT current_database()",
		"SELECT current_user",
		"SELECT now()",
	}

	for i, query := range customQueries {
		fmt.Printf("   Executando query customizada %d...\n", i+1)
		_, err := conn.Query(ctx, query)
		if err != nil {
			fmt.Printf("   ❌ Erro na query customizada %d: %v\n", i+1, err)
		} else {
			fmt.Printf("   ✅ Query customizada %d executada com sucesso\n", i+1)
		}
	}

	// Mostrar hooks registrados
	fmt.Println("\n   📋 Hooks Registrados:")
	registeredHooks := hookManager.ListHooks()
	for hookType, hooks := range registeredHooks {
		fmt.Printf("   - Tipo %d: %d hooks\n", hookType, len(hooks))
	}

	return nil
}

// Estruturas de suporte

type PerformanceMetrics struct {
	QueryCount    int64
	TotalTime     time.Duration
	SlowQueries   int64
	AverageTime   time.Duration
	SlowThreshold time.Duration
}

type AuditEntry struct {
	Timestamp time.Time
	Operation string
	Query     string
	Args      []interface{}
	Duration  time.Duration
	Success   bool
	Error     string
}

type AuditLog struct {
	Entries []AuditEntry
}

type ErrorStats struct {
	TotalErrors      int64
	ConnectionErrors int64
	QueryErrors      int64
	TimeoutErrors    int64
	OtherErrors      int64
}

type ErrorCategory int

const (
	ErrorCategoryConnection ErrorCategory = iota
	ErrorCategoryQuery
	ErrorCategoryTimeout
	ErrorCategoryOther
)

// Funções utilitárias

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

func containsSuspiciousPattern(query string) bool {
	// Simular detecção de padrões suspeitos
	patterns := []string{"LIKE '%'", "SELECT *", "pg_stat_activity"}
	for _, pattern := range patterns {
		if len(query) > len(pattern) && query[:len(pattern)] == pattern ||
			len(query) > len(pattern) && query[len(query)-len(pattern):] == pattern {
			return true
		}
	}
	return false
}

func containsSensitiveOperation(query string) bool {
	// Simular detecção de operações sensíveis
	return len(query) > 10 && (query[:6] == "SELECT" || query[:6] == "INSERT" || query[:6] == "UPDATE" || query[:6] == "DELETE")
}

func containsSecurityRisk(query string) bool {
	// Simular detecção de riscos de segurança
	risks := []string{"pg_stat_activity", "pg_tables", "DROP", "ALTER"}
	for _, risk := range risks {
		if len(query) >= len(risk) {
			for i := 0; i <= len(query)-len(risk); i++ {
				if query[i:i+len(risk)] == risk {
					return true
				}
			}
		}
	}
	return false
}

func categorizeError(err error) ErrorCategory {
	if err == nil {
		return ErrorCategoryOther
	}

	errorStr := err.Error()
	if len(errorStr) > 10 {
		switch {
		case errorStr[:10] == "connection":
			return ErrorCategoryConnection
		case errorStr[:5] == "query":
			return ErrorCategoryQuery
		case errorStr[:7] == "timeout":
			return ErrorCategoryTimeout
		default:
			return ErrorCategoryOther
		}
	}
	return ErrorCategoryOther
}

func isRetryableError(err error) bool {
	// Simular detecção de erros retriáveis
	if err == nil {
		return false
	}

	errorStr := err.Error()
	retryablePatterns := []string{"timeout", "connection", "temporary"}
	for _, pattern := range retryablePatterns {
		if len(errorStr) >= len(pattern) {
			for i := 0; i <= len(errorStr)-len(pattern); i++ {
				if errorStr[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}

func isSelectQuery(query string) bool {
	return len(query) >= 6 && query[:6] == "SELECT"
}
