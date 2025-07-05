# Makefile para o ISIS Golang Tracer

.PHONY: help build test bench clean setup-env start-env stop-env examples lint

# Variáveis
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Help
help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build
build: ## Compila o código
	@echo "🔨 Compilando..."
	go build -v ./tracer/...

# Test
test: ## Executa os testes
	@echo "🧪 Executando testes..."
	go test -v ./tracer/...

# Benchmark
bench: ## Executa benchmarks
	@echo "⚡ Executando benchmarks..."
	go test -bench=. -benchmem ./tracer/...

# Test com coverage
test-coverage: ## Executa testes com coverage
	@echo "📊 Executando testes com coverage..."
	go test -coverprofile=coverage.out ./tracer/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report gerado em coverage.html"

# Clean
clean: ## Limpa arquivos temporários
	@echo "🧹 Limpando..."
	go clean -testcache
	rm -f coverage.out coverage.html

# Setup do ambiente de desenvolvimento
setup-env: ## Configura o ambiente de desenvolvimento
	@echo "🚀 Configurando ambiente..."
	@if command -v docker > /dev/null 2>&1; then \
		echo "✅ Docker encontrado"; \
	else \
		echo "❌ Docker não encontrado. Instale o Docker primeiro."; \
		exit 1; \
	fi
	@if [ -f ./tracer/setup.sh ]; then \
		chmod +x ./tracer/setup.sh; \
		cd tracer && ./setup.sh; \
	else \
		echo "❌ Script setup.sh não encontrado"; \
		exit 1; \
	fi

# Inicia o ambiente
start-env: ## Inicia os serviços de tracing
	@echo "🐳 Iniciando ambiente..."
	cd tracer && docker-compose up -d

# Para o ambiente
stop-env: ## Para os serviços de tracing
	@echo "🛑 Parando ambiente..."
	cd tracer && docker-compose down

# Logs do ambiente
logs-env: ## Mostra logs dos serviços
	@echo "📋 Mostrando logs..."
	cd tracer && docker-compose logs -f

# Status do ambiente
status-env: ## Mostra status dos serviços
	@echo "📊 Status dos serviços:"
	cd tracer && docker-compose ps

# Executa exemplos
examples: ## Executa os exemplos
	@echo "🧪 Executando exemplos..."
	go run ./tracer/examples

# Lint
lint: ## Executa linting
	@echo "🔍 Executando linting..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./tracer/...; \
	else \
		echo "⚠️  golangci-lint não encontrado. Executando go vet..."; \
		go vet ./tracer/...; \
	fi

# Formata código
fmt: ## Formata o código
	@echo "💄 Formatando código..."
	go fmt ./tracer/...

# Moderniza módulos
mod-tidy: ## Limpa e atualiza go.mod
	@echo "📦 Limpando módulos..."
	go mod tidy

# Atualiza dependências
mod-update: ## Atualiza todas as dependências
	@echo "🔄 Atualizando dependências..."
	go get -u ./tracer/...
	go mod tidy

# Verifica vulnerabilidades
security: ## Verifica vulnerabilidades
	@echo "🔒 Verificando vulnerabilidades..."
	@if command -v govulncheck > /dev/null 2>&1; then \
		govulncheck ./tracer/...; \
	else \
		echo "⚠️  govulncheck não encontrado. Instale com: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Gera documentação
docs: ## Gera documentação
	@echo "📚 Gerando documentação..."
	@if command -v godoc > /dev/null 2>&1; then \
		echo "Acesse http://localhost:6060/pkg/github.com/fsvxavier/nexs-lib/tracer/"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc não encontrado. Instale com: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Instalação completa
install: mod-tidy build test ## Instalação completa (tidy + build + test)
	@echo "✅ Instalação completa realizada com sucesso!"

# CI Pipeline
ci: lint test bench ## Pipeline de CI (lint + test + bench)
	@echo "✅ Pipeline de CI executada com sucesso!"

# Release checklist
release-check: ci test-coverage security ## Checklist de release
	@echo "✅ Checklist de release executada com sucesso!"
	@echo "📋 Arquivos de release:"
	@echo "   - coverage.html (coverage report)"
	@echo "   - ./tracer/ (código principal)"
	@echo "   - ./tracer/README.md (documentação)"
	@echo "   - ./tracer/examples/ (exemplos)"

# Debug
debug: ## Mostra informações de debug
	@echo "🐛 Informações de debug:"
	@echo "GOPATH: $(GOPATH)"
	@echo "GOBIN: $(GOBIN)"
	@echo "GOOS: $(shell go env GOOS)"
	@echo "GOARCH: $(shell go env GOARCH)"
	@echo "Go version: $(shell go version)"
	@echo "Docker version:"
	@docker --version || echo "Docker não encontrado"
	@echo "Estrutura do projeto:"
	@find ./tracer -name "*.go" | head -10

# All
all: clean mod-tidy build test examples ## Executa tudo (clean + tidy + build + test + examples)
