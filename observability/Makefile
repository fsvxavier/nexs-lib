# Nexs Observability Makefile
# Central automation for logger, tracer, and infraestructure

.PHONY: help setup test clean validate infra-up infra-down docs examples

# Default target
help: ## Show this help message
	@echo "Nexs Observability Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "infraestructure Commands:"
	@echo "  make infra-up           # Start observability infraestructure"
	@echo "  make infra-down         # Stop observability infraestructure"
	@echo "  make infra-status       # Check infraestructure status"
	@echo "  make infra-logs         # View infraestructure logs"
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev-logger         # Setup for logger development"
	@echo "  make dev-tracer         # Setup for tracer development"
	@echo "  make dev-full           # Setup complete development environment"

# Setup and validation
setup: validate ## Setup development environment
	@echo "🚀 Setting up Nexs Observability development environment..."
	@go mod tidy
	@go mod download
	@echo "✅ Setup complete!"

validate: ## Validate project structure and dependencies
	@echo "🔍 Validating Nexs Observability project..."
	@go version
	@go mod verify
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker not found. Please install Docker."; exit 1; }
	@echo "✅ Validation passed!"

# Testing
test: ## Run all tests
	@echo "🧪 Running all Nexs Observability tests..."
	@cd logger && go test -v ./...
	@cd tracer && go test -v ./...
	@echo "✅ All tests passed!"

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	@cd logger && go test -v -coverprofile=coverage.out ./...
	@cd tracer && go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=logger/coverage.out -o logger/coverage.html
	@go tool cover -html=tracer/coverage.out -o tracer/coverage.html
	@echo "✅ Coverage reports generated!"

test-integration: infra-up ## Run integration tests with real infraestructure
	@echo "🔄 Waiting for infraestructure to be ready..."
	@sleep 15
	@echo "🧪 Running integration tests..."
	@cd tracer && go test -tags=integration -v ./...
	@cd logger && go test -tags=integration -v ./...
	@echo "✅ Integration tests completed!"

# infraestructure management (delegate to infraestructure/Makefile)
infra-up: ## Start observability infraestructure
	@cd infraestructure && $(MAKE) infra-up

infra-down: ## Stop observability infraestructure
	@cd infraestructure && $(MAKE) infra-down

infra-status: ## Show infraestructure status
	@cd infraestructure && $(MAKE) infra-status

infra-logs: ## Show infraestructure logs
	@cd infraestructure && $(MAKE) infra-logs

infra-clean: ## Clean infraestructure
	@cd infraestructure && $(MAKE) infra-clean

infra-reset: ## Reset infraestructure
	@cd infraestructure && $(MAKE) infra-reset

infra-urls: ## Show service URLs
	@cd infraestructure && $(MAKE) infra-urls

infra-health: ## Check infraestructure health
	@cd infraestructure && $(MAKE) infra-health

# Development workflows
dev-logger: ## Setup environment for logger development
	@echo "📝 Setting up logger development environment..."
	@cd infraestructure && $(MAKE) infra-up GROUP=logger
	@cd infraestructure && $(MAKE) infra-up GROUP=databases
	@echo "✅ Logger development environment ready!"
	@echo "   Kibana: http://localhost:5601"
	@echo "   Elasticsearch: http://localhost:9200"

dev-tracer: ## Setup environment for tracer development
	@echo "📊 Setting up tracer development environment..."
	@cd infraestructure && $(MAKE) infra-up GROUP=tracer
	@cd infraestructure && $(MAKE) infra-up GROUP=databases
	@echo "✅ Tracer development environment ready!"
	@echo "   Jaeger: http://localhost:16686"
	@echo "   Grafana: http://localhost:3000"

dev-full: ## Setup complete development environment
	@echo "🚀 Setting up complete development environment..."
	@cd infraestructure && $(MAKE) infra-up
	@echo "✅ Complete development environment ready!"
	@cd infraestructure && $(MAKE) infra-urls

dev-stop: ## Stop development environment
	@cd infraestructure && $(MAKE) infra-down
	@echo "🛑 Development environment stopped."

# Examples
examples: ## Run all examples
	@echo "💡 Running logger examples..."
	@cd logger/examples && ./run_all_examples.sh
	@echo "💡 Running tracer examples..."
	@cd tracer/examples && for dir in */; do echo "Running $$dir..."; cd $$dir && timeout 10s go run main.go & sleep 2; kill %% 2>/dev/null || true; cd ..; done
	@echo "✅ All examples completed!"

