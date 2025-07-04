package example

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// ExemploCriacaoErros demonstra a criação de diferentes tipos de erros
func ExemploCriacaoErros() {
	// Erro simples
	err1 := domainerrors.New("E001", "Erro simples de teste")
	fmt.Println(err1.Error()) // Saída: [E001] Erro simples de teste

	// Erro com causa
	erroBase := errors.New("erro original")
	err2 := domainerrors.NewWithError("E002", "Falha ao processar", erroBase)
	fmt.Println(err2.Error()) // Saída: [E002] Falha ao processar: erro original

	// Erro com stack trace
	fmt.Println("Stack trace:")
	fmt.Println(err2.FormatStackTrace())
}

// ExemploTiposErros demonstra a criação de erros de tipos específicos
func ExemploTiposErros() {
	// Erro de validação
	validationErr := domainerrors.NewValidationError("Falha na validação", nil)
	validationErr.WithField("email", "Email inválido")
	validationErr.WithField("senha", "Senha deve ter pelo menos 8 caracteres")
	fmt.Println(validationErr.Error()) // [VALIDATION_ERROR] Falha na validação
	fmt.Printf("Campos inválidos: %+v\n", validationErr.ValidatedFields)

	// Erro de recurso não encontrado
	notFoundErr := domainerrors.NewNotFoundError("Usuário não encontrado")
	notFoundErr.WithResource("user", "123")
	fmt.Println(notFoundErr.Error()) // [NOT_FOUND] Usuário não encontrado
	fmt.Printf("Recurso não encontrado: %s:%s\n", notFoundErr.ResourceType, notFoundErr.ResourceID)

	// Erro de regra de negócio
	businessErr := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Saldo insuficiente para a operação")
	fmt.Println(businessErr.Error()) // [INSUFFICIENT_FUNDS] Saldo insuficiente para a operação

	// Erro de banco de dados
	dbError := errors.New("constraint violation on table 'users'")
	databaseErr := domainerrors.NewDatabaseError("Falha ao inserir usuário", dbError)
	databaseErr.WithOperation("INSERT", "users")
	fmt.Println(databaseErr.Error()) // [INFRA_ERROR] Falha ao inserir usuário: constraint violation on table 'users'

	// Erro de serviço externo
	extErr := errors.New("timeout connecting to service")
	externalErr := domainerrors.NewExternalServiceError("payment-gateway", "Falha ao processar pagamento", extErr)
	externalErr.WithStatusCode(http.StatusGatewayTimeout)
	fmt.Println(externalErr.Error()) // [EXTERNAL_ERROR] Falha ao processar pagamento: timeout connecting to service
}

// ExemploWrapErrors demonstra o empilhamento de erros
func ExemploWrapErrors() {
	// Criando erro base
	baseErr := errors.New("conexão recusada")

	// Camada de infraestrutura
	infraErr := domainerrors.NewInfrastructureError("database", "Falha ao conectar", baseErr)

	// Camada de repositório
	repoErr := domainerrors.New("REPO_ERROR", "Erro ao acessar repositório").Wrap("Tentativa de consulta", infraErr)

	// Camada de serviço
	serviceErr := domainerrors.New("SERVICE_ERROR", "Falha no serviço").Wrap("Busca de usuário", repoErr)

	// Camada de controle
	finalErr := domainerrors.New("API_ERROR", "Erro na API").Wrap("Processamento da requisição", serviceErr)

	// Exibindo erro completo
	fmt.Println("Erro final:")
	fmt.Println(finalErr.Error())

	// Exibindo stack trace formatado
	fmt.Println("\nStack trace:")
	fmt.Println(finalErr.FormatStackTrace())
}

// ExemploUtilidades demonstra o uso das funções utilitárias
func ExemploUtilidades() {
	// Criando alguns erros para teste
	notFoundErr := domainerrors.NewNotFoundError("Produto não encontrado")
	validationErr := domainerrors.NewValidationError("Dados inválidos", nil)
	businessErr := domainerrors.NewBusinessError("LIMIT_EXCEEDED", "Limite excedido")

	// Verificando tipos de erro
	fmt.Printf("É erro not found? %v\n", domainerrors.IsNotFoundError(notFoundErr))
	fmt.Printf("É erro de validação? %v\n", domainerrors.IsValidationError(validationErr))
	fmt.Printf("É erro de negócio? %v\n", domainerrors.IsBusinessError(businessErr))

	// Obtendo código HTTP
	fmt.Printf("Status code para notFoundErr: %d\n", notFoundErr.StatusCode())
	fmt.Printf("Status code para validationErr: %d\n", validationErr.StatusCode())
	fmt.Printf("Status code para businessErr: %d\n", businessErr.StatusCode())

	// Trabalhando com stack de erros
	stack := domainerrors.NewErrorStack()
	stack.Push(notFoundErr)
	stack.Push(validationErr)
	stack.Push(businessErr)

	fmt.Println("\nStack de erros formatado:")
	fmt.Println(stack.Format())
}

