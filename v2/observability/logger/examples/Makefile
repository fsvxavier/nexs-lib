# Makefile para Logger v2 Examples
# Seguindo práticas de engenharia sênior para automação

.PHONY: all build test bench clean examples help deps check lint

# Variáveis
GO_VERSION := 1.21
EXAMPLES := basic structured context-aware async middleware providers microservices web-app
COVERAGE_TARGET := 98

# Cores para output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# Target padrão
all: check deps build test

# Ajuda
help:
	@echo "$(GREEN)Logger v2 Examples - Makefile$(NC)"
	@echo ""
	@echo "Targets disponíveis:"
	@echo "  $(YELLOW)all$(NC)        - Executa check, deps, build e test"
	@echo "  $(YELLOW)check$(NC)      - Verifica pré-requisitos"
	@echo "  $(YELLOW)deps$(NC)       - Baixa dependências"
	@echo "  $(YELLOW)build$(NC)      - Compila todos os exemplos"
	@echo "  $(YELLOW)test$(NC)       - Executa testes (quando disponíveis)"
	@echo "  $(YELLOW)bench$(NC)      - Executa benchmarks (quando disponíveis)"
	@echo "  $(YELLOW)examples$(NC)   - Executa todos os exemplos"
	@echo "  $(YELLOW)lint$(NC)       - Executa linting"
	@echo "  $(YELLOW)clean$(NC)      - Remove arquivos temporários"
	@echo "  $(YELLOW)coverage$(NC)   - Gera relatório de cobertura"
	@echo "  $(YELLOW)help$(NC)       - Mostra esta ajuda"

# Verifica pré-requisitos
check:
	@echo "$(GREEN)Verificando pré-requisitos...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Erro: Go não está instalado$(NC)"; exit 1; }
	@go version | grep -q "go1.2[1-9]" || go version | grep -q "go1.[3-9][0-9]" || { echo "$(RED)Erro: Go 1.21+ necessário$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Go está instalado e atualizado$(NC)"

# Baixa dependências
deps: check
	@echo "$(GREEN)Baixando dependências...$(NC)"
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			echo "Processando $$example..."; \
			cd $$example && go mod tidy && cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Dependências atualizadas$(NC)"

# Compila todos os exemplos
build: deps
	@echo "$(GREEN)Compilando exemplos...$(NC)"
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ] && [ -f "$$example/main.go" ]; then \
			echo "Compilando $$example..."; \
			cd $$example && go build -o ../build/$$example . && cd .. || { echo "$(RED)Erro compilando $$example$(NC)"; exit 1; }; \
		fi; \
	done
	@echo "$(GREEN)✓ Todos os exemplos compilados$(NC)"

# Executa testes (quando disponíveis)
test: build
	@echo "$(GREEN)Executando testes...$(NC)"
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			cd $$example; \
			if ls *_test.go >/dev/null 2>&1; then \
				echo "Testando $$example..."; \
				go test -v . || { echo "$(RED)Erro nos testes de $$example$(NC)"; exit 1; }; \
			else \
				echo "$(YELLOW)Sem testes para $$example$(NC)"; \
			fi; \
			cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Testes concluídos$(NC)"

# Executa benchmarks
bench: build
	@echo "$(GREEN)Executando benchmarks...$(NC)"
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			cd $$example; \
			if ls *_test.go >/dev/null 2>&1; then \
				echo "Benchmark $$example..."; \
				go test -bench=. -benchmem . || true; \
			fi; \
			cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Benchmarks concluídos$(NC)"

# Executa todos os exemplos
examples: build
	@echo "$(GREEN)Executando todos os exemplos...$(NC)"
	@./run_all.sh

# Linting
lint: check
	@echo "$(GREEN)Executando linting...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(YELLOW)golangci-lint não instalado, pulando...$(NC)"; exit 0; }
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			echo "Linting $$example..."; \
			cd $$example && golangci-lint run . && cd .. || { echo "$(RED)Erro de lint em $$example$(NC)"; exit 1; }; \
		fi; \
	done
	@echo "$(GREEN)✓ Linting concluído$(NC)"

# Relatório de cobertura
coverage: build
	@echo "$(GREEN)Gerando relatório de cobertura...$(NC)"
	@mkdir -p coverage
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			cd $$example; \
			if ls *_test.go >/dev/null 2>&1; then \
				echo "Cobertura $$example..."; \
				go test -coverprofile=../coverage/$$example.out . || true; \
				go tool cover -html=../coverage/$$example.out -o ../coverage/$$example.html || true; \
			fi; \
			cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Relatórios de cobertura gerados em coverage/$(NC)"

# Limpeza
clean:
	@echo "$(GREEN)Limpando arquivos temporários...$(NC)"
	@rm -rf build/
	@rm -rf coverage/
	@rm -f output_*.log
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ]; then \
			cd $$example && go clean && cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Limpeza concluída$(NC)"

# Setup inicial
setup: check
	@echo "$(GREEN)Configurando ambiente...$(NC)"
	@mkdir -p build
	@mkdir -p coverage
	@echo "$(GREEN)✓ Ambiente configurado$(NC)"

# Validação completa
validate: clean setup lint test bench examples
	@echo "$(GREEN)✓ Validação completa concluída$(NC)"

# Instalação de ferramentas de desenvolvimento
install-tools:
	@echo "$(GREEN)Instalando ferramentas de desenvolvimento...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/cover@latest
	@echo "$(GREEN)✓ Ferramentas instaladas$(NC)"

# Release build (otimizado)
release: clean
	@echo "$(GREEN)Compilando para release...$(NC)"
	@mkdir -p release
	@for example in $(EXAMPLES); do \
		if [ -d "$$example" ] && [ -f "$$example/main.go" ]; then \
			echo "Release build $$example..."; \
			cd $$example && go build -ldflags="-s -w" -o ../release/$$example . && cd ..; \
		fi; \
	done
	@echo "$(GREEN)✓ Release builds concluídos$(NC)"

# Informações do sistema
info:
	@echo "$(GREEN)Informações do Sistema:$(NC)"
	@echo "Go version: $$(go version)"
	@echo "GOOS: $$(go env GOOS)"
	@echo "GOARCH: $$(go env GOARCH)"
	@echo "GOPATH: $$(go env GOPATH)"
	@echo "GOROOT: $$(go env GOROOT)"
	@echo "Examples: $(EXAMPLES)"
	@echo "Coverage target: $(COVERAGE_TARGET)%"