examples-logger: ## Run logger examples
	@cd logger/examples && ./run_all_examples.sh

examples-tracer: ## Run tracer examples against infraestructure
	@echo "💡 Testing tracer examples against infraestructure..."
	@$(MAKE) dev-tracer
	@sleep 10
	@cd tracer/examples && for dir in */; do \
		echo "Testing $$dir..."; \
		cd $$dir && timeout 30s go run main.go & \
		PID=$$!; sleep 5; \
		curl -f http://localhost:8080/health >/dev/null 2>&1 && echo "✅ $$dir working" || echo "❌ $$dir failed"; \
		kill $$PID 2>/dev/null || true; \
		cd ..; \
	done

# Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@go doc -all ./logger > docs/logger-api.md 2>/dev/null || echo "Logger docs generated"
	@go doc -all ./tracer > docs/tracer-api.md 2>/dev/null || echo "Tracer docs generated"
	@echo "✅ Documentation updated!"

docs-serve: ## Serve documentation locally
	@echo "📖 Serving documentation..."
	@command -v godoc >/dev/null 2>&1 || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Documentation available at: http://localhost:6060/pkg/"
	@godoc -http=:6060

# Benchmarks and performance
bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	@cd logger && go test -bench=. -benchmem ./...
	@cd tracer && go test -bench=. -benchmem ./...

bench-integration: infra-up ## Run benchmarks against real infraestructure
	@echo "⚡ Running integration benchmarks..."
	@sleep 15
	@cd tracer && go test -bench=. -benchmem -tags=integration ./...

# Code quality
lint: ## Run linters
	@echo "🔍 Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run ./...
	@echo "✅ Linting passed!"

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || echo "goimports not available, using go fmt only"

mod-tidy: ## Tidy go modules
	@echo "📦 Tidying modules..."
	@go mod tidy
	@go mod verify

# Build and install
build: ## Build all components
	@echo "🔨 Building components..."
	@cd logger && go build ./...
	@cd tracer && go build ./...
	@echo "✅ Build successful!"

install: ## Install as Go module
	@echo "📥 Installing Nexs Observability..."
	@go install ./...

# Cleanup
clean: ## Clean build artifacts and caches
	@echo "🧹 Cleaning up..."
	@go clean -cache
	@go clean -modcache
	@rm -f logger/coverage.out logger/coverage.html
	@rm -f tracer/coverage.out tracer/coverage.html
	@cd infraestructure && $(MAKE) infra-clean
	@echo "✅ Cleanup complete!"

clean-all: clean ## Deep clean including infraestructure
	@cd infraestructure && $(MAKE) docker-prune
	@echo "✅ Deep cleanup complete!"

# Monitoring and debugging
monitor: ## Open monitoring UIs
	@echo "📊 Opening monitoring interfaces..."
	@cd infraestructure && $(MAKE) monitor-traces
	@cd infraestructure && $(MAKE) monitor-metrics
	@cd infraestructure && $(MAKE) monitor-logs

debug-logger: ## Debug logger with real infraestructure
	@$(MAKE) dev-logger
	@echo "🐛 Logger debug environment ready!"
	@echo "   View logs: make infra-logs SERVICE=kibana"
	@echo "   Test: cd logger/examples && go run basic/main.go"

debug-tracer: ## Debug tracer with real infraestructure
	@$(MAKE) dev-tracer
	@echo "🐛 Tracer debug environment ready!"
	@echo "   View traces: make infra-logs SERVICE=jaeger"
	@echo "   Test: cd tracer/examples && go run datadog/main.go"

# CI/CD support
ci-test: ## Run tests for CI environment
	@echo "🚀 Running CI tests..."
	@go mod verify
	@go vet ./...
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

ci-integration: ## Run integration tests for CI
	@echo "🚀 Running CI integration tests..."
	@$(MAKE) infra-up
	@sleep 30  # More time for CI environment
	@go test -tags=integration -v -timeout=120s ./...
	@$(MAKE) infra-down

# Release and versioning
version: ## Show version information
	@echo "Nexs Observability Library"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"

# Help for infraestructure commands
infra-help: ## Show infraestructure-specific help
	@cd infraestructure && $(MAKE) help

# Quick commands
quick-test: ## Quick test without infraestructure
	@go test ./logger/... ./tracer/...

quick-start: ## Quick start for development
	@$(MAKE) validate
	@$(MAKE) dev-full
	@echo "🎉 Quick start complete! Ready for development."

# Default goal
.DEFAULT_GOAL := help