// ExemploRegistroErros demonstra o uso do registro de códigos de erro
func ExemploRegistroErros() {
	// Criando um registro de códigos de erro
	registry := domainerrors.NewErrorCodeRegistry()

	// Registrando alguns códigos
	registry.Register("AUTH001", "Credenciais inválidas", http.StatusUnauthorized)
	registry.Register("AUTH002", "Token expirado", http.StatusUnauthorized)
	registry.Register("PROD001", "Produto não encontrado", http.StatusNotFound)

	// Usando o registro para criar erros
	baseErr := errors.New("usuário ou senha incorretos")
	err := registry.WrapWithCode("AUTH001", baseErr)

	fmt.Println(err.Error())

	// Verificando se um código existe
	code, exists := registry.Get("PROD001")
	if exists {
		fmt.Printf("Código PROD001: %s (Status %d)\n", code.Description, code.StatusCode)
	}
}

// ExemploRecuperacaoPanico demonstra o uso da função de recuperação de pânicos
func ExemploRecuperacaoPanico() {
	err := domainerrors.RecoverMiddleware(func() error {
		// Simulando um código que pode gerar pânico
		m := map[string]string(nil)
		return errors.New(m["chave-inexistente"]) // Isso vai causar pânico
	})

	fmt.Println("Erro recuperado do pânico:")
	fmt.Println(err.Error())

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		fmt.Printf("Tipo do erro: %s\n", domainErr.Type)
	}
}

// ExemploNovosErrosDominio demonstra o uso dos novos tipos de erro adicionados
func ExemploNovosErrosDominio() {
	// Erro de conflito
	conflictErr := domainerrors.NewConflictError("Email já está em uso")
	conflictErr.WithConflictingResource("user", "email duplicado")
	fmt.Println("Conflito:", conflictErr.Error())
	fmt.Printf("Status: %d\n", conflictErr.StatusCode())

	// Erro de limite de taxa
	rateLimitErr := domainerrors.NewRateLimitError("Muitas tentativas de login")
	rateLimitErr.WithRateLimit(10, 0, "2025-01-01T15:00:00Z", "60s")
	fmt.Println("Rate Limit:", rateLimitErr.Error())
	fmt.Printf("Limite: %d, Restante: %d, Retry após: %s\n",
		rateLimitErr.Limit, rateLimitErr.Remaining, rateLimitErr.RetryAfter)

	// Erro de circuit breaker
	circuitErr := domainerrors.NewCircuitBreakerError("payment-api", "Serviço indisponível")
	circuitErr.WithCircuitState("OPEN", 5)
	fmt.Println("Circuit Breaker:", circuitErr.Error())
	fmt.Printf("Estado: %s, Falhas: %d\n", circuitErr.State, circuitErr.Failures)

	// Erro de configuração
	configErr := domainerrors.NewConfigurationError("Configuração de banco inválida")
	configErr.WithConfigDetails("database.port", "invalid", "número entre 1-65535")
	fmt.Println("Configuração:", configErr.Error())

	// Erro de segurança
	securityErr := domainerrors.NewSecurityError("Tentativa de acesso suspeita")
	securityErr.WithSecurityContext("login_brute_force", "HIGH")
	securityErr.WithClientInfo("curl/7.68.0", "192.168.1.100")
	fmt.Println("Segurança:", securityErr.Error())

	// Erro de recurso esgotado
	resourceErr := domainerrors.NewResourceExhaustedError("memory", "Memória insuficiente")
	resourceErr.WithResourceLimits(2048, 2048, "MB")
	fmt.Println("Recurso esgotado:", resourceErr.Error())
	fmt.Printf("Uso: %d/%d %s\n", resourceErr.Current, resourceErr.Limit, resourceErr.Unit)

	// Erro de dependência
	depErr := errors.New("connection timeout")
	dependencyErr := domainerrors.NewDependencyError("elasticsearch", "Falha na busca", depErr)
	dependencyErr.WithDependencyInfo("search_engine", "7.10.0", "UNHEALTHY")
	fmt.Println("Dependência:", dependencyErr.Error())

	// Erro de serialização
	serErr := errors.New("invalid JSON structure")
	serializationErr := domainerrors.NewSerializationError("JSON", "Falha ao serializar dados", serErr)
	serializationErr.WithTypeInfo("user.birthdate", "time.Time", "string")
	fmt.Println("Serialização:", serializationErr.Error())

	// Erro de cache
	cacheBaseErr := errors.New("redis connection lost")
	cacheErr := domainerrors.NewCacheError("redis", "SET", "Falha ao armazenar no cache", cacheBaseErr)
	cacheErr.WithCacheDetails("session:abc123", "3600s")
	fmt.Println("Cache:", cacheErr.Error())

	// Erro de workflow
	workflowErr := domainerrors.NewWorkflowError("order-fulfillment", "inventory_check", "Estoque insuficiente")
	workflowErr.WithStateInfo("checking_inventory", "inventory_confirmed")
	fmt.Println("Workflow:", workflowErr.Error())

	// Erro de migração
	migBaseErr := errors.New("table already exists")
	migrationErr := domainerrors.NewMigrationError("v2.1.0", "create_audit_table", "Falha na migração", migBaseErr)
	migrationErr.WithMigrationDetails("up", 0)
	fmt.Println("Migração:", migrationErr.Error())
}

// ExemploVerificacaoTiposNovos demonstra como verificar os novos tipos de erro
func ExemploVerificacaoTiposNovos() {
	// Criando diferentes tipos de erro
	conflictErr := domainerrors.NewConflictError("Conflito de dados")
	rateLimitErr := domainerrors.NewRateLimitError("Limite excedido")
	securityErr := domainerrors.NewSecurityError("Acesso negado")

	erros := []error{conflictErr, rateLimitErr, securityErr}

	for i, err := range erros {
		fmt.Printf("Erro %d:\n", i+1)
		fmt.Printf("  É conflito? %v\n", domainerrors.IsConflictError(err))
		fmt.Printf("  É rate limit? %v\n", domainerrors.IsRateLimitError(err))
		fmt.Printf("  É segurança? %v\n", domainerrors.IsSecurityError(err))
		fmt.Printf("  Status HTTP: %d\n", domainerrors.GetStatusCode(err))
		fmt.Println()
	}
}

// ExemploTratamentoErrosComplexos demonstra o tratamento de erros complexos com múltiplas camadas
func ExemploTratamentoErrosComplexos() {
	// Simulando uma cadeia de erros complexa

	// 1. Erro de dependência (Redis)
	redisErr := errors.New("connection refused")
	_ = domainerrors.NewCacheError("redis", "GET", "Cache não disponível", redisErr)

	// 2. Erro de fallback para banco de dados
	dbErr := errors.New("connection pool exhausted")
	resourceErr := domainerrors.NewResourceExhaustedError("db_connections", "Pool de conexões esgotado")
	resourceErr.WithResourceLimits(100, 100, "connections")

	// 3. Erro de circuit breaker
	circuitErr := domainerrors.NewCircuitBreakerError("user-service", "Serviço de usuários indisponível")
	circuitErr.WithCircuitState("OPEN", 10)

	// 4. Erro de workflow
	workflowErr := domainerrors.NewWorkflowError("user-registration", "profile_creation", "Falha ao criar perfil")
	workflowErr.Wrap("Erro no circuit breaker", circuitErr)
	workflowErr.Wrap("Erro de recurso esgotado", resourceErr)
	workflowErr.Wrap("Erro de banco de dados", dbErr)

	// Exibindo a cadeia completa de erros
	fmt.Println("Erro final:")
	fmt.Println(workflowErr.Error())

	fmt.Println("\nStack trace formatado:")
	fmt.Println(workflowErr.FormatStackTrace())

	// Verificando tipos na cadeia
	fmt.Printf("É erro de workflow? %v\n", domainerrors.IsWorkflowError(workflowErr))
	fmt.Printf("É erro de circuit breaker? %v\n", domainerrors.IsCircuitBreakerError(workflowErr))
	fmt.Printf("É erro de cache? %v\n", domainerrors.IsCacheError(workflowErr))

	// Código de status final
	fmt.Printf("Status HTTP: %d\n", workflowErr.StatusCode())
}
